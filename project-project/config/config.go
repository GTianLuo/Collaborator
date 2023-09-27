package config

import (
	"bytes"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"os"
	"project-common/logs"
	"project-project/internal/dao"
	"project-project/internal/database/gorms"
	//"project-project/internal/database/gorms"
)

var AppConf = InitConfig()

type ServerConfig struct {
	Name string
	Addr string
}
type Config struct {
	viper       *viper.Viper
	Gc          *GrpcConfig
	EtcdConfig  *EtcdConfig
	Sc          *ServerConfig
	MysqlConfig *MysqlConfig
	JwtConfig   *JwtConfig
	DbConfig    *DbConfig
}
type GrpcConfig struct {
	Name    string
	Addr    string
	Version string
	Weight  int
}
type EtcdConfig struct {
	Addrs []string
}
type MysqlConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Db       string
}
type JwtConfig struct {
	AccessExp     int
	RefreshExp    int
	AccessSecret  string
	RefreshSecret string
}

func InitConfig() *Config {
	conf := &Config{viper: viper.New()}
	//加入nacos
	nacos := InitNacosClient()
	configYaml, err := nacos.confClient.GetConfig(vo.ConfigParam{
		DataId: "config.yaml",
		Group:  BC.NacosConfig.Group,
	})
	if err != nil {
		log.Fatalln(err)
	}
	conf.viper.SetConfigType("yaml")
	if configYaml != "" {
		err := conf.viper.ReadConfig(bytes.NewBuffer([]byte(configYaml)))
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("load nacos config")
		err = nacos.confClient.ListenConfig(vo.ConfigParam{
			DataId: "config.yaml",
			Group:  BC.NacosConfig.Group,
			OnChange: func(namespace, group, dataId, data string) {
				log.Println("listen nacos config change", data)
				//监听变化
				err = conf.viper.ReadConfig(bytes.NewBuffer([]byte(data)))
				if err != nil {
					log.Printf("listen nacos config parse err %s \n", err.Error())
				}
				//重新载入配置
				conf.ReLoadAllConfig()
			},
		})
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		workDir, _ := os.Getwd()
		conf.viper.SetConfigName("config")
		conf.viper.AddConfigPath(workDir + "/config")
		conf.viper.AddConfigPath("D:/go/project/ms_project/project-project/config")
		err := conf.viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}
	conf.ReLoadAllConfig()
	return conf
}

func (c *Config) InitZapLog() {
	//从配置中读取日志配置，初始化日志
	lc := &logs.LogConfig{
		DebugFileName: c.viper.GetString("zap.debugFileName"),
		InfoFileName:  c.viper.GetString("zap.infoFileName"),
		WarnFileName:  c.viper.GetString("zap.warnFileName"),
		MaxSize:       c.viper.GetInt("maxSize"),
		MaxAge:        c.viper.GetInt("maxAge"),
		MaxBackups:    c.viper.GetInt("maxBackups"),
	}
	err := logs.InitLogger(lc)
	if err != nil {
		log.Fatalln(err)
	}
}

