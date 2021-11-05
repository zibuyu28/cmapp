// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: cdriver.proto

package driver

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type Node_StateE int32

const (
	Node_Handling Node_StateE = 0
	Node_Normal   Node_StateE = 1
	Node_Abnormal Node_StateE = 2
)

// Enum value maps for Node_StateE.
var (
	Node_StateE_name = map[int32]string{
		0: "Handling",
		1: "Normal",
		2: "Abnormal",
	}
	Node_StateE_value = map[string]int32{
		"Handling": 0,
		"Normal":   1,
		"Abnormal": 2,
	}
)

func (x Node_StateE) Enum() *Node_StateE {
	p := new(Node_StateE)
	*p = x
	return p
}

func (x Node_StateE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Node_StateE) Descriptor() protoreflect.EnumDescriptor {
	return file_cdriver_proto_enumTypes[0].Descriptor()
}

func (Node_StateE) Type() protoreflect.EnumType {
	return &file_cdriver_proto_enumTypes[0]
}

func (x Node_StateE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Node_StateE.Descriptor instead.
func (Node_StateE) EnumDescriptor() ([]byte, []int) {
	return file_cdriver_proto_rawDescGZIP(), []int{0, 0}
}

type Chain_StateE int32

const (
	Chain_Handling Chain_StateE = 0
	Chain_Normal   Chain_StateE = 1
	Chain_Abnormal Chain_StateE = 2
)

// Enum value maps for Chain_StateE.
var (
	Chain_StateE_name = map[int32]string{
		0: "Handling",
		1: "Normal",
		2: "Abnormal",
	}
	Chain_StateE_value = map[string]int32{
		"Handling": 0,
		"Normal":   1,
		"Abnormal": 2,
	}
)

func (x Chain_StateE) Enum() *Chain_StateE {
	p := new(Chain_StateE)
	*p = x
	return p
}

func (x Chain_StateE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Chain_StateE) Descriptor() protoreflect.EnumDescriptor {
	return file_cdriver_proto_enumTypes[1].Descriptor()
}

func (Chain_StateE) Type() protoreflect.EnumType {
	return &file_cdriver_proto_enumTypes[1]
}

func (x Chain_StateE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Chain_StateE.Descriptor instead.
func (Chain_StateE) EnumDescriptor() ([]byte, []int) {
	return file_cdriver_proto_rawDescGZIP(), []int{1, 0}
}

type Node struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string            `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	UUID       string            `protobuf:"bytes,2,opt,name=UUID,proto3" json:"UUID,omitempty"`
	Type       string            `protobuf:"bytes,3,opt,name=Type,proto3" json:"Type,omitempty"`
	State      Node_StateE       `protobuf:"varint,5,opt,name=State,proto3,enum=driver.Node_StateE" json:"State,omitempty"` // 1处理中，2正常，3异常
	Message    string            `protobuf:"bytes,6,opt,name=Message,proto3" json:"Message,omitempty"`
	MachineID  int32             `protobuf:"varint,7,opt,name=MachineID,proto3" json:"MachineID,omitempty"`
	ChainID    int32             `protobuf:"varint,8,opt,name=ChainID,proto3" json:"ChainID,omitempty"`
	Tags       []string          `protobuf:"bytes,9,rep,name=Tags,proto3" json:"Tags,omitempty"`
	CustomInfo map[string]string `protobuf:"bytes,10,rep,name=CustomInfo,proto3" json:"CustomInfo,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Node) Reset() {
	*x = Node{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cdriver_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Node) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Node) ProtoMessage() {}

func (x *Node) ProtoReflect() protoreflect.Message {
	mi := &file_cdriver_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Node.ProtoReflect.Descriptor instead.
func (*Node) Descriptor() ([]byte, []int) {
	return file_cdriver_proto_rawDescGZIP(), []int{0}
}

func (x *Node) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Node) GetUUID() string {
	if x != nil {
		return x.UUID
	}
	return ""
}

func (x *Node) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Node) GetState() Node_StateE {
	if x != nil {
		return x.State
	}
	return Node_Handling
}

