package blsbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	signatureschemes2 "github.com/incognitochain/incognito-chain/consensus_v2/signatureschemes"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
)

type ProposeMessageEnvironment struct {
	block                            types.BlockInterface
	previousBlock                    types.BlockInterface
	committees                       []incognitokey.CommitteePublicKey
	signingCommittees                []incognitokey.CommitteePublicKey
	userKeySet                       []signatureschemes2.MiningKey
	NumberOfFixedShardBlockValidator int
	proposerPublicBLSMiningKey       string
}

func NewProposeMessageEnvironment(block types.BlockInterface, previousBlock types.BlockInterface, committees []incognitokey.CommitteePublicKey, signingCommittees []incognitokey.CommitteePublicKey, userKeySet []signatureschemes2.MiningKey, numberOfFixedShardBlockValidator int, proposerPublicBLSMiningKey string) *ProposeMessageEnvironment {
	return &ProposeMessageEnvironment{block: block, previousBlock: previousBlock, committees: committees, signingCommittees: signingCommittees, userKeySet: userKeySet, NumberOfFixedShardBlockValidator: numberOfFixedShardBlockValidator, proposerPublicBLSMiningKey: proposerPublicBLSMiningKey}
}

type SendProposeBlockEnvironment struct {
	finalityProof    *multiview.FinalityProof
	isValidRePropose bool
	userProposeKey   signatureschemes2.MiningKey
	peerID           string
}

func NewSendProposeBlockEnvironment(finalityProof *multiview.FinalityProof, isValidRePropose bool, userProposeKey signatureschemes2.MiningKey, peerID string) *SendProposeBlockEnvironment {
	return &SendProposeBlockEnvironment{finalityProof: finalityProof, isValidRePropose: isValidRePropose, userProposeKey: userProposeKey, peerID: peerID}
}

type IProposeMessageRule interface {
	HandleBFTProposeMessage(env *ProposeMessageEnvironment, propose *BFTPropose) (*ProposeBlockInfo, error)
	CreateProposeBFTMessage(env *SendProposeBlockEnvironment, block types.BlockInterface) (*BFTPropose, error)
	GetValidFinalityProof(block types.BlockInterface, currentTimeSlot int64) (*multiview.FinalityProof, bool, string)
	HandleCleanMem(finalView uint64)
	FinalityProof() map[string]map[int64]string
}

type ProposeRuleLemma1 struct {
	logger common.Logger
}

func (p ProposeRuleLemma1) FinalityProof() map[string]map[int64]string {
	return make(map[string]map[int64]string)
}

func NewProposeRuleLemma1(logger common.Logger) *ProposeRuleLemma1 {
	return &ProposeRuleLemma1{logger: logger}
}

func (p ProposeRuleLemma1) HandleCleanMem(finalView uint64) {
	return
}

func (p ProposeRuleLemma1) HandleBFTProposeMessage(env *ProposeMessageEnvironment, propose *BFTPropose) (*ProposeBlockInfo, error) {
	return &ProposeBlockInfo{
		block:                   env.block,
		Votes:                   make(map[string]*BFTVote),
		Committees:              incognitokey.DeepCopy(env.committees),
		SigningCommittees:       incognitokey.DeepCopy(env.signingCommittees),
		UserKeySet:              signatureschemes2.DeepCopyMiningKeyArray(env.userKeySet),
		ProposerMiningKeyBase58: env.proposerPublicBLSMiningKey,
	}, nil
}

func (p ProposeRuleLemma1) CreateProposeBFTMessage(env *SendProposeBlockEnvironment, block types.BlockInterface) (*BFTPropose, error) {

	var bftPropose = new(BFTPropose)
	blockData, _ := json.Marshal(block)

	bftPropose.FinalityProof = *multiview.NewFinalityProof()
	bftPropose.ReProposeHashSignature = ""
	bftPropose.Block = blockData
	bftPropose.PeerID = env.peerID

	return bftPropose, nil
}

