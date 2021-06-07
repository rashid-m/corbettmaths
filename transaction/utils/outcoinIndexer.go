package utils

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/privacy"
)

type CoinIndexer struct {
	sem            *semaphore.Weighted
	numWorkers     int
	ManagedOTAKeys *sync.Map
	db             incdb.Database
	accessTokens   map[string][]privacy.OTAKey
	idxQueue       []IndexParams
	IdxChan        chan IndexParams
	statusChan     chan JobStatus
}

// NewOutCoinIndexer creates a new full node's caching instance for faster output coin retrieval.
func NewOutCoinIndexer(numWorkers int64, db incdb.Database) (*CoinIndexer, error) {
	// view key :-> indexing status
	// 2 means indexer finished
	// while < 2 : `balance` & `createTx` RPCs are not available
	// viewKey map will be loaded from db
	sem := semaphore.NewWeighted(numWorkers)

	m := &sync.Map{}
	// load from db once after startup
	loadedKeysRaw, err := rawdbv2.GetIndexedOTAKeys(db)
	if err == nil {
		for _, b := range loadedKeysRaw {
			var temp [64]byte
			copy(temp[:], b[0:64])
			m.Store(temp, 2)
		}
	}
	return &CoinIndexer{sem: sem, ManagedOTAKeys: m, db: db, accessTokens: nil}, nil
}

// IsAuthorizedUser checks if a user is authorized to use the enhanced cache.
func (ci *CoinIndexer) IsAuthorizedUser(accessToken string, otaKey privacy.OTAKey) bool {
	if otaList, ok := ci.accessTokens[accessToken]; ok { // valid access token
		if otaList == nil {
			otaList = make([]privacy.OTAKey, 0)
		}

		otaList = append(otaList, otaKey)

		return true
	}

	return false
}

// RemoveOTAKey removes an OTAKey from the cached database.
//
// TODO: remove cached output coins, access token.
func (ci *CoinIndexer) RemoveOTAKey(accessToken string, otaKey privacy.OTAKey) error {
	keyBytes := OTAKeyToRaw(otaKey)
	err := rawdbv2.DeleteIndexedOTAKey(ci.db, keyBytes[:])
	if err != nil {
		return err
	}
	ci.ManagedOTAKeys.Delete(keyBytes)

	return nil
}

// AddOTAKey adds a new OTAKey to the cache list.
func (ci *CoinIndexer) AddOTAKey(otaKey privacy.OTAKey) error {
	keyBytes := OTAKeyToRaw(otaKey)
	err := rawdbv2.StoreIndexedOTAKey(ci.db, keyBytes[:])
	if err != nil {
		return err
	}
	ci.ManagedOTAKeys.Store(keyBytes, 2)
	return nil
}

// IsQueueFull checks if the current indexing queue is full.
func (ci *CoinIndexer) IsQueueFull() bool {
	return len(ci.idxQueue) >= ci.numWorkers
}

// ReIndexOutCoin re-scans all output coins from fromHeight to toHeight and adds them to the cache if they belong to vk.
func (ci *CoinIndexer) ReIndexOutCoin(idxParams IndexParams) {
	status := JobStatus{
		otaKey: idxParams.OTAKey,
		err:    nil,
	}

	vkb := OTAKeyToRaw(idxParams.OTAKey)
	Logger.Log.Infof("[CoinIndexer] Re-index output coins for key %x", idxParams.OTAKey)
	keyExists, processing := ci.HasOTAKey(vkb)
	if keyExists {
		if processing == 1 {
			Logger.Log.Errorf("ota key %v is being processed", idxParams.OTAKey)
			ci.statusChan <- status
			return
		}
		// resetting entries for this key is reserved for debugging RPCs
		if processing == 2 && !idxParams.IsReset {
			Logger.Log.Errorf("ota key %v has been processed and isReset = false", idxParams.OTAKey)

			ci.statusChan <- status
			return
		}
	}
	ci.ManagedOTAKeys.Store(vkb, 1)
	defer func() {
		if r := recover(); r != nil {
			Logger.Log.Errorf("[CoinIndexer] Recovered from: %v\n", r)
		}
		if exists, processing := ci.HasOTAKey(vkb); exists && processing == 1 {
			ci.ManagedOTAKeys.Delete(vkb)
		}
	}()
	var allOutputCoins []privacy.Coin

	for height := idxParams.FromHeight; height <= idxParams.ToHeight; {
		nextHeight := height + MaxOutcoinQueryInterval

		ctx, cancel := context.WithTimeout(context.Background(), OutcoinReindexerTimeout*time.Second)
		defer cancel()

		err := ci.sem.Acquire(ctx, 1)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] semaphore acquiring error: %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		// query token output coins
		currentOutputCoinsToken, err := QueryDbCoinVer2(idxParams.OTAKey, idxParams.ShardID, &common.ConfidentialAssetID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying token coins from db - %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		// query PRV output coins
		currentOutputCoinsPRV, err := QueryDbCoinVer2(idxParams.OTAKey, idxParams.ShardID, &common.PRVCoinID, height, nextHeight-1, idxParams.TxDb, getCoinFilterByOTAKey())
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] Error while querying PRV coins from db - %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}

		ci.sem.Release(1)

		Logger.Log.Infof("[CoinIndexer] Key %x - %d to %d : found %d PRV + %d pToken coins\n", vkb, height, nextHeight-1, len(currentOutputCoinsPRV), len(currentOutputCoinsToken))

		allOutputCoins = append(allOutputCoins, append(currentOutputCoinsToken, currentOutputCoinsPRV...)...)
		height = nextHeight
	}

	// write
	err := rawdbv2.StoreIndexedOTAKey(ci.db, vkb[:])
	if err == nil {
		err = ci.StoreIndexedOutputCoins(idxParams.OTAKey, allOutputCoins, idxParams.ShardID)
		if err != nil {
			Logger.Log.Errorf("[CoinIndexer] StoreIndexedOutCoins error: %v\n", err)

			status.err = err
			ci.statusChan <- status
			return
		}
	} else {
		Logger.Log.Errorf("[CoinIndexer] StoreIndexedOTAKey error: %v\n", err)

		status.err = err
		ci.statusChan <- status
		return
	}

	ci.ManagedOTAKeys.Store(vkb, 2)
	Logger.Log.Infof("[CoinIndexer] Indexing complete for key %x\n", vkb)

	status.err = nil
	ci.statusChan <- status
	return
}

