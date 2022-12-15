package syncker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2"
	"os"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/wire"
)

type BeaconPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type BeaconSyncProcess struct {
	status              string //stop, running
	isCommittee         bool
	isCatchUp           bool
	beaconPeerStates    map[string]BeaconPeerState //sender -> state
	beaconPeerStateCh   chan *wire.MessagePeerState
	blockchain          *blockchain.BlockChain
	network             Network
	chain               Chain
	beaconPool          *BlkPool
	actionCh            chan func()
	lastCrossShardState map[byte]map[byte]uint64
	lastInsert          string
	consensus           peerv2.ConsensusData
}

func NewBeaconSyncProcess(network Network, consensus peerv2.ConsensusData, bc *blockchain.BlockChain, chain BeaconChainInterface) *BeaconSyncProcess {

	var isOutdatedBlock = func(blk interface{}) bool {
		if blk.(*types.BeaconBlock).GetHeight() < chain.GetFinalViewHeight() {
			return true
		}
		return false
	}

	s := &BeaconSyncProcess{
		status:              STOP_SYNC,
		blockchain:          bc,
		network:             network,
		chain:               chain,
		beaconPool:          NewBlkPool("BeaconPool", isOutdatedBlock),
		beaconPeerStates:    make(map[string]BeaconPeerState),
		beaconPeerStateCh:   make(chan *wire.MessagePeerState, 100),
		actionCh:            make(chan func()),
		lastCrossShardState: make(map[byte]map[byte]uint64),
		consensus:           consensus,
	}
	go s.syncBeacon()
	go s.insertBeaconBlockFromPool()
	go s.updateConfirmCrossShard()
	go s.syncFinishSyncMessage()
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		lastHeight := s.chain.GetBestViewHeight()
		for {
			select {
			case f := <-s.actionCh:
				f()
			case beaconPeerState := <-s.beaconPeerStateCh:
				Logger.Debugf("Got new beacon peerstate, last height %v", beaconPeerState.Beacon.Height)
				s.beaconPeerStates[beaconPeerState.SenderID] = BeaconPeerState{
					Timestamp:      beaconPeerState.Timestamp,
					BestViewHash:   beaconPeerState.Beacon.BlockHash.String(),
					BestViewHeight: beaconPeerState.Beacon.Height,
				}
				s.chain.SetReady(true)
			case <-ticker.C:
				for sender, ps := range s.beaconPeerStates {
					if ps.Timestamp < time.Now().Unix()-20 {
						delete(s.beaconPeerStates, sender)
					}
				}
				if lastHeight != s.chain.GetBestViewHeight() {
					s.lastInsert = time.Now().Format("2006-01-02T15:04:05-0700")
					lastHeight = s.chain.GetBestViewHeight()
				}

			}
			if s.status != RUNNING_SYNC {
				time.Sleep(time.Second)
				continue
			}
		}
	}()

	return s
}

func (s *BeaconSyncProcess) start() {
	if s.status == RUNNING_SYNC {
		return
	}
	s.status = RUNNING_SYNC

}

func (s *BeaconSyncProcess) stop() {
	s.status = STOP_SYNC
}

