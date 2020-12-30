package blockchain

import (
	"bytes"
	"errors"
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"math/big"
	"strconv"

	rCommon "github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wallet"
)

// NOTE: for whole bridge's deposit process, anytime an error occurs it will be logged for debugging and the request will be skipped for retry later. No error will be returned so that the network can still continue to process others.

func buildInstruction(metaType int, shardID byte, instStatus string, contentStr string) []string {
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		instStatus,
		contentStr,
	}
}

func getShardIDFromPaymentAddress(addressStr string) (byte, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(addressStr)
	if err != nil {
		return byte(0), err
	}
	if len(keyWallet.KeySet.PaymentAddress.Pk) == 0 {
		return byte(0), errors.New("Payment address' public key must not be empty")
	}
	// calculate shard ID
	lastByte := keyWallet.KeySet.PaymentAddress.Pk[len(keyWallet.KeySet.PaymentAddress.Pk)-1]
	shardID := common.GetShardIDFromLastByte(lastByte)
	return shardID, nil
}

func (blockchain *BlockChain) buildInstructionsForContractingReq(
	contentStr string,
	shardID byte,
	metaType int,
) ([][]string, error) {
	inst := buildInstruction(metaType, shardID, "accepted", contentStr)
	return [][]string{inst}, nil
}

