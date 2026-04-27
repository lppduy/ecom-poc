package repository

import (
	"github.com/lppduy/ecom-poc/services/cart/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormCartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *GormCartRepository {
	return &GormCartRepository{db: db}
}

func (r *GormCartRepository) AddItem(userID, productID string, quantity int) error {
	item := domain.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}
	return r.db.Create(&item).Error
}

func (r *GormCartRepository) ListItems(userID string) ([]domain.CartItem, error) {
	var items []domain.CartItem
	err := r.db.Where("user_id = ?", userID).Order("id asc").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormCartRepository) Clear(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.CartItem{}).Error
}

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func InitSchema(db *gorm.DB) error {
	return db.AutoMigrate(&domain.CartItem{})
}
