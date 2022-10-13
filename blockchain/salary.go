package blockchain

import (
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/transaction"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (blockchain *BlockChain) addShardRewardRequestToBeacon(beaconBlock *types.BeaconBlock, rewardStateDB *statedb.StateDB, bestState *BeaconBestState) error {
	//get shard block version that confirmed by this beacon block
	shardsBlockVersions := map[int]map[uint64]int{}
	for _, sID := range blockchain.GetShardIDs() {
		shardsBlockVersions[sID] = map[uint64]int{}
		sStates := beaconBlock.Body.ShardState[byte(sID)]
		for _, sState := range sStates {
			shardsBlockVersions[sID][sState.Height] = sState.Version
		}
	}

	for _, inst := range beaconBlock.Body.Instructions {
		if len(inst) <= 2 {
			continue
		}
		version := beaconBlock.Header.Version
		if inst[0] == instruction.ACCEPT_BLOCK_REWARD_V3_ACTION {
			acceptBlockRewardIns, err := instruction.ValidateAndImportAcceptBlockRewardV3InstructionFromString(inst)
			if err != nil {
				return err
			}
			if shardBlockVersions, ok := shardsBlockVersions[int(acceptBlockRewardIns.ShardID())]; ok {
				if blkVersion, ok := shardBlockVersions[acceptBlockRewardIns.ShardBlockHeight()]; ok {
					version = blkVersion
				}
			}
			rewardAmount, err := blockchain.GetRewardAmount(acceptBlockRewardIns.ShardID(), version, acceptBlockRewardIns.ShardBlockHeight())
			if err != nil {
				return err
			} else {
				if beaconBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
					if bestState.RewardMinted+rewardAmount > config.Param().MaxReward {
						if config.Param().MaxReward > bestState.RewardMinted {
							rewardAmount = config.Param().MaxReward - bestState.RewardMinted
						} else {
							rewardAmount = 0
						}
					}
					bestState.RewardMinted += rewardAmount
				}
			}

			acceptBlockRewardIns.TxsFee()[common.PRVCoinID] += rewardAmount

			for key, value := range acceptBlockRewardIns.TxsFee() {
				if value != 0 {
					err = statedb.AddShardRewardRequestMultiset(
						rewardStateDB,
						beaconBlock.Header.Epoch,
						acceptBlockRewardIns.ShardID(),
						acceptBlockRewardIns.SubsetID(),
						key, value)
					if err != nil {
						return err
					}
				}
			}
			continue
		}
		if instruction.IsConsensusInstruction(inst[0]) {
			continue
		}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		if metaType == instruction.ACCEPT_BLOCK_REWARD_V1_ACTION {
			acceptedBlkRewardInfo, err := instruction.NewAcceptedBlockRewardV1FromString(inst[2])
			if err != nil {
				return err
			}

			if acceptedBlkRewardInfo.TxsFee == nil {
				acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
			}
			if shardBlockVersions, ok := shardsBlockVersions[int(acceptedBlkRewardInfo.ShardID)]; ok {
				if blkVersion, ok := shardBlockVersions[acceptedBlkRewardInfo.ShardBlockHeight]; ok {
					version = blkVersion
				}
			}
			rewardAmount, err := blockchain.GetRewardAmount(acceptedBlkRewardInfo.ShardID, version, acceptedBlkRewardInfo.ShardBlockHeight)
			if err != nil {
				return err
			} else {
				Logger.log.Infof("Test reward Got %v reward for block %v of shard %v ", rewardAmount, acceptedBlkRewardInfo.ShardBlockHeight, acceptedBlkRewardInfo.ShardID)
				if beaconBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
					if bestState.RewardMinted+rewardAmount > config.Param().MaxReward {
						if config.Param().MaxReward > bestState.RewardMinted {
							rewardAmount = config.Param().MaxReward - bestState.RewardMinted
						} else {
							rewardAmount = 0
						}
					}
					bestState.RewardMinted += rewardAmount
				}
				Logger.log.Infof("\tTest reward Got %v reward after check max reward, minted %v, max %v", rewardAmount, bestState.RewardMinted, config.Param().MaxReward)
			}
			acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] += rewardAmount

			for key, value := range acceptedBlkRewardInfo.TxsFee {
				if value != 0 {
					err = statedb.AddShardRewardRequest(
						rewardStateDB,
						beaconBlock.Header.Epoch,
						acceptedBlkRewardInfo.ShardID,
						key, value)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) processSalaryInstructions(curView *ShardBestState, rewardStateDB, consensusStateDB *statedb.StateDB, beaconBlocks []*types.BeaconBlock, confirmBeaconHeight uint64, shardID byte) error {
	rewInsV2 := []metadata.BeaconRewardInfo{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {

			if len(l) <= 2 {
				continue
			}

			if l[0] == instruction.SHARD_RECEIVE_REWARD_V3_ACTION {
				shardReceiveRewardV3, err := instruction.ValidateAndImportShardReceiveRewardV3InstructionFromString(l)
				if err != nil {
					Logger.log.Debug(err)
					continue
				}
				if shardReceiveRewardV3.Epoch() != 0 {
					if confirmBeaconHeight < config.Param().ConsensusParam.EnableSlashingHeightV2 {
						cInfos, err := blockchain.GetAllCommitteeStakeInfo(shardReceiveRewardV3.Epoch())
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
						shardSubsetStakerInfo := getCommitteeToPayRewardMultiset(cInfos[int(shardReceiveRewardV3.ShardID())], shardReceiveRewardV3)
						err = blockchain.addShardCommitteeReward(rewardStateDB, shardID, shardReceiveRewardV3.Reward(), shardSubsetStakerInfo)
						if err != nil {
							return err
						}
					} else {
						cInfosV2, err := blockchain.GetAllCommitteeStakeInfoSlashingVersion(shardReceiveRewardV3.Epoch())
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
						shardSubsetStakerInfo := getCommitteeToPayRewardMultisetSlashingVersion(cInfosV2[int(shardReceiveRewardV3.ShardID())], shardReceiveRewardV3)
						beaconBestState := blockchain.BeaconChain.GetBestView().(*BeaconBestState)
						nonSlashingCInfosV2, err := beaconBestState.GetNonSlashingCommittee(shardSubsetStakerInfo, shardReceiveRewardV3.Epoch(), shardReceiveRewardV3.ShardID())
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
						err = blockchain.addShardCommitteeRewardSlashingVersion(rewardStateDB, shardID, shardReceiveRewardV3.Reward(), nonSlashingCInfosV2)
						if err != nil {
							return err
						}
					}
				}
				continue
			}

			if instruction.IsConsensusInstruction(l[0]) {
				continue
			}

			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			instType, err := strconv.Atoi(l[0])
			if err != nil {
				return NewBlockChainError(ProcessSalaryInstructionsError, err)
			}
			if beaconBlock.GetHeight() >= config.Param().ConsensusParam.StakingFlowV4Height {
				Logger.log.Infof("test salary aaaaaaaaaaaaaaaaaaaa")
				switch instType {
				case metadata.BeaconRewardRequestMeta:
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						Logger.log.Infof("test salary %v", err)
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					rewInsV2 = append(rewInsV2, *beaconBlkRewardInfo)
					continue
				}
			}
			if shardToProcess == int(shardID) {
				switch instType {
				case metadata.BeaconRewardRequestMeta:
					if beaconBlock.GetHeight() >= config.Param().ConsensusParam.StakingFlowV4Height {
						continue
					}
					beaconBlkRewardInfo, err := metadata.NewBeaconBlockRewardInfoFromStr(l[3])
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}

					for key := range beaconBlkRewardInfo.BeaconReward {
						Logger.log.Criticalf("Add Committee Reward BeaconReward, Public Key %+v, reward %+v, token %+v, shardProcess %v", beaconBlkRewardInfo.PayToPublicKey, beaconBlkRewardInfo.BeaconReward[key], key, shardID)
						err = statedb.AddCommitteeReward(rewardStateDB, beaconBlkRewardInfo.PayToPublicKey, beaconBlkRewardInfo.BeaconReward[key], key)
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
					}
					continue

				case metadata.IncDAORewardRequestMeta:
					incDAORewardInfo, err := metadata.NewIncDAORewardInfoFromStr(l[3])
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					keyWalletDevAccount, err := wallet.Base58CheckDeserialize(config.Param().IncognitoDAOAddress)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					for key := range incDAORewardInfo.IncDAOReward {
						tempPublicKey := base58.Base58Check{}.Encode(keyWalletDevAccount.KeySet.PaymentAddress.Pk, common.Base58Version)
						Logger.log.Criticalf("Add Committee Reward IncDAOReward, Public Key %+v, reward %+v, token %+v", tempPublicKey, incDAORewardInfo.IncDAOReward[key], key)
						err = statedb.AddCommitteeReward(rewardStateDB, tempPublicKey, incDAORewardInfo.IncDAOReward[key], key)
						if err != nil {
							return NewBlockChainError(ProcessSalaryInstructionsError, err)
						}
					}
					continue
				}
			}
			switch instType {
			case instruction.SHARD_RECEIVE_REWARD_V1_ACTION:
				shardRewardInfo, err := instruction.NewShardReceiveRewardV1FromString(l[3])
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
				if confirmBeaconHeight < config.Param().ConsensusParam.EnableSlashingHeightV2 {
					cInfos, err := blockchain.GetAllCommitteeStakeInfo(shardRewardInfo.Epoch)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					err = blockchain.addShardCommitteeReward(rewardStateDB, shardID, shardRewardInfo.ShardReward, cInfos[int(shardToProcess)])
					if err != nil {
						return err
					}
				} else {
					cInfosV2, err := blockchain.GetAllCommitteeStakeInfoSlashingVersion(shardRewardInfo.Epoch)
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					beaconBestState := blockchain.BeaconChain.GetBestView().(*BeaconBestState)
					nonSlashingCInfosV2, err := beaconBestState.GetNonSlashingCommittee(cInfosV2[int(shardToProcess)], shardRewardInfo.Epoch, byte(shardToProcess))
					if err != nil {
						return NewBlockChainError(ProcessSalaryInstructionsError, err)
					}
					err = blockchain.addShardCommitteeRewardSlashingVersion(rewardStateDB, shardID, shardRewardInfo.ShardReward, nonSlashingCInfosV2)
					if err != nil {
						return err
					}
				}
				continue
			}

		}
	}
	if len(rewInsV2) > 0 {
		bBestState := blockchain.GetBeaconBestState()
		bConsensusDB := bBestState.GetBeaconConsensusStateDB()
		if err := processBeaconRewardV2(curView, rewInsV2, shardID, rewardStateDB, consensusStateDB, bConsensusDB); err != nil {
			Logger.log.Infof("test salary %v", err)
			return err
		}
	}
	return nil
}