func (blockchain *BlockChain) buildInstructionsForIssuingReq(
	beaconBestState *BeaconBestState,
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	Logger.log.Info("[Centralized bridge token issuance] Starting...")
	instructions := [][]string{}
	issuingReqAction, err := metadata.ParseIssuingInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while parsing issuing action content: ", err)
		return nil, nil
	}

	issuingReq := issuingReqAction.Meta
	issuingTokenID := issuingReq.TokenID
	issuingTokenName := issuingReq.TokenName
	rejectedInst := buildInstruction(metaType, shardID, "rejected", issuingReqAction.TxReqID.String())

	if !ac.CanProcessCIncToken(issuingTokenID) {
		Logger.log.Warnf("WARNING: The issuing token (%s) was already used in the current block.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInAllShards(beaconBestState, issuingTokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	ok, err := statedb.CanProcessCIncToken(stateDB, issuingTokenID, privacyTokenExisted)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if !ok {
		Logger.log.Warnf("WARNING: The issuing token (%s) was already used in the previous blocks.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}
	pTokenInitInfo, err := statedb.GetPTokenInit(stateDB, issuingTokenID.String())
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the centralized token's existence against initialized privacy custom token on beacon: ", err)
		return append(instructions, rejectedInst), nil
	}
	if pTokenInitInfo != nil { // existed
		Logger.log.Warnf("WARNING: The centralized token (%s) was already existed in privacy custom token list in the previous blocks on beacon.", issuingTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	if len(issuingReq.ReceiverAddress.Pk) == 0 {
		Logger.log.Warn("WARNING: invalid receiver address")
		return append(instructions, rejectedInst), nil
	}
	lastByte := issuingReq.ReceiverAddress.Pk[len(issuingReq.ReceiverAddress.Pk)-1]
	receivingShardID := common.GetShardIDFromLastByte(lastByte)

	issuingAcceptedInst := metadata.IssuingAcceptedInst{
		ShardID:         receivingShardID,
		DepositedAmount: issuingReq.DepositedAmount,
		ReceiverAddr:    issuingReq.ReceiverAddress,
		IncTokenID:      issuingTokenID,
		IncTokenName:    issuingTokenName,
		TxReqID:         issuingReqAction.TxReqID,
	}
	issuingAcceptedInstBytes, err := json.Marshal(issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while marshaling issuingAccepted instruction: ", err)
		return append(instructions, rejectedInst), nil
	}

	ac.CBridgeTokens = append(ac.CBridgeTokens, &issuingTokenID)
	returnedInst := buildInstruction(metaType, shardID, "accepted", base64.StdEncoding.EncodeToString(issuingAcceptedInstBytes))
	return append(instructions, returnedInst), nil
}

func (blockchain *BlockChain) buildInstructionsForIssuingETHReq(
	beaconBestState *BeaconBestState,
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	Logger.log.Info("[Decentralized bridge token issuance] Starting...")
	instructions := [][]string{}
	issuingETHReqAction, err := metadata.ParseETHIssuingInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while parsing issuing action content: ", err)
		return nil, nil
	}
	md := issuingETHReqAction.Meta
	rejectedInst := buildInstruction(metaType, shardID, "rejected", issuingETHReqAction.TxReqID.String())

	ethReceipt := issuingETHReqAction.ETHReceipt
	if ethReceipt == nil {
		Logger.log.Warn("WARNING: eth receipt is null.")
		return append(instructions, rejectedInst), nil
	}

	// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
	// so must build unique eth tx as combination of block hash and tx index.
	uniqETHTx := append(md.BlockHash[:], []byte(strconv.Itoa(int(md.TxIndex)))...)
	isUsedInBlock := metadata.IsETHTxHashUsedInBlock(uniqETHTx, ac.UniqETHTxsUsed)
	if isUsedInBlock {
		Logger.log.Warn("WARNING: already issued for the hash in current block: ", uniqETHTx)
		return append(instructions, rejectedInst), nil
	}
	isIssued, err := statedb.IsETHTxHashIssued(stateDB, uniqETHTx)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the eth tx hash is issued or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if isIssued {
		Logger.log.Warn("WARNING: already issued for the hash in previous blocks: ", uniqETHTx)
		return append(instructions, rejectedInst), nil
	}

	logMap, err := metadata.PickAndParseLogMapFromReceipt(ethReceipt, blockchain.config.ChainParams.EthContractAddressStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while parsing log map from receipt: ", err)
		return append(instructions, rejectedInst), nil
	}
	if logMap == nil {
		Logger.log.Warn("WARNING: could not find log map out from receipt")
		return append(instructions, rejectedInst), nil
	}

	logMapBytes, _ := json.Marshal(logMap)
	Logger.log.Warn("INFO: eth logMap json - ", string(logMapBytes))

	// the token might be ETH/ERC20
	ethereumAddr, ok := logMap["token"].(rCommon.Address)
	if !ok {
		Logger.log.Warn("WARNING: could not parse eth token id from log map.")
		return append(instructions, rejectedInst), nil
	}
	ethereumToken := ethereumAddr.Bytes()
	canProcess, err := ac.CanProcessTokenPair(ethereumToken, md.IncTokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while checking it can process for token pair on the current block or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if !canProcess {
		Logger.log.Warn("WARNING: pair of incognito token id & ethereum's id is invalid in current block")
		return append(instructions, rejectedInst), nil
	}
	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInAllShards(beaconBestState, md.IncTokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking it can process for the incognito token or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	isValid, err := statedb.CanProcessTokenPair(stateDB, ethereumToken, md.IncTokenID, privacyTokenExisted)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while checking it can process for token pair on the previous blocks or not: ", err)
		return append(instructions, rejectedInst), nil
	}
	if !isValid {
		Logger.log.Warn("WARNING: pair of incognito token id & ethereum's id is invalid with previous blocks")
		return append(instructions, rejectedInst), nil
	}

	pTokenInitInfo, err := statedb.GetPTokenInit(stateDB, md.IncTokenID.String())
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the decentralized token's existence against initialized privacy custom token on beacon: ", err)
		return append(instructions, rejectedInst), nil
	}
	if pTokenInitInfo != nil { // existed
		Logger.log.Warnf("WARNING: The decentralized token (%s) was already existed in privacy custom token list in the previous blocks on beacon.", md.IncTokenID.String())
		return append(instructions, rejectedInst), nil
	}

	addressStr, ok := logMap["incognitoAddress"].(string)
	if !ok {
		Logger.log.Warn("WARNING: could not parse incognito address from eth log map.")
		return append(instructions, rejectedInst), nil
	}
	amt, ok := logMap["amount"].(*big.Int)
	if !ok {
		Logger.log.Warn("WARNING: could not parse amount from eth log map.")
		return append(instructions, rejectedInst), nil
	}
	amount := uint64(0)
	if bytes.Equal(rCommon.HexToAddress(common.EthAddrStr).Bytes(), ethereumToken) {
		// convert amt from wei (10^18) to nano eth (10^9)
		amount = big.NewInt(0).Div(amt, big.NewInt(1000000000)).Uint64()
	} else { // ERC20
		amount = amt.Uint64()
	}

	receivingShardID, err := getShardIDFromPaymentAddress(addressStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while getting shard id from payment address: ", err)
		return append(instructions, rejectedInst), nil
	}

	issuingETHAcceptedInst := metadata.IssuingETHAcceptedInst{
		ShardID:         receivingShardID,
		IssuingAmount:   amount,
		ReceiverAddrStr: addressStr,
		IncTokenID:      md.IncTokenID,
		TxReqID:         issuingETHReqAction.TxReqID,
		UniqETHTx:       uniqETHTx,
		ExternalTokenID: ethereumToken,
	}
	issuingETHAcceptedInstBytes, err := json.Marshal(issuingETHAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while marshaling issuingETHAccepted instruction: ", err)
		return append(instructions, rejectedInst), nil
	}
	ac.UniqETHTxsUsed = append(ac.UniqETHTxsUsed, uniqETHTx)
	ac.DBridgeTokenPair[md.IncTokenID.String()] = ethereumToken

	acceptedInst := buildInstruction(metaType, shardID, "accepted", base64.StdEncoding.EncodeToString(issuingETHAcceptedInstBytes))
	return append(instructions, acceptedInst), nil
}

func (blockGenerator *BlockGenerator) buildIssuanceTx(contentStr string, producerPrivateKey *privacy.PrivateKey, shardID byte, shardView *ShardBestState, beaconView *BeaconBestState) (metadata.Transaction, error) {
	Logger.log.Info("[Centralized bridge token issuance] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while decoding content string of accepted issuance instruction: ", err)
		return nil, nil
	}
	var issuingAcceptedInst metadata.IssuingAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while unmarshaling accepted issuance instruction: ", err)
		return nil, nil
	}

	if shardID != issuingAcceptedInst.ShardID {
		return nil, nil
	}
	issuingRes := metadata.NewIssuingResponse(
		issuingAcceptedInst.TxReqID,
		metadata.IssuingResponseMeta,
	)
	receiver := &privacy.PaymentInfo{
		Amount:         issuingAcceptedInst.DepositedAmount,
		PaymentAddress: issuingAcceptedInst.ReceiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingAcceptedInst.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   issuingAcceptedInst.IncTokenName,
		PropertySymbol: issuingAcceptedInst.IncTokenName,
		Amount:         issuingAcceptedInst.DepositedAmount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []*privacy.InputCoin{},
		Mintable:       true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			shardView.GetCopiedTransactionStateDB(),
			issuingRes,
			false,
			false,
			shardID,
			nil,
			beaconView.GetBeaconFeatureStateDB()))

	if initErr != nil {
		Logger.log.Warn("WARNING: an error occured while initializing response tx: ", initErr)
		return nil, nil
	}
	Logger.log.Info("[Centralized token issuance] Create tx ok.")
	return resTx, nil
}

