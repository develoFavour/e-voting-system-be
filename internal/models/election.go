package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ElectionStatus string

const (
	ElectionStatusPending ElectionStatus = "pending"
	ElectionStatusLive    ElectionStatus = "live"
	ElectionStatusClosed  ElectionStatus = "closed"
)

type Election struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Status      ElectionStatus     `bson:"status" json:"status"`
	StartTime   *time.Time         `bson:"start_time,omitempty" json:"start_time"`
	EndTime     *time.Time         `bson:"end_time,omitempty" json:"end_time"`
	Duration    int                `bson:"duration" json:"duration"` // Duration in minutes
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedBy   primitive.ObjectID `bson:"created_by" json:"created_by"`
}

type StartElectionRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Duration    int    `json:"duration" binding:"required"` // Duration in minutes
}