//helper function to access map in atomic way
func (s *BeaconSyncProcess) getBeaconPeerStates() map[string]BeaconPeerState {
	res := make(chan map[string]BeaconPeerState)
	s.actionCh <- func() {
		ps := make(map[string]BeaconPeerState)
		for k, v := range s.beaconPeerStates {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

type LastCrossShardBeaconProcess struct {
	BeaconHeight        uint64
	LastCrossShardState map[byte]map[byte]uint64
}

//watching confirm beacon block and update cross shard info (which beacon confirm crossshard block N of shard X)
func (s *BeaconSyncProcess) updateConfirmCrossShard() {
	state := rawdbv2.GetLastBeaconStateConfirmCrossShard(s.chain.GetDatabase())
	lastBeaconStateConfirmCrossX := new(LastCrossShardBeaconProcess)
	_ = json.Unmarshal(state, &lastBeaconStateConfirmCrossX)
	lastBeaconHeightConfirmCrossX := uint64(1)
	if lastBeaconStateConfirmCrossX.BeaconHeight != 0 {
		s.lastCrossShardState = lastBeaconStateConfirmCrossX.LastCrossShardState
		lastBeaconHeightConfirmCrossX = lastBeaconStateConfirmCrossX.BeaconHeight
	}
	fmt.Println("lastBeaconHeightConfirmCrossX", lastBeaconHeightConfirmCrossX)
	for {
		if s.status != RUNNING_SYNC {
			time.Sleep(time.Second)
			continue
		}
		if lastBeaconHeightConfirmCrossX > s.chain.GetFinalViewHeight() {
			//fmt.Println("DEBUG:larger than final view", s.chain.GetFinalViewHeight())
			time.Sleep(time.Second * 5)
			continue
		}
		beaconBlock, err := s.blockchain.FetchConfirmBeaconBlockByHeight(lastBeaconHeightConfirmCrossX)
		if err != nil || beaconBlock == nil {
			//fmt.Println("DEBUG: cannot find beacon block", lastBeaconHeightConfirmCrossX)
			time.Sleep(time.Second * 5)
			continue
		}
		err = processBeaconForConfirmmingCrossShard(s.chain.GetDatabase(), beaconBlock, s.lastCrossShardState)
		if err == nil {
			lastBeaconHeightConfirmCrossX++
			if lastBeaconHeightConfirmCrossX%1000 == 0 {
				Logger.Info("store lastBeaconHeightConfirmCrossX", lastBeaconHeightConfirmCrossX)
				rawdbv2.StoreLastBeaconStateConfirmCrossShard(s.chain.GetDatabase(), LastCrossShardBeaconProcess{lastBeaconHeightConfirmCrossX, s.lastCrossShardState})
			}
		} else {
			fmt.Println(err)
			time.Sleep(time.Second * 5)
		}

	}
}

func processBeaconForConfirmmingCrossShard(database incdb.Database, beaconBlock *types.BeaconBlock, lastCrossShardState map[byte]map[byte]uint64) error {
	if beaconBlock != nil && beaconBlock.Body.ShardState != nil {
		for fromShard, shardBlocks := range beaconBlock.Body.ShardState {
			for _, shardBlock := range shardBlocks {
				for _, toShard := range shardBlock.CrossShard {

					if fromShard == toShard {
						continue
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
					waitHeight := shardBlock.Height

					info := blockchain.NextCrossShardInfo{
						waitHeight,
						shardBlock.Hash.String(),
						beaconBlock.GetHeight(),
						beaconBlock.Hash().String(),
					}
					//Logger.Info("DEBUG: processBeaconForConfirmmingCrossShard ", fromShard, toShard, info)
					b, _ := json.Marshal(info)
					Logger.Info("debug StoreCrossShardNextHeight", fromShard, toShard, lastHeight, string(b))
					err := rawdbv2.StoreCrossShardNextHeight(database, fromShard, toShard, lastHeight, b)
					if err != nil {
						return err
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
				}
			}
		}
	}
	return nil
}

//periodically check pool and insert into pool (in case some fork block or block come early)
var insertBeaconTimeCache, _ = lru.New(1000)

func (s *BeaconSyncProcess) insertBeaconBlockFromPool() {

	insertCnt := 0
	defer func() {
		if insertCnt > 0 {
			s.insertBeaconBlockFromPool()
		} else {
			time.AfterFunc(time.Second*2, s.insertBeaconBlockFromPool)
		}
	}()
	//Logger.Debugf("insertBeaconBlockFromPool Start")
	//loop all current views, if there is any block connect to the view
	for _, viewHash := range s.chain.GetAllViewHash() {
		blks := s.beaconPool.GetBlockByPrevHash(viewHash)
		for _, blk := range blks {
			if blk == nil {
				continue
			}
			//Logger.Debugf("insertBeaconBlockFromPool blk %v %v", blk.GetHeight(), blk.Hash().String())
			//if already insert and error, last time insert is < 10s then we skip
			insertTime, ok := insertBeaconTimeCache.Get(viewHash.String())
			if ok && time.Since(insertTime.(time.Time)).Seconds() < 10 {
				continue
			}

			insertBeaconTimeCache.Add(viewHash.String(), time.Now())
			insertCnt++
			//must validate this block when insert

			if err := s.chain.InsertBlock(blk.(types.BlockInterface), true); err != nil {
				Logger.Error("Insert beacon block from pool fail", blk.GetHeight(), blk.Hash(), err)
				continue
			}
			s.beaconPool.RemoveBlock(blk)
		}
	}

}

//sync beacon
func (s *BeaconSyncProcess) syncBeacon() {
	for {
		requestCnt := 0
		if s.status != RUNNING_SYNC {
			s.isCatchUp = false
			time.Sleep(time.Second)
			continue
		}

		for peerID, pState := range s.getBeaconPeerStates() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		//last check, if we still need to sync more
		if requestCnt > 0 {
			s.isCatchUp = false
		} else {
			if len(s.beaconPeerStates) > 0 {
				s.isCatchUp = true
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *BeaconSyncProcess) streamFromPeer(peerID string, pState BeaconPeerState) (requestCnt int) {
	if pState.processed {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	blockBuffer := []types.BlockInterface{}
	defer func() {
		if requestCnt == 0 {
			pState.processed = true
		}
		cancel()
	}()

	toHeight := pState.BestViewHeight
	//process param

	//fullnode delay 1 block (make sure insert final block)
	if os.Getenv("FULLNODE") != "" {
		toHeight = toHeight - 1
		if toHeight <= s.chain.GetBestViewHeight() {
			return
		}
	}

	//if is behind, and
	//if peerstate show fork, sync that block
	if pState.BestViewHeight < s.chain.GetBestViewHeight() || (pState.BestViewHeight == s.chain.GetBestViewHeight() && s.chain.GetBestViewHash() == pState.BestViewHash) {
		return
	}

	if pState.BestViewHeight == s.chain.GetBestViewHeight() && s.chain.GetBestViewHash() != pState.BestViewHash {
		for _, h := range s.chain.GetAllViewHash() { //check if block exist in multiview, then return
			if h.String() == pState.BestViewHash {
				return
			}
		}
	}

	if pState.BestViewHeight > s.chain.GetBestViewHeight() {
		requestCnt++
		peerID = ""
	}

	//incase, we have long multiview chain, just sync last 100 block (very low probability that we have fork more than 100 blocks)
	fromHeight := s.chain.GetFinalViewHeight() + 1
	if s.chain.GetBestViewHeight()-100 > fromHeight {
		fromHeight = s.chain.GetBestViewHeight()
	}

	//stream
	ch, err := s.network.RequestBeaconBlocksViaStream(ctx, peerID, fromHeight, toHeight)
	if err != nil || ch == nil {
		fmt.Println("Syncker: create channel fail")
		requestCnt = 0
		return
	}

	//receive
	insertTime := time.Now()
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				blockBuffer = append(blockBuffer, blk)
			}

			if uint64(len(blockBuffer)) >= blockchain.DefaultMaxBlkReqPerPeer || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*2000)) {
				insertBlkCnt := 0
				for {
					time1 := time.Now()

					/*for _, v := range blockBuffer {*/
					//Logger.Infof("[config] v height %v proposetime %v", v.GetHeight(), v.GetProposeTime())
					/*}*/

					if successBlk, err := InsertBatchBlock(s.chain, blockBuffer); err != nil {
						if successBlk == 0 {
							fmt.Println(err)
						}
						return
					} else {
						insertBlkCnt += successBlk
						Logger.Infof("Syncker Insert %d beacon block (from %d to %d) elaspse %f \n", successBlk, blockBuffer[0].GetHeight(), blockBuffer[len(blockBuffer)-1].GetHeight(), time.Since(time1).Seconds())
						if successBlk == 0 {
							return
						}
						if successBlk < len(blockBuffer) {
							blockBuffer = blockBuffer[successBlk:]
						} else {
							break
						}
					}
				}
				insertTime = time.Now()
				blockBuffer = []types.BlockInterface{}
			}
			if isNil(blk) && len(blockBuffer) == 0 {
				return
			}
		}
	}
}

func (s *BeaconSyncProcess) syncFinishSyncMessage() {
	ts := s.blockchain.BeaconChain.GetBestView().GetCurrentTimeSlot()
	sleepTime := time.Duration(ts/2) * time.Second
	for {
		validCnt := 0
		for i := 0; i < s.blockchain.GetActiveShardNumber(); i++ {
			committeeView := s.blockchain.BeaconChain.GetBestView().(*blockchain.BeaconBestState)
			if committeeView.CommitteeStateVersion() >= committeestate.STAKING_FLOW_V4 {
				shardView := s.blockchain.ShardChain[i].GetBestView().(*blockchain.ShardBestState)
				convertedTimeslot := time.Duration(ts) * time.Second
				now := time.Now().Unix()
				ceiling := now + int64(5*convertedTimeslot.Seconds())
				floor := now - int64(5*convertedTimeslot.Seconds())
				if floor <= shardView.BestBlock.Header.Timestamp &&
					shardView.BestBlock.Header.Timestamp <= ceiling {
					validCnt++
				}
			}
		}
		if validCnt == s.blockchain.GetActiveShardNumber() {
			s.trySendFinishSyncMessage()
		}
		time.Sleep(sleepTime)
	}

}

func (s *BeaconSyncProcess) trySendFinishSyncMessage() {
	committeeView := s.blockchain.BeaconChain.GetBestView().(*blockchain.BeaconBestState)
	validatorFromUserKeys, syncValidator := committeeView.ExtractFinishSyncingValidators(
		s.consensus.GetValidators(), 255)
	finishedSyncValidators := []string{}
	finishedSyncSignatures := [][]byte{}
	for i, v := range validatorFromUserKeys {
		signature, err := v.MiningKey.BriSignData([]byte(wire.CmdMsgFinishSync))
		if err != nil {
			continue
		}
		finishedSyncSignatures = append(finishedSyncSignatures, signature)
		finishedSyncValidators = append(finishedSyncValidators, syncValidator[i])
	}
	if len(finishedSyncValidators) == 0 {
		return
	}
	Logger.Infof("Send Finish Sync Message, beacon key %+v \n signature %+v", finishedSyncValidators, finishedSyncSignatures)
	msg := wire.NewMessageFinishSync(finishedSyncValidators, finishedSyncSignatures, 255)
	if err := s.network.PublishMessageToShard(msg, common.BeaconChainSyncID); err != nil {
		Logger.Errorf("trySendFinishSyncMessage Public Message to Chain %+v, error %+v", common.BeaconChainSyncID, err)
	}
}
