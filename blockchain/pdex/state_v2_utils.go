package pdex

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type Share struct {
	amount             uint64
	accessOTA          []byte
	tradingFees        map[common.Hash]uint64
	lastLPFeesPerShare map[common.Hash]*big.Int
}

func (share *Share) Amount() uint64 {
	return share.amount
}

func (share *Share) LastLPFeesPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range share.lastLPFeesPerShare {
		res[k] = big.NewInt(0).Set(v)
	}
	return res
}

func (share *Share) TradingFees() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range share.tradingFees {
		res[k] = v
	}
	return res
}

func NewShare() *Share {
	return &Share{
		amount:             0,
		tradingFees:        map[common.Hash]uint64{},
		lastLPFeesPerShare: map[common.Hash]*big.Int{},
	}
}

func NewShareWithValue(
	amount uint64,
	accessOTA []byte,
	tradingFees map[common.Hash]uint64,
	lastLPFeesPerShare map[common.Hash]*big.Int,
) *Share {
	return &Share{
		amount:             amount,
		accessOTA:          accessOTA,
		tradingFees:        tradingFees,
		lastLPFeesPerShare: lastLPFeesPerShare,
	}
}

func (share *Share) Clone() *Share {
	res := NewShare()
	res.accessOTA = share.accessOTA
	res.amount = share.amount
	res.tradingFees = map[common.Hash]uint64{}
	for k, v := range share.tradingFees {
		res.tradingFees[k] = v
	}
	res.lastLPFeesPerShare = map[common.Hash]*big.Int{}
	for k, v := range share.lastLPFeesPerShare {
		res.lastLPFeesPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (share *Share) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Amount             uint64                   `json:"Amount"`
		AccessOTA          []byte                   `json:"AccessOTA,omitempty"`
		TradingFees        map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{
		AccessOTA:          share.accessOTA,
		Amount:             share.amount,
		TradingFees:        share.tradingFees,
		LastLPFeesPerShare: share.lastLPFeesPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (share *Share) UnmarshalJSON(data []byte) error {
	temp := struct {
		Amount             uint64                   `json:"Amount"`
		AccessOTA          []byte                   `json:"AccessOTA,omitempty"`
		TradingFees        map[common.Hash]uint64   `json:"TradingFees"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastLPFeesPerShare"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	share.accessOTA = temp.AccessOTA
	share.amount = temp.Amount
	share.tradingFees = temp.TradingFees
	share.lastLPFeesPerShare = temp.LastLPFeesPerShare
	return nil
}

func (share *Share) getDiff(
	compareShare *Share,
	shareChange *v2utils.ShareChange,
) *v2utils.ShareChange {
	newShareChange := shareChange
	if newShareChange == nil {
		newShareChange = v2utils.NewShareChange()
	}
	if compareShare == nil {
		newShareChange.IsChanged = true
		for tokenID := range share.tradingFees {
			newShareChange.TradingFees[tokenID.String()] = true
		}
		for tokenID := range share.lastLPFeesPerShare {
			newShareChange.LastLPFeesPerShare[tokenID.String()] = true
		}
	} else {
		if share.amount != compareShare.amount || !bytes.Equal(share.accessOTA, compareShare.accessOTA) {
			newShareChange.IsChanged = true
		}
		newShareChange.TradingFees = v2utils.GetChangedElementsFromMapUint64(share.tradingFees, compareShare.tradingFees)
		newShareChange.LastLPFeesPerShare = v2utils.GetChangedElementsFromMapBigInt(share.lastLPFeesPerShare, compareShare.lastLPFeesPerShare)
	}

	return newShareChange
}

func (share *Share) setAccessOTA(accessOTA []byte) {
	share.accessOTA = accessOTA
}

func (share *Share) isValidAccessOTA(burntOTA metadataPdexv3.AccessOTA) (bool, error) {
	if !bytes.Equal(share.accessOTA, burntOTA.ToBytesS()) {
		return false, errors.New("Not valid access OTA")
	}
	return true, nil
}

func (share *Share) isEmpty() bool {
	if share.amount != 0 {
		return false
	}
	for _, v := range share.tradingFees {
		if v != 0 {
			return false
		}
	}
	//TODO: @tin verify with last liquidity mining later
	return true
}

type Staker struct {
	liquidity           uint64
	accessOTA           []byte
	rewards             map[common.Hash]uint64
	lastRewardsPerShare map[common.Hash]*big.Int
}

func (staker *Staker) Liquidity() uint64 {
	return staker.liquidity
}

func (staker *Staker) LastRewardsPerShare() map[common.Hash]*big.Int {
	res := make(map[common.Hash]*big.Int)
	for k, v := range staker.lastRewardsPerShare {
		res[k] = big.NewInt(0).Set(v)
	}
	return res
}

func (staker *Staker) Rewards() map[common.Hash]uint64 {
	res := make(map[common.Hash]uint64)
	for k, v := range staker.rewards {
		res[k] = v
	}
	return res
}

func (staker *Staker) AccessOTA() []byte {
	return staker.accessOTA
}

func (staker *Staker) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Liquidity           uint64                   `json:"Liquidity"`
		AccessOTA           []byte                   `json:"AccessOTA,omitempty"`
		Rewards             map[common.Hash]uint64   `json:"Rewards"`
		LastRewardsPerShare map[common.Hash]*big.Int `json:"LastRewardsPerShare"`
	}{
		Liquidity:           staker.liquidity,
		AccessOTA:           staker.accessOTA,
		Rewards:             staker.rewards,
		LastRewardsPerShare: staker.lastRewardsPerShare,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (staker *Staker) UnmarshalJSON(data []byte) error {
	temp := struct {
		Liquidity          uint64                   `json:"Liquidity"`
		AccessOTA          []byte                   `json:"AccessOTA,omitempty"`
		Rewards            map[common.Hash]uint64   `json:"Rewards"`
		LastLPFeesPerShare map[common.Hash]*big.Int `json:"LastRewardsPerShare"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	staker.accessOTA = temp.AccessOTA
	staker.liquidity = temp.Liquidity
	staker.rewards = temp.Rewards
	staker.lastRewardsPerShare = temp.LastLPFeesPerShare
	return nil
}

func NewStaker() *Staker {
	return &Staker{
		rewards:             make(map[common.Hash]uint64),
		lastRewardsPerShare: make(map[common.Hash]*big.Int),
	}
}

func NewStakerWithValue(
	liquidity uint64, accessOTA []byte,
	rewards map[common.Hash]uint64, lastLPFeesPerShare map[common.Hash]*big.Int,
) *Staker {
	return &Staker{
		accessOTA:           accessOTA,
		liquidity:           liquidity,
		rewards:             rewards,
		lastRewardsPerShare: lastLPFeesPerShare,
	}
}

func (staker *Staker) Clone() *Staker {
	res := NewStaker()
	res.liquidity = staker.liquidity
	res.accessOTA = staker.accessOTA
	for k, v := range staker.rewards {
		res.rewards[k] = v
	}
	for k, v := range staker.lastRewardsPerShare {
		res.lastRewardsPerShare[k] = new(big.Int).Set(v)
	}
	return res
}

func (staker *Staker) getDiff(
	compareStaker *Staker, stakerChange *v2utils.StakerChange,
) *v2utils.StakerChange {
	newStakerChange := stakerChange
	if compareStaker == nil {
		newStakerChange.IsChanged = true
		for tokenID := range staker.rewards {
			newStakerChange.Rewards[tokenID.String()] = true
		}
		for tokenID := range staker.lastRewardsPerShare {
			newStakerChange.LastRewardsPerShare[tokenID.String()] = true
		}
	} else {
		if staker.liquidity != compareStaker.liquidity || !bytes.Equal(staker.accessOTA, compareStaker.accessOTA) {
			newStakerChange.IsChanged = true
		}
		newStakerChange.LastRewardsPerShare = v2utils.GetChangedElementsFromMapBigInt(staker.lastRewardsPerShare, compareStaker.lastRewardsPerShare)
		newStakerChange.Rewards = v2utils.GetChangedElementsFromMapUint64(staker.rewards, compareStaker.rewards)
	}
	return newStakerChange
}

func (staker *Staker) setAccessOTA(accessOTA []byte) {
	staker.accessOTA = accessOTA
}

func (staker *Staker) isValidAccessOTA(burntOTA metadataPdexv3.AccessOTA) (bool, error) {
	if !bytes.Equal(staker.accessOTA, burntOTA.ToBytesS()) {
		return false, errors.New("Not valid access OTA")
	}
	return true, nil
}

func (staker *Staker) isEmpty() bool {
	if staker.liquidity != 0 {
		return false
	}
	for _, v := range staker.rewards {
		if v != 0 {
			return false
		}
	}
	return true
}

func addStakingPoolState(
	stakingPoolStates map[string]*StakingPoolState, stakingPoolIDs map[string]uint,
) map[string]*StakingPoolState {
	for k := range stakingPoolIDs {
		if stakingPoolStates[k] == nil {
			stakingPoolStates[k] = NewStakingPoolState()
		}
	}
	return stakingPoolStates
}

func (share *Share) updateToDB(
	env StateEnvironment, poolPairID, nftID string, shareChange *v2utils.ShareChange,
) error {
	if shareChange.IsChanged {
		nftID, err := common.Hash{}.NewHashFromStr(nftID)
		err = statedb.StorePdexv3Share(
			env.StateDB(), poolPairID,
			*nftID,
			share.amount, share.accessOTA,
		)
		if err != nil {
			return err
		}
	}
	for tokenID, isChanged := range shareChange.TradingFees {
		if isChanged {
			tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
			if err != nil {
				return err
			}
			if tradingFee, found := share.tradingFees[*tokenHash]; found {
				err := statedb.StorePdexv3ShareTradingFee(
					env.StateDB(), poolPairID, nftID,
					statedb.NewPdexv3ShareTradingFeeStateWithValue(*tokenHash, tradingFee),
				)
				if err != nil {
					return err
				}
			} else {
				err := statedb.DeletePdexv3ShareTradingFee(
					env.StateDB(), poolPairID, nftID, tokenID,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	for tokenID, isChanged := range shareChange.LastLPFeesPerShare {
		if isChanged {
			tokenHash, err := common.Hash{}.NewHashFromStr(tokenID)
			if err != nil {
				return err
			}
			if lastLPFeesPerShare, found := share.lastLPFeesPerShare[*tokenHash]; found {
				err := statedb.StorePdexv3ShareLastLpFeePerShare(
					env.StateDB(), poolPairID, nftID,
					statedb.NewPdexv3ShareLastLpFeePerShareStateWithValue(*tokenHash, lastLPFeesPerShare),
				)
				if err != nil {
					return err
				}
			} else {
				err := statedb.DeletePdexv3ShareLastLpFeePerShare(
					env.StateDB(), poolPairID, nftID, tokenID,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (share *Share) deleteFromDB(
	env StateEnvironment,
	poolPairID, nftID string,
	shareChange *v2utils.ShareChange,
) error {
	err := statedb.DeletePdexv3Share(env.StateDB(), poolPairID, nftID)
	if err != nil {
		return err
	}
	return nil
}

type Reward map[common.Hash]uint64 // tokenID -> amount

type OrderReward struct {
	accessOTA          []byte
	uncollectedRewards Reward
}

func (orderReward *OrderReward) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AccessOTA          []byte `json:"AccessOTA,omitempty"`
		UncollectedRewards Reward `json:"UncollectedRewards"`
	}{
		AccessOTA:          orderReward.accessOTA,
		UncollectedRewards: orderReward.uncollectedRewards,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (orderReward *OrderReward) UnmarshalJSON(data []byte) error {
	temp := struct {
		AccessOTA          []byte `json:"AccessOTA,omitempty"`
		UncollectedRewards Reward `json:"UncollectedRewards"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	orderReward.accessOTA = temp.AccessOTA
	orderReward.uncollectedRewards = temp.UncollectedRewards
	return nil
}

func (orderReward *OrderReward) AccessOTA() []byte {
	return orderReward.accessOTA
}

func (orderReward *OrderReward) UncollectedRewards() Reward {
	res := Reward{}
	for k, v := range orderReward.uncollectedRewards {
		res[k] = v
	}
	return res
}

func (orderReward *OrderReward) AddReward(tokenID common.Hash, amount uint64) {
	oldAmount := uint64(0)
	if _, ok := orderReward.uncollectedRewards[tokenID]; ok {
		oldAmount = orderReward.uncollectedRewards[tokenID]
	}
	orderReward.uncollectedRewards[tokenID] = oldAmount + amount
}

func NewOrderReward() *OrderReward {
	return &OrderReward{
		uncollectedRewards: make(map[common.Hash]uint64),
	}
}

func NewOrderRewardWithValue(
	accessOTA []byte,
	uncollectedRewards map[common.Hash]uint64,
) *OrderReward {
	return &OrderReward{
		accessOTA:          accessOTA,
		uncollectedRewards: uncollectedRewards,
	}
}

func (orderReward *OrderReward) Clone() *OrderReward {
	res := NewOrderReward()
	for k, v := range orderReward.uncollectedRewards {
		res.uncollectedRewards[k] = v
	}
	res.accessOTA = orderReward.accessOTA
	return res
}

func (orderReward *OrderReward) getDiff(
	compareOrderReward *OrderReward,
	orderRewardChange *v2utils.OrderRewardChange,
) *v2utils.OrderRewardChange {
	newOrderRewardChange := orderRewardChange
	if newOrderRewardChange == nil {
		newOrderRewardChange = v2utils.NewOrderRewardChange()
	}
	if compareOrderReward == nil {
		for tokenID := range orderReward.uncollectedRewards {
			newOrderRewardChange.UncollectedReward[tokenID.String()] = true
		}
		newOrderRewardChange.IsChanged = true
	} else {
		if !bytes.Equal(orderReward.accessOTA, compareOrderReward.accessOTA) {
			newOrderRewardChange.IsChanged = true
		}
		newOrderRewardChange.UncollectedReward = v2utils.GetChangedElementsFromMapUint64(orderReward.uncollectedRewards, compareOrderReward.uncollectedRewards)
	}
	return newOrderRewardChange
}

type MakingVolume struct {
	volume map[string]*big.Int // nftID -> amount
}

func (makingVolume *MakingVolume) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		Volume map[string]*big.Int `json:"Volume"`
	}{
		Volume: makingVolume.volume,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (makingVolume *MakingVolume) UnmarshalJSON(data []byte) error {
	temp := struct {
		Volume map[string]*big.Int `json:"Volume"`
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	makingVolume.volume = temp.Volume
	return nil
}

func (makingVolume *MakingVolume) AddVolume(nftID string, amount *big.Int) {
	oldAmount := big.NewInt(0)
	if _, ok := makingVolume.volume[nftID]; ok {
		oldAmount = makingVolume.volume[nftID]
	}
	makingVolume.volume[nftID] = big.NewInt(0).Add(oldAmount, amount)
}

func NewMakingVolume() *MakingVolume {
	return &MakingVolume{
		volume: make(map[string]*big.Int),
	}
}

func (makingVolume *MakingVolume) Clone() *MakingVolume {
	res := NewMakingVolume()
	for k, v := range makingVolume.volume {
		res.volume[k] = new(big.Int).Set(v)
	}
	return res
}

func (makingVolume *MakingVolume) getDiff(
	compareMakingVolume *MakingVolume,
	makingVolumeChange *v2utils.MakingVolumeChange,
) *v2utils.MakingVolumeChange {
	newMakingVolumeChange := makingVolumeChange
	if newMakingVolumeChange == nil {
		newMakingVolumeChange = v2utils.NewMakingVolumeChange()
	}
	if compareMakingVolume == nil {
		for nftID := range makingVolume.volume {
			newMakingVolumeChange.Volume[nftID] = true
		}
	} else {
		newMakingVolumeChange.Volume = v2utils.GetChangedElementsFromMapStringBigInt(makingVolume.volume, compareMakingVolume.volume)
	}
	return newMakingVolumeChange
}
