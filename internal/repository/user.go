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
		options.FindOne().SetCollation(&options.Collation{
			Locale:   "en",
			Strength: 2,
		}),
	).Decode(&user)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// FindByEmail finds a user by email address
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(
		context.Background(),
		bson.M{"email": email},
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

// FindByPasswordResetTokenHash finds a user with a valid reset token hash
func (r *UserRepository) FindByPasswordResetTokenHash(tokenHash string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(
		context.Background(),
		bson.M{
			"password_reset_token_hash": tokenHash,
			"password_reset_expires_at": bson.M{"$gt": time.Now()},
		},
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
			"$unset": bson.M{
				"accreditation_rejection_reason": "",
			},
		},
	)

	return err
}

// RejectAccreditation updates a user's accreditation status and stores the rejection reason.
func (r *UserRepository) RejectAccreditation(id, reason string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"status":                         models.StatusRejected,
				"accreditation_rejection_reason": reason,
				"updated_at":                     time.Now(),
			},
		},
	)

	return err
}

// SetPasswordResetToken stores a password reset token hash and its expiry
func (r *UserRepository) SetPasswordResetToken(id, tokenHash string, expiresAt time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"password_reset_token_hash": tokenHash,
				"password_reset_expires_at": expiresAt,
				"updated_at":                time.Now(),
			},
		},
	)

	return err
}

// UpdatePassword updates a user's password and clears any outstanding reset token
func (r *UserRepository) UpdatePassword(id, passwordHash string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"password_hash": passwordHash,
				"updated_at":    time.Now(),
			},
			"$unset": bson.M{
				"password_reset_token_hash": "",
				"password_reset_expires_at": "",
			},
		},
	)

	return err
}

// ClearPasswordResetToken clears any stored password reset token for a user
func (r *UserRepository) ClearPasswordResetToken(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{
			"$unset": bson.M{
				"password_reset_token_hash": "",
				"password_reset_expires_at": "",
			},
			"$set": bson.M{
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

// FindManagedUsers returns student users whose accreditation has been processed.
func (r *UserRepository) FindManagedUsers() ([]*models.User, error) {
	cursor, err := r.collection.Find(
		context.Background(),
		bson.M{
			"role": models.RoleStudent,
			"status": bson.M{
				"$in": []models.UserStatus{models.StatusApproved, models.StatusRejected},
			},
		},
		options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}),
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

// FindByRole gets all users with a specific role
func (r *UserRepository) FindByRole(role models.UserRole) ([]*models.User, error) {
	cursor, err := r.collection.Find(
		context.Background(),
		bson.M{"role": role},
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

// DeleteByID removes a user document by ID.
func (r *UserRepository) DeleteByID(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objectID},
	)

	return err
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
