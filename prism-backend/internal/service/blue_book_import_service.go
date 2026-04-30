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
	blueBookImportSheetInput            = "Input Data"
	blueBookImportSheetEA               = "Relasi - EA"
	blueBookImportSheetIA               = "Relasi - IA"
	blueBookImportSheetLocations        = "Relasi - Locations"
	blueBookImportSheetNationalPriority = "Relasi - National Priority"
	blueBookImportSheetProjectCost      = "Relasi - Project Cost"
	blueBookImportSheetLenderIndication = "Relasi - Lender Indication"
)

type blueBookImportProjectDraft struct {
	row                   int
	bbCode                string
	projectName           string
	programTitleID        pgtype.UUID
	bappenasPartnerIDs    []string
	duration              *int32
	objective             *string
	scopeOfWork           *string
	outputs               *string
	outcomes              *string
	projectIdentityID     pgtype.UUID
	executingAgencyIDs    []string
	implementingAgencyIDs []string
	locationIDs           []string
	nationalPriorityIDs   []string
	projectCosts          []model.ProjectCostItem
	lenderIndications     []model.LenderIndicationItem
	skipExisting          bool
	errors                []string
}

type blueBookImportRelationRow struct {
	row     int
	sheet   string
	code    string
	label   string
	draft   *blueBookImportProjectDraft
	status  string
	message string
}

func (d *blueBookImportProjectDraft) addError(message string) {
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

func (d *blueBookImportProjectDraft) failed() bool {
	return len(d.errors) > 0
}

func (s *BlueBookService) PreviewBlueBookProjects(ctx context.Context, bbID pgtype.UUID, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processBlueBookProjectsWorkbook(ctx, bbID, fileName, reader, size, false)
}

func (s *BlueBookService) ImportBlueBookProjects(ctx context.Context, bbID pgtype.UUID, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processBlueBookProjectsWorkbook(ctx, bbID, fileName, reader, size, true)
}

func (s *BlueBookService) processBlueBookProjectsWorkbook(ctx context.Context, bbID pgtype.UUID, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
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
	blueBook, err := qtx.GetBlueBook(ctx, bbID)
	if err != nil {
		return nil, mapNotFound(err, "Blue Book tidak ditemukan")
	}

	masterSvc := &MasterService{db: s.db, queries: s.queries}
	lookups, err := masterSvc.loadMasterImportLookups(ctx, qtx)
	if err != nil {
		return nil, err
	}

	response, createdIDs, err := s.buildBlueBookImportPreview(ctx, qtx, workbook, lookups, blueBook, fileName)
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
		return nil, apperrors.Internal("Gagal menyimpan hasil import Blue Book")
	}

	if s.broker != nil {
		for _, id := range createdIDs {
			s.broker.Publish("bb_project.created", map[string]string{"id": id})
		}
	}

	return response, nil
}

