package blockchain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

/*
	This hot fix include:
	- init build process when init chain state
	- build process implementation (proecssnewstakingtx.go)
	- database access (accessor_newstakingtx.go)
	- replace process when updateshardbeststate
*/

const NEWSTAKINGTX_HEIGHT_SWITCH = 1000000

func (blockchain *BlockChain) buildNewStakingTx() {
	bDB := blockchain.GetDatabase()
	stakingInfo, err := rawdbv2.GetMapStakingTxNew(bDB)

	//no data in database => init one
	if err != nil {
		stakingInfo = &rawdbv2.StakingTXInfo{
			MStakingTX: make(map[int]map[string]string), //shardID -> (committee->txid)
			Height:     2,
		}
		for i := 0; i < blockchain.GetActiveShardNumber(); i++ {
			stakingInfo.MStakingTX[i] = make(map[string]string)
		}
	}

	//fetch each beacon block and process stakingtx for all shard, until we get checkpoint
	for {
		nextRequestHeight := stakingInfo.Height + 500
		//if > checkpoint
		if nextRequestHeight > NEWSTAKINGTX_HEIGHT_SWITCH {
			nextRequestHeight = NEWSTAKINGTX_HEIGHT_SWITCH
		}

		//if > current beacon height (only fetch to the best block)
		if nextRequestHeight > blockchain.GetBeaconHeight() {
			nextRequestHeight = blockchain.GetBeaconHeight()
		}

		fmt.Println("NEWTX: get blocks", stakingInfo.Height+1, nextRequestHeight)

		//if beacon dont have new block, wait 1 second
		if nextRequestHeight < stakingInfo.Height+1 {
			time.Sleep(time.Second)
			continue
		}

		blocks, err := FetchBeaconBlockFromHeight(bDB, stakingInfo.Height+1, nextRequestHeight)
		if err != nil {
			Logger.log.Error(err)
			panic(err)
		}

		//process beacon blocks
		for _, block := range blocks {
			newMap, err := blockchain.processStakingTxFromBeaconBlock(stakingInfo.MStakingTX, block)
			if err != nil {
				Logger.log.Error(err)
				panic(err)
			}
			stakingInfo.Height = block.GetHeight()
			stakingInfo.MStakingTX = newMap

			//backup at every 100 processed block if less than checkpoint height
			if block.GetHeight() < NEWSTAKINGTX_HEIGHT_SWITCH && block.GetHeight()%500 == 0 {
				err := rawdbv2.StoreMapStakingTxNew(bDB, block.GetHeight(), newMap)
				if err != nil {
					Logger.log.Error(err)
					panic("Store stakingtx map error")
				}
			}

			//if we reach checkpoint, store map and return
			if block.GetHeight() == NEWSTAKINGTX_HEIGHT_SWITCH {
				err := rawdbv2.StoreMapStakingTxNew(bDB, block.GetHeight(), newMap)
				if err != nil {
					Logger.log.Error(err)
					panic("Store stakingtx map error")
				}
				return
			}
		}

	}
}

func (blockchain *BlockChain) processStakingTxFromBeaconBlock(curMap map[int]map[string]string, bcBlk *BeaconBlock) (map[int]map[string]string, error) {
	beaconConsensusRootHash, err := blockchain.GetBeaconConsensusRootHash(blockchain.GetDatabase(), bcBlk.GetHeight())
	if err != nil {
		return nil, fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found ,error %+v", bcBlk.GetHeight(), err)
	}
	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	if err != nil {
		return nil, err
	}
	_, autoStaking := statedb.GetRewardReceiverAndAutoStaking(beaconConsensusStateDB, blockchain.GetShardIDs())
	for _, l := range bcBlk.Body.Instructions {
		switch l[0] {
		case SwapAction:
			for _, outPublicKey := range strings.Split(l[2], ",") {
				// If out public key has auto staking then ignore this public key
				res, ok := autoStaking[outPublicKey]
				if ok && res {
					continue
				}
				sid, err := strconv.Atoi(l[4])
				if err != nil {
					panic(err)
				}
				for k := range curMap {
					delete(curMap[k], outPublicKey)
				}
			}
		case StakeAction:
			switch l[2] {
			case "shard":
				shard := strings.Split(l[1], ",")
				newShardCandidates := []string{}
				newShardCandidates = append(newShardCandidates, shard...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						shardID, _, _, _, err := blockchain.GetTransactionByHash(*txHash)
						if err != nil {
							continue
						}
						// update stakingtx for this transaction
						curMap[int(shardID)][newShardCandidates[i]] = v
					}
				}
			default:
				continue
			}
		default:
			continue
		}
	}
	return curMap, nil
}
