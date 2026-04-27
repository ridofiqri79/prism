package middleware

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID, module, action string) (bool, error)
}

type denyAllPermissionChecker struct{}

func (denyAllPermissionChecker) HasPermission(context.Context, string, string, string) (bool, error) {
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

			allowed, err := permissionChecker.HasPermission(c.Request().Context(), user.ID, module, action)
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
