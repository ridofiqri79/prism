package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
)

func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *apperrors.AppError
	var httpErr *echo.HTTPError

	switch {
	case errors.As(err, &appErr):
		_ = c.JSON(appErr.StatusCode, map[string]any{"error": appErr})
	case errors.As(err, &httpErr):
		_ = c.JSON(httpErr.Code, map[string]any{
			"error": map[string]any{
				"code":    "HTTP_ERROR",
				"message": fmt.Sprintf("%v", httpErr.Message),
			},
		})
	default:
		log.Error().Err(err).Msg("unhandled error")
		_ = c.JSON(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{
				"code":    "INTERNAL_ERROR",
				"message": "Terjadi kesalahan, silakan coba lagi",
			},
		})
	}
}
