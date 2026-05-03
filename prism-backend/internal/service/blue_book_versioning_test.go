package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type blueBookVersioningTestEnv struct {
	ctx          context.Context
	pool         *pgxpool.Pool
	queries      *queries.Queries
	service      *BlueBookService
	period       queries.Period
	programTitle queries.ProgramTitle
	ea           queries.Institution
	ia           queries.Institution
	region       queries.Region
}

func TestBlueBookVersioningAllowsSameCodeAcrossRevisionsButRejectsDuplicateInDocument(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-001", "Flood Control")

	_, err := env.service.CreateBBProject(env.ctx, mustParseUUID(t, original.ID), env.bbProjectRequest("BB-001", "Duplicate"))
	assertAppErrorCode(t, err, "CONFLICT")

	revision := env.createBlueBook(t, 1, &original.ID)
	projects, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, revision.ID), model.BBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListBBProjects(revision) error = %v", err)
	}
	if len(projects.Data) != 1 {
		t.Fatalf("revision projects = %d, want 1", len(projects.Data))
	}
	cloned := projects.Data[0]
	if cloned.BBCode != sourceProject.BBCode {
		t.Fatalf("cloned BBCode = %q, want %q", cloned.BBCode, sourceProject.BBCode)
	}
	if cloned.ProjectIdentityID != sourceProject.ProjectIdentityID {
		t.Fatalf("cloned identity = %s, want %s", cloned.ProjectIdentityID, sourceProject.ProjectIdentityID)
	}

	_, err = env.service.CreateBBProject(env.ctx, mustParseUUID(t, revision.ID), env.bbProjectRequest("BB-001", "Duplicate in revision"))
	assertAppErrorCode(t, err, "CONFLICT")
}

