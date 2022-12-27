// Copyright (c) 2016 The thaibaoautonomous developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	// estimateFeeDepth is the maximum number of blocks before a transaction
	// is confirmed that we want to track.
	estimateFeeDepth = 200

	// estimateFeeBinSize is the number of txs stored in each bin.
	estimateFeeBinSize = 100

	// estimateFeeMaxReplacements is the max number of replacements that
	// can be made by the txs found in a given block.
	estimateFeeMaxReplacements = 10

	// DefaultEstimateFeeMaxRollback is the default number of rollbacks
	// allowed by the fee estimator for orphaned blocks.
	DefaultEstimateFeeMaxRollback = 2

	// DefaultEstimateFeeMinRegisteredBlocks is the default minimum
	// number of blocks which must be observed by the fee estimator before
	// it will provide fee estimations.
	DefaultEstimateFeeMinRegisteredBlocks = 3
)

var (
	// EstimateFeeDatabaseKey is the key that we use to
	// store the fee estimator in the database.
	EstimateFeeDatabaseKey = []byte("estimatefee")
)

// NewCoinPerKilobyte creates a CoinPerByte from an Amount and a
// size in bytes.
func NewCoinPerKilobyte(fee uint64, size uint64) uint64 {
	return uint64(float64(fee) / float64(size))
}

// observedTransaction represents an observed transaction and some
// additional data required for the fee estimation algorithm.
type observedTransaction struct {
	// A transaction hash.
	hash common.Hash

	// The PRV fee per kilobyte of the transaction in coins.
	feeRate uint64

	// The token fee per kilobyte of the transaction in coins.
	feeRateForToken map[common.Hash]uint64

	// The block height when it was observed.
	observed uint64

	// The height of the block in which it was mined.
	// If the transaction has not yet been mined, it is zero.
	mined uint64
}

func (o *observedTransaction) Serialize(w io.Writer) {
	binary.Write(w, binary.BigEndian, o.hash)
	binary.Write(w, binary.BigEndian, o.feeRate)
	binary.Write(w, binary.BigEndian, o.feeRateForToken)
	binary.Write(w, binary.BigEndian, o.observed)
	binary.Write(w, binary.BigEndian, o.mined)
}

func deserializeObservedTransaction(r io.Reader) (*observedTransaction, error) {
	ot := observedTransaction{}

	// The first 32 bytes should be a hash.
	binary.Read(r, binary.BigEndian, &ot.hash)

	// The next 8 are feeRate
	binary.Read(r, binary.BigEndian, &ot.feeRate)

	// The next 8 are feeRateForToken
	binary.Read(r, binary.BigEndian, &ot.feeRateForToken)

	// And next there are two uint32's.
	binary.Read(r, binary.BigEndian, &ot.observed)
	binary.Read(r, binary.BigEndian, &ot.mined)

	return &ot, nil
}

// registeredBlock has the hash of a block and the list of transactions
// it mined which had been previously observed by the feeEstimator. It
// is used if Rollback is called to reverse the effect of registering
// a block.
type registeredBlock struct {
	hash         common.Hash
	transactions []*observedTransaction
}

func (rb *registeredBlock) serialize(w io.Writer, txs map[*observedTransaction]uint32) {
	binary.Write(w, binary.BigEndian, rb.hash)

	binary.Write(w, binary.BigEndian, uint32(len(rb.transactions)))
	for _, o := range rb.transactions {
		binary.Write(w, binary.BigEndian, txs[o])
	}
}

// feeEstimator manages the data necessary to create
// fee estimations. It is safe for concurrent access.
type FeeEstimator struct {
	maxRollback uint32
	binSize     int32

	// The maximum number of replacements that can be made in a single
	// bin per block. Default is estimateFeeMaxReplacements
	maxReplacements int32

	// The minimum number of blocks that can be registered with the fee
	// estimator before it will provide answers.
	minRegisteredBlocks uint32

	// The last known height.
	lastKnownHeight uint64

	// The number of blocks that have been registered.
	numBlocksRegistered uint32

	mtx      sync.RWMutex
	observed map[common.Hash]*observedTransaction
	bin      [estimateFeeDepth][]*observedTransaction

	// The cached estimates.
	cached []uint64

	// Transactions that have been removed from the bins. This allows us to
	// revert in case of an orphaned block.
	dropped []*registeredBlock

	// min fee which be needed for payment on tx(per Kb data)
	limitFee uint64

	// min fee per transaction
	minFeePerTx uint64

	// specified fee for special transactions
	specifiedFeeTx uint64
}

