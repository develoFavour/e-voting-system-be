package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ActivityType string

const (
	ActivityTypeElectionStarted ActivityType = "election_started"
	ActivityTypeElectionEnded   ActivityType = "election_ended"
	ActivityTypeVoterApproved   ActivityType = "voter_approved"
	ActivityTypeVoterRejected   ActivityType = "voter_rejected"
	ActivityTypeCandidateAdded  ActivityType = "candidate_added"
	ActivityTypePositionAdded   ActivityType = "position_added"
)

type Activity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      ActivityType       `bson:"type" json:"type"`
	Message   string             `bson:"message" json:"message"`
	AdminID   primitive.ObjectID `bson:"admin_id,omitempty" json:"admin_id"`
	AdminName string             `bson:"admin_name" json:"admin_name"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
