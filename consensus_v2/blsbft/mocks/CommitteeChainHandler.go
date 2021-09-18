// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	common "github.com/incognitochain/incognito-chain/common"
	incognitokey "github.com/incognitochain/incognito-chain/incognitokey"

	mock "github.com/stretchr/testify/mock"

	multiview "github.com/incognitochain/incognito-chain/multiview"
)

// CommitteeChainHandler is an autogenerated mock type for the CommitteeChainHandler type
type CommitteeChainHandler struct {
	mock.Mock
}

// CommitteesFromViewHashForShard provides a mock function with given fields: committeeHash, shardID
func (_m *CommitteeChainHandler) CommitteesFromViewHashForShard(committeeHash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
	ret := _m.Called(committeeHash, shardID)

	var r0 []incognitokey.CommitteePublicKey
	if rf, ok := ret.Get(0).(func(common.Hash, byte) []incognitokey.CommitteePublicKey); ok {
		r0 = rf(committeeHash, shardID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]incognitokey.CommitteePublicKey)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(common.Hash, byte) error); ok {
		r1 = rf(committeeHash, shardID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FinalView provides a mock function with given fields:
func (_m *CommitteeChainHandler) FinalView() multiview.View {
	ret := _m.Called()

	var r0 multiview.View
	if rf, ok := ret.Get(0).(func() multiview.View); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(multiview.View)
		}
	}

	return r0
}