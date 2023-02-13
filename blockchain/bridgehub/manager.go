package bridgehub

type Manager struct {
	state     *BridgeHubState
	producer  stateProducer
	processor stateProcessor
}

func NewManager() *Manager {
	return &Manager{
		state:     NewBrigdeHubState(),
		producer:  stateProducer{},
		processor: stateProcessor{},
	}
}

// func (m *Manager) Clone() *Manager {
// 	return &Manager{
// 		state: m.state.Clone(),
// 	}
// }

func (m *Manager) State() *BridgeHubState {
	return m.state
}

func NewManagerWithValue(state *BridgeHubState) *Manager {
	return &Manager{
		state: state,
	}
}

//TODO:

// func (m *Manager) BuildInstructions(env StateEnvironment) ([][]string, *metadata.AccumulatedValues, error) {
// 	res := [][]string{}
// 	insts := [][]string{}
// 	var err error
// 	ac := env.AccumulatedValues()

// 	// build instruction for convert actions
// 	for shardID, actions := range env.StakeActions() {
// 		for _, action := range actions {
// 			insts, m.state, err = m.producer.convert(action, m.state, env.StateDBs(), byte(shardID))
// 			if err != nil {
// 				return [][]string{}, nil, err
// 			}
// 			res = append(res, insts...)
// 		}
// 	}

// 	return res, ac, nil
// }

// func (m *Manager) BuildNewUnshieldInstructions(stateDB *statedb.StateDB, beaconHeight uint64, unshieldActionForProducers []UnshieldActionForProducer) ([][]string, error) {
// 	res := [][]string{}
// 	insts := [][]string{}
// 	var err error

// 	// build instruction for new unshielding actions
// 	for _, a := range unshieldActionForProducers {
// 		switch a.Action.Meta.GetType() {
// 		case metadataCommon.BurningUnifiedTokenRequestMeta:
// 			insts, m.state, err = m.producer.unshield(a, m.state, beaconHeight, stateDB)
// 			if err != nil {
// 				return [][]string{}, err
// 			}
// 			res = append(res, insts...)
// 		case metadataCommon.BurnForCallRequestMeta:
// 			insts, m.state, err = m.producer.burnForCall(a, m.state, beaconHeight, stateDB)
// 			if err != nil {
// 				return [][]string{}, err
// 			}
// 			res = append(res, insts...)
// 		default:
// 		}
// 	}

// 	Logger.log.Info("bridgeagg new unshield instructions:", res)
// 	return res, nil
// }

// func (m *Manager) Process(insts [][]string, sDB *statedb.StateDB) (map[string]bool, error) {
// 	// init bridge agg param if it's nil
// 	if m.state.param == nil {
// 		m.InitBridgeAggParamDefault()
// 	}

// 	// process insts
// 	updatingInfoByTokenID := map[common.Hash]metadata.UpdatingInfo{}
// 	bridgeAggUnshieldTxIDs := map[string]bool{}
// 	for _, content := range insts {
// 		if len(content) == 0 {
// 			continue // Empty instruction
// 		}
// 		metaType, err := strconv.Atoi(content[0])
// 		if err != nil {
// 			continue // Not error, just not bridgeagg instructions
// 		}
// 		if !metadataBridge.IsBridgeAggMetaType(metaType) {
// 			continue // Not error, just not bridgeagg instructions
// 		}
// 		if metaType == metadataCommon.BridgeAggAddTokenMeta {
// 			m.state, err = m.processor.addToken(content, m.state, sDB)
// 			if err != nil {
// 				return bridgeAggUnshieldTxIDs, err
// 			}
// 			continue
// 		}
// 		if len(content) != 4 {
// 			continue // Not error, just not bridgeagg instructions
// 		}

// 		inst := metadataCommon.NewInstruction()
// 		if err := inst.FromStringSlice(content); err != nil {
// 			return bridgeAggUnshieldTxIDs, err
// 		}

// 		switch inst.MetaType {
// 		case metadataCommon.BridgeAggConvertTokenToUnifiedTokenRequestMeta:
// 			m.state, updatingInfoByTokenID, err = m.processor.convert(*inst, m.state, sDB, updatingInfoByTokenID)
// 		case metadataCommon.IssuingUnifiedTokenRequestMeta:
// 			m.state, updatingInfoByTokenID, err = m.processor.shield(*inst, m.state, sDB, updatingInfoByTokenID)
// 		case metadataCommon.BurningUnifiedTokenRequestMeta:
// 			m.state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, err = m.processor.unshield(*inst, m.state, sDB, updatingInfoByTokenID, bridgeAggUnshieldTxIDs)
// 		case metadataCommon.BridgeAggModifyParamMeta:
// 			m.state, err = m.processor.modifyParam(*inst, m.state, sDB)
// 		case metadataCommon.BurnForCallRequestMeta:
// 			m.state, updatingInfoByTokenID, bridgeAggUnshieldTxIDs, err = m.processor.burnForCall(*inst, m.state, sDB, updatingInfoByTokenID, bridgeAggUnshieldTxIDs)
// 		case metadataCommon.IssuingReshieldResponseMeta:
// 			m.state, updatingInfoByTokenID, err = m.processor.reshield(*inst, m.state, sDB, updatingInfoByTokenID)
// 		}
// 		if err != nil {
// 			return bridgeAggUnshieldTxIDs, err
// 		}
// 	}

