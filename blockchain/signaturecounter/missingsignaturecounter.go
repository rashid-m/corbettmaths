package signaturecounter

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Penalty struct {
	MinPercent   uint
	ForceUnstake bool
}

type MissingSignature struct {
	VotePercent uint
	ActualTotal uint
	Missing     uint
}

func NewMissingSignature() MissingSignature {
	return MissingSignature{
		ActualTotal: 0,
		Missing:     0,
	}
}

var defaultRule = []Penalty{
	{
		MinPercent:   50,
		ForceUnstake: true,
	},
}

func NewPenalty() Penalty {
	return Penalty{}
}

func (p Penalty) IsEmpty() bool {
	return reflect.DeepEqual(p, NewPenalty())
}

type IMissingSignatureCounter interface {
	MissingSignature() map[string]MissingSignature
	Penalties() []Penalty
	AddMissingSignature(validationData string, shardID int, committees []incognitokey.CommitteePublicKey) error
	AddPreviousMissignSignature(prevValidationData string, shardID int) error
	GetAllSlashingPenaltyWithActualTotalBlock() map[string]Penalty
	GetAllSlashingPenaltyWithExpectedTotalBlock(map[string]uint) map[string]Penalty
	Reset(committees []string)
	CommitteeChange(committees []string)
	Copy() IMissingSignatureCounter
}

type MissingSignatureCounter struct {
	missingSignature map[string]MissingSignature
	penalties        []Penalty
	lock             *sync.RWMutex

	lastShardStateValidatorCommittee map[int][]string
	lastShardStateValidatorIndex     map[int][]int
}

func (s *MissingSignatureCounter) Penalties() []Penalty {
	return s.penalties
}

func (s *MissingSignatureCounter) MissingSignature() map[string]MissingSignature {
	s.lock.RLock()
	defer s.lock.RUnlock()
	missingSignature := make(map[string]MissingSignature)
	for k, v := range s.missingSignature {
		missingSignature[k] = v
	}
	return missingSignature
}

func NewDefaultSignatureCounter(committees []string) *MissingSignatureCounter {
	missingSignature := make(map[string]MissingSignature)
	for _, v := range committees {
		missingSignature[v] = NewMissingSignature()
	}
	return &MissingSignatureCounter{
		missingSignature:                 missingSignature,
		penalties:                        defaultRule,
		lock:                             new(sync.RWMutex),
		lastShardStateValidatorCommittee: make(map[int][]string),
		lastShardStateValidatorIndex:     make(map[int][]int),
	}
}

func (s *MissingSignatureCounter) AddMissingSignature(data string, shardID int, toBeSignedCommittees []incognitokey.CommitteePublicKey) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	validationData, err := consensustypes.DecodeValidationData(data)
	if err != nil {
		return err
	}
	tempToBeSignedCommittees, _ := incognitokey.CommitteeKeyListToString(toBeSignedCommittees)
	signedCommittees := make(map[string]struct{})
	for _, idx := range validationData.ValidatiorsIdx {
		if idx >= len(tempToBeSignedCommittees) {
			return fmt.Errorf("Idx = %+v, greater than len(toBeSignedCommittees) = %+v", idx, len(tempToBeSignedCommittees))
		}
		signedCommittees[tempToBeSignedCommittees[idx]] = struct{}{}
	}
	for _, toBeSignedCommittee := range tempToBeSignedCommittees {
		missingSignature, ok := s.missingSignature[toBeSignedCommittee]
		if !ok {
			// skip toBeSignedCommittee not in current list
			continue
		}
		missingSignature.ActualTotal++
		if _, ok := signedCommittees[toBeSignedCommittee]; !ok {
			missingSignature.Missing++
		}
		s.missingSignature[toBeSignedCommittee] = missingSignature
	}
	s.lastShardStateValidatorCommittee[shardID] = tempToBeSignedCommittees
	s.lastShardStateValidatorIndex[shardID] = validationData.ValidatiorsIdx
	return nil
}

func (s *MissingSignatureCounter) AddPreviousMissignSignature(data string, shardID int) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	tempToBeSignedCommittees := s.lastShardStateValidatorCommittee[shardID]
	prevShardStateValidatorIndex := s.lastShardStateValidatorIndex[shardID]
	if len(prevShardStateValidatorIndex) == 0 || len(tempToBeSignedCommittees) == 0 {
		return nil
	}

	prevValidationData, err := consensustypes.DecodeValidationData(data)
	if err != nil {
		return err
	}
	uncountCommittees := make(map[string]struct{})

	if len(prevValidationData.ValidatiorsIdx) <= len(prevShardStateValidatorIndex) {
		return nil
	}

	//find index that is not count in previous validator index
	for _, idx := range prevValidationData.ValidatiorsIdx {
		if idx >= len(tempToBeSignedCommittees) {
			return fmt.Errorf("Idx = %+v, greater than len(toBeSignedCommittees) = %+v", idx, len(tempToBeSignedCommittees))
		}
		if common.IndexOfInt(idx, prevShardStateValidatorIndex) == -1 {
			uncountCommittees[tempToBeSignedCommittees[idx]] = struct{}{}
		}
	}

	//revert missing counter
	for _, toBeSignedCommittee := range tempToBeSignedCommittees {
		missingSignature, ok := s.missingSignature[toBeSignedCommittee]
		if !ok {
			// skip toBeSignedCommittee not in current list
			continue
		}
		if _, ok := uncountCommittees[toBeSignedCommittee]; ok {
			if missingSignature.Missing > 0 {
				missingSignature.Missing--
			}
		}
		s.missingSignature[toBeSignedCommittee] = missingSignature
	}

	return nil
}

