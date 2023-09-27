package repo

import (
	"context"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/database"
)

type MemberRepo interface {
	GetMemberByEmail(ctx context.Context, email string) (bool, error)
	GetMemberByAccount(ctx context.Context, account string) (bool, error)
	GetMemberByMobile(ctx context.Context, mobile string) (bool, error)
	SaveMember(conn database.Dbconn, ctx context.Context, member *member.Member) error
	FindMember(ctx context.Context, account string) (*member.Member, error)
	FindMemberById(ctx context.Context, id int64) (*member.Member, error)
	FindMemberByIds(background context.Context, ids []int64) (list []*member.Member, err error)
}
