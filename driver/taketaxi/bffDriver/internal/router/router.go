package router

import (
	"driver/taketaxi/bffDriver/internal/handler"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewRouter(client *rpcclient.DriverClient, apiKey string, signKey string, aiCfg *config.AiConfig, dhCfg *config.DigitalHumanConfig, rdb *redis.Client, mongoDb *mongo.Database) *gin.Engine {
	r := gin.Default()
	h := handler.NewDriverHandler(client, apiKey, signKey)
	ai := handler.NewAiHandler(aiCfg, client, dhCfg, rdb, apiKey, signKey, mongoDb)
	r.GET("/api/v1/drivers", h.List)
	r.GET("/api/v1/drivers/:id", h.Get)
	r.POST("/api/v1/drivers", h.Create)
	r.PUT("/api/v1/drivers/:id", h.Update)
	r.DELETE("/api/v1/drivers/:id", h.Delete)

	// 司机上线/听单/下线/位置上报
	r.POST("/api/v1/drivers/:id/go-online", h.GoOnline)
	r.POST("/api/v1/drivers/:id/set-idle", h.SetIdle)
	r.POST("/api/v1/drivers/:id/start-listening", h.StartListening)
	r.POST("/api/v1/drivers/:id/go-offline", h.GoOffline)
	r.POST("/api/v1/drivers/:id/report-location", h.ReportLocation)

	// 派单/接单
	r.POST("/api/v1/orders/dispatch", h.DispatchOrder)
	r.POST("/api/v1/orders/:id/accept", h.AcceptOrder)
	r.POST("/api/v1/orders/:id/reject", h.RejectOrder)
	r.POST("/api/v1/orders/:id/cancel", h.CancelOrder)
	r.POST("/api/v1/orders/:id/arrive", h.DriverArrive)
	r.POST("/api/v1/orders/:id/verify-passenger", h.VerifyPassengerPhone)
	r.POST("/api/v1/orders/:id/start-trip", h.StartTrip)
	r.POST("/api/v1/orders/:id/end-trip", h.EndTrip)

	// 抢单池
	r.POST("/api/v1/pool/list", h.ListPoolOrders)
	r.POST("/api/v1/orders/:id/grab", h.GrabOrder)

	// 订单查询
	r.GET("/api/v1/drivers/:id/orders", h.ListOrders)
	r.GET("/api/v1/orders/:id/detail", h.GetOrder)

	// 路径规划
	r.POST("/api/v1/route/plan", h.RoutePlan)

	// AI
	r.POST("/api/v1/ai/chat", ai.Chat)
	r.POST("/api/v1/ai/voice-chat", ai.VoiceChat)
	r.GET("/api/v1/ai/tts", ai.TTS)
	return r
}
