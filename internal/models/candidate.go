package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Candidate struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ElectionID primitive.ObjectID `bson:"election_id,omitempty" json:"electionId"`
	PositionID primitive.ObjectID `bson:"position_id,omitempty" json:"positionId"`
	Name       string             `bson:"name" json:"name"`
	Position   string             `bson:"position" json:"position"`
	Party      string             `bson:"party" json:"party"`
	Manifesto  string             `bson:"manifesto" json:"manifesto"`
	Department string             `bson:"department" json:"department"`
	Level      string             `bson:"level" json:"level"`
	ImageURL   string             `bson:"image_url" json:"imageUrl"`
	VoteCount  int                `bson:"vote_count" json:"voteCount"`
	CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
}

type AddCandidateRequest struct {
	Name       string `json:"name" binding:"required"`
	Position   string `json:"position" binding:"required"`
	Party      string `json:"party" binding:"required"`
	Manifesto  string `json:"manifesto" binding:"required"`
	Department string `json:"department" binding:"required"`
	Level      string `json:"level" binding:"required"`
	ImageURL   string `json:"imageUrl"`
}
