package main

import (
	"github.com/gin-gonic/gin"
	srv "project-common"
	_ "test.com/project-user/api"
	"test.com/project-user/config"
	"test.com/project-user/router"
)

func main() {

	r := gin.Default()
	//从配置中读取日志配置，初始化日志
	config.AppConf.InitZapLog()
	router.InitRouter(r)
	gc := router.RegisterGrpc()
	router.RegisterEtcdServer()
	stop := func() {
		gc.Stop()
	}
	srv.Run(r, "web", config.AppConf.Sc.Addr, stop)
}