// NewFeeEstimator creates a feeEstimator for which at most maxRollback blocks
// can be unregistered and which returns an error unless minRegisteredBlocks
// have been registered with it.
func NewFeeEstimator(maxRollback, minRegisteredBlocks uint32, limitFee, minFeePerTx, specifiedFeeTx uint64) *FeeEstimator {
	return &FeeEstimator{
		maxRollback:         maxRollback,
		minRegisteredBlocks: minRegisteredBlocks,
		lastKnownHeight:     unminedHeight,
		binSize:             estimateFeeBinSize,
		maxReplacements:     estimateFeeMaxReplacements,
		observed:            make(map[common.Hash]*observedTransaction),
		dropped:             make([]*registeredBlock, 0, maxRollback),
		limitFee:            limitFee,
		minFeePerTx:         minFeePerTx,
		specifiedFeeTx:      specifiedFeeTx,
	}
}

// ObserveTransaction is called when a new transaction is observed in the mempool.
func (ef *FeeEstimator) ObserveTransaction(t *TxDesc) {
	ef.mtx.Lock()
	defer ef.mtx.Unlock()

	// If we haven't seen a block yet we don't know when this one arrived,
	// so we ignore it.
	if ef.lastKnownHeight == unminedHeight {
		return
	}

	hash := *t.Desc.Tx.Hash()
	if _, ok := ef.observed[hash]; !ok {
		size := t.Desc.Tx.GetTxActualSize()

		feeRateForToken := make(map[common.Hash]uint64)
		if t.Desc.Tx.GetType() == common.TxCustomTokenPrivacyType || t.Desc.Tx.GetType() == common.TxCustomTokenPrivacyType {
			tokenID := t.Desc.Tx.GetTokenID()
			tokenFee := t.Desc.FeeToken
			feeRateForToken[*tokenID] = NewCoinPerKilobyte(tokenFee, size)
		}

		ef.observed[hash] = &observedTransaction{
			hash:            hash,
			feeRate:         NewCoinPerKilobyte(uint64(t.Desc.Fee), size),
			feeRateForToken: feeRateForToken,
			observed:        t.Desc.Height,
			mined:           unminedHeight,
		}
		Logger.log.Info("Observe Transaction success", t.Desc.Tx.Hash())
	}
}

