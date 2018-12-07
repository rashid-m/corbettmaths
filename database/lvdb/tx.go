package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/transaction"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/database"

	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"math/big"
)

// StoreSerialNumbers - store list serialNumbers by chainID
func (db *db) StoreSerialNumbers(serialNumber []byte, chainId byte) error {
	key := db.getKey(string(serialNumbersPrefix), "")
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrayData [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrayData); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	arrayData = append(arrayData, serialNumber)
	b, err := json.Marshal(arrayData)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

// FetchSerialNumbers - Get list SerialNumbers by chainID
func (db *db) FetchSerialNumbers(chainID byte) ([][]byte, error) {
	key := db.getKey(string(serialNumbersPrefix), "")
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrayData [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrayData); err != nil {
			return make([][]byte, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return arrayData, nil
}

// HasSerialNumber - Check serialNumber in list SerialNumbers by chainID
func (db *db) HasSerialNumber(serialNumber []byte, chainID byte) (bool, error) {
	listSerialNumbers, err := db.FetchSerialNumbers(chainID)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listSerialNumbers {
		if bytes.Equal(item, serialNumber) {
			return true, nil
		}
	}
	return false, nil
}

// CleanSerialNumbers - clear all list serialNumber in DB
func (db *db) CleanSerialNumbers() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(serialNumbersPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

// StoreCommitments - store list commitments by chainID
func (db *db) StoreCommitments(commitments []byte, chainId byte) error {
	key := db.getKey(string(commitmentsPrefix), "")
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	txs = append(txs, commitments)
	b, err := json.Marshal(txs)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

// FetchCommitments - Get list commitments by chainID
func (db *db) FetchCommitments(chainId byte) ([][]byte, error) {
	key := db.getKey(string(commitmentsPrefix), "")
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	return txs, nil
}

// HasCommitment - Check commitment in list commitments by chainID
func (db *db) HasCommitment(commitment []byte, chainId byte) (bool, error) {
	listCommitments, err := db.FetchCommitments(chainId)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listCommitments {
		if bytes.Equal(item, commitment) {
			return true, nil
		}
	}
	return false, nil
}

// CleanCommitments - clear all list commitments in DB
func (db *db) CleanCommitments() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(commitmentsPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

// StoreSerialNumbers - store list serialNumbers by chainID
func (db *db) StoreSNDerivators(data big.Int, chainID byte) error {
	key := db.getKey(string(snderivatorsPrefix), "")
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var arrData []big.Int
	if len(res) > 0 {
		if err := json.Unmarshal(res, &arrData); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	arrData = append(arrData, data)
	b, err := json.Marshal(arrData)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

// FetchSerialNumbers - Get list all SnDerivators by chainID
func (db *db) FetchSNDerivator(chainID byte) ([]big.Int, error) {
	key := db.getKey(string(snderivatorsPrefix), "")
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([]big.Int, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs []big.Int
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([]big.Int, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return txs, nil
}

// HasSNDerivator - Check SnDerivator in list SnDerivators by chainID
func (db *db) HasSNDerivator(data big.Int, chainID byte) (bool, error) {
	listSNDDerivators, err := db.FetchSNDerivator(chainID)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listSNDDerivators {
		if item.Cmp(&data) == 0 {
			return true, nil
		}
	}
	return false, nil
}

// CleanCommitments - clear all list commitments in DB
func (db *db) CleanSNDerivator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(snderivatorsPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

// StoreFeeEstimator - Store data for FeeEstimator object
func (db *db) StoreFeeEstimator(val []byte, chainId byte) error {
	if err := db.put(append(feeEstimator, chainId), val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

// GetFeeEstimator - Get data for FeeEstimator object as a json in byte format
func (db *db) GetFeeEstimator(chainId byte) ([]byte, error) {
	b, err := db.lvdb.Get(append(feeEstimator, chainId), nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return b, err
}

// CleanFeeEstimator - Clear FeeEstimator
func (db *db) CleanFeeEstimator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(feeEstimator), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

/*
	StoreTransactionIndex
	Store tx detail location
  Key: prefixTx-txHash
	H: blockHash-blockIndex
*/
func (db *db) StoreTransactionIndex(txId *common.Hash, blockHash *common.Hash, index int) error {
	key := string(transactionKeyPrefix) + txId.String()
	value := blockHash.String() + string(splitter) + strconv.Itoa(index)
	fmt.Println("Key in StoreTransactionIndex", key)
	fmt.Println("H in StoreTransactionIndex", value)
	if err := db.lvdb.Put([]byte(key), []byte(value), nil); err != nil {
		return err
	}

	return nil
}

/*
  Get Transaction by ID
*/

func (db *db) GetTransactionIndexById(txId *common.Hash) (*common.Hash, int, error) {
	fmt.Println("TxID in GetTransactionById", txId.String())
	key := string(transactionKeyPrefix) + txId.String()
	_, err := db.hasValue([]byte(key))
	if err != nil {
		fmt.Println("ERROR in finding transaction id", txId.String(), err)
		return nil, -1, err
	}

	res, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return nil, -1, err;
	}
	reses := strings.Split(string(res), (string(splitter)))
	hash, err := common.Hash{}.NewHashFromStr(reses[0])
	if err != nil {
		return nil, -1, err;
	}
	index, err := strconv.Atoi(reses[1])
	if err != nil {
		return nil, -1, err;
	}
	fmt.Println("BlockHash", hash, "Transaction index", index)
	return hash, index, nil
}

/*
	Store Transaction in Light mode
	1. Key -> value : prefix(privateky)privateKey-[-]-chainId-[-]-(999999999 - blockHeight)-[-]-(999999999 - txIndex) 		-> 		tx
	2. Key -> value :							prefix(transaction)txHash 												->  	privateKey-chainId-blockHeight-txIndex

*/
func (db *db) StoreTransactionLightMode(privateKey *privacy.SpendingKey, chainId byte, blockHeight int32, txIndex int, unspentTx *transaction.Tx) error {
	tempChainId := []byte{}
	tempChainId = append(tempChainId, chainId)
	temp3ChainId := int(chainId)
	temp2ChainId := string(int(chainId))
	fmt.Println("StoreTransactionLightMode", privateKey, temp3ChainId, temp2ChainId, blockHeight, txIndex)
	reverseBlockHeight := make([]byte, 4)
	binary.LittleEndian.PutUint32(reverseBlockHeight, uint32(bigNumber-blockHeight))

	// Uncomment this to test little endian encoding and decoding
	/*
	buf := bytes.NewBuffer(reverseBlockHeight)
	var temp uint32
	err1 := binary.Read(buf, binary.LittleEndian, &temp)
	if err1 != nil {
		return err1
	}
	fmt.Println("Testing encoding and decoding uint32 little endian", uint32(999999999-temp), blockHeight)
	*/

	//fmt.Println("StoreTransactionLightMode reverseBlockHeight in byte", reverseBlockHeight, []byte(string(reverseBlockHeight)))

	reverseTxIndex := make([]byte, 4)
	binary.LittleEndian.PutUint32(reverseTxIndex, uint32(bigNumberTx-int32(txIndex)))

	key1 := string(privateKeyPrefix) + privateKey.String() + string(splitter) + string(int(chainId)) + string(splitter) + string(reverseBlockHeight) + string(splitter) + string(reverseTxIndex)
	key2 := string(transactionKeyPrefix) + unspentTx.Hash().String()

	if ok, _ := db.hasValue([]byte(key1)); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("tx %s already exists", key1))
	}
	if ok, _ := db.hasValue([]byte(key2)); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("tx %s already exists", key2))
	}

	value, err := json.Marshal(unspentTx)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.put([]byte(key1), value); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.put([]byte(key2), []byte(key1)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}

	fmt.Println("Storing Transaction in light mode: txLocation -> tx", key1, unspentTx)
	fmt.Println("Storing Transaction in light mode: txHash -> txLocation", key2, key1)
	return nil
}

/*
	Get Transaction in Light mode
	Get transaction by prefix(privateKey)privateKey, this prefix help to get all transaction belong to that privatekey
	1. Key -> value : prefix(privateky)-privateKey-chainId-(999999999 - blockHeight)-(999999999 - txIndex) 		-> 		tx

*/
func (db *db) GetTransactionLightModeByPrivateKey(privateKey *privacy.SpendingKey) (map[byte][]transaction.Tx, error) {
	prefix := []byte(string(privateKeyPrefix) + privateKey.String())
	iter := db.lvdb.NewIterator(util.BytesPrefix(prefix), nil)
	results := make(map[byte][]transaction.Tx)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Println("GetTransactionLightModeByPrivateKey, key", string(key))
		reses := strings.Split(string(key), string(splitter))
		tempChainId, _ := strconv.Atoi(reses[2])
		chainId := byte(tempChainId)
		fmt.Println("GetTransactionLightModeByPrivateKey, chainId", chainId)
		tx := transaction.Tx{}
		err := json.Unmarshal(value, &tx)
		if err != nil {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
		}
		results[chainId] = append(results[chainId], tx)
	}
	iter.Release()
	return results, nil
}

/*
	Key: transactionPrefix-txHash
  H: txLocation
  tx: tx object in byte
*/
func (db *db) GetTransactionLightModeByHash(txId *common.Hash) ([]byte, []byte, error) {
	key := string(transactionKeyPrefix) + txId.String()
	fmt.Println("GetTransactionLightModeByHash - key", key)
	_, err := db.hasValue([]byte(key))
	if err != nil {
		fmt.Println("ERROR in finding transaction id", txId.String(), err)
		return nil, nil, err
	}
	value, err := db.lvdb.Get([]byte(key), nil)
	fmt.Println("GetTransactionLightModeByHash - value", value)
	if err != nil {
		return nil, nil, err;
	}
	_, err1 := db.hasValue([]byte(value))
	if err1 != nil {
		fmt.Println("ERROR in finding location transaction id", txId.String(), err1)
		return nil, nil, err
	}
	tx, err := db.lvdb.Get([]byte(value), nil)
	if err != nil {
		return nil, nil, err;
	}
	return value, tx, nil
}