func (x *Node) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Node) GetMachineID() int32 {
	if x != nil {
		return x.MachineID
	}
	return 0
}

func (x *Node) GetChainID() int32 {
	if x != nil {
		return x.ChainID
	}
	return 0
}

func (x *Node) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Node) GetCustomInfo() map[string]string {
	if x != nil {
		return x.CustomInfo
	}
	return nil
}

type Chain struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string            `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	UUID       string            `protobuf:"bytes,2,opt,name=UUID,proto3" json:"UUID,omitempty"`
	Type       string            `protobuf:"bytes,3,opt,name=Type,proto3" json:"Type,omitempty"`
	Version    string            `protobuf:"bytes,4,opt,name=Version,proto3" json:"Version,omitempty"`
	State      Chain_StateE      `protobuf:"varint,5,opt,name=State,proto3,enum=driver.Chain_StateE" json:"State,omitempty"` // 1处理中，2正常，3异常
	DriverID   int32             `protobuf:"varint,6,opt,name=DriverID,proto3" json:"DriverID,omitempty"`
	Tags       []string          `protobuf:"bytes,7,rep,name=Tags,proto3" json:"Tags,omitempty"`
	CustomInfo map[string]string `protobuf:"bytes,8,rep,name=CustomInfo,proto3" json:"CustomInfo,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Nodes      []*Node           `protobuf:"bytes,9,rep,name=Nodes,proto3" json:"Nodes,omitempty"`
}

func (x *Chain) Reset() {
	*x = Chain{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cdriver_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Chain) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Chain) ProtoMessage() {}

func (x *Chain) ProtoReflect() protoreflect.Message {
	mi := &file_cdriver_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Chain.ProtoReflect.Descriptor instead.
func (*Chain) Descriptor() ([]byte, []int) {
	return file_cdriver_proto_rawDescGZIP(), []int{1}
}

func (x *Chain) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Chain) GetUUID() string {
	if x != nil {
		return x.UUID
	}
	return ""
}

func (x *Chain) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *Chain) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *Chain) GetState() Chain_StateE {
	if x != nil {
		return x.State
	}
	return Chain_Handling
}

func (x *Chain) GetDriverID() int32 {
	if x != nil {
		return x.DriverID
	}
	return 0
}

func (x *Chain) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *Chain) GetCustomInfo() map[string]string {
	if x != nil {
		return x.CustomInfo
	}
	return nil
}

func (x *Chain) GetNodes() []*Node {
	if x != nil {
		return x.Nodes
	}
	return nil
}

var File_cdriver_proto protoreflect.FileDescriptor

