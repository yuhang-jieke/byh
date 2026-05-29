package handler

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/common/constants"
	pb "driver/taketaxi/common/kitexGen/driver"
	"driver/taketaxi/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AiHandler struct {
	cfg          *config.AiConfig
	client       *rpcclient.DriverClient
	dhCfg        *config.DigitalHumanConfig
	rdb          *redis.Client
	amapKey      string
	amapSignKey  string
	mongoDb      *mongo.Database
	baiduToken   string
	baiduTokenAt time.Time
	baiduMu      sync.Mutex
	httpClient   *http.Client
}

func NewAiHandler(cfg *config.AiConfig, client *rpcclient.DriverClient, dhCfg *config.DigitalHumanConfig, rdb *redis.Client, amapKey, amapSignKey string, mongoDb *mongo.Database) *AiHandler {
	return &AiHandler{
		cfg:         cfg,
		client:      client,
		dhCfg:       dhCfg,
		rdb:         rdb,
		amapKey:     amapKey,
		amapSignKey: amapSignKey,
		mongoDb:     mongoDb,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

type chatRequest struct {
	Message  string `json:"message"`
	DriverID int64  `json:"driver_id"`
}

type deepseekResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type weatherData struct {
	City          string
	Temp          string
	Weather       string
	WindDirection string
	WindPower     string
	Humidity      string
}

type lastIntentInfo struct {
	IntentName string `json:"intent_name"`
	API        string `json:"api"`
	UserText   string `json:"user_text"`
}

var triggerCache = map[string]*regexp.Regexp{}

func matchTrigger(text string, pattern string) bool {
	if strings.HasPrefix(pattern, "^") || strings.HasSuffix(pattern, "$") {
		re, ok := triggerCache[pattern]
		if !ok {
			re = regexp.MustCompile(pattern)
			triggerCache[pattern] = re
		}
		return re.MatchString(text)
	}

	if containsRegexMeta(pattern) {
		re, ok := triggerCache[pattern]
		if !ok {
			var err error
			re, err = regexp.Compile(pattern)
			if err != nil {
				return strings.Contains(text, pattern)
			}
			triggerCache[pattern] = re
		}
		return re.MatchString(text)
	}

	return strings.Contains(text, pattern)
}

func containsRegexMeta(s string) bool {
	for _, mc := range []string{".", "*", "+", "?", "|", "(", ")", "[", "]", "{", "}", "\\"} {
		if strings.Contains(s, mc) {
			return true
		}
	}
	return false
}

func (h *AiHandler) dispatchAPI(c *gin.Context, apiName string, driverID int64, userText string) string {
	ctx := c.Request.Context()
	today := time.Now().Format("2006-01-02")

	switch apiName {
	case "GetTodayStats":
		resp, err := h.client.GetTodayStats(ctx, &pb.GetTodayStatsReq{DriverId: driverID, Date: today})
		if err != nil {
			return "查询失败，请稍后再试"
		}
		if resp.CompletedOrders == 0 {
			return "你今天还没有完成订单，加油出车接单吧！"
		}
		return fmt.Sprintf("你今天接了 %d 单，收入 %.2f 元。", resp.CompletedOrders, resp.TotalEarnings)

	case "GetWallet":
		resp, err := h.client.GetWallet(ctx, &pb.GetWalletReq{DriverId: driverID})
		if err != nil {
			return "查询失败，请稍后再试"
		}
		return fmt.Sprintf("你的钱包余额 %.2f 元，累计收入 %.2f 元。", resp.Balance, resp.TotalIncome)

	case "GetDriverStatus":
		resp, err := h.client.Get(ctx, &pb.GetDriverReq{Id: driverID})
		if err != nil {
			return "查询失败，请稍后再试"
		}
		desc := h.getStatusDescription(c, driverID, resp.WorkStatus)
		return fmt.Sprintf("你目前%s。", desc)

	case "GetCurrentOrder":
		resp, err := h.client.GetCurrentOrder(ctx, &pb.GetCurrentOrderReq{DriverId: driverID})
		if err != nil || resp == nil || !resp.Success || resp.OrderId == 0 {
			return "你目前没有进行中的订单。"
		}
		return fmt.Sprintf("当前订单：从 %s 到 %s，乘客 %s，预计费用 %.2f 元。",
			resp.OriginAddress, resp.DestAddress, resp.PassengerName, resp.EstimateFee)

	case "GetServiceScore":
		resp, err := h.client.Get(ctx, &pb.GetDriverReq{Id: driverID})
		if err != nil {
			return "查询失败，请稍后再试"
		}
		return fmt.Sprintf("你的服务评分是 %.1f 分。", resp.ServiceScore)

	case "ListOrders":
		resp, err := h.client.ListOrders(ctx, &pb.ListOrdersReq{DriverId: driverID, IsAll: true})
		if err != nil || resp == nil || !resp.Success {
			return "查询失败，请稍后再试"
		}
		for _, item := range resp.Items {
			if item.Status == int32(constants.OrderStatusCompleted) {
				return fmt.Sprintf("上一单：从%s到%s，预估%.0f元，约%.1f公里。", item.OriginAddress, item.DestAddress, item.EstimateFee, item.DistanceKm)
			}
		}
		return "暂时没有历史订单记录。"

	case "ListOrdersByDate":
		date := extractDate(userText)
		if date == "" {
			return "请告诉我具体日期，比如「昨天的订单」或「5月2号的订单」"
		}
		resp, err := h.client.ListOrders(ctx, &pb.ListOrdersReq{DriverId: driverID, Date: date})
		if err != nil || resp == nil || !resp.Success {
			return "查询失败，请稍后再试"
		}
		var completed, total int
		for _, item := range resp.Items {
			total++
			if item.Status == int32(constants.OrderStatusCompleted) {
				completed++
			}
		}
		if total == 0 {
			return date + "没有订单记录。"
		}
		return fmt.Sprintf("%s共接了%d单，完成了%d单", date, total, completed)

	case "ListPoolOrders":
		resp, err := h.client.ListPoolOrders(ctx, &pb.ListPoolOrdersReq{DriverId: driverID, Page: 1, PageSize: 30})
		if err != nil || resp == nil || !resp.Success {
			return "查询失败，请稍后再试"
		}

		ongoing, _ := h.client.GetCurrentOrder(ctx, &pb.GetCurrentOrderReq{DriverId: driverID})
		if ongoing != nil && ongoing.Success && ongoing.OrderId > 0 {
			return "您当前有进行中的订单，先完成当前订单再查看抢单池吧。"
		}

		var driverLat, driverLng float64
		if h.mongoDb != nil {
			mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			var loc struct {
				Lat float64 `bson:"lat"`
				Lng float64 `bson:"lng"`
			}
			if err := h.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err == nil && loc.Lat != 0 {
				driverLat = loc.Lat
				driverLng = loc.Lng
			}
		}

		type orderWithDist struct {
			item   *pb.PoolOrderItem
			distKm float64
		}

		var filtered []orderWithDist
		if driverLat != 0 {
			type calc struct {
				item   *pb.PoolOrderItem
				distKm float64
			}
			var calcs []calc
			for _, item := range resp.Items {
				d := -1.0
				if item.OriginLat != 0 {
					d = haversine(driverLat, driverLng, item.OriginLat, item.OriginLng)
				}
				calcs = append(calcs, calc{item: item, distKm: d})
			}

			mult := h.getPoolRadiusMultiplier(ctx, driverID)
			for _, radius := range []float64{8 * mult, 15 * mult, 30 * mult} {
				for _, c := range calcs {
					if c.distKm <= radius || c.distKm < 0 {
						filtered = append(filtered, orderWithDist{item: c.item, distKm: c.distKm})
					}
				}
				if len(filtered) > 0 {
					break
				}
				filtered = nil
			}
		} else {
			for _, item := range resp.Items {
				filtered = append(filtered, orderWithDist{item: item, distKm: 0})
				if len(filtered) >= 5 {
					break
				}
			}
		}

		if len(filtered) > 5 {
			filtered = filtered[:5]
		}

		if len(filtered) == 0 {
			if h.getPoolRadiusMultiplier(ctx, driverID) > 1 {
				h.resetPoolRadius(ctx, driverID)
				return "扩大范围后也没有找到更多订单，已恢复默认搜索范围。"
			}
			return "附近暂时没有可抢订单，稍后再看看吧。"
		}

		var poolParts []string
		for i, od := range filtered {
			distStr := ""
			if driverLat != 0 && od.distKm > 0 {
				if od.distKm < 1.0 {
					distStr = fmt.Sprintf("，距您%.0f米", od.distKm*1000)
				} else {
					distStr = fmt.Sprintf("，距您%.1fkm", od.distKm)
				}
			}
			poolParts = append(poolParts, fmt.Sprintf("%d. %s→%s，预估%.0f元，全程%.1fkm%s",
				i+1, od.item.OriginAddress, od.item.DestAddress, od.item.EstimateFee, od.item.EstimateDistance/1000, distStr))
		}
		return fmt.Sprintf("附近有%d个订单：%s。您想抢第几个？", len(filtered), strings.Join(poolParts, "；"))

	case "GrabPoolOrder":
		poolResp, err := h.client.ListPoolOrders(ctx, &pb.ListPoolOrdersReq{DriverId: driverID, Page: 1, PageSize: 30})
		if err != nil || poolResp == nil || !poolResp.Success || len(poolResp.Items) == 0 {
			return "抢单池暂时没有订单。"
		}

		ongoing, _ := h.client.GetCurrentOrder(ctx, &pb.GetCurrentOrderReq{DriverId: driverID})
		if ongoing != nil && ongoing.Success && ongoing.OrderId > 0 {
			return "您当前有进行中的订单，不能同时抢多个单，先完成当前订单吧。"
		}

		var grabDriverLat, grabDriverLng float64
		if h.mongoDb != nil {
			mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			var loc struct {
				Lat float64 `bson:"lat"`
				Lng float64 `bson:"lng"`
			}
			if err := h.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err == nil && loc.Lat != 0 {
				grabDriverLat = loc.Lat
				grabDriverLng = loc.Lng
			}
		}

		type grabItem struct {
			item   *pb.PoolOrderItem
			distKm float64
		}
		var grabFiltered []grabItem
		if grabDriverLat != 0 {
			type grabCalc struct {
				item   *pb.PoolOrderItem
				distKm float64
			}
			var grabCalcs []grabCalc
			for _, item := range poolResp.Items {
				d := -1.0
				if item.OriginLat != 0 {
					d = haversine(grabDriverLat, grabDriverLng, item.OriginLat, item.OriginLng)
				}
				grabCalcs = append(grabCalcs, grabCalc{item: item, distKm: d})
			}
			mult := h.getPoolRadiusMultiplier(ctx, driverID)
			for _, radius := range []float64{8 * mult, 15 * mult, 30 * mult} {
				for _, c := range grabCalcs {
					if c.distKm <= radius || c.distKm < 0 {
						grabFiltered = append(grabFiltered, grabItem{item: c.item, distKm: c.distKm})
					}
				}
				if len(grabFiltered) > 0 {
					break
				}
				grabFiltered = nil
			}
		} else {
			for _, item := range poolResp.Items {
				grabFiltered = append(grabFiltered, grabItem{item: item, distKm: 0})
				if len(grabFiltered) >= 5 {
					break
				}
			}
		}

		if len(grabFiltered) > 5 {
			grabFiltered = grabFiltered[:5]
		}

		if len(grabFiltered) == 0 {
			return "附近暂时没有可抢订单。"
		}

		idx, ok := parseGrabIndex(userText)
		if !ok || idx < 1 || idx > len(grabFiltered) {
			var grabParts []string
			for i, g := range grabFiltered {
				gDistStr := ""
				if grabDriverLat != 0 && g.distKm > 0 {
					if g.distKm < 1.0 {
						gDistStr = fmt.Sprintf("，距您%.0f米", g.distKm*1000)
					} else {
						gDistStr = fmt.Sprintf("，距您%.1fkm", g.distKm)
					}
				}
				grabParts = append(grabParts, fmt.Sprintf("%d. %s→%s，预估%.0f元，全程%.1fkm%s",
					i+1, g.item.OriginAddress, g.item.DestAddress, g.item.EstimateFee, g.item.EstimateDistance/1000, gDistStr))
			}
			return fmt.Sprintf("附近有%d个订单：%s。您想抢第几个？", len(grabParts), strings.Join(grabParts, "；"))
		}
		order := grabFiltered[idx-1].item
		grabResp, err := h.client.GrabOrder(ctx, &pb.GrabOrderReq{OrderId: order.OrderId, DriverId: driverID})
		if err != nil {
			return "抢单失败，请稍后再试"
		}
		if grabResp.Success {
			return fmt.Sprintf("抢到了！从%s到%s，预估%.0f元，出发吧！", order.OriginAddress, order.DestAddress, order.EstimateFee)
		}
		return fmt.Sprintf("没抢到%s到%s的单，被别人抢先了，看看其他的吧。", order.OriginAddress, order.DestAddress)

	case "FindHotAreas":
		return h.findHotAreas(c, driverID, userText)

	case "GoOnline":
		resp, err := h.client.GoOnline(ctx, &pb.GoOnlineReq{DriverId: driverID})
		if err != nil {
			return "操作失败，请稍后再试"
		}
		if resp.Success {
			return "已出车，开始听单接单！"
		}
		return resp.Message

	case "GoOffline":
		resp, err := h.client.GoOffline(ctx, &pb.GoOfflineReq{DriverId: driverID})
		if err != nil {
			return "操作失败，请稍后再试"
		}
		if resp.Success {
			return "已收车下线，辛苦了！"
		}
		return resp.Message

	case "StartListening":
		resp, err := h.client.StartListening(ctx, &pb.StartListeningReq{DriverId: driverID, Lat: 0, Lng: 0})
		if err != nil {
			return "操作失败，请稍后再试"
		}
		if resp.Success {
			return "已开始听单，有订单会立即通知你！"
		}
		return resp.Message

	case "DriverArrive":
		return "请先在订单页操作到达上车点。"

	case "StartTrip":
		return "请先在订单页操作开始行程。"

	case "EndTrip":
		return "请先在订单页操作结束行程。"

	case "CancelOrder":
		return "请在订单页操作取消订单。"

	case "QueryWeather":
		return h.fetchWeather(c, driverID, userText)

	case "QueryNearbyToilet":
		return h.queryNearbyPOI(c, driverID, "公共厕所|公厕", "附近没找到公共厕所，可以去加油站或商场看看")
	case "QueryNearbyGas":
		return h.queryNearbyPOI(c, driverID, "加油站", "附近没找到加油站")
	case "QueryNearbyCharger":
		return h.queryNearbyPOI(c, driverID, "充电桩", "附近没找到充电桩")
	case "QueryNearbyFood":
		return h.queryNearbyPOI(c, driverID, "餐厅|快餐|美食", "附近没找到餐厅")
	case "QueryNearbyParking":
		return h.queryNearbyPOI(c, driverID, "停车场", "附近没找到停车场")
	case "QueryNearbyMarket":
		return h.queryNearbyPOI(c, driverID, "超市|便利店", "附近没找到超市")

	default:
		return ""
	}
}

// ---------- Chat ----------

func (h *AiHandler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	reply := h.chatProcess(c, req.DriverID, strings.TrimSpace(req.Message))
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}

// VoiceChat 语音对话接口，返回音频
func (h *AiHandler) VoiceChat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	reply := h.chatProcess(c, req.DriverID, strings.TrimSpace(req.Message))
	audio, err := h.textToSpeech(reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tts failed"})
		return
	}
	c.Data(http.StatusOK, "audio/mp3", audio)
}

func (h *AiHandler) chatProcess(c *gin.Context, driverID int64, text string) string {
	var reply string
	intentMatched := false
	var matchedIntent *config.DigitalHumanIntent

	if h.dhCfg != nil && driverID > 0 {
		for _, intent := range h.dhCfg.Intents {
			matched := false
			for _, trigger := range intent.Triggers {
				if matchTrigger(text, trigger) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			switch intent.Action {
			case "say":
				h.resetPoolRadius(c.Request.Context(), driverID)
				reply = intent.Reply
				intentMatched = true
			case "call_api":
				if intent.API != "ListPoolOrders" {
					h.resetPoolRadius(c.Request.Context(), driverID)
				}
				reply = h.dispatchAPI(c, intent.API, driverID, text)
				if reply != "" {
					intentMatched = true
				}
			case "llm_chat":
			}
			if intentMatched {
				matchedIntent = &intent
			}
			break
		}
	}

	if !intentMatched && driverID > 0 {
		reply = h.inferFromContext(c, &intentMatched, driverID, text)
	}

	if !intentMatched {
		history := h.loadHistory(c.Request.Context(), driverID)
		driverCtx := h.buildDriverContext(c, driverID, text)

		sysPrompt := "你是一个开网约车司机的AI助手，名叫花小猪助手，说话简洁亲切，语气像朋友一样自然。回答限制在80字以内。"
		if h.dhCfg != nil && h.dhCfg.Name != "" {
			sysPrompt = h.dhCfg.BuildSystemPrompt()
		}
		if driverCtx != "" {
			sysPrompt = sysPrompt + "。" + driverCtx
		}
		sysPrompt += "【注意】我给你的实时数据优先于对话历史，如果历史消息中的数据与最新实时数据不一致，以我给你的实时数据为准。"
		sysPrompt += "你只能引用当前对话中我明确给你的数据（包括当前司机信息和对话历史中的API返回结果），不得自己编造任何订单明细、金额、评分、操作结果等数据。你没有调用任何接口的能力，用户要求操作时请引导用户使用「帮我抢」「出车」等指定话术。"

		messages := make([]map[string]string, 0, len(history)+2)
		messages = append(messages, map[string]string{"role": "system", "content": sysPrompt})
		messages = append(messages, history...)
		messages = append(messages, map[string]string{"role": "user", "content": text})

		body, _ := json.Marshal(map[string]any{
			"model":    h.cfg.LlmModel,
			"messages": messages,
			"stream":   false,
		})

		httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+h.cfg.LlmApiKey)

		resp, err := h.httpClient.Do(httpReq)
		if err == nil {
			defer resp.Body.Close()
			var dsResp deepseekResp
			if err := json.NewDecoder(resp.Body).Decode(&dsResp); err == nil && len(dsResp.Choices) > 0 {
				reply = dsResp.Choices[0].Message.Content
			}
		}
		if reply == "" {
			reply = "这个我暂时查不到，您说点别的吧。"
		}
	}

	if driverID > 0 && reply != "" {
		h.saveConversation(c.Request.Context(), driverID, text, reply)
	}

	h.saveLastIntent(c.Request.Context(), driverID, text, matchedIntent)

	return reply
}

// ---------- Redis 对话上下文 ----------

const (
	dhCtxKeyPrefix         = "dh:ctx:%d"
	dhCtxMaxLen            = 20
	dhCtxTTL               = 1 * time.Hour
	dhLastIntentKeyPrefix  = "dh:last:intent:%d"
	dhPoolRadiusKeyPrefix  = "dh:pool:radius:%d"
	dhPoolRadiusMax        = 8.0
)

func (h *AiHandler) loadHistory(ctx context.Context, driverID int64) []map[string]string {
	if h.rdb == nil || driverID <= 0 {
		return nil
	}
	key := fmt.Sprintf(dhCtxKeyPrefix, driverID)
	entries, err := h.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil || len(entries) == 0 {
		return nil
	}
	result := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		var msg map[string]string
		if err := json.Unmarshal([]byte(entry), &msg); err != nil {
			continue
		}
		if msg["role"] == "" || msg["content"] == "" {
			continue
		}
		result = append(result, msg)
	}
	return result
}

