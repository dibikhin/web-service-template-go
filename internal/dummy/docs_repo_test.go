package dummy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"
)

var (
	testMongoClient *mongo.Client
)

func Test_usersDocRepo_Insert(t *testing.T) {
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
			name:    "Positive: Insert user",
			args:    args{name: "testname567"},
			want:    domain.UserID("345678987654"),
			wantErr: false,
		},
	}

	idGeneratorMock := &mocks.IDGenerator{}
	idGeneratorMock.EXPECT().NewID().Return("345678987654").Times(1)

	col := testMongoClient.Database("test_dummy").Collection("users")

	r := usersDocRepo{
		col:         col,
		idGenerator: idGeneratorMock,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			got, err := r.Insert(context.Background(), tt.args.name)

			assert.Equal(tt.want, got)
			assert.Equal(tt.wantErr, err != nil, err)

			idGeneratorMock.AssertExpectations(t)
		})
	}
}

func Test_usersDocRepo_Update(t *testing.T) {
	type args struct {
		id   domain.UserID
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Positive: Update user",
			args:    args{id: domain.UserID("1"), name: "testname123"},
			wantErr: false,
		},
	}

	r := usersDocRepo{
		col: testMongoClient.Database("test_dummy").Collection("users"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			err := r.Update(context.Background(), tt.args.id, tt.args.name)

			assert.Equal(tt.wantErr, err != nil, err)
		})
	}
}
