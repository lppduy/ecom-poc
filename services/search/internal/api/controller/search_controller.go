package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/search/internal/api/httpx"
	"github.com/lppduy/ecom-poc/services/search/internal/client"
	"github.com/lppduy/ecom-poc/services/search/internal/domain"
	"github.com/lppduy/ecom-poc/services/search/internal/service"
)

type SearchController struct {
	service       service.SearchService
	catalogClient client.CatalogClient
}

func NewSearchController(svc service.SearchService, catalogClient client.CatalogClient) *SearchController {
	return &SearchController{service: svc, catalogClient: catalogClient}
}

func (ctrl *SearchController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GET /search?q=iphone&minPrice=1000&maxPrice=50000
func (ctrl *SearchController) Search(c *gin.Context) {
	q := c.Query("q")
	minPrice := queryInt(c, "minPrice")
	maxPrice := queryInt(c, "maxPrice")

	result, err := ctrl.service.Search(q, minPrice, maxPrice)
	if err != nil {
		httpx.InternalError(c, "search failed")
		return
	}

	httpx.OK(c, result)
}

// POST /search/reindex — pulls all products from catalog and re-indexes into ES
func (ctrl *SearchController) Reindex(c *gin.Context) {
	products, err := ctrl.catalogClient.FetchAllProducts()
	if err != nil {
		httpx.InternalError(c, "failed to fetch products from catalog")
		return
	}

	if err := ctrl.service.BulkIndex(products); err != nil {
		httpx.InternalError(c, "bulk index failed")
		return
	}

	httpx.OK(c, gin.H{"indexed": len(products)})
}

// POST /search/index — index a single product (called manually or by catalog event)
func (ctrl *SearchController) IndexOne(c *gin.Context) {
	var p domain.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		httpx.BadRequest(c, "invalid json")
		return
	}
	if err := ctrl.service.IndexProduct(p); err != nil {
		httpx.InternalError(c, "index failed")
		return
	}
	httpx.OK(c, gin.H{"message": "indexed"})
}

func queryInt(c *gin.Context, key string) int {
	v := c.Query(key)
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
}
