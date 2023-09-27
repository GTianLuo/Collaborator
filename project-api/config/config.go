package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"project-common/logs"
)

var AppConf = InitConfig()

type Config struct {
	viper      *viper.Viper
	EtcdConfig *EtcdConfig
}
type GrpcConfig struct {
	Name string
	Addr string
}
type EtcdConfig struct {
	Addrs []string
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
	conf.ReadEtcdConfig()
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
func (c *Config) ReadEtcdConfig() {
	ec := &EtcdConfig{}
	var addrs []string
	c.viper.UnmarshalKey("etcd.addrs", &addrs)
	ec.Addrs = addrs
	c.EtcdConfig = ec
}
