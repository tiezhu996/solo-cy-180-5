// Package config 集中解析环境变量配置。
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config 保存服务运行所需的全部配置。
type Config struct {
	ServerPort   string `env:"SERVER_PORT" envDefault:"8080"`
	RunMode      string `env:"RUN_MODE" envDefault:"release"`
	DBHost       string `env:"DB_HOST" envDefault:"127.0.0.1"`
	DBPort       string `env:"DB_PORT" envDefault:"3306"`
	DBName       string `env:"DB_NAME" envDefault:"oralhistory_db"`
	DBUser       string `env:"DB_USER" envDefault:"oralhistory_user"`
	DBPassword   string `env:"DB_PASSWORD" envDefault:"oralhistory_pwd"`
	JWTSecret    string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTExpireH   int    `env:"JWT_EXPIRE_HOURS" envDefault:"72"`
	RedisAddr    string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	RedisPass    string `env:"REDIS_PASSWORD" envDefault:""`
	RedisDB      int    `env:"REDIS_DB" envDefault:"0"`
	MinIOEndpoint  string `env:"MINIO_ENDPOINT" envDefault:"127.0.0.1:9000"`
	MinIOAccessKey string `env:"MINIO_ACCESS_KEY" envDefault:"minioadmin"`
	MinIOSecretKey string `env:"MINIO_SECRET_KEY" envDefault:"minioadmin"`
	MinIOBucket    string `env:"MINIO_BUCKET" envDefault:"oralhistory-audio"`
	MinIOUseSSL    bool   `env:"MINIO_USE_SSL" envDefault:"false"`
}

// Load 从环境变量加载配置。
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

// DSN 返回 MySQL 连接串。
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