func (h *AiHandler) saveConversation(ctx context.Context, driverID int64, userMsg, reply string) {
	if h.rdb == nil || driverID <= 0 {
		return
	}
	key := fmt.Sprintf(dhCtxKeyPrefix, driverID)

	userEntry, _ := json.Marshal(map[string]string{
		"role":    "user",
		"content": userMsg,
	})
	replyEntry, _ := json.Marshal(map[string]string{
		"role":    "assistant",
		"content": reply,
	})

	pipe := h.rdb.Pipeline()
	pipe.RPush(ctx, key, string(userEntry), string(replyEntry))
	pipe.LTrim(ctx, key, int64(-dhCtxMaxLen), -1)
	pipe.Expire(ctx, key, dhCtxTTL)
	pipe.Exec(ctx)
}

func (h *AiHandler) saveLastIntent(ctx context.Context, driverID int64, userText string, matchedIntent *config.DigitalHumanIntent) {
	if h.rdb == nil || driverID <= 0 {
		return
	}
	if matchedIntent == nil {
		return
	}
	info := lastIntentInfo{
		IntentName: matchedIntent.Name,
		API:        matchedIntent.API,
		UserText:   userText,
	}
	data, _ := json.Marshal(info)
	key := fmt.Sprintf(dhLastIntentKeyPrefix, driverID)
	h.rdb.Set(ctx, key, data, dhCtxTTL)
}

