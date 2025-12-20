package repository

import (
	"context"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CandidateRepository struct {
	collection *mongo.Collection
}

func NewCandidateRepository(db *mongo.Database) *CandidateRepository {
	return &CandidateRepository{
		collection: db.Collection("candidates"),
	}
}

// Create creates a new candidate
func (r *CandidateRepository) Create(candidate *models.Candidate) error {
	candidate.CreatedAt = time.Now()
	candidate.VoteCount = 0

	result, err := r.collection.InsertOne(context.Background(), candidate)
	if err != nil {
		return err
	}

	candidate.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindAll gets all candidates
func (r *CandidateRepository) FindAll() ([]*models.Candidate, error) {
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var candidates []*models.Candidate
	if err = cursor.All(context.Background(), &candidates); err != nil {
		return nil, err
	}

	return candidates, nil
}

// FindByPosition gets candidates for a specific position
func (r *CandidateRepository) FindByPosition(position string) ([]*models.Candidate, error) {
	cursor, err := r.collection.Find(
		context.Background(),
		bson.M{"position": position},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var candidates []*models.Candidate
	if err = cursor.All(context.Background(), &candidates); err != nil {
		return nil, err
	}

	return candidates, nil
}

// IncrementVoteCount increments vote count for a candidate
func (r *CandidateRepository) IncrementVoteCount(candidateID string) error {
	objectID, err := primitive.ObjectIDFromHex(candidateID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$inc": bson.M{"vote_count": 1}},
	)

	return err
}
