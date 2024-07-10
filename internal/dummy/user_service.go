package dummy

import (
	"context"
	"fmt"

	"ws-dummy-go/internal/dummy/domain"
)

// UserService provides operations on dummys.
type UserService interface {
	CreateUser(ctx context.Context, name string) (domain.UserID, error)
}

func NewUserService(kv UsersKVRepo, sql UsersSQLRepo, docs UsersDocsRepo) UserService {
	return userService{kv, sql, docs}
}

type userService struct {
	kvRepo   UsersKVRepo
	sqlRepo  UsersSQLRepo
	docsRepo UsersDocsRepo
}

func (s userService) CreateUser(ctx context.Context, name string) (domain.UserID, error) {
	id1, err := s.sqlRepo.Insert(ctx, name)
	if err != nil {
		return "", fmt.Errorf("creating user in sql repo: %w", err)
	}
	id2, err := s.kvRepo.Set(ctx, name)
	if err != nil {
		return "", fmt.Errorf("creating user in kv repo: %w", err)
	}
	id3, err := s.docsRepo.Insert(ctx, name)
	if err != nil {
		return "", fmt.Errorf("creating user in docs repo: %w", err)
	}
	return domain.UserID(fmt.Sprintf("%s-%s-%s", id1, id2, id3)), nil
}
