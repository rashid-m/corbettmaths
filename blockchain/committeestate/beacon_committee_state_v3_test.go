package committeestate

import (
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate/mocks"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBeaconCommitteeStateV3_processSwapShardInstruction(t *testing.T) {
	initTestParams()
	initLog()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey, *incKey2, *incKey3, *incKey6,
			*incKey8, *incKey9, *incKey10, *incKey11},
		map[string]privacy.PaymentAddress{
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey.GetIncKeyBase58():   paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey3.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey9.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey11.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key0:  true,
			key:   true,
			key2:  true,
			key3:  true,
			key6:  true,
			key8:  true,
			key9:  true,
			key10: false,
			key11: true,
		},
		map[string]common.Hash{
			key0:  *hash,
			key:   *hash,
			key2:  *hash,
			key3:  *hash,
			key6:  *hash,
			key8:  *hash,
			key9:  *hash,
			key10: *hash,
			key11: *hash,
		},
	)
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		swapShardInstruction     *instruction.SwapShardInstruction
		env                      *BeaconCommitteeStateEnvironment
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
		oldState                 BeaconCommitteeState
		newState                 BeaconCommitteeState
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
		want1              *instruction.ReturnStakeInstruction
		wantErr            bool
	}{
		//TODO: @hung add testcase
		// Testcase 1: no swap out, only swap in
		// Testcase 2: only swap out, no swap in
		// Testcase 3: only slash, no swap in
		// Testcase 4: only slash, swap in
		// Testcase 5: only normal swap out, no swap in, mix stop auto stake = true && false
		// Testcase 6: only normal swap out, swap in, mix stop auto stake = true && false
		// Testcase 7: > 1 slash and > 1 normal swap out, swap in, mix stop auto stake = true && false
		// Testcase 8: one slash and one normal swap out, swap in 1
		// Testcase 9: normal swap out, swap in, one node, stop autostake = false
		// Testcase 10: normal swap out, swap in, one node, stop autostake = true
		// Testcase 11: > 1 slash and > 1 normal swap out, swap in > 1, mix stop auto stake = true && false
		{
			name: "[valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key4, key5, key6, key7,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key8, key9, key10, key11,
							},
							1: []string{
								key12, key13, key14, key15,
							},
						},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  false,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: true,
							key11: false,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
				syncPool: map[byte][]string{
					0: []string{key16, key17},
					1: []string{key18, key19},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key4, key5, key7, key12,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key8, key9, key10, key11,
							},
							1: []string{
								key13, key6, key14, key15,
							},
						},
						autoStake: map[string]bool{
							key0:  true,
							key:   true,
							key2:  true,
							key3:  false,
							key6:  true,
							key8:  true,
							key9:  true,
							key10: true,
							key11: false,
						},
					},
					swapRule: NewSwapRuleV3(),
				},
				syncPool: map[byte][]string{
					0: []string{key16, key17},
					1: []string{key18, key19},
				},
			},
			args: args{
				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
				env: &BeaconCommitteeStateEnvironment{
					ConsensusStateDB:                 sDB,
					BeaconHeight:                     500000,
					ActiveShards:                     2,
					RandomNumber:                     1,
					MinShardCommitteeSize:            4,
					MaxShardCommitteeSize:            4,
					NumberOfFixedShardBlockValidator: 2,
					MissingSignaturePenalty:          map[string]signaturecounter.Penalty{},
					shardCommittee: map[byte][]string{
						0: []string{
							key0, key, key2, key3,
						},
						1: []string{
							key4, key5, key6, key7,
						},
					},
					shardSubstitute: map[byte][]string{
						0: []string{
							key8, key9, key10, key11,
						},
						1: []string{
							key12, key13, key14, key15,
						},
					},
				},
				swapShardInstruction: &instruction.SwapShardInstruction{
					InPublicKeys:        []string{key12},
					InPublicKeyStructs:  []incognitokey.CommitteePublicKey{*incKey12},
					OutPublicKeys:       []string{key6},
					OutPublicKeyStructs: []incognitokey.CommitteePublicKey{*incKey6},
					ChainID:             1,
				},
				committeeChange: &CommitteeChange{
					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
					ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeRemoved:  map[byte][]incognitokey.CommitteePublicKey{},
					ShardCommitteeAdded:    map[byte][]incognitokey.CommitteePublicKey{},
					SlashingCommittee: map[byte][]string{
						1: []string{},
					},
				},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							shardCommittee: map[byte][]string{
								0: []string{
									key0, key, key2, key3,
								},
								1: []string{
									key4, key5, key6, key7,
								},
							},
							shardSubstitute: map[byte][]string{
								0: []string{
									key8, key9, key10, key11,
								},
								1: []string{
									key12, key13, key14, key15,
								},
							},
							autoStake: map[string]bool{
								key0:  true,
								key:   true,
								key2:  true,
								key3:  false,
								key6:  true,
								key8:  true,
								key9:  true,
								key10: true,
								key11: false,
							},
						},
						swapRule: NewSwapRuleV3(),
					},
					syncPool: map[byte][]string{
						0: []string{key16, key17},
						1: []string{key18, key19},
					},
				},
			},
			want: &CommitteeChange{
				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey6},
				},
				ShardCommitteeRemoved: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey6},
				},
				ShardCommitteeAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey12},
				},
				ShardSubstituteRemoved: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{*incKey12},
				},
				SlashingCommittee: map[byte][]string{
					1: []string{},
				},
			},
			want1:   &instruction.ReturnStakeInstruction{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			got, got1, err := b.processSwapShardInstruction(tt.args.swapShardInstruction, tt.args.env, tt.args.committeeChange, tt.args.returnStakingInstruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() b.syncPool = %v, tt.fieldsAfterProcess.syncPool %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("BeaconCommitteeStateV3.processSwapShardInstruction() b.beaconCommitteeStateV2 = %v, tt.fieldsAfterProcess.beaconCommitteeStateV2 %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAssignWithRandomInstruction(t *testing.T) {

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		rand              int64
		numberOfValidator []int
		committeeChange   *CommitteeChange
		oldState          BeaconCommitteeState
		beaconHeight      uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "valid input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
					},
					shardCommonPool: []string{
						key0, key, key2, key3,
					},
					numberOfAssignedCandidates: 2,
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
					},
					shardCommonPool: []string{
						key2, key3,
					},
					numberOfAssignedCandidates: 0,
				},
				syncPool: map[byte][]string{
					0: []string{
						key,
					},
					1: []string{
						key0,
					},
				},
			},
			args: args{
				rand:              1000,
				numberOfValidator: []int{8, 8},
				committeeChange: &CommitteeChange{
					SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{},
				},
				beaconHeight: 1000,
			},
			want: &CommitteeChange{
				NextEpochShardCandidateRemoved: []incognitokey.CommitteePublicKey{
					*incKey0, *incKey,
				},
				SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
					0: []incognitokey.CommitteePublicKey{*incKey},
					1: []incognitokey.CommitteePublicKey{*incKey0},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.processAssignWithRandomInstruction(tt.args.rand, tt.args.numberOfValidator, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processAssignWithRandomInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b = %v, tt.fieldsAfterProcess %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
			}
		})
	}
}