// RegisterBlock informs the fee estimator of a new block to take into account.
func (ef *FeeEstimator) RegisterBlock(block *types.ShardBlock) error {
	ef.mtx.Lock()
	defer ef.mtx.Unlock()

	// The previous sorted list is invalid, so delete it.
	ef.cached = nil

	height := block.Header.Height
	if height != ef.lastKnownHeight+1 && ef.lastKnownHeight != unminedHeight {
		Logger.log.Errorf("RegisterBlock: intermediate block not recorded; current height is %d; new height is %d, shardID %d",
			ef.lastKnownHeight, height, block.Header.ShardID)
		return fmt.Errorf("RegisterBlock: intermediate block not recorded; current height is %d; new height is %d",
			ef.lastKnownHeight, height)
	}

	// Update the last known height.
	ef.lastKnownHeight = height
	ef.numBlocksRegistered++

	// Randomly order txs in block.
	transactions := make(map[*common.Hash]bool)
	for _, t := range block.Body.Transactions {
		transactions[t.Hash()] = true
	}

	// Count the number of replacements we make per bin so that we don't
	// replace too many.
	var replacementCounts [estimateFeeDepth]int

	// Keep track of which txs were dropped in case of an orphan block.
	dropped := &registeredBlock{
		hash:         *block.Hash(),
		transactions: make([]*observedTransaction, 0, 100),
	}

	// Go through the txs in the block.
	for t := range transactions {

		// Have we observed this tx in the mempool?
		o, ok := ef.observed[*t]
		if !ok {
			Logger.log.Info("RegisterBlock: Not in observed", block.Header.ShardID, t.String())
			continue
		}

		// Put the observed tx in the oppropriate bin.
		blocksToConfirm := height - o.observed - 1

		// This shouldn't happen if the fee estimator works correctly,
		// but return an error if it does.
		if o.mined != unminedHeight {
			Logger.log.Error("RegisterBlock: Estimate fee: transaction ", t.String(), " has already been mined", block.Header.ShardID, t.String())
			return errors.New("Transaction has already been mined")
		}

		// This shouldn't happen but check just in case to avoid
		// an out-of-bounds array index later.
		if blocksToConfirm >= estimateFeeDepth {
			Logger.log.Info("RegisterBlock: This shouldn't happen but check just in case to avoid an out-of-bounds array index later.", block.Header.ShardID, t.String())
			continue
		}

		// Make sure we do not replace too many transactions per min.
		if replacementCounts[blocksToConfirm] == int(ef.maxReplacements) {
			Logger.log.Info("RegisterBlock: Make sure we do not replace too many transactions per min", block.Header.ShardID, t.String())
			continue
		}

		o.mined = height

		replacementCounts[blocksToConfirm]++

		bin := ef.bin[blocksToConfirm]

		// Remove a random element and replace it with this new tx.
		if len(bin) == int(ef.binSize) {
			// Don't drop transactions we have just added from this same block.
			l := int(ef.binSize) - replacementCounts[blocksToConfirm]
			drop := rand.Intn(l)
			dropped.transactions = append(dropped.transactions, bin[drop])

			bin[drop] = bin[l-1]
			bin[l-1] = o
		} else {
			bin = append(bin, o)
		}
		ef.bin[blocksToConfirm] = bin
	}

	// Go through the mempool for txs that have been in too long.
	for hash, o := range ef.observed {
		if o.mined == unminedHeight && height-o.observed >= estimateFeeDepth {
			delete(ef.observed, hash)
		}
	}

	// Add dropped list to history.
	if ef.maxRollback == 0 {
		return nil
	}

	if uint32(len(ef.dropped)) == ef.maxRollback {
		ef.dropped = append(ef.dropped[1:], dropped)
	} else {
		ef.dropped = append(ef.dropped, dropped)
	}
	Logger.log.Debugf("Estimate fee register block success for block hash %s", block.Header.ShardID, block.Hash().String())
	return nil
}

// LastKnownHeight returns the height of the last block which was registered.
func (ef *FeeEstimator) LastKnownHeight() uint64 {
	ef.mtx.Lock()
	defer ef.mtx.Unlock()

	return ef.lastKnownHeight
}

// Rollback unregisters a recently registered block from the feeEstimator.
// This can be used to reverse the effect of an orphaned block on the fee
// estimator. The maximum number of rollbacks allowed is given by
// maxRollbacks.
//
// Note: not everything can be rolled back because some transactions are
// deleted if they have been observed too long ago. That means the result
// of Rollback won't always be exactly the same as if the last block had not
// happened, but it should be close enough.
func (ef *FeeEstimator) Rollback(hash *common.Hash) error {
	ef.mtx.Lock()
	defer ef.mtx.Unlock()

	// Find this block in the stack of recent registered blocks.
	var n int
	for n = 1; n <= len(ef.dropped); n++ {
		if ef.dropped[len(ef.dropped)-n].hash.IsEqual(hash) {
			break
		}
	}

	if n > len(ef.dropped) {
		return errors.New("no such block was recently registered")
	}

	for i := 0; i < n; i++ {
		ef.rollback()
	}

	return nil
}

