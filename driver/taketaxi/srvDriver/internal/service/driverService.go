package service

import (
	"context"
	driver "driver/taketaxi/common/kitexGen/driver"
	"driver/taketaxi/srvDriver/internal/model"
	"driver/taketaxi/srvDriver/internal/repository"
)

type DriverService struct{ repo *repository.DriverRepo }

func NewDriverService(repo *repository.DriverRepo) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) Create(ctx context.Context, req *driver.CreateDriverReq) (*driver.CreateDriverResp, error) {
	m := &model.Driver{Name: req.Name}
	return &driver.CreateDriverResp{Id: int64(m.ID)}, s.repo.Create(ctx, m)
}
func (s *DriverService) Get(ctx context.Context, req *driver.GetDriverReq) (*driver.GetDriverResp, error) {
	d, err := s.repo.GetDriverSByDriverId(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &driver.GetDriverResp{
		Id:           d.DriverId,
		Name:         d.Nickname,
		Mobile:       d.Mobile,
		Nickname:     d.Nickname,
		WorkStatus:   int32(d.WorkStatus),
		ServiceScore: d.ServiceScore,
		OrderCount:   int32(d.OrderCount),
		TotalIncome:  d.TotalIncome,
		Status:       int32(d.Status),
	}, nil
}
func (s *DriverService) List(ctx context.Context, req *driver.ListDriverReq) (*driver.ListDriverResp, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var items []*driver.DriverItem
	for _, m := range list {
		items = append(items, &driver.DriverItem{Id: int64(m.ID), Name: m.Name, Status: int32(m.Status)})
	}
	return &driver.ListDriverResp{Items: items}, nil
}
func (s *DriverService) Update(ctx context.Context, req *driver.UpdateDriverReq) (*driver.UpdateDriverResp, error) {
	m, err := s.repo.GetByID(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		m.Name = req.Name
	}
	return &driver.UpdateDriverResp{Success: true}, s.repo.Update(ctx, m)
}
func (s *DriverService) Delete(ctx context.Context, req *driver.DeleteDriverReq) (*driver.DeleteDriverResp, error) {
	return &driver.DeleteDriverResp{Success: true}, s.repo.Delete(ctx, uint(req.Id))
}

func (s *DriverService) GetTodayStats(ctx context.Context, req *driver.GetTodayStatsReq) (*driver.TodayStatsResp, error) {
	summary, err := s.repo.GetTodayStats(ctx, req.DriverId, req.Date)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return &driver.TodayStatsResp{Success: true}, nil
	}
	return &driver.TodayStatsResp{
		Success:        true,
		CompletedOrders: int32(summary.CompleteCount),
		TotalEarnings:  summary.TotalIncome,
		OnlineSeconds:  int32(summary.OnlineDuration),
		TotalKm:        float64(summary.TotalDistance) / 1000,
	}, nil
}

func (s *DriverService) GetWallet(ctx context.Context, req *driver.GetWalletReq) (*driver.WalletResp, error) {
	wallet, err := s.repo.GetWallet(ctx, req.DriverId)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return &driver.WalletResp{Success: true}, nil
	}
	return &driver.WalletResp{
		Success:     true,
		Balance:     wallet.Balance,
		TotalIncome: wallet.TotalIncome,
		FrozenAmount: wallet.FrozenAmount,
	}, nil
}
