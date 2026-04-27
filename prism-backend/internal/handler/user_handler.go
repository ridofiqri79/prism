package handler

import (
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) List(c echo.Context) error {
	params := model.PaginationParams{
		Page:  parseIntQuery(c, "page", 1),
		Limit: parseIntQuery(c, "limit", 20),
		Sort:  c.QueryParam("sort"),
		Order: c.QueryParam("order"),
	}

	res, err := h.service.ListUsers(c.Request().Context(), params)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *UserHandler) Get(c echo.Context) error {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	res, err := h.service.GetUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.UserResponse]{Data: res})
}

func (h *UserHandler) Create(c echo.Context) error {
	var req model.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "body", Message: "JSON tidak valid"})
	}

	res, err := h.service.CreateUser(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, model.DataResponse[*model.UserResponse]{Data: res})
}

func (h *UserHandler) Update(c echo.Context) error {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	var req model.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "body", Message: "JSON tidak valid"})
	}

	res, err := h.service.UpdateUser(c.Request().Context(), userID, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[*model.UserResponse]{Data: res})
}

func (h *UserHandler) Delete(c echo.Context) error {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	if err := h.service.DeleteUser(c.Request().Context(), userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) GetPermissions(c echo.Context) error {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	res, err := h.service.GetPermissions(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.DataResponse[[]model.PermissionInfo]{Data: res})
}

func (h *UserHandler) UpdatePermissions(c echo.Context) error {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	var req model.UpdatePermissionsRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.Validation(apperrors.FieldError{Field: "body", Message: "JSON tidak valid"})
	}

	if err := h.service.UpdatePermissions(c.Request().Context(), userID, req); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func parseIDParam(c echo.Context, name string) (pgtype.UUID, error) {
	id, err := model.ParseUUID(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, apperrors.Validation(apperrors.FieldError{Field: name, Message: "UUID tidak valid"})
	}
	return id, nil
}

func parseIntQuery(c echo.Context, name string, fallback int) int {
	value := c.QueryParam(name)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
