package metadata

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BlockRewardAcceptInstruction struct {
	BeaconSalary uint64
}

type BeaconRewardInfo struct {
	BeaconReward   map[common.Hash]uint64
	PayToPublicKey string

	DelegatorReward map[common.Hash]uint64
	// InfoHash       *common.Hash
}

func BuildInstForBeaconReward(reward map[common.Hash]uint64, payToPublicKey []byte) ([]string, error) {
	publicKeyString := base58.Base58Check{}.Encode(payToPublicKey, common.ZeroByte)
	beaconRewardInfo := BeaconRewardInfo{
		PayToPublicKey: publicKeyString,
		BeaconReward:   reward,
	}

	contentStr, err := json.Marshal(beaconRewardInfo)
	if err != nil {
		return nil, NewMetadataTxError(BeaconBlockRewardBuildInstructionForBeaconBlockRewardError, err)
	}

	returnedInst := []string{
		strconv.Itoa(BeaconRewardRequestMeta),
		strconv.Itoa(int(common.GetShardIDFromLastByte(payToPublicKey[len(payToPublicKey)-1]))),
		"beaconRewardInst",
		string(contentStr),
	}

	return returnedInst, nil
}

func BuildInstForBeaconRewardV2(bRew, dRew map[common.Hash]uint64, payToPublicKey string) ([]string, error) {
	beaconRewardInfo := BeaconRewardInfo{
		PayToPublicKey:  payToPublicKey,
		BeaconReward:    bRew,
		DelegatorReward: dRew,
	}

	contentStr, err := json.Marshal(beaconRewardInfo)
	if err != nil {
		return nil, NewMetadataTxError(BeaconBlockRewardBuildInstructionForBeaconBlockRewardError, err)
	}

	returnedInst := []string{
		strconv.Itoa(BeaconRewardRequestMeta),
		strconv.Itoa(int(common.GetShardIDFromLastByte(payToPublicKey[len(payToPublicKey)-1]))),
		"beaconRewardInst",
		string(contentStr),
	}

	return returnedInst, nil
}

func NewBeaconBlockRewardInfoFromStr(inst string) (*BeaconRewardInfo, error) {
	Ins := &BeaconRewardInfo{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, NewMetadataTxError(BeaconBlockRewardNewBeaconBlockRewardInfoFromStrError, err)
	}
	return Ins, nil
}

type BeaconBlockSalaryRes struct {
	MetadataBase
	BeaconBlockHeight uint64
	ProducerAddress   *privacy.PaymentAddress
	InfoHash          *common.Hash
}

type BeaconBlockSalaryInfo struct {
	BeaconSalary      uint64
	PayToAddress      *privacy.PaymentAddress
	BeaconBlockHeight uint64
	InfoHash          *common.Hash
}

func (sbsRes BeaconBlockSalaryRes) CheckTransactionFee(tr Transaction, minFee uint64, beaconHeight int64, db *statedb.StateDB) bool {
	// no need to have fee for this tx
	return true
}

func (sbsRes BeaconBlockSalaryRes) ValidateTxWithBlockChain(tx Transaction, chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, shardID byte, transactionStateDB *statedb.StateDB) (bool, error) {
	// no need to validate tx with blockchain, just need to validate with request tx (via RequestedTxID) in current block
	return false, nil
}

func (sbsRes BeaconBlockSalaryRes) ValidateSanityData(chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever, beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {
	_, err := AssertPaymentAddressAndTxVersion(sbsRes.ProducerAddress, tx.GetVersion())
	if err != nil {
		return false, false, err
	}
	return false, true, nil
}

func (sbsRes BeaconBlockSalaryRes) ValidateMetadataByItself() bool {
	// The validation just need to check at tx level, so returning true here
	return true
}

func (sbsRes BeaconBlockSalaryRes) Hash() *common.Hash {
	record := sbsRes.ProducerAddress.String()
	record += string(sbsRes.BeaconBlockHeight)
	record += sbsRes.InfoHash.String()

	// final hash
	record += sbsRes.MetadataBase.Hash().String()
	hash := common.HashH([]byte(record))
	return &hash
}
