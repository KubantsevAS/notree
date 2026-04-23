package dto

type UpdateProfileRequest struct {
	Username  *string `json:"username"`
	AvatarUrl *string `json:"avatar_url"`
}
