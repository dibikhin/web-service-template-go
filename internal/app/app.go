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
	"ws-dummy-go/internal/dummy/middleware"
)

const (
	appName      = "ws-dummy-go"
	poolMaxConns = 10
)

func Run() {
	file := flag.String("config", "dev.env", "config file")
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
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
		Namespace: "dummy_group",
		Subsystem: "ws_dummy_go",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys)

	requestLatency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "dummy_group",
		Subsystem: "ws_dummy_go",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, fieldKeys)

	var svc dummy.UserService
	{
		// Postgres
		pgPool, err := pgxpool.New(context.Background(), composePostgresURL(cfg.Postgres))
		if err != nil {
			logger.Log("msg", "connecting to postgres", "err", err)
			return
		}
		if err := pgPool.Ping(context.Background()); err != nil {
			logger.Log("msg", "pinging postgres", "err", err)
			return
		}
		s := pgPool.Stat()
		logger.Log("msg", "postgres pool connected", "total", s.TotalConns(), "max", s.MaxConns())

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
				logger.Log("msg", "closing redis client", "err", err)
				return
			}
			logger.Log("msg", "redis client closed")
		}()

		// Mongo
		mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			cfg.Mongo.Username, cfg.Mongo.Password, cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.Database,
		)
		mongoClient, err := mongo.Connect(context.Background(), options.Client().
			ApplyURI(mongoURI).
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
		docsRepo := dummy.NewUsersDocsRepo(dummyCollection, dummy.NewRandIDGenerator())
		kvRepo := dummy.NewUsersKVRepo(redisClient, dummy.NewRandIDGenerator())
		sqlRepo := dummy.NewUsersSQLRepo(pgPool)

		svc = dummy.NewUserService(kvRepo, sqlRepo, docsRepo)
	}

	svc = middleware.NewLoggingMiddleware(logger)(svc)
	svc = middleware.NewInstrumentingMiddleware(requestCount, requestLatency)(svc)

	createUserHandler := httptransport.NewServer(
		middleware.Recovery(logger)(
			middleware.MakeCreateUserEndpoint(svc),
		),
		middleware.DecodingRecovery(logger)(
			middleware.DecodeCreateUserRequest,
		),
		httptransport.EncodeJSONResponse,
		httptransport.ServerBefore(middleware.RequestID),
		httptransport.ServerBefore(middleware.RequestLogging(logger, cfg.Mode)),
		httptransport.ServerAfter(middleware.SetRequestID),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(middleware.ErrorEncoder()),
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

func composePostgresURL(cfg PostgresConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?connect_timeout=%d&pool_max_conns=%d&application_name=%s&sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Timeout, poolMaxConns, appName,
	)
}
