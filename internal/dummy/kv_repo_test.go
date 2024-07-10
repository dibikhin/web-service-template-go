package dummy

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-kit/log"
	"github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"
)

var testRedisClient *redis.Client

func TestMain(m *testing.M) {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Log("msg", "could not connect to docker", "err", err)
		return
	}
	if err := pool.Client.Ping(); err != nil {
		logger.Log("msg", "could not ping docker", "err", err)
		return
	}
	password := "mytestpassword"

	resource, err := pool.Run("bitnami/redis", "latest", []string{"REDIS_PASSWORD=" + password})
	if err != nil {
		logger.Log("msg", "could not start resource", "err", err)
		return
	}
	if err := pool.Retry(func() error {
		testRedisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("localhost:%s", resource.GetPort("6379/tcp")),
			Password: password,
			DB:       0,
		})

		return testRedisClient.Ping(context.Background()).Err()
	}); err != nil {
		logger.Log("msg", "could not connect to redis", "err", err)
	}

	defer func() {
		if err := pool.Purge(resource); err != nil {
			logger.Log("msg", "could not purge resource", "err", err)
		}
	}()

	m.Run()
}

func Test_usersKVRepo_Set(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    domain.UserID
		wantErr bool
	}{
		{
			name:    "Positive: Set user",
			args:    args{name: "testname123"},
			want:    domain.UserID("0987654321"),
			wantErr: false,
		},
	}

	idGeneratorMock := &mocks.IDGenerator{}
	idGeneratorMock.EXPECT().NewID().Return("0987654321").Times(1)

	r := usersKVRepo{
		client:      testRedisClient,
		idGenerator: idGeneratorMock,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			got, err := r.Set(context.Background(), tt.args.name)

			assert.Equal(tt.want, got)
			assert.Equal(tt.wantErr, err != nil)

			idGeneratorMock.AssertExpectations(t)
		})
	}
}
