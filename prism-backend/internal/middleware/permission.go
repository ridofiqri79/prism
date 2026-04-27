package middleware

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID model.AuthUser, module, action string) (bool, error)
}

type denyAllPermissionChecker struct{}

func (denyAllPermissionChecker) HasPermission(context.Context, model.AuthUser, string, string) (bool, error) {
	return false, nil
}

var permissionChecker PermissionChecker = denyAllPermissionChecker{}

func SetPermissionChecker(checker PermissionChecker) {
	if checker == nil {
		permissionChecker = denyAllPermissionChecker{}
		return
	}

	permissionChecker = checker
}

type databasePermissionChecker struct {
	queries *queries.Queries
}

func NewDatabasePermissionChecker(queries *queries.Queries) PermissionChecker {
	return &databasePermissionChecker{queries: queries}
}

func (d *databasePermissionChecker) HasPermission(ctx context.Context, user model.AuthUser, module, action string) (bool, error) {
	perm, err := d.queries.GetUserPermissionByModule(ctx, queries.GetUserPermissionByModuleParams{
		UserID: user.ID,
		Module: module,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, apperrors.Internal("Gagal mengecek permission")
	}

	switch action {
	case "create":
		return perm.CanCreate, nil
	case "read":
		return perm.CanRead, nil
	case "update":
		return perm.CanUpdate, nil
	case "delete":
		return perm.CanDelete, nil
	default:
		return false, fmt.Errorf("unsupported permission action %q", action)
	}
}

func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*model.AuthUser)
			if !ok || user == nil {
				return apperrors.Unauthorized("User tidak ditemukan")
			}

			if user.Role != "ADMIN" {
				return apperrors.Forbidden("Hanya ADMIN yang dapat mengakses resource ini")
			}

			return next(c)
		}
	}
}

func Require(module, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*model.AuthUser)
			if !ok || user == nil {
				return apperrors.Unauthorized("User tidak ditemukan")
			}

			if user.Role == "ADMIN" {
				return next(c)
			}

			if !isSupportedAction(action) {
				return fmt.Errorf("unsupported permission action %q", action)
			}

			allowed, err := permissionChecker.HasPermission(c.Request().Context(), *user, module, action)
			if err != nil {
				return err
			}

			if !allowed {
				return apperrors.Forbidden("Anda tidak memiliki akses ke resource ini")
			}

			return next(c)
		}
	}
}

func isSupportedAction(action string) bool {
	switch action {
	case "create", "read", "update", "delete":
		return true
	default:
		return false
	}
}
