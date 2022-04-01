package bridgeagg

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducer struct {
}

func (sp *stateProducer) modifyListTokens(
	contentStr string,
	unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	sDBs map[int]*statedb.StateDB,
	shardID byte,
) ([]string, map[common.Hash]map[uint]*Vault, error) {
	action := metadataCommon.NewAction()
	md := &metadataBridge.ModifyListToken{}
	action.Meta = md
	err := action.FromString(contentStr)
	if err != nil {
		return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(OtherError, err)
	}
	inst := metadataCommon.NewInstructionWithValue(
		metadataCommon.BridgeAggModifyListTokenMeta,
		common.AcceptedStatusStr,
		shardID,
		utils.EmptyString,
	)
	rejectContent := metadataCommon.NewRejectContentWithValue(action.TxReqID, 0, nil)

	for unifiedTokenID, vaults := range md.NewListTokens {
		_, found := unifiedTokenInfos[unifiedTokenID]
		if !found {
			unifiedTokenInfos[unifiedTokenID] = make(map[uint]*Vault)
		}
		for _, vault := range vaults {
			if _, found := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]; !found {
				rejectContent.ErrorCode = ErrCodeMessage[NotFoundNetworkIDError].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(NotFoundTokenIDInNetworkError, err)
				}
				return temp, unifiedTokenInfos, nil

			}
			newRewardReserve := vault.RewardReserve
			v := unifiedTokenInfos[unifiedTokenID][vault.NetworkID()]
			lastUpdatedRewardReserve := v.LastUpdatedRewardReserve()
			currentRewardReserve := v.CurrentRewardReserve()
			if newRewardReserve < lastUpdatedRewardReserve-currentRewardReserve {
				Logger.log.Warnf("tx %s UpdatedRewardReserve is invalid", action.TxReqID.String())
				rejectContent.ErrorCode = ErrCodeMessage[InvalidRewardReserveError].Code
				temp, err := inst.StringSliceWithRejectContent(rejectContent)
				if err != nil {
					return []string{}, unifiedTokenInfos, NewBridgeAggErrorWithValue(InvalidRewardReserveError, err)
				}
				return temp, unifiedTokenInfos, nil
			}
			v.SetLastUpdatedRewardReserve(vault.RewardReserve)
			v.SetCurrentRewardReserve(vault.RewardReserve)
			unifiedTokenInfos[unifiedTokenID][vault.NetworkID()] = v
		}
	}
	acceptedContent := metadataBridge.AcceptedModifyListToken{
		ModifyListToken: *md,
		TxReqID:         action.TxReqID,
	}
	contentBytes, err := json.Marshal(acceptedContent)
	if err != nil {
		return []string{}, unifiedTokenInfos, err
	}
	inst.Content = base64.StdEncoding.EncodeToString(contentBytes)
	return inst.StringSlice(), unifiedTokenInfos, nil
}

