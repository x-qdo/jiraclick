package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const ServiceName = "jiraclick"

var envBindings = []string{
	"debug",
	"rabbitmq.url",
	"httphandler.port",
	"metrics.port",
	"postgres.url",
	"postgres.insecure",
	"otel.exporter.endpoint",
}

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
	Metrics struct {
		Port string `yaml:"port"`
	} `yaml:"metrics"`
	OTel struct {
		Exporter struct {
			Endpoint string `yaml:"endpoint"`
		} `yaml:"exporter"`
	} `yaml:"otel"`
}

func NewConfig() (*Config, error) {
	if err := bindEnvs(envBindings); err != nil {
		return nil, err
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var c *Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return c, nil
}

func bindEnvs(envKeys []string) error {
	for _, envKey := range envKeys {
		if err := viper.BindEnv(envKey); err != nil {
			return fmt.Errorf("error binding '%s': %w", envKey, err)
		}
	}

	return nil
}
