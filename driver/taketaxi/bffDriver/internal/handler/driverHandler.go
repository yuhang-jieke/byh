package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/bffDriver/internal/service"
	pb "driver/taketaxi/common/kitexGen/driver"

	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	client      *rpcclient.DriverClient
	routeSvc    *service.RouteService
}

func NewDriverHandler(client *rpcclient.DriverClient, apiKey string, signKey string) *DriverHandler {
	return &DriverHandler{
		client:   client,
		routeSvc: service.NewRouteService(apiKey, signKey),
	}
}

// RoutePlan 路径规划
func (h *DriverHandler) RoutePlan(c *gin.Context) {
	var req service.RoutePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证坐标有效性
	if req.OriginLat == 0 || req.OriginLng == 0 || req.DestLat == 0 || req.DestLng == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinates"})
		return
	}

	resp, err := h.routeSvc.PlanRoute(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": resp.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"distance": resp.Distance,
		"duration": resp.Duration,
		"polyline": resp.Polyline,
	})
}

func (h *DriverHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.Create(c.Request.Context(), &pb.CreateDriverReq{Name: req.Name})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": resp.Id})
}

func (h *DriverHandler) List(c *gin.Context) {
	resp, err := h.client.List(c.Request.Context(), &pb.ListDriverReq{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"drivers": resp.Items})
}

func (h *DriverHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.client.Get(c.Request.Context(), &pb.GetDriverReq{Id: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DriverHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.Update(c.Request.Context(), &pb.UpdateDriverReq{Id: id, Name: req.Name})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DriverHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.client.Delete(c.Request.Context(), &pb.DeleteDriverReq{Id: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GoOnline 司机出车上线
func (h *DriverHandler) GoOnline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.client.GoOnline(c.Request.Context(), &pb.GoOnlineReq{DriverId: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SetIdle 停止听单 (收车但不下线)
func (h *DriverHandler) SetIdle(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.client.SetIdle(c.Request.Context(), &pb.SetIdleReq{DriverId: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// StartListening 开始听单
func (h *DriverHandler) StartListening(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Lat float64 `json:"lat" binding:"required"`
		Lng float64 `json:"lng" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.StartListening(c.Request.Context(), &pb.StartListeningReq{
		DriverId: id, Lat: req.Lat, Lng: req.Lng,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GoOffline 收车下线
func (h *DriverHandler) GoOffline(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	resp, err := h.client.GoOffline(c.Request.Context(), &pb.GoOfflineReq{DriverId: id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReportLocation 位置上报
func (h *DriverHandler) ReportLocation(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Lat      float64 `json:"lat" binding:"required"`
		Lng      float64 `json:"lng" binding:"required"`
		Heading  float64 `json:"heading"`
		Speed    float64 `json:"speed"`
		Status   int32   `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.ReportLocation(c.Request.Context(), &pb.ReportLocationReq{
		DriverId: id, Lat: req.Lat, Lng: req.Lng,
		Heading: req.Heading, Speed: req.Speed, Status: req.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DispatchOrder 派单
func (h *DriverHandler) DispatchOrder(c *gin.Context) {
	var req struct {
		OrderID     int64   `json:"order_id" binding:"required"`
		OrderNo     string  `json:"order_no"`
		ServiceType int32   `json:"service_type" binding:"required"`
		OriginLat   float64 `json:"origin_lat" binding:"required"`
		OriginLng   float64 `json:"origin_lng" binding:"required"`
		PassengerID int64   `json:"passenger_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.DispatchOrder(c.Request.Context(), &pb.DispatchOrderReq{
		OrderId:     req.OrderID,
		OrderNo:     req.OrderNo,
		ServiceType: req.ServiceType,
		OriginLat:   req.OriginLat,
		OriginLng:   req.OriginLng,
		PassengerId: req.PassengerID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// AcceptOrder 司机接单
func (h *DriverHandler) AcceptOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	fmt.Printf("[BFF AcceptOrder] orderID=%d err=%v\n", orderID, err)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("[BFF AcceptOrder] calling gRPC: orderID=%d driverID=%d\n", orderID, req.DriverID)
	resp, err := h.client.AcceptOrder(c.Request.Context(), &pb.AcceptOrderReq{
		OrderId:  orderID,
		DriverId: req.DriverID,
	})
	fmt.Printf("[BFF AcceptOrder] gRPC resp: success=%v errCode=%v message=%v err=%v\n", resp.Success, resp.ErrCode, resp.Message, err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RejectOrder 司机拒绝接单
func (h *DriverHandler) RejectOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.RejectOrder(c.Request.Context(), &pb.RejectOrderReq{
		OrderId:  orderID,
		DriverId: req.DriverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CancelOrder 司机取消订单
func (h *DriverHandler) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID     int64  `json:"driver_id" binding:"required"`
		CancelReason string `json:"cancel_reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.CancelOrder(c.Request.Context(), &pb.CancelOrderReq{
		OrderId:      orderID,
		DriverId:     req.DriverID,
		CancelReason: req.CancelReason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DriverArrive 司机已到达
func (h *DriverHandler) DriverArrive(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.DriverArrive(c.Request.Context(), &pb.DriverArriveReq{
		OrderId:  orderID,
		DriverId: req.DriverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// VerifyPassengerPhone 验证乘客手机号后四位
func (h *DriverHandler) VerifyPassengerPhone(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID    int64  `json:"driver_id" binding:"required"`
		PhoneLast4 string `json:"phone_last4" binding:"required,len=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.VerifyPassengerPhone(c.Request.Context(), &pb.VerifyPassengerPhoneReq{
		OrderId:    orderID,
		DriverId:   req.DriverID,
		PhoneLast4: req.PhoneLast4,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// StartTrip 开始行程
func (h *DriverHandler) StartTrip(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.StartTrip(c.Request.Context(), &pb.StartTripReq{
		OrderId:  orderID,
		DriverId: req.DriverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// EndTrip 到达目的地
func (h *DriverHandler) EndTrip(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.EndTrip(c.Request.Context(), &pb.EndTripReq{
		OrderId:  orderID,
		DriverId: req.DriverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListPoolOrders 查看抢单池
func (h *DriverHandler) ListPoolOrders(c *gin.Context) {
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
		Page     int32 `json:"page"`
		PageSize int32 `json:"page_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}
	resp, err := h.client.ListPoolOrders(c.Request.Context(), &pb.ListPoolOrdersReq{
		DriverId: req.DriverID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GrabOrder 抢单
func (h *DriverHandler) GrabOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	var req struct {
		DriverID int64 `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.GrabOrder(c.Request.Context(), &pb.GrabOrderReq{
		OrderId:  orderID,
		DriverId: req.DriverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListOrders 司机订单列表
func (h *DriverHandler) ListOrders(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver id"})
		return
	}

	date := c.DefaultQuery("date", "")
	cursor, _ := strconv.ParseInt(c.DefaultQuery("cursor", "0"), 10, 32)
	isAll := c.Query("is_all") == "true"

	resp, err := h.client.ListOrders(c.Request.Context(), &pb.ListOrdersReq{
		DriverId: id,
		Date:     date,
		Cursor:   int32(cursor),
		IsAll:    isAll,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetOrder 订单详情
func (h *DriverHandler) GetOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	driverID, err := strconv.ParseInt(c.Query("driver_id"), 10, 64)
	if err != nil || driverID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id required"})
		return
	}
	resp, err := h.client.GetOrder(c.Request.Context(), &pb.GetOrderReq{
		OrderId:  orderID,
		DriverId: driverID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
