package task_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"project-common/errs"
	"project-common/tms"
	task_service_v1 "project-grpc/task"
	"project-grpc/user/login"
	"project-project/internal/dao"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/tran"
	"project-project/internal/repo"
	"project-project/internal/rpc"
	"project-project/pkg/model"
	"strconv"
	"time"
)

type TaskService struct {
	task_service_v1.UnimplementedTaskServiceServer
	cache                  repo.Cache
	transaction            tran.Transaction
	taskStagesTemplateRepo repo.TaskStagesTemplateRepo
	taskStagesRepo         repo.TaskStagesRepo
	projectRepo            repo.ProjectRepo
	taskRepo               repo.TaskRepo
	projectLogRepo         repo.ProjectLogRepo
	taskWorkTimeRepo       repo.TaskWorkTimeRepo
	fileRepo               repo.FileRepo
	sourceLinkRepo         repo.SourceLinkRepo
}

func New() *TaskService {
	return &TaskService{
		cache:                  dao.Rc,
		transaction:            dao.NewTransactionUser(),
		taskStagesTemplateRepo: dao.NewTaskStagesTemplateDao(),
		taskStagesRepo:         dao.NewTaskStagesDao(),
		projectRepo:            dao.NewProjectDao(),
		taskRepo:               dao.NewTaskDao(),
		projectLogRepo:         dao.NewProjectLogDao(),
		taskWorkTimeRepo:       dao.NewTaskWorkTimeDao(),
		fileRepo:               dao.NewFileDao(),
		sourceLinkRepo:         dao.NewSourceLinkDao(),
	}
}

