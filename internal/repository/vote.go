package repository

import (
	"context"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type VoteRepository struct {
	collection *mongo.Collection
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	return &VoteRepository{
		collection: db.Collection("votes"),
	}
}

// Create creates a new vote record
func (r *VoteRepository) Create(vote *models.Vote) error {
	vote.Timestamp = time.Now()

	result, err := r.collection.InsertOne(context.Background(), vote)
	if err != nil {
		return err
	}

	vote.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByUserID finds a vote by user ID
func (r *VoteRepository) FindByUserID(userID string) (*models.Vote, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var vote models.Vote
	err = r.collection.FindOne(
		context.Background(),
		bson.M{"user_id": objectID},
	).Decode(&vote)

	if err != nil {
		return nil, err
	}

	return &vote, nil
}

// Count returns total number of votes cast
func (r *VoteRepository) Count() (int64, error) {
	return r.collection.CountDocuments(context.Background(), bson.M{})
}

// FindAll gets all votes (for admin results)
func (r *VoteRepository) FindAll() ([]*models.Vote, error) {
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var votes []*models.Vote
	if err = cursor.All(context.Background(), &votes); err != nil {
		return nil, err
	}

	return votes, nil
}
