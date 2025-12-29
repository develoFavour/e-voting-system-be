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

type PositionRepository struct {
	collection *mongo.Collection
}

func NewPositionRepository(db *mongo.Database) *PositionRepository {
	return &PositionRepository{collection: db.Collection("positions")}
}

func (r *PositionRepository) Create(position *models.Position) error {
	position.CreatedAt = time.Now()
	if position.MaxSelections <= 0 {
		position.MaxSelections = 1
	}

	res, err := r.collection.InsertOne(context.Background(), position)
	if err != nil {
		return err
	}

	position.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *PositionRepository) FindByElectionID(electionID string) ([]*models.Position, error) {
	oid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return nil, err
	}

	cur, err := r.collection.Find(
		context.Background(),
		bson.M{"election_id": oid},
		options.Find().SetSort(bson.D{{"order", 1}, {"created_at", 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var positions []*models.Position
	if err := cur.All(context.Background(), &positions); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *PositionRepository) FindByID(id string) (*models.Position, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var pos models.Position
	err = r.collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&pos)
	if err != nil {
		return nil, err
	}

	return &pos, nil
}

func (r *PositionRepository) FindUnscoped() ([]*models.Position, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"election_id": bson.M{"$exists": false}},
			{"election_id": primitive.NilObjectID},
		},
	}

	cur, err := r.collection.Find(
		context.Background(),
		filter,
		options.Find().SetSort(bson.D{{"order", 1}, {"created_at", 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var positions []*models.Position
	if err := cur.All(context.Background(), &positions); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *PositionRepository) FindByElectionIDAndName(electionID string, name string) (*models.Position, error) {
	eid, err := primitive.ObjectIDFromHex(electionID)
	if err != nil {
		return nil, err
	}

	var pos models.Position
	err = r.collection.FindOne(context.Background(), bson.M{"election_id": eid, "name": name}).Decode(&pos)
	if err != nil {
		return nil, err
	}

	return &pos, nil
}

func (r *PositionRepository) FindUnscopedByName(name string) (*models.Position, error) {
	filter := bson.M{
		"name": name,
		"$or": []bson.M{
			{"election_id": bson.M{"$exists": false}},
			{"election_id": primitive.NilObjectID},
		},
	}

	var pos models.Position
	err := r.collection.FindOne(context.Background(), filter, options.FindOne().SetSort(bson.D{{"created_at", -1}})).Decode(&pos)
	if err != nil {
		return nil, err
	}

	return &pos, nil
}

func (r *PositionRepository) AttachUnscopedToElection(electionID string) error {
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
