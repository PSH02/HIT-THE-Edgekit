package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv      string    `mapstructure:"app_env" validate:"required,oneof=local dev staging prod"`
	HTTPAddr    string    `mapstructure:"http_addr" validate:"required"`
	GRPCAddr    string    `mapstructure:"grpc_addr" validate:"required"`
	DatabaseURL string    `mapstructure:"database_url" validate:"required"`
	RedisURL    string    `mapstructure:"redis_url" validate:"required"`
	LogLevel    string    `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
	JWT         JWTConfig `mapstructure:"jwt"`
	RateLimit   RateLimit `mapstructure:"rate_limit"`
}

type JWTConfig struct {
	PrivateKeyPath string        `mapstructure:"private_key_path" validate:"required"`
	PublicKeyPath  string        `mapstructure:"public_key_path" validate:"required"`
	AccessTTL      time.Duration `mapstructure:"access_ttl"`
	RefreshTTL     time.Duration `mapstructure:"refresh_ttl"`
}

type RateLimit struct {
	Rate   int `mapstructure:"rate" validate:"required,min=1"`
	Window int `mapstructure:"window" validate:"required,min=1"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvPrefix("EDGEKIT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.JWT.AccessTTL == 0 {
		cfg.JWT.AccessTTL = time.Duration(v.GetInt("jwt.access_ttl")) * time.Second
	}
	if cfg.JWT.RefreshTTL == 0 {
		cfg.JWT.RefreshTTL = time.Duration(v.GetInt("jwt.refresh_ttl")) * time.Second
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app_env", "local")
	v.SetDefault("http_addr", ":8080")
	v.SetDefault("grpc_addr", ":50051")
	v.SetDefault("log_level", "info")
	v.SetDefault("jwt.access_ttl", 3600)
	v.SetDefault("jwt.refresh_ttl", 604800)
	v.SetDefault("rate_limit.rate", 100)
	v.SetDefault("rate_limit.window", 60)
}
