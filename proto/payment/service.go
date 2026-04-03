// Package paymentpb contains the gRPC types for the Payment service.
package paymentpb

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ─── Types ───────────────────────────────────────────────────────────────────

type ProcessPaymentRequest struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
	// In production: card token from Stripe/Razorpay
	CardToken string `json:"card_token"`
}

type ProcessPaymentResponse struct {
	PaymentID string  `json:"payment_id"`
	Status    string  `json:"status"` // success|failed|pending
	Message   string  `json:"message"`
	Amount    float64 `json:"amount"`
}

type GetPaymentRequest struct {
	PaymentID string `json:"payment_id"`
}

type Payment struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	Gateway     string    `json:"gateway"`
	ProcessedAt time.Time `json:"processed_at"`
}

// ─── Server ──────────────────────────────────────────────────────────────────

type PaymentServiceServer interface {
	ProcessPayment(context.Context, *ProcessPaymentRequest) (*ProcessPaymentResponse, error)
	GetPayment(context.Context, *GetPaymentRequest) (*Payment, error)
	mustEmbedUnimplementedPaymentServiceServer()
}

type UnimplementedPaymentServiceServer struct{}

func (UnimplementedPaymentServiceServer) ProcessPayment(context.Context, *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessPayment not implemented")
}
func (UnimplementedPaymentServiceServer) GetPayment(context.Context, *GetPaymentRequest) (*Payment, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPayment not implemented")
}
func (UnimplementedPaymentServiceServer) mustEmbedUnimplementedPaymentServiceServer() {}

func RegisterPaymentServiceServer(s grpc.ServiceRegistrar, srv PaymentServiceServer) {
	s.RegisterService(&PaymentService_ServiceDesc, srv)
}

// ─── Client ──────────────────────────────────────────────────────────────────

type PaymentServiceClient interface {
	ProcessPayment(ctx context.Context, req *ProcessPaymentRequest, opts ...grpc.CallOption) (*ProcessPaymentResponse, error)
	GetPayment(ctx context.Context, req *GetPaymentRequest, opts ...grpc.CallOption) (*Payment, error)
}

type paymentServiceClient struct{ cc grpc.ClientConnInterface }

func NewPaymentServiceClient(cc grpc.ClientConnInterface) PaymentServiceClient {
	return &paymentServiceClient{cc}
}
func (c *paymentServiceClient) ProcessPayment(ctx context.Context, in *ProcessPaymentRequest, opts ...grpc.CallOption) (*ProcessPaymentResponse, error) {
	out := new(ProcessPaymentResponse)
	return out, c.cc.Invoke(ctx, "/payment.PaymentService/ProcessPayment", in, out, opts...)
}
func (c *paymentServiceClient) GetPayment(ctx context.Context, in *GetPaymentRequest, opts ...grpc.CallOption) (*Payment, error) {
	out := new(Payment)
	return out, c.cc.Invoke(ctx, "/payment.PaymentService/GetPayment", in, out, opts...)
}

// ─── ServiceDesc ─────────────────────────────────────────────────────────────

var PaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessPayment",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(ProcessPaymentRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(PaymentServiceServer).ProcessPayment(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/ProcessPayment"},
					func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(PaymentServiceServer).ProcessPayment(ctx, req.(*ProcessPaymentRequest))
					})
			},
		},
		{
			MethodName: "GetPayment",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(GetPaymentRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(PaymentServiceServer).GetPayment(ctx, in)
				}
				return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/payment.PaymentService/GetPayment"},
					func(ctx context.Context, req interface{}) (interface{}, error) {
						return srv.(PaymentServiceServer).GetPayment(ctx, req.(*GetPaymentRequest))
					})
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment/payment.proto",
}
