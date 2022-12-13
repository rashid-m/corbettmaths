package statedb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

// @NOTE this struct is view object only
type CommitteeData struct {
	bDelegateData []byte
	bLockingData  []byte
	bPerformance  []byte
}

func NewCommitteeData() *CommitteeData {
	return &CommitteeData{}
}

func NewCommitteeDataWithValue(
	delegateData []byte,
	lockingData []byte,
	performance []byte,
) *CommitteeData {
	return &CommitteeData{
		bDelegateData: delegateData,
		bLockingData:  lockingData,
		bPerformance:  performance,
	}
}

func (c CommitteeData) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		BDlg []byte
		BLck []byte
	}{
		BDlg: c.bDelegateData,
		BLck: c.bLockingData,
	})
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (c *CommitteeData) UnmarshalJSON(data []byte) error {
	temp := struct {
		BDlg []byte
		BLck []byte
	}{}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	c.bDelegateData = temp.BDlg
	c.bLockingData = temp.BLck
	return nil
}

func (s *CommitteeData) SetBeaconDelegateData(d []byte) {
	s.bDelegateData = d
}

func (s *CommitteeData) SetBeaconLockingData(l []byte) {
	s.bLockingData = l
}

func (s *CommitteeData) SetBeaconPerformanceData(p []byte) {
	s.bPerformance = p
}

func (s *CommitteeData) BeaconDelegateData() []byte {
	return s.bDelegateData
}

func (s *CommitteeData) BeaconLockingData() []byte {
	return s.bLockingData
}

func (s *CommitteeData) BeaconPerformanceData() []byte {
	return s.bPerformance
}

type CommitteeStateDataObject struct {
	db *StateDB
	// Write caches.
	trie Trie // storage trie, which becomes non-nil on first access

	version    int
	stateHash  common.Hash
	stateInfo  *CommitteeData
	objectType int
	deleted    bool

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error
}

func newCommitteeStateDataObject(db *StateDB, hash common.Hash) *CommitteeStateDataObject {
	return &CommitteeStateDataObject{
		version:    defaultVersion,
		db:         db,
		stateHash:  hash,
		stateInfo:  &CommitteeData{},
		objectType: CommitteeDataObjectType,
		deleted:    false,
	}
}

func newCommitteeStateDataObjectWithValue(db *StateDB, key common.Hash, data interface{}) (*CommitteeStateDataObject, error) {
	var committeeData = NewCommitteeData()
	var ok bool
	var dataBytes []byte
	if dataBytes, ok = data.([]byte); ok {
		err := json.Unmarshal(dataBytes, committeeData)
		if err != nil {
			return nil, err
		}
	} else {
		committeeData, ok = data.(*CommitteeData)
		if !ok {
			return nil, fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
		}
	}
	return &CommitteeStateDataObject{
		version:    defaultVersion,
		stateHash:  key,
		stateInfo:  committeeData,
		db:         db,
		objectType: CommitteeDataObjectType,
		deleted:    false,
	}, nil
}

func (c CommitteeStateDataObject) GetVersion() int {
	return c.version
}

// setError remembers the first non-nil error it is called with.
func (c *CommitteeStateDataObject) SetError(err error) {
	if c.dbErr == nil {
		c.dbErr = err
	}
}

func (c CommitteeStateDataObject) GetTrie(db DatabaseAccessWarper) Trie {
	return c.trie
}

func (c *CommitteeStateDataObject) SetValue(data interface{}) error {
	committeeData, ok := data.(*CommitteeData)
	if !ok {
		return fmt.Errorf("%+v, got type %+v", ErrInvalidStakerInfoType, reflect.TypeOf(data))
	}
	c.stateInfo = committeeData
	return nil
}

func (c CommitteeStateDataObject) GetValue() interface{} {
	return c.stateInfo
}

func (c CommitteeStateDataObject) GetValueBytes() []byte {
	data := c.GetValue()
	value, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal all shard committee")
	}
	return value
}

func (c CommitteeStateDataObject) GetHash() common.Hash {
	return c.stateHash
}

func (c CommitteeStateDataObject) GetType() int {
	return c.objectType
}

// MarkDelete will delete an object in trie
func (c *CommitteeStateDataObject) MarkDelete() {
	c.deleted = true
}

// reset all shard committee value into default value
func (c *CommitteeStateDataObject) Reset() bool {
	c.stateInfo = NewCommitteeData()
	return true
}

func (c CommitteeStateDataObject) IsDeleted() bool {
	return c.deleted
}

// value is either default or nil
func (c CommitteeStateDataObject) IsEmpty() bool {
	temp := NewCommitteeData()
	return reflect.DeepEqual(temp, c.stateInfo) || c.stateInfo == nil
}
