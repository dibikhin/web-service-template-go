package dummy

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"

	"github.com/go-kit/log"
)

var testRedisClient *redis.Client

func TestMain(m *testing.M) {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stderr)
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
			name: "Positive: Set user",
			want: domain.UserID("0987654321"),
		},
	}

	idGetterMock := &mocks.IDGetter{}
	idGetterMock.EXPECT().GetID().Return("0987654321").Times(1)

	r := usersKVRepo{
		client:   testRedisClient,
		idGetter: idGetterMock,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Set(context.Background(), tt.args.name)

			if (err != nil) != tt.wantErr {
				t.Errorf("usersKVRepo.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("usersKVRepo.Set() = %v, want %v", got, tt.want)
			}
			idGetterMock.AssertExpectations(t)
		})
	}
}
