package statedb

import (
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func TrackPdexv3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte, statusContent []byte) error {
	key := GeneratePdexv3StatusObjectKey(statusType, statusSuffix)
	value := NewPdexv3StatusStateWithValue(statusType, statusSuffix, statusContent)
	err := stateDB.SetStateObject(Pdexv3StatusObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePdexv3StatusError, err)
	}
	return nil
}

func GetPdexv3Status(stateDB *StateDB, statusType []byte, statusSuffix []byte) ([]byte, error) {
	key := GeneratePdexv3StatusObjectKey(statusType, statusSuffix)
	s, has, err := stateDB.getPdexv3StatusByKey(key)
	if err != nil {
		return []byte{}, NewStatedbError(GetPdexv3StatusError, err)
	}
	if !has {
		return []byte{}, NewStatedbError(GetPdexv3StatusError, fmt.Errorf("status %+v with prefix %+v not found", string(statusType), statusSuffix))
	}
	return s.statusContent, nil
}

func StorePdexv3Params(
	stateDB *StateDB,
	defaultFeeRateBPS uint,
	feeRateBPS map[string]uint,
	prvDiscountPercent uint,
	tradingProtocolFeePercent uint,
	tradingStakingPoolRewardPercent uint,
	pdexRewardPoolPairsShare map[string]uint,
	stakingPoolsShare map[string]uint,
	stakingRewardTokens []common.Hash,
	mintNftRequireAmount uint64,
	maxOrdersPerNft uint,
) error {
	key := GeneratePdexv3ParamsObjectKey()
	value := NewPdexv3ParamsWithValue(
		defaultFeeRateBPS,
		feeRateBPS,
		prvDiscountPercent,
		tradingProtocolFeePercent,
		tradingStakingPoolRewardPercent,
		pdexRewardPoolPairsShare,
		stakingPoolsShare,
		stakingRewardTokens,
		mintNftRequireAmount,
		maxOrdersPerNft,
	)
	err := stateDB.SetStateObject(Pdexv3ParamsObjectType, key, value)
	if err != nil {
		return NewStatedbError(StorePdexv3ParamsError, err)
	}
	return nil
}

func GetPdexv3Params(stateDB *StateDB) (*Pdexv3Params, error) {
	key := GeneratePdexv3ParamsObjectKey()
	s, has, err := stateDB.getPdexv3ParamsByKey(key)
	if err != nil {
		return nil, NewStatedbError(GetPdexv3ParamsError, err)
	}
	if !has {
		return nil, NewStatedbError(GetPdexv3ParamsError, fmt.Errorf("could not found pDex v3 params in statedb"))
	}
	return s, nil
}

func StorePdexv3WaitingContributions(
	stateDB *StateDB,
	contributions map[string]rawdbv2.Pdexv3Contribution,
) error {
	for k, v := range contributions {
		key := GeneratePdexv3ContributionObjectKey(k)
		value := NewPdexv3ContributionStateWithValue(
			v, k,
		)
		err := stateDB.SetStateObject(Pdexv3ContributionObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePdexv3ContributionError, err)
		}
	}
	return nil
}

func DeletePdexv3WaitingContributions(
	stateDB *StateDB,
	pairHashes []string,
) error {
	for _, pairHash := range pairHashes {
		key := GeneratePdexv3ContributionObjectKey(pairHash)
		if !stateDB.MarkDeleteStateObject(Pdexv3ContributionObjectType, key) {
			fmt.Printf("Can't delete contributions with pair hash %v maybe 2 contributions have been matched\n", pairHash)
		}
	}
	return nil
}

func StorePdexv3Share(
	stateDB *StateDB, poolPairID string, nftID common.Hash,
	amount uint64, tradingFees map[common.Hash]uint64, lastLPFeesPerShare map[common.Hash]*big.Int,
) error {
	state := NewPdexv3ShareStateWithValue(nftID, amount, tradingFees, lastLPFeesPerShare)
	key := GeneratePdexv3ShareObjectKey(poolPairID, nftID.String())
	return stateDB.SetStateObject(Pdexv3ShareObjectType, key, state)
}

