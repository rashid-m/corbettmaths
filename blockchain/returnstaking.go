package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	"strings"

	"github.com/incognitochain/incognito-chain/incognitokey"

	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/instruction"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

type returnStakingInfo struct {
	SwapoutPubKey string
	FunderAddress privacy.PaymentAddress
	SharedRandom  []byte
	StakingTx     metadata.Transaction
	StakingAmount uint64
}

func (blockchain *BlockChain) buildReturnStakingTxFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	[]metadata.Transaction,
	[][]string,
	error,
) {
	responsedTxs := []metadata.Transaction{}
	responsedHashTxs := map[common.Hash]struct{}{} // capture hash of responsed tx
	errorInstructions := [][]string{}
	var err error
	mReturnStakingInfo, errIns, err := blockchain.getReturnStakingInfoFromBeaconInstructions(
		curView,
		beaconBlocks,
		shardID,
	)
	errorInstructions = append(errorInstructions, errIns...)
	if err != nil {
		return nil, nil, err
	}
	for txStakingHash, returnInfo := range mReturnStakingInfo {
		txReturn, returnAmount, err := blockchain.buildReturnStakingAmountTx(
			curView,
			&returnInfo,
			producerPrivateKey,
			shardID,
		)
		Logger.log.Debugf("Return Staking Amount %v for funder %+v of candidate %v, staking transaction hash %+v, shardID %+v, err: %v", returnAmount, returnInfo.FunderAddress.String(), returnInfo.SwapoutPubKey, txStakingHash.String(), shardID, err)
		if err != nil {
			return nil, nil, err
		}
		txReturnHash := *txReturn.Hash()
		if _, ok := responsedHashTxs[txReturnHash]; ok {
			err = errors.Errorf("Double tx return staking from instruction for tx staking %v, swapout pubkey %v", txStakingHash, returnInfo.SwapoutPubKey)
			return nil, nil, err
		}
		responsedHashTxs[txReturnHash] = struct{}{}
		responsedTxs = append(responsedTxs, txReturn)
	}
	return responsedTxs, errorInstructions, nil
}

func (blockchain *BlockChain) ValidateReturnStakingTxFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
	shardBlock *types.ShardBlock,
	shardID byte,
) error {
	if shardID == 1 && shardBlock.GetHeight() == 432620 {
		return nil
	}
	if shardID == 0 && shardBlock.GetHeight() == 502419 {
		return nil
	}
	mReturnStakingInfoGot := map[common.Hash]returnStakingInfo{}
	returnStakingTxs := map[common.Hash]struct{}{}
	for _, tx := range shardBlock.Body.Transactions {
		if tx.GetType() == common.TxReturnStakingType {
			txHash := tx.Hash()
			returnMeta, ok := tx.GetMetadata().(*metadata.ReturnStakingMetadata)
			if !ok {
				return errors.Errorf("Can not parse metadata of tx %v to ReturnStaking Metadata", tx.Hash().String())
			}
			if _, ok := returnStakingTxs[*txHash]; ok {
				return errors.Errorf("Double tx return staking from instruction for tx staking %v", returnMeta.TxID)
			}
			returnStakingTxs[*txHash] = struct{}{}
			isMinted, mintCoin, coinID, err := tx.GetTxMintData()
			if err != nil || !isMinted {
				return errors.Errorf("this is not tx mint for return staking. Error %v", err)
			}
			if ok := mintCoin.CheckCoinValid(returnMeta.StakerAddress, returnMeta.SharedRandom, config.Param().StakingAmountShard); !ok {
				return errors.Errorf("mint data is invalid: Address %v; Amount %v", returnMeta.StakerAddress, mintCoin.GetValue())
			}
			if coinID.String() != common.PRVIDStr {
				return errors.Errorf("return staking tx only mints prv. Error token %v", coinID.String())
			}
			h, err := common.Hash{}.NewHashFromStr(returnMeta.TxID)
			if err != nil {
				Logger.log.Errorf("returnStaking hash %v error: %v\n", returnMeta.TxID)
				return err
			}
			mReturnStakingInfoGot[*h] = returnStakingInfo{
				FunderAddress: returnMeta.StakerAddress,
				StakingAmount: mintCoin.GetValue(),
			}
		}
	}

	mReturnStakingInfoWanted, _, err := blockchain.getReturnStakingInfoFromBeaconInstructions(
		curView,
		beaconBlocks,
		shardID,
	)
	if err != nil {
		return err
	}

	if len(mReturnStakingInfoGot) != len(mReturnStakingInfoWanted) {
		return errors.Errorf("List return staking tx of producer (len %v) and validator (len %v) not match", len(mReturnStakingInfoGot), len(mReturnStakingInfoWanted))
	}
	for txStakingHash, returnInfoWanted := range mReturnStakingInfoWanted {
		if returnInfoGot, ok := mReturnStakingInfoGot[txStakingHash]; ok {
			if (returnInfoGot.FunderAddress.String() != returnInfoWanted.FunderAddress.String()) || (returnInfoGot.StakingAmount != returnInfoWanted.StakingAmount) {
				keyWL := wallet.KeyWallet{}
				keyWL.KeySet.PaymentAddress = returnInfoGot.FunderAddress
				payment := keyWL.Base58CheckSerialize(wallet.PaymentAddressType)
				return errors.Errorf("Validator want to return for funder %v using tx staking %v with amount %v but producer return for %v with amount %v", payment, returnInfoWanted.StakingTx.Hash().String(), returnInfoWanted.StakingAmount, returnInfoGot.FunderAddress.String(), returnInfoGot.StakingAmount)
			}
			continue
		}
		return errors.Errorf("Validator want to return for funder %v using tx staking %v but producer dont do it", returnInfoWanted.FunderAddress.String(), returnInfoWanted.StakingTx.Hash().String())
	}
	return nil
}

