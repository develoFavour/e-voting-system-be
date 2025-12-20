package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var database *mongo.Database

// Connect establishes connection to MongoDB
func Connect(uri, dbName string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	database = client.Database(dbName)
	log.Println("✅ Connected to MongoDB!")

	// Create indexes
	createIndexes()

	return database, nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}

// GetDatabase returns the database instance
func GetDatabase() *mongo.Database {
	return database
}

// createIndexes creates necessary database indexes
func createIndexes() {
	ctx := context.Background()

	// Users collection indexes
	usersCollection := database.Collection("users")
	usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"matric_number": 1},
		Options: options.Index().SetUnique(true),
	})

	// Candidates collection indexes
	candidatesCollection := database.Collection("candidates")
	candidatesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"position": 1},
	})

	// Votes collection indexes
	votesCollection := database.Collection("votes")
	votesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"user_id": 1},
		Options: options.Index().SetUnique(true),
	})

	log.Println("✅ Database indexes created")
}
