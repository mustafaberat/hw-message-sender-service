package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	Message  MessageConfig  `mapstructure:"message"`
	Log      LogConfig      `mapstructure:"log"`
	Webhook  WebhookConfig  `mapstructure:"webhook"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
}

type RedisConfig struct {
	Address            string        `mapstructure:"address"`
	DB                 int           `mapstructure:"db"`
	PoolSize           int           `mapstructure:"poolSize"`
	PoolTimeout        time.Duration `mapstructure:"poolTimeout"`
	MinIdleConnection  int           `mapstructure:"minIdleConnection"`
	MessageCacheTTL    time.Duration `mapstructure:"messageCacheTTL"`
	ServiceStatusKey   string        `mapstructure:"serviceStatusKey"`
	SentMessagesPrefix string        `mapstructure:"sentMessagesPrefix"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type MessageConfig struct {
	BatchSize       int           `mapstructure:"batchSize"`
	ProcessInterval time.Duration `mapstructure:"processInterval"`
	MaxContentLen   int           `mapstructure:"maxContentLen"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type WebhookConfig struct {
	URL     string        `mapstructure:"url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

func Parse() (*Config, error) {
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	if err := viper.BindEnv("server.port", "SERVER_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var SERVER_PORT: %w", err)
	}
	if err := viper.BindEnv("server.readTimeout", "SERVER_READ_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var SERVER_READ_TIMEOUT: %w", err)
	}
	if err := viper.BindEnv("server.writeTimeout", "SERVER_WRITE_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var SERVER_WRITE_TIMEOUT: %w", err)
	}

	if err := viper.BindEnv("redis.address", "REDIS_ADDRESS"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_ADDRESS: %w", err)
	}
	if err := viper.BindEnv("redis.db", "REDIS_DB"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_DB: %w", err)
	}
	if err := viper.BindEnv("redis.poolSize", "REDIS_POOL_SIZE"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_POOL_SIZE: %w", err)
	}
	if err := viper.BindEnv("redis.poolTimeout", "REDIS_POOL_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_POOL_TIMEOUT: %w", err)
	}
	if err := viper.BindEnv("redis.minIdleConnection", "REDIS_MIN_IDLE_CONNECTION"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_MIN_IDLE_CONNECTION: %w", err)
	}
	if err := viper.BindEnv("redis.messageCacheTTL", "REDIS_MESSAGE_CACHE_TTL"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_MESSAGE_CACHE_TTL: %w", err)
	}
	if err := viper.BindEnv("redis.serviceStatusKey", "REDIS_SERVICE_STATUS_KEY"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_SERVICE_STATUS_KEY: %w", err)
	}
	if err := viper.BindEnv("redis.sentMessagesPrefix", "REDIS_SENT_MESSAGES_PREFIX"); err != nil {
		return nil, fmt.Errorf("failed to bind env var REDIS_SENT_MESSAGES_PREFIX: %w", err)
	}

	if err := viper.BindEnv("database.host", "DATABASE_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind env var DATABASE_HOST: %w", err)
	}
	if err := viper.BindEnv("database.port", "DATABASE_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var DATABASE_PORT: %w", err)
	}
	if err := viper.BindEnv("database.user", "DATABASE_USER"); err != nil {
		return nil, fmt.Errorf("failed to bind env var DATABASE_USER: %w", err)
	}
	if err := viper.BindEnv("database.password", "DATABASE_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind env var DATABASE_PASSWORD: %w", err)
	}
	if err := viper.BindEnv("database.dbname", "DATABASE_DBNAME"); err != nil {
		return nil, fmt.Errorf("failed to bind env var DATABASE_DBNAME: %w", err)
	}
	if err := viper.BindEnv("database.sslmode", "DATABASE_SSLMODE"); err != nil {
		return nil, fmt.Errorf("failed to bind env var DATABASE_SSLMODE: %w", err)
	}

	if err := viper.BindEnv("message.batchSize", "MESSAGE_BATCH_SIZE"); err != nil {
		return nil, fmt.Errorf("failed to bind env var MESSAGE_BATCH_SIZE: %w", err)
	}
	if err := viper.BindEnv("message.processInterval", "MESSAGE_PROCESS_INTERVAL"); err != nil {
		return nil, fmt.Errorf("failed to bind env var MESSAGE_PROCESS_INTERVAL: %w", err)
	}
	if err := viper.BindEnv("message.maxContentLen", "MESSAGE_MAX_CONTENT_LEN"); err != nil {
		return nil, fmt.Errorf("failed to bind env var MESSAGE_MAX_CONTENT_LEN: %w", err)
	}

	if err := viper.BindEnv("log.level", "LOG_LEVEL"); err != nil {
		return nil, fmt.Errorf("failed to bind env var LOG_LEVEL: %w", err)
	}
	if err := viper.BindEnv("log.format", "LOG_FORMAT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var LOG_FORMAT: %w", err)
	}

	if err := viper.BindEnv("webhook.url", "WEBHOOK_URL"); err != nil {
		return nil, fmt.Errorf("failed to bind env var WEBHOOK_URL: %w", err)
	}
	if err := viper.BindEnv("webhook.timeout", "WEBHOOK_TIMEOUT"); err != nil {
		return nil, fmt.Errorf("failed to bind env var WEBHOOK_TIMEOUT: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
