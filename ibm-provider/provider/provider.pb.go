// *******************************************************************************
// * IBM Confidential
// * OCO Source Materials
// * IBM Cloud Container Service, 5737-D43
// * (C) Copyright IBM Corp. 2020 All Rights Reserved.
// * The source code for this program is not  published or otherwise divested of
// * its trade secrets, irrespective of what has been deposited with
// * the U.S. Copyright Office.
// ******************************************************************************/

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.24.0
// 	protoc        v3.12.4
// source: provider/provider.proto

package provider

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// The provider type request
type ProviderTypeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *ProviderTypeRequest) Reset() {
	*x = ProviderTypeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_provider_provider_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProviderTypeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProviderTypeRequest) ProtoMessage() {}

func (x *ProviderTypeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_provider_provider_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProviderTypeRequest.ProtoReflect.Descriptor instead.
func (*ProviderTypeRequest) Descriptor() ([]byte, []int) {
	return file_provider_provider_proto_rawDescGZIP(), []int{0}
}

func (x *ProviderTypeRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

// The provider type reply
type ProviderTypeReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *ProviderTypeReply) Reset() {
	*x = ProviderTypeReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_provider_provider_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProviderTypeReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProviderTypeReply) ProtoMessage() {}

func (x *ProviderTypeReply) ProtoReflect() protoreflect.Message {
	mi := &file_provider_provider_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProviderTypeReply.ProtoReflect.Descriptor instead.
func (*ProviderTypeReply) Descriptor() ([]byte, []int) {
	return file_provider_provider_proto_rawDescGZIP(), []int{1}
}

func (x *ProviderTypeReply) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

// The VPC cloud service endpoint request
type VPCSvcEndpointRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *VPCSvcEndpointRequest) Reset() {
	*x = VPCSvcEndpointRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_provider_provider_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VPCSvcEndpointRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VPCSvcEndpointRequest) ProtoMessage() {}

func (x *VPCSvcEndpointRequest) ProtoReflect() protoreflect.Message {
	mi := &file_provider_provider_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VPCSvcEndpointRequest.ProtoReflect.Descriptor instead.
func (*VPCSvcEndpointRequest) Descriptor() ([]byte, []int) {
	return file_provider_provider_proto_rawDescGZIP(), []int{2}
}

func (x *VPCSvcEndpointRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

// The VPC cloud service endpoint reply
type VPCSvcEndpointReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cse string `protobuf:"bytes,1,opt,name=cse,proto3" json:"cse,omitempty"`
}

func (x *VPCSvcEndpointReply) Reset() {
	*x = VPCSvcEndpointReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_provider_provider_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VPCSvcEndpointReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VPCSvcEndpointReply) ProtoMessage() {}

func (x *VPCSvcEndpointReply) ProtoReflect() protoreflect.Message {
	mi := &file_provider_provider_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VPCSvcEndpointReply.ProtoReflect.Descriptor instead.
func (*VPCSvcEndpointReply) Descriptor() ([]byte, []int) {
	return file_provider_provider_proto_rawDescGZIP(), []int{3}
}

func (x *VPCSvcEndpointReply) GetCse() string {
	if x != nil {
		return x.Cse
	}
	return ""
}

var File_provider_provider_proto protoreflect.FileDescriptor

var file_provider_provider_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x22, 0x25, 0x0a, 0x13, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x27, 0x0a, 0x11, 0x50, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x22, 0x27, 0x0a, 0x15, 0x56, 0x50, 0x43, 0x53, 0x76, 0x63, 0x45, 0x6e, 0x64,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x27, 0x0a, 0x13,
	0x56, 0x50, 0x43, 0x53, 0x76, 0x63, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x63, 0x73, 0x65, 0x32, 0xad, 0x01, 0x0a, 0x0b, 0x49, 0x42, 0x4d, 0x50, 0x72, 0x6f,
	0x76, 0x69, 0x64, 0x65, 0x72, 0x12, 0x4f, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x64, 0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64,
	0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x4d, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x56, 0x50, 0x43,
	0x43, 0x53, 0x45, 0x12, 0x1f, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e, 0x56,
	0x50, 0x43, 0x53, 0x76, 0x63, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2e,
	0x56, 0x50, 0x43, 0x53, 0x76, 0x63, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x7d, 0x0a, 0x14, 0x69, 0x6f, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x2e, 0x69, 0x62, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x42, 0x0b, 0x49,
	0x42, 0x4d, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x50, 0x01, 0x5a, 0x56, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x69, 0x62, 0x6d, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x6c, 0x63,
	0x68, 0x65, 0x6d, 0x79, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x2f,
	0x61, 0x72, 0x6d, 0x61, 0x64, 0x61, 0x2d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2d, 0x73,
	0x33, 0x66, 0x73, 0x2d, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x69,
	0x62, 0x6d, 0x2d, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_provider_provider_proto_rawDescOnce sync.Once
	file_provider_provider_proto_rawDescData = file_provider_provider_proto_rawDesc
)

func file_provider_provider_proto_rawDescGZIP() []byte {
	file_provider_provider_proto_rawDescOnce.Do(func() {
		file_provider_provider_proto_rawDescData = protoimpl.X.CompressGZIP(file_provider_provider_proto_rawDescData)
	})
	return file_provider_provider_proto_rawDescData
}

var file_provider_provider_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_provider_provider_proto_goTypes = []interface{}{
	(*ProviderTypeRequest)(nil),   // 0: provider.ProviderTypeRequest
	(*ProviderTypeReply)(nil),     // 1: provider.ProviderTypeReply
	(*VPCSvcEndpointRequest)(nil), // 2: provider.VPCSvcEndpointRequest
	(*VPCSvcEndpointReply)(nil),   // 3: provider.VPCSvcEndpointReply
}
var file_provider_provider_proto_depIdxs = []int32{
	0, // 0: provider.IBMProvider.GetProviderType:input_type -> provider.ProviderTypeRequest
	2, // 1: provider.IBMProvider.GetVPCSvcEndpoint:input_type -> provider.VPCSvcEndpointRequest
	1, // 2: provider.IBMProvider.GetProviderType:output_type -> provider.ProviderTypeReply
	3, // 3: provider.IBMProvider.GetVPCSvcEndpoint:output_type -> provider.VPCSvcEndpointReply
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_provider_provider_proto_init() }
func file_provider_provider_proto_init() {
	if File_provider_provider_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_provider_provider_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProviderTypeRequest); i {
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
		file_provider_provider_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProviderTypeReply); i {
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
		file_provider_provider_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VPCSvcEndpointRequest); i {
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
		file_provider_provider_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VPCSvcEndpointReply); i {
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
			RawDescriptor: file_provider_provider_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_provider_provider_proto_goTypes,
		DependencyIndexes: file_provider_provider_proto_depIdxs,
		MessageInfos:      file_provider_provider_proto_msgTypes,
	}.Build()
	File_provider_provider_proto = out.File
	file_provider_provider_proto_rawDesc = nil
	file_provider_provider_proto_goTypes = nil
	file_provider_provider_proto_depIdxs = nil
}