// rollback rolls back the effect of the last block in the stack
// of registered blocks.
func (ef *FeeEstimator) rollback() {
	// The previous sorted list is invalid, so delete it.
	ef.cached = nil

	// pop the last list of dropped txs from the stack.
	last := len(ef.dropped) - 1
	if last == -1 {
		// Cannot really happen because the exported calling function
		// only rolls back a block already known to be in the list
		// of dropped transactions.
		return
	}

	dropped := ef.dropped[last]

	// where we are in each bin as we replace txs?
	var replacementCounters [estimateFeeDepth]int

	// Go through the txs in the dropped block.
	for _, o := range dropped.transactions {
		// Which bin was this tx in?
		blocksToConfirm := o.mined - o.observed - 1

		bin := ef.bin[blocksToConfirm]

		var counter = replacementCounters[blocksToConfirm]

		// Continue to go through that bin where we left off.
		for {
			if counter >= len(bin) {
				// Panic, as we have entered an unrecoverable invalid state.
				panic(errors.New("illegal state: cannot rollback dropped transaction"))
			}

			prev := bin[counter]

			if prev.mined == ef.lastKnownHeight {
				prev.mined = unminedHeight

				bin[counter] = o

				counter++
				break
			}

			counter++
		}

		replacementCounters[blocksToConfirm] = counter
	}

	// Continue going through bins to find other txs to remove
	// which did not replace any other when they were entered.
	for i, j := range replacementCounters {
		for {
			l := len(ef.bin[i])
			if j >= l {
				break
			}

			prev := ef.bin[i][j]

			if prev.mined == ef.lastKnownHeight {
				prev.mined = unminedHeight

				newBin := append(ef.bin[i][0:j], ef.bin[i][j+1:l]...)
				// leak but it causes a panic when it is uncommented.
				// ef.bin[i][j] = nil
				ef.bin[i] = newBin

				continue
			}

			j++
		}
	}

	ef.dropped = ef.dropped[0:last]

	// The number of blocks the fee estimator has seen is decrimented.
	ef.numBlocksRegistered--
	ef.lastKnownHeight--
}

// estimateFeeSet is a set of txs that can that is sorted
// by the fee per kb rate.
// inherit from golang sorter
type estimateFeeSet struct {
	feeRate         []uint64
	feeRateForToken map[common.Hash][]uint64
	bin             [estimateFeeDepth]uint32
}

func (b *estimateFeeSet) Len() int { return len(b.feeRate) }

func (b *estimateFeeSet) Less(i, j int) bool {
	return b.feeRate[i] > b.feeRate[j]
}

func (b *estimateFeeSet) Swap(i, j int) {
	b.feeRate[i], b.feeRate[j] = b.feeRate[j], b.feeRate[i]
}

// estimateFee returns the estimated fee for a transaction
// to confirm in confirmations blocks from now, given
// the data set we have collected.
func (b *estimateFeeSet) estimateFee(confirmations int) uint64 {
	if confirmations <= 0 {
		return uint64(math.Inf(1))
	}

	if confirmations > estimateFeeDepth {
		return 0
	}

	// We don't have any transactions!
	if len(b.feeRate) == 0 {
		return 0
	}

	var min, max = 0, 0
	for i := 0; i < confirmations-1; i++ {
		min += int(b.bin[i])
	}

	max = min + int(b.bin[confirmations-1]) - 1
	if max < min {
		max = min
	}
	feeIndex := (min + max) / 2
	if feeIndex >= len(b.feeRate) {
		feeIndex = len(b.feeRate) - 1
	}

	return b.feeRate[feeIndex]
}

// estimateFee returns the estimated fee for a transaction
// to confirm in confirmations blocks from now, given
// the data set we have collected.
func (b *estimateFeeSet) estimateFeeForToken(confirmations int, tokenId *common.Hash) uint64 {
	if confirmations <= 0 {
		return uint64(math.Inf(1))
	}

	if confirmations > estimateFeeDepth {
		return 0
	}

	// We don't have any privacy token transactions!
	if len(b.feeRateForToken[*tokenId]) == 0 {
		return 0
	}

	// sort feeRateForToken

	var min, max = 0, 0
	for i := 0; i < confirmations-1; i++ {
		min += int(b.bin[i])
	}

	max = min + int(b.bin[confirmations-1]) - 1
	if max < min {
		max = min
	}
	feeIndex := (min + max) / 2
	if feeIndex >= len(b.feeRateForToken[*tokenId]) {
		feeIndex = len(b.feeRateForToken[*tokenId]) - 1
	}

	return b.feeRateForToken[*tokenId][feeIndex]
}

// newEstimateFeeSet creates a temporary data structure that
// can be used to find all fee estimates.
func (ef *FeeEstimator) newEstimateFeeSet(tokenID *common.Hash) *estimateFeeSet {
	set := &estimateFeeSet{}

	capacity := 0
	for i, b := range ef.bin {
		l := len(b)
		set.bin[i] = uint32(l)
		capacity += l
	}

	set.feeRate = make([]uint64, capacity)
	set.feeRateForToken = make(map[common.Hash][]uint64)

	i := 0
	for _, b := range ef.bin {
		for _, o := range b {
			set.feeRate[i] = o.feeRate

			if tokenID != nil {
				for key, value := range o.feeRateForToken {
					if key.IsEqual(tokenID) {
						set.feeRateForToken[key] = append(set.feeRateForToken[key], value)
					}
				}
			}

			i++
		}
	}

	if tokenID == nil {
		sort.Sort(set)
	} else {
	}

	return set
}

