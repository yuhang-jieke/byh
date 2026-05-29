package handler

import (
	"context"
	"fmt"

	"driver/taketaxi/common/constants"
	"driver/taketaxi/common/errors"
	driver "driver/taketaxi/common/kitexGen/driver"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/srvDriver/internal/cache"
	"driver/taketaxi/srvDriver/internal/repository"
	"driver/taketaxi/srvDriver/internal/service"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type DriverHandler struct {
	svc         *service.DriverService
	onlineSvc   *service.OnlineService
	dispatchSvc *service.DispatchService
	poolSvc     *service.PoolService
	orderSvc    *service.OrderService
}

func NewDriverHandler(db *mongo.Database, repo *repository.DriverRepo, cfg *config.DispatchConfig, poolCache *cache.PoolCache) *DriverHandler {
	return &DriverHandler{
		svc:         service.NewDriverService(repo),
		onlineSvc:   service.NewOnlineService(repo, db),
		dispatchSvc: service.NewDispatchService(repo, db, cfg, poolCache),
		poolSvc:     service.NewPoolService(repo, db, poolCache),
		orderSvc:    service.NewOrderService(repo),
	}
}

func (h *DriverHandler) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	return h.svc.Create(ctx, req)
}
func (h *DriverHandler) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	return h.svc.Get(ctx, req)
}
func (h *DriverHandler) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	return h.svc.List(ctx, req)
}
func (h *DriverHandler) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	return h.svc.Update(ctx, req)
}
func (h *DriverHandler) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return h.svc.Delete(ctx, req)
}

