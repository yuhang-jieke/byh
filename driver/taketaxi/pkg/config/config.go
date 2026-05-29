package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Registry RegistryConfig `yaml:"registry"`
	Mongo    MongoConfig    `yaml:"mongo"`
	Dispatch DispatchConfig `yaml:"dispatch"`
	Amap     AmapConfig     `yaml:"amap"`
	Ai       AiConfig       `yaml:"ai"`
}

type ServerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	GRPCHost string `yaml:"grpc_host"`
	GRPCPort int    `yaml:"grpc_port"`
}

type DatabaseConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	Password string `yaml:"Password"`
	Database int    `yaml:"Database"`
}

type RegistryConfig struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
}

type MongoConfig struct {
	Uri      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type DispatchConfig struct {
	RadiusKm             float64 `yaml:"radius_km"`
	MinServiceScore      float64 `yaml:"min_service_score"`
	ArriveCheckRadius    float64 `yaml:"arrive_check_radius"`
	EndTripCheckRadius   float64 `yaml:"end_trip_check_radius"`
	ExpandedRadiusKm     float64 `yaml:"expanded_radius_km"`
	MaxExpandedAttempts  int     `yaml:"max_expanded_attempts"`
	RejectionThreshold   int     `yaml:"rejection_threshold"`
	DispatchTimeoutSec   int     `yaml:"dispatch_timeout_sec"`
	PoolTimeoutSec       int     `yaml:"pool_timeout_sec"`
	PoolRetryIntervalSec int     `yaml:"pool_retry_interval_sec"`
}

type AmapConfig struct {
	WebAPIKey    string `yaml:"web_api_key"`
	WebAPISignKey string `yaml:"web_api_sign_key"`
}

type AiConfig struct {
	LlmApiKey    string `yaml:"llm_api_key"`
	LlmModel     string `yaml:"llm_model"`
	TtsApiKey    string `yaml:"tts_api_key"`
	TtsSecretKey string `yaml:"tts_secret_key"`
	TtsApiUrl    string `yaml:"tts_api_url"`
}

// DigitalHumanIntent 一个意图定义
type DigitalHumanIntent struct {
	Name     string   `yaml:"name"`
	Triggers []string `yaml:"triggers"`
	Action   string   `yaml:"action"` // call_api | say | llm_chat
	API      string   `yaml:"api,omitempty"`
	Reply    string   `yaml:"reply,omitempty"`
}

// FallbackConfig LLM 兜底配置
type FallbackConfig struct {
	Action             string   `yaml:"action"`
	SystemPromptExtras []string `yaml:"system_prompt_extras"`
}

// DigitalHumanConfig 数字人完整配置
type DigitalHumanConfig struct {
	Name               string                `yaml:"name"`
	Role               string                `yaml:"role"`
	Tone               string                `yaml:"tone"`
	MaxResponseLen     int                   `yaml:"max_response_length"`
	StatusDescriptions map[string]string    `yaml:"status_descriptions"`
	Intents            []DigitalHumanIntent  `yaml:"intents"`
	Fallback           FallbackConfig        `yaml:"fallback"`
	ResponseRules      []string              `yaml:"response_rules"`
}

func (c *DigitalHumanConfig) BuildSystemPrompt() string {
	var parts []string
	parts = append(parts, fmt.Sprintf(
		"你是一个%s，名叫%s，%s。回答限制在%d字以内。",
		c.Role, c.Name, c.Tone, c.MaxResponseLen,
	))
	parts = append(parts, c.ResponseRules...)
	parts = append(parts, c.Fallback.SystemPromptExtras...)
	return strings.Join(parts, "。")
}

func LoadDigitalHuman(path string) (*DigitalHumanConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg DigitalHumanConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
