package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/srvDriver/internal/cache"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DispatchService struct {
	repo      *repository.DriverRepo
	mongoDb   *mongo.Database
	cfg       *config.DispatchConfig
	poolCache *cache.PoolCache
}

func NewDispatchService(repo *repository.DriverRepo, mongoDb *mongo.Database, cfg *config.DispatchConfig, poolCache *cache.PoolCache) *DispatchService {
	return &DispatchService{repo: repo, mongoDb: mongoDb, cfg: cfg, poolCache: poolCache}
}

// Dispatch 派单核心逻辑：找附近司机，逐条件过滤，成功则 CAS 抢单
// 如果所有候选司机都拒绝或附近无合适司机，自动放入抢单池
func (s *DispatchService) Dispatch(ctx context.Context, orderID int64, serviceType int32, originLat, originLng float64, passengerID int64) error {
	radiusKm := s.cfg.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 3.0
	}
	radiusMeters := radiusKm * 1000.0

	err := s.tryDispatch(ctx, orderID, serviceType, originLat, originLng, passengerID, 0, radiusMeters)
	if err != nil {
		fmt.Printf("[Dispatch] order %d no eligible driver found, move to pool: %v\n", orderID, err)
		s.moveToPool(ctx, orderID)
	}
	return nil
}

// RejectOrder 司机拒绝接单（CAS 重置状态，然后重新派单给下一个候选）
func (s *DispatchService) RejectOrder(ctx context.Context, orderID, driverID int64) error {
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusDispatched, driverID).
		Updates(map[string]interface{}{
			"status":    constants.OrderStatusPending,
			"driver_id": 0,
		})
	if result.Error != nil {
		return fmt.Errorf("reject order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法拒绝")
	}

	s.logDispatchResult(ctx, orderID, driverID, 0, 0, 0, bson.M{"note": "司机拒绝接单"})

	// 重新派单给下一个候选司机（排除当前司机）
	s.retryDispatch(ctx, orderID, driverID)
	return nil
}

// retryDispatch 重新派单（跳过指定 driverID），失败则放入抢单池
func (s *DispatchService) retryDispatch(ctx context.Context, orderID, excludeDriverID int64) {
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ?", orderID, constants.OrderStatusPending).First(&order).Error; err != nil {
		fmt.Printf("[retryDispatch] order %d not found or status changed, err=%v\n", orderID, err)
		return
	}

	radiusKm := s.cfg.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 3.0
	}
	radiusMeters := radiusKm * 1000.0

	err := s.tryDispatch(ctx, orderID, int32(order.ServiceType), order.OriginLat, order.OriginLng, order.PassengerId, excludeDriverID, radiusMeters)
	if err != nil {
		fmt.Printf("[retryDispatch] order %d all drivers exhausted, moving to pool\n", orderID)
		s.moveToPool(ctx, orderID)
	} else {
		fmt.Printf("[retryDispatch] order %d re-dispatch success\n", orderID)
	}
}

