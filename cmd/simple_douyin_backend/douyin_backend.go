package main

import (
	"fmt"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/init/router"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// 用于单机的极简版抖音后端程序

// initAll 初始化所有的部分
func initAll() {
	initialization.InitConfig()
	initialization.InitDB()
	initialization.InitOSS()
	initialization.InitRDB()
	logger.InitLogger(initialization.LogConf)

	jwt.InitJwt()
}

func main() {
	initAll()
	hServer := server.Default(server.WithHostPorts(fmt.Sprintf("127.0.0.1:%s", initialization.Port)))

	router.InitRouter(hServer)
	hServer.Spin()
}