//func TestBeaconCommitteeStateV3_processAfterNormalSwap(t *testing.T) {
//	initTestParams()
//	initLog()
//
//	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
//
//	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
//	assert.Nil(t, err)
//
//	hash, err := common.Hash{}.NewHashFromStr("123")
//	statedb.StoreStakerInfo(
//		sDB,
//		[]incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey12, *incKey13},
//		map[string]privacy.PaymentAddress{
//			incKey10.GetIncKeyBase58(): paymentAddress,
//			incKey11.GetIncKeyBase58(): paymentAddress,
//			incKey12.GetIncKeyBase58(): paymentAddress,
//			incKey13.GetIncKeyBase58(): paymentAddress,
//		},
//		map[string]bool{
//			key10: true,
//			key11: true,
//			key12: false,
//			key13: true,
//		},
//		map[string]common.Hash{
//			key10: *hash,
//			key11: *hash,
//			key12: *hash,
//			key13: *hash,
//		},
//	)
//
//	type fields struct {
//		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
//		syncPool                         map[byte][]string
//	}
//	type args struct {
//		env                      *BeaconCommitteeStateEnvironment
//		outPublicKeys            []string
//		committeeChange          *CommitteeChange
//		returnStakingInstruction *instruction.ReturnStakeInstruction
//		oldState                 BeaconCommitteeState
//		newState                 BeaconCommitteeState
//	}
//	tests := []struct {
//		name               string
//		fields             fields
//		fieldsAfterProcess fields
//		args               args
//		want               *CommitteeChange
//		want1              *instruction.ReturnStakeInstruction
//		wantErr            bool
//	}{
//		{
//			name: "[valid input]",
//			fields: fields{
//				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
//					beaconCommitteeStateBase: beaconCommitteeStateBase{
//						shardSubstitute: map[byte][]string{
//							0: []string{
//								key0, key, key2, key3,
//							},
//							1: []string{
//								key4, key5, key6, key7,
//							},
//						},
//						autoStake: map[string]bool{
//							key10: true,
//							key11: true,
//							key12: false,
//							key13: true,
//						},
//					},
//				},
//				syncPool: map[byte][]string{
//					0: []string{key8},
//					1: []string{key9},
//				},
//			},
//			fieldsAfterProcess: fields{
//				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
//					beaconCommitteeStateBase: beaconCommitteeStateBase{
//						shardSubstitute: map[byte][]string{
//							0: []string{
//								key0, key11, key, key2, key13, key10, key3,
//							},
//							1: []string{
//								key4, key5, key6, key7,
//							},
//						},
//						autoStake: map[string]bool{
//							key10: true,
//							key11: true,
//							key13: true,
//						},
//					},
//				},
//				syncPool: map[byte][]string{
//					0: []string{key8},
//					1: []string{key9},
//				},
//			},
//			args: args{
//				returnStakingInstruction: &instruction.ReturnStakeInstruction{},
//				env: &BeaconCommitteeStateEnvironment{
//					ConsensusStateDB: sDB,
//					BeaconHeight:     500000,
//					ShardID:          0,
//					ActiveShards:     2,
//					RandomNumber:     1,
//				},
//				outPublicKeys: []string{key10, key11, key12, key13},
//				committeeChange: &CommitteeChange{
//					ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{},
//				},
//				oldState: &BeaconCommitteeStateV3{
//					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
//						beaconCommitteeStateBase: beaconCommitteeStateBase{
//							shardSubstitute: map[byte][]string{
//								0: []string{
//									key0, key, key2, key3,
//								},
//								1: []string{
//									key4, key5, key6, key7,
//								},
//							},
//							autoStake: map[string]bool{
//								key10: true,
//								key11: true,
//								key12: false,
//								key13: true,
//							},
//						},
//					},
//					syncPool: map[byte][]string{
//						0: []string{key8},
//						1: []string{key9},
//					},
//				},
//			},
//			want: &CommitteeChange{
//				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
//					0: []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey13},
//				},
//				RemovedStaker: []string{key12},
//			},
//			want1: &instruction.ReturnStakeInstruction{
//				PublicKeys:       []string{key12},
//				PublicKeysStruct: []incognitokey.CommitteePublicKey{*incKey12},
//				StakingTXIDs:     []string{hash.String()},
//				StakingTxHashes:  []common.Hash{*hash},
//				PercentReturns:   []uint{100},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			b := &BeaconCommitteeStateV3{
//				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
//				syncPool:                         tt.fields.syncPool,
//			}
//			got, got1, err := b.processAfterNormalSwap(tt.args.env, tt.args.outPublicKeys, tt.args.committeeChange, tt.args.returnStakingInstruction)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() got = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(got1, tt.want1) {
//				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() got1 = %v, want %v", got1, tt.want1)
//			}
//			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
//				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b.syncPool = %v, tt.fieldsAfterProcess.syncPool %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
//			}
//			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase.shardSubstitute[0], tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute[0]) {
//				t.Errorf("BeaconCommitteeStateV3.processAfterNormalSwap() b.beaconCommitteeStateV2 = %v, tt.fieldsAfterProcess.beaconCommitteeStateV2 %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
//			}
//		})
//	}
//}

