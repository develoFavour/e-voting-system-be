package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStatus string
type UserRole string

const (
	StatusPending  UserStatus = "PENDING"
	StatusApproved UserStatus = "APPROVED"
	StatusRejected UserStatus = "REJECTED"

	RoleStudent UserRole = "STUDENT"
	RoleAdmin   UserRole = "ADMIN"
)

type User struct {
	ID                           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MatricNumber                 string             `bson:"matric_number" json:"matricNumber"`
	FullName                     string             `bson:"full_name" json:"fullName"`
	Email                        string             `bson:"email,omitempty" json:"email,omitempty"`
	Department                   string             `bson:"department" json:"department"`
	Faculty                      string             `bson:"faculty" json:"faculty"`
	PasswordHash                 string             `bson:"password_hash" json:"-"`
	PasswordResetTokenHash       string             `bson:"password_reset_token_hash,omitempty" json:"-"`
	PasswordResetExpiresAt       *time.Time         `bson:"password_reset_expires_at,omitempty" json:"-"`
	AccreditationRejectionReason string             `bson:"accreditation_rejection_reason,omitempty" json:"accreditationRejectionReason,omitempty"`
	IDCardURL                    string             `bson:"id_card_url" json:"idCardUrl"`
	Status                       UserStatus         `bson:"status" json:"status"`
	Role                         UserRole           `bson:"role" json:"role"`
	HasVoted                     bool               `bson:"has_voted" json:"hasVoted"`
	CreatedAt                    time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt                    time.Time          `bson:"updated_at" json:"updatedAt"`
}

// RegisterRequest represents the accreditation registration payload
type RegisterRequest struct {
	MatricNumber string `json:"matricNumber" binding:"required"`
	FullName     string `json:"fullName" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Department   string `json:"department" binding:"required"`
	Faculty      string `json:"faculty" binding:"required"`
	Password     string `json:"password" binding:"required,min=6"`
	IDCardURL    string `json:"idCardUrl" binding:"required"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	MatricNumber string `json:"matricNumber" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	MatricNumber string `json:"matricNumber" binding:"required"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type RejectAccreditationRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// LoginResponse contains JWT token and user info
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
