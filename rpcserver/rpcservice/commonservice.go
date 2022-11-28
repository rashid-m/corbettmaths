package rpcservice

import (
	"bytes"
	"fmt"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/incognitochain/incognito-chain/metadata/evmcaller"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func NewContractingRequestMetadata(senderPrivateKeyStr string, tokenReceivers interface{}, tokenID string) (*metadata.ContractingRequest, *RPCError) {
	senderKey, err := wallet.Base58CheckDeserialize(senderPrivateKeyStr)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	paymentAddr := senderKey.KeySet.PaymentAddress

	_, voutsAmount, err := CreateCustomTokenPrivacyReceiverArray(tokenReceivers)
	if err != nil {
		return nil, NewRPCError(RPCInvalidParamsError, err)
	}
	tokenIDHash, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	meta, _ := metadata.NewContractingRequest(
		paymentAddr,
		uint64(voutsAmount),
		*tokenIDHash,
		metadata.ContractingRequestMeta,
	)

	return meta, nil
}

func CreateCustomTokenPrivacyReceiverArray(dataReceiver interface{}) ([]*privacy.PaymentInfo, int64, error) {
	if dataReceiver == nil {
		return nil, 0, errors.Errorf("data receiver is in valid")
	}
	paymentInfos := []*privacy.PaymentInfo{}
	voutsAmount := int64(0)
	receivers, ok := dataReceiver.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	for key, value := range receivers {
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil {
			return nil, 0, errors.Errorf("payment info %+v is invalid. Error %v\n", key, err)
		}
		if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
			return nil, 0, errors.Errorf("public key in payment info %+v is invalid\n", key)
		}
		amount, err := common.AssertAndConvertNumber(value)
		if err != nil {
			return nil, 0, errors.Errorf("amount payment address is invalid. Error %v\n", err)
		}
		temp := &privacy.PaymentInfo{
			PaymentAddress: keyWallet.KeySet.PaymentAddress,
			Amount:         amount,
		}
		paymentInfos = append(paymentInfos, temp)
		voutsAmount += int64(temp.Amount)
	}
	return paymentInfos, voutsAmount, nil
}

// CreateCustomTokenPrivacyReceiverArrayV2 parses data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenPrivacyReceiverArrayV2(dataReceiver interface{}) ([]*privacy.PaymentInfo, int64, error) {
	if dataReceiver == nil {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	result := []*privacy.PaymentInfo{}
	voutsAmount := int64(0)
	receivers, ok := dataReceiver.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	for key, value := range receivers {
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil {
			Logger.log.Errorf("Invalid key in CreateCustomTokenPrivacyReceiverArray %+v", key)
			return nil, 0, err
		}
		keySet := keyWallet.KeySet

		amount, err := common.AssertAndConvertStrToNumber(value)
		if err != nil {
			return nil, 0, err
		}

		temp := &privacy.PaymentInfo{
			PaymentAddress: keySet.PaymentAddress,
			Amount:         amount,
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Amount)
	}
	return result, voutsAmount, nil
}

func CreateCustomTokenPrivacyBurningReceiverArray(dataReceiver interface{}, bcr metadata.ChainRetriever, beaconHeight uint64) ([]*privacy.PaymentInfo, int64, error) {
	if dataReceiver == nil {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	var result []*privacy.PaymentInfo
	voutsAmount := int64(0)
	receivers, ok := dataReceiver.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}

	burningAddress := bcr.GetBurningAddress(beaconHeight)
	keyWalletBurningAccount, err := wallet.Base58CheckDeserialize(burningAddress)
	if err != nil {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress

	for key, value := range receivers {
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil {
			Logger.log.Errorf("Invalid key in CreateCustomTokenPrivacyReceiverArray %+v", key)
			return nil, 0, err
		}
		if !bytes.Equal(keyWallet.KeySet.PaymentAddress.Pk[:], paymentAddressBurningAccount.Pk[:]) {
			continue
		}

		amount, err := common.AssertAndConvertNumber(value)
		if err != nil {
			return nil, 0, errors.Errorf("amount payment address is invalid. Error %v\n", err)
		}
		temp := &privacy.PaymentInfo{
			PaymentAddress: keyWallet.KeySet.PaymentAddress,
			Amount:         amount,
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Amount)
	}
	return result, voutsAmount, nil
}

func NewBurningRequestMetadata(
	senderPrivateKeyStr string,
	tokenReceivers interface{},
	tokenID string,
	tokenName string,
	remoteAddress string,
	burningMetaType int,
	bcr metadata.ChainRetriever,
	beaconHeight uint64,
	version int8,
	networkID uint,
	expectedAmount uint64,
	isDepositToSC *bool,
) (*metadataBridge.BurningRequest, *RPCError) {
	senderKey, err := wallet.Base58CheckDeserialize(senderPrivateKeyStr)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	err = senderKey.KeySet.InitFromPrivateKey(&senderKey.KeySet.PrivateKey)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	paymentAddr := senderKey.KeySet.PaymentAddress

	_, voutsAmount, err := CreateCustomTokenPrivacyBurningReceiverArray(tokenReceivers, bcr, beaconHeight)
	if err != nil {
		return nil, NewRPCError(RPCInvalidParamsError, err)
	}
	tokenIDHash, err := common.Hash{}.NewHashFromStr(tokenID)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}

	meta, err := metadataBridge.NewBurningRequest(
		paymentAddr,
		uint64(voutsAmount),
		*tokenIDHash,
		tokenName,
		remoteAddress,
		burningMetaType,
	)
	if err != nil {
		return nil, NewRPCError(UnexpectedError, err)
	}
	if version == 1 {
		meta.BurnerAddress.OTAPublic = nil
	}

	return meta, nil
}

