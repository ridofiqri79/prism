package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestDashboardFilterParsesPhase0Params(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(
		http.MethodGet,
		"/dashboard/summary?period_id=period-1&publish_year=2026&lender_id=lender-1&institution_id=institution-1&include_history=true",
		nil,
	)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	filter, err := dashboardFilter(c)
	if err != nil {
		t.Fatalf("dashboardFilter returned error: %v", err)
	}
	if filter.PeriodID == nil || *filter.PeriodID != "period-1" {
		t.Fatalf("period_id = %v, want period-1", filter.PeriodID)
	}
	if filter.PublishYear == nil || *filter.PublishYear != 2026 {
		t.Fatalf("publish_year = %v, want 2026", filter.PublishYear)
	}
	if filter.LenderID == nil || *filter.LenderID != "lender-1" {
		t.Fatalf("lender_id = %v, want lender-1", filter.LenderID)
	}
	if filter.InstitutionID == nil || *filter.InstitutionID != "institution-1" {
		t.Fatalf("institution_id = %v, want institution-1", filter.InstitutionID)
	}
	if !filter.IncludeHistory {
		t.Fatal("include_history = false, want true")
	}
}

func TestDashboardGreenBookReadinessFilterParsesParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(
		http.MethodGet,
		"/dashboard/green-book-readiness?publish_year=2026&green_book_id=gb-1&institution_id=inst-1&lender_id=lender-1&readiness_status=cofinancing",
		nil,
	)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	filter, err := dashboardGreenBookReadinessFilter(c)
	if err != nil {
		t.Fatalf("dashboardGreenBookReadinessFilter returned error: %v", err)
	}
	if filter.PublishYear == nil || *filter.PublishYear != 2026 {
		t.Fatalf("publish_year = %v, want 2026", filter.PublishYear)
	}
	if filter.GreenBookID == nil || *filter.GreenBookID != "gb-1" {
		t.Fatalf("green_book_id = %v, want gb-1", filter.GreenBookID)
	}
	if filter.InstitutionID == nil || *filter.InstitutionID != "inst-1" {
		t.Fatalf("institution_id = %v, want inst-1", filter.InstitutionID)
	}
	if filter.LenderID == nil || *filter.LenderID != "lender-1" {
		t.Fatalf("lender_id = %v, want lender-1", filter.LenderID)
	}
	if filter.ReadinessStatus == nil || *filter.ReadinessStatus != "COFINANCING" {
		t.Fatalf("readiness_status = %v, want COFINANCING", filter.ReadinessStatus)
	}
}

func TestDashboardGreenBookReadinessFilterRejectsInvalidStatus(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/green-book-readiness?readiness_status=DONE", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, err := dashboardGreenBookReadinessFilter(c); err == nil {
		t.Fatal("dashboardGreenBookReadinessFilter returned nil error for invalid status")
	}
}

func TestDashboardLenderFinancingMixFilterParsesParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(
		http.MethodGet,
		"/dashboard/lender-financing-mix?lender_type=ksa&lender_id=lender-1&currency=jpy&period_id=period-1&publish_year=2026",
		nil,
	)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	filter, err := dashboardLenderFinancingMixFilter(c)
	if err != nil {
		t.Fatalf("dashboardLenderFinancingMixFilter returned error: %v", err)
	}
	if filter.LenderType == nil || *filter.LenderType != "KSA" {
		t.Fatalf("lender_type = %v, want KSA", filter.LenderType)
	}
	if filter.LenderID == nil || *filter.LenderID != "lender-1" {
		t.Fatalf("lender_id = %v, want lender-1", filter.LenderID)
	}
	if filter.Currency == nil || *filter.Currency != "JPY" {
		t.Fatalf("currency = %v, want JPY", filter.Currency)
	}
	if filter.PeriodID == nil || *filter.PeriodID != "period-1" {
		t.Fatalf("period_id = %v, want period-1", filter.PeriodID)
	}
	if filter.PublishYear == nil || *filter.PublishYear != 2026 {
		t.Fatalf("publish_year = %v, want 2026", filter.PublishYear)
	}
}