var file_cdriver_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x06, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x1a, 0x0d, 0x6d, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x82, 0x03, 0x0a, 0x04, 0x4e, 0x6f, 0x64, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x55, 0x55, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x55, 0x55, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x29, 0x0a, 0x05, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x64, 0x72, 0x69,
	0x76, 0x65, 0x72, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x52,
	0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x1c, 0x0a, 0x09, 0x4d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x49, 0x44, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x09, 0x4d, 0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x49, 0x44, 0x12, 0x18,
	0x0a, 0x07, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x07, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x61, 0x67, 0x73,
	0x18, 0x09, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x54, 0x61, 0x67, 0x73, 0x12, 0x3c, 0x0a, 0x0a,
	0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x2e, 0x43,
	0x75, 0x73, 0x74, 0x6f, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a,
	0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x3d, 0x0a, 0x0f, 0x43, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x30, 0x0a, 0x06, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x45, 0x12, 0x0c, 0x0a, 0x08, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x69, 0x6e, 0x67, 0x10,
	0x00, 0x12, 0x0a, 0x0a, 0x06, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x01, 0x12, 0x0c, 0x0a,
	0x08, 0x41, 0x62, 0x6e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x02, 0x22, 0x8d, 0x03, 0x0a, 0x05,
	0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x55, 0x55, 0x49,
	0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x55, 0x55, 0x49, 0x44, 0x12, 0x12, 0x0a,
	0x04, 0x54, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x2a, 0x0a, 0x05, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x64, 0x72, 0x69,
	0x76, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45,
	0x52, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x44, 0x72, 0x69, 0x76, 0x65,
	0x72, 0x49, 0x44, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x44, 0x72, 0x69, 0x76, 0x65,
	0x72, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x61, 0x67, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x04, 0x54, 0x61, 0x67, 0x73, 0x12, 0x3d, 0x0a, 0x0a, 0x43, 0x75, 0x73, 0x74, 0x6f,
	0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x64, 0x72,
	0x69, 0x76, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x2e, 0x43, 0x75, 0x73, 0x74, 0x6f,
	0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x43, 0x75, 0x73, 0x74,
	0x6f, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x22, 0x0a, 0x05, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x18,
	0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x4e,
	0x6f, 0x64, 0x65, 0x52, 0x05, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x1a, 0x3d, 0x0a, 0x0f, 0x43, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x30, 0x0a, 0x06, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x45, 0x12, 0x0c, 0x0a, 0x08, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x69, 0x6e, 0x67, 0x10,
	0x00, 0x12, 0x0a, 0x0a, 0x06, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x01, 0x12, 0x0c, 0x0a,
	0x08, 0x41, 0x62, 0x6e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x02, 0x32, 0x95, 0x01, 0x0a, 0x0b,
	0x43, 0x68, 0x61, 0x69, 0x6e, 0x44, 0x72, 0x69, 0x76, 0x65, 0x72, 0x12, 0x2b, 0x0a, 0x09, 0x49,
	0x6e, 0x69, 0x74, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x0d, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65,
	0x72, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72,
	0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x22, 0x00, 0x12, 0x31, 0x0a, 0x0f, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x45, 0x78, 0x65, 0x63, 0x12, 0x0d, 0x2e, 0x64, 0x72,
	0x69, 0x76, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x1a, 0x0d, 0x2e, 0x64, 0x72, 0x69,
	0x76, 0x65, 0x72, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x26, 0x0a, 0x04, 0x45,
	0x78, 0x69, 0x74, 0x12, 0x0d, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x00, 0x42, 0x0a, 0x5a, 0x08, 0x2e, 0x2f, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cdriver_proto_rawDescOnce sync.Once
	file_cdriver_proto_rawDescData = file_cdriver_proto_rawDesc
)

func file_cdriver_proto_rawDescGZIP() []byte {
	file_cdriver_proto_rawDescOnce.Do(func() {
		file_cdriver_proto_rawDescData = protoimpl.X.CompressGZIP(file_cdriver_proto_rawDescData)
	})
	return file_cdriver_proto_rawDescData
}

