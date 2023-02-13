package main

import (
	pbfavorite "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/favorite/route"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/cronUtils"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"google.golang.org/grpc"
	"net"
)

const (
	port = ":50051"
)

func initAll() {
	initialization.InitConfig()
	initialization.InitDB()
	initialization.InitOSS()
	initialization.InitRDB()
	initialization.InitKafkaServer()
	initialization.InitKafkaClient()

	//Init Utils
	logger.InitLogger(initialization.LogConf)
	jwt.InitJwt()
	cronUtils.InitCron()

	//Init lower Levels
	dao.DaoInitialization()
}

func main() {
	initAll()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.GlobalLogger.Fatal().Err(err)
	}
	s := grpc.NewServer()
	pbfavorite.RegisterFavoriteInfoServer(s, service.GetFavoriteServiceInstance())
	if err = s.Serve(lis); err != nil {
		panic(err)
	}
}
