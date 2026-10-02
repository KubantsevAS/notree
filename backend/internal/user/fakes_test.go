package user_test

import (
	"context"
	"database/sql"

	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/jackc/pgx/v5/pgtype"
)

type userStoreFake struct {
	getUserByIdResult           *userDb.UsersPublic
	getUserByIdErr              error
	getUserPasswordHashResult   string
	getUserPasswordHashErr      error
	createUserResult            pgtype.UUID
	createUserErr               error
	setVerificationTokenErr     error
	setVerificationTokenParams  []userDb.SetVerificationTokenParams
	updateUserPasswordErr       error
	updateUserPasswordParams    []userDb.UpdateUserPasswordParams
	updateUserProfileResult     *userDb.UpdateUserProfileRow
	updateUserProfileErr        error
	updateUserProfileParams     []userDb.UpdateUserProfileParams
	updateUserPreferencesResult *userDb.UpdateUserPreferencesRow
	updateUserPreferencesErr    error
	updateUserPreferencesParams []userDb.UpdateUserPreferencesParams
	verifyEmailByTokenResult    pgtype.UUID
	verifyEmailByTokenErr       error
	verifyEmailByTokenParams    []userDb.VerifyEmailByTokenParams
	verifyEmailAlreadyVerified  bool
}

func (r *userStoreFake) GetUserById(ctx context.Context, id pgtype.UUID) (userDb.UsersPublic, error) {
	if r.getUserByIdErr != nil {
		return userDb.UsersPublic{}, r.getUserByIdErr
	}
	if r.getUserByIdResult == nil {
		return userDb.UsersPublic{}, sql.ErrNoRows
	}
	return *r.getUserByIdResult, nil
}

func (r *userStoreFake) CreateUser(ctx context.Context, params userDb.CreateUserParams) (pgtype.UUID, error) {
	if r.createUserErr != nil {
		return pgtype.UUID{}, r.createUserErr
	}
	return r.createUserResult, nil
}

func (r *userStoreFake) GetUserPasswordHashById(ctx context.Context, id pgtype.UUID) (string, error) {
	if r.getUserPasswordHashErr != nil {
		return "", r.getUserPasswordHashErr
	}
	return r.getUserPasswordHashResult, nil
}

func (r *userStoreFake) SetVerificationToken(ctx context.Context, params userDb.SetVerificationTokenParams) error {
	r.setVerificationTokenParams = append(r.setVerificationTokenParams, params)
	return r.setVerificationTokenErr
}

func (r *userStoreFake) UpdateUserPassword(ctx context.Context, params userDb.UpdateUserPasswordParams) error {
	r.updateUserPasswordParams = append(r.updateUserPasswordParams, params)
	return r.updateUserPasswordErr
}

func (r *userStoreFake) UpdateUserPreferences(ctx context.Context, params userDb.UpdateUserPreferencesParams) (userDb.UpdateUserPreferencesRow, error) {
	r.updateUserPreferencesParams = append(r.updateUserPreferencesParams, params)
	if r.updateUserPreferencesErr != nil {
		return userDb.UpdateUserPreferencesRow{}, r.updateUserPreferencesErr
	}
	if r.updateUserPreferencesResult == nil {
		return userDb.UpdateUserPreferencesRow{}, sql.ErrNoRows
	}
	return *r.updateUserPreferencesResult, nil
}

func (r *userStoreFake) UpdateUserProfile(ctx context.Context, params userDb.UpdateUserProfileParams) (userDb.UpdateUserProfileRow, error) {
	r.updateUserProfileParams = append(r.updateUserProfileParams, params)
	if r.updateUserProfileErr != nil {
		return userDb.UpdateUserProfileRow{}, r.updateUserProfileErr
	}
	if r.updateUserProfileResult == nil {
		return userDb.UpdateUserProfileRow{}, sql.ErrNoRows
	}
	return *r.updateUserProfileResult, nil
}

func (r *userStoreFake) VerifyEmailByToken(ctx context.Context, params userDb.VerifyEmailByTokenParams) (pgtype.UUID, error) {
	r.verifyEmailByTokenParams = append(r.verifyEmailByTokenParams, params)
	if r.verifyEmailByTokenErr != nil {
		return pgtype.UUID{}, r.verifyEmailByTokenErr
	}
	return r.verifyEmailByTokenResult, nil
}

type fakeVerificationMailer struct {
	sent chan string
}

func (f *fakeVerificationMailer) SendPasswordReset(_ context.Context, _ string, _ string) error {
	return nil
}

func (f *fakeVerificationMailer) SendVerificationEmail(_ context.Context, _ string, token string) error {
	if f.sent != nil {
		f.sent <- token
	}
	return nil
}
