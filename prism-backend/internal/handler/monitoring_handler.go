package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type MonitoringHandler struct {
	service *service.MonitoringService
}

func NewMonitoringHandler(service *service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{service: service}
}

func (h *MonitoringHandler) List(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	res, err := h.service.ListMonitoring(c.Request().Context(), laID, paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *MonitoringHandler) Get(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	res, err := h.service.GetMonitoring(c.Request().Context(), laID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.MonitoringResponse]{Data: res})
}

func (h *MonitoringHandler) Create(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	var req model.MonitoringRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.CreateMonitoring(c.Request().Context(), laID, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, model.DataResponse[*model.MonitoringResponse]{Data: res})
}

func (h *MonitoringHandler) Update(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	var req model.MonitoringRequest
	if err := bind(c, &req); err != nil {
		return err
	}
	res, err := h.service.UpdateMonitoring(c.Request().Context(), laID, id, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.MonitoringResponse]{Data: res})
}

func (h *MonitoringHandler) Delete(c echo.Context) error {
	laID, err := parseIDParam(c, "laId")
	if err != nil {
		return err
	}
	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteMonitoring(c.Request().Context(), laID, id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
