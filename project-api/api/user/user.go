package user

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model/user"
	common "project-common"
	"project-common/errs"
	"project-grpc/user/login"
	"time"
)

type HandlerUser struct {
}

func New() *HandlerUser {
	return &HandlerUser{}
}
func (u *HandlerUser) register(c *gin.Context) {
	result := &common.Result{}
	var req user.RegisterReq
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "参数格式错误"))
		return
	}
	if err := req.Verify(); err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "参数不合法"))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &login.RegisterMessage{
		Name:     req.Name,
		Mobile:   req.Mobile,
		Password: req.Password,
		Captcha:  req.Captcha,
		Email:    req.Email,
	}
	_, err = LoginServiceClient.Register(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success(""))
}
func (h *HandlerUser) getCaptcha(ctx *gin.Context) {
	result := &common.Result{}
	mobile := ctx.PostForm("mobile")
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	captchaResponse, err := LoginServiceClient.GetCaptcha(c, &login.CaptchaMessage{
		Mobile: mobile,
	})
	fmt.Println(captchaResponse)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}

	ctx.JSON(200, result.Success(captchaResponse.Code))
}

func (u *HandlerUser) login(ctx *gin.Context) {
	//接收参数
	result := &common.Result{}
	var req user.LoginReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}
	//调用RPC
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	loginReponse, err := LoginServiceClient.Login(c, &login.LoginMessage{
		Account:  req.Account,
		Password: req.Password,
	})
	//fmt.Println(captchaResponse)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	rsp := &user.LoginRsp{}
	err = copier.Copy(rsp, loginReponse)
	ctx.JSON(200, result.Success(rsp))
}

func (u *HandlerUser) myOrgList(c *gin.Context) {
	result := &common.Result{}
	token := c.GetHeader("Authorization")
	//验证用户是否已经登录
	mem, err2 := LoginServiceClient.TokenVerify(context.Background(), &login.TokenVerifyMessage{Token: token})
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	list, err2 := LoginServiceClient.MyOrgList(context.Background(), &login.UserMessage{MemId: mem.Member.Id})
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	if list.OrganizationList == nil {
		c.JSON(http.StatusOK, result.Success([]*user.OrganizationList{}))
		return
	}
	var orgs []*user.OrganizationList
	copier.Copy(&orgs, list.OrganizationList)
	c.JSON(http.StatusOK, result.Success(orgs))
}
