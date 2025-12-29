package repository

import (
	"context"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ActivityRepository struct {
	collection *mongo.Collection
}

func NewActivityRepository(db *mongo.Database) *ActivityRepository {
	return &ActivityRepository{
		collection: db.Collection("activities"),
	}
}

func (r *ActivityRepository) Create(activity *models.Activity) error {
	activity.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(context.Background(), activity)
	return err
}

func (r *ActivityRepository) FindRecent(limit int64) ([]*models.Activity, error) {
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var activities []*models.Activity
	if err := cursor.All(context.Background(), &activities); err != nil {
		return nil, err
	}

	return activities, nil
}
