package gorms

import (
	"context"
	//"fmt"
	//"gorm.io/driver/mysql"
	"gorm.io/gorm"
	//"gorm.io/gorm/logger"
	//"project-project/config"
)

var _db *gorm.DB

//func init() {
//	//配置MySQL连接参数
//	username := config.AppConf.MysqlConfig.Username //账号
//	password := config.AppConf.MysqlConfig.Password //密码
//	host := config.AppConf.MysqlConfig.Host         //数据库地址，可以是Ip或者域名
//	port := config.AppConf.MysqlConfig.Port         //数据库端口
//	Dbname := config.AppConf.MysqlConfig.Db         //数据库名
//	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
//	var err error
//	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//		Logger: logger.Default.LogMode(logger.Info),
//	})
//	if err != nil {
//		panic("连接数据库失败, error=" + err.Error())
//	}
//}

func GetDB() *gorm.DB {
	return _db
}

type GormConn struct {
	DB *gorm.DB
	TX *gorm.DB
}

func New() *GormConn {
	return &GormConn{DB: GetDB()}
}
func NewTran() *GormConn {
	return &GormConn{DB: GetDB(), TX: GetDB()}
}
func (g *GormConn) Default(ctx context.Context) *gorm.DB {
	return g.DB.Session(&gorm.Session{Context: ctx})
}
func (g *GormConn) Begin() {
	g.TX = g.TX.Begin()
}
func (g *GormConn) Rollback() {
	g.TX.Rollback()
}
func (g *GormConn) Commit() {
	g.TX.Commit()
}
func (g *GormConn) Tran(ctx context.Context) *gorm.DB {
	return g.TX.WithContext(ctx)
}
func SetDB(db *gorm.DB) {
	_db = db
}