// tryDispatch 核心派单逻辑（可指定排除司机和搜索半径）
func (s *DispatchService) tryDispatch(ctx context.Context, orderID int64, serviceType int32, originLat, originLng float64, passengerID, excludeDriverID int64, radiusMeters float64) error {
	minScore := s.cfg.MinServiceScore
	if minScore <= 0 {
		minScore = 60.0
	}

	candidates, err := s.findNearbyDrivers(ctx, originLat, originLng, radiusMeters)
	if err != nil {
		return fmt.Errorf("query nearby drivers: %w", err)
	}
	if len(candidates) == 0 {
		return errors.NewDispatchRejectError(constants.DispatchTooFar, "附近没有可用司机")
	}

	for _, c := range candidates {
		driverID := int64(c["driver_id"].(int64))

		if excludeDriverID > 0 && driverID == excludeDriverID {
			continue
		}

		driver, err := s.repo.GetDriverSByDriverId(ctx, driverID)
		if err != nil {
			continue
		}

		// C1: 听单中
		if int8(driver.WorkStatus) != constants.WorkStatusListening {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchNotListening, originLat, originLng, c)
			continue
		}

		// C6: 服务分
		if driver.ServiceScore < minScore {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchLowScore, originLat, originLng, c)
			continue
		}

		// C4: 无进行中订单
		hasOrder, err := s.repo.HasOngoingOrder(ctx, driverID)
		if err != nil || hasOrder {
			if hasOrder {
				s.logDispatchResult(ctx, orderID, driverID, constants.DispatchHasOngoingOrder, originLat, originLng, c)
			}
			continue
		}

		// C5: 接单上限
		if driver.DailyOrderLimit > 0 {
			today := time.Now().Format("2006-01-02")
			todayCount, err := s.repo.GetTodayOrderCount(ctx, driverID, today)
			if err != nil || todayCount >= driver.DailyOrderLimit {
				if todayCount >= driver.DailyOrderLimit {
					s.logDispatchResult(ctx, orderID, driverID, constants.DispatchDailyLimit, originLat, originLng, c)
				}
				continue
			}
		}

		// C2: 距离
		distance := haversine(originLat, originLng, c["lat"].(float64), c["lng"].(float64))
		if distance > radiusMeters {
			s.logDispatchResult(ctx, orderID, driverID, constants.DispatchTooFar, originLat, originLng, c)
			continue
		}

		// C3: 车型
		vehicle, err := s.repo.GetVehicleByDriverIdAndStatus(ctx, driverID)
		if err != nil || int8(vehicle.ServiceType) != int8(serviceType) {
			if vehicle != nil {
				s.logDispatchResult(ctx, orderID, driverID, constants.DispatchVehicleMismatch, originLat, originLng, c)
			}
			continue
		}

		// C7: CAS 抢占
		affected, err := s.claimOrder(ctx, orderID, driverID)
		if err != nil || affected == 0 {
			continue
		}

		s.logDispatchResult(ctx, orderID, driverID, 0, originLat, originLng, c)
		return nil
	}

	return errors.NewDispatchRejectError(constants.DispatchTooFar, "附近没有符合条件的司机")
}

// moveToPool 将订单自动放入抢单池（所有候选司机拒绝/无合适司机时）
func (s *DispatchService) moveToPool(ctx context.Context, orderID int64) {
	now := time.Now()
	timeout := s.cfg.PoolTimeoutSec
	if timeout <= 0 {
		timeout = constants.PoolOrderTimeoutSec
	}
	expireAt := now.Add(time.Duration(timeout) * time.Second)

	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = 0", orderID, constants.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":          constants.OrderStatusPool,
			"pool_entered_at": now,
			"pool_expire_at":  expireAt,
			"pool_reason":     constants.PoolReasonAllRejected,
		})
	if result.Error != nil {
		fmt.Printf("[moveToPool] order %d db error: %v\n", orderID, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		return // CAS failed, order was claimed by another driver
	}

	// 写入 Redis 缓存
	if s.poolCache != nil {
		var order model.Order
		if err := s.repo.GetDB().Where("order_id = ?", orderID).First(&order).Error; err == nil {
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
	}

	fmt.Printf("[moveToPool] order %d moved to pool, expire_at=%v\n", orderID, expireAt)
}

// AcceptOrder 司机接单
func (s *DispatchService) AcceptOrder(ctx context.Context, orderID, driverID int64) error {
	fmt.Printf("[AcceptOrder] called: orderID=%d driverID=%d\n", orderID, driverID)
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusDispatched, driverID).
		Update("status", constants.OrderStatusAccepted)
	if result.Error != nil {
		return fmt.Errorf("accept order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法接单")
	}

	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ?", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("query order: %w", err)
	}

	now := time.Now()
	if err := s.repo.GetDB().Exec(
		"INSERT INTO trip_service (order_id, driver_id, passenger_id, accept_time) VALUES (?, ?, ?, ?)",
		orderID, driverID, order.PassengerId, now,
	).Error; err != nil {
		fmt.Printf("[AcceptOrder] create trip_service failed: %v\n", err)
		return fmt.Errorf("create trip service: %w", err)
	}

	return nil
}

