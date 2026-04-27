package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type AuthService struct {
	queries   *queries.Queries
	jwtSecret string
	expiresIn int
}

func NewAuthService(queries *queries.Queries, jwtSecret string, expiresIn int) *AuthService {
	return &AuthService{
		queries:   queries,
		jwtSecret: jwtSecret,
		expiresIn: expiresIn,
	}
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, apperrors.Validation(
			apperrors.FieldError{Field: "username", Message: "wajib diisi"},
			apperrors.FieldError{Field: "password", Message: "wajib diisi"},
		)
	}

	user, err := s.queries.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperrors.Unauthorized("Username atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperrors.Unauthorized("Username atau password salah")
	}

	expiresAt := time.Now().Add(time.Duration(s.expiresIn) * time.Second)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      model.UUIDToString(user.ID),
		"username": user.Username,
		"role":     user.Role,
		"exp":      expiresAt.Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat token")
	}

	return &model.LoginResponse{
		AccessToken: tokenStr,
		ExpiresIn:   s.expiresIn,
		User:        toUserInfo(user),
	}, nil
}

func (s *AuthService) GetMe(ctx context.Context, userID pgtype.UUID) (*model.MeResponse, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.Unauthorized("User tidak ditemukan")
		}
		return nil, apperrors.Internal("Gagal mengambil data user")
	}

	if !user.IsActive {
		return nil, apperrors.Unauthorized("User tidak aktif")
	}

	perms, err := s.queries.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil permission user")
	}

	res := &model.MeResponse{
		UserInfo:    toUserInfo(user),
		Permissions: make([]model.PermissionInfo, 0, len(perms)),
	}
	for _, perm := range perms {
		res.Permissions = append(res.Permissions, toPermissionInfo(perm))
	}

	return res, nil
}

func toUserInfo(user queries.AppUser) model.UserInfo {
	return model.UserInfo{
		ID:       model.UUIDToString(user.ID),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}
}

func toPermissionInfo(perm queries.UserPermission) model.PermissionInfo {
	return model.PermissionInfo{
		Module:    perm.Module,
		CanCreate: perm.CanCreate,
		CanRead:   perm.CanRead,
		CanUpdate: perm.CanUpdate,
		CanDelete: perm.CanDelete,
	}
}
