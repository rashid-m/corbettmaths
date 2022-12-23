package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	"path"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdb_consensus"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/txpool"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

type ShardChain struct {
	shardID   int
	multiView multiview.MultiView

	BlockGen     *BlockGenerator
	Blockchain   *BlockChain
	BlockStorage *BlockStorage
	hashHistory  *lru.Cache
	ChainName    string
	Ready        bool

	TxPool      txpool.TxPool
	TxsVerifier txpool.TxVerifier

	insertLock sync.Mutex
}

func NewShardChain(
	shardID int,
	multiView multiview.MultiView,
	blockGen *BlockGenerator,
	blockchain *BlockChain,
	chainName string,
	tp txpool.TxPool,
	tv txpool.TxVerifier,
) *ShardChain {
	cfg := config.Config()
	ffPath := path.Join(cfg.DataDir, cfg.DatabaseDir, fmt.Sprintf("shard%v", shardID), "blockstorage")
	bs := NewBlockStorage(blockchain.GetShardChainDatabase(byte(shardID)), ffPath, shardID, false)
	chain := &ShardChain{
		shardID:      shardID,
		multiView:    multiView,
		BlockGen:     blockGen,
		BlockStorage: bs,
		Blockchain:   blockchain,
		ChainName:    chainName,
		TxPool:       tp,
		TxsVerifier:  tv,
	}

	return chain
}

func (chain *ShardChain) GetInsertLock() *sync.Mutex {
	insertLock := chain.insertLock
	return &insertLock
}

func (chain *ShardChain) GetDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) GetMultiView() multiview.MultiView {
	return chain.multiView
}

func (chain *ShardChain) CloneMultiView() multiview.MultiView {
	return chain.multiView.Clone()
}

func (chain *ShardChain) SetMultiView(multiView multiview.MultiView) {
	chain.multiView = multiView
}

func (chain *ShardChain) GetFinalView() multiview.View {
	return chain.multiView.GetFinalView()
}

func (chain *ShardChain) GetBestView() multiview.View {
	return chain.multiView.GetBestView()
}

func (chain *ShardChain) GetViewByHash(hash common.Hash) multiview.View {
	return chain.multiView.GetViewByHash(hash)
}

func (chain *ShardChain) GetBestState() *ShardBestState {
	return chain.multiView.GetBestView().(*ShardBestState)
}

func (chain *ShardChain) AddView(view multiview.View) bool {
	curBestView := chain.multiView.GetBestView()
	added, err := chain.multiView.(*multiview.ShardMultiView).AddView(view)
	if (curBestView != nil) && (added == 1) {
		go func(chain *ShardChain, curBestView multiview.View) {
			sBestView := chain.GetBestState()
			if (time.Now().Unix() - sBestView.GetBlockTime()) > (int64(15 * sBestView.GetCurrentTimeSlot())) {
				return
			}
			if (curBestView.GetHash().String() != sBestView.GetHash().String()) && (chain.TxPool != nil) {
				bcHash := sBestView.GetBeaconHash()
				bcView, err := chain.Blockchain.GetBeaconViewStateDataFromBlockHash(bcHash, true, false, false)
				if err != nil {
					Logger.log.Errorf("Can not get beacon view from hash %, sView Hash %v, err %v", bcHash.String(), sBestView.GetHash().String(), err)
				} else {
					chain.TxPool.FilterWithNewView(chain.Blockchain, sBestView, bcView)
				}
			}
		}(chain, curBestView)
	}
	return err == nil
}

func (s *ShardChain) GetEpoch() uint64 {
	return s.GetBestState().Epoch
}

func (s *ShardChain) InsertBatchBlock([]types.BlockInterface) (int, error) {
	panic("implement me")
}

func (s *ShardChain) GetCrossShardState() map[byte]uint64 {

	res := make(map[byte]uint64)
	for index, key := range s.GetBestState().BestCrossShard {
		res[index] = key
	}
	return res
}

