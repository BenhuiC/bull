package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var Cfg config

type config struct {
	Base     Base           `yaml:"base"`
	Http     HttpCfg        `yaml:"http"`
	Database DatabaseCfg    `yaml:"database"`
	Service  ServiceCfg     `yaml:"services"`
	Hasher   HasherConfig   `yaml:"hasher"`
	Kube     KubeConfig     `yaml:"kube"`
	Es       EsConfig       `yaml:"elastic"`
}

type Base struct {
	Env           string `yaml:"env"`
}

type HttpCfg struct {
	ListenAddr string `yaml:"listenAddr"`
}

type DatabaseCfg struct {
	DB               string `yaml:"db"`
	WorkerRedisURL   string `yaml:"workerRedisURL"`
	RedisInstanceURL string `yaml:"redisInstanceURL"`
}

type ServiceCfg struct {
	Stub ServerConfig `yaml:"stub"`
}

func InitConfig(v *viper.Viper) (err error) {
	err = v.Unmarshal(&Cfg, func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})
	return
}

type HasherConfig struct {
	Salt     string `yaml:"salt"`
	MinLen   int    `yaml:"minLen"`
	Alphabet string `yaml:"alphabet"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
	AK   string `yaml:"ak"`
	SK   string `yaml:"sk"`
	Host string `yaml:"host"`
}

type KubeConfig struct {
	Config    string `yaml:"config"`
	Namespace string `yaml:"namespace"`
	JobTTL    int32  `yaml:"jobTTL"`
}

type EsConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
