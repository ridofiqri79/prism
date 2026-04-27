package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type DashboardHandler struct {
	service *service.DashboardService
}

func NewDashboardHandler(service *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

func (h *DashboardHandler) Summary(c echo.Context) error {
	res, err := h.service.GetSummary(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardSummary]{Data: res})
}

func (h *DashboardHandler) MonitoringSummary(c echo.Context) error {
	filter, err := monitoringSummaryFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetMonitoringSummary(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.MonitoringSummary]{Data: res})
}

func monitoringSummaryFilter(c echo.Context) (model.MonitoringSummaryFilter, error) {
	filter := model.MonitoringSummaryFilter{}
	if value := strings.TrimSpace(c.QueryParam("budget_year")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "budget_year", Message: "harus angka"})
		}
		year := int32(parsed)
		filter.BudgetYear = &year
	}
	if value := strings.TrimSpace(c.QueryParam("quarter")); value != "" {
		quarter := strings.ToUpper(value)
		filter.Quarter = &quarter
	}
	if value := strings.TrimSpace(c.QueryParam("lender_id")); value != "" {
		filter.LenderID = &value
	}
	return filter, nil
}
