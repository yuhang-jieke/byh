package repository

import (
	"context"
	"driver/taketaxi/common/constants"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/srvDriver/internal/model"

	"gorm.io/gorm"
)

type DriverRepo struct{ db *gorm.DB }

func NewDriverRepo(db *gorm.DB) *DriverRepo {
	if db == nil {
		db, _ = database.NewDB(nil)
	}
	return &DriverRepo{db: db}
}

func (r *DriverRepo) GetDB() *gorm.DB {
	return r.db
}

func (r *DriverRepo) Create(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *DriverRepo) GetByID(ctx context.Context, id uint) (*model.Driver, error) {
	var m model.Driver
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *DriverRepo) List(ctx context.Context) ([]*model.Driver, error) {
	var list []*model.Driver
	return list, r.db.WithContext(ctx).Find(&list).Error
}
func (r *DriverRepo) Update(ctx context.Context, m *model.Driver) error {
	return r.db.WithContext(ctx).Save(m).Error
}
func (r *DriverRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Driver{}, id).Error
}

// GetDriverSByDriverId 查询司机详情
func (r *DriverRepo) GetDriverSByDriverId(ctx context.Context, driverId int64) (*model.DriverS, error) {
	var driver model.DriverS
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverId).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

// GetRealnameByDriverId 查询实名认证
func (r *DriverRepo) GetRealnameByDriverId(ctx context.Context, driverId int64) (*model.DriverRealname, error) {
	var realname model.DriverRealname
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverId).Order("id desc").First(&realname).Error
	if err != nil {
		return nil, err
	}
	return &realname, nil
}

// GetVehicleByDriverId 查询车辆认证
func (r *DriverRepo) GetVehicleByDriverId(ctx context.Context, driverId int64) (*model.DriverVehicle, error) {
	var vehicle model.DriverVehicle
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverId).Order("id desc").First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// UpdateWorkStatus CAS 更新工作状态
func (r *DriverRepo) UpdateWorkStatus(ctx context.Context, driverId int64, fromStatus, toStatus int8) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.DriverS{}).
		Where("driver_id = ? AND work_status = ?", driverId, fromStatus).
		Update("work_status", toStatus)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// CreateStatusLog 写入状态变更日志
func (r *DriverRepo) CreateStatusLog(ctx context.Context, log *model.DriverStatusLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// CreateOnlineLog 写入出车/收车日志
func (r *DriverRepo) CreateOnlineLog(ctx context.Context, log *model.DriverOnlineLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// CreateDispatchLog 写入派单日志
func (r *DriverRepo) CreateDispatchLog(ctx context.Context, log *model.DispatchLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// HasOngoingOrder 判断司机是否有进行中订单
func (r *DriverRepo) HasOngoingOrder(ctx context.Context, driverId int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Order{}).
		Where("driver_id = ? AND status IN ?", driverId, []int8{constants.OrderStatusPending, constants.OrderStatusDispatched, constants.OrderStatusAccepted, constants.OrderStatusArrived, constants.OrderStatusInTrip}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetTodayOrderCount 查询司机今日已完成订单数
func (r *DriverRepo) GetTodayOrderCount(ctx context.Context, driverId int64, date string) (int, error) {
	var summary model.DriverStatisticsSummary
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND DATE(stat_date) = ?", driverId, date).
		First(&summary).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return summary.OrderCount, nil
}

// GetVehicleByDriverIdAndStatus 查询司机已通过认证的车辆
func (r *DriverRepo) GetVehicleByDriverIdAndStatus(ctx context.Context, driverId int64) (*model.DriverVehicle, error) {
	var vehicle model.DriverVehicle
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND status = ?", driverId, constants.AuthStatusApproved).
		Order("id desc").
		First(&vehicle).Error
	if err != nil {
		return nil, err
	}
	return &vehicle, nil
}

// GetTodayStats 查询司机今日统计汇总
func (r *DriverRepo) GetTodayStats(ctx context.Context, driverId int64, date string) (*model.DriverStatisticsSummary, error) {
	var summary model.DriverStatisticsSummary
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND DATE(stat_date) = ?", driverId, date).
		First(&summary).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &summary, nil
}

// GetCurrentOrder 查询司机当前进行中的订单（accepted/arrived/in_trip）
func (r *DriverRepo) GetCurrentOrder(ctx context.Context, driverId int64) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).
		Where("driver_id = ? AND status IN ?", driverId,
			[]int8{constants.OrderStatusAccepted, constants.OrderStatusArrived, constants.OrderStatusInTrip}).
		Order("created_at desc").
		First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// GetWallet 查询司机钱包
func (r *DriverRepo) GetWallet(ctx context.Context, driverId int64) (*model.DriverWallet, error) {
	var wallet model.DriverWallet
	err := r.db.WithContext(ctx).
		Where("driver_id = ?", driverId).
		First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}
