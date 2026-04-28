package client

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	catalogpb "github.com/lppduy/ecom-poc/gen/catalog"
	"github.com/lppduy/ecom-poc/services/search/internal/domain"
)

// CatalogGRPCClient fetches products from catalog via server-side streaming gRPC.
// Catalog streams products one-by-one; client collects them until EOF.
// This is the active implementation replacing CatalogHTTPClient for reindexing.
type CatalogGRPCClient struct {
	client catalogpb.CatalogServiceClient
	conn   *grpc.ClientConn
}

func NewCatalogGRPCClient(addr string) (*CatalogGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &CatalogGRPCClient{
		client: catalogpb.NewCatalogServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *CatalogGRPCClient) Close() error {
	return c.conn.Close()
}

// FetchAllProducts opens a server-side streaming call and collects every Product
// message until the server closes the stream (EOF).
func (c *CatalogGRPCClient) FetchAllProducts() ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := c.client.StreamProducts(ctx, &catalogpb.StreamProductsRequest{})
	if err != nil {
		return nil, err
	}

	var products []domain.Product
	for {
		p, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		products = append(products, domain.Product{
			ID:    p.Id,
			Name:  p.Name,
			Price: int(p.Price),
		})
	}
	return products, nil
}
