package service

import (
	"context"
	"time"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OnlineService struct {
	repo    *repository.DriverRepo
	mongoDb *mongo.Database
}

func NewOnlineService(repo *repository.DriverRepo, mongoDb *mongo.Database) *OnlineService {
	return &OnlineService{repo: repo, mongoDb: mongoDb}
}

// GoOnline 司机出车上线（离线 → 听单中，直接进入接单状态）
func (s *OnlineService) GoOnline(ctx context.Context, driverId int64) error {
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverId)
	if err != nil {
		return err
	}

	// 条件 5: 账号状态
	if driver.Status != constants.AccountStatusNormal {
		return errors.NewOnlineCheckError(errors.ErrCodeAccount, "您的账号状态异常，无法出车接单")
	}

	// 条件 6: 当前必须离线
	if int8(driver.WorkStatus) != constants.WorkStatusOffline {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "您已处于在线状态")
	}

	// 条件 2: 实名审核
	realname, err := s.repo.GetRealnameByDriverId(ctx, driverId)
	if err != nil || realname.Status != constants.AuthStatusApproved {
		return errors.NewOnlineCheckError(errors.ErrCodeRealname, "您的资质未审核通过，无法出车接单")
	}

	// 条件 3: 车辆认证
	vehicle, err := s.repo.GetVehicleByDriverId(ctx, driverId)
	if err != nil || vehicle.Status != constants.AuthStatusApproved {
		return errors.NewOnlineCheckError(errors.ErrCodeVehicle, "您的资质未审核通过，无法出车接单")
	}

	// CAS 更新：work_status 0 → 2 (直接进入听单状态)
	affected, err := s.repo.UpdateWorkStatus(ctx, driverId, constants.WorkStatusOffline, constants.WorkStatusListening)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "操作失败，您的状态已变更")
	}

	// 记录状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverId,
		FromStatus: constants.WorkStatusOffline,
		ToStatus:   constants.WorkStatusListening,
		Reason:     "司机出车上线 (进入听单)",
	})

	// 记录出车日志
	s.repo.CreateOnlineLog(ctx, &model.DriverOnlineLog{
		DriverId:   driverId,
		OnlineTime: time.Now(),
		CityId:     driver.CityId,
	})

	return nil
}

// SetIdle 停止听单（听单中 → 在线空闲）
func (s *OnlineService) SetIdle(ctx context.Context, driverId int64) error {
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverId)
	if err != nil {
		return err
	}

	// 必须是听单中状态才能停止听单
	if int8(driver.WorkStatus) != constants.WorkStatusListening {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "您当前不在听单状态")
	}

	// 检查是否有进行中订单
	hasOrder, err := s.repo.HasOngoingOrder(ctx, driverId)
	if err != nil {
		return err
	}
	if hasOrder {
		return errors.NewOnlineCheckError(errors.ErrCodeOngoing, "您当前有进行中的订单，无法停止听单")
	}

	// CAS 更新：work_status 2 → 1
	affected, err := s.repo.UpdateWorkStatus(ctx, driverId, constants.WorkStatusListening, constants.WorkStatusOnline)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "操作失败，您的状态已变更")
	}

	// 记录状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverId,
		FromStatus: constants.WorkStatusListening,
		ToStatus:   constants.WorkStatusOnline,
		Reason:     "停止听单 (保留在线)",
	})

	return nil
}

// StartListening 开始听单（在线 → 听单中）
func (s *OnlineService) StartListening(ctx context.Context, driverId int64, lat, lng float64) error {
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverId)
	if err != nil {
		return err
	}

	if int8(driver.WorkStatus) != constants.WorkStatusOnline {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "请先出车上线")
	}

	// CAS 更新：work_status 1 → 2
	affected, err := s.repo.UpdateWorkStatus(ctx, driverId, constants.WorkStatusOnline, constants.WorkStatusListening)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "操作失败，您的状态已变更")
	}

	// 写入初始位置到 MongoDB
	s.saveLocationToMongo(ctx, driverId, lat, lng, driver.CityId, 1)

	// 记录状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverId,
		FromStatus: constants.WorkStatusOnline,
		ToStatus:   constants.WorkStatusListening,
		Reason:     "开始听单",
	})

	return nil
}

// GoOffline 收车下线（在线/听单中 → 离线）
func (s *OnlineService) GoOffline(ctx context.Context, driverId int64) error {
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverId)
	if err != nil {
		return err
	}

	if int8(driver.WorkStatus) == constants.WorkStatusOffline {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "您已处于离线状态")
	}

	// 有进行中订单不允许下线
	hasOrder, err := s.repo.HasOngoingOrder(ctx, driverId)
	if err != nil {
		return err
	}
	if hasOrder {
		return errors.NewOnlineCheckError(errors.ErrCodeOngoing, "您当前有进行中的订单，无法收车")
	}

	fromStatus := int8(driver.WorkStatus)
	affected, err := s.repo.UpdateWorkStatus(ctx, driverId, fromStatus, constants.WorkStatusOffline)
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "操作失败，您的状态已变更")
	}

	// 记录状态日志
	s.repo.CreateStatusLog(ctx, &model.DriverStatusLog{
		DriverId:   driverId,
		FromStatus: fromStatus,
		ToStatus:   constants.WorkStatusOffline,
		Reason:     "司机收车下线",
	})

	// 删除 MongoDB 中的位置信息
	s.deleteLocationFromMongo(ctx, driverId)

	return nil
}

// ReportLocation 位置上报（在线状态均可）
func (s *OnlineService) ReportLocation(ctx context.Context, driverId int64, lat, lng, heading, speed float64, status int8) error {
	driver, err := s.repo.GetDriverSByDriverId(ctx, driverId)
	if err != nil {
		return err
	}

	if int8(driver.WorkStatus) <= constants.WorkStatusOffline {
		return errors.NewOnlineCheckError(errors.ErrCodeOnline, "请先出车后再上报位置")
	}

	s.saveLocationToMongo(ctx, driverId, lat, lng, driver.CityId, status)
	return nil
}

// saveLocationToMongo 保存司机位置到 MongoDB
func (s *OnlineService) saveLocationToMongo(ctx context.Context, driverId int64, lat, lng float64, cityId int64, status int8) {
	if s.mongoDb == nil {
		return
	}
	mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	collection := s.mongoDb.Collection("driver_local")
	now := time.Now().Unix()
	doc := bson.M{
		"driver_id":  driverId,
		"lat":        lat,
		"lng":        lng,
		"status":     status,
		"city_id":    cityId,
		"updated_at": now,
		"loc": bson.M{
			"type":        "Point",
			"coordinates": []float64{lng, lat},
		},
	}
	opts := options.Replace().SetUpsert(true)
	collection.ReplaceOne(mongoCtx, bson.M{"driver_id": driverId}, doc, opts)
}

// deleteLocationFromMongo 删除 MongoDB 中的司机位置
func (s *OnlineService) deleteLocationFromMongo(ctx context.Context, driverId int64) {
	if s.mongoDb == nil {
		return
	}
	mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	collection := s.mongoDb.Collection("driver_local")
	collection.DeleteOne(mongoCtx, bson.M{"driver_id": driverId})
}
