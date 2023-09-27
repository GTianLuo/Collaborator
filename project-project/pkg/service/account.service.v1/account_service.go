package account_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"project-common/errs"
	"project-grpc/account"
	"project-project/internal/dao"
	"project-project/internal/database/tran"
	"project-project/internal/domain"
	"project-project/internal/repo"
	"strconv"
)

type AccountService struct {
	account.UnimplementedAccountServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	accountDomain     *domain.AccountDomain
	projectAuthDomain *domain.ProjectAuthDomain
}

func New() *AccountService {
	return &AccountService{
		cache:             dao.Rc,
		transaction:       dao.NewTransactionUser(),
		accountDomain:     domain.NewAccountDomain(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

func (a *AccountService) Account(ctx context.Context, msg *account.AccountReqMessage) (*account.AccountResponse, error) {
	//1. 去account表查询account
	//2. 去auth表查询authList
	accountList, total, err := a.accountDomain.AccountList(
		msg.OrganizationCode,
		msg.MemberId,
		msg.Page,
		msg.PageSize,
		msg.DepartmentCode,
		msg.SearchType)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	OrganizationInt, _ := strconv.ParseInt(msg.OrganizationCode, 10, 64)
	authList, err := a.projectAuthDomain.AuthList(OrganizationInt)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var maList []*account.MemberAccount
	copier.Copy(&maList, accountList)
	var prList []*account.ProjectAuth
	copier.Copy(&prList, authList)
	return &account.AccountResponse{
		AccountList: maList,
		AuthList:    prList,
		Total:       total,
	}, nil
}
