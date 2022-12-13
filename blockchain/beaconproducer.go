package blockchain

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/portal"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	"github.com/incognitochain/incognito-chain/syncker/finishsync"
)

type duplicateKeyStakeInstruction struct {
	instructions []*instruction.StakeInstruction
}

func (inst *duplicateKeyStakeInstruction) add(newInst *duplicateKeyStakeInstruction) {
	inst.instructions = append(inst.instructions, newInst.instructions...)
}

type shardInstruction struct {
	stakeInstructions         []*instruction.StakeInstruction
	unstakeInstructions       []*instruction.UnstakeInstruction
	swapInstructions          map[byte][]*instruction.SwapInstruction
	stopAutoStakeInstructions []*instruction.StopAutoStakeInstruction
	redelegateInstructions    []*instruction.ReDelegateInstruction
	addStakingInstructions    []*instruction.AddStakingInstruction
}

func newShardInstruction() *shardInstruction {
	return &shardInstruction{
		swapInstructions: make(map[byte][]*instruction.SwapInstruction),
	}
}

func (shardInstruction *shardInstruction) add(newShardInstruction *shardInstruction) {
	shardInstruction.stakeInstructions = append(shardInstruction.stakeInstructions, newShardInstruction.stakeInstructions...)
	shardInstruction.unstakeInstructions = append(shardInstruction.unstakeInstructions, newShardInstruction.unstakeInstructions...)
	shardInstruction.stopAutoStakeInstructions = append(shardInstruction.stopAutoStakeInstructions, newShardInstruction.stopAutoStakeInstructions...)
	shardInstruction.redelegateInstructions = append(shardInstruction.redelegateInstructions, newShardInstruction.redelegateInstructions...)
	shardInstruction.addStakingInstructions = append(shardInstruction.addStakingInstructions, newShardInstruction.addStakingInstructions...)
	for shardID, swapInstructions := range newShardInstruction.swapInstructions {
		shardInstruction.swapInstructions[shardID] = append(shardInstruction.swapInstructions[shardID], swapInstructions...)
	}
}

// NewBlockBeacon create new beacon block
func (blockchain *BlockChain) NewBlockBeacon(
	curView *BeaconBestState,
	version int, proposer string, round int, startTime int64,
	prevValidationData string,
) (*types.BeaconBlock, error) {
	Logger.log.Infof("⛏ Creating Beacon Block %+v", curView.BeaconHeight+1)
	var err error
	var epoch uint64
	//var isNextEpoch bool = false
	newBeaconBlock := types.NewBeaconBlock()
	copiedCurView := NewBeaconBestState()

	err = copiedCurView.cloneBeaconBestStateFrom(curView)
	if err != nil {
		return nil, err
	}

	epoch, _ = blockchain.GetEpochNextHeight(copiedCurView.BeaconHeight)
	Logger.log.Infof("New Beacon Block, height %+v, epoch %+v", copiedCurView.BeaconHeight+1, epoch)
	newBeaconBlock.Header = types.NewBeaconHeader(
		version,
		copiedCurView.BeaconHeight+1,
		epoch,
		round,
		startTime,
		copiedCurView.BestBlockHash,
		copiedCurView.ConsensusAlgorithm,
		proposer,
		proposer,
		prevValidationData,
	)

	if version >= types.INSTANT_FINALITY_VERSION {
		if blockchain.shouldBeaconGenerateBridgeInstruction(copiedCurView) {
			processBridgeFromBlock := copiedCurView.LastBlockProcessBridge + 1
			newBeaconBlock.Header.ProcessBridgeFromBlock = &processBridgeFromBlock
		}
	}

	BLogger.log.Infof("Producing block: %d (epoch %d)", newBeaconBlock.Header.Height, newBeaconBlock.Header.Epoch)
	//=====END Build Header Essential Data=====
	portalParams := portal.GetPortalParams()
	allShardBlocks := blockchain.GetShardBlockForBeaconProducer(copiedCurView.BestShardHeight)

	//dequeueInst := copiedCurView.generateOutdatedDequeueInstruction()

	instructions, shardStates, err := blockchain.GenerateBeaconBlockBody(
		newBeaconBlock,
		copiedCurView,
		*portalParams,
		allShardBlocks,
	)
	if err != nil {
		return nil, NewBlockChainError(GenerateInstructionError, err)
	}

	finishSyncInstructions := copiedCurView.generateFinishSyncInstruction()
	instructions = addFinishInstruction(instructions, finishSyncInstructions)

	enableFeatureInstructions, _ := copiedCurView.generateEnableFeatureInstructions()
	instructions = append(instructions, enableFeatureInstructions...)

	newBeaconBlock.Body = types.NewBeaconBody(shardStates, instructions)

	// Process new block with new view
	_, hashes, _, incurredInstructions, err := copiedCurView.updateBeaconBestState(newBeaconBlock, blockchain)
	if err != nil {
		return nil, err
	}

	instructions = append(instructions, incurredInstructions...)
	newBeaconBlock.Body.SetInstructions(instructions)
	if len(newBeaconBlock.Body.Instructions) != 0 {
		Logger.log.Info("Beacon Produce: Beacon Instruction", newBeaconBlock.Body.Instructions)
	}

	// calculate hash
	tempInstructionArr := []string{}
	for _, strs := range instructions {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	instructionHash, err := generateHashFromStringArray(tempInstructionArr)
	if err != nil {
		return nil, NewBlockChainError(GenerateInstructionHashError, err)
	}
	shardStatesHash, err := generateHashFromShardState(shardStates, curView.CommitteeStateVersion())
	if err != nil {
		return nil, NewBlockChainError(GenerateShardStateError, err)
	}
	// Instruction merkle root
	flattenInsts, err := FlattenAndConvertStringInst(instructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, err)
	}
	// add hash to header
	newBeaconBlock.Header.AddBeaconHeaderHash(
		instructionHash,
		shardStatesHash,
		types.GetKeccak256MerkleRoot(flattenInsts),
		hashes.BeaconCommitteeAndValidatorHash,
		hashes.BeaconCandidateHash,
		hashes.ShardCandidateHash,
		hashes.ShardCommitteeAndValidatorHash,
		hashes.AutoStakeHash,
		hashes.ShardSyncValidatorsHash,
	)
	return newBeaconBlock, nil
}

// beacon should only generate bridge (unshield) instruction when curView is finality (generate instructions from checkpoint to curView)
func (blockchain *BlockChain) shouldBeaconGenerateBridgeInstruction(curView *BeaconBestState) bool {
	if curView.GetBlock().Hash().IsEqual(blockchain.BeaconChain.GetFinalView().GetBlock().Hash()) {
		return true
	}
	return false
}

func (blockchain *BlockChain) generateBridgeInstruction(
	curView *BeaconBestState,
	currentShardStateBlock map[byte][]*types.ShardBlock,
	newBeaconBlock *types.BeaconBlock,
) ([][]string, error) {
	keys := []int{}
	bridgeInstructions := [][]string{}
	for shardID := range currentShardStateBlock {
		keys = append(keys, int(shardID))
	}
	sort.Ints(keys)

	for _, v := range keys {
		shardID := byte(v)
		for _, shardBlock := range currentShardStateBlock[shardID] {
			actions, err := CreateShardBridgeUnshieldActionsFromTxs(shardBlock.Body.Transactions, blockchain,
				shardID, shardBlock.Header.Height, shardBlock.Header.BeaconHeight)
			if err != nil {
				BLogger.log.Errorf("Build bridge unshield instructions failed: %s", err.Error())
				return nil, err
			}

			// build bridge unshield instructions
			bridgeInstructionForBlock, err := blockchain.buildBridgeInstructions(
				curView.GetBeaconFeatureStateDB(),
				shardID,
				actions,
				newBeaconBlock.GetHeight(),
			)
			if err != nil {
				BLogger.log.Errorf("Build bridge unshield confirm instructions failed: %s", err.Error())
				return nil, err
			}
			bridgeInstructions = append(bridgeInstructions, bridgeInstructionForBlock...)
		}
	}

	return bridgeInstructions, nil
}

