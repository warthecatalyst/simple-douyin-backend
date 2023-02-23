package main

import (
	"fmt"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/init/router"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/cronUtils"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// initAll 初始化所有的部分
func initAll() {
	//Init basic operators
	initialization.InitConfig()

	//Init Utils
	logger.InitLogger(initialization.LogConf)
	jwt.InitJwt()
	cronUtils.InitCron()
}

func main() {
	initAll()
	hServer := server.Default(server.WithHostPorts(fmt.Sprintf("127.0.0.1:%s", initialization.Port)))

	router.InitRouter(hServer)
	hServer.Spin()
}
