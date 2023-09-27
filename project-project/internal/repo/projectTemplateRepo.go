package repo

import (
	"context"
	"project-project/internal/data"
)

type ProjectTemplateRepo interface {
	FindProjectTemplateSystem(ctx context.Context, page int64, size int64) ([]data.ProjectTemplate, int64, error)
	FindProjectTemplateCustom(ctx context.Context, memId int64, organizationCode int64, page int64, size int64) ([]data.ProjectTemplate, int64, error)
	FindProjectTemplateAll(ctx context.Context, organizationCode int64, page int64, size int64) ([]data.ProjectTemplate, int64, error)
}