func (p ProposeRuleLemma1) GetValidFinalityProof(block types.BlockInterface, currentTimeSlot int64) (*multiview.FinalityProof, bool, string) {
	return multiview.NewFinalityProof(), false, ""
}

type ProposeRuleLemma2 struct {
	logger                 common.Logger
	nextBlockFinalityProof map[string]map[int64]string
	chain                  Chain
}

func (p ProposeRuleLemma2) FinalityProof() map[string]map[int64]string {
	return p.nextBlockFinalityProof
}

func NewProposeRuleLemma2(logger common.Logger, nextBlockFinalityProof map[string]map[int64]string, chain Chain) *ProposeRuleLemma2 {
	return &ProposeRuleLemma2{logger: logger, nextBlockFinalityProof: nextBlockFinalityProof, chain: chain}
}

func (p ProposeRuleLemma2) HandleCleanMem(finalView uint64) {
	for temp := range p.nextBlockFinalityProof {
		hash := common.Hash{}.NewHashFromStr2(temp)
		block, err := p.chain.GetBlockByHash(hash)
		if err == nil {
			if block.GetHeight() < finalView {
				delete(p.nextBlockFinalityProof, temp)
			}
		}
	}
}

func (p ProposeRuleLemma2) HandleBFTProposeMessage(env *ProposeMessageEnvironment, proposeMsg *BFTPropose) (*ProposeBlockInfo, error) {
	isValidLemma2 := false
	var err error
	var isReProposeFirstBlockNextHeight = false
	var isFirstBlockNextHeight = false
	var reProposeProof *multiview.ReProposeProof = nil
	isFirstBlockNextHeight = p.isFirstBlockNextHeight(env.previousBlock, env.block)
	if isFirstBlockNextHeight {
		err := p.verifyLemma2FirstBlockNextHeight(proposeMsg, env.block)
		if err != nil {
			p.logger.Error("Verify lemma2 first block next height error", err)
			return nil, err
		}
		isValidLemma2 = true
	} else {
		isReProposeFirstBlockNextHeight = p.isReProposeFromFirstBlockNextHeight(env.previousBlock, env.block, env.committees, env.NumberOfFixedShardBlockValidator)
		if isReProposeFirstBlockNextHeight {
			reProposeProof, isValidLemma2, err = p.verifyLemma2ReProposeBlockNextHeight(proposeMsg, env.block, env.committees, env.NumberOfFixedShardBlockValidator)
			if err != nil {
				p.logger.Error("Verify lemma2 error", err)
				return nil, err
			}
		}
	}

	proposeBlockInfo := newProposeBlockForProposeMsgLemma2(
		proposeMsg,
		env.block,
		env.committees,
		env.signingCommittees,
		env.userKeySet,
		env.proposerPublicBLSMiningKey,
		isValidLemma2,
		reProposeProof,
	)
	//get vote for this propose block (case receive vote faster)
	votes, err := GetVotesByBlockHashFromDB(env.block.Hash().String())
	if err != nil {
		p.logger.Error("Cannot get vote by block hash for rebuild", err)
		return nil, err
	}
	proposeBlockInfo.Votes = votes

	p.logger.Infof("HandleBFTProposeMessage Lemma 2, receive Block height %+v, hash %+v, finality height %+v, isValidLemma2 %+v",
		env.block.GetHeight(), env.block.Hash().String(), env.block.GetFinalityHeight(), isValidLemma2)
	if isValidLemma2 {
		if err := p.addFinalityProof(env.block, proposeMsg.ReProposeHashSignature, proposeMsg.FinalityProof); err != nil {
			return nil, err
		}
	}

	return proposeBlockInfo, nil
}

