package service

import (
	"strconv"
	"testing"
	"time"

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

func TestDashboardAnalyticsRisksUseDatabaseBackedDrilldowns(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	dkService := NewDKService(env.pool, env.queries, nil)
	laService := NewLAService(env.pool, env.queries, nil)
	analytics := NewDashboardAnalyticsService(env.queries)

	lender := env.createKSALender(t, "KSA Risk Fund", "KRF")
	originalBlueBook := env.createBlueBook(t, 0, nil)
	oldLinkedBBProject := env.createBBProject(t, originalBlueBook.ID, "BB-DA-LINK", "Risk Linked Project")
	env.createBBProject(t, originalBlueBook.ID, "BB-DA-STOP", "Risk Stopped Project")
	revisionBlueBook := env.createBlueBook(t, 1, &originalBlueBook.ID)
	latestLinkedBBProject := env.findBBProjectByCode(t, revisionBlueBook.ID, "BB-DA-LINK")
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	gbProject := env.createGBProjectWithFundingLender(t, gbService, greenBook.ID, oldLinkedBBProject.ID, "GB-DA-003", "Risk Analytics GB", lender)
	dk := env.createDaftarKegiatan(t, dkService, "DK-DA-003")

	currentYear, currentQuarter := dashboardAnalyticsCurrentBudgetPeriod(time.Now())
	lowAbsorption := env.createAnalyticsLoanAgreement(t, dkService, laService, dk.ID, gbProject.ID, lender, "LA-LOW-ABS", "2025-01-01", "2030-12-31", "2030-12-31")
	env.createAnalyticsMonitoring(t, lowAbsorption.ID, currentYear, currentQuarter, 200, 40)

	zeroPlanned := env.createAnalyticsLoanAgreement(t, dkService, laService, dk.ID, gbProject.ID, lender, "LA-ZERO-PLANNED", "2025-01-01", "2030-12-31", "2030-12-31")
	env.createAnalyticsMonitoring(t, zeroPlanned.ID, currentYear, currentQuarter, 0, 25)

	env.createAnalyticsLoanAgreement(t, dkService, laService, dk.ID, gbProject.ID, lender, "LA-NO-MON", "2025-01-01", "2030-12-31", "2030-12-31")
	env.createAnalyticsLoanAgreement(t, dkService, laService, dk.ID, gbProject.ID, lender, "LA-FUTURE-NO-MON", futureDate(1, 0, 0), futureDate(4, 0, 0), futureDate(4, 0, 0))

	closingDate := futureDate(0, 6, 0)
	closingRisk := env.createAnalyticsLoanAgreement(t, dkService, laService, dk.ID, gbProject.ID, lender, "LA-CLOSING", "2025-01-01", closingDate, closingDate)
	env.createAnalyticsMonitoring(t, closingRisk.ID, currentYear, currentQuarter, 100, 70)

	originalClosingDate := futureDate(3, 0, 0)
	extendedClosingDate := futureDate(3, 3, 0)
	extended := env.createAnalyticsLoanAgreement(t, dkService, laService, dk.ID, gbProject.ID, lender, "LA-EXTENDED", "2025-01-01", originalClosingDate, extendedClosingDate)
	env.createAnalyticsMonitoring(t, extended.ID, currentYear, currentQuarter, 100, 100)

	risks, err := analytics.Risks(env.ctx, model.DashboardAnalyticsFilter{})
	if err != nil {
		t.Fatalf("Risks() error = %v", err)
	}

	if risks.Summary.LowAbsorptionCount != 1 {
		t.Fatalf("low absorption count = %d, want 1", risks.Summary.LowAbsorptionCount)
	}
	if findLoanAgreementRisk(risks.Watchlists.LowAbsorptionProjects, "LA-LOW-ABS") == nil {
		t.Fatalf("low absorption watchlist = %+v, want LA-LOW-ABS", risks.Watchlists.LowAbsorptionProjects)
	}
	if findLoanAgreementRisk(risks.Watchlists.LowAbsorptionProjects, "LA-ZERO-PLANNED") != nil {
		t.Fatalf("planned zero item should not be flagged as low absorption: %+v", risks.Watchlists.LowAbsorptionProjects)
	}

	if risks.Summary.EffectiveWithoutMonitoringCount != 1 {
		t.Fatalf("effective without monitoring count = %d, want 1", risks.Summary.EffectiveWithoutMonitoringCount)
	}
	effectiveMissing := findLoanAgreementRisk(risks.Watchlists.EffectiveWithoutMonitoring, "LA-NO-MON")
	if effectiveMissing == nil {
		t.Fatalf("effective without monitoring watchlist = %+v, want LA-NO-MON", risks.Watchlists.EffectiveWithoutMonitoring)
	}
	if effectiveMissing.Drilldown.Target != "monitoring" || !drilldownHas(effectiveMissing.Drilldown.Query, "risk_codes", "EFFECTIVE_WITHOUT_MONITORING") {
		t.Fatalf("effective missing drilldown = %+v, want monitoring risk code", effectiveMissing.Drilldown)
	}
	if findLoanAgreementRisk(risks.Watchlists.EffectiveWithoutMonitoring, "LA-FUTURE-NO-MON") != nil {
		t.Fatalf("future effective date must not be flagged effective-without-monitoring: %+v", risks.Watchlists.EffectiveWithoutMonitoring)
	}

	closing := findLoanAgreementRisk(risks.Watchlists.ClosingRisks, "LA-CLOSING")
	if closing == nil {
		t.Fatalf("closing risk watchlist = %+v, want LA-CLOSING", risks.Watchlists.ClosingRisks)
	}
	if closing.AbsorptionPct >= closingRiskAbsorptionThreshold {
		t.Fatalf("closing absorption = %v, want below %v", closing.AbsorptionPct, closingRiskAbsorptionThreshold)
	}
	if closing.MonthsToClosing > int(defaultClosingMonthsThreshold) {
		t.Fatalf("months to closing = %d, want within threshold %d", closing.MonthsToClosing, defaultClosingMonthsThreshold)
	}

	extendedItem := findLoanAgreementRisk(risks.Watchlists.ExtendedLoans, "LA-EXTENDED")
	if extendedItem == nil {
		t.Fatalf("extended loan watchlist = %+v, want LA-EXTENDED", risks.Watchlists.ExtendedLoans)
	}
	if extendedItem.ExtensionDays <= 0 || risks.ExtendedLoanInsight.AverageExtensionDays <= 0 {
		t.Fatalf("extended insight = %+v and item = %+v, want computed extension days", risks.ExtendedLoanInsight, extendedItem)
	}

	plannedZeroIssue := findDataQualityIssue(risks.DataQuality, "PLANNED_ZERO_REALIZED_POSITIVE")
	if plannedZeroIssue == nil {
		t.Fatalf("data quality = %+v, want PLANNED_ZERO_REALIZED_POSITIVE", risks.DataQuality)
	}
	if plannedZeroIssue.Drilldown.Target != "monitoring" || !drilldownHas(plannedZeroIssue.Drilldown.Query, "data_quality_codes", "PLANNED_ZERO_REALIZED_POSITIVE") {
		t.Fatalf("planned zero drilldown = %+v, want monitoring data quality query", plannedZeroIssue.Drilldown)
	}

	bbBottleneck := findPipelineBottleneck(risks.Watchlists.PipelineBottlenecks, "BB")
	if bbBottleneck == nil {
		t.Fatalf("pipeline bottlenecks = %+v, want BB bottleneck", risks.Watchlists.PipelineBottlenecks)
	}
	if bbBottleneck.ProjectCount != 1 {
		t.Fatalf("BB bottleneck count = %d, want latest logical project count 1", bbBottleneck.ProjectCount)
	}
	if latestLinkedBBProject.ID == "" {
		t.Fatal("latest linked BB project should exist")
	}
	if bbBottleneck.Drilldown.Target != "projects" || !drilldownHas(bbBottleneck.Drilldown.Query, "pipeline_statuses", "BB") {
		t.Fatalf("BB bottleneck drilldown = %+v, want projects pipeline_statuses=BB", bbBottleneck.Drilldown)
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

func (env *blueBookVersioningTestEnv) createAnalyticsLoanAgreement(t *testing.T, dkService *DKService, laService *LAService, dkID, gbProjectID string, lender queries.Lender, loanCode, effectiveDate, originalClosingDate, closingDate string) *model.LoanAgreementResponse {
	t.Helper()
	dkProject := env.createDKProject(t, dkService, dkID, gbProjectID, lender)
	req := env.loanAgreementRequest(dkProject.ID, lender, loanCode)
	req.EffectiveDate = effectiveDate
	req.OriginalClosingDate = originalClosingDate
	req.ClosingDate = closingDate
	res, err := laService.CreateLoanAgreement(env.ctx, req)
	if err != nil {
		t.Fatalf("CreateLoanAgreement(%s) error = %v", loanCode, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) findBBProjectByCode(t *testing.T, blueBookID, code string) *model.BBProjectResponse {
	t.Helper()
	projects, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, blueBookID), model.BBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 50})
	if err != nil {
		t.Fatalf("ListBBProjects(%s) error = %v", blueBookID, err)
	}
	for i := range projects.Data {
		if projects.Data[i].BBCode == code {
			return &projects.Data[i]
		}
	}
	t.Fatalf("BB project code %s not found in revision %s; projects=%+v", code, blueBookID, projects.Data)
	return nil
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

func dashboardAnalyticsCurrentBudgetPeriod(now time.Time) (int32, string) {
	year := int32(now.Year())
	switch month := now.Month(); {
	case month >= time.April && month <= time.June:
		return year, "TW1"
	case month >= time.July && month <= time.September:
		return year, "TW2"
	case month >= time.October && month <= time.December:
		return year, "TW3"
	default:
		return year - 1, "TW4"
	}
}

func futureDate(years, months, days int) string {
	return time.Now().AddDate(years, months, days).Format("2006-01-02")
}

func findLoanAgreementRisk(items []model.DashboardAnalyticsLoanAgreementRiskItem, loanCode string) *model.DashboardAnalyticsLoanAgreementRiskItem {
	for i := range items {
		if items[i].LoanCode == loanCode {
			return &items[i]
		}
	}
	return nil
}

func findDataQualityIssue(items []model.DashboardAnalyticsDataQualityItem, code string) *model.DashboardAnalyticsDataQualityItem {
	for i := range items {
		if items[i].Code == code {
			return &items[i]
		}
	}
	return nil
}

func findPipelineBottleneck(items []model.DashboardAnalyticsPipelineBottleneckItem, stage string) *model.DashboardAnalyticsPipelineBottleneckItem {
	for i := range items {
		if items[i].Stage == stage {
			return &items[i]
		}
	}
	return nil
}

func drilldownHas(query map[string][]string, key, want string) bool {
	for _, value := range query[key] {
		if value == want {
			return true
		}
	}
	return false
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
