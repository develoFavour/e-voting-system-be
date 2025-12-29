package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Position struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ElectionID    primitive.ObjectID `bson:"election_id,omitempty" json:"electionId"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description,omitempty" json:"description,omitempty"`
	Order         int                `bson:"order" json:"order"`
	MaxSelections int                `bson:"max_selections" json:"maxSelections"`
	CreatedAt     time.Time          `bson:"created_at" json:"createdAt"`
}

type AddPositionRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	Order         int    `json:"order"`
	MaxSelections int    `json:"maxSelections"`
}
