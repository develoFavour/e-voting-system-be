package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vote struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID            primitive.ObjectID `bson:"user_id" json:"userId"`
	ElectionID        primitive.ObjectID `bson:"election_id" json:"electionId"`
	EncryptedVoteData string             `bson:"encrypted_vote_data" json:"-"`
	Timestamp         time.Time          `bson:"timestamp" json:"timestamp"`
	Hash              string             `bson:"hash" json:"hash"`
}

// VoteData represents the decrypted vote structure
type VoteData struct {
	Selections map[string]string `json:"selections"`
}

// CastVoteRequest represents the vote submission payload
type CastVoteRequest struct {
	Selections map[string]string `json:"selections" binding:"required"`
}
