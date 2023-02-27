package bridgehub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"strconv"
)

type StakePRVRequest struct {
	ExtChainID   string      `json:"ExtChainID"`
	StakeAmount  uint64      `json:"StakeAmount"` // must be equal to vout value
	TokenID      common.Hash `json:"TokenID"`
	BridgePubKey string      `json:"BridgePubKey"` // staker's key
	metadataCommon.MetadataBase
}

type StakePRVRequestContentInst struct {
	ExtChainID       string      `json:"ExtChainID"`
	BridgePoolPubKey string      `json:"BridgePoolPubKey"` // TSS pubkey
	StakeAmount      uint64      `json:"StakeAmount"`      // must be equal to vout value
	TokenID          common.Hash `json:"TokenID"`
	BridgeID         string      `json:"BridgeID,omitempty"`
	TxReqID          string      `json:"TxReqID"`
}

type StakeReqAction struct {
	Meta          StakePRVRequest `json:"meta"`
	RequestedTxID *common.Hash    `json:"RequestedTxID"`
}

func NewStakePRVRequest(
	bridgePubKey string,
	stakeAmount uint64,
	tokenID common.Hash,
	metaType int,
) (*StakePRVRequest, error) {
	metadataBase := metadataCommon.MetadataBase{
		Type: metaType,
	}
	burningReq := &StakePRVRequest{
		BridgePubKey: bridgePubKey,
		StakeAmount:  stakeAmount,
		TokenID:      tokenID,
	}
	burningReq.MetadataBase = metadataBase
	return burningReq, nil
}

func (bReq StakePRVRequest) ValidateTxWithBlockChain(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	return true, nil
}

func (bReq StakePRVRequest) ValidateSanityData(chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, beaconHeight uint64, tx metadataCommon.Transaction) (bool, bool, error) {
	// check trigger feature or not
	if shardViewRetriever.GetTriggeredFeature()[metadataCommon.BridgeHubFeatureName] == 0 {
		return false, false, fmt.Errorf("Bridge Hub Feature has not been enabled yet %v", bReq.Type)
	}

	if bReq.StakeAmount == 0 {
		return false, false, fmt.Errorf("wrong request info's staked amount")
	}

	isBurned, burnCoin, burnedTokenID, err := tx.GetTxBurnData()
	if err != nil || !isBurned {
		return false, false, fmt.Errorf("it is not transaction stake. Error %v", err)
	}

	if !bytes.Equal(burnedTokenID[:], bReq.TokenID[:]) || bReq.TokenID.String() != common.PRVIDStr {
		return false, false, fmt.Errorf("wrong request info's token id and token staked")
	}

	burnAmount := burnCoin.GetValue()
	if burnAmount != bReq.StakeAmount {
		return false, false, fmt.Errorf("stake amount is incorrect %v", burnAmount)
	}

	return true, true, nil
}

func (bReq StakePRVRequest) ValidateMetadataByItself() bool {
	return bReq.Type == metadataCommon.StakePRVRequestMeta
}

func (bReq StakePRVRequest) Hash() *common.Hash {
	rawBytes, _ := json.Marshal(bReq)
	hash := common.HashH(rawBytes)
	return &hash
}

func (bReq *StakePRVRequest) BuildReqActions(tx metadataCommon.Transaction, chainRetriever metadataCommon.ChainRetriever, shardViewRetriever metadataCommon.ShardViewRetriever, beaconViewRetriever metadataCommon.BeaconViewRetriever, shardID byte, shardHeight uint64) ([][]string, error) {
	actionContent := map[string]interface{}{
		"meta":          *bReq,
		"RequestedTxID": tx.Hash(),
	}
	actionContentBytes, err := json.Marshal(actionContent)
	if err != nil {
		return [][]string{}, err
	}
	actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
	action := []string{strconv.Itoa(bReq.Type), actionContentBase64Str}
	return [][]string{action}, nil
}

func (bReq *StakePRVRequest) CalculateSize() uint64 {
	return metadataCommon.CalculateSize(bReq)
}
