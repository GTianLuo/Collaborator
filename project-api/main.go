package main

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	_ "project-api/api"
	"project-api/router"
	"project-api/tracing"
	srv "project-common"
	"test.com/project-user/config"
)

func main() {
	r := gin.Default()
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	//从配置中读取日志配置，初始化日志
	config.AppConf.InitZapLog()
	r.Use(otelgin.Middleware("project-api"))
	router.InitRouter(r)
	srv.Run(r, "web", ":80", nil)
}
