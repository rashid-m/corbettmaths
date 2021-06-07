package utils

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type JobStatus struct {
	otaKey privacy.OTAKey
	err    error
}

type IndexParams struct {
	FromHeight uint64
	ToHeight   uint64
	OTAKey     privacy.OTAKey
	TxDb       *statedb.StateDB
	ShardID    byte
	IsReset    bool
}

type CoinMatcher func(*privacy.CoinV2, map[string]interface{}) bool

// getCoinFilterByOTAKey returns a functions that filters if an output coin belongs to an OTAKey.
func getCoinFilterByOTAKey() CoinMatcher {
	return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool {
		entry, exists := kvargs["otaKey"]
		if !exists {
			return false
		}
		vk, ok := entry.(privacy.OTAKey)
		if !ok {
			return false
		}
		ks := &incognitokey.KeySet{}
		ks.OTAKey = vk

		pass, _ := c.DoesCoinBelongToKeySet(ks)
		return pass
	}
}

// GetCoinFilterByOTAKeyAndToken returns a functions that filters if an output coin is of a specific token and belongs to an OTAKey.
func GetCoinFilterByOTAKeyAndToken() CoinMatcher {
	return func(c *privacy.CoinV2, kvargs map[string]interface{}) bool {
		entry, exists := kvargs["otaKey"]
		if !exists {
			return false
		}
		vk, ok := entry.(privacy.OTAKey)
		if !ok {
			return false
		}
		entry, exists = kvargs["tokenID"]
		if !exists {
			return false
		}
		tokenID, ok := entry.(*common.Hash)
		if !ok {
			return false
		}
		ks := &incognitokey.KeySet{}
		ks.OTAKey = vk

		if pass, sharedSecret := c.DoesCoinBelongToKeySet(ks); pass {
			pass, _ = c.ValidateAssetTag(sharedSecret, tokenID)
			return pass
		}
		return false
	}
}

func GetNextLowerHeight(upper, floor uint64) uint64 {
	if upper > MaxOutcoinQueryInterval+floor {
		return upper - MaxOutcoinQueryInterval
	}
	return floor
}

func OTAKeyToRaw(vk privacy.OTAKey) [64]byte {
	var result [64]byte
	copy(result[0:32], vk.GetOTASecretKey().ToBytesS())
	copy(result[32:64], vk.GetPublicSpend().ToBytesS())
	return result
}

func OTAKeyFromRaw(b [64]byte) privacy.OTAKey {
	result := &privacy.OTAKey{}
	result.SetOTASecretKey(b[0:32])
	result.SetPublicSpend(b[32:64])
	return *result
}

func QueryDbCoinVer1(pubKey []byte, shardID byte, tokenID *common.Hash, db *statedb.StateDB) ([]privacy.Coin, error) {
	outCoinsBytes, err := statedb.GetOutcoinsByPubkey(db, *tokenID, pubKey, shardID)
	if err != nil {
		Logger.Log.Error("GetOutcoinsBytesByKeyset Get by PubKey", err)
		return nil, err
	}
	var outCoins []privacy.Coin
	for _, item := range outCoinsBytes {
		outCoin := &privacy.CoinV1{}
		err := outCoin.SetBytes(item)
		if err != nil {
			Logger.Log.Errorf("Cannot create coin from byte %v", err)
			return nil, err
		}
		outCoins = append(outCoins, outCoin)
	}
	return outCoins, nil
}

func QueryDbCoinVer2(otaKey privacy.OTAKey, shardID byte, tokenID *common.Hash, shardHeight, destHeight uint64, db *statedb.StateDB, filters ...CoinMatcher) ([]privacy.Coin, error) {
	var outCoins []privacy.Coin
	// avoid overlap; unless lower height is 0
	start := shardHeight + 1
	if shardHeight == 0 {
		start = 0
	}
	for height := start; height <= destHeight; height += 1 {
		currentHeightCoins, err := statedb.GetOTACoinsByHeight(db, *tokenID, shardID, height)
		if err != nil {
			Logger.Log.Error("Get outcoins ver 2 bytes by keyset get by height", err)
			return nil, err
		}
		params := make(map[string]interface{})
		params["otaKey"] = otaKey
		params["db"] = db
		params["tokenID"] = tokenID
		for _, coinBytes := range currentHeightCoins {
			cv2 := &privacy.CoinV2{}
			err := cv2.SetBytes(coinBytes)
			if err != nil {
				Logger.Log.Error("Get outcoins ver 2 from bytes", err)
				return nil, err
			}
			pass := true
			for _, f := range filters {
				if !f(cv2, params) {
					pass = false
				}
			}
			if pass {
				outCoins = append(outCoins, cv2)
			}
		}
	}
	return outCoins, nil
}