func (blockGenerator *BlockGenerator) buildETHIssuanceTx(contentStr string, producerPrivateKey *privacy.PrivateKey, shardID byte, shardView *ShardBestState, beaconView *BeaconBestState) (metadata.Transaction, error) {
	Logger.log.Info("[Decentralized bridge token issuance] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while decoding content string of ETH accepted issuance instruction: ", err)
		return nil, nil
	}
	var issuingETHAcceptedInst metadata.IssuingETHAcceptedInst
	err = json.Unmarshal(contentBytes, &issuingETHAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while unmarshaling ETH accepted issuance instruction: ", err)
		return nil, nil
	}

	if shardID != issuingETHAcceptedInst.ShardID {
		return nil, nil
	}
	key, err := wallet.Base58CheckDeserialize(issuingETHAcceptedInst.ReceiverAddrStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while deserializing receiver address string: ", err)
		return nil, nil
	}
	receiver := &privacy.PaymentInfo{
		Amount:         issuingETHAcceptedInst.IssuingAmount,
		PaymentAddress: key.KeySet.PaymentAddress,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], issuingETHAcceptedInst.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID: propID.String(),
		// PropertyName:   common.PETHTokenName,
		// PropertySymbol: common.PETHTokenName,
		Amount:      issuingETHAcceptedInst.IssuingAmount,
		TokenTxType: transaction.CustomTokenInit,
		Receiver:    []*privacy.PaymentInfo{receiver},
		TokenInput:  []*privacy.InputCoin{},
		Mintable:    true,
	}

	issuingETHRes := metadata.NewIssuingETHResponse(
		issuingETHAcceptedInst.TxReqID,
		issuingETHAcceptedInst.UniqETHTx,
		issuingETHAcceptedInst.ExternalTokenID,
		metadata.IssuingETHResponseMeta,
	)
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			shardView.GetCopiedTransactionStateDB(),
			issuingETHRes,
			false,
			false,
			shardID, nil,
			beaconView.GetBeaconFeatureStateDB()))

	if initErr != nil {
		Logger.log.Warn("WARNING: an error occured while initializing response tx: ", initErr)
		return nil, nil
	}
	Logger.log.Info("[Decentralized bridge token issuance] Create tx ok.")
	return resTx, nil
}

