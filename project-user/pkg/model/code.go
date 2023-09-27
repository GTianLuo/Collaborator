package model

import (
	"project-common/errs"
)

var (
	NoLegalMobile             = errs.NewError(2001, "手机号不合法")
	CapthaError               = errs.NewError(2002, "验证码不正确")
	RedisError                = errs.NewError(-100, "redis Error")
	DBError                   = errs.NewError(-200, "DB Error")
	EmailExist                = errs.NewError(2002, "邮箱已经存在")
	NameExist                 = errs.NewError(2003, "用户名已经存在")
	MobileExist               = errs.NewError(2004, "手机号已经存在")
	UnLogin                   = errs.NewError(2006, "未登录")
	Normal                    = 1
	AccountOrPasswordNotRight = errs.NewError(2005, "账号或者密码不正确")
)
