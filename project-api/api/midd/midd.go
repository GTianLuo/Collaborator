package midd

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"project-api/api/user"
	common "project-common"
	"project-common/errs"
	"project-grpc/user/login"
	"time"
)

func TokenVerify() func(c *gin.Context) {
	return func(c *gin.Context) {
		result := &common.Result{}
		token := c.GetHeader("Authorization")
		//验证用户是否已经登录
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		memberFind, err2 := user.LoginServiceClient.TokenVerify(ctx, &login.TokenVerifyMessage{Token: token})
		if err2 != nil {
			code, msg := errs.ParseGrpcError(err2)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort()
			return
		}
		c.Set("memberId", memberFind.Member.Id)
		c.Set("memberName", memberFind.Member.Name)
		c.Set("organizationCode", memberFind.Member.OrganizationCode)
		c.Next()
	}
}
