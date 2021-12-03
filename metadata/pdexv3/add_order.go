package pdexv3

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

// AddOrderRequest
type AddOrderRequest struct {
	TokenToSell         common.Hash                         `json:"TokenToSell"`
	PoolPairID          string                              `json:"PoolPairID"`
	SellAmount          uint64                              `json:"SellAmount"`
	MinAcceptableAmount uint64                              `json:"MinAcceptableAmount"`
	Receiver            map[common.Hash]privacy.OTAReceiver `json:"Receiver"`
	NftID               common.Hash                         `json:"NftID"`
	metadataCommon.MetadataBase
}

func NewAddOrderRequest(
	tokenToSell common.Hash,
	pairID string,
	sellAmount uint64,
	minAcceptableAmount uint64,
	recv map[common.Hash]privacy.OTAReceiver,
	nftID common.Hash,
	metaType int,
) (*AddOrderRequest, error) {
	r := &AddOrderRequest{
		TokenToSell:         tokenToSell,
		PoolPairID:          pairID,
		SellAmount:          sellAmount,
		MinAcceptableAmount: minAcceptableAmount,
		Receiver:            recv,
		NftID:               nftID,
		MetadataBase: metadataCommon.MetadataBase{
			Type: metaType,
		},
	}
	return r, nil
}

func (req AddOrderRequest) NextAccessOTA() *AccessOTA {
	result := AccessOTA{}
	err := result.FromBytes([32]byte(req.NftID))
	if err != nil {
		return nil
	}
	return &result
}

func (req AddOrderRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	err := beaconViewRetriever.IsValidPoolPairID(req.PoolPairID)
	if err != nil {
		return false, err
	}
	
	err = beaconViewRetriever.IsValidNftID(req.NftID.String())
	if err != nil {
		if req.NextAccessOTA() == nil {
			return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, err)
		}
		valid, err1 := ValidPdexv3Access(nil, *req.NextAccessOTA(), tx, common.PRVCoinID, transactionStateDB)
		if valid {
			return true, nil
		}
		return false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, err1))
	}
	return true, nil
}

func (req AddOrderRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	if !chainRetriever.IsAfterPdexv3CheckPoint(beaconHeight) {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Feature pdexv3 has not been activated yet"))
	}

	// OTAReceiver check
	for _, item := range req.Receiver {
		if !item.IsValid() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("Invalid OTAReceiver %v", item))
		}
		if tx.GetSenderAddrLastByte() != item.GetShardID() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Invalid shard %d for Receiver - must equal sender shard",
					item.GetShardID()))
		}
	}

	// Burned coin check
	isBurn, burnedPRVCoin, burnedCoin, burnedTokenID, err := tx.GetTxFullBurnData()
	if err != nil || !isBurn {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Burned coins not found in trade request - %v", err))
	}
	if *burnedTokenID != req.TokenToSell {
		return false, false, metadataCommon.NewMetadataTxError(
			metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("Burned token ID mismatch - %v vs %v on metadata", *burnedTokenID, req.TokenToSell))
	}
	burnedTokenList := []common.Hash{*burnedTokenID}
	if burnedPRVCoin != nil {
		burnedTokenList = append(burnedTokenList, common.PRVCoinID)
	}
	for _, tokenID := range burnedTokenList {
		_, exists := req.Receiver[tokenID]
		if !exists {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Missing refund OTAReceiver for token %v", tokenID))
		}
	}

	if req.MinAcceptableAmount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("MinAcceptableAmount cannot be 0"))
	}
	if req.SellAmount == 0 {
		return false, false, metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError,
			fmt.Errorf("SellAmount cannot be 0"))
	}

	// Type vs burned token id + amount check
	switch tx.GetType() {
	case common.TxNormalType:
		// PRV must be burned
		if req.TokenToSell != common.PRVCoinID || burnedPRVCoin == nil {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned token invalid - must be PRV"))
		}
		// range check before adding
		if req.SellAmount > burnedPRVCoin.GetValue() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Sell amount invalid - must not exceed %d PRV burned", burnedPRVCoin.GetValue()))
		}
		if req.SellAmount != burnedPRVCoin.GetValue() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Sell amount invalid - must equal burned amount %d PRV after fee",
					burnedPRVCoin.GetValue()))
		}
	case common.TxCustomTokenPrivacyType:
		if req.TokenToSell != *burnedTokenID || burnedCoin == nil || burnedPRVCoin != nil {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Burned token invalid - must be %v", req.TokenToSell))
		}
		if req.SellAmount != burnedCoin.GetValue() {
			return false, false, metadataCommon.NewMetadataTxError(
				metadataCommon.PDEInvalidMetadataValueError,
				fmt.Errorf("Sell amount invalid - must equal burned amount %d after fee",
					burnedCoin.GetValue()))
		}

	default:
		return false, false, fmt.Errorf("Invalid transaction type %v for trade request", tx.GetType())
	}
	return true, true, nil
}

func (req AddOrderRequest) ValidateMetadataByItself() bool {
	return req.Type == metadataCommon.Pdexv3AddOrderRequestMeta
}

func (req AddOrderRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(req)
	hash := common.HashH([]byte(rawBytes))
	return &hash
}

func (req *AddOrderRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(req)
}

func (req *AddOrderRequest) GetOTADeclarations() []metadataCommon.OTADeclaration {
	var result []metadataCommon.OTADeclaration
	for currentTokenID, val := range req.Receiver {
		if currentTokenID != common.PRVCoinID {
			currentTokenID = common.ConfidentialAssetID
		}
		result = append(result, metadataCommon.OTADeclaration{
			PublicKey: val.PublicKey.ToBytes(), TokenID: currentTokenID,
		})
	}
	return result
}
