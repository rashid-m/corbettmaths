package rpcclient

import "github.com/incognitochain/incognito-chain/common"

type RPCClientLogger struct {
	common.Logger
}

func (self *RPCClientLogger) Init(inst common.Logger) {
	self.Logger = inst
}

// Global instant to use
var Logger = RPCClientLogger{}