// generateBridgeAggInstruction creates bridge agg unshield instructions for FINALIZED unshield reqs
func (blockchain *BlockChain) generateBridgeAggInstruction(
	curView *BeaconBestState,
	currentShardStateBlockForBridgeAgg map[uint64]map[byte][]*types.ShardBlock,
	newBeaconBlock *types.BeaconBlock,
) ([][]string, error) {
	bridgeAggInstructions := [][]string{}
	unshieldActions := []bridgeagg.UnshieldActionForProducer{}

	for beaconHeight, shardBlkMaps := range currentShardStateBlockForBridgeAgg {
		for shardID, shardBlks := range shardBlkMaps {
			for _, shardBlk := range shardBlks {
				actions, err := CreateShardBridgeAggUnshieldActionsFromTxs(
					shardBlk.Body.Transactions, blockchain,
					shardID, shardBlk.Header.Height, shardBlk.Header.BeaconHeight)
				if err != nil {
					BLogger.log.Errorf("Build bridge agg unshield actions failed: %s", err.Error())
					return nil, err
				}
				unshieldActions = append(unshieldActions,
					bridgeagg.BuildUnshieldActionForProducerFromInsts(actions, shardID, beaconHeight)...)
			}
		}
	}

	// sort unshieldActions by beaconHeight ascending and TxID ascending
	sort.SliceStable(unshieldActions, func(i, j int) bool {
		if unshieldActions[i].BeaconHeight == unshieldActions[j].BeaconHeight {
			return unshieldActions[i].TxReqID.String() < unshieldActions[j].TxReqID.String()
		}
		return unshieldActions[i].BeaconHeight < unshieldActions[j].BeaconHeight
	})

	// build bridge aggregator unshield instructions
	newInsts, err := curView.bridgeAggManager.BuildNewUnshieldInstructions(curView.GetBeaconFeatureStateDB(), newBeaconBlock.GetHeight(), unshieldActions)
	if err != nil {
		BLogger.log.Errorf("Build bridge agg unshield instructions failed: %s", err.Error())
		return nil, err
	}
	bridgeAggInstructions = append(bridgeAggInstructions, newInsts...)

	return bridgeAggInstructions, nil
}

// GenerateBeaconBlockBody generate beacon instructions and shard states
func (blockchain *BlockChain) GenerateBeaconBlockBody(
	newBeaconBlock *types.BeaconBlock,
	curView *BeaconBestState,
	portalParams portal.PortalParams,
	allShardBlocks map[byte][]*types.ShardBlock,
) ([][]string, map[byte][]types.ShardState, error) {
	bridgeInstructions := [][]string{}
	bridgeAggInstructions := [][]string{}
	acceptedRewardInstructions := [][]string{}
	statefulActionsByShardID := map[byte][][]string{}
	shardStates := make(map[byte][]types.ShardState)
	shardInstruction := newShardInstruction()
	duplicateKeyStakeInstructions := &duplicateKeyStakeInstruction{}
	validStakePublicKeys := []string{}
	validUnstakePublicKeys := make(map[string]bool)
	rewardForCustodianByEpoch := map[common.Hash]uint64{}
	rewardByEpochInstruction := [][]string{}
	pdexReward := uint64(0)
	committeeChange := committeestate.NewCommitteeChange()

	keys := []int{}
	for shardID, shardBlocks := range allShardBlocks {
		strs := fmt.Sprintf("GetShardState shardID: %+v, Height", shardID)
		for _, shardBlock := range shardBlocks {
			strs += fmt.Sprintf(" %d", shardBlock.Header.Height)
		}
		Logger.log.Info(strs)
		keys = append(keys, int(shardID))
	}
	sort.Ints(keys)
	//Shard block is a map ShardId -> array of shard block

	allPdexTxs := make(map[uint]map[byte][]metadata.Transaction)

	for _, v := range keys {
		shardID := byte(v)
		shardBlocks := allShardBlocks[shardID]
		for _, shardBlock := range shardBlocks {
			shardState, newShardInstruction, newDuplicateKeyStakeInstruction,
				acceptedRewardInstruction, statefulActions,
				pdexTxs, err := blockchain.GetShardStateFromBlock(
				curView, curView.BeaconHeight+1, shardBlock, shardID, validUnstakePublicKeys, validStakePublicKeys)
			if err != nil {
				return [][]string{}, shardStates, err
			}
			shardStates[shardID] = append(shardStates[shardID], shardState[shardID])
			duplicateKeyStakeInstructions.add(newDuplicateKeyStakeInstruction)
			shardInstruction.add(newShardInstruction)
			acceptedRewardInstructions = append(acceptedRewardInstructions, acceptedRewardInstruction)
			// group stateful actions by shardID
			for _, v := range newShardInstruction.stakeInstructions {
				validStakePublicKeys = append(validStakePublicKeys, v.PublicKeys...)
			}
			_, found := statefulActionsByShardID[shardID]
			if !found {
				statefulActionsByShardID[shardID] = statefulActions
			} else {
				statefulActionsByShardID[shardID] = append(statefulActionsByShardID[shardID], statefulActions...)
			}
			for version, txs := range pdexTxs {
				if allPdexTxs[version] == nil {
					allPdexTxs[version] = make(map[byte][]metadata.Transaction)
				}
				allPdexTxs[version][shardID] = append(allPdexTxs[version][shardID], txs...)
			}
			for _, ins := range shardInstruction.redelegateInstructions {
				Logger.log.Infof("Shard inst: %+v %+v", ins.CommitteePublicKeys, ins.DelegateList)
			}
		}
	}

	// remove duplicate PreValidation data in ShardState
	for _, ss := range shardStates {
		for i := len(ss) - 1; i > 0; i-- {
			if i == 0 {
				break
			}
			ss[i].PreviousValidationData = ""
		}
	}

	// build stateful instructions
	statefulInsts, err := blockchain.buildStatefulInstructions(
		curView,
		curView.featureStateDB,
		statefulActionsByShardID,
		newBeaconBlock.Header.Height,
		rewardForCustodianByEpoch,
		portalParams,
		shardStates,
		allPdexTxs,
		pdexReward,
	)
	if err != nil {
		return nil, nil, err
	}

	// build bridge unshielding instruction
	retrievedShardBlockForBridge := allShardBlocks
	retrievedShardBlockForBridgeAgg := map[uint64]map[byte][]*types.ShardBlock{
		newBeaconBlock.Header.Height: allShardBlocks,
	}

	if newBeaconBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION {
		if blockchain.shouldBeaconGenerateBridgeInstruction(curView) {
			//get data from checkpoint to final view
			Logger.log.Infof("[Bridge Debug] Checking bridge for beacon block %v %v", curView.LastBlockProcessBridge+1, blockchain.BeaconChain.GetFinalView().GetHeight())
			retrievedShardBlockForBridge, retrievedShardBlockForBridgeAgg, err = blockchain.GetShardBlockForBridge(curView.LastBlockProcessBridge+1, *blockchain.BeaconChain.GetFinalView().GetHash(), newBeaconBlock, shardStates)
			if err != nil {
				return nil, nil, NewBlockChainError(BuildBridgeError, err)
			}
		}
	}
	Logger.log.Infof("[Bridge Debug] retrievedShardBlockForBridge %+v", retrievedShardBlockForBridge)
	Logger.log.Infof("[Bridge Debug] retrievedShardBlockForBridgeAgg %+v", retrievedShardBlockForBridgeAgg)

	bridgeInstructions, err = blockchain.generateBridgeInstruction(curView, retrievedShardBlockForBridge, newBeaconBlock)
	Logger.log.Info("[Bridge Debug] Generate bridge unshield instruction", len(bridgeInstructions))
	if err != nil {
		return nil, nil, NewBlockChainError(BuildBridgeError, err)
	}
	bridgeAggInstructions, err = blockchain.generateBridgeAggInstruction(curView, retrievedShardBlockForBridgeAgg, newBeaconBlock)
	Logger.log.Info("[Bridge Debug] Generate bridge agg unshield instruction", len(bridgeAggInstructions))
	if err != nil {
		return nil, nil, NewBlockChainError(BuildBridgeAggError, err)
	}

	bridgeInstructions = append(bridgeInstructions, statefulInsts...)
	bridgeInstructions = append(bridgeInstructions, bridgeAggInstructions...)
	shardInstruction.compose()

	//outdatedPendingValidator := map[int][]int{}
	//if dequeueInst != nil && len(dequeueInst.DequeueList) > 0 {
	//	outdatedPendingValidator = dequeueInst.DequeueList
	//}
	instructions, err := curView.GenerateInstruction(
		newBeaconBlock.Header.Height, shardInstruction, duplicateKeyStakeInstructions,
		bridgeInstructions, acceptedRewardInstructions,
		blockchain, shardStates, committeeChange,
	)
	if err != nil {
		return nil, nil, err
	}
	if len(instructions) > 0 {
		BLogger.log.Infof("Producer instructions: %+v", instructions)
	}

	if blockchain.IsFirstBeaconHeightInEpoch(newBeaconBlock.Header.Height) {
		featureStateDB := curView.GetBeaconFeatureStateDB()
		cloneBeaconBestState, err := blockchain.GetClonedBeaconBestState()
		if err != nil {
			return nil, nil, NewBlockChainError(CloneBeaconBestStateError, err)
		}
		totalLockedCollateral, err := portalprocessv3.GetTotalLockedCollateralInEpoch(
			featureStateDB,
			cloneBeaconBestState.portalStateV3)
		if err != nil {
			return nil, nil, NewBlockChainError(GetTotalLockedCollateralError, err)
		}
		portalParamsV3 := portalParams.GetPortalParamsV3(newBeaconBlock.GetHeight())
		isSplitRewardForCustodian := totalLockedCollateral > 0
		percentCustodianRewards := portalParamsV3.MaxPercentCustodianRewards
		if totalLockedCollateral < portalParamsV3.MinLockCollateralAmountInEpoch {
			percentCustodianRewards = portalParamsV3.MinPercentCustodianRewards
		}

		isSplitRewardForPdex := curView.BeaconHeight >= config.Param().PDexParams.Pdexv3BreakPointHeight

		pdexRewardPercent := uint(0)
		if isSplitRewardForPdex {
			pdexRewardPercent = curView.pdeStates[pdex.AmplifierVersion].Reader().Params().DAOContributingPercent
		}

		rewardByEpochInstruction, rewardForCustodianByEpoch, pdexReward, err = blockchain.buildRewardInstructionByEpoch(
			curView,
			newBeaconBlock.Header.Height,
			curView.Epoch,
			isSplitRewardForCustodian,
			percentCustodianRewards,
			isSplitRewardForPdex,
			pdexRewardPercent,
			newBeaconBlock.Header.Version,
			committeeChange,
		)
		if err != nil {
			return nil, nil, NewBlockChainError(BuildRewardInstructionError, err)
		}
	}

	if len(rewardByEpochInstruction) != 0 {
		instructions = append(instructions, rewardByEpochInstruction...)
	}

	return instructions, shardStates, nil
}

