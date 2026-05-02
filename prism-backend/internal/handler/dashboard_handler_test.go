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
		"/dashboard/summary?period_id=period-1&publish_year=2026&budget_year=2027&quarter=tw2&lender_id=lender-1&institution_id=institution-1&include_history=true",
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
	if filter.BudgetYear == nil || *filter.BudgetYear != 2027 {
		t.Fatalf("budget_year = %v, want 2027", filter.BudgetYear)
	}
	if filter.Quarter == nil || *filter.Quarter != "TW2" {
		t.Fatalf("quarter = %v, want TW2", filter.Quarter)
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

func TestDashboardFilterRejectsInvalidQuarter(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/summary?quarter=Q1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if _, err := dashboardFilter(c); err == nil {
		t.Fatal("dashboardFilter returned nil error for invalid quarter")
	}
}
