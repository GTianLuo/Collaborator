package interceptor

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"project-common/encrypts"
	"project-grpc/user/login"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/repo"
	"time"
)

type Interceptor struct {
	cache    repo.Cache
	cacheMap map[string]any
}

func NewInterceptor() *Interceptor {
	cacheMap := make(map[string]any)
	cacheMap["/login.service.v1.LoginService/MyOrgList"] = &login.OrgListResponse{}
	cacheMap["/login.service.v1.LoginService/FindMemInfoById"] = &login.MemberMessage{}
	return &Interceptor{cache: dao.Rc, cacheMap: cacheMap}
}
func (i *Interceptor) CacheInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		respType := i.cacheMap[info.FullMethod]
		if respType == nil {
			return handler(ctx, req)
		}
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))
		respJson, err := i.cache.Get(c, info.FullMethod+"::"+cacheKey)
		if err == nil {
			json.Unmarshal([]byte(respJson), &respType)
			return respType, nil
		}
		resp, err = handler(ctx, req)
		if resp == nil {
			return
		}
		bytes, _ := json.Marshal(resp)
		i.cache.Put(c, info.FullMethod+"::"+cacheKey, string(bytes), 5*time.Minute)
		return
	})
}

func CacheInterceptorFunc() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		i := NewInterceptor()
		respType := i.cacheMap[info.FullMethod]
		if respType == nil {
			return handler(ctx, req)
		}
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))
		respJson, err := i.cache.Get(c, info.FullMethod+"::"+cacheKey)
		if err == nil {
			json.Unmarshal([]byte(respJson), &respType)
			return respType, nil
		}
		resp, err = handler(ctx, req)
		if resp == nil {
			return
		}
		bytes, _ := json.Marshal(resp)
		i.cache.Put(c, info.FullMethod+"::"+cacheKey, string(bytes), 5*time.Minute)
		return
	}
}
