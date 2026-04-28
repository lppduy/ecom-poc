package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/payment/internal/api/response"
	"github.com/lppduy/ecom-poc/services/payment/internal/domain"
	"github.com/lppduy/ecom-poc/services/payment/internal/service"
)

type PaymentController struct {
	svc service.PaymentService
}

func NewPaymentController(svc service.PaymentService) *PaymentController {
	return &PaymentController{svc: svc}
}

type createPaymentRequest struct {
	OrderID string  `json:"orderId" binding:"required"`
	Amount  float64 `json:"amount"  binding:"required,gt=0"`
}

type callbackRequest struct {
	Result string `json:"result" binding:"required"`
}

func (ctrl *PaymentController) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := ctrl.svc.CreatePayment(req.OrderID, req.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidOrderID) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "failed to create payment")
		return
	}
	response.Created(c, p)
}

func (ctrl *PaymentController) GetPayment(c *gin.Context) {
	id := c.Param("id")
	p, err := ctrl.svc.GetPayment(id)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			response.NotFound(c, "payment not found")
			return
		}
		response.InternalError(c, "failed to get payment")
		return
	}
	response.OK(c, p)
}

// Callback simulates the payment gateway webhook.
// POST /payments/:id/callback {"result":"success"|"fail"}
func (ctrl *PaymentController) Callback(c *gin.Context) {
	id := c.Param("id")
	var req callbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := ctrl.svc.Callback(id, req.Result)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPaymentNotFound):
			response.NotFound(c, "payment not found")
		case errors.Is(err, domain.ErrAlreadyProcessed):
			response.Conflict(c, "payment already processed")
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}
	response.OK(c, p)
}
