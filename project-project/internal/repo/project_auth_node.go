package repo

import (
	"context"
	"project-project/internal/database"
)

type ProjectAuthNodeRepo interface {
	FindNodeStringList(ctx context.Context, authId int64) ([]string, error)
	DeleteByAuthId(background context.Context, conn database.Dbconn, authId int64) error
	Save(background context.Context, conn database.Dbconn, authId int64, nodes []string) error
}
