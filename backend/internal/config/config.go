package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string `mapstructure:"ENVIRONMENT"`
	EnableTls   bool   `mapstructure:"ENABLE_TLS"`
	TlsCert     string `mapstructure:"TLS_CERT"`
	TlsKey      string `mapstructure:"TLS_KEY"`

	DbDriver    string `mapstructure:"DB_DRIVER"`
	DbHost      string `mapstructure:"DB_HOST"`
	DbUser      string `mapstructure:"DB_USER"`
	DbPassword  string `mapstructure:"DB_PASSWORD"`
	DbName      string `mapstructure:"DB_NAME"`
	DbPort      string `mapstructure:"DB_PORT"`
	DbEnableSsl bool   `mapstructure:"DB_ENABLE_SSL"`

	ServerPort string `mapstructure:"SERVER_PORT"`

	KeycloakBaseUrl  string `mapstructure:"KEYCLOAK_BASE_URL"`
	KeycloakRealm    string `mapstructure:"KEYCLOAK_REALM"`
	KeycloakClientId string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	KeycloakSecret   string `mapstructure:"KEYCLOAK_SECRET"`
	KeycloakToken    string `mapstructure:"KEYCLOAK_TOKEN"`

	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func (config *Config) DbSource() string {
	dbPassEscaped := url.QueryEscape(config.DbPassword)

	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", config.DbUser, dbPassEscaped, config.DbHost, config.DbPort, config.DbName, func() string {
		if config.DbEnableSsl {
			return "require"
		}
		return "disable"
	}())
}