var file_cdriver_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_cdriver_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_cdriver_proto_goTypes = []interface{}{
	(Node_StateE)(0),  // 0: driver.Node.StateE
	(Chain_StateE)(0), // 1: driver.Chain.StateE
	(*Node)(nil),      // 2: driver.Node
	(*Chain)(nil),     // 3: driver.Chain
	nil,               // 4: driver.Node.CustomInfoEntry
	nil,               // 5: driver.Chain.CustomInfoEntry
	(*Empty)(nil),     // 6: driver.Empty
}
var file_cdriver_proto_depIdxs = []int32{
	0, // 0: driver.Node.State:type_name -> driver.Node.StateE
	4, // 1: driver.Node.CustomInfo:type_name -> driver.Node.CustomInfoEntry
	1, // 2: driver.Chain.State:type_name -> driver.Chain.StateE
	5, // 3: driver.Chain.CustomInfo:type_name -> driver.Chain.CustomInfoEntry
	2, // 4: driver.Chain.Nodes:type_name -> driver.Node
	6, // 5: driver.ChainDriver.InitChain:input_type -> driver.Empty
	3, // 6: driver.ChainDriver.CreateChainExec:input_type -> driver.Chain
	6, // 7: driver.ChainDriver.Exit:input_type -> driver.Empty
	3, // 8: driver.ChainDriver.InitChain:output_type -> driver.Chain
	6, // 9: driver.ChainDriver.CreateChainExec:output_type -> driver.Empty
	6, // 10: driver.ChainDriver.Exit:output_type -> driver.Empty
	8, // [8:11] is the sub-list for method output_type
	5, // [5:8] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_cdriver_proto_init() }
func file_cdriver_proto_init() {
	if File_cdriver_proto != nil {
		return
	}
	file_mdriver_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_cdriver_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Node); i {
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
		file_cdriver_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Chain); i {
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
			RawDescriptor: file_cdriver_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cdriver_proto_goTypes,
		DependencyIndexes: file_cdriver_proto_depIdxs,
		EnumInfos:         file_cdriver_proto_enumTypes,
		MessageInfos:      file_cdriver_proto_msgTypes,
	}.Build()
	File_cdriver_proto = out.File
	file_cdriver_proto_rawDesc = nil
	file_cdriver_proto_goTypes = nil
	file_cdriver_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ChainDriverClient is the client API for ChainDriver service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChainDriverClient interface {
	// InitChain create a chain to store
	InitChain(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Chain, error)
	// CreateChainExec execute create chain action
	CreateChainExec(ctx context.Context, in *Chain, opts ...grpc.CallOption) (*Empty, error)
	// Exit driver exit
	Exit(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type chainDriverClient struct {
	cc grpc.ClientConnInterface
}

func NewChainDriverClient(cc grpc.ClientConnInterface) ChainDriverClient {
	return &chainDriverClient{cc}
}

func (c *chainDriverClient) InitChain(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Chain, error) {
	out := new(Chain)
	err := c.cc.Invoke(ctx, "/driver.ChainDriver/InitChain", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainDriverClient) CreateChainExec(ctx context.Context, in *Chain, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/driver.ChainDriver/CreateChainExec", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chainDriverClient) Exit(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/driver.ChainDriver/Exit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChainDriverServer is the server API for ChainDriver service.
type ChainDriverServer interface {
	// InitChain create a chain to store
	InitChain(context.Context, *Empty) (*Chain, error)
	// CreateChainExec execute create chain action
	CreateChainExec(context.Context, *Chain) (*Empty, error)
	// Exit driver exit
	Exit(context.Context, *Empty) (*Empty, error)
}

// UnimplementedChainDriverServer can be embedded to have forward compatible implementations.
type UnimplementedChainDriverServer struct {
}

func (*UnimplementedChainDriverServer) InitChain(context.Context, *Empty) (*Chain, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitChain not implemented")
}
func (*UnimplementedChainDriverServer) CreateChainExec(context.Context, *Chain) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChainExec not implemented")
}
func (*UnimplementedChainDriverServer) Exit(context.Context, *Empty) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exit not implemented")
}

func RegisterChainDriverServer(s *grpc.Server, srv ChainDriverServer) {
	s.RegisterService(&_ChainDriver_serviceDesc, srv)
}

func _ChainDriver_InitChain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainDriverServer).InitChain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/driver.ChainDriver/InitChain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainDriverServer).InitChain(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChainDriver_CreateChainExec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Chain)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainDriverServer).CreateChainExec(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/driver.ChainDriver/CreateChainExec",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainDriverServer).CreateChainExec(ctx, req.(*Chain))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChainDriver_Exit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChainDriverServer).Exit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/driver.ChainDriver/Exit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChainDriverServer).Exit(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _ChainDriver_serviceDesc = grpc.ServiceDesc{
	ServiceName: "driver.ChainDriver",
	HandlerType: (*ChainDriverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitChain",
			Handler:    _ChainDriver_InitChain_Handler,
		},
		{
			MethodName: "CreateChainExec",
			Handler:    _ChainDriver_CreateChainExec_Handler,
		},
		{
			MethodName: "Exit",
			Handler:    _ChainDriver_Exit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cdriver.proto",
}
