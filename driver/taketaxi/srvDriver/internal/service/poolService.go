package service

import (
	"context"
	"fmt"
	"time"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	"driver/taketaxi/srvDriver/internal/cache"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PoolService struct {
	repo      *repository.DriverRepo
	mongoDb   *mongo.Database
	poolCache *cache.PoolCache
}

func NewPoolService(repo *repository.DriverRepo, mongoDb *mongo.Database, poolCache *cache.PoolCache) *PoolService {
	return &PoolService{repo: repo, mongoDb: mongoDb, poolCache: poolCache}
}

// IsSpecialOrder 判断订单是否为特殊订单（直接进抢单池，不走智能派单）
func IsSpecialOrder(originLat, originLng, destLat, destLng float64, estimateDistance int) bool {
	destDistance := haversine(originLat, originLng, destLat, destLng)
	if destDistance > 50000.0 {
		return true
	}
	if estimateDistance > 50000 {
		return true
	}
	return false
}

// PoolOrderResult ListPoolOrders 返回结果
type PoolOrderResult struct {
	Order       model.Order
	SecondsLeft int64
}

// ListPoolOrders 查看抢单池中的订单（status=7 抢单池中）
func (s *PoolService) ListPoolOrders(ctx context.Context, driverID int64, page, pageSize int32) ([]PoolOrderResult, int32, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > constants.PoolPageSizeMax {
		pageSize = constants.PoolPageSizeDefault
	}

	db := s.repo.GetDB()
	now := time.Now()

	// 先将已超时池订单批量取消
	db.Model(&model.Order{}).
		Where("status = ? AND pool_expire_at IS NOT NULL AND pool_expire_at < ?",
			constants.OrderStatusPool, now).
		Updates(map[string]interface{}{
			"status":        constants.OrderStatusCancelled,
			"cancel_by":     3,
			"cancel_time":   now,
			"cancel_reason": "抢单超时",
		})

	// 查询池中未超时订单：status=7, 且未过期
	var orders []model.Order
	offset := (page - 1) * pageSize
	if err := db.Where("status = ? AND (pool_expire_at IS NULL OR pool_expire_at > ?)",
		constants.OrderStatusPool, now).
		Order("pool_entered_at ASC"). // 等待时间最长的优先
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("list pool orders: %w", err)
	}

	// 总数
	var total int64
	if err := db.Model(&model.Order{}).Where("status = ? AND (pool_expire_at IS NULL OR pool_expire_at > ?)",
		constants.OrderStatusPool, now).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count pool orders: %w", err)
	}

	results := make([]PoolOrderResult, 0, len(orders))
	for _, o := range orders {
		secondsLeft := int64(0)
		if o.PoolExpireAt != nil {
			secondsLeft = int64(time.Until(*o.PoolExpireAt).Seconds())
			if secondsLeft < 0 {
				secondsLeft = 0
			}
		}
		results = append(results, PoolOrderResult{
			Order:       o,
			SecondsLeft: secondsLeft,
		})
	}

	return results, int32(total), nil
}

