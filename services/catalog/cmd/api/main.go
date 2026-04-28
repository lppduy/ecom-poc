package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	catalogpb "github.com/lppduy/ecom-poc/gen/catalog"
	"github.com/lppduy/ecom-poc/services/catalog/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/catalog/internal/config"
	cataloggrpc "github.com/lppduy/ecom-poc/services/catalog/internal/grpc"
	"github.com/lppduy/ecom-poc/services/catalog/internal/repository"
	"github.com/lppduy/ecom-poc/services/catalog/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := repository.InitSchema(db); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	repo := repository.NewProductRepository(db)
	if err := repo.SeedDefaultsIfEmpty(); err != nil {
		log.Fatalf("failed to seed products: %v", err)
	}

	productService := service.NewProductService(repo)

	// gRPC server — serves StreamProducts for search service reindexing
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("catalog: gRPC listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		catalogpb.RegisterCatalogServiceServer(grpcServer, cataloggrpc.NewCatalogGRPCServer(productService))
		log.Printf("catalog gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("catalog: gRPC serve: %v", err)
		}
	}()

	productController := controller.NewProductController(productService)
	router := gin.Default()
	productController.RegisterRoutes(router)

	addr := ":" + cfg.Port
	fmt.Printf("catalog HTTP server started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