func (s *BlueBookService) buildBlueBookImportPreview(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, blueBook queries.GetBlueBookRow, fileName string) (*model.MasterImportResponse, []string, error) {
	inputResult := model.MasterImportSheetResult{Sheet: blueBookImportSheetInput}
	projects, projectsByCode, err := s.parseBlueBookInputRows(ctx, qtx, workbook, lookups, blueBook, &inputResult)
	if err != nil {
		return nil, nil, err
	}

	relationResults := map[string]model.MasterImportSheetResult{
		blueBookImportSheetEA:               {Sheet: blueBookImportSheetEA},
		blueBookImportSheetIA:               {Sheet: blueBookImportSheetIA},
		blueBookImportSheetLocations:        {Sheet: blueBookImportSheetLocations},
		blueBookImportSheetNationalPriority: {Sheet: blueBookImportSheetNationalPriority},
		blueBookImportSheetProjectCost:      {Sheet: blueBookImportSheetProjectCost},
		blueBookImportSheetLenderIndication: {Sheet: blueBookImportSheetLenderIndication},
	}
	relationRows := make([]blueBookImportRelationRow, 0)

	relationRows = append(relationRows, s.parseBlueBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, blueBookImportSheetEA, "executing_agency_name", roleExecutingAgency)...)
	relationRows = append(relationRows, s.parseBlueBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, blueBookImportSheetIA, "implementing_agency_name", roleImplementingAgency)...)
	relationRows = append(relationRows, s.parseBlueBookLocationRelation(workbook, lookups, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseBlueBookNationalPriorityRelation(workbook, lookups, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseBlueBookProjectCostRelation(workbook, projectsByCode, relationResults)...)
	relationRows = append(relationRows, s.parseBlueBookLenderIndicationRelation(workbook, lookups, projectsByCode, relationResults)...)

	for _, draft := range projects {
		if draft.skipExisting || draft.failed() {
			continue
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
			addImportSkipped(&inputResult, draft.row, fmt.Sprintf("%s - %s", draft.bbCode, draft.projectName))
		case draft.failed():
			addImportError(&inputResult, draft.row, strings.Join(draft.errors, "; "))
		default:
			isRevisionSnapshot := draft.projectIdentityID.Valid
			identityID := draft.projectIdentityID
			if !identityID.Valid {
				identity, err := qtx.CreateProjectIdentity(ctx)
				if err != nil {
					return nil, nil, fromPg(err)
				}
				identityID = identity.ID
			}
			created, err := qtx.CreateBBProject(ctx, queries.CreateBBProjectParams{
				BlueBookID:        blueBook.ID,
				ProjectIdentityID: identityID,
				ProgramTitleID:    draft.programTitleID,
				BbCode:            draft.bbCode,
				ProjectName:       draft.projectName,
				Duration:          int4Ptr(draft.duration),
				Objective:         nullableTextPtr(draft.objective),
				ScopeOfWork:       nullableTextPtr(draft.scopeOfWork),
				Outputs:           nullableTextPtr(draft.outputs),
				Outcomes:          nullableTextPtr(draft.outcomes),
			})
			if err != nil {
				return nil, nil, fromPg(err)
			}
			req := model.CreateBBProjectRequest{
				ExecutingAgencyIDs:    draft.executingAgencyIDs,
				ImplementingAgencyIDs: draft.implementingAgencyIDs,
				BappenasPartnerIDs:    draft.bappenasPartnerIDs,
				LocationIDs:           draft.locationIDs,
				NationalPriorityIDs:   draft.nationalPriorityIDs,
				ProjectCosts:          draft.projectCosts,
				LenderIndications:     draft.lenderIndications,
			}
			if err := s.replaceBBProjectChildren(ctx, qtx, created.ID, req); err != nil {
				return nil, nil, err
			}
			createdIDs = append(createdIDs, model.UUIDToString(created.ID))
			message := "Created new logical BB Project"
			if isRevisionSnapshot {
				message = "Created revision snapshot for existing logical BB Project"
			}
			addImportCreatedWithMessage(&inputResult, draft.row, fmt.Sprintf("%s - %s", draft.bbCode, draft.projectName), message)
		}
	}

	for _, relation := range relationRows {
		result := relationResults[relation.sheet]
		status, message := relationStatus(relation)
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
			relationResults[blueBookImportSheetEA],
			relationResults[blueBookImportSheetIA],
			relationResults[blueBookImportSheetLocations],
			relationResults[blueBookImportSheetNationalPriority],
			relationResults[blueBookImportSheetProjectCost],
			relationResults[blueBookImportSheetLenderIndication],
		},
	}
	recalculateImportTotals(response)

	return response, createdIDs, nil
}

