package repository

import (
	"context"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VoteRepository struct {
	collection *mongo.Collection
}

func NewVoteRepository(db *mongo.Database) *VoteRepository {
	r := &VoteRepository{
		collection: db.Collection("votes"),
	}

	// Drop the old single-field user_id index if it exists (legacy)
	_, _ = r.collection.Indexes().DropOne(context.Background(), "user_id_1")

	// Ensure compound unique index on (user_id, election_id) for per-election voting
	_, _ = r.collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{"user_id", 1}, {"election_id", 1}},
			Options: options.Index().SetUnique(true),
		},
	)

	return r
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

func (r *VoteRepository) FindByUserAndElectionID(userID string, electionID string) (*models.Vote, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return nil, err
	}

	var vote models.Vote
	err = r.collection.FindOne(
		context.Background(),
		bson.M{"user_id": uid, "election_id": eid},
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

func (r *VoteRepository) CountByElectionID(electionID string) (int64, error) {
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return 0, err
	}
	return r.collection.CountDocuments(context.Background(), bson.M{"election_id": eid})
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

func (r *VoteRepository) FindAllByElectionID(electionID string) ([]*models.Vote, error) {
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(context.Background(), bson.M{"election_id": eid})
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
