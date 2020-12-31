package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Debug   bool `yaml:"debug"`
	ClickUp struct {
		Host          string `yaml:"host"`
		Token         string `yaml:"token"`
		List          string `yaml:"list"`
		WebhookSecret string `yaml:"webhooksecret"`
	} `yaml:"clickup"`
	Jira     map[string]JiraInstance `yaml:"jira"`
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
}

type JiraInstance struct {
	Username string `yaml:"username"`
	APIToken string `yaml:"apitoken"`
	BaseURL  string `yaml:"baseurl"`
	Project  string `yaml:"project"`
}

func NewConfig(path string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/jiraclick/")
	viper.AddConfigPath(path)
	if err := viper.BindEnv("debug"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("rabbitmq.url"); err != nil {
		return nil, err
	}
	if err := viper.BindEnv("httphandler.port"); err != nil {
		return nil, err
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		// Handle errors reading the config file
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	var c *Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return c, nil
}
