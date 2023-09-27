package project_service_v1

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/errs"
	"project-common/tms"
	project_service_v1 "project-grpc/project"
	"project-grpc/user/login"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/data/menu"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/internal/rpc"
	"project-project/pkg/model"
	"strconv"
	"time"
)

type ProjectService struct {
	project_service_v1.UnimplementedProjectServiceServer
	Logincache             repo.Cache
	Transaction            tran.Transaction
	MenuRepo               repo.MenuRepo
	ProjectRepo            repo.ProjectRepo
	ProjectTemplateRepo    repo.ProjectTemplateRepo
	TaskStagesTemplateRepo repo.TaskStagesTemplateRepo
	TaskStagesRepo         repo.TaskStagesRepo
	ProjectLogRepo         repo.ProjectLogRepo
	TaskRepo               repo.TaskRepo
}

func New() *ProjectService {
	return &ProjectService{
		Logincache:             dao.Rc,
		Transaction:            dao.NewTransactionUser(),
		MenuRepo:               dao.NewMenuDao(),
		ProjectRepo:            dao.NewProjectDao(),
		ProjectTemplateRepo:    dao.NewProjectTemplateDao(),
		TaskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		TaskStagesRepo:         dao.NewTaskStagesDao(),
		ProjectLogRepo:         dao.NewProjectLogDao(),
		TaskRepo:               dao.NewTaskDao(),
	}
}
func (p *ProjectService) Index(ctx context.Context, msg *project_service_v1.IndexMessage) (*project_service_v1.IndexResponse, error) {
	menus, err := p.MenuRepo.FindMenus(context.Background())
	if err != nil {
		zap.L().Error("Index db FindMenus error")
		return nil, errs.GrpcError(model.DBError)
	}
	childs := menu.CovertChild(menus)
	var mms []*project_service_v1.MenuMessage
	copier.Copy(&mms, childs)
	return &project_service_v1.IndexResponse{
		Menus: mms,
	}, nil
}
func (p *ProjectService) FindProjectByMemId(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.MyProjectResponse, error) {
	memberId := msg.MemberId
	page := msg.Page
	pageSize := msg.PageSize
	var pms []*data.ProAndMember
	var total int64
	var err error
	if msg.SelectBy == "" || msg.SelectBy == "my" {
		//TODO 做过改动 BUG优先级Up
		pms, total, err = p.ProjectRepo.FindProjectByMemId(ctx, memberId, page, pageSize, "and deleted = 0")
	}
	if msg.SelectBy == "archive" {
		pms, total, err = p.ProjectRepo.FindProjectByMemId(ctx, memberId, page, pageSize, "and archive = 1")
	}
	if msg.SelectBy == "deleted" {
		pms, total, err = p.ProjectRepo.FindProjectByMemId(ctx, memberId, page, pageSize, "and deleted = 1")
	}
	if msg.SelectBy == "collect" {
		pms, total, err = p.ProjectRepo.FindCollectProjectByMemId(ctx, memberId, page, pageSize)
	}
	//fmt.Println(page)
	//fmt.Println(pageSize)

	if err != nil {
		zap.L().Error("menu findAll error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if pms == nil {
		return &project_service_v1.MyProjectResponse{Pm: []*project_service_v1.ProjectMessage{}, Total: total}, nil
	}
	var pmm []*project_service_v1.ProjectMessage
	copier.Copy(&pmm, pms)
	fmt.Println(pmm)
	for _, v := range pmm {
		v.Code = strconv.FormatInt(v.Id, 10)
		pam := data.ToMap(pms)[v.Id]
		v.AccessControlType = pam.GetAccessControlType()
		//	v.OrganizationCode, _ = encrypts.EncryptInt64(pam.OrganizationCode, model.AESKey)
		v.JoinTime = tms.FormatByMill(pam.JoinTime)
		v.OwnerName = msg.MemberName
		v.Order = int32(pam.Sort)
		v.CreateTime = tms.FormatByMill(pam.CreateTime)
	}
	return &project_service_v1.MyProjectResponse{
		Pm:    pmm,
		Total: total,
	}, nil
}
func (ps *ProjectService) FindProjectTemplate(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.ProjectTemplateResponse, error) {
	//organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	page := msg.Page
	pageSize := msg.PageSize
	var pts []data.ProjectTemplate
	var total int64
	var err error
	if msg.ViewType == -1 {
		pts, total, err = ps.ProjectTemplateRepo.FindProjectTemplateAll(ctx, organizationCode, page, pageSize)
	}
	if msg.ViewType == 1 {
		pts, total, err = ps.ProjectTemplateRepo.FindProjectTemplateSystem(ctx, page, pageSize)
	}
	if msg.ViewType == 0 {
		pts, total, err = ps.ProjectTemplateRepo.FindProjectTemplateCustom(ctx, msg.MemberId, organizationCode, page, pageSize)
	}
	if err != nil {
		zap.L().Error("FindProjectTemplate error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//查询task stages数据库
	tsts, err := ps.TaskStagesTemplateRepo.FindInProTemIds(ctx, data.ToProjectTemplateIds(pts))
	if err != nil {
		zap.L().Error("FindProjectTemplate FindInProTemIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var ptAll []*data.ProjectTemplateAll
	for _, v := range pts {
		ptAll = append(ptAll, v.Convert(data.CovertProjectMap(tsts)[v.Id]))
	}
	var ptRsp []*project_service_v1.ProjectTemplateMessage
	copier.Copy(&ptRsp, ptAll)
	return &project_service_v1.ProjectTemplateResponse{Ptm: ptRsp, Total: total}, nil
}
func (ps *ProjectService) SaveProject(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.SaveProjectMessage, error) {
	//	organizationCodeStr, _ := encrypts.Decrypt(msg.OrganizationCode, model.AESKey)
	organizationCode, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	//	templateCodeStr, _ := encrypts.Decrypt(msg.TemplateCode, model.AESKey)
	//templateCode, _ := strconv.ParseInt(msg.TemplateCode, 10, 64)
	var pr = &data.Project{
		Name:              msg.Name,
		Description:       msg.Description,
		TemplateCode:      msg.TemplateCode,
		CreateTime:        time.Now().UnixMilli(),
		Cover:             "https://img2.baidu.com/it/u=792555388,2449797505&fm=253&fmt=auto&app=138&f=JPEG?w=667&h=500",
		Deleted:           0,
		Archive:           0,
		OrganizationCode:  organizationCode,
		AccessControlType: model.Open,
		TaskBoardTheme:    model.Simple,
	}
	var rsp *project_service_v1.SaveProjectMessage
	err := ps.Transaction.Action(func(conn database.Dbconn) error {
		err := ps.ProjectRepo.SaveProject(ctx, conn, pr)
		if err != nil {
			zap.L().Error("SaveProject Save error", zap.Error(err))
			return model.DBError
		}
		pm := &data.MemberProject{
			ProjectCode: pr.Id,
			MemberCode:  msg.MemberId,
			JoinTime:    time.Now().UnixMilli(),
			IsOwner:     msg.MemberId,
			Authorize:   "",
		}
		err = ps.ProjectRepo.SaveProjectMember(ctx, conn, pm)
		if err != nil {
			zap.L().Error("SaveProject SaveProjectMember error", zap.Error(err))
			return model.DBError
		}
		//	code, _ := encrypts.EncryptInt64(pr.Id, model.AESKey)
		rsp = &project_service_v1.SaveProjectMessage{
			Id:               pr.Id,
			Code:             strconv.FormatInt(pr.Id, 10),
			OrganizationCode: strconv.FormatInt(organizationCode, 10),
			Name:             pr.Name,
			Cover:            pr.Cover,
			CreateTime:       tms.FormatByMill(pr.CreateTime),
			TaskBoardTheme:   pr.TaskBoardTheme,
		}
		//持久化任务步骤
		templateCode, _ := strconv.ParseInt(msg.TemplateCode, 10, 64)
		templates, err := ps.TaskStagesTemplateRepo.FindByProjectTemplate(ctx, templateCode)
		if err != nil {
			zap.L().Error("project SaveProject taskStagesTemplate FindByProjectTemplate error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		for index, v := range templates {
			stages := &data.TaskStages{
				Name:        v.Name,
				Description: "",
				Sort:        index,
				CreateTime:  time.Now().UnixMilli(),
				ProjectCode: pr.Id,
				Deleted:     model.NoDeleted,
			}
			err = ps.TaskStagesRepo.Save(ctx, conn, stages)
			if err != nil {
				zap.L().Error("project SaveProject SaveTaskStages error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
		}
		return nil
	})
	if err != nil {
		return nil, errs.GrpcError(err.(*errs.BError))
	}
	return rsp, nil
}
func (ps *ProjectService) FindProjectDetail(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.ProjectDetailMessage, error) {
	projectCodeStr := msg.ProjectCode
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	memberId := msg.MemberId
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	projectAndMember, err := ps.ProjectRepo.FindProjectByPIdAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindProjectByPIdAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	ownerId := projectAndMember.IsOwner
	member, err := rpc.LoginServiceClient.FindMemInfoById(c, &login.UserMessage{MemId: ownerId})
	if err != nil {
		zap.L().Error("project rpc FindProjectDetail FindMemInfoById error", zap.Error(err))
		return nil, err
	}
	//去user模块去找了
	//TODO 优化 收藏的时候 可以放入redis
	isCollect, err := ps.ProjectRepo.FindCollectByPidAndMemId(c, projectCode, memberId)
	if err != nil {
		zap.L().Error("project FindProjectDetail FindCollectByPidAndMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if isCollect {
		projectAndMember.Collected = model.Collected
	}
	var detailMsg = &project_service_v1.ProjectDetailMessage{}
	idStr := strconv.FormatInt(projectAndMember.Id, 10)
	orgCodestr := strconv.FormatInt(projectAndMember.OrganizationCode, 10)
	copier.Copy(detailMsg, projectAndMember)
	detailMsg.OwnerAvatar = member.Avatar
	detailMsg.OwnerName = member.Name
	detailMsg.Code = idStr
	detailMsg.AccessControlType = projectAndMember.GetAccessControlType()
	detailMsg.OrganizationCode = orgCodestr
	detailMsg.Order = int32(projectAndMember.Sort)
	detailMsg.CreateTime = tms.FormatByMill(projectAndMember.CreateTime)
	return detailMsg, nil
}
func (ps *ProjectService) UpdateDeletedProject(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.DeletedProjectResponse, error) {
	projectCodeStr := msg.ProjectCode
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := ps.ProjectRepo.UpdateDeletedProject(c, projectCode, msg.Deleted)
	if err != nil {
		zap.L().Error("project RecycleProject DeleteProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project_service_v1.DeletedProjectResponse{}, nil
}
func (ps *ProjectService) UpdateCollectProject(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.CollectProjectResponse, error) {
	projectCodeStr := msg.ProjectCode
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var err error
	if "collect" == msg.CollectType {
		pc := &data.ProjectCollection{
			ProjectCode: projectCode,
			MemberCode:  msg.MemberId,
			CreateTime:  time.Now().UnixMilli(),
		}
		err = ps.ProjectRepo.SaveProjectCollect(c, pc)
	}
	if "cancel" == msg.CollectType {
		err = ps.ProjectRepo.DeleteProjectCollect(c, msg.MemberId, projectCode)
	}
	if err != nil {
		zap.L().Error("project UpdateCollectProject SaveProjectCollect error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project_service_v1.CollectProjectResponse{}, nil
}
func (ps *ProjectService) UpdateProject(ctx context.Context, msg *project_service_v1.UpdateProjectMessage) (*project_service_v1.UpdateProjectResponse, error) {
	projectCodeStr := msg.ProjectCode
	fmt.Println(msg.ProjectCode)
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	fmt.Println(projectCode)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	proj := &data.Project{
		Id:                 projectCode,
		Name:               msg.Name,
		Description:        msg.Description,
		Cover:              msg.Cover,
		TaskBoardTheme:     msg.TaskBoardTheme,
		Prefix:             msg.Prefix,
		Private:            int(msg.Private),
		OpenPrefix:         int(msg.OpenPrefix),
		OpenBeginTime:      int(msg.OpenBeginTime),
		OpenTaskPrivate:    int(msg.OpenTaskPrivate),
		Schedule:           msg.Schedule,
		AutoUpdateSchedule: int(msg.AutoUpdateSchedule),
	}
	err := ps.ProjectRepo.UpdateProject(c, proj)
	if err != nil {
		zap.L().Error("project UpdateProject::UpdateProject error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &project_service_v1.UpdateProjectResponse{}, nil
}
func (ps *ProjectService) GetLogBySelfProject(ctx context.Context, msg *project_service_v1.ProjectRpcMessage) (*project_service_v1.ProjectLogResponse, error) {
	//根据用户id查询当前的用户的日志表

	projectLogs, total, err := ps.ProjectLogRepo.FindLogByMemberCode(context.Background(), msg.MemberId, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindLogByMemberCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//查询项目信息
	pIdList := make([]int64, len(projectLogs))
	mIdList := make([]int64, len(projectLogs))
	taskIdList := make([]int64, len(projectLogs))
	for _, v := range projectLogs {
		pIdList = append(pIdList, v.ProjectCode)
		mIdList = append(mIdList, v.MemberCode)
		taskIdList = append(taskIdList, v.SourceCode)
	}
	projects, err := ps.ProjectRepo.FindProjectByIds(context.Background(), pIdList)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindProjectByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pMap := make(map[int64]*data.Project)
	for _, v := range projects {
		pMap[v.Id] = v
	}
	messageList, _ := rpc.LoginServiceClient.FindMemInfoByIds(context.Background(), &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	tasks, err := ps.TaskRepo.FindTaskByIds(context.Background(), taskIdList)
	if err != nil {
		zap.L().Error("project ProjectService::GetLogBySelfProject projectLogRepo.FindTaskByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	tMap := make(map[int64]*data.Task)
	for _, v := range tasks {
		tMap[v.Id] = v
	}
	var list []*data.IndexProjectLogDisplay
	for _, v := range projectLogs {
		display := v.ToIndexDisplay()
		display.ProjectName = pMap[v.ProjectCode].Name
		display.MemberAvatar = mMap[v.MemberCode].Avatar
		display.MemberName = mMap[v.MemberCode].Name
		display.TaskName = tMap[v.SourceCode].Name
		list = append(list, display)
	}
	var msgList []*project_service_v1.ProjectLogMessage
	copier.Copy(&msgList, list)
	return &project_service_v1.ProjectLogResponse{List: msgList, Total: total}, nil
}