// GetShardStateFromBlock get state (information) from shard-to-beacon block
// state will be presented as instruction
//
//	Return Params:
//	1. ShardState
//	2. Stake Instruction
//	3. Swap Instruction
//	4. Bridge Instruction
//	5. Accepted BlockReward Instruction
//	6. StopAutoStakingInstruction
func (blockchain *BlockChain) GetShardStateFromBlock(
	curView *BeaconBestState,
	newBeaconHeight uint64,
	shardBlock *types.ShardBlock,
	shardID byte,
	validUnstakePublicKeys map[string]bool,
	validStakePublicKeys []string,
) (map[byte]types.ShardState, *shardInstruction, *duplicateKeyStakeInstruction,
	[]string, [][]string, map[uint][]metadata.Transaction, error) {
	//Variable Declaration
	shardStates := make(map[byte]types.ShardState)
	duplicateKeyStakeInstruction := &duplicateKeyStakeInstruction{}

	acceptedRewardInstruction := curView.getAcceptBlockRewardInstruction(shardID, shardBlock, blockchain)

	prevShardBlockValidatorIndex := ""
	if curView.BestBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION {
		prevShardBlock, _, err := blockchain.ShardChain[shardID].BlockStorage.GetBlockWithLatestValidationData(shardBlock.GetPrevHash())
		if err != nil {
			return nil, nil, nil, nil, nil, nil, errors.New("Cannot find previous shard block for get validator index")
		}
		prevShardBlockValidatorIndex = prevShardBlock.(*types.ShardBlock).ValidationData
	}

	//Get Shard State from Block
	shardStates[shardID] = types.NewShardState(
		shardBlock.ValidationData,
		prevShardBlockValidatorIndex,
		shardBlock.Header.CommitteeFromBlock,
		shardBlock.Header.Height,
		shardBlock.Header.Hash(),
		shardBlock.Header.CrossShardBitMap,
		shardBlock.Header.ProposeTime,
		shardBlock.Header.Version,
	)
	instructions, pdexTxs, err := CreateShardInstructionsFromTransactionAndInstruction(
		shardBlock.Body.Transactions, blockchain,
		shardID, shardBlock.Header.Height, shardBlock.Header.BeaconHeight, true,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	instructions = append(instructions, shardBlock.Body.Instructions...)

	shardInstruction := curView.preProcessInstructionsFromShardBlock(instructions, shardID)

	allCommitteeValidatorCandidate := []string{}
	if len(shardInstruction.stopAutoStakeInstructions) != 0 || len(shardInstruction.unstakeInstructions) != 0 ||
		len(shardInstruction.stakeInstructions) != 0 || len(shardInstruction.redelegateInstructions) != 0 {
		allCommitteeValidatorCandidate = curView.getAllCommitteeValidatorCandidateFlattenList()
	}

	shardInstruction, duplicateKeyStakeInstruction = curView.
		processStakeInstructionFromShardBlock(shardInstruction, validStakePublicKeys, allCommitteeValidatorCandidate)
	shardInstruction = curView.processStopAutoStakeInstructionFromShardBlock(shardInstruction, allCommitteeValidatorCandidate)
	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v shard redelegate inst %+v\n", shardBlock.Header.Height, shardID, len(shardInstruction.redelegateInstructions))
	shardInstruction = curView.processReDelegateInstructionFromShardBlock(shardInstruction, allCommitteeValidatorCandidate)
	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v shard redelegate inst %+v\n", shardBlock.Header.Height, shardID, len(shardInstruction.redelegateInstructions))
	shardInstruction = curView.processUnstakeInstructionFromShardBlock(
		shardInstruction, allCommitteeValidatorCandidate, shardID, validUnstakePublicKeys)
	// for _, redelegateins := range shardInstruction.redelegateInstructions {
	// 	Logger.log.Infof("Becon Produce: Got Shard Block %+v Shard %+v shard inst %+v %+v\n", shardBlock.Header.Height, shardID, redelegateins.CommitteePublicKeys, redelegateins.DelegateList)
	// }
	// Collect stateful actions
	statefulActions := collectStatefulActions(instructions)

	return shardStates, shardInstruction, duplicateKeyStakeInstruction, acceptedRewardInstruction, statefulActions, pdexTxs, nil
}

func (curView *BeaconBestState) getAcceptBlockRewardInstruction(
	shardID byte,
	shardBlock *types.ShardBlock,
	blockchain *BlockChain,
) []string {
	if shardBlock.Header.BeaconHeight >= config.Param().ConsensusParam.BlockProducingV3Height && shardBlock.GetVersion() < types.INSTANT_FINALITY_VERSION_V2 {
		timeSlot := curView.CalculateTimeSlot(shardBlock.GetProposeTime())
		subsetID := GetSubsetIDFromProposerTimeV2(
			timeSlot,
			curView.GetShardProposerLength(),
		)
		acceptedRewardInstruction := instruction.NewAcceptBlockRewardV3WithValue(
			byte(subsetID), shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)

		return acceptedRewardInstruction.String()
	} else {
		acceptedBlockRewardInfo := instruction.NewAcceptBlockRewardV1WithValue(
			shardID, shardBlock.Header.TotalTxsFee, shardBlock.Header.Height)
		acceptedRewardInstruction, err := acceptedBlockRewardInfo.String()
		if err != nil {
			// if err then ignore accepted reward instruction
			Logger.log.Error(NewBlockChainError(GenerateInstructionError, err))
			return []string{}
		}

		return acceptedRewardInstruction
	}
}

// GenerateInstruction generate instruction for new beacon block
func (curView *BeaconBestState) GenerateInstruction(
	newBeaconHeight uint64,
	shardInstruction *shardInstruction,
	duplicateKeyStakeInstruction *duplicateKeyStakeInstruction,
	bridgeInstructions [][]string,
	acceptedRewardInstructions [][]string,
	blockchain *BlockChain,
	shardsState map[byte][]types.ShardState,
	committeeChange *committeestate.CommitteeChange,
) ([][]string, error) {
	instructions := [][]string{}
	instructions = append(instructions, bridgeInstructions...)
	instructions = append(instructions, acceptedRewardInstructions...)

	// Stake
	for _, stakeInstruction := range shardInstruction.stakeInstructions {
		instructions = append(instructions, stakeInstruction.ToString())
	}

	// Duplicate Staking Instruction
	for _, stakeInstruction := range duplicateKeyStakeInstruction.instructions {
		if len(stakeInstruction.TxStakes) > 0 {
			returnStakingIns := instruction.NewReturnStakeInsWithValue(
				stakeInstruction.PublicKeys,
				stakeInstruction.TxStakes,
			)
			instructions = append(instructions, returnStakingIns.ToString())
		}
	}

	// Shard Swap: both abnormal or normal swap
	var keys []int
	for k := range shardInstruction.swapInstructions {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, tempSwapInstruction := range shardInstruction.swapInstructions[byte(shardID)] {
			instructions = append(instructions, tempSwapInstruction.ToString())
		}
	}

	// Random number for Assign Instruction
	if blockchain.IsGreaterThanRandomTime(newBeaconHeight) && !curView.IsGetRandomNumber {
		randomInstructionGenerator := curView.beaconCommitteeState.(committeestate.RandomInstructionsGenerator)
		randomInstruction, randomNumber := randomInstructionGenerator.GenerateRandomInstructions(&committeestate.BeaconCommitteeStateEnvironment{
			BeaconHash:    curView.BestBlockHash,
			BestShardHash: curView.BestShardHash,
			ActiveShards:  curView.ActiveShards,
		})
		instructions = append(instructions, randomInstruction.ToString())
		Logger.log.Infof("Beacon Producer found Random Instruction at Block Height %+v, %+v", randomInstruction, newBeaconHeight)

		if curView.CommitteeStateVersion() == committeestate.SELF_SWAP_SHARD_VERSION {
			env := committeestate.NewBeaconCommitteeStateEnvironmentForAssigningToPendingList(
				randomNumber,
				config.Param().SwapCommitteeParam.AssignOffset,
				newBeaconHeight,
			)
			assignInstructionGenerator := curView.beaconCommitteeState.(*committeestate.BeaconCommitteeStateV1)
			assignInstructions := assignInstructionGenerator.GenerateAssignInstructions(env)
			for _, assignInstruction := range assignInstructions {
				instructions = append(instructions, assignInstruction.ToString())
			}
			Logger.log.Info("assignInstructions:", assignInstructions)
		}
	}

	// Unstake
	for _, unstakeInstruction := range shardInstruction.unstakeInstructions {
		instructions = append(instructions, unstakeInstruction.ToString())
	}

	// Generate swap shard instruction at block height %chainParamEpoch == 0
	if curView.CommitteeStateVersion() == committeestate.SELF_SWAP_SHARD_VERSION {
		if blockchain.IsLastBeaconHeightInEpoch(newBeaconHeight) {
			BeaconCommittee := curView.GetBeaconCommittee()
			beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(BeaconCommittee)
			if err != nil {
				Logger.log.Error(err)
			}
			epoch := blockchain.GetEpochByHeight(newBeaconHeight)
			if common.IndexOfUint64(epoch, config.Param().ConsensusParam.EpochBreakPointSwapNewKey) > -1 {
				swapBeaconInstructions, beaconCommittee := createBeaconSwapActionForKeyListV2(beaconCommitteeStr, curView.MinBeaconCommitteeSize, epoch)
				instructions = append(instructions, swapBeaconInstructions)
				beaconRootInst, _ := buildBeaconSwapConfirmInstruction(beaconCommittee, newBeaconHeight)
				instructions = append(instructions, beaconRootInst)
			}
		}
	} else {
		//swap outdated pending validator to syncing pool at last height of epoch
		//if blockchain.IsLastBeaconHeightInEpoch(newBeaconHeight) {
		//	if len(outdatedPendingValidator) > 0 {
		//		dequeueInst := instruction.NewDequeueInstructionWithValue(instruction.OUTDATED_DEQUEUE_REASON, outdatedPendingValidator)
		//		instructions = append(instructions, dequeueInst.ToString())
		//	}
		//}

		//swap validator in committee to pending validator
		if blockchain.IsFirstBeaconHeightInEpoch(newBeaconHeight) {
			// Generate request shard swap instruction, only available after upgrade to BeaconCommitteeEngineV2
			env := curView.NewBeaconCommitteeStateEnvironment()
			env.LatestShardsState = shardsState
			var swapShardInstructionsGenerator committeestate.SwapShardInstructionsGenerator
			switch curView.beaconCommitteeState.Version() {
			case committeestate.STAKING_FLOW_V2:
				swapShardInstructionsGenerator = curView.beaconCommitteeState.(*committeestate.BeaconCommitteeStateV2)
			case committeestate.STAKING_FLOW_V3:
				swapShardInstructionsGenerator = curView.beaconCommitteeState.(*committeestate.BeaconCommitteeStateV3)
			case committeestate.STAKING_FLOW_V4:
				swapShardInstructionsGenerator = curView.beaconCommitteeState.(*committeestate.BeaconCommitteeStateV4)
			}
			swapShardInstructions, slashedChange, err := swapShardInstructionsGenerator.GenerateSwapShardInstructions(env)
			if err != nil {
				return [][]string{}, err
			}
			for _, swapShardInstruction := range swapShardInstructions {
				if !swapShardInstruction.IsEmpty() {
					instructions = append(instructions, swapShardInstruction.ToString())
				}
			}
			committeeChange = slashedChange
		}
	}

	// Stop Auto Stake
	for _, stopAutoStakeInstruction := range shardInstruction.stopAutoStakeInstructions {
		instructions = append(instructions, stopAutoStakeInstruction.ToString())
	}
	Logger.log.Infof("Beacon %v generate instructions from inst %+v %+v", curView.GetBeaconHeight(), len(shardInstruction.redelegateInstructions), shardInstruction.redelegateInstructions)

	for _, redelegateInstruction := range shardInstruction.redelegateInstructions {
		instructions = append(instructions, redelegateInstruction.ToString())
	}

	for _, addStakingInstruction := range shardInstruction.addStakingInstructions {
		instructions = append(instructions, addStakingInstruction.ToString())
	}

	return instructions, nil
}

func addFinishInstruction(
	instructions, res [][]string) [][]string {

	instructions = append(instructions, res...)

	return instructions
}

////generate dequeue instruction , to push node into sync pool
//func (curView *BeaconBestState) generateOutdatedDequeueInstruction() *instruction.DequeueInstruction {
//
//	expectedContainFeature := []string{}
//	unTriggerFeatures := curView.getUntriggerFeature()
//
//	for _, feature := range unTriggerFeatures {
//		autoEnableFeatureInfo, ok := config.Param().AutoEnableFeature[feature]
//		if !ok {
//			continue
//		}
//		//check timing condition
//		if uint64(autoEnableFeatureInfo.ForceBlockHeight) > curView.BeaconHeight {
//			continue
//		}
//
//		expectedContainFeature = append(expectedContainFeature, feature)
//	}
//
//	//loop all shard pending validators, check if the validator code is latest or not
//	outdatedValidatorIndex := map[int][]int{} // shardID -> idnex
//	for cid := 0; cid < curView.ActiveShards; cid++ {
//		pendingList, err := incognitokey.CommitteeKeyListToString(curView.GetAShardPendingValidator(byte(cid)))
//		committeeList, err := incognitokey.CommitteeKeyListToString(curView.GetAShardCommittee(byte(cid)))
//		if err != nil {
//			Logger.log.Infof("Get Committee from shard %v error %v", cid, err)
//			return nil
//		}
//		for validatorIndex, cpk := range pendingList {
//			if DefaultFeatureStat.containExpectedFeature(cpk, expectedContainFeature) == false {
//				outdatedValidatorIndex[cid] = append(outdatedValidatorIndex[cid], validatorIndex)
//			}
//		}
//		if len(outdatedValidatorIndex[cid]) >= int((float64(len(pendingList)+len(committeeList)))*DEQUEUE_THRESHOLD_PERCENT) {
//			Logger.log.Infof("Chain %v cannot generate dequeue, not enough updated node, outdate %v , validator: %v", cid, len(outdatedValidatorIndex[cid]), (len(pendingList) + len(committeeList)))
//			delete(outdatedValidatorIndex, cid)
//		}
//	}
//
//	return instruction.NewDequeueInstructionWithValue(instruction.OUTDATED_DEQUEUE_REASON, outdatedValidatorIndex)
//}

func (curView *BeaconBestState) generateEnableFeatureInstructions() ([][]string, []string) {
	instructions := [][]string{}
	enableFeature := []string{}
	// get valid untrigger feature
	unTriggerFeatures := curView.getUntriggerFeature(false)

	Logger.log.Info("[committee-state] unTriggerFeatures:", unTriggerFeatures)

	for _, feature := range unTriggerFeatures {
		autoEnableFeatureInfo, ok := config.Param().AutoEnableFeature[feature]
		if !ok {
			continue
		}
		Logger.log.Info("[committee-state] feature:", feature)
		Logger.log.Info("[committee-state] autoEnableFeatureInfo:", autoEnableFeatureInfo)
		if uint64(autoEnableFeatureInfo.MinTriggerBlockHeight) > curView.BeaconHeight {
			continue
		}

		// check proposer threshold
		invalidCondition := false
		featureStatReport := DefaultFeatureStat.Report(curView)
		if featureStatReport.CommitteeStat[feature] == nil {
			continue
		}
		beaconProposerSize := len(curView.GetCommittee())
		//if number of beacon proposer update < 95%, not generate inst
		if featureStatReport.CommitteeStat[feature][-1] < uint64(math.Ceil(float64(beaconProposerSize)*95/100)) {
			continue
		}

		//if number of each shard committee update < 95%, not generate inst
		for chainID := 0; chainID < curView.ActiveShards; chainID++ {
			shardCommitteeSize := len(curView.GetAShardCommittee(byte(chainID)))
			if featureStatReport.CommitteeStat[feature][chainID] < uint64(math.Ceil(float64(shardCommitteeSize)*89/100)) {
				invalidCondition = true
				break
			}
		}

		if invalidCondition {
			continue
		}

		//check validator threshold
		if featureStatReport.ValidatorStat[feature] != nil {
			for chainID, size := range featureStatReport.ValidatorSize {
				if featureStatReport.ValidatorStat[feature][chainID] < uint64(math.Ceil(float64(size*autoEnableFeatureInfo.RequiredPercentage)/100)) {
					invalidCondition = true
					break
				}
			}
		}
		if invalidCondition {
			continue
		}

		Logger.log.Info("[committee-state] add feature:", feature)

		enableFeature = append(enableFeature, feature)
	}

	if len(enableFeature) > 0 && curView.BeaconHeight == GetFirstBeaconHeightInEpoch(curView.Epoch) {
		//generate instruction for valid condition
		inst := instruction.NewEnableFeatureInstructionWithValue(enableFeature)
		instructions = append(instructions, inst.ToString())
	}
	return instructions, enableFeature
}

func (curView *BeaconBestState) halfPendingCycleEpoch(sid byte) uint64 {
	swapOffset := uint64(curView.MaxShardCommitteeSize / 8)
	halfPendingCycleEpoch := math.Ceil(float64(len(curView.GetShardPendingValidator()[sid])) / float64(2*swapOffset))
	return uint64(halfPendingCycleEpoch)
}

func (curView *BeaconBestState) generateFinishSyncInstruction() [][]string {
	//get validators in sync pool that contain latest code
	syncVal := make(map[byte][]string)
	for shardID, validators := range curView.beaconCommitteeState.GetSyncingValidators() {
		validValidator := []incognitokey.CommitteePublicKey{}
		for _, v := range validators {
			validatorStr, _ := v.ToBase58()
			if DefaultFeatureStat.IsContainLatestFeature(curView, validatorStr) {
				//fmt.Println("add ", validatorStr, "to valid Sync val")
				validValidator = append(validValidator, v)
			}
		}
		syncVal[shardID], _ = incognitokey.CommitteeKeyListToString(validValidator)
	}

	//get valid waiting validator
	validWaitingValidator := map[byte][]string{}
	for sid, validators := range syncVal {
		halfPendingCycleEpoch := curView.halfPendingCycleEpoch(sid)
		for _, validator := range validators {
			info, exists, err := curView.GetStakerInfo(validator)
			if !exists || err != nil {
				panic("Error when generateFinishSyncInstruction. This must not occur!")
			}
			if curView.BeaconHeight >= info.BeaconConfirmHeight()+halfPendingCycleEpoch*config.Param().EpochParam.NumberOfBlockInEpoch {
				validWaitingValidator[sid] = append(validWaitingValidator[sid], validator)
			}
		}
	}

	finishSyncInstructions := finishsync.DefaultFinishSyncMsgPool.Instructions(validWaitingValidator, curView.BeaconHeight)
	instructions := [][]string{}

	for _, finishSyncInstruction := range finishSyncInstructions {
		if !finishSyncInstruction.IsEmpty() {
			instructions = append(instructions, finishSyncInstruction.ToString())
		}
	}

	return instructions
}

func filterEnableFeatureInstruction(instructions [][]string) [][]string {
	enableFeatureInstructions := [][]string{}
	for _, v := range instructions {
		if v[0] == instruction.ENABLE_FEATURE {
			enableFeatureInstructions = append(enableFeatureInstructions, v)
		}
	}
	return enableFeatureInstructions
}

//func filterDequeueInstruction(instructions [][]string, reason string) (*instruction.DequeueInstruction, error) {
//	for _, v := range instructions {
//		if v[0] == instruction.DEQUEUE && v[1] == reason {
//			return instruction.ValidateAndImportDequeueInstructionFromString(v)
//		}
//	}
//	return nil, nil
//}

func (curView *BeaconBestState) filterAndVerifyFinishSyncInstruction(instructions [][]string) ([][]string, error) {

	finishSyncInstructions := [][]string{}

	syncValidators := curView.GetSyncingValidatorsString()
	for _, v := range instructions {
		if v[0] == instruction.FINISH_SYNC_ACTION {
			inst, err := instruction.ValidateAndImportFinishSyncInstructionFromString(v)
			if err != nil {
				return nil, err
			}
			finishSyncInstructions = append(finishSyncInstructions, v)

			//verify staker have valid waiting time
			for _, validator := range inst.PublicKeys {
				info, exists, err := curView.GetStakerInfo(validator)
				if !exists || err != nil {
					fmt.Println("finishSyncInstructions", v[2])
					panic("Error when generateFinishSyncInstruction. This must not occur!")
				}

				//loop sync pool, check if validator is exist and having valid waiting time
				exist := false
				for sid, vals := range syncValidators {
					if common.IndexOfStr(validator, vals) != -1 {
						halfPendingCycleEpoch := curView.halfPendingCycleEpoch(sid)
						if curView.BeaconHeight < info.BeaconConfirmHeight()+halfPendingCycleEpoch*config.Param().EpochParam.NumberOfBlockInEpoch {
							return nil, fmt.Errorf("Not valid waiting time for syncing validator, current beacon %v, expect valid height %v", curView.BeaconHeight, info.BeaconConfirmHeight()+halfPendingCycleEpoch*config.Param().EpochParam.NumberOfBlockInEpoch)
						} else {
							exist = true
						}
					}
				}

				//cannot find validator in sync pool
				if !exist {
					return nil, fmt.Errorf("Cannot find validator %v in syncing pool", validator)
				}
			}
		}
	}

	return finishSyncInstructions, nil
}

func createBeaconSwapActionForKeyListV2(
	beaconCommittees []string,
	minCommitteeSize int,
	epoch uint64,
) ([]string, []string) {
	swapInstruction, newBeaconCommittees := GetBeaconSwapInstructionKeyListV2(epoch)
	remainBeaconCommittees := beaconCommittees[minCommitteeSize:]
	return swapInstruction, append(newBeaconCommittees, remainBeaconCommittees...)
}

func (beaconBestState *BeaconBestState) preProcessInstructionsFromShardBlock(instructions [][]string, shardID byte) *shardInstruction {
	shardInstruction := newShardInstruction()
	// extract instructions

	waitingValidatorsList, err := incognitokey.CommitteeKeyListToString(beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom())
	if err != nil {
		return shardInstruction
	}

	for _, inst := range instructions {
		if len(inst) > 0 {
			if inst[0] == instruction.STAKE_ACTION {
				if err := instruction.ValidateStakeInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Stake Instruction Error %+v", err)
					continue
				}
				tempStakeInstruction := instruction.ImportStakeInstructionFromString(inst)
				shardInstruction.stakeInstructions = append(shardInstruction.stakeInstructions, tempStakeInstruction)
			}
			if inst[0] == instruction.SWAP_ACTION {
				// validate swap instruction
				// only allow shard to swap committee for it self
				if err := instruction.ValidateSwapInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Swap Instruction Error %+v", err)
					continue
				}
				tempSwapInstruction := instruction.ImportSwapInstructionFromString(inst)
				shardInstruction.swapInstructions[shardID] = append(shardInstruction.swapInstructions[shardID], tempSwapInstruction)
			}
			if inst[0] == instruction.STOP_AUTO_STAKE_ACTION {
				if err := instruction.ValidateStopAutoStakeInstructionSanity(inst); err != nil {
					Logger.log.Errorf("SKIP Stop Auto Stake Instruction Error %+v", err)
					continue
				}
				tempStopAutoStakeInstruction := instruction.ImportStopAutoStakeInstructionFromString(inst)
				for i := 0; i < len(tempStopAutoStakeInstruction.CommitteePublicKeys); i++ {
					v := tempStopAutoStakeInstruction.CommitteePublicKeys[i]
					check, ok := beaconBestState.GetAutoStakingList()[v]
					if !ok {
						Logger.log.Errorf("[stop-autoStaking] Committee %s is not found or has already been unstaked:", v)
					}
					if !ok || !check {
						tempStopAutoStakeInstruction.DeleteSingleElement(i)
						i--
					}
				}
				if len(tempStopAutoStakeInstruction.CommitteePublicKeys) != 0 {
					shardInstruction.stopAutoStakeInstructions = append(shardInstruction.stopAutoStakeInstructions, tempStopAutoStakeInstruction)
				}
			}
			if inst[0] == instruction.RE_DELEGATE {
				Logger.log.Info("Beacon Producer/ Process redelegate List ", inst)
				if redelegateIns, err := instruction.ValidateAndImportReDelegateInstructionFromString(inst); err != nil {
					Logger.log.Errorf("SKIP Redelegate Instruction Error %+v", err)
					continue
				} else {
					if len(redelegateIns.CommitteePublicKeys) != 0 {
						shardInstruction.redelegateInstructions = append(shardInstruction.redelegateInstructions, redelegateIns)
					}
				}
			}
			if inst[0] == instruction.ADD_STAKING_ACTION {
				Logger.log.Infof("Beacon Producer/ Process add staking list %v", inst)
				if addStakingIns, err := instruction.ValidateAndImportAddStakingInstructionFromString(inst); err != nil {
					Logger.log.Errorf("SKIP Add Staking Instruction Error %+v", err)
					continue
				} else {
					if len(addStakingIns.CommitteePublicKeys) != 0 {
						shardInstruction.addStakingInstructions = append(shardInstruction.addStakingInstructions, addStakingIns)
					}
				}
			}
			if inst[0] == instruction.UNSTAKE_ACTION {
				if err := instruction.ValidateUnstakeInstructionSanity(inst); err != nil {
					Logger.log.Errorf("[unstaking] SKIP Stop Auto Stake Instruction Error %+v", err)
					continue
				}
				tempUnstakeInstruction := instruction.ImportUnstakeInstructionFromString(inst)
				for i := 0; i < len(tempUnstakeInstruction.CommitteePublicKeys); i++ {
					v := tempUnstakeInstruction.CommitteePublicKeys[i]
					index := common.IndexOfStr(v, waitingValidatorsList)
					if index == -1 {
						check, ok := beaconBestState.GetAutoStakingList()[v]
						if !ok {
							Logger.log.Errorf("[unstaking] Committee %s is not found or has already been unstaked:", v)
						}
						if !ok || !check {
							tempUnstakeInstruction.DeleteSingleElement(i)
							i--
						}
					}
				}
				if len(tempUnstakeInstruction.CommitteePublicKeys) != 0 {
					shardInstruction.unstakeInstructions = append(shardInstruction.unstakeInstructions, tempUnstakeInstruction)
				}
			}
		}
	}

	if len(shardInstruction.stakeInstructions) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List ", shardInstruction.stakeInstructions)
	}
	if len(shardInstruction.addStakingInstructions) != 0 {
		Logger.log.Info("Beacon Producer/ Process Stakers List:")
		for _, inst := range shardInstruction.addStakingInstructions {
			Logger.log.Infof("add staking inst %v", inst.ToString())
		}
	}
	if len(shardInstruction.swapInstructions[shardID]) != 0 {
		Logger.log.Info("Beacon Producer/ Process Swap List ", shardInstruction.swapInstructions[shardID])
	}

	return shardInstruction
}

