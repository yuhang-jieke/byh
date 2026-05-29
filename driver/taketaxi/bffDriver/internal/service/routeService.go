package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// RoutePlanRequest 路径规划请求
type RoutePlanRequest struct {
	OriginLat float64 `json:"origin_lat"`
	OriginLng float64 `json:"origin_lng"`
	DestLat   float64 `json:"dest_lat"`
	DestLng   float64 `json:"dest_lng"`
}

// RoutePlanResponse 路径规划响应
type RoutePlanResponse struct {
	Success  bool      `json:"success"`
	Error    string    `json:"error,omitempty"`
	Distance int64     `json:"distance"`  // 单位：米
	Duration int64     `json:"duration"`  // 单位：秒
	Polyline []LngLat  `json:"polyline"`  // 路线坐标点
}

// LngLat 经纬度坐标
type LngLat struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// RouteService 路径规划服务
type RouteService struct {
	apiKey  string
	signKey string
	client  *http.Client
}

// NewRouteService 创建路径规划服务
func NewRouteService(apiKey, signKey string) *RouteService {
	return &RouteService{
		apiKey:  apiKey,
		signKey: signKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PlanRoute 规划驾车路线
func (s *RouteService) PlanRoute(req RoutePlanRequest) (*RoutePlanResponse, error) {
	baseURL := "https://restapi.amap.com/v3/direction/driving"

	// 构造参数列表（按 key 字典序排列）
	type kv struct{ k, v string }
	args := []kv{
		{"destination", fmt.Sprintf("%.6f,%.6f", req.DestLng, req.DestLat)},
		{"extensions", "base"},
		{"key", s.apiKey},
		{"origin", fmt.Sprintf("%.6f,%.6f", req.OriginLng, req.OriginLat)},
		{"output", "json"},
	}
	sort.Slice(args, func(i, j int) bool { return args[i].k < args[j].k })

	// 用原始值（不 URL 编码）计算签名
	var rawParts []string
	for _, a := range args {
		rawParts = append(rawParts, a.k+"="+a.v)
	}
	queryStrRaw := strings.Join(rawParts, "&")

	var requestURL string
	if s.signKey != "" {
		sigInput := queryStrRaw + s.signKey
		hash := md5.Sum([]byte(sigInput))
		sig := hex.EncodeToString(hash[:])
		log.Printf("[路线规划] 签名输入: %s", sigInput)
		log.Printf("[路线规划] 签名结果: %s", sig)

		// 构建实际请求 URL（URL 编码 + sig）
		params := url.Values{}
		for _, a := range args {
			params.Set(a.k, a.v)
		}
		requestURL = baseURL + "?" + params.Encode() + "&sig=" + sig
	} else {
		params := url.Values{}
		for _, a := range args {
			params.Set(a.k, a.v)
		}
		requestURL = baseURL + "?" + params.Encode()
	}
	log.Printf("[路线规划] 完整 URL: %s", requestURL)

	resp, err := s.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	// 解析高德 API 响应
	var amapResp struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		Infocode string `json:"infocode"`
		Route    struct {
			Paths []struct {
				Distance string `json:"distance"`
				Duration string `json:"duration"`
				Steps    []struct {
					Polyline string `json:"polyline"`
				} `json:"steps"`
			} `json:"paths"`
		} `json:"route"`
	}

	if err := json.Unmarshal(body, &amapResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	// 检查错误
	if amapResp.Status != "1" {
		return &RoutePlanResponse{
			Success: false,
			Error:   fmt.Sprintf("status=%s, info=%s, infocode=%s", amapResp.Status, amapResp.Info, amapResp.Infocode),
		}, nil
	}

	// 没有路线数据
	if len(amapResp.Route.Paths) == 0 {
		return &RoutePlanResponse{
			Success: false,
			Error:   "no route found",
		}, nil
	}

	path := amapResp.Route.Paths[0]

	// 解析 polyline 坐标
	polyline := s.parsePolyline(path.Steps)

	// 解析距离和时长
	var distance, duration int64
	fmt.Sscanf(path.Distance, "%d", &distance)
	fmt.Sscanf(path.Duration, "%d", &duration)

	return &RoutePlanResponse{
		Success:  true,
		Distance: distance,
		Duration: duration,
		Polyline: polyline,
	}, nil
}

// parsePolyline 解析高德 polyline 格式
// 高德 polyline 格式：经度 1 纬度 1;经度 2 纬度 2;...
func (s *RouteService) parsePolyline(steps []struct {
	Polyline string `json:"polyline"`
}) []LngLat {
	var result []LngLat

	for _, step := range steps {
		if step.Polyline == "" {
			continue
		}
		// 按分号分割坐标点
		points := step.Polyline
		// 坐标点之间用分号分隔
		for i := 0; i < len(points); {
			var lng, lat float64
			j := i
			// 解析经度
			for j < len(points) && points[j] != ',' {
				j++
			}
			if j > i {
				fmt.Sscanf(points[i:j], "%f", &lng)
			}
			j++ // 跳过逗号
			// 解析纬度
			k := j
			for k < len(points) && points[k] != ';' {
				k++
			}
			if k > j {
				fmt.Sscanf(points[j:k], "%f", &lat)
			}
			if lng != 0 || lat != 0 {
				result = append(result, LngLat{Lng: lng, Lat: lat})
			}
			if k >= len(points) {
				break
			}
			i = k + 1 // 跳过分号
		}
	}

	return result
}
