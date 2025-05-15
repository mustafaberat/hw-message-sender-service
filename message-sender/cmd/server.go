package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/oklog/run"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"message-sender/config"
	"message-sender/repository/postgres"
	redisrepo "message-sender/repository/redis"
	"message-sender/service"
	"message-sender/transport/http"
)

func Run() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	logger := initLogger(cfg.Log.Level)
	defer func() {
		_ = logger.Sync()
	}()

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Address,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  cfg.Redis.PoolTimeout,
		MinIdleConns: cfg.Redis.MinIdleConnection,
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	redisRepo := redisrepo.NewRepository(redisClient, &cfg.Redis)

	postgresRepo, err := postgres.NewRepository(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer postgresRepo.Close()

	if err := postgresRepo.InitSchema(context.Background()); err != nil {
		logger.Fatal("Failed to initialize database schema", zap.Error(err))
	}

	messageSvc := service.NewMessageProcessor(
		postgresRepo,
		redisRepo,
		redisRepo,
		logger,
		cfg,
	)

	httpServer := http.NewServer(cfg, logger, messageSvc)

	var g run.Group

	g.Add(
		func() error {
			return httpServer.Start()
		},
		func(err error) {
			if err := httpServer.Stop(); err != nil {
				logger.Error("Failed to stop HTTP server", zap.Error(err))
			}
		},
	)

	g.Add(
		func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			sig := <-c
			return fmt.Errorf("received signal: %v", sig)
		},
		func(err error) {
		},
	)

	logger.Info("Starting message-sender service")
	if err := g.Run(); err != nil {
		logger.Error("Application error", zap.Error(err))

		postgresRepo.Close()
		logger.Fatal("Application error", zap.Error(err))
	}
}

func initLogger(level string) *zap.Logger {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	return logger
}