// GrabOrder 司机抢单
func (s *PoolService) GrabOrder(ctx context.Context, orderID, driverID int64) error {
	// 1-7: 司机校验（同原逻辑）
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverID)
	if err != nil {
		return fmt.Errorf("get driver: %w", err)
	}
	if driver.Status != constants.AccountStatusNormal {
		return errors.NewBusinessError(GrabErrAccountAbnormal, "您的账号状态异常，无法抢单")
	}
	realname, err := s.repo.GetRealnameByDriverId(ctx, driverID)
	if err != nil || realname.Status != constants.AuthStatusApproved {
		return errors.NewBusinessError(GrabErrRealname, "您的资质未审核通过，无法抢单")
	}
	vehicle, err := s.repo.GetVehicleByDriverId(ctx, driverID)
	if err != nil || vehicle.Status != constants.AuthStatusApproved {
		return errors.NewBusinessError(GrabErrVehicle, "您的车辆认证未通过，无法抢单")
	}
	if int8(driver.WorkStatus) != constants.WorkStatusListening {
		return errors.NewBusinessError(GrabErrNotListening, "请先开始听单后再抢单")
	}
	hasOrder, err := s.repo.HasOngoingOrder(ctx, driverID)
	if err != nil {
		return fmt.Errorf("check ongoing order: %w", err)
	}
	if hasOrder {
		return errors.NewBusinessError(GrabErrHasOrder, "您当前有进行中的订单，无法抢单")
	}
	if driver.DailyOrderLimit > 0 {
		today := time.Now().Format("2006-01-02")
		todayCount, err := s.repo.GetTodayOrderCount(ctx, driverID, today)
		if err != nil {
			return fmt.Errorf("get today order count: %w", err)
		}
		if todayCount >= driver.DailyOrderLimit {
			return errors.NewBusinessError(constants.GrabErrDailyLimit, "您已达到当日接单上限，无法继续抢单")
		}
	}

	// 8: 查订单
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return errors.NewBusinessError(constants.GrabErrOrderNotFound, "订单不存在")
	}

	// 9: 校验订单是否在抢单池中（status=7 抢单池中）
	if order.Status != constants.OrderStatusPool {
		return errors.NewBusinessError(constants.GrabErrNotInPool, "订单不在抢单池中")
	}

	// 10: 校验超时
	if order.PoolExpireAt != nil && time.Now().After(*order.PoolExpireAt) {
		return errors.NewBusinessError(constants.GrabErrTimeout, "抢单超时，订单已取消")
	}

	// 11: 城市匹配
	if order.CityId > 0 && driver.CityId > 0 && order.CityId != driver.CityId {
		return errors.NewBusinessError(constants.GrabErrCityMismatch, "该订单不在您的服务城市，无法抢单")
	}

	// 12: CAS 抢占：status=7 → status=2(已接单), driver_id=司机ID（跳过 status=1，抢单即确认）
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusPool, 0).
		Updates(map[string]interface{}{
			"driver_id": driverID,
			"status":    constants.OrderStatusAccepted,
		})
	if result.Error != nil {
		return fmt.Errorf("grab order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewBusinessError(constants.GrabErrGrabbed, "订单已被抢走")
	}

	// 13: 从 Redis 缓存中移除
	if s.poolCache != nil {
		_ = s.poolCache.RemoveFromPool(context.Background(), orderID, order.CityId)
	}

	// 14: 创建行程记录（抢单即确认，直接生成 trip_service）
	now := time.Now()
	if err := s.repo.GetDB().Exec(
		"INSERT INTO trip_service (order_id, driver_id, passenger_id, accept_time) VALUES (?, ?, ?, ?)",
		orderID, driverID, order.PassengerId, now,
	).Error; err != nil {
		fmt.Printf("[GrabOrder] create trip_service failed: %v\n", err)
	}

	// 16: 写派单日志
	s.repo.CreateDispatchLog(ctx, &model.DispatchLog{
		OrderId:      orderID,
		DriverId:     driverID,
		DispatchType: 2, // 抢单
		DispatchTime: now,
		Result:       1, // 接受
		ResponseTime: now,
	})

	// 17: 写状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: constants.WorkStatusListening,
		ToStatus:   constants.WorkStatusListening,
		Reason:     fmt.Sprintf("抢单成功 order_id=%d", orderID),
	})

	return nil
}

// OpsAddToPool 运营端手动将订单放入抢单池（status=0 → status=7）
func (s *PoolService) OpsAddToPool(ctx context.Context, orderID int64, timeoutSec int) error {
	if timeoutSec <= 0 {
		timeoutSec = 1800
	}
	now := time.Now()
	expireAt := now.Add(time.Duration(timeoutSec) * time.Second)

	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = 0", orderID, constants.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":          constants.OrderStatusPool,
			"pool_entered_at": now,
			"pool_expire_at":  expireAt,
			"pool_reason":     100, // 运营手动操作
		})
	if result.Error != nil {
		return fmt.Errorf("ops add to pool: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewBusinessError(1, "订单状态不允许放入抢单池（非待派单或已分配司机）")
	}

	// 写入 Redis 缓存
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ?", orderID).First(&order).Error; err == nil && s.poolCache != nil {
		go func() {
			_ = s.poolCache.AddToPool(context.Background(), orderID, order.CityId, map[string]interface{}{
				"order_id":          orderID,
				"origin_lat":        order.OriginLat,
				"origin_lng":        order.OriginLng,
				"origin_address":    order.OriginAddress,
				"dest_address":      order.DestAddress,
				"estimate_fee":      order.EstimateFee,
				"estimate_distance": order.EstimateDistance,
				"service_type":      order.ServiceType,
				"city_id":           order.CityId,
			})
		}()
	}

	fmt.Printf("[OpsAddToPool] order %d manually added to pool\n", orderID)
	return nil
}

// Grab 错误码
const (
	GrabErrAccountAbnormal = 5010
	GrabErrRealname        = 5011
	GrabErrVehicle         = 5012
	GrabErrNotListening    = 5013
	GrabErrHasOrder        = 5014
)