// estimates returns the set of all fee estimates from 1 to estimateFeeDepth
// confirmations from now.
func (ef *FeeEstimator) estimates(tokenID *common.Hash) []uint64 {
	set := ef.newEstimateFeeSet(tokenID)

	estimates := make([]uint64, estimateFeeDepth)
	for i := 0; i < estimateFeeDepth; i++ {
		if tokenID != nil {
			estimates[i] = set.estimateFeeForToken(i+1, tokenID)
		} else {
			estimates[i] = set.estimateFee(i + 1)
		}
	}

	return estimates
}

// EstimateFee estimates the fee per byte to have a tx confirmed a given
// number of blocks from now.
func (ef *FeeEstimator) EstimateFee(numBlocks uint64, tokenId *common.Hash) (uint64, error) {
	ef.mtx.Lock()
	defer ef.mtx.Unlock()

	// If the number of registered blocks is below the minimum, return
	// an error.
	if ef.numBlocksRegistered < ef.minRegisteredBlocks {
		return 0, errors.New("not enough blocks have been observed")
	}

	if numBlocks == 0 {
		return 0, errors.New("cannot confirm transaction in zero blocks")
	}

	if numBlocks > estimateFeeDepth {
		return 0, fmt.Errorf(
			"can only estimate fees for up to %d blocks from now",
			estimateFeeBinSize)
	}

	// If there are no cached results, generate them.
	if ef.cached == nil {
		ef.cached = ef.estimates(tokenId)
	}

	result := ef.cached[int(numBlocks)-1]
	return result, nil
}

// In case the format for the serialized version of the feeEstimator changes,
// we use a version number. If the version number changes, it does not make
// sense to try to upgrade a previous version to a new version. Instead, just
// start fee estimation over.
const estimateFeeSaveVersion = 1

func deserializeRegisteredBlock(r io.Reader, txs map[uint32]*observedTransaction) (*registeredBlock, error) {
	var lenTransactions uint32

	rb := &registeredBlock{}
	binary.Read(r, binary.BigEndian, &rb.hash)
	binary.Read(r, binary.BigEndian, &lenTransactions)

	rb.transactions = make([]*observedTransaction, lenTransactions)

	for i := uint32(0); i < lenTransactions; i++ {
		var index uint32
		binary.Read(r, binary.BigEndian, &index)
		rb.transactions[i] = txs[index]
	}

	return rb, nil
}

// FeeEstimatorState represents a saved feeEstimator that can be
// restored with data from an earlier session of the program.
type FeeEstimatorState []byte

// observedTxSet is a set of txs that can that is sorted
// by hash. It exists for serialization purposes so that
// a serialized state always comes out the same.
type observedTxSet []*observedTransaction

func (q observedTxSet) Len() int { return len(q) }

func (q observedTxSet) Less(i, j int) bool {
	return strings.Compare(q[i].hash.String(), q[j].hash.String()) < 0
}

func (q observedTxSet) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

