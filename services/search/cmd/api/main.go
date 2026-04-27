package main

import (
	"fmt"
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/search/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/search/internal/api/routes"
	"github.com/lppduy/ecom-poc/services/search/internal/client"
	"github.com/lppduy/ecom-poc/services/search/internal/config"
	"github.com/lppduy/ecom-poc/services/search/internal/repository"
	"github.com/lppduy/ecom-poc/services/search/internal/service"
)

func main() {
	cfg := config.Load()

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.ESAddress},
	})
	if err != nil {
		log.Fatalf("failed to create ES client: %v", err)
	}

	// Verify ES is reachable.
	res, err := esClient.Info()
	if err != nil {
		log.Fatalf("failed to connect to Elasticsearch: %v", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Fatalf("Elasticsearch returned error: %s", res.String())
	}
	log.Printf("connected to Elasticsearch at %s", cfg.ESAddress)

	repo := repository.NewESSearchRepository(esClient)
	searchService := service.NewSearchService(repo)
	catalogClient := client.NewCatalogHTTPClient(cfg.CatalogBaseURL)

	// Reindex products from catalog on startup (best-effort).
	go func() {
		products, err := catalogClient.FetchAllProducts()
		if err != nil {
			log.Printf("warn: startup reindex failed: %v", err)
			return
		}
		if err := searchService.BulkIndex(products); err != nil {
			log.Printf("warn: startup bulk index failed: %v", err)
			return
		}
		log.Printf("indexed %d products from catalog on startup", len(products))
	}()

	searchController := controller.NewSearchController(searchService, catalogClient)

	router := gin.Default()
	routes.Register(router, searchController)

	addr := ":" + cfg.Port
	fmt.Printf("service started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