func (s *BlueBookService) parseBlueBookInputRows(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, blueBook queries.GetBlueBookRow, result *model.MasterImportSheetResult) ([]*blueBookImportProjectDraft, map[string]*blueBookImportProjectDraft, error) {
	rows, ok := workbook.importRows(blueBookImportSheetInput, []string{"program_title", "bb_code", "project_name"})
	if !ok {
		addImportError(result, 0, "Sheet Input Data tidak ditemukan")
		return nil, map[string]*blueBookImportProjectDraft{}, nil
	}
	if hasImportHeaderError(result, rows) {
		return nil, map[string]*blueBookImportProjectDraft{}, nil
	}

	projects := make([]*blueBookImportProjectDraft, 0, len(rows))
	projectsByCode := make(map[string]*blueBookImportProjectDraft, len(rows))
	seenCodes := map[string]struct{}{}

	for _, row := range rows {
		draft := &blueBookImportProjectDraft{
			row:         row.number,
			bbCode:      strings.TrimSpace(row.value("bb_code")),
			projectName: strings.TrimSpace(row.value("project_name")),
			objective:   row.optionalString("objective"),
			scopeOfWork: row.optionalString("scope_of_work"),
			outputs:     row.optionalString("outputs"),
			outcomes:    row.optionalString("outcomes"),
		}
		duration, err := parseImportOptionalPositiveInt32(row.value("duration"))
		if err != nil {
			draft.addError("Duration harus berupa jumlah bulan positif")
		}
		draft.duration = duration
		projects = append(projects, draft)

		if draft.bbCode == "" {
			draft.addError("BB Code wajib diisi")
			continue
		}
		codeKey := normalizeLookupKey(draft.bbCode)
		if _, exists := seenCodes[codeKey]; exists {
			draft.addError("BB Code duplikat di workbook")
			continue
		}
		seenCodes[codeKey] = struct{}{}
		projectsByCode[codeKey] = draft

		existing, err := qtx.GetBBProjectByBlueBookAndCode(ctx, queries.GetBBProjectByBlueBookAndCodeParams{BlueBookID: blueBook.ID, Lower: draft.bbCode})
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa BB Code")
		}
		if err == nil && existing.ID.Valid {
			draft.skipExisting = true
			continue
		}
		previous, err := qtx.FindPreviousBBProjectByCodeForBlueBook(ctx, queries.FindPreviousBBProjectByCodeForBlueBookParams{ID: blueBook.ID, Lower: draft.bbCode})
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa histori BB Code")
		}
		if err == nil && previous.ID.Valid {
			draft.projectIdentityID = previous.ProjectIdentityID
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

		partnerName := row.value("bappenas_partners")
		if partnerName != "" {
			for _, name := range splitImportNames(partnerName) {
				partner, exists := lookups.bappenasPartnersByName[normalizeLookupKey(name)]
				if !exists {
					draft.addError(fmt.Sprintf("Mitra Kerja Bappenas %q belum ada di master data", name))
					continue
				}
				draft.bappenasPartnerIDs = append(draft.bappenasPartnerIDs, model.UUIDToString(partner.ID))
			}
		}
	}

	return projects, projectsByCode, nil
}

func (s *BlueBookService) parseBlueBookInstitutionRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, sheetName, nameHeader, role string) []blueBookImportRelationRow {
	result := relationResults[sheetName]
	rows, ok := workbook.importRows(sheetName, []string{"bb_code", nameHeader})
	if !ok {
		addImportError(&result, 0, fmt.Sprintf("Sheet %s tidak ditemukan", sheetName))
		relationResults[sheetName] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[sheetName] = result
		return nil
	}

	relations := make([]blueBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("bb_code")
		name := row.value(nameHeader)
		label := fmt.Sprintf("%s - %s", code, name)
		relation := blueBookImportRelationRow{row: row.number, sheet: sheetName, code: code, label: label}
		draft := projectDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code dan nama institution wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("%s baris %d tidak lengkap", sheetName, row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code tidak ada di sheet Input Data"
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

func (s *BlueBookService) parseBlueBookLocationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetLocations]
	rows, ok := workbook.importRows(blueBookImportSheetLocations, []string{"bb_code", "location_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Locations tidak ditemukan")
		relationResults[blueBookImportSheetLocations] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[blueBookImportSheetLocations] = result
		return nil
	}

	relations := make([]blueBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("bb_code")
		name := row.value("location_name")
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetLocations, code: code, label: fmt.Sprintf("%s - %s", code, name)}
		draft := projectDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code dan Location Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Locations baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code tidak ada di sheet Input Data"
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

	relationResults[blueBookImportSheetLocations] = result
	return relations
}

func (s *BlueBookService) parseBlueBookNationalPriorityRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetNationalPriority]
	rows, ok := workbook.importRows(blueBookImportSheetNationalPriority, []string{"bb_code", "national_priority_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - National Priority tidak ditemukan")
		relationResults[blueBookImportSheetNationalPriority] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[blueBookImportSheetNationalPriority] = result
		return nil
	}

	relations := make([]blueBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("bb_code")
		name := row.value("national_priority_name")
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetNationalPriority, code: code, label: fmt.Sprintf("%s - %s", code, name)}
		draft := projectDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code dan National Priority Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - National Priority baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		priority, exists := lookups.nationalPriorityByTitle(name)
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("National Priority %q tidak ada di master data", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + model.UUIDToString(priority.ID)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		draft.nationalPriorityIDs = append(draft.nationalPriorityIDs, model.UUIDToString(priority.ID))
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[blueBookImportSheetNationalPriority] = result
	return relations
}