func (beaconBestState *BeaconBestState) processStakeInstructionFromShardBlock(
	shardInstructions *shardInstruction, validStakePublicKeys []string, allCommitteeValidatorCandidate []string) (
	*shardInstruction, *duplicateKeyStakeInstruction) {

	duplicateKeyStakeInstruction := &duplicateKeyStakeInstruction{}
	newShardInstructions := shardInstructions
	newStakeInstructions := []*instruction.StakeInstruction{}
	stakeShardPublicKeys := []string{}
	stakeShardDelegateList := []string{}
	stakeShardTx := []string{}
	stakeShardRewardReceiver := []string{}
	stakeShardAutoStaking := []bool{}
	stakeBeaconPublicKeys := []string{}
	stakeBeaconTx := []string{}
	stakeBeaconRewardReceiver := []string{}
	stakeBeaconAutoStaking := []bool{}
	stakeBeaconStakingAmount := []uint64{}

	// Process Stake Instruction form Shard Block
	// Validate stake instruction => extract only valid stake instruction
	for _, stakeInstruction := range shardInstructions.stakeInstructions {
		if stakeInstruction.Chain == instruction.SHARD_INST {
			tempStakePublicKey := make([]string, len(stakeInstruction.PublicKeys))
			copy(tempStakePublicKey, stakeInstruction.PublicKeys)
			duplicateStakePublicKeys := []string{}
			// list of stake public keys and stake transaction and reward receiver must have equal length

			tempStakePublicKey = beaconBestState.GetValidStakers(tempStakePublicKey, false)
			tempStakePublicKey = common.GetValidStaker(stakeShardPublicKeys, tempStakePublicKey)
			tempStakePublicKey = common.GetValidStaker(validStakePublicKeys, tempStakePublicKey)
			tempStakePublicKey = common.GetValidStaker(allCommitteeValidatorCandidate, tempStakePublicKey)

			if len(tempStakePublicKey) > 0 {
				stakeShardPublicKeys = append(stakeShardPublicKeys, tempStakePublicKey...)
				for i, v := range stakeInstruction.PublicKeys {
					if common.IndexOfStr(v, tempStakePublicKey) > -1 {
						stakeShardTx = append(stakeShardTx, stakeInstruction.TxStakes[i])
						stakeShardRewardReceiver = append(stakeShardRewardReceiver, stakeInstruction.RewardReceivers[i])
						stakeShardAutoStaking = append(stakeShardAutoStaking, stakeInstruction.AutoStakingFlag[i])
						stakeShardDelegateList = append(stakeShardDelegateList, stakeInstruction.DelegateList[i])
					}
				}
			}

			if beaconBestState.beaconCommitteeState.Version() != committeestate.SELF_SWAP_SHARD_VERSION &&
				(len(stakeInstruction.PublicKeys) != len(tempStakePublicKey)) {
				duplicateStakePublicKeys = committeestate.DifferentElementStrings(stakeInstruction.PublicKeys, tempStakePublicKey)
				if len(duplicateStakePublicKeys) > 0 {
					stakingTxs := []string{}
					autoStaking := []bool{}
					rewardReceivers := []string{}
					delegateList := []string{}
					for i, v := range stakeInstruction.PublicKeys {
						if common.IndexOfStr(v, duplicateStakePublicKeys) > -1 {
							stakingTxs = append(stakingTxs, stakeInstruction.TxStakes[i])
							rewardReceivers = append(rewardReceivers, stakeInstruction.RewardReceivers[i])
							autoStaking = append(autoStaking, stakeInstruction.AutoStakingFlag[i])
							delegateList = append(delegateList, stakeInstruction.DelegateList[i])
						}
					}
					duplicateStakeInstruction := instruction.NewShardStakeInstructionWithValue(
						duplicateStakePublicKeys,
						stakeInstruction.Chain,
						stakingTxs,
						rewardReceivers,
						autoStaking,
						delegateList,
					)
					duplicateKeyStakeInstruction.instructions = append(duplicateKeyStakeInstruction.instructions, duplicateStakeInstruction)
				}
			}
		} else {
			tempStakePublicKey := make([]string, len(stakeInstruction.PublicKeys))
			copy(tempStakePublicKey, stakeInstruction.PublicKeys)
			duplicateStakePublicKeys := []string{}
			// list of stake public keys and stake transaction and reward receiver must have equal length

			tempStakePublicKey = beaconBestState.GetValidStakers(tempStakePublicKey, true)
			tempStakePublicKey = common.GetValidStaker(stakeShardPublicKeys, tempStakePublicKey)
			tempStakePublicKey = common.GetValidStaker(validStakePublicKeys, tempStakePublicKey)
			// tempStakePublicKey = common.GetValidStaker(allCommitteeValidatorCandidate, tempStakePublicKey)

			if len(tempStakePublicKey) > 0 {
				stakeBeaconPublicKeys = append(stakeBeaconPublicKeys, tempStakePublicKey...)
				for i, v := range stakeInstruction.PublicKeys {
					if common.IndexOfStr(v, tempStakePublicKey) > -1 {
						stakeBeaconTx = append(stakeBeaconTx, stakeInstruction.TxStakes[i])
						stakeBeaconRewardReceiver = append(stakeBeaconRewardReceiver, stakeInstruction.RewardReceivers[i])
						stakeBeaconAutoStaking = append(stakeBeaconAutoStaking, stakeInstruction.AutoStakingFlag[i])
						stakeBeaconStakingAmount = append(stakeBeaconStakingAmount, stakeInstruction.StakingAmount[i])
					}
				}
			}

			if beaconBestState.beaconCommitteeState.Version() != committeestate.SELF_SWAP_SHARD_VERSION &&
				(len(stakeInstruction.PublicKeys) != len(tempStakePublicKey)) {
				duplicateStakePublicKeys = committeestate.DifferentElementStrings(stakeInstruction.PublicKeys, tempStakePublicKey)
				if len(duplicateStakePublicKeys) > 0 {
					stakingTxs := []string{}
					autoStaking := []bool{}
					rewardReceivers := []string{}
					stakingAmountList := []uint64{}
					for i, v := range stakeInstruction.PublicKeys {
						if common.IndexOfStr(v, duplicateStakePublicKeys) > -1 {
							stakingTxs = append(stakingTxs, stakeInstruction.TxStakes[i])
							rewardReceivers = append(rewardReceivers, stakeInstruction.RewardReceivers[i])
							autoStaking = append(autoStaking, stakeInstruction.AutoStakingFlag[i])
							stakingAmountList = append(stakingAmountList, stakeInstruction.StakingAmount[i])
						}
					}
					duplicateStakeInstruction := instruction.NewBeaconStakeInstructionWithValue(
						duplicateStakePublicKeys,
						// stakeInstruction.Chain,
						stakingTxs,
						rewardReceivers,
						autoStaking,
						stakingAmountList,
					)
					duplicateKeyStakeInstruction.instructions = append(duplicateKeyStakeInstruction.instructions, duplicateStakeInstruction)
				}
			}
		}
	}

	if len(stakeShardPublicKeys) > 0 {
		// tempValidStakePublicKeys = append(tempValidStakePublicKeys, stakeShardPublicKeys...)
		tempStakeShardInstruction := instruction.NewShardStakeInstructionWithValue(
			stakeShardPublicKeys,
			instruction.SHARD_INST,
			stakeShardTx, stakeShardRewardReceiver,
			stakeShardAutoStaking,
			stakeShardDelegateList,
		)
		newStakeInstructions = append(newStakeInstructions, tempStakeShardInstruction)
	}
	if len(stakeBeaconPublicKeys) > 0 {
		// tempValidStakePublicKeys = append(tempValidStakePublicKeys, stakeShardPublicKeys...)
		tempStakeBeaconInstruction := instruction.NewBeaconStakeInstructionWithValue(
			stakeBeaconPublicKeys,
			stakeBeaconTx,
			stakeBeaconRewardReceiver,
			stakeBeaconAutoStaking,
			stakeBeaconStakingAmount,
		)
		newStakeInstructions = append(newStakeInstructions, tempStakeBeaconInstruction)
	}

	newShardInstructions.stakeInstructions = newStakeInstructions
	for _, ins := range newStakeInstructions {
		Logger.log.Infof("New staker: %+v %+v ", ins.DelegateList, ins.TxStakes)
	}
	return newShardInstructions, duplicateKeyStakeInstruction
}

