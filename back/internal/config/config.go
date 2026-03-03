package config

import (
        "fmt"
        "search-job/pkg/postgres"
        "strconv"

        "github.com/spf13/viper"
)

func NewConfig() (*Config, error) {
        viper.SetConfigName("config")
        viper.SetConfigType("yaml")

        // Добавляем пути поиска
        viper.AddConfigPath(".")
        viper.AddConfigPath("..")
        viper.AddConfigPath("../..")

        if err := viper.ReadInConfig(); err != nil {
                return nil, fmt.Errorf("error reading config file: %w", err)
        }

        return &Config{
                IsProd: viper.GetBool("server.isProd"),
                Web: &webParams{
                        Port: viper.GetUint16("server.port"),
                },
                Postgres: &postgres.ConnectionData{
                        User:     viper.GetString("server.pg.user"),
                        Password: viper.GetString("server.pg.password"),
                        Host:     viper.GetString("server.pg.host"),
                        Port:     viper.GetUint16("server.pg.port"),
                        DBName:   viper.GetString("server.pg.database"),
                        SSLMode:  viper.GetString("server.pg.sslmode"),
                },
                External: &ExternalConfig{  // ДОБАВЛЕНО
                        URL:     viper.GetString("server.externalApi.url"),
                        Timeout: viper.GetInt("server.externalApi.timeout"),
                },
        }, nil
}

type Config struct {
        IsProd   bool
        Web      *webParams
        Postgres *postgres.ConnectionData
        External *ExternalConfig  // ДОБАВЛЕНО
}

type webParams struct {
        Port uint16
}

type ExternalConfig struct {  // ДОБАВЛЕНО
        URL     string
        Timeout int
}

func (cfg *Config) GetWebPort() string {
        if cfg == nil || cfg.Web == nil {
                return ""
        }
        return strconv.Itoa(int(cfg.Web.Port))
}