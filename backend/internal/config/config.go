package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment        string `mapstructure:"ENVIRONMENT"`
	EnableTls          bool   `mapstructure:"ENABLE_TLS"`
	TlsCert            string `mapstructure:"TLS_CERT"`
	TlsKey             string `mapstructure:"TLS_KEY"`
	KeycloakCACertPath string `mapstructure:"TLS_KEYCLOAK_CA_CERT_PATH"`

	DbDriver    string `mapstructure:"DB_DRIVER"`
	DbHost      string `mapstructure:"DB_HOST"`
	DbUser      string `mapstructure:"DB_USER"`
	DbPassword  string `mapstructure:"DB_PASSWORD"`
	DbName      string `mapstructure:"DB_NAME"`
	DbPort      string `mapstructure:"DB_PORT"`
	DbEnableSsl bool   `mapstructure:"DB_ENABLE_SSL"`

	ServerPort string `mapstructure:"SERVER_PORT"`

	KeycloakBaseUrl      string `mapstructure:"KEYCLOAK_BASE_URL"`
	KeycloakRealm        string `mapstructure:"KEYCLOAK_REALM"`
	KeycloakClientID     string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	KeycloakClientSecret string `mapstructure:"KEYCLOAK_SECRET"`
	KeycloakToken        string `mapstructure:"KEYCLOAK_TOKEN"`
	KeycloakRedirectUri  string `mapstructure:"KEYCLOAK_REDIRECT_URI"`

	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("SERVER_PORT", "9000")
	viper.SetDefault("DB_DRIVER", "postgres")
	viper.SetDefault("DB_ENABLE_SSL", false)
	viper.SetDefault("ACCESS_TOKEN_DURATION", "15m")
	viper.SetDefault("REFRESH_TOKEN_DURATION", "168h")

	if path != "" {
		viper.SetConfigName("app")
		viper.SetConfigType("env")
		viper.AddConfigPath(path)

		// DO NOT fail if missing
		_ = viper.ReadInConfig()
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (config *Config) DbSource() string {
	password := ""
	if config.DbPassword != "" {
		password = url.QueryEscape(config.DbPassword)
	}

	sslMode := "disable"
	if config.DbEnableSsl {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		config.DbUser,
		password,
		config.DbHost,
		config.DbPort,
		config.DbName,
		sslMode,
	)
}

func (c *Config) Validate() error {
	if c.DbHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.KeycloakBaseUrl == "" {
		return fmt.Errorf("KEYCLOAK_BASE_URL is required")
	}
	if _, err := url.Parse(c.KeycloakBaseUrl); err != nil {
		return fmt.Errorf("invalid KEYCLOAK_BASE_URL")
	}
	return nil
}