func getCommitteeToPayRewardMultiset(
	committees []*statedb.ShardStakerInfo,
	shardReceiveRewardV3 *instruction.ShardReceiveRewardV3,
) []*statedb.ShardStakerInfo {
	res := []*statedb.ShardStakerInfo{}
	for i, v := range committees {
		if i%MaxSubsetCommittees == int(shardReceiveRewardV3.SubsetID()) {
			res = append(res, v)
		}
	}
	return res
}

func processBeaconRewardV2(
	curView *ShardBestState,
	beaconBlkRewardInfos []metadata.BeaconRewardInfo,
	shardID byte,
	sRDB *statedb.StateDB,
	sCDB *statedb.StateDB,
	bCDB *statedb.StateDB, // Beacon Best state db
) error {
	sSlashed := map[byte][]string{}
	gotSlashedData := false
	sCommitteeStr := []string{}
	delegatorRewM := map[string]map[common.Hash]uint64{}
	bCommitteeRewM := map[string]map[common.Hash]uint64{}
	BLogger.log.Infof("beaconrew Process beacon reward for epoch %v, block %v, len ins %v ", curView.Epoch, curView.ShardHeight, len(beaconBlkRewardInfos))
	for _, rewIns := range beaconBlkRewardInfos {
		if !gotSlashedData {
			sSlashed = statedb.GetSlashingCommittee(bCDB, rewIns.Epoch)
			gotSlashedData = true
		}
		BLogger.log.Infof("beaconrew ins %+v ", rewIns)
		delegatorRewM[rewIns.PayToPublicKey] = rewIns.DelegatorReward
		bCommitteeRewM[rewIns.PayToPublicKey] = rewIns.BeaconReward
	}
	sDelegateInfo, ok, err := statedb.GetAllShardStakersInfo(sCDB)
	if !ok {
		return errors.Errorf("Can not found shard %v staker info", shardID)
	}

	if err != nil {
		return err
	}
	sDelegateStateM := sDelegateInfo.MapDelegate()
	sCommitteeInfoM := sDelegateInfo.IsCommittee()
	for pk, isCommittee := range sCommitteeInfoM {
		if _, ok := sDelegateStateM[pk]; ok {
			if isCommittee {
				sCommitteeStr = append(sCommitteeStr, pk)
			}
		}
	}
	for _, committees := range sSlashed {
		for _, pk := range committees {
			if _, ok := sDelegateStateM[pk]; ok {
				delete(sDelegateStateM, pk)
			}
		}
	}
	sRewardReceiverM := sDelegateInfo.RewardReceiver()
	sStakerRewM, sCommitteeRewM := calculateRewardForDelegators(shardID, curView.Epoch, curView.ShardHeight, sCommitteeStr, delegatorRewM, sDelegateStateM)
	BLogger.log.Infof("xstaker %v, shardID %v", len(sDelegateInfo.MapDelegate()), shardID)
	for k, v := range sDelegateInfo.MapDelegate() {
		BLogger.log.Infof("xstaker %v delegate to %v", k, v)
	}
	if err = addShardCommitteeRewardV2(shardID, sRDB, sStakerRewM, sCommitteeRewM, sRewardReceiverM); err != nil {
		return err
	}

	return addBeaconCommitteeRewardV2(sRDB, bCommitteeRewM, shardID)
}