func (s MissingSignatureCounter) GetAllSlashingPenaltyWithActualTotalBlock() map[string]Penalty {
	s.lock.Lock()
	defer s.lock.Unlock()

	penalties := make(map[string]Penalty)
	for key, numberOfMissingSig := range s.missingSignature {
		votedPercent, penalty := getSlashingPenalty(numberOfMissingSig.Missing, numberOfMissingSig.ActualTotal, s.penalties)
		if !penalty.IsEmpty() {
			penalties[key] = penalty
		}
		numberOfMissingSig.VotePercent = votedPercent
		s.missingSignature[key] = numberOfMissingSig
	}
	return penalties
}

func (s MissingSignatureCounter) GetAllSlashingPenaltyWithExpectedTotalBlock(expectedTotalBlocks map[string]uint) map[string]Penalty {
	s.lock.Lock()
	defer s.lock.Unlock()

	penalties := make(map[string]Penalty)
	for key, expectedTotalBlock := range expectedTotalBlocks {
		var penalty Penalty
		var votedPercent uint
		missingSignature, ok := s.missingSignature[key]
		if !ok {
			votedPercent, penalty = getSlashingPenalty(expectedTotalBlock, expectedTotalBlock, s.penalties)
		} else {
			signedBlock := missingSignature.ActualTotal - missingSignature.Missing
			missingBlock := uint(0)
			if signedBlock > expectedTotalBlock {
				missingBlock = 0
			} else {
				missingBlock = expectedTotalBlock - signedBlock
			}
			votedPercent, penalty = getSlashingPenalty(missingBlock, expectedTotalBlock, s.penalties)
		}
		if !penalty.IsEmpty() {
			penalties[key] = penalty
		}
		missingSignature.VotePercent = votedPercent
		s.missingSignature[key] = missingSignature
	}

	return penalties
}

func getSlashingPenalty(
	numberOfMissingSig, total uint,
	penalties []Penalty,
) (
	uint,
	Penalty,
) {
	votePercent := uint(100)
	penalty := NewPenalty()
	if total == 0 {
		return votePercent, penalty
	}
	missedPercent := numberOfMissingSig * 100 / total
	votePercent -= missedPercent
	for _, penaltyLevel := range penalties {
		if missedPercent >= penaltyLevel.MinPercent {
			penalty = penaltyLevel
		}
	}
	return votePercent, penalty
}

func (s *MissingSignatureCounter) Reset(committees []string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	missingSignature := make(map[string]MissingSignature)
	for _, v := range committees {
		missingSignature[v] = NewMissingSignature()
	}

	s.missingSignature = missingSignature
	s.lastShardStateValidatorCommittee = make(map[int][]string)
	s.lastShardStateValidatorIndex = make(map[int][]int)
}

func (s *MissingSignatureCounter) CommitteeChange(newCommittees []string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	missingSignature := make(map[string]MissingSignature)
	for _, v := range newCommittees {
		res, ok := s.missingSignature[v]
		if !ok {
			missingSignature[v] = NewMissingSignature()
		} else {
			missingSignature[v] = res
		}
	}

	s.missingSignature = missingSignature
}

func (s *MissingSignatureCounter) Copy() IMissingSignatureCounter {
	s.lock.RLock()
	defer s.lock.RUnlock()

	newS := &MissingSignatureCounter{
		missingSignature:                 make(map[string]MissingSignature),
		penalties:                        make([]Penalty, len(s.penalties)),
		lastShardStateValidatorCommittee: make(map[int][]string),
		lastShardStateValidatorIndex:     make(map[int][]int),
		lock:                             new(sync.RWMutex),
	}
	copy(newS.penalties, s.penalties)
	for sid, _ := range s.lastShardStateValidatorCommittee {
		newS.lastShardStateValidatorCommittee[sid] = make([]string, len(s.lastShardStateValidatorCommittee[sid]))
		newS.lastShardStateValidatorIndex[sid] = make([]int, len(s.lastShardStateValidatorIndex[sid]))
		copy(newS.lastShardStateValidatorCommittee[sid], s.lastShardStateValidatorCommittee[sid])
		copy(newS.lastShardStateValidatorIndex[sid], s.lastShardStateValidatorIndex[sid])
	}

	for k, v := range s.missingSignature {
		newS.missingSignature[k] = v
	}
	return newS
}
