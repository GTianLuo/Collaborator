package login_service_v1

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"log"
	common "project-common"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/repo"
	"time"
)

type LoginService struct {
	UnimplementedLoginServiceServer
	cache repo.Cache
}

func New() *LoginService {
	return &LoginService{
		cache: dao.Rc,
	}
}
func (ls *LoginService) GetCaptcha(ctx context.Context, msg *CaptchaMessage) (*CaptchaResponse, error) {
	//rsp := &common.Result{}
	//mobile := ctx.PostForm("mobile")
	mobile := msg.Mobile
	if !common.VerifyMobile(mobile) {
		return nil, errors.New("手机号不合法")
	}
	code := "123456"
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功，发送短信")
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := ls.cache.Put(c, "REGISTER_"+mobile, code, 15*time.Minute)
		if err != nil {
			zap.L().Error("验证码存入redis出错，")
			fmt.Println(err)
		}
		log.Println("将手机号和验证码存入redis成功 REGISTER_%S : %S", mobile, code)
	}()
	return &CaptchaResponse{Code: code}, nil
}