func GetETHHeaderByHash(ethBlockHash string) (*types.Header, error) {
	gethParam := config.Param().GethParam
	if len(gethParam.Host) < 1 {
		return nil, errors.New("Invalid param geth")
	}
	res, err := evmcaller.GetEVMHeaderResultMultipleHosts(rCommon.HexToHash(ethBlockHash), gethParam.Host, metadata.EVMConfirmationBlocks)
	if err != nil {
		return nil, err
	}
	return &res.Header, nil
}

// GetKeySetFromPrivateKeyParams - deserialize a private key string
// into keyWallet object and fill all keyset in keywallet with private key
// return key set and shard ID
func GetKeySetFromPrivateKeyParams(privateKeyWalletStr string) (*incognitokey.KeySet, byte, error) {
	// deserialize to crate keywallet object which contain private key
	keyWallet, err := wallet.Base58CheckDeserialize(privateKeyWalletStr)
	if err != nil {
		return nil, byte(0), err
	}
	if keyWallet.KeySet.PrivateKey != nil {
		return GetKeySetFromPrivateKey(keyWallet.KeySet.PrivateKey)
	} else {
		pk := keyWallet.KeySet.OTAKey.GetPublicSpend()
		if pk == nil {
			return nil, byte(0), errors.New("OTA Public Key not found")
		}
		pkb := pk.ToBytesS()
		shardID := common.GetShardIDFromLastByte(pkb[len(pkb)-1])
		return &keyWallet.KeySet, shardID, nil
	}
}

// GetKeySetFromPrivateKeyParams - deserialize a private key string
// into keyWallet object and fill all keyset in keywallet with private key
// return key set and shard ID
func GetKeySetFromPrivateKey(privateKey privacy.PrivateKey) (*incognitokey.KeySet, byte, error) {
	keySet := new(incognitokey.KeySet)
	// fill paymentaddress and readonly key with privatekey
	err := keySet.InitFromPrivateKey(&privateKey)
	if err != nil {
		return nil, byte(0), err
	}

	// calculate shard ID
	lastByte := keySet.PaymentAddress.Pk[len(keySet.PaymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)

	return keySet, shardID, nil
}

// GetKeySetFromPaymentAddressParam - deserialize a key string(wallet serialized)
// into keyWallet - this keywallet may contain
func GetKeySetFromPaymentAddressParam(paymentAddressStr string) (*incognitokey.KeySet, byte, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(paymentAddressStr)
	if err != nil {
		return nil, byte(0), err
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return nil, byte(0), errors.New("invalid payment address string")
	}

	// calculate shard ID
	lastByte := keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)

	return &keyWallet.KeySet, shardID, nil
}

func NewPaymentInfosFromReceiversParam(receiversParam map[string]interface{}) ([]*privacy.PaymentInfo, error) {
	paymentInfos := make([]*privacy.PaymentInfo, 0)
	for paymentAddressStr, amount := range receiversParam {
		keyWalletReceiver, err := wallet.Base58CheckDeserialize(paymentAddressStr)
		if err != nil {
			return nil, err
		}
		paymentInfo := &privacy.PaymentInfo{
			Amount:         uint64(amount.(float64)),
			PaymentAddress: keyWalletReceiver.KeySet.PaymentAddress,
		}
		paymentInfos = append(paymentInfos, paymentInfo)
	}

	return paymentInfos, nil
}

func GetStakingAmount(stakingType int, stakingShardAmountParam uint64) uint64 {
	amount := uint64(0)
	stakingData, _ := metadata.NewStakingMetadata(metadata.ShardStakingMeta, "", "", stakingShardAmountParam, "", "", true)
	// if stakingType == 1 {
	// 	amount = stakingData.GetBeaconStakeAmount()
	// }
	if stakingType == 0 {
		amount = stakingData.GetShardStateAmount()
	}

	return amount
}

func HashToIdenticon(hashStrs []interface{}) ([]string, error) {
	result := make([]string, 0)
	for _, hash := range hashStrs {
		temp, err := common.Hash{}.NewHashFromStr(hash.(string))
		if err != nil {
			return nil, errors.New("Hash string is invalid")
		}
		result = append(result, common.Render(temp.GetBytes()))
	}

	return result, nil
}

func GenerateTokenID(network string, name string) (common.Hash, error) {
	point := privacy.HashToPoint([]byte(network + "-" + name))
	hash := new(common.Hash)
	err := hash.SetBytes(point.ToBytesS())
	if err != nil {
		Logger.log.Error("Wrong param %+v", err)
		return common.Hash{}, err
	}
	return *hash, nil
}

func GetInputSerialnumber(inputCoins []privacy.PlainCoin) ([]string, error) {

	if inputCoins == nil || len(inputCoins) == 0 {
		return nil, errors.New("list of input coins is empty")
	}
	tmp := make([]string, len(inputCoins))
	for i, input := range inputCoins {
		tmp[i] = jsonresult.OperationPointPtrToBase58(input.GetKeyImage())
	}
	return tmp, nil
}
