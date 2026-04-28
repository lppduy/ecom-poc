package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventorypb "github.com/lppduy/ecom-poc/gen/inventory"
	"github.com/lppduy/ecom-poc/services/order/internal/domain"
)

type InventoryGRPCClient struct {
	conn   *grpc.ClientConn
	client inventorypb.InventoryServiceClient
}

func NewInventoryGRPCClient(addr string) (*InventoryGRPCClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("inventory grpc dial %s: %w", addr, err)
	}
	return &InventoryGRPCClient{
		conn:   conn,
		client: inventorypb.NewInventoryServiceClient(conn),
	}, nil
}

func (c *InventoryGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *InventoryGRPCClient) Reserve(orderID int64, items []domain.OrderItem) error {
	pbItems := make([]*inventorypb.OrderItem, len(items))
	for i, it := range items {
		pbItems[i] = &inventorypb.OrderItem{
			ProductId: it.ProductID,
			Quantity:  int32(it.Quantity),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Reserve(ctx, &inventorypb.ReserveRequest{
		OrderId: orderID,
		Items:   pbItems,
	})
	return err
}

func (c *InventoryGRPCClient) Release(orderID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Release(ctx, &inventorypb.OrderIDRequest{OrderId: orderID})
	return err
}

func (c *InventoryGRPCClient) Confirm(orderID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.Confirm(ctx, &inventorypb.OrderIDRequest{OrderId: orderID})
	return err
}
