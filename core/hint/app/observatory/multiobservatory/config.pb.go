// Code generated manually for v2xray merged core.
// source: app/observatory/multiobservatory/config.proto

package multiobservatory

import (
protoreflect "google.golang.org/protobuf/reflect/protoreflect"
protoimpl "google.golang.org/protobuf/runtime/protoimpl"
reflect "reflect"
sync "sync"

)

// ObserverConfig defines a single observatory group with a dedicated balancer tag.
type ObserverConfig struct {
state           protoimpl.MessageState `protogen:"open.v1"`
Tag             string                 `protobuf:"bytes,1,opt,name=tag,proto3" json:"tag,omitempty"`
ProbeUrl        string                 `protobuf:"bytes,2,opt,name=probe_url,json=probeUrl,proto3" json:"probe_url,omitempty"`
ProbeInterval   int64                  `protobuf:"varint,3,opt,name=probe_interval,json=probeInterval,proto3" json:"probe_interval,omitempty"`
SubjectSelector []string               `protobuf:"bytes,4,rep,name=subject_selector,json=subjectSelector,proto3" json:"subject_selector,omitempty"`
unknownFields   protoimpl.UnknownFields
sizeCache       protoimpl.SizeCache
}

func (x *ObserverConfig) Reset() {
*x = ObserverConfig{}
mi := &file_app_observatory_multiobservatory_config_proto_msgTypes[0]
ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
ms.StoreMessageInfo(mi)
}
func (x *ObserverConfig) String() string         { return protoimpl.X.MessageStringOf(x) }
func (*ObserverConfig) ProtoMessage()             {}

func (x *ObserverConfig) ProtoReflect() protoreflect.Message {
mi := &file_app_observatory_multiobservatory_config_proto_msgTypes[0]
if x != nil {
ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
if ms.LoadMessageInfo() == nil {
ms.StoreMessageInfo(mi)
}
return ms
}
return mi.MessageOf(x)
}
func (*ObserverConfig) Descriptor() ([]byte, []int) {
return file_app_observatory_multiobservatory_config_proto_rawDescGZIP(), []int{0}
}
func (x *ObserverConfig) GetTag() string {
if x != nil { return x.Tag }; return ""
}
func (x *ObserverConfig) GetProbeUrl() string {
if x != nil { return x.ProbeUrl }; return ""
}
func (x *ObserverConfig) GetProbeInterval() int64 {
if x != nil { return x.ProbeInterval }; return 0
}
func (x *ObserverConfig) GetSubjectSelector() []string {
if x != nil { return x.SubjectSelector }; return nil
}

// Config is the top-level multiObservatory feature config.
// Each Observers entry creates an independent observatory for a balancer group.
type Config struct {
state         protoimpl.MessageState `protogen:"open.v1"`
Observers     []*ObserverConfig      `protobuf:"bytes,1,rep,name=observers,proto3" json:"observers,omitempty"`
unknownFields protoimpl.UnknownFields
sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
*x = Config{}
mi := &file_app_observatory_multiobservatory_config_proto_msgTypes[1]
ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
ms.StoreMessageInfo(mi)
}
func (x *Config) String() string  { return protoimpl.X.MessageStringOf(x) }
func (*Config) ProtoMessage()      {}

func (x *Config) ProtoReflect() protoreflect.Message {
mi := &file_app_observatory_multiobservatory_config_proto_msgTypes[1]
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
return file_app_observatory_multiobservatory_config_proto_rawDescGZIP(), []int{1}
}
func (x *Config) GetObservers() []*ObserverConfig {
if x != nil { return x.Observers }; return nil
}

var File_app_observatory_multiobservatory_config_proto protoreflect.FileDescriptor