func (s *ShardChain) GetAllViewHash() (res []common.Hash) {
	for _, v := range s.multiView.GetAllViewsWithBFS() {
		res = append(res, *v.GetHash())
	}
	return
}

func (s *ShardChain) GetBestViewHeight() uint64 {
	return s.CurrentHeight()
}

func (s *ShardChain) GetFinalViewHeight() uint64 {
	return s.GetFinalView().GetHeight()
}

func (s *ShardChain) GetBestViewHash() string {
	return s.GetBestState().BestBlockHash.String()
}

func (s *ShardChain) GetFinalViewHash() string {
	return s.GetFinalView().(*ShardBestState).BestBlockHash.String()
}
func (chain *ShardChain) GetLastBlockTimeStamp() int64 {
	return chain.GetBestState().BestBlock.Header.Timestamp
}

func (chain *ShardChain) GetMinBlkInterval() time.Duration {
	return chain.GetBestState().BlockInterval
}

func (chain *ShardChain) GetMaxBlkCreateTime() time.Duration {
	return chain.GetBestState().BlockMaxCreateTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Ready
}

func (chain *ShardChain) SetReady(ready bool) {
	chain.Ready = ready
}

func (chain *ShardChain) CurrentHeight() uint64 {
	return chain.GetBestState().BestBlock.Header.Height
}

func (chain *ShardChain) GetCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.GetBestState().shardCommitteeState.GetShardCommittee()...)
}

func (chain *ShardChain) GetLastCommittee() []incognitokey.CommitteePublicKey {
	v := chain.multiView.GetViewByHash(*chain.GetBestView().GetPreviousHash())
	if v == nil {
		return nil
	}
	result := []incognitokey.CommitteePublicKey{}
	return append(result, v.GetCommittee()...)
}

func (chain *ShardChain) GetCommitteeByHeight(h uint64) ([]incognitokey.CommitteePublicKey, error) {
	bcStateRootHash := chain.Blockchain.GetBeaconBestState().ConsensusStateDBRootHash
	bcDB := chain.Blockchain.GetBeaconChainDatabase()
	bcStateDB, err := statedb.NewWithPrefixTrie(bcStateRootHash, statedb.NewDatabaseAccessWarper(bcDB))
	if err != nil {
		return nil, err
	}
	return statedb.GetOneShardCommittee(bcStateDB, byte(chain.shardID)), nil
}

func (chain *ShardChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.GetBestState().shardCommitteeState.GetShardSubstitute()...)
}

func (chain *ShardChain) GetCommitteeSize() int {
	return len(chain.GetBestState().shardCommitteeState.GetShardCommittee())
}