// ---------- 抢单池半径倍数 ----------

func (h *AiHandler) getPoolRadiusMultiplier(ctx context.Context, driverID int64) float64 {
	if h.rdb == nil || driverID <= 0 {
		return 1
	}
	key := fmt.Sprintf(dhPoolRadiusKeyPrefix, driverID)
	val, err := h.rdb.Get(ctx, key).Float64()
	if err != nil || val <= 0 {
		return 1
	}
	if val > dhPoolRadiusMax {
		return dhPoolRadiusMax
	}
	return val
}

func (h *AiHandler) expandPoolRadius(ctx context.Context, driverID int64) {
	if h.rdb == nil || driverID <= 0 {
		return
	}
	key := fmt.Sprintf(dhPoolRadiusKeyPrefix, driverID)
	cur := h.getPoolRadiusMultiplier(ctx, driverID)
	next := cur * 2
	if next > dhPoolRadiusMax {
		next = dhPoolRadiusMax
	}
	h.rdb.Set(ctx, key, next, dhCtxTTL)
}

func (h *AiHandler) resetPoolRadius(ctx context.Context, driverID int64) {
	if h.rdb == nil || driverID <= 0 {
		return
	}
	key := fmt.Sprintf(dhPoolRadiusKeyPrefix, driverID)
	h.rdb.Del(ctx, key)
}

