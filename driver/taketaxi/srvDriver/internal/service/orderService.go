package service

import (
	"context"
	"fmt"
	"time"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	driver "driver/taketaxi/common/kitexGen/driver"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"

	"gorm.io/gorm"
)

type OrderService struct {
	repo *repository.DriverRepo
	db   *gorm.DB
}

func NewOrderService(repo *repository.DriverRepo) *OrderService {
	return &OrderService{repo: repo, db: repo.GetDB()}
}

// ListOrders 查询司机订单列表
func (s *OrderService) ListOrders(ctx context.Context, driverID int64, date string, cursor int32, isAll bool) ([]model.Order, error) {
	const pageSize = 20
	if cursor < 0 {
		cursor = 0
	}

	query := s.db.WithContext(ctx).Model(&model.Order{}).
		Where("driver_id = ? AND status IN ?", driverID,
			[]int8{
				constants.OrderStatusDispatched,
				constants.OrderStatusAccepted,
				constants.OrderStatusArrived,
				constants.OrderStatusInTrip,
				constants.OrderStatusCompleted,
				constants.OrderStatusCancelled,
			})

	if !isAll && date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, errors.NewBusinessError(1, "日期格式错误")
		}
		next := t.Add(24 * time.Hour)
		query = query.Where("created_at >= ? AND created_at < ?", t, next)
	} else if !isAll {
		today := time.Now().Format("2006-01-02")
		start, _ := time.Parse("2006-01-02", today)
		query = query.Where("created_at >= ? OR status IN ?", start,
			[]int8{
				constants.OrderStatusDispatched,
				constants.OrderStatusAccepted,
				constants.OrderStatusArrived,
				constants.OrderStatusInTrip,
			})
	}

	var orders []model.Order
	if err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(int(cursor)).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	return orders, nil
}

// GetOrder 查询订单详情（含行程节点、评价）
func (s *OrderService) GetOrder(ctx context.Context, orderID, driverID int64) (*model.Order, *model.TripService, *model.OrderEvaluation, error) {
	var order model.Order
	if err := s.db.WithContext(ctx).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		First(&order).Error; err != nil {
		return nil, nil, nil, errors.NewBusinessError(1, "订单不存在")
	}

	var trip model.TripService
	_ = s.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&trip).Error

	var evaluation model.OrderEvaluation
	_ = s.db.WithContext(ctx).
		Where("order_id = ? AND driver_id = ?", orderID, driverID).
		First(&evaluation).Error

	return &order, &trip, &evaluation, nil
}

func (s *OrderService) GetCurrentOrder(ctx context.Context, req *driver.GetCurrentOrderReq) (*driver.CurrentOrderResp, error) {
	order, err := s.repo.GetCurrentOrder(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return &driver.CurrentOrderResp{Success: true}, nil
	}
	return &driver.CurrentOrderResp{
		Success:        true,
		OrderId:        order.OrderId,
		OrderNo:        order.OrderNo,
		Status:         int32(order.Status),
		OriginAddress:  order.OriginAddress,
		OriginLat:      order.OriginLat,
		OriginLng:      order.OriginLng,
		DestAddress:    order.DestAddress,
		DestLat:        order.DestLat,
		DestLng:        order.DestLng,
		PassengerName:  order.PassengerName,
		PassengerMobile: order.PassengerMobile,
		EstimateFee:    order.EstimateFee,
	}, nil
}
