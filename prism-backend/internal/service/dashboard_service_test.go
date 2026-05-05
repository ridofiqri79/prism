package service

import (
	"testing"

	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestDataQualityIssueTypesIncludePhase7MainRules(t *testing.T) {
	required := []string{
		"BB_WITHOUT_BAPPENAS_PARTNER",
		"BB_INDICATION_WITHOUT_LOI",
		"LOI_WITHOUT_GB",
		"GB_WITHOUT_BB_REFERENCE",
		"GB_WITHOUT_FUNDING_SOURCE",
		"GB_WITHOUT_DISBURSEMENT_PLAN",
		"GB_WITHOUT_ACTIVITY",
		"DK_WITHOUT_FINANCING_DETAIL",
		"DK_WITHOUT_ACTIVITY_DETAIL",
		"DK_WITHOUT_LA",
		"LA_NOT_EFFECTIVE",
		"CURRENCY_USD_MISMATCH",
	}

	for _, issueType := range required {
		if _, ok := dataQualityIssueTypes[issueType]; !ok {
			t.Fatalf("dataQualityIssueTypes missing %s", issueType)
		}
	}
}

func TestOptionalDataQualityIssueTypeAcceptsRequiredIssueTypes(t *testing.T) {
	accepted := []string{
		"BB_WITHOUT_BAPPENAS_PARTNER",
		"GB_WITHOUT_FUNDING_SOURCE",
		"DK_WITHOUT_LA",
		"LA_NOT_EFFECTIVE",
		"CURRENCY_USD_MISMATCH",
	}

	for _, issueType := range accepted {
		value := issueType
		if _, err := optionalDataQualityIssueType(&value); err != nil {
			t.Fatalf("optionalDataQualityIssueType(%s) returned error: %v", issueType, err)
		}
	}
}

func TestOptionalDataQualityIssueTypeRejectsUnknownIssueType(t *testing.T) {
	value := "UNKNOWN_RULE"
	if _, err := optionalDataQualityIssueType(&value); err == nil {
		t.Fatal("optionalDataQualityIssueType returned nil error for unknown issue type")
	}
}

func TestDataQualityIssueSummaryCountsSeverityAndAuditEvents(t *testing.T) {
	auditEvents := 7
	summary := dataQualityIssueSummary([]model.DataQualityIssueItem{
		{Severity: "error"},
		{Severity: "warning"},
		{Severity: "warning"},
		{Severity: "info"},
	}, &auditEvents)

	if summary.TotalIssues != 4 {
		t.Fatalf("TotalIssues = %d, want 4", summary.TotalIssues)
	}
	if summary.ErrorCount != 1 || summary.WarningCount != 2 || summary.InfoCount != 1 {
		t.Fatalf("severity counts = error:%d warning:%d info:%d, want 1/2/1", summary.ErrorCount, summary.WarningCount, summary.InfoCount)
	}
	if summary.AuditEvents == nil || *summary.AuditEvents != 7 {
		t.Fatalf("AuditEvents = %v, want 7", summary.AuditEvents)
	}
}
