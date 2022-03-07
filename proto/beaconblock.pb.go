// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.10.0
// source: proto/beaconblock.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ShardStateBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ValidationData     []byte `protobuf:"bytes,1,opt,name=ValidationData,proto3" json:"ValidationData,omitempty"`
	CommitteeFromBlock []byte `protobuf:"bytes,2,opt,name=CommitteeFromBlock,proto3" json:"CommitteeFromBlock,omitempty"`
	Height             uint64 `protobuf:"varint,3,opt,name=Height,proto3" json:"Height,omitempty"`
	Hash               []byte `protobuf:"bytes,4,opt,name=Hash,proto3" json:"Hash,omitempty"`
	CrossShard         []byte `protobuf:"bytes,5,opt,name=CrossShard,proto3" json:"CrossShard,omitempty"`
	ProposerTime       int64  `protobuf:"zigzag64,6,opt,name=ProposerTime,proto3" json:"ProposerTime,omitempty"`
	Version            int32  `protobuf:"varint,7,opt,name=Version,proto3" json:"Version,omitempty"`
}

func (x *ShardStateBytes) Reset() {
	*x = ShardStateBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_beaconblock_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShardStateBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShardStateBytes) ProtoMessage() {}

func (x *ShardStateBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_beaconblock_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShardStateBytes.ProtoReflect.Descriptor instead.
func (*ShardStateBytes) Descriptor() ([]byte, []int) {
	return file_proto_beaconblock_proto_rawDescGZIP(), []int{0}
}

func (x *ShardStateBytes) GetValidationData() []byte {
	if x != nil {
		return x.ValidationData
	}
	return nil
}

func (x *ShardStateBytes) GetCommitteeFromBlock() []byte {
	if x != nil {
		return x.CommitteeFromBlock
	}
	return nil
}

func (x *ShardStateBytes) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *ShardStateBytes) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *ShardStateBytes) GetCrossShard() []byte {
	if x != nil {
		return x.CrossShard
	}
	return nil
}

func (x *ShardStateBytes) GetProposerTime() int64 {
	if x != nil {
		return x.ProposerTime
	}
	return 0
}

func (x *ShardStateBytes) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

type ListShardStateBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []*ShardStateBytes `protobuf:"bytes,1,rep,name=Data,proto3" json:"Data,omitempty"`
}

func (x *ListShardStateBytes) Reset() {
	*x = ListShardStateBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_beaconblock_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListShardStateBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListShardStateBytes) ProtoMessage() {}

func (x *ListShardStateBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_beaconblock_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListShardStateBytes.ProtoReflect.Descriptor instead.
func (*ListShardStateBytes) Descriptor() ([]byte, []int) {
	return file_proto_beaconblock_proto_rawDescGZIP(), []int{1}
}

func (x *ListShardStateBytes) GetData() []*ShardStateBytes {
	if x != nil {
		return x.Data
	}
	return nil
}

type BeaconBodyBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ShardState    map[int32]*ListShardStateBytes `protobuf:"bytes,1,rep,name=ShardState,proto3" json:"ShardState,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Instrucstions []*InstrucstionTmp             `protobuf:"bytes,2,rep,name=Instrucstions,proto3" json:"Instrucstions,omitempty"`
}

func (x *BeaconBodyBytes) Reset() {
	*x = BeaconBodyBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_beaconblock_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BeaconBodyBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BeaconBodyBytes) ProtoMessage() {}

func (x *BeaconBodyBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_beaconblock_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BeaconBodyBytes.ProtoReflect.Descriptor instead.
func (*BeaconBodyBytes) Descriptor() ([]byte, []int) {
	return file_proto_beaconblock_proto_rawDescGZIP(), []int{2}
}

func (x *BeaconBodyBytes) GetShardState() map[int32]*ListShardStateBytes {
	if x != nil {
		return x.ShardState
	}
	return nil
}

func (x *BeaconBodyBytes) GetInstrucstions() []*InstrucstionTmp {
	if x != nil {
		return x.Instrucstions
	}
	return nil
}

type BeaconHeaderBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Version                         int32  `protobuf:"varint,1,opt,name=Version,proto3" json:"Version,omitempty"`
	Height                          uint64 `protobuf:"varint,2,opt,name=Height,proto3" json:"Height,omitempty"`
	Epoch                           uint64 `protobuf:"varint,3,opt,name=Epoch,proto3" json:"Epoch,omitempty"`
	Round                           int32  `protobuf:"varint,4,opt,name=Round,proto3" json:"Round,omitempty"`
	Timestamp                       int64  `protobuf:"zigzag64,5,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	PreviousBlockHash               []byte `protobuf:"bytes,6,opt,name=PreviousBlockHash,proto3" json:"PreviousBlockHash,omitempty"`
	InstructionHash                 []byte `protobuf:"bytes,7,opt,name=InstructionHash,proto3" json:"InstructionHash,omitempty"`
	ShardStateHash                  []byte `protobuf:"bytes,8,opt,name=ShardStateHash,proto3" json:"ShardStateHash,omitempty"`
	InstructionMerkleRoot           []byte `protobuf:"bytes,9,opt,name=InstructionMerkleRoot,proto3" json:"InstructionMerkleRoot,omitempty"`
	BeaconCommitteeAndValidatorRoot []byte `protobuf:"bytes,10,opt,name=BeaconCommitteeAndValidatorRoot,proto3" json:"BeaconCommitteeAndValidatorRoot,omitempty"`
	BeaconCandidateRoot             []byte `protobuf:"bytes,11,opt,name=BeaconCandidateRoot,proto3" json:"BeaconCandidateRoot,omitempty"`
	ShardCandidateRoot              []byte `protobuf:"bytes,12,opt,name=ShardCandidateRoot,proto3" json:"ShardCandidateRoot,omitempty"`
	ShardCommitteeAndValidatorRoot  []byte `protobuf:"bytes,13,opt,name=ShardCommitteeAndValidatorRoot,proto3" json:"ShardCommitteeAndValidatorRoot,omitempty"`
	AutoStakingRoot                 []byte `protobuf:"bytes,14,opt,name=AutoStakingRoot,proto3" json:"AutoStakingRoot,omitempty"`
	ShardSyncValidatorRoot          []byte `protobuf:"bytes,15,opt,name=ShardSyncValidatorRoot,proto3" json:"ShardSyncValidatorRoot,omitempty"`
	ConsensusType                   string `protobuf:"bytes,16,opt,name=ConsensusType,proto3" json:"ConsensusType,omitempty"`
	Producer                        int32  `protobuf:"varint,17,opt,name=Producer,proto3" json:"Producer,omitempty"`
	Proposer                        int32  `protobuf:"varint,18,opt,name=Proposer,proto3" json:"Proposer,omitempty"`
	ProposeTime                     int64  `protobuf:"zigzag64,19,opt,name=ProposeTime,proto3" json:"ProposeTime,omitempty"`
	FinalityHeight                  uint64 `protobuf:"varint,20,opt,name=FinalityHeight,proto3" json:"FinalityHeight,omitempty"`
}

