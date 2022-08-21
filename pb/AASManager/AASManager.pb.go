// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0-devel
// 	protoc        v3.11.4
// source: AASManager/AASManager.proto

package AASManager

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	AASMessage "github.com/lyesteven/go-framework/pb/AASMessage"
	reflect "reflect"
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

var File_AASManager_AASManager_proto protoreflect.FileDescriptor

var file_AASManager_AASManager_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x41, 0x41, 0x53, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x41, 0x41, 0x53,
	0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x41,
	0x41, 0x53, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x1a, 0x1b, 0x41, 0x41, 0x53, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x2f, 0x41, 0x41, 0x53, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0x69, 0x0a, 0x0a, 0x41, 0x41, 0x53, 0x4d, 0x61, 0x6e,
	0x61, 0x67, 0x65, 0x72, 0x12, 0x5b, 0x0a, 0x12, 0x49, 0x6e, 0x76, 0x6f, 0x6b, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x79, 0x41, 0x41, 0x53, 0x12, 0x20, 0x2e, 0x41, 0x41, 0x53,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x49, 0x6e, 0x76, 0x6f, 0x6b, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x41,
	0x41, 0x53, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x49, 0x6e, 0x76, 0x6f, 0x6b, 0x65,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x42, 0x0f, 0x5a, 0x0d, 0x70, 0x62, 0x2f, 0x41, 0x41, 0x53, 0x4d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_AASManager_AASManager_proto_goTypes = []interface{}{
	(*AASMessage.InvokeServiceRequest)(nil),  // 0: AASMessage.InvokeServiceRequest
	(*AASMessage.InvokeServiceResponse)(nil), // 1: AASMessage.InvokeServiceResponse
}
var file_AASManager_AASManager_proto_depIdxs = []int32{
	0, // 0: AASManager.AASManager.InvokeServiceByAAS:input_type -> AASMessage.InvokeServiceRequest
	1, // 1: AASManager.AASManager.InvokeServiceByAAS:output_type -> AASMessage.InvokeServiceResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_AASManager_AASManager_proto_init() }
func file_AASManager_AASManager_proto_init() {
	if File_AASManager_AASManager_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_AASManager_AASManager_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_AASManager_AASManager_proto_goTypes,
		DependencyIndexes: file_AASManager_AASManager_proto_depIdxs,
	}.Build()
	File_AASManager_AASManager_proto = out.File
	file_AASManager_AASManager_proto_rawDesc = nil
	file_AASManager_AASManager_proto_goTypes = nil
	file_AASManager_AASManager_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AASManagerClient is the client API for AASManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AASManagerClient interface {
	InvokeServiceByAAS(ctx context.Context, in *AASMessage.InvokeServiceRequest, opts ...grpc.CallOption) (*AASMessage.InvokeServiceResponse, error)
}

type aASManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewAASManagerClient(cc grpc.ClientConnInterface) AASManagerClient {
	return &aASManagerClient{cc}
}

func (c *aASManagerClient) InvokeServiceByAAS(ctx context.Context, in *AASMessage.InvokeServiceRequest, opts ...grpc.CallOption) (*AASMessage.InvokeServiceResponse, error) {
	out := new(AASMessage.InvokeServiceResponse)
	err := c.cc.Invoke(ctx, "/AASManager.AASManager/InvokeServiceByAAS", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AASManagerServer is the server API for AASManager service.
type AASManagerServer interface {
	InvokeServiceByAAS(context.Context, *AASMessage.InvokeServiceRequest) (*AASMessage.InvokeServiceResponse, error)
}

// UnimplementedAASManagerServer can be embedded to have forward compatible implementations.
type UnimplementedAASManagerServer struct {
}

func (*UnimplementedAASManagerServer) InvokeServiceByAAS(context.Context, *AASMessage.InvokeServiceRequest) (*AASMessage.InvokeServiceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InvokeServiceByAAS not implemented")
}

func RegisterAASManagerServer(s *grpc.Server, srv AASManagerServer) {
	s.RegisterService(&_AASManager_serviceDesc, srv)
}

func _AASManager_InvokeServiceByAAS_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AASMessage.InvokeServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AASManagerServer).InvokeServiceByAAS(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/AASManager.AASManager/InvokeServiceByAAS",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AASManagerServer).InvokeServiceByAAS(ctx, req.(*AASMessage.InvokeServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _AASManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "AASManager.AASManager",
	HandlerType: (*AASManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InvokeServiceByAAS",
			Handler:    _AASManager_InvokeServiceByAAS_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "AASManager/AASManager.proto",
}