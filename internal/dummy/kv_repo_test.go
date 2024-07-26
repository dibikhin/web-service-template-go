package dummy

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"
)

var (
	testRedisClient *redis.Client
)

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
			// TODO: t.Parallel()
			assert := assert.New(t)

			got, err := r.Set(context.Background(), tt.args.name)

			assert.Equal(tt.want, got)
			assert.Equal(tt.wantErr, err != nil, err)

			idGeneratorMock.AssertExpectations(t)
		})
	}
}