// ---------- 上下文推断 ----------

func (h *AiHandler) loadLastIntent(ctx context.Context, driverID int64) *lastIntentInfo {
	if h.rdb == nil || driverID <= 0 {
		return nil
	}
	key := fmt.Sprintf(dhLastIntentKeyPrefix, driverID)
	data, err := h.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var info lastIntentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil
	}
	return &info
}

func (h *AiHandler) inferFromContext(c *gin.Context, intentMatched *bool, driverID int64, text string) string {
	last := h.loadLastIntent(c.Request.Context(), driverID)
	if last == nil {
		return ""
	}

	if last.API == "QueryWeather" {
		city := extractCity(text)
		if city != "" {
			*intentMatched = true
			return h.fetchWeather(c, driverID, text)
		}
		if (strings.Contains(text, "那") || strings.Contains(text, "这里")) &&
			(strings.Contains(text, "天气") || strings.Contains(text, "天") || strings.Contains(text, "下")) {
			*intentMatched = true
			return h.fetchWeather(c, driverID, text)
		}
	}

	if last.API == "GetTodayStats" {
		date := extractDate(text)
		if date != "" {
			*intentMatched = true
			return h.dispatchAPI(c, "ListOrdersByDate", driverID, text)
		}
		if strings.Contains(text, "那") &&
			(strings.Contains(text, "天") || strings.Contains(text, "周") || strings.Contains(text, "月")) {
			date := extractDate(text)
			if date != "" {
				*intentMatched = true
				return h.dispatchAPI(c, "ListOrdersByDate", driverID, text)
			}
		}
	}

	if last.API == "ListPoolOrders" {
		_, ok := parseGrabIndex(text)
		if ok {
			*intentMatched = true
			return h.dispatchAPI(c, "GrabPoolOrder", driverID, text)
		}
		if (strings.Contains(text, "抢") && !strings.Contains(text, "不")) || strings.Contains(text, "来") {
			*intentMatched = true
			return h.dispatchAPI(c, "GrabPoolOrder", driverID, text)
		}
		if strings.Contains(text, "还有") || strings.Contains(text, "再") || strings.Contains(text, "刷") || strings.Contains(text, "新单") || strings.Contains(text, "远") {
			h.expandPoolRadius(c.Request.Context(), driverID)
			*intentMatched = true
			return h.dispatchAPI(c, "ListPoolOrders", driverID, text)
		}
		if strings.Contains(text, "还原") || strings.Contains(text, "重置") || strings.Contains(text, "默认") {
			h.resetPoolRadius(c.Request.Context(), driverID)
			*intentMatched = true
			return h.dispatchAPI(c, "ListPoolOrders", driverID, text)
		}
	}

	if last.API == "FindHotAreas" {
		if strings.Contains(text, "其他") {
			*intentMatched = true
			return h.findHotAreas(c, driverID, text)
		}
	}

	if last.API == "GrabPoolOrder" {
		if strings.Contains(text, "第") || strings.Contains(text, "确认") || strings.Contains(text, "抢") || strings.Contains(text, "接") {
			*intentMatched = true
			return h.dispatchAPI(c, "GrabPoolOrder", driverID, text)
		}
	}

	return ""
}

