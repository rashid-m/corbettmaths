package instruction

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

// ReturnStakeInstruction :
// format: "return", "key1,key2,key3", "1231231,312312321,12312321", "100,100,100,100", "0,0,0,1"
type ReturnStakeInstruction struct {
	PublicKeys       []string
	PublicKeysStruct []incognitokey.CommitteePublicKey
	StakingTXIDs     []string
	StakingTxHashes  []common.Hash
	PercentReturns   []uint
	Reasons          []byte
}

func NewReturnStakeInsWithValue(
	publicKeys []string,
	txStake []string,
) *ReturnStakeInstruction {
	rsI := &ReturnStakeInstruction{}
	rsI, _ = rsI.SetPublicKeys(publicKeys)
	rsI, _ = rsI.SetStakingTXIDs(txStake)
	for _, _ = range publicKeys {
		rsI.PercentReturns = append(rsI.PercentReturns, 100)
	}
	return rsI
}

func NewReturnStakeIns() *ReturnStakeInstruction {
	return &ReturnStakeInstruction{}
}

func (rsI *ReturnStakeInstruction) IsEmpty() bool {
	return reflect.DeepEqual(rsI, NewReturnStakeIns()) ||
		len(rsI.PublicKeysStruct) == 0 && len(rsI.PublicKeys) == 0
}

func (rsI *ReturnStakeInstruction) SetPublicKeys(publicKeys []string) (*ReturnStakeInstruction, error) {
	if publicKeys == nil {
		return nil, errors.New("Public Keys Are Null")
	}
	rsI.PublicKeys = publicKeys
	publicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return nil, err
	}
	rsI.PublicKeysStruct = publicKeyStructs
	return rsI, nil
}

func (rsI *ReturnStakeInstruction) SetStakingTXIDs(txIDs []string) (*ReturnStakeInstruction, error) {
	if txIDs == nil {
		return nil, errors.New("Tx Hashes Are Null")
	}
	rsI.StakingTXIDs = txIDs
	rsI.StakingTxHashes = make([]common.Hash, len(txIDs))
	for i, v := range rsI.StakingTXIDs {
		temp, err := common.Hash{}.NewHashFromStr(v)
		if err != nil {
			return rsI, err
		}
		rsI.StakingTxHashes[i] = *temp
	}
	return rsI, nil
}

func (rsI *ReturnStakeInstruction) SetPercentReturns(percentReturns []uint) error {
	rsI.PercentReturns = percentReturns
	return nil
}

func (rsI *ReturnStakeInstruction) SetReasons(reason []byte) error {
	rsI.Reasons = reason
	return nil
}

func (rsI *ReturnStakeInstruction) GetType() string {
	return RETURN_ACTION
}

func (rsI *ReturnStakeInstruction) GetPercentReturns() []uint {
	return rsI.PercentReturns
}

func (rsI *ReturnStakeInstruction) GetStakingTX() []string {
	return rsI.StakingTXIDs
}

func (rsI *ReturnStakeInstruction) GetPublicKey() []string {
	return rsI.PublicKeys
}

func (rsI *ReturnStakeInstruction) GetReason() []byte {
	return rsI.Reasons
}

func (rsI *ReturnStakeInstruction) ToString() []string {
	returnStakeInsStr := []string{RETURN_ACTION}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.PublicKeys, SPLITTER))
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(rsI.StakingTXIDs, SPLITTER))
	percentReturnsStr := make([]string, len(rsI.PercentReturns))
	for i, v := range rsI.PercentReturns {
		percentReturnsStr[i] = strconv.Itoa(int(v))
	}
	returnStakeInsStr = append(returnStakeInsStr, strings.Join(percentReturnsStr, SPLITTER))
	if len(rsI.Reasons) != 0 {
		reasonStrs := make([]string, len(rsI.Reasons))
		for i, v := range rsI.Reasons {
			reasonStrs[i] = strconv.Itoa(int(v))
		}
		returnStakeInsStr = append(returnStakeInsStr, strings.Join(reasonStrs, SPLITTER))

	}
	return returnStakeInsStr
}

func (rsI *ReturnStakeInstruction) AddNewRequest(publicKey string, stakingTx string) {
	rsI.PublicKeys = append(rsI.PublicKeys, publicKey)
	publicKeyStruct, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{publicKey})
	rsI.PublicKeysStruct = append(rsI.PublicKeysStruct, publicKeyStruct[0])
	rsI.StakingTXIDs = append(rsI.StakingTXIDs, stakingTx)
	stakingTxHash, _ := common.Hash{}.NewHashFromStr(stakingTx)
	rsI.StakingTxHashes = append(rsI.StakingTxHashes, *stakingTxHash)
	rsI.PercentReturns = append(rsI.PercentReturns, 100)
	rsI.Reasons = append(rsI.Reasons, 255)
}

func ValidateAndImportReturnStakingInstructionFromString(instruction []string) (*ReturnStakeInstruction, error) {
	if err := ValidateReturnStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnStakingInstructionFromString(instruction)
}

func BuildReturnStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateReturnBeaconStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnBeaconStakingInstructionFromString(instruction)
}

func ImportReturnStakingInstructionFromString(instruction []string) (*ReturnStakeInstruction, error) {
	returnStakingIns := NewReturnStakeIns()
	var err error
	returnStakingIns, err = returnStakingIns.SetPublicKeys(strings.Split(instruction[1], SPLITTER))
	if err != nil {
		return nil, err
	}

	returnStakingIns, err = returnStakingIns.SetStakingTXIDs(strings.Split(instruction[2], SPLITTER))
	if err != nil {
		return nil, err
	}

	percentRetunrsStr := strings.Split(instruction[3], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		percentReturns[i] = uint(tempPercent)
	}
	returnStakingIns.SetPercentReturns(percentReturns)
	if len(instruction) == 5 {
		reasonsStr := strings.Split(instruction[4], SPLITTER)
		reasons := make([]byte, len(reasonsStr))
		for i, v := range reasonsStr {
			reason, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			reasons[i] = byte(reason)
		}
		returnStakingIns.SetReasons(reasons)
	}
	return returnStakingIns, err
}

func ValidateReturnStakingInstructionSanity(instruction []string) error {
	if (len(instruction) != 4) && (len(instruction) != 5) {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != RETURN_ACTION {
		return fmt.Errorf("invalid return staking action, %+v", instruction)
	}
	publicKeys := strings.Split(instruction[1], SPLITTER)
	_, err := incognitokey.CommitteeBase58KeyListToStruct(publicKeys)
	if err != nil {
		return err
	}
	txStakings := strings.Split(instruction[2], SPLITTER)
	for _, txStaking := range txStakings {
		_, err := common.Hash{}.NewHashFromStr(txStaking)
		if err != nil {
			log.Println("err:", err)
			return fmt.Errorf("invalid tx return staking %+v", err)
		}
	}
	percentRetunrsStr := strings.Split(instruction[3], SPLITTER)
	percentReturns := make([]uint, len(percentRetunrsStr))
	for i, v := range percentRetunrsStr {
		tempPercent, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid percent return %+v", err)
		}
		percentReturns[i] = uint(tempPercent)
	}
	if len(publicKeys) != len(txStakings) {
		return fmt.Errorf("invalid public key & tx staking txs length, %+v", instruction)
	}
	if len(percentReturns) != len(txStakings) {
		return fmt.Errorf("invalid reward percentReturns & tx stakings length, %+v", instruction)
	}
	if len(percentReturns) != len(publicKeys) {
		return fmt.Errorf("invalid reward percentReturns & public Keys length, %+v", instruction)
	}
	return nil
}
