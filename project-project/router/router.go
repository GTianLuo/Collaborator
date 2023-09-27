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
	"project-grpc/account"
	"project-grpc/auth"
	"project-grpc/department"
	project_service_v1 "project-grpc/project"
	task_service_v1 "project-grpc/task"
	account_service_v1 "project-project/pkg/service/account.service.v1"
	auth_service_v1 "project-project/pkg/service/auth.service.v1"
	department_service_v1 "project-project/pkg/service/department.service.v1"
	project_service "project-project/pkg/service/project.service.v1"
	task_service "project-project/pkg/service/task.service.v1"
	"test.com/project-user/config"
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
			project_service_v1.RegisterProjectServiceServer(g, project_service.New())
			task_service_v1.RegisterTaskServiceServer(g, task_service.New())
			account.RegisterAccountServiceServer(g, account_service_v1.New())
			auth.RegisterAuthServiceServer(g, auth_service_v1.New())
			department.RegisterDepartmentServiceServer(g, department_service_v1.New())
		},
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			otelgrpc.UnaryServerInterceptor(),
			//interceptor.New().CacheInterceptor(),
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