func (blockchain *BlockChain) buildReturnStakingAmountTx(
	curView *ShardBestState,
	info *returnStakingInfo,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) (
	metadata.Transaction,
	uint64,
	error,
) {
	txStakingHash := info.StakingTx.Hash().String()
	returnStakingMeta := metadata.NewReturnStaking(
		txStakingHash,
		info.FunderAddress,
		metadata.ReturnStakingMeta,
	)

	txParam := transaction.TxSalaryOutputParams{
		Amount:          info.StakingAmount,
		ReceiverAddress: &info.FunderAddress,
		TokenID:         &common.PRVCoinID,
		Type:            common.TxReturnStakingType,
	}

	makeMD := func(c privacy.Coin) metadata.Metadata {
		if c != nil && c.GetSharedRandom() != nil {
			returnStakingMeta.SetSharedRandom(c.GetSharedRandom().ToBytesS())
		}
		return returnStakingMeta
	}
	returnStakingTx, err := txParam.BuildTxSalary(producerPrivateKey, curView.GetCopiedTransactionStateDB(), makeMD)
	if err != nil {
		return nil, 0, errors.Errorf("cannot init return staking tx. Error %v", err)
	}
	// returnStakingTx.SetType()
	return returnStakingTx, info.StakingAmount, nil
}