func (chain *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.GetBestState().shardCommitteeState.GetShardCommittee() {
		if key.GetMiningKeyBase58(chain.GetBestState().ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *ShardChain) GetLastProposerIndex() int {
	return chain.GetBestState().ShardProposerIdx
}

func (chain *ShardChain) CreateNewBlock(
	version int, proposer string, round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash) (types.BlockInterface, error) {
	Logger.log.Infof("Begin Start New Block Shard %+v", time.Now())
	curView := chain.GetBestState()
	newBlock, err := chain.Blockchain.NewBlockShard(
		curView,
		version, proposer, round,
		startTime, committees, committeeViewHash)
	Logger.log.Infof("Finish New Block Shard %+v", time.Now())
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	if version >= types.MULTI_VIEW_VERSION {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}
	newBlock.Header.FinalityHeight = 0
	if version >= types.LEMMA2_VERSION {
		previousBlock, err := chain.GetBlockByHash(newBlock.Header.PreviousBlockHash)
		if err != nil {
			fmt.Println("Cannot find block", newBlock.Header.PreviousBlockHash)
			return nil, err
		}
		prevShardBlk, ok := previousBlock.(*types.ShardBlock)
		if !ok {
			return nil, errors.New("Can not get shard block")
		}
		previousProposeTimeSlot := curView.CalculateTimeSlot(prevShardBlk.GetProposeTime())
		previousProduceTimeSlot := curView.CalculateTimeSlot(prevShardBlk.GetProduceTime())
		currentTimeSlot := curView.CalculateTimeSlot(newBlock.Header.ProposeTime)

		// if previous block is finalized or same propose/produce timeslot
		// and current block is produced/proposed next block time
		if newBlock.Header.Timestamp == newBlock.Header.ProposeTime &&
			newBlock.Header.Producer == newBlock.Header.Proposer &&
			previousProposeTimeSlot+1 == currentTimeSlot {
			if version >= types.INSTANT_FINALITY_VERSION {
				if previousBlock.GetFinalityHeight() != 0 || previousProposeTimeSlot == previousProduceTimeSlot {
					newBlock.Header.FinalityHeight = newBlock.Header.Height
				}
			} else {
				newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
			}
		}
	}

	Logger.log.Infof("[dcs] new block header proposer %v proposerTime %v", newBlock.Header.Proposer, newBlock.Header.ProposeTime)

	Logger.log.Infof("Finish Create New Block")
	return newBlock, nil
}

func (chain *ShardChain) CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64, isValidRePropose bool) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.ShardBlock)
	json.Unmarshal(b, &newBlock)

	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime
	version := newBlock.Header.Version
	newBlock.Header.FinalityHeight = 0

	if version >= types.LEMMA2_VERSION {
		// if previous block is finality or same produce/propose
		// and valid lemma2
		previousBlock, err := chain.GetBlockByHash(newBlock.Header.PreviousBlockHash)
		if err != nil {
			return nil, err
		}
		if isValidRePropose {
			if version >= types.INSTANT_FINALITY_VERSION {
				curView := chain.GetBestView().(*ShardBestState)
				previousProposeTimeSlot := curView.CalculateTimeSlot(previousBlock.GetProposeTime())
				previousProduceTimeSlot := curView.CalculateTimeSlot(previousBlock.GetProduceTime())
				if previousBlock.GetFinalityHeight() != 0 || previousProposeTimeSlot == previousProduceTimeSlot {
					newBlock.Header.FinalityHeight = newBlock.Header.Height
				}
			} else {
				newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
			}
		}
	}

	return newBlock, nil
}

// func (chain *ShardChain) ValidateAndInsertBlock(block common.BlockInterface) error {
// 	//@Bahamoot review later
// 	chain.lock.Lock()
// 	defer chain.lock.Unlock()
// 	var shardBestState ShardBestState
// 	shardBlock := block.(*ShardBlock)
// 	shardBestState.cloneShardBestStateFrom(chain.BestState)
// 	producerPublicKey := shardBlock.Header.Producer
// 	producerPosition := (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
// 	tempProducer := shardBestState.ShardCommittee[producerPosition].GetMiningKeyBase58(shardBestState.ConsensusAlgorithm)
// 	if strings.Compare(tempProducer, producerPublicKey) != 0 {
// 		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
// 	}
// 	if err := chain.ValidateBlockSignatures(block, shardBestState.ShardCommittee); err != nil {
// 		return err
// 	}
// 	return chain.Blockchain.InsertShardBlock(shardBlock, false)
// }

func (chain *ShardChain) ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey, numOfFixNode int) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committees, numOfFixNode); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) InsertBlock(block types.BlockInterface, shouldValidate bool) error {
	err := chain.Blockchain.InsertShardBlock(block.(*types.ShardBlock), shouldValidate)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	return nil
}

func (chain *ShardChain) InsertAndBroadcastBlock(block types.BlockInterface) error {

	go chain.Blockchain.config.Server.PushBlockToAll(block, "", false)

	if err := chain.InsertBlock(block, false); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) InsertWithPrevValidationData(block types.BlockInterface, newValidationData string) error {

	if newValidationData != "" {
		linkView := chain.GetMultiView().GetViewByHash(block.GetPrevHash())
		if linkView == nil {
			return errors.New("InsertWithPrevValidationData fail! Cannot find previous block hash" + block.GetPrevHash().String())
		}
		if err := chain.ReplacePreviousValidationData(block.GetPrevHash(), *linkView.GetBlock().(*types.ShardBlock).ProposeHash(), nil, newValidationData); err != nil {
			return err
		}
	}

	if err := chain.InsertBlock(block, true); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) InsertAndBroadcastBlockWithPrevValidationData(block types.BlockInterface, newValidationData string) error {

	go chain.Blockchain.config.Server.PushBlockToAll(block, newValidationData, false)

	return chain.InsertWithPrevValidationData(block, newValidationData)
}

