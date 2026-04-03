// Package productpb contains the gRPC types for the Product service.
package productpb

// CreateProductRequest is the payload for adding a product to the catalog.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
}

// CreateProductResponse is returned after a product is created.
type CreateProductResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// GetProductRequest looks up a product by ID.
type GetProductRequest struct {
	ID string `json:"id"`
}

// Product is the full product record.
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
}

// ListProductsRequest requests the product catalog (supports pagination).
type ListProductsRequest struct {
	Page  int32 `json:"page"`
	Limit int32 `json:"limit"`
}

// ListProductsResponse returns a page of products.
type ListProductsResponse struct {
	Products []*Product `json:"products"`
	Total    int64      `json:"total"`
}

// DeductStockRequest reduces stock after an order is placed.
type DeductStockRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

// DeductStockResponse confirms the deduction.
type DeductStockResponse struct {
	Success      bool  `json:"success"`
	RemainingStock int32 `json:"remaining_stock"`
}