func TestCreateBBProjectReusesPreviousIdentityWhenCodeMatchesRevisionSource(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-REUSE", "Flood Control")

	if err := env.queries.SupersedeBlueBooksByPeriod(env.ctx, env.period.ID); err != nil {
		t.Fatalf("SupersedeBlueBooksByPeriod error = %v", err)
	}
	revision, err := env.queries.CreateBlueBook(env.ctx, queries.CreateBlueBookParams{
		PeriodID:       env.period.ID,
		PublishDate:    testDate(2026, time.March, 1),
		RevisionNumber: 1,
		RevisionYear:   pgtype.Int4{Int32: 2026, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateBlueBook(manual revision) error = %v", err)
	}

	created, err := env.service.CreateBBProject(env.ctx, revision.ID, env.bbProjectRequest("BB-REUSE", "Flood Control Updated"))
	if err != nil {
		t.Fatalf("CreateBBProject(reuse code) error = %v", err)
	}
	if created.ID == sourceProject.ID {
		t.Fatal("created project reused source snapshot id")
	}
	if created.ProjectIdentityID != sourceProject.ProjectIdentityID {
		t.Fatalf("created identity = %s, want %s", created.ProjectIdentityID, sourceProject.ProjectIdentityID)
	}
}

func TestCreateBlueBookRejectsDuplicatePeriodAndVersion(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	original := env.createBlueBook(t, 0, nil)

	_, err := env.service.CreateBlueBook(env.ctx, model.BlueBookRequest{
		PeriodID:       model.UUIDToString(env.period.ID),
		PublishDate:    "2026-03-01",
		RevisionNumber: 0,
	})
	assertAppErrorCode(t, err, "CONFLICT")

	revision := env.createBlueBook(t, 1, &original.ID)
	revisionYear := int32(2026)
	_, err = env.service.CreateBlueBook(env.ctx, model.BlueBookRequest{
		PeriodID:           model.UUIDToString(env.period.ID),
		ReplacesBlueBookID: &original.ID,
		PublishDate:        "2026-04-01",
		RevisionNumber:     1,
		RevisionYear:       &revisionYear,
	})
	assertAppErrorCode(t, err, "CONFLICT")

	current, err := env.service.GetBlueBook(env.ctx, mustParseUUID(t, revision.ID))
	if err != nil {
		t.Fatalf("GetBlueBook(revision) error = %v", err)
	}
	if current.Status != "active" {
		t.Fatalf("revision status after duplicate create = %s, want active", current.Status)
	}
}

func TestBlueBookRevisionClonePreservesIdentityAndChildren(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-001", "Flood Control")
	revision := env.createBlueBook(t, 1, &original.ID)

	projects, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, revision.ID), model.BBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListBBProjects(revision) error = %v", err)
	}
	if len(projects.Data) != 1 {
		t.Fatalf("revision projects = %d, want 1", len(projects.Data))
	}

	cloned := projects.Data[0]
	if cloned.ID == sourceProject.ID {
		t.Fatal("cloned project reused source snapshot id")
	}
	if cloned.ProjectIdentityID != sourceProject.ProjectIdentityID {
		t.Fatalf("cloned identity = %s, want %s", cloned.ProjectIdentityID, sourceProject.ProjectIdentityID)
	}
	if len(cloned.ExecutingAgencies) != 1 || len(cloned.ImplementingAgencies) != 1 || len(cloned.Locations) != 1 {
		t.Fatalf("cloned relations lengths EA=%d IA=%d locations=%d, want 1 each", len(cloned.ExecutingAgencies), len(cloned.ImplementingAgencies), len(cloned.Locations))
	}
	if len(cloned.ProjectCosts) != 1 || cloned.ProjectCosts[0].FundingType != "Foreign" || cloned.ProjectCosts[0].FundingCategory != "Loan" {
		t.Fatalf("cloned project costs = %+v, want one Foreign/Loan row", cloned.ProjectCosts)
	}
	if !cloned.IsLatest || cloned.HasNewerRevision {
		t.Fatalf("cloned latest flags = is_latest:%v has_newer:%v, want latest with no newer", cloned.IsLatest, cloned.HasNewerRevision)
	}

	sourceAfterRevision, err := env.service.GetBBProject(env.ctx, mustParseUUID(t, original.ID), mustParseUUID(t, sourceProject.ID))
	if err != nil {
		t.Fatalf("GetBBProject(source after revision) error = %v", err)
	}
	if sourceAfterRevision.IsLatest || !sourceAfterRevision.HasNewerRevision {
		t.Fatalf("source latest flags = is_latest:%v has_newer:%v, want historical with newer revision", sourceAfterRevision.IsLatest, sourceAfterRevision.HasNewerRevision)
	}
}

func TestDeleteBBProjectHardDeletesWhenNoDownstreamAndAuditsChildren(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	blueBook := env.createBlueBook(t, 0, nil)
	project := env.createBBProject(t, blueBook.ID, "BB-DEL-001", "Wrong Input")
	projectID := mustParseUUID(t, project.ID)

	if err := env.service.DeleteBBProject(env.ctx, mustParseUUID(t, blueBook.ID), projectID, staffDeleteUser()); err != nil {
		t.Fatalf("DeleteBBProject(no downstream) error = %v", err)
	}
	if _, err := env.queries.GetBBProject(env.ctx, projectID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetBBProject after hard delete error = %v, want pgx.ErrNoRows", err)
	}
	assertAuditDeleteExists(t, env, "bb_project", project.ID)
	assertAuditDeleteExists(t, env, "bb_project_institution", project.ID)
	assertAuditDeleteExists(t, env, "bb_project_location", project.ID)
	assertAuditDeleteExists(t, env, "bb_project_cost", project.ProjectCosts[0].ID)
}