// this get consensus data for beacon
func (chain *ShardChain) GetBlockConsensusData() map[int]types.BlockConsensusData {
	consensusData := map[int]types.BlockConsensusData{}
	bestViewBlock := chain.multiView.GetBestView().GetBlock().(*types.ShardBlock)
	consensusData[chain.shardID] = types.BlockConsensusData{
		BlockHash:      *bestViewBlock.Hash(),
		BlockHeight:    bestViewBlock.GetHeight(),
		FinalityHeight: bestViewBlock.GetFinalityHeight(),
		Proposer:       bestViewBlock.GetProposer(),
		ProposerTime:   bestViewBlock.GetProposeTime(),
		ValidationData: bestViewBlock.ValidationData,
	}

	blk, _, err := chain.Blockchain.BeaconChain.BlockStorage.GetBlock(*chain.Blockchain.BeaconChain.multiView.GetExpectedFinalView().GetHash())
	if err != nil {
		panic(err)
	}
	beaconBlk := blk.(*types.BeaconBlock)
	consensusData[-1] = types.BlockConsensusData{
		BlockHash:      *beaconBlk.Hash(),
		BlockHeight:    beaconBlk.GetHeight(),
		FinalityHeight: beaconBlk.GetFinalityHeight(),
		Proposer:       beaconBlk.GetProposer(),
		ProposerTime:   beaconBlk.GetProposeTime(),
		ValidationData: beaconBlk.ValidationData,
	}

	return consensusData
}

// this is only call when insert block successfully, the previous block is replace
func (chain *ShardChain) ReplacePreviousValidationData(previousBlockHash common.Hash, previousProposeHash common.Hash, _ []incognitokey.CommitteePublicKey, newValidationData string) error {
	if hasBlock := chain.BlockStorage.IsExisted(previousBlockHash); !hasBlock {
		// This block is not inserted yet, no need to replace
		Logger.log.Errorf("Replace previous validation data fail! Cannot find find block in db " + previousBlockHash.String())
		return nil
	}

	blk, err := chain.GetBlockByHash(previousBlockHash)
	if err != nil {
		return NewBlockChainError(ReplacePreviousValidationDataError, err)
	}
	shardBlock := blk.(*types.ShardBlock)

	if !previousProposeHash.IsEqual(shardBlock.ProposeHash()) {
		Logger.log.Errorf("Replace previous validation data fail! Propose hash not correct, data for" + previousProposeHash.String() + " got " + shardBlock.ProposeHash().String())
		return nil
	}

	decodedOldValidationData, err := consensustypes.DecodeValidationData(shardBlock.ValidationData)
	if err != nil {
		return NewBlockChainError(ReplacePreviousValidationDataError, err)
	}

	decodedNewValidationData, err := consensustypes.DecodeValidationData(newValidationData)
	if err != nil {
		return NewBlockChainError(ReplacePreviousValidationDataError, err)
	}

	if len(decodedNewValidationData.ValidatiorsIdx) > len(decodedOldValidationData.ValidatiorsIdx) {
		Logger.log.Infof("SHARD %+v | Shard Height %+v, Replace Previous ValidationData new number of signatures %+v (old %+v)",
			shardBlock.Header.ShardID, shardBlock.Header.Height, len(decodedNewValidationData.ValidatiorsIdx), len(decodedOldValidationData.ValidatiorsIdx))
	} else {
		return nil
	}

	// validate block before rewrite to
	replaceBlockHash := *shardBlock.Hash()
	shardBlock.ValidationData = newValidationData
	committees, err := chain.GetCommitteeV2(shardBlock)
	if err != nil {
		return err
	}
	if err = chain.ValidateBlockSignatures(shardBlock, committees, chain.GetBestView().GetProposerLength()); err != nil {
		return err
	}
	//rewrite to database
	if err = chain.BlockStorage.ReplaceBlock(shardBlock); err != nil {
		return err
	}
	//update multiview
	view := chain.multiView.GetViewByHash(replaceBlockHash)
	if view != nil {
		view.ReplaceBlock(shardBlock)
	} else {
		fmt.Println("Cannot find shard view", replaceBlockHash.String())
	}
	return nil
}

