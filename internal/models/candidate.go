package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Candidate struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Position  string             `bson:"position" json:"position"`
	Party     string             `bson:"party" json:"party"`
	Manifesto string             `bson:"manifesto" json:"manifesto"`
	ImageURL  string             `bson:"image_url" json:"imageUrl"`
	VoteCount int                `bson:"vote_count" json:"voteCount"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
}

type AddCandidateRequest struct {
	Name      string `json:"name" binding:"required"`
	Position  string `json:"position" binding:"required"`
	Party     string `json:"party" binding:"required"`
	Manifesto string `json:"manifesto" binding:"required"`
	ImageURL  string `json:"imageUrl"`
}
