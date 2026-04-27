package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/lppduy/ecom-poc/services/cart/internal/domain"
	"github.com/redis/go-redis/v9"
)

const cartTTL = 24 * time.Hour

type RedisCartRepository struct {
	client *redis.Client
}

func NewRedisCartRepository(client *redis.Client) *RedisCartRepository {
	return &RedisCartRepository{client: client}
}

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func cartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}

// AddItem stores productId→quantity in a Redis Hash and refreshes TTL.
func (r *RedisCartRepository) AddItem(userID, productID string, quantity int) error {
	ctx := context.Background()
	key := cartKey(userID)
	if err := r.client.HSet(ctx, key, productID, quantity).Err(); err != nil {
		return err
	}
	return r.client.Expire(ctx, key, cartTTL).Err()
}

// ListItems returns all items in the cart Hash as CartItem slice.
func (r *RedisCartRepository) ListItems(userID string) ([]domain.CartItem, error) {
	ctx := context.Background()
	data, err := r.client.HGetAll(ctx, cartKey(userID)).Result()
	if err != nil {
		return nil, err
	}
	items := make([]domain.CartItem, 0, len(data))
	for productID, qtyStr := range data {
		qty, _ := strconv.Atoi(qtyStr)
		items = append(items, domain.CartItem{
			UserID:    userID,
			ProductID: productID,
			Quantity:  qty,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	return items, nil
}

// Clear removes the entire cart key.
func (r *RedisCartRepository) Clear(userID string) error {
	return r.client.Del(context.Background(), cartKey(userID)).Err()
}
