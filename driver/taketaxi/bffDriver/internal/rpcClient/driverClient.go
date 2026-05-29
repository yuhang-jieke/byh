package rpcclient

import (
	"context"

	driver "driver/taketaxi/common/kitexGen/driver"
	"driver/taketaxi/common/kitexGen/driver/driverservice"

	"github.com/cloudwego/kitex/client"
)

type DriverClient struct {
	client driverservice.Client
}

func NewDriverClient(addr string) (*DriverClient, error) {
	cli, err := driverservice.NewClient("DriverService", client.WithHostPorts(addr))
	if err != nil {
		return nil, err
	}
	return &DriverClient{client: cli}, nil
}

func (c *DriverClient) Close() {}

func (c *DriverClient) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	return c.client.Create(ctx, req)
}

func (c *DriverClient) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	return c.client.Get(ctx, req)
}

func (c *DriverClient) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	return c.client.List(ctx, req)
}

func (c *DriverClient) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	return c.client.Update(ctx, req)
}

func (c *DriverClient) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return c.client.Delete(ctx, req)
}

func (c *DriverClient) GoOnline(ctx context.Context, req *driver.GoOnlineReq) (*driver.GoOnlineResp, error) {
	return c.client.GoOnline(ctx, req)
}

func (c *DriverClient) SetIdle(ctx context.Context, req *driver.SetIdleReq) (*driver.SetIdleResp, error) {
	return c.client.SetIdle(ctx, req)
}

func (c *DriverClient) StartListening(ctx context.Context, req *driver.StartListeningReq) (*driver.StartListeningResp, error) {
	return c.client.StartListening(ctx, req)
}

func (c *DriverClient) GoOffline(ctx context.Context, req *driver.GoOfflineReq) (*driver.GoOfflineResp, error) {
	return c.client.GoOffline(ctx, req)
}

func (c *DriverClient) ReportLocation(ctx context.Context, req *driver.ReportLocationReq) (*driver.ReportLocationResp, error) {
	return c.client.ReportLocation(ctx, req)
}

func (c *DriverClient) DispatchOrder(ctx context.Context, req *driver.DispatchOrderReq) (*driver.DispatchOrderResp, error) {
	return c.client.DispatchOrder(ctx, req)
}

func (c *DriverClient) AcceptOrder(ctx context.Context, req *driver.AcceptOrderReq) (*driver.AcceptOrderResp, error) {
	return c.client.AcceptOrder(ctx, req)
}

func (c *DriverClient) RejectOrder(ctx context.Context, req *driver.RejectOrderReq) (*driver.RejectOrderResp, error) {
	return c.client.RejectOrder(ctx, req)
}

func (c *DriverClient) CancelOrder(ctx context.Context, req *driver.CancelOrderReq) (*driver.CancelOrderResp, error) {
	return c.client.CancelOrder(ctx, req)
}

func (c *DriverClient) DriverArrive(ctx context.Context, req *driver.DriverArriveReq) (*driver.DriverArriveResp, error) {
	return c.client.DriverArrive(ctx, req)
}

func (c *DriverClient) VerifyPassengerPhone(ctx context.Context, req *driver.VerifyPassengerPhoneReq) (*driver.VerifyPassengerPhoneResp, error) {
	return c.client.VerifyPassengerPhone(ctx, req)
}

func (c *DriverClient) StartTrip(ctx context.Context, req *driver.StartTripReq) (*driver.StartTripResp, error) {
	return c.client.StartTrip(ctx, req)
}

func (c *DriverClient) EndTrip(ctx context.Context, req *driver.EndTripReq) (*driver.EndTripResp, error) {
	return c.client.EndTrip(ctx, req)
}

func (c *DriverClient) ListPoolOrders(ctx context.Context, req *driver.ListPoolOrdersReq) (*driver.ListPoolOrdersResp, error) {
	return c.client.ListPoolOrders(ctx, req)
}

func (c *DriverClient) GrabOrder(ctx context.Context, req *driver.GrabOrderReq) (*driver.GrabOrderResp, error) {
	return c.client.GrabOrder(ctx, req)
}

func (c *DriverClient) ListOrders(ctx context.Context, req *driver.ListOrdersReq) (*driver.ListOrdersResp, error) {
	return c.client.ListOrders(ctx, req)
}

func (c *DriverClient) GetOrder(ctx context.Context, req *driver.GetOrderReq) (*driver.GetOrderResp, error) {
	return c.client.GetOrder(ctx, req)
}

func (c *DriverClient) GetTodayStats(ctx context.Context, req *driver.GetTodayStatsReq) (*driver.TodayStatsResp, error) {
	return c.client.GetTodayStats(ctx, req)
}

func (c *DriverClient) GetCurrentOrder(ctx context.Context, req *driver.GetCurrentOrderReq) (*driver.CurrentOrderResp, error) {
	return c.client.GetCurrentOrder(ctx, req)
}

func (c *DriverClient) GetWallet(ctx context.Context, req *driver.GetWalletReq) (*driver.WalletResp, error) {
	return c.client.GetWallet(ctx, req)
}