func (beaconBestState *BeaconBestState) processStopAutoStakeInstructionFromShardBlock(
	shardInstructions *shardInstruction, allCommitteeValidatorCandidate []string) *shardInstruction {

	stopAutoStakingPublicKeys := []string{}
	stopAutoStakeInstructions := []*instruction.StopAutoStakeInstruction{}

	for _, stopAutoStakeInstruction := range shardInstructions.stopAutoStakeInstructions {
		for _, tempStopAutoStakingPublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
			if common.IndexOfStr(tempStopAutoStakingPublicKey, allCommitteeValidatorCandidate) > -1 {
				stopAutoStakingPublicKeys = append(stopAutoStakingPublicKeys, tempStopAutoStakingPublicKey)
			}
		}
	}

	if len(stopAutoStakingPublicKeys) > 0 {
		tempStopAutoStakeInstruction := instruction.NewStopAutoStakeInstructionWithValue(stopAutoStakingPublicKeys)
		stopAutoStakeInstructions = append(stopAutoStakeInstructions, tempStopAutoStakeInstruction)
	}

	shardInstructions.stopAutoStakeInstructions = stopAutoStakeInstructions
	return shardInstructions
}

func (beaconBestState *BeaconBestState) processReDelegateInstructionFromShardBlock(
	shardInstructions *shardInstruction, allCommitteeValidatorCandidate []string) *shardInstruction {

	redelegateList := [2][]string{}
	redelegateIns := []*instruction.ReDelegateInstruction{}

	for _, redelegateInstruction := range shardInstructions.redelegateInstructions {
		for idx, tempReDelegatePublicKey := range redelegateInstruction.CommitteePublicKeys {
			if common.IndexOfStr(tempReDelegatePublicKey, allCommitteeValidatorCandidate) > -1 {
				redelegateList[0] = append(redelegateList[0], tempReDelegatePublicKey)
				redelegateList[1] = append(redelegateList[1], redelegateInstruction.DelegateList[idx])
			}
		}
	}

	if len(redelegateList[0]) > 0 {
		tempReDelegateIns := instruction.NewReDelegateInstructionWithValue(redelegateList[0], redelegateList[1])
		redelegateIns = append(redelegateIns, tempReDelegateIns)
	}

	shardInstructions.redelegateInstructions = redelegateIns
	return shardInstructions
}