func (t *TaskService) TaskStages(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskStagesResponse, error) {
	projectCodeStr := msg.ProjectCode
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	page := msg.Page
	pageSize := msg.PageSize
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskStages, total, err := t.taskStagesRepo.FindByProjectCode(c, projectCode, page, pageSize)
	if err != nil {
		zap.L().Error("project task TaskStages FindByProjectCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	tsMap := data.ToTaskStagesMap(taskStages)
	var resp []*task_service_v1.TaskStagesMessage
	copier.Copy(&resp, taskStages)
	for _, v := range resp {
		stages := tsMap[int(v.Id)]
		v.Code = string(v.Id)
		v.CreateTime = tms.FormatByMill(stages.CreateTime)
		v.ProjectCode = msg.ProjectCode
	}
	return &task_service_v1.TaskStagesResponse{
		List:  resp,
		Total: total,
	}, nil
}
func (t *TaskService) MemberProjectList(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.MemberProjectResponse, error) {
	projectCodeStr := msg.ProjectCode
	projectCode, _ := strconv.ParseInt(projectCodeStr, 10, 64)
	page := msg.Page
	pageSize := msg.PageSize
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberInfos, total, err := t.projectRepo.FindMemberInfoByProjectCode(c, projectCode, page, pageSize)
	if err != nil {
		zap.L().Error("project task TaskStages FindByProjectCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pmMap := data.ToProjectMemberInfoMap(memberInfos)
	var resp []*task_service_v1.MemberProjectMessage
	copier.Copy(&resp, memberInfos)
	for _, v := range resp {
		pm := pmMap[v.MemberCode]
		v.Code = strconv.FormatInt(v.MemberCode, 10)
		if pm.MemberCode == pm.IsOwner {
			v.IsOwner = 1
		} else {
			v.IsOwner = 0
		}
	}
	return &task_service_v1.MemberProjectResponse{
		List:  resp,
		Total: total,
	}, nil
}
func (t *TaskService) TaskList(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskListResponse, error) {
	stageCode, _ := strconv.ParseInt(msg.StageCode, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskList, err := t.taskRepo.FindTaskByStageCode(c, int(stageCode))
	if err != nil {
		zap.L().Error("project task TaskList FindTaskByStageCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var taskDisplayList []*data.TaskDisplay
	for _, v := range taskList {
		display := v.ToTaskDisplay()
		tm, err := t.taskRepo.FindTaskMemberByTaskId(ctx, v.Id, msg.MemberId)
		if err != nil {
			zap.L().Error("project task TaskList FindTaskMemberByTaskId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		if tm == nil {
			display.CanRead = model.NoCanRead
		} else {
			display.CanRead = model.CanRead
		}
		taskDisplayList = append(taskDisplayList, display)
	}
	var taskMessageList []*task_service_v1.TaskMessage
	copier.Copy(&taskMessageList, taskDisplayList)
	return &task_service_v1.TaskListResponse{List: taskMessageList}, nil
}
func (t *TaskService) SaveTask(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskMessage, error) {
	//先检查业务
	if msg.Name == "" {
		return nil, errs.GrpcError(model.TaskNameNotNull)
	}
	stageCode := msg.StageCode
	stageCodeInt, _ := strconv.ParseInt(stageCode, 10, 64)
	taskStages, err := t.taskStagesRepo.FindById(ctx, int(stageCodeInt))
	if err != nil {
		zap.L().Error("project task SaveTask taskStagesRepo.FindById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskStages == nil {
		return nil, errs.GrpcError(model.TaskStagesNotNull)
	}
	projectCode := msg.ProjectCode
	projectCodeInt, _ := strconv.ParseInt(projectCode, 10, 64)
	project, err := t.projectRepo.FindProjectById(ctx, projectCodeInt)
	if err != nil {
		zap.L().Error("project task SaveTask projectRepo.FindProjectById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if project.Deleted == model.Deleted {
		return nil, errs.GrpcError(model.ProjectAlreadyDeleted)
	}

	maxIdNum, err := t.taskRepo.FindTaskMaxIdNum(ctx, projectCodeInt)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskMaxIdNum error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	maxSort, err := t.taskRepo.FindTaskSort(ctx, projectCodeInt, stageCodeInt)
	if err != nil {
		zap.L().Error("project task SaveTask taskRepo.FindTaskSort error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	assignTo := msg.AssignTo
	assignToInt, _ := strconv.ParseInt(assignTo, 10, 64)
	var ts = &data.Task{
		Name:        msg.Name,
		CreateTime:  time.Now().UnixMilli(),
		CreateBy:    msg.MemberId,
		AssignTo:    assignToInt,
		ProjectCode: projectCodeInt,
		StageCode:   int(stageCodeInt),
		IdNum:       int(maxIdNum + 1),
		Private:     project.OpenTaskPrivate,
		Sort:        int(maxSort + 1),
		BeginTime:   time.Now().UnixMilli(),
		EndTime:     time.Now().Add(2 * 24 * time.Hour).UnixMilli(),
	}
	err = t.transaction.Action(func(conn database.Dbconn) error {
		err = t.taskRepo.SaveTask(ctx, conn, ts)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTask error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		tm := &data.TaskMember{
			MemberCode: msg.MemberId,
			TaskCode:   ts.Id,
			IsExecutor: model.Executor,
			JoinTime:   time.Now().UnixMilli(),
			IsOwner:    model.Owner,
		}
		err = t.taskRepo.SaveTaskMember(ctx, conn, tm)
		if err != nil {
			zap.L().Error("project task SaveTask taskRepo.SaveTaskMember error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	display := ts.ToTaskDisplay()
	tm := &task_service_v1.TaskMessage{}
	copier.Copy(tm, display)
	return tm, nil
}
func (t *TaskService) TaskSort(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskSortResponse, error) {
	preTaskCode := msg.PreTaskCode
	preTaskCodeInt, _ := strconv.ParseInt(preTaskCode, 10, 64)
	stageCode := msg.StageCode
	stageCodeInt, _ := strconv.ParseInt(stageCode, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ts, err := t.taskRepo.FindTaskById(c, preTaskCodeInt)
	if err != nil {
		zap.L().Error("project task TaskSort taskRepo.FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	isChange := false
	if ts.StageCode != int(stageCodeInt) {
		//任务步骤变化 移动到其他步骤
		ts.StageCode = int(stageCodeInt)
		isChange = true
	}
	err = t.transaction.Action(func(conn database.Dbconn) error {
		nextTaskCode := msg.NextTaskCode
		if nextTaskCode != "" {
			//顺序变了 需要互换位置
			nextTaskId, _ := strconv.ParseInt(nextTaskCode, 10, 64)
			nextTs, err := t.taskRepo.FindTaskById(c, nextTaskId)
			if err != nil {
				zap.L().Error("project task TaskSort nextTaskId taskRepo.FindTaskById error", zap.Error(err))
				return errs.GrpcError(model.DBError)
			}
			sort := ts.Sort
			ts.Sort = nextTs.Sort
			nextTs.Sort = sort
			isChange = true
			err = t.taskRepo.UpdateTaskSort(ctx, conn, nextTs)
			if err != nil {
				return err
			}
		}
		if isChange {
			err := t.taskRepo.UpdateTaskSort(ctx, conn, ts)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &task_service_v1.TaskSortResponse{}, nil
}
func (t *TaskService) MyTaskList(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.MyTaskListResponse, error) {
	var tsList []*data.Task
	var err error
	var total int64
	if msg.TaskType == 1 {
		//我执行的
		tsList, total, err = t.taskRepo.FindTaskByAssignTo(ctx, msg.MemberId, int(msg.Type))
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByAssignTo error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if msg.TaskType == 2 {
		//我参与的
		tsList, total, err = t.taskRepo.FindTaskByMemberCode(ctx, msg.MemberId, int(msg.Type))
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByMemberCode error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if msg.TaskType == 3 {
		//我创建的
		tsList, total, err = t.taskRepo.FindTaskByCreateBy(ctx, msg.MemberId, int(msg.Type))
		if err != nil {
			zap.L().Error("project task MyTaskList taskRepo.FindTaskByCreateBy error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
	}
	if tsList == nil || len(tsList) <= 0 {
		return &task_service_v1.MyTaskListResponse{List: nil, Total: 0}, nil
	}
	var pids []int64
	var mids []int64
	for _, v := range tsList {
		pids = append(pids, v.ProjectCode)
		mids = append(mids, v.AssignTo)
	}
	pList, err := t.projectRepo.FindProjectByIds(ctx, pids)
	projectMap := data.ToProjectMap(pList)

	mList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{
		MIds: mids,
	})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range mList.List {
		mMap[v.Id] = v
	}
	var mtdList []*data.MyTaskDisplay
	for _, v := range tsList {
		memberMessage := mMap[v.AssignTo]
		name := memberMessage.Name
		avatar := memberMessage.Avatar
		mtd := v.ToMyTaskDisplay(projectMap[v.ProjectCode], name, avatar)
		mtdList = append(mtdList, mtd)
	}
	var myMsgs []*task_service_v1.MyTaskMessage
	copier.Copy(&myMsgs, mtdList)
	return &task_service_v1.MyTaskListResponse{List: myMsgs, Total: total}, nil
}
func (t *TaskService) ReadTask(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskMessage, error) {
	taskCode, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskInfo, err := t.taskRepo.FindTaskById(c, taskCode)
	if err != nil {
		zap.L().Error("project task ReadTask taskRepo FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if taskInfo == nil {
		return &task_service_v1.TaskMessage{}, nil
	}
	display := taskInfo.ToTaskDisplay()
	if taskInfo.Private == 1 {
		//代表隐私模式
		taskMember, err := t.taskRepo.FindTaskMemberByTaskId(ctx, taskInfo.Id, msg.MemberId)
		if err != nil {
			zap.L().Error("project task TaskList taskRepo.FindTaskMemberByTaskId error", zap.Error(err))
			return nil, errs.GrpcError(model.DBError)
		}
		if taskMember != nil {
			display.CanRead = model.CanRead
		} else {
			display.CanRead = model.NoCanRead
		}
	}
	pj, err := t.projectRepo.FindProjectById(c, taskInfo.ProjectCode)
	display.ProjectName = pj.Name
	taskStages, err := t.taskStagesRepo.FindById(c, taskInfo.StageCode)
	display.StageName = taskStages.Name
	// in ()
	memberMessage, err := rpc.LoginServiceClient.FindMemInfoById(ctx, &login.UserMessage{MemId: taskInfo.AssignTo})
	if err != nil {
		zap.L().Error("project task TaskList LoginServiceClient.FindMemInfoById error", zap.Error(err))
		return nil, err
	}
	e := data.Executor{
		Name:   memberMessage.Name,
		Avatar: memberMessage.Avatar,
	}
	display.Executor = e
	var taskMessage = &task_service_v1.TaskMessage{}
	copier.Copy(taskMessage, display)
	return taskMessage, nil
}
func (t *TaskService) ListTaskMember(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskMemberList, error) {
	taskCode := msg.TaskCode
	taskCodeInt, _ := strconv.ParseInt(taskCode, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	taskMemberPage, total, err := t.taskRepo.FindTaskMemberPage(c, taskCodeInt, msg.Page, msg.PageSize)
	if err != nil {
		zap.L().Error("project task TaskList taskRepo.FindTaskMemberPage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var mids []int64
	for _, v := range taskMemberPage {
		mids = append(mids, v.MemberCode)
	}
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(ctx, &login.UserMessage{MIds: mids})
	mMap := make(map[int64]*login.MemberMessage, len(messageList.List))
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	var taskMemeberMemssages []*task_service_v1.TaskMemberMessage
	for _, v := range taskMemberPage {
		tm := &task_service_v1.TaskMemberMessage{}
		tm.Code = strconv.FormatInt(v.MemberCode, 10)
		tm.Id = v.Id
		message := mMap[v.MemberCode]
		tm.Name = message.Name
		tm.Avatar = message.Avatar
		tm.IsExecutor = int32(v.IsExecutor)
		tm.IsOwner = int32(v.IsOwner)
		taskMemeberMemssages = append(taskMemeberMemssages, tm)
	}
	return &task_service_v1.TaskMemberList{List: taskMemeberMemssages, Total: total}, nil
}
func (t *TaskService) TaskLog(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskLogList, error) {
	taskCode, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	all := msg.All
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var list []*data.ProjectLog
	var total int64
	var err error
	if all == 1 {
		//显示全部
		list, total, err = t.projectLogRepo.FindLogByTaskCode(c, taskCode, int(msg.Comment))
	}
	if all == 0 {
		//分页
		list, total, err = t.projectLogRepo.FindLogByTaskCodePage(c, taskCode, int(msg.Comment), int(msg.Page), int(msg.PageSize))
	}
	if err != nil {
		zap.L().Error("project task TaskLog projectLogRepo.FindLogByTaskCodePage error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if total == 0 {
		return &task_service_v1.TaskLogList{}, nil
	}
	var displayList []*data.ProjectLogDisplay
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(c, &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*task_service_v1.TaskLog
	copier.Copy(&l, displayList)
	return &task_service_v1.TaskLogList{List: l, Total: total}, nil
}
func (t *TaskService) TaskWorkTimeList(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskWorkTimeResponse, error) {
	taskCode, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var list []*data.TaskWorkTime
	var err error
	list, err = t.taskWorkTimeRepo.FindWorkTimeList(c, taskCode)
	if err != nil {
		zap.L().Error("project task TaskWorkTimeList taskWorkTimeRepo.FindWorkTimeList error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(list) == 0 {
		return &task_service_v1.TaskWorkTimeResponse{}, nil
	}
	var displayList []*data.TaskWorkTimeDisplay
	var mIdList []int64
	for _, v := range list {
		mIdList = append(mIdList, v.MemberCode)
	}
	messageList, err := rpc.LoginServiceClient.FindMemInfoByIds(c, &login.UserMessage{MIds: mIdList})
	mMap := make(map[int64]*login.MemberMessage)
	for _, v := range messageList.List {
		mMap[v.Id] = v
	}
	for _, v := range list {
		display := v.ToDisplay()
		message := mMap[v.MemberCode]
		m := data.Member{}
		m.Name = message.Name
		m.Id = message.Id
		m.Avatar = message.Avatar
		m.Code = message.Code
		display.Member = m
		displayList = append(displayList, display)
	}
	var l []*task_service_v1.TaskWorkTime
	copier.Copy(&l, displayList)
	return &task_service_v1.TaskWorkTimeResponse{List: l, Total: int64(len(l))}, nil
}
func (t *TaskService) SaveTaskWorkTime(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.SaveTaskWorkTimeResponse, error) {
	tmt := &data.TaskWorkTime{}
	tmt.BeginTime = msg.BeginTime
	tmt.Num = int(msg.Num)
	tmt.Content = msg.Content
	taskCodeInt, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	tmt.TaskCode = taskCodeInt
	tmt.MemberCode = msg.MemberId
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := t.taskWorkTimeRepo.Save(c, tmt)
	if err != nil {
		zap.L().Error("project task SaveTaskWorkTime taskWorkTimeRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task_service_v1.SaveTaskWorkTimeResponse{}, nil
}
func (t *TaskService) SaveTaskFile(ctx context.Context, msg *task_service_v1.TaskFileReqMessage) (*task_service_v1.TaskFileResponse, error) {
	taskCode, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	//存file表
	OrganizationCodeInt, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	TaskCodeInt, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	ProjectCodeInt, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	var f = &data.File{
		PathName:         msg.PathName,
		Title:            msg.FileName,
		Extension:        msg.Extension,
		Size:             int(msg.Size),
		ObjectType:       "",
		OrganizationCode: OrganizationCodeInt,
		TaskCode:         TaskCodeInt,
		ProjectCode:      ProjectCodeInt,
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Downloads:        0,
		Extra:            "",
		Deleted:          model.NoDeleted,
		FileType:         msg.FileType,
		FileUrl:          msg.FileUrl,
		DeletedTime:      0,
	}
	err := t.fileRepo.Save(context.Background(), f)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	//存入source_link
	sl := &data.SourceLink{
		SourceType:       "file",
		SourceCode:       f.Id,
		LinkType:         "task",
		LinkCode:         taskCode,
		OrganizationCode: OrganizationCodeInt,
		CreateBy:         msg.MemberId,
		CreateTime:       time.Now().UnixMilli(),
		Sort:             0,
	}
	err = t.sourceLinkRepo.Save(context.Background(), sl)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.Save error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	return &task_service_v1.TaskFileResponse{}, nil
}
func (t *TaskService) TaskSources(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.TaskSourceResponse, error) {
	taskCode, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	sourceLinks, err := t.sourceLinkRepo.FindByTaskCode(context.Background(), taskCode)
	if err != nil {
		zap.L().Error("project task SaveTaskFile sourceLinkRepo.FindByTaskCode error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(sourceLinks) == 0 {
		return &task_service_v1.TaskSourceResponse{}, nil
	}
	var fIdList []int64
	for _, v := range sourceLinks {
		fIdList = append(fIdList, v.SourceCode)
	}
	files, err := t.fileRepo.FindByIds(context.Background(), fIdList)
	if err != nil {
		zap.L().Error("project task SaveTaskFile fileRepo.FindByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	fMap := make(map[int64]*data.File)
	for _, v := range files {
		fMap[v.Id] = v
	}
	var list []*data.SourceLinkDisplay
	for _, v := range sourceLinks {
		list = append(list, v.ToDisplay(fMap[v.SourceCode]))
	}
	var slMsg []*task_service_v1.TaskSourceMessage
	copier.Copy(&slMsg, list)
	return &task_service_v1.TaskSourceResponse{List: slMsg}, nil
}
func (t *TaskService) CreateComment(ctx context.Context, msg *task_service_v1.TaskReqMessage) (*task_service_v1.CreateCommentResponse, error) {
	taskCode, _ := strconv.ParseInt(msg.TaskCode, 10, 64)
	taskById, err := t.taskRepo.FindTaskById(context.Background(), taskCode)
	if err != nil {
		zap.L().Error("project task CreateComment fileRepo.FindTaskById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	pl := &data.ProjectLog{
		MemberCode:   msg.MemberId,
		Content:      msg.CommentContent,
		Remark:       msg.CommentContent,
		Type:         "createComment",
		CreateTime:   time.Now().UnixMilli(),
		SourceCode:   taskCode,
		ActionType:   "task",
		ToMemberCode: 0,
		IsComment:    model.Comment,
		ProjectCode:  taskById.ProjectCode,
		Icon:         "plus",
		IsRobot:      0,
	}
	t.projectLogRepo.SaveProjectLog(pl)
	return &task_service_v1.CreateCommentResponse{}, nil
}
