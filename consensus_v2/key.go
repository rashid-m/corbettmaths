package consensus_v2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/blsmultisig"
	"github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes/bridgesig"
	"github.com/incognitochain/incognito-chain/incognitokey"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func GetMiningKeyFromPrivateSeed(privateSeed string) (*signatureschemes2.MiningKey, error) {
	var miningKey signatureschemes2.MiningKey
	privateSeedBytes, _, err := base58.Base58Check{}.Decode(privateSeed)
	if err != nil {
		return nil, NewConsensusError(LoadKeyError, err)
	}
	blsPriKey, blsPubKey := blsmultisig.KeyGen(privateSeedBytes)
	miningKey.PriKey = map[string][]byte{}
	miningKey.PubKey = map[string][]byte{}
	miningKey.PriKey[common.BlsConsensus] = blsmultisig.SKBytes(blsPriKey)
	miningKey.PubKey[common.BlsConsensus] = blsmultisig.PKBytes(blsPubKey)
	bridgePriKey, bridgePubKey := bridgesig.KeyGen(privateSeedBytes)
	miningKey.PriKey[common.BridgeConsensus] = bridgesig.SKBytes(&bridgePriKey)
	miningKey.PubKey[common.BridgeConsensus] = bridgesig.PKBytes(&bridgePubKey)
	return &miningKey, nil
}

func (engine *Engine) GetCurrentMiningPublicKey() (publickey string, keyType string) {
	if engine != nil && engine.GetMiningPublicKeys() != nil {
		name := engine.consensusName
		pubkey := engine.GetMiningPublicKeys().GetMiningKeyBase58(name)
		return pubkey, name
	}
	return "", ""
}

func (engine *Engine) GetMiningPublicKeyByConsensus(consensusName string) (publickey string, err error) {
	keyBytes := map[string][]byte{}
	if engine != nil && engine.currentMiningProcess != nil {
		keytype := engine.currentMiningProcess.GetConsensusName()
		lightweightKey, exist := engine.GetMiningPublicKeys().MiningPubKey[common.BridgeConsensus]
		if !exist {
			return "", blsbft.NewConsensusError(blsbft.LoadKeyError, errors.New("Lightweight key not found"))
		}
		keyBytes[keytype], exist = engine.GetMiningPublicKeys().MiningPubKey[keytype]
		if !exist {
			return "", blsbft.NewConsensusError(blsbft.LoadKeyError, errors.New("Key not found"))
		}
		keyBytes[common.BridgeConsensus] = lightweightKey
		// pubkey := engine.userMiningPublicKeys[keytype]
		// return pubkey.GetMiningKeyBase58(keytype), keytype
	}
	res, err := json.Marshal(keyBytes)
	if err != nil {
		return "", blsbft.NewConsensusError(blsbft.UnExpectedError, err)
	}
	return string(res), nil
}

func (s *Engine) GetMiningPublicKeys() *incognitokey.CommitteePublicKey {
	return s.userMiningPublicKeys
}

func (s *Engine) GetNodeMiningPublicKeys() (userPks []*incognitokey.CommitteePublicKey) {
	for _, v := range s.validators {
		userPks = append(userPks, v.MiningKey.GetPublicKey())
	}
	return userPks
}

// legacy code -> get all key type  of 1 mining key
func (engine *Engine) GetAllMiningPublicKeys() []string {
	var keys []string
	for keyType, _ := range engine.userMiningPublicKeys.MiningPubKey {
		keys = append(keys, fmt.Sprintf("%v:%v", keyType, engine.userMiningPublicKeys.GetMiningKeyBase58(keyType)))
	}
	return keys
}

func (engine *Engine) SignDataWithCurrentMiningKey(
	data []byte,
) (
	publicKeyStr string,
	publicKeyType string,
	signature string,
	err error,
) {
	publicKeyStr = ""
	publicKeyType = ""
	signature = ""
	if engine != nil && engine.currentMiningProcess != nil {
		publicKeyType = engine.config.Blockchain.GetBeaconBestState().ConsensusAlgorithm
		publicKeyStr, err = engine.GetMiningPublicKeyByConsensus(publicKeyType)
		if err != nil {
			return
		}
		signature, err = engine.currentMiningProcess.SignData(data)
	}
	return
}

