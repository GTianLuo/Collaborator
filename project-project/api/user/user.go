package user

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"net/http"
	common "project-common"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/repo"
	"time"
)

type HandlerUser struct {
	cache repo.Cache
}

func New() *HandlerUser {
	return &HandlerUser{
		cache: dao.Rc,
	}
}

func (h *HandlerUser) getCaptcha(ctx *gin.Context) {
	rsp := &common.Result{}
	mobile := ctx.PostForm("mobile")
	if !common.VerifyMobile(mobile) {
		ctx.JSON(200, "手机号不合法")
		return
	}
	code := "123456"
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功，发送短信")
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := h.cache.Put(c, "REGISTER_"+mobile, code, 15*time.Minute)
		if err != nil {
			zap.L().Error("验证码存入redis出错，")
		}
		log.Println("将手机号和验证码存入redis成功 REGISTER_%S : %S", mobile, code)
	}()
	ctx.JSON(http.StatusOK, rsp.Success("123456"))
	//ctx.JSON(200, rsp.Success("123456"))
}
