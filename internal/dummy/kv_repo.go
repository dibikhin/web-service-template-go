package dummy

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"ws-dummy-go/internal/dummy/domain"
)

type UsersKVRepo interface {
	Set(ctx context.Context, name string) (domain.UserID, error)
}

func NewUsersKVRepo(c *redis.Client, g IDGenerator) UsersKVRepo {
	return usersKVRepo{
		client:      c,
		idGenerator: g,
	}
}

type usersKVRepo struct {
	client      *redis.Client
	idGenerator IDGenerator
}

func (r usersKVRepo) Set(ctx context.Context, name string) (domain.UserID, error) {
	newID := r.idGenerator.NewID()

	if err := r.client.Set(ctx, name, newID, 0).Err(); err != nil {
		return "", fmt.Errorf("setting key: %w", err)
	}
	id, err := r.client.Get(ctx, name).Result()
	if err != nil {
		if err == redis.Nil {
			return "", domain.NewNotFoundError("user not found")
		}
		return "", fmt.Errorf("getting key: %w", err)
	}
	return domain.UserID(id), nil
}
