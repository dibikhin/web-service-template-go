package app

import (
	"context"
	"fmt"

	"github.com/go-kit/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func composePostgresURL(cfg PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?connect_timeout=%d&pool_max_conns=%d&application_name=%s&sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Timeout, poolMaxConns, appName,
	)
}

func connectPostgres(logger log.Logger, cfg PostgresConfig) (
	pool *pgxpool.Pool, ok bool, teardown func(),
) {
	pool, err := pgxpool.New(context.Background(), composePostgresURL(cfg))
	if err != nil {
		logger.Log("msg", "connecting to postgres", "err", err)
		return nil, false, nil
	}
	if err := pool.Ping(context.Background()); err != nil {
		logger.Log("msg", "pinging postgres", "err", err)
		return nil, false, nil
	}
	s := pool.Stat()
	logger.Log("msg", "postgres pool connected", "total", s.TotalConns(), "max", s.MaxConns())

	return pool, true, func() {
		pool.Close()
		logger.Log("msg", "postgres pool closed")
	}
}

func connectRedis(logger log.Logger, cfg RedisConfig) (
	client *redis.Client, ok bool, teardown func(),
) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           0,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  0, // 0 = 3s
		WriteTimeout: 0, // 0 = 3s
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Log("msg", "pinging redis", "err", err)
		return nil, false, nil
	}
	logger.Log("msg", "redis connected")

	return client, true, func() {
		if err := client.Close(); err != nil {
			logger.Log("msg", "closing redis client", "err", err)
			return
		}
		logger.Log("msg", "redis client closed")
	}
}

func connectMongo(logger log.Logger, cfg MongoConfig) (
	client *mongo.Client, ok bool, teardown func(),
) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)
	client, err := mongo.Connect(context.Background(), options.Client().
		ApplyURI(uri).
		SetConnectTimeout(cfg.Timeout).
		SetServerSelectionTimeout(cfg.Timeout).
		SetTimeout(cfg.Timeout))
	if err != nil {
		logger.Log("msg", "connecting to mongodb", "err", err)
		return nil, false, nil
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		logger.Log("msg", "pinging mongodb", "err", err)
		return nil, false, nil
	}
	logger.Log("msg", "mongodb connected")

	return client, true, func() {
		if err := client.Disconnect(context.Background()); err != nil {
			logger.Log("msg", "disconnecting from mongodb", "err", err)
			return
		}
		logger.Log("msg", "mongodb client disconnected")
	}
}