func TestDashboardLenderFinancingMixFilterRejectsInvalidLenderType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/lender-financing-mix?lender_type=ExportCredit", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, err := dashboardLenderFinancingMixFilter(c); err == nil {
		t.Fatal("dashboardLenderFinancingMixFilter returned nil error for invalid lender_type")
	}
}

func TestDashboardKLPortfolioPerformanceFilterParsesParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(
		http.MethodGet,
		"/dashboard/kl-portfolio-performance?institution_id=inst-1&institution_role=implementing%20agency&period_id=period-1&publish_year=2026&sort_by=risk_count",
		nil,
	)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	filter, err := dashboardKLPortfolioPerformanceFilter(c)
	if err != nil {
		t.Fatalf("dashboardKLPortfolioPerformanceFilter returned error: %v", err)
	}
	if filter.InstitutionID == nil || *filter.InstitutionID != "inst-1" {
		t.Fatalf("institution_id = %v, want inst-1", filter.InstitutionID)
	}
	if filter.InstitutionRole == nil || *filter.InstitutionRole != "Implementing Agency" {
		t.Fatalf("institution_role = %v, want Implementing Agency", filter.InstitutionRole)
	}
	if filter.PeriodID == nil || *filter.PeriodID != "period-1" {
		t.Fatalf("period_id = %v, want period-1", filter.PeriodID)
	}
	if filter.PublishYear == nil || *filter.PublishYear != 2026 {
		t.Fatalf("publish_year = %v, want 2026", filter.PublishYear)
	}
	if filter.SortBy == nil || *filter.SortBy != "risk_count" {
		t.Fatalf("sort_by = %v, want risk_count", filter.SortBy)
	}
}

func TestDashboardKLPortfolioPerformanceFilterRejectsInvalidSort(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/kl-portfolio-performance?sort_by=name", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, err := dashboardKLPortfolioPerformanceFilter(c); err == nil {
		t.Fatal("dashboardKLPortfolioPerformanceFilter returned nil error for invalid sort_by")
	}
}

func TestDashboardDataQualityGovernanceFilterParsesParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(
		http.MethodGet,
		"/dashboard/data-quality-governance?severity=WARNING&module=gb_project&issue_type=gb_without_activity&only_unresolved=true&audit_days=45",
		nil,
	)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	filter, err := dashboardDataQualityGovernanceFilter(c)
	if err != nil {
		t.Fatalf("dashboardDataQualityGovernanceFilter returned error: %v", err)
	}
	if filter.Severity == nil || *filter.Severity != "warning" {
		t.Fatalf("severity = %v, want warning", filter.Severity)
	}
	if filter.Module == nil || *filter.Module != "gb_project" {
		t.Fatalf("module = %v, want gb_project", filter.Module)
	}
	if filter.IssueType == nil || *filter.IssueType != "GB_WITHOUT_ACTIVITY" {
		t.Fatalf("issue_type = %v, want GB_WITHOUT_ACTIVITY", filter.IssueType)
	}
	if !filter.OnlyUnresolved {
		t.Fatal("only_unresolved = false, want true")
	}
	if filter.AuditDays != 45 {
		t.Fatalf("audit_days = %d, want 45", filter.AuditDays)
	}
}

func TestDashboardDataQualityGovernanceFilterDefaultsAuditDays(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/data-quality-governance", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	filter, err := dashboardDataQualityGovernanceFilter(c)
	if err != nil {
		t.Fatalf("dashboardDataQualityGovernanceFilter returned error: %v", err)
	}
	if filter.AuditDays != 30 {
		t.Fatalf("audit_days = %d, want 30", filter.AuditDays)
	}
}

func TestDashboardDataQualityGovernanceFilterRejectsInvalidSeverity(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/data-quality-governance?severity=critical", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, err := dashboardDataQualityGovernanceFilter(c); err == nil {
		t.Fatal("dashboardDataQualityGovernanceFilter returned nil error for invalid severity")
	}
}

func TestDashboardDataQualityGovernanceFilterRejectsInvalidAuditDays(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/data-quality-governance?audit_days=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, err := dashboardDataQualityGovernanceFilter(c); err == nil {
		t.Fatal("dashboardDataQualityGovernanceFilter returned nil error for invalid audit_days")
	}
}
