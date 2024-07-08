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
	Insert(name string) (domain.UserID, error)
}

func NewUsersDocsRepo(c *mongo.Collection) UsersDocsRepo {
	return usersDocRepo{
		col: c,
	}
}

type usersDocRepo struct {
	col *mongo.Collection
}

func (r usersDocRepo) Insert(name string) (domain.UserID, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	res, err := r.col.InsertOne(ctx, bson.D{
		{Key: "name", Value: name}, {Key: "created_at", Value: time.Now()},
	})
	if err != nil {
		return "", fmt.Errorf("inserting a doc: %w", err)
	}
	return domain.UserID(fmt.Sprintf("%v", res.InsertedID)), nil
}
