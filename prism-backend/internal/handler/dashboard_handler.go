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
	filter, err := dashboardFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetSummary(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardSummary]{Data: res})
}

func (h *DashboardHandler) StageFunnel(c echo.Context) error {
	filter, err := dashboardFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetStageFunnel(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[[]model.StageMetric]{Data: res})
}

func (h *DashboardHandler) MonitoringRollup(c echo.Context) error {
	filter, err := dashboardFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetMonitoringRollup(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[[]model.TimeSeriesPoint]{Data: res})
}

func (h *DashboardHandler) FilterOptions(c echo.Context) error {
	res, err := h.service.GetFilterOptions(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[model.DashboardFilterOptions]{Data: res})
}

func (h *DashboardHandler) ExecutivePortfolio(c echo.Context) error {
	filter, err := dashboardFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetExecutivePortfolio(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.ExecutivePortfolioDashboard]{Data: res})
}

func (h *DashboardHandler) PipelineBottleneck(c echo.Context) error {
	filter, err := dashboardPipelineFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetPipelineBottleneck(c.Request().Context(), filter, paginationParams(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func dashboardFilter(c echo.Context) (model.DashboardFilterRequest, error) {
	filter := model.DashboardFilterRequest{}
	if value := strings.TrimSpace(c.QueryParam("period_id")); value != "" {
		filter.PeriodID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("publish_year")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "publish_year", Message: "harus angka"})
		}
		year := int32(parsed)
		filter.PublishYear = &year
	}
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
		if quarter != "TW1" && quarter != "TW2" && quarter != "TW3" && quarter != "TW4" {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "quarter", Message: "harus TW1, TW2, TW3, atau TW4"})
		}
		filter.Quarter = &quarter
	}
	if value := strings.TrimSpace(c.QueryParam("lender_id")); value != "" {
		filter.LenderID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("institution_id")); value != "" {
		filter.InstitutionID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("include_history")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "include_history", Message: "harus true atau false"})
		}
		filter.IncludeHistory = parsed
	}
	return filter, nil
}

func dashboardPipelineFilter(c echo.Context) (model.PipelineBottleneckFilterRequest, error) {
	filter := model.PipelineBottleneckFilterRequest{}
	if value := strings.TrimSpace(c.QueryParam("stage")); value != "" {
		stage := strings.ToUpper(value)
		filter.Stage = &stage
	}
	if value := strings.TrimSpace(c.QueryParam("period_id")); value != "" {
		filter.PeriodID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("publish_year")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "publish_year", Message: "harus angka"})
		}
		year := int32(parsed)
		filter.PublishYear = &year
	}
	if value := strings.TrimSpace(c.QueryParam("institution_id")); value != "" {
		filter.InstitutionID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("lender_id")); value != "" {
		filter.LenderID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("min_age_days")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed < 0 {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "min_age_days", Message: "harus angka positif"})
		}
		days := int32(parsed)
		filter.MinAgeDays = &days
	}
	return filter, nil
}
