package conf

import (
	"flag"

	"github.com/spf13/viper"
)

var (
	configFile = flag.String("config", "configs", "path to the configuration file")
)

type Option func(*conf)

type conf struct {
	configFileType string
	configFilename string
}

var c = &conf{
	configFileType: "yaml",
	configFilename: "config",
}

func apply(opts ...Option) *conf {
	newConf := c
	for _, opt := range opts {
		opt(newConf)
	}
	return newConf
}

func WithFileType(fileType string) Option {
	return func(c *conf) {
		c.configFileType = fileType
	}
}

func WithFileName(filename string) Option {
	return func(c *conf) {
		c.configFilename = filename
	}
}

func Init(opts ...Option) {
	cur := apply(opts...)
	viper.SetConfigType(cur.configFileType)
	viper.AddConfigPath(*configFile)
	viper.SetConfigName(cur.configFilename)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
