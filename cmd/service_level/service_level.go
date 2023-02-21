package main

import (
	pbuser "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/user"
	pbvideo "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/video"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/service"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/cronUtils"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"google.golang.org/grpc"
	"net"
	"sync"
)

func initAll() {
	initialization.InitConfig()
	initialization.InitOSS()
	initialization.InitRDB()
	initialization.InitKafkaServer()

	//Init Utils
	logger.InitLogger(initialization.LogConf)
	cronUtils.InitCron()
}

var wg sync.WaitGroup

func main() {
	initAll()

	// 开一个协程监听UserService端口
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", initialization.RpcCSConf.UserServicePort)
		if err != nil {
			logger.GlobalLogger.Fatal().Err(err)
		} else {
			logger.GlobalLogger.Printf("Successfully Listen At port %v", initialization.RpcCSConf.UserServicePort)
		}
		s := grpc.NewServer()
		pbuser.RegisterUserServiceInfoServer(s, service.GetUserServiceInstance())
		logger.GlobalLogger.Printf("Successfully register userServiceInfo Server")
		if err = s.Serve(lis); err != nil {
			logger.GlobalLogger.Printf("Serving userInfo error")
			panic(err)
		}
	}()

	//开一个协程监听VideoService端口
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", initialization.RpcCSConf.VideoServicePort)
		if err != nil {
			logger.GlobalLogger.Fatal().Err(err)
		} else {
			logger.GlobalLogger.Printf("Successfully Listen At port %v", initialization.RpcCSConf.VideoServicePort)
		}
		s := grpc.NewServer()
		pbvideo.RegisterVideoServiceInfoServer(s, service.GetVideoServiceInstance())
		logger.GlobalLogger.Printf("Successfully register VideoServiceInfo Server")
		if err = s.Serve(lis); err != nil {
			logger.GlobalLogger.Printf("Serving videoInfo error")
			panic(err)
		}
	}()
	wg.Wait()
}
