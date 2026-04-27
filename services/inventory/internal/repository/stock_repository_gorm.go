package repository

import (
	"errors"

	"github.com/lppduy/ecom-poc/services/inventory/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormStockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) *GormStockRepository {
	return &GormStockRepository{db: db}
}

func (r *GormStockRepository) Reserve(orderID int64, items []domain.ReserveItem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Idempotency: if any reservation already exists for this order, skip.
		var count int64
		tx.Model(&domain.Reservation{}).Where("order_id = ?", orderID).Count(&count)
		if count > 0 {
			return nil
		}

		for _, item := range items {
			var stock domain.Stock
			// Lock the row for update to prevent race conditions.
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("product_id = ?", item.ProductID).
				First(&stock).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrProductNotFound
				}
				return err
			}

			if stock.Available() < item.Quantity {
				return domain.ErrInsufficientStock
			}

			if err := tx.Model(&domain.Stock{}).
				Where("product_id = ?", item.ProductID).
				Update("reserved", gorm.Expr("reserved + ?", item.Quantity)).Error; err != nil {
				return err
			}

			reservation := domain.Reservation{
				OrderID:   orderID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Status:    domain.ReservationStatusReserved,
			}
			if err := tx.Create(&reservation).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormStockRepository) Release(orderID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var reservations []domain.Reservation
		if err := tx.Where("order_id = ? AND status = ?", orderID, domain.ReservationStatusReserved).
			Find(&reservations).Error; err != nil {
			return err
		}
		if len(reservations) == 0 {
			return nil // already released or never reserved
		}

		for _, res := range reservations {
			if err := tx.Model(&domain.Stock{}).
				Where("product_id = ?", res.ProductID).
				Update("reserved", gorm.Expr("reserved - ?", res.Quantity)).Error; err != nil {
				return err
			}
		}

		return tx.Model(&domain.Reservation{}).
			Where("order_id = ? AND status = ?", orderID, domain.ReservationStatusReserved).
			Update("status", domain.ReservationStatusReleased).Error
	})
}

func (r *GormStockRepository) Confirm(orderID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var reservations []domain.Reservation
		if err := tx.Where("order_id = ? AND status = ?", orderID, domain.ReservationStatusReserved).
			Find(&reservations).Error; err != nil {
			return err
		}
		if len(reservations) == 0 {
			return nil // already confirmed or never reserved
		}

		for _, res := range reservations {
			// Deduct from both quantity and reserved simultaneously.
			if err := tx.Model(&domain.Stock{}).
				Where("product_id = ?", res.ProductID).
				Updates(map[string]interface{}{
					"quantity": gorm.Expr("quantity - ?", res.Quantity),
					"reserved": gorm.Expr("reserved - ?", res.Quantity),
				}).Error; err != nil {
				return err
			}
		}

		return tx.Model(&domain.Reservation{}).
			Where("order_id = ? AND status = ?", orderID, domain.ReservationStatusReserved).
			Update("status", domain.ReservationStatusConfirmed).Error
	})
}

func (r *GormStockRepository) GetStock(productID string) (domain.Stock, bool, error) {
	var stock domain.Stock
	err := r.db.Where("product_id = ?", productID).First(&stock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Stock{}, false, nil
		}
		return domain.Stock{}, false, err
	}
	return stock, true, nil
}

func (r *GormStockRepository) SeedDefaultsIfEmpty() error {
	var count int64
	if err := r.db.Model(&domain.Stock{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seed := []domain.Stock{
		{ProductID: "sku_iphone_15_128", Quantity: 100},
		{ProductID: "sku_airpods_pro_2", Quantity: 50},
	}
	return r.db.Create(&seed).Error
}
