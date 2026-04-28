package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/lppduy/ecom-poc/services/cart/internal/service"
)

type CartController struct {
	service service.CartService
}

func NewCartController(svc service.CartService) *CartController {
	return &CartController{service: svc}
}

func (c *CartController) RegisterRoutes(router *gin.Engine, jwtSecret string) {
	router.GET("/health", c.health)

	protected := router.Group("/", jwtutil.RequireAuth(jwtSecret))
	{
		protected.GET("/cart", c.getCart)
		protected.POST("/cart/items", c.addCartItem)
		protected.POST("/cart/clear", c.clearCart)
	}
}

type addCartItemRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

func (c *CartController) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *CartController) addCartItem(ctx *gin.Context) {
	userID := jwtutil.GetUserID(ctx)

	var req addCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if err := c.service.AddItem(userID, req.ProductID, req.Quantity); err != nil {
		if errors.Is(err, service.ErrMissingProductID) || errors.Is(err, service.ErrInvalidQuantity) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add item"})
		return
	}

	items, err := c.service.GetItems(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list items"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "item added to cart", "items": items})
}

func (c *CartController) getCart(ctx *gin.Context) {
	userID := jwtutil.GetUserID(ctx)
	items, err := c.service.GetItems(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cart"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (c *CartController) clearCart(ctx *gin.Context) {
	userID := jwtutil.GetUserID(ctx)
	if err := c.service.Clear(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear cart"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}
