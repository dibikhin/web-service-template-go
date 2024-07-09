package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/jackc/pgx/v5/pgxpool"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ws-dummy-go/internal/dummy"
	dummymw "ws-dummy-go/internal/dummy/middleware"
)

const AppName = "ws-dummy-go"

func Run() {
	var (
		file = flag.String("config", "dev.env", "config file")
	)
	flag.Parse()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	logger.Log("msg", "server starting...")
	defer logger.Log("msg", "server shut down")

	cfg, err := LoadConfig(*file)
	if err != nil {
		logger.Log("msg", "loading config", "err", err)
		return
	}
	logger.Log("msg", "config loaded")

	fieldKeys := []string{"method", "error"}

	requestCount := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "my_group",
		Subsystem: "dummy_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "my_group",
		Subsystem: "dummy_service",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	var svc dummy.UserService
	{
		// Postgres
		pgConnString := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?connect_timeout=%d&pool_max_conns=%d&application_name=%s",
			cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port,
			cfg.Postgres.Database, cfg.Postgres.Timeout, 10, AppName,
		)
		pgPool, err := pgxpool.New(context.Background(), pgConnString)
		if err != nil {
			logger.Log("msg", "connecting to postgres", "err", err)
			return
		}
		if err := pgPool.Ping(context.Background()); err != nil {
			logger.Log("msg", "pinging postgres", "err", err)
			return
		}
		logger.Log("msg", "postgres pool connected")
		s := pgPool.Stat()
		logger.Log("msg", "pool conns", "total", s.TotalConns(), "max", s.MaxConns(), "acquired", s.AcquiredConns())

		defer func() {
			pgPool.Close()
			logger.Log("msg", "postgres pool closed")
		}()

		// Redis
		redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		redisClient := redis.NewClient(&redis.Options{
			Addr:         redisAddr,
			Password:     cfg.Redis.Password,
			DB:           0,
			DialTimeout:  cfg.Redis.Timeout,
			ReadTimeout:  0, // 0 = 3s
			WriteTimeout: 0, // 0 = 3s
		})
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			logger.Log("msg", "pinging redis", "err", err)
			return
		}
		logger.Log("msg", "redis connected")

		defer func() {
			if err := redisClient.Close(); err != nil {
				logger.Log("msg", "closing redis connection", "err", err)
				return
			}
			logger.Log("msg", "redis client closed")
		}()

		// Mongo
		mongoConnURI := fmt.Sprintf("mongodb://%s:%d/%s", cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.Database)

		mongoClient, err := mongo.Connect(context.Background(), options.Client().
			ApplyURI(mongoConnURI).
			SetConnectTimeout(cfg.Mongo.Timeout).
			SetServerSelectionTimeout(cfg.Mongo.Timeout).
			SetTimeout(cfg.Mongo.Timeout))
		if err != nil {
			logger.Log("msg", "connecting to mongodb", "err", err)
			return
		}
		if err := mongoClient.Ping(context.Background(), nil); err != nil {
			logger.Log("msg", "pinging mongodb", "err", err)
			return
		}
		logger.Log("msg", "mongodb connected")

		defer func() {
			if err := mongoClient.Disconnect(context.Background()); err != nil {
				logger.Log("msg", "disconnecting from mongodb", "err", err)
				return
			}
			logger.Log("msg", "mongodb client disconnected")
		}()

		dummyCollection := mongoClient.Database(cfg.Mongo.Database).Collection("users")

		// Repos
		docsRepo := dummy.NewUsersDocsRepo(dummyCollection, cfg.Mongo.Timeout)
		kvRepo := dummy.NewUsersKVRepo(redisClient, dummy.NewIDGetter())
		sqlRepo := dummy.NewUsersSQLRepo(pgPool)

		svc = dummy.NewUserService(kvRepo, sqlRepo, docsRepo)
	}

	svc = dummymw.NewLoggingMiddleware(logger)(svc)
	svc = dummymw.NewInstrumentingMiddleware(requestCount, requestLatency)(svc)

	logErrHandler := httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger))

	// todord propagate request id, return it

	createUserHandler := httptransport.NewServer(
		dummymw.RecoveryMiddleware(logger)(
			dummymw.MakeCreateUserEndpoint(svc),
		),
		dummymw.RequestIDMiddleware()(
			dummymw.MakeLoggingMiddleware(logger, cfg.Mode)(
				dummymw.DecodeCreateUserRequest,
			)),
		httptransport.EncodeJSONResponse,
		logErrHandler,
	)

	server := &http.Server{
		Addr: cfg.Port,
	}
	http.Handle("/createUser", createUserHandler)
	http.Handle("/metrics", promhttp.Handler())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log("msg", "HTTP", "addr", cfg.Port)

		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Log("msg", "server closed")
			} else {
				logger.Log("err", err)
			}
		}
		sigs <- syscall.SIGUSR1 // Just reusing the channel
	}()

	s := <-sigs
	logger.Log("msg", "got signal", "signal", s)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log("msg", "server shutting down", "err", err)
	}
}
