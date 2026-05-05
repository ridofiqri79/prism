package service

import (
	"strings"
	"testing"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestJourneyReturnsConcretePathRevisionMetadataAndFundingDetails(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	dkService := NewDKService(env.pool, env.queries, nil)
	laService := NewLAService(env.pool, env.queries, nil)
	journeyService := NewJourneyService(env.queries)

	lender := env.createTestLender(t, "Journey Development Bank", "JDB")
	blueBook := env.createBlueBook(t, 0, nil)
	bbReq := env.bbProjectRequest("BB-JRN-001", "Journey Concrete Path")
	remarks := "Indicative funding"
	bbReq.LenderIndications = []model.LenderIndicationItem{{
		LenderID: model.UUIDToString(lender.ID),
		Remarks:  &remarks,
	}}
	bbProject, err := env.service.CreateBBProject(env.ctx, mustParseUUID(t, blueBook.ID), bbReq)
	if err != nil {
		t.Fatalf("CreateBBProject(journey) error = %v", err)
	}
	loiNumber := "LoI-001"
	if _, err := env.service.CreateLoI(env.ctx, mustParseUUID(t, bbProject.ID), model.LoIRequest{
		LenderID:     model.UUIDToString(lender.ID),
		Subject:      "Journey LoI",
		Date:         "2025-03-10",
		LetterNumber: &loiNumber,
	}); err != nil {
		t.Fatalf("CreateLoI(journey) error = %v", err)
	}

	greenBook := env.createGreenBook(t, gbService, 0, nil)
	gbProject := env.createGBProjectWithFundingLender(t, gbService, greenBook.ID, bbProject.ID, "GB-JRN-001", "Journey GB", lender)
	dk := env.createDaftarKegiatan(t, dkService, "DK-JRN-001")
	dkProject := env.createDKProject(t, dkService, dk.ID, gbProject.ID, lender)
	laReq := env.loanAgreementRequest(dkProject.ID, lender, "LA-JRN-001")
	laReq.AgreementDate = "2025-01-01"
	laReq.EffectiveDate = "2025-02-01"
	laReq.OriginalClosingDate = "2030-12-31"
	laReq.ClosingDate = "2031-01-15"
	la, err := laService.CreateLoanAgreement(env.ctx, laReq)
	if err != nil {
		t.Fatalf("CreateLoanAgreement(journey) error = %v", err)
	}
	if _, err := env.queries.CreateMonitoring(env.ctx, queries.CreateMonitoringParams{
		LoanAgreementID:    mustParseUUID(t, la.ID),
		BudgetYear:         2025,
		Quarter:            "TW1",
		ExchangeRateUsdIdr: numericFromFloat(15000),
		ExchangeRateLaIdr:  numericFromFloat(15000),
		PlannedLa:          numericFromFloat(200),
		PlannedUsd:         numericFromFloat(200),
		PlannedIdr:         numericFromFloat(3000000),
		RealizedLa:         numericFromFloat(100),
		RealizedUsd:        numericFromFloat(100),
		RealizedIdr:        numericFromFloat(1500000),
	}); err != nil {
		t.Fatalf("CreateMonitoring(journey) error = %v", err)
	}

	blueBookRevision := env.createBlueBook(t, 1, &blueBook.ID)
	env.importBBProjectFromBlueBook(t, blueBookRevision.ID, blueBook.ID, bbProject.ID)
	greenBookRevision := env.createGreenBook(t, gbService, 1, &greenBook.ID)
	env.createGBProjectWithFundingLender(t, gbService, greenBookRevision.ID, bbProject.ID, "GB-JRN-001", "Journey GB Revision", lender)

	journey, err := journeyService.GetProjectJourney(env.ctx, mustParseUUID(t, bbProject.ID))
	if err != nil {
		t.Fatalf("GetProjectJourney() error = %v", err)
	}
	if journey.BBProject.ID != bbProject.ID {
		t.Fatalf("journey BB id = %s, want %s", journey.BBProject.ID, bbProject.ID)
	}
	if journey.BBProject.IsLatest || !journey.BBProject.HasNewerRevision {
		t.Fatalf("BB revision flags = latest:%v newer:%v, want historical with newer", journey.BBProject.IsLatest, journey.BBProject.HasNewerRevision)
	}
	if !strings.Contains(journey.BBProject.LatestBlueBookRevisionLabel, "Revisi ke-1") {
		t.Fatalf("latest BB label = %q, want revision label", journey.BBProject.LatestBlueBookRevisionLabel)
	}
	if len(journey.BBProject.LenderIndications) != 1 || journey.BBProject.LenderIndications[0].Remarks == nil {
		t.Fatalf("lender indications = %+v, want one item with remarks", journey.BBProject.LenderIndications)
	}
	if len(journey.LoI) != 1 || journey.LoI[0].LetterNumber == nil || *journey.LoI[0].LetterNumber != loiNumber {
		t.Fatalf("loi = %+v, want letter number %s", journey.LoI, loiNumber)
	}
	if len(journey.GBProjects) != 1 {
		t.Fatalf("GBProjects length = %d, want 1", len(journey.GBProjects))
	}
	gbNode := journey.GBProjects[0]
	if gbNode.ID != gbProject.ID {
		t.Fatalf("journey GB id = %s, want concrete %s", gbNode.ID, gbProject.ID)
	}
	if gbNode.IsLatest || !gbNode.HasNewerRevision {
		t.Fatalf("GB revision flags = latest:%v newer:%v, want historical with newer", gbNode.IsLatest, gbNode.HasNewerRevision)
	}
	if len(gbNode.FundingSources) != 1 || gbNode.FundingSources[0].LoanUSD != 100 {
		t.Fatalf("funding sources = %+v, want one USD 100 loan", gbNode.FundingSources)
	}
	if len(gbNode.DKProjects) != 1 || gbNode.DKProjects[0].ID != dkProject.ID {
		t.Fatalf("DK projects = %+v, want concrete DK %s", gbNode.DKProjects, dkProject.ID)
	}
	if gbNode.DKProjects[0].DaftarKegiatan == nil || gbNode.DKProjects[0].DaftarKegiatan.LetterNumber == nil || *gbNode.DKProjects[0].DaftarKegiatan.LetterNumber != "DK-JRN-001" {
		t.Fatalf("DK header = %+v, want letter number DK-JRN-001", gbNode.DKProjects[0].DaftarKegiatan)
	}
	laNode := gbNode.DKProjects[0].LoanAgreement
	if laNode == nil {
		t.Fatal("loan agreement is nil, want populated")
	}
	if !laNode.IsExtended || laNode.ExtensionDays <= 0 || laNode.Currency != "USD" || laNode.AmountUSD != 100 {
		t.Fatalf("loan agreement = %+v, want extended USD 100", laNode)
	}
	if len(laNode.Monitoring) != 1 || laNode.Monitoring[0].AbsorptionPct != 50 {
		t.Fatalf("monitoring = %+v, want one row with 50%% absorption", laNode.Monitoring)
	}
}
