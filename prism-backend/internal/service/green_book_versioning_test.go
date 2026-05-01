package service

import (
	"bytes"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

func TestGreenBookVersioningAllowsSameCodeAcrossRevisionsButRejectsDuplicateInDocument(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	_, sourceBBProject, latestBBProject := env.createBBRevisionPair(t)
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	env.createGBProject(t, gbService, greenBook.ID, sourceBBProject.ID, "GB-001", "Flood Control GB")

	_, err := gbService.CreateGBProject(env.ctx, mustParseUUID(t, greenBook.ID), env.gbProjectRequest(sourceBBProject.ID, "GB-001", "Duplicate GB"))
	assertAppErrorCode(t, err, "CONFLICT")

	revision := env.createGreenBook(t, gbService, 1, &greenBook.ID)
	projects, err := gbService.ListGBProjects(env.ctx, mustParseUUID(t, revision.ID), model.GBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListGBProjects(revision) error = %v", err)
	}
	if len(projects.Data) != 1 {
		t.Fatalf("revision projects = %d, want 1", len(projects.Data))
	}
	if projects.Data[0].GBCode != "GB-001" {
		t.Fatalf("cloned GB code = %q, want GB-001", projects.Data[0].GBCode)
	}
	if len(projects.Data[0].BBProjects) != 1 || projects.Data[0].BBProjects[0].ID != latestBBProject.ID {
		t.Fatalf("cloned BB relation = %+v, want latest BB %s", projects.Data[0].BBProjects, latestBBProject.ID)
	}

	_, err = gbService.CreateGBProject(env.ctx, mustParseUUID(t, revision.ID), env.gbProjectRequest(sourceBBProject.ID, "GB-001", "Duplicate in revision"))
	assertAppErrorCode(t, err, "CONFLICT")
}

func TestCreateGreenBookRejectsDuplicatePublishYearAndRevisionNumber(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	original := env.createGreenBook(t, gbService, 0, nil)

	_, err := gbService.CreateGreenBook(env.ctx, model.GreenBookRequest{
		PublishYear:    2026,
		RevisionNumber: 0,
	})
	assertAppErrorCode(t, err, "CONFLICT")

	revision := env.createGreenBook(t, gbService, 1, &original.ID)
	_, err = gbService.CreateGreenBook(env.ctx, model.GreenBookRequest{
		PublishYear:         2026,
		ReplacesGreenBookID: &original.ID,
		RevisionNumber:      1,
	})
	assertAppErrorCode(t, err, "CONFLICT")

	current, err := gbService.GetGreenBook(env.ctx, mustParseUUID(t, revision.ID))
	if err != nil {
		t.Fatalf("GetGreenBook(revision) error = %v", err)
	}
	if current.Status != "active" {
		t.Fatalf("revision status after duplicate create = %s, want active", current.Status)
	}
}

func TestCreateGBProjectResolvesOldBBInputToLatestBBSnapshot(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	_, oldBBProject, latestBBProject := env.createBBRevisionPair(t)
	greenBook := env.createGreenBook(t, gbService, 0, nil)

	created := env.createGBProject(t, gbService, greenBook.ID, oldBBProject.ID, "GB-001", "Flood Control GB")
	if len(created.BBProjects) != 1 {
		t.Fatalf("created BB project relations = %d, want 1", len(created.BBProjects))
	}
	if created.BBProjects[0].ID != latestBBProject.ID {
		t.Fatalf("stored BB project ID = %s, want latest %s", created.BBProjects[0].ID, latestBBProject.ID)
	}
	if !created.BBProjects[0].IsLatest || created.BBProjects[0].HasNewerRevision {
		t.Fatalf("stored BB latest flags = %+v, want latest without newer revision", created.BBProjects[0])
	}
}

func TestGreenBookRevisionClonePreservesIdentityAndUsesLatestBB(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	_, oldBBProject, latestBBProject := env.createBBRevisionPair(t)
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	sourceGBProject := env.createGBProject(t, gbService, greenBook.ID, oldBBProject.ID, "GB-001", "Flood Control GB")
	revision := env.createGreenBook(t, gbService, 1, &greenBook.ID)

	projects, err := gbService.ListGBProjects(env.ctx, mustParseUUID(t, revision.ID), model.GBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListGBProjects(revision) error = %v", err)
	}
	if len(projects.Data) != 1 {
		t.Fatalf("revision projects = %d, want 1", len(projects.Data))
	}

	cloned := projects.Data[0]
	if cloned.ID == sourceGBProject.ID {
		t.Fatal("cloned GB Project reused source snapshot id")
	}
	if cloned.GBProjectIdentityID != sourceGBProject.GBProjectIdentityID {
		t.Fatalf("cloned GB identity = %s, want %s", cloned.GBProjectIdentityID, sourceGBProject.GBProjectIdentityID)
	}
	if len(cloned.BBProjects) != 1 || cloned.BBProjects[0].ID != latestBBProject.ID {
		t.Fatalf("cloned BB relation = %+v, want latest BB %s", cloned.BBProjects, latestBBProject.ID)
	}
	if len(cloned.ExecutingAgencies) != 1 || len(cloned.ImplementingAgencies) != 1 || len(cloned.Locations) != 1 {
		t.Fatalf("cloned relations lengths EA=%d IA=%d locations=%d, want 1 each", len(cloned.ExecutingAgencies), len(cloned.ImplementingAgencies), len(cloned.Locations))
	}
	if !cloned.IsLatest || cloned.HasNewerRevision {
		t.Fatalf("cloned latest flags = is_latest:%v has_newer:%v, want latest without newer", cloned.IsLatest, cloned.HasNewerRevision)
	}

	sourceAfterRevision, err := gbService.GetGBProject(env.ctx, mustParseUUID(t, greenBook.ID), mustParseUUID(t, sourceGBProject.ID))
	if err != nil {
		t.Fatalf("GetGBProject(source after revision) error = %v", err)
	}
	if sourceAfterRevision.IsLatest || !sourceAfterRevision.HasNewerRevision {
		t.Fatalf("source latest flags = is_latest:%v has_newer:%v, want historical with newer revision", sourceAfterRevision.IsLatest, sourceAfterRevision.HasNewerRevision)
	}
}

func TestDeleteGBProjectHardDeletesWhenNoDownstreamAndAuditsChildren(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	blueBook := env.createBlueBook(t, 0, nil)
	bbProject := env.createBBProject(t, blueBook.ID, "BB-DEL-GB-001", "BB For Delete")
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	gbProject := env.createGBProject(t, gbService, greenBook.ID, bbProject.ID, "GB-DEL-001", "Wrong GB")
	gbProjectID := mustParseUUID(t, gbProject.ID)

	if err := gbService.DeleteGBProject(env.ctx, mustParseUUID(t, greenBook.ID), gbProjectID, staffDeleteUser()); err != nil {
		t.Fatalf("DeleteGBProject(no downstream) error = %v", err)
	}
	if _, err := env.queries.GetGBProject(env.ctx, gbProjectID); err != pgx.ErrNoRows {
		t.Fatalf("GetGBProject after hard delete error = %v, want pgx.ErrNoRows", err)
	}
	assertAuditDeleteExists(t, env, "gb_project", gbProject.ID)
	assertAuditDeleteExists(t, env, "gb_project_bb_project", gbProject.ID)
	assertAuditDeleteExists(t, env, "gb_project_institution", gbProject.ID)
	assertAuditDeleteExists(t, env, "gb_activity", gbProject.Activities[0].ID)
}

func TestDeleteGBProjectRejectsDKDownstreamAndShowsRelations(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	blueBook := env.createBlueBook(t, 0, nil)
	bbProject := env.createBBProject(t, blueBook.ID, "BB-USED-GB-001", "BB Used By GB")
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	gbProject := env.createGBProject(t, gbService, greenBook.ID, bbProject.ID, "GB-USED-001", "Used GB")
	env.createDKProjectLinkedToGB(t, gbProject.ID)

	err := gbService.DeleteGBProject(env.ctx, mustParseUUID(t, greenBook.ID), mustParseUUID(t, gbProject.ID), staffDeleteUser())
	assertAppErrorCode(t, err, "FORBIDDEN")
	assertAppErrorHasDetailField(t, err, "daftar_kegiatan_project")

	err = gbService.DeleteGBProject(env.ctx, mustParseUUID(t, greenBook.ID), mustParseUUID(t, gbProject.ID), adminDeleteUser())
	assertAppErrorCode(t, err, "CONFLICT")
	assertAppErrorHasDetailField(t, err, "daftar_kegiatan_project")
}

func TestGreenBookRevisionCloneMapsFundingAllocationToClonedActivity(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	_, oldBBProject, _ := env.createBBRevisionPair(t)
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	sourceGBProject := env.createGBProject(t, gbService, greenBook.ID, oldBBProject.ID, "GB-001", "Flood Control GB")
	revision := env.createGreenBook(t, gbService, 1, &greenBook.ID)

	projects, err := gbService.ListGBProjects(env.ctx, mustParseUUID(t, revision.ID), model.GBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListGBProjects(revision) error = %v", err)
	}
	cloned := projects.Data[0]
	if len(sourceGBProject.Activities) != 2 || len(sourceGBProject.FundingAllocations) != 2 {
		t.Fatalf("source activities/allocations = %d/%d, want 2/2", len(sourceGBProject.Activities), len(sourceGBProject.FundingAllocations))
	}
	if len(cloned.Activities) != 2 || len(cloned.FundingAllocations) != 2 {
		t.Fatalf("cloned activities/allocations = %d/%d, want 2/2", len(cloned.Activities), len(cloned.FundingAllocations))
	}

	sourceActivityIDs := map[string]struct{}{}
	for _, activity := range sourceGBProject.Activities {
		sourceActivityIDs[activity.ID] = struct{}{}
	}
	for _, allocation := range cloned.FundingAllocations {
		if _, exists := sourceActivityIDs[allocation.GBActivityID]; exists {
			t.Fatalf("cloned allocation points to source activity id %s", allocation.GBActivityID)
		}
		if allocation.ActivityName == "Design" && allocation.Services != 10 {
			t.Fatalf("Design services = %.2f, want 10", allocation.Services)
		}
		if allocation.ActivityName == "Construction" && allocation.Constructions != 20 {
			t.Fatalf("Construction constructions = %.2f, want 20", allocation.Constructions)
		}
	}
}

func TestGetGBProjectHistoryReturnsOrderedSnapshotsAndConcreteBBRelations(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	_, oldBBProject, latestBBProject := env.createBBRevisionPair(t)
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	sourceGBProject := env.createGBProject(t, gbService, greenBook.ID, oldBBProject.ID, "GB-001", "Flood Control GB")
	env.createGreenBook(t, gbService, 1, &greenBook.ID)

	history, err := gbService.GetGBProjectHistory(env.ctx, mustParseUUID(t, sourceGBProject.ID))
	if err != nil {
		t.Fatalf("GetGBProjectHistory() error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("history length = %d, want 2: %+v", len(history), history)
	}
	if history[0].ID != sourceGBProject.ID {
		t.Fatalf("history[0].ID = %s, want source %s", history[0].ID, sourceGBProject.ID)
	}
	if history[0].BookStatus != "superseded" || history[0].IsLatest {
		t.Fatalf("history[0] status/latest = %s/%v, want superseded/not latest", history[0].BookStatus, history[0].IsLatest)
	}
	if history[1].BookStatus != "active" || !history[1].IsLatest {
		t.Fatalf("history[1] status/latest = %s/%v, want active/latest", history[1].BookStatus, history[1].IsLatest)
	}
	if history[0].GBProjectIdentityID != history[1].GBProjectIdentityID {
		t.Fatalf("history identities differ: %s vs %s", history[0].GBProjectIdentityID, history[1].GBProjectIdentityID)
	}
	if len(history[1].BBProjects) != 1 || history[1].BBProjects[0].ID != latestBBProject.ID {
		t.Fatalf("history latest BB relation = %+v, want latest BB %s", history[1].BBProjects, latestBBProject.ID)
	}
}

func TestGreenBookImportReusesPreviousIdentityResolvesLatestBBAndScopedInstitutionPath(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	_, oldBBProject, latestBBProject := env.createBBRevisionPair(t)
	sourceGreenBook := env.createGreenBook(t, gbService, 0, nil)
	sourceGBProject := env.createGBProject(t, gbService, sourceGreenBook.ID, oldBBProject.ID, "GB-001", "Flood Control GB")
	if err := env.queries.SupersedeGreenBooksByPublishYear(env.ctx, 2026); err != nil {
		t.Fatalf("SupersedeGreenBooksByPublishYear() error = %v", err)
	}
	targetGreenBook, err := env.queries.CreateGreenBook(env.ctx, queries.CreateGreenBookParams{
		PublishYear:         2026,
		ReplacesGreenBookID: mustParseUUID(t, sourceGreenBook.ID),
		RevisionNumber:      1,
	})
	if err != nil {
		t.Fatalf("CreateGreenBook(target revision) error = %v", err)
	}
	_, _ = env.createScopedInstitution(t, "Kementerian C", "Sekretariat Utama")
	targetInstitution, targetPath := env.createScopedInstitution(t, "Kementerian D", "Sekretariat Utama")
	lender := env.createMultilateralLender(t, "World Test Bank")

	ambiguousWorkbook := buildGreenBookRevisionImportWorkbook(t, env, "GB-001", "Flood Control GB Revision", "BB-001", "Sekretariat Utama", env.ia.Name, lender.Name, "")
	ambiguousRes, err := gbService.PreviewGreenBookProjects(env.ctx, targetGreenBook.ID, "green-book-projects.xlsx", bytes.NewReader(ambiguousWorkbook), int64(len(ambiguousWorkbook)))
	if err != nil {
		t.Fatalf("PreviewGreenBookProjects(ambiguous) error = %v", err)
	}
	eaSheet := findImportSheet(t, ambiguousRes, greenBookImportSheetEA)
	if eaSheet.Failed == 0 || !importSheetHasMessage(eaSheet, "ambigu") {
		t.Fatalf("EA sheet did not report ambiguous institution: %+v", eaSheet)
	}

	workbook := buildGreenBookRevisionImportWorkbook(t, env, "GB-001", "Flood Control GB Revision", "BB-001", targetPath, env.ia.Name, lender.Name, targetPath)
	res, err := gbService.ImportGreenBookProjects(env.ctx, targetGreenBook.ID, "green-book-projects.xlsx", bytes.NewReader(workbook), int64(len(workbook)))
	if err != nil {
		t.Fatalf("ImportGreenBookProjects(scoped) error = %v", err)
	}
	if res.TotalFailed != 0 {
		t.Fatalf("TotalFailed = %d, response = %+v", res.TotalFailed, res)
	}
	imported, err := env.queries.GetGBProjectByGreenBookAndCode(env.ctx, queries.GetGBProjectByGreenBookAndCodeParams{GreenBookID: targetGreenBook.ID, Lower: "GB-001"})
	if err != nil {
		t.Fatalf("GetGBProjectByGreenBookAndCode(target) error = %v", err)
	}
	if model.UUIDToString(imported.GbProjectIdentityID) != sourceGBProject.GBProjectIdentityID {
		t.Fatalf("imported GB identity = %s, want %s", model.UUIDToString(imported.GbProjectIdentityID), sourceGBProject.GBProjectIdentityID)
	}
	detail, err := gbService.GetGBProject(env.ctx, targetGreenBook.ID, imported.ID)
	if err != nil {
		t.Fatalf("GetGBProject(imported) error = %v", err)
	}
	if len(detail.BBProjects) != 1 || detail.BBProjects[0].ID != latestBBProject.ID {
		t.Fatalf("imported BB relation = %+v, want latest BB %s", detail.BBProjects, latestBBProject.ID)
	}
	if !gbDetailHasInstitution(detail.ExecutingAgencies, targetInstitution.ID) {
		t.Fatalf("executing agencies = %+v, want scoped institution %s", detail.ExecutingAgencies, model.UUIDToString(targetInstitution.ID))
	}
	if len(detail.FundingSources) != 1 || detail.FundingSources[0].Institution == nil || detail.FundingSources[0].Institution.ID != model.UUIDToString(targetInstitution.ID) {
		t.Fatalf("funding sources = %+v, want scoped institution %s", detail.FundingSources, model.UUIDToString(targetInstitution.ID))
	}
}

func TestGreenBookImportRejectsDuplicateGBCodeInWorkbook(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	workbook := simpleXLSXWorkbook{Sheets: []simpleXLSXSheet{
		{
			Name: greenBookImportSheetInput,
			Rows: [][]simpleXLSXCell{
				headerRow("Program Title (*)", "GB Code (*)", "Project Name (*)"),
				textRow(env.programTitle.Title, "GB-DUP", "Duplicate A"),
				textRow(env.programTitle.Title, "GB-DUP", "Duplicate B"),
			},
			Columns:       columns(28, 18, 48),
			ShowGridLines: false,
		},
		emptyImportSheet(greenBookImportSheetBBProject, "GB Code (*)", "BB Code (*)"),
		emptyImportSheet(greenBookImportSheetEA, "GB Code (*)", "Executing Agency Name (*)"),
		emptyImportSheet(greenBookImportSheetIA, "GB Code (*)", "Implementing Agency Name (*)"),
		emptyImportSheet(greenBookImportSheetLocations, "GB Code (*)", "Location Name (*)"),
		emptyImportSheet(greenBookImportSheetActivities, "GB Code (*)", "Activity No (*)", "Activity Name (*)", "Implementation Location", "PIU", "Sort Order"),
		emptyImportSheet(greenBookImportSheetFundingSource, "GB Code (*)", "Lender Name (*)", "Institution Name", "Currency", "Loan Original", "Grant Original", "Local Original", "Loan USD", "Grant USD", "Local USD"),
		emptyImportSheet(greenBookImportSheetDisbursementPlan, "GB Code (*)", "Year (*)", "Amount USD"),
		emptyImportSheet(greenBookImportSheetFundingAllocation, "GB Code (*)", "Activity No (*)", "Services", "Constructions", "Goods", "Trainings", "Other"),
	}}
	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		t.Fatalf("buildSimpleXLSX() error = %v", err)
	}

	res, err := gbService.PreviewGreenBookProjects(env.ctx, mustParseUUID(t, greenBook.ID), "green-book-projects.xlsx", bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("PreviewGreenBookProjects() error = %v", err)
	}
	inputSheet := findImportSheet(t, res, greenBookImportSheetInput)
	if inputSheet.Failed == 0 || !importSheetHasMessage(inputSheet, "GB Code duplikat di workbook") {
		t.Fatalf("input sheet did not report duplicate GB Code: %+v", inputSheet)
	}
}

func (env *blueBookVersioningTestEnv) createBBRevisionPair(t *testing.T) (*model.BlueBookResponse, *model.BBProjectResponse, *model.BBProjectResponse) {
	t.Helper()
	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-001", "Flood Control")
	revision := env.createBlueBook(t, 1, &original.ID)
	projects, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, revision.ID), model.BBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListBBProjects(BB revision) error = %v", err)
	}
	if len(projects.Data) != 1 {
		t.Fatalf("BB revision projects = %d, want 1", len(projects.Data))
	}
	return original, sourceProject, &projects.Data[0]
}

