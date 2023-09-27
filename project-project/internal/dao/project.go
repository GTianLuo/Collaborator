package dao

import (
	"context"
	"errors"
	"fmt"
	"project-project/internal/data"
	"project-project/internal/database"
	"project-project/internal/database/gorms"
	"strconv"
)

type ProjectDao struct {
	conn *gorms.GormConn
}

func (p *ProjectDao) FindProjectById(ctx context.Context, id int64) (pj *data.Project, err error) {
	err = p.conn.Default(ctx).Where("id=?", id).Find(&pj).Error
	return
}
func (p *ProjectDao) FindMemberInfoByProjectCode(ctx context.Context, projectCode int64, page int64, size int64) ([]*data.ProjectMemberInfo, int64, error) {
	sql := "select a.project_code,a.member_code,a.is_owner,b.`name`,b.avatar,b.email from ms_project_member a, ms_member b where a.member_code=b.id and project_code = ? limit ?,?"
	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	db := session.Raw(sql, readProjectCode, (page-1)*size, size)
	var mis []*data.ProjectMemberInfo
	err := db.Scan(&mis).Error
	var total int64
	sqlCount := "select count(*) from ms_project_member a, ms_member b where a.member_code=b.id and project_code = ?"
	dbCount := session.Raw(sqlCount, readProjectCode)
	err = dbCount.Scan(&total).Error
	return mis, total, err
}
func (p *ProjectDao) UpdateProject(ctx context.Context, proj *data.Project) error {
	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", proj.Id)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	proj.Id, _ = strconv.ParseInt(readProjectCode, 10, 64)
	return p.conn.Default(ctx).Where("id = ?", proj.Id).Updates(&proj).Error
}

func (p *ProjectDao) SaveProjectCollect(ctx context.Context, pc *data.ProjectCollection) error {
	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", pc.ProjectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	count := session.Exec("insert into ms_project_collection (`project_code`,`member_code`,`create_time`) values (?,?,?)", readProjectCode, pc.MemberCode, pc.CreateTime)
	if count.RowsAffected > 0 {
		return nil
	}
	return errors.New("更新失败")
}

func (p *ProjectDao) DeleteProjectCollect(ctx context.Context, memId int64, projectCode int64) error {
	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	return session.Where("member_code=? and project_code=?", memId, readProjectCode).Delete(&data.ProjectCollection{}).Error
}

func (p *ProjectDao) UpdateDeletedProject(ctx context.Context, code int64, deleted bool) error {
	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", code)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	var err error
	if deleted {
		err = session.Model(&data.Project{}).Where("id=?", readProjectCode).Update("deleted", 1).Error
	} else {
		err = session.Model(&data.Project{}).Where("id=?", readProjectCode).Update("deleted", 0).Error
	}
	return err
}
func (p *ProjectDao) FindProjectByPIdAndMemId(ctx context.Context, projectCode int64, id int64) (*data.ProAndMember, error) {
	var pm *data.ProAndMember

	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	raw := session.Raw(fmt.Sprintf("select * from ms_project a, ms_project_member b where a.id=b.project_code and b.member_code=? and b.project_code = ? limit 1"), id, readProjectCode)
	err := raw.Scan(&pm).Error
	return pm, err
}

func (p *ProjectDao) FindCollectByPidAndMemId(ctx context.Context, projectCode int64, id int64) (bool, error) {
	var count int64
	session := p.conn.Default(ctx)
	db2 := session.Raw("select project_code from ms_project_member where id = ?", projectCode)
	var readProjectCode string
	db2.Scan(&readProjectCode)
	db := session.Raw(fmt.Sprintf("select count(*) project_code from ms_project_collection where member_code = ? and project_code = ?"), id, readProjectCode)
	//	var mp *pro.ProAndMember
	err := db.Scan(&count).Error
	//session.Model(&pro.MemberProject{}).Where(fmt.Sprintf("member_code=? %s", condition), memId).Count(&total)
	return count > 0, err
}

func (p *ProjectDao) SaveProject(ctx context.Context, conn database.Dbconn, pr *data.Project) error {
	p.conn = conn.(*gorms.GormConn)
	return p.conn.Tran(ctx).Save(&pr).Error
}

func (p *ProjectDao) SaveProjectMember(ctx context.Context, conn database.Dbconn, pm *data.MemberProject) error {
	p.conn = conn.(*gorms.GormConn)
	return p.conn.Tran(ctx).Save(&pm).Error
}

func (p *ProjectDao) FindCollectProjectByMemId(ctx context.Context, id int64, page int64, size int64) ([]*data.ProAndMember, int64, error) {
	session := p.conn.Default(ctx)
	index := (page - 1) * size
	db := session.Raw(fmt.Sprintf("select * from ms_project where id in (select project_code from ms_project_collection where member_code = ?) limit ?,?"), id, index, size)
	var mp []*data.ProAndMember
	err := db.Scan(&mp).Error
	var total int64
	session.Model(&data.ProjectCollection{}).Where("member_code=? ", id).Count(&total)
	return mp, total, err
}

func (p *ProjectDao) FindProjectByMemId(ctx context.Context, memId int64, page int64, size int64, condition string) ([]*data.ProAndMember, int64, error) {
	session := p.conn.Default(ctx)
	index := (page - 1) * size
	db := session.Raw(fmt.Sprintf("select * from ms_project a, ms_project_member b where a.id=b.project_code and b.member_code=? %s limit ?,?", condition), memId, index, size)
	var mp []*data.ProAndMember
	err := db.Scan(&mp).Error
	var total int64
	//session.Model(&pro.MemberProject{}).Where(fmt.Sprintf("member_code=? %s", condition), memId).Count(&total)
	db = session.Raw(fmt.Sprintf("select * from ms_project a, ms_project_member b where a.id=b.project_code and b.member_code=? %s", condition), memId)
	db.Scan(&total)
	return mp, total, err
}
func (p *ProjectDao) FindProjectByIds(ctx context.Context, pids []int64) (list []*data.Project, err error) {
	session := p.conn.Default(ctx)
	err = session.Model(&data.Project{}).Where("id in (?)", pids).Find(&list).Error
	return
}
func NewProjectDao() *ProjectDao {
	return &ProjectDao{
		conn: gorms.New(),
	}
}