// ---------- 司机实时上下文（RPC 查 MySQL） ----------

func (h *AiHandler) buildDriverContext(c *gin.Context, driverID int64, userText string) string {
	if driverID <= 0 {
		return ""
	}
	ctx := c.Request.Context()
	var parts []string

	driverResp, err := h.client.Get(ctx, &pb.GetDriverReq{Id: driverID})
	if err == nil && driverResp != nil {
		desc := h.getStatusDescription(c, driverID, driverResp.WorkStatus)
		parts = append(parts, desc)
		if driverResp.ServiceScore > 0 {
			parts = append(parts, fmt.Sprintf("服务评分%.1f分", driverResp.ServiceScore))
		}
	}

	today := time.Now().Format("2006-01-02")
	statsResp, err := h.client.GetTodayStats(ctx, &pb.GetTodayStatsReq{DriverId: driverID, Date: today})
	if err == nil && statsResp != nil && statsResp.Success && statsResp.CompletedOrders > 0 {
		parts = append(parts, fmt.Sprintf("今天完成了%d单，收入%.2f元", statsResp.CompletedOrders, statsResp.TotalEarnings))
	}

	orderResp, err := h.client.GetCurrentOrder(ctx, &pb.GetCurrentOrderReq{DriverId: driverID})
	if err == nil && orderResp != nil && orderResp.Success && orderResp.OrderId > 0 {
		parts = append(parts, fmt.Sprintf("当前有进行中的订单：从%s到%s，预估价%.0f元", orderResp.OriginAddress, orderResp.DestAddress, orderResp.EstimateFee))
	}

	if len(parts) == 0 {
		return ""
	}
	return "当前司机信息：" + strings.Join(parts, "，") + "。"
}

// ---------- 天气查询 ----------

func (h *AiHandler) fetchWeather(c *gin.Context, driverID int64, userText string) string {
	city := extractCity(userText)

	if city == "" && h.mongoDb != nil && driverID > 0 {
		mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var loc struct {
			Lat float64 `bson:"lat"`
			Lng float64 `bson:"lng"`
		}
		collection := h.mongoDb.Collection("driver_local")
		if err := collection.FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err == nil && loc.Lat != 0 {
			city = h.regeoCity(loc.Lat, loc.Lng)
		}
	}

	if city == "" {
		return "无法获取您的位置，暂时查不到天气信息。你可以说「北京天气」查指定城市的天气~"
	}

	data, errMsg := h.queryWeatherData(city)
	if data == nil {
		return errMsg
	}
	return h.buildWeatherReply(data)
}

var citySuffixPat = regexp.MustCompile(`([^\s,，、的]+?)(?:市|省)`)

var (
	cityDict     map[string]bool
	sortedCities []string
	loadCities   sync.Once
)

