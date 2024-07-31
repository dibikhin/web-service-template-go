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
	Update(ctx context.Context, id domain.UserID, name string) error
}

func NewUsersDocsRepo(c *mongo.Collection, g IDGenerator) UsersDocsRepo {
	return usersDocRepo{
		col:         c,
		idGenerator: g,
	}
}

type usersDocRepo struct {
	col         *mongo.Collection
	idGenerator IDGenerator
}

func (r usersDocRepo) Insert(ctx context.Context, name string) (domain.UserID, error) {
	newID := r.idGenerator.NewID()

	res, err := r.col.InsertOne(ctx, bson.D{
		{Key: "_id", Value: newID},
		{Key: "name", Value: name},
		{Key: "created_at", Value: time.Now()},
	})
	if err != nil {
		return "", fmt.Errorf("inserting a doc: %w", err)
	}
	return domain.UserID(fmt.Sprintf("%v", res.InsertedID)), nil
}

func (r usersDocRepo) Update(ctx context.Context, id domain.UserID, name string) error {
	if _, err := r.col.UpdateOne(ctx, bson.D{
		{Key: "_id", Value: string(id)},
	}, bson.D{
		{Key: "$set", Value: bson.D{{Key: "name", Value: name}}},
	}); err != nil {
		return fmt.Errorf("updating a doc: %w", err)
	}
	return nil
}
