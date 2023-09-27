package repo

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database"
)

type TaskStagesTemplateRepo interface {
	FindInProTemIds(ctx context.Context, ids []int) ([]data.MsTaskStagesTemplate, error)
	FindByProjectTemplate(ctx context.Context, projectTemplateCode int64) ([]data.MsTaskStagesTemplate, error)
}
type TaskStagesRepo interface {
	Save(ctx context.Context, conn database.Dbconn, stages *data.TaskStages) error
	FindByProjectCode(ctx context.Context, projectCode int64, page int64, size int64) ([]*data.TaskStages, int64, error)
	FindById(ctx context.Context, stageCode int) (*data.TaskStages, error)
}

type TaskRepo interface {
	FindTaskByStageCode(ctx context.Context, stageCode int) ([]*data.Task, error)
	FindTaskMaxIdNum(ctx context.Context, projectCode int64) (int64, error)
	FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (int64, error)
	SaveTask(ctx context.Context, conn database.Dbconn, ts *data.Task) error
	SaveTaskMember(ctx context.Context, conn database.Dbconn, tm *data.TaskMember) error
	FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberCode int64) (*data.TaskMember, error)
	//	FindTaskByCreateBy(ctx context.Context, memberId int64, done int) (tList []*task.Task, total int64, err error)
	FindTaskByAssignTo(ctx context.Context, memberId int64, done int) ([]*data.Task, int64, error)
	FindTaskByMemberCode(ctx context.Context, memberId int64, done int) (tList []*data.Task, total int64, err error)
	FindTaskByCreateBy(ctx context.Context, memberId int64, done int) (tList []*data.Task, total int64, err error)
	FindTaskById(ctx context.Context, taskCode int64) (*data.Task, error)
	UpdateTaskSort(ctx context.Context, conn database.Dbconn, ts *data.Task) error
	FindTaskMemberPage(ctx context.Context, taskCode int64, page int64, size int64) (list []*data.TaskMember, total int64, err error)
	FindTaskByIds(background context.Context, taskIdList []int64) (list []*data.Task, err error)
}
