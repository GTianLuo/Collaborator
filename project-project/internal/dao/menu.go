package dao

import (
	"context"
	"project-project/internal/data/menu"
	"project-project/internal/database/gorms"
)

type MenuDao struct {
	conn *gorms.GormConn
}

func (m *MenuDao) FindMenus(ctx context.Context) (pms []*menu.ProjectMenu, err error) {
	session := m.conn.Default(ctx)
	err = session.Model(&menu.ProjectMenu{}).Order("pid,sort asc, id asc").Find(&pms).Error
	return
}

func NewMenuDao() *MenuDao {
	return &MenuDao{
		conn: gorms.New(),
	}
}
