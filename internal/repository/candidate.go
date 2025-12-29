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

func (r *CandidateRepository) FindUnscoped() ([]*models.Candidate, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"election_id": bson.M{"$exists": false}},
			{"election_id": primitive.NilObjectID},
		},
	}

	cursor, err := r.collection.Find(
		context.Background(),
		filter,
		options.Find().SetSort(bson.D{{"created_at", -1}}),
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

func (r *CandidateRepository) FindByElectionID(electionID string) ([]*models.Candidate, error) {
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(context.Background(), bson.M{"election_id": eid})
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

func (r *CandidateRepository) FindByElectionAndPosition(electionID string, position string) ([]*models.Candidate, error) {
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.collection.Find(
		context.Background(),
		bson.M{"election_id": eid, "position": position},
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

func (r *CandidateRepository) FindByID(id string) (*models.Candidate, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var candidate models.Candidate
	err = r.collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&candidate)
	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

func (r *CandidateRepository) AttachUnscopedToElection(electionID string) error {
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"$or": []bson.M{
			{"election_id": bson.M{"$exists": false}},
			{"election_id": primitive.NilObjectID},
		},
	}

	_, err = r.collection.UpdateMany(
		context.Background(),
		filter,
		bson.M{"$set": bson.M{"election_id": eid}},
	)
	return err
}

func (r *CandidateRepository) SetPositionID(candidateID string, positionID primitive.ObjectID) error {
	objectID, err := primitive.ObjectIDFromHex(candidateID)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"position_id": positionID}},
	)
	return err
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

// Count returns total number of candidates
func (r *CandidateRepository) Count() (int64, error) {
	count, err := r.collection.CountDocuments(context.Background(), bson.M{})
	return count, err
}
