package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	poolOrderKey     = "pool:orders:%d"       // ZSET: 按城市分组的池订单集合, score=时间戳, member=order_id
	poolOrderInfoKey = "pool:order:%d"         // HASH: 订单详细信息缓存
	poolNewOrderChan = "channel:pool:new_order" // 新订单通知 channel
	poolOrderTTL     = 2 * time.Hour           // 缓存 TTL（比池超时长即可）
)

type PoolCache struct {
	rdb *redis.Client
}

func NewPoolCache(rdb *redis.Client) *PoolCache {
	return &PoolCache{rdb: rdb}
}

// AddToPool 将订单加入抢单池缓存
func (c *PoolCache) AddToPool(ctx context.Context, orderID, cityID int64, info map[string]interface{}) error {
	pipe := c.rdb.Pipeline()

	now := float64(time.Now().Unix())

	// ZSET: pool:orders:{city_id} → score=当前时间戳 member=order_id
	pipe.ZAdd(ctx, fmt.Sprintf(poolOrderKey, cityID), redis.Z{
		Score:  now,
		Member: orderID,
	})

	// HASH: pool:order:{order_id} → 订单基础信息
	infoKey := fmt.Sprintf(poolOrderInfoKey, orderID)
	infoMap := make(map[string]interface{}, len(info))
	for k, v := range info {
		infoMap[k] = v
	}
	pipe.HSet(ctx, infoKey, infoMap)
	pipe.Expire(ctx, infoKey, poolOrderTTL)

	// PUBLISH: 通知所有订阅方有新订单进池
	msg, _ := json.Marshal(map[string]interface{}{
		"order_id": orderID,
		"city_id":  cityID,
		"time":     time.Now().Unix(),
	})
	pipe.Publish(ctx, poolNewOrderChan, string(msg))

	_, err := pipe.Exec(ctx)
	return err
}

// RemoveFromPool 从抢单池缓存中移除订单
func (c *PoolCache) RemoveFromPool(ctx context.Context, orderID, cityID int64) error {
	pipe := c.rdb.Pipeline()
	pipe.ZRem(ctx, fmt.Sprintf(poolOrderKey, cityID), orderID)
	pipe.Del(ctx, fmt.Sprintf(poolOrderInfoKey, orderID))
	_, err := pipe.Exec(ctx)
	return err
}

// GetPoolOrderIDs 获取按时间排序的池订单 ID 列表（最近优先）
func (c *PoolCache) GetPoolOrderIDs(ctx context.Context, cityID int64, start, stop int64) ([]int64, error) {
	key := fmt.Sprintf(poolOrderKey, cityID)
	// ZRangeArgs with Rev: 按 score 降序（最新进池优先）
	members, err := c.rdb.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     key,
		Start:   start,
		Stop:    stop,
		Rev:     true,
	}).Result()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0, len(members))
	for _, m := range members {
		var id int64
		fmt.Sscanf(m, "%d", &id)
		ids = append(ids, id)
	}
	return ids, nil
}

// GetPoolOrderCount 获取城市池订单总数
func (c *PoolCache) GetPoolOrderCount(ctx context.Context, cityID int64) (int64, error) {
	return c.rdb.ZCard(ctx, fmt.Sprintf(poolOrderKey, cityID)).Result()
}

// IsOrderInPool 判断订单是否还在池中
func (c *PoolCache) IsOrderInPool(ctx context.Context, orderID int64) (bool, error) {
	n, err := c.rdb.Exists(ctx, fmt.Sprintf(poolOrderInfoKey, orderID)).Result()
	return n > 0, err
}

// PublishNewOrder 广播新池订单通知（轻量级消息）
func (c *PoolCache) PublishNewOrder(ctx context.Context, orderID, cityID int64) error {
	msg, _ := json.Marshal(map[string]interface{}{
		"order_id": orderID,
		"city_id":  cityID,
		"time":     time.Now().Unix(),
	})
	return c.rdb.Publish(ctx, poolNewOrderChan, string(msg)).Err()
}
