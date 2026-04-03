// Package orderpb contains the gRPC types for the Order service.
package orderpb

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Types ───────────────────────────────────────────────────────────────────

type CreateOrderRequest struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type CreateOrderResponse struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
	Message string  `json:"message"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"` // for authorization
}

type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	Quantity  int32     `json:"quantity"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // pending|paid|cancelled
	CreatedAt time.Time `json:"created_at"`
}

type ListOrdersRequest struct {
	UserID string `json:"user_id"`
	Page   int32  `json:"page"`
	Limit  int32  `json:"limit"`
}

type ListOrdersResponse struct {
	Orders []*Order `json:"orders"`
	Total  int64    `json:"total"`
}

type UpdateOrderStatusRequest struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type UpdateOrderStatusResponse struct {
	Success bool `json:"success"`
}

// ─── Server ──────────────────────────────────────────────────────────────────

type OrderServiceServer interface {
	CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(context.Context, *GetOrderRequest) (*Order, error)
	ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersResponse, error)
	UpdateOrderStatus(context.Context, *UpdateOrderStatusRequest) (*UpdateOrderStatusResponse, error)
	mustEmbedUnimplementedOrderServiceServer()
}

type UnimplementedOrderServiceServer struct{}

func (UnimplementedOrderServiceServer) CreateOrder(context.Context, *CreateOrderRequest) (*CreateOrderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrder not implemented")
}
func (UnimplementedOrderServiceServer) GetOrder(context.Context, *GetOrderRequest) (*Order, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrder not implemented")
}
func (UnimplementedOrderServiceServer) ListOrders(context.Context, *ListOrdersRequest) (*ListOrdersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrders not implemented")
}
func (UnimplementedOrderServiceServer) UpdateOrderStatus(context.Context, *UpdateOrderStatusRequest) (*UpdateOrderStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOrderStatus not implemented")
}
func (UnimplementedOrderServiceServer) mustEmbedUnimplementedOrderServiceServer() {}

func RegisterOrderServiceServer(s grpc.ServiceRegistrar, srv OrderServiceServer) {
	s.RegisterService(&OrderService_ServiceDesc, srv)
}

// ─── Client ──────────────────────────────────────────────────────────────────

type OrderServiceClient interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderResponse, error)
	GetOrder(ctx context.Context, req *GetOrderRequest, opts ...grpc.CallOption) (*Order, error)
	ListOrders(ctx context.Context, req *ListOrdersRequest, opts ...grpc.CallOption) (*ListOrdersResponse, error)
	UpdateOrderStatus(ctx context.Context, req *UpdateOrderStatusRequest, opts ...grpc.CallOption) (*UpdateOrderStatusResponse, error)
}

type orderServiceClient struct{ cc grpc.ClientConnInterface }

func NewOrderServiceClient(cc grpc.ClientConnInterface) OrderServiceClient { return &orderServiceClient{cc} }

func (c *orderServiceClient) CreateOrder(ctx context.Context, in *CreateOrderRequest, opts ...grpc.CallOption) (*CreateOrderResponse, error) {
	out := new(CreateOrderResponse)
	return out, c.cc.Invoke(ctx, "/order.OrderService/CreateOrder", in, out, opts...)
}
func (c *orderServiceClient) GetOrder(ctx context.Context, in *GetOrderRequest, opts ...grpc.CallOption) (*Order, error) {
	out := new(Order)
	return out, c.cc.Invoke(ctx, "/order.OrderService/GetOrder", in, out, opts...)
}
func (c *orderServiceClient) ListOrders(ctx context.Context, in *ListOrdersRequest, opts ...grpc.CallOption) (*ListOrdersResponse, error) {
	out := new(ListOrdersResponse)
	return out, c.cc.Invoke(ctx, "/order.OrderService/ListOrders", in, out, opts...)
}
func (c *orderServiceClient) UpdateOrderStatus(ctx context.Context, in *UpdateOrderStatusRequest, opts ...grpc.CallOption) (*UpdateOrderStatusResponse, error) {
	out := new(UpdateOrderStatusResponse)
	return out, c.cc.Invoke(ctx, "/order.OrderService/UpdateOrderStatus", in, out, opts...)
}

// ─── Handlers & ServiceDesc ──────────────────────────────────────────────────

func _makeHandler[T any](method string, call func(interface{}, context.Context, *T) (interface{}, error)) grpc.MethodDesc {
	return grpc.MethodDesc{
		MethodName: method,
		Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
			in := new(T)
			if err := dec(in); err != nil {
				return nil, err
			}
			if interceptor == nil {
				return call(srv, ctx, in)
			}
			return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/" + method},
				func(ctx context.Context, req interface{}) (interface{}, error) { return call(srv, ctx, req.(*T)) })
		},
	}
}

var OrderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "order.OrderService",
	HandlerType: (*OrderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOrder",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(CreateOrderRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(OrderServiceServer).CreateOrder(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/CreateOrder"},
					func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(OrderServiceServer).CreateOrder(ctx, req.(*CreateOrderRequest))
					})
			},
		},
		{
			MethodName: "GetOrder",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(GetOrderRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(OrderServiceServer).GetOrder(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/GetOrder"},
					func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(OrderServiceServer).GetOrder(ctx, req.(*GetOrderRequest))
					})
			},
		},
		{
			MethodName: "ListOrders",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(ListOrdersRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(OrderServiceServer).ListOrders(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/ListOrders"},
					func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(OrderServiceServer).ListOrders(ctx, req.(*ListOrdersRequest))
					})
			},
		},
		{
			MethodName: "UpdateOrderStatus",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(UpdateOrderStatusRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(OrderServiceServer).UpdateOrderStatus(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/UpdateOrderStatus"},
					func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(OrderServiceServer).UpdateOrderStatus(ctx, req.(*UpdateOrderStatusRequest))
					})
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "order/order.proto",
}
