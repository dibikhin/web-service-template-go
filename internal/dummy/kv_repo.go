package dummy

import (
	"context"
	"fmt"

	"ws-dummy-go/internal/dummy/domain"

	"github.com/redis/go-redis/v9"
)

type UsersKVRepo interface {
	Set(ctx context.Context, name string) (domain.UserID, error)
}

func NewUsersKVRepo(c *redis.Client, g IDGetter) UsersKVRepo {
	return usersKVRepo{
		client:   c,
		idGetter: g,
	}
}

type usersKVRepo struct {
	client   *redis.Client
	idGetter IDGetter
}

func (r usersKVRepo) Set(ctx context.Context, name string) (domain.UserID, error) {
	newID := r.idGetter.GetID()

	if err := r.client.Set(ctx, name, newID, 0).Err(); err != nil {
		return "", fmt.Errorf("setting key: %w", err)
	}
	id, err := r.client.Get(ctx, name).Result()
	if err != nil {
		return "", fmt.Errorf("getting key: %w", err)
	}
	return domain.UserID(id), nil
}
