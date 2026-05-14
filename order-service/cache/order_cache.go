package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yessenchik/order-service/domain"
	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewOrderCache(addr string, ttl time.Duration) *OrderCache {
	return &OrderCache{rdb: redis.NewClient(&redis.Options{Addr: addr}), ttl: ttl}
}

func orderKey(orderID string) string { return fmt.Sprintf("order:%s", orderID) }

func (c *OrderCache) Get(ctx context.Context, orderID string) (*domain.Order, bool) {
	data, err := c.rdb.Get(ctx, orderKey(orderID)).Bytes()
	if err != nil {
		return nil, false
	}
	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, false
	}
	return &order, true
}

func (c *OrderCache) Set(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, orderKey(order.ID), data, c.ttl).Err()
}

func (c *OrderCache) Delete(ctx context.Context, orderID string) error {
	return c.rdb.Del(ctx, orderKey(orderID)).Err()
}
