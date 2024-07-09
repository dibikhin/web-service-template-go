package dummy

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"ws-dummy-go/internal/dummy/domain"
)

type UsersDocsRepo interface {
	Insert(ctx context.Context, name string) (domain.UserID, error)
}

func NewUsersDocsRepo(c *mongo.Collection, t time.Duration) UsersDocsRepo {
	return usersDocRepo{
		col:     c,
		timeout: t,
	}
}

type usersDocRepo struct {
	col     *mongo.Collection
	timeout time.Duration
}

func (r usersDocRepo) Insert(ctx context.Context, name string) (domain.UserID, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.col.InsertOne(ctx, bson.D{
		{Key: "name", Value: name}, {Key: "created_at", Value: time.Now()},
	})
	if err != nil {
		return "", fmt.Errorf("inserting a doc: %w", err)
	}
	return domain.UserID(fmt.Sprintf("%v", res.InsertedID)), nil
}