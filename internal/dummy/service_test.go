package dummy

import (
	"errors"
	"testing"

	"ws-dummy-go/internal/dummy/domain"
	"ws-dummy-go/internal/mocks"
)

func Test_userService_Create(t *testing.T) {
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
				sqlRepo.EXPECT().Insert(testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepo.EXPECT().Set(testname).Return(domain.UserID("2"), nil).
					Once()
				docsRepo.EXPECT().Insert(testname).Return(domain.UserID("3"), nil).
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
				sqlRepo.EXPECT().Insert(testname).Return("", mockError).
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
				sqlRepo.EXPECT().Insert(testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepo.EXPECT().Set(testname).Return("", mockError).
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
				sqlRepo.EXPECT().Insert(testname).Return(domain.UserID("1"), nil).
					Once()
				kvRepo.EXPECT().Set(testname).Return(domain.UserID("2"), nil).
					Once()
				docsRepo.EXPECT().Insert(testname).Return("", mockError).
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

			tt.arrange()
			got, err := s.Create(tt.args.name)

			if (err != nil) != tt.wantErr {
				t.Errorf("userService.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("userService.Create() = %v, want %v", got, tt.want)
			}
			kvRepo.AssertExpectations(t)
			sqlRepo.AssertExpectations(t)
			docsRepo.AssertExpectations(t)
		})
	}
}