func extractCity(text string) string {
	loadCities.Do(func() {
		cityDict = map[string]bool{
			"北京": true, "上海": true, "天津": true, "重庆": true,
			"香港": true, "澳门": true,
			"河北": true, "山西": true, "辽宁": true, "吉林": true, "黑龙江": true,
			"江苏": true, "浙江": true, "安徽": true, "福建": true, "江西": true,
			"山东": true, "河南": true, "湖北": true, "湖南": true, "广东": true,
			"海南": true, "四川": true, "贵州": true, "云南": true, "陕西": true,
			"甘肃": true, "青海": true, "台湾": true,
			"内蒙古": true, "广西": true, "西藏": true, "宁夏": true, "新疆": true,
			"石家庄": true, "唐山": true, "秦皇岛": true, "邯郸": true, "邢台": true,
			"保定": true, "张家口": true, "承德": true, "沧州": true, "廊坊": true, "衡水": true,
			"太原": true, "大同": true, "阳泉": true, "长治": true, "晋城": true,
			"朔州": true, "晋中": true, "运城": true, "忻州": true, "临汾": true, "吕梁": true,
			"呼和浩特": true, "包头": true, "乌海": true, "赤峰": true, "通辽": true,
			"鄂尔多斯": true, "呼伦贝尔": true, "巴彦淖尔": true, "乌兰察布": true,
			"沈阳": true, "大连": true, "鞍山": true, "抚顺": true, "本溪": true,
			"丹东": true, "锦州": true, "营口": true, "阜新": true, "辽阳": true,
			"盘锦": true, "铁岭": true, "朝阳": true, "葫芦岛": true,
			"长春": true, "四平": true, "辽源": true, "通化": true,
			"白山": true, "松原": true, "白城": true, "延边": true,
			"哈尔滨": true, "齐齐哈尔": true, "鸡西": true, "鹤岗": true, "双鸭山": true,
			"大庆": true, "伊春": true, "佳木斯": true, "七台河": true, "牡丹江": true,
			"黑河": true, "绥化": true,
			"南京": true, "无锡": true, "徐州": true, "常州": true, "苏州": true,
			"南通": true, "连云港": true, "淮安": true, "盐城": true, "扬州": true,
			"镇江": true, "泰州": true, "宿迁": true,
			"杭州": true, "宁波": true, "温州": true, "嘉兴": true, "湖州": true,
			"绍兴": true, "金华": true, "衢州": true, "舟山": true, "台州": true, "丽水": true,
			"合肥": true, "芜湖": true, "蚌埠": true, "淮南": true, "马鞍山": true,
			"淮北": true, "铜陵": true, "安庆": true, "黄山": true, "滁州": true,
			"阜阳": true, "宿州": true, "六安": true, "亳州": true, "池州": true, "宣城": true,
			"福州": true, "厦门": true, "莆田": true, "三明": true, "泉州": true,
			"漳州": true, "南平": true, "龙岩": true, "宁德": true,
			"南昌": true, "景德镇": true, "萍乡": true, "九江": true, "新余": true,
			"鹰潭": true, "赣州": true, "吉安": true, "宜春": true, "抚州": true, "上饶": true,
			"济南": true, "青岛": true, "淄博": true, "枣庄": true, "东营": true,
			"烟台": true, "潍坊": true, "济宁": true, "泰安": true, "威海": true,
			"日照": true, "临沂": true, "德州": true, "聊城": true, "滨州": true, "菏泽": true,
			"郑州": true, "开封": true, "洛阳": true, "平顶山": true, "安阳": true,
			"鹤壁": true, "新乡": true, "焦作": true, "濮阳": true, "许昌": true,
			"漯河": true, "三门峡": true, "南阳": true, "商丘": true, "信阳": true,
			"周口": true, "驻马店": true,
			"武汉": true, "黄石": true, "十堰": true, "宜昌": true, "襄阳": true,
			"鄂州": true, "荆门": true, "孝感": true, "荆州": true, "黄冈": true,
			"咸宁": true, "随州": true, "恩施": true,
			"长沙": true, "株洲": true, "湘潭": true, "衡阳": true, "邵阳": true,
			"岳阳": true, "常德": true, "张家界": true, "益阳": true, "郴州": true,
			"永州": true, "怀化": true, "娄底": true, "湘西": true,
			"广州": true, "韶关": true, "深圳": true, "珠海": true, "汕头": true,
			"佛山": true, "江门": true, "湛江": true, "茂名": true, "肇庆": true,
			"惠州": true, "梅州": true, "汕尾": true, "河源": true, "阳江": true,
			"清远": true, "东莞": true, "中山": true, "潮州": true, "揭阳": true, "云浮": true,
			"南宁": true, "柳州": true, "桂林": true, "梧州": true, "北海": true,
			"防城港": true, "钦州": true, "贵港": true, "玉林": true, "百色": true,
			"贺州": true, "河池": true, "来宾": true, "崇左": true,
			"海口": true, "三亚": true, "儋州": true,
			"成都": true, "自贡": true, "攀枝花": true, "泸州": true, "德阳": true,
			"绵阳": true, "广元": true, "遂宁": true, "内江": true, "乐山": true,
			"南充": true, "眉山": true, "宜宾": true, "广安": true, "达州": true,
			"雅安": true, "巴中": true, "资阳": true,
			"贵阳": true, "六盘水": true, "遵义": true, "安顺": true, "毕节": true, "铜仁": true,
			"昆明": true, "曲靖": true, "玉溪": true, "保山": true, "昭通": true,
			"丽江": true, "普洱": true, "临沧": true, "大理": true,
			"楚雄": true, "红河": true, "文山": true, "西双版纳": true, "德宏": true,
			"拉萨": true, "日喀则": true, "昌都": true, "林芝": true, "山南": true, "那曲": true,
			"西安": true, "铜川": true, "宝鸡": true, "咸阳": true, "渭南": true,
			"延安": true, "汉中": true, "榆林": true, "安康": true, "商洛": true,
			"兰州": true, "嘉峪关": true, "金昌": true, "白银": true, "天水": true,
			"武威": true, "张掖": true, "平凉": true, "酒泉": true, "庆阳": true,
			"定西": true, "陇南": true,
			"西宁": true, "海东": true,
			"银川": true, "石嘴山": true, "吴忠": true, "固原": true, "中卫": true,
			"乌鲁木齐": true, "克拉玛依": true, "吐鲁番": true, "哈密": true,
			"阿克苏": true, "喀什": true, "和田": true, "伊犁": true, "昌吉": true, "石河子": true,
		}
		sortedCities = make([]string, 0, len(cityDict))
		for name := range cityDict {
			sortedCities = append(sortedCities, name)
		}
		sort.Slice(sortedCities, func(i, j int) bool {
			return len(sortedCities[i]) > len(sortedCities[j])
		})
	})

	matches := citySuffixPat.FindAllStringSubmatch(text, -1)
	var best string
	for _, m := range matches {
		if cityDict[m[1]] {
			if m[2] == "市" {
				best = m[1]
			} else if best == "" {
				best = m[1]
			}
		}
	}
	if best != "" {
		return best
	}

	for _, name := range sortedCities {
		if strings.Contains(text, name) {
			return name
		}
	}
	return ""
}

var datePat = regexp.MustCompile(`(\d{1,2})月(\d{1,2})(?:号|日)`)