func (x *BeaconHeaderBytes) Reset() {
	*x = BeaconHeaderBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_beaconblock_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BeaconHeaderBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BeaconHeaderBytes) ProtoMessage() {}

func (x *BeaconHeaderBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_beaconblock_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BeaconHeaderBytes.ProtoReflect.Descriptor instead.
func (*BeaconHeaderBytes) Descriptor() ([]byte, []int) {
	return file_proto_beaconblock_proto_rawDescGZIP(), []int{3}
}

func (x *BeaconHeaderBytes) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *BeaconHeaderBytes) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *BeaconHeaderBytes) GetEpoch() uint64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *BeaconHeaderBytes) GetRound() int32 {
	if x != nil {
		return x.Round
	}
	return 0
}

func (x *BeaconHeaderBytes) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *BeaconHeaderBytes) GetPreviousBlockHash() []byte {
	if x != nil {
		return x.PreviousBlockHash
	}
	return nil
}

func (x *BeaconHeaderBytes) GetInstructionHash() []byte {
	if x != nil {
		return x.InstructionHash
	}
	return nil
}

func (x *BeaconHeaderBytes) GetShardStateHash() []byte {
	if x != nil {
		return x.ShardStateHash
	}
	return nil
}

func (x *BeaconHeaderBytes) GetInstructionMerkleRoot() []byte {
	if x != nil {
		return x.InstructionMerkleRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetBeaconCommitteeAndValidatorRoot() []byte {
	if x != nil {
		return x.BeaconCommitteeAndValidatorRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetBeaconCandidateRoot() []byte {
	if x != nil {
		return x.BeaconCandidateRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetShardCandidateRoot() []byte {
	if x != nil {
		return x.ShardCandidateRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetShardCommitteeAndValidatorRoot() []byte {
	if x != nil {
		return x.ShardCommitteeAndValidatorRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetAutoStakingRoot() []byte {
	if x != nil {
		return x.AutoStakingRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetShardSyncValidatorRoot() []byte {
	if x != nil {
		return x.ShardSyncValidatorRoot
	}
	return nil
}

func (x *BeaconHeaderBytes) GetConsensusType() string {
	if x != nil {
		return x.ConsensusType
	}
	return ""
}

func (x *BeaconHeaderBytes) GetProducer() int32 {
	if x != nil {
		return x.Producer
	}
	return 0
}

func (x *BeaconHeaderBytes) GetProposer() int32 {
	if x != nil {
		return x.Proposer
	}
	return 0
}

func (x *BeaconHeaderBytes) GetProposeTime() int64 {
	if x != nil {
		return x.ProposeTime
	}
	return 0
}

func (x *BeaconHeaderBytes) GetFinalityHeight() uint64 {
	if x != nil {
		return x.FinalityHeight
	}
	return 0
}

type BeaconBlockBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ValidationData []byte             `protobuf:"bytes,1,opt,name=ValidationData,proto3" json:"ValidationData,omitempty"`
	Header         *BeaconHeaderBytes `protobuf:"bytes,2,opt,name=Header,proto3" json:"Header,omitempty"`
	Body           *BeaconBodyBytes   `protobuf:"bytes,3,opt,name=Body,proto3" json:"Body,omitempty"`
}

func (x *BeaconBlockBytes) Reset() {
	*x = BeaconBlockBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_beaconblock_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BeaconBlockBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BeaconBlockBytes) ProtoMessage() {}

func (x *BeaconBlockBytes) ProtoReflect() protoreflect.Message {
	mi := &file_proto_beaconblock_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BeaconBlockBytes.ProtoReflect.Descriptor instead.
func (*BeaconBlockBytes) Descriptor() ([]byte, []int) {
	return file_proto_beaconblock_proto_rawDescGZIP(), []int{4}
}

func (x *BeaconBlockBytes) GetValidationData() []byte {
	if x != nil {
		return x.ValidationData
	}
	return nil
}

func (x *BeaconBlockBytes) GetHeader() *BeaconHeaderBytes {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *BeaconBlockBytes) GetBody() *BeaconBodyBytes {
	if x != nil {
		return x.Body
	}
	return nil
}

var File_proto_beaconblock_proto protoreflect.FileDescriptor

var file_proto_beaconblock_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x62, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x73, 0x68, 0x61, 0x72, 0x64, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xf3, 0x01, 0x0a, 0x0f, 0x53, 0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x2e, 0x0a,
	0x12, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x12, 0x43, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x74, 0x65, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x16, 0x0a,
	0x06, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x48,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x48, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x48, 0x61, 0x73, 0x68, 0x12, 0x1e, 0x0a, 0x0a, 0x43, 0x72, 0x6f,
	0x73, 0x73, 0x53, 0x68, 0x61, 0x72, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x43,
	0x72, 0x6f, 0x73, 0x73, 0x53, 0x68, 0x61, 0x72, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x50, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x12, 0x52,
	0x0c, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x72, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3b, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x24,
	0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x04,
	0x44, 0x61, 0x74, 0x61, 0x22, 0xe0, 0x01, 0x0a, 0x0f, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x42,
	0x6f, 0x64, 0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x40, 0x0a, 0x0a, 0x53, 0x68, 0x61, 0x72,
	0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x42,
	0x65, 0x61, 0x63, 0x6f, 0x6e, 0x42, 0x6f, 0x64, 0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x2e, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a,
	0x53, 0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x36, 0x0a, 0x0d, 0x49, 0x6e,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x73, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x10, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x73, 0x74, 0x69, 0x6f, 0x6e,
	0x54, 0x6d, 0x70, 0x52, 0x0d, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x73, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x1a, 0x53, 0x0a, 0x0f, 0x53, 0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x53, 0x68, 0x61,
	0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xc3, 0x06, 0x0a, 0x11, 0x42, 0x65, 0x61, 0x63,
	0x6f, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x48, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05,
	0x45, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x12, 0x52, 0x09,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x2c, 0x0a, 0x11, 0x50, 0x72, 0x65,
	0x76, 0x69, 0x6f, 0x75, 0x73, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x11, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x48, 0x61, 0x73, 0x68, 0x12, 0x28, 0x0a, 0x0f, 0x49, 0x6e, 0x73, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x61, 0x73, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0f, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x61, 0x73,
	0x68, 0x12, 0x26, 0x0a, 0x0e, 0x53, 0x68, 0x61, 0x72, 0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x48,
	0x61, 0x73, 0x68, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x53, 0x68, 0x61, 0x72, 0x64,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x48, 0x61, 0x73, 0x68, 0x12, 0x34, 0x0a, 0x15, 0x49, 0x6e, 0x73,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f,
	0x6f, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x15, 0x49, 0x6e, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x72, 0x6b, 0x6c, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12,
	0x48, 0x0a, 0x1f, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74,
	0x65, 0x65, 0x41, 0x6e, 0x64, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x6f,
	0x6f, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x1f, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e,
	0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x41, 0x6e, 0x64, 0x56, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x30, 0x0a, 0x13, 0x42, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74,
	0x18, 0x0b, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x13, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x43, 0x61,
	0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x2e, 0x0a, 0x12, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x52, 0x6f, 0x6f,
	0x74, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x12, 0x53, 0x68, 0x61, 0x72, 0x64, 0x43, 0x61,
	0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x46, 0x0a, 0x1e, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x41, 0x6e, 0x64,
	0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x0d, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x1e, 0x53, 0x68, 0x61, 0x72, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74,
	0x74, 0x65, 0x65, 0x41, 0x6e, 0x64, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x52,
	0x6f, 0x6f, 0x74, 0x12, 0x28, 0x0a, 0x0f, 0x41, 0x75, 0x74, 0x6f, 0x53, 0x74, 0x61, 0x6b, 0x69,
	0x6e, 0x67, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x41, 0x75,
	0x74, 0x6f, 0x53, 0x74, 0x61, 0x6b, 0x69, 0x6e, 0x67, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x36, 0x0a,
	0x16, 0x53, 0x68, 0x61, 0x72, 0x64, 0x53, 0x79, 0x6e, 0x63, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x52, 0x6f, 0x6f, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x16, 0x53,
	0x68, 0x61, 0x72, 0x64, 0x53, 0x79, 0x6e, 0x63, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f,
	0x72, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x43, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73,
	0x75, 0x73, 0x54, 0x79, 0x70, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x43, 0x6f,
	0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x18, 0x11, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x50, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x65, 0x72, 0x18, 0x12, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x50, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x65, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x54, 0x69,
	0x6d, 0x65, 0x18, 0x13, 0x20, 0x01, 0x28, 0x12, 0x52, 0x0b, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x26, 0x0a, 0x0e, 0x46, 0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74,
	0x79, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x14, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0e, 0x46,
	0x69, 0x6e, 0x61, 0x6c, 0x69, 0x74, 0x79, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x22, 0x8c, 0x01,
	0x0a, 0x10, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x42, 0x79, 0x74,
	0x65, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x44, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x56, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x12, 0x2a, 0x0a, 0x06, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x42, 0x65, 0x61,
	0x63, 0x6f, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x06,
	0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x42, 0x65, 0x61, 0x63, 0x6f, 0x6e, 0x42, 0x6f, 0x64,
	0x79, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x04, 0x42, 0x6f, 0x64, 0x79, 0x42, 0x09, 0x5a, 0x07,
	0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_beaconblock_proto_rawDescOnce sync.Once
	file_proto_beaconblock_proto_rawDescData = file_proto_beaconblock_proto_rawDesc
)

func file_proto_beaconblock_proto_rawDescGZIP() []byte {
	file_proto_beaconblock_proto_rawDescOnce.Do(func() {
		file_proto_beaconblock_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_beaconblock_proto_rawDescData)
	})
	return file_proto_beaconblock_proto_rawDescData
}

var file_proto_beaconblock_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_proto_beaconblock_proto_goTypes = []interface{}{
	(*ShardStateBytes)(nil),     // 0: ShardStateBytes
	(*ListShardStateBytes)(nil), // 1: ListShardStateBytes
	(*BeaconBodyBytes)(nil),     // 2: BeaconBodyBytes
	(*BeaconHeaderBytes)(nil),   // 3: BeaconHeaderBytes
	(*BeaconBlockBytes)(nil),    // 4: BeaconBlockBytes
	nil,                         // 5: BeaconBodyBytes.ShardStateEntry
	(*InstrucstionTmp)(nil),     // 6: InstrucstionTmp
}
var file_proto_beaconblock_proto_depIdxs = []int32{
	0, // 0: ListShardStateBytes.Data:type_name -> ShardStateBytes
	5, // 1: BeaconBodyBytes.ShardState:type_name -> BeaconBodyBytes.ShardStateEntry
	6, // 2: BeaconBodyBytes.Instrucstions:type_name -> InstrucstionTmp
	3, // 3: BeaconBlockBytes.Header:type_name -> BeaconHeaderBytes
	2, // 4: BeaconBlockBytes.Body:type_name -> BeaconBodyBytes
	1, // 5: BeaconBodyBytes.ShardStateEntry.value:type_name -> ListShardStateBytes
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_proto_beaconblock_proto_init() }
func file_proto_beaconblock_proto_init() {
	if File_proto_beaconblock_proto != nil {
		return
	}
	file_proto_shardblock_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_proto_beaconblock_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShardStateBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_beaconblock_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListShardStateBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_beaconblock_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BeaconBodyBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_beaconblock_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BeaconHeaderBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_beaconblock_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BeaconBlockBytes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_beaconblock_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_beaconblock_proto_goTypes,
		DependencyIndexes: file_proto_beaconblock_proto_depIdxs,
		MessageInfos:      file_proto_beaconblock_proto_msgTypes,
	}.Build()
	File_proto_beaconblock_proto = out.File
	file_proto_beaconblock_proto_rawDesc = nil
	file_proto_beaconblock_proto_goTypes = nil
	file_proto_beaconblock_proto_depIdxs = nil
}