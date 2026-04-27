package repository

import (
	"github.com/lppduy/ecom-poc/services/catalog/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List() ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Order("id asc").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) SeedDefaultsIfEmpty() error {
	var count int64
	if err := r.db.Model(&domain.Product{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seed := []domain.Product{
		{ID: "sku_iphone_15_128", Name: "iPhone 15 128GB", Price: 19990000},
		{ID: "sku_airpods_pro_2", Name: "AirPods Pro 2", Price: 5990000},
	}
	return r.db.Create(&seed).Error
}

func NewDB(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func InitSchema(db *gorm.DB) error {
	return db.AutoMigrate(&domain.Product{})
}