func extractDate(text string) string {
	now := time.Now()
	if strings.Contains(text, "昨天") {
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	if strings.Contains(text, "前天") {
		return now.AddDate(0, 0, -2).Format("2006-01-02")
	}
	if m := datePat.FindStringSubmatch(text); len(m) == 3 {
		month, _ := strconv.Atoi(m[1])
		day, _ := strconv.Atoi(m[2])
		return time.Date(now.Year(), time.Month(month), day, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	}
	return ""
}

var chineseNum = map[string]int{"一": 1, "二": 2, "三": 3, "四": 4, "五": 5}
var grabNumPat = regexp.MustCompile(`第([一二三四五\d])[^\d]*|^抢([一二三四五\d])`)

func parseGrabIndex(text string) (int, bool) {
	if m := grabNumPat.FindStringSubmatch(text); len(m) > 1 {
		for _, g := range m[1:] {
			if g == "" {
				continue
			}
			if n, ok := chineseNum[g]; ok {
				return n, true
			}
			if n, err := strconv.Atoi(g); err == nil {
				return n, true
			}
		}
	}
	return 0, false
}

func (h *AiHandler) regeoCity(lat, lng float64) string {
	locStr := fmt.Sprintf("%.6f,%.6f", lng, lat)
	params := map[string]string{
		"key":      h.amapKey,
		"location": locStr,
		"radius":   "1000",
	}
	sig := h.buildAmapSig(params)
	urlStr := fmt.Sprintf("https://restapi.amap.com/v3/geocode/regeo?key=%s&location=%s&radius=1000&sig=%s",
		url.QueryEscape(h.amapKey), url.QueryEscape(locStr), sig)

	resp, err := h.httpClient.Get(urlStr)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status    string `json:"status"`
		Regeocode *struct {
			AddressComponent *struct {
				City     string `json:"city"`
				CityCode string `json:"citycode"`
				AdCode   string `json:"adcode"`
				Province string `json:"province"`
			} `json:"addressComponent"`
		} `json:"regeocode"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Status != "1" || result.Regeocode == nil {
		return ""
	}

	if result.Regeocode.AddressComponent.City != "" {
		return result.Regeocode.AddressComponent.City
	}
	return result.Regeocode.AddressComponent.Province
}

func (h *AiHandler) buildAmapSig(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(params[k])
	}
	b.WriteString(h.amapSignKey)
	return fmt.Sprintf("%x", md5.Sum([]byte(b.String())))
}

func (h *AiHandler) queryWeatherData(city string) (*weatherData, string) {
	params := map[string]string{
		"key":        h.amapKey,
		"city":       city,
		"extensions": "base",
	}
	sig := h.buildAmapSig(params)
	urlStr := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?key=%s&city=%s&extensions=base&sig=%s",
		url.QueryEscape(h.amapKey), url.QueryEscape(city), sig)

	resp, err := h.httpClient.Get(urlStr)
	if err != nil {
		return nil, "天气服务暂时不可用，稍后再试试吧~"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
		Info   string `json:"info"`
		Lives  []struct {
			Province      string `json:"province"`
			City          string `json:"city"`
			Weather       string `json:"weather"`
			Temperature   string `json:"temperature"`
			WindDirection string `json:"winddirection"`
			WindPower     string `json:"windpower"`
			Humidity      string `json:"humidity"`
		} `json:"lives"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Status != "1" || len(result.Lives) == 0 {
		return nil, "天气服务暂时不可用，稍后再试试吧~"
	}

	live := result.Lives[0]
	return &weatherData{
		City:          live.City,
		Temp:          live.Temperature,
		Weather:       live.Weather,
		WindDirection: live.WindDirection,
		WindPower:     live.WindPower,
		Humidity:      live.Humidity,
	}, ""
}

func (h *AiHandler) buildWeatherReply(data *weatherData) string {
	sysPrompt := `你是花小猪助手，根据实时天气数据用一句话告诉司机天气情况和出车建议。
	说话像朋友聊天一样自然，可以适当用语气词（呢、哦、～），但不要用"亲""哈"。
	好天气鼓励出车，恶劣天气提醒安全，根据天气严重程度调整语气。
	如实引用数据，不要编造。`

	userMsg := fmt.Sprintf("城市：%s\n温度：%s°C\n天气：%s\n风向：%s\n风力：%s级\n湿度：%s%%",
		data.City, data.Temp, data.Weather, data.WindDirection, data.WindPower, data.Humidity)

	messages := []map[string]string{
		{"role": "system", "content": sysPrompt},
		{"role": "user", "content": userMsg},
	}

	body, _ := json.Marshal(map[string]any{
		"model":    h.cfg.LlmModel,
		"messages": messages,
		"stream":   false,
	})

	httpReq, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Sprintf("%s %s°C，%s。", data.City, data.Temp, data.Weather)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.cfg.LlmApiKey)

	resp, err := h.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Sprintf("%s %s°C，%s。", data.City, data.Temp, data.Weather)
	}
	defer resp.Body.Close()

	var dsResp deepseekResp
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil || len(dsResp.Choices) == 0 {
		return fmt.Sprintf("%s %s°C，%s。", data.City, data.Temp, data.Weather)
	}

	return strings.TrimSpace(dsResp.Choices[0].Message.Content)
}

// ---------- 热点区域查询 ----------

func (h *AiHandler) findHotAreas(c *gin.Context, driverID int64, userText string) string {
	var lat, lng float64
	if driverID > 0 && h.mongoDb != nil {
		mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var loc struct {
			Lat float64 `bson:"lat"`
			Lng float64 `bson:"lng"`
		}
		if err := h.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err == nil && loc.Lat != 0 {
			lat = loc.Lat
			lng = loc.Lng
		}
	}
	if lat == 0 {
		return "暂时获取不到您的位置，无法推荐等单区域。"
	}

	radius := 3000
	re := h.rdb
	if re != nil {
		rkey := fmt.Sprintf("dh:hot:radius:%d", driverID)
		if val, err := re.Get(c.Request.Context(), rkey).Int(); err == nil && val > 0 {
			if strings.Contains(userText, "远") || strings.Contains(userText, "大") || strings.Contains(userText, "再") {
				radius = val * 2
			}
		}
	}
	if radius == 3000 && (strings.Contains(userText, "远") || strings.Contains(userText, "大")) {
		radius = 8000
	}

	if re != nil {
		re.Set(c.Request.Context(), fmt.Sprintf("dh:hot:radius:%d", driverID), radius, dhCtxTTL)
	}

	var hotAreas []string
	for _, t := range []string{"060100", "050201", "120201"} {
		names := h.searchPOI(lng, lat, t, 3, radius)
		hotAreas = append(hotAreas, names...)
	}

	if len(hotAreas) == 0 {
		return "附近暂时没有找到热点区域，建议去商圈或者写字楼附近看看。"
	}
	if len(hotAreas) > 3 {
		hotAreas = hotAreas[:3]
	}
	if radius > 3000 {
		return fmt.Sprintf("再远一点的%s人流量也不小，可以去看看。", strings.Join(hotAreas, "、"))
	}
	return fmt.Sprintf("附近%s人流量比较大，建议去那边等单。", strings.Join(hotAreas, "、"))
}

