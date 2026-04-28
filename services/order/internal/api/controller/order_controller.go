package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	responsedto "github.com/lppduy/ecom-poc/services/order/internal/api/dto/response"
	"github.com/lppduy/ecom-poc/services/order/internal/api/httpx"
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
	httpx.OK(ctx, gin.H{"status": "ok"})
}

func (c *OrderController) CreateOrder(ctx *gin.Context) {
	userID := jwtutil.GetUserID(ctx)
	if userID == "" {
		httpx.BadRequest(ctx, "unauthenticated")
		return
	}

	idempotencyKey := ctx.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		httpx.BadRequest(ctx, "Idempotency-Key header is required")
		return
	}

	created, existed, err := c.service.CreateOrder(ctx.Request.Context(), userID, idempotencyKey)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyCart) {
			httpx.BadRequest(ctx, "cart is empty")
			return
		}
		httpx.InternalError(ctx, "failed to create order")
		return
	}

	resp := responsedto.FromDomain(created)
	if existed {
		httpx.OK(ctx, resp)
		return
	}
	httpx.Created(ctx, resp)
}

func (c *OrderController) GetOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		httpx.BadRequest(ctx, "order id is required")
		return
	}

	found, ok, err := c.service.GetOrder(id)
	if err != nil {
		httpx.InternalError(ctx, "failed to query order")
		return
	}
	if !ok {
		httpx.NotFound(ctx, "order not found")
		return
	}

	httpx.OK(ctx, responsedto.FromDomain(found))
}

func (c *OrderController) ConfirmOrder(ctx *gin.Context) {
	c.handleTransition(ctx, c.service.ConfirmOrder)
}

func (c *OrderController) FailOrder(ctx *gin.Context) {
	c.handleTransition(ctx, c.service.FailOrder)
}

func (c *OrderController) handleTransition(ctx *gin.Context, fn func(string) (domain.Order, error)) {
	id := ctx.Param("id")
	if id == "" {
		httpx.BadRequest(ctx, "order id is required")
		return
	}

	updated, err := fn(id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			httpx.NotFound(ctx, "order not found")
		case errors.Is(err, domain.ErrInvalidTransition):
			httpx.BadRequest(ctx, "invalid status transition")
		default:
			httpx.InternalError(ctx, "failed to update order")
		}
		return
	}

	httpx.OK(ctx, responsedto.FromDomain(updated))
}