func calculateRewardForDelegators(
	sID byte,
	epoch uint64,
	blockH uint64,
	sCommitee []string,
	delegatorRewM map[string]map[common.Hash]uint64,
	sDelegateStateM map[string]string,
) (
	stakerLockedRewM map[string]map[common.Hash]uint64,
	stakerAvailableRewM map[string]map[common.Hash]uint64,
) {
	stakerRewM := map[string]map[common.Hash]uint64{}
	stakerAvailableRewM = map[string]map[common.Hash]uint64{}
	stakerLockedRewM = map[string]map[common.Hash]uint64{}
	for sPK, dPK := range sDelegateStateM { //staker public key & delegate public key
		BLogger.log.Infof("testdistribute0 sID %v epoch %v-%v %v %v", sID, epoch, blockH, sPK, dPK)
		if rew, ok := delegatorRewM[dPK]; ok {
			stakerRewM[sPK] = rew
		}
	}
	for _, cPK := range sCommitee {
		if rew, ok := stakerRewM[cPK]; ok {
			stakerAvailableRewM[cPK] = rew
			delete(stakerRewM, cPK)
		}
	}
	stakerLockedRewM = stakerRewM
	return stakerLockedRewM, stakerAvailableRewM
}

func getCommitteeToPayRewardMultisetSlashingVersion(
	committees []*statedb.StakerInfoSlashingVersion,
	shardReceiveRewardV3 *instruction.ShardReceiveRewardV3,
) []*statedb.StakerInfoSlashingVersion {
	res := []*statedb.StakerInfoSlashingVersion{}
	for i, v := range committees {
		if i%MaxSubsetCommittees == int(shardReceiveRewardV3.SubsetID()) {
			res = append(res, v)
		}
	}
	return res
}

func addShardCommitteeRewardV2(
	shardID byte,
	rewardStateDB *statedb.StateDB,
	stakerLockedRewM map[string]map[common.Hash]uint64,
	stakerAvailableRewM map[string]map[common.Hash]uint64,
	rewReceiverM map[string]privacy.PaymentAddress,
) (
	err error,
) {
	for cPK, receiverAddr := range rewReceiverM {
		receiverPK := base58.Base58Check{}.Encode(receiverAddr.Pk, common.Base58Version)
		if reward, ok := stakerLockedRewM[cPK]; ok {
			if err = statedb.AddDelegatorReward(rewardStateDB, receiverPK, reward); err != nil {
				return NewBlockChainError(ProcessSalaryInstructionsError, err)
			}
		}
		if reward, ok := stakerAvailableRewM[cPK]; ok {
			if err = statedb.AddShardCommitteeReward(rewardStateDB, receiverPK, reward); err != nil {
				return NewBlockChainError(ProcessSalaryInstructionsError, err)
			}
		}
	}
	return nil
}

func slashedShardCommitteeReward(
	shardID byte,
	rewardStateDB *statedb.StateDB,
	slashCommittees []string,
	rewReceiverM map[string]privacy.PaymentAddress,
) (
	err error,
) {
	for _, pk := range slashCommittees {
		if receiverAddr, ok := rewReceiverM[pk]; ok {
			receiverPK := base58.Base58Check{}.Encode(receiverAddr.Pk, common.Base58Version)
			if err = statedb.SlashShardCommitteeReward(rewardStateDB, receiverPK); err != nil {
				return NewBlockChainError(ProcessSalaryInstructionsError, err)
			}
			delete(rewReceiverM, pk)
		}
	}
	return nil
}