func (env *blueBookVersioningTestEnv) createGreenBook(t *testing.T, service *GreenBookService, revisionNumber int32, replacesID *string) *model.GreenBookResponse {
	t.Helper()
	res, err := service.CreateGreenBook(env.ctx, model.GreenBookRequest{
		PublishYear:         2026,
		ReplacesGreenBookID: replacesID,
		RevisionNumber:      revisionNumber,
	})
	if err != nil {
		t.Fatalf("CreateGreenBook(revision %d) error = %v", revisionNumber, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) createGBProject(t *testing.T, service *GreenBookService, greenBookID, bbProjectID, code, name string) *model.GBProjectResponse {
	t.Helper()
	res, err := service.CreateGBProject(env.ctx, mustParseUUID(t, greenBookID), env.gbProjectRequest(bbProjectID, code, name))
	if err != nil {
		t.Fatalf("CreateGBProject(%s) error = %v", code, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) createDKProjectLinkedToGB(t *testing.T, gbProjectID string) queries.DkProject {
	t.Helper()
	date := pgtype.Date{Time: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Valid: true}
	dk, err := env.queries.CreateDaftarKegiatan(env.ctx, queries.CreateDaftarKegiatanParams{
		LetterNumber: pgtype.Text{String: "DK-DEL-001", Valid: true},
		Subject:      "DK Delete Dependency",
		Date:         date,
	})
	if err != nil {
		t.Fatalf("CreateDaftarKegiatan error = %v", err)
	}
	project, err := env.queries.CreateDKProject(env.ctx, queries.CreateDKProjectParams{
		DkID:           dk.ID,
		ProgramTitleID: env.programTitle.ID,
		InstitutionID:  env.ea.ID,
		Duration:       pgtype.Int4{Int32: 12, Valid: true},
		Objectives:     pgtype.Text{String: "Dependency", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateDKProject error = %v", err)
	}
	if err := env.queries.AddDKProjectGBProject(env.ctx, queries.AddDKProjectGBProjectParams{
		DkProjectID: project.ID,
		GbProjectID: mustParseUUID(t, gbProjectID),
	}); err != nil {
		t.Fatalf("AddDKProjectGBProject error = %v", err)
	}
	return project
}

func (env *blueBookVersioningTestEnv) gbProjectRequest(bbProjectID, code, name string) model.CreateGBProjectRequest {
	programTitleID := model.UUIDToString(env.programTitle.ID)
	designOrder := int32(1)
	constructionOrder := int32(2)
	return model.CreateGBProjectRequest{
		ProgramTitleID:        &programTitleID,
		GBCode:                code,
		ProjectName:           name,
		BBProjectIDs:          []string{bbProjectID},
		ExecutingAgencyIDs:    []string{model.UUIDToString(env.ea.ID)},
		ImplementingAgencyIDs: []string{model.UUIDToString(env.ia.ID)},
		LocationIDs:           []string{model.UUIDToString(env.region.ID)},
		Activities: []model.GBActivityItem{
			{ActivityName: "Design", SortOrder: &designOrder},
			{ActivityName: "Construction", SortOrder: &constructionOrder},
		},
		DisbursementPlan: []model.GBDisbursementPlanItem{
			{Year: 2026, AmountUSD: 100},
			{Year: 2027, AmountUSD: 200},
		},
		FundingAllocations: []model.GBFundingAllocationItem{
			{ActivityIndex: 0, Services: 10, Constructions: 1, Goods: 2, Trainings: 3, Other: 4},
			{ActivityIndex: 1, Services: 5, Constructions: 20, Goods: 6, Trainings: 7, Other: 8},
		},
	}
}

func (env *blueBookVersioningTestEnv) createMultilateralLender(t *testing.T, name string) queries.Lender {
	t.Helper()
	lender, err := env.queries.CreateLender(env.ctx, queries.CreateLenderParams{
		CountryID: pgtype.UUID{},
		Name:      name,
		ShortName: pgtype.Text{},
		Type:      "Multilateral",
	})
	if err != nil {
		t.Fatalf("CreateLender(%s) error = %v", name, err)
	}
	return lender
}

func buildGreenBookRevisionImportWorkbook(t *testing.T, env *blueBookVersioningTestEnv, code, projectName, bbCode, executingAgencyRef, implementingAgencyRef, lenderName, fundingInstitutionRef string) []byte {
	t.Helper()
	fundingRows := [][]simpleXLSXCell{
		headerRow("GB Code (*)", "Lender Name (*)", "Institution Name", "Currency", "Loan Original", "Grant Original", "Local Original", "Loan USD", "Grant USD", "Local USD"),
	}
	if lenderName != "" {
		fundingRows = append(fundingRows, textRow(code, lenderName, fundingInstitutionRef, "USD", "100", "0", "0", "100", "0", "0"))
	}
	workbook := simpleXLSXWorkbook{Sheets: []simpleXLSXSheet{
		{
			Name: greenBookImportSheetInput,
			Rows: [][]simpleXLSXCell{
				headerRow("Program Title (*)", "GB Code (*)", "Project Name (*)"),
				textRow(env.programTitle.Title, code, projectName),
			},
			Columns:       columns(28, 18, 48),
			ShowGridLines: false,
		},
		{
			Name: greenBookImportSheetBBProject,
			Rows: [][]simpleXLSXCell{
				headerRow("GB Code (*)", "BB Code (*)"),
				textRow(code, bbCode),
			},
			Columns:       columns(18, 18),
			ShowGridLines: false,
		},
		{
			Name: greenBookImportSheetEA,
			Rows: [][]simpleXLSXCell{
				headerRow("GB Code (*)", "Executing Agency Name (*)"),
				textRow(code, executingAgencyRef),
			},
			Columns:       columns(18, 44),
			ShowGridLines: false,
		},
		{
			Name: greenBookImportSheetIA,
			Rows: [][]simpleXLSXCell{
				headerRow("GB Code (*)", "Implementing Agency Name (*)"),
				textRow(code, implementingAgencyRef),
			},
			Columns:       columns(18, 44),
			ShowGridLines: false,
		},
		{
			Name: greenBookImportSheetLocations,
			Rows: [][]simpleXLSXCell{
				headerRow("GB Code (*)", "Location Name (*)"),
				textRow(code, env.region.Name),
			},
			Columns:       columns(18, 36),
			ShowGridLines: false,
		},
		emptyImportSheet(greenBookImportSheetActivities, "GB Code (*)", "Activity No (*)", "Activity Name (*)", "Implementation Location", "PIU", "Sort Order"),
		{
			Name:          greenBookImportSheetFundingSource,
			Rows:          fundingRows,
			Columns:       columns(18, 36, 44, 14, 18, 18, 18, 18, 18, 18),
			ShowGridLines: false,
		},
		emptyImportSheet(greenBookImportSheetDisbursementPlan, "GB Code (*)", "Year (*)", "Amount USD"),
		emptyImportSheet(greenBookImportSheetFundingAllocation, "GB Code (*)", "Activity No (*)", "Services", "Constructions", "Goods", "Trainings", "Other"),
	}}
	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		t.Fatalf("buildSimpleXLSX() error = %v", err)
	}
	return data
}

func gbDetailHasInstitution(rows []model.InstitutionResponse, institutionID pgtype.UUID) bool {
	for _, row := range rows {
		if row.ID == model.UUIDToString(institutionID) {
			return true
		}
	}
	return false
}
