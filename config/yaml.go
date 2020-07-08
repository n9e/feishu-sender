package config

import (
	"fmt"

	"github.com/toolkits/pkg/file"
)

type Config struct {
	Logger   loggerSection   `yaml:"logger"`
	Sms      smsSection      `yaml:"sms"`
	Consumer consumerSection `yaml:"consumer"`
	Redis    redisSection    `yaml:"redis"`
}

type loggerSection struct {
	Dir       string `yaml:"dir"`
	Level     string `yaml:"level"`
	KeepHours uint   `yaml:"keepHours"`
}

type redisSection struct {
	Addr    string         `yaml:"addr"`
	Pass    string         `yaml:"pass"`
	Idle    int            `yaml:"idle"`
	DB      int            `yaml:"db"`
	Timeout timeoutSection `yaml:"timeout"`
}

type timeoutSection struct {
	Conn  int `yaml:"conn"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
}

type consumerSection struct {
	Queue  string `yaml:"queue"`
	Worker int    `yaml:"worker"`
}

type smsSection struct {
	Url string `yaml:"url"`
}

var yaml Config

func Get() Config {
	return yaml
}

func ParseConfig(yf string) error {
	err := file.ReadYaml(yf, &yaml)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", yf, err)
	}
	return nil
}
