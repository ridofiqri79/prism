package service

import (
	"testing"

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
	projects, err := gbService.ListGBProjects(env.ctx, mustParseUUID(t, revision.ID), model.PaginationParams{Page: 1, Limit: 10})
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

	projects, err := gbService.ListGBProjects(env.ctx, mustParseUUID(t, revision.ID), model.PaginationParams{Page: 1, Limit: 10})
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

func TestGreenBookRevisionCloneMapsFundingAllocationToClonedActivity(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	_, oldBBProject, _ := env.createBBRevisionPair(t)
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	sourceGBProject := env.createGBProject(t, gbService, greenBook.ID, oldBBProject.ID, "GB-001", "Flood Control GB")
	revision := env.createGreenBook(t, gbService, 1, &greenBook.ID)

	projects, err := gbService.ListGBProjects(env.ctx, mustParseUUID(t, revision.ID), model.PaginationParams{Page: 1, Limit: 10})
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

func (env *blueBookVersioningTestEnv) createBBRevisionPair(t *testing.T) (*model.BlueBookResponse, *model.BBProjectResponse, *model.BBProjectResponse) {
	t.Helper()
	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-001", "Flood Control")
	revision := env.createBlueBook(t, 1, &original.ID)
	projects, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, revision.ID), model.PaginationParams{Page: 1, Limit: 10})
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
