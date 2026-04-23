package dto

import "encoding/json"

type UpdateProfileRequest struct {
	Username  *string `json:"username"`
	AvatarUrl *string `json:"avatar_url"`
}

type UpdateUserPreferencesRequest struct {
	Locale      *string          `json:"locale"`
	Timezone    *string          `json:"timezone"`
	Preferences *json.RawMessage `json:"preferences"`
}
