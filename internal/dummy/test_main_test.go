package dummy

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMain(m *testing.M) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Log("msg", "connecting to docker", "err", err)
		return
	}
	if err := pool.Client.Ping(); err != nil {
		logger.Log("msg", "pinging docker", "err", err)
		return
	}
	password := "mytestpassword"

	redisImage, err := pool.Run("bitnami/redis", "7.0", []string{"REDIS_PASSWORD=" + password})
	if err != nil {
		logger.Log("msg", "starting redis", "err", err)
		return
	}
	mongoImage, err := pool.Run("bitnami/mongodb", "7.0", nil)
	if err != nil {
		logger.Log("msg", "starting mongodb", "err", err)
		return
	}

	if err := pool.Retry(func() error {
		testRedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("localhost:%s", redisImage.GetPort("6379/tcp")),
			Password: password,
			DB:       0,
		})
		return testRedisClient.Ping(context.Background()).Err()
	}); err != nil {
		logger.Log("msg", "connecting to redis", "err", err)
	}
	if err := pool.Retry(func() error {
		testMongoClient, err = mongo.Connect(context.Background(), options.Client().
			ApplyURI(fmt.Sprintf("mongodb://localhost:%s", mongoImage.GetPort("27017/tcp"))).
			SetConnectTimeout(1*time.Second).
			SetServerSelectionTimeout(1*time.Second).
			SetTimeout(1*time.Second))
		if err != nil {
			return err
		}
		return testMongoClient.Ping(context.Background(), nil)
	}); err != nil {
		logger.Log("msg", "connecting to mongodb", "err", err)
	}

	defer func() {
		if err := testRedisClient.Close(); err != nil {
			logger.Log("msg", "closing redis client", "err", err)
		}
		if err := testMongoClient.Disconnect(context.Background()); err != nil {
			logger.Log("msg", "disconnecting from mongodb", "err", err)
		}

		if err := pool.Purge(redisImage); err != nil {
			logger.Log("msg", "purging redis", "err", err)
		}
		if err := pool.Purge(mongoImage); err != nil {
			logger.Log("msg", "purging mongodb", "err", err)
		}
	}()

	m.Run()
}
