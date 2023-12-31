package department_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/errs"
	"project-grpc/department"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/domain"
	"project-project/internal/repo"
	"strconv"
)

type DepartmentService struct {
	department.UnimplementedDepartmentServiceServer
	cache            repo.Cache
	transaction      tran.Transaction
	departmentDomain *domain.DepartmentDomain
}

func New() *DepartmentService {
	return &DepartmentService{
		cache:            dao.Rc,
		transaction:      dao.NewTransactionUser(),
		departmentDomain: domain.NewDepartmentDomain(),
	}
}
func (d *DepartmentService) List(ctx context.Context, msg *department.DepartmentReqMessage) (*department.ListDepartmentMessage, error) {
	organizationCode, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	var parentDepartmentCode int64
	if msg.ParentDepartmentCode != "" {
		parentDepartmentCode, _ = strconv.ParseInt(msg.ParentDepartmentCode, 10, 64)
	}
	dps, total, err := d.departmentDomain.List(
		organizationCode,
		parentDepartmentCode,
		msg.Page,
		msg.PageSize)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var list []*department.DepartmentMessage
	copier.Copy(&list, dps)
	return &department.ListDepartmentMessage{List: list, Total: total}, nil
}

func (d *DepartmentService) Save(ctx context.Context, msg *department.DepartmentReqMessage) (*department.DepartmentMessage, error) {
	organizationCode, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	var departmentCode int64
	if msg.DepartmentCode != "" {
		departmentCode, _ = strconv.ParseInt(msg.DepartmentCode, 10, 64)
	}
	var parentDepartmentCode int64
	if msg.ParentDepartmentCode != "" {
		parentDepartmentCode, _ = strconv.ParseInt(msg.ParentDepartmentCode, 10, 64)
	}
	dp, err := d.departmentDomain.Save(
		organizationCode,
		departmentCode,
		parentDepartmentCode,
		msg.Name)
	if err != nil {
		return &department.DepartmentMessage{}, errs.GrpcError(err)
	}
	var res = &department.DepartmentMessage{}
	copier.Copy(res, dp)
	return res, nil
}

func (d *DepartmentService) Read(ctx context.Context, msg *department.DepartmentReqMessage) (*department.DepartmentMessage, error) {
	//organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	departmentCode, _ := strconv.ParseInt(msg.DepartmentCode, 10, 64)
	dp, err := d.departmentDomain.FindDepartmentById(departmentCode)
	if err != nil {
		return &department.DepartmentMessage{}, errs.GrpcError(err)
	}
	var res = &department.DepartmentMessage{}
	copier.Copy(res, dp.ToDisplay())
	return res, nil
}
