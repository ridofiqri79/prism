package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "body", Message: "JSON tidak valid"})
	}

	res, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.LoginResponse]{Data: res})
}

func (h *AuthHandler) Logout(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) Me(c echo.Context) error {
	user, ok := c.Get("user").(*model.AuthUser)
	if !ok || user == nil {
		return apperrors.Unauthorized("User tidak ditemukan")
	}

	res, err := h.service.GetMe(c.Request().Context(), user.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.MeResponse]{Data: res})
}
