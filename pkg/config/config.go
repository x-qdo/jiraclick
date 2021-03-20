package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Debug    bool `yaml:"debug"`
	RabbitMQ struct {
		URL      string `yaml:"url"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Vhost    string `yaml:"vhost"`
	} `yaml:"rabbitmq"`
	HTTPHandler struct {
		Port string `yaml:"port"`
	} `yaml:"httphandler"`
	Postgres struct {
		URL      string `yaml:"url"`
		Insecure bool   `yaml:"insecure"`
	} `yaml:"postgres"`
}

func NewConfig() (*Config, error) {
	if err := viper.BindEnv("debug"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("rabbitmq.url"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("httphandler.port"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("postgres.url"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("postgres.insecure"); err != nil {
		return nil, err
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var c *Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return c, nil
}
