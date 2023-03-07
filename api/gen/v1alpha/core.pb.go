// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: v1alpha/core.proto

// Additional imports go here

package core

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

type CheckPermissionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Subject      string `protobuf:"bytes,1,opt,name=subject,proto3" json:"subject,omitempty"`
	Operation    string `protobuf:"bytes,2,opt,name=operation,proto3" json:"operation,omitempty"`
	Resourcetype string `protobuf:"bytes,3,opt,name=resourcetype,proto3" json:"resourcetype,omitempty"`
	Resourceid   string `protobuf:"bytes,4,opt,name=resourceid,proto3" json:"resourceid,omitempty"`
}

func (x *CheckPermissionRequest) Reset() {
	*x = CheckPermissionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1alpha_core_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckPermissionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckPermissionRequest) ProtoMessage() {}

func (x *CheckPermissionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1alpha_core_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckPermissionRequest.ProtoReflect.Descriptor instead.
func (*CheckPermissionRequest) Descriptor() ([]byte, []int) {
	return file_v1alpha_core_proto_rawDescGZIP(), []int{0}
}

func (x *CheckPermissionRequest) GetSubject() string {
	if x != nil {
		return x.Subject
	}
	return ""
}

func (x *CheckPermissionRequest) GetOperation() string {
	if x != nil {
		return x.Operation
	}
	return ""
}

func (x *CheckPermissionRequest) GetResourcetype() string {
	if x != nil {
		return x.Resourcetype
	}
	return ""
}

func (x *CheckPermissionRequest) GetResourceid() string {
	if x != nil {
		return x.Resourceid
	}
	return ""
}

type CheckPermissionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result      bool   `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
	Description string `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *CheckPermissionResponse) Reset() {
	*x = CheckPermissionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1alpha_core_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckPermissionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckPermissionResponse) ProtoMessage() {}

func (x *CheckPermissionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1alpha_core_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckPermissionResponse.ProtoReflect.Descriptor instead.
func (*CheckPermissionResponse) Descriptor() ([]byte, []int) {
	return file_v1alpha_core_proto_rawDescGZIP(), []int{1}
}

func (x *CheckPermissionResponse) GetResult() bool {
	if x != nil {
		return x.Result
	}
	return false
}

func (x *CheckPermissionResponse) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

// CreateSeatsRequest assuming we get the userId etc from the requestor in the authorization header to validate if an "admin" can actually add licenses.
type CreateSeatsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TenantId  string   `protobuf:"bytes,1,opt,name=tenantId,proto3" json:"tenantId,omitempty"`   //tenantId of subjects
	Subjects  []string `protobuf:"bytes,2,rep,name=subjects,proto3" json:"subjects,omitempty"`   //list of subjects to add
	ServiceId string   `protobuf:"bytes,3,opt,name=serviceId,proto3" json:"serviceId,omitempty"` //id of service to add subjects to as "licensed users"
}

func (x *CreateSeatsRequest) Reset() {
	*x = CreateSeatsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1alpha_core_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateSeatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSeatsRequest) ProtoMessage() {}

func (x *CreateSeatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1alpha_core_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSeatsRequest.ProtoReflect.Descriptor instead.
func (*CreateSeatsRequest) Descriptor() ([]byte, []int) {
	return file_v1alpha_core_proto_rawDescGZIP(), []int{2}
}

func (x *CreateSeatsRequest) GetTenantId() string {
	if x != nil {
		return x.TenantId
	}
	return ""
}

func (x *CreateSeatsRequest) GetSubjects() []string {
	if x != nil {
		return x.Subjects
	}
	return nil
}

func (x *CreateSeatsRequest) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

type CreateSeatsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CreateSeatsResponse) Reset() {
	*x = CreateSeatsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1alpha_core_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateSeatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSeatsResponse) ProtoMessage() {}

func (x *CreateSeatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1alpha_core_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateSeatsResponse.ProtoReflect.Descriptor instead.
func (*CreateSeatsResponse) Descriptor() ([]byte, []int) {
	return file_v1alpha_core_proto_rawDescGZIP(), []int{3}
}

var File_v1alpha_core_proto protoreflect.FileDescriptor

var file_v1alpha_core_proto_rawDesc = []byte{
	0x0a, 0x12, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x22, 0x94, 0x01, 0x0a, 0x16, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x50, 0x65, 0x72, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73,
	0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6f, 0x70, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x69, 0x64, 0x22, 0x53, 0x0a, 0x17, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x6a, 0x0a,
	0x12, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x49, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x08, 0x73, 0x75, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x22, 0x15, 0x0a, 0x13, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x53, 0x65, 0x61, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x32, 0x71, 0x0a, 0x0f, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x12, 0x5e, 0x0a, 0x0f, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x50, 0x65, 0x72, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x23, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x50,
	0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x32, 0x62, 0x0a, 0x0c, 0x53, 0x65, 0x61, 0x74, 0x73, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x52, 0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x61,
	0x74, 0x73, 0x12, 0x1f, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x53, 0x65, 0x61, 0x74, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x40, 0x5a, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x52, 0x65, 0x64, 0x48, 0x61, 0x74, 0x49, 0x6e, 0x73, 0x69,
	0x67, 0x68, 0x74, 0x73, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x7a, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x3b, 0x63, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_v1alpha_core_proto_rawDescOnce sync.Once
	file_v1alpha_core_proto_rawDescData = file_v1alpha_core_proto_rawDesc
)

func file_v1alpha_core_proto_rawDescGZIP() []byte {
	file_v1alpha_core_proto_rawDescOnce.Do(func() {
		file_v1alpha_core_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1alpha_core_proto_rawDescData)
	})
	return file_v1alpha_core_proto_rawDescData
}

var file_v1alpha_core_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_v1alpha_core_proto_goTypes = []interface{}{
	(*CheckPermissionRequest)(nil),  // 0: api.v1alpha.CheckPermissionRequest
	(*CheckPermissionResponse)(nil), // 1: api.v1alpha.CheckPermissionResponse
	(*CreateSeatsRequest)(nil),      // 2: api.v1alpha.CreateSeatsRequest
	(*CreateSeatsResponse)(nil),     // 3: api.v1alpha.CreateSeatsResponse
}
var file_v1alpha_core_proto_depIdxs = []int32{
	0, // 0: api.v1alpha.CheckPermission.CheckPermission:input_type -> api.v1alpha.CheckPermissionRequest
	2, // 1: api.v1alpha.SeatsService.CreateSeats:input_type -> api.v1alpha.CreateSeatsRequest
	1, // 2: api.v1alpha.CheckPermission.CheckPermission:output_type -> api.v1alpha.CheckPermissionResponse
	3, // 3: api.v1alpha.SeatsService.CreateSeats:output_type -> api.v1alpha.CreateSeatsResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_v1alpha_core_proto_init() }
func file_v1alpha_core_proto_init() {
	if File_v1alpha_core_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1alpha_core_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckPermissionRequest); i {
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
		file_v1alpha_core_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckPermissionResponse); i {
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
		file_v1alpha_core_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateSeatsRequest); i {
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
		file_v1alpha_core_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateSeatsResponse); i {
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
			RawDescriptor: file_v1alpha_core_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_v1alpha_core_proto_goTypes,
		DependencyIndexes: file_v1alpha_core_proto_depIdxs,
		MessageInfos:      file_v1alpha_core_proto_msgTypes,
	}.Build()
	File_v1alpha_core_proto = out.File
	file_v1alpha_core_proto_rawDesc = nil
	file_v1alpha_core_proto_goTypes = nil
	file_v1alpha_core_proto_depIdxs = nil
}