func (engine *Engine) VerifyData(data []byte, sig string, publicKey string, consensusType string) error {
	mapPublicKey := map[string][]byte{}
	err := json.Unmarshal([]byte(publicKey), &mapPublicKey)
	if err != nil {
		return blsbft.NewConsensusError(blsbft.LoadKeyError, err)
	}
	return engine.currentMiningProcess.ValidateData(data, sig, string(mapPublicKey[common.BridgeConsensus]))
}

func (engine *Engine) ValidateProducerPosition(
	blk types.BlockInterface,
	lastProposerIdx int,
	committee []incognitokey.CommitteePublicKey,
	lenProposers int,
	produceTimeSlot int64,
	proposeTimeSlot int64,
) error {

	//check producer,proposer,agg sig with this version
	producerPosition := blsbft.GetProposerIndexByRound(lastProposerIdx, blk.GetRound(), len(committee))
	if blk.GetVersion() == types.BFT_VERSION {
		tempProducer, err := committee[producerPosition].ToBase58()
		if err != nil {
			return fmt.Errorf("Cannot base58 a committee")
		}
		if strings.Compare(tempProducer, blk.GetProducer()) != 0 {
			return fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, blk.GetProducer())
		}
	} else {
		//validate producer
		producer := blk.GetProducer()
		tempProducerID := blockchain.GetProposerByTimeSlot(produceTimeSlot, lenProposers)
		tempProducer := committee[tempProducerID]
		b58Str, _ := tempProducer.ToBase58()
		if strings.Compare(b58Str, producer) != 0 {
			return fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", b58Str, producer)
		}

		//validate proposer
		proposer := blk.GetProposer()
		tempProducerID = blockchain.GetProposerByTimeSlot(proposeTimeSlot, lenProposers)
		tempProducer = committee[tempProducerID]
		b58Str, _ = tempProducer.ToBase58()
		if strings.Compare(b58Str, proposer) != 0 {
			return fmt.Errorf("Expect Proposer Public Key to be equal but get %+v From Index, %+v From Header", b58Str, proposer)
		}
	}

	return nil
}

func (engine *Engine) ValidateProducerSig(block types.BlockInterface, consensusType string) error {
	if block.GetVersion() == types.BFT_VERSION {
		return blsbft.ValidateProducerSigV1(block)
	} else {
		return blsbft.ValidateProducerSigV2(block)
	}
}

func (engine *Engine) ValidateBlockCommitteSig(block types.BlockInterface, committees []incognitokey.CommitteePublicKey, numFixNode int) error {
	return blsbft.ValidateCommitteeSig(block, committees, numFixNode)
}

func GenMiningKeyFromPrivateKey(privateKey string) (string, error) {
	privateSeed, err := LoadUserKeyFromIncPrivateKey(privateKey)
	if err != nil {
		return "", err
	}
	return privateSeed, nil
}

func (engine *Engine) ExtractBridgeValidationData(block types.BlockInterface) ([][]byte, []int, error) {
	return blsbft.ExtractBridgeValidationData(block)
}

func (engine *Engine) ExtractPortalV4ValidationData(block types.BlockInterface) ([]*portalprocessv4.PortalSig, error) {
	if block.GetVersion() >= 2 {
		return blsbft.ExtractPortalV4ValidationData(block)
	}
	return nil, blsbft.NewConsensusError(blsbft.ConsensusTypeNotExistError, errors.New(block.GetConsensusType()))
}

func LoadUserKeyFromIncPrivateKey(privateKey string) (string, error) {
	wl, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return "", NewConsensusError(LoadKeyError, err)
	}
	privateSeedBytes := common.HashB(common.HashB(wl.KeySet.PrivateKey))
	if err != nil {
		return "", NewConsensusError(LoadKeyError, err)
	}
	privateSeed := base58.Base58Check{}.Encode(privateSeedBytes, common.Base58Version)
	return privateSeed, nil
}