// consensusData contain beacon finality consensus data
func (chain *ShardChain) VerifyFinalityAndReplaceBlockConsensusData(consensusData types.BlockConsensusData) error {
	replaceBlockHash := consensusData.BlockHash
	//retrieve block from database and replace consensus field
	blk, err := chain.GetBlockByHash(replaceBlockHash)
	if blk == nil {
		return fmt.Errorf("Shard %v Cannot find shard block %v", chain.shardID, replaceBlockHash.String())
	}
	shardBlk := blk.(*types.ShardBlock)
	if shardBlk.GetVersion() < types.INSTANT_FINALITY_VERSION {
		return nil
	}
	shardBlk.Header.Proposer = consensusData.Proposer
	shardBlk.Header.ProposeTime = consensusData.ProposerTime
	shardBlk.Header.FinalityHeight = consensusData.FinalityHeight
	shardBlk.ValidationData = consensusData.ValidationData

	// validate block before rewrite to
	committees, err := chain.GetCommitteeV2(shardBlk)
	if err != nil {
		return err
	}
	if err = chain.ValidateBlockSignatures(shardBlk, committees, chain.GetBestView().GetProposerLength()); err != nil {
		return err
	}

	//replace block if improve finality
	if ok, err := chain.multiView.ReplaceBlockIfImproveFinality(shardBlk); !ok {
		return err
	}
	b, _ := json.Marshal(consensusData)
	Logger.log.Info("Replace shard block improving finality", chain.shardID, string(b))

	//rewrite to database
	if err = chain.BlockStorage.StoreBlock(shardBlk); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) GetBlockByHash(hash common.Hash) (types.BlockInterface, error) {
	block, _, err := chain.BlockStorage.GetBlock(hash)
	return block, err
}

func (chain *ShardChain) CheckExistedBlk(block types.BlockInterface) bool {
	blkHash := block.Hash()
	return chain.BlockStorage.IsExisted(*blkHash)
}

func (chain *ShardChain) GetActiveShardNumber() int {
	return 0
}

func (chain *ShardChain) GetChainName() string {
	return chain.ChainName
}

func (chain *ShardChain) GetConsensusType() string {
	return chain.GetBestState().ConsensusAlgorithm
}

func (chain *ShardChain) GetShardID() int {
	return chain.shardID
}

func (chain *ShardChain) IsBeaconChain() bool {
	return false
}

func (chain *ShardChain) UnmarshalBlock(blockString []byte) (types.BlockInterface, error) {
	var shardBlk types.ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return &shardBlk, nil
}

func (chain *ShardChain) ValidatePreSignBlock(block types.BlockInterface, signingCommittees, committees []incognitokey.CommitteePublicKey) error {
	return chain.Blockchain.VerifyPreSignShardBlock(block.(*types.ShardBlock), signingCommittees, committees, byte(block.(*types.ShardBlock).GetShardID()))
}

func (chain *ShardChain) GetAllView() []multiview.View {
	return chain.multiView.GetAllViewsWithBFS()
}

func (chain *ShardChain) GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams {
	return chain.Blockchain.GetPortalParamsV4(beaconHeight)
}