func TestDeleteBBProjectRejectsDownstreamAndShowsRelations(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	gbService := NewGreenBookService(env.pool, env.queries, nil)

	blueBook := env.createBlueBook(t, 0, nil)
	bbProject := env.createBBProject(t, blueBook.ID, "BB-USED-001", "Used Input")
	greenBook := env.createGreenBook(t, gbService, 0, nil)
	env.createGBProject(t, gbService, greenBook.ID, bbProject.ID, "GB-USED-001", "Used GB")

	err := env.service.DeleteBBProject(env.ctx, mustParseUUID(t, blueBook.ID), mustParseUUID(t, bbProject.ID), staffDeleteUser())
	assertAppErrorCode(t, err, "FORBIDDEN")
	assertAppErrorHasDetailField(t, err, "green_book_project")

	err = env.service.DeleteBBProject(env.ctx, mustParseUUID(t, blueBook.ID), mustParseUUID(t, bbProject.ID), adminDeleteUser())
	assertAppErrorCode(t, err, "CONFLICT")
	assertAppErrorHasDetailField(t, err, "green_book_project")
}

func TestListBBProjectsSupportsSearchAndFilters(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	blueBook := env.createBlueBook(t, 0, nil)
	floodProject := env.createBBProject(t, blueBook.ID, "BB-001", "Flood Control")
	agricultureEA, err := env.queries.CreateInstitution(env.ctx, queries.CreateInstitutionParams{ParentID: pgtype.UUID{}, Name: "Ministry of Agriculture", ShortName: pgtype.Text{String: "MoA", Valid: true}, Level: "Kementerian/Badan/Lembaga"})
	if err != nil {
		t.Fatalf("CreateInstitution(agriculture EA) error = %v", err)
	}
	province, err := env.queries.CreateRegion(env.ctx, queries.CreateRegionParams{Code: "ID-11", Name: "Aceh", Type: "PROVINCE", ParentCode: pgtype.Text{String: "ID", Valid: true}})
	if err != nil {
		t.Fatalf("CreateRegion(province) error = %v", err)
	}

	foodRequest := env.bbProjectRequest("BB-002", "Food Estate")
	foodRequest.ExecutingAgencyIDs = []string{model.UUIDToString(agricultureEA.ID)}
	foodRequest.LocationIDs = []string{model.UUIDToString(province.ID)}
	foodProject, err := env.service.CreateBBProject(env.ctx, mustParseUUID(t, blueBook.ID), foodRequest)
	if err != nil {
		t.Fatalf("CreateBBProject(food) error = %v", err)
	}

	searchResult, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, blueBook.ID), model.BBProjectListFilter{}, model.PaginationParams{Page: 1, Limit: 10, Search: "Agriculture"})
	if err != nil {
		t.Fatalf("ListBBProjects(search) error = %v", err)
	}
	assertProjectListIDs(t, searchResult.Data, foodProject.ID)

	eaResult, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, blueBook.ID), model.BBProjectListFilter{ExecutingAgencyIDs: []string{model.UUIDToString(env.ea.ID)}}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListBBProjects(EA filter) error = %v", err)
	}
	assertProjectListIDs(t, eaResult.Data, floodProject.ID)

	locationResult, err := env.service.ListBBProjects(env.ctx, mustParseUUID(t, blueBook.ID), model.BBProjectListFilter{LocationIDs: []string{model.UUIDToString(province.ID)}}, model.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("ListBBProjects(location filter) error = %v", err)
	}
	assertProjectListIDs(t, locationResult.Data, foodProject.ID)
}