func (beaconBestState *BeaconBestState) processUnstakeInstructionFromShardBlock(
	shardInstructions *shardInstruction,
	allCommitteeValidatorCandidate []string,
	shardID byte,
	validUnstakePublicKeys map[string]bool) *shardInstruction {
	unstakePublicKeys := []string{}
	unstakeInstructions := []*instruction.UnstakeInstruction{}

	for _, unstakeInstruction := range shardInstructions.unstakeInstructions {
		for _, tempUnstakePublicKey := range unstakeInstruction.CommitteePublicKeys {
			if _, ok := validUnstakePublicKeys[tempUnstakePublicKey]; ok {
				Logger.log.Errorf("SHARD %v | UNSTAKE duplicated unstake instruction %+v ", shardID, tempUnstakePublicKey)
				continue
			}
			if common.IndexOfStr(tempUnstakePublicKey, allCommitteeValidatorCandidate) > -1 {
				unstakePublicKeys = append(unstakePublicKeys, tempUnstakePublicKey)
			}
			validUnstakePublicKeys[tempUnstakePublicKey] = true
		}
	}
	if len(unstakePublicKeys) > 0 {
		tempUnstakeInstruction := instruction.NewUnstakeInstructionWithValue(unstakePublicKeys)
		tempUnstakeInstruction.SetCommitteePublicKeys(unstakePublicKeys)
		unstakeInstructions = append(unstakeInstructions, tempUnstakeInstruction)
	}

	shardInstructions.unstakeInstructions = unstakeInstructions
	return shardInstructions

}

