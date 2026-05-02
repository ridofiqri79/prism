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

type DashboardAnalyticsHandler struct {
	service *service.DashboardAnalyticsService
}

func NewDashboardAnalyticsHandler(svc *service.DashboardAnalyticsService) *DashboardAnalyticsHandler {
	return &DashboardAnalyticsHandler{service: svc}
}

// ------ Overview ------

func (h *DashboardAnalyticsHandler) Overview(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	res, err := h.service.GetOverview(c.Request().Context(), f)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsOverview]{Data: res})
}

// ------ Institutions ------

func (h *DashboardAnalyticsHandler) Institutions(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	page, limit, err := parsePagination(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetInstitutions(c.Request().Context(), f, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsInstitutionsResponse]{Data: res})
}

// ------ Lenders ------

func (h *DashboardAnalyticsHandler) Lenders(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	page, limit, err := parsePagination(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetLenders(c.Request().Context(), f, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsLendersResponse]{Data: res})
}

// ------ Absorption ------

func (h *DashboardAnalyticsHandler) Absorption(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	page, limit, err := parsePagination(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetAbsorption(c.Request().Context(), f, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsAbsorptionResponse]{Data: res})
}

// ------ Yearly ------

func (h *DashboardAnalyticsHandler) Yearly(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	page, limit, err := parsePagination(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetYearly(c.Request().Context(), f, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsYearlyResponse]{Data: res})
}

// ------ Lender Proportion ------

func (h *DashboardAnalyticsHandler) LenderProportion(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	res, err := h.service.GetLenderProportion(c.Request().Context(), f)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardAnalyticsLenderProportionResponse]{Data: res})
}

// ------ Risks ------

func (h *DashboardAnalyticsHandler) Risks(c echo.Context) error {
	f := parseAnalyticsFilter(c)
	page, limit, err := parsePagination(c)
	if err != nil {
		return err
	}
	res, err := h.service.GetRisks(c.Request().Context(), f, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, model.DataResponse[*model.DashboardRisksResponse]{Data: res})
}

// ------ Filter Parser ------

func parseAnalyticsFilter(c echo.Context) model.DashboardAnalyticsFilter {
	f := model.DashboardAnalyticsFilter{}
	if v := parseOptionalInt32(c, "budget_year"); v != nil {
		f.BudgetYear = v
	}
	if v := parseOptionalEnum(c, "quarter", []string{"TW1", "TW2", "TW3", "TW4"}); v != "" {
		f.Quarter = &v
	}
	f.LenderIDs = parseMultiValue(c, "lender_ids")
	f.LenderTypes = parseMultiValue(c, "lender_types")
	f.InstitutionIDs = parseMultiValue(c, "institution_ids")
	f.PipelineStatuses = parseMultiValue(c, "pipeline_statuses")
	f.ProjectStatuses = parseMultiValue(c, "project_statuses")
	f.RegionIDs = parseMultiValue(c, "region_ids")
	f.ProgramTitleIDs = parseMultiValue(c, "program_title_ids")
	if v := parseOptionalFloat64(c, "foreign_loan_min"); v != nil {
		f.ForeignLoanMin = v
	}
	if v := parseOptionalFloat64(c, "foreign_loan_max"); v != nil {
		f.ForeignLoanMax = v
	}
	if v := strings.TrimSpace(c.QueryParam("include_history")); v != "" {
		f.IncludeHistory = v == "true" || v == "1"
	}
	return f
}

func parsePagination(c echo.Context) (page, limit int, err error) {
	page = 1
	limit = 20
	if v := strings.TrimSpace(c.QueryParam("page")); v != "" {
		page, err = strconv.Atoi(v)
		if err != nil || page < 1 {
			return 0, 0, apperrors.Validation(apperrors.FieldError{Field: "page", Message: "harus angka >= 1"})
		}
	}
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit < 1 || limit > 100 {
			return 0, 0, apperrors.Validation(apperrors.FieldError{Field: "limit", Message: "harus angka 1-100"})
		}
	}
	return
}

func parseOptionalInt32(c echo.Context, key string) *int32 {
	v := strings.TrimSpace(c.QueryParam(key))
	if v == "" { return nil }
	parsed, err := strconv.ParseInt(v, 10, 32)
	if err != nil { return nil }
	val := int32(parsed)
	return &val
}

func parseOptionalFloat64(c echo.Context, key string) *float64 {
	v := strings.TrimSpace(c.QueryParam(key))
	if v == "" { return nil }
	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil { return nil }
	return &parsed
}

func parseOptionalEnum(c echo.Context, key string, allowed []string) string {
	v := strings.ToUpper(strings.TrimSpace(c.QueryParam(key)))
	if v == "" { return "" }
	for _, a := range allowed {
		if v == a { return v }
	}
	return ""
}

func parseMultiValue(c echo.Context, key string) []string {
	var raw []string
	params := c.QueryParams()
	if vals, ok := params[key]; ok { raw = append(raw, vals...) }
	if vals, ok := params[key+"[]"]; ok { raw = append(raw, vals...) }
	if len(raw) == 0 { return nil }
	var result []string
	for _, val := range raw {
		for _, p := range strings.Split(val, ",") {
			p = strings.TrimSpace(p)
			if p != "" { result = append(result, p) }
		}
	}
	if len(result) == 0 { return nil }
	return result
}