func (s *BlueBookService) parseBlueBookProjectCostRelation(workbook *xlsxWorkbook, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetProjectCost]
	rows, ok := workbook.importRows(blueBookImportSheetProjectCost, []string{"bb_code", "funding_type", "funding_category"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Project Cost tidak ditemukan")
		relationResults[blueBookImportSheetProjectCost] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[blueBookImportSheetProjectCost] = result
		return nil
	}

	relations := make([]blueBookImportRelationRow, 0, len(rows))
	for _, row := range rows {
		code := row.value("bb_code")
		fundingTypeRaw := row.value("funding_type")
		category := row.value("funding_category")
		label := fmt.Sprintf("%s - %s/%s", code, fundingTypeRaw, category)
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetProjectCost, code: code, label: label}
		draft := projectDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || fundingTypeRaw == "" || category == "" {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code, Funding Type, dan Funding Category wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Project Cost baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		fundingType, ok := normalizeProjectFundingType(fundingTypeRaw)
		if !ok {
			relation.status = masterImportStatusFailed
			relation.message = "Funding Type harus Foreign atau Counterpart"
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
		draft.projectCosts = append(draft.projectCosts, model.ProjectCostItem{FundingType: fundingType, FundingCategory: category, AmountUSD: amount})
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[blueBookImportSheetProjectCost] = result
	return relations
}

func (s *BlueBookService) parseBlueBookLenderIndicationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetLenderIndication]
	rows, ok := workbook.importRows(blueBookImportSheetLenderIndication, []string{"bb_code", "lender_name"})
	if !ok {
		addImportError(&result, 0, "Sheet Relasi - Lender Indication tidak ditemukan")
		relationResults[blueBookImportSheetLenderIndication] = result
		return nil
	}
	if hasImportHeaderError(&result, rows) {
		relationResults[blueBookImportSheetLenderIndication] = result
		return nil
	}

	relations := make([]blueBookImportRelationRow, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		code := row.value("bb_code")
		name := row.value("lender_name")
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetLenderIndication, code: code, label: fmt.Sprintf("%s - %s", code, name)}
		draft := projectDraftByCode(projectsByCode, code)
		relation.draft = draft
		if code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code dan Lender Name wajib diisi"
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Lender Indication baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		lender, exists := lookups.lendersByName[normalizeLookupKey(name)]
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q belum ada di master data", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := normalizeLookupKey(code) + "|" + model.UUIDToString(lender.ID)
		if _, exists := seen[key]; exists {
			relation.status = masterImportStatusSkip
			relation.message = "Duplikat relasi di workbook, dilewati"
			relations = append(relations, relation)
			continue
		}
		seen[key] = struct{}{}
		remarks := row.optionalString("keterangan")
		if remarks == nil {
			remarks = row.optionalString("remarks")
		}
		draft.lenderIndications = append(draft.lenderIndications, model.LenderIndicationItem{LenderID: model.UUIDToString(lender.ID), Remarks: remarks})
		relation.status = masterImportStatusCreate
		relations = append(relations, relation)
	}

	relationResults[blueBookImportSheetLenderIndication] = result
	return relations
}

func projectDraftByCode(projectsByCode map[string]*blueBookImportProjectDraft, code string) *blueBookImportProjectDraft {
	if strings.TrimSpace(code) == "" {
		return nil
	}
	return projectsByCode[normalizeLookupKey(code)]
}

func relationStatus(relation blueBookImportRelationRow) (string, string) {
	if relation.status == masterImportStatusFailed || relation.status == masterImportStatusSkip {
		return relation.status, relation.message
	}
	if relation.draft == nil {
		return masterImportStatusFailed, "BB Code tidak ada di sheet Input Data"
	}
	if relation.draft.skipExisting {
		return masterImportStatusSkip, "Project sudah ada, relasi dilewati"
	}
	if relation.draft.failed() {
		return masterImportStatusFailed, "Project terkait gagal validasi"
	}
	return masterImportStatusCreate, ""
}

func normalizeProjectFundingType(value string) (string, bool) {
	switch normalizeLookupKey(value) {
	case "foreign":
		return "Foreign", true
	case "counterpart":
		return "Counterpart", true
	default:
		return "", false
	}
}

func parseImportFloat(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	value = strings.ReplaceAll(value, ",", "")
	return strconv.ParseFloat(value, 64)
}

func splitImportNames(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';'
	})
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func recalculateImportTotals(response *model.MasterImportResponse) {
	response.TotalInserted = 0
	response.TotalSkipped = 0
	response.TotalFailed = 0
	for _, sheet := range response.Sheets {
		response.TotalInserted += sheet.Inserted
		response.TotalSkipped += sheet.Skipped
		response.TotalFailed += sheet.Failed
	}
}
