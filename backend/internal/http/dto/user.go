package dto

import "encoding/json"

type UpdateProfileRequest struct {
	Username  *string `json:"username" validate:"omitempty,max=100"`
	AvatarUrl *string `json:"avatar_url" validate:"omitempty,max=500"`
}

type UpdateUserPreferencesRequest struct {
	Locale      *string          `json:"locale" validate:"omitempty,max=10"`
	Timezone    *string          `json:"timezone" validate:"omitempty,max=50"`
	Preferences *json.RawMessage `json:"preferences" validate:"omitempty,json_object"`
}
