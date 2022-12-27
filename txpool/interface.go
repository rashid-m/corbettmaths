package txpool

import (
	"context"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"time"
)

type BeaconViewRetriever interface {
	GetAllCommitteeValidatorCandidate() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error)
	GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error)
	GetAutoStakingList() map[string]bool
	GetBeaconFeatureStateDB() *statedb.StateDB
	GetBeaconRewardStateDB() *statedb.StateDB
	GetBeaconSlashStateDB() *statedb.StateDB
}

type ShardViewRetriever interface {
	GetEpoch() uint64
	GetBeaconHeight() uint64
	GetStakingTx() map[string]string
	ListShardPrivacyTokenAndPRV() []common.Hash
	GetShardRewardStateDB() *statedb.StateDB
	GetCopiedFeatureStateDB() *statedb.StateDB
}

type TxPoolManager interface {
	GetShardTxsPool(shardID byte) (TxPool, error)
	StartShardTxsPool(shardID byte) error
	StopShardTxsPool(shardID byte) error
}

type PrefetchInterface interface {
	context.Context
	DecreaseNumTXRemain()
	GetMaxTime() time.Duration
	GetMaxSize() uint64
	IsRunning() bool
	GetNumTxRemain() int64
}
type TxPool interface {
	UpdateTxVerifier(tv TxVerifier)
	Start()
	Stop()
	GetInbox() chan metadata.Transaction
	IsRunning() bool
	RemoveTxs(txHashes []string)
	GetTxsTranferForNewBlock(
		cView metadata.ChainRetriever,
		sView metadata.ShardViewRetriever,
		bcView metadata.BeaconViewRetriever,
		ctx PrefetchInterface,
	) []metadata.Transaction
	FilterWithNewView(
		cView metadata.ChainRetriever,
		sView metadata.ShardViewRetriever,
		bcView metadata.BeaconViewRetriever,
	)
	CheckValidatedTxs(
		txs []metadata.Transaction,
	) (
		valid []metadata.Transaction,
		needValidate []metadata.Transaction,
	)
	CheckDoubleSpendWithCurMem(
		target metadata.Transaction,
	) (
		bool,
		bool,
		string,
		[]string,
	)
	snapshotPool() TxsData
	snapshotPoolOutCoin() map[common.Hash]interface{}
	getTxByHash(txID string) metadata.Transaction
	RemoveTx(txHash string)
}

type BlockTxsVerifier interface {
	FullValidateTransactions(
		txP TxPool,
		sView interface{},
		bcView interface{},
		txs []metadata.Transaction,
	) (bool, error)
	ValidateBatchRangeProof([]metadata.Transaction) (bool, error)
}

type FeeEstimator interface {
	RegisterBlock(block *types.ShardBlock) error
	EstimateFee(numBlocks uint64, tokenId *common.Hash) (uint64, error)
	GetLimitFeeForNativeToken() uint64
	GetMinFeePerTx() uint64
	GetSpecifiedFeeTx() uint64
}

type TxVerifier interface {
	ValidateWithoutChainstate(metadata.Transaction) (bool, error)
	ValidateWithChainState(
		tx metadata.Transaction,
		chainRetriever metadata.ChainRetriever,
		shardViewRetriever metadata.ShardViewRetriever,
		beaconViewRetriever metadata.BeaconViewRetriever,
		beaconHeight uint64,
	) (bool, error)
	FullValidateTransactions(
		chainRetriever metadata.ChainRetriever,
		shardViewRetriever metadata.ShardViewRetriever,
		beaconViewRetriever metadata.BeaconViewRetriever,
		txs []metadata.Transaction,
	) (bool, error)
	LoadCommitment(
		tx metadata.Transaction,
		shardViewRetriever metadata.ShardViewRetriever,
	) (bool, error)
	PrepareDataForTxs(
		validTxs []metadata.Transaction,
		newTxs []metadata.Transaction,
		shardViewRetriever metadata.ShardViewRetriever,
	) (bool, error)
	UpdateTransactionStateDB(
		newSDB *statedb.StateDB,
	)
	UpdateFeeEstimator(
		estimator FeeEstimator,
	)
}
