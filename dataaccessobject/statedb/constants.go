package statedb

// version
const (
	defaultVersion = 0
)

// Object type
const (
	TestObjectType = iota
	CommitteeObjectType
	CommitteeRewardObjectType
	RewardRequestObjectType
	BlackListProducerObjectType
	SerialNumberObjectType
	CommitmentObjectType
	CommitmentIndexObjectType
	CommitmentLengthObjectType
	SNDerivatorObjectType
	OutputCoinObjectType
	OTACoinObjectType
	OTACoinIndexObjectType
	OTACoinLengthObjectType
	OnetimeAddressObjectType
	TokenObjectType
	WaitingPDEContributionObjectType
	PDEPoolPairObjectType
	PDEShareObjectType
	PDEStatusObjectType
	BridgeEthTxObjectType
	BridgeTokenInfoObjectType
	BridgeStatusObjectType
	BurningConfirmObjectType
	TokenTransactionObjectType

	// portal
	//final exchange rates
	PortalFinalExchangeRatesStateObjectType
	//waiting porting request
	PortalWaitingPortingRequestObjectType
	//liquidation
	PortalLiquidationPoolObjectType
	PortalStatusObjectType
	CustodianStateObjectType
	WaitingRedeemRequestObjectType
	PortalRewardInfoObjectType
	LockedCollateralStateObjectType
	RewardFeatureStateObjectType

	// PDEX v2
	PDETradingFeeObjectType

	StakerObjectType

	// Portal v3
	PortalExternalTxObjectType
	PortalConfirmProofObjectType
	PortalUnlockOverRateCollaterals

	// Portal v4
	PortalV4StatusObjectType
	PortalV4UTXOObjectType
	PortalV4ShieldRequestObjectType
	PortalWaitingUnshieldObjectType
	PortalProcessedUnshieldRequestBatchObjectType

	SlashingCommitteeObjectType

	// bsc bridge
	BridgeBSCTxObjectType
)

// Prefix length
const (
	prefixHashKeyLength = 12
	prefixKeyLength     = 20
)

// Committee Role
const (
	NextEpochShardCandidate = iota
	NextEpochBeaconCandidate
	CurrentEpochShardCandidate
	CurrentEpochBeaconCandidate
	SubstituteValidator
	CurrentValidator
	CommonBeaconPool
	CommonShardPool
	BeaconPool
	ShardPool
	BeaconCommittee
	ShardCommittee
)
const (
	BeaconChainID    = -1
	CandidateChainID = -2
)

// PDE Track Status type
const (
	WaitingContributionStatus = iota
	TradeStatus
	WithdrawStatus
)

// bridge
const (
	BridgeMinusOperator = "-"
	BridgePlusOperator  = "+"
)

// commitment
var (
	zeroBigInt = []byte("zero")
)

// token type
const (
	InitToken = iota
	CrossShardToken
	BridgeToken
	UnknownToken
)
