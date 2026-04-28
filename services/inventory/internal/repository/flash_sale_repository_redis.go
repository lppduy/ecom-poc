package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ErrSoldOut = errors.New("sold out")

const flashSaleKeyPrefix = "flash_sale:stock:"

// FlashSaleRepository handles atomic stock operations for flash sales using Redis.
// Redis is single-threaded so DECR/INCRBY are guaranteed atomic without extra locking.
type FlashSaleRepository struct {
	client *redis.Client
}

func NewFlashSaleRepository(client *redis.Client) *FlashSaleRepository {
	return &FlashSaleRepository{client: client}
}

func key(productID string) string {
	return flashSaleKeyPrefix + productID
}

// Init seeds the flash sale stock for a product. Safe to call multiple times (SET only if not exists).
func (r *FlashSaleRepository) Init(ctx context.Context, productID string, quantity int64) error {
	// SETNX - only set if key doesn't exist (idempotent init)
	ok, err := r.client.SetNX(ctx, key(productID), quantity, 0).Result()
	if err != nil {
		return fmt.Errorf("flash sale init: %w", err)
	}
	if !ok {
		return fmt.Errorf("flash sale for product %s already initialised", productID)
	}
	return nil
}

// ForceInit resets the flash sale stock (for testing/admin use).
func (r *FlashSaleRepository) ForceInit(ctx context.Context, productID string, quantity int64) error {
	return r.client.Set(ctx, key(productID), quantity, 0).Err()
}

// Reserve atomically decrements stock by qty.
// Returns ErrSoldOut if stock would go below 0.
func (r *FlashSaleRepository) Reserve(ctx context.Context, productID string, qty int64) error {
	result, err := r.client.DecrBy(ctx, key(productID), qty).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("flash sale not initialised for product %s", productID)
		}
		return fmt.Errorf("flash sale reserve: %w", err)
	}
	if result < 0 {
		// Roll back - put stock back
		r.client.IncrBy(ctx, key(productID), qty)
		return ErrSoldOut
	}
	return nil
}

// Stock returns current available flash sale stock.
func (r *FlashSaleRepository) Stock(ctx context.Context, productID string) (int64, error) {
	v, err := r.client.Get(ctx, key(productID)).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, fmt.Errorf("flash sale not initialised for product %s", productID)
		}
		return 0, err
	}
	return v, nil
}