// Save records the current state of the feeEstimator to a []byte that
// can be restored later.
func (ef *FeeEstimator) Save() FeeEstimatorState {
	ef.mtx.Lock()
	defer ef.mtx.Unlock()

	w := bytes.NewBuffer(make([]byte, 0))

	binary.Write(w, binary.BigEndian, uint32(estimateFeeSaveVersion))

	// Insert basic parameters.
	binary.Write(w, binary.BigEndian, &ef.maxRollback)
	binary.Write(w, binary.BigEndian, &ef.binSize)
	binary.Write(w, binary.BigEndian, &ef.maxReplacements)
	binary.Write(w, binary.BigEndian, &ef.minRegisteredBlocks)
	binary.Write(w, binary.BigEndian, &ef.lastKnownHeight)
	binary.Write(w, binary.BigEndian, &ef.numBlocksRegistered)
	binary.Write(w, binary.BigEndian, &ef.limitFee)

	// Put all the observed transactions in a sorted list.
	var txCount uint32
	ots := make([]*observedTransaction, len(ef.observed))
	for hash := range ef.observed {
		ots[txCount] = ef.observed[hash]
		txCount++
	}

	sort.Sort(observedTxSet(ots))

	txCount = 0
	observed := make(map[*observedTransaction]uint32)
	binary.Write(w, binary.BigEndian, uint32(len(ef.observed)))
	for _, ot := range ots {
		ot.Serialize(w)
		observed[ot] = txCount
		txCount++
	}

	// Save all the right bins.
	for _, list := range ef.bin {

		binary.Write(w, binary.BigEndian, uint32(len(list)))

		for _, o := range list {
			binary.Write(w, binary.BigEndian, observed[o])
		}
	}

	// Dropped transactions.
	binary.Write(w, binary.BigEndian, uint32(len(ef.dropped)))
	for _, registered := range ef.dropped {
		registered.serialize(w, observed)
	}

	// CommitAll the tx and return.
	return FeeEstimatorState(w.Bytes())
}

// RestoreFeeEstimator takes a FeeEstimatorState that was previously
// returned by Save and restores it to a feeEstimator
func RestoreFeeEstimator(data FeeEstimatorState) (*FeeEstimator, error) {
	r := bytes.NewReader([]byte(data))

	// Check version
	var version uint32
	err := binary.Read(r, binary.BigEndian, &version)
	if err != nil {
		return nil, err
	}
	if version != estimateFeeSaveVersion {
		return nil, fmt.Errorf("Incorrect version: expected %d found %d", estimateFeeSaveVersion, version)
	}

	ef := &FeeEstimator{
		observed: make(map[common.Hash]*observedTransaction),
	}

	// Read basic parameters.
	binary.Read(r, binary.BigEndian, &ef.maxRollback)
	binary.Read(r, binary.BigEndian, &ef.binSize)
	binary.Read(r, binary.BigEndian, &ef.maxReplacements)
	binary.Read(r, binary.BigEndian, &ef.minRegisteredBlocks)
	binary.Read(r, binary.BigEndian, &ef.lastKnownHeight)
	binary.Read(r, binary.BigEndian, &ef.numBlocksRegistered)
	binary.Read(r, binary.BigEndian, &ef.limitFee)

	// Read transactions.
	var numObserved uint32
	observed := make(map[uint32]*observedTransaction)
	binary.Read(r, binary.BigEndian, &numObserved)
	for i := uint32(0); i < numObserved; i++ {
		ot, err := deserializeObservedTransaction(r)
		if err != nil {
			return nil, err
		}
		observed[i] = ot
		ef.observed[ot.hash] = ot
	}

	// Read bins.
	for i := 0; i < estimateFeeDepth; i++ {
		var numTransactions uint32
		binary.Read(r, binary.BigEndian, &numTransactions)
		bin := make([]*observedTransaction, numTransactions)
		for j := uint32(0); j < numTransactions; j++ {
			var index uint32
			binary.Read(r, binary.BigEndian, &index)

			var exists bool
			bin[j], exists = observed[index]
			if !exists {
				return nil, fmt.Errorf("Invalid transaction reference %d", index)
			}
		}
		ef.bin[i] = bin
	}

	// Read dropped transactions.
	var numDropped uint32
	binary.Read(r, binary.BigEndian, &numDropped)
	ef.dropped = make([]*registeredBlock, numDropped)
	for i := uint32(0); i < numDropped; i++ {
		var err error
		ef.dropped[int(i)], err = deserializeRegisteredBlock(r, observed)
		if err != nil {
			return nil, err
		}
	}
	Logger.log.Debugf("Fee estimator data when restoring #%d", ef)
	return ef, nil
}

// returns the limit fee (min fee per kb)
func (ef FeeEstimator) GetLimitFeeForNativeToken() uint64 {
	return ef.limitFee
}

// returns the min fee per tx
func (ef FeeEstimator) GetMinFeePerTx() uint64 {
	return ef.minFeePerTx
}

// returns the specified fee tx
func (ef FeeEstimator) GetSpecifiedFeeTx() uint64 {
	return ef.specifiedFeeTx
}
