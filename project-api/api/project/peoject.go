package project

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"project-api/pkg/model"
	"project-api/pkg/model/menu"
	"project-api/pkg/model/project"
	common "project-common"
	"project-common/errs"
	project_service_v1 "project-grpc/project"
	"strconv"
	"time"
)

type HandlerProject struct {
}

func NewProjectHanlder() *HandlerProject {
	return &HandlerProject{}
}
func (p *HandlerProject) Index(ctx *gin.Context) {
	result := &common.Result{}
	ctxTime, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project_service_v1.IndexMessage{}
	response, err := ProjectServiceClient.Index(ctxTime, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var ms []*menu.Menu
	copier.Copy(&ms, response.Menus)
	ctx.JSON(http.StatusOK, result.Success(ms))
}

func (p *HandlerProject) myProjectList(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	idAny, _ := c.Get("memberId")
	id := idAny.(int64)
	nameAny, _ := c.Get("memberName")
	name := nameAny.(string)
	selectBy := c.PostForm("selectBy")
	var page = &model.Page{}
	page.Bind(c)
	pm, err := ProjectServiceClient.FindProjectByMemId(ctx, &project_service_v1.ProjectRpcMessage{
		MemberId:   id,
		Page:       page.Page,
		PageSize:   page.PageSize,
		SelectBy:   selectBy,
		MemberName: name,
	})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var pam []*project.ProAndMember
	copier.Copy(&pam, pm.Pm)
	if pam == nil {
		pam = []*project.ProAndMember{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pam,
		"total": pm.Total,
	}))
}

func (u *HandlerProject) projectTemplate(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberId := c.GetInt64("memberId")
	memberName := c.GetString("memberName")
	organizationCode := c.GetString("organizationCode")
	var page = &model.Page{}
	page.Bind(c)
	viewTypeStr := c.PostForm("viewType")
	viewType, _ := strconv.ParseInt(viewTypeStr, 10, 64)
	projectTemplateRsp, err := ProjectServiceClient.FindProjectTemplate(ctx,
		&project_service_v1.ProjectRpcMessage{
			MemberId:         memberId,
			MemberName:       memberName,
			OrganizationCode: organizationCode,
			Page:             page.Page,
			PageSize:         page.PageSize,
			ViewType:         int32(viewType)})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var pts []*project.ProjectTemplate
	copier.Copy(&pts, projectTemplateRsp.Ptm)
	if pts == nil {
		pts = []*project.ProjectTemplate{}
	}
	c.JSON(http.StatusOK, result.Success(
		gin.H{
			"list":  pts,
			"total": projectTemplateRsp.Total,
		}))
}
func (u *HandlerProject) projectSave(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberId := c.GetInt64("memberId")
	organizationCode := c.GetString("organizationCode")
	var req *project.SaveProjectRequest
	c.ShouldBind(&req)
	msg := &project_service_v1.ProjectRpcMessage{
		MemberId:         memberId,
		Name:             req.Name,
		OrganizationCode: organizationCode,
		Description:      req.Description,
		TemplateCode:     req.TemplateCode,
		Id:               int64(req.Id)}
	saveProjectMessage, err := ProjectServiceClient.SaveProject(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var sp *project.SaveProject
	copier.Copy(&sp, saveProjectMessage)
	c.JSON(http.StatusOK, result.Success(sp))
}

func (p *HandlerProject) readProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	detail, err := ProjectServiceClient.FindProjectDetail(ctx, &project_service_v1.ProjectRpcMessage{ProjectCode: projectCode, MemberId: memberId})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	pd := &project.ProjectDetail{}
	copier.Copy(pd, detail)
	c.JSON(http.StatusOK, result.Success(pd))
}

func (p *HandlerProject) recycleProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &project_service_v1.ProjectRpcMessage{ProjectCode: projectCode, Deleted: true})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) recoveryProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &project_service_v1.ProjectRpcMessage{ProjectCode: projectCode, Deleted: false})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) collectProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	collectType := c.PostForm("type")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateCollectProject(ctx, &project_service_v1.ProjectRpcMessage{ProjectCode: projectCode, CollectType: collectType, MemberId: memberId})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) editProject(c *gin.Context) {
	result := &common.Result{}
	var req *project.ProjectReq
	err := c.ShouldBind(&req)
	if err != nil {
		fmt.Println(err)
	}
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project_service_v1.UpdateProjectMessage{}
	copier.Copy(&msg, req)
	msg.ProjectCode = strconv.FormatInt(req.ProjectCode, 10)
	msg.MemberId = memberId
	fmt.Println(msg)
	_, err = ProjectServiceClient.UpdateProject(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) getLogBySelfProject(c *gin.Context) {
	result := &common.Result{}
	var page = &model.Page{}
	page.Bind(c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project_service_v1.ProjectRpcMessage{
		MemberId: c.GetInt64("memberId"),
		Page:     page.Page,
		PageSize: page.PageSize,
	}
	projectLogResponse, err := ProjectServiceClient.GetLogBySelfProject(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*model.ProjectLog
	copier.Copy(&list, projectLogResponse.List)
	if list == nil {
		list = []*model.ProjectLog{}
	}
	c.JSON(http.StatusOK, result.Success(list))
}
func (p *HandlerProject) nodeList(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	response, err := ProjectServiceClient.NodeList(ctx, &project_service_v1.ProjectRpcMessage{})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	var list []*model.ProjectNodeTree
	copier.Copy(&list, response.Nodes)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"nodes": list,
	}))
}
func (p *HandlerProject) FindProjectByMemberId(memberId int64, projectCode string, taskCode string) (*project.Project, bool, bool, *errs.BError) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project_service_v1.ProjectRpcMessage{
		MemberId:    memberId,
		ProjectCode: projectCode,
		TaskCode:    taskCode,
	}
	projectResponse, err := ProjectServiceClient.FindProjectByMemberId(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		return nil, false, false, errs.NewError(errs.ErrorCode(code), msg)
	}
	if projectResponse.Project == nil {
		return nil, false, false, nil
	}
	pr := &project.Project{}
	copier.Copy(pr, projectResponse.Project)
	return pr, true, projectResponse.IsOwner, nil
}
