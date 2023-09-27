package user

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"project-api/router"
)

func init() {
	router.Register(&RouterUser{})
}

type RouterUser struct {
}

func (*RouterUser) Route(r *gin.Engine) {
	InitRpcUserClient()
	h := New()
	zap.L().Info("接口开始执行")
	r.POST("/project/login/getCaptcha", h.getCaptcha)
	r.POST("/project/login/register", h.register)
	r.POST("/project/login", h.login)
	r.POST("/project/organization/_getOrgList", h.myOrgList)
}
