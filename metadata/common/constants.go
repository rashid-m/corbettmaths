package common

import (
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

const (
	InvalidMeta = 1

	IssuingRequestMeta     = 24
	IssuingResponseMeta    = 25
	ContractingRequestMeta = 26
	IssuingETHRequestMeta  = 80
	IssuingETHResponseMeta = 81

	ShardBlockReward             = 36
	AcceptedBlockRewardInfoMeta  = 37
	ShardBlockSalaryResponseMeta = 38
	BeaconRewardRequestMeta      = 39
	BeaconSalaryResponseMeta     = 40
	ReturnStakingMeta            = 41
	IncDAORewardRequestMeta      = 42
	ShardBlockRewardRequestMeta  = 43
	WithDrawRewardRequestMeta    = 44
	WithDrawRewardResponseMeta   = 45

	//staking
	ShardStakingMeta    = 63
	StopAutoStakingMeta = 127
	BeaconStakingMeta   = 64
	UnStakingMeta       = 210

	// Incognito -> Ethereum bridge
	BeaconSwapConfirmMeta = 70
	BridgeSwapConfirmMeta = 71
	BurningRequestMeta    = 27
	BurningRequestMetaV2  = 240
	BurningConfirmMeta    = 72
	BurningConfirmMetaV2  = 241

	// pde
	PDEContributionMeta                   = 90
	PDETradeRequestMeta                   = 91
	PDETradeResponseMeta                  = 92
	PDEWithdrawalRequestMeta              = 93
	PDEWithdrawalResponseMeta             = 94
	PDEContributionResponseMeta           = 95
	PDEPRVRequiredContributionRequestMeta = 204
	PDECrossPoolTradeRequestMeta          = 205
	PDECrossPoolTradeResponseMeta         = 206
	PDEFeeWithdrawalRequestMeta           = 207
	PDEFeeWithdrawalResponseMeta          = 208
	PDETradingFeesDistributionMeta        = 209

	// pDEX v3
	Pdexv3ModifyParamsMeta                  = 270
	Pdexv3AddLiquidityRequestMeta           = 271
	Pdexv3AddLiquidityResponseMeta          = 272
	Pdexv3WithdrawLiquidityRequestMeta      = 273
	Pdexv3WithdrawLiquidityResponseMeta     = 274
	Pdexv3TradeRequestMeta                  = 275
	Pdexv3TradeResponseMeta                 = 276
	Pdexv3AddOrderRequestMeta               = 277
	Pdexv3AddOrderResponseMeta              = 278
	Pdexv3WithdrawOrderRequestMeta          = 279
	Pdexv3WithdrawOrderResponseMeta         = 280
	Pdexv3UserMintNftRequestMeta            = 281
	Pdexv3UserMintNftResponseMeta           = 282
	Pdexv3MintNftRequestMeta                = 283
	Pdexv3MintNftResponseMeta               = 284
	Pdexv3StakingRequestMeta                = 285
	Pdexv3StakingResponseMeta               = 286
	Pdexv3UnstakingRequestMeta              = 287
	Pdexv3UnstakingResponseMeta             = 288
	Pdexv3WithdrawLPFeeRequestMeta          = 289
	Pdexv3WithdrawLPFeeResponseMeta         = 290
	Pdexv3WithdrawProtocolFeeRequestMeta    = 291
	Pdexv3WithdrawProtocolFeeResponseMeta   = 292
	Pdexv3MintPDEXGenesisMeta               = 293
	Pdexv3MintBlockRewardMeta               = 294
	Pdexv3DistributeStakingRewardMeta       = 295
	Pdexv3WithdrawStakingRewardRequestMeta  = 296
	Pdexv3WithdrawStakingRewardResponseMeta = 297

	// portal
	PortalCustodianDepositMeta                  = 100
	PortalRequestPortingMeta                    = 101
	PortalUserRequestPTokenMeta                 = 102
	PortalCustodianDepositResponseMeta          = 103
	PortalUserRequestPTokenResponseMeta         = 104
	PortalExchangeRatesMeta                     = 105
	PortalRedeemRequestMeta                     = 106
	PortalRedeemRequestResponseMeta             = 107
	PortalRequestUnlockCollateralMeta           = 108
	PortalCustodianWithdrawRequestMeta          = 110
	PortalCustodianWithdrawResponseMeta         = 111
	PortalLiquidateCustodianMeta                = 112
	PortalLiquidateCustodianResponseMeta        = 113
	PortalLiquidateTPExchangeRatesMeta          = 114
	PortalExpiredWaitingPortingReqMeta          = 116
	PortalRewardMeta                            = 117
	PortalRequestWithdrawRewardMeta             = 118
	PortalRequestWithdrawRewardResponseMeta     = 119
	PortalRedeemFromLiquidationPoolMeta         = 120
	PortalRedeemFromLiquidationPoolResponseMeta = 121
	PortalCustodianTopupMeta                    = 122
	PortalCustodianTopupResponseMeta            = 123
	PortalTotalRewardCustodianMeta              = 124
	PortalPortingResponseMeta                   = 125
	PortalReqMatchingRedeemMeta                 = 126
	PortalPickMoreCustodianForRedeemMeta        = 128
	PortalCustodianTopupMetaV2                  = 129
	PortalCustodianTopupResponseMetaV2          = 130

	// Portal v3
	PortalCustodianDepositMetaV3                  = 131
	PortalCustodianWithdrawRequestMetaV3          = 132
	PortalRewardMetaV3                            = 133
	PortalRequestUnlockCollateralMetaV3           = 134
	PortalLiquidateCustodianMetaV3                = 135
	PortalLiquidateByRatesMetaV3                  = 136
	PortalRedeemFromLiquidationPoolMetaV3         = 137
	PortalRedeemFromLiquidationPoolResponseMetaV3 = 138
	PortalCustodianTopupMetaV3                    = 139
	PortalTopUpWaitingPortingRequestMetaV3        = 140
	PortalRequestPortingMetaV3                    = 141
	PortalRedeemRequestMetaV3                     = 142
	PortalUnlockOverRateCollateralsMeta           = 143

	// Incognito => Ethereum's SC for portal
	PortalCustodianWithdrawConfirmMetaV3         = 170
	PortalRedeemFromLiquidationPoolConfirmMetaV3 = 171
	PortalLiquidateRunAwayCustodianConfirmMetaV3 = 172

	//Note: don't use this metadata type for others
	PortalResetPortalDBMeta = 199

	// relaying
	RelayingBNBHeaderMeta = 200
	RelayingBTCHeaderMeta = 201

	PortalTopUpWaitingPortingRequestMeta  = 202
	PortalTopUpWaitingPortingResponseMeta = 203

	// incognito mode for smart contract
	BurningForDepositToSCRequestMeta   = 96
	BurningForDepositToSCRequestMetaV2 = 242
	BurningConfirmForDepositToSCMeta   = 97
	BurningConfirmForDepositToSCMetaV2 = 243

	// PORTAL V4
	PortalV4ShieldingRequestMeta      = 260
	PortalV4ShieldingResponseMeta     = 261
	PortalV4UnshieldingRequestMeta    = 262
	PortalV4UnshieldingResponseMeta   = 263
	PortalV4UnshieldBatchingMeta      = 264
	PortalV4FeeReplacementRequestMeta = 265
	PortalV4SubmitConfirmedTxMeta     = 266
	PortalV4ConvertVaultRequestMeta   = 267

	InitTokenRequestMeta  = 244
	InitTokenResponseMeta = 245

	// incognito mode for bsc
	IssuingBSCRequestMeta  = 250
	IssuingBSCResponseMeta = 251
	BurningPBSCRequestMeta = 252
	BurningBSCConfirmMeta  = 253
)

var minerCreatedMetaTypes = []int{
	ShardBlockReward,
	BeaconSalaryResponseMeta,
	IssuingResponseMeta,
	IssuingETHResponseMeta,
	IssuingBSCResponseMeta,
	ReturnStakingMeta,
	WithDrawRewardResponseMeta,
	PDETradeResponseMeta,
	PDECrossPoolTradeResponseMeta,
	PDEWithdrawalResponseMeta,
	PDEFeeWithdrawalResponseMeta,
	PDEContributionResponseMeta,
	PortalUserRequestPTokenResponseMeta,
	PortalCustodianDepositResponseMeta,
	PortalRedeemRequestResponseMeta,
	PortalCustodianWithdrawResponseMeta,
	PortalLiquidateCustodianResponseMeta,
	PortalRequestWithdrawRewardResponseMeta,
	PortalRedeemFromLiquidationPoolResponseMeta,
	PortalCustodianTopupResponseMeta,
	PortalCustodianTopupResponseMetaV2,
	PortalPortingResponseMeta,
	PortalTopUpWaitingPortingResponseMeta,
	PortalRedeemFromLiquidationPoolResponseMetaV3,
	PortalV4ShieldingResponseMeta,
	PortalV4UnshieldingResponseMeta,
	InitTokenResponseMeta,
	Pdexv3AddLiquidityResponseMeta,
	Pdexv3MintNftResponseMeta,
	Pdexv3UserMintNftResponseMeta,
	Pdexv3WithdrawLiquidityResponseMeta,
	Pdexv3TradeResponseMeta,
	Pdexv3AddOrderResponseMeta,
	Pdexv3WithdrawOrderResponseMeta,
	Pdexv3WithdrawLPFeeResponseMeta,
	Pdexv3WithdrawProtocolFeeResponseMeta,
	Pdexv3MintPDEXGenesisMeta,
	Pdexv3StakingResponseMeta,
	Pdexv3UnstakingResponseMeta,
	Pdexv3WithdrawStakingRewardResponseMeta,
}

// Special rules for shardID: stored as 2nd param of instruction of BeaconBlock
const (
	AllShards  = -1
	BeaconOnly = -2
)

/*var (*/
//// if the blockchain is running in Docker container
//// then using GETH_NAME env's value (aka geth container name)
//// otherwise using localhost
//EthereumLightNodeHost     = utils.GetEnv("GETH_NAME", "127.0.0.1")
//EthereumLightNodeProtocol = utils.GetEnv("GETH_PROTOCOL", "http")
//EthereumLightNodePort     = utils.GetEnv("GETH_PORT", "8545")
/*)*/

const (
	StopAutoStakingAmount = 0
	EVMConfirmationBlocks = 15
)

var AcceptedWithdrawRewardRequestVersion = []int{0, 1}

var portalMetaTypesV3 = []int{
	PortalCustodianDepositMeta,
	PortalRequestPortingMeta,
	PortalUserRequestPTokenMeta,
	PortalCustodianDepositResponseMeta,
	PortalUserRequestPTokenResponseMeta,
	PortalExchangeRatesMeta,
	PortalRedeemRequestMeta,
	PortalRedeemRequestResponseMeta,
	PortalRequestUnlockCollateralMeta,
	PortalCustodianWithdrawRequestMeta,
	PortalCustodianWithdrawResponseMeta,
	PortalLiquidateCustodianMeta,
	PortalLiquidateCustodianResponseMeta,
	PortalLiquidateTPExchangeRatesMeta,
	PortalExpiredWaitingPortingReqMeta,
	PortalRewardMeta,
	PortalRequestWithdrawRewardMeta,
	PortalRequestWithdrawRewardResponseMeta,
	PortalRedeemFromLiquidationPoolMeta,
	PortalRedeemFromLiquidationPoolResponseMeta,
	PortalCustodianTopupMeta,
	PortalCustodianTopupResponseMeta,
	PortalTotalRewardCustodianMeta,
	PortalPortingResponseMeta,
	PortalReqMatchingRedeemMeta,
	PortalPickMoreCustodianForRedeemMeta,
	PortalCustodianTopupMetaV2,
	PortalCustodianTopupResponseMetaV2,

	// Portal v3
	PortalCustodianDepositMetaV3,
	PortalCustodianWithdrawRequestMetaV3,
	PortalRewardMetaV3,
	PortalRequestUnlockCollateralMetaV3,
	PortalLiquidateCustodianMetaV3,
	PortalLiquidateByRatesMetaV3,
	PortalRedeemFromLiquidationPoolMetaV3,
	PortalRedeemFromLiquidationPoolResponseMetaV3,
	PortalCustodianTopupMetaV3,
	PortalTopUpWaitingPortingRequestMetaV3,
	PortalRequestPortingMetaV3,
	PortalRedeemRequestMetaV3,
	PortalCustodianWithdrawConfirmMetaV3,
	PortalRedeemFromLiquidationPoolConfirmMetaV3,
	PortalLiquidateRunAwayCustodianConfirmMetaV3,
	PortalResetPortalDBMeta,

	PortalTopUpWaitingPortingRequestMeta,
	PortalTopUpWaitingPortingResponseMeta,
}

var portalRelayingMetaTypes = []int{
	RelayingBNBHeaderMeta,
	RelayingBTCHeaderMeta,
}

var bridgeMetas = []string{
	strconv.Itoa(BeaconSwapConfirmMeta),
	strconv.Itoa(BridgeSwapConfirmMeta),
	strconv.Itoa(BurningConfirmMeta),
	strconv.Itoa(BurningConfirmForDepositToSCMeta),
	strconv.Itoa(BurningConfirmMetaV2),
	strconv.Itoa(BurningConfirmForDepositToSCMetaV2),
	strconv.Itoa(BurningBSCConfirmMeta),
}

var portalV4MetaTypes = []int{
	PortalV4ShieldingRequestMeta,
	PortalV4ShieldingResponseMeta,
	PortalV4UnshieldingRequestMeta,
	PortalV4UnshieldBatchingMeta,
	PortalV4FeeReplacementRequestMeta,
	PortalV4SubmitConfirmedTxMeta,
	PortalV4ConvertVaultRequestMeta,
}

// NOTE: add new records when add new feature flags
var FeatureFlagWithMetaTypes = map[string][]int{
	common.PortalRelayingFlag: portalRelayingMetaTypes,
	common.PortalV3Flag:       portalMetaTypesV3,
	common.PortalV4Flag:       portalV4MetaTypes,
}