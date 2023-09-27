package model

import "github.com/gin-gonic/gin"

type Page struct {
	Page     int64
	PageSize int64
}

func (p *Page) Bind(c *gin.Context) {
	c.ShouldBind(&p)
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
}
