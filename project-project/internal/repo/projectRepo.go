package repo

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database"
)

type ProjectRepo interface {
	FindProjectByMemId(ctx context.Context, memId int64, page int64, size int64, condition string) ([]*data.ProAndMember, int64, error)
	FindCollectProjectByMemId(ctx context.Context, id int64, page int64, size int64) ([]*data.ProAndMember, int64, error)
	SaveProject(ctx context.Context, conn database.Dbconn, pr *data.Project) error
	SaveProjectMember(ctx context.Context, conn database.Dbconn, pm *data.MemberProject) error
	FindProjectByPIdAndMemId(ctx context.Context, projectCode int64, id int64) (*data.ProAndMember, error)
	FindCollectByPidAndMemId(ctx context.Context, projectCode int64, id int64) (bool, error)
	UpdateDeletedProject(ctx context.Context, code int64, deleted bool) error
	SaveProjectCollect(ctx context.Context, pc *data.ProjectCollection) error
	DeleteProjectCollect(ctx context.Context, mem int64, projectCode int64) error
	UpdateProject(ctx context.Context, proj *data.Project) error
	FindMemberInfoByProjectCode(ctx context.Context, project int64, page int64, size int64) ([]*data.ProjectMemberInfo, int64, error)
	FindProjectById(ctx context.Context, id int64) (pj *data.Project, err error)
	FindProjectByIds(ctx context.Context, pids []int64) (list []*data.Project, err error)
	//FindProjectTemplateAll(ctx context.Context, organizationCode int64, page int64, size int64) ([]project.ProjectTemplate, int64, error)
}
