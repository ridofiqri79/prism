package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/service"
)

type DashboardAnalyticsHandler struct {
	service *service.DashboardAnalyticsService
}

func NewDashboardAnalyticsHandler(service *service.DashboardAnalyticsService) *DashboardAnalyticsHandler {
	return &DashboardAnalyticsHandler{service: service}
}

func (h *DashboardAnalyticsHandler) Overview(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.Overview(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsOverviewResponse]{Data: res})
}

func (h *DashboardAnalyticsHandler) Institutions(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.Institutions(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsInstitutionsResponse]{Data: res})
}

func (h *DashboardAnalyticsHandler) Lenders(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.Lenders(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsLendersResponse]{Data: res})
}

func (h *DashboardAnalyticsHandler) Absorption(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.Absorption(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsAbsorptionResponse]{Data: res})
}

func (h *DashboardAnalyticsHandler) Yearly(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.Yearly(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsYearlyResponse]{Data: res})
}

func (h *DashboardAnalyticsHandler) LenderProportion(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.LenderProportion(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsLenderProportionResponse]{Data: res})
}

func (h *DashboardAnalyticsHandler) Risks(c echo.Context) error {
	filter, err := dashboardAnalyticsFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.Risks(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsRisksResponse]{Data: res})
}

func dashboardAnalyticsFilter(c echo.Context) (model.DashboardAnalyticsFilter, error) {
	filter := model.DashboardAnalyticsFilter{
		LenderIDs:        queryValues(c, "lender_ids", "lender_ids[]"),
		LenderTypes:      queryValues(c, "lender_types", "lender_types[]"),
		InstitutionIDs:   queryValues(c, "institution_ids", "institution_ids[]"),
		PipelineStatuses: queryValues(c, "pipeline_statuses", "pipeline_statuses[]"),
		ProjectStatuses:  queryValues(c, "project_statuses", "project_statuses[]"),
		RegionIDs:        queryValues(c, "region_ids", "region_ids[]"),
		ProgramTitleIDs:  queryValues(c, "program_title_ids", "program_title_ids[]"),
	}

	if value := strings.TrimSpace(c.QueryParam("budget_year")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, dashboardAnalyticsValidation("budget_year", "harus angka")
		}
		year := int32(parsed)
		filter.BudgetYear = &year
	}
	if value := strings.TrimSpace(c.QueryParam("quarter")); value != "" {
		quarter := strings.ToUpper(value)
		if err := dashboardAnalyticsAllowed("quarter", quarter, "TW1", "TW2", "TW3", "TW4"); err != nil {
			return filter, err
		}
		filter.Quarter = &quarter
	}
	if value := strings.TrimSpace(c.QueryParam("foreign_loan_min")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return filter, dashboardAnalyticsValidation("foreign_loan_min", "harus angka")
		}
		filter.ForeignLoanMin = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("foreign_loan_max")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return filter, dashboardAnalyticsValidation("foreign_loan_max", "harus angka")
		}
		filter.ForeignLoanMax = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("low_absorption_threshold")); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed < 0 || parsed > 100 {
			return filter, dashboardAnalyticsValidation("low_absorption_threshold", "harus angka 0 sampai 100")
		}
		filter.LowAbsorptionThreshold = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("closing_months_threshold")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			return filter, dashboardAnalyticsValidation("closing_months_threshold", "harus angka lebih dari 0")
		}
		threshold := int32(parsed)
		filter.ClosingMonthsThreshold = &threshold
	}
	if value := strings.TrimSpace(c.QueryParam("stale_monitoring_quarters")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			return filter, dashboardAnalyticsValidation("stale_monitoring_quarters", "harus angka lebih dari 0")
		}
		threshold := int32(parsed)
		filter.StaleMonitoringQuarters = &threshold
	}
	if filter.ForeignLoanMin != nil && filter.ForeignLoanMax != nil && *filter.ForeignLoanMin > *filter.ForeignLoanMax {
		return filter, dashboardAnalyticsValidation("foreign_loan_max", "harus lebih besar dari nilai minimum")
	}
	if value := strings.TrimSpace(c.QueryParam("include_history")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return filter, dashboardAnalyticsValidation("include_history", "harus true atau false")
		}
		filter.IncludeHistory = parsed
	}

	if err := dashboardAnalyticsUUIDs("lender_ids", filter.LenderIDs); err != nil {
		return filter, err
	}
	if err := dashboardAnalyticsUUIDs("institution_ids", filter.InstitutionIDs); err != nil {
		return filter, err
	}
	if err := dashboardAnalyticsUUIDs("region_ids", filter.RegionIDs); err != nil {
		return filter, err
	}
	if err := dashboardAnalyticsUUIDs("program_title_ids", filter.ProgramTitleIDs); err != nil {
		return filter, err
	}
	if err := dashboardAnalyticsAllowedValues("lender_types", filter.LenderTypes, "Bilateral", "Multilateral", "KSA"); err != nil {
		return filter, err
	}
	if err := dashboardAnalyticsAllowedValues("pipeline_statuses", filter.PipelineStatuses, "BB", "GB", "DK", "LA", "Monitoring"); err != nil {
		return filter, err
	}
	if err := dashboardAnalyticsAllowedValues("project_statuses", filter.ProjectStatuses, "Pipeline", "Ongoing"); err != nil {
		return filter, err
	}

	return filter, nil
}

func dashboardAnalyticsUUIDs(field string, values []string) error {
	for _, value := range values {
		if _, err := model.ParseUUID(value); err != nil {
			return dashboardAnalyticsValidation(field, "UUID tidak valid")
		}
	}
	return nil
}

func dashboardAnalyticsAllowedValues(field string, values []string, allowed ...string) error {
	for _, value := range values {
		if err := dashboardAnalyticsAllowed(field, value, allowed...); err != nil {
			return err
		}
	}
	return nil
}

func dashboardAnalyticsAllowed(field, value string, allowed ...string) error {
	for _, item := range allowed {
		if value == item {
			return nil
		}
	}
	return dashboardAnalyticsValidation(field, "nilai tidak valid")
}

func dashboardAnalyticsValidation(field, message string) error {
	return apperrors.Validation(apperrors.FieldError{Field: field, Message: message})
}