func TestGetBBProjectHistoryReturnsOrderedSnapshots(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-001", "Flood Control")
	env.createBlueBook(t, 1, &original.ID)

	history, err := env.service.GetBBProjectHistory(env.ctx, mustParseUUID(t, sourceProject.ID))
	if err != nil {
		t.Fatalf("GetBBProjectHistory() error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("history length = %d, want 2: %+v", len(history), history)
	}
	if history[0].ID != sourceProject.ID {
		t.Fatalf("history[0].ID = %s, want source %s", history[0].ID, sourceProject.ID)
	}
	if history[0].BookStatus != "superseded" || history[0].IsLatest {
		t.Fatalf("history[0] status/latest = %s/%v, want superseded/not latest", history[0].BookStatus, history[0].IsLatest)
	}
	if history[1].BookStatus != "active" || !history[1].IsLatest {
		t.Fatalf("history[1] status/latest = %s/%v, want active/latest", history[1].BookStatus, history[1].IsLatest)
	}
	if history[0].ProjectIdentityID != history[1].ProjectIdentityID {
		t.Fatalf("history identities differ: %s vs %s", history[0].ProjectIdentityID, history[1].ProjectIdentityID)
	}
}

func TestBlueBookImportReusesPreviousIdentityForRevisionSnapshot(t *testing.T) {
	env := setupBlueBookVersioningTest(t)

	original := env.createBlueBook(t, 0, nil)
	sourceProject := env.createBBProject(t, original.ID, "BB-001", "Flood Control")
	if err := env.queries.SupersedeBlueBooksByPeriod(env.ctx, env.period.ID); err != nil {
		t.Fatalf("SupersedeBlueBooksByPeriod() error = %v", err)
	}
	targetRevision, err := env.queries.CreateBlueBook(env.ctx, queries.CreateBlueBookParams{
		PeriodID:           env.period.ID,
		ReplacesBlueBookID: mustParseUUID(t, original.ID),
		PublishDate:        testDate(2026, time.February, 1),
		RevisionNumber:     1,
		RevisionYear:       pgtype.Int4{Int32: 2026, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateBlueBook(target revision) error = %v", err)
	}

	workbook := buildBlueBookRevisionImportWorkbook(t, env, "BB-001", "Flood Control Revision")
	res, err := env.service.ImportBlueBookProjects(env.ctx, targetRevision.ID, "blue-book-projects.xlsx", bytes.NewReader(workbook), int64(len(workbook)))
	if err != nil {
		t.Fatalf("ImportBlueBookProjects() error = %v", err)
	}
	if res.TotalFailed != 0 {
		t.Fatalf("TotalFailed = %d, response = %+v", res.TotalFailed, res)
	}
	inputSheet := findImportSheet(t, res, blueBookImportSheetInput)
	if inputSheet.Inserted != 1 || len(inputSheet.Rows) != 1 {
		t.Fatalf("input sheet inserted/rows = %d/%d, want 1/1: %+v", inputSheet.Inserted, len(inputSheet.Rows), inputSheet)
	}
	if !strings.Contains(inputSheet.Rows[0].Message, "revision snapshot") {
		t.Fatalf("input row message = %q, want revision snapshot message", inputSheet.Rows[0].Message)
	}

	imported, err := env.queries.GetBBProjectByBlueBookAndCode(env.ctx, queries.GetBBProjectByBlueBookAndCodeParams{BlueBookID: targetRevision.ID, Lower: "BB-001"})
	if err != nil {
		t.Fatalf("GetBBProjectByBlueBookAndCode(target) error = %v", err)
	}
	if model.UUIDToString(imported.ProjectIdentityID) != sourceProject.ProjectIdentityID {
		t.Fatalf("imported identity = %s, want %s", model.UUIDToString(imported.ProjectIdentityID), sourceProject.ProjectIdentityID)
	}
}

func TestBlueBookImportRejectsDuplicateBBCodeInWorkbook(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	blueBook := env.createBlueBook(t, 0, nil)
	workbook := simpleXLSXWorkbook{Sheets: []simpleXLSXSheet{
		{
			Name: blueBookImportSheetInput,
			Rows: [][]simpleXLSXCell{
				headerRow("Program Title (*)", "BB Code (*)", "Project Name (*)"),
				textRow(env.programTitle.Title, "BB-DUP", "Duplicate A"),
				textRow(env.programTitle.Title, "BB-DUP", "Duplicate B"),
			},
			Columns:       columns(28, 18, 48),
			ShowGridLines: false,
		},
		emptyImportSheet(blueBookImportSheetEA, "BB Code (*)", "Executing Agency Name (*)"),
		emptyImportSheet(blueBookImportSheetIA, "BB Code (*)", "Implementing Agency Name (*)"),
		emptyImportSheet(blueBookImportSheetLocations, "BB Code (*)", "Location Name (*)"),
		emptyImportSheet(blueBookImportSheetNationalPriority, "BB Code (*)", "National Priority Name (*)"),
		emptyImportSheet(blueBookImportSheetProjectCost, "BB Code (*)", "Funding Type (*)", "Funding Category (*)", "Amount USD"),
		emptyImportSheet(blueBookImportSheetLenderIndication, "BB Code (*)", "Lender Name (*)", "Remarks"),
	}}
	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		t.Fatalf("buildSimpleXLSX() error = %v", err)
	}

	res, err := env.service.PreviewBlueBookProjects(env.ctx, mustParseUUID(t, blueBook.ID), "blue-book-projects.xlsx", bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("PreviewBlueBookProjects() error = %v", err)
	}
	inputSheet := findImportSheet(t, res, blueBookImportSheetInput)
	if inputSheet.Failed == 0 || !importSheetHasMessage(inputSheet, "BB Code duplikat di workbook") {
		t.Fatalf("input sheet did not report duplicate BB Code: %+v", inputSheet)
	}
}

func TestBlueBookImportResolvesScopedInstitutionPath(t *testing.T) {
	env := setupBlueBookVersioningTest(t)
	blueBook := env.createBlueBook(t, 0, nil)
	_, _ = env.createScopedInstitution(t, "Kementerian A", "Sekretariat Utama")
	targetInstitution, targetPath := env.createScopedInstitution(t, "Kementerian B", "Sekretariat Utama")

	ambiguousWorkbook := buildBlueBookImportWorkbookWithInstitutionRefs(t, env, "BB-AMB", "Ambiguous Institution", "Sekretariat Utama", env.ia.Name)
	ambiguousRes, err := env.service.PreviewBlueBookProjects(env.ctx, mustParseUUID(t, blueBook.ID), "blue-book-projects.xlsx", bytes.NewReader(ambiguousWorkbook), int64(len(ambiguousWorkbook)))
	if err != nil {
		t.Fatalf("PreviewBlueBookProjects(ambiguous) error = %v", err)
	}
	eaSheet := findImportSheet(t, ambiguousRes, blueBookImportSheetEA)
	if eaSheet.Failed == 0 || !importSheetHasMessage(eaSheet, "ambigu") {
		t.Fatalf("EA sheet did not report ambiguous institution: %+v", eaSheet)
	}

	workbook := buildBlueBookImportWorkbookWithInstitutionRefs(t, env, "BB-SCOPED", "Scoped Institution", targetPath, env.ia.Name)
	res, err := env.service.ImportBlueBookProjects(env.ctx, mustParseUUID(t, blueBook.ID), "blue-book-projects.xlsx", bytes.NewReader(workbook), int64(len(workbook)))
	if err != nil {
		t.Fatalf("ImportBlueBookProjects(scoped) error = %v", err)
	}
	if res.TotalFailed != 0 {
		t.Fatalf("TotalFailed = %d, response = %+v", res.TotalFailed, res)
	}
	imported, err := env.queries.GetBBProjectByBlueBookAndCode(env.ctx, queries.GetBBProjectByBlueBookAndCodeParams{BlueBookID: mustParseUUID(t, blueBook.ID), Lower: "BB-SCOPED"})
	if err != nil {
		t.Fatalf("GetBBProjectByBlueBookAndCode(scoped) error = %v", err)
	}
	institutions, err := env.queries.GetBBProjectInstitutions(env.ctx, imported.ID)
	if err != nil {
		t.Fatalf("GetBBProjectInstitutions(scoped) error = %v", err)
	}
	if !hasInstitutionRole(institutions, targetInstitution.ID, roleExecutingAgency) {
		t.Fatalf("imported institutions = %+v, want scoped EA %s", institutions, model.UUIDToString(targetInstitution.ID))
	}
}

func setupBlueBookVersioningTest(t *testing.T) *blueBookVersioningTestEnv {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping PostgreSQL integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	databaseURL := blueBookVersioningTestDatabaseURL()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("PostgreSQL test database unavailable: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("PostgreSQL test database unavailable: %v", err)
	}

	schemaName := fmt.Sprintf("prism_test_%d", time.Now().UnixNano())
	if _, err := adminPool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public`); err != nil {
		adminPool.Close()
		t.Fatalf("create uuid extension: %v", err)
	}
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+schemaName); err != nil {
		adminPool.Close()
		t.Fatalf("create test schema: %v", err)
	}

	ddl := readPrismDDL(t)
	conn, err := adminPool.Acquire(ctx)
	if err != nil {
		adminPool.Close()
		t.Fatalf("acquire setup connection: %v", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, "SET search_path TO "+schemaName+", public;\n"+ddl).ReadAll()
	conn.Release()
	if err != nil {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
		adminPool.Close()
		t.Fatalf("apply DDL to test schema: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		adminPool.Close()
		t.Fatalf("parse test database URL: %v", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schemaName + ",public"
	cfg.MaxConns = 4
	cfg.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		adminPool.Close()
		t.Fatalf("create test pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		adminPool.Close()
		t.Fatalf("ping test pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaName+" CASCADE")
		adminPool.Close()
	})

	q := queries.New(pool)
	env := &blueBookVersioningTestEnv{
		ctx:     ctx,
		pool:    pool,
		queries: q,
		service: NewBlueBookService(pool, q, nil),
	}
	env.seedFixtures(t)
	return env
}

func blueBookVersioningTestDatabaseURL() string {
	if value := strings.TrimSpace(os.Getenv("PRISM_TEST_DATABASE_URL")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")); value != "" {
		return value
	}
	return "postgres://prism:prism_secret@localhost:5432/prism_dev?sslmode=disable"
}

func readPrismDDL(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, "docs", "prism_ddl.sql")
		data, err := os.ReadFile(candidate)
		if err == nil {
			return string(data)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("docs/prism_ddl.sql not found from test working directory")
	return ""
}

func (env *blueBookVersioningTestEnv) seedFixtures(t *testing.T) {
	t.Helper()
	var err error
	env.period, err = env.queries.CreatePeriod(env.ctx, queries.CreatePeriodParams{Name: "2025-2029", YearStart: 2025, YearEnd: 2029})
	if err != nil {
		t.Fatalf("CreatePeriod() error = %v", err)
	}
	env.programTitle, err = env.queries.CreateProgramTitle(env.ctx, queries.CreateProgramTitleParams{ParentID: pgtype.UUID{}, Title: "Flood Management"})
	if err != nil {
		t.Fatalf("CreateProgramTitle() error = %v", err)
	}
	env.ea, err = env.queries.CreateInstitution(env.ctx, queries.CreateInstitutionParams{ParentID: pgtype.UUID{}, Name: "Ministry of Works", ShortName: pgtype.Text{String: "MW", Valid: true}, Level: "Kementerian/Badan/Lembaga"})
	if err != nil {
		t.Fatalf("CreateInstitution(EA) error = %v", err)
	}
	env.ia, err = env.queries.CreateInstitution(env.ctx, queries.CreateInstitutionParams{ParentID: pgtype.UUID{}, Name: "Directorate of Water", ShortName: pgtype.Text{String: "DW", Valid: true}, Level: "Eselon I"})
	if err != nil {
		t.Fatalf("CreateInstitution(IA) error = %v", err)
	}
	env.region, err = env.queries.CreateRegion(env.ctx, queries.CreateRegionParams{Code: "ID", Name: "Nasional", Type: "COUNTRY", ParentCode: pgtype.Text{}})
	if err != nil {
		t.Fatalf("CreateRegion() error = %v", err)
	}
}

func (env *blueBookVersioningTestEnv) createBlueBook(t *testing.T, revisionNumber int32, replacesID *string) *model.BlueBookResponse {
	t.Helper()
	var revisionYear *int32
	if revisionNumber > 0 {
		value := int32(2026)
		revisionYear = &value
	}
	res, err := env.service.CreateBlueBook(env.ctx, model.BlueBookRequest{
		PeriodID:           model.UUIDToString(env.period.ID),
		ReplacesBlueBookID: replacesID,
		PublishDate:        fmt.Sprintf("2026-%02d-01", revisionNumber+1),
		RevisionNumber:     revisionNumber,
		RevisionYear:       revisionYear,
	})
	if err != nil {
		t.Fatalf("CreateBlueBook(revision %d) error = %v", revisionNumber, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) createBBProject(t *testing.T, blueBookID, code, name string) *model.BBProjectResponse {
	t.Helper()
	res, err := env.service.CreateBBProject(env.ctx, mustParseUUID(t, blueBookID), env.bbProjectRequest(code, name))
	if err != nil {
		t.Fatalf("CreateBBProject(%s) error = %v", code, err)
	}
	return res
}

func (env *blueBookVersioningTestEnv) createScopedInstitution(t *testing.T, rootName, childName string) (queries.Institution, string) {
	t.Helper()
	root, err := env.queries.CreateInstitution(env.ctx, queries.CreateInstitutionParams{ParentID: pgtype.UUID{}, Name: rootName, ShortName: pgtype.Text{}, Level: "Kementerian/Badan/Lembaga"})
	if err != nil {
		t.Fatalf("CreateInstitution(%s) error = %v", rootName, err)
	}
	child, err := env.queries.CreateInstitution(env.ctx, queries.CreateInstitutionParams{ParentID: root.ID, Name: childName, ShortName: pgtype.Text{}, Level: "Eselon I"})
	if err != nil {
		t.Fatalf("CreateInstitution(%s under %s) error = %v", childName, rootName, err)
	}
	return child, fmt.Sprintf("%s; %s;", childName, rootName)
}

func (env *blueBookVersioningTestEnv) bbProjectRequest(code, name string) model.CreateBBProjectRequest {
	programTitleID := model.UUIDToString(env.programTitle.ID)
	return model.CreateBBProjectRequest{
		ProgramTitleID:        &programTitleID,
		BBCode:                code,
		ProjectName:           name,
		ExecutingAgencyIDs:    []string{model.UUIDToString(env.ea.ID)},
		ImplementingAgencyIDs: []string{model.UUIDToString(env.ia.ID)},
		LocationIDs:           []string{model.UUIDToString(env.region.ID)},
		ProjectCosts: []model.ProjectCostItem{{
			FundingType:     "Foreign",
			FundingCategory: "Loan",
			AmountUSD:       1000000,
		}},
	}
}

func buildBlueBookRevisionImportWorkbook(t *testing.T, env *blueBookVersioningTestEnv, code, projectName string) []byte {
	t.Helper()
	return buildBlueBookImportWorkbookWithInstitutionRefs(t, env, code, projectName, env.ea.Name, env.ia.Name)
}

func buildBlueBookImportWorkbookWithInstitutionRefs(t *testing.T, env *blueBookVersioningTestEnv, code, projectName, executingAgencyRef, implementingAgencyRef string) []byte {
	t.Helper()
	workbook := simpleXLSXWorkbook{Sheets: []simpleXLSXSheet{
		{
			Name: blueBookImportSheetInput,
			Rows: [][]simpleXLSXCell{
				headerRow("Program Title (*)", "BB Code (*)", "Project Name (*)"),
				textRow(env.programTitle.Title, code, projectName),
			},
			Columns:       columns(28, 18, 48),
			ShowGridLines: false,
		},
		{
			Name: blueBookImportSheetEA,
			Rows: [][]simpleXLSXCell{
				headerRow("BB Code (*)", "Executing Agency Name (*)"),
				textRow(code, executingAgencyRef),
			},
			Columns:       columns(18, 44),
			ShowGridLines: false,
		},
		{
			Name: blueBookImportSheetIA,
			Rows: [][]simpleXLSXCell{
				headerRow("BB Code (*)", "Implementing Agency Name (*)"),
				textRow(code, implementingAgencyRef),
			},
			Columns:       columns(18, 44),
			ShowGridLines: false,
		},
		{
			Name: blueBookImportSheetLocations,
			Rows: [][]simpleXLSXCell{
				headerRow("BB Code (*)", "Location Name (*)"),
				textRow(code, env.region.Name),
			},
			Columns:       columns(18, 36),
			ShowGridLines: false,
		},
		emptyImportSheet(blueBookImportSheetNationalPriority, "BB Code (*)", "National Priority Name (*)"),
		emptyImportSheet(blueBookImportSheetProjectCost, "BB Code (*)", "Funding Type (*)", "Funding Category (*)", "Amount USD"),
		emptyImportSheet(blueBookImportSheetLenderIndication, "BB Code (*)", "Lender Name (*)", "Remarks"),
	}}
	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		t.Fatalf("buildSimpleXLSX() error = %v", err)
	}
	return data
}

func importSheetHasMessage(sheet model.MasterImportSheetResult, want string) bool {
	for _, row := range sheet.Rows {
		if strings.Contains(row.Message, want) {
			return true
		}
	}
	for _, row := range sheet.Errors {
		if strings.Contains(row.Message, want) {
			return true
		}
	}
	return false
}

func hasInstitutionRole(rows []queries.GetBBProjectInstitutionsRow, institutionID pgtype.UUID, role string) bool {
	for _, row := range rows {
		if row.Role == role && model.UUIDToString(row.ID) == model.UUIDToString(institutionID) {
			return true
		}
	}
	return false
}

func emptyImportSheet(name string, headers ...string) simpleXLSXSheet {
	return simpleXLSXSheet{
		Name:          name,
		Rows:          [][]simpleXLSXCell{headerRow(headers...)},
		Columns:       columns(22, 36, 30, 18),
		ShowGridLines: false,
	}
}

func findImportSheet(t *testing.T, res *model.MasterImportResponse, name string) model.MasterImportSheetResult {
	t.Helper()
	for _, sheet := range res.Sheets {
		if sheet.Sheet == name {
			return sheet
		}
	}
	t.Fatalf("import sheet %q not found in %+v", name, res.Sheets)
	return model.MasterImportSheetResult{}
}

func assertAppErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError %s, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("AppError code = %s, want %s; message=%q", appErr.Code, code, appErr.Message)
	}
}

func assertAppErrorHasDetailField(t *testing.T, err error, field string) {
	t.Helper()
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError with detail field %s, got %T: %v", field, err, err)
	}
	for _, detail := range appErr.Details {
		if detail.Field == field {
			return
		}
	}
	t.Fatalf("AppError details did not include field %s: %+v", field, appErr.Details)
}

func assertAuditDeleteExists(t *testing.T, env *blueBookVersioningTestEnv, tableName, recordID string) {
	t.Helper()
	var count int64
	if err := env.pool.QueryRow(env.ctx, `SELECT COUNT(*) FROM audit_log WHERE table_name = $1 AND record_id = $2 AND action = 'DELETE'`, tableName, recordID).Scan(&count); err != nil {
		t.Fatalf("count audit log for %s/%s: %v", tableName, recordID, err)
	}
	if count == 0 {
		t.Fatalf("audit log missing DELETE for %s/%s", tableName, recordID)
	}
}

func staffDeleteUser() *model.AuthUser {
	return &model.AuthUser{Role: "STAFF"}
}

func adminDeleteUser() *model.AuthUser {
	return &model.AuthUser{Role: "ADMIN"}
}

func assertProjectListIDs(t *testing.T, projects []model.BBProjectResponse, ids ...string) {
	t.Helper()
	if len(projects) != len(ids) {
		t.Fatalf("project count = %d, want %d: %+v", len(projects), len(ids), projects)
	}
	for index, id := range ids {
		if projects[index].ID != id {
			t.Fatalf("project[%d].ID = %s, want %s", index, projects[index].ID, id)
		}
	}
}

func mustParseUUID(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	parsed, err := model.ParseUUID(value)
	if err != nil {
		t.Fatalf("parse UUID %q: %v", value, err)
	}
	return parsed
}

func testDate(year int, month time.Month, day int) pgtype.Date {
	return pgtype.Date{Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC), Valid: true}
}
