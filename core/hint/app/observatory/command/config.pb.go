// Code generated manually for v2raya-core.
// Defines v2ray-compatible observatory command messages.
// Proto package: v2ray.core.app.observatory.command
// Source: hint/app/observatory/command/config.proto

package command

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

// GetOutboundStatusRequest is the v2ray-compatible request for querying observatory.
// Tag field (proto field 1, string) is wire-identical to v2ray's GetOutboundStatusRequest,
// so v2rayA's ObservatoryProducer requests are decoded correctly without protowire extraction.
type GetOutboundStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tag           string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetOutboundStatusRequest) Reset() {
	*x = GetOutboundStatusRequest{}
	mi := &file_hint_app_observatory_command_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}
func (x *GetOutboundStatusRequest) String() string { return protoimpl.X.MessageStringOf(x) }
func (*GetOutboundStatusRequest) ProtoMessage()    {}
func (x *GetOutboundStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_hint_app_observatory_command_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}
func (*GetOutboundStatusRequest) Descriptor() ([]byte, []int) {
	return file_hint_app_observatory_command_config_proto_rawDescGZIP(), []int{0}
}
func (x *GetOutboundStatusRequest) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

// Config is the empty marker config for registering this service in xray's commander.
type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_hint_app_observatory_command_config_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}
func (x *Config) String() string { return protoimpl.X.MessageStringOf(x) }
func (*Config) ProtoMessage()    {}
func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_hint_app_observatory_command_config_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}
func (*Config) Descriptor() ([]byte, []int) {
	return file_hint_app_observatory_command_config_proto_rawDescGZIP(), []int{1}
}

var File_hint_app_observatory_command_config_proto protoreflect.FileDescriptor

// rawDesc encodes the FileDescriptorProto for (len=205):
//
//	syntax = "proto3";
//	package v2ray.core.app.observatory.command;
//	option go_package = "github.com/v2rayA/v2raya-core/hint/app/observatory/command";
//	message GetOutboundStatusRequest { string tag = 1; }
//	message Config {}
var file_hint_app_observatory_command_config_proto_rawDesc = []byte{
	0x0a, 0x29, 0x68, 0x69, 0x6e, 0x74, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72,
	0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2f, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x22, 0x76, 0x32, 0x72,
	0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6f, 0x62, 0x73, 0x65,
	0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22,
	0x2c, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x4f, 0x75, 0x74, 0x62, 0x6f, 0x75, 0x6e, 0x64, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x74,
	0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x61, 0x67, 0x22, 0x08, 0x0a,
	0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x32, 0x72, 0x61, 0x79, 0x41, 0x2f, 0x76, 0x32, 0x72,
	0x61, 0x79, 0x61, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x68, 0x69, 0x6e, 0x74, 0x2f, 0x61, 0x70,
	0x70, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_hint_app_observatory_command_config_proto_rawDescOnce sync.Once
	file_hint_app_observatory_command_config_proto_rawDescData []byte
)

func file_hint_app_observatory_command_config_proto_rawDescGZIP() []byte {
	file_hint_app_observatory_command_config_proto_rawDescOnce.Do(func() {
		file_hint_app_observatory_command_config_proto_rawDescData = protoimpl.X.CompressGZIP(
			file_hint_app_observatory_command_config_proto_rawDesc,
		)
	})
	return file_hint_app_observatory_command_config_proto_rawDescData
}

var file_hint_app_observatory_command_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_hint_app_observatory_command_config_proto_goTypes = []any{
	(*GetOutboundStatusRequest)(nil), // 0: v2ray.core.app.observatory.command.GetOutboundStatusRequest
	(*Config)(nil),                   // 1: v2ray.core.app.observatory.command.Config
}
var file_hint_app_observatory_command_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_hint_app_observatory_command_config_proto_init() }

func file_hint_app_observatory_command_config_proto_init() {
	if File_hint_app_observatory_command_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_hint_app_observatory_command_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_hint_app_observatory_command_config_proto_goTypes,
		DependencyIndexes: file_hint_app_observatory_command_config_proto_depIdxs,
		MessageInfos:      file_hint_app_observatory_command_config_proto_msgTypes,
	}.Build()
	File_hint_app_observatory_command_config_proto = out.File
	file_hint_app_observatory_command_config_proto_goTypes = nil
	file_hint_app_observatory_command_config_proto_depIdxs = nil
}