func (blockchain *BlockChain) getReturnStakingInfoFromBeaconInstructions(
	curView *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
	shardID byte,
) (
	map[common.Hash]returnStakingInfo,
	[][]string,
	error,
) {
	res := map[common.Hash]returnStakingInfo{}
	beaconConsensusStateDB := &statedb.StateDB{}
	beaconConsensusRootHash := common.Hash{}
	errorInstructions := [][]string{}
	beaconView := blockchain.BeaconChain.GetFinalView().(*BeaconBestState)
	var err error
	for _, beaconBlock := range beaconBlocks {
		beaconConsensusStateDB = nil
		for _, l := range beaconBlock.Body.Instructions {
			switch l[0] {
			case instruction.SWAP_ACTION:
				if beaconConsensusStateDB == nil {
					beaconConsensusRootHash, err = blockchain.GetBeaconConsensusRootHash(beaconView, beaconBlock.GetHeight()-1)
					if err != nil {
						return nil, nil, NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found, error %+v", beaconBlock.GetHeight(), err))
					}
					beaconConsensusStateDB, err = statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
				}
				for _, outPublicKey := range strings.Split(l[2], ",") {
					if len(outPublicKey) == 0 {
						continue
					}
					key := incognitokey.CommitteePublicKey{}
					err := key.FromBase58(outPublicKey)
					if err != nil {
						return nil, nil, NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Cannot parse outpubickey %v", outPublicKey))
					}
					_, has, err := statedb.IsInShardCandidateForNextEpoch(beaconConsensusStateDB, key)
					if has { //still in committee process (next epoch)
						continue
					}
					_, has, err = statedb.IsInShardCandidateForCurrentEpoch(beaconConsensusStateDB, key)
					if has { //still in committee process (current epoch: swap and random at same time -> validator will go into this queue)
						continue
					}

					//dont have shard candidate for next epoch => kickout => return staking amount
					stakerInfo, has, err := statedb.GetStakerInfo(beaconConsensusStateDB, outPublicKey)
					if err != nil || !has || stakerInfo == nil {
						Logger.log.Errorf("fmt.Errorf(\"Cannot get staker info for outpubickey %v\", outPublicKey), error %+v", outPublicKey, err)
						//return nil, nil, NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Cannot get staker info for outpubickey %v", outPublicKey))
						continue
					}

					if stakerInfo.AutoStaking() {
						continue
						//return nil, nil, NewBlockChainError(ProcessSalaryInstructionsError, fmt.Errorf("Beacon kick out this key, but autostaking still true %v", outPublicKey))
					}

					// If autostaking or staker who not has tx staking, do nothing
					if stakerInfo.TxStakingID() == common.HashH([]byte{0}) {
						continue
					}
					Logger.log.Info("stakerInfo.TxStakingID().String():", stakerInfo.TxStakingID().String())
					if _, ok := res[stakerInfo.TxStakingID()]; ok {
						err = errors.Errorf("Dupdate return staking using tx staking %v", stakerInfo.TxStakingID().String())
						return nil, nil, err
					}
					txData, err := blockchain.ShardChain[shardID].BlockStorage.GetStakingTx(stakerInfo.TxStakingID())
					if err != nil {
						continue
					}
					txMeta, ok := txData.GetMetadata().(*metadata.StakingMetadata)
					if !ok {
						Logger.log.Error("Can not parse meta data of this tx %v", txData.Hash().String())
						errorInstructions = append(errorInstructions, l)
						continue
					}
					keyWallet, err := wallet.Base58CheckDeserialize(txMeta.FunderPaymentAddress)
					if err != nil {
						Logger.log.Error("SA: cannot get payment address", txMeta, shardID)
						errorInstructions = append(errorInstructions, l)
						continue
					}
					Logger.log.Info("SA: build salary tx", txMeta.FunderPaymentAddress, shardID)
					paymentShardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
					if paymentShardID != shardID {
						err = NewBlockChainError(WrongShardIDError, fmt.Errorf("Staking Payment Address ShardID %+v, Not From Current Shard %+v", paymentShardID, shardID))
						errorInstructions = append(errorInstructions, l)
						Logger.log.Error(err)
						continue
					}
					res[stakerInfo.TxStakingID()] = returnStakingInfo{
						SwapoutPubKey: outPublicKey,
						FunderAddress: keyWallet.KeySet.PaymentAddress,
						StakingTx:     txData,
						StakingAmount: txMeta.StakingAmountShard,
					}
				}
			case instruction.RETURN_ACTION:
				returnStakingIns, err := instruction.ValidateAndImportReturnStakingInstructionFromString(l)
				if err != nil {
					Logger.log.Errorf("SKIP Return staking instruction %+v, error %+v", returnStakingIns, err)
					continue
				}
				for i, v := range returnStakingIns.GetPublicKey() {
					txHash := returnStakingIns.StakingTxHashes[i]
					txData, err := blockchain.ShardChain[shardID].BlockStorage.GetStakingTx(txHash)
					if err != nil {
						continue
					}
					txMeta, ok := txData.GetMetadata().(*metadata.StakingMetadata)
					if !ok {
						Logger.log.Errorf("Can not parse meta data of this tx %v", txData.Hash().String())
						errorInstructions = append(errorInstructions, l)
						continue
					}
					keyWallet, err := wallet.Base58CheckDeserialize(txMeta.FunderPaymentAddress)
					if err != nil {
						Logger.log.Error("SA: cannot get payment address", txMeta, shardID)
						errorInstructions = append(errorInstructions, l)
						continue
					}
					Logger.log.Info("SA: build salary tx", txMeta.FunderPaymentAddress, shardID)
					paymentShardID := common.GetShardIDFromLastByte(keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1])
					if paymentShardID != shardID {
						err = NewBlockChainError(WrongShardIDError, fmt.Errorf("Staking Payment Address ShardID %+v, Not From Current Shard %+v", paymentShardID, shardID))
						errorInstructions = append(errorInstructions, l)
						Logger.log.Error(err)
						continue
					}
					res[txHash] = returnStakingInfo{
						SwapoutPubKey: v,
						FunderAddress: keyWallet.KeySet.PaymentAddress,
						StakingTx:     txData,
						StakingAmount: txMeta.StakingAmountShard,
					}
				}
			}
		}
	}
	return res, errorInstructions, nil
}
