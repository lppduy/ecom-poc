package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/catalog/internal/service"
)

type ProductController struct {
	service service.ProductService
}

func NewProductController(service service.ProductService) *ProductController {
	return &ProductController{service: service}
}

func (c *ProductController) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", c.health)
	router.GET("/products", c.listProducts)
}

func (c *ProductController) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *ProductController) listProducts(ctx *gin.Context) {
	products, err := c.service.ListProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}
	ctx.JSON(http.StatusOK, products)
}
