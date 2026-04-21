package service

import "github.com/KubantsevAS/notree/backend/internal/db/user"

type UserService struct {
	db *user.Queries
}

func NewUserService(db *user.Queries) *UserService {
	return &UserService{db: db}
}