// 以下方法保持原样：CancelOrder, DriverArrive, VerifyPassengerPhone, StartTrip, EndTrip

func (s *DispatchService) CancelOrder(ctx context.Context, orderID, driverID int64, reason string) error {
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND driver_id = ?", orderID, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单不存在")
	}
	if order.Status < constants.OrderStatusAccepted || order.Status > constants.OrderStatusInTrip {
		return errors.NewDispatchRejectError(1, "当前订单状态不允许取消")
	}

	now := time.Now()
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND driver_id = ? AND status IN ?", orderID, driverID,
			[]int8{constants.OrderStatusAccepted, constants.OrderStatusArrived, constants.OrderStatusInTrip}).
		Updates(map[string]interface{}{
			"status":        constants.OrderStatusCancelled,
			"cancel_by":     2,
			"cancel_reason": reason,
			"cancel_time":   now,
		})
	if result.Error != nil {
		return fmt.Errorf("cancel order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法取消")
	}

	s.repo.CreateDispatchLog(ctx, &model.DispatchLog{
		OrderId:      orderID,
		DriverId:     driverID,
		DispatchType: 2,
		DispatchTime: now,
		Result:       3,
		ResponseTime: now,
		RejectReason: reason,
	})
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: order.Status,
		ToStatus:   constants.OrderStatusCancelled,
		Reason:     fmt.Sprintf("司机主动取消 order_id=%d reason=%s", orderID, reason),
	})
	return nil
}

func (s *DispatchService) DriverArrive(ctx context.Context, orderID, driverID int64) error {
	arriveRadius := s.cfg.ArriveCheckRadius
	if arriveRadius <= 0 {
		arriveRadius = 30.0
	}

	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusAccepted, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法确认到达")
	}

	if s.mongoDb == nil {
		return fmt.Errorf("mongodb not available")
	}
	mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var loc struct {
		Lat float64 `bson:"lat"`
		Lng float64 `bson:"lng"`
	}
	if err := s.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err != nil {
		return errors.NewDispatchRejectError(1, "无法获取司机位置")
	}

	distance := haversine(order.OriginLat, order.OriginLng, loc.Lat, loc.Lng)
	if distance > arriveRadius {
		return errors.NewDispatchRejectError(1, fmt.Sprintf("您距离上车点还有%d米，请靠近后再确认", int(distance)))
	}

	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusAccepted, driverID).
		Update("status", constants.OrderStatusArrived)
	if result.Error != nil {
		return fmt.Errorf("update order status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法确认到达")
	}

	s.repo.GetDB().Model(&model.TripService{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Update("arrive_time", time.Now())
	return nil
}

func (s *DispatchService) VerifyPassengerPhone(ctx context.Context, orderID, driverID int64, phoneLast4 string) error {
	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusArrived, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法验证")
	}
	if len(order.PassengerMobile) < 4 {
		return errors.NewDispatchRejectError(constants.ErrCodePhoneMismatch, "订单手机号格式异常")
	}
	mobileLast4 := order.PassengerMobile[len(order.PassengerMobile)-4:]
	if mobileLast4 != phoneLast4 {
		return errors.NewDispatchRejectError(constants.ErrCodePhoneMismatch, "乘客手机号后四位不正确")
	}
	return nil
}