func (h *AiHandler) searchPOI(lng, lat float64, poiType string, limit, radius int) []string {
	locStr := fmt.Sprintf("%.6f,%.6f", lng, lat)
	params := map[string]string{
		"key":      h.amapKey,
		"location": locStr,
		"types":    poiType,
		"radius":   strconv.Itoa(radius),
		"offset":   strconv.Itoa(limit),
	}
	sig := h.buildAmapSig(params)
	urlStr := fmt.Sprintf("https://restapi.amap.com/v3/place/around?key=%s&location=%s&types=%s&radius=%d&offset=%d&sig=%s",
		url.QueryEscape(h.amapKey), url.QueryEscape(locStr), url.QueryEscape(poiType), radius, limit, sig)

	resp, err := h.httpClient.Get(urlStr)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
		POIs   []struct {
			Name string `json:"name"`
		} `json:"pois"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Status != "1" {
		return nil
	}

	var names []string
	for _, p := range result.POIs {
		if len(names) >= limit {
			break
		}
		names = append(names, p.Name)
	}
	return names
}

// searchPOIByKeyword 用关键词搜索附近 POI
func (h *AiHandler) searchPOIByKeyword(lng, lat float64, keyword string, limit, radius int) []string {
	locStr := fmt.Sprintf("%.6f,%.6f", lng, lat)
	params := map[string]string{
		"key":      h.amapKey,
		"location": locStr,
		"keywords": keyword,
		"radius":   strconv.Itoa(radius),
		"offset":   strconv.Itoa(limit),
	}
	sig := h.buildAmapSig(params)
	urlStr := fmt.Sprintf("https://restapi.amap.com/v3/place/around?key=%s&location=%s&keywords=%s&radius=%d&offset=%d&sig=%s",
		url.QueryEscape(h.amapKey), url.QueryEscape(locStr), url.QueryEscape(keyword), radius, limit, sig)

	resp, err := h.httpClient.Get(urlStr)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status string `json:"status"`
		POIs   []struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"pois"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Status != "1" {
		return nil
	}

	var names []string
	for _, p := range result.POIs {
		if len(names) >= limit {
			break
		}
		name := p.Name
		if name == "公共厕所" && p.Address != "" {
			name = p.Address
		}
		names = append(names, name)
	}
	return names
}

// queryNearbyPOI 通用搜索附近 POI
func (h *AiHandler) queryNearbyPOI(c *gin.Context, driverID int64, keyword, fallback string) string {
	var lat, lng float64
	if driverID > 0 && h.mongoDb != nil {
		mongoCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var loc struct {
			Lat float64 `bson:"lat"`
			Lng float64 `bson:"lng"`
		}
		if err := h.mongoDb.Collection("driver_local").FindOne(mongoCtx, bson.M{"driver_id": driverID}).Decode(&loc); err == nil && loc.Lat != 0 {
			lat = loc.Lat
			lng = loc.Lng
		}
	}
	if lat == 0 {
		return "暂时获取不到您的位置，附近加油站、商场一般都有，您可以留意一下。"
	}

	names := h.searchPOIByKeyword(lng, lat, keyword, 3, 2000)
	if len(names) == 0 {
		return fallback
	}
	if len(names) > 3 {
		names = names[:3]
	}
	return fmt.Sprintf("您附近有：%s。", strings.Join(names, "、"))
}

// ---------- Baidu TTS ----------

func (h *AiHandler) getBaiduToken() (string, error) {
	h.baiduMu.Lock()
	defer h.baiduMu.Unlock()

	if h.baiduToken != "" && time.Since(h.baiduTokenAt) < 25*24*time.Hour {
		return h.baiduToken, nil
	}

	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("client_id", h.cfg.TtsApiKey)
	v.Set("client_secret", h.cfg.TtsSecretKey)

	resp, err := h.httpClient.PostForm("https://aip.baidubce.com/oauth/2.0/token", v)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", nil
	}

	h.baiduToken = result.AccessToken
	h.baiduTokenAt = time.Now()
	return h.baiduToken, nil
}

// textToSpeech 将文本转为语音音频
func (h *AiHandler) textToSpeech(text string) ([]byte, error) {
	if len([]rune(text)) > 500 {
		text = string([]rune(text)[:500])
	}

	token, err := h.getBaiduToken()
	if err != nil || token == "" {
		return nil, fmt.Errorf("tts auth failed")
	}

	v := url.Values{}
	v.Set("tex", text)
	v.Set("lan", "zh")
	v.Set("ctp", "1")
	v.Set("tok", token)
	v.Set("per", "3")
	v.Set("spd", "5")
	v.Set("pit", "5")
	v.Set("vol", "5")
	v.Set("cuid", "huaxiaozhu")

	resp, err := h.httpClient.PostForm(h.cfg.TtsApiUrl, v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct != "" && ct != "audio/mp3" {
		return nil, fmt.Errorf("tts error: %s", ct)
	}

	return io.ReadAll(resp.Body)
}

func (h *AiHandler) TTS(c *gin.Context) {
	text := c.Query("text")
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text required"})
		return
	}
	audio, err := h.textToSpeech(text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "audio/mp3", audio)
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (h *AiHandler) getStatusDescription(ctx *gin.Context, driverID int64, workStatus int32) string {
	if h.dhCfg == nil || h.dhCfg.StatusDescriptions == nil {
		switch workStatus {
		case int32(constants.WorkStatusOffline):
			return "已收车离线"
		case int32(constants.WorkStatusOnline):
			return "空闲在线"
		default:
			return "忙碌"
		}
	}

	switch workStatus {
	case int32(constants.WorkStatusOffline):
		return h.dhCfg.StatusDescriptions["offline"]
	case int32(constants.WorkStatusOnline):
		return h.dhCfg.StatusDescriptions["idle"]
	case int32(constants.WorkStatusListening):
		order, err := h.client.GetCurrentOrder(ctx, &pb.GetCurrentOrderReq{DriverId: driverID})
		if err != nil || order == nil || !order.Success || order.OrderId == 0 {
			return h.dhCfg.StatusDescriptions["listening"]
		}
		switch order.Status {
		case int32(constants.OrderStatusAccepted):
			return h.dhCfg.StatusDescriptions["accepted"]
		case int32(constants.OrderStatusArrived):
			return h.dhCfg.StatusDescriptions["arrived"]
		case int32(constants.OrderStatusInTrip):
			return h.dhCfg.StatusDescriptions["in_trip"]
		default:
			return h.dhCfg.StatusDescriptions["listening"]
		}
	default:
		return "未知状态"
	}
}