// CommitteesV2 get committees by block for shardChain
// Input block must be ShardBlock
func (chain *ShardChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	var isShardView bool
	var shardView *ShardBestState
	shardView, isShardView = chain.GetViewByHash(block.GetPrevHash()).(*ShardBestState)
	if !isShardView {
		shardView = chain.GetBestState()
	}
	shardBlock, isShardBlock := block.(*types.ShardBlock)
	if !isShardBlock {
		return []incognitokey.CommitteePublicKey{}, fmt.Errorf("Shard Chain NOT insert Shard Block Types")
	}
	_, signingCommittees, err := shardView.getSigningCommittees(shardBlock, chain.Blockchain)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, err
	}
	return signingCommittees, nil
}

func (chain *ShardChain) CommitteeStateVersion() int {
	return chain.GetBestState().shardCommitteeState.Version()
}

// BestViewCommitteeFromBlock ...
func (chain *ShardChain) BestViewCommitteeFromBlock() common.Hash {
	return chain.GetBestState().CommitteeFromBlock()
}

func (chain *ShardChain) GetChainDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) CommitteeEngineVersion() int {
	return chain.multiView.GetBestView().CommitteeStateVersion()
}

// ProposerByTimeSlot ...
func (chain *ShardChain) GetProposerByTimeSlotFromCommitteeList(ts int64, committees []incognitokey.CommitteePublicKey) (incognitokey.CommitteePublicKey, int) {
	proposer, proposerIndex := GetProposer(
		ts,
		committees,
		chain.GetBestState().GetProposerLength(),
	)
	return proposer, proposerIndex
}

func (chain *ShardChain) GetSigningCommittees(
	proposerIndex int, committees []incognitokey.CommitteePublicKey, blockVersion int,
) []incognitokey.CommitteePublicKey {
	res := []incognitokey.CommitteePublicKey{}
	if blockVersion >= types.BLOCK_PRODUCINGV3_VERSION && blockVersion < types.INSTANT_FINALITY_VERSION_V2 {
		res = FilterSigningCommitteeV3(committees, proposerIndex)
	} else {
		res = append(res, committees...)
	}
	return res
}

func (chain *ShardChain) GetFinalityProof(hash common.Hash) (*types.ShardBlock, map[string]interface{}, error) {

	shardBlock, err := chain.GetBlockByHash(hash)
	if err != nil {
		return nil, nil, err
	}

	m, err := rawdb_consensus.GetShardFinalityProof(rawdb_consensus.GetConsensusDatabase(), byte(chain.shardID), hash)
	if err != nil {
		return nil, nil, err
	}

	return shardBlock.(*types.ShardBlock), m, nil
}

//
//func (chain *ShardChain) CalculateTimeSlot(curTime int64) int64 {
//	return chain.GetBestView().(*ShardBestState).TSManager.calculateTimeslot(curTime)
//}

//func (chain *ShardChain) UpdateArchorTime(beaconHeight uint64, shardBlock *types.ShardBlock) {
//	timeSlot := chain.CalculateTimeSlot(beaconHeight, shardBlock.GetProduceTime())
//	archorTime := chain.archorTime
//	if _, ok := archorTime.archorMap[beaconHeight]; ok {
//		return
//	}
//	archorTime.heights = append(archorTime.heights, beaconHeight)
//	archorTime.archorMap[beaconHeight] = struct {
//		timeLock int64
//		timeSlot int64
//	}{
//		timeLock: shardBlock.GetProduceTime(),
//		timeSlot: timeSlot,
//	}
//}

//func (chain *ShardChain) InitArchorTime() {
//	chain.archorTime = struct {
//		archorMap map[uint64]struct {
//			timeLock int64
//			timeSlot int64
//		}
//		heights []uint64
//	}{
//		archorMap: map[uint64]struct {
//			timeLock int64
//			timeSlot int64
//		}{},
//		heights: []uint64{},
//	}
//	chain.archorTime.heights = []uint64{0}
//	chain.archorTime.archorMap[0] = struct {
//		timeLock int64
//		timeSlot int64
//	}{
//		timeLock: 0,
//		timeSlot: 0,
//	}
//}
