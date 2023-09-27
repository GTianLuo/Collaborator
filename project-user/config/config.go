package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
	"os"
	"project-common/logs"
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
	v := viper.New()
	conf := &Config{viper: v}
	workDir, _ := os.Getwd()
	conf.viper.SetConfigName("app")
	conf.viper.SetConfigType("yml")
	conf.viper.AddConfigPath(workDir + "/config")
	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	conf.ReadGrpcConfig()
	conf.ReadEtcdConfig()
	conf.ReadServerConfig()
	conf.ReadMysqlConfig()
	conf.ReadJwtConfig()
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
