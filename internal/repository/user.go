package repository

import (
	"context"
	"time"

	"github.com/develoFavour/e-voting-system-be/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByMatricNumber finds a user by matriculation number
func (r *UserRepository) FindByMatricNumber(matricNumber string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(
		context.Background(),
		bson.M{"matric_number": matricNumber},
	).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = r.collection.FindOne(
		context.Background(),
		bson.M{"_id": objectID},
	).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateStatus updates user accreditation status
func (r *UserRepository) UpdateStatus(id string, status models.UserStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)

	return err
}

// FindPendingAccreditation gets all pending accreditation requests
func (r *UserRepository) FindPendingAccreditation() ([]*models.User, error) {
	cursor, err := r.collection.Find(
		context.Background(),
		bson.M{"status": models.StatusPending},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var users []*models.User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}

	return users, nil
}

// MarkAsVoted marks a user as having voted (ATOMIC OPERATION)
func (r *UserRepository) MarkAsVoted(userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	// Atomic find and update - prevents double voting
	result := r.collection.FindOneAndUpdate(
		context.Background(),
		bson.M{
			"_id":       objectID,
			"has_voted": false, // Only update if hasn't voted
		},
		bson.M{
			"$set": bson.M{
				"has_voted":  true,
				"updated_at": time.Now(),
			},
		},
	)

	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

// Count returns total number of users
func (r *UserRepository) Count() (int64, error) {
	count, err := r.collection.CountDocuments(context.Background(), bson.M{})
	return count, err
}

// CountByStatus returns number of users with a specific status
func (r *UserRepository) CountByStatus(status models.UserStatus) (int64, error) {
	count, err := r.collection.CountDocuments(context.Background(), bson.M{"status": status})
	return count, err
}

// GetDepartmentStats returns count of users by department
func (r *UserRepository) GetDepartmentStats() ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$department"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}

	cursor, err := r.collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stats []bson.M
	if err = cursor.All(context.Background(), &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetFacultyStats returns count of users by faculty
func (r *UserRepository) GetFacultyStats() ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$faculty"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}

	cursor, err := r.collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stats []bson.M
	if err = cursor.All(context.Background(), &stats); err != nil {
		return nil, err
	}
	return stats, nil
}
