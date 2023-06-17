package main

import (
	"fmt"
	"log"
	"time"

	env "github.com/caarlos0/env/v8"
)

type AuthConfig struct {
	AuthPath   string `env:"AUTH_PATH" envDefault:"/.launcher/auth"`
	EnableAuth bool   `env:"ENABLE_AUTH" envDefault:"false"`
	Username   string `env:"AUTH_USERNAME"`
	Password   string `env:"AUTH_PASSWORD"`
}

type Ec2Config struct {
	Region          string `env:"AWS_REGION"`
	ImageId         string `env:"EC2_IMAGE_ID"`
	InstanceType    string `env:"EC2_INSTANCE_TYPE"`
	KeyName         string `env:"AWS_KEY_NAME"`
	SecurityGroupId string `env:"AWS_SECURITY_GROUP_ID"`
	StartScript     string `env:"EC2_SCRIPT"`
	Tag             string `env:"EC2_TAG" envDefault:"created-by-launcher"`
	DiskSize        int64  `env:"EC2_DISK_SIZE" envDefault:"16"`
	Port            int    `env:"EC2_PORT" envDefault:"0"`
	UsePrivateDns   bool   `env:"AWS_USE_PRIVATE_DNS" envDefault:"false"`
}

type Config struct {
	Port     string        `env:"PORT" envDefault:"7890"`
	Host     string        `env:"HOST"`
	LogLevel string        `env:"LOG_LEVEL" envDefault:"INFO"`
	CacheTtl time.Duration `env:"CACHE_TTL" envDefault:"1800s"`
	WaitTime time.Duration `env:"LAUNCH_WAIT_TIME" envDefault:"31s"`

	Ec2Config  Ec2Client
	AuthConfig AuthConfig
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%v:%v", c.Host, c.Port)
}

func ConfigFromEnv() Config {
	c := Config{}

	err := env.Parse(&c)
	if err != nil {
		log.Fatal(err)
	}

	authConfig := &c.AuthConfig
	if authConfig.EnableAuth && (authConfig.Username == "" || authConfig.Password == "") {
		log.Fatal("AUTH_USERNAME and AUTH_PASSWORD must be set if ENABLE_AUTH is set")
	}

	return c
}
