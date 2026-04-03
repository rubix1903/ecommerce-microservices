package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	userpb "github.com/ecommerce/microservices/proto/user"
	productpb "github.com/ecommerce/microservices/proto/product"
	orderpb "github.com/ecommerce/microservices/proto/order"
	paymentpb "github.com/ecommerce/microservices/proto/payment"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func grpcErr(c *gin.Context, err error) {
	c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
}

// ── User handlers ─────────────────────────────────────────────────────────────

type UserHTTPHandler struct{ client userpb.UserServiceClient }

func NewUserHTTPHandler(c userpb.UserServiceClient) *UserHTTPHandler { return &UserHTTPHandler{c} }

func (h *UserHTTPHandler) Register(c *gin.Context) {
	var req userpb.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.Register(c.Request.Context(), &req)
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *UserHTTPHandler) Login(c *gin.Context) {
	var req userpb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.Login(c.Request.Context(), &req)
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHTTPHandler) GetUser(c *gin.Context) {
	resp, err := h.client.GetUser(c.Request.Context(), &userpb.GetUserRequest{ID: c.Param("id")})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── Product handlers ──────────────────────────────────────────────────────────

type ProductHTTPHandler struct{ client productpb.ProductServiceClient }

func NewProductHTTPHandler(c productpb.ProductServiceClient) *ProductHTTPHandler {
	return &ProductHTTPHandler{c}
}

func (h *ProductHTTPHandler) CreateProduct(c *gin.Context) {
	var req productpb.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *ProductHTTPHandler) GetProduct(c *gin.Context) {
	resp, err := h.client.GetProduct(c.Request.Context(), &productpb.GetProductRequest{ID: c.Param("id")})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *ProductHTTPHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	resp, err := h.client.ListProducts(c.Request.Context(), &productpb.ListProductsRequest{
		Page: int32(page), Limit: int32(limit),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── Order handlers ────────────────────────────────────────────────────────────

type OrderHTTPHandler struct{ client orderpb.OrderServiceClient }

func NewOrderHTTPHandler(c orderpb.OrderServiceClient) *OrderHTTPHandler { return &OrderHTTPHandler{c} }

func (h *OrderHTTPHandler) CreateOrder(c *gin.Context) {
	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int32  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetString("user_id")
	resp, err := h.client.CreateOrder(c.Request.Context(), &orderpb.CreateOrderRequest{
		UserID:    userID,
		ProductID: body.ProductID,
		Quantity:  body.Quantity,
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *OrderHTTPHandler) GetOrder(c *gin.Context) {
	resp, err := h.client.GetOrder(c.Request.Context(), &orderpb.GetOrderRequest{
		OrderID: c.Param("id"),
		UserID:  c.GetString("user_id"),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrderHTTPHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	resp, err := h.client.ListOrders(c.Request.Context(), &orderpb.ListOrdersRequest{
		UserID: c.GetString("user_id"),
		Page:   int32(page),
		Limit:  int32(limit),
	})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── Payment handlers ──────────────────────────────────────────────────────────

type PaymentHTTPHandler struct{ client paymentpb.PaymentServiceClient }

func NewPaymentHTTPHandler(c paymentpb.PaymentServiceClient) *PaymentHTTPHandler {
	return &PaymentHTTPHandler{c}
}

func (h *PaymentHTTPHandler) GetPayment(c *gin.Context) {
	resp, err := h.client.GetPayment(c.Request.Context(), &paymentpb.GetPaymentRequest{PaymentID: c.Param("id")})
	if err != nil {
		grpcErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