func (sp *stateProducer) convert(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault, sDBs map[int]*statedb.StateDB, shardID byte,
) (resInst [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	var errorType int
	var contents [][]byte
	resUnifiedTokenInfos = unifiedTokenInfos
	action := metadataCommon.NewAction()
	md := &metadataBridge.ConvertTokenToUnifiedTokenRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}

	defer func() {
		if err != nil {
			rejectedConvertRequest := metadataBridge.RejectedConvertTokenToUnifiedToken{
				TokenID:  md.TokenID,
				Amount:   md.Amount,
				Receiver: md.Receiver,
			}
			Logger.log.Warnf("Convert token with tx %s err %v", action.TxReqID.String(), err)
			content, err := json.Marshal(rejectedConvertRequest)
			if err != nil {
				errorType = OtherError
				return
			}
			contents = append(contents, content)
		}
		resInst, err = buildInstruction(
			metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta,
			errorType, contents, action.TxReqID, shardID, err,
		)
		if err != nil {
			Logger.log.Warnf("Cannot buildInstruction with tx %s err %v", action.TxReqID.String(), err)
		}
		err = nil
	}()

	if _, found := resUnifiedTokenInfos[md.UnifiedTokenID]; !found {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	if vault, found := resUnifiedTokenInfos[md.UnifiedTokenID][md.NetworkID]; !found {
		errorType = NotFoundNetworkIDError
		return
	} else {
		if vault.tokenID.String() != md.TokenID.String() {
			errorType = NotFoundTokenIDInNetworkError
			return
		}
		err = vault.convert(md.Amount)
		if err != nil {
			Logger.log.Warnf("Invalid convert amount error: %v tx %s", err, action.TxReqID.String())
			errorType = InvalidConvertAmountError
			return
		}
		resUnifiedTokenInfos[md.UnifiedTokenID][md.NetworkID] = vault
		acceptedContent := metadataBridge.AcceptedConvertTokenToUnifiedToken{
			ConvertTokenToUnifiedTokenRequest: *md,
			TxReqID:                           action.TxReqID,
		}
		var content []byte
		content, err = json.Marshal(acceptedContent)
		if err != nil {
			errorType = OtherError
			return
		}
		contents = append(contents, content)
	}
	return
}

func (sp *stateProducer) shield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	ac *metadata.AccumulatedValues, shardID byte,
	stateDBs map[int]*statedb.StateDB,
) (resInst [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	var errorType int
	var contents [][]byte
	resUnifiedTokenInfos = unifiedTokenInfos
	action := metadataCommon.NewAction()
	md := &metadataBridge.ShieldRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	defer func() {
		if err != nil {
			Logger.log.Warnf("tx %s shield with err %v", action.TxReqID.String(), err)
			contents = append(contents, action.TxReqID.Bytes())
		}
		resInst, err = buildInstruction(
			metadataCommon.IssuingUnifiedTokenRequestMeta,
			errorType, contents, action.TxReqID, shardID, err,
		)
		if err != nil {
			Logger.log.Warnf("Cannot build tx %s with err %v", action.TxReqID.String(), err)
		}
		err = nil
	}()

	vaults, err := CloneVaults(resUnifiedTokenInfos, md.IncTokenID)
	if err != nil {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	tempAC := new(metadata.AccumulatedValues)
	*tempAC = *ac
	var receiveShardID byte
	acceptedShieldRequestData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	acceptedShieldRequestRewardData := make([]metadataBridge.AcceptedShieldRequestData, len(md.Data))
	var rewardAmount uint64

	for index, data := range md.Data {
		switch data.NetworkID {
		case common.BSCNetworkID, common.ETHNetworkID, common.PLGNetworkID, common.FTMNetworkID:
			blockHash := rCommon.Hash{}
			err = blockHash.UnmarshalText([]byte(data.BlockHash))
			if err != nil {
				errorType = OtherError
				return
			}
			actualAmount, reward, receivingShardID, token, uniqTX, et, e := shieldEVM(
				md.IncTokenID, data.NetworkID, tempAC, shardID,
				action.TxReqID, vaults, stateDBs,
				action.ExtraData[index], blockHash, data.TxIndex,
			)
			if e != nil {
				errorType = et
				err = e
				return
			}
			acceptedShieldRequestData[index].IssuingAmount = actualAmount - reward
			acceptedShieldRequestRewardData[index].IssuingAmount = reward
			acceptedShieldRequestData[index].ExternalTokenID = token
			acceptedShieldRequestData[index].UniqTx = uniqTX
			acceptedShieldRequestData[index].NetworkID = data.NetworkID
			acceptedShieldRequestRewardData[index].NetworkID = data.NetworkID
			receiveShardID = receivingShardID
			rewardAmount += reward
		case common.DefaultNetworkID:
			errorType = OtherError
			err = errors.New("Invalid networkID")
			return
		default:
			errorType = OtherError
			err = errors.New("Not found networkID")
			return
		}
	}
	contents, err = buildAcceptedShieldContents(
		acceptedShieldRequestData, acceptedShieldRequestRewardData,
		md.PaymentAddress, md.IncTokenID, action.TxReqID, receiveShardID, rewardAmount != 0,
	)
	if err != nil {
		errorType = OtherError
		return
	}
	resUnifiedTokenInfos[md.IncTokenID] = vaults
	ac = tempAC
	return
}

func (sp *stateProducer) unshield(
	contentStr string, unifiedTokenInfos map[common.Hash]map[uint]*Vault,
	beaconHeight uint64, shardID byte,
	stateDB *statedb.StateDB,
) (resInsts [][]string, resUnifiedTokenInfos map[common.Hash]map[uint]*Vault, err error) {
	action := metadataCommon.NewAction()
	md := &metadataBridge.UnshieldRequest{}
	action.Meta = md
	err = action.FromString(contentStr)
	if err != nil {
		err = NewBridgeAggErrorWithValue(OtherError, err)
		return
	}
	resUnifiedTokenInfos = unifiedTokenInfos
	var errorType int
	var contents [][]byte
	var burningInsts [][]string
	var buringAmount uint64
	defer func() {
		if err != nil {
			rejectedUnshieldRequest := metadataBridge.RejectedUnshieldRequest{
				TokenID:  md.TokenID,
				Amount:   buringAmount,
				Receiver: md.Receiver,
			}
			Logger.log.Warnf("Unshield with tx %s err %v", action.TxReqID.String(), err)
			content, err := json.Marshal(rejectedUnshieldRequest)
			if err != nil {
				errorType = OtherError
				return
			}
			contents = append(contents, content)
		}
		insts, err := buildInstruction(
			metadataCommon.BurningUnifiedTokenRequestMeta,
			errorType, contents, action.TxReqID, shardID, err,
		)
		if err != nil {
			Logger.log.Warnf("Cannot buildInstruction with tx %s err %v", action.TxReqID.String(), err)
		}
		err = nil
		resInsts = append(resInsts, insts...)
		if len(burningInsts) != 0 {
			resInsts = append(resInsts, burningInsts...)
		}
	}()

	vaults, err := CloneVaults(resUnifiedTokenInfos, md.TokenID)
	if err != nil {
		errorType = NotFoundTokenIDInNetworkError
		return
	}
	var listAcceptedUnshieldRequestData []metadataBridge.AcceptedUnshieldRequestData

	for _, data := range md.Data {
		buringAmount += data.BurningAmount
		switch data.NetworkID {
		case common.BSCNetworkID, common.ETHNetworkID, common.PLGNetworkID, common.FTMNetworkID:
			externalTokenID, unshieldAmount, amount, fee, burningMetaType, et, e := unshieldEVM(data, stateDB, vaults, md.TokenID, action.TxReqID)
			if e != nil {
				errorType = et
				err = e
				burningInsts = [][]string{}
				return
			}
			h := big.NewInt(0).SetUint64(beaconHeight)
			burningInst := []string{
				strconv.Itoa(burningMetaType),
				strconv.Itoa(int(common.BridgeShardID)),
				base58.Base58Check{}.Encode(externalTokenID, 0x00),
				data.RemoteAddress,
				base58.Base58Check{}.Encode(unshieldAmount.Bytes(), 0x00),
				action.TxReqID.String(),
				base58.Base58Check{}.Encode(md.TokenID[:], 0x00),
				base58.Base58Check{}.Encode(h.Bytes(), 0x00),
			}
			burningInsts = append(burningInsts, burningInst)
			acceptedUnshieldRequestData := metadataBridge.AcceptedUnshieldRequestData{
				Fee:           fee,
				Amount:        amount,
				NetworkID:     data.NetworkID,
				IsDepositToSC: data.IsDepositToSC,
			}
			listAcceptedUnshieldRequestData = append(listAcceptedUnshieldRequestData, acceptedUnshieldRequestData)
		case common.DefaultNetworkID:
			errorType = OtherError
			err = errors.New("Invalid networkID")
			return
		default:
			errorType = OtherError
			err = errors.New("Not found networkID")
			return
		}
	}

	acceptedContent := metadataBridge.AcceptedUnshieldRequest{
		TokenID: md.TokenID,
		TxReqID: action.TxReqID,
		Data:    listAcceptedUnshieldRequestData,
	}
	content, err := json.Marshal(acceptedContent)
	if err != nil {
		errorType = OtherError
		return
	}
	contents = append(contents, content)

	return
}
