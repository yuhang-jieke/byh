package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"driver/taketaxi/common/kitexGen/driver/driverservice"
	"driver/taketaxi/pkg/config"
	"driver/taketaxi/pkg/database"
	"driver/taketaxi/pkg/mongodb"
	"driver/taketaxi/pkg/redis"
	"driver/taketaxi/srvDriver/internal/cache"
	"driver/taketaxi/srvDriver/internal/handler"
	"driver/taketaxi/srvDriver/internal/repository"

	"github.com/cloudwego/kitex/server"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "../configs/config.yaml", "config file")
}

func main() {
	flag.Parse()
	cfg, err := config.Load(confPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	// Redis 初始化（用于抢单池缓存）
	rdb := redis.NewRedisClient(&cfg.Redis)
	poolCache := cache.NewPoolCache(rdb)

	mongoDb, closeMongo := mongodb.NewMongoDB(cfg.Mongo.Uri, cfg.Mongo.Database)
	if closeMongo != nil {
		defer closeMongo()
	}

	repo := repository.NewDriverRepo(db)
	h := handler.NewDriverHandler(mongoDb, repo, &cfg.Dispatch, poolCache)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("resolve addr: %v", err)
	}
	svr := driverservice.NewServer(h, server.WithServiceAddr(tcpAddr))
	log.Println("Starting on", addr)
	if err := svr.Run(); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
