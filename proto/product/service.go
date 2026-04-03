package productpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductServiceServer is implemented by the product-service.
type ProductServiceServer interface {
	CreateProduct(context.Context, *CreateProductRequest) (*CreateProductResponse, error)
	GetProduct(context.Context, *GetProductRequest) (*Product, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsResponse, error)
	DeductStock(context.Context, *DeductStockRequest) (*DeductStockResponse, error)
	mustEmbedUnimplementedProductServiceServer()
}

// UnimplementedProductServiceServer for forward compatibility.
type UnimplementedProductServiceServer struct{}

func (UnimplementedProductServiceServer) CreateProduct(context.Context, *CreateProductRequest) (*CreateProductResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProduct not implemented")
}
func (UnimplementedProductServiceServer) GetProduct(context.Context, *GetProductRequest) (*Product, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProduct not implemented")
}
func (UnimplementedProductServiceServer) ListProducts(context.Context, *ListProductsRequest) (*ListProductsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProducts not implemented")
}
func (UnimplementedProductServiceServer) DeductStock(context.Context, *DeductStockRequest) (*DeductStockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeductStock not implemented")
}
func (UnimplementedProductServiceServer) mustEmbedUnimplementedProductServiceServer() {}

func RegisterProductServiceServer(s grpc.ServiceRegistrar, srv ProductServiceServer) {
	s.RegisterService(&ProductService_ServiceDesc, srv)
}

// ProductServiceClient is used by other services.
type ProductServiceClient interface {
	CreateProduct(ctx context.Context, req *CreateProductRequest, opts ...grpc.CallOption) (*CreateProductResponse, error)
	GetProduct(ctx context.Context, req *GetProductRequest, opts ...grpc.CallOption) (*Product, error)
	ListProducts(ctx context.Context, req *ListProductsRequest, opts ...grpc.CallOption) (*ListProductsResponse, error)
	DeductStock(ctx context.Context, req *DeductStockRequest, opts ...grpc.CallOption) (*DeductStockResponse, error)
}

type productServiceClient struct{ cc grpc.ClientConnInterface }

func NewProductServiceClient(cc grpc.ClientConnInterface) ProductServiceClient {
	return &productServiceClient{cc}
}

func (c *productServiceClient) CreateProduct(ctx context.Context, in *CreateProductRequest, opts ...grpc.CallOption) (*CreateProductResponse, error) {
	out := new(CreateProductResponse)
	return out, c.cc.Invoke(ctx, "/product.ProductService/CreateProduct", in, out, opts...)
}
func (c *productServiceClient) GetProduct(ctx context.Context, in *GetProductRequest, opts ...grpc.CallOption) (*Product, error) {
	out := new(Product)
	return out, c.cc.Invoke(ctx, "/product.ProductService/GetProduct", in, out, opts...)
}
func (c *productServiceClient) ListProducts(ctx context.Context, in *ListProductsRequest, opts ...grpc.CallOption) (*ListProductsResponse, error) {
	out := new(ListProductsResponse)
	return out, c.cc.Invoke(ctx, "/product.ProductService/ListProducts", in, out, opts...)
}
func (c *productServiceClient) DeductStock(ctx context.Context, in *DeductStockRequest, opts ...grpc.CallOption) (*DeductStockResponse, error) {
	out := new(DeductStockResponse)
	return out, c.cc.Invoke(ctx, "/product.ProductService/DeductStock", in, out, opts...)
}

// Handlers

func _ProductService_CreateProduct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateProductRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).CreateProduct(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/product.ProductService/CreateProduct"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(ProductServiceServer).CreateProduct(ctx, req.(*CreateProductRequest))
		})
}
func _ProductService_GetProduct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProductRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).GetProduct(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/product.ProductService/GetProduct"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(ProductServiceServer).GetProduct(ctx, req.(*GetProductRequest))
		})
}
func _ProductService_ListProducts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProductsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).ListProducts(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/product.ProductService/ListProducts"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(ProductServiceServer).ListProducts(ctx, req.(*ListProductsRequest))
		})
}
func _ProductService_DeductStock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeductStockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).DeductStock(ctx, in)
	}
	return interceptor(ctx, in, &grpc.UnaryServerInfo{Server: srv, FullMethod: "/product.ProductService/DeductStock"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(ProductServiceServer).DeductStock(ctx, req.(*DeductStockRequest))
		})
}

var ProductService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "product.ProductService",
	HandlerType: (*ProductServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateProduct", Handler: _ProductService_CreateProduct_Handler},
		{MethodName: "GetProduct", Handler: _ProductService_GetProduct_Handler},
		{MethodName: "ListProducts", Handler: _ProductService_ListProducts_Handler},
		{MethodName: "DeductStock", Handler: _ProductService_DeductStock_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "product/product.proto",
}
