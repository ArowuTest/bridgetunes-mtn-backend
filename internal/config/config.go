package config

import (
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	JWT      JWTConfig
	MTN      MTNConfig
	SMS      SMSConfig
	LogLevel string
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         string
	AllowedHosts []string
}

// MongoDBConfig holds MongoDB-specific configuration
type MongoDBConfig struct {
	URI      string
	Database string
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret    string
	ExpiresIn int
}

// MTNConfig holds MTN API-specific configuration
type MTNConfig struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	MockAPI    bool
}

// SMSConfig holds SMS gateway-specific configuration
type SMSConfig struct {
	MTNGateway      MTNGatewayConfig
	KodobeGateway   KodobeGatewayConfig
	DefaultGateway  string
	MockSMSGateway  bool
}

// MTNGatewayConfig holds MTN SMS gateway-specific configuration
type MTNGatewayConfig struct {
	BaseURL    string
	APIKey     string
	APISecret  string
}

// KodobeGatewayConfig holds Kodobe SMS gateway-specific configuration
type KodobeGatewayConfig struct {
	BaseURL  string
	APIKey   string
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file is not found, we'll use environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	viper.SetDefault("Server.Port", "4000")
	viper.SetDefault("Server.AllowedHosts", []string{"localhost:3000"})
	viper.SetDefault("MongoDB.URI", "mongodb+srv://fsanus20111:wXVTvRfaCtcd5W7t@cluster0.llhkakp.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	viper.SetDefault("MongoDB.Database", "bridgetunes-mtn")
	viper.SetDefault("JWT.ExpiresIn", 24*60*60) // 24 hours
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("MTN.MockAPI", true)
	viper.SetDefault("SMS.DefaultGateway", "mtn")
	viper.SetDefault("SMS.MockSMSGateway", true)
}
