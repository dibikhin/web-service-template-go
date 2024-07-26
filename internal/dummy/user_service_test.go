package dummy

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"
)

func Test_userService_CreateUser(t *testing.T) {
	kvRepoMock := &mocks.UsersKVRepo{}
	sqlRepoMock := &mocks.UsersSQLRepo{}
	docsRepoMock := &mocks.UsersDocsRepo{}

	s := NewUserService(kvRepoMock, sqlRepoMock, docsRepoMock)

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
				sqlRepoMock.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepoMock.EXPECT().Set(mock.Anything, testname).Return(domain.UserID("2"), nil).
					Once()
				docsRepoMock.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("3"), nil).
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
				sqlRepoMock.EXPECT().Insert(mock.Anything, testname).Return("", mockError).
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
				sqlRepoMock.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepoMock.EXPECT().Set(mock.Anything, testname).Return("", mockError).
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
				sqlRepoMock.EXPECT().Insert(mock.Anything, testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepoMock.EXPECT().Set(mock.Anything, testname).Return(domain.UserID("2"), nil).
					Once()
				docsRepoMock.EXPECT().Insert(mock.Anything, testname).Return("", mockError).
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
			assert.Equal(tt.wantErr, err != nil, err)

			kvRepoMock.AssertExpectations(t)
			sqlRepoMock.AssertExpectations(t)
			docsRepoMock.AssertExpectations(t)
		})
	}
}
