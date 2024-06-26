// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: spark/v1/rpc_get_job_status.proto

package sparkv1

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

type GetJobStatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (x *GetJobStatusRequest) Reset() {
	*x = GetJobStatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_spark_v1_rpc_get_job_status_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetJobStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetJobStatusRequest) ProtoMessage() {}

func (x *GetJobStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_spark_v1_rpc_get_job_status_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetJobStatusRequest.ProtoReflect.Descriptor instead.
func (*GetJobStatusRequest) Descriptor() ([]byte, []int) {
	return file_spark_v1_rpc_get_job_status_proto_rawDescGZIP(), []int{0}
}

func (x *GetJobStatusRequest) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

type GetJobStatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	JobId        string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Status       string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	ErrorMessage string `protobuf:"bytes,3,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *GetJobStatusResponse) Reset() {
	*x = GetJobStatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_spark_v1_rpc_get_job_status_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetJobStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetJobStatusResponse) ProtoMessage() {}

func (x *GetJobStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_spark_v1_rpc_get_job_status_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetJobStatusResponse.ProtoReflect.Descriptor instead.
func (*GetJobStatusResponse) Descriptor() ([]byte, []int) {
	return file_spark_v1_rpc_get_job_status_proto_rawDescGZIP(), []int{1}
}

func (x *GetJobStatusResponse) GetJobId() string {
	if x != nil {
		return x.JobId
	}
	return ""
}

func (x *GetJobStatusResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *GetJobStatusResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

var File_spark_v1_rpc_get_job_status_proto protoreflect.FileDescriptor

var file_spark_v1_rpc_get_job_status_proto_rawDesc = []byte{
	0x0a, 0x21, 0x73, 0x70, 0x61, 0x72, 0x6b, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x70, 0x63, 0x5f, 0x67,
	0x65, 0x74, 0x5f, 0x6a, 0x6f, 0x62, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x70, 0x61, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x22, 0x2c, 0x0a,
	0x13, 0x47, 0x65, 0x74, 0x4a, 0x6f, 0x62, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x22, 0x6a, 0x0a, 0x14, 0x47,
	0x65, 0x74, 0x4a, 0x6f, 0x62, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x6a, 0x6f, 0x62, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x6a, 0x6f, 0x62, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42, 0x77, 0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x2e, 0x73,
	0x70, 0x61, 0x72, 0x6b, 0x2e, 0x76, 0x31, 0x42, 0x14, 0x52, 0x70, 0x63, 0x47, 0x65, 0x74, 0x4a,
	0x6f, 0x62, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a,
	0x10, 0x73, 0x70, 0x61, 0x72, 0x6b, 0x2f, 0x76, 0x31, 0x3b, 0x73, 0x70, 0x61, 0x72, 0x6b, 0x76,
	0x31, 0xa2, 0x02, 0x03, 0x53, 0x58, 0x58, 0xaa, 0x02, 0x08, 0x53, 0x70, 0x61, 0x72, 0x6b, 0x2e,
	0x56, 0x31, 0xca, 0x02, 0x08, 0x53, 0x70, 0x61, 0x72, 0x6b, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14,
	0x53, 0x70, 0x61, 0x72, 0x6b, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x53, 0x70, 0x61, 0x72, 0x6b, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_spark_v1_rpc_get_job_status_proto_rawDescOnce sync.Once
	file_spark_v1_rpc_get_job_status_proto_rawDescData = file_spark_v1_rpc_get_job_status_proto_rawDesc
)

func file_spark_v1_rpc_get_job_status_proto_rawDescGZIP() []byte {
	file_spark_v1_rpc_get_job_status_proto_rawDescOnce.Do(func() {
		file_spark_v1_rpc_get_job_status_proto_rawDescData = protoimpl.X.CompressGZIP(file_spark_v1_rpc_get_job_status_proto_rawDescData)
	})
	return file_spark_v1_rpc_get_job_status_proto_rawDescData
}

var file_spark_v1_rpc_get_job_status_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_spark_v1_rpc_get_job_status_proto_goTypes = []interface{}{
	(*GetJobStatusRequest)(nil),  // 0: spark.v1.GetJobStatusRequest
	(*GetJobStatusResponse)(nil), // 1: spark.v1.GetJobStatusResponse
}
var file_spark_v1_rpc_get_job_status_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_spark_v1_rpc_get_job_status_proto_init() }
func file_spark_v1_rpc_get_job_status_proto_init() {
	if File_spark_v1_rpc_get_job_status_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_spark_v1_rpc_get_job_status_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetJobStatusRequest); i {
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
		file_spark_v1_rpc_get_job_status_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetJobStatusResponse); i {
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
			RawDescriptor: file_spark_v1_rpc_get_job_status_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_spark_v1_rpc_get_job_status_proto_goTypes,
		DependencyIndexes: file_spark_v1_rpc_get_job_status_proto_depIdxs,
		MessageInfos:      file_spark_v1_rpc_get_job_status_proto_msgTypes,
	}.Build()
	File_spark_v1_rpc_get_job_status_proto = out.File
	file_spark_v1_rpc_get_job_status_proto_rawDesc = nil
	file_spark_v1_rpc_get_job_status_proto_goTypes = nil
	file_spark_v1_rpc_get_job_status_proto_depIdxs = nil
}
