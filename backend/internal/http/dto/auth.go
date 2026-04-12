package dto

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type RegisterResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}
