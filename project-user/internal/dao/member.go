package dao

import (
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/gorms"
)

type MemberDao struct {
	conn *gorms.GormConn
}

func (m *MemberDao) GetMemberByEmail(ctx context.Context, email string) (bool, error) {
	//TODO implement me
	var count int64
	err := m.conn.Default(ctx).Model(&member.Member{}).Where("email=?", email).Count(&count).Error
	return count > 0, err
}

func (m *MemberDao) GetMemberByAccount(ctx context.Context, account string) (bool, error) {
	//TODO implement me
	var count int64
	err := m.conn.Default(ctx).Model(&member.Member{}).Where("name=?", account).Count(&count).Error
	return count > 0, err
}

func (m *MemberDao) GetMemberByMobile(ctx context.Context, mobile string) (bool, error) {
	//TODO implement me
	var count int64
	err := m.conn.Default(ctx).Model(&member.Member{}).Where("mobile=?", mobile).Count(&count).Error
	return count > 0, err
}
func (m *MemberDao) SaveMember(conn database.Dbconn, ctx context.Context, member *member.Member) error {
	m.conn = conn.(*gorms.GormConn)
	return m.conn.Tran(ctx).Create(member).Error
}
func (m *MemberDao) FindMember(ctx context.Context, account string) (*member.Member, error) {
	var res *member.Member
	err := m.conn.Default(ctx).Model(&member.Member{}).Where("email=?", account).Find(&res).Error
	return res, err
}
func (m *MemberDao) FindMemberById(ctx context.Context, id int64) (*member.Member, error) {
	var mem *member.Member
	err := m.conn.Default(ctx).Where("id=?", id).First(&mem).Error
	if err == gorm.ErrRecordNotFound {
		return nil, err
	}
	return mem, err
}
func (m *MemberDao) FindMemberByIds(background context.Context, ids []int64) (list []*member.Member, err error) {
	if len(ids) <= 0 {
		zap.L().Info("FindMember len ids <= 0")
		return nil, nil
	}
	err = m.conn.Default(background).Model(&member.Member{}).Where("id in (?)", ids).First(&list).Error
	return
}

func NewMemberDao() *MemberDao {
	return &MemberDao{
		conn: gorms.New(),
	}
}
