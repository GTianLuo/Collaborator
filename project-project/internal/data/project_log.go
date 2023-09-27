package data

import (
	"github.com/jinzhu/copier"
	"project-common/tms"
	"strconv"
)

type ProjectLog struct {
	Id           int64
	MemberCode   int64
	Content      string
	Remark       string
	Type         string
	CreateTime   int64
	SourceCode   int64
	ActionType   string
	ToMemberCode int64
	IsComment    int
	ProjectCode  int64
	Icon         string
	IsRobot      int
}

func (*ProjectLog) TableName() string {
	return "ms_project_log"
}

type ProjectLogDisplay struct {
	Id           int64
	MemberCode   string
	Content      string
	Remark       string
	Type         string
	CreateTime   string
	SourceCode   string
	ActionType   string
	ToMemberCode string
	IsComment    int
	ProjectCode  string
	Icon         string
	IsRobot      int
	Member       Member
}

func (l *ProjectLog) ToDisplay() *ProjectLogDisplay {
	pd := &ProjectLogDisplay{}
	copier.Copy(pd, l)
	pd.MemberCode = strconv.FormatInt(l.MemberCode, 10)
	pd.ToMemberCode = strconv.FormatInt(l.ToMemberCode, 10)
	pd.ProjectCode = strconv.FormatInt(l.ProjectCode, 10)
	pd.CreateTime = tms.FormatByMill(l.CreateTime)
	pd.SourceCode = strconv.FormatInt(l.SourceCode, 10)
	return pd
}

type IndexProjectLogDisplay struct {
	Content      string
	Remark       string
	CreateTime   string
	SourceCode   string
	IsComment    int
	ProjectCode  string
	MemberAvatar string
	MemberName   string
	ProjectName  string
	TaskName     string
}

func (l *ProjectLog) ToIndexDisplay() *IndexProjectLogDisplay {
	pd := &IndexProjectLogDisplay{}
	copier.Copy(pd, l)
	pd.ProjectCode = strconv.FormatInt(l.ProjectCode, 10)
	pd.CreateTime = tms.FormatByMill(l.CreateTime)
	pd.SourceCode = strconv.FormatInt(l.SourceCode, 10)
	return pd
}
