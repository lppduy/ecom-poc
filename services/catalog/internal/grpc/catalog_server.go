package grpc

import (
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogpb "github.com/lppduy/ecom-poc/gen/catalog"
	"github.com/lppduy/ecom-poc/services/catalog/internal/service"
)

// CatalogGRPCServer implements catalogpb.CatalogServiceServer.
type CatalogGRPCServer struct {
	catalogpb.UnimplementedCatalogServiceServer
	productService service.ProductService
}

func NewCatalogGRPCServer(svc service.ProductService) *CatalogGRPCServer {
	return &CatalogGRPCServer{productService: svc}
}

// StreamProducts streams all products one-by-one to the caller.
// Server-side streaming: catalog sends N Product messages, then closes the stream.
func (s *CatalogGRPCServer) StreamProducts(
	req *catalogpb.StreamProductsRequest,
	stream catalogpb.CatalogService_StreamProductsServer,
) error {
	products, err := s.productService.ListProducts()
	if err != nil {
		return status.Errorf(codes.Internal, "list products: %v", err)
	}

	for _, p := range products {
		if err := stream.Send(&catalogpb.Product{
			Id:    p.ID,
			Name:  p.Name,
			Price: int32(p.Price),
		}); err != nil {
			return err
		}
		log.Printf("[catalog-grpc] streamed product id=%s name=%s", p.ID, p.Name)
	}
	return nil
}
