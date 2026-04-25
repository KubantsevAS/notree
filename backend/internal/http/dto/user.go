package dto

import (
	"encoding/json"
	"time"
)

type GetProfileResponse struct {
	ID              string           `json:"id"`
	Email           string           `json:"email"`
	Username        *string          `json:"username"`
	AvatarUrl       *string          `json:"avatar_url"`
	Timezone        *string          `json:"timezone"`
	Locale          *string          `json:"locale"`
	Preferences     *json.RawMessage `json:"preferences" swaggertype:"object"`
	IsEmailVerified *bool            `json:"is_email_verified"`
	LastLoginAt     *time.Time       `json:"last_login_at"`
	CreatedAt       *time.Time       `json:"created_at"`
	UpdatedAt       *time.Time       `json:"updated_at"`
}

type UpdateUserProfileRequest struct {
	Username  *string `json:"username" validate:"omitempty,max=100"`
	AvatarUrl *string `json:"avatar_url" validate:"omitempty,max=500"`
}

type UpdateUserProfileResponse struct {
	Username  *string    `json:"username"`
	AvatarUrl *string    `json:"avatar_url"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UpdateUserPreferencesRequest struct {
	Locale      *string          `json:"locale" validate:"omitempty,max=10"`
	Timezone    *string          `json:"timezone" validate:"omitempty,max=50"`
	Preferences *json.RawMessage `json:"preferences" validate:"omitempty,json_object" swaggertype:"object"`
}

type UpdateUserPreferencesResponse struct {
	Locale      *string          `json:"locale"`
	Timezone    *string          `json:"timezone"`
	Preferences *json.RawMessage `json:"preferences" swaggertype:"object"`
	UpdatedAt   *time.Time       `json:"updated_at"`
}
