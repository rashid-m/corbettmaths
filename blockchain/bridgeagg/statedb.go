package bridgeagg

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func CheckTokenIDExisted(sDBs map[int]*statedb.StateDB, tokenID common.Hash) error {
	for _, sDB := range sDBs {
		if statedb.PrivacyTokenIDExisted(sDB, tokenID) {
			return nil
		}
	}
	return fmt.Errorf("Cannot find tokenID %s in network", tokenID.String())
}

func InitStateFromDB(sDB *statedb.StateDB) (*State, error) {
	unifiedTokenStates, err := statedb.GetBridgeAggUnifiedTokens(sDB)
	if err != nil {
		return nil, err
	}
	unifiedTokenInfos := make(map[common.Hash]map[uint]*Vault)
	for _, unifiedTokenState := range unifiedTokenStates {
		unifiedTokenInfos[unifiedTokenState.TokenID()] = make(map[uint]*Vault)
		convertTokens, err := statedb.GetBridgeAggConvertedTokens(sDB, unifiedTokenState.TokenID())
		if err != nil {
			return nil, err
		}
		for _, convertToken := range convertTokens {
			state, err := statedb.GetBridgeAggVault(sDB, unifiedTokenState.TokenID(), convertToken.TokenID())
			if err != nil {
				state = statedb.NewBridgeAggVaultState()
			}
			vault := NewVaultWithValue(*state, nil, convertToken.TokenID()) // TODO: find externalTokenID here
			unifiedTokenInfos[unifiedTokenState.TokenID()][convertToken.NetworkID()] = vault
		}
	}
	return NewStateWithValue(unifiedTokenInfos), nil
}
