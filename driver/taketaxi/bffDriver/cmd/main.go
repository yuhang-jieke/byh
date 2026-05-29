package main

import (
	"context"
	"driver/taketaxi/bffDriver/internal/router"
	"driver/taketaxi/bffDriver/internal/rpcClient"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/mongodb"
	"flag"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "../configs/config.yaml", "config file")
}

func main() {
	flag.Parse()
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatal(err)
	}
	grpcAddr := fmt.Sprintf("%s:%d", cfg.Server.GRPCHost, cfg.Server.GRPCPort)
	client, err := rpcclient.NewDriverClient(grpcAddr)
	if err != nil {
		log.Fatalf("failed to create gRPC client: %v", err)
	}
	defer client.Close()
	dhCfg, err := config.LoadDigitalHuman("../configs/digital_human.yaml")
	if err != nil {
		log.Printf("warn: digital_human.yaml not loaded: %v", err)
		dhCfg = &config.DigitalHumanConfig{}
	}

	// Redis（非必须，不可用时降级，数字人对话上下文丢失）
	var rdb *redis.Client
	if cfg.Redis.Host != "" {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.Database,
		})
		if err := rdb.Ping(context.Background()).Err(); err != nil {
			log.Printf("warn: Redis unavailable, DH history disabled: %v", err)
			rdb = nil
		}
	}

	// MongoDB（非必须，天气查询需要获取司机实时位置）
	mongoDb, closeMongo := mongodb.NewMongoDB(cfg.Mongo.Uri, cfg.Mongo.Database)
	if closeMongo != nil {
		defer closeMongo()
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("BFF starting on %s", addr)
	router.NewRouter(client, cfg.Amap.WebAPIKey, cfg.Amap.WebAPISignKey, &cfg.Ai, dhCfg, rdb, mongoDb).Run(addr)
}
