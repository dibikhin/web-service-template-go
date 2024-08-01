package app

import (
	"context"
	"flag"

	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"

	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
		pgPool, ok, teardownPG := connectPostgres(logger, cfg.Postgres)
		if !ok {
			return
		}
		defer teardownPG()

		redisClient, ok, teardownRedis := connectRedis(logger, cfg.Redis)
		if !ok {
			return
		}
		defer teardownRedis()

		mongoClient, ok, teardownMongo := connectMongo(logger, cfg.Mongo)
		if !ok {
			return
		}
		defer teardownMongo()

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
			middleware.MakeDecodeCreateUserRequest(logger),
		),
		httptransport.EncodeJSONResponse,
		httptransport.ServerBefore(middleware.RequestID),
		httptransport.ServerBefore(middleware.RequestLogging(logger, cfg.Mode)),
		httptransport.ServerAfter(middleware.SetRequestID),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(middleware.ErrorEncoder()),
	)

	updateUserHandler := httptransport.NewServer(
		middleware.Recovery(logger)(
			middleware.MakeUpdateUserEndpoint(svc),
		),
		middleware.DecodingRecovery(logger)(
			middleware.MakeDecodeUpdateUserRequest(logger),
		),
		httptransport.EncodeJSONResponse,
		httptransport.ServerBefore(middleware.RequestID),
		httptransport.ServerBefore(middleware.RequestLogging(logger, cfg.Mode)),
		httptransport.ServerAfter(middleware.SetRequestID),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(middleware.ErrorEncoder()),
	)

	http.Handle("/createUser", createUserHandler)
	http.Handle("/updateUser", updateUserHandler)

	http.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr: cfg.Port,
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log("msg", "HTTP", "addr", cfg.Port)

		if err := server.ListenAndServe(); err != nil {
			logger.Log("msg", "listening", "err", err)
		}
		logger.Log("msg", "server cleaning up...")

		sigs <- syscall.SIGUSR1 // Reusing the channel
	}()

	s := <-sigs
	logger.Log("msg", "got signal", "sig", s)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log("msg", "server shutting down", "err", err)
	}
}