func (ci *CoinIndexer) GetIndexedOutCoin(viewKey privacy.OTAKey, tokenID *common.Hash, txDb *statedb.StateDB, shardID byte) ([]privacy.Coin, int, error) {
	vkb := OTAKeyToRaw(viewKey)
	Logger.Log.Infof("Retrieve re-indexed coins for %x from db %v", vkb, ci.db)
	_, processing := ci.HasOTAKey(vkb)
	if processing == 1 {
		return nil, 1, errors.New(fmt.Sprintf("View Key %x not ready : Sync still in progress", viewKey))
	}
	if processing == 0 {
		// this is a new view key
		return nil, 0, errors.New(fmt.Sprintf("View Key %x not synced", viewKey))
	}
	ocBytes, err := rawdbv2.GetOutCoinsByIndexedOTAKey(ci.db, common.ConfidentialAssetID, shardID, vkb[:])
	if err != nil {
		return nil, 0, err
	}
	params := make(map[string]interface{})
	params["otaKey"] = viewKey
	params["tokenID"] = tokenID
	filter := GetCoinFilterByOTAKeyAndToken()
	var result []privacy.Coin
	for _, cb := range ocBytes {
		temp := &privacy.CoinV2{}
		err := temp.SetBytes(cb)
		if err != nil {
			return nil, 0, errors.New("Coin by View Key storage is corrupted")
		}
		if filter(temp, params) {
			// eliminate forked coins
			if dbHasOta, _, err := statedb.HasOnetimeAddress(txDb, *tokenID, temp.GetPublicKey().ToBytesS()); dbHasOta && err == nil {
				result = append(result, temp)
			}
		}
	}
	return result, 2, nil
}

func (ci *CoinIndexer) StoreIndexedOutputCoins(viewKey privacy.OTAKey, outputCoins []privacy.Coin, shardID byte) error {
	var ocBytes [][]byte
	for _, c := range outputCoins {
		ocBytes = append(ocBytes, c.Bytes())
	}
	vkb := OTAKeyToRaw(viewKey)
	Logger.Log.Infof("Store %d indexed coins to db %v", len(ocBytes), ci.db)
	// all token and PRV coins are grouped together; match them to desired tokenID upon retrieval
	err := rawdbv2.StoreIndexedOutCoins(ci.db, common.ConfidentialAssetID, vkb[:], ocBytes, shardID)
	if err != nil {
		return err
	}
	return err
}

func (ci *CoinIndexer) Serve() {
	numWorking := 0
	for {
		select {
		case status := <-ci.statusChan:
			numWorking--
			if status.err != nil {
				Logger.Log.Errorf("IndexOutCoin for otaKey %v failed: %v\n", status.otaKey, status.err)
				ci.RemoveOTAKey("", status.otaKey)
			} else {
				Logger.Log.Infof("Finished indexing output coins for otaKey: %v\n", status.otaKey)
			}

		case idxParams := <-ci.IdxChan:
			Logger.Log.Infof("New authorized OTAKey received: %v\n", idxParams.OTAKey)
			//do something
			ci.idxQueue = append(ci.idxQueue, idxParams)
		default:
			if numWorking < ci.numWorkers && len(ci.idxQueue) > 0 {
				numWorking++

				idxParams := ci.idxQueue[0]
				ci.idxQueue = ci.idxQueue[1:]

				go ci.ReIndexOutCoin(idxParams)
			}
		}
	}
}

func (ci *CoinIndexer) HasOTAKey(k [64]byte) (bool, int) {
	var result int
	val, ok := ci.ManagedOTAKeys.Load(k)
	if ok {
		result, ok = val.(int)
	}
	return ok, result
}
