package userpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Server interface ────────────────────────────────────────────────────────

// UserServiceServer is implemented by the user-service.
type UserServiceServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterResponse, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
	mustEmbedUnimplementedUserServiceServer()
}

// UnimplementedUserServiceServer must be embedded for forward compatibility.
type UnimplementedUserServiceServer struct{}

func (UnimplementedUserServiceServer) Register(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedUserServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedUserServiceServer) GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedUserServiceServer) mustEmbedUnimplementedUserServiceServer() {}

// RegisterUserServiceServer registers the implementation with a gRPC server.
func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	s.RegisterService(&UserService_ServiceDesc, srv)
}

// ─── Client interface ────────────────────────────────────────────────────────

// UserServiceClient is used by other services to call the user-service.
type UserServiceClient interface {
	Register(ctx context.Context, req *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	GetUser(ctx context.Context, req *GetUserRequest, opts ...grpc.CallOption) (*GetUserResponse, error)
}

type userServiceClient struct{ cc grpc.ClientConnInterface }

// NewUserServiceClient dials the user-service and returns a client.
func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	return out, c.cc.Invoke(ctx, "/user.UserService/Register", in, out, opts...)
}
func (c *userServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	return out, c.cc.Invoke(ctx, "/user.UserService/Login", in, out, opts...)
}
func (c *userServiceClient) GetUser(ctx context.Context, in *GetUserRequest, opts ...grpc.CallOption) (*GetUserResponse, error) {
	out := new(GetUserResponse)
	return out, c.cc.Invoke(ctx, "/user.UserService/GetUser", in, out, opts...)
}

// ─── Server handlers & ServiceDesc ──────────────────────────────────────────

func _UserService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Register(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/user.UserService/Register"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(UserServiceServer).Register(ctx, req.(*RegisterRequest))
		})
}

func _UserService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Login(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/user.UserService/Login"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(UserServiceServer).Login(ctx, req.(*LoginRequest))
		})
}

func _UserService_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).GetUser(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/user.UserService/GetUser"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(UserServiceServer).GetUser(ctx, req.(*GetUserRequest))
		})
}

// UserService_ServiceDesc describes the UserService gRPC service.
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Register", Handler: _UserService_Register_Handler},
		{MethodName: "Login", Handler: _UserService_Login_Handler},
		{MethodName: "GetUser", Handler: _UserService_GetUser_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "user/user.proto",
}
