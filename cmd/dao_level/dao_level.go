package main

import (
	pbuser "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/user"
	pbvideo "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/video"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/dao"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/cronUtils"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"google.golang.org/grpc"
	"net"
	"sync"
)

func initAll() {
	initialization.InitConfig()
	initialization.InitDB()
	initialization.InitKafkaClient()

	//Init Utils
	logger.InitLogger(initialization.LogConf)
	cronUtils.InitCron()

	//Init lower Levels
	dao.DaoInitialization()
}

var wg sync.WaitGroup

func main() {
	initAll()

	// 开一个协程监听UserDao端口
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", initialization.RpcSDConf.UserServicePort)
		if err != nil {
			logger.GlobalLogger.Fatal().Err(err)
		} else {
			logger.GlobalLogger.Printf("Successfully Listen At port %v", initialization.RpcSDConf.UserServicePort)
		}
		s := grpc.NewServer()
		pbuser.RegisterUserDaoInfoServer(s, dao.GetUserDaoInstance())
		logger.GlobalLogger.Printf("Successfully register userInfo Server")
		if err = s.Serve(lis); err != nil {
			logger.GlobalLogger.Printf("Serving userInfo error")
			panic(err)
		}
	}()

	// 开一个协程监听videoDao端口
	wg.Add(1)
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", initialization.RpcSDConf.VideoServicePort)
		if err != nil {
			logger.GlobalLogger.Fatal().Err(err)
		} else {
			logger.GlobalLogger.Printf("Successfully Listen At port %v", initialization.RpcSDConf.VideoServicePort)
		}
		s := grpc.NewServer()
		pbvideo.RegisterVideoDaoInfoServer(s, dao.GetVideoDaoInstance())
		logger.GlobalLogger.Printf("Successfully register videoInfo Server")
		if err = s.Serve(lis); err != nil {
			logger.GlobalLogger.Printf("Serving videoInfo error")
			panic(err)
		}
	}()

	wg.Wait()
}