func addBeaconCommitteeRewardV2(
	rewardStateDB *statedb.StateDB,
	bCommitteeRewM map[string]map[common.Hash]uint64,
	shardID byte,
) (
	err error,
) {
	for bcCPK, reward := range bCommitteeRewM {
		bcCommitteePubkeyStruct := incognitokey.CommitteePublicKey{}
		if err := bcCommitteePubkeyStruct.FromString(bcCPK); err != nil {
			return err
		}
		if common.GetShardIDFromLastByte(bcCommitteePubkeyStruct.IncPubKey[len(bcCommitteePubkeyStruct.IncPubKey)-1]) == shardID {
			bIPKStr := bcCommitteePubkeyStruct.GetIncKeyBase58()
			if err = statedb.AddBeaconCommitteeReward(rewardStateDB, bIPKStr, reward); err != nil {
				return err
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) addShardCommitteeReward(
	rewardStateDB *statedb.StateDB,
	shardID byte,
	reward map[common.Hash]uint64,
	cStakeInfos []*statedb.ShardStakerInfo,
) (
	err error,
) {
	committeeSize := len(cStakeInfos)
	for _, candidate := range cStakeInfos {
		if common.GetShardIDFromLastByte(candidate.RewardReceiver().Pk[common.PublicKeySize-1]) == shardID {
			for key, value := range reward {
				tempPK := base58.Base58Check{}.Encode(candidate.RewardReceiver().Pk, common.Base58Version)
				Logger.log.Criticalf("Add Committee Reward ShardCommitteeReward, Public Key %+v, reward %+v, token %+v", tempPK, value/uint64(committeeSize), key)
				err = statedb.AddCommitteeReward(rewardStateDB, tempPK, value/uint64(committeeSize), key)
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) addShardCommitteeRewardSlashingVersion(
	rewardStateDB *statedb.StateDB,
	shardID byte,
	reward map[common.Hash]uint64,
	cStakeInfos []*statedb.StakerInfoSlashingVersion,
) (
	err error,
) {
	committeeSize := len(cStakeInfos)
	for _, candidate := range cStakeInfos {
		if common.GetShardIDFromLastByte(candidate.RewardReceiver().Pk[common.PublicKeySize-1]) == shardID {
			for key, value := range reward {
				tempPK := base58.Base58Check{}.Encode(candidate.RewardReceiver().Pk, common.Base58Version)
				Logger.log.Criticalf("Add Committee Reward ShardCommitteeReward, Public Key %+v, reward %+v, token %+v", tempPK, value/uint64(committeeSize), key)
				err = statedb.AddCommitteeReward(rewardStateDB, tempPK, value/uint64(committeeSize), key)
				if err != nil {
					return NewBlockChainError(ProcessSalaryInstructionsError, err)
				}
			}
		}
	}
	return nil
}

func (blockchain *BlockChain) calculateRewardMultiset(
	splitRewardRuleProcessor committeestate.SplitRewardRuleProcessor,
	curView *BeaconBestState,
	maxBeaconBlockCreation uint64,
	maxSubsetsCommittee int,
	beaconHeight uint64,
	epoch uint64,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
) (map[common.Hash]uint64,
	[][]map[common.Hash]uint64,
	map[common.Hash]uint64,
	map[common.Hash]uint64, error,
) {
	allCoinID := statedb.GetAllTokenIDForRewardMultiset(curView.rewardStateDB, epoch)
	currentBeaconYear, err := blockchain.GetYearOfBeacon(beaconHeight)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	percentForIncognitoDAO := getPercentForIncognitoDAOV2(currentBeaconYear)
	totalRewardForShardSubset := make([][]map[common.Hash]uint64, curView.ActiveShards)
	totalRewards := make([][]map[common.Hash]uint64, curView.ActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForIncDAO := map[common.Hash]uint64{}
	totalRewardForCustodian := map[common.Hash]uint64{}

	for shardID := 0; shardID < curView.ActiveShards; shardID++ {
		totalRewardForShardSubset[shardID] = make([]map[common.Hash]uint64, maxSubsetsCommittee)
		totalRewards[shardID] = make([]map[common.Hash]uint64, maxSubsetsCommittee)
		for subsetID := 0; subsetID < maxSubsetsCommittee; subsetID++ {
			if totalRewards[shardID][subsetID] == nil {
				totalRewards[shardID][subsetID] = map[common.Hash]uint64{}
			}
			if totalRewardForShardSubset[shardID][subsetID] == nil {
				totalRewardForShardSubset[shardID][subsetID] = map[common.Hash]uint64{}
			}

			for _, coinID := range allCoinID {
				totalRewards[shardID][subsetID][coinID], err = statedb.GetRewardOfShardByEpochMultiset(
					curView.rewardStateDB, epoch,
					byte(shardID), byte(subsetID), coinID)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				if totalRewards[shardID][subsetID][coinID] == 0 {
					delete(totalRewards[shardID][subsetID], coinID)
				}
			}

			env := committeestate.NewSplitRewardEnvironmentMultiset(
				byte(shardID),
				byte(subsetID),
				byte(maxSubsetsCommittee),
				beaconHeight,
				totalRewards[shardID][subsetID],
				isSplitRewardForCustodian,
				percentCustodianRewards,
				percentForIncognitoDAO,
				curView.GetBeaconCommittee(),
				curView.GetShardCommittee(),
				curView.GetCommitteeState(),
			)

			Logger.log.Info("[dcs] env.MaxSubsetCommittees:", env.MaxSubsetCommittees)
			rewardForBeacon, rewardForShardSubset, rewardForDAO, rewardForCustodian, err := splitRewardRuleProcessor.SplitReward(env)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			plusMap(rewardForBeacon, totalRewardForBeacon)
			plusMap(rewardForShardSubset, totalRewardForShardSubset[shardID][subsetID])
			plusMap(rewardForDAO, totalRewardForIncDAO)
			plusMap(rewardForCustodian, totalRewardForCustodian)
		}
	}

	return totalRewardForBeacon, totalRewardForShardSubset, totalRewardForIncDAO, totalRewardForCustodian, nil
}

func (blockchain *BlockChain) calculateReward(
	splitRewardRuleProcessor committeestate.SplitRewardRuleProcessor,
	curView *BeaconBestState,
	numberOfActiveShards int,
	beaconHeight uint64,
	epoch uint64,
	rewardStateDB *statedb.StateDB,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
) (map[common.Hash]uint64,
	[]map[common.Hash]uint64,
	map[common.Hash]uint64,
	map[common.Hash]uint64, error,
) {
	allCoinID := statedb.GetAllTokenIDForReward(rewardStateDB, epoch)
	currentBeaconYear, err := blockchain.GetYearOfBeacon(beaconHeight)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	percentForIncognitoDAO := getPercentForIncognitoDAOV2(currentBeaconYear)
	totalRewardForShard := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewards := make([]map[common.Hash]uint64, numberOfActiveShards)
	totalRewardForBeacon := map[common.Hash]uint64{}
	totalRewardForIncDAO := map[common.Hash]uint64{}
	totalRewardForCustodian := map[common.Hash]uint64{}

	for id := 0; id < numberOfActiveShards; id++ {
		if totalRewards[id] == nil {
			totalRewards[id] = map[common.Hash]uint64{}
		}
		if totalRewardForShard[id] == nil {
			totalRewardForShard[id] = map[common.Hash]uint64{}
		}

		for _, coinID := range allCoinID {
			totalRewards[id][coinID], err = statedb.GetRewardOfShardByEpoch(rewardStateDB, epoch, byte(id), coinID)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if totalRewards[id][coinID] == 0 {
				delete(totalRewards[id], coinID)
			}
		}

		env := committeestate.NewSplitRewardEnvironmentV1(
			byte(id),
			beaconHeight,
			totalRewards[id],
			isSplitRewardForCustodian,
			percentCustodianRewards,
			percentForIncognitoDAO,
			numberOfActiveShards,
			curView.GetBeaconCommittee(),
			curView.GetShardCommittee(),
			curView.GetCommitteeState(),
		)
		rewardForBeacon, rewardForShard, rewardForDAO, rewardForCustodian, err := splitRewardRuleProcessor.SplitReward(env)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		plusMap(rewardForBeacon, totalRewardForBeacon)
		plusMap(rewardForShard, totalRewardForShard[id])
		plusMap(rewardForDAO, totalRewardForIncDAO)
		plusMap(rewardForCustodian, totalRewardForCustodian)
	}

	return totalRewardForBeacon, totalRewardForShard, totalRewardForIncDAO, totalRewardForCustodian, nil
}

func (blockchain *BlockChain) buildRewardInstructionByEpoch(
	curView *BeaconBestState,
	blkHeight, epoch uint64,
	isSplitRewardForCustodian bool,
	percentCustodianRewards uint64,
	isSplitRewardForPdex bool,
	pdexRewardPercent uint,
	blockVersion int,
	committeeChange *committeestate.CommitteeChange,
) ([][]string, map[common.Hash]uint64, uint64, error) {

	//Declare variables
	var resInst [][]string
	var err error
	var instRewardForBeacons [][]string
	var instRewardForIncDAO [][]string
	var instRewardForShards [][]string

	beaconBestView := blockchain.BeaconChain.GetBestView().(*BeaconBestState)

	totalRewardForBeacon := make(map[common.Hash]uint64)
	totalRewardForShard := make([]map[common.Hash]uint64, beaconBestView.ActiveShards)
	totalRewardForShardSubset := make([][]map[common.Hash]uint64, beaconBestView.ActiveShards)
	totalRewardForCustodian := make(map[common.Hash]uint64)
	totalRewardForIncDAO := make(map[common.Hash]uint64)
	rewardForPdex := uint64(0)

	if blockVersion >= types.BLOCK_PRODUCINGV3_VERSION && blockVersion < types.INSTANT_FINALITY_VERSION_V2 {
		splitRewardRuleProcessor := committeestate.GetRewardSplitRule(blockVersion)
		totalRewardForBeacon,
			totalRewardForShardSubset,
			totalRewardForIncDAO,
			totalRewardForCustodian,
			err = blockchain.calculateRewardMultiset(
			splitRewardRuleProcessor,
			curView,
			uint64(config.Param().BlockTime.MaxBeaconBlockCreation.Seconds()),
			MaxSubsetCommittees,
			blkHeight,
			epoch,
			isSplitRewardForCustodian,
			percentCustodianRewards,
		)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
		instRewardForShards, err = blockchain.buildInstructionRewardForShardsV3(epoch, totalRewardForShardSubset)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}

	} else {
		splitRewardRuleProcessor := committeestate.GetRewardSplitRule(blockVersion)
		totalRewardForBeacon,
			totalRewardForShard,
			totalRewardForIncDAO,
			totalRewardForCustodian,
			err = blockchain.calculateReward(
			splitRewardRuleProcessor,
			curView,
			curView.ActiveShards, blkHeight, epoch,
			curView.GetBeaconRewardStateDB(),
			isSplitRewardForCustodian, percentCustodianRewards,
		)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
		instRewardForShards, err = blockchain.buildInstructionRewardForShards(epoch, totalRewardForShard)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
	}

	if len(totalRewardForBeacon) > 0 {
		if curView.GetBeaconHeight() >= config.Param().ConsensusParam.StakingFlowV4Height {
			instRewardForBeacons, err = curView.buildInstRewardForBeaconDelegation(epoch, totalRewardForBeacon, committeeChange)
			if err != nil {
				return nil, nil, rewardForPdex, err
			}
		} else {
			instRewardForBeacons, err = curView.buildInstRewardForBeacons(epoch, totalRewardForBeacon)
			if err != nil {
				return nil, nil, rewardForPdex, err
			}
		}

	}

	if len(totalRewardForIncDAO) > 0 {
		prvReward := totalRewardForIncDAO[common.PRVCoinID]
		if prvReward > 0 && isSplitRewardForPdex {
			temp := new(big.Int).Mul(new(big.Int).SetUint64(prvReward), new(big.Int).SetUint64(uint64(pdexRewardPercent)))
			temp = temp.Div(temp, big.NewInt(100))
			if !temp.IsUint64() {
				return nil, nil, rewardForPdex, errors.New("Reward for Pdex is not uint64")
			}
			rewardForPdex = temp.Uint64()
			totalRewardForIncDAO[common.PRVCoinID] = prvReward - rewardForPdex
		}
		instRewardForIncDAO, err = blockchain.buildInstRewardForIncDAO(epoch, totalRewardForIncDAO)
		if err != nil {
			return nil, nil, rewardForPdex, err
		}
	}

	resInst = common.AppendSliceString(instRewardForBeacons, instRewardForIncDAO, instRewardForShards)
	return resInst, totalRewardForCustodian, rewardForPdex, nil
}

// buildInstRewardForBeacons create reward instruction for beacons
func (beaconBestState *BeaconBestState) buildInstRewardForBeacons(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	baseRewards := map[common.Hash]uint64{}
	for key, value := range totalReward {
		baseRewards[key] = value / uint64(len(beaconBestState.GetBeaconCommittee()))
	}
	for _, beaconpublickey := range beaconBestState.GetBeaconCommittee() {
		// indicate reward pubkey
		singleInst, err := metadata.BuildInstForBeaconReward(baseRewards, beaconpublickey.GetNormalKey())
		if err != nil {
			Logger.log.Errorf("BuildInstForBeaconReward error %+v\n Totalreward: %+v, epoch: %+v, reward: %+v\n", err, totalReward, epoch, baseRewards)
			return nil, err
		}
		resInst = append(resInst, singleInst)
	}
	return resInst, nil
}

func (beaconBestState *BeaconBestState) buildInstRewardForBeaconDelegation(epoch uint64, totalReward map[common.Hash]uint64, committeeChange *committeestate.CommitteeChange) ([][]string, error) {
	bCStateI := beaconBestState.GetCommitteeState()
	bCStateV4, ok := bCStateI.(*committeestate.BeaconCommitteeStateV4)
	if !ok {
		return nil, errors.Errorf("Instruction for beacon delegation just work with beacon committee state version > 3")
	}
	validators := bCStateV4.GetBeaconCommitteeString()
	dState := bCStateV4.GetDelegateState()
	rState := bCStateV4.Reputation
	slashed := map[string]interface{}{}
	for _, slashcommittees := range committeeChange.SlashingCommittee {
		for _, pkSlashed := range slashcommittees {
			slashed[pkSlashed] = nil
		}
	}
	for pk := range slashed {
		for bcPK, dInfo := range dState {
			if _, ok := dInfo.CurrentDelegatorsDetails[pk]; ok {
				delete(dInfo.CurrentDelegatorsDetails, pk)
				dInfo.CurrentDelegators--
			}
			dState[bcPK] = dInfo
		}
	}

	detailBCReward, detailDelegatorsReward := calculateRewardForBeacon(validators, dState, rState, totalReward)

	resInst := [][]string{}
	keys := []string{}
	for k := range detailBCReward {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, beaconpublickey := range keys {
		rew := detailBCReward[beaconpublickey]
		delegatorReward, ok := detailDelegatorsReward[beaconpublickey]
		if !ok {
			delegatorReward = map[common.Hash]uint64{}
		}
		singleInst, err := metadata.BuildInstForBeaconRewardV2(rew, delegatorReward, beaconpublickey, epoch)
		if err != nil {
			Logger.log.Errorf("BuildInstForBeaconReward error %+v\n Totalreward: %+v, epoch: %+v, reward: %+v\n", err, totalReward, epoch, rew)
			return nil, err
		}
		resInst = append(resInst, singleInst)
	}
	return resInst, nil
}

func (blockchain *BlockChain) buildInstRewardForIncDAO(epoch uint64, totalReward map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	devRewardInst, err := metadata.BuildInstForIncDAOReward(totalReward, config.Param().IncognitoDAOAddress)
	if err != nil {
		Logger.log.Errorf("buildInstRewardForIncDAO error %+v\n Totalreward: %+v, epoch: %+v\n", err, totalReward, epoch)
		return nil, err
	}
	resInst = append(resInst, devRewardInst)
	return resInst, nil
}

func (blockchain *BlockChain) buildInstructionRewardForShardsV3(epoch uint64, totalRewards [][]map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}

	for shardID, v := range totalRewards {
		for subsetID, reward := range v {
			if len(reward) > 0 {
				shardSubsetReward := instruction.NewShardReceiveRewardV3WithValue(reward, epoch, byte(shardID), byte(subsetID))
				shardSubsetRewardInst := shardSubsetReward.String()
				resInst = append(resInst, shardSubsetRewardInst)
			}
		}
	}

	return resInst, nil
}

func (blockchain *BlockChain) buildInstructionRewardForShards(epoch uint64, totalRewards []map[common.Hash]uint64) ([][]string, error) {
	resInst := [][]string{}
	for i, reward := range totalRewards {
		if len(reward) > 0 {
			shardRewardInst, err := instruction.NewShardReceiveRewardV1WithValue(reward, epoch, byte(i))
			if err != nil {
				Logger.log.Errorf("BuildInstForShardReward error %+v\n Totalreward: %+v, epoch: %+v\n; shard:%+v", err, reward, epoch, byte(i))
				return nil, err
			}
			resInst = append(resInst, shardRewardInst...)
		}
	}
	return resInst, nil
}

func (blockchain *BlockChain) buildWithDrawTransactionResponse(view *ShardBestState, txRequest *metadata.Transaction, blkProducerPrivateKey *privacy.PrivateKey, shardID byte) (metadata.Transaction, error) {
	if (*txRequest).GetMetadataType() != metadata.WithDrawRewardRequestMeta {
		return nil, errors.New("Can not understand this request!")
	}
	requestDetail := (*txRequest).GetMetadata().(*metadata.WithDrawRewardRequest)
	tempPublicKey := base58.Base58Check{}.Encode(requestDetail.PaymentAddress.Pk, common.Base58Version)
	amount, err := statedb.GetCommitteeReward(view.GetShardRewardStateDB(), tempPublicKey, requestDetail.TokenID)
	if (amount == 0) || (err != nil) {
		return nil, errors.New("Not enough reward")
	}
	responseMeta, err := metadata.NewWithDrawRewardResponse(requestDetail, (*txRequest).Hash())
	if err != nil {
		return nil, err
	}
	txParam := transaction.TxSalaryOutputParams{Amount: amount, ReceiverAddress: &requestDetail.PaymentAddress, TokenID: &requestDetail.TokenID}
	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			responseMeta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return responseMeta
	}

	salaryTx, err := txParam.BuildTxSalary(blkProducerPrivateKey, view.GetCopiedTransactionStateDB(), makeMD)
	if err != nil {
		return nil, errors.Errorf("cannot init salary tx. Error %v", err)
	}
	salaryTx.SetType(common.TxRewardType)
	return salaryTx, nil
}

func (blockchain *BlockChain) GetBlockTimeByBlockVersion(blkVersion int) (int64, error) {
	blockTimeMap := config.Param().BlockTimeParam
	defaultBlockTime := blockTimeMap[BLOCKTIME_DEFAULT]
	blockVersionEnableFeature := map[string]int{}
	sortBlockVersionEnableFeature := []string{}

	//from TriggeredFeature in beacon beststate, we extract feature that related to adjusting block time
	for feature, enableBlockHeight := range blockchain.BeaconChain.GetBestView().(*BeaconBestState).TriggeredFeature {
		if _, ok := config.Param().BlockTimeParam[feature]; ok {
			beaconBlock, err := blockchain.GetBeaconBlockByHeightV1(enableBlockHeight)
			if err != nil {
				return 0, fmt.Errorf("Cannot find beacon block %v", enableBlockHeight)
			}
			blockVersionEnableFeature[feature] = beaconBlock.GetVersion() //this is version before adjusting block time
			sortBlockVersionEnableFeature = append(sortBlockVersionEnableFeature, feature)
		}
	}

	//then sort by version
	sort.Slice(sortBlockVersionEnableFeature, func(i, j int) bool {
		if blockVersionEnableFeature[sortBlockVersionEnableFeature[i]] > blockVersionEnableFeature[sortBlockVersionEnableFeature[j]] {
			return true
		}
		return false
	})

	//return the corresponding blocktime of blkVersion
	for _, feature := range sortBlockVersionEnableFeature {
		if blkVersion > blockVersionEnableFeature[feature] { //version must be greater than the version before adjusting
			if t, ok := blockTimeMap[feature]; !ok {
				return 0, fmt.Errorf("Cannot find block time param %v", feature)
			} else {
				return t, nil
			}
		}
	}

	return defaultBlockTime, nil
}

func (blockchain *BlockChain) GetBasicRewardByVersion(version int) (uint64, error) {
	blockTimeMap := config.Param().BlockTimeParam
	defaultBlockTime := blockTimeMap[BLOCKTIME_DEFAULT]
	curBlockTime, err := blockchain.GetBlockTimeByBlockVersion(version)
	if err != nil {
		return 0, err
	}
	return config.Param().BasicReward * uint64(curBlockTime) / uint64(defaultBlockTime), nil
}

func (blockchain *BlockChain) getRewardAmount(blkHeight uint64) uint64 {
	blockBeaconInterval := config.Param().BlockTime.MinBeaconBlockInterval.Seconds()
	blockInYear := getNoBlkPerYear(uint64(blockBeaconInterval))
	n := (blkHeight - 1) / blockInYear
	reward := uint64(config.Param().BasicReward)
	for ; n > 0; n-- {
		reward *= 91
		reward /= 100
	}
	return reward
}

func (blockchain *BlockChain) GetRewardAmount(shardID byte, shardVersion int, shardHeight uint64) (uint64, error) {
	if shardVersion < types.ADJUST_BLOCKTIME_VERSION {
		return blockchain.getRewardAmount(shardHeight), nil
	}
	basicReward, err := blockchain.GetBasicRewardByVersion(shardVersion)
	if err != nil {
		return 0, err
	}
	yearOfShardHeight, err := blockchain.GetYearOfShard(shardID, shardHeight)
	if err != nil {
		return 0, err
	}
	return blockchain.getRewardAmountV2(basicReward, yearOfShardHeight), nil

}

func (blockchain *BlockChain) getRewardAmountV2(basicReward, year uint64) uint64 {
	n := year
	reward := basicReward
	for ; n > 0; n-- {
		reward *= 91
		reward /= 100
	}
	return reward
}

func (blockchain *BlockChain) GetYearOfBeacon(blockHeight uint64) (uint64, error) {

	bView := blockchain.GetBeaconBestState()
	if bView == nil {
		return 0, errors.Errorf("Can not get beacon view for get reward amount at block beacon %v", blockHeight)
	}
	featureManager := bView.TSManager
	return getYearOfBlockChain(&featureManager, blockHeight), nil
}

func (blockchain *BlockChain) GetYearOfShard(sID byte, blockHeight uint64) (uint64, error) {

	bView := blockchain.GetBeaconBestState()
	if bView == nil {
		return 0, errors.Errorf("Can not get beacon view for get reward amount at block beacon %v", blockHeight)
	}
	if _, ok := bView.ShardTSManager[sID]; !ok {
		bView.ShardTSManager[sID] = new(TSManager)
	}
	featureManager := bView.ShardTSManager[sID]
	return getYearOfBlockChain(featureManager, blockHeight), nil
}

func getYearOfBlockChain(featuresManager *TSManager, blkHeight uint64) uint64 {
	anchors := featuresManager.Anchors
	startBlkHeight := uint64(0)
	endBlkHeight := uint64(0)
	prevBlockTime := config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]
	defBlockTime := config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]
	totalBlock := uint64(0)
	if len(anchors) > 0 {
		for _, anchor := range anchors {
			endBlkHeight = anchor.BlockHeight
			if endBlkHeight > blkHeight {
				endBlkHeight = blkHeight
			}
			// fmt.Printf("%v - %v - %v - %v\n", prevBlockTime, startBlkHeight, endBlkHeight, (endBlkHeight-startBlkHeight)*uint64(prevBlockTime)/uint64(defBlockTime))
			totalBlock += (endBlkHeight - startBlkHeight) * uint64(prevBlockTime) / uint64(defBlockTime)
			startBlkHeight = endBlkHeight
			prevBlockTime = config.Param().BlockTimeParam[anchor.Feature]
		}
	}
	if endBlkHeight < blkHeight {
		endBlkHeight = blkHeight
		// fmt.Printf("%v - %v - %v - %v\n", prevBlockTime, startBlkHeight, endBlkHeight, (endBlkHeight-startBlkHeight)*uint64(prevBlockTime)/uint64(defBlockTime))
		totalBlock += (endBlkHeight - startBlkHeight) * uint64(prevBlockTime) / uint64(defBlockTime)
	}
	// fmt.Printf("%v - %v\n", totalBlock, getNoBlkPerYear(uint64(defBlockTime)))
	return totalBlock / getNoBlkPerYear(uint64(defBlockTime))
}

func getNoBlkPerYear(blockCreationTimeSeconds uint64) uint64 {
	return (365.25 * 24 * 60 * 60) / blockCreationTimeSeconds
}

func GetBlockTimeInterval(blkTimeFeature string) int64 {
	blockTimeMap := config.Param().BlockTimeParam
	if blkTime, ok := blockTimeMap[blkTimeFeature]; ok {
		return int64(blkTime)
	}
	return int64(config.Param().BlockTime.MinBeaconBlockInterval.Seconds())

}

func GetNumberBlkPerYear(blkTimeFeature string) uint64 {
	return getNoBlkPerYear(uint64(GetBlockTimeInterval(blkTimeFeature)))
}

func getPercentForIncognitoDAO(blockHeight, blkPerYear uint64) int {
	year := (blockHeight - 1) / blkPerYear
	if year > (UpperBoundPercentForIncDAO - LowerBoundPercentForIncDAO) {
		return LowerBoundPercentForIncDAO
	} else {
		return UpperBoundPercentForIncDAO - int(year)
	}
}

func getPercentForIncognitoDAOV2(year uint64) int {
	if year > (UpperBoundPercentForIncDAO - LowerBoundPercentForIncDAO) {
		return LowerBoundPercentForIncDAO
	} else {
		return UpperBoundPercentForIncDAO - int(year)
	}
}

// plusMap(src, dst): dst = dst + src
func plusMap(src, dst map[common.Hash]uint64) {
	if src != nil {
		for key, value := range src {
			dst[key] += value
		}
	}
}

func calculateRewardForBeacon(
	bCommittee []string,
	dState map[string]committeestate.BeaconDelegatorInfo,
	rState map[string]uint64,
	totalRewardM map[common.Hash]uint64,
) (
	forBeacon map[string]map[common.Hash]uint64,
	forDelegators map[string]map[common.Hash]uint64,
) {
	forBeacon = map[string]map[common.Hash]uint64{}
	forDelegators = map[string]map[common.Hash]uint64{}
	totalCredits := 0
	for _, val := range bCommittee {
		totalCredits += 2
		if dInfo, ok := dState[val]; ok {
			totalCredits += dInfo.CurrentDelegators
		}
	}
	for _, val := range bCommittee {
		//TODO modify it
		bcStakingCredit := 2
		rewM := map[common.Hash]uint64{}
		dInfo, hasDInfo := dState[val]
		rInfo, hasRInfo := rState[val]
		rewBeacon := map[common.Hash]uint64{}
		rewDelegators := map[common.Hash]uint64{}
		for tokenID, totalReward := range totalRewardM {
			cre := bcStakingCredit
			if hasDInfo {
				cre += dInfo.CurrentDelegators
			}
			rew := totalReward * uint64(cre) / uint64(totalCredits)
			if hasRInfo {
				rewM[tokenID] = rew * rInfo / 1000
			} else {
				rewM[tokenID] = rew
			}
			if rewM[tokenID] > 0 {
				rewBeacon[tokenID] = rewM[tokenID] * uint64(bcStakingCredit)
				if hasDInfo {
					rewDelegators[tokenID] = rewM[tokenID]
				}
			}
		}
		forBeacon[val] = rewBeacon
		forDelegators[val] = rewDelegators
	}
	return forBeacon, forDelegators
}