func (s *DispatchService) StartTrip(ctx context.Context, orderID, driverID int64) error {
	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusArrived, driverID).
		Update("status", constants.OrderStatusInTrip)
	if result.Error != nil {
		return fmt.Errorf("start trip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法开始行程")
	}
	s.repo.GetDB().Model(&model.TripService{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Update("start_time", time.Now())
	return nil
}

func (s *DispatchService) EndTrip(ctx context.Context, orderID, driverID int64) error {
	endTripRadius := s.cfg.EndTripCheckRadius
	if endTripRadius <= 0 {
		endTripRadius = 30.0
	}

	var order model.Order
	if err := s.repo.GetDB().Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusInTrip, driverID).First(&order).Error; err != nil {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法结束行程")
	}

	var trip model.TripService
	if err := s.repo.GetDB().Where("order_id = ? AND driver_id = ?", orderID, driverID).First(&trip).Error; err != nil {
		return errors.NewDispatchRejectError(1, "行程记录不存在")
	}

	if s.mongoDb == nil {
		return fmt.Errorf("mongodb not available")
	}
	mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var loc struct {
		Lat float64 `bson:"lat"`
		Lng float64 `bson:"lng"`
	}
	if err := s.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err != nil {
		return errors.NewDispatchRejectError(1, "无法获取司机位置")
	}

	distance := haversine(loc.Lat, loc.Lng, order.DestLat, order.DestLng)
	if distance > endTripRadius {
		return errors.NewDispatchRejectError(1, fmt.Sprintf("您距离目的地还有%d米，请到达后再点击结束", int(distance)))
	}

	now := time.Now()
	actualDistance := int(haversine(order.OriginLat, order.OriginLng, order.DestLat, order.DestLng))
	tripDuration := 0
	if !trip.StartTime.IsZero() {
		tripDuration = int(now.Sub(trip.StartTime).Seconds())
	}

	result := s.repo.GetDB().Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = ?", orderID, constants.OrderStatusInTrip, driverID).
		Updates(map[string]interface{}{
			"status":          constants.OrderStatusCompleted,
			"actual_distance": actualDistance,
			"actual_duration": tripDuration,
		})
	if result.Error != nil {
		return fmt.Errorf("end trip: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NewDispatchRejectError(1, "订单状态已变更，无法结束行程")
	}

	s.repo.GetDB().Model(&model.TripService{}).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		Updates(map[string]interface{}{
			"end_time":      now,
			"trip_duration": tripDuration,
			"trip_distance": actualDistance,
		})

	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: constants.OrderStatusInTrip,
		ToStatus:   constants.OrderStatusCompleted,
		Reason:     fmt.Sprintf("到达目的地 order_id=%d distance=%dm", orderID, int(distance)),
	})
	return nil
}

// findNearbyDrivers 从 MongoDB 查找 radius 内的司机
func (s *DispatchService) findNearbyDrivers(ctx context.Context, lat, lng, maxDistance float64) ([]bson.M, error) {
	if s.mongoDb == nil {
		return nil, fmt.Errorf("mongodb not available")
	}

	collection := s.mongoDb.Collection("driver_local")
	filter := bson.M{
		"loc": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lng, lat},
				},
				"$maxDistance": maxDistance,
			},
		},
		"updated_at": bson.M{
			"$gte": time.Now().Unix() - 300,
		},
	}

	opts := options.Find().SetLimit(50)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// claimOrder CAS 抢占：status=0, driver_id=0 → status=1, driver_id=driverID
func (s *DispatchService) claimOrder(ctx context.Context, orderID, driverID int64) (int64, error) {
	result := s.repo.GetDB().WithContext(ctx).
		Model(&model.Order{}).
		Where("order_id = ? AND status = ? AND driver_id = 0", orderID, constants.OrderStatusPending).
		Updates(map[string]interface{}{
			"driver_id": driverID,
			"status":    constants.OrderStatusDispatched,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// logDispatchResult 记录派单日志
func (s *DispatchService) logDispatchResult(ctx context.Context, orderID, driverID int64, rejectCode int, originLat, originLng float64, candidate bson.M) {
	driverLat, _ := candidate["lat"].(float64)
	driverLng, _ := candidate["lng"].(float64)
	distance := int(haversine(originLat, originLng, driverLat, driverLng))

	reason := fmt.Sprintf("派单失败 order_id=%d code=%d distance=%dm", orderID, rejectCode, distance)
	if rejectCode == 0 {
		reason = fmt.Sprintf("派单成功 order_id=%d distance=%dm", orderID, distance)
	}

	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverID,
		FromStatus: constants.WorkStatusListening,
		ToStatus:   constants.WorkStatusListening,
		Reason:     reason,
	})
}

// haversine 计算两点间距离(米)
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
