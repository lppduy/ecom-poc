package response

import "github.com/lppduy/ecom-poc/services/order/internal/domain"

type OrderResponse struct {
	ID     int64  `json:"id"`
	UserID string `json:"userId"`
	Status string `json:"status"`
}

func FromDomain(order domain.Order) OrderResponse {
	return OrderResponse{
		ID:     order.ID,
		UserID: order.UserID,
		Status: order.Status,
	}
}
