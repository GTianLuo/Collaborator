package data

import (
	"project-common/tms"
	//	"project-project/internal/data/task"
	"strconv"
)

type Project struct {
	Id                 int64
	Cover              string
	Name               string
	Description        string
	AccessControlType  int
	WhiteList          string
	Sort               int
	Deleted            int
	TemplateCode       string
	Schedule           float64
	CreateTime         int64
	OrganizationCode   int64
	DeletedTime        string
	Private            int
	Prefix             string
	OpenPrefix         int
	Archive            int
	ArchiveTime        int64
	OpenBeginTime      int
	OpenTaskPrivate    int
	TaskBoardTheme     string
	BeginTime          int64
	EndTime            int64
	AutoUpdateSchedule int
}

func (*Project) TableName() string {
	return "ms_project"
}

type MemberProject struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   string
}
type ProAndMember struct {
	Project
	Id          int64
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   string
	OwnerName   string
	Collected   int
}

func (*MemberProject) TableName() string {
	return "ms_project_member"
}
func ToProjectMap(list []*Project) map[int64]*Project {
	m := make(map[int64]*Project, len(list))
	for _, v := range list {
		m[v.Id] = v
	}
	return m
}

func (m *Project) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

type ProjectCollection struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	CreateTime  int64
	IsOwner     int64
	Authorize   string
}

func (*ProjectCollection) TableName() string {
	return "ms_project_collection"
}
func (m *ProAndMember) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

func ToMap(orgs []*ProAndMember) map[int64]*ProAndMember {
	m := make(map[int64]*ProAndMember)
	for _, v := range orgs {
		m[v.Id] = v
	}
	return m
}

type ProjectTemplate struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       int64
	OrganizationCode int64
	Cover            string
	MemberCode       int64
	IsSystem         int
}

func (*ProjectTemplate) TableName() string {
	return "ms_project_template"
}

type ProjectTemplateAll struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       string
	OrganizationCode string
	Cover            string
	MemberCode       string
	IsSystem         int
	TaskStages       []*TaskStagesOnlyName
	Code             string
}

func (pt ProjectTemplate) Convert(taskStages []*TaskStagesOnlyName) *ProjectTemplateAll {
	//	organizationCode, _ := encrypts.EncryptInt64(pt.OrganizationCode, model.AESKey)
	//	memberCode, _ := encrypts.EncryptInt64(pt.MemberCode, model.AESKey)
	//	code, _ := encrypts.EncryptInt64(int64(pt.Id), model.AESKey)
	pta := &ProjectTemplateAll{
		Id:               pt.Id,
		Name:             pt.Name,
		Description:      pt.Description,
		Sort:             pt.Sort,
		CreateTime:       tms.FormatByMill(pt.CreateTime),
		OrganizationCode: strconv.FormatInt(pt.OrganizationCode, 10),
		Cover:            pt.Cover,
		MemberCode:       strconv.FormatInt(pt.MemberCode, 10),
		IsSystem:         pt.IsSystem,
		TaskStages:       taskStages,
		Code:             strconv.Itoa(pt.Id),
	}
	return pta
}
func ToProjectTemplateIds(pts []ProjectTemplate) []int {
	var ids []int
	for _, v := range pts {
		ids = append(ids, v.Id)
	}
	return ids
}

type ProjectMemberInfo struct {
	ProjectCode int64
	MemberCode  int64
	Name        string
	Avatar      string
	IsOwner     int64
	Email       string
}

func ToProjectMemberInfoMap(pm []*ProjectMemberInfo) map[int64]*ProjectMemberInfo {
	m := make(map[int64]*ProjectMemberInfo)
	for _, v := range pm {
		m[v.MemberCode] = v
	}
	return m
}
