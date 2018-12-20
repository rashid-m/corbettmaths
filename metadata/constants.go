package metadata

const (
	LoanKeyDigestLength = 32
	LoanKeyLength       = 32
)

const (
	InvalidMeta = iota

	LoanRequestMeta
	LoanResponseMeta
	LoanWithdrawMeta
	LoanUnlockMeta
	LoanPaymentMeta
	DividendMeta
	CrowdsaleRequestMeta
	CrowdsaleResponseMeta
	CrowdsalePaymentMeta

	// CMB
	CMBInitRequestMeta
	CMBInitResponseMeta // offchain multisig
	CMBInitRefundMeta   // miner
	CMBDepositRequestMeta
	CMBDepositResponseMeta // offchain multisig
	CMBDepositRefundMeta   // miner
	CMBWithdrawRequestMeta
	CMBWithdrawResponseMeta // offchain multisig

	BuyFromGOVRequestMeta
	BuyFromGOVResponseMeta
	BuyBackRequestMeta
	BuyBackResponseMeta
	IssuingRequestMeta
	IssuingResponseMeta
	ContractingRequestMeta
	OracleFeedMeta
	OracleRewardMeta
	RefundMeta

	//Voting
	SubmitDCBProposalMeta
	VoteDCBProposalMeta
	VoteDCBBoardMeta
	AcceptDCBProposalMeta
	AcceptDCBBoardMeta

	SubmitGOVProposalMeta
	VoteGOVProposalMeta
	VoteGOVBoardMeta
	AcceptGOVProposalMeta
	AcceptGOVBoardMeta
	SendInitDCBVoteTokenMeta
	SendInitGOVVoteTokenMeta
	SealedLv1DCBBallotMeta
	SealedLv2DCBBallotMeta
	SealedLv3DCBBallotMeta
	NormalDCBBallotMetaFromSealer
	NormalDCBBallotMetaFromOwner
	SealedLv1GOVBallotMeta
	SealedLv2GOVBallotMeta
	SealedLv3GOVBallotMeta
	NormalGOVBallotMetaFromSealer
	NormalGOVBallotMetaFromOwner
)

const (
	MaxDivTxsPerBlock = 1000
	PayoutFrequency   = 1000 // Payout dividend every 1000 blocks
)
