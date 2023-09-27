package project

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
	"project-api/api/midd"
	"project-api/router"
)

func init() {
	log.Println("init project router")
	router.Register(&RouterProject{})
}

type RouterProject struct {
}

func (*RouterProject) Route(r *gin.Engine) {
	InitRpcUserClient()
	h := NewProjectHanlder()
	t := NewTask()
	zap.L().Info("接口开始执行")
	group := r.Group("/project/index")
	group.Use(midd.TokenVerify())
	group.POST("", h.Index)
	group1 := r.Group("/project/project")
	group1.Use(midd.TokenVerify())
	group1.POST("/selfList", h.myProjectList)
	group1.POST("", h.myProjectList)
	group2 := r.Group("/project")
	group2.Use(midd.TokenVerify())
	group2.POST("/project_template", h.projectTemplate)
	group2.POST("/project/save", h.projectSave)
	group2.POST("/project/read", h.readProject)
	group2.POST("/project/recycle", h.recycleProject)
	group2.POST("/project/recovery", h.recoveryProject)
	group2.POST("/project_collect/collect", h.collectProject)
	group2.POST("/project/edit", h.editProject)
	group2.POST("/task_stages", t.taskStages)
	group2.POST("/project_member/index", t.taskMemberList)
	group2.POST("/task_stages/tasks", t.taskList)
	group2.POST("/task/selfList", t.myTaskList)
	group2.POST("/task/save", t.saveTask)
	group2.POST("/task/sort", t.taskSort)
	group2.POST("/task/read", t.readTask)
	group2.POST("/task_member", t.listTaskMember)
	group2.POST("/task/taskLog", t.taskLog)
	group2.POST("/task/_taskWorkTimeList", t.taskWorkTimeList)
	group2.POST("task/saveTaskWorkTime", t.saveTaskWorkTime)
	group2.POST("/file/uploadFiles", t.uploadFiles)
	group2.POST("/task/taskSources", t.taskSources)
	group2.POST("/task/createComment", t.createComment)
	group2.POST("/project/getLogBySelfProject", h.getLogBySelfProject)
	a := NewAccount()
	group2.POST("/account", a.account)
	d := NewDepartment()
	group2.POST("/department", d.department)
	group2.POST("/department/save", d.save)
	group2.POST("/department/read", d.read)
	auth := NewAuth()
	group2.POST("/auth", auth.authList)
	group.POST("/auth/apply", auth.apply)
	menu := NewMenu()
	group.POST("/menu/menu", menu.menuList)
}
