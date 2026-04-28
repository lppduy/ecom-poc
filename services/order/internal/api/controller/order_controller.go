package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	responsedto "github.com/lppduy/ecom-poc/services/order/internal/api/dto/response"
	"github.com/lppduy/ecom-poc/services/order/internal/api/response"
	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"github.com/lppduy/ecom-poc/services/order/internal/service"
)

type OrderController struct {
	service service.OrderService
}

func NewOrderController(service service.OrderService) *OrderController {
	return &OrderController{service: service}
}

func (c *OrderController) Health(ctx *gin.Context) {
	response.OK(ctx, gin.H{"status": "ok"})
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	userID := jwtutil.GetUserID(ctx)
	if userID == "" {
		response.BadRequest(ctx, "unauthenticated")
		return
	}

	idempotencyKey := ctx.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		response.BadRequest(ctx, "Idempotency-Key header is required")
		return
	}

	created, existed, err := c.service.CreateOrder(ctx.Request.Context(), userID, idempotencyKey)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) {
			response.BadRequest(ctx, "cart is empty")
			return
		}
		response.InternalError(ctx, "failed to create order")
		return
	}

	resp := responsedto.FromDomain(created)
	if existed {
		response.OK(ctx, resp)
		return
	}
	response.Created(ctx, resp)
}

func (c *OrderController) GetOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.BadRequest(ctx, "order id is required")
		return
	}

	found, ok, err := c.service.GetOrder(id)
	if err != nil {
		response.InternalError(ctx, "failed to query order")
		return
	}
	if !ok {
		response.NotFound(ctx, "order not found")
		return
	}

	response.OK(ctx, responsedto.FromDomain(found))
}
