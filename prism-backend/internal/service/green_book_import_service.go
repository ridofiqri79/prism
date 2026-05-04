package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

const (
	greenBookImportSheetInput             = "Input Data"
	greenBookImportSheetBBProject         = "Relasi - BB Project"
	greenBookImportSheetEA                = "Relasi - EA"
	greenBookImportSheetIA                = "Relasi - IA"
	greenBookImportSheetLocations         = "Relasi - Locations"
	greenBookImportSheetActivities        = "Relasi - Activities"
	greenBookImportSheetFundingSource     = "Relasi - Funding Source"
	greenBookImportSheetDisbursementPlan  = "Relasi - Disbursement Plan"
	greenBookImportSheetFundingAllocation = "Relasi - Funding Allocation"
)

type greenBookImportProjectDraft struct {
	row                   int
	gbCode                string
	projectName           string
	programTitleID        pgtype.UUID
	duration              *int32
	objective             *string
	scopeOfProject        *string
	gbProjectIdentityID   pgtype.UUID
	bbProjectIDs          []string
	executingAgencyIDs    []string
	implementingAgencyIDs []string
	locationIDs           []string
	activities            []model.GBActivityItem
	activityIndexByNo     map[string]int
	fundingSources        []model.GBFundingSourceItem
	disbursementPlan      []model.GBDisbursementPlanItem
	disbursementYears     map[int32]struct{}
	allocationByActivity  map[int]model.GBFundingAllocationItem
	skipExisting          bool
	errors                []string
}

type greenBookImportRelationRow struct {
	row     int
	sheet   string
	code    string
	label   string
	draft   *greenBookImportProjectDraft
	status  string
	message string
}

func (d *greenBookImportProjectDraft) addError(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	for _, existing := range d.errors {
		if existing == message {
			return
		}
	}
	d.errors = append(d.errors, message)
}

func (d *greenBookImportProjectDraft) failed() bool {
	return len(d.errors) > 0
}

func (d *greenBookImportProjectDraft) fundingAllocations() []model.GBFundingAllocationItem {
	allocations := make([]model.GBFundingAllocationItem, 0, len(d.activities))
	for index := range d.activities {
		item := model.GBFundingAllocationItem{ActivityIndex: index}
		if existing, ok := d.allocationByActivity[index]; ok {
			item = existing
			item.ActivityIndex = index
		}
		allocations = append(allocations, item)
	}
	return allocations
}

func (s *GreenBookService) PreviewGreenBookProjects(ctx context.Context, gbID pgtype.UUID, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processGreenBookProjectsWorkbook(ctx, gbID, fileName, reader, size, false)
}

func (s *GreenBookService) ImportGreenBookProjects(ctx context.Context, gbID pgtype.UUID, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processGreenBookProjectsWorkbook(ctx, gbID, fileName, reader, size, true)
}

func (s *GreenBookService) processGreenBookProjectsWorkbook(ctx context.Context, gbID pgtype.UUID, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
	if !strings.HasSuffix(strings.ToLower(fileName), ".xlsx") {
		return nil, validation("file", "file harus berformat .xlsx")
	}
	if size > maxMasterImportFileSize {
		return nil, validation("file", "ukuran file maksimal 20 MB")
	}

	data, err := io.ReadAll(io.LimitReader(reader, maxMasterImportFileSize+1))
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca file import")
	}
	if len(data) == 0 {
		return nil, validation("file", "file kosong")
	}
	if len(data) > maxMasterImportFileSize {
		return nil, validation("file", "ukuran file maksimal 20 MB")
	}

	workbook, err := readXLSXWorkbook(data)
	if err != nil {
		return nil, validation("file", "format workbook tidak valid")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal memulai transaksi import")
	}
	defer tx.Rollback(ctx)

	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return nil, apperrors.Internal("Gagal menyiapkan audit user")
	}

	qtx := s.queries.WithTx(tx)
	greenBook, err := qtx.GetGreenBook(ctx, gbID)
	if err != nil {
		return nil, mapNotFound(err, "Green Book tidak ditemukan")
	}

	masterSvc := &MasterService{db: s.db, queries: s.queries}
	lookups, err := masterSvc.loadMasterImportLookups(ctx, qtx)
	if err != nil {
		return nil, err
	}

	response, createdIDs, err := s.buildGreenBookImportPreview(ctx, qtx, workbook, lookups, greenBook, fileName)
	if err != nil {
		return nil, err
	}

	if !shouldCommit {
		return response, nil
	}
	if response.TotalFailed > 0 {
		return nil, validation("file", "Perbaiki error preview sebelum eksekusi import")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.Internal("Gagal menyimpan hasil import Green Book")
	}

	if s.broker != nil {
		for _, id := range createdIDs {
			s.broker.Publish("gb_project.created", map[string]string{"id": id})
		}
	}

	return response, nil
}

