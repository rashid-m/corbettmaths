package pdexv3

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type ReceiverInfo struct {
	Address privacy.OTAReceiver `json:"Address"`
	Amount  uint64              `json:"Amount"`
}

// Check if the given OTA address is a valid address and has the expected shard ID
func isValidOTAReceiver(receiverAddress privacy.OTAReceiver, expectedShardID byte) (privacy.OTAReceiver, error) {
	if !receiverAddress.IsValid() {
		return receiverAddress, errors.New("ReceiverAddress is invalid")
	}

	pkb := receiverAddress.PublicKey.ToBytesS()
	currentShardID := common.GetShardIDFromLastByte(pkb[len(pkb)-1])
	if currentShardID != expectedShardID {
		return receiverAddress, errors.New("ReceiverAddress shard ID is wrong")
	}

	return receiverAddress, nil
}

type AccessOTA privacy.Point

func (ota AccessOTA) MarshalJSON() ([]byte, error) {
	return json.Marshal(ota.ToBytesS())
}

func (ota *AccessOTA) UnmarshalJSON(data []byte) error {
	var b []byte
	err := json.Unmarshal(data, &b)
	if err != nil {
		return err
	}
	return ota.FromBytesS(b)
}

func (ota AccessOTA) ToBytes() [32]byte {
	return privacy.Point(ota).ToBytes()
}

func (ota *AccessOTA) FromBytes(data [32]byte) error {
	_, err := (*privacy.Point)(ota).FromBytes(data)
	return err
}

func (ota AccessOTA) ToBytesS() []byte {
	temp := ota.ToBytes()
	return temp[:]
}

func (ota *AccessOTA) FromBytesS(data []byte) error {
	if len(data) != 32 {
		return fmt.Errorf("Invalid AccessOTA byte length %d", len(data))
	}
	var temp [32]byte
	copy(temp[:], data)
	*ota = AccessOTA{}
	err := ota.FromBytes(temp)
	return err
}

type BurnInputReader interface {
	DerivableBurnInput(*statedb.StateDB) (map[common.Hash]privacy.Point, error)
}

func ValidPdexv3Access(burnOTA *AccessOTA, nextOTA AccessOTA, tx metadataCommon.Transaction, accessTokenID common.Hash, transactionStateDB *statedb.StateDB) (bool, error) {
	if accessTokenID != common.PRVCoinID {
		accessTokenID = common.ConfidentialAssetID
	}
	// check that nextOTA is valid
	p := (*privacy.Point)(&nextOTA)
	if !p.PointValid() {
		return false, fmt.Errorf("invalid point - next AccessOTA %x", nextOTA.ToBytesS())
	}
	rawPubkey := p.ToBytesS()

	exists, otaStatus, err := statedb.HasOnetimeAddress(transactionStateDB, accessTokenID, rawPubkey)
	if err != nil {
		return false, err
	}
	// allow only OTAs from stored coins & not OTA-declarations
	if !exists {
		return false, fmt.Errorf("next AccessOTA %s does not exist", p.String())
	}
	if otaStatus == statedb.OTA_STATUS_OCCUPIED {
		return false, fmt.Errorf("next AccessOTA %s has invalid status %v", statedb.OTA_STATUS_OCCUPIED)
	}

	if burnOTA == nil {
		// accept metadata that declare nextAccessOTA the 1st time
		return true, nil
	}
	// check that burnOTA is valid
	{
		burnOTAPoint := (*privacy.Point)(burnOTA)
		if !burnOTAPoint.PointValid() {
			return false, fmt.Errorf("invalid point - burnt AccessOTA")
		}
		rawPubkey := burnOTAPoint.ToBytesS()

		exists, otaStatus, err := statedb.HasOnetimeAddress(transactionStateDB, accessTokenID, rawPubkey)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, fmt.Errorf("burnt AccessOTA %s does not exist", burnOTAPoint.String())
		}
		if otaStatus == statedb.OTA_STATUS_OCCUPIED {
			return false, fmt.Errorf("burnt AccessOTA %s has invalid status %v", statedb.OTA_STATUS_OCCUPIED)
		}

		// check that burnOTA is a tx input. Param tx cannot be nil
		btx, ok := tx.(BurnInputReader)
		if !ok {
			return false, fmt.Errorf("cannot identify burn input for tx %s", tx.Hash().String())
		}
		burnMap, err := btx.DerivableBurnInput(transactionStateDB)
		if err != nil {
			return false, fmt.Errorf("cannot identify burn input for tx %s: %v", tx.Hash().String(), err)
		}
		if burnInputCoinPubkey, exists := burnMap[common.PdexAccessCoinID]; !exists || !privacy.IsPointEqual(burnOTAPoint, &burnInputCoinPubkey) {
			return false, fmt.Errorf("burn missing: tx %s must burn AccessOTA %s from metadata", tx.Hash().String(), burnOTAPoint.String())
		}
		// check that burn output is PdexAccess, with sufficient amount
		isBurn, _, burnedTokenCoin, burnedTokenID, err := tx.GetTxFullBurnData()
		if !isBurn || *burnedTokenID != common.PdexAccessCoinID || burnedTokenCoin.GetValue() < MinPdexv3AccessBurn {
			return false, fmt.Errorf("invalid burn output (%v, %d) for Pdexv3 Access", burnedTokenID, burnedTokenCoin.GetValue())
		}
	}

	return true, nil
}

type AccessOption struct {
	NextOTA  *AccessOTA  `json:"NextOTA,omitempty"`
	BurntOTA *AccessOTA  `json:"BurntOTA,omitempty"`
	NftID    common.Hash `json:"NftID"`
}

func NewAccessOption() *AccessOption {
	return &AccessOption{}
}

func NewAccessOptionWithValue(nextOTA, burntOTA *AccessOTA, nftID common.Hash) *AccessOption {
	return &AccessOption{
		NextOTA:  nextOTA,
		BurntOTA: burntOTA,
		NftID:    nftID,
	}
}

func (a *AccessOption) IsValid(
	tx metadataCommon.Transaction,
	beaconViewRetriever metadataCommon.BeaconViewRetriever,
	transactionStateDB *statedb.StateDB,
	isWithdrawRequest bool,
) error {
	err := beaconViewRetriever.IsValidNftID(a.NftID.String())
	if err != nil {
		if isWithdrawRequest && a.BurntOTA == nil {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, errors.New("burn OTA is null")))
		}
		if a.NextOTA == nil {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, errors.New("next OTA is null")))
		}
		valid, err1 := ValidPdexv3Access(a.BurntOTA, *a.NextOTA, tx, common.ConfidentialAssetID, transactionStateDB)
		if !valid {
			return metadataCommon.NewMetadataTxError(metadataCommon.PDEInvalidMetadataValueError, fmt.Errorf("%v - %v", err, err1))
		}
		return nil
	}
	if a.BurntOTA != nil || a.NextOTA != nil {
		return errors.New("NftID and nextOTA can not be exist at the same time")
	}
	return nil
}

func (a *AccessOption) IsEmpty(isWithdrawRequest bool) bool {
	if isWithdrawRequest && a.BurntOTA == nil {
		return true
	}
	return a.NftID.IsZeroValue() && a.NextOTA == nil
}

func (a *AccessOption) UseNft() bool {
	return !a.NftID.IsZeroValue()
}
