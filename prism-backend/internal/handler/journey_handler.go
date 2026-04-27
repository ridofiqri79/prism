package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type JourneyHandler struct {
	service *service.JourneyService
}

func NewJourneyHandler(service *service.JourneyService) *JourneyHandler {
	return &JourneyHandler{service: service}
}

func (h *JourneyHandler) GetJourney(c echo.Context) error {
	bbProjectID, err := parseIDParam(c, "bbProjectId")
	if err != nil {
		return err
	}
	res, err := h.service.GetProjectJourney(c.Request().Context(), bbProjectID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.JourneyResponse]{Data: res})
}
