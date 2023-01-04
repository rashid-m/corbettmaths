package metadata

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/wallet"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

// UnStakingMetadata : unstaking metadata
type UnStakingMetadata struct {
	MetadataBaseWithSignature
	CommitteePublicKey string
}

func (meta *UnStakingMetadata) Hash() *common.Hash {
	record := strconv.Itoa(meta.Type)
	data := []byte(record)
	hash := common.HashH(data)
	return &hash
}

func (meta *UnStakingMetadata) HashWithoutSig() *common.Hash {
	return meta.MetadataBaseWithSignature.Hash()
}

func (*UnStakingMetadata) ShouldSignMetaData() bool { return true }

// NewUnStakingMetadata : Constructor of UnStakingMetadata struct
func NewUnStakingMetadata(committeePublicKey string) (*UnStakingMetadata, error) {
	metadataBase := NewMetadataBaseWithSignature(UnStakingMeta)
	return &UnStakingMetadata{
		MetadataBaseWithSignature: *metadataBase,
		CommitteePublicKey:        committeePublicKey,
	}, nil
}

// ValidateMetadataByItself Validate data format/type in unStakingMetadata
func (unStakingMetadata *UnStakingMetadata) ValidateMetadataByItself() bool {
	CommitteePublicKey := new(incognitokey.CommitteePublicKey)
	if unStakingMetadata.Type != UnStakingMeta {
		return false
	}
	if err := CommitteePublicKey.FromString(unStakingMetadata.CommitteePublicKey); err != nil {
		return false
	}
	if !CommitteePublicKey.CheckSanityData() {
		return false
	}
	return true
}

// ValidateTxWithBlockChain Validate Condition to Request Unstake With Blockchain
// - Requested Committee Publickey is in candidate, pending validator,
// - Requested Committee Publickey is in staking tx list,
// - Requester (sender of tx) must be address, which create staking transaction for current requested committee public key
func (unStakingMetadata UnStakingMetadata) ValidateTxWithBlockChain(tx Transaction,
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever,
	shardID byte,
	transactionStateDB *statedb.StateDB,
) (bool, error) {
	requestedPublicKey := unStakingMetadata.CommitteePublicKey

	stakerInfo, has, _ := beaconViewRetriever.GetStakerInfo(requestedPublicKey)
	if has {
		_, _, _, _, stakingTx, err := chainRetriever.GetTransactionByHash(stakerInfo.TxStakingID())
		if err != nil {
			return false, NewMetadataTxError(UnStakingRequestStakingTransactionNotFoundError, err)
		}

		stakingMetadata := stakingTx.GetMetadata().(*StakingMetadata)
		funderPaymentAddress := stakingMetadata.FunderPaymentAddress
		funderWallet, err := wallet.Base58CheckDeserialize(funderPaymentAddress)
		if err != nil || funderWallet == nil {
			return false, errors.New("Invalid Funder Payment Address, Failed to Deserialized Into Key Wallet")
		}

		if ok, err := unStakingMetadata.MetadataBaseWithSignature.VerifyMetadataSignature(funderWallet.KeySet.PaymentAddress.Pk, tx); !ok || err != nil {
			return false, NewMetadataTxError(ConsensusMetadataInvalidTransactionSenderError, fmt.Errorf("CheckAuthorizedSender fail"))
		}

		waitingValidatorsList, err := incognitokey.CommitteeKeyListToString(beaconViewRetriever.CandidateWaitingForNextRandom())
		if err != nil {
			return false, err
		}
		index := common.IndexOfStr(requestedPublicKey, waitingValidatorsList)
		if index == -1 {
			if !stakerInfo.AutoStaking() {
				return false, NewMetadataTxError(UnstakingRequestAlreadyUnstake, errors.New("Public Key Has Already Been Unstaked"))
			}
		}
		return true, nil
	}

	beaconStakerInfo, has, _ := beaconViewRetriever.GetBeaconStakerInfo(requestedPublicKey)
	if has {
		if beaconStakerInfo.Unstaking() {
			return false, NewMetadataTxError(UnstakingRequestAlreadyUnstake, errors.New("Public Key Has Already Been Unstaked"))
		}

		_, _, _, _, stakingTx, err := chainRetriever.GetTransactionByHash(beaconStakerInfo.StakingTxList()[0])
		if err != nil {
			return false, NewMetadataTxError(UnStakingRequestStakingTransactionNotFoundError, err)
		}

		stakingMetadata := stakingTx.GetMetadata().(*StakingMetadata)
		funderPaymentAddress := stakingMetadata.FunderPaymentAddress
		funderWallet, err := wallet.Base58CheckDeserialize(funderPaymentAddress)
		if err != nil || funderWallet == nil {
			return false, errors.New("Invalid Funder Payment Address, Failed to Deserialized Into Key Wallet")
		}

		if ok, err := unStakingMetadata.MetadataBaseWithSignature.VerifyMetadataSignature(funderWallet.KeySet.PaymentAddress.Pk, tx); !ok || err != nil {
			return false, NewMetadataTxError(ConsensusMetadataInvalidTransactionSenderError, fmt.Errorf("CheckAuthorizedSender fail"))
		}
		return true, nil
	}
	return false, NewMetadataTxError(UnStakingRequestNotFoundStakerInfoError, fmt.Errorf("Committee Publickey %+v has not staked yet", requestedPublicKey))
}

// ValidateSanityData :
func (unStakingMetadata UnStakingMetadata) ValidateSanityData(
	chainRetriever ChainRetriever, shardViewRetriever ShardViewRetriever,
	beaconViewRetriever BeaconViewRetriever, beaconHeight uint64, tx Transaction) (bool, bool, error) {

	if !unStakingMetadata.ValidateMetadataByItself() {
		return false, false, errors.New("Fail To Validate Metadata By Itself")
	}

	if tx.IsPrivacy() {
		return false, false, errors.New("Unstaking Request Transaction Is No Privacy Transaction")
	}
	return true, true, nil
}

// GetType :
func (unStakingMetadata UnStakingMetadata) GetType() int {
	return unStakingMetadata.Type
}

// CalculateSize :
func (unStakingMetadata *UnStakingMetadata) CalculateSize() uint64 {
	return calculateSize(unStakingMetadata)
}
