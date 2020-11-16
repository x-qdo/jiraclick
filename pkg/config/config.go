package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Debug   bool `yaml:"debug"`
	ClickUp struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
		List  string `yaml:"list"`
	} `yaml:"clickup"`
	RabbitMQ struct {
		URL      string `yaml:"url"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Vhost    string `yaml:"vhost"`
	} `yaml:"rabbitmq"`
	HttpHandler struct {
		Port string `yaml:"port"`
	} `yaml:"httphandler"`
}

func NewConfig(path string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(path)
	viper.BindEnv("rabbitmq.url")
	viper.BindEnv("httphandler.port")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	var c *Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return c, nil
}