func (s *GreenBookService) buildGreenBookImportPreview(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, greenBook queries.GetGreenBookRow, fileName string) (*model.MasterImportResponse, []string, error) {
	inputResult := model.MasterImportSheetResult{Sheet: greenBookImportSheetInput}
	projects, projectsByCode, err := s.parseGreenBookInputRows(ctx, qtx, workbook, lookups, greenBook, &inputResult)
	if err != nil {
		return nil, nil, err
	}

	relationResults := map[string]model.MasterImportSheetResult{
		greenBookImportSheetBBProject:         {Sheet: greenBookImportSheetBBProject},
		greenBookImportSheetEA:                {Sheet: greenBookImportSheetEA},
		greenBookImportSheetIA:                {Sheet: greenBookImportSheetIA},
		greenBookImportSheetLocations:         {Sheet: greenBookImportSheetLocations},
		greenBookImportSheetActivities:        {Sheet: greenBookImportSheetActivities},
		greenBookImportSheetFundingSource:     {Sheet: greenBookImportSheetFundingSource},
		greenBookImportSheetDisbursementPlan:  {Sheet: greenBookImportSheetDisbursementPlan},
		greenBookImportSheetFundingAllocation: {Sheet: greenBookImportSheetFundingAllocation},
	}
	relationRows := make([]greenBookImportRelationRow, 0)

	relationRows = append(relationRows, s.parseGreenBookBBProjectRelation(workbook, lookups, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseGreenBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, greenBookImportSheetEA, "executing_agency_name", roleExecutingAgency)...)
	relationRows = append(relationRows, s.parseGreenBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, greenBookImportSheetIA, "implementing_agency_name", roleImplementingAgency)...)
	relationRows = append(relationRows, s.parseGreenBookLocationRelation(workbook, lookups, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseGreenBookActivitiesRelation(workbook, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseGreenBookFundingSourceRelation(workbook, lookups, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseGreenBookDisbursementPlanRelation(workbook, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseGreenBookFundingAllocationRelation(workbook, projectsByCode, relationResults)...)

	for _, draft := range projects {
		if draft.skipExisting || draft.failed() {
			continue
		}
		if len(draft.bbProjectIDs) == 0 {
			draft.addError("Minimal 1 BB Project wajib diisi")
		}
		if len(draft.executingAgencyIDs) == 0 {
			draft.addError("Executing Agency wajib diisi")
		}
		if len(draft.implementingAgencyIDs) == 0 {
			draft.addError("Implementing Agency wajib diisi")
		}
		if len(draft.locationIDs) == 0 {
			draft.addError("Location wajib diisi")
		}
	}

	createdIDs := make([]string, 0)
	for _, draft := range projects {
		switch {
		case draft.skipExisting:
			addImportSkipped(&inputResult, draft.row, fmt.Sprintf("%s - %s", draft.gbCode, draft.projectName))
		case draft.failed():
			addImportError(&inputResult, draft.row, strings.Join(draft.errors, "; "))
		default:
			isRevisionSnapshot := draft.gbProjectIdentityID.Valid
			identityID := draft.gbProjectIdentityID
			if !identityID.Valid {
				identity, err := qtx.CreateGBProjectIdentity(ctx)
				if err != nil {
					return nil, nil, fromPg(err)
				}
				identityID = identity.ID
			}
			created, err := qtx.CreateGBProject(ctx, queries.CreateGBProjectParams{
				GreenBookID:         greenBook.ID,
				GbProjectIdentityID: identityID,
				ProgramTitleID:      draft.programTitleID,
				GbCode:              draft.gbCode,
				ProjectName:         draft.projectName,
				Duration:            int4Ptr(draft.duration),
				Objective:           nullableTextPtr(draft.objective),
				ScopeOfProject:      nullableTextPtr(draft.scopeOfProject),
			})
			if err != nil {
				return nil, nil, fromPg(err)
			}
			req := model.CreateGBProjectRequest{
				BBProjectIDs:          draft.bbProjectIDs,
				ExecutingAgencyIDs:    draft.executingAgencyIDs,
				ImplementingAgencyIDs: draft.implementingAgencyIDs,
				LocationIDs:           draft.locationIDs,
				Activities:            draft.activities,
				FundingSources:        draft.fundingSources,
				DisbursementPlan:      draft.disbursementPlan,
				FundingAllocations:    draft.fundingAllocations(),
			}
			if err := s.replaceGBProjectChildren(ctx, qtx, created.ID, req); err != nil {
				return nil, nil, err
			}
			createdIDs = append(createdIDs, model.UUIDToString(created.ID))
			message := "Created new logical GB Project"
			if isRevisionSnapshot {
				message = "Created revision snapshot for existing logical GB Project"
			}
			addImportCreatedWithMessage(&inputResult, draft.row, fmt.Sprintf("%s - %s", draft.gbCode, draft.projectName), message)
		}
	}

	for _, relation := range relationRows {
		result := relationResults[relation.sheet]
		status, message := greenBookRelationStatus(relation)
		addImportRow(&result, relation.row, status, relation.label, message)
		if status == masterImportStatusCreate {
			result.Inserted++
		}
		if status == masterImportStatusSkip {
			result.Skipped++
		}
		if status == masterImportStatusFailed {
			result.Failed++
			result.Errors = append(result.Errors, model.MasterImportRowError{Row: relation.row, Message: message})
		}
		relationResults[relation.sheet] = result
	}

	response := &model.MasterImportResponse{
		FileName: fileName,
		Sheets: []model.MasterImportSheetResult{
			inputResult,
			relationResults[greenBookImportSheetBBProject],
			relationResults[greenBookImportSheetEA],
			relationResults[greenBookImportSheetIA],
			relationResults[greenBookImportSheetLocations],
			relationResults[greenBookImportSheetActivities],
			relationResults[greenBookImportSheetFundingSource],
			relationResults[greenBookImportSheetDisbursementPlan],
			relationResults[greenBookImportSheetFundingAllocation],
		},
	}
	recalculateImportTotals(response)

	return response, createdIDs, nil
}

func (s *GreenBookService) parseGreenBookInputRows(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, greenBook queries.GetGreenBookRow, result *model.MasterImportSheetResult) ([]*greenBookImportProjectDraft, map[string]*greenBookImportProjectDraft, error) {
	rows, ok := workbook.importRows(greenBookImportSheetInput, []string{"program_title", "gb_code", "project_name"})
	if !ok {
		addImportError(result, 0, "Sheet Input Data tidak ditemukan")
		return nil, map[string]*greenBookImportProjectDraft{}, nil
	}
	if hasImportHeaderError(result, rows) {
		return nil, map[string]*greenBookImportProjectDraft{}, nil
	}

	projects := make([]*greenBookImportProjectDraft, 0, len(rows))
	projectsByCode := make(map[string]*greenBookImportProjectDraft, len(rows))
	seenCodes := map[string]struct{}{}

	for _, row := range rows {
		draft := &greenBookImportProjectDraft{
			row:                  row.number,
			gbCode:               strings.TrimSpace(row.value("gb_code")),
			projectName:          strings.TrimSpace(row.value("project_name")),
			objective:            row.optionalString("objective"),
			scopeOfProject:       row.optionalString("scope_of_project"),
			activityIndexByNo:    map[string]int{},
			disbursementYears:    map[int32]struct{}{},
			allocationByActivity: map[int]model.GBFundingAllocationItem{},
		}
		duration, err := parseImportOptionalPositiveInt32(row.value("duration"))
		if err != nil {
			draft.addError("Duration harus berupa jumlah bulan positif")
		}
		draft.duration = duration
		projects = append(projects, draft)

		if draft.gbCode == "" {
			draft.addError("GB Code wajib diisi")
			continue
		}
		codeKey := normalizeLookupKey(draft.gbCode)
		if _, exists := seenCodes[codeKey]; exists {
			draft.addError("GB Code duplikat di workbook")
			continue
		}
		seenCodes[codeKey] = struct{}{}
		projectsByCode[codeKey] = draft

		existing, err := qtx.GetGBProjectByGreenBookAndCode(ctx, queries.GetGBProjectByGreenBookAndCodeParams{GreenBookID: greenBook.ID, Lower: draft.gbCode})
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa GB Code")
		}
		if err == nil && existing.ID.Valid {
			draft.skipExisting = true
			continue
		}
		previous, err := qtx.FindPreviousGBProjectByCodeForGreenBook(ctx, queries.FindPreviousGBProjectByCodeForGreenBookParams{ID: greenBook.ID, Lower: draft.gbCode})
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa histori GB Code")
		}
		if err == nil && previous.ID.Valid {
			draft.gbProjectIdentityID = previous.GbProjectIdentityID
		}

		if draft.projectName == "" {
			draft.addError("Project Name wajib diisi")
		}

		programTitle := row.value("program_title")
		if programTitle == "" {
			draft.addError("Program Title wajib diisi")
		} else if ref, exists := lookups.programTitlesByTitle[normalizeLookupKey(programTitle)]; exists {
			draft.programTitleID = ref.ID
		} else {
			draft.addError(fmt.Sprintf("Program Title %q belum ada di master data", programTitle))
		}
	}

	return projects, projectsByCode, nil
}

func (s *GreenBookService) parseGreenBookBBProjectRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []greenBookImportRelationRow {
	result := relationResults[greenBookImportSheetBBProject]
	rows, ok := workbook.importRows(greenBookImportSheetBBProject, []string{"gb_code", "bb_code"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - BB Project tidak ditemukan")
		relationResults[greenBookImportSheetBBProject] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[greenBookImportSheetBBProject] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("gb_code")
		bbCode := row.value("bb_code")
		relation := greenBookImportRelationRow{row: row.number, sheet: greenBookImportSheetBBProject, code: code, label: fmt.Sprintf("%s - %s", code, bbCode)}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || bbCode == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code dan BB Code wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - BB Project baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		bbProject, exists := lookups.bbProjectByCode(bbCode)
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("BB Project %q belum ada atau tidak active", bbCode)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + model.UUIDToString(bbProject.ID)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.bbProjectIDs = append(draft.bbProjectIDs, model.UUIDToString(bbProject.ID))
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[greenBookImportSheetBBProject] = result
	return relations
}

func (s *GreenBookService) parseGreenBookInstitutionRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, sheetName, nameHeader, role string) []greenBookImportRelationRow {
	result := relationResults[sheetName]
	rows, ok := workbook.importRows(sheetName, []string{"gb_code", nameHeader})
	if !ok {
		addImportError(&result, 0, fmt.Sprintf("Sheet %s tidak ditemukan", sheetName))
		relationResults[sheetName] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[sheetName] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("gb_code")
		name := row.value(nameHeader)
		label := fmt.Sprintf("%s - %s", code, name)
		relation := greenBookImportRelationRow{row: row.number, sheet: sheetName, code: code, label: label}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code dan nama institution wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("%s baris %d tidak lengkap", sheetName, row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		institution, exists, ambiguous := lookups.lookupInstitutionReference(name)
		if ambiguous {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Institution %q ambigu karena ada lebih dari satu institution dengan nama sama", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Institution %q belum ada di master data", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + model.UUIDToString(institution.ID) + "|" + role
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		if role == roleExecutingAgency {
			draft.executingAgencyIDs = append(draft.executingAgencyIDs, model.UUIDToString(institution.ID))
		} else {
			draft.implementingAgencyIDs = append(draft.implementingAgencyIDs, model.UUIDToString(institution.ID))
		}
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[sheetName] = result
	return relations
}

func (s *GreenBookService) parseGreenBookLocationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []greenBookImportRelationRow {
	result := relationResults[greenBookImportSheetLocations]
	rows, ok := workbook.importRows(greenBookImportSheetLocations, []string{"gb_code", "location_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Locations tidak ditemukan")
		relationResults[greenBookImportSheetLocations] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[greenBookImportSheetLocations] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("gb_code")
		name := row.value("location_name")
		relation := greenBookImportRelationRow{row: row.number, sheet: greenBookImportSheetLocations, code: code, label: fmt.Sprintf("%s - %s", code, name)}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code dan Location Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Locations baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		region, exists := lookups.regionByNameOrCode(name)
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Location %q belum ada di master region", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + model.UUIDToString(region.ID)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.locationIDs = append(draft.locationIDs, model.UUIDToString(region.ID))
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[greenBookImportSheetLocations] = result
	return relations
}

func (s *GreenBookService) parseGreenBookActivitiesRelation(workbook *xlsxWorkbook, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []greenBookImportRelationRow {
	result := relationResults[greenBookImportSheetActivities]
	rows, ok := workbook.importRows(greenBookImportSheetActivities, []string{"gb_code", "activity_no", "activity_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Activities tidak ditemukan")
		relationResults[greenBookImportSheetActivities] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[greenBookImportSheetActivities] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	for _, row := range rows {
		code := row.value("gb_code")
		activityNo := row.value("activity_no")
		activityName := row.value("activity_name")
		label := fmt.Sprintf("%s - %s - %s", code, activityNo, activityName)
		relation := greenBookImportRelationRow{row: row.number, sheet: greenBookImportSheetActivities, code: code, label: label}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || activityNo == "" || activityName == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code, Activity No, dan Activity Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Activities baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		activityKey := greenBookActivityKey(activityNo)
		if _, exists := draft.activityIndexByNo[activityKey]; exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Activity No %q duplikat untuk GB Code %s", activityNo, code)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		sortOrder := int32(len(draft.activities))
		if rawSort := row.value("sort_order"); rawSort != "" {
			parsed, err := parseImportInt(rawSort)
			if err != nil {
				relation.status = masterImportStatusFailed
				relation.message = "Sort Order wajib berupa angka"
				draft.addError(relation.message)
				relations = append(relations, relation)
				continue
			}
			sortOrder = int32(parsed)
		}
		draft.activityIndexByNo[activityKey] = len(draft.activities)
		draft.activities = append(draft.activities, model.GBActivityItem{
			ActivityName:           activityName,
			ImplementationLocation: row.optionalString("implementation_location"),
			PIU:                    row.optionalString("piu"),
			SortOrder:              &sortOrder,
		})
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[greenBookImportSheetActivities] = result
	return relations
}

func (s *GreenBookService) parseGreenBookFundingSourceRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []greenBookImportRelationRow {
	result := relationResults[greenBookImportSheetFundingSource]
	rows, ok := workbook.importRows(greenBookImportSheetFundingSource, []string{"gb_code", "lender_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Funding Source tidak ditemukan")
		relationResults[greenBookImportSheetFundingSource] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[greenBookImportSheetFundingSource] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("gb_code")
		lenderName := row.value("lender_name")
		institutionName := row.value("institution_name")
		relation := greenBookImportRelationRow{row: row.number, sheet: greenBookImportSheetFundingSource, code: code, label: fmt.Sprintf("%s - %s", code, lenderName)}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || lenderName == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code dan Lender Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Funding Source baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		lender, exists, ambiguous := lookups.lookupLenderReference(lenderName)
		if ambiguous {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q ambigu karena cocok dengan lebih dari satu short_name di master data", lenderName)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q belum ada di master data", lenderName)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		var institutionID *string
		institutionKey := ""
		if institutionName != "" {
			institution, exists, ambiguous := lookups.lookupInstitutionReference(institutionName)
			if ambiguous {
				relation.status = masterImportStatusFailed
				relation.message = fmt.Sprintf("Institution %q ambigu karena ada lebih dari satu institution dengan nama sama", institutionName)
				draft.addError(relation.message)
				relations = append(relations, relation)
				continue
			}
			if !exists {
				relation.status = masterImportStatusFailed
				relation.message = fmt.Sprintf("Institution %q belum ada di master data", institutionName)
				draft.addError(relation.message)
				relations = append(relations, relation)
				continue
			}
			id := model.UUIDToString(institution.ID)
			institutionID = &id
			institutionKey = id
		}
		currency, err := parseDKImportCurrency(row.value("currency"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = err.Error()
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + model.UUIDToString(lender.ID) + "|" + institutionKey + "|" + currency
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		loanUSD, err := parseImportFloat(row.value("loan_usd"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Loan USD wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		loanOriginal, err := parseImportFloat(row.value("loan_original"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Loan Original wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		if currency == "USD" && row.value("loan_original") == "" {
			loanOriginal = loanUSD
		}
		grantUSD, err := parseImportFloat(row.value("grant_usd"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Grant USD wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		grantOriginal, err := parseImportFloat(row.value("grant_original"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Grant Original wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		if currency == "USD" && row.value("grant_original") == "" {
			grantOriginal = grantUSD
		}
		localUSD, err := parseImportFloat(row.value("local_usd"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Local USD wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		localOriginal, err := parseImportFloat(row.value("local_original"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Local Original wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		if currency == "USD" && row.value("local_original") == "" {
			localOriginal = localUSD
		}
		seen[key] = struct{}{}
		draft.fundingSources = append(draft.fundingSources, model.GBFundingSourceItem{
			LenderID:      model.UUIDToString(lender.ID),
			InstitutionID: institutionID,
			Currency:      currency,
			LoanOriginal:  loanOriginal,
			GrantOriginal: grantOriginal,
			LocalOriginal: localOriginal,
			LoanUSD:       loanUSD,
			GrantUSD:      grantUSD,
			LocalUSD:      localUSD,
		})
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[greenBookImportSheetFundingSource] = result
	return relations
}

func (s *GreenBookService) parseGreenBookDisbursementPlanRelation(workbook *xlsxWorkbook, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []greenBookImportRelationRow {
	result := relationResults[greenBookImportSheetDisbursementPlan]
	rows, ok := workbook.importRows(greenBookImportSheetDisbursementPlan, []string{"gb_code", "year"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Disbursement Plan tidak ditemukan")
		relationResults[greenBookImportSheetDisbursementPlan] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[greenBookImportSheetDisbursementPlan] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	for _, row := range rows {
		code := row.value("gb_code")
		yearRaw := row.value("year")
		relation := greenBookImportRelationRow{row: row.number, sheet: greenBookImportSheetDisbursementPlan, code: code, label: fmt.Sprintf("%s - %s", code, yearRaw)}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || yearRaw == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code dan Year wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Disbursement Plan baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		year, err := parseImportInt(yearRaw)
		if err != nil || year <= 0 {
			relation.status = masterImportStatusFailed
			relation.message = "Year wajib berupa angka tahun"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		yearValue := int32(year)
		if _, exists := draft.disbursementYears[yearValue]; exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Tahun %d duplikat di disbursement plan", year)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		amount, err := parseImportFloat(row.value("amount_usd"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Amount USD wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		draft.disbursementYears[yearValue] = struct{}{}
		draft.disbursementPlan = append(draft.disbursementPlan, model.GBDisbursementPlanItem{Year: yearValue, AmountUSD: amount})
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[greenBookImportSheetDisbursementPlan] = result
	return relations
}

func (s *GreenBookService) parseGreenBookFundingAllocationRelation(workbook *xlsxWorkbook, projectsByCode map[string]*greenBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []greenBookImportRelationRow {
	result := relationResults[greenBookImportSheetFundingAllocation]
	rows, ok := workbook.importRows(greenBookImportSheetFundingAllocation, []string{"gb_code", "activity_no"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Funding Allocation tidak ditemukan")
		relationResults[greenBookImportSheetFundingAllocation] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[greenBookImportSheetFundingAllocation] = result
		return nil
	}

	relations := make([]greenBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("gb_code")
		activityNo := row.value("activity_no")
		relation := greenBookImportRelationRow{row: row.number, sheet: greenBookImportSheetFundingAllocation, code: code, label: fmt.Sprintf("%s - Activity %s", code, activityNo)}
		draft := greenBookDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || activityNo == "" {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code dan Activity No wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Funding Allocation baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "GB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		activityIndex, exists := draft.activityIndexByNo[greenBookActivityKey(activityNo)]
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Activity No %q tidak ada di sheet Relasi - Activities", activityNo)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + strconv.Itoa(activityIndex)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		services, err := parseImportFloat(row.value("services"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Services wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		constructions, err := parseImportFloat(row.value("constructions"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Constructions wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		goods, err := parseImportFloat(row.value("goods"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Goods wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		trainings, err := parseImportFloat(row.value("trainings"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Trainings wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		other, err := parseImportFloat(row.value("other"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = "Other wajib berupa angka"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.allocationByActivity[activityIndex] = model.GBFundingAllocationItem{
			ActivityIndex: activityIndex,
			Services:      services,
			Constructions: constructions,
			Goods:         goods,
			Trainings:     trainings,
			Other:         other,
		}
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[greenBookImportSheetFundingAllocation] = result
	return relations
}

func greenBookDraftByCode(projectsByCode map[string]*greenBookImportProjectDraft, code string) *greenBookImportProjectDraft {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	return projectsByCode[normalizeLookupKey(code)]
}

func greenBookRelationStatus(relation greenBookImportRelationRow) (string, string) {
	if relation.status == masterImportStatusFailed || relation.status == masterImportStatusSkip {
		return relation.status, relation.message
	}
	if relation.draft == nil {
		return masterImportStatusFailed, "GB Code tidak ada di sheet Input Data"
	}
	if relation.draft.skipExisting {
		return masterImportStatusSkip, "Project sudah ada, relasi dilewati"
	}
	if relation.draft.failed() {
		return masterImportStatusFailed, "Project terkait gagal validasi"
	}
	return masterImportStatusCreate, ""
}

func greenBookActivityKey(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := parseImportInt(value); err == nil {
		return strconv.Itoa(parsed)
	}
	return normalizeLookupKey(value)
}
