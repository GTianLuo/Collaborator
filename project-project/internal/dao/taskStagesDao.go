package dao

import (
	"context"
	"gorm.io/gorm"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type TaskStagesDao struct {
	conn *gorms.GormConn
}

func (t *TaskStagesDao) FindByProjectCode(ctx context.Context, projectCode int64, page int64, size int64) ([]*data.TaskStages, int64, error) {
	session := t.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	var stages []*data.TaskStages
	err := session.Model(&data.TaskStages{}).Where("project_code=? and deleted=?", readProjectCode, 0).Order("sort asc").Limit(int(size)).Offset(int((page - 1) * size)).Find(&stages).Error
	var total int64
	err = session.Model(&data.TaskStages{}).Where("project_code=?", readProjectCode).Count(&total).Error
	return stages, total, err
}
func (t *TaskStagesDao) Save(ctx context.Context, conn database.Dbconn, stages *data.TaskStages) error {
	t.conn = conn.(*gorms.GormConn)
	session := t.conn.Tran(ctx)
	return session.Save(&stages).Error
}
func (t *TaskStagesDao) FindById(ctx context.Context, stageCode int) (ts *data.TaskStages, err error) {
	err = t.conn.Default(ctx).Where("id=?", stageCode).Find(&ts).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return
}

func NewTaskStagesDao() *TaskStagesDao {
	return &TaskStagesDao{
		conn: gorms.New(),
	}
}
