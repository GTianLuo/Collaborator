package dao

import (
	"context"
	"gorm.io/gorm"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
)

type TaskDao struct {
	conn *gorms.GormConn
}

func (t *TaskDao) FindTaskMemberByTaskId(ctx context.Context, taskCode int64, memberCode int64) (*data.TaskMember, error) {
	var tm *data.TaskMember
	err := t.conn.Default(ctx).Where("task_code=? and member_code=?", taskCode, memberCode).Limit(1).Find(&tm).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return tm, err
}

func (t *TaskDao) SaveTaskMember(ctx context.Context, conn database.Dbconn, tm *data.TaskMember) error {
	t.conn = conn.(*gorms.GormConn)
	return t.conn.Tran(ctx).Save(&tm).Error
}

func (t *TaskDao) SaveTask(ctx context.Context, conn database.Dbconn, ts *data.Task) error {
	t.conn = conn.(*gorms.GormConn)
	session := t.conn.Tran(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", ts.ProjectCode)
	var readProjectCode int64
	db2.Scan(&readProjectCode)
	ts.ProjectCode = readProjectCode
	return session.Save(&ts).Error
}

func (t *TaskDao) FindTaskSort(ctx context.Context, projectCode int64, stageCode int64) (v int64, err error) {
	session := t.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	m := make(map[string]*int64)
	err = session.Model(&data.Task{}).Where("project_code=? and stage_code=?", readProjectCode, stageCode).Select("max(sort) as sort").Take(&m).Error
	if m["sort"] == nil {
		return 0, nil
	}
	v = *m["sort"]
	return
}

func (t *TaskDao) FindTaskMaxIdNum(ctx context.Context, projectCode int64) (v int64, err error) {
	session := t.conn.Default(ctx)
	m := make(map[string]*int64)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	err = session.Model(&data.Task{}).Where("project_code=?", readProjectCode).Select("max(id_num) as maxIdNum").Take(&m).Error
	if m["maxIdNum"] == nil {
		return 0, nil
	}
	v = *m["maxIdNum"]
	return
}

func (t *TaskDao) FindTaskByStageCode(ctx context.Context, stageCode int) (taskList []*data.Task, err error) {
	session := t.conn.Default(ctx)
	err = session.Model(&data.Task{}).Where("stage_code=?", stageCode).Find(&taskList).Error
	return
}
func (t *TaskDao) UpdateTaskSort(ctx context.Context, conn database.Dbconn, ts *data.Task) error {
	t.conn = conn.(*gorms.GormConn)
	err := t.conn.Tran(ctx).Model(&data.Task{}).
		Where("id=?", ts.Id).
		Select("sort", "stage_code").
		Updates(&ts).
		Error
	return err
}
func (t *TaskDao) FindTaskByIds(background context.Context, taskIdList []int64) (list []*data.Task, err error) {
	session := t.conn.Default(background)
	err = session.Model(&data.Task{}).Where("id in (?)", taskIdList).Find(&list).Error
	return
}
func (t *TaskDao) FindTaskMemberPage(ctx context.Context, taskCode int64, page int64, size int64) (list []*data.TaskMember, total int64, err error) {
	session := t.conn.Default(ctx)
	offset := (page - 1) * size
	err = session.Model(&data.TaskMember{}).
		Where("task_code=?", taskCode).
		Limit(int(size)).Offset(int(offset)).
		Find(&list).Error
	err = session.Model(&data.TaskMember{}).
		Where("task_code=?", taskCode).
		Count(&total).Error
	return
}
func (t *TaskDao) FindTaskById(ctx context.Context, taskCode int64) (ts *data.Task, err error) {
	session := t.conn.Default(ctx)
	err = session.Where("id=?", taskCode).Take(&ts).Error
	return
}
func (t *TaskDao) FindTaskByCreateBy(ctx context.Context, memberId int64, done int) (tList []*data.Task, total int64, err error) {
	session := t.conn.Default(ctx)
	err = session.Model(&data.Task{}).Where("create_by=? and deleted=0 and done=?", memberId, done).Find(&tList).Error
	err = session.Model(&data.Task{}).Where("create_by=? and deleted=0 and done=?", memberId, done).Count(&total).Error
	return
}

func (t *TaskDao) FindTaskByMemberCode(ctx context.Context, memberId int64, done int) (tList []*data.Task, total int64, err error) {
	session := t.conn.Default(ctx)
	sql := "select a.* from ms_task a,ms_task_member b where a.id=b.task_code and member_code=? and a.deleted=0 and a.done=?"
	raw := session.Model(&data.Task{}).Raw(sql, memberId, done)
	err = raw.Scan(&tList).Error
	if err != nil {
		return nil, 0, err
	}
	sqlCount := "select count(*) from ms_task a,ms_task_member b where a.id=b.task_code and member_code=? and a.deleted=0 and a.done=?"
	rawCount := session.Model(&data.Task{}).Raw(sqlCount, memberId, done)
	err = rawCount.Scan(&total).Error
	return
}

func (t *TaskDao) FindTaskByAssignTo(ctx context.Context, memberId int64, done int) (tsList []*data.Task, total int64, err error) {
	session := t.conn.Default(ctx)
	err = session.Model(&data.Task{}).Where("assign_to=? and deleted=0 and done=?", memberId, done).Find(&tsList).Error
	err = session.Model(&data.Task{}).Where("assign_to=? and deleted=0 and done=?", memberId, done).Count(&total).Error
	return
}
func NewTaskDao() *TaskDao {
	return &TaskDao{
		conn: gorms.New(),
	}
}