// isFirstBlockNextHeight verify firstBlockNextHeight
// producer timeslot is proposer timeslot
// producer is proposer
// producer timeslot = previous proposer timeslot + 1
func (p *ProposeRuleLemma2) isFirstBlockNextHeight(
	previousBlock types.BlockInterface,
	block types.BlockInterface,
) bool {

	if block.GetProposeTime() != block.GetProduceTime() {
		return false
	}

	if block.GetProposer() != block.GetProducer() {
		return false
	}

	previousProposerTimeSlot := common.CalculateTimeSlot(previousBlock.GetProposeTime())
	producerTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())

	if producerTimeSlot != previousProposerTimeSlot+1 {
		return false
	}

	return true
}

// isReProposeFromFirstBlockNextHeight verify a block is re-propose from first block next height
// producer timeslot is first block next height
// proposer timeslot > producer timeslot
// proposer is correct
func (p *ProposeRuleLemma2) isReProposeFromFirstBlockNextHeight(
	previousBlock types.BlockInterface,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	numberOfFixedShardBlockValidator int,
) bool {

	previousProposerTimeSlot := common.CalculateTimeSlot(previousBlock.GetProposeTime())
	producerTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
	proposerTimeSlot := common.CalculateTimeSlot(block.GetProposeTime())

	if producerTimeSlot != previousProposerTimeSlot+1 {
		return false
	}

	if proposerTimeSlot <= producerTimeSlot {
		return false
	}

	wantProposer, _ := GetProposerByTimeSlotFromCommitteeList(proposerTimeSlot, committees, numberOfFixedShardBlockValidator)
	wantProposerBase58, _ := wantProposer.ToBase58()
	if block.GetProposer() != wantProposerBase58 {
		return false
	}

	return true
}

func (p *ProposeRuleLemma2) verifyLemma2FirstBlockNextHeight(
	proposeMsg *BFTPropose,
	block types.BlockInterface,
) error {

	isValid, err := multiview.VerifyReProposeHashSignatureFromBlock(proposeMsg.ReProposeHashSignature, block)
	if err != nil {
		return err
	}
	if !isValid {
		return fmt.Errorf("Invalid FirstBlockNextHeight ReproposeHashSignature %+v, proposer %+v",
			proposeMsg.ReProposeHashSignature, block.GetProposer())
	}

	finalityHeight := block.GetFinalityHeight()
	previousBlockHeight := block.GetHeight() - 1
	if finalityHeight != previousBlockHeight {
		return fmt.Errorf("Invalid FirstBlockNextHeight FinalityHeight expect %+v, but got %+v",
			previousBlockHeight, finalityHeight)
	}

	return nil
}

func (p *ProposeRuleLemma2) verifyLemma2ReProposeBlockNextHeight(
	proposeMsg *BFTPropose,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	numberOfFixedShardBlockVaildator int,
) (*multiview.ReProposeProof, bool, error) {

	isValid, err := multiview.VerifyReProposeHashSignatureFromBlock(proposeMsg.ReProposeHashSignature, block)
	if err != nil {
		return nil, false, err
	}
	if !isValid {
		return nil, false, fmt.Errorf("Invalid ReProposeBlockNextHeight ReproposeHashSignature %+v, proposer %+v",
			proposeMsg.ReProposeHashSignature, block.GetProposer())
	}

	reProposeProof, isValidProof, err := p.verifyFinalityProof(proposeMsg, block, committees, numberOfFixedShardBlockVaildator)
	if err != nil {
		return nil, false, err
	}

	finalityHeight := block.GetFinalityHeight()
	if isValidProof {
		previousBlockHeight := block.GetHeight() - 1
		if finalityHeight != previousBlockHeight {
			return nil, false, fmt.Errorf("Invalid ReProposeBlockNextHeight FinalityHeight expect %+v, but got %+v",
				previousBlockHeight, finalityHeight)
		}
	} else {
		if finalityHeight != 0 {
			return nil, false, fmt.Errorf("Invalid ReProposeBlockNextHeight FinalityHeight expect %+v, but got %+v",
				0, finalityHeight)
		}
	}

	return reProposeProof, isValidProof, nil
}

