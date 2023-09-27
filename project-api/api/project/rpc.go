package project

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"project-api/config"
	"project-common/discovery"
	"project-common/logs"
	"project-grpc/account"
	"project-grpc/auth"
	"project-grpc/department"
	"project-grpc/menu"
	project_service_v1 "project-grpc/project"
	task_service_v1 "project-grpc/task"
)

var ProjectServiceClient project_service_v1.ProjectServiceClient
var TaskServiceClient task_service_v1.TaskServiceClient
var AccountServiceClient account.AccountServiceClient
var DepartmentServiceClient department.DepartmentServiceClient
var AuthServiceClient auth.AuthServiceClient
var MenuServiceClient menu.MenuServiceClient

func InitRpcUserClient() {
	etcdRegister := discovery.NewResolver(config.AppConf.EtcdConfig.Addrs, logs.LG)
	resolver.Register(etcdRegister)
	conn, err := grpc.Dial(
		"etcd:///user",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	ProjectServiceClient = project_service_v1.NewProjectServiceClient(conn)
	TaskServiceClient = task_service_v1.NewTaskServiceClient(conn)
	AccountServiceClient = account.NewAccountServiceClient(conn)
	DepartmentServiceClient = department.NewDepartmentServiceClient(conn)
	AuthServiceClient = auth.NewAuthServiceClient(conn)
	MenuServiceClient = menu.NewMenuServiceClient(conn)
}
