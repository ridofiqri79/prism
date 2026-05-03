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

func (h *DashboardHandler) GreenBookReadiness(c echo.Context) error {
	filter, err := dashboardGreenBookReadinessFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetGreenBookReadiness(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.GreenBookReadinessDashboard]{Data: res})
}

func (h *DashboardHandler) LenderFinancingMix(c echo.Context) error {
	filter, err := dashboardLenderFinancingMixFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetLenderFinancingMix(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.LenderFinancingMixDashboard]{Data: res})
}

func (h *DashboardHandler) KLPortfolioPerformance(c echo.Context) error {
	filter, err := dashboardKLPortfolioPerformanceFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetKLPortfolioPerformance(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.KLPortfolioPerformanceDashboard]{Data: res})
}

func (h *DashboardHandler) LADisbursement(c echo.Context) error {
	filter, err := dashboardLADisbursementFilter(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetLADisbursement(c.Request().Context(), filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.LADisbursementDashboard]{Data: res})
}

func (h *DashboardHandler) DataQualityGovernance(c echo.Context) error {
	filter, err := dashboardDataQualityGovernanceFilter(c)
	if err != nil {
		return err
	}
	user, _ := c.Get("user").(*model.AuthUser)
	includeAudit := user != nil && user.Role == "ADMIN"
	res, err := h.service.GetDataQualityGovernance(c.Request().Context(), filter, includeAudit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DataQualityGovernanceDashboard]{Data: res})
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

func dashboardGreenBookReadinessFilter(c echo.Context) (model.GreenBookReadinessFilterRequest, error) {
	filter := model.GreenBookReadinessFilterRequest{}
	if value := strings.TrimSpace(c.QueryParam("publish_year")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "publish_year", Message: "harus angka"})
		}
		year := int32(parsed)
		filter.PublishYear = &year
	}
	if value := strings.TrimSpace(c.QueryParam("green_book_id")); value != "" {
		filter.GreenBookID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("institution_id")); value != "" {
		filter.InstitutionID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("lender_id")); value != "" {
		filter.LenderID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("readiness_status")); value != "" {
		status := strings.ToUpper(value)
		switch status {
		case "READY", "PARTIAL", "INCOMPLETE", "COFINANCING":
			filter.ReadinessStatus = &status
		default:
			return filter, apperrors.Validation(apperrors.FieldError{Field: "readiness_status", Message: "harus READY, PARTIAL, INCOMPLETE, atau COFINANCING"})
		}
	}
	return filter, nil
}

func dashboardLenderFinancingMixFilter(c echo.Context) (model.LenderFinancingMixFilterRequest, error) {
	filter := model.LenderFinancingMixFilterRequest{}
	if value := strings.TrimSpace(c.QueryParam("lender_type")); value != "" {
		switch strings.ToLower(value) {
		case "bilateral":
			lenderType := "Bilateral"
			filter.LenderType = &lenderType
		case "multilateral":
			lenderType := "Multilateral"
			filter.LenderType = &lenderType
		case "ksa":
			lenderType := "KSA"
			filter.LenderType = &lenderType
		default:
			return filter, apperrors.Validation(apperrors.FieldError{Field: "lender_type", Message: "harus Bilateral, Multilateral, atau KSA"})
		}
	}
	if value := strings.TrimSpace(c.QueryParam("lender_id")); value != "" {
		filter.LenderID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("currency")); value != "" {
		currency := strings.ToUpper(value)
		filter.Currency = &currency
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
	if value := strings.TrimSpace(c.QueryParam("budget_year")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "budget_year", Message: "harus angka"})
		}
		year := int32(parsed)
		filter.BudgetYear = &year
	}
	return filter, nil
}

func dashboardKLPortfolioPerformanceFilter(c echo.Context) (model.KLPortfolioPerformanceFilterRequest, error) {
	filter := model.KLPortfolioPerformanceFilterRequest{}
	if value := strings.TrimSpace(c.QueryParam("institution_id")); value != "" {
		filter.InstitutionID = &value
	}
	if value := strings.TrimSpace(c.QueryParam("institution_role")); value != "" {
		switch strings.ToLower(value) {
		case "executing agency":
			role := "Executing Agency"
			filter.InstitutionRole = &role
		case "implementing agency":
			role := "Implementing Agency"
			filter.InstitutionRole = &role
		default:
			return filter, apperrors.Validation(apperrors.FieldError{Field: "institution_role", Message: "harus Executing Agency atau Implementing Agency"})
		}
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
	if value := strings.TrimSpace(c.QueryParam("sort_by")); value != "" {
		switch value {
		case "pipeline_usd", "la_commitment_usd", "absorption_pct", "risk_count":
			filter.SortBy = &value
		default:
			return filter, apperrors.Validation(apperrors.FieldError{Field: "sort_by", Message: "sort_by tidak valid"})
		}
	}
	return filter, nil
}

func dashboardLADisbursementFilter(c echo.Context) (model.LADisbursementFilterRequest, error) {
	filter := model.LADisbursementFilterRequest{}
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
	if value := strings.TrimSpace(c.QueryParam("is_extended")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "is_extended", Message: "harus true atau false"})
		}
		filter.IsExtended = &parsed
	}
	if value := strings.TrimSpace(c.QueryParam("closing_months")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "closing_months", Message: "harus angka"})
		}
		months := int32(parsed)
		if months != 3 && months != 6 && months != 12 {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "closing_months", Message: "harus 3, 6, atau 12"})
		}
		filter.ClosingMonths = &months
	}
	if value := strings.TrimSpace(c.QueryParam("risk_level")); value != "" {
		riskLevel := strings.ToLower(value)
		if riskLevel != "low" && riskLevel != "medium" && riskLevel != "high" {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "risk_level", Message: "harus low, medium, atau high"})
		}
		filter.RiskLevel = &riskLevel
	}
	return filter, nil
}

func dashboardDataQualityGovernanceFilter(c echo.Context) (model.DataQualityGovernanceFilterRequest, error) {
	filter := model.DataQualityGovernanceFilterRequest{AuditDays: 30}
	if value := strings.TrimSpace(c.QueryParam("severity")); value != "" {
		severity := strings.ToLower(value)
		if severity != "info" && severity != "warning" && severity != "error" {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "severity", Message: "harus info, warning, atau error"})
		}
		filter.Severity = &severity
	}
	if value := strings.TrimSpace(c.QueryParam("module")); value != "" {
		filter.Module = &value
	}
	if value := strings.TrimSpace(c.QueryParam("issue_type")); value != "" {
		issueType := strings.ToUpper(value)
		filter.IssueType = &issueType
	}
	if value := strings.TrimSpace(c.QueryParam("only_unresolved")); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "only_unresolved", Message: "harus true atau false"})
		}
		filter.OnlyUnresolved = parsed
	}
	if value := strings.TrimSpace(c.QueryParam("audit_days")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil || parsed <= 0 {
			return filter, apperrors.Validation(apperrors.FieldError{Field: "audit_days", Message: "harus angka positif"})
		}
		filter.AuditDays = int32(parsed)
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
