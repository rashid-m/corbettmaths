package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/config"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	portalcommonv3 "github.com/incognitochain/incognito-chain/portal/portalv3/common"
	portalcommonv4 "github.com/incognitochain/incognito-chain/portal/portalv4/common"
	"github.com/incognitochain/incognito-chain/privacy"

	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
)

// NewBlockShard Create New block Shard:
//  1. Identify Beacon State for this Shard Block: Beacon Hash & Beacon Height & Epoch
//     + Get Beacon Block (B) from Beacon Best State (from Beacon Chain of Shard Node)
//     + Beacon Block (B) must have the same epoch With New Shard Block (S):
//     + If Beacon Block (B) have different height previous shard block PS (previous of S)
//     Then Beacon Block (B) epoch greater than Shard Block (S) epoch exact 1 value
//     BUT This only works if Shard Best State have the Beacon Height divisible by epoch
//     + Ex: 1 epoch has 50 block
//     Example 1:
//     shard block with
//     height 10,
//     epoch: 1,
//     beacon block height: 49
//     then shard block with
//     height 11 must have
//     epoch: 1,
//     beacon block height: must be 49 or 50
//     Example 2:
//     shard block with
//     height 10,
//     epoch: 1,
//     beacon block height: 50
//     then shard block with
//     height is 11 can have 2 option:
//     a. epoch: 1, if beacon block height remain 50
//     b. epoch: 2, and beacon block must in range from 51-100
//     Can have beacon block with height > 100
//  2. Build Shard Block Body:
//     a. Get Cross Transaction from other shard && Build Cross Shard Tx Custom Token Transaction (if exist)
//     b. Get Transactions for New Block
//     c. Process Assign Instructions from Beacon Blocks
//     c. Generate Instructions
//  3. Build Shard Block Essential Data for Header
//  4. Update Cloned ShardBestState with New Shard Block
//  5. Create Root Hash from New Shard Block and updated Clone Shard Beststate Data
func (blockchain *BlockChain) NewBlockShard(curView *ShardBestState,
	version int, proposer string, round int, start int64,
	committees []incognitokey.CommitteePublicKey,
	committeeFinalViewHash common.Hash) (*types.ShardBlock, error) {
	var (
		newShardBlockBeginTime            = time.Now()
		transactionsForNewBlock           = make([]metadata.Transaction, 0)
		newShardBlock                     = types.NewShardBlock()
		shardInstructions                 = [][]string{}
		isOldBeaconHeight                 = false
		tempPrivateKey                    = blockchain.config.BlockGen.createTempKeyset()
		shardBestState                    = NewShardBestState()
		shardID                           = curView.ShardID
		currentCommitteePublicKeys        = []string{}
		currentCommitteePublicKeysStructs = []incognitokey.CommitteePublicKey{}
		committeeFromBlockHash            = common.Hash{}
		beaconProcessHeight               uint64
		err                               error
	)
	Logger.log.Criticalf("⛏ Creating Shard Block %+v", curView.ShardHeight+1)
	//check if expected final view is not confirmed by beacon for too far
	beaconFinalView := blockchain.BeaconChain.GetFinalView().(*BeaconBestState)
	if beaconFinalView.BestShardHeight[shardID]+100 < blockchain.ShardChain[shardID].multiView.GetExpectedFinalView().GetHeight() {
		return nil, fmt.Errorf("Shard %v | Wait for beacon shardstate %v, the unconfirmed view %v is too far away",
			shardID, beaconFinalView.BestShardHeight[shardID], curView.BestBlock.Hash().String())
	}

	//check if bestview is on the same branch with beacon shardstate
	blockByView, _ := blockchain.GetShardBlockByView(curView, beaconFinalView.BestShardHeight[shardID], shardID)
	if blockByView == nil {
		return nil, fmt.Errorf("Shard %v | Cannot get block by height: %v from view %v",
			shardID, beaconFinalView.BestShardHeight[shardID], curView.BestBlock.Hash().String())
	}
	if blockByView.GetHeight() != 1 && blockByView.Hash().String() != beaconFinalView.BestShardHash[shardID].String() {
		//update view
		if err := blockchain.ShardChain[shardID].multiView.FinalizeView(beaconFinalView.BestShardHash[shardID]); err != nil {
			//request missing view
			blockchain.config.Server.RequestMissingViewViaStream("", [][]byte{beaconFinalView.BestShardHash[shardID].Bytes()}, int(shardID), blockchain.ShardChain[shardID].GetChainName())
		}
		return nil, fmt.Errorf("Shard %v | Create block from view that is not on same branch with beacon shardstate, expect view %v height %v, got %v",
			shardID, beaconFinalView.BestShardHash[shardID].String(), beaconFinalView.BestShardHeight[shardID], curView.BestBlock.Hash().String())
	}

	// Clone best state value into new variable
	if err := shardBestState.cloneShardBestStateFrom(curView); err != nil {
		return nil, err
	}
	BLogger.log.Infof("Producing block: %d", shardBestState.ShardHeight+1)
	currentPendingValidators := shardBestState.GetShardPendingValidator()

	getBeaconFinalHeightForProcess := func() uint64 {
		view := blockchain.BeaconChain.GetFinalView().(*BeaconBestState)
		height := view.GetHeight()
		if height > shardBestState.BeaconHeight && height-shardBestState.BeaconHeight > MAX_BEACON_BLOCK {
			height = shardBestState.BeaconHeight + MAX_BEACON_BLOCK
		}
		return height
	}
	beaconProcessHeight = getBeaconFinalHeightForProcess()

	if shardBestState.CommitteeStateVersion() == committeestate.STAKING_FLOW_V2 {
		if beaconProcessHeight > config.Param().ConsensusParam.StakingFlowV3Height {
			beaconProcessHeight = config.Param().ConsensusParam.StakingFlowV3Height
		}
	}

	if shardBestState.CommitteeStateVersion() == committeestate.SELF_SWAP_SHARD_VERSION {
		currentCommitteePublicKeysStructs = shardBestState.GetShardCommittee()
		if beaconProcessHeight > config.Param().ConsensusParam.StakingFlowV2Height {
			beaconProcessHeight = config.Param().ConsensusParam.StakingFlowV2Height
		}
	} else {
		if beaconProcessHeight <= shardBestState.BeaconHeight {
			Logger.log.Info("Waiting For Beacon Produce Block beaconProcessHeight %+v shardBestState.BeaconHeight %+v",
				beaconProcessHeight, shardBestState.BeaconHeight)
			time.Sleep(time.Duration(shardBestState.GetCurrentTimeSlot()/5) * time.Second)
			beaconProcessHeight = getBeaconFinalHeightForProcess()
			if beaconProcessHeight <= shardBestState.BeaconHeight { //cannot receive beacon block after waiting
				return nil, errors.New("Waiting For Beacon Produce Block")
			}
		}

		currentCommitteePublicKeysStructs = committees
		committeeFinalViewBlock, _, err := blockchain.GetBeaconBlockByHash(committeeFinalViewHash)
		if err != nil {
			return nil, err
		}
		if !shardBestState.CommitteeFromBlock().IsZeroValue() {
			oldCommitteesPubKeys, _ := incognitokey.CommitteeKeyListToString(shardBestState.GetCommittee())
			currentCommitteePublicKeys, _ = incognitokey.CommitteeKeyListToString(currentCommitteePublicKeysStructs)
			temp := committeestate.DifferentElementStrings(oldCommitteesPubKeys, currentCommitteePublicKeys)
			if len(temp) != 0 {
				oldCommitteeFromBlock, _, err := blockchain.GetBeaconBlockByHash(shardBestState.CommitteeFromBlock())
				if err != nil {
					return nil, err
				}
				if oldCommitteeFromBlock.Header.Height >= committeeFinalViewBlock.Header.Height {
					return nil, NewBlockChainError(WrongBlockHeightError,
						fmt.Errorf("Height of New Shard Block's Committee From Block %+v is smaller than current Committee From Block View %+v",
							committeeFinalViewBlock.Hash(), oldCommitteeFromBlock.Hash()))
				}
				committeeFromBlockHash = committeeFinalViewHash
			} else {
				committeeFromBlockHash = shardBestState.CommitteeFromBlock()
			}
		} else {
			committeeFromBlockHash = committeeFinalViewHash
		}
	}

	if shardBestState.shardCommitteeState.Version() == committeestate.SELF_SWAP_SHARD_VERSION {
		currentCommitteePublicKeys, err = incognitokey.CommitteeKeyListToString(currentCommitteePublicKeysStructs)
		if err != nil {
			return nil, err
		}
	}

	beaconHash, err := blockchain.BeaconChain.BlockStorage.GetFinalizedBeaconBlock(beaconProcessHeight)
	if err != nil {
		Logger.log.Errorf("Beacon block %+v not found", beaconProcessHeight)
		return nil, NewBlockChainError(FetchBeaconBlockHashError, err)
	}

	blk, _, err := blockchain.BeaconChain.BlockStorage.GetBlock(*beaconHash)
	if err != nil {
		return nil, err
	}
	beaconBlock := blk.(*types.BeaconBlock)

	epoch := beaconBlock.Header.Epoch
	Logger.log.Infof("Get Beacon Block With Height %+v, Shard BestState %+v", beaconProcessHeight, shardBestState.BeaconHeight)
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain, shardBestState.BeaconHeight+1, beaconProcessHeight)
	if err != nil {
		return nil, err
	}
	if beaconProcessHeight == shardBestState.BeaconHeight {
		isOldBeaconHeight = true
	}

	// Get Transaction For new Block
	// Get Cross output coin from other shard && produce cross shard transaction
	crossTransactions := blockchain.config.BlockGen.getCrossShardData(shardBestState, beaconProcessHeight)
	Logger.log.Critical("Cross Transaction: ", crossTransactions)
	// Get Transaction for new block
	// // startStep = time.Now()
	blockCreationLeftOver := curView.BlockMaxCreateTime - time.Since(newShardBlockBeginTime)
	txsToAddFromBlock, err := blockchain.config.BlockGen.getTransactionForNewBlock(
		curView,
		&tempPrivateKey,
		shardID,
		beaconBlocks,
		blockCreationLeftOver,
		beaconProcessHeight,
	)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, txsToAddFromBlock...)
	// build txs with metadata
	transactionsForNewBlock, err = blockchain.BuildResponseTransactionFromTxsWithMetadata(shardBestState, transactionsForNewBlock, &tempPrivateKey, shardID)
	// process instruction from beacon block

	beaconInstructions, _, err := blockchain.
		extractInstructionsFromBeacon(beaconBlocks, shardBestState.ShardID)
	if err != nil {
		return nil, err
	}

	shardPendingValidatorStr, err := incognitokey.
		CommitteeKeyListToString(currentPendingValidators)
	if err != nil {
		return nil, err
	}

	if shardBestState.shardCommitteeState.Version() == committeestate.SELF_SWAP_SHARD_VERSION {
		env := committeestate.NewShardCommitteeStateEnvironmentForAssignInstruction(
			beaconInstructions,
			curView.ShardID,
			shardBestState.NumberOfFixedShardBlockValidator,
			shardBestState.ShardHeight+1,
		)

		assignInstructionProcessor := shardBestState.shardCommitteeState.(committeestate.AssignInstructionProcessor)
		addedSubstitutes := assignInstructionProcessor.ProcessAssignInstructions(env)

		currentPendingValidators, err = updateCommitteesWithAddedAndRemovedListValidator(currentPendingValidators,
			addedSubstitutes)

		shardPendingValidatorStr, err = incognitokey.CommitteeKeyListToString(currentPendingValidators)
		if err != nil {
			return nil, NewBlockChainError(ProcessInstructionFromBeaconError, err)
		}
	}

	shardInstructions, _, _, err = blockchain.generateInstruction(shardBestState, shardID,
		beaconProcessHeight, isOldBeaconHeight, beaconBlocks,
		shardPendingValidatorStr, currentCommitteePublicKeys)
	if err != nil {
		return nil, NewBlockChainError(GenerateInstructionError, err)
	}

	if len(shardInstructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", shardInstructions)
	}

	newShardBlock.BuildShardBlockBody(shardInstructions, crossTransactions, transactionsForNewBlock)
	//==========Build Essential Header Data=========
	// producer key
	producerKey := proposer
	producerPubKeyStr := proposer
	totalTxsFee := shardBestState.shardCommitteeState.BuildTotalTxsFeeFromTxs(newShardBlock.Body.Transactions)
	crossShards, err := CreateCrossShardByteArray(newShardBlock.Body.Transactions, shardID)
	if err != nil {
		return nil, err
	}

	newShardBlock.Header = types.ShardHeader{
		Producer:           producerKey, //committeeMiningKeys[producerPosition],
		ProducerPubKeyStr:  producerPubKeyStr,
		ShardID:            shardID,
		Version:            version,
		PreviousBlockHash:  shardBestState.BestBlockHash,
		Height:             shardBestState.ShardHeight + 1,
		Round:              round,
		Epoch:              epoch,
		CrossShardBitMap:   crossShards,
		BeaconHeight:       beaconProcessHeight,
		BeaconHash:         *beaconHash,
		TotalTxsFee:        totalTxsFee,
		ConsensusType:      shardBestState.ConsensusAlgorithm,
		CommitteeFromBlock: committeeFromBlockHash,
	}
	//============Update Shard BestState=============
	// startStep = time.Now()
	_, hashes, _, err := shardBestState.updateShardBestState(blockchain, newShardBlock, beaconBlocks, currentCommitteePublicKeysStructs)
	if err != nil {
		return nil, err
	}
	//============Build Header=============
	// Build Root Hash for Header
	merkleRoots := types.Merkle{}.BuildMerkleTreeStore(newShardBlock.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	crossTransactionRoot, err := CreateMerkleCrossTransaction(newShardBlock.Body.CrossTransactions)
	if err != nil {
		return nil, err
	}
	txInstructions, _, err := CreateShardInstructionsFromTransactionAndInstruction(newShardBlock.Body.Transactions, blockchain, shardID, newShardBlock.Header.Height, newShardBlock.Header.BeaconHeight, false)
	if err != nil {
		return nil, err
	}

	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range shardInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	instructionsHash, err := generateHashFromStringArray(totalInstructions)
	if err != nil {
		return nil, NewBlockChainError(InstructionsHashError, err)
	}

	// Instruction merkle root
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(shardInstructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := types.GetKeccak256MerkleRoot(insts)
	// shard tx root
	_, shardTxMerkleData := types.CreateShardTxRoot(newShardBlock.Body.Transactions)
	// Add Root Hash To Header
	newShardBlock.Header.TxRoot = *merkleRoot
	newShardBlock.Header.ShardTxRoot = shardTxMerkleData[len(shardTxMerkleData)-1]
	newShardBlock.Header.CrossTransactionRoot = *crossTransactionRoot
	newShardBlock.Header.InstructionsRoot = instructionsHash
	newShardBlock.Header.CommitteeRoot = hashes.ShardCommitteeHash
	newShardBlock.Header.PendingValidatorRoot = hashes.ShardSubstituteHash
	newShardBlock.Header.StakingTxRoot = common.Hash{}
	newShardBlock.Header.Timestamp = start

	copy(newShardBlock.Header.InstructionMerkleRoot[:], instMerkleRoot)
	return newShardBlock, nil
}

// getTransactionForNewBlock get transaction for new block
// 1. Get pending transaction from blockgen
// 2. Keep valid tx & Removed error tx
// 3. Build response Transaction For Shard
// 4. Build response Transaction For Beacon
// 5. Return valid transaction from pending, response transactions from shard and beacon
func (blockGenerator *BlockGenerator) getTransactionForNewBlock(
	curView *ShardBestState,
	privatekey *privacy.PrivateKey,
	shardID byte,
	beaconBlocks []*types.BeaconBlock,
	blockCreationLeftOver time.Duration,
	beaconHeight uint64,
) (
	[]metadata.Transaction,
	error,
) {
	var err error
	st := time.Now()
	chain := blockGenerator.chain.ShardChain[shardID]
	maxSize := uint64(4096) //kB
	var responseTxsBeacon []metadata.Transaction
	var errInstructions [][]string
	responseTxsBeacon, errInstructions, err = blockGenerator.buildResponseTxsFromBeaconInstructions(curView, beaconBlocks, privatekey, shardID)
	if err != nil {
		return nil, err
	}
	bView := &BeaconBestState{}
	bView, err = blockGenerator.chain.GetBeaconViewStateDataFromBlockHash(curView.BestBeaconHash, true, false, false)

	if err != nil {
		return nil, NewBlockChainError(CloneBeaconBestStateError, err)
	}
	blockCreationLeftOver = blockCreationLeftOver - time.Now().Sub(st)
	st = time.Now()
	txsToAdd := []metadata.Transaction{}
	totalTxsReminder := int64(config.Param().TransactionInBlockParam.Upper)
	if curView.BestBlock.GetVersion() >= types.INSTANT_FINALITY_VERSION_V2 {
		totalTxsReminder = curView.MaxTxsPerBlockRemainder - int64(len(responseTxsBeacon))
		if totalTxsReminder < 0 {
			totalTxsReminder = 0
		} else {
			if totalTxsReminder > int64(config.Param().TransactionInBlockParam.Upper) {
				totalTxsReminder = int64(config.Param().TransactionInBlockParam.Upper)
			}
		}
	}
	if !blockGenerator.chain.config.usingNewPool {
		txToRemove := []metadata.Transaction{}
		txsToAdd, txToRemove, _ = blockGenerator.getPendingTransaction(
			shardID,
			beaconBlocks,
			blockCreationLeftOver.Nanoseconds(),
			bView.BeaconHeight,
			curView,
			totalTxsReminder,
		)
		if len(txsToAdd) == 0 {
			Logger.log.Info("Creating empty block...")
		}
		go blockGenerator.txPool.RemoveTx(txToRemove, false)
	} else {
		txsToAdd = chain.TxPool.GetTxsTranferForNewBlock(
			blockGenerator.chain,
			curView,
			bView,
			maxSize,
			blockCreationLeftOver*4,
			blockCreationLeftOver,
			totalTxsReminder,
		)
	}
	if len(txsToAdd) > 0 {
		Logger.log.Infof("SHARD %v | Crawling %v txs for block %v cost %v", shardID, len(txsToAdd), curView.ShardHeight+1, time.Since(st))
	}
	txsToAdd = append(txsToAdd, responseTxsBeacon...)
	if len(errInstructions) > 0 {
		Logger.log.Error("List error instructions, which can not create tx", errInstructions)
	}
	return txsToAdd, nil
}

// buildResponseTxsFromBeaconInstructions builds response txs from beacon instructions
func (blockGenerator *BlockGenerator) buildResponseTxsFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, [][]string, error) {
	responsedTxs := []metadata.Transaction{}
	responsedHashTxs := []common.Hash{} // capture hash of responsed tx
	errorInstructions := [][]string{}   // capture error instruction -> which instruction can not create tx

	for _, beaconBlock := range beaconBlocks {
		blockHash := beaconBlock.Header.Hash()
		beaconRootHashes, err := GetBeaconRootsHashByBlockHash(
			blockGenerator.chain.GetBeaconChainDatabase(), blockHash)
		if err != nil {
			return nil, nil, err
		}
		featureStateDB, err := statedb.NewWithPrefixTrie(
			beaconRootHashes.FeatureStateDBRootHash,
			statedb.NewDatabaseAccessWarper(blockGenerator.chain.GetBeaconChainDatabase()),
		)
		if err != nil {
			return nil, nil, err
		}

		for _, inst := range beaconBlock.Body.Instructions {
			if len(inst) <= 2 {
				continue
			}
			if instruction.IsConsensusInstruction(inst[0]) {
				continue
			}
			metaType, err := strconv.Atoi(inst[0])
			if err != nil {
				continue
			}
			var newTx metadata.Transaction
			switch metaType {
			case metadata.InitTokenRequestMeta:
				if len(inst) == 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildTokenInitAcceptedTx(inst[3], producerPrivateKey, shardID, curView)
				}
			case metadata.IssuingETHRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingETHResponseMeta, false)
				}
			case metadata.IssuingBSCRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingBSCResponseMeta, false)
				}
			case metadata.IssuingRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB)
				}
			case metadata.IssuingPRVERC20RequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingPRVERC20ResponseMeta, true)
				}
			case metadata.IssuingPRVBEP20RequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingPRVBEP20ResponseMeta, true)
				}
			case metadata.IssuingPLGRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingPLGResponseMeta, false)
				}
			case metadata.IssuingFantomRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingFantomResponseMeta, false)
				}
			case metadata.IssuingAuroraRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingAuroraResponseMeta, false)
				}
			case metadata.IssuingAvaxRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingAvaxResponseMeta, false)
				}

			case metadata.IssuingNearRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildBridgeIssuanceTx(inst[3], producerPrivateKey, shardID, curView, featureStateDB, metadata.IssuingNearResponseMeta, false)
				}

			// portal
			case metadata.PortalRequestPortingMeta, metadata.PortalRequestPortingMetaV3:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestRejectedChainStatus {
					newTx, err = curView.buildPortalRefundPortingFeeTx(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalCustodianDepositMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestRefundChainStatus {
					newTx, err = curView.buildPortalRefundCustodianDepositTx(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalUserRequestPTokenMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestAcceptedChainStatus {
					newTx, err = curView.buildPortalAcceptedRequestPTokensTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
				//custodian withdraw
			case metadata.PortalCustodianWithdrawRequestMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestAcceptedChainStatus {
					newTx, err = curView.buildPortalCustodianWithdrawRequest(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalRedeemRequestMeta, metadata.PortalRedeemRequestMetaV3:
				if len(inst) >= 4 && (inst[2] == portalcommonv3.PortalRequestRejectedChainStatus || inst[2] == portalcommonv3.PortalRedeemReqCancelledByLiquidationChainStatus) {
					newTx, err = curView.buildPortalRejectedRedeemRequestTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
				//liquidation: redeem ptoken
			case metadata.PortalRedeemFromLiquidationPoolMeta:
				if len(inst) >= 4 {
					if inst[2] == portalcommonv3.PortalProducerInstSuccessChainStatus {
						newTx, err = curView.buildPortalRedeemLiquidateExchangeRatesRequestTx(inst[3], producerPrivateKey, shardID)
					} else if inst[2] == portalcommonv3.PortalRequestRejectedChainStatus {
						newTx, err = curView.buildPortalRefundRedeemLiquidateExchangeRatesTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
					}
				}
			case metadata.PortalLiquidateCustodianMeta, metadata.PortalLiquidateCustodianMetaV3:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalProducerInstSuccessChainStatus {
					newTx, err = curView.buildPortalLiquidateCustodianResponseTx(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalRequestWithdrawRewardMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestAcceptedChainStatus {
					newTx, err = curView.buildPortalAcceptedWithdrawRewardTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
				//liquidation: custodian deposit
			case metadata.PortalCustodianTopupMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestRejectedChainStatus {
					newTx, err = curView.buildPortalLiquidationCustodianDepositReject(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalCustodianTopupMetaV2:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestRejectedChainStatus {
					newTx, err = curView.buildPortalLiquidationCustodianDepositRejectV2(inst[3], producerPrivateKey, shardID)
				}
			//
			case metadata.PortalTopUpWaitingPortingRequestMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv3.PortalRequestRejectedChainStatus {
					newTx, err = curView.buildPortalRejectedTopUpWaitingPortingTx(inst[3], producerPrivateKey, shardID)
				}
			//redeem from liquidation pool
			case metadata.PortalRedeemFromLiquidationPoolMetaV3:
				if len(inst) >= 4 {
					if inst[2] == portalcommonv3.PortalProducerInstSuccessChainStatus {
						newTx, err = curView.buildPortalRedeemLiquidateExchangeRatesRequestTxV3(inst[3], producerPrivateKey, shardID)
					} else if inst[2] == portalcommonv3.PortalRequestRejectedChainStatus {
						newTx, err = curView.buildPortalRefundRedeemLiquidateExchangeRatesTxV3(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
					}
				}
			// portal v4
			case metadataCommon.PortalV4ShieldingRequestMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv4.PortalV4RequestAcceptedChainStatus {
					newTx, err = curView.buildPortalAcceptedShieldingRequestTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
			case metadataCommon.PortalV4UnshieldingRequestMeta:
				if len(inst) >= 4 && inst[2] == portalcommonv4.PortalV4RequestRefundedChainStatus {
					newTx, err = curView.buildPortalRefundedUnshieldingRequestTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}

			default:
				if metadataCommon.IsPDEType(metaType) {
					pdeTxBuilderV1 := pdex.TxBuilderV1{}
					newTx, err = pdeTxBuilderV1.Build(
						metaType,
						inst,
						producerPrivateKey,
						shardID,
						curView.GetCopiedTransactionStateDB(),
					)
				} else if metadataCommon.IsPdexv3Type(metaType) {
					pdeTxBuilderV2 := pdex.TxBuilderV2{}
					newTx, err = pdeTxBuilderV2.Build(
						metaType,
						inst,
						producerPrivateKey,
						shardID,
						curView.GetCopiedTransactionStateDB(),
						beaconBlock.Header.Height,
					)
				} else if metadataBridge.IsBridgeAggMetaType(metaType) {
					newTx, err = bridgeagg.TxBuilder{}.Build(
						metaType, inst, producerPrivateKey, shardID, curView.GetCopiedTransactionStateDB(),
					)
				}
			}
			if err != nil {
				return nil, nil, err
			}
			if newTx != nil {
				newTxHash := *newTx.Hash()
				if ok, _ := common.SliceExists(responsedHashTxs, newTxHash); ok {
					data, _ := json.Marshal(newTx)
					Logger.log.Error("Double tx from instruction", inst, string(data))
					errorInstructions = append(errorInstructions, inst)
					//continue
				}
				responsedTxs = append(responsedTxs, newTx)
				responsedHashTxs = append(responsedHashTxs, newTxHash)
			}
		}
	}
	returnStakingTxs, errIns, err := blockGenerator.chain.buildReturnStakingTxFromBeaconInstructions(
		curView,
		beaconBlocks,
		producerPrivateKey,
		shardID,
	)
	if err != nil {
		return nil, nil, err
	}
	responsedTxs = append(responsedTxs, returnStakingTxs...)
	errorInstructions = append(errorInstructions, errIns...)
	return responsedTxs, errorInstructions, nil
}

// generateInstruction create instruction for new shard block
// Swap: at the end of beacon epoch
// Brigde: at the end of beacon epoch
// Return params:
// #1: instruction list
// #2: shardpendingvalidator
// #3: shardcommittee
// #4: error
func (blockchain *BlockChain) generateInstruction(view *ShardBestState,
	shardID byte, beaconHeight uint64,
	isOldBeaconHeight bool, beaconBlocks []*types.BeaconBlock,
	shardPendingValidators []string, shardCommittees []string) ([][]string, []string, []string, error) {
	var (
		instructions                      = [][]string{}
		bridgeSwapConfirmInst             = []string{}
		swapOrConfirmShardSwapInstruction = []string{}
		err                               error
	)
	// if this beacon height has been seen already then DO NOT generate any more instruction
	if view.shardCommitteeState.Version() == committeestate.SELF_SWAP_SHARD_VERSION {
		if blockchain.IsLastBeaconHeightInEpoch(beaconHeight) && isOldBeaconHeight == false {
			Logger.log.Info("ShardPendingValidator", shardPendingValidators)
			Logger.log.Info("ShardCommittee", shardCommittees)
			Logger.log.Info("MaxShardCommitteeSize", view.MaxShardCommitteeSize)
			Logger.log.Info("ShardID", shardID)

			numberOfFixedShardBlockValidators := view.NumberOfFixedShardBlockValidator

			maxShardCommitteeSize := view.MaxShardCommitteeSize - numberOfFixedShardBlockValidators
			var minShardCommitteeSize int
			if view.MinShardCommitteeSize-numberOfFixedShardBlockValidators < 0 {
				minShardCommitteeSize = 0
			} else {
				minShardCommitteeSize = view.MinShardCommitteeSize - numberOfFixedShardBlockValidators
			}
			epoch := blockchain.GetEpochByHeight(beaconHeight)
			if common.IndexOfUint64(epoch, config.Param().ConsensusParam.EpochBreakPointSwapNewKey) > -1 {
				swapOrConfirmShardSwapInstruction, shardCommittees = createShardSwapActionForKeyListV2(
					shardCommittees,
					numberOfFixedShardBlockValidators,
					config.Param().ActiveShards,
					shardID,
					epoch,
				)
			} else {
				tempSwapInstruction := instruction.NewSwapInstruction()
				env := committeestate.NewShardCommitteeStateEnvironmentForSwapInstruction(
					view.ShardHeight,
					shardID,
					maxShardCommitteeSize,
					minShardCommitteeSize,
					config.Param().SwapCommitteeParam.Offset,
					config.Param().SwapCommitteeParam.SwapOffset,
					numberOfFixedShardBlockValidators,
				)
				swapInstructionGenerator := view.shardCommitteeState.(committeestate.SwapInstructionGenerator)
				tempSwapInstruction, shardPendingValidators, shardCommittees, err =
					swapInstructionGenerator.GenerateSwapInstructions(env)
				if err != nil {
					Logger.log.Error(err)
					return instructions, shardPendingValidators, shardCommittees, err
				}
				if !tempSwapInstruction.IsEmpty() {
					swapOrConfirmShardSwapInstruction = tempSwapInstruction.ToString()
				}
			}
		}
	}

	if len(swapOrConfirmShardSwapInstruction) > 0 {
		instructions = append(instructions, swapOrConfirmShardSwapInstruction)
	}

	if len(bridgeSwapConfirmInst) > 0 {
		instructions = append(instructions, bridgeSwapConfirmInst)
		Logger.log.Infof("Build bridge swap confirm inst: %s \n", bridgeSwapConfirmInst)
	}
	// Pick BurningConfirm inst and save to bridge block
	bridgeID := byte(common.BridgeShardID)
	if shardID == bridgeID { // Pick burning confirm inst for V1
		prevBlock := view.BestBlock
		height := view.ShardHeight + 1
		confirmInsts := pickBurningConfirmInstructionV1(beaconBlocks, height)
		if len(confirmInsts) > 0 {
			bid := []uint64{}
			for _, b := range beaconBlocks {
				bid = append(bid, b.Header.Height)
			}
			Logger.log.Infof("Picked burning confirm inst: %s %d %v\n", confirmInsts, prevBlock.Header.Height+1, bid)
			instructions = append(instructions, confirmInsts...)
		}
	}

	return instructions, shardPendingValidators, shardCommittees, nil
}

// getCrossShardData get cross shard data from cross shard block
//  1. Get Cross Shard Block and Validate
//     a. Get Valid Cross Shard Block from Cross Shard Pool
//     b. Get Current Cross Shard State: Last Cross Shard Block From other Shard (FS) to this shard (TS) (Ex: last cross shard block from Shard 0 to Shard 1)
//     c. Get Next Cross Shard Block Height from other Shard (FS) to this shard (TS)
//     + Using FetchCrossShardNextHeight function in Database to determine next block height
//     d. Fetch Other Shard (FS) Committee at Next Cross Shard Block Height for Validation
//  2. Validate
//     a. Get Next Cross Shard Height from Database
//     b. Cross Shard Block Height is Next Cross Shard Height from Database (if miss Cross Shard Block according to beacon bytemap then stop discard the rest)
//     c. Verify Cross Shard Block Signature
//  3. After validation:
//     - Process valid block to extract:
//     + Cross output coin
//     + Cross Normal Token
func (blockGenerator *BlockGenerator) getCrossShardData(curView *ShardBestState, currentBeaconHeight uint64) map[byte][]types.CrossTransaction {
	crossTransactions := make(map[byte][]types.CrossTransaction)
	// get cross shard block
	toShard := curView.ShardID
	var allCrossShardBlock = make([][]*types.CrossShardBlock, config.Param().ActiveShards)
	for sid, v := range blockGenerator.syncker.GetCrossShardBlocksForShardProducer(curView, nil) {
		heightList := make([]uint64, len(v))
		for i, b := range v {
			allCrossShardBlock[sid] = append(allCrossShardBlock[sid], b.(*types.CrossShardBlock))
			heightList[i] = b.(*types.CrossShardBlock).GetHeight()
		}
		Logger.log.Infof("Shard %v, GetCrossShardBlocksForShardProducer from shard %v: %v", toShard, sid, heightList)
	}

	// allCrossShardBlock => already short
	for _, crossShardBlock := range allCrossShardBlock {
		for _, blk := range crossShardBlock {
			crossTransaction := types.CrossTransaction{
				OutputCoin:       blk.CrossOutputCoin,
				TokenPrivacyData: blk.CrossTxTokenPrivacyData,
				BlockHash:        *blk.Hash(),
				BlockHeight:      blk.Header.Height,
			}
			crossTransactions[blk.Header.ShardID] = append(crossTransactions[blk.Header.ShardID], crossTransaction)
			fmt.Println("Check cross shard data", crossTransaction.OutputCoin)
		}
	}
	return crossTransactions
}

/*
Verify Transaction with these condition: defined in mempool.go
*/
func (blockGenerator *BlockGenerator) getPendingTransaction(
	shardID byte,
	beaconBlocks []*types.BeaconBlock,
	blockCreationTimeLeftOver int64,
	beaconHeight uint64,
	curView *ShardBestState,
	maxTxs int64,
) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	spareTime := SpareTime * time.Millisecond
	maxBlockCreationTimeLeftTime := blockCreationTimeLeftOver - spareTime.Nanoseconds()
	startTime := time.Now()
	sourceTxns := blockGenerator.GetPendingTxsV2(shardID)
	var elasped int64
	Logger.log.Infof("Number of transaction get from Block Generator: %v; Maximum txs in this block %v", len(sourceTxns), maxTxs)
	isEmpty := blockGenerator.chain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		return []metadata.Transaction{}, []metadata.Transaction{}, 0
	}
	currentSize := uint64(0)
	preparedTxForNewBlock := []metadata.Transaction{}
	for _, tx := range sourceTxns {
		tempSize := tx.GetTxActualSize()
		if currentSize+tempSize >= common.MaxBlockSize {
			break
		}
		preparedTxForNewBlock = append(preparedTxForNewBlock, tx)
		elasped = time.Since(startTime).Nanoseconds()
		if elasped >= maxBlockCreationTimeLeftTime {
			Logger.log.Info("Shard Producer/Elapsed, Break: ", elasped)
			break
		}
	}
	listBatchTxs := []metadata.Transaction{}
	for index, tx := range preparedTxForNewBlock {
		elasped = time.Since(startTime).Nanoseconds()
		if elasped >= maxBlockCreationTimeLeftTime {
			Logger.log.Info("Shard Producer/Elapsed, Break: ", elasped)
			break
		}
		listBatchTxs = append(listBatchTxs, tx)
		if ((index+1)%TransactionBatchSize == 0) || (index == len(preparedTxForNewBlock)-1) {
			tempTxDesc, err := blockGenerator.chain.config.TempTxPool.MaybeAcceptBatchTransactionForBlockProducing(shardID, listBatchTxs, int64(beaconHeight), curView)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Verify Batch Transaction for new block error %+v", shardID, err)
				for _, tx2 := range listBatchTxs {
					if blockGenerator.chain.config.TempTxPool.HaveTransaction(tx2.Hash()) {
						continue
					}
					txShardID := common.GetShardIDFromLastByte(tx2.GetSenderAddrLastByte())
					if txShardID != shardID {
						continue
					}
					tempTxDesc, err := blockGenerator.chain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx2, int64(beaconHeight), curView)
					if err != nil {
						txToRemove = append(txToRemove, tx2)
						continue
					}
					tempTx := tempTxDesc.Tx
					tempSize := tempTx.GetTxActualSize()
					if currentSize+tempSize >= common.MaxBlockSize {
						break
					}
					if len(txsToAdd)+1 > int(maxTxs) {
						break
					}
					totalFee += tempTx.GetTxFee()
					currentSize += tempSize
					txsToAdd = append(txsToAdd, tempTx)
				}
			}
			for _, tempToAddTxDesc := range tempTxDesc {
				tempTx := tempToAddTxDesc.Tx
				totalFee += tempTx.GetTxFee()
				tempSize := tempTx.GetTxActualSize()
				if currentSize+tempSize >= common.MaxBlockSize {
					break
				}
				if len(txsToAdd)+1 > int(maxTxs) {
					break
				}
				currentSize += tempSize
				txsToAdd = append(txsToAdd, tempTx)
			}
			// reset list batch and add new txs
			listBatchTxs = []metadata.Transaction{}

		} else {
			continue
		}
	}
	Logger.log.Criticalf(" 🔎 %+v transactions for New Block from pool \n", len(txsToAdd))
	blockGenerator.chain.config.TempTxPool.EmptyPool()
	return txsToAdd, txToRemove, totalFee
}

func (blockGenerator *BlockGenerator) createTempKeyset() privacy.PrivateKey {
	b := common.RandBytes(common.HashSize)
	return privacy.GeneratePrivateKey(b)
}

func createShardSwapActionForKeyListV2(
	shardCommittees []string,
	minCommitteeSize int,
	activeShard int,
	shardID byte,
	epoch uint64,
) ([]string, []string) {
	swapInstruction, newShardCommittees := GetShardSwapInstructionKeyListV2(epoch, minCommitteeSize, activeShard)
	remainShardCommittees := shardCommittees[minCommitteeSize:]
	return swapInstruction[shardID], append(newShardCommittees[shardID], remainShardCommittees...)
}

// extractInstructionsFromBeacon : preprcess for beacon instructions before move to handle it in committee state
// Store stakingtx address and return it back to outside
// Only process for instruction not stake instruction
func (blockchain *BlockChain) extractInstructionsFromBeacon(
	beaconBlocks []*types.BeaconBlock,
	shardID byte) ([][]string, map[string]string, error) {

	instructions := [][]string{}
	stakingTx := make(map[string]string)
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// Get Staking Tx
			// assume that stake instruction already been validated by beacon committee

			switch l[0] {
			case instruction.STAKE_ACTION:
				if l[2] == "shard" {
					shard := strings.Split(l[1], ",")
					newShardCandidates := []string{}
					newShardCandidates = append(newShardCandidates, shard...)
					if len(l) == 6 {
						for i, v := range strings.Split(l[3], ",") {
							txHash, err := common.Hash{}.NewHashFromStr(v)
							if err != nil {
								continue
							}
							_, _, _, err = blockchain.GetTransactionByHashWithShardID(*txHash, shardID)
							if err != nil {
								continue
							}
							// if transaction belong to this shard then add to shard beststate
							stakingTx[newShardCandidates[i]] = v
						}
						instructions = append(instructions, l)
					}
				}
				if l[2] == "beacon" {
					beacon := strings.Split(l[1], ",")
					newBeaconCandidates := []string{}
					newBeaconCandidates = append(newBeaconCandidates, beacon...)
					if len(l) == 6 {
						for i, v := range strings.Split(l[3], ",") {
							txHash, err := common.Hash{}.NewHashFromStr(v)
							if err != nil {
								continue
							}
							_, _, _, err = blockchain.GetTransactionByHashWithShardID(*txHash, shardID)
							if err != nil {
								continue
							}
							// if transaction belong to this shard then add to shard beststate
							stakingTx[newBeaconCandidates[i]] = v
						}
						// instructions = append(instructions, l)
					}
				}

			case instruction.SWAP_SHARD_ACTION:
				//Only process swap shard action for that shard
				swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(l)
				if err != nil {
					Logger.log.Errorf("Fail to ValidateAndImportSwapShardInstructionFromString %v", err)
					continue
				}
				if byte(swapShardInstruction.ChainID) != shardID {
					continue
				}
				instructions = append(instructions, l)
			case instruction.ASSIGN_ACTION:
				//Only process swap shard action for that shard
				assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(l)
				if err != nil {
					Logger.log.Errorf("Fail to ValidateAndImportSwapShardInstructionFromString %v", err)
					continue
				}
				if byte(assignInstruction.ChainID) != shardID {
					continue
				}
				instructions = append(instructions, l)
			default:
				instructions = append(instructions, l)
				continue
			}
		}
	}

	return instructions, stakingTx, nil
}

// CreateShardInstructionsFromTransactionAndInstruction create inst from transactions in shard block
func CreateShardInstructionsFromTransactionAndInstruction(
	transactions []metadata.Transaction,
	bc *BlockChain, shardID byte,
	shardHeight, beaconHeight uint64,
	shouldCollectPdexTxs bool,
) (instructions [][]string, pdexTxs map[uint][]metadata.Transaction, err error) {
	// Generate stake action
	stakeShardPublicKey := []string{}
	stakeShardTxID := []string{}
	stakeShardRewardReceiver := []string{}
	stakeShardAutoStaking := []string{}
	stopAutoStaking := []string{}
	unstaking := []string{}

	addStake_cpk := []string{}
	addStake_amount := []uint64{}
	addStake_tx := []string{}

	if shouldCollectPdexTxs {
		pdexTxs = make(map[uint][]metadata.Transaction)
	}

	for _, tx := range transactions {
		metadataValue := tx.GetMetadata()
		if metadataValue != nil {
			if beaconHeight >= config.Param().PDexParams.Pdexv3BreakPointHeight && metadata.IsPdexv3Tx(metadataValue) {
				if shouldCollectPdexTxs {
					pdexTxs[pdex.AmplifierVersion] = append(pdexTxs[pdex.AmplifierVersion], tx)
				}
			} else {
				actionPairs, err := metadataValue.BuildReqActions(tx, bc, nil, bc.BeaconChain.GetFinalView().(*BeaconBestState), shardID, shardHeight)
				Logger.log.Infof("Build Request Action Pairs %+v, metadata value %+v", actionPairs, metadataValue)
				if err != nil {
					Logger.log.Errorf("Build Request Action Error %+v", err)
					return nil, nil, fmt.Errorf("Build Request Action Error %+v", err)
				}
				if shouldCollectPdexTxs {
					if metadata.IsPDETx(metadataValue) {
						pdexTxs[pdex.BasicVersion] = append(pdexTxs[pdex.BasicVersion], tx)
					}
				}
				instructions = append(instructions, actionPairs...)
			}
		}
		switch tx.GetMetadataType() {
		case metadata.ShardStakingMeta:
			stakingMetadata, ok := tx.GetMetadata().(*metadata.StakingMetadata)
			if !ok {
				return nil, nil, fmt.Errorf("Expect metadata type to be *metadata.StakingMetadata but get %+v", reflect.TypeOf(tx.GetMetadata()))
			}
			stakeShardPublicKey = append(stakeShardPublicKey, stakingMetadata.CommitteePublicKey)
			stakeShardTxID = append(stakeShardTxID, tx.Hash().String())
			stakeShardRewardReceiver = append(stakeShardRewardReceiver, stakingMetadata.RewardReceiverPaymentAddress)
			if len(stakingMetadata.CommitteePublicKey) == 0 {
				continue
			}
			if stakingMetadata.AutoReStaking {
				stakeShardAutoStaking = append(stakeShardAutoStaking, "true")
			} else {
				stakeShardAutoStaking = append(stakeShardAutoStaking, "false")
			}
		case metadata.BeaconStakingMeta:
			stakingMetadata, ok := tx.GetMetadata().(*metadata.StakingMetadata)
			if !ok {
				return nil, nil, fmt.Errorf("Expect metadata type to be *metadata.StakingMetadata but get %+v", reflect.TypeOf(tx.GetMetadata()))
			}
			inst := []string{
				instruction.STAKE_ACTION,
				stakingMetadata.CommitteePublicKey,
				instruction.BEACON_INST, "-1",
				stakingMetadata.RewardReceiverPaymentAddress,
				"true",
			}
			instructions = append(instructions, inst)
		case metadata.StopAutoStakingMeta:
			stopAutoStakingMetadata, ok := tx.GetMetadata().(*metadata.StopAutoStakingMetadata)
			if !ok {
				return nil, nil, fmt.Errorf("Expect metadata type to be *metadata.StopAutoStakingMetadata but get %+v", reflect.TypeOf(tx.GetMetadata()))
			}
			if len(stopAutoStakingMetadata.CommitteePublicKey) != 0 {
				stopAutoStaking = append(stopAutoStaking, stopAutoStakingMetadata.CommitteePublicKey)
			}
		case metadata.UnStakingMeta:
			unstakingMetadata, ok := tx.GetMetadata().(*metadata.UnStakingMetadata)
			if !ok {
				return nil, nil, fmt.Errorf("Expect metadata type to be *metadata.UnstakingMetadata but get %+v", reflect.TypeOf(tx.GetMetadata()))
			}
			if len(unstakingMetadata.CommitteePublicKey) != 0 {
				unstaking = append(unstaking, unstakingMetadata.CommitteePublicKey)
			}
		case metadata.AddStakingMeta:
			stakingMetadata, ok := tx.GetMetadata().(*metadata.AddStakingMetadata)
			if !ok {
				return nil, nil, fmt.Errorf("Expect metadata type to be *metadata.StakingMetadata but get %+v", reflect.TypeOf(tx.GetMetadata()))
			}
			addStake_cpk = append(addStake_cpk, stakingMetadata.CommitteePublicKey)
			addStake_amount = append(addStake_amount, stakingMetadata.AddStakingAmount)
			addStake_tx = append(addStake_tx, stakingMetadata.Hash().String())
		}
	}

	if !reflect.DeepEqual(stakeShardPublicKey, []string{}) {
		if len(stakeShardPublicKey) != len(stakeShardTxID) &&
			len(stakeShardTxID) != len(stakeShardRewardReceiver) &&
			len(stakeShardRewardReceiver) != len(stakeShardAutoStaking) {
			return nil, nil, NewBlockChainError(StakeInstructionError,
				fmt.Errorf("Expect public key list (length %+v) and reward receiver list (length %+v), auto restaking (length %+v) to be equal", len(stakeShardPublicKey), len(stakeShardRewardReceiver), len(stakeShardAutoStaking)))
		}
		stakeShardPublicKey, err = incognitokey.ConvertToBase58ShortFormat(stakeShardPublicKey)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed To Convert Stake Shard Public Key to Base58 Short Form")
		}
		// ["stake", "pubkey1,pubkey2,..." "shard" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." "flag1,flag2,..."]
		inst := []string{
			instruction.STAKE_ACTION,
			strings.Join(stakeShardPublicKey, ","),
			instruction.SHARD_INST, strings.Join(stakeShardTxID, ","),
			strings.Join(stakeShardRewardReceiver, ","),
			strings.Join(stakeShardAutoStaking, ","),
		}
		instructions = append(instructions, inst)
	}

	if !reflect.DeepEqual(stopAutoStaking, []string{}) {
		// ["stopautostaking" "pubkey1,pubkey2,..."]
		validStopAutoStaking := []string{}
		for _, v := range stopAutoStaking {
			validStopAutoStaking = append(validStopAutoStaking, v)
		}
		inst := []string{instruction.STOP_AUTO_STAKE_ACTION, strings.Join(validStopAutoStaking, ",")}
		instructions = append(instructions, inst)
	}

	if !reflect.DeepEqual(unstaking, []string{}) {
		// ["unstake" "pubkey1,pubkey2,..."]
		inst := []string{instruction.UNSTAKE_ACTION, strings.Join(unstaking, ",")}
		instructions = append(instructions, inst)
	}

	if len(addStake_cpk) > 0 {
		inst := instruction.NewAddStakingInstructionWithValue(addStake_cpk, addStake_amount, addStake_tx)
		instructions = append(instructions, inst.ToString())
	}
	return instructions, pdexTxs, nil
}

// CreateShardBridgeUnshieldActionsFromTxs create bridge unshield insts from transactions in shard block
func CreateShardBridgeUnshieldActionsFromTxs(
	transactions []metadata.Transaction,
	bc *BlockChain, shardID byte,
	shardHeight, beaconHeight uint64,
) ([][]string, error) {
	bridgeActions := [][]string{}
	for _, tx := range transactions {
		metadataValue := tx.GetMetadata()
		if metadataValue == nil {
			continue
		}
		if metadataCommon.IsBridgeUnshieldMetaType(tx.GetMetadataType()) {
			actionPairs, err := metadataValue.BuildReqActions(tx, bc, nil, bc.BeaconChain.GetFinalView().(*BeaconBestState), shardID, shardHeight)
			Logger.log.Infof("Build Shard Bridge Unshield instruction %+v, metadata value %+v", actionPairs, metadataValue)
			if err != nil {
				Logger.log.Errorf("Build Shard Bridge Unshield Error %+v", err)
				return nil, fmt.Errorf("Build Shard Bridge Unshield Error %+v", err)
			}
			bridgeActions = append(bridgeActions, actionPairs...)
		}
	}
	return bridgeActions, nil
}

// CreateShardBridgeAggUnshieldActionsFromTxs create bridge agg unshield insts from transactions in shard block
func CreateShardBridgeAggUnshieldActionsFromTxs(
	transactions []metadata.Transaction,
	bc *BlockChain, shardID byte,
	shardHeight, beaconHeight uint64,
) ([][]string, error) {
	bridgeAggActions := [][]string{}
	for _, tx := range transactions {
		metadataValue := tx.GetMetadata()
		if metadataValue == nil {
			continue
		}
		if metadataCommon.IsBridgeAggUnshieldMetaType(tx.GetMetadataType()) {
			actionPairs, err := metadataValue.BuildReqActions(tx, bc, nil, bc.BeaconChain.GetFinalView().(*BeaconBestState), shardID, shardHeight)
			Logger.log.Infof("Build Shard Bridge Agg Unshield instruction %+v, metadata value %+v", actionPairs, metadataValue)
			if err != nil {
				Logger.log.Errorf("Build Shard Bridge Agg Unshield Error %+v", err)
				return nil, fmt.Errorf("Build Shard Bridge Agg Unshield Error %+v", err)
			}
			bridgeAggActions = append(bridgeAggActions, actionPairs...)
		}
	}
	return bridgeAggActions, nil
}