func (h *DriverHandler) GoOnline(ctx context.Context, req *driver.GoOnlineReq) (*driver.GoOnlineResp, error) {
	err := h.onlineSvc.GoOnline(ctx, req.DriverId)
	if err != nil {
		return &driver.GoOnlineResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.GoOnlineResp{Success: true}, nil
}

func (h *DriverHandler) SetIdle(ctx context.Context, req *driver.SetIdleReq) (*driver.SetIdleResp, error) {
	err := h.onlineSvc.SetIdle(ctx, req.DriverId)
	if err != nil {
		return &driver.SetIdleResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.SetIdleResp{Success: true}, nil
}

func (h *DriverHandler) StartListening(ctx context.Context, req *driver.StartListeningReq) (*driver.StartListeningResp, error) {
	err := h.onlineSvc.StartListening(ctx, req.DriverId, req.Lat, req.Lng)
	if err != nil {
		return &driver.StartListeningResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.StartListeningResp{Success: true}, nil
}

func (h *DriverHandler) GoOffline(ctx context.Context, req *driver.GoOfflineReq) (*driver.GoOfflineResp, error) {
	err := h.onlineSvc.GoOffline(ctx, req.DriverId)
	if err != nil {
		return &driver.GoOfflineResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.GoOfflineResp{Success: true}, nil
}

func (h *DriverHandler) ReportLocation(ctx context.Context, req *driver.ReportLocationReq) (*driver.ReportLocationResp, error) {
	err := h.onlineSvc.ReportLocation(ctx, req.DriverId, req.Lat, req.Lng, req.Heading, req.Speed, int8(req.Status))
	if err != nil {
		return &driver.ReportLocationResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.ReportLocationResp{Success: true}, nil
}

func (h *DriverHandler) DispatchOrder(ctx context.Context, req *driver.DispatchOrderReq) (*driver.DispatchOrderResp, error) {
	err := h.dispatchSvc.Dispatch(ctx, req.OrderId, req.ServiceType, req.OriginLat, req.OriginLng, req.PassengerId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.DispatchOrderResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.DispatchOrderResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.DispatchOrderResp{Success: true}, nil
}

func (h *DriverHandler) AcceptOrder(ctx context.Context, req *driver.AcceptOrderReq) (*driver.AcceptOrderResp, error) {
	fmt.Printf("[Handler AcceptOrder] called: OrderId=%d DriverId=%d\n", req.OrderId, req.DriverId)
	err := h.dispatchSvc.AcceptOrder(ctx, req.OrderId, req.DriverId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.AcceptOrderResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.AcceptOrderResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.AcceptOrderResp{Success: true}, nil
}

func (h *DriverHandler) RejectOrder(ctx context.Context, req *driver.RejectOrderReq) (*driver.RejectOrderResp, error) {
	err := h.dispatchSvc.RejectOrder(ctx, req.OrderId, req.DriverId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.RejectOrderResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.RejectOrderResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.RejectOrderResp{Success: true}, nil
}

func (h *DriverHandler) CancelOrder(ctx context.Context, req *driver.CancelOrderReq) (*driver.CancelOrderResp, error) {
	err := h.dispatchSvc.CancelOrder(ctx, req.OrderId, req.DriverId, req.CancelReason)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.CancelOrderResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.CancelOrderResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.CancelOrderResp{Success: true}, nil
}

func (h *DriverHandler) DriverArrive(ctx context.Context, req *driver.DriverArriveReq) (*driver.DriverArriveResp, error) {
	err := h.dispatchSvc.DriverArrive(ctx, req.OrderId, req.DriverId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.DriverArriveResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.DriverArriveResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.DriverArriveResp{Success: true}, nil
}

func (h *DriverHandler) VerifyPassengerPhone(ctx context.Context, req *driver.VerifyPassengerPhoneReq) (*driver.VerifyPassengerPhoneResp, error) {
	err := h.dispatchSvc.VerifyPassengerPhone(ctx, req.OrderId, req.DriverId, req.PhoneLast4)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.VerifyPassengerPhoneResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.VerifyPassengerPhoneResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.VerifyPassengerPhoneResp{Success: true}, nil
}

func (h *DriverHandler) StartTrip(ctx context.Context, req *driver.StartTripReq) (*driver.StartTripResp, error) {
	err := h.dispatchSvc.StartTrip(ctx, req.OrderId, req.DriverId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.StartTripResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.StartTripResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.StartTripResp{Success: true}, nil
}

func (h *DriverHandler) EndTrip(ctx context.Context, req *driver.EndTripReq) (*driver.EndTripResp, error) {
	err := h.dispatchSvc.EndTrip(ctx, req.OrderId, req.DriverId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.EndTripResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.EndTripResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.EndTripResp{Success: true}, nil
}

// ListPoolOrders 查看抢单池
func (h *DriverHandler) ListPoolOrders(ctx context.Context, req *driver.ListPoolOrdersReq) (*driver.ListPoolOrdersResp, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > constants.PoolPageSizeMax {
		pageSize = constants.PoolPageSizeDefault
	}

	results, total, err := h.poolSvc.ListPoolOrders(ctx, req.DriverId, page, pageSize)
	if err != nil {
		return &driver.ListPoolOrdersResp{Success: false}, nil
	}

	pbItems := make([]*driver.PoolOrderItem, 0, len(results))
	for _, r := range results {
		o := r.Order
		pbItems = append(pbItems, &driver.PoolOrderItem{
			OrderId:          o.OrderId,
			OrderNo:          o.OrderNo,
			ServiceType:      int32(o.ServiceType),
			OriginLat:        o.OriginLat,
			OriginLng:        o.OriginLng,
			OriginAddress:    o.OriginAddress,
			DestLat:          o.DestLat,
			DestLng:          o.DestLng,
			DestAddress:      o.DestAddress,
			PassengerId:      o.PassengerId,
			PassengerName:    o.PassengerName,
			EstimateDistance: float64(o.EstimateDistance),
			EstimateFee:      o.EstimateFee,
			CreatedAt:        o.CreatedAt.Unix(),
			SecondsLeft:      r.SecondsLeft,
		})
	}

	return &driver.ListPoolOrdersResp{
		Success: true,
		Items:   pbItems,
		Total:   total,
	}, nil
}

// GrabOrder 抢单
func (h *DriverHandler) GrabOrder(ctx context.Context, req *driver.GrabOrderReq) (*driver.GrabOrderResp, error) {
	err := h.poolSvc.GrabOrder(ctx, req.OrderId, req.DriverId)
	if err != nil {
		if e, ok := err.(*errors.BusinessError); ok {
			return &driver.GrabOrderResp{Success: false, ErrCode: int32(e.Code), Message: e.Message}, nil
		}
		return &driver.GrabOrderResp{Success: false, Message: err.Error()}, nil
	}
	return &driver.GrabOrderResp{Success: true, OrderId: req.OrderId}, nil
}

func (h *DriverHandler) ListOrders(ctx context.Context, req *driver.ListOrdersReq) (*driver.ListOrdersResp, error) {
	orders, err := h.orderSvc.ListOrders(ctx, req.DriverId, req.Date, req.Cursor, req.IsAll)
	if err != nil {
		return &driver.ListOrdersResp{Success: false}, nil
	}

	items := make([]*driver.OrderItem, 0, len(orders))
	for _, o := range orders {
		item := &driver.OrderItem{
			OrderId:       o.OrderId,
			OrderNo:       o.OrderNo,
			ServiceType:   int32(o.ServiceType),
			OriginAddress: o.OriginAddress,
			DestAddress:   o.DestAddress,
			Status:        int32(o.Status),
			CreatedAt:     o.CreatedAt.Unix(),
			EstimateFee:   o.EstimateFee,
			DriverId:      o.DriverId,
			OriginLat:     o.OriginLat,
			OriginLng:     o.OriginLng,
			DestLat:       o.DestLat,
			DestLng:       o.DestLng,
		}
		if o.ActualDistance > 0 {
			item.DistanceKm = float64(o.ActualDistance) / 1000.0
		} else if o.EstimateDistance > 0 {
			item.DistanceKm = float64(o.EstimateDistance) / 1000.0
		}
		if o.ActualDuration > 0 {
			item.DurationMin = float64(o.ActualDuration) / 60.0
		} else if o.EstimateDuration > 0 {
			item.DurationMin = float64(o.EstimateDuration) / 60.0
		}
		items = append(items, item)
	}

	return &driver.ListOrdersResp{
		Success: true,
		Items:   items,
	}, nil
}

func (h *DriverHandler) GetOrder(ctx context.Context, req *driver.GetOrderReq) (*driver.GetOrderResp, error) {
	order, trip, evaluation, err := h.orderSvc.GetOrder(ctx, req.OrderId, req.DriverId)
	if err != nil {
		return &driver.GetOrderResp{Success: false, Message: err.Error()}, nil
	}

	resp := &driver.GetOrderResp{
		Success:         true,
		OrderNo:         order.OrderNo,
		OrderId:         order.OrderId,
		Status:          int32(order.Status),
		CreatedAt:       order.CreatedAt.Unix(),
		OriginAddress:   order.OriginAddress,
		DestAddress:     order.DestAddress,
		PassengerName:   order.PassengerName,
		PassengerMobile: order.PassengerMobile,
		TotalFee:        order.ActualFee,
		PayType:         int32(order.PayType),
		BaseFee:         order.BaseFee,
		DistanceFee:     order.DistanceFee,
		DurationFee:     order.DurationFee,
		WaitFee:         order.WaitFee,
		OriginLat:       order.OriginLat,
		OriginLng:       order.OriginLng,
		DestLat:         order.DestLat,
		DestLng:         order.DestLng,
	}

	if order.ActualDistance > 0 {
		resp.DistanceKm = float64(order.ActualDistance) / 1000.0
	}
	if order.ActualDuration > 0 {
		resp.DurationMin = float64(order.ActualDuration) / 60.0
	}

	if !trip.EndTime.IsZero() {
		resp.CompletedAt = trip.EndTime.Unix()
	}

	if evaluation.Id > 0 {
		resp.PassengerScore = int32(evaluation.PassengerScore)
		resp.PassengerComment = evaluation.PassengerComment
	}

	nodes := make([]*driver.OrderNode, 0)
	if !trip.AcceptTime.IsZero() {
		nodes = append(nodes, &driver.OrderNode{Name: "接单", Time: trip.AcceptTime.Unix()})
	}
	if !trip.ArriveTime.IsZero() {
		nodes = append(nodes, &driver.OrderNode{Name: "到达上车点", Time: trip.ArriveTime.Unix()})
	}
	if !trip.StartTime.IsZero() {
		nodes = append(nodes, &driver.OrderNode{Name: "开始行程", Time: trip.StartTime.Unix()})
	}
	if !trip.EndTime.IsZero() {
		nodes = append(nodes, &driver.OrderNode{Name: "到达目的地", Time: trip.EndTime.Unix()})
	}
	resp.Nodes = nodes

	return resp, nil
}

func (h *DriverHandler) GetTodayStats(ctx context.Context, req *driver.GetTodayStatsReq) (*driver.TodayStatsResp, error) {
	return h.svc.GetTodayStats(ctx, req)
}

func (h *DriverHandler) GetCurrentOrder(ctx context.Context, req *driver.GetCurrentOrderReq) (*driver.CurrentOrderResp, error) {
	return h.orderSvc.GetCurrentOrder(ctx, req)
}

func (h *DriverHandler) GetWallet(ctx context.Context, req *driver.GetWalletReq) (*driver.WalletResp, error) {
	return h.svc.GetWallet(ctx, req)
}
