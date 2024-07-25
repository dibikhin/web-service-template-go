package dummy

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	username = "mytestuser"
	password = "mytestpassword"
	timeout  = 30 // seconds
)

func TestMain(m *testing.M) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	logger.Log("msg", "connecting to docker")

	// Dockertest
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Log("msg", "connecting to docker", "err", err)
		return
	}
	if err := pool.Client.Ping(); err != nil {
		logger.Log("msg", "pinging docker", "err", err)
		return
	}
	pool.MaxWait = timeout * time.Second

	defer func() {
		logger.Log("msg", "disconnected from docker")
	}()

	// Run images
	redisImage, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bitnami/redis",
		Tag:        "7.0",
		Env: []string{
			"ALLOW_EMPTY_PASSWORD=yes",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Log("msg", "starting redis", "err", err)
		return
	}
	mongoImage, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bitnami/mongodb",
		Tag:        "7.0",
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Log("msg", "starting mongodb", "err", err)
		return
	}
	pgImage, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "bitnami/postgresql",
		Tag:        "16",
		Env: []string{
			"POSTGRESQL_USERNAME=" + username,
			"POSTGRESQL_PASSWORD=" + password,
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		logger.Log("msg", "starting postgres", "err", err)
		return
	}

	redisImage.Expire(timeout)
	mongoImage.Expire(timeout)
	pgImage.Expire(timeout)

	// Connect to images
	if err := pool.Retry(func() error {
		testRedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("localhost:%s", redisImage.GetPort("6379/tcp")),
			Password: password,
			DB:       0,
		})
		return testRedisClient.Ping(context.Background()).Err()
	}); err != nil {
		logger.Log("msg", "connecting to redis", "err", err)
		return
	}
	logger.Log("msg", "connected to redis")

	defer func() {
		if err := testRedisClient.Close(); err != nil {
			logger.Log("msg", "closing redis client", "err", err)
		}
		if err := pool.Purge(redisImage); err != nil {
			logger.Log("msg", "purging redis", "err", err)
		}
	}()

	if err := pool.Retry(func() error {
		uri := fmt.Sprintf("mongodb://localhost:%s", mongoImage.GetPort("27017/tcp"))

		testMongoClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri).
			SetConnectTimeout(1*time.Second).
			SetServerSelectionTimeout(1*time.Second).
			SetTimeout(1*time.Second))

		if err != nil {
			return err
		}
		return testMongoClient.Ping(context.Background(), nil)
	}); err != nil {
		logger.Log("msg", "connecting to mongodb", "err", err)
		return
	}
	logger.Log("msg", "connected to mongodb")

	defer func() {
		if err := testMongoClient.Disconnect(context.Background()); err != nil {
			logger.Log("msg", "disconnecting from mongodb", "err", err)
		}
		if err := pool.Purge(mongoImage); err != nil {
			logger.Log("msg", "purging mongodb", "err", err)
		}
	}()

	pgURL := fmt.Sprintf(
		"postgres://%s:%s@localhost:%s/postgres?connect_timeout=%d&sslmode=disable",
		username, password, pgImage.GetPort("5432/tcp"), 10,
	)
	if err := pool.Retry(func() error {
		testPostgresPool, err = pgxpool.New(context.Background(), pgURL)
		if err != nil {
			return err
		}
		return testPostgresPool.Ping(context.Background())
	}); err != nil {
		logger.Log("msg", "connecting to postgres", "err", err)
		return
	}
	logger.Log("msg", "connected to postgres")

	defer func() {
		testPostgresPool.Close()

		if err := pool.Purge(pgImage); err != nil {
			logger.Log("msg", "purging postgres", "err", err)
		}
	}()

	// Migrate
	mg, err := migrate.New("file://../../migrations", pgURL)
	if err != nil {
		logger.Log("msg", "initializing migrations", "err", err)
		return
	}
	if err := mg.Up(); err != nil {
		logger.Log("msg", "running migrations up", "err", err)
		return
	}

	m.Run()
}