func (c *Config) InitRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     c.viper.GetString("redis.host") + ":" + c.viper.GetString("redis.port"),
		Password: c.viper.GetString("redis.password"), // no password set
		DB:       c.viper.GetInt("db"),                // use default DB
	}
}
func (c *Config) ReadServerConfig() {
	sc := &ServerConfig{}
	sc.Addr = c.viper.GetString("server.addr")
	sc.Name = c.viper.GetString("server.name")
	c.Sc = sc
}
func (c *Config) ReadGrpcConfig() {
	gc := &GrpcConfig{}
	gc.Name = c.viper.GetString("grpc.name")
	gc.Addr = c.viper.GetString("grpc.addr")
	gc.Version = c.viper.GetString("grpc.version")
	gc.Weight = c.viper.GetInt("grpc.weight")
	c.Gc = gc
}
func (c *Config) ReadJwtConfig() {
	jc := &JwtConfig{}
	jc.AccessSecret = c.viper.GetString("jwt.accessSecret")
	jc.RefreshExp = c.viper.GetInt("jwt.refreshExp")
	jc.AccessExp = c.viper.GetInt("jwt.accessExp")
	jc.RefreshSecret = c.viper.GetString("jwt.refreshSecret")
	c.JwtConfig = jc
}
func (c *Config) ReadEtcdConfig() {
	ec := &EtcdConfig{}
	var addrs []string
	c.viper.UnmarshalKey("etcd.addrs", &addrs)
	ec.Addrs = addrs
	c.EtcdConfig = ec
}
func (c *Config) ReadMysqlConfig() {
	mc := &MysqlConfig{
		Username: c.viper.GetString("mysql.username"),
		Password: c.viper.GetString("mysql.password"),
		Host:     c.viper.GetString("mysql.host"),
		Port:     c.viper.GetInt("mysql.port"),
		Db:       c.viper.GetString("mysql.db"),
	}
	c.MysqlConfig = mc
}
func (c *Config) ReLoadAllConfig() {
	c.ReadServerConfig()
	c.InitZapLog()
	c.ReadGrpcConfig()
	c.ReadEtcdConfig()
	c.ReadMysqlConfig()
	c.ReadJwtConfig()
	//重新创建相关的客户端
	c.ReConnRedis()
	c.ReConnMysql()
}
func (c *Config) ReConnRedis() {
	rdb := redis.NewClient(c.InitRedisOptions())
	rc := &dao.RedisCache{
		Rdb: rdb,
	}
	dao.Rc = rc
}

type DbConfig struct {
	Separation bool
	Master     MysqlConfig
	Slave      []MysqlConfig
}

var _db *gorm.DB

func (c *Config) InitDbConfig() {
	mc := DbConfig{}
	mc.Separation = c.viper.GetBool("db.separation")
	var slaves []MysqlConfig
	err := c.viper.UnmarshalKey("db.slave", &slaves)
	if err != nil {
		panic(err)
	}
	master := MysqlConfig{
		Username: c.viper.GetString("db.master.username"),
		Password: c.viper.GetString("db.master.password"),
		Host:     c.viper.GetString("db.master.host"),
		Port:     c.viper.GetInt("db.master.port"),
		Db:       c.viper.GetString("db.master.db"),
	}
	mc.Master = master
	mc.Slave = slaves
	c.DbConfig = &mc
}
func (c *Config) ReConnMysql() {
	if c.DbConfig.Separation {
		//读写分离配置
		username := c.DbConfig.Master.Username //账号
		password := c.DbConfig.Master.Password //密码
		host := c.DbConfig.Master.Host         //数据库地址，可以是Ip或者域名
		port := c.DbConfig.Master.Port         //数据库端口
		Dbname := c.DbConfig.Master.Db         //数据库名
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
		var err error
		_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			panic("连接数据库失败, error=" + err.Error())
		}
		replicas := []gorm.Dialector{}
		for _, v := range c.DbConfig.Slave {
			username := v.Username //账号
			password := v.Password //密码
			host := v.Host         //数据库地址，可以是Ip或者域名
			port := v.Port         //数据库端口
			Dbname := v.Db         //数据库名
			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
			cfg := mysql.Config{
				DSN: dsn,
			}
			replicas = append(replicas, mysql.New(cfg))
		}
		_db.Use(dbresolver.Register(dbresolver.Config{
			Sources: []gorm.Dialector{mysql.New(mysql.Config{
				DSN: dsn,
			})},
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		}).
			SetMaxIdleConns(10).
			SetMaxOpenConns(200))
	} else {
		//配置MySQL连接参数
		username := c.MysqlConfig.Username //账号
		password := c.MysqlConfig.Password //密码
		host := c.MysqlConfig.Host         //数据库地址，可以是Ip或者域名
		port := c.MysqlConfig.Port         //数据库端口
		Dbname := c.MysqlConfig.Db         //数据库名
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
		var err error
		_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			panic("连接数据库失败, error=" + err.Error())
		}
	}
	gorms.SetDB(_db)
}