func TestBeaconCommitteeStateV3_assignRandomlyToSubstituteList(t *testing.T) {

	initTestParams()
	initLog()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		candidates      []string
		rand            int64
		shardID         byte
		committeeChange *CommitteeChange
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		//TODO: @hung
		// testcase 1: substitute list is empty
		// testcase 1: substitute list is not empty
		// testcase 1: candidate = 1
		// testcase 6: candidate > 1
		{
			name: "substitute list is empty",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key2},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}),
		},
		{
			name: "substitute list is not empty",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key3},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key2, key3},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}),
		},
		{
			name: "substitute list is empty, > 1 candidate",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key6, key4, key2},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2, key4, key6},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2, key4, key6}),
		}, {
			name: "substitute list is not empty, > 1 candidate",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key5},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key2, key6, key4, key5},
							1: []string{key},
						},
					},
				},
				syncPool: map[byte][]string{},
			},
			args: args{
				candidates:      []string{key2, key4, key6},
				rand:            1000,
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2, key4, key6}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignRandomlyToSubstituteList(tt.args.candidates, tt.args.rand, tt.args.shardID, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToPending() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.shardSubstitute, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute) {
				t.Errorf("BeaconCommitteeStateV3.assignToPending() b = %v, tt.fieldsAfterProcess %v", b.shardSubstitute, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase.shardSubstitute)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignToSync(t *testing.T) {
	initTestParams()
	initLog()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		shardID         byte
		candidates      []string
		committeeChange *CommitteeChange
		beaconHeight    uint64
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "not empty list candidates",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{},
				syncPool: map[byte][]string{
					0: []string{
						key0, key, key2, key3,
					},
					1: []string{
						key4, key5, key6, key7,
					},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{},
				syncPool: map[byte][]string{
					0: []string{
						key0, key, key2, key3,
					},
					1: []string{
						key4, key5, key6, key7, key8, key9, key10,
					},
				},
			},
			args: args{
				shardID:    1,
				candidates: []string{key8, key9, key10},
				committeeChange: &CommitteeChange{
					SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
						1: []incognitokey.CommitteePublicKey{},
					},
				},
				beaconHeight: 1000,
			},
			want: &CommitteeChange{
				SyncingPoolAdded: map[byte][]incognitokey.CommitteePublicKey{
					1: []incognitokey.CommitteePublicKey{
						*incKey8, *incKey9, *incKey10,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignToSync(tt.args.shardID, tt.args.candidates, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.assignToSync() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
				t.Errorf("BeaconCommitteeStateV3.assignToSync() b = %v, tt.fieldsAfterProcess %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_clone(t *testing.T) {
	initTestParams()
	initLog()

	mutex := &sync.RWMutex{}
	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	txHash, _ := common.Hash{}.NewHashFromStr("123")

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
		finishedSyncValidators           map[byte][]string
	}
	tests := []struct {
		name   string
		fields fields
		want   *BeaconCommitteeStateV3
	}{
		{
			name: "[Valid input]",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{
							key0, key, key2, key3,
						},
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						autoStake: map[string]bool{
							key:  true,
							key0: false,
							key2: true,
							key3: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():  paymentAddress,
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey3.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *txHash,
							key0: *txHash,
							key2: *txHash,
							key3: *txHash,
						},
						mu: mutex,
					},
					shardCommonPool: []string{
						key, key0, key2, key3,
					},
					numberOfAssignedCandidates: 1,
					swapRule:                   NewSwapRuleV3(),
				},
				syncPool: map[byte][]string{
					0: []string{
						key, key0, key2, key3,
					},
					1: []string{
						key, key0, key2, key3,
					},
				},
				finishedSyncValidators: map[byte][]string{
					0: []string{
						key, key0, key2, key3,
					},
					1: []string{
						key, key0, key2, key3,
					},
				},
			},
			want: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{
							key0, key, key2, key3,
						},
						shardCommittee: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						shardSubstitute: map[byte][]string{
							0: []string{
								key0, key, key2, key3,
							},
							1: []string{
								key0, key, key2, key3,
							},
						},
						autoStake: map[string]bool{
							key:  true,
							key0: false,
							key2: true,
							key3: true,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey.GetIncKeyBase58():  paymentAddress,
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey3.GetIncKeyBase58(): paymentAddress,
						},
						stakingTx: map[string]common.Hash{
							key:  *txHash,
							key0: *txHash,
							key2: *txHash,
							key3: *txHash,
						},
						mu: mutex,
					},
					shardCommonPool: []string{
						key, key0, key2, key3,
					},
					numberOfAssignedCandidates: 1,
					swapRule:                   NewSwapRuleV3(),
				},
				syncPool: map[byte][]string{
					0: []string{
						key, key0, key2, key3,
					},
					1: []string{
						key, key0, key2, key3,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.clone(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.clone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processUnstakeInstruction(t *testing.T) {

	initLog()
	initTestParams()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey2, *incKey4, *incKey6, *incKey8, *incKey10, *incKey12,
		},
		map[string]privacy.PaymentAddress{
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey4.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key0:  true,
			key2:  true,
			key4:  true,
			key6:  true,
			key8:  true,
			key10: true,
			key12: true,
		},
		map[string]common.Hash{
			key0:  *hash,
			key2:  *hash,
			key4:  *hash,
			key6:  *hash,
			key8:  *hash,
			key10: *hash,
			key12: *hash,
		},
	)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
		finishedSyncValidators           map[byte][]string
	}
	type args struct {
		unstakeInstruction *instruction.UnstakeInstruction
		env                *BeaconCommitteeStateEnvironment
		committeeChange    *CommitteeChange
		oldState           BeaconCommitteeState
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *CommitteeChange
	}{
		{
			name: "Valid Input",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					shardCommonPool: []string{key0, key},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key2, key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			fieldsAfterProcess: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					shardCommonPool: []string{key0, key},
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{},
						shardCommittee: map[byte][]string{
							0: []string{key2, key3},
							1: []string{key4, key5},
						},
						shardSubstitute: map[byte][]string{
							0: []string{key6, key7},
							1: []string{key8, key9},
						},
						autoStake: map[string]bool{
							key0:  true,
							key2:  true,
							key4:  true,
							key6:  true,
							key8:  true,
							key10: true,
							key12: true,
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key10, key11},
					1: []string{key12, key13},
				},
			},
			args: args{
				unstakeInstruction: &instruction.UnstakeInstruction{
					CommitteePublicKeys: []string{key0, key2, key4, key6, key8, key10, key12, key14},
				},
				env: &BeaconCommitteeStateEnvironment{
					newValidators: []string{
						key0, key, key2, key3, key4, key5, key6,
						key7, key8, key9, key10, key11, key12, key13,
					},
				},
				committeeChange: &CommitteeChange{},
				oldState: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						shardCommonPool: []string{key0, key},
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key2, key3},
								1: []string{key4, key5},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key6, key7},
								1: []string{key8, key9},
							},
							autoStake: map[string]bool{
								key0:  true,
								key2:  true,
								key4:  true,
								key6:  true,
								key8:  true,
								key10: true,
								key12: true,
							},
						},
					},
					syncPool: map[byte][]string{
						0: []string{key10, key11},
						1: []string{key12, key13},
					},
				},
			},
			want: &CommitteeChange{
				StopAutoStake: []string{key0, key2, key4, key6, key8, key10, key12},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			got := b.processUnstakeInstruction(tt.args.unstakeInstruction, tt.args.env, tt.args.committeeChange, tt.args.oldState)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.processUnstakeInstruction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestBeaconCommitteeStateV3_processFinishSyncInstruction(t *testing.T) {
//
//	initLog()
//	initTestParams()
//
//	type fields struct {
//		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
//		syncPool                         map[byte][]string
//		finishedSyncValidators           map[byte][]string
//	}
//	type args struct {
//		finishSyncInstruction *instruction.FinishSyncInstruction
//		env                   *BeaconCommitteeStateEnvironment
//		committeeChange       *CommitteeChange
//		oldState              BeaconCommitteeState
//	}
//	tests := []struct {
//		name               string
//		fields             fields
//		fieldsAfterProcess fields
//		args               args
//		want               *CommitteeChange
//		wantErr            bool
//	}{
//		{
//			name: "Valid Input",
//			fields: fields{
//				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
//					beaconCommitteeStateBase: beaconCommitteeStateBase{
//						shardSubstitute: map[byte][]string{
//							0: []string{},
//							1: []string{
//								key4, key5, key6, key7,
//							},
//						},
//					},
//				},
//				syncPool: map[byte][]string{
//					0: []string{},
//					1: []string{key0, key, key2, key3},
//				},
//				finishedSyncValidators: map[byte][]string{
//					0: []string{},
//					1: []string{
//						key0, key2,
//					},
//				},
//			},
//			fieldsAfterProcess: fields{
//				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
//					beaconCommitteeStateBase: beaconCommitteeStateBase{
//						shardSubstitute: map[byte][]string{
//							0: []string{},
//							1: []string{
//								key4, key0, key2, key5, key6, key7,
//							},
//						},
//					},
//				},
//				syncPool: map[byte][]string{
//					0: []string{},
//					1: []string{key, key3},
//				},
//				finishedSyncValidators: map[byte][]string{
//					0: []string{},
//					1: []string{},
//				},
//			},
//			args: args{
//				finishSyncInstruction: &instruction.FinishSyncInstruction{
//					ChainID: 1,
//					PublicKeys: []string{
//						key0, key2,
//					},
//					PublicKeysStruct: []incognitokey.CommitteePublicKey{
//						*incKey0, *incKey2,
//					},
//				},
//				env: &BeaconCommitteeStateEnvironment{},
//				committeeChange: &CommitteeChange{
//					SyncingPoolRemoved:     map[byte][]incognitokey.CommitteePublicKey{},
//					ShardSubstituteAdded:   map[byte][]incognitokey.CommitteePublicKey{},
//					FinishedSyncValidators: map[byte][]string{},
//				},
//				oldState: &BeaconCommitteeStateV3{
//					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
//						beaconCommitteeStateBase: beaconCommitteeStateBase{
//							shardSubstitute: map[byte][]string{
//								0: []string{},
//								1: []string{
//									key4, key5, key6, key7,
//								},
//							},
//						},
//					},
//					syncPool: map[byte][]string{
//						0: []string{},
//						1: []string{key0, key, key2, key3},
//					},
//				},
//			},
//			want: &CommitteeChange{
//				SyncingPoolRemoved: map[byte][]incognitokey.CommitteePublicKey{
//					1: []incognitokey.CommitteePublicKey{
//						*incKey0, *incKey2,
//					},
//				},
//				ShardSubstituteAdded: map[byte][]incognitokey.CommitteePublicKey{
//					1: []incognitokey.CommitteePublicKey{
//						*incKey0, *incKey2,
//					},
//				},
//				FinishedSyncValidators: map[byte][]string{
//					1: []string{key0, key2},
//				},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			b := &BeaconCommitteeStateV3{
//				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
//				syncPool:                         tt.fields.syncPool,
//			}
//			got := b.processFinishSyncInstruction(tt.args.finishSyncInstruction, tt.args.env, tt.args.committeeChange)
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("BeaconCommitteeStateV3.processFinishSyncInstruction() = %v, want %v", got, tt.want)
//			}
//			if !reflect.DeepEqual(b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase) {
//				t.Errorf("BeaconCommitteeStateV3.beaconCommitteeStateV2 = %v, want %v", b.beaconCommitteeStateSlashingBase, tt.fieldsAfterProcess.beaconCommitteeStateSlashingBase)
//			}
//			if !reflect.DeepEqual(b.syncPool, tt.fieldsAfterProcess.syncPool) {
//				t.Errorf("BeaconCommitteeStateV3.syncPool = %v, want %v", b.syncPool, tt.fieldsAfterProcess.syncPool)
//			}
//		})
//	}
//}

func TestBeaconCommitteeStateV3_UpdateCommitteeState(t *testing.T) {
	initLog()
	initTestParams()

	finalMutex := &sync.RWMutex{}

	type fields struct {
		BeaconCommitteeStateV3 *BeaconCommitteeStateV3
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, _ := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey6, *incKey7, *incKey17,
		},
		map[string]privacy.PaymentAddress{
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey7.GetIncKeyBase58():  paymentAddress,
			incKey17.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key6:  true,
			key7:  false,
			key17: true,
		},
		map[string]common.Hash{
			key6:  *hash,
			key7:  *hash,
			key17: *hash,
		},
	)

	randomInstructionCommitteeChange := NewCommitteeChange()
	randomInstructionCommitteeChange.NextEpochShardCandidateRemoved =
		append(randomInstructionCommitteeChange.NextEpochShardCandidateRemoved, []incognitokey.CommitteePublicKey{*incKey0, *incKey}...)
	randomInstructionCommitteeChange.SyncingPoolAdded[0] = []incognitokey.CommitteePublicKey{*incKey}
	randomInstructionCommitteeChange.SyncingPoolAdded[1] = []incognitokey.CommitteePublicKey{*incKey0}

	unstakeInstructionCommitteeChange := NewCommitteeChange()
	unstakeInstructionCommitteeChange.StopAutoStake = []string{key}

	finishSyncInstructionCommitteeChange := NewCommitteeChange()
	finishSyncInstructionCommitteeChange.FinishedSyncValidators[1] = []string{key15}
	finishSyncInstructionCommitteeChange.SyncingPoolRemoved[1] = []incognitokey.CommitteePublicKey{*incKey15}
	finishSyncInstructionCommitteeChange.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{*incKey15}

	swapShardInstructionCommitteeChange := NewCommitteeChange()
	swapShardInstructionCommitteeChange.RemovedStaker = []string{key7, key6}
	swapShardInstructionCommitteeChange.SlashingCommittee[1] = []string{key6}
	swapShardInstructionCommitteeChange.ShardSubstituteAdded[1] = []incognitokey.CommitteePublicKey{*incKey17}
	swapShardInstructionCommitteeChange.ShardCommitteeAdded[1] = []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey18}
	swapShardInstructionCommitteeChange.ShardCommitteeRemoved[1] = []incognitokey.CommitteePublicKey{*incKey6, *incKey7, *incKey17}
	swapShardInstructionCommitteeChange.ShardSubstituteRemoved[1] = []incognitokey.CommitteePublicKey{*incKey10, *incKey11, *incKey18}

	swapRule := &mocks.SwapRule{}
	swapRule.On("Process", uint8(1),
		[]string{key6, key7, key17},
		[]string{key10, key11, key18},
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int"),
		map[string]signaturecounter.Penalty{}).Return(
		&instruction.SwapShardInstruction{
			OutPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey6, *incKey7, *incKey17,
			},
			OutPublicKeys: []string{key6, key7, key17},
			InPublicKeyStructs: []incognitokey.CommitteePublicKey{
				*incKey10, *incKey11, *incKey18,
			},
			InPublicKeys: []string{key10, key11, key18},
			ChainID:      1,
			Type:         instruction.SWAP_BY_END_EPOCH,
		},
		[]string{key10, key11, key18}, []string{}, []string{key6}, []string{key7, key17})
	swapRule.On("Version").Return(swapRuleTestVersion)

	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *BeaconCommitteeStateHash
		want1              *CommitteeChange
		want2              [][]string
		wantErr            bool
	}{
		{
			name: "Process Random Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 2,
					},
					syncPool: map[byte][]string{
						0: []string{},
						1: []string{},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake:      map[string]bool{},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key},
						1: []string{key0},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.RANDOM_ACTION,
							"800000",
							"120000",
							"350000",
							"190000",
						},
					},
				},
			},
			want1:   randomInstructionCommitteeChange,
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Unstake Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58():  paymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key:  *hash,
								key7: *hash,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: true,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{
								incKey.GetIncKeyBase58():  paymentAddress,
								incKey7.GetIncKeyBase58(): paymentAddress,
							},
							stakingTx: map[string]common.Hash{
								key:  *hash,
								key7: *hash,
							},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.UNSTAKE_ACTION,
							strings.Join([]string{key, key7}, instruction.SPLITTER),
						},
					},
					newValidators: []string{
						key0, key, key2, key3, key4, key5, key6, key7, key8,
						key9, key10, key11, key12, key13, key14, key15,
					},
				},
			},
			want1:   NewCommitteeChange(),
			want2:   [][]string{},
			wantErr: false,
		},
		{
			name: "Process Finish Sync Instruction",
			fields: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14, key15},
					},
				},
			},
			fieldsAfterProcess: fields{
				BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
					beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
						beaconCommitteeStateBase: beaconCommitteeStateBase{
							mu:              finalMutex,
							beaconCommittee: []string{},
							shardCommittee: map[byte][]string{
								0: []string{key4, key5},
								1: []string{key6, key7},
							},
							shardSubstitute: map[byte][]string{
								0: []string{key8, key9},
								1: []string{key10, key15, key11},
							},
							autoStake: map[string]bool{
								key:  true,
								key7: false,
							},
							rewardReceiver: map[string]privacy.PaymentAddress{},
							stakingTx:      map[string]common.Hash{},
						},
						shardCommonPool: []string{
							key0, key, key2, key3,
						},
						numberOfAssignedCandidates: 0,
					},
					syncPool: map[byte][]string{
						0: []string{key12, key13},
						1: []string{key14},
					},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					ActiveShards: 2,
					BeaconInstructions: [][]string{
						[]string{
							instruction.FINISH_SYNC_ACTION,
							"1",
							strings.Join([]string{key15}, instruction.SPLITTER),
						},
					},
				},
			},
			want1:   finishSyncInstructionCommitteeChange,
			want2:   [][]string{},
			wantErr: false,
		},
		//{
		//	name: "Process Swap Shard Instruction",
		//	fields: fields{
		//		BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
		//			beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
		//				beaconCommitteeStateBase: beaconCommitteeStateBase{
		//					mu:              finalMutex,
		//					beaconCommittee: []string{},
		//					shardCommittee: map[byte][]string{
		//						0: []string{key4, key5, key16},
		//						1: []string{key6, key7, key17},
		//					},
		//					shardSubstitute: map[byte][]string{
		//						0: []string{key8, key9},
		//						1: []string{key10, key11, key18},
		//					},
		//					autoStake: map[string]bool{
		//						key6:  true,  //slashing
		//						key7:  false, //swap out
		//						key17: true,  // return to substitute
		//					},
		//					rewardReceiver: map[string]privacy.PaymentAddress{
		//						incKey6.GetIncKeyBase58():  paymentAddress,
		//						incKey7.GetIncKeyBase58():  paymentAddress,
		//						incKey17.GetIncKeyBase58(): paymentAddress,
		//					},
		//					stakingTx: map[string]common.Hash{
		//						key6:  *hash, //slashing
		//						key7:  *hash, //swap out
		//						key17: *hash, // return to substitute
		//					},
		//				},
		//				shardCommonPool: []string{
		//					key0, key, key2, key3,
		//				},
		//				numberOfAssignedCandidates: 0,
		//				swapRule:                   swapRule,
		//			},
		//			syncPool: map[byte][]string{
		//				0: []string{key12, key13},
		//				1: []string{key14, key15},
		//			},
		//		},
		//	},
		//	fieldsAfterProcess: fields{
		//		BeaconCommitteeStateV3: &BeaconCommitteeStateV3{
		//			beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
		//				beaconCommitteeStateBase: beaconCommitteeStateBase{
		//					mu:              finalMutex,
		//					beaconCommittee: []string{},
		//					shardCommittee: map[byte][]string{
		//						0: []string{key4, key5},
		//						1: []string{key10, key11, key18},
		//					},
		//					shardSubstitute: map[byte][]string{
		//						0: []string{key8, key9},
		//						1: []string{key17},
		//					},
		//					autoStake: map[string]bool{
		//						key17: true,
		//					},
		//					rewardReceiver: map[string]privacy.PaymentAddress{
		//						incKey17.GetIncKeyBase58(): paymentAddress,
		//					},
		//					stakingTx: map[string]common.Hash{
		//						key17: *hash, // return to substitute
		//					},
		//				},
		//				shardCommonPool: []string{
		//					key0, key, key2, key3,
		//				},
		//				numberOfAssignedCandidates: 0,
		//				swapRule:                   swapRule,
		//			},
		//			syncPool: map[byte][]string{
		//				0: []string{key12, key13},
		//				1: []string{key14, key15},
		//			},
		//		},
		//	},
		//	args: args{
		//		env: &BeaconCommitteeStateEnvironment{
		//			ConsensusStateDB: sDB,
		//			ActiveShards:     2,
		//			BeaconInstructions: [][]string{
		//				[]string{
		//					instruction.SWAP_SHARD_ACTION,
		//					strings.Join([]string{key10, key11, key18}, instruction.SPLITTER),
		//					strings.Join([]string{key6, key7, key17}, instruction.SPLITTER),
		//					"1",
		//					strconv.Itoa(instruction.SWAP_BY_END_EPOCH),
		//				},
		//			},
		//			MinShardCommitteeSize:            4,
		//			MaxShardCommitteeSize:            4,
		//			NumberOfFixedShardBlockValidator: 0,
		//			MissingSignaturePenalty: map[string]signaturecounter.Penalty{
		//				key6: signaturecounter.NewPenalty(),
		//			},
		//		},
		//	},
		//	want1: swapShardInstructionCommitteeChange,
		//	want2: [][]string{
		//		[]string{
		//			instruction.RETURN_ACTION,
		//			strings.Join([]string{key7, key6}, instruction.SPLITTER),
		//			strings.Join([]string{hash.String(), hash.String()}, instruction.SPLITTER),
		//			strings.Join([]string{"100", "100"}, instruction.SPLITTER),
		//		},
		//	},
		//	wantErr: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.BeaconCommitteeStateV3
			_, got1, got2, err := b.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
			gotCommitteeState := b
			wantCommitteeState := tt.fieldsAfterProcess.BeaconCommitteeStateV3
			if !reflect.DeepEqual(gotCommitteeState, wantCommitteeState) {
				t.Errorf("got = %v, want %v", gotCommitteeState, wantCommitteeState)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_UpdateCommitteeState_MultipleInstructions(t *testing.T) {
	type fields struct {
		BeaconCommitteeStateV3 BeaconCommitteeStateV3
	}
	type args struct {
		env *BeaconCommitteeStateEnvironment
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess fields
		args               args
		want               *BeaconCommitteeStateHash
		want1              *CommitteeChange
		want2              [][]string
		wantErr            bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.fields.BeaconCommitteeStateV3
			_, got1, got2, err := b.UpdateCommitteeState(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeEngineV3.UpdateCommitteeState() got2 = %v, want %v", got2, tt.want2)
			}
			gotCommitteeState := b
			wantCommitteeState := tt.fieldsAfterProcess.BeaconCommitteeStateV3
			if !reflect.DeepEqual(gotCommitteeState, wantCommitteeState) {
				t.Errorf("got = %v, want %v", gotCommitteeState, wantCommitteeState)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processFinishSyncInstruction(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		finishSyncInstruction *instruction.FinishSyncInstruction
		env                   *BeaconCommitteeStateEnvironment
		committeeChange       *CommitteeChange
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess *BeaconCommitteeStateV3
		args               args
		want               *CommitteeChange
	}{
		{
			name: "remove one validator, sync pool not empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key4, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key4}),
		},
		{
			name: "remove one validator, sync pool is empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key4},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key4, key},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key4}),
		},
		{
			name: "remove multiple validator, sync pool is empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4, key5},
					1: []string{key13, key14},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key3, key0, key4, key5, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key13, key14},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4, key3, key5}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddSyncingPoolRemoved(0, []string{key3}).
				AddSyncingPoolRemoved(0, []string{key5}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key3}).
				AddFinishedSyncValidators(0, []string{key5}).
				AddShardSubstituteAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key3}).
				AddShardSubstituteAdded(0, []string{key5}),
		},
		{
			name: "remove multiple validator, sync pool not empty => assign to pending",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key0, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4, key5},
					1: []string{key13, key14},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key3, key0, key4, key},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key5},
					1: []string{key13, key14},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4, key3}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddSyncingPoolRemoved(0, []string{key3}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key3}).
				AddShardSubstituteAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key3}),
		},
		{
			name: "remove multiple validator, sync pool not empty => pending is empty before assign",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{key3, key4, key5},
					1: []string{key13, key14},
				},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key3, key5, key4},
							1: []string{key10, key11},
						},
					},
				},
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key13, key14},
				},
			},
			args: args{
				env: &BeaconCommitteeStateEnvironment{
					RandomNumber: 1001,
				},
				committeeChange:       NewCommitteeChange(),
				finishSyncInstruction: instruction.NewFinishSyncInstructionWithValue(0, []string{key4, key3, key5}),
			},
			want: NewCommitteeChange().
				AddSyncingPoolRemoved(0, []string{key4}).
				AddSyncingPoolRemoved(0, []string{key3}).
				AddSyncingPoolRemoved(0, []string{key5}).
				AddFinishedSyncValidators(0, []string{key4}).
				AddFinishedSyncValidators(0, []string{key3}).
				AddFinishedSyncValidators(0, []string{key5}).
				AddShardSubstituteAdded(0, []string{key4}).
				AddShardSubstituteAdded(0, []string{key3}).
				AddShardSubstituteAdded(0, []string{key5}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.processFinishSyncInstruction(tt.args.finishSyncInstruction, tt.args.env, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processFinishSyncInstruction() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("processFinishSyncInstruction() fieldsAfterProcess got = %v, want %v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_removeValidatorsFromSyncPool(t *testing.T) {
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		validators []string
		shardID    byte
	}
	tests := []struct {
		name               string
		fields             fields
		fieldsAfterProcess *BeaconCommitteeStateV3
		args               args
	}{
		{
			name: "remove 2 validators, 1 syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key5, key0},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key10, key11, key12},
				},
			},
		},
		{
			name: "remove 1 validators, 2 syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key0},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key5},
					1: []string{key10, key11, key12},
				},
			},
		},
		{
			name: "remove validators not in syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key6},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key0, key5},
					1: []string{key10, key11, key12},
				},
			},
		},
		{
			name: "remove all validators must be removed from syncPool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key, key3, key2, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key5, key0, key2, key, key3},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{},
					1: []string{key10, key11, key12},
				},
			},
		},
		{
			name: "remove 3 validators, 5 sync pool",
			fields: fields{
				syncPool: map[byte][]string{
					0: []string{key0, key, key3, key2, key5},
					1: []string{key10, key11, key12},
				},
			},
			args: args{
				shardID:    0,
				validators: []string{key5, key0, key3},
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				syncPool: map[byte][]string{
					0: []string{key, key2},
					1: []string{key10, key11, key12},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			b.removeValidatorsFromSyncPool(tt.args.validators, tt.args.shardID)
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("removeValidatorsFromSyncPool() got %+v, want %+v", b, tt.fieldsAfterProcess)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_processAfterNormalSwap(t *testing.T) {

	initLog()
	initTestParams()

	paymentAddress := privacy.GeneratePaymentAddress([]byte{1})
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, wrarperDB)
	assert.Nil(t, err)

	hash, err := common.Hash{}.NewHashFromStr("123")
	statedb.StoreStakerInfo(
		sDB,
		[]incognitokey.CommitteePublicKey{
			*incKey0, *incKey2, *incKey4, *incKey6, *incKey8, *incKey10, *incKey12,
		},
		map[string]privacy.PaymentAddress{
			incKey0.GetIncKeyBase58():  paymentAddress,
			incKey2.GetIncKeyBase58():  paymentAddress,
			incKey4.GetIncKeyBase58():  paymentAddress,
			incKey6.GetIncKeyBase58():  paymentAddress,
			incKey8.GetIncKeyBase58():  paymentAddress,
			incKey10.GetIncKeyBase58(): paymentAddress,
			incKey12.GetIncKeyBase58(): paymentAddress,
		},
		map[string]bool{
			key0:  true,
			key2:  true,
			key4:  true,
			key6:  false,
			key8:  false,
			key10: false,
			key12: false,
		},
		map[string]common.Hash{
			key0:  *hash,
			key2:  *hash,
			key4:  *hash,
			key6:  *hash,
			key8:  *hash,
			key10: *hash,
			key12: *hash,
		},
	)

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		env                      *BeaconCommitteeStateEnvironment
		outPublicKeys            []string
		committeeChange          *CommitteeChange
		returnStakingInstruction *instruction.ReturnStakeInstruction
	}
	tests := []struct {
		name                        string
		fields                      fields
		fieldsAfterProcess          *BeaconCommitteeStateV3
		args                        args
		wantCommitteeChange         *CommitteeChange
		want1ReturnStakeInstruction *instruction.ReturnStakeInstruction
		wantErr                     bool
	}{
		{
			name: "1 stop auto stake = false, 1 return stake, no assign back",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						stakingTx: map[string]common.Hash{
							key6: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey6.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key6: false,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						stakingTx:      map[string]common.Hash{},
						rewardReceiver: map[string]privacy.PaymentAddress{},
						autoStake:      map[string]bool{},
					},
				},
				syncPool: make(map[byte][]string),
			},
			args: args{
				committeeChange: NewCommitteeChange(),
				env: &BeaconCommitteeStateEnvironment{
					ShardID:          0,
					ConsensusStateDB: sDB,
				},
				outPublicKeys:            []string{key6},
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddRemovedStaker(key6),
			want1ReturnStakeInstruction: instruction.NewReturnStakeInsWithValue([]string{key6}, []string{hash.String()}),
			wantErr:                     false,
		},
		{
			name: "stop auto stake = true, no return stake, assign back",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key2: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey2.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key2: true,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key, key2},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key2: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey2.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key2: true,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			args: args{
				committeeChange: NewCommitteeChange(),
				env: &BeaconCommitteeStateEnvironment{
					ShardID:          0,
					ConsensusStateDB: sDB,
				},
				outPublicKeys:            []string{key2},
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}),
			want1ReturnStakeInstruction: instruction.NewReturnStakeIns(),
			wantErr:                     false,
		},
		{
			name: "both stop auto stake = true, false, has assign back and return inst",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key0: *hash,
							key2: *hash,
							key6: *hash,
							key8: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
							incKey6.GetIncKeyBase58(): paymentAddress,
							incKey8.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key0: true,
							key2: true,
							key6: false,
							key8: false,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			fieldsAfterProcess: &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: map[byte][]string{
							0: []string{key, key2, key0},
							1: []string{key3},
						},
						stakingTx: map[string]common.Hash{
							key0: *hash,
							key2: *hash,
						},
						rewardReceiver: map[string]privacy.PaymentAddress{
							incKey0.GetIncKeyBase58(): paymentAddress,
							incKey2.GetIncKeyBase58(): paymentAddress,
						},
						autoStake: map[string]bool{
							key0: true,
							key2: true,
						},
					},
				},
				syncPool: make(map[byte][]string),
			},
			args: args{
				committeeChange: NewCommitteeChange(),
				env: &BeaconCommitteeStateEnvironment{
					ShardID:          0,
					ConsensusStateDB: sDB,
				},
				outPublicKeys:            []string{key2, key0, key6, key8},
				returnStakingInstruction: instruction.NewReturnStakeIns(),
			},
			wantCommitteeChange: NewCommitteeChange().
				AddShardSubstituteAdded(0, []string{key2}).
				AddShardSubstituteAdded(0, []string{key0}).
				AddRemovedStaker(key6).
				AddRemovedStaker(key8),
			want1ReturnStakeInstruction: instruction.NewReturnStakeInsWithValue([]string{key6, key8}, []string{hash.String(), hash.String()}),
			wantErr:                     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			got, got1, err := b.processAfterNormalSwap(tt.args.env, tt.args.outPublicKeys, tt.args.committeeChange, tt.args.returnStakingInstruction)
			if (err != nil) != tt.wantErr {
				t.Errorf("processAfterNormalSwap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantCommitteeChange) {
				t.Errorf("processAfterNormalSwap() got = %v, want %v", got, tt.wantCommitteeChange)
			}
			if !reflect.DeepEqual(b, tt.fieldsAfterProcess) {
				t.Errorf("processAfterNormalSwap() fieldsAfterProcess got = %v, want %v", b, tt.fieldsAfterProcess)
			}
			if !reflect.DeepEqual(got1, tt.want1ReturnStakeInstruction) {
				t.Errorf("processAfterNormalSwap() got1 = %v, want %v", got1, tt.want1ReturnStakeInstruction)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_assignBackToSubstituteList(t *testing.T) {
	testcase3CommitteeChange := NewCommitteeChange().AddShardSubstituteAdded(0, []string{key0, key})
	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		candidates      []string
		shardID         byte
		committeeChange *CommitteeChange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *CommitteeChange
	}{
		{
			name: "one candidate, empty committee change",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: make(map[byte][]string),
					},
				},
			},
			args: args{
				candidates: []string{
					key0,
				},
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().AddShardSubstituteAdded(0, []string{key0}),
		},
		{
			name: "two candidate, empty committee change",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: make(map[byte][]string),
					},
				},
			},
			args: args{
				candidates: []string{
					key0, key,
				},
				shardID:         0,
				committeeChange: NewCommitteeChange(),
			},
			want: NewCommitteeChange().AddShardSubstituteAdded(0, []string{key0, key}),
		},
		{
			name: "two candidate, not empty committee change",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						shardSubstitute: make(map[byte][]string),
					},
				},
			},
			args: args{
				candidates: []string{
					key5, key6,
				},
				shardID:         0,
				committeeChange: testcase3CommitteeChange,
			},
			want: testcase3CommitteeChange.AddShardSubstituteAdded(0, []string{key5, key6}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			if got := b.assignBackToSubstituteList(tt.args.candidates, tt.args.shardID, tt.args.committeeChange); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("assignBackToSubstituteList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeaconCommitteeStateV3_SplitReward(t *testing.T) {

	initLog()
	initTestParams()

	type fields struct {
		beaconCommitteeStateSlashingBase beaconCommitteeStateSlashingBase
		syncPool                         map[byte][]string
	}
	type args struct {
		env *SplitRewardEnvironment
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[common.Hash]uint64
		want1   map[common.Hash]uint64
		want2   map[common.Hash]uint64
		want3   map[common.Hash]uint64
		wantErr bool
	}{
		{
			name: "Year 1",
			fields: fields{
				beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
					beaconCommitteeStateBase: beaconCommitteeStateBase{
						beaconCommittee: []string{
							key0, key, key2, key3, key, key2, key3,
						},
						shardCommittee: map[byte][]string{
							0: []string{},
							1: []string{
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3, key4, key5,
								key0, key, key2, key3,
							},
							2: []string{},
							3: []string{},
							4: []string{},
							5: []string{},
							6: []string{},
							7: []string{},
						},
						mu: &sync.RWMutex{},
					},
				},
			},
			args: args{
				env: &SplitRewardEnvironment{
					DAOPercent:                10,
					ShardID:                   1,
					IsSplitRewardForCustodian: false,
					PercentCustodianReward:    0,
					TotalReward: map[common.Hash]uint64{
						common.PRVCoinID: 546998,
					},
					MaxSubsetCommittees: 2,
					SubsetID:            1,
				},
			},
			want: map[common.Hash]uint64{
				common.PRVCoinID: 13104,
			},
			want1: map[common.Hash]uint64{
				common.PRVCoinID: 479195,
			},
			want2: map[common.Hash]uint64{
				common.PRVCoinID: 54699,
			},
			want3:   map[common.Hash]uint64{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeaconCommitteeStateV3{
				beaconCommitteeStateSlashingBase: tt.fields.beaconCommitteeStateSlashingBase,
				syncPool:                         tt.fields.syncPool,
			}
			got, got1, got2, got3, err := b.SplitReward(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("BeaconCommitteeStateV3.SplitReward() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
