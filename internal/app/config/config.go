package config

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

// Config Структура конфигурации;
// Содержит все конфигурационные данные о сервисе;
// автоподгружается при изменении исходного файла
type Config struct {
	VKToken    string
	Redis      RedisConfig
	BITOPToken string
	BITOP      BitopConfig
	Bot        BotConfig
}

type BitopConfig struct {
	SiteAdress string
	Protocol   string
	PathPath   string
	PathSearch string
}

type RedisConfig struct {
	// from config file
	TTL         int
	DialTimeout int
	ReadTimeout int
	// from env
	Host     string
	Port     int
	User     string
	Password string
}

type BotConfig struct {
	GroupID string
	ChatID  int
}

// NewConfig Создаёт новый объект конфигурации, загружая данные из файла конфигурации
func NewConfig(ctx context.Context) (*Config, error) {
	var err error
	cfg := &Config{}

	configName := "config"

	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("REDIS_PASSWORD"); token == "" {
		return nil, fmt.Errorf("can't find REDIS_PASSWORD")
	} else {
		cfg.Redis.Password = token
	}

	if token := os.Getenv("REDIS_HOST"); token == "" {
		return nil, fmt.Errorf("can't find REDIS_HOST")
	} else {
		cfg.Redis.Host = token
	}

	if token := os.Getenv("REDIS_PORT"); token == "" {
		return nil, fmt.Errorf("can't find REDIS_PORT")
	} else {
		cfg.Redis.Port, err = strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("redis port must be int value: %w", err)
		}
	}

	if token := os.Getenv("VK_TOKEN"); token == "" {
		return nil, fmt.Errorf("can't find VK_TOKEN")
	} else {
		cfg.VKToken = token
	}

	if token := os.Getenv("BITOP_TOKEN"); token == "" {
		return nil, fmt.Errorf("can't find BITOP_TOKEN")
	} else {
		cfg.BITOPToken = token
	}

	log.Info("config parsed")

	return cfg, nil
}
