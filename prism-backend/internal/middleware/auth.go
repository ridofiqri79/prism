package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type JWTClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func Auth(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get(echo.HeaderAuthorization)
			if header == "" {
				return apperrors.Unauthorized("Token tidak ditemukan")
			}

			tokenString, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(tokenString) == "" {
				return apperrors.Unauthorized("Format token tidak valid")
			}

			claims := new(JWTClaims)
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				return apperrors.Unauthorized("Token tidak valid atau kedaluwarsa")
			}

			if claims.Subject == "" || claims.Role == "" {
				return apperrors.Unauthorized("Token tidak valid atau kedaluwarsa")
			}

			c.Set("user", &model.AuthUser{
				ID:       claims.Subject,
				Username: claims.Username,
				Role:     claims.Role,
			})

			return next(c)
		}
	}
}
