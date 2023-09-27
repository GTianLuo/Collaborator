package main

import (
	"github.com/gin-gonic/gin"
	srv "project-common"
	_ "project-project/api"
	"project-project/config"
	"project-project/internal/rpc"
	"project-project/router"
)

func main() {
	r := gin.Default()
	//从配置中读取日志配置，初始化日志
	config.AppConf.InitZapLog()
	rpc.InitRpcUserClient()
	router.InitRouter(r)
	gc := router.RegisterGrpc()
	router.RegisterEtcdServer()
	stop := func() {
		gc.Stop()
	}
	srv.Run(r, "project", config.AppConf.Sc.Addr, stop)
}
