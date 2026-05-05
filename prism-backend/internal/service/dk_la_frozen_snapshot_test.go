package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestDKLAFrozenSnapshotResolvesLatestAtCreateAndKeepsStoredConcreteVersion(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	dkService := NewDKService(env.pool, env.queries, nil)
	laService := NewLAService(env.pool, env.queries, nil)

	legacyLender := env.createTestLender(t, "Legacy Development Bank", "LDB")
	storedLender := env.createTestLender(t, "Stored Development Bank", "SDB")
	newerLender := env.createTestLender(t, "New Revision Bank", "NRB")

	_, oldBBProject, _ := env.createBBRevisionPair(t)

	originalGB := env.createGreenBook(t, gbService, 0, nil)
	originalGBProject := env.createGBProjectWithFundingLender(t, gbService, originalGB.ID, oldBBProject.ID, "GB-001", "Flood Control GB", legacyLender)

	revision1GB := env.createGreenBook(t, gbService, 1, &originalGB.ID)
	revision1Project := env.createGBProjectWithFundingLender(t, gbService, revision1GB.ID, oldBBProject.ID, "GB-001", "Flood Control GB Revision 1", storedLender)

	dk := env.createDaftarKegiatan(t, dkService, "DK-001")
	_, err := dkService.CreateDKProject(env.ctx, mustParseUUID(t, dk.ID), env.dkProjectRequest(originalGBProject.ID, legacyLender))
	assertAppErrorCode(t, err, "BUSINESS_RULE_ERROR")

	dkProject := env.createDKProject(t, dkService, dk.ID, originalGBProject.ID, storedLender)
	if len(dkProject.GBProjects) != 1 {
		t.Fatalf("DK GB project relations = %d, want 1", len(dkProject.GBProjects))
	}
	if dkProject.GBProjects[0].ID != revision1Project.ID {
		t.Fatalf("stored DK GB project ID = %s, want latest-at-create %s", dkProject.GBProjects[0].ID, revision1Project.ID)
	}
	if !dkProject.GBProjects[0].IsLatest || dkProject.GBProjects[0].HasNewerRevision {
		t.Fatalf("stored DK GB latest flags before later revision = %+v, want latest without newer revision", dkProject.GBProjects[0])
	}

	env.assertDKAllowedLenders(t, dkProject.ID, []queries.Lender{storedLender}, []queries.Lender{legacyLender, newerLender})

	revision2GB := env.createGreenBook(t, gbService, 2, &revision1GB.ID)
	revision2Project := env.createGBProjectWithFundingLender(t, gbService, revision2GB.ID, oldBBProject.ID, "GB-001", "Flood Control GB Revision 2", newerLender)

	reloadedDKProject, err := dkService.GetDKProject(env.ctx, mustParseUUID(t, dk.ID), mustParseUUID(t, dkProject.ID))
	if err != nil {
		t.Fatalf("GetDKProject(after newer GB revision) error = %v", err)
	}
	if len(reloadedDKProject.GBProjects) != 1 {
		t.Fatalf("reloaded DK GB project relations = %d, want 1", len(reloadedDKProject.GBProjects))
	}
	if reloadedDKProject.GBProjects[0].ID != revision1Project.ID {
		t.Fatalf("reloaded DK GB project ID = %s, want frozen concrete %s; latest revision is %s", reloadedDKProject.GBProjects[0].ID, revision1Project.ID, revision2Project.ID)
	}
	if reloadedDKProject.GBProjects[0].IsLatest || !reloadedDKProject.GBProjects[0].HasNewerRevision {
		t.Fatalf("reloaded DK GB latest flags = %+v, want historical with newer revision", reloadedDKProject.GBProjects[0])
	}

	env.assertDKAllowedLenders(t, dkProject.ID, []queries.Lender{storedLender}, []queries.Lender{legacyLender, newerLender})

	_, err = laService.CreateLoanAgreement(env.ctx, env.loanAgreementRequest(dkProject.ID, newerLender, "LA-BAD"))
	assertAppErrorCode(t, err, "BUSINESS_RULE_ERROR")

	la, err := laService.CreateLoanAgreement(env.ctx, env.loanAgreementRequest(dkProject.ID, storedLender, "LA-001"))
	if err != nil {
		t.Fatalf("CreateLoanAgreement(stored lender) error = %v", err)
	}
	if la.DKProjectID != dkProject.ID {
		t.Fatalf("LA DKProjectID = %s, want stored DK project %s", la.DKProjectID, dkProject.ID)
	}
	if la.Lender.ID != model.UUIDToString(storedLender.ID) {
		t.Fatalf("LA lender = %s, want stored lender %s", la.Lender.ID, model.UUIDToString(storedLender.ID))
	}
}

func (env *blueBookVersioningTestEnv) createTestLender(t *testing.T, name, shortName string) queries.Lender {
	t.Helper()
	lender, err := env.queries.CreateLender(env.ctx, queries.CreateLenderParams{
		CountryID: pgtype.UUID{},
		Name:      name,
		ShortName: pgtype.Text{String: shortName, Valid: true},
		Type:      "Multilateral",
	})
	if err != nil {
		t.Fatalf("CreateLender(%s) error = %v", name, err)
	}
	return lender
}

