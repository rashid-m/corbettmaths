package statedb

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func InsertETHTxHashIssued(stateDB *StateDB, uniqueEthTx []byte) error {
	key := GenerateBridgeEthTxObjectKey(uniqueEthTx)
	value := NewBridgeEthTxStateWithValue(uniqueEthTx)
	err := stateDB.SetStateObject(BridgeEthTxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertETHTxHashIssuedError, err)
	}
	return nil
}

func IsETHTxHashIssued(stateDB *StateDB, uniqueEthTx []byte) (bool, error) {
	key := GenerateBridgeEthTxObjectKey(uniqueEthTx)
	ethTxState, has, err := stateDB.getBridgeEthTxState(key)
	if err != nil {
		return false, NewStatedbError(IsETHTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(ethTxState.UniqueEthTx(), uniqueEthTx) != 0 {
		panic("same key wrong value")
	}
	return has, nil
}

func InsertBSCTxHashIssued(stateDB *StateDB, uniqueBSCTx []byte) error {
	key := GenerateBridgeBSCTxObjectKey(uniqueBSCTx)
	value := NewBridgeBSCTxStateWithValue(uniqueBSCTx)
	err := stateDB.SetStateObject(BridgeBSCTxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertBSCTxHashIssuedError, err)
	}
	return nil
}

func IsBSCTxHashIssued(stateDB *StateDB, uniqueBSCTx []byte) (bool, error) {
	key := GenerateBridgeBSCTxObjectKey(uniqueBSCTx)
	bscTxState, has, err := stateDB.getBridgeBSCTxState(key)
	if err != nil {
		return false, NewStatedbError(IsBSCTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(bscTxState.UniqueBSCTx(), uniqueBSCTx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func InsertPLGTxHashIssued(stateDB *StateDB, uniquePLGTx []byte) error {
	key := GenerateBridgePLGTxObjectKey(uniquePLGTx)
	value := NewBridgePLGTxStateWithValue(uniquePLGTx)
	err := stateDB.SetStateObject(BridgePLGTxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertPLGTxHashIssuedError, err)
	}
	return nil
}

func IsPLGTxHashIssued(stateDB *StateDB, uniquePLGTx []byte) (bool, error) {
	key := GenerateBridgePLGTxObjectKey(uniquePLGTx)
	bscTxState, has, err := stateDB.getBridgePLGTxState(key)
	if err != nil {
		return false, NewStatedbError(IsPLGTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(bscTxState.UniquePLGTx(), uniquePLGTx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func InsertPRVEVMTxHashIssued(stateDB *StateDB, uniquePRVEVMTx []byte) error {
	key := GenerateBrigePRVEVMObjectKey(uniquePRVEVMTx)
	value := NewBrigePRVEVMStateWithValue(uniquePRVEVMTx)
	err := stateDB.SetStateObject(BridgePRVEVMObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertPRVEVMTxHashIssuedError, err)
	}
	return nil
}

func IsPRVEVMTxHashIssued(stateDB *StateDB, uniquePRVEVMTx []byte) (bool, error) {
	key := GenerateBrigePRVEVMObjectKey(uniquePRVEVMTx)
	prvEVMTxState, has, err := stateDB.getBridgePRVEVMState(key)
	if err != nil {
		return false, NewStatedbError(BridgeInsertPRVEVMTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(prvEVMTxState.UniquePRVEVMTx(), uniquePRVEVMTx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func InsertFTMTxHashIssued(stateDB *StateDB, uniqueFTMTx []byte) error {
	key := GenerateBridgeFTMTxObjectKey(uniqueFTMTx)
	value := NewBridgeFTMTxStateWithValue(uniqueFTMTx)
	err := stateDB.SetStateObject(BridgeFTMTxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertFTMTxHashIssuedError, err)
	}
	return nil
}

func IsFTMTxHashIssued(stateDB *StateDB, uniqueFTMTx []byte) (bool, error) {
	key := GenerateBridgeFTMTxObjectKey(uniqueFTMTx)
	ftmTxState, has, err := stateDB.getBridgeFTMTxState(key)
	if err != nil {
		return false, NewStatedbError(IsFTMTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(ftmTxState.UniqueFTMTx(), uniqueFTMTx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func InsertNEARTxHashIssued(stateDB *StateDB, uniqueNEARTx []byte) error {
	key := GenerateBridgeNEARTxObjectKey(uniqueNEARTx)
	value := NewBridgeNEARTxStateWithValue(uniqueNEARTx)
	err := stateDB.SetStateObject(BridgeNEARTxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertNEARTxHashIssuedError, err)
	}
	return nil
}

func IsNEARTxHashIssued(stateDB *StateDB, uniqueNEARTx []byte) (bool, error) {
	key := GenerateBridgeNEARTxObjectKey(uniqueNEARTx)
	nearTxState, has, err := stateDB.getBridgeNEARTxState(key)
	if err != nil {
		return false, NewStatedbError(IsNEARTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(nearTxState.UniqueNEARTx(), uniqueNEARTx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func InsertAURORATxHashIssued(stateDB *StateDB, uniqueAURORATx []byte) error {
	key := GenerateBridgeAURORATxObjectKey(uniqueAURORATx)
	value := NewBridgeAURORATxStateWithValue(uniqueAURORATx)
	err := stateDB.SetStateObject(BridgeAURORATxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertAURORATxHashIssuedError, err)
	}
	return nil
}

func IsAURORATxHashIssued(stateDB *StateDB, uniqueAURORATx []byte) (bool, error) {
	key := GenerateBridgeAURORATxObjectKey(uniqueAURORATx)
	auroraTxState, has, err := stateDB.getBridgeAURORATxState(key)
	if err != nil {
		return false, NewStatedbError(IsAURORATxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(auroraTxState.UniqueAURORATx(), uniqueAURORATx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func InsertAVAXTxHashIssued(stateDB *StateDB, uniqueAVAXTx []byte) error {
	key := GenerateBridgeAVAXTxObjectKey(uniqueAVAXTx)
	value := NewBridgeAVAXTxStateWithValue(uniqueAVAXTx)
	err := stateDB.SetStateObject(BridgeAVAXTxObjectType, key, value)
	if err != nil {
		return NewStatedbError(BridgeInsertAVAXTxHashIssuedError, err)
	}
	return nil
}

func IsAVAXTxHashIssued(stateDB *StateDB, uniqueAVAXTx []byte) (bool, error) {
	key := GenerateBridgeAVAXTxObjectKey(uniqueAVAXTx)
	avaxTxState, has, err := stateDB.getBridgeAVAXTxState(key)
	if err != nil {
		return false, NewStatedbError(IsAVAXTxHashIssuedError, err)
	}
	if !has {
		return false, nil
	}
	if bytes.Compare(avaxTxState.UniqueAVAXTx(), uniqueAVAXTx) != 0 {
		panic("same key wrong value")
	}
	return true, nil
}

func CanProcessCIncToken(stateDB *StateDB, incTokenID common.Hash, privacyTokenExisted bool) (bool, error) {
	dBridgeTokenExisted, err := IsBridgeTokenExistedByType(stateDB, incTokenID, false)
	if err != nil {
		return false, NewStatedbError(CanProcessCIncTokenError, err)
	}
	if dBridgeTokenExisted {
		return false, nil
	}
	cBridgeTokenExisted, err := IsBridgeTokenExistedByType(stateDB, incTokenID, true)
	if err != nil {
		return false, NewStatedbError(CanProcessCIncTokenError, err)
	}
	if !cBridgeTokenExisted && privacyTokenExisted {
		return false, nil
	}
	return true, nil
}

func IsBridgeTokenExistedByType(stateDB *StateDB, incTokenID common.Hash, isCentralized bool) (bool, error) {
	key := GenerateBridgeTokenInfoObjectKey(isCentralized, incTokenID)
	tokenInfoState, has, err := stateDB.getBridgeTokenInfoState(key)
	if err != nil {
		return false, NewStatedbError(IsBridgeTokenExistedByTypeError, err)
	}
	if !has {
		return false, nil
	}
	tempIncoTokenID := tokenInfoState.IncTokenID()
	if !tempIncoTokenID.IsEqual(&incTokenID) || tokenInfoState.IsCentralized() != isCentralized {
		panic("same key wrong value")
	}
	return has, nil
}

func GetBridgeTokenByType(stateDB *StateDB, incTokenID common.Hash, isCentralized bool) (*BridgeTokenInfoState, bool, error) {
	return getBridgeTokenByType(stateDB, incTokenID, isCentralized)
}

func getBridgeTokenByType(stateDB *StateDB, incTokenID common.Hash, isCentralized bool) (*BridgeTokenInfoState, bool, error) {
	key := GenerateBridgeTokenInfoObjectKey(isCentralized, incTokenID)
	tokenInfoState, has, err := stateDB.getBridgeTokenInfoState(key)
	if err != nil {
		return nil, false, err
	}
	if !has {
		return tokenInfoState, false, nil
	}
	tempIncoTokenID := tokenInfoState.IncTokenID()
	if !tempIncoTokenID.IsEqual(&incTokenID) || tokenInfoState.IsCentralized() != isCentralized {
		panic("same key wrong value")
	}
	return tokenInfoState, has, nil
}

func CanProcessTokenPair(stateDB *StateDB, externalTokenID []byte, incTokenID common.Hash, privacyTokenExisted bool) (bool, error) {
	if len(externalTokenID) == 0 || len(incTokenID[:]) == 0 {
		return false, nil
	}
	cBridgeTokenExisted, err := IsBridgeTokenExistedByType(stateDB, incTokenID, true)
	if err != nil {
		return false, NewStatedbError(CanProcessTokenPairError, err)
	}
	if cBridgeTokenExisted {
		log.Println("WARNING: inc token was existed in centralized token set")
		return false, nil
	}
	dBridgeTokenExisted, err := IsBridgeTokenExistedByType(stateDB, incTokenID, false)
	if err != nil {
		return false, NewStatedbError(CanProcessTokenPairError, err)
	}
	log.Println("INFO: whether inc token was existed in decentralized token set: ", dBridgeTokenExisted)
	if !dBridgeTokenExisted && privacyTokenExisted {
		log.Println("WARNING: failed at condition 1: ", dBridgeTokenExisted, privacyTokenExisted)
		return false, nil
	}
	bridgeTokenInfoState, has, err := getBridgeTokenByType(stateDB, incTokenID, false)
	if err != nil {
		return false, NewStatedbError(CanProcessTokenPairError, err)
	}
	if has {
		if bytes.Compare(bridgeTokenInfoState.ExternalTokenID(), externalTokenID) == 0 {
			return true, nil
		}
		log.Println("WARNING: failed at condition 2:", bridgeTokenInfoState.ExternalTokenID()[:], externalTokenID[:])
		return false, nil
	}
	bridgeTokenInfoStates := stateDB.getAllBridgeTokenInfoState(false)
	for _, tempBridgeTokenInfoState := range bridgeTokenInfoStates {
		if bytes.Compare(tempBridgeTokenInfoState.ExternalTokenID(), externalTokenID) != 0 {
			continue
		}
		log.Println("WARNING: failed at condition 3:", tempBridgeTokenInfoState.ExternalTokenID()[:], externalTokenID[:])
		return false, nil
	}
	// both tokens are not existed -> can create new one
	return true, nil
}

func UpdateBridgeTokenInfo(stateDB *StateDB, incTokenID common.Hash, externalTokenID []byte, isCentralized bool, updatingAmount uint64, updateType string) error {
	dataaccessobject.Logger.Log.Infof("Update bridge token %v, isCentralized %v, amount %v\n", incTokenID.String(), isCentralized, updatingAmount)
	bridgeTokenInfoState, has, err := getBridgeTokenByType(stateDB, incTokenID, isCentralized)
	if err != nil {
		return NewStatedbError(UpdateBridgeTokenInfoError, err)
	}
	if !has {
		bridgeTokenInfoState.SetIncTokenID(incTokenID)
		bridgeTokenInfoState.SetExternalTokenID(externalTokenID)
		bridgeTokenInfoState.SetIsCentralized(isCentralized)
		if updateType == BridgeMinusOperator {
			bridgeTokenInfoState.SetAmount(0)
		} else {
			bridgeTokenInfoState.SetAmount(updatingAmount)
		}
		dataaccessobject.Logger.Log.Infof("Store Privacy Bridge Token %+v", incTokenID)
	} else {
		if updateType == BridgePlusOperator {
			bridgeTokenInfoState.SetAmount(bridgeTokenInfoState.Amount() + updatingAmount)
		} else if bridgeTokenInfoState.Amount() <= updatingAmount {
			bridgeTokenInfoState.SetAmount(0)
		} else {
			bridgeTokenInfoState.SetAmount(bridgeTokenInfoState.Amount() - updatingAmount)
		}
	}
	key := GenerateBridgeTokenInfoObjectKey(isCentralized, incTokenID)
	value := NewBridgeTokenInfoStateWithValue(bridgeTokenInfoState.IncTokenID(), bridgeTokenInfoState.ExternalTokenID(), bridgeTokenInfoState.Amount(), bridgeTokenInfoState.Network(), bridgeTokenInfoState.IsCentralized())
	err = stateDB.SetStateObject(BridgeTokenInfoObjectType, key, value)
	if err != nil {
		return NewStatedbError(UpdateBridgeTokenInfoError, err)
	}
	return nil
}

func GetAllBridgeTokens(stateDB *StateDB) ([]byte, error) {
	cBridgeTokenInfoStates := stateDB.getAllBridgeTokenInfoState(true)
	dBridgeTokenInfoStates := stateDB.getAllBridgeTokenInfoState(false)
	bridgeTokenInfos := []*rawdbv2.BridgeTokenInfo{}
	bridgeTokenInfoStates := append(cBridgeTokenInfoStates, dBridgeTokenInfoStates...)
	for _, bridgeTokenInfoState := range bridgeTokenInfoStates {
		tokenID := bridgeTokenInfoState.IncTokenID()
		tempBridgeTokenInfo := rawdbv2.NewBridgeTokenInfo(&tokenID, bridgeTokenInfoState.Amount(), bridgeTokenInfoState.ExternalTokenID(), bridgeTokenInfoState.Network(), bridgeTokenInfoState.IsCentralized())
		bridgeTokenInfos = append(bridgeTokenInfos, tempBridgeTokenInfo)
	}
	res, err := json.Marshal(bridgeTokenInfos)
	if err != nil {
		return []byte{}, NewStatedbError(GetAllBridgeTokensError, err)
	}
	return res, nil
}

func TrackBridgeReqWithStatus(stateDB *StateDB, txReqID common.Hash, status byte) error {
	key := GenerateBridgeStatusObjectKey(txReqID)
	value := NewBridgeStatusStateWithValue(txReqID, status)
	err := stateDB.SetStateObject(BridgeStatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(TrackBridgeReqWithStatusError, err)
	}
	return nil
}

func GetBridgeReqWithStatus(stateDB *StateDB, txReqID common.Hash) (byte, error) {
	key := GenerateBridgeStatusObjectKey(txReqID)
	bridgeStatusState, has, err := stateDB.getBridgeStatusState(key)
	if err != nil {
		return 0, NewStatedbError(GetBridgeReqWithStatusError, err)
	}
	if !has {
		return common.BridgeRequestNotFoundStatus, nil
	}
	tempTxReqID := bridgeStatusState.TxReqID()
	if !tempTxReqID.IsEqual(&txReqID) {
		panic("same key wrong value")
	}
	return bridgeStatusState.Status(), nil
}

func IsBridgeToken(stateDB *StateDB, tokenID common.Hash) (
	isBridgeTokens bool,
	err error,
) {
	isBridgeTokens, err = IsBridgeTokenExistedByType(stateDB, tokenID, true)
	if err != nil {
		return false, err
	}
	if !isBridgeTokens {
		return IsBridgeTokenExistedByType(stateDB, tokenID, false)
	}
	return isBridgeTokens, err
}

func GetBridgeTokens(stateDB *StateDB) ([]*rawdbv2.BridgeTokenInfo, error) {
	cBridgeTokenInfoStates := stateDB.getAllBridgeTokenInfoState(true)
	dBridgeTokenInfoStates := stateDB.getAllBridgeTokenInfoState(false)
	bridgeTokenInfos := []*rawdbv2.BridgeTokenInfo{}
	bridgeTokenInfoStates := append(cBridgeTokenInfoStates, dBridgeTokenInfoStates...)
	for _, bridgeTokenInfoState := range bridgeTokenInfoStates {
		tokenID := bridgeTokenInfoState.IncTokenID()
		tempBridgeTokenInfo := rawdbv2.NewBridgeTokenInfo(&tokenID, bridgeTokenInfoState.Amount(), bridgeTokenInfoState.ExternalTokenID(), bridgeTokenInfoState.Network(), bridgeTokenInfoState.IsCentralized())
		bridgeTokenInfos = append(bridgeTokenInfos, tempBridgeTokenInfo)
	}
	return bridgeTokenInfos, nil
}
