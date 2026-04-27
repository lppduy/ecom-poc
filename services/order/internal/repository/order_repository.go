package repository

import (
	"context"
	"errors"

	"github.com/lppduy/ecom-poc/services/order/internal/domain"
	"gorm.io/gorm"
)

var ErrDuplicateIdempotencyKey = errors.New("duplicate idempotency key")

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) FindByID(id string) (domain.Order, bool, error) {
	var found domain.Order
	err := r.db.Where("id = ?", id).First(&found).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Order{}, false, nil
		}
		return domain.Order{}, false, err
	}
	return found, true, nil
}

func (r *OrderRepository) FindByIdempotencyKey(key string) (domain.Order, bool, error) {
	var found domain.Order
	err := r.db.Where("idempotency_key = ?", key).First(&found).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Order{}, false, nil
		}
		return domain.Order{}, false, err
	}
	return found, true, nil
}

func (r *OrderRepository) CreateWithItems(ctx context.Context, userID, idempotencyKey string, items []domain.OrderItem) (domain.Order, error) {
	var created domain.Order
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		created = domain.Order{
			UserID:         userID,
			Status:         "PENDING",
			IdempotencyKey: idempotencyKey,
		}
		if err := tx.Create(&created).Error; err != nil {
			if isDuplicateKeyError(err) {
				return ErrDuplicateIdempotencyKey
			}
			return err
		}

		for _, item := range items {
			record := domain.OrderItem{
				OrderID:   created.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}
	return created, nil
}