func (env *blueBookVersioningTestEnv) createGBProjectWithFundingLender(t *testing.T, service *GreenBookService, greenBookID, bbProjectID, code, name string, lender queries.Lender) *model.GBProjectResponse {
	t.Helper()
	res, err := service.CreateGBProject(env.ctx, mustParseUUID(t, greenBookID), env.gbProjectRequestWithFundingLender(bbProjectID, code, name, lender))
	if err != nil {
		t.Fatalf("CreateGBProjectWithFundingLender(%s) error = %v", code, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) updateGBProjectFundingLender(t *testing.T, service *GreenBookService, greenBookID, gbProjectID, bbProjectID, code, name string, lender queries.Lender) *model.GBProjectResponse {
	t.Helper()
	res, err := service.UpdateGBProject(env.ctx, mustParseUUID(t, greenBookID), mustParseUUID(t, gbProjectID), env.gbProjectRequestWithFundingLender(bbProjectID, code, name, lender))
	if err != nil {
		t.Fatalf("UpdateGBProjectFundingLender(%s) error = %v", code, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) gbProjectRequestWithFundingLender(bbProjectID, code, name string, lender queries.Lender) model.CreateGBProjectRequest {
	req := env.gbProjectRequest(bbProjectID, code, name)
	req.FundingSources = []model.GBFundingSourceItem{{
		LenderID: model.UUIDToString(lender.ID),
		LoanUSD:  100,
	}}
	return req
}

func (env *blueBookVersioningTestEnv) singleGBProject(t *testing.T, service *GreenBookService, greenBookID string) *model.GBProjectResponse {
	t.Helper()
	projects, err := service.ListGBProjects(env.ctx, mustParseUUID(t, greenBookID), model.GBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListGBProjects(%s) error = %v", greenBookID, err)
	}
	if len(projects.Data) != 1 {
		t.Fatalf("ListGBProjects(%s) length = %d, want 1", greenBookID, len(projects.Data))
	}
	return &projects.Data[0]
}

func (env *blueBookVersioningTestEnv) createDaftarKegiatan(t *testing.T, service *DKService, letterNumber string) *model.DaftarKegiatanResponse {
	t.Helper()
	res, err := service.CreateDaftarKegiatan(env.ctx, model.DaftarKegiatanRequest{
		LetterNumber: &letterNumber,
		Subject:      "DK Frozen Snapshot",
		Date:         "2026-03-01",
	})
	if err != nil {
		t.Fatalf("CreateDaftarKegiatan(%s) error = %v", letterNumber, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) createDKProject(t *testing.T, service *DKService, dkID, gbProjectID string, lender queries.Lender) *model.DKProjectResponse {
	t.Helper()
	res, err := service.CreateDKProject(env.ctx, mustParseUUID(t, dkID), env.dkProjectRequest(gbProjectID, lender))
	if err != nil {
		t.Fatalf("CreateDKProject(%s) error = %v", gbProjectID, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) dkProjectRequest(gbProjectID string, lender queries.Lender) model.CreateDKProjectRequest {
	programTitleID := model.UUIDToString(env.programTitle.ID)
	institutionID := model.UUIDToString(env.ea.ID)
	locationID := model.UUIDToString(env.region.ID)
	lenderID := model.UUIDToString(lender.ID)
	duration := int32(36)
	objectives := "Keep downstream relation frozen to the stored GB snapshot"

	return model.CreateDKProjectRequest{
		ProgramTitleID: &programTitleID,
		InstitutionID:  &institutionID,
		ProjectName:    "Frozen DK Snapshot Project",
		Duration:       &duration,
		Objectives:     &objectives,
		GBProjectIDs:   []string{gbProjectID},
		LocationIDs:    []string{locationID},
		FinancingDetails: []model.DKFinancingDetailItem{{
			LenderID:       &lenderID,
			Currency:       "USD",
			AmountOriginal: 100,
			AmountUSD:      100,
		}},
		LoanAllocations: []model.DKLoanAllocationItem{{
			InstitutionID:  &institutionID,
			Currency:       "USD",
			AmountOriginal: 100,
			AmountUSD:      100,
		}},
		ActivityDetails: []model.DKActivityDetailItem{{
			ActivityNumber: 1,
			ActivityName:   "Civil Works",
		}},
	}
}

func (env *blueBookVersioningTestEnv) loanAgreementRequest(dkProjectID string, lender queries.Lender, loanCode string) model.LoanAgreementRequest {
	return model.LoanAgreementRequest{
		DKProjectID:         dkProjectID,
		LenderID:            model.UUIDToString(lender.ID),
		LoanCode:            loanCode,
		AgreementDate:       "2026-05-01",
		EffectiveDate:       "2026-06-01",
		OriginalClosingDate: "2030-12-31",
		ClosingDate:         "2030-12-31",
		Currency:            "USD",
		AmountOriginal:      100,
		AmountUSD:           100,
	}
}

func (env *blueBookVersioningTestEnv) assertDKAllowedLenders(t *testing.T, dkProjectID string, wantPresent []queries.Lender, wantAbsent []queries.Lender) {
	t.Helper()
	allowed, err := env.queries.GetAllowedLenderIDsForDK(env.ctx, mustParseUUID(t, dkProjectID))
	if err != nil {
		t.Fatalf("GetAllowedLenderIDsForDK(%s) error = %v", dkProjectID, err)
	}
	got := make(map[string]struct{}, len(allowed))
	for _, id := range allowed {
		got[model.UUIDToString(id)] = struct{}{}
	}
	for _, lender := range wantPresent {
		id := model.UUIDToString(lender.ID)
		if _, ok := got[id]; !ok {
			t.Fatalf("allowed lenders missing %s; got %v", id, got)
		}
	}
	for _, lender := range wantAbsent {
		id := model.UUIDToString(lender.ID)
		if _, ok := got[id]; ok {
			t.Fatalf("allowed lenders unexpectedly include %s; got %v", id, got)
		}
	}
}
