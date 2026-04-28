package service

import (
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

const (
	dkImportSheetHeader          = "Daftar Kegiatan"
	dkImportSheetInput           = "Input Data"
	dkImportSheetGBProject       = "Relasi - GB Project"
	dkImportSheetLocations       = "Relasi - Locations"
	dkImportSheetFinancingDetail = "Relasi - Financing Detail"
	dkImportSheetLoanAllocation  = "Relasi - Loan Allocation"
	dkImportSheetActivityDetail  = "Relasi - Activity Detail"
)

type dkImportHeaderDraft struct {
	row          int
	dkKey        string
	letterNumber string
	subject      string
	date         pgtype.Date
	createdID    pgtype.UUID
	skipExisting bool
	errors       []string
	projects     []*dkImportProjectDraft
}

type dkImportProjectDraft struct {
	row             int
	header          *dkImportHeaderDraft
	dkKey           string
	projectKey      string
	programTitleID  pgtype.UUID
	institutionID   pgtype.UUID
	duration        *string
	objectives      *string
	gbProjectIDs    []string
	gbProjectUUIDs  []pgtype.UUID
	locationIDs     []string
	financing       []model.DKFinancingDetailItem
	loanAllocations []model.DKLoanAllocationItem
	activityDetails []model.DKActivityDetailItem
	activityNumbers map[int32]struct{}
	errors          []string
}

type dkImportRelationRow struct {
	row        int
	sheet      string
	dkKey      string
	projectKey string
	label      string
	draft      *dkImportProjectDraft
	status     string
	message    string
}

type dkAllowedLenderCache map[string]map[string]queries.ListAllowedLenderReferencesByGBProjectRow

func (d *dkImportHeaderDraft) addError(message string) {
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

func (d *dkImportHeaderDraft) failed() bool {
	return d == nil || len(d.errors) > 0
}

func (d *dkImportProjectDraft) addError(message string) {
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

func (d *dkImportProjectDraft) failed() bool {
	return d == nil || len(d.errors) > 0
}

func (d *dkImportProjectDraft) skipped() bool {
	return d != nil && d.header != nil && d.header.skipExisting
}

func (s *DKService) PreviewDaftarKegiatanImport(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processDaftarKegiatanWorkbook(ctx, fileName, reader, size, false)
}

func (s *DKService) ImportDaftarKegiatan(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processDaftarKegiatanWorkbook(ctx, fileName, reader, size, true)
}

func (s *DKService) processDaftarKegiatanWorkbook(ctx context.Context, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
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
	masterSvc := &MasterService{db: s.db, queries: s.queries}
	lookups, err := masterSvc.loadMasterImportLookups(ctx, qtx)
	if err != nil {
		return nil, err
	}

	response, createdIDs, err := s.buildDaftarKegiatanImportPreview(ctx, qtx, workbook, lookups, fileName)
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
		return nil, apperrors.Internal("Gagal menyimpan hasil import Daftar Kegiatan")
	}

	if s.broker != nil {
		for _, id := range createdIDs {
			s.broker.Publish("daftar_kegiatan.created", map[string]string{"id": id})
		}
	}

	return response, nil
}

func (s *DKService) buildDaftarKegiatanImportPreview(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, fileName string) (*model.MasterImportResponse, []string, error) {
	headerResult := model.MasterImportSheetResult{Sheet: dkImportSheetHeader}
	headers, headersByKey, err := s.parseDKHeaderRows(ctx, qtx, workbook, &headerResult)
	if err != nil {
		return nil, nil, err
	}

	inputResult := model.MasterImportSheetResult{Sheet: dkImportSheetInput}
	projects, projectsByKey, err := s.parseDKProjectInputRows(workbook, lookups, headersByKey, &inputResult)
	if err != nil {
		return nil, nil, err
	}

	relationResults := map[string]model.MasterImportSheetResult{
		dkImportSheetGBProject:       {Sheet: dkImportSheetGBProject},
		dkImportSheetLocations:       {Sheet: dkImportSheetLocations},
		dkImportSheetFinancingDetail: {Sheet: dkImportSheetFinancingDetail},
		dkImportSheetLoanAllocation:  {Sheet: dkImportSheetLoanAllocation},
		dkImportSheetActivityDetail:  {Sheet: dkImportSheetActivityDetail},
	}
	relationRows := make([]dkImportRelationRow, 0)
	allowedCache := dkAllowedLenderCache{}

	relationRows = append(relationRows, s.parseDKGBProjectRelation(workbook, lookups, headersByKey, projectsByKey, relationResults)...)
	relationRows = append(relationRows, s.parseDKLocationRelation(workbook, lookups, headersByKey, projectsByKey, relationResults)...)
	financingRows, err := s.parseDKFinancingDetailRelation(ctx, qtx, workbook, lookups, headersByKey, projectsByKey, relationResults, allowedCache)
	if err != nil {
		return nil, nil, err
	}
	relationRows = append(relationRows, financingRows...)
	relationRows = append(relationRows, s.parseDKLoanAllocationRelation(workbook, lookups, headersByKey, projectsByKey, relationResults)...)
	relationRows = append(relationRows, s.parseDKActivityDetailRelation(workbook, headersByKey, projectsByKey, relationResults)...)

	for _, draft := range projects {
		if draft.skipped() || draft.failed() || draft.header == nil || draft.header.failed() {
			continue
		}
		if len(draft.gbProjectIDs) == 0 {
			draft.addError("Minimal 1 GB Project wajib diisi")
		}
		if len(draft.locationIDs) == 0 {
			draft.addError("Location wajib diisi")
		}
		if len(draft.financing) == 0 {
			draft.addError("Financing Detail wajib diisi")
		}
		if len(draft.loanAllocations) == 0 {
			draft.addError("Loan Allocation wajib diisi")
		}
		if len(draft.activityDetails) == 0 {
			draft.addError("Activity Detail wajib diisi")
		}
	}

	createdHeaderIDs := make([]string, 0)
	for _, header := range headers {
		switch {
		case header.skipExisting:
			addImportSkipped(&headerResult, header.row, dkHeaderLabel(header))
		case header.failed():
			addImportError(&headerResult, header.row, strings.Join(header.errors, "; "))
		default:
			created, err := qtx.CreateDaftarKegiatan(ctx, queries.CreateDaftarKegiatanParams{
				LetterNumber: nullableTextPtr(&header.letterNumber),
				Subject:      header.subject,
				Date:         header.date,
			})
			if err != nil {
				return nil, nil, fromPg(err)
			}
			header.createdID = created.ID
			createdHeaderIDs = append(createdHeaderIDs, model.UUIDToString(created.ID))
			addImportCreated(&headerResult, header.row, dkHeaderLabel(header))
		}
	}

	for _, draft := range projects {
		switch {
		case draft.skipped():
			addImportSkipped(&inputResult, draft.row, dkProjectLabel(draft))
		case draft.header == nil:
			addImportError(&inputResult, draft.row, "DK Key tidak ada di sheet Daftar Kegiatan")
		case draft.header.failed():
			addImportError(&inputResult, draft.row, "Header Daftar Kegiatan terkait gagal validasi")
		case draft.failed():
			addImportError(&inputResult, draft.row, strings.Join(draft.errors, "; "))
		default:
			created, err := qtx.CreateDKProject(ctx, queries.CreateDKProjectParams{
				DkID:           draft.header.createdID,
				ProgramTitleID: draft.programTitleID,
				InstitutionID:  draft.institutionID,
				Duration:       nullableTextPtr(draft.duration),
				Objectives:     nullableTextPtr(draft.objectives),
			})
			if err != nil {
				return nil, nil, fromPg(err)
			}
			req := model.CreateDKProjectRequest{
				ProgramTitleID:   stringPtrFromUUID(draft.programTitleID),
				InstitutionID:    stringPtrFromUUID(draft.institutionID),
				Duration:         draft.duration,
				Objectives:       draft.objectives,
				GBProjectIDs:     draft.gbProjectIDs,
				LocationIDs:      draft.locationIDs,
				FinancingDetails: draft.financing,
				LoanAllocations:  draft.loanAllocations,
				ActivityDetails:  draft.activityDetails,
			}
			if err := s.replaceDKProjectChildren(ctx, qtx, created.ID, req); err != nil {
				return nil, nil, err
			}
			addImportCreated(&inputResult, draft.row, dkProjectLabel(draft))
		}
	}

	for _, relation := range relationRows {
		result := relationResults[relation.sheet]
		status, message := dkRelationStatus(relation)
		addImportRow(&result, relation.row, status, relation.label, message)
		switch status {
		case masterImportStatusCreate:
			result.Inserted++
		case masterImportStatusSkip:
			result.Skipped++
		case masterImportStatusFailed:
			result.Failed++
			result.Errors = append(result.Errors, model.MasterImportRowError{Row: relation.row, Message: message})
		}
		relationResults[relation.sheet] = result
	}

	response := &model.MasterImportResponse{
		FileName: fileName,
		Sheets: []model.MasterImportSheetResult{
			headerResult,
			inputResult,
			relationResults[dkImportSheetGBProject],
			relationResults[dkImportSheetLocations],
			relationResults[dkImportSheetFinancingDetail],
			relationResults[dkImportSheetLoanAllocation],
			relationResults[dkImportSheetActivityDetail],
		},
	}
	recalculateImportTotals(response)

	return response, createdHeaderIDs, nil
}

func (s *DKService) parseDKHeaderRows(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, result *model.MasterImportSheetResult) ([]*dkImportHeaderDraft, map[string]*dkImportHeaderDraft, error) {
	rows, ok := workbook.importRows(dkImportSheetHeader, []string{"dk_key", "letter_number", "subject", "date"})
	if !ok {
		addImportError(result, 0, "Sheet Daftar Kegiatan tidak ditemukan")
		return nil, map[string]*dkImportHeaderDraft{}, nil
	}
	if hasImportHeaderError(result, rows) {
		return nil, map[string]*dkImportHeaderDraft{}, nil
	}

	headers := make([]*dkImportHeaderDraft, 0, len(rows))
	headersByKey := make(map[string]*dkImportHeaderDraft, len(rows))
	seenDKKeys := map[string]struct{}{}
	seenLetterNumbers := map[string]struct{}{}

	for _, row := range rows {
		header := &dkImportHeaderDraft{
			row:          row.number,
			dkKey:        row.value("dk_key"),
			letterNumber: row.value("letter_number"),
			subject:      row.value("subject"),
		}
		headers = append(headers, header)

		if header.dkKey == "" {
			header.addError("DK Key wajib diisi")
		} else {
			key := normalizeLookupKey(header.dkKey)
			if _, exists := seenDKKeys[key]; exists {
				header.addError("DK Key duplikat di workbook")
			} else {
				seenDKKeys[key] = struct{}{}
				headersByKey[key] = header
			}
		}

		if header.letterNumber == "" {
			header.addError("Letter Number wajib diisi untuk import")
			continue
		}
		letterKey := normalizeLookupKey(header.letterNumber)
		if _, exists := seenLetterNumbers[letterKey]; exists {
			header.addError("Letter Number duplikat di workbook")
			continue
		}
		seenLetterNumbers[letterKey] = struct{}{}

		existing, err := qtx.GetDaftarKegiatanByLetterNumber(ctx, header.letterNumber)
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa Letter Number Daftar Kegiatan")
		}
		if err == nil && existing.ID.Valid {
			header.skipExisting = true
			continue
		}

		if header.subject == "" {
			header.addError("Subject wajib diisi")
		}
		dateValue := row.value("date")
		if dateValue == "" {
			header.addError("Date wajib diisi")
		} else {
			date, err := parseDKImportDate(dateValue)
			if err != nil {
				header.addError("Date harus berupa tanggal valid")
			} else {
				header.date = date
			}
		}
	}

	return headers, headersByKey, nil
}

func (s *DKService) parseDKProjectInputRows(workbook *xlsxWorkbook, lookups *masterImportLookups, headersByKey map[string]*dkImportHeaderDraft, result *model.MasterImportSheetResult) ([]*dkImportProjectDraft, map[string]*dkImportProjectDraft, error) {
	rows, ok := workbook.importRows(dkImportSheetInput, []string{"dk_key", "project_key", "executing_agency_name"})
	if !ok {
		addImportError(result, 0, "Sheet Input Data tidak ditemukan")
		return nil, map[string]*dkImportProjectDraft{}, nil
	}
	if hasImportHeaderError(result, rows) {
		return nil, map[string]*dkImportProjectDraft{}, nil
	}

	projects := make([]*dkImportProjectDraft, 0, len(rows))
	projectsByKey := make(map[string]*dkImportProjectDraft, len(rows))
	seenProjectKeys := map[string]struct{}{}

	for _, row := range rows {
		draft := &dkImportProjectDraft{
			row:             row.number,
			dkKey:           row.value("dk_key"),
			projectKey:      row.value("project_key"),
			duration:        row.optionalString("duration"),
			objectives:      row.optionalString("objectives"),
			activityNumbers: map[int32]struct{}{},
		}
		projects = append(projects, draft)

		header := headersByKey[normalizeLookupKey(draft.dkKey)]
		draft.header = header
		if header != nil {
			header.projects = append(header.projects, draft)
		}

		if draft.dkKey == "" || draft.projectKey == "" {
			if draft.skipped() {
				continue
			}
			draft.addError("DK Key dan Project Key wajib diisi")
			continue
		}
		if header == nil {
			draft.addError("DK Key tidak ada di sheet Daftar Kegiatan")
			continue
		}
		projectKey := dkProjectKey(draft.dkKey, draft.projectKey)
		if _, exists := seenProjectKeys[projectKey]; exists {
			if !header.skipExisting {
				draft.addError("Project Key duplikat untuk DK Key yang sama")
			}
			continue
		}
		seenProjectKeys[projectKey] = struct{}{}
		projectsByKey[projectKey] = draft

		if header.skipExisting || header.failed() {
			continue
		}

		programTitle := row.value("program_title")
		if programTitle != "" {
			ref, exists := lookups.programTitlesByTitle[normalizeLookupKey(programTitle)]
			if !exists {
				draft.addError(fmt.Sprintf("Program Title %q belum ada di master data", programTitle))
			} else {
				draft.programTitleID = ref.ID
			}
		}

		agencyName := row.value("executing_agency_name")
		if agencyName == "" {
			draft.addError("Executing Agency Name wajib diisi")
		} else if institution, exists := lookups.institutionsByName[normalizeLookupKey(agencyName)]; exists {
			draft.institutionID = institution.ID
		} else {
			draft.addError(fmt.Sprintf("Executing Agency %q belum ada di master institution", agencyName))
		}
	}

	return projects, projectsByKey, nil
}

func (s *DKService) parseDKGBProjectRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, headersByKey map[string]*dkImportHeaderDraft, projectsByKey map[string]*dkImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []dkImportRelationRow {
	result := relationResults[dkImportSheetGBProject]
	rows, ok := workbook.importRows(dkImportSheetGBProject, []string{"dk_key", "project_key", "gb_code"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - GB Project tidak ditemukan")
		relationResults[dkImportSheetGBProject] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[dkImportSheetGBProject] = result
		return nil
	}

	relations := make([]dkImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		dkKey := row.value("dk_key")
		projectKey := row.value("project_key")
		gbCode := row.value("gb_code")
		relation := dkImportRelationRow{row: row.number, sheet: dkImportSheetGBProject, dkKey: dkKey, projectKey: projectKey, label: fmt.Sprintf("%s/%s - %s", dkKey, projectKey, gbCode)}
		draft := dkProjectDraftByKey(projectsByKey, dkKey, projectKey)
		relation.draft = draft
		if shouldSkipDKRelation(headersByKey, draft, dkKey) {
			relation.status = masterImportStatusSkip
			relation.message = "Daftar Kegiatan sudah ada, relasi dilewati"
			relations = append(relations, relation)
			continue
		}
		if dkKey == "" || projectKey == "" || gbCode == "" {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key, Project Key, dan GB Code wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - GB Project baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key/Project Key tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		gbProject, exists := lookups.gbProjectByCode(gbCode)
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("GB Project %q belum ada atau tidak active", gbCode)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := dkProjectKey(dkKey, projectKey) + "|" + model.UUIDToString(gbProject.ID)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.gbProjectIDs = append(draft.gbProjectIDs, model.UUIDToString(gbProject.ID))
		draft.gbProjectUUIDs = append(draft.gbProjectUUIDs, gbProject.ID)
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[dkImportSheetGBProject] = result
	return relations
}

func (s *DKService) parseDKLocationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, headersByKey map[string]*dkImportHeaderDraft, projectsByKey map[string]*dkImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []dkImportRelationRow {
	result := relationResults[dkImportSheetLocations]
	rows, ok := workbook.importRows(dkImportSheetLocations, []string{"dk_key", "project_key", "location_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Locations tidak ditemukan")
		relationResults[dkImportSheetLocations] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[dkImportSheetLocations] = result
		return nil
	}

	relations := make([]dkImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		dkKey := row.value("dk_key")
		projectKey := row.value("project_key")
		locationName := row.value("location_name")
		relation := dkImportRelationRow{row: row.number, sheet: dkImportSheetLocations, dkKey: dkKey, projectKey: projectKey, label: fmt.Sprintf("%s/%s - %s", dkKey, projectKey, locationName)}
		draft := dkProjectDraftByKey(projectsByKey, dkKey, projectKey)
		relation.draft = draft
		if shouldSkipDKRelation(headersByKey, draft, dkKey) {
			relation.status = masterImportStatusSkip
			relation.message = "Daftar Kegiatan sudah ada, relasi dilewati"
			relations = append(relations, relation)
			continue
		}
		if dkKey == "" || projectKey == "" || locationName == "" {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key, Project Key, dan Location Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Locations baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key/Project Key tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		region, exists := lookups.regionByNameOrCode(locationName)
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Location %q belum ada di master region", locationName)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := dkProjectKey(dkKey, projectKey) + "|" + model.UUIDToString(region.ID)
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

	relationResults[dkImportSheetLocations] = result
	return relations
}

func (s *DKService) parseDKFinancingDetailRelation(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, headersByKey map[string]*dkImportHeaderDraft, projectsByKey map[string]*dkImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, allowedCache dkAllowedLenderCache) ([]dkImportRelationRow, error) {
	result := relationResults[dkImportSheetFinancingDetail]
	rows, ok := workbook.importRows(dkImportSheetFinancingDetail, []string{"dk_key", "project_key", "lender_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Financing Detail tidak ditemukan")
		relationResults[dkImportSheetFinancingDetail] = result
		return nil, nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[dkImportSheetFinancingDetail] = result
		return nil, nil
	}

	relations := make([]dkImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		dkKey := row.value("dk_key")
		projectKey := row.value("project_key")
		lenderName := row.value("lender_name")
		relation := dkImportRelationRow{row: row.number, sheet: dkImportSheetFinancingDetail, dkKey: dkKey, projectKey: projectKey, label: fmt.Sprintf("%s/%s - %s", dkKey, projectKey, lenderName)}
		draft := dkProjectDraftByKey(projectsByKey, dkKey, projectKey)
		relation.draft = draft
		if shouldSkipDKRelation(headersByKey, draft, dkKey) {
			relation.status = masterImportStatusSkip
			relation.message = "Daftar Kegiatan sudah ada, relasi dilewati"
			relations = append(relations, relation)
			continue
		}
		if dkKey == "" || projectKey == "" || lenderName == "" {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key, Project Key, dan Lender Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Financing Detail baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key/Project Key tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		lender, exists := lookups.lendersByName[normalizeLookupKey(lenderName)]
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q belum ada di master data", lenderName)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		allowed, err := s.allowedLendersForDKImportProject(ctx, qtx, draft, allowedCache)
		if err != nil {
			return nil, err
		}
		if _, ok := allowed[model.UUIDToString(lender.ID)]; !ok {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q tidak berasal dari GB Project terkait", lenderName)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		currency, err := parseDKImportCurrency(row.value("currency"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = err.Error()
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		item, err := parseDKFinancingDetailItem(row, lender.ID, currency)
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = err.Error()
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := dkProjectKey(dkKey, projectKey) + "|" + dkFinancingFingerprint(item)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.financing = append(draft.financing, item)
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[dkImportSheetFinancingDetail] = result
	return relations, nil
}

func (s *DKService) parseDKLoanAllocationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, headersByKey map[string]*dkImportHeaderDraft, projectsByKey map[string]*dkImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []dkImportRelationRow {
	result := relationResults[dkImportSheetLoanAllocation]
	rows, ok := workbook.importRows(dkImportSheetLoanAllocation, []string{"dk_key", "project_key", "institution_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Loan Allocation tidak ditemukan")
		relationResults[dkImportSheetLoanAllocation] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[dkImportSheetLoanAllocation] = result
		return nil
	}

	relations := make([]dkImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		dkKey := row.value("dk_key")
		projectKey := row.value("project_key")
		institutionName := row.value("institution_name")
		relation := dkImportRelationRow{row: row.number, sheet: dkImportSheetLoanAllocation, dkKey: dkKey, projectKey: projectKey, label: fmt.Sprintf("%s/%s - %s", dkKey, projectKey, institutionName)}
		draft := dkProjectDraftByKey(projectsByKey, dkKey, projectKey)
		relation.draft = draft
		if shouldSkipDKRelation(headersByKey, draft, dkKey) {
			relation.status = masterImportStatusSkip
			relation.message = "Daftar Kegiatan sudah ada, relasi dilewati"
			relations = append(relations, relation)
			continue
		}
		if dkKey == "" || projectKey == "" || institutionName == "" {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key, Project Key, dan Institution Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Loan Allocation baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key/Project Key tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		institution, exists := lookups.institutionsByName[normalizeLookupKey(institutionName)]
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Institution %q belum ada di master data", institutionName)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		currency, err := parseDKImportCurrency(row.value("currency"))
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = err.Error()
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		item, err := parseDKLoanAllocationItem(row, institution.ID, currency)
		if err != nil {
			relation.status = masterImportStatusFailed
			relation.message = err.Error()
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := dkProjectKey(dkKey, projectKey) + "|" + dkLoanAllocationFingerprint(item)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.loanAllocations = append(draft.loanAllocations, item)
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[dkImportSheetLoanAllocation] = result
	return relations
}

func (s *DKService) parseDKActivityDetailRelation(workbook *xlsxWorkbook, headersByKey map[string]*dkImportHeaderDraft, projectsByKey map[string]*dkImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []dkImportRelationRow {
	result := relationResults[dkImportSheetActivityDetail]
	rows, ok := workbook.importRows(dkImportSheetActivityDetail, []string{"dk_key", "project_key", "activity_no", "activity_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Activity Detail tidak ditemukan")
		relationResults[dkImportSheetActivityDetail] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[dkImportSheetActivityDetail] = result
		return nil
	}

	relations := make([]dkImportRelationRow, 0, len(rows))
	for _, row := range rows {
		dkKey := row.value("dk_key")
		projectKey := row.value("project_key")
		activityNoRaw := row.value("activity_no")
		activityName := row.value("activity_name")
		relation := dkImportRelationRow{row: row.number, sheet: dkImportSheetActivityDetail, dkKey: dkKey, projectKey: projectKey, label: fmt.Sprintf("%s/%s - %s", dkKey, projectKey, activityNoRaw)}
		draft := dkProjectDraftByKey(projectsByKey, dkKey, projectKey)
		relation.draft = draft
		if shouldSkipDKRelation(headersByKey, draft, dkKey) {
			relation.status = masterImportStatusSkip
			relation.message = "Daftar Kegiatan sudah ada, relasi dilewati"
			relations = append(relations, relation)
			continue
		}
		if dkKey == "" || projectKey == "" || activityNoRaw == "" || activityName == "" {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key, Project Key, Activity No, dan Activity Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Activity Detail baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "DK Key/Project Key tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		activityNo, err := parseImportInt(activityNoRaw)
		if err != nil || activityNo <= 0 {
			relation.status = masterImportStatusFailed
			relation.message = "Activity No wajib berupa angka lebih dari 0"
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		activityNumber := int32(activityNo)
		if _, exists := draft.activityNumbers[activityNumber]; exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Activity No %d duplikat untuk Project Key %s", activityNo, projectKey)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		draft.activityNumbers[activityNumber] = struct{}{}
		draft.activityDetails = append(draft.activityDetails, model.DKActivityDetailItem{ActivityNumber: activityNumber, ActivityName: activityName})
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[dkImportSheetActivityDetail] = result
	return relations
}

func (s *DKService) allowedLendersForDKImportProject(ctx context.Context, qtx *queries.Queries, draft *dkImportProjectDraft, cache dkAllowedLenderCache) (map[string]queries.ListAllowedLenderReferencesByGBProjectRow, error) {
	allowed := map[string]queries.ListAllowedLenderReferencesByGBProjectRow{}
	if draft == nil {
		return allowed, nil
	}
	for _, gbProjectID := range draft.gbProjectUUIDs {
		gbID := model.UUIDToString(gbProjectID)
		items, exists := cache[gbID]
		if !exists {
			rows, err := qtx.ListAllowedLenderReferencesByGBProject(ctx, gbProjectID)
			if err != nil {
				return nil, apperrors.Internal("Gagal membaca allowed lender GB Project")
			}
			items = map[string]queries.ListAllowedLenderReferencesByGBProjectRow{}
			for _, row := range rows {
				items[model.UUIDToString(row.LenderID)] = row
			}
			cache[gbID] = items
		}
		for lenderID, row := range items {
			allowed[lenderID] = row
		}
	}
	return allowed, nil
}

func parseDKFinancingDetailItem(row importRow, lenderID pgtype.UUID, currency string) (model.DKFinancingDetailItem, error) {
	amountOriginal, err := parseDKImportAmount(row.value("amount_original"), "Amount Original")
	if err != nil {
		return model.DKFinancingDetailItem{}, err
	}
	grantOriginal, err := parseDKImportAmount(row.value("grant_original"), "Grant Original")
	if err != nil {
		return model.DKFinancingDetailItem{}, err
	}
	counterpartOriginal, err := parseDKImportAmount(row.value("counterpart_original"), "Counterpart Original")
	if err != nil {
		return model.DKFinancingDetailItem{}, err
	}
	amountUSD, err := parseDKImportAmount(row.value("amount_usd"), "Amount USD")
	if err != nil {
		return model.DKFinancingDetailItem{}, err
	}
	grantUSD, err := parseDKImportAmount(row.value("grant_usd"), "Grant USD")
	if err != nil {
		return model.DKFinancingDetailItem{}, err
	}
	counterpartUSD, err := parseDKImportAmount(row.value("counterpart_usd"), "Counterpart USD")
	if err != nil {
		return model.DKFinancingDetailItem{}, err
	}
	lenderIDText := model.UUIDToString(lenderID)
	return model.DKFinancingDetailItem{
		LenderID:            &lenderIDText,
		Currency:            currency,
		AmountOriginal:      amountOriginal,
		GrantOriginal:       grantOriginal,
		CounterpartOriginal: counterpartOriginal,
		AmountUSD:           amountUSD,
		GrantUSD:            grantUSD,
		CounterpartUSD:      counterpartUSD,
		Remarks:             row.optionalString("remarks"),
	}, nil
}

func parseDKLoanAllocationItem(row importRow, institutionID pgtype.UUID, currency string) (model.DKLoanAllocationItem, error) {
	amountOriginal, err := parseDKImportAmount(row.value("amount_original"), "Amount Original")
	if err != nil {
		return model.DKLoanAllocationItem{}, err
	}
	grantOriginal, err := parseDKImportAmount(row.value("grant_original"), "Grant Original")
	if err != nil {
		return model.DKLoanAllocationItem{}, err
	}
	counterpartOriginal, err := parseDKImportAmount(row.value("counterpart_original"), "Counterpart Original")
	if err != nil {
		return model.DKLoanAllocationItem{}, err
	}
	amountUSD, err := parseDKImportAmount(row.value("amount_usd"), "Amount USD")
	if err != nil {
		return model.DKLoanAllocationItem{}, err
	}
	grantUSD, err := parseDKImportAmount(row.value("grant_usd"), "Grant USD")
	if err != nil {
		return model.DKLoanAllocationItem{}, err
	}
	counterpartUSD, err := parseDKImportAmount(row.value("counterpart_usd"), "Counterpart USD")
	if err != nil {
		return model.DKLoanAllocationItem{}, err
	}
	institutionIDText := model.UUIDToString(institutionID)
	return model.DKLoanAllocationItem{
		InstitutionID:       &institutionIDText,
		Currency:            currency,
		AmountOriginal:      amountOriginal,
		GrantOriginal:       grantOriginal,
		CounterpartOriginal: counterpartOriginal,
		AmountUSD:           amountUSD,
		GrantUSD:            grantUSD,
		CounterpartUSD:      counterpartUSD,
		Remarks:             row.optionalString("remarks"),
	}, nil
}

func parseDKImportCurrency(value string) (string, error) {
	currency := strings.ToUpper(strings.TrimSpace(value))
	if currency == "" {
		return "USD", nil
	}
	if len(currency) != 3 {
		return "", fmt.Errorf("Currency harus kode 3 huruf")
	}
	for _, char := range currency {
		if char < 'A' || char > 'Z' {
			return "", fmt.Errorf("Currency harus kode 3 huruf")
		}
	}
	return currency, nil
}

func parseDKImportDate(value string) (pgtype.Date, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Date{}, fmt.Errorf("Date wajib diisi")
	}
	if date, err := parseDate(value, "date"); err == nil {
		return date, nil
	}

	serial, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64)
	if err != nil || math.IsNaN(serial) || math.IsInf(serial, 0) {
		return pgtype.Date{}, fmt.Errorf("Date harus berupa tanggal valid")
	}

	day := int(math.Floor(serial))
	if day <= 0 || day == 60 {
		return pgtype.Date{}, fmt.Errorf("Date harus berupa tanggal valid")
	}
	if day > 59 {
		day--
	}

	base := time.Date(1899, time.December, 31, 0, 0, 0, 0, time.UTC)
	return pgtype.Date{Time: base.AddDate(0, 0, day), Valid: true}, nil
}

func parseDKImportAmount(value, label string) (float64, error) {
	amount, err := parseImportFloat(value)
	if err != nil {
		return 0, fmt.Errorf("%s wajib berupa angka", label)
	}
	if amount < 0 {
		return 0, fmt.Errorf("%s tidak boleh negatif", label)
	}
	return amount, nil
}

func dkRelationStatus(relation dkImportRelationRow) (string, string) {
	if relation.status == masterImportStatusFailed || relation.status == masterImportStatusSkip {
		return relation.status, relation.message
	}
	if relation.draft == nil {
		return masterImportStatusFailed, "DK Key/Project Key tidak ada di sheet Input Data"
	}
	if relation.draft.skipped() {
		return masterImportStatusSkip, "Daftar Kegiatan sudah ada, relasi dilewati"
	}
	if relation.draft.header == nil || relation.draft.header.failed() {
		return masterImportStatusFailed, "Header Daftar Kegiatan terkait gagal validasi"
	}
	if relation.draft.failed() {
		return masterImportStatusFailed, "Project terkait gagal validasi"
	}
	return masterImportStatusCreate, ""
}

func dkHeaderLabel(header *dkImportHeaderDraft) string {
	if header == nil {
		return "Daftar Kegiatan"
	}
	if header.subject == "" {
		return header.letterNumber
	}
	return fmt.Sprintf("%s - %s", header.letterNumber, header.subject)
}

func dkProjectLabel(draft *dkImportProjectDraft) string {
	if draft == nil {
		return "DK Project"
	}
	if draft.dkKey == "" && draft.projectKey == "" {
		return "DK Project"
	}
	if draft.dkKey == "" {
		return draft.projectKey
	}
	if draft.projectKey == "" {
		return draft.dkKey
	}
	return fmt.Sprintf("%s/%s", draft.dkKey, draft.projectKey)
}

func dkProjectKey(dkKey, projectKey string) string {
	return normalizeLookupKey(dkKey) + "|" + normalizeLookupKey(projectKey)
}

func dkProjectDraftByKey(projectsByKey map[string]*dkImportProjectDraft, dkKey, projectKey string) *dkImportProjectDraft {
	if strings.TrimSpace(dkKey) == "" || strings.TrimSpace(projectKey) == "" {
		return nil
	}
	return projectsByKey[dkProjectKey(dkKey, projectKey)]
}

func shouldSkipDKRelation(headersByKey map[string]*dkImportHeaderDraft, draft *dkImportProjectDraft, dkKey string) bool {
	if draft != nil {
		return draft.skipped()
	}
	header := headersByKey[normalizeLookupKey(dkKey)]
	return header != nil && header.skipExisting
}

func dkFinancingFingerprint(item model.DKFinancingDetailItem) string {
	lenderID := ""
	if item.LenderID != nil {
		lenderID = *item.LenderID
	}
	return strings.Join([]string{
		lenderID,
		item.Currency,
		strconv.FormatFloat(item.AmountOriginal, 'f', 2, 64),
		strconv.FormatFloat(item.GrantOriginal, 'f', 2, 64),
		strconv.FormatFloat(item.CounterpartOriginal, 'f', 2, 64),
		strconv.FormatFloat(item.AmountUSD, 'f', 2, 64),
		strconv.FormatFloat(item.GrantUSD, 'f', 2, 64),
		strconv.FormatFloat(item.CounterpartUSD, 'f', 2, 64),
		optionalValue(item.Remarks),
	}, "|")
}

func dkLoanAllocationFingerprint(item model.DKLoanAllocationItem) string {
	institutionID := ""
	if item.InstitutionID != nil {
		institutionID = *item.InstitutionID
	}
	return strings.Join([]string{
		institutionID,
		item.Currency,
		strconv.FormatFloat(item.AmountOriginal, 'f', 2, 64),
		strconv.FormatFloat(item.GrantOriginal, 'f', 2, 64),
		strconv.FormatFloat(item.CounterpartOriginal, 'f', 2, 64),
		strconv.FormatFloat(item.AmountUSD, 'f', 2, 64),
		strconv.FormatFloat(item.GrantUSD, 'f', 2, 64),
		strconv.FormatFloat(item.CounterpartUSD, 'f', 2, 64),
		optionalValue(item.Remarks),
	}, "|")
}

func optionalValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
