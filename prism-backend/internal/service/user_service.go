package service

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type UserService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
}

func NewUserService(db *pgxpool.Pool, queries *queries.Queries) *UserService {
	return &UserService{db: db, queries: queries}
}

func (s *UserService) ListUsers(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.UserResponse], error) {
	page, limit := normalizePagination(params.Page, params.Limit)
	offset := (page - 1) * limit

	users, err := s.queries.ListUsers(ctx, queries.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar user")
	}

	total, err := s.queries.CountUsers(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung user")
	}

	data := make([]model.UserResponse, 0, len(users))
	for _, user := range users {
		data = append(data, toUserResponse(user))
	}

	return &model.ListResponse[model.UserResponse]{
		Data: data,
		Meta: model.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID pgtype.UUID) (*model.UserResponse, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("User tidak ditemukan")
		}
		return nil, apperrors.Internal("Gagal mengambil data user")
	}

	res := toUserResponse(user)
	return &res, nil
}

func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
	if err := validateCreateUser(req); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat password hash")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)

	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return nil, apperrors.Internal("Gagal menyiapkan audit user")
	}

	user, err := s.queries.WithTx(tx).CreateUser(ctx, queries.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
	})
	if err != nil {
		return nil, apperrors.FromPgError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.Internal("Gagal menyimpan user")
	}

	res := toUserResponse(user)
	return &res, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID pgtype.UUID, req model.UpdateUserRequest) (*model.UserResponse, error) {
	if err := validateUpdateUser(req); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)

	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return nil, apperrors.Internal("Gagal menyiapkan audit user")
	}

	user, err := s.queries.WithTx(tx).UpdateUser(ctx, queries.UpdateUserParams{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
		IsActive: req.IsActive,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("User tidak ditemukan")
		}
		return nil, apperrors.FromPgError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.Internal("Gagal menyimpan user")
	}

	res := toUserResponse(user)
	return &res, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID pgtype.UUID) error {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return apperrors.NotFound("User tidak ditemukan")
		}
		return apperrors.Internal("Gagal mengambil data user")
	}

	_, err = s.UpdateUser(ctx, userID, model.UpdateUserRequest{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		IsActive: false,
	})
	return err
}

func (s *UserService) GetPermissions(ctx context.Context, userID pgtype.UUID) ([]model.PermissionInfo, error) {
	if _, err := s.queries.GetUserByID(ctx, userID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("User tidak ditemukan")
		}
		return nil, apperrors.Internal("Gagal mengambil data user")
	}

	perms, err := s.queries.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil permission user")
	}

	data := make([]model.PermissionInfo, 0, len(perms))
	for _, perm := range perms {
		data = append(data, toPermissionInfo(perm))
	}

	return data, nil
}

func (s *UserService) UpdatePermissions(ctx context.Context, userID pgtype.UUID, req model.UpdatePermissionsRequest) error {
	if req.Permissions == nil {
		return apperrors.Validation(apperrors.FieldError{Field: "permissions", Message: "wajib diisi"})
	}

	if _, err := s.queries.GetUserByID(ctx, userID); err != nil {
		if err == pgx.ErrNoRows {
			return apperrors.NotFound("User tidak ditemukan")
		}
		return apperrors.Internal("Gagal mengambil data user")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.Internal("Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)

	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return apperrors.Internal("Gagal menyiapkan audit user")
	}

	qtx := s.queries.WithTx(tx)
	if err := qtx.DeleteUserPermissions(ctx, userID); err != nil {
		return apperrors.Internal("Gagal menghapus permission lama")
	}

	for _, perm := range req.Permissions {
		if strings.TrimSpace(perm.Module) == "" {
			return apperrors.Validation(apperrors.FieldError{Field: "permissions.module", Message: "wajib diisi"})
		}

		if _, err := qtx.CreateUserPermission(ctx, queries.CreateUserPermissionParams{
			UserID:    userID,
			Module:    perm.Module,
			CanCreate: perm.CanCreate,
			CanRead:   perm.CanRead,
			CanUpdate: perm.CanUpdate,
			CanDelete: perm.CanDelete,
		}); err != nil {
			return apperrors.FromPgError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.Internal("Gagal menyimpan permission")
	}

	return nil
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func validateCreateUser(req model.CreateUserRequest) error {
	var fields []apperrors.FieldError
	if len(strings.TrimSpace(req.Username)) < 3 {
		fields = append(fields, apperrors.FieldError{Field: "username", Message: "minimal 3 karakter"})
	}
	if !strings.Contains(req.Email, "@") {
		fields = append(fields, apperrors.FieldError{Field: "email", Message: "format email tidak valid"})
	}
	if len(req.Password) < 8 {
		fields = append(fields, apperrors.FieldError{Field: "password", Message: "minimal 8 karakter"})
	}
	if !isValidRole(req.Role) {
		fields = append(fields, apperrors.FieldError{Field: "role", Message: "harus ADMIN atau STAFF"})
	}
	if len(fields) > 0 {
		return apperrors.Validation(fields...)
	}
	return nil
}

func validateUpdateUser(req model.UpdateUserRequest) error {
	var fields []apperrors.FieldError
	if len(strings.TrimSpace(req.Username)) < 3 {
		fields = append(fields, apperrors.FieldError{Field: "username", Message: "minimal 3 karakter"})
	}
	if !strings.Contains(req.Email, "@") {
		fields = append(fields, apperrors.FieldError{Field: "email", Message: "format email tidak valid"})
	}
	if !isValidRole(req.Role) {
		fields = append(fields, apperrors.FieldError{Field: "role", Message: "harus ADMIN atau STAFF"})
	}
	if len(fields) > 0 {
		return apperrors.Validation(fields...)
	}
	return nil
}

func isValidRole(role string) bool {
	return role == "ADMIN" || role == "STAFF"
}

func toUserResponse(user queries.AppUser) model.UserResponse {
	return model.UserResponse{
		ID:        model.UUIDToString(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		IsActive:  user.IsActive,
		CreatedAt: formatTime(user.CreatedAt),
		UpdatedAt: formatTime(user.UpdatedAt),
	}
}

func formatTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
