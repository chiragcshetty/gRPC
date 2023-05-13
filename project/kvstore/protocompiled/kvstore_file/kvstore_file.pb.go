// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.12.4
// source: kvstore_file.proto

package kvstore_file

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

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp int32  `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	FileChunk string `protobuf:"bytes,2,opt,name=fileChunk,proto3" json:"fileChunk,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_file_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_file_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_kvstore_file_proto_rawDescGZIP(), []int{0}
}

func (x *Response) GetTimestamp() int32 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Response) GetFileChunk() string {
	if x != nil {
		return x.FileChunk
	}
	return ""
}

type Record struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Using a string instead (note: '[]byte' can't be used as map key in golang)
	Key int32 `protobuf:"varint,1,opt,name=key,proto3" json:"key,omitempty"`
	// string filename = 2;
	ClientId  string `protobuf:"bytes,2,opt,name=clientId,proto3" json:"clientId,omitempty"`
	TimeStamp int32  `protobuf:"varint,3,opt,name=timeStamp,proto3" json:"timeStamp,omitempty"`
	Value     string `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Record) Reset() {
	*x = Record{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_file_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Record) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record) ProtoMessage() {}

func (x *Record) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_file_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record.ProtoReflect.Descriptor instead.
func (*Record) Descriptor() ([]byte, []int) {
	return file_kvstore_file_proto_rawDescGZIP(), []int{1}
}

func (x *Record) GetKey() int32 {
	if x != nil {
		return x.Key
	}
	return 0
}

func (x *Record) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *Record) GetTimeStamp() int32 {
	if x != nil {
		return x.TimeStamp
	}
	return 0
}

func (x *Record) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type AckMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ack by server to a Set sent by client
	Ack bool `protobuf:"varint,1,opt,name=ack,proto3" json:"ack,omitempty"`
}

func (x *AckMsg) Reset() {
	*x = AckMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kvstore_file_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AckMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AckMsg) ProtoMessage() {}

func (x *AckMsg) ProtoReflect() protoreflect.Message {
	mi := &file_kvstore_file_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AckMsg.ProtoReflect.Descriptor instead.
func (*AckMsg) Descriptor() ([]byte, []int) {
	return file_kvstore_file_proto_rawDescGZIP(), []int{2}
}

func (x *AckMsg) GetAck() bool {
	if x != nil {
		return x.Ack
	}
	return false
}

var File_kvstore_file_proto protoreflect.FileDescriptor

var file_kvstore_file_proto_rawDesc = []byte{
	0x0a, 0x12, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x22, 0x46, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x1c, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x22, 0x6a,
	0x0a, 0x06, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x53, 0x74,
	0x61, 0x6d, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x53,
	0x74, 0x61, 0x6d, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x1a, 0x0a, 0x06, 0x41, 0x63,
	0x6b, 0x4d, 0x73, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x63, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x03, 0x61, 0x63, 0x6b, 0x32, 0xc8, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x12, 0x41, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x54, 0x65, 0x73, 0x74, 0x12, 0x18, 0x2e, 0x6b, 0x76,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x52,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x1a, 0x1a, 0x2e, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x18, 0x2e, 0x6b, 0x76, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x52, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x1a, 0x1a, 0x2e, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x30, 0x01, 0x12, 0x3b, 0x0a, 0x03, 0x53, 0x65, 0x74, 0x12, 0x18, 0x2e, 0x6b, 0x76,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x52,
	0x65, 0x63, 0x6f, 0x72, 0x64, 0x1a, 0x18, 0x2e, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x46, 0x69, 0x6c, 0x65, 0x2e, 0x41, 0x63, 0x6b, 0x4d, 0x73, 0x67, 0x22,
	0x00, 0x42, 0x1c, 0x5a, 0x1a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6d, 0x70, 0x69, 0x6c,
	0x65, 0x64, 0x2f, 0x6b, 0x76, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kvstore_file_proto_rawDescOnce sync.Once
	file_kvstore_file_proto_rawDescData = file_kvstore_file_proto_rawDesc
)

func file_kvstore_file_proto_rawDescGZIP() []byte {
	file_kvstore_file_proto_rawDescOnce.Do(func() {
		file_kvstore_file_proto_rawDescData = protoimpl.X.CompressGZIP(file_kvstore_file_proto_rawDescData)
	})
	return file_kvstore_file_proto_rawDescData
}

var file_kvstore_file_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_kvstore_file_proto_goTypes = []interface{}{
	(*Response)(nil), // 0: kvstoreProtoFile.Response
	(*Record)(nil),   // 1: kvstoreProtoFile.Record
	(*AckMsg)(nil),   // 2: kvstoreProtoFile.AckMsg
}
var file_kvstore_file_proto_depIdxs = []int32{
	1, // 0: kvstoreProtoFile.Store.GetTest:input_type -> kvstoreProtoFile.Record
	1, // 1: kvstoreProtoFile.Store.Get:input_type -> kvstoreProtoFile.Record
	1, // 2: kvstoreProtoFile.Store.Set:input_type -> kvstoreProtoFile.Record
	0, // 3: kvstoreProtoFile.Store.GetTest:output_type -> kvstoreProtoFile.Response
	0, // 4: kvstoreProtoFile.Store.Get:output_type -> kvstoreProtoFile.Response
	2, // 5: kvstoreProtoFile.Store.Set:output_type -> kvstoreProtoFile.AckMsg
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_kvstore_file_proto_init() }
func file_kvstore_file_proto_init() {
	if File_kvstore_file_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kvstore_file_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
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
		file_kvstore_file_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Record); i {
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
		file_kvstore_file_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AckMsg); i {
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
			RawDescriptor: file_kvstore_file_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_kvstore_file_proto_goTypes,
		DependencyIndexes: file_kvstore_file_proto_depIdxs,
		MessageInfos:      file_kvstore_file_proto_msgTypes,
	}.Build()
	File_kvstore_file_proto = out.File
	file_kvstore_file_proto_rawDesc = nil
	file_kvstore_file_proto_goTypes = nil
	file_kvstore_file_proto_depIdxs = nil
}
