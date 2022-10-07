package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type S3 struct {
	Endpoint  string `yaml:"endpoint"`
	Scheme    string `yaml:"scheme"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
}

type Server struct {
	Listen string `yaml:"listen"`
}
type Site struct {
	Domain    string `yaml:"domain"`
	SubFolder string `yaml:"subFolder"`
}
type Config struct {
	S3     *S3     `yaml:"s3"`
	Sites  []Site  `yaml:"sites"`
	Server *Server `yaml:"server"`
}

var (
	S3Config     *S3
	SitesConfig  []Site
	ServerConfig *Server
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("fatal error config file: %v ", err)
	}
	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	S3Config = c.S3
	ServerConfig = c.Server
}