func (shardInstruction *shardInstruction) compose() {
	stakeShardInstruction := &instruction.StakeInstruction{}
	stakeBeaconInstruction := &instruction.StakeInstruction{}
	unstakeInstruction := &instruction.UnstakeInstruction{}
	stopAutoStakeInstruction := &instruction.StopAutoStakeInstruction{}
	redelegateInstruction := &instruction.ReDelegateInstruction{}
	addStakingInstruction := &instruction.AddStakingInstruction{}
	unstakeKeys := map[string]bool{}

	for _, v := range shardInstruction.stakeInstructions {
		if v.IsEmpty() {
			continue
		}
		if v.Chain == instruction.SHARD_INST {
			stakeShardInstruction.PublicKeys = append(stakeShardInstruction.PublicKeys, v.PublicKeys...)
			stakeShardInstruction.PublicKeyStructs = append(stakeShardInstruction.PublicKeyStructs, v.PublicKeyStructs...)
			stakeShardInstruction.TxStakeHashes = append(stakeShardInstruction.TxStakeHashes, v.TxStakeHashes...)
			stakeShardInstruction.TxStakes = append(stakeShardInstruction.TxStakes, v.TxStakes...)
			stakeShardInstruction.RewardReceivers = append(stakeShardInstruction.RewardReceivers, v.RewardReceivers...)
			stakeShardInstruction.RewardReceiverStructs = append(stakeShardInstruction.RewardReceiverStructs, v.RewardReceiverStructs...)
			stakeShardInstruction.Chain = v.Chain
			stakeShardInstruction.AutoStakingFlag = append(stakeShardInstruction.AutoStakingFlag, v.AutoStakingFlag...)
			stakeShardInstruction.DelegateList = append(stakeShardInstruction.DelegateList, v.DelegateList...)
		} else {
			stakeBeaconInstruction.PublicKeys = append(stakeBeaconInstruction.PublicKeys, v.PublicKeys...)
			stakeBeaconInstruction.PublicKeyStructs = append(stakeBeaconInstruction.PublicKeyStructs, v.PublicKeyStructs...)
			stakeBeaconInstruction.TxStakeHashes = append(stakeBeaconInstruction.TxStakeHashes, v.TxStakeHashes...)
			stakeBeaconInstruction.TxStakes = append(stakeBeaconInstruction.TxStakes, v.TxStakes...)
			stakeBeaconInstruction.RewardReceivers = append(stakeBeaconInstruction.RewardReceivers, v.RewardReceivers...)
			stakeBeaconInstruction.RewardReceiverStructs = append(stakeBeaconInstruction.RewardReceiverStructs, v.RewardReceiverStructs...)
			stakeBeaconInstruction.AutoStakingFlag = append(stakeBeaconInstruction.AutoStakingFlag, v.AutoStakingFlag...)
			stakeBeaconInstruction.StakingAmount = append(stakeBeaconInstruction.StakingAmount, v.StakingAmount...)
			stakeBeaconInstruction.Chain = v.Chain
		}
	}

	for _, v := range shardInstruction.redelegateInstructions {
		if v.IsEmpty() {
			continue
		}
		redelegateInstruction.CommitteePublicKeys = append(redelegateInstruction.CommitteePublicKeys, v.CommitteePublicKeys...)
		redelegateInstruction.CommitteePublicKeysStruct = append(redelegateInstruction.CommitteePublicKeysStruct, v.CommitteePublicKeysStruct...)
		redelegateInstruction.DelegateList = append(redelegateInstruction.DelegateList, v.DelegateList...)
		redelegateInstruction.DelegateListStruct = append(redelegateInstruction.DelegateListStruct, v.DelegateListStruct...)
	}

	for _, v := range shardInstruction.unstakeInstructions {
		if v.IsEmpty() {
			continue
		}
		for _, key := range v.CommitteePublicKeys {
			unstakeKeys[key] = true
		}
		unstakeInstruction.CommitteePublicKeys = append(unstakeInstruction.CommitteePublicKeys, v.CommitteePublicKeys...)
		unstakeInstruction.CommitteePublicKeysStruct = append(unstakeInstruction.CommitteePublicKeysStruct, v.CommitteePublicKeysStruct...)
	}

	for _, v := range shardInstruction.stopAutoStakeInstructions {
		if v.IsEmpty() {
			continue
		}

		committeePublicKeys := []string{}
		for _, key := range v.CommitteePublicKeys {
			if !unstakeKeys[key] {
				committeePublicKeys = append(committeePublicKeys, key)
			}
		}

		stopAutoStakeInstruction.CommitteePublicKeys = append(stopAutoStakeInstruction.CommitteePublicKeys, committeePublicKeys...)
	}
	for _, v := range shardInstruction.addStakingInstructions {
		if v.IsEmpty() {
			continue
		}

		committeePublicKeys := []string{}
		addStakingAmounts := []uint64{}
		stakingTxs := []string{}
		for id, key := range v.CommitteePublicKeys {
			if !unstakeKeys[key] {
				committeePublicKeys = append(committeePublicKeys, key)
			}
			addStakingAmounts = append(addStakingAmounts, v.StakingAmount[id])
			stakingTxs = append(stakingTxs, v.StakingTxIDs[id])
		}

		addStakingInstruction.CommitteePublicKeys = append(addStakingInstruction.CommitteePublicKeys, committeePublicKeys...)
		addStakingInstruction.StakingAmount = append(addStakingInstruction.StakingAmount, addStakingAmounts...)
		addStakingInstruction.StakingTxIDs = append(addStakingInstruction.StakingTxIDs, stakingTxs...)
	}

	shardInstruction.stakeInstructions = []*instruction.StakeInstruction{}
	shardInstruction.unstakeInstructions = []*instruction.UnstakeInstruction{}
	shardInstruction.stopAutoStakeInstructions = []*instruction.StopAutoStakeInstruction{}
	shardInstruction.redelegateInstructions = []*instruction.ReDelegateInstruction{}

	if !stakeShardInstruction.IsEmpty() {
		shardInstruction.stakeInstructions = append(shardInstruction.stakeInstructions, stakeShardInstruction)
	}
	if !stakeBeaconInstruction.IsEmpty() {
		shardInstruction.stakeInstructions = append(shardInstruction.stakeInstructions, stakeBeaconInstruction)
	}

	if !unstakeInstruction.IsEmpty() {
		shardInstruction.unstakeInstructions = append(shardInstruction.unstakeInstructions, unstakeInstruction)
	}

	if !stopAutoStakeInstruction.IsEmpty() {
		shardInstruction.stopAutoStakeInstructions = append(shardInstruction.stopAutoStakeInstructions, stopAutoStakeInstruction)
	}
	if !redelegateInstruction.IsEmpty() {
		shardInstruction.redelegateInstructions = append(shardInstruction.redelegateInstructions, redelegateInstruction)
	}
	if !addStakingInstruction.IsEmpty() {
		shardInstruction.addStakingInstructions = append(shardInstruction.addStakingInstructions, addStakingInstruction)
	}
}
