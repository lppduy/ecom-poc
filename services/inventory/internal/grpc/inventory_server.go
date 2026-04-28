package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	inventorypb "github.com/lppduy/ecom-poc/gen/inventory"
	"github.com/lppduy/ecom-poc/services/inventory/internal/domain"
	"github.com/lppduy/ecom-poc/services/inventory/internal/service"
)

// InventoryServer implements the gRPC InventoryServiceServer interface.
type InventoryServer struct {
	inventorypb.UnimplementedInventoryServiceServer
	svc service.InventoryService
}

func NewInventoryServer(svc service.InventoryService) *InventoryServer {
	return &InventoryServer{svc: svc}
}

func (s *InventoryServer) Reserve(ctx context.Context, req *inventorypb.ReserveRequest) (*inventorypb.ReserveResponse, error) {
	items := make([]domain.ReserveItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = domain.ReserveItem{
			ProductID: it.ProductId,
			Quantity:  int(it.Quantity),
		}
	}

	if err := s.svc.Reserve(req.OrderId, items); err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientStock):
			return nil, status.Errorf(codes.ResourceExhausted, "insufficient stock: %v", err)
		case errors.Is(err, domain.ErrProductNotFound):
			return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
		default:
			return nil, status.Errorf(codes.Internal, "reserve failed: %v", err)
		}
	}

	return &inventorypb.ReserveResponse{Success: true, Message: "stock reserved"}, nil
}

func (s *InventoryServer) Confirm(ctx context.Context, req *inventorypb.OrderIDRequest) (*inventorypb.CommonResponse, error) {
	if err := s.svc.Confirm(req.OrderId); err != nil {
		return nil, status.Errorf(codes.Internal, "confirm failed: %v", err)
	}
	return &inventorypb.CommonResponse{Success: true}, nil
}

func (s *InventoryServer) Release(ctx context.Context, req *inventorypb.OrderIDRequest) (*inventorypb.CommonResponse, error) {
	if err := s.svc.Release(req.OrderId); err != nil {
		return nil, status.Errorf(codes.Internal, "release failed: %v", err)
	}
	return &inventorypb.CommonResponse{Success: true}, nil
}