// rawDesc is the serialized FileDescriptorProto for this proto file (len=409).
var file_app_observatory_multiobservatory_config_proto_rawDesc = []byte{
0x0a, 0x2d, 0x61, 0x70, 0x70, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72,
0x79, 0x2f, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f,
0x72, 0x79, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
0x2a, 0x78, 0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6f,
0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x6d, 0x75, 0x6c, 0x74, 0x69,
0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x22, 0x91, 0x01, 0x0a, 0x0e,
0x4f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x10,
0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x74, 0x61, 0x67,
0x12, 0x1b, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20,
0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x25, 0x0a,
0x0e, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18,
0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x70, 0x72, 0x6f, 0x62, 0x65, 0x49, 0x6e, 0x74, 0x65,
0x72, 0x76, 0x61, 0x6c, 0x12, 0x29, 0x0a, 0x10, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f,
0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f,
0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x53, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x22,
0x62, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x58, 0x0a, 0x09, 0x6f, 0x62, 0x73,
0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3a, 0x2e, 0x78,
0x72, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x61, 0x70, 0x70, 0x2e, 0x6f, 0x62, 0x73,
0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x6f, 0x62,
0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x4f, 0x62, 0x73, 0x65, 0x72, 0x76,
0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x09, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76,
0x65, 0x72, 0x73, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
0x6d, 0x2f, 0x78, 0x74, 0x6c, 0x73, 0x2f, 0x78, 0x72, 0x61, 0x79, 0x2d, 0x63, 0x6f, 0x72, 0x65,
0x2f, 0x61, 0x70, 0x70, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72, 0x79,
0x2f, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x61, 0x74, 0x6f, 0x72,
0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
file_app_observatory_multiobservatory_config_proto_rawDescOnce sync.Once
file_app_observatory_multiobservatory_config_proto_rawDescData []byte
)

func file_app_observatory_multiobservatory_config_proto_rawDescGZIP() []byte {
file_app_observatory_multiobservatory_config_proto_rawDescOnce.Do(func() {
file_app_observatory_multiobservatory_config_proto_rawDescData = protoimpl.X.CompressGZIP(
file_app_observatory_multiobservatory_config_proto_rawDesc,
)
})
return file_app_observatory_multiobservatory_config_proto_rawDescData
}

var file_app_observatory_multiobservatory_config_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_app_observatory_multiobservatory_config_proto_goTypes = []any{
(*ObserverConfig)(nil), // 0: xray.core.app.observatory.multiobservatory.ObserverConfig
(*Config)(nil),         // 1: xray.core.app.observatory.multiobservatory.Config
}
var file_app_observatory_multiobservatory_config_proto_depIdxs = []int32{
0, // 0: xray.core.app.observatory.multiobservatory.Config.observers:type_name -> xray.core.app.observatory.multiobservatory.ObserverConfig
1, // [1:1] is the sub-list for method output_type
1, // [1:1] is the sub-list for method input_type
1, // [1:1] is the sub-list for extension type_name
1, // [1:1] is the sub-list for extension extendee
0, // [0:1] is the sub-list for field type_name
}

func init() { file_app_observatory_multiobservatory_config_proto_init() }

func file_app_observatory_multiobservatory_config_proto_init() {
if File_app_observatory_multiobservatory_config_proto != nil {
return
}
type x struct{}
out := protoimpl.TypeBuilder{
File: protoimpl.DescBuilder{
GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
RawDescriptor: file_app_observatory_multiobservatory_config_proto_rawDesc,
NumEnums:      0,
NumMessages:   2,
NumExtensions: 0,
NumServices:   0,
},
GoTypes:           file_app_observatory_multiobservatory_config_proto_goTypes,
DependencyIndexes: file_app_observatory_multiobservatory_config_proto_depIdxs,
MessageInfos:      file_app_observatory_multiobservatory_config_proto_msgTypes,
}.Build()
File_app_observatory_multiobservatory_config_proto = out.File
file_app_observatory_multiobservatory_config_proto_goTypes = nil
file_app_observatory_multiobservatory_config_proto_depIdxs = nil

}
