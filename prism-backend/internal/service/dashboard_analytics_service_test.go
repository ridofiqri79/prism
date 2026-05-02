package service

import (
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestDashboardAnalyticsAggregatesUseLatestSnapshotAndLegalLenderStages(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	dkService := NewDKService(env.pool, env.queries, nil)
	laService := NewLAService(env.pool, env.queries, nil)
	analytics := NewDashboardAnalyticsService(env.queries)

	ksaLender := env.createKSALender(t, "Saudi Fund for Development", "SFD")
	indicationOnlyLender := env.createTestLender(t, "Indication Only Bank", "IOB")

	_, oldBBProject, latestBBProject := env.createBBRevisionPair(t)
	if _, err := env.queries.CreateLenderIndication(env.ctx, queries.CreateLenderIndicationParams{
		BbProjectID: mustParseUUID(t, latestBBProject.ID),
		LenderID:    indicationOnlyLender.ID,
		Remarks:     pgtype.Text{String: "pipeline indication only", Valid: true},
	}); err != nil {
		t.Fatalf("CreateLenderIndication() error = %v", err)
	}

	greenBook := env.createGreenBook(t, gbService, 0, nil)
	gbProject := env.createGBProjectWithFundingLender(t, gbService, greenBook.ID, oldBBProject.ID, "GB-DA-001", "Analytics GB", ksaLender)
	dk := env.createDaftarKegiatan(t, dkService, "DK-DA-001")
	dkProject := env.createDKProject(t, dkService, dk.ID, gbProject.ID, ksaLender)
	la, err := laService.CreateLoanAgreement(env.ctx, env.loanAgreementRequest(dkProject.ID, ksaLender, "LA-DA-001"))
	if err != nil {
		t.Fatalf("CreateLoanAgreement() error = %v", err)
	}
	env.createAnalyticsMonitoring(t, la.ID, 2025, "TW2", 0, 50)
	env.createAnalyticsMonitoring(t, la.ID, 2025, "TW1", 200, 100)

	overview, err := analytics.Overview(env.ctx, model.DashboardAnalyticsFilter{})
	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Portfolio.ProjectCount != 1 {
		t.Fatalf("overview project count = %d, want latest logical count 1", overview.Portfolio.ProjectCount)
	}
	if overview.Portfolio.AssignmentCount != 1 {
		t.Fatalf("overview assignment count = %d, want 1", overview.Portfolio.AssignmentCount)
	}
	if overview.Portfolio.AbsorptionPct != 75 {
		t.Fatalf("overview absorption = %v, want 75", overview.Portfolio.AbsorptionPct)
	}

	withHistory := true
	historyOverview, err := analytics.Overview(env.ctx, model.DashboardAnalyticsFilter{IncludeHistory: withHistory})
	if err != nil {
		t.Fatalf("Overview(include_history) error = %v", err)
	}
	if historyOverview.Portfolio.ProjectCount != 1 || historyOverview.Portfolio.AssignmentCount != 2 {
		t.Fatalf("history overview project/assignment = %d/%d, want logical project 1 and two snapshot assignments", historyOverview.Portfolio.ProjectCount, historyOverview.Portfolio.AssignmentCount)
	}

	lenders, err := analytics.Lenders(env.ctx, model.DashboardAnalyticsFilter{})
	if err != nil {
		t.Fatalf("Lenders() error = %v", err)
	}
	if len(lenders.Items) != 1 {
		t.Fatalf("lenders length = %d, want only legal agreement lender", len(lenders.Items))
	}
	if lenders.Items[0].Lender.ID != ksaLenderID(t, ksaLender) {
		t.Fatalf("lender item = %+v, want legal KSA lender", lenders.Items[0].Lender)
	}
	if len(lenders.LenderInstitutionMatrix) != 1 || lenders.LenderInstitutionMatrix[0].Lender.ID != ksaLenderID(t, ksaLender) {
		t.Fatalf("matrix = %+v, want Loan Agreement lender only", lenders.LenderInstitutionMatrix)
	}

	proportion, err := analytics.LenderProportion(env.ctx, model.DashboardAnalyticsFilter{})
	if err != nil {
		t.Fatalf("LenderProportion() error = %v", err)
	}
	if !dashboardProportionHasType(proportion.ByStage, "KSA") {
		t.Fatalf("lender proportion = %+v, want KSA as separate lender type", proportion.ByStage)
	}

	year := int32(2025)
	yearly, err := analytics.Yearly(env.ctx, model.DashboardAnalyticsFilter{BudgetYear: &year})
	if err != nil {
		t.Fatalf("Yearly() error = %v", err)
	}
	if len(yearly.Items) != 2 || yearly.Items[0].Quarter != "TW1" || yearly.Items[1].Quarter != "TW2" {
		t.Fatalf("yearly order = %+v, want TW1 then TW2", yearly.Items)
	}

	quarter := "TW1"
	yearlyFiltered, err := analytics.Yearly(env.ctx, model.DashboardAnalyticsFilter{BudgetYear: &year, Quarter: &quarter})
	if err != nil {
		t.Fatalf("Yearly(filtered) error = %v", err)
	}
	if len(yearlyFiltered.Items) != 1 || yearlyFiltered.Items[0].Quarter != "TW1" {
		t.Fatalf("yearly filtered = %+v, want only TW1", yearlyFiltered.Items)
	}

	lenderTypes := []string{"KSA"}
	institutionIDs := []string{model.UUIDToString(env.ea.ID)}
	filteredInstitutions, err := analytics.Institutions(env.ctx, model.DashboardAnalyticsFilter{LenderTypes: lenderTypes, InstitutionIDs: institutionIDs})
	if err != nil {
		t.Fatalf("Institutions(filtered) error = %v", err)
	}
	if filteredInstitutions.Summary.InstitutionCount != 1 {
		t.Fatalf("filtered institution count = %d, want 1", filteredInstitutions.Summary.InstitutionCount)
	}

	zeroQuarter := "TW2"
	absorption, err := analytics.Absorption(env.ctx, model.DashboardAnalyticsFilter{BudgetYear: &year, Quarter: &zeroQuarter})
	if err != nil {
		t.Fatalf("Absorption(TW2) error = %v", err)
	}
	if absorption.Summary.AbsorptionPct != 0 {
		t.Fatalf("absorption planned zero pct = %v, want 0", absorption.Summary.AbsorptionPct)
	}
	if len(absorption.ByProject) != 1 || absorption.ByProject[0].Status != "low" {
		t.Fatalf("absorption project items = %+v, want low status", absorption.ByProject)
	}
}

func (env *blueBookVersioningTestEnv) createKSALender(t *testing.T, name, shortName string) queries.Lender {
	t.Helper()
	country, err := env.queries.CreateCountry(env.ctx, queries.CreateCountryParams{Name: "Saudi Arabia", Code: "SA"})
	if err != nil {
		t.Fatalf("CreateCountry(KSA) error = %v", err)
	}
	lender, err := env.queries.CreateLender(env.ctx, queries.CreateLenderParams{
		CountryID: country.ID,
		Name:      name,
		ShortName: pgtype.Text{String: shortName, Valid: true},
		Type:      "KSA",
	})
	if err != nil {
		t.Fatalf("CreateLender(KSA) error = %v", err)
	}
	return lender
}

func (env *blueBookVersioningTestEnv) createAnalyticsMonitoring(t *testing.T, loanAgreementID string, budgetYear int32, quarter string, plannedUSD, realizedUSD float64) {
	t.Helper()
	if _, err := env.queries.CreateMonitoring(env.ctx, queries.CreateMonitoringParams{
		LoanAgreementID:    mustParseUUID(t, loanAgreementID),
		BudgetYear:         budgetYear,
		Quarter:            quarter,
		ExchangeRateUsdIdr: testNumeric(t, 16000),
		ExchangeRateLaIdr:  testNumeric(t, 16000),
		PlannedLa:          testNumeric(t, plannedUSD),
		PlannedUsd:         testNumeric(t, plannedUSD),
		PlannedIdr:         testNumeric(t, plannedUSD*16000),
		RealizedLa:         testNumeric(t, realizedUSD),
		RealizedUsd:        testNumeric(t, realizedUSD),
		RealizedIdr:        testNumeric(t, realizedUSD*16000),
	}); err != nil {
		t.Fatalf("CreateMonitoring(%s) error = %v", quarter, err)
	}
}

func testNumeric(t *testing.T, value float64) pgtype.Numeric {
	t.Helper()
	var numeric pgtype.Numeric
	if err := numeric.Scan(strconv.FormatFloat(value, 'f', -1, 64)); err != nil {
		t.Fatalf("numeric scan %v: %v", value, err)
	}
	return numeric
}

func ksaLenderID(t *testing.T, lender queries.Lender) string {
	t.Helper()
	return model.UUIDToString(lender.ID)
}

func dashboardProportionHasType(stages []model.DashboardAnalyticsLenderProportionStage, lenderType string) bool {
	for _, stage := range stages {
		for _, item := range stage.Items {
			if item.Type == lenderType {
				return true
			}
		}
	}
	return false
}
