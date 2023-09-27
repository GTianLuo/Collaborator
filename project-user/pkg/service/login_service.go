package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"log"
	common "project-common"
	"project-common/encrypts"
	"project-common/errs"
	"project-common/jwts"
	"project-common/tms"
	"project-grpc/user/login"
	"strconv"
	"strings"
	"test.com/project-user/config"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/data/organization"
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/tran"
	"test.com/project-user/internal/repo"
	"test.com/project-user/pkg/model"
	"time"
)

type LoginService struct {
	login.UnimplementedLoginServiceServer
	Logincache       repo.Cache
	LoginmemberRepo  repo.MemberRepo
	OrganizationRepo repo.OrganizationRepo
	Transaction      tran.Transaction
}

func New() *LoginService {
	return &LoginService{
		Logincache:       dao.Rc,
		LoginmemberRepo:  dao.NewMemberDao(),
		OrganizationRepo: dao.NewOrganizationDao(),
		Transaction:      dao.NewTransactionUser(),
	}
}

func (ls *LoginService) GetCaptcha(ctx context.Context, msg *login.CaptchaMessage) (*login.CaptchaResponse, error) {
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
		err := ls.Logincache.Put(c, "REGISTER_"+mobile, code, 15*time.Minute)
		if err != nil {
			zap.L().Error("验证码存入redis出错，")
			fmt.Println(err)
		}
		log.Println("将手机号和验证码存入redis成功 REGISTER_%S : %S", mobile, code)
	}()
	return &login.CaptchaResponse{Code: code}, nil
}
func (ls *LoginService) Register(ctx context.Context, msg *login.RegisterMessage) (*login.RegisterResponse, error) {
	redisCode, err := ls.Logincache.Get(context.Background(), model.RegisterRedisKey+msg.Mobile)
	if err != nil {
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.RedisError)
	}
	if redisCode != msg.Captcha {
		return nil, errs.GrpcError(model.CapthaError)
	}
	exist, err := ls.LoginmemberRepo.GetMemberByEmail(ctx, msg.Email)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.EmailExist)
	}
	exist2, err := ls.LoginmemberRepo.GetMemberByAccount(ctx, msg.Name)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exist2 {
		return nil, errs.GrpcError(model.NameExist)
	}
	exist3, err := ls.LoginmemberRepo.GetMemberByMobile(ctx, msg.Mobile)
	if err != nil {
		return nil, errs.GrpcError(model.DBError)
	}
	if exist3 {
		return nil, errs.GrpcError(model.MobileExist)
	}
	pwd := encrypts.Md5(msg.Password)
	//事务开始执行
	err = ls.Transaction.Action(func(conn database.Dbconn) error {
		mem := &member.Member{
			Account:       msg.Name,
			Password:      pwd,
			Name:          msg.Name,
			Mobile:        msg.Mobile,
			Email:         msg.Email,
			CreateTime:    time.Now().UnixMilli(),
			LastLoginTime: time.Now().UnixMilli(),
			Status:        model.Normal,
		}
		err = ls.LoginmemberRepo.SaveMember(conn, ctx, mem)
		if err != nil {
			return errs.GrpcError(model.DBError)
		}
		org := &organization.Organization{
			Name:       mem.Name + "个人项目",
			MemberId:   mem.Id,
			CreateTime: time.Now().UnixMilli(),
			Personal:   1,
			Avatar:     "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fc-ssl.dtstatic.com%2Fuploads%2Fblog%2F202103%2F31%2F20210331160001_9a852.thumb.1000_0.jpg&refer=http%3A%2F%2Fc-ssl.dtstatic.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1673017724&t=ced22fc74624e6940fd6a89a21d30cc5",
		}
		err = ls.OrganizationRepo.SaveOrganization(conn, ctx, org)
		if err != nil {
			zap.L().Error("register SaveOrganization db err", zap.Error(err))
			return model.DBError
		}
		return nil
	})
	return &login.RegisterResponse{}, err
}
func (ls *LoginService) Login(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	if ctx == nil {
		zap.L().Error("ctx is nil")
	}
	//fmt.Printf("%v sdojiawdjo", ctx)
	//查询账号密码是否正确
	findMember, err := ls.LoginmemberRepo.FindMember(ctx, msg.Account)
	if err != nil {
		zap.L().Error("login FindMember db err", zap.Error(err))
		return nil, model.DBError
	}
	if findMember.Password != encrypts.Md5(msg.Password) {
		return nil, model.AccountOrPasswordNotRight
	}
	memMsg := &login.MemberMessage{}
	err = copier.Copy(&memMsg, findMember)
	memMsg.Code = strconv.FormatInt(findMember.Id, 10)
	memMsg.LastLoginTime = tms.FormatByMill(findMember.LastLoginTime)
	//根据用户id查询组织
	orgs, err := ls.OrganizationRepo.FindOrganizationByMemId(ctx, findMember.Id)
	if err != nil {
		zap.L().Error("login FindOrg db err", zap.Error(err))
		return nil, model.DBError
	}
	var orgsMessage []*login.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	for _, org := range orgsMessage {
		org.Code = strconv.FormatInt(org.Id, 10)
		org.OwnerCode = memMsg.Code
		org.CreateTime = tms.FormatByMill(organization.ToMap(orgs)[org.Id].CreateTime)
	}
	//用jwt生成token，返回
	accessExp := time.Duration(config.AppConf.JwtConfig.AccessExp*3600*24) * time.Second
	refreExp := time.Duration(config.AppConf.JwtConfig.RefreshExp*3600*24) * time.Second
	token := jwts.CreateToken(strconv.FormatInt(findMember.Id, 10), config.AppConf.JwtConfig.AccessSecret, accessExp, refreExp)
	tokenList := &login.TokenMessage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		AccessTokenExp: int64(token.AccessExp),
		TokenType:      "bearer",
	}
	go func(ctx context.Context) {
		err := ls.Logincache.Put(ctx, "test2", "test", time.Second*120).Error()
		zap.L().Error(err)
		marshal, _ := json.Marshal(findMember)
		ls.Logincache.Put(context.Background(), model.Member+"::"+strconv.FormatInt(findMember.Id, 10), string(marshal), accessExp)
		orgsJson, _ := json.Marshal(orgs)
		ls.Logincache.Put(context.Background(), model.MemberOrganization+"::"+strconv.FormatInt(findMember.Id, 10), string(orgsJson), accessExp)
	}(ctx)
	if len(orgs) > 0 {
		memMsg.OrganizationCode = strconv.FormatInt(orgs[0].Id, 10)
	}
	return &login.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgsMessage,
		TokenList:        tokenList,
	}, nil
}
func (ls *LoginService) TokenVerify(ctx context.Context, msg *login.TokenVerifyMessage) (*login.TokenVerifyResponse, error) {
	token := msg.Token
	println(token)
	if strings.Contains(token, "bearer") {
		token = strings.ReplaceAll(token, "bearer ", "")
	}
	tokenKey, err := jwts.ParseToken(token, config.AppConf.JwtConfig.AccessSecret)
	if err != nil {
		zap.L().Error("login TokenVerify Error", zap.Error(err))
		return nil, errs.GrpcError(model.UnLogin)
	}
	memJson, err := ls.Logincache.Get(context.Background(), model.Member+"::"+tokenKey)
	if err != nil {
		zap.L().Error("login TokenVerify cache get member error Error", zap.Error(err))
		return nil, errs.GrpcError(model.UnLogin)
	}
	if memJson == "" {
		zap.L().Error("login TokenVerify cache get member expire Error", zap.Error(err))
		return nil, errs.GrpcError(model.UnLogin)
	}
	memberById := &member.Member{}
	json.Unmarshal([]byte(memJson), memberById)
	//下行查询数据库 后来还是放弃了 因为登录请求 没登陆重新登陆即可 不应在每次token检查的时候都要访问数据库资源
	/*id, _ := strconv.ParseInt(tokenKey, 10, 64)
	byId, err := ls.LoginmemberRepo.FindMemberById(ctx, id)
	if err != nil {
		zap.L().Error("login FindMemberById db err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	orgs, err := ls.OrganizationRepo.FindOrganizationByMemId(ctx, byId.Id)
	if err != nil {
		zap.L().Error("login FindOrg db err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}*/
	orgsJson, err := ls.Logincache.Get(context.Background(), model.MemberOrganization+"::"+tokenKey)
	var orgs []*organization.Organization
	json.Unmarshal([]byte(orgsJson), orgs)
	memMsg := &login.MemberMessage{}
	copier.Copy(memMsg, memberById)
	if len(orgs) > 0 {
		memMsg.OrganizationCode = strconv.FormatInt(orgs[0].Id, 10)
	}
	return &login.TokenVerifyResponse{Member: memMsg}, nil
}
func (l *LoginService) MyOrgList(ctx context.Context, msg *login.UserMessage) (*login.OrgListResponse, error) {
	memId := msg.MemId
	orgs, err := l.OrganizationRepo.FindOrganizationByMemId(ctx, memId)
	if err != nil {
		zap.L().Error("MyOrgList FindOrganizationByMemId err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgsMessage []*login.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	//for _, org := range orgsMessage {
	//	org., _ = encrypts.EncryptInt64(org.Id, model.AESKey)
	//}
	return &login.OrgListResponse{OrganizationList: orgsMessage}, nil
}
func (ls *LoginService) FindMemInfoById(ctx context.Context, msg *login.UserMessage) (*login.MemberMessage, error) {
	memberById, err := ls.LoginmemberRepo.FindMemberById(context.Background(), msg.MemId)
	if err != nil {
		zap.L().Error("TokenVerify db FindMemberById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	memMsg := &login.MemberMessage{}
	copier.Copy(memMsg, memberById)
	memMsg.Code = strconv.FormatInt(memMsg.Id, 10)
	orgs, err := ls.OrganizationRepo.FindOrganizationByMemId(context.Background(), memberById.Id)
	if err != nil {
		zap.L().Error("TokenVerify db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode = strconv.FormatInt(orgs[0].Id, 10)
	}
	memMsg.CreateTime = tms.FormatByMill(memberById.CreateTime)
	return memMsg, nil
}
func (ls *LoginService) FindMemInfoByIds(ctx context.Context, msg *login.UserMessage) (*login.MemberMessageList, error) {
	memberList, err := ls.LoginmemberRepo.FindMemberByIds(context.Background(), msg.MIds)
	if err != nil {
		zap.L().Error("FindMemInfoByIds db memberRepo.FindMemberByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if memberList == nil || len(memberList) <= 0 {
		return &login.MemberMessageList{List: nil}, nil
	}
	mMap := make(map[int64]*member.Member)
	for _, v := range memberList {
		mMap[v.Id] = v
	}
	var memMsgs []*login.MemberMessage
	copier.Copy(&memMsgs, memberList)
	for _, v := range memMsgs {
		m := mMap[v.Id]
		v.CreateTime = tms.FormatByMill(m.CreateTime)
		v.Code = strconv.FormatInt(v.Id, 10)
	}

	return &login.MemberMessageList{List: memMsgs}, nil
}
