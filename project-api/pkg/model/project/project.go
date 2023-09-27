package project

type Project struct {
	Id                 int64   `json:"id"`
	Cover              string  `json:"cover"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	AccessControlType  string  `json:"access_control_type"`
	WhiteList          string  `json:"white_list"`
	Order              int     `json:"order"`
	Deleted            int     `json:"deleted"`
	TemplateCode       string  `json:"template_code"`
	Schedule           float64 `json:"schedule"`
	CreateTime         string  `json:"create_time"`
	OrganizationCode   string  `json:"organization_code"`
	DeletedTime        string  `json:"deleted_time"`
	Private            int     `json:"private"`
	Prefix             string  `json:"prefix"`
	OpenPrefix         int     `json:"open_prefix"`
	Archive            int     `json:"archive"`
	ArchiveTime        int64   `json:"archive_time"`
	OpenBeginTime      int     `json:"open_begin_time"`
	OpenTaskPrivate    int     `json:"open_task_private"`
	TaskBoardTheme     string  `json:"task_board_theme"`
	BeginTime          int64   `json:"begin_time"`
	EndTime            int64   `json:"end_time"`
	AutoUpdateSchedule int     `json:"auto_update_schedule"`
	Code               string  `json:"code"`
}
type ProjectReq struct {
	ProjectCode        int64   `json:"projectCode" form:"projectCode"`
	Cover              string  `json:"cover" form:"cover"`
	Name               string  `json:"name" form:"name"`
	Description        string  `json:"description" form:"description"`
	AccessControlType  string  `json:"access_control_type" form:"access_control_type"`
	WhiteList          string  `json:"white_list" form:"white_list"`
	Schedule           float64 `json:"schedule" form:"schedule"`
	Private            int     `json:"private" form:"private"`
	Prefix             string  `json:"prefix" form:"prefix"`
	OpenPrefix         int     `json:"open_prefix" form:"open_prefix"`
	OpenBeginTime      int     `json:"open_begin_time" form:"open_begin_time"`
	OpenTaskPrivate    int     `json:"open_task_private" form:"open_task_private"`
	TaskBoardTheme     string  `json:"task_board_theme" form:"task_board_theme"`
	AutoUpdateSchedule int     `json:"auto_update_schedule" form:"auto_update_schedule"`
}

type MemberProject struct {
	Id          int64  `json:"id"`
	ProjectCode int64  `json:"project_code"`
	MemberCode  int64  `json:"member_code"`
	JoinTime    string `json:"join_time"`
	IsOwner     int64  `json:"is_owner"`
	Authorize   string `json:"authorize"`
}

type ProAndMember struct {
	Project
	ProjectCode int64  `json:"project_code"`
	MemberCode  int64  `json:"member_code"`
	JoinTime    int64  `json:"join_time"`
	IsOwner     int64  `json:"is_owner"`
	Authorize   string `json:"authorize"`
	OwnerName   string `json:"owner_name"`
	Collected   int    `json:"collected"`
}
type ProjectTemplate struct {
	Id               int                   `json:"id"`
	Name             string                `json:"name"`
	Description      string                `json:"description"`
	Sort             int                   `json:"sort"`
	CreateTime       string                `json:"create_time"`
	OrganizationCode string                `json:"organization_code"`
	Cover            string                `json:"cover"`
	MemberCode       string                `json:"member_code"`
	IsSystem         int                   `json:"is_system"`
	TaskStages       []*TaskStagesOnlyName `json:"task_stages"`
	Code             string                `json:"code"`
}
type ProjectDetail struct {
	Project
	OwnerName   string `json:"owner_name"`
	Collected   int    `json:"collected"`
	OwnerAvatar string `json:"owner_avatar"`
}
type SaveProjectRequest struct {
	Name         string `json:"name" form:"name"`
	TemplateCode string `json:"templateCode" form:"templateCode"`
	Description  string `json:"description" form:"description"`
	Id           int    `json:"id" form:"id"`
}

type SaveProject struct {
	Id               int64  `json:"id"`
	Cover            string `json:"cover"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Code             string `json:"code"`
	CreateTime       string `json:"create_time"`
	TaskBoardTheme   string `json:"task_board_theme"`
	OrganizationCode string `json:"organization_code"`
}
type TaskStagesOnlyName struct {
	Name string `json:"name"`
}
type MemberProjectResp struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Avatar  string `json:"avatar"`
	Code    string `json:"code"`
	IsOwner int    `json:"isOwner"`
}
type ProjectInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
type Executor struct {
	name   string `json:"name"`
	avatar string `json:"avatar"`
}

type MyTaskDisplay struct {
	ProjectCode        string      `json:"project_code"`
	Name               string      `json:"name"`
	Pri                int         `json:"pri"`
	ExecuteStatus      string      `json:"execute_status"`
	Description        string      `json:"description"`
	CreateBy           string      `json:"create_by"`
	DoneBy             string      `json:"done_by"`
	DoneTime           string      `json:"done_time"`
	CreateTime         string      `json:"create_time"`
	AssignTo           string      `json:"assign_to"`
	Deleted            int         `json:"deleted"`
	StageCode          string      `json:"stage_code"`
	TaskTag            string      `json:"task_tag"`
	Done               int         `json:"done"`
	BeginTime          string      `json:"begin_time"`
	EndTime            string      `json:"end_time"`
	RemindTime         string      `json:"remind_time"`
	Pcode              string      `json:"pcode"`
	Sort               int         `json:"sort"`
	Like               int         `json:"like"`
	Star               int         `json:"star"`
	DeletedTime        string      `json:"deleted_time"`
	Private            int         `json:"private"`
	IdNum              int         `json:"id_num"`
	Path               string      `json:"path"`
	Schedule           int         `json:"schedule"`
	VersionCode        string      `json:"version_code"`
	FeaturesCode       string      `json:"features_code"`
	WorkTime           int         `json:"work_time"`
	Status             int         `json:"status"`
	Code               string      `json:"code"`
	ProjectName        string      `json:"project_name"`
	Cover              string      `json:"cover"`
	AccessControlType  string      `json:"access_control_type"`
	WhiteList          string      `json:"white_list"`
	Order              int         `json:"order"`
	TemplateCode       string      `json:"template_code"`
	OrganizationCode   string      `json:"organization_code"`
	Prefix             string      `json:"prefix"`
	OpenPrefix         int         `json:"open_prefix"`
	Archive            int         `json:"archive"`
	ArchiveTime        string      `json:"archive_time"`
	OpenBeginTime      int         `json:"open_begin_time"`
	OpenTaskPrivate    int         `json:"open_task_private"`
	TaskBoardTheme     string      `json:"task_board_theme"`
	AutoUpdateSchedule int         `json:"auto_update_schedule"`
	HasUnDone          int         `json:"hasUnDone"`
	ParentDone         int         `json:"parentDone"`
	PriText            string      `json:"priText"`
	Executor           Executor    `json:"executor"`
	ProjectInfo        ProjectInfo `json:"projectInfo"`
}