func StorePdexv3PoolPair(
	stateDB *StateDB,
	poolPairID string,
	poolPair rawdbv2.Pdexv3PoolPair,
) error {
	state := NewPdexv3PoolPairStateWithValue(poolPairID, poolPair)
	key := GeneratePdexv3PoolPairObjectKey(poolPairID)
	return stateDB.SetStateObject(Pdexv3PoolPairObjectType, key, state)
}

func StorePdexv3NftIDs(stateDB *StateDB, nftIDs map[string]uint64) error {
	for k, v := range nftIDs {
		key := GeneratePdexv3NftObjectKey(k)
		nftHash, err := common.Hash{}.NewHashFromStr(k)
		if err != nil {
			return err
		}
		value := NewPdexv3NftStateWithValue(*nftHash, v)
		err = stateDB.SetStateObject(Pdexv3NftObjectType, key, value)
		if err != nil {
			return NewStatedbError(StorePdexv3NftsError, err)
		}
	}
	return nil
}

func StorePdexv3StakingPool(
	stateDB *StateDB,
	tokenID string,
	totalAmount uint64,
	rewardsPerShare map[common.Hash]*big.Int,
) error {
	state := NewPdexv3StakingPoolStateWithValue(tokenID, totalAmount, rewardsPerShare)
	key := GeneratePdexv3StakingPoolObjectKey(tokenID)
	return stateDB.SetStateObject(Pdexv3StakingPoolObjectType, key, state)
}

func StorePdexv3Staker(stateDB *StateDB, stakingPoolID, nftID string, state *Pdexv3StakerState) error {
	key := GeneratePdexv3StakerObjectKey(stakingPoolID, nftID)
	return stateDB.SetStateObject(Pdexv3StakerObjectType, key, state)
}

func StorePdexv3Order(stateDB *StateDB, orderState Pdexv3OrderState) error {
	v := orderState.Value()
	key := GeneratePdexv3OrderObjectKey(orderState.PoolPairID(), v.Id())
	return stateDB.SetStateObject(Pdexv3OrderObjectType, key, &orderState)
}

func DeletePdexv3Order(stateDB *StateDB, pairID, orderID string) error {
	key := GeneratePdexv3OrderObjectKey(pairID, orderID)
	if !stateDB.MarkDeleteStateObject(Pdexv3OrderObjectType, key) {
		return fmt.Errorf("Cannot delete order with ID %v - %v", pairID, orderID)
	}
	return nil
}

func GetPdexv3WaitingContributions(stateDB *StateDB) (map[string]rawdbv2.Pdexv3Contribution, error) {
	prefixHash := GetPdexv3WaitingContributionsPrefix()
	return stateDB.iterateWithPdexv3Contributions(prefixHash)
}

func GetPdexv3PoolPairs(stateDB *StateDB) (map[string]Pdexv3PoolPairState, error) {
	prefixHash := GetPdexv3PoolPairsPrefix()
	return stateDB.iterateWithPdexv3PoolPairs(prefixHash)
}

func GetPdexv3Shares(stateDB *StateDB, poolPairID string) (
	map[string]Pdexv3ShareState, error,
) {
	prefixHash := generatePdexv3ShareObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3Shares(prefixHash)
}

func GetPdexv3NftIDs(stateDB *StateDB) (map[string]uint64, error) {
	prefixHash := GetPdexv3NftPrefix()
	return stateDB.iterateWithPdexv3Nfts(prefixHash)
}

func GetPdexv3Orders(stateDB *StateDB, poolPairID string) (
	map[string]Pdexv3OrderState,
	error,
) {
	prefixHash := generatePdexv3OrderObjectPrefix(poolPairID)
	return stateDB.iterateWithPdexv3Orders(prefixHash)
}

func GetPdexv3StakingPools(stateDB *StateDB) (map[string]*Pdexv3StakingPoolState, error) {
	prefixHash := GetPdexv3StakingPoolsPrefix()
	return stateDB.iterateWithPdexv3StakingPools(prefixHash)
}

func GetPdexv3Stakers(stateDB *StateDB, stakingPoolID string) (map[string]Pdexv3StakerState, error) {
	prefixHash := generatePdexv3StakerObjectPrefix(stakingPoolID)
	return stateDB.iterateWithPdexv3Stakers(prefixHash)
}