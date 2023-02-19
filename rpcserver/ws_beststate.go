package rpcserver

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (wsServer *WsServer) handleSubscribeShardBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	shardID := byte(arrayParams[0].(float64))
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.ShardBeststateTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Beacon Beststate Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.ShardBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.ShardBestState)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.ShardBestState, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				shardBestState, err := wsServer.blockService.GetShardBestStateByShardID(shardID)
				if err != nil {
					err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
					cResult <- RpcSubResult{Error: err}
					return
				}
				block, err := wsServer.config.BlockChain.ShardChain[shardID].GetBlockByHash(shardBestState.BestBlockHash)
				if err != nil || block == nil {
					err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
					cResult <- RpcSubResult{Error: err}
					return
				}
				shardBestState.BestBlock = block.(*types.ShardBlock)
				shardBestStateResult := jsonresult.NewGetShardBestState(shardBestState)
				cResult <- RpcSubResult{Result: shardBestStateResult, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Shard Beststate"}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubscribeBeaconBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe Beacon Beststate", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 0 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain NO params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.BeaconBeststateTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Beacon Beststate Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.BeaconBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBestState)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBestState, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				allViews := []*blockchain.BeaconBestState{}
				beaconBestState := &blockchain.BeaconBestState{}
				beaconDB := wsServer.blockService.BlockChain.GetBeaconChainDatabase()
				beaconViews, err := rawdbv2.GetBeaconViews(beaconDB)
				if err != nil {
					cResult <- RpcSubResult{Error: rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)}
				}

				err = json.Unmarshal(beaconViews, &allViews)
				if err != nil {
					cResult <- RpcSubResult{Error: rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)}
				}

				for _, v := range allViews {
					// TODO: 0xkraken: why false?
					err := v.RestoreBeaconViewStateFromHash(wsServer.GetBlockchain(), true, false, false, false)
					if err != nil {
						cResult <- RpcSubResult{Error: rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)}
					}
					beaconBestState = v
				}
				beaconBestStateResult := jsonresult.NewGetBeaconBestState(beaconBestState)
				if err != nil {
					err := rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
					cResult <- RpcSubResult{Error: err}
				} else {
					cResult <- RpcSubResult{Result: beaconBestStateResult, Error: nil}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Beacon Beststate"}}
				return
			}
		}
	}
}
func (wsServer *WsServer) handleSubscribeBeaconBestStateFromMem(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe Beacon Beststate", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 0 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain NO params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.BeaconBeststateTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Beacon Beststate Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.BeaconBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				_, ok := msg.Value.(*blockchain.BeaconBestState)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBestState, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				beaconBestStateResult := jsonresult.NewGetBeaconBestState(wsServer.blockService.BlockChain.GetBeaconBestState())
				if err != nil {
					err := rpcservice.NewRPCError(rpcservice.GetClonedBeaconBestStateError, err)
					cResult <- RpcSubResult{Error: err}
				} else {
					cResult <- RpcSubResult{Result: beaconBestStateResult, Error: nil}
				}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Beacon Beststate"}}
				return
			}
		}
	}
}
