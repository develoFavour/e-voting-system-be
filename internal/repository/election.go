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

type ElectionRepository struct {
	collection *mongo.Collection
}

func NewElectionRepository(db *mongo.Database) *ElectionRepository {
	return &ElectionRepository{
		collection: db.Collection("elections"),
	}
}

// Create creates a new election
func (r *ElectionRepository) Create(election *models.Election) error {
	election.CreatedAt = time.Now()
	election.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(context.Background(), election)
	if err != nil {
		return err
	}

	election.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetCurrentElection returns the currently active or most recent election
func (r *ElectionRepository) GetCurrentElection() (*models.Election, error) {
	var election models.Election

	// Find the most recent election (sorted by created_at desc)
	err := r.collection.FindOne(
		context.Background(),
		bson.M{},
		options.FindOne().SetSort(bson.D{{"created_at", -1}}),
	).Decode(&election)

	if err != nil {
		return nil, err
	}

	return &election, nil
}

// GetLiveElection returns the currently live election (status=live)
func (r *ElectionRepository) GetLiveElection() (*models.Election, error) {
	var election models.Election

	err := r.collection.FindOne(
		context.Background(),
		bson.M{"status": models.ElectionStatusLive},
		options.FindOne().SetSort(bson.D{{"created_at", -1}}),
	).Decode(&election)
	if err != nil {
		return nil, err
	}

	return &election, nil
}

// UpdateStatus updates the election status
func (r *ElectionRepository) UpdateStatus(id string, status models.ElectionStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Add timestamps based on status
	if status == models.ElectionStatusLive {
		update["$set"].(bson.M)["start_time"] = time.Now()
	} else if status == models.ElectionStatusClosed {
		update["$set"].(bson.M)["end_time"] = time.Now()
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		update,
	)

	return err
}

// StartElection starts an election with duration
func (r *ElectionRepository) StartElection(id string, duration int) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"status":     models.ElectionStatusLive,
				"start_time": startTime,
				"end_time":   endTime,
				"duration":   duration,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}

// GetAllElections returns all elections, sorted by creation date (newest first)
func (r *ElectionRepository) GetAllElections() ([]*models.Election, error) {
	var elections []*models.Election

	cursor, err := r.collection.Find(
		context.Background(),
		bson.M{},
		options.Find().SetSort(bson.D{{"created_at", -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var election models.Election
		if err := cursor.Decode(&election); err != nil {
			continue // Skip invalid documents
		}
		elections = append(elections, &election)
	}

	return elections, nil
}

// GetElectionById returns a specific election by ID
func (r *ElectionRepository) GetElectionById(id string) (*models.Election, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var election models.Election
	err = r.collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&election)

	if err != nil {
		return nil, err
	}

	return &election, nil
}

// CheckAndEndExpiredElections checks if any live elections have expired and ends them
func (r *ElectionRepository) CheckAndEndExpiredElections() error {
	now := time.Now()

	// Find all live elections that should have ended
	filter := bson.M{
		"status":   models.ElectionStatusLive,
		"end_time": bson.M{"$lte": now},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     models.ElectionStatusClosed,
			"end_time":   now,
			"updated_at": now,
		},
	}

	_, err := r.collection.UpdateMany(
		context.Background(),
		filter,
		update,
	)

	return err
}