func (p *ProposeRuleLemma2) verifyFinalityProof(
	proposeMsg *BFTPropose,
	block types.BlockInterface,
	committees []incognitokey.CommitteePublicKey,
	numberOfFixedShardBlockValidator int,
) (*multiview.ReProposeProof, bool, error) {

	finalityProof := proposeMsg.FinalityProof

	previousBlockHash := block.GetPrevHash()
	producer := block.GetProducer()
	rootHash := block.GetAggregateRootHash()
	beginTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
	currentTimeSlot := common.CalculateTimeSlot(block.GetProposeTime())

	if int(currentTimeSlot-beginTimeSlot) != len(finalityProof.ReProposeHashSignature) {
		p.logger.Infof("Failed to verify finality proof, expect number of proof %+v, but got %+v",
			int(currentTimeSlot-beginTimeSlot), len(finalityProof.ReProposeHashSignature))
		return nil, false, nil
	}

	proposerBase58List := []string{}
	for reProposeTimeSlot := beginTimeSlot; reProposeTimeSlot < currentTimeSlot; reProposeTimeSlot++ {
		reProposer, _ := GetProposerByTimeSlotFromCommitteeList(reProposeTimeSlot, committees, numberOfFixedShardBlockValidator)
		reProposerBase58, _ := reProposer.ToBase58()
		proposerBase58List = append(proposerBase58List, reProposerBase58)
	}

	err := finalityProof.Verify(
		previousBlockHash,
		producer,
		beginTimeSlot,
		proposerBase58List,
		rootHash,
	)
	if err != nil {
		return nil, false, err
	}

	reProposeProof := multiview.NewReProposeProof(&finalityProof, proposerBase58List)

	return reProposeProof, true, nil
}

func (p *ProposeRuleLemma2) addFinalityProof(
	block types.BlockInterface,
	reProposeHashSignature string,
	proof multiview.FinalityProof,
) error {
	previousHash := block.GetPrevHash()
	beginTimeSlot := common.CalculateTimeSlot(block.GetProduceTime())
	currentTimeSlot := common.CalculateTimeSlot(block.GetProposeTime())

	if currentTimeSlot-beginTimeSlot > MAX_FINALITY_PROOF {
		return nil
	}

	nextBlockFinalityProof, ok := p.nextBlockFinalityProof[previousHash.String()]
	if !ok {
		nextBlockFinalityProof = make(map[int64]string)
	}

	nextBlockFinalityProof[currentTimeSlot] = reProposeHashSignature
	p.logger.Infof("Add Finality Proof | Block %+v, %+v, Current Block Sig for Timeslot: %+v",
		block.GetHeight(), block.Hash().String(), currentTimeSlot)

	index := 0
	var err error
	for timeSlot := beginTimeSlot; timeSlot < currentTimeSlot; timeSlot++ {
		_, ok := nextBlockFinalityProof[timeSlot]
		if !ok {
			nextBlockFinalityProof[timeSlot], err = proof.GetProofByIndex(index)
			if err != nil {
				return err
			}
			p.logger.Infof("Add Finality Proof | Block %+v, %+v, Previous Proof for Timeslot: %+v",
				block.GetHeight(), block.Hash().String(), timeSlot)
		}
		index++
	}

	p.nextBlockFinalityProof[previousHash.String()] = nextBlockFinalityProof

	return nil
}

//ProposerByTimeSlot ...
func GetProposerByTimeSlotFromCommitteeList(ts int64, committees []incognitokey.CommitteePublicKey, numberOfFixedShardBlockValidator int) (incognitokey.CommitteePublicKey, int) {
	proposer, proposerIndex := GetProposer(
		ts,
		committees,
		numberOfFixedShardBlockValidator,
	)
	return proposer, proposerIndex
}

