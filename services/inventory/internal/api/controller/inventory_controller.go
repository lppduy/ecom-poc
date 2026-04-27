package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/inventory/internal/api/httpx"
	"github.com/lppduy/ecom-poc/services/inventory/internal/domain"
	"github.com/lppduy/ecom-poc/services/inventory/internal/service"
)

type InventoryController struct {
	service service.InventoryService
}

func NewInventoryController(svc service.InventoryService) *InventoryController {
	return &InventoryController{service: svc}
}

type reserveRequest struct {
	OrderID int64 `json:"orderId" binding:"required"`
	Items   []struct {
		ProductID string `json:"productId" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
}

type orderIDRequest struct {
	OrderID int64 `json:"orderId" binding:"required"`
}

func (ctrl *InventoryController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (ctrl *InventoryController) Reserve(c *gin.Context) {
	var req reserveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	items := make([]domain.ReserveItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.ReserveItem{ProductID: it.ProductID, Quantity: it.Quantity}
	}

	if err := ctrl.service.Reserve(req.OrderID, items); err != nil {
		if errors.Is(err, domain.ErrInsufficientStock) {
			httpx.Conflict(c, err.Error())
			return
		}
		if errors.Is(err, domain.ErrProductNotFound) {
			httpx.NotFound(c, err.Error())
			return
		}
		httpx.InternalError(c, "reserve failed")
		return
	}

	httpx.OK(c, gin.H{"message": "stock reserved"})
}

func (ctrl *InventoryController) Release(c *gin.Context) {
	var req orderIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.service.Release(req.OrderID); err != nil {
		httpx.InternalError(c, "release failed")
		return
	}

	httpx.OK(c, gin.H{"message": "stock released"})
}

func (ctrl *InventoryController) Confirm(c *gin.Context) {
	var req orderIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.BadRequest(c, err.Error())
		return
	}

	if err := ctrl.service.Confirm(req.OrderID); err != nil {
		httpx.InternalError(c, "confirm failed")
		return
	}

	httpx.OK(c, gin.H{"message": "stock confirmed"})
}

func (ctrl *InventoryController) GetStock(c *gin.Context) {
	productID := c.Param("productId")

	stock, found, err := ctrl.service.GetStock(productID)
	if err != nil {
		httpx.InternalError(c, "failed to get stock")
		return
	}
	if !found {
		httpx.NotFound(c, "product not found")
		return
	}

	httpx.OK(c, gin.H{
		"productId": stock.ProductID,
		"quantity":  stock.Quantity,
		"reserved":  stock.Reserved,
		"available": stock.Available(),
	})
}
