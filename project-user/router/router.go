package router

import (
	"github.com/gin-gonic/gin"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
	"project-common/discovery"
	"project-common/logs"
	login_service_v1 "project-grpc/user/login"
	"project-project/config"
	"test.com/project-user/internal/interceptor"
	login_service "test.com/project-user/pkg/service"
)

type Router interface {
	Route(r *gin.Engine)
}
type RegisterRouter struct {
}

func New() *RegisterRouter {
	return &RegisterRouter{}
}
func (*RegisterRouter) Route(ro Router, r *gin.Engine) {
	ro.Route(r)
}

var routers []Router

func InitRouter(r *gin.Engine) {
	/*	rg := New()
		rg.Route(&user.RouterUser{}, r)*/
	for _, ro := range routers {
		ro.Route(r)
	}
}
func Register(ro ...Router) {
	routers = append(routers, ro...)
}

type gRPCConfig struct {
	Addr         string
	RegisterFunc func(*grpc.Server)
}

func RegisterGrpc() *grpc.Server {
	c := gRPCConfig{
		Addr: config.AppConf.Gc.Addr,
		RegisterFunc: func(g *grpc.Server) {
			login_service_v1.RegisterLoginServiceServer(g, login_service.New())
		},
	}

	//newInterceptor := interceptor.NewInterceptor()
	var s = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			otelgrpc.UnaryServerInterceptor(),
			interceptor.CacheInterceptorFunc(),
		)),
	)
	c.RegisterFunc(s)
	lis, err := net.Listen("tcp", config.AppConf.Gc.Addr)
	if err != nil {
		log.Println("cannot listen")
	}
	go func() {
		err = s.Serve(lis)
		if err != nil {
			log.Println("server started error", err)
			return
		}
	}()
	return s
}
func RegisterEtcdServer() {
	etcdRegister := discovery.NewResolver(config.AppConf.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	info := discovery.Server{
		Name:    config.AppConf.Gc.Name,
		Addr:    config.AppConf.Gc.Addr,
		Version: config.AppConf.Gc.Version,
		Weight:  int64(config.AppConf.Gc.Weight),
	}
	r := discovery.NewRegister(config.AppConf.EtcdConfig.Addrs, logs.LG)
	_, err := r.Register(info, 2)
	if err != nil {
		log.Fatalln(err)
	}
}
