package common

import (
	"context"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(r *gin.Engine, srvName string, addr string, stop func()) {
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	//保证下面的优雅启停
	go func() {
		log.Printf("%s running in %s \n", srvName, srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()
	//植入两个信号 分别是ctrl c 和 关闭程序的时候触发
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting Down project web server...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if stop != nil {
		stop()
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("web server Shutdown,cause by :", err)
	}
	select {
	case <-ctx.Done():
		log.Println("关闭超时")
	default:
		log.Println("关闭成功")
	}
	log.Println("web server stop success")
}
