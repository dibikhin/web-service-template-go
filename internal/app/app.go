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
	"time"

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
			"postgres://%s:%s@%s:%d/%s",
			cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Database,
		)
		pgPool, err := pgxpool.New(context.Background(), pgConnString)
		if err != nil {
			logger.Log("msg", "connecting to postgres", "err", err)
			return
		}
		ctxPG, cancelPG := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelPG()

		if err := pgPool.Ping(ctxPG); err != nil {
			logger.Log("msg", "pinging postgres", "err", err)
			return
		}
		logger.Log("msg", "postgres pool connected")

		defer func() {
			pgPool.Close()
			logger.Log("msg", "postgres pool closed")
		}()

		// Redis
		redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		redisClient := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: cfg.Redis.Password,
			DB:       0,
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
		ctxM, cancelM := context.WithTimeout(context.Background(), cfg.Mongo.Timeout)
		defer cancelM()

		mongoClient, err := mongo.Connect(ctxM, options.Client().ApplyURI(fmt.Sprintf(
			"mongodb://%s:%d/%s", cfg.Mongo.Host, cfg.Mongo.Port, cfg.Mongo.Database)),
		)
		if err != nil {
			logger.Log("msg", "connecting to mongodb", "err", err)
			return
		}
		if err := mongoClient.Ping(context.TODO(), nil); err != nil {
			logger.Log("msg", "pinging mongodb", "err", err)
			return
		}
		logger.Log("msg", "mongodb connected")

		defer func() {
			ctxM1, cancelM1 := context.WithTimeout(context.TODO(), cfg.Mongo.Timeout)
			defer cancelM1()

			if err := mongoClient.Disconnect(ctxM1); err != nil {
				logger.Log("msg", "disconnecting from mongodb", "err", err)
				return
			}
			logger.Log("msg", "mongodb client disconnected")
		}()

		dummyCollection := mongoClient.Database(cfg.Mongo.Database).Collection("users")

		// Repos
		docsRepo := dummy.NewUsersDocsRepo(dummyCollection)
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
		dummymw.MakeLoggingMiddleware(logger, cfg.Mode)(
			dummymw.DecodeCreateUserRequest,
		),
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

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log("msg", "server shutting down", "err", err)
	}
}
