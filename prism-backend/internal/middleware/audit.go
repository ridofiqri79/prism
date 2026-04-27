package middleware

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/model"
)

type auditUserIDContextKey struct{}

func SetAuditUser(_ *pgxpool.Pool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, ok := c.Get("user").(*model.AuthUser)
			if ok && user != nil && user.ID.Valid {
				ctx := context.WithValue(c.Request().Context(), auditUserIDContextKey{}, model.UUIDToString(user.ID))
				c.SetRequest(c.Request().WithContext(ctx))
			}

			return next(c)
		}
	}
}

func ApplyAuditUser(ctx context.Context, tx pgx.Tx) error {
	userID, ok := AuditUserID(ctx)
	if !ok || userID == "" {
		return nil
	}

	_, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID)
	return err
}

func AuditUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(auditUserIDContextKey{}).(string)
	return userID, ok
}
