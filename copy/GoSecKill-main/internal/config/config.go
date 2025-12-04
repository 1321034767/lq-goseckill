package config

import (
	"github.com/spf13/viper"
)

var config *AppConfig

// AppConfig holds the application configuration
type AppConfig struct {
	// Application specific configurations

	// LogConfig holds the log configuration
	LogConfig LogConfig `mapstructure:"log"`

	// Server holds the server configuration
	Server Server `mapstructure:"server"`

	// Database holds the database configuration
	Database Database `mapstructure:"database"`

	// RabbitMQ holds the RabbitMQ configuration
	RabbitMQ RabbitMQ `mapstructure:"rabbitmq"`
}

type LogConfig struct {
	DeveloperMode   bool `mapstructure:"developerMode"`
	OutputToConsole bool `mapstructure:"outputToConsole"`
}

type Server struct {
	AdminPort  string `mapstructure:"adminPort"`
	ServerPort string `mapstructure:"ServerPort"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RabbitMQ struct {
	url string `mapstructure:"url"`
}

// LoadConfig loads application configuration from file and environment variables.
func LoadConfig(configPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configPath)

	// Read configuration from the file
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	// Override with environment variables
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&config); err != nil {
		return err
	}

	return nil
}