// 	for _, updatingInfo := range updatingInfoByTokenID {
// 		var updatingAmt uint64
// 		var updatingType string
// 		if updatingInfo.CountUpAmt > updatingInfo.DeductAmt {
// 			updatingAmt = updatingInfo.CountUpAmt - updatingInfo.DeductAmt
// 			updatingType = "+"
// 		}
// 		if updatingInfo.CountUpAmt < updatingInfo.DeductAmt {
// 			updatingAmt = updatingInfo.DeductAmt - updatingInfo.CountUpAmt
// 			updatingType = "-"
// 		}
// 		err := statedb.UpdateBridgeTokenInfo(
// 			sDB,
// 			updatingInfo.TokenID,
// 			updatingInfo.ExternalTokenID,
// 			updatingInfo.IsCentralized,
// 			updatingAmt,
// 			updatingType,
// 		)
// 		if err != nil {
// 			return bridgeAggUnshieldTxIDs, err
// 		}
// 	}
// 	return bridgeAggUnshieldTxIDs, nil
// }

// func (m *Manager) UpdateToDB(sDB *statedb.StateDB, newUnifiedTokens map[common.Hash]bool) error {
// 	// store new unifiedTokens
// 	for unifiedTokenID := range newUnifiedTokens {
// 		err := statedb.StoreBridgeAggUnifiedToken(
// 			sDB,
// 			unifiedTokenID,
// 			statedb.NewBridgeAggUnifiedTokenStateWithValue(unifiedTokenID),
// 		)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// store updated vaults
// 	for unifiedTokenID, vaults := range m.state.unifiedTokenVaults {
// 		for tokenID, vault := range vaults {
// 			err := statedb.StoreBridgeAggVault(sDB, unifiedTokenID, tokenID, vault)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// store new waiting unshield reqs
// 	for unifiedTokenID, unshieldReqs := range m.state.newWaitingUnshieldReqs {
// 		for _, req := range unshieldReqs {
// 			err := statedb.StoreBridgeAggWaitingUnshieldReq(sDB, unifiedTokenID, req.GetUnshieldID(), req)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	// store new param
// 	if m.state.param != nil {
// 		err := statedb.StoreBridgeAggParam(sDB, m.state.param)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// delete waiting unshield reqs
// 	err := statedb.DeleteBridgeAggWaitingUnshieldReqs(sDB, m.state.deletedWaitingUnshieldReqKeyHashes)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (m *Manager) GetDiffState(state *State) (*State, map[common.Hash]bool, error) {
// 	return m.state.GetDiff(state)
// }

// func (m *Manager) BuildAddTokenInstruction(beaconHeight uint64, sDBs map[int]*statedb.StateDB, ac *metadata.AccumulatedValues, triggeredFeature map[string]uint64) ([][]string, *metadata.AccumulatedValues, error) {
// 	res := [][]string{}
// 	temp := []string{}
// 	var err error

// 	checkpoints := []uint64{}
// 	validCheckpoints := []uint64{}
// 	for k := range triggeredFeature {
// 		if len(k) > len(unifiedTokenTriggerPrefix) {
// 			if bytes.Equal([]byte(k[:len(unifiedTokenTriggerPrefix)]), []byte(unifiedTokenTriggerPrefix)) {
// 				checkpoint, err := strconv.ParseUint(k[len(unifiedTokenTriggerPrefix):], 10, 64)
// 				if err != nil {
// 					return [][]string{}, ac, err
// 				}
// 				checkpoints = append(checkpoints, checkpoint)
// 			}
// 		}
// 	}
// 	if len(checkpoints) == 0 {
// 		return [][]string{}, ac, nil
// 	}
// 	sort.Slice(checkpoints, func(i, j int) bool {
// 		return checkpoints[i] < checkpoints[j]
// 	})

// 	// after beacon generate autoenablefeature instruction, TriggerFeature will mark the height of the trigger time.
// 	// we only need to process once at the mark time (block after trigger)
// 	//ex: trigger at 8, block 9 will process punified config
// 	for _, checkpoint := range checkpoints {
// 		if beaconHeight == triggeredFeature[unifiedTokenTriggerPrefix+strconv.FormatUint(checkpoint, 10)]+1 {
// 			validCheckpoints = append(validCheckpoints, checkpoint)
// 		}
// 	}
// 	if len(validCheckpoints) == 0 {
// 		return [][]string{}, ac, nil
// 	}
// 	for _, checkpoint := range validCheckpoints {
// 		temp, m.state, ac, err = m.producer.addToken(m.state, beaconHeight, sDBs, ac, checkpoint)
// 		if err != nil {
// 			return res, nil, err
// 		}
// 		if len(temp) != 0 {
// 			res = append(res, temp)
// 		}
// 	}
// 	return res, ac, err
// }

// func (m *Manager) InitBridgeAggParamDefault() error {
// 	if m.state.param != nil {
// 		return errors.New("Can not set bridge agg param to valued param")
// 	}
// 	m.state.param = statedb.NewBridgeAggParamStateWithValue(config.Param().BridgeAggParam.DefaultPercentFeeWithDecimal)
// 	return nil
// }
