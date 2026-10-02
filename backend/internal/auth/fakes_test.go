package auth_test

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/auth"
	"github.com/KubantsevAS/notree/backend/internal/config"
	authDb "github.com/KubantsevAS/notree/backend/internal/db/auth"
	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

func newService(store *authStoreFake, userStore *userStoreFake, mailer *fakeMailer) *auth.Service {
	return auth.NewService(&config.Config{JWT: config.JWTConfig{Secret: testutil.TestSecret}}, store, userStore, mailer)
}

type authStoreFake struct {
	getRefreshTokenResult authDb.RefreshToken
	getRefreshTokenErr    error
	createRefreshTokenErr error
	createRefreshParams   []authDb.CreateRefreshTokenParams
	deleteRefreshTokenErr error
	deleteRefreshTokenArg []string
}

func (f *authStoreFake) CreateRefreshToken(ctx context.Context, params authDb.CreateRefreshTokenParams) error {
	f.createRefreshParams = append(f.createRefreshParams, params)
	return f.createRefreshTokenErr
}

func (f *authStoreFake) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	f.deleteRefreshTokenArg = append(f.deleteRefreshTokenArg, tokenHash)
	return f.deleteRefreshTokenErr
}

func (f *authStoreFake) GetRefreshToken(ctx context.Context, tokenHash string) (authDb.RefreshToken, error) {
	return f.getRefreshTokenResult, f.getRefreshTokenErr
}

type userStoreFake struct {
	getUserByEmailResult           userDb.User
	getUserByEmailErr              error
	createUserResult               pgtype.UUID
	createUserErr                  error
	createUserParams               []userDb.CreateUserParams
	setResetPasswordTokenParams    []userDb.SetResetPasswordTokenParams
	setResetPasswordTokenErr       error
	getUserIdByResetPasswordResult pgtype.UUID
	getUserIdByResetPasswordErr    error
	updateUserPasswordParams       []userDb.UpdateUserPasswordParams
	updateUserPasswordErr          error
}

func (f *userStoreFake) GetUserByEmail(_ context.Context, _ string) (userDb.User, error) {
	return f.getUserByEmailResult, f.getUserByEmailErr
}

func (f *userStoreFake) CreateUser(_ context.Context, params userDb.CreateUserParams) (pgtype.UUID, error) {
	f.createUserParams = append(f.createUserParams, params)
	return f.createUserResult, f.createUserErr
}

func (f *userStoreFake) SetResetPasswordToken(_ context.Context, params userDb.SetResetPasswordTokenParams) error {
	f.setResetPasswordTokenParams = append(f.setResetPasswordTokenParams, params)
	return f.setResetPasswordTokenErr
}

func (f *userStoreFake) GetUserIdByResetPasswordToken(_ context.Context, _ pgtype.Text) (pgtype.UUID, error) {
	return f.getUserIdByResetPasswordResult, f.getUserIdByResetPasswordErr
}

func (f *userStoreFake) UpdateUserPassword(_ context.Context, params userDb.UpdateUserPasswordParams) error {
	f.updateUserPasswordParams = append(f.updateUserPasswordParams, params)
	return f.updateUserPasswordErr
}

type fakeMailer struct {
	sent chan string
}

func (m *fakeMailer) SendPasswordReset(ctx context.Context, email string, token string) error {
	if m.sent != nil {
		m.sent <- token
	}
	return nil
}

func (m *fakeMailer) SendVerificationEmail(ctx context.Context, email string, token string) error {
	return nil
}
