package dummy

import (
	"context"
	"errors"
	"testing"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_userService_CreateUser(t *testing.T) {
	kvRepo := &mocks.UsersKVRepo{}
	sqlRepo := &mocks.UsersSQLRepo{}
	docsRepo := &mocks.UsersDocsRepo{}

	s := NewUserService(kvRepo, sqlRepo, docsRepo)

	testname := "testname123"
	mockError := errors.New("mock error")

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		arrange func()
		args    args
		want    domain.UserID
		wantErr bool
	}{
		{
			name: "Positive: Create user",
			arrange: func() {
				sqlRepo.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepo.EXPECT().Set(mock.Anything, testname).Return(domain.UserID("2"), nil).
					Once()
				docsRepo.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("3"), nil).
					Once()
			},
			args: args{
				name: testname,
			},
			want:    domain.UserID("1-2-3"),
			wantErr: false,
		},
		{
			name: "Negative: Creating in sql repo fails",
			arrange: func() {
				sqlRepo.EXPECT().Insert(mock.Anything, testname).Return("", mockError).
					Once()
			},
			args: args{
				name: testname,
			},
			want:    domain.UserID(""),
			wantErr: true,
		},
		{
			name: "Negative: Creating in kv repo fails",
			arrange: func() {
				sqlRepo.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepo.EXPECT().Set(mock.Anything, testname).Return("", mockError).
					Once()
			},
			args: args{
				name: testname,
			},
			want:    domain.UserID(""),
			wantErr: true,
		},
		{
			name: "Negative: Creating in docs repo fails",
			arrange: func() {
				sqlRepo.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepo.EXPECT().Set(mock.Anything, testname).Return(domain.UserID("2"), nil).
					Once()
				docsRepo.EXPECT().Insert(mock.Anything, testname).Return("", mockError).
					Once()
			},
			args: args{
				name: testname,
			},
			want:    domain.UserID(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			tt.arrange()
			got, err := s.CreateUser(context.Background(), tt.args.name)

			assert.Equal(tt.want, got)
			assert.Equal(tt.wantErr, err != nil)

			kvRepo.AssertExpectations(t)
			sqlRepo.AssertExpectations(t)
			docsRepo.AssertExpectations(t)
		})
	}
}
