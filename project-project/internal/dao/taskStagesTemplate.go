package dao

import (
	"context"
	"project-project/internal/data"
	"project-project/internal/database/gorms"
)

type TaskStagesTemplateDao struct {
	conn *gorms.GormConn
}

func (t *TaskStagesTemplateDao) FindByProjectTemplate(ctx context.Context, projectTemplateCode int64) ([]data.MsTaskStagesTemplate, error) {
	var tsts []data.MsTaskStagesTemplate
	session := t.conn.Default(ctx)
	err := session.Where("project_template_code = ?", projectTemplateCode).Order("sort desc,id asc").Find(&tsts).Error
	return tsts, err
}
func (t *TaskStagesTemplateDao) FindInProTemIds(ctx context.Context, ids []int) ([]data.MsTaskStagesTemplate, error) {
	var tsts []data.MsTaskStagesTemplate
	session := t.conn.Default(ctx)
	err := session.Where("project_template_code in ?", ids).Find(&tsts).Error
	return tsts, err
}

func NewTaskStagesTemplateDao() *TaskStagesTemplateDao {
	return &TaskStagesTemplateDao{
		conn: gorms.New(),
	}
}