func (blockchain *BlockChain) buildInstructionsForInitPTokenReq(
	beaconBestState *BeaconBestState,
	stateDB *statedb.StateDB,
	contentStr string,
	shardID byte,
	metaType int,
	ac *metadata.AccumulatedValues,
) ([][]string, error) {
	Logger.log.Info("[Init privacy custom token] Starting...")
	instructions := [][]string{}
	initPTokenReqAction, err := metadata.ParseInitPTokenInstContent(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while parsing init privacy custom token action content: ", err)
		return nil, nil
	}

	initPTokenReq := initPTokenReqAction.Meta
	tokenID := initPTokenReqAction.TokenID
	tokenName := initPTokenReq.TokenName
	tokenSymbol := initPTokenReq.TokenSymbol
	rejectedInst := buildInstruction(metaType, shardID, "rejected", initPTokenReqAction.TxReqID.String())

	// check existence in the current block (on mem)
	if !ac.CanProcessPTokenInitialization(tokenID) {
		Logger.log.Warnf("WARNING: The privacy custom token (%s) was already used in the current block.", tokenID.String())
		return append(instructions, rejectedInst), nil
	}

	// check existence in previous blocks (on blockchain's db)
	privacyTokenExisted, err := blockchain.PrivacyTokenIDExistedInAllShards(beaconBestState, tokenID)
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the initializing token's existence on all shards: ", err)
		return append(instructions, rejectedInst), nil
	}
	if privacyTokenExisted {
		Logger.log.Warnf("WARNING: The initializing token (%s) was already existed in the previous blocks on all shards.", tokenID.String())
		return append(instructions, rejectedInst), nil
	}

	pTokenInitInfo, err := statedb.GetPTokenInit(stateDB, tokenID.String())
	if err != nil {
		Logger.log.Warn("WARNING: an issue occured while checking the initializing token's existence on beacon: ", err)
		return append(instructions, rejectedInst), nil
	}
	if pTokenInitInfo != nil { // existed
		Logger.log.Warnf("WARNING: The initializing token (%s) was already existed in the previous blocks on beacon.", tokenID.String())
		return append(instructions, rejectedInst), nil
	}

	lastByte := initPTokenReq.ReceiverAddress.Pk[len(initPTokenReq.ReceiverAddress.Pk)-1]
	receivingShardID := common.GetShardIDFromLastByte(lastByte)
	initPTokenAcceptedInst := metadata.InitPTokenAcceptedInst{
		ShardID:         receivingShardID,
		Amount: 				 initPTokenReq.Amount,
		ReceiverAddr:    initPTokenReq.ReceiverAddress,
		IncTokenID:      tokenID,
		IncTokenName:    tokenName,
		IncTokenSymbol:  tokenSymbol,
		TxReqID:         initPTokenReqAction.TxReqID,
	}
	initPTokenAcceptedInstBytes, err := json.Marshal(initPTokenAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while marshaling initPTokenAccepted instruction: ", err)
		return append(instructions, rejectedInst), nil
	}

	ac.InitializedPTokenIDs = append(ac.InitializedPTokenIDs, tokenID.String())
	returnedInst := buildInstruction(metaType, shardID, "accepted", base64.StdEncoding.EncodeToString(initPTokenAcceptedInstBytes))
	return append(instructions, returnedInst), nil
}

func (blockGenerator *BlockGenerator) buildPTokenInitializationTx(
	contentStr string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	shardView *ShardBestState,
	beaconView *BeaconBestState,
) (metadata.Transaction, error) {
	Logger.log.Info("[Init privacy custom token] Starting...")
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while decoding content string of accepted privacy token init instruction: ", err)
		return nil, nil
	}
	var initPTokenAcceptedInst metadata.InitPTokenAcceptedInst
	err = json.Unmarshal(contentBytes, &initPTokenAcceptedInst)
	if err != nil {
		Logger.log.Warn("WARNING: an error occured while unmarshaling accepted privacy token init instruction: ", err)
		return nil, nil
	}

	if shardID != initPTokenAcceptedInst.ShardID {
		return nil, nil
	}
	initPTokenRes := metadata.NewInitPTokenResponse(
		initPTokenAcceptedInst.TxReqID,
		metadata.InitPTokenResponseMeta,
	)
	receiver := &privacy.PaymentInfo{
		Amount:         initPTokenAcceptedInst.Amount,
		PaymentAddress: initPTokenAcceptedInst.ReceiverAddr,
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], initPTokenAcceptedInst.IncTokenID[:])
	propID := common.Hash(propertyID)
	tokenParams := &transaction.CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   initPTokenAcceptedInst.IncTokenName,
		PropertySymbol: initPTokenAcceptedInst.IncTokenSymbol,
		Amount:         initPTokenAcceptedInst.Amount,
		TokenTxType:    transaction.CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []*privacy.InputCoin{},
		Mintable:       true,
	}
	resTx := &transaction.TxCustomTokenPrivacy{}
	initErr := resTx.Init(
		transaction.NewTxPrivacyTokenInitParams(producerPrivateKey,
			[]*privacy.PaymentInfo{},
			nil,
			0,
			tokenParams,
			shardView.GetCopiedTransactionStateDB(),
			initPTokenRes,
			false,
			false,
			shardID,
			nil,
			beaconView.GetBeaconFeatureStateDB()))

	if initErr != nil {
		Logger.log.Warn("WARNING: an error occured while initializing response tx: ", initErr)
		return nil, nil
	}
	Logger.log.Info("[Init privacy custom token] Create tx ok.")
	return resTx, nil
}