func GetProposer(
	ts int64, committees []incognitokey.CommitteePublicKey,
	lenProposers int) (incognitokey.CommitteePublicKey, int) {
	id := GetProposerByTimeSlot(ts, lenProposers)
	return committees[id], id
}

func GetProposerByTimeSlot(ts int64, committeeLen int) int {
	id := int(ts) % committeeLen
	return id
}

func (p ProposeRuleLemma2) CreateProposeBFTMessage(env *SendProposeBlockEnvironment, block types.BlockInterface) (*BFTPropose, error) {

	reProposeHashSignature, err := multiview.CreateReProposeHashSignature(
		env.userProposeKey.PriKey[common.BridgeConsensus], block)

	if err != nil {
		return nil, err
	}

	blockData, _ := json.Marshal(block)
	var bftPropose = new(BFTPropose)

	if env.isValidRePropose {
		bftPropose.FinalityProof = *env.finalityProof
	} else {
		bftPropose.FinalityProof = *multiview.NewFinalityProof()
	}
	bftPropose.ReProposeHashSignature = reProposeHashSignature

	bftPropose.Block = blockData
	bftPropose.PeerID = env.peerID

	return bftPropose, nil
}

func (p ProposeRuleLemma2) GetValidFinalityProof(block types.BlockInterface, currentTimeSlot int64) (*multiview.FinalityProof, bool, string) {
	if block == nil {
		return multiview.NewFinalityProof(), false, "block is nil"
	}

	finalityData, ok := p.nextBlockFinalityProof[block.GetPrevHash().String()]
	if !ok {
		return multiview.NewFinalityProof(), false, "finality proof not found"
	}

	finalityProof := multiview.NewFinalityProof()

	producerTime := block.GetProduceTime()
	producerTimeTimeSlot := common.CalculateTimeSlot(producerTime)

	if currentTimeSlot-producerTimeTimeSlot > MAX_FINALITY_PROOF {
		return finalityProof, false, fmt.Sprintf("exceed max finality proof, required %+v proofs", currentTimeSlot-producerTimeTimeSlot)
	}

	for i := producerTimeTimeSlot; i < currentTimeSlot; i++ {
		reProposeHashSignature, ok := finalityData[i]
		if !ok {
			return multiview.NewFinalityProof(), false, "invalid re-propose hash signature"
		}
		finalityProof.AddProof(reProposeHashSignature)
	}

	return finalityProof, true, ""
}

type NoHandleProposeMessageRule struct {
	logger common.Logger
}

func NewNoHandleProposeMessageRule(logger common.Logger) *NoHandleProposeMessageRule {
	return &NoHandleProposeMessageRule{logger: logger}
}

func (n NoHandleProposeMessageRule) HandleBFTProposeMessage(env *ProposeMessageEnvironment, propose *BFTPropose) (*ProposeBlockInfo, error) {
	n.logger.Debug("using no-handle-propose-message rule, HandleBFTProposeMessage don't work ")
	return new(ProposeBlockInfo), errors.New("using no handle propose message rule")
}

func (n NoHandleProposeMessageRule) CreateProposeBFTMessage(env *SendProposeBlockEnvironment, block types.BlockInterface) (*BFTPropose, error) {
	n.logger.Debug("using no-handle-propose-message rule, CreateProposeBFTMessage don't work ")
	return new(BFTPropose), errors.New("using no handle propose message rule")
}

func (n NoHandleProposeMessageRule) GetValidFinalityProof(block types.BlockInterface, currentTimeSlot int64) (*multiview.FinalityProof, bool, string) {
	return multiview.NewFinalityProof(), false, ""
}

func (n NoHandleProposeMessageRule) HandleCleanMem(finalView uint64) {
	return
}

func (n NoHandleProposeMessageRule) FinalityProof() map[string]map[int64]string {
	return make(map[string]map[int64]string)
}
