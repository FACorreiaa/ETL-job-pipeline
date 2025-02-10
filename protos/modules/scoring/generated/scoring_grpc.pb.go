// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: scoring.proto

package generated

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ScoringService_CalculateScores_FullMethodName       = "/scoringpb.ScoringService/CalculateScores"
	ScoringService_CalculateScoresStream_FullMethodName = "/scoringpb.ScoringService/CalculateScoresStream"
)

// ScoringServiceClient is the client API for ScoringService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ScoringServiceClient interface {
	CalculateScores(ctx context.Context, in *CalculateRequest, opts ...grpc.CallOption) (*CalculateResponse, error)
	CalculateScoresStream(ctx context.Context, in *CalculateRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[CompanyScore], error)
}

type scoringServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewScoringServiceClient(cc grpc.ClientConnInterface) ScoringServiceClient {
	return &scoringServiceClient{cc}
}

func (c *scoringServiceClient) CalculateScores(ctx context.Context, in *CalculateRequest, opts ...grpc.CallOption) (*CalculateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CalculateResponse)
	err := c.cc.Invoke(ctx, ScoringService_CalculateScores_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *scoringServiceClient) CalculateScoresStream(ctx context.Context, in *CalculateRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[CompanyScore], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &ScoringService_ServiceDesc.Streams[0], ScoringService_CalculateScoresStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[CalculateRequest, CompanyScore]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ScoringService_CalculateScoresStreamClient = grpc.ServerStreamingClient[CompanyScore]

// ScoringServiceServer is the server API for ScoringService service.
// All implementations must embed UnimplementedScoringServiceServer
// for forward compatibility.
type ScoringServiceServer interface {
	CalculateScores(context.Context, *CalculateRequest) (*CalculateResponse, error)
	CalculateScoresStream(*CalculateRequest, grpc.ServerStreamingServer[CompanyScore]) error
	mustEmbedUnimplementedScoringServiceServer()
}

// UnimplementedScoringServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedScoringServiceServer struct{}

func (UnimplementedScoringServiceServer) CalculateScores(context.Context, *CalculateRequest) (*CalculateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CalculateScores not implemented")
}
func (UnimplementedScoringServiceServer) CalculateScoresStream(*CalculateRequest, grpc.ServerStreamingServer[CompanyScore]) error {
	return status.Errorf(codes.Unimplemented, "method CalculateScoresStream not implemented")
}
func (UnimplementedScoringServiceServer) mustEmbedUnimplementedScoringServiceServer() {}
func (UnimplementedScoringServiceServer) testEmbeddedByValue()                        {}

// UnsafeScoringServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ScoringServiceServer will
// result in compilation errors.
type UnsafeScoringServiceServer interface {
	mustEmbedUnimplementedScoringServiceServer()
}

func RegisterScoringServiceServer(s grpc.ServiceRegistrar, srv ScoringServiceServer) {
	// If the following call pancis, it indicates UnimplementedScoringServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ScoringService_ServiceDesc, srv)
}

func _ScoringService_CalculateScores_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CalculateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ScoringServiceServer).CalculateScores(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ScoringService_CalculateScores_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ScoringServiceServer).CalculateScores(ctx, req.(*CalculateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ScoringService_CalculateScoresStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CalculateRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ScoringServiceServer).CalculateScoresStream(m, &grpc.GenericServerStream[CalculateRequest, CompanyScore]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type ScoringService_CalculateScoresStreamServer = grpc.ServerStreamingServer[CompanyScore]

// ScoringService_ServiceDesc is the grpc.ServiceDesc for ScoringService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ScoringService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "scoringpb.ScoringService",
	HandlerType: (*ScoringServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CalculateScores",
			Handler:    _ScoringService_CalculateScores_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CalculateScoresStream",
			Handler:       _ScoringService_CalculateScoresStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "scoring.proto",
}
