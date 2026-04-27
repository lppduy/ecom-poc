package controller

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/cart/internal/service"
)

type CartController struct {
	service       service.CartService
	defaultUserID string
}

func NewCartController(service service.CartService, defaultUserID string) *CartController {
	return &CartController{service: service, defaultUserID: defaultUserID}
}

func (c *CartController) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", c.health)
	router.GET("/cart", c.getCart)
	router.POST("/cart/items", c.addCartItem)
	router.POST("/cart/clear", c.clearCart)
}

type addCartItemRequest struct {
	UserID    string `json:"userId"`
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

func (c *CartController) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *CartController) addCartItem(ctx *gin.Context) {
	var req addCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	userID := req.UserID
	if userID == "" {
		userID = c.resolveUserID(ctx)
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
	userID := c.resolveUserID(ctx)
	items, err := c.service.GetItems(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cart"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"items": items})
}

func (c *CartController) clearCart(ctx *gin.Context) {
	userID := c.resolveUserID(ctx)
	if err := c.service.Clear(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear cart"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}

func (c *CartController) resolveUserID(ctx *gin.Context) string {
	if q := strings.TrimSpace(ctx.Query("userId")); q != "" {
		return q
	}
	if h := strings.TrimSpace(ctx.GetHeader("X-User-Id")); h != "" {
		return h
	}
	return c.defaultUserID
}
