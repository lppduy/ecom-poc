package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/order/internal/service"
)

type OrderController struct {
	service *service.OrderService
}

func NewOrderController(service *service.OrderService) *OrderController {
	return &OrderController{service: service}
}

func (c *OrderController) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", c.health)
	router.POST("/orders", c.createOrder)
	router.GET("/orders/:id", c.getOrder)
}

type createOrderRequest struct {
	UserID string `json:"userId"`
}

func (c *OrderController) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *OrderController) createOrder(ctx *gin.Context) {
	var req createOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.UserID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	idempotencyKey := ctx.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	created, existed, err := c.service.CreateOrder(ctx.Request.Context(), req.UserID, idempotencyKey)
	if err != nil {
		if errors.Is(err, service.ErrEmptyCart) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	if existed {
		ctx.JSON(http.StatusOK, created)
		return
	}
	ctx.JSON(http.StatusCreated, created)
}

func (c *OrderController) getOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id is required"})
		return
	}

	found, ok, err := c.service.GetOrder(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query order"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	ctx.JSON(http.StatusOK, found)
}
