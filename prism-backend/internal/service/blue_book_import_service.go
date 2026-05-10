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
	blueBookImportSheetHeader           = "Blue Book"
	blueBookImportSheetInput            = "Input Data"
	blueBookImportSheetEA               = "Relasi - EA"
	blueBookImportSheetIA               = "Relasi - IA"
	blueBookImportSheetLocations        = "Relasi - Locations"
	blueBookImportSheetNationalPriority = "Relasi - National Priority"
	blueBookImportSheetProjectCost      = "Relasi - Project Cost"
	blueBookImportSheetLenderIndication = "Relasi - Lender Indication"
)

type blueBookImportParseMode struct {
	requireBlueBookKey bool
	singleTarget       *blueBookImportTargetDraft
	targetsByKey       map[string]*blueBookImportTargetDraft
}

type blueBookImportTargetDraft struct {
	row                int
	key                string
	periodName         string
	periodID           pgtype.UUID
	publishDate        pgtype.Date
	revisionNumber     int32
	revisionYear       pgtype.Int4
	status             string
	replacesRef        string
	replacesBlueBookID pgtype.UUID
	replacesTarget     *blueBookImportTargetDraft
	blueBookID         pgtype.UUID
	existing           bool
	skipExisting       bool
	created            bool
	errors             []string
}

type blueBookImportProjectDraft struct {
	row                   int
	blueBookKey           string
	target                *blueBookImportTargetDraft
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
	bbKey   string
	code    string
	label   string
	draft   *blueBookImportProjectDraft
	status  string
	message string
}

func (d *blueBookImportTargetDraft) addError(message string) {
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

func (d *blueBookImportTargetDraft) failed() bool {
	return d == nil || len(d.errors) > 0
}

func (d *blueBookImportTargetDraft) label() string {
	if d == nil {
		return "Blue Book"
	}
	label := strings.TrimSpace(d.key)
	if label == "" {
		label = strings.TrimSpace(d.periodName)
	}
	if label == "" {
		return "Blue Book"
	}
	return label
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

func (s *BlueBookService) PreviewMultiBlueBookImport(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processMultiBlueBookWorkbook(ctx, fileName, reader, size, false)
}

func (s *BlueBookService) ImportMultiBlueBook(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processMultiBlueBookWorkbook(ctx, fileName, reader, size, true)
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

func (s *BlueBookService) processMultiBlueBookWorkbook(ctx context.Context, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
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

	response, createdIDs, err := s.buildMultiBlueBookImportPreview(ctx, qtx, workbook, lookups, fileName)
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
	target := &blueBookImportTargetDraft{
		periodName:     blueBook.PeriodName,
		periodID:       blueBook.PeriodID,
		publishDate:    blueBook.PublishDate,
		revisionNumber: blueBook.RevisionNumber,
		revisionYear:   blueBook.RevisionYear,
		status:         blueBook.Status,
		blueBookID:     blueBook.ID,
	}
	mode := blueBookImportParseMode{singleTarget: target}
	inputResult := model.MasterImportSheetResult{Sheet: blueBookImportSheetInput}
	projects, projectsByCode, err := s.parseBlueBookInputRows(ctx, qtx, workbook, lookups, mode, &inputResult)
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

	relationRows = append(relationRows, s.parseBlueBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, mode, blueBookImportSheetEA, "executing_agency_name", roleExecutingAgency)...)
	relationRows = append(relationRows, s.parseBlueBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, mode, blueBookImportSheetIA, "implementing_agency_name", roleImplementingAgency)...)
	relationRows = append(relationRows, s.parseBlueBookLocationRelation(workbook, lookups, projectsByCode, relationResults, mode)...)
	relationRows = append(relationRows, s.parseBlueBookNationalPriorityRelation(workbook, lookups, projectsByCode, relationResults, mode)...)
	relationRows = append(relationRows, s.parseBlueBookProjectCostRelation(workbook, projectsByCode, relationResults, mode)...)
	relationRows = append(relationRows, s.parseBlueBookLenderIndicationRelation(workbook, lookups, projectsByCode, relationResults, mode)...)

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
		label := blueBookProjectImportLabel(draft)
		switch {
		case draft.skipExisting:
			addImportSkipped(&inputResult, draft.row, label)
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
				BlueBookID:        draft.target.blueBookID,
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
			addImportCreatedWithMessage(&inputResult, draft.row, label, message)
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

func (s *BlueBookService) buildMultiBlueBookImportPreview(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, fileName string) (*model.MasterImportResponse, []string, error) {
	headerResult := model.MasterImportSheetResult{Sheet: blueBookImportSheetHeader}
	targets, targetsByKey, err := s.parseBlueBookHeaderRows(ctx, qtx, workbook, lookups, &headerResult)
	if err != nil {
		return nil, nil, err
	}

	for _, target := range targets {
		if target.failed() {
			continue
		}
		if err := s.ensureBlueBookImportTarget(ctx, qtx, target, map[string]bool{}); err != nil {
			return nil, nil, err
		}
	}

	for _, target := range targets {
		switch {
		case target.failed():
			addImportError(&headerResult, target.row, strings.Join(target.errors, "; "))
		case target.skipExisting:
			addImportSkipped(&headerResult, target.row, target.label())
		default:
			addImportCreatedWithMessage(&headerResult, target.row, target.label(), "Created Blue Book header")
		}
	}

	mode := blueBookImportParseMode{requireBlueBookKey: true, targetsByKey: targetsByKey}
	inputResult := model.MasterImportSheetResult{Sheet: blueBookImportSheetInput}
	projects, projectsByCode, err := s.parseBlueBookInputRows(ctx, qtx, workbook, lookups, mode, &inputResult)
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

	relationRows = append(relationRows, s.parseBlueBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, mode, blueBookImportSheetEA, "executing_agency_name", roleExecutingAgency)...)
	relationRows = append(relationRows, s.parseBlueBookInstitutionRelation(workbook, lookups, projectsByCode, relationResults, mode, blueBookImportSheetIA, "implementing_agency_name", roleImplementingAgency)...)
	relationRows = append(relationRows, s.parseBlueBookLocationRelation(workbook, lookups, projectsByCode, relationResults, mode)...)
	relationRows = append(relationRows, s.parseBlueBookNationalPriorityRelation(workbook, lookups, projectsByCode, relationResults, mode)...)
	relationRows = append(relationRows, s.parseBlueBookProjectCostRelation(workbook, projectsByCode, relationResults, mode)...)
	relationRows = append(relationRows, s.parseBlueBookLenderIndicationRelation(workbook, lookups, projectsByCode, relationResults, mode)...)

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
		label := blueBookProjectImportLabel(draft)
		switch {
		case draft.skipExisting:
			addImportSkipped(&inputResult, draft.row, label)
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
				BlueBookID:        draft.target.blueBookID,
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
			addImportCreatedWithMessage(&inputResult, draft.row, label, message)
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
			headerResult,
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

func (s *BlueBookService) parseBlueBookHeaderRows(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, result *model.MasterImportSheetResult) ([]*blueBookImportTargetDraft, map[string]*blueBookImportTargetDraft, error) {
	rows, ok := workbook.importRows(blueBookImportSheetHeader, []string{"blue_book_key", "period_name", "publish_date", "status"})
	if !ok {
		addImportError(result, 0, "Sheet Blue Book tidak ditemukan")
		return nil, map[string]*blueBookImportTargetDraft{}, nil
	}
	if hasImportHeaderError(result, rows) {
		return nil, map[string]*blueBookImportTargetDraft{}, nil
	}

	targets := make([]*blueBookImportTargetDraft, 0, len(rows))
	targetsByKey := make(map[string]*blueBookImportTargetDraft, len(rows))
	versions := map[string]*blueBookImportTargetDraft{}

	for _, row := range rows {
		target := &blueBookImportTargetDraft{
			row:         row.number,
			key:         strings.TrimSpace(row.value("blue_book_key")),
			replacesRef: strings.TrimSpace(row.value("replaces_blue_book_ref")),
		}
		targets = append(targets, target)

		if target.key == "" {
			target.addError("Blue Book Key wajib diisi")
		} else {
			key := normalizeLookupKey(target.key)
			if existing, exists := targetsByKey[key]; exists {
				target.addError("Blue Book Key duplikat di workbook")
				existing.addError("Blue Book Key duplikat di workbook")
			} else {
				targetsByKey[key] = target
			}
		}

		periodName := row.value("period_name")
		if periodName == "" {
			target.addError("Period Name wajib diisi")
		} else if period, exists := lookups.periodByLabel(periodName); exists {
			target.periodName = period.Name
			target.periodID = period.ID
		} else {
			target.addError(fmt.Sprintf("Period Name %q belum ada di master data", periodName))
		}

		publishDate, err := parseDate(row.value("publish_date"), "publish_date")
		if err != nil {
			target.addError("Publish Date wajib diisi dengan format YYYY-MM-DD")
		} else {
			target.publishDate = publishDate
		}

		revisionNumber := 0
		if rawRevision := row.value("revision_number"); rawRevision != "" {
			parsed, err := parseImportInt(rawRevision)
			if err != nil || parsed < 0 {
				target.addError("Revision Number wajib berupa angka 0 atau lebih")
			} else {
				revisionNumber = parsed
			}
		}
		target.revisionNumber = int32(revisionNumber)

		if rawRevisionYear := row.value("revision_year"); rawRevisionYear != "" {
			parsed, err := parseImportInt(rawRevisionYear)
			if err != nil || parsed <= 0 {
				target.addError("Revision Year wajib berupa angka tahun")
			} else {
				target.revisionYear = pgtype.Int4{Int32: int32(parsed), Valid: true}
			}
		}

		status, err := parseBlueBookImportStatus(row.value("status"))
		if err != nil {
			target.addError(err.Error())
		} else {
			target.status = status
		}

		if target.periodID.Valid {
			versionKey := blueBookVersionImportKey(target.periodID, target.revisionNumber, target.revisionYear)
			if existing, exists := versions[versionKey]; exists {
				target.addError("Period, Revision Number, dan Revision Year duplikat di workbook")
				existing.addError("Period, Revision Number, dan Revision Year duplikat di workbook")
			} else {
				versions[versionKey] = target
			}
		}
	}

	for _, target := range targets {
		if target.replacesRef == "" {
			continue
		}
		if referenced, exists := targetsByKey[normalizeLookupKey(target.replacesRef)]; exists {
			if referenced == target {
				target.addError("Replaces Blue Book Ref tidak boleh mengarah ke Blue Book Key yang sama")
				continue
			}
			if referenced.periodID.Valid && target.periodID.Valid && model.UUIDToString(referenced.periodID) != model.UUIDToString(target.periodID) {
				target.addError("Replaces Blue Book Ref harus berasal dari Period yang sama")
				continue
			}
			target.replacesTarget = referenced
			continue
		}

		id, err := model.ParseUUID(target.replacesRef)
		if err != nil {
			target.addError("Replaces Blue Book Ref harus Blue Book Key workbook atau UUID Blue Book")
			continue
		}
		source, err := qtx.GetBlueBook(ctx, id)
		if err != nil {
			target.addError("Blue Book sumber revisi tidak ditemukan")
			continue
		}
		if target.periodID.Valid && model.UUIDToString(source.PeriodID) != model.UUIDToString(target.periodID) {
			target.addError("Replaces Blue Book Ref harus berasal dari Period yang sama")
			continue
		}
		target.replacesBlueBookID = id
	}

	for _, target := range targets {
		if !target.periodID.Valid {
			continue
		}
		existing, err := qtx.GetBlueBookByPeriodAndVersion(ctx, queries.GetBlueBookByPeriodAndVersionParams{
			PeriodID:          target.periodID,
			RevisionNumber:    target.revisionNumber,
			RevisionYearValid: target.revisionYear.Valid,
			RevisionYear:      target.revisionYear.Int32,
		})
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa versi Blue Book")
		}
		if err == pgx.ErrNoRows {
			continue
		}

		target.blueBookID = existing.ID
		target.existing = true
		target.skipExisting = true
		if target.status != "" && existing.Status != target.status {
			target.addError("Status Blue Book existing tidak sama dengan workbook")
		}
		if !blueBookDateEqual(existing.PublishDate, target.publishDate) {
			target.addError("Publish Date Blue Book existing tidak sama dengan workbook")
		}
	}

	for _, target := range targets {
		if !target.existing {
			continue
		}
		existing, err := qtx.GetBlueBook(ctx, target.blueBookID)
		if err != nil {
			return nil, nil, apperrors.Internal("Gagal memeriksa versi Blue Book")
		}
		if target.replacesTarget != nil && !target.replacesTarget.blueBookID.Valid {
			target.addError("Replaces Blue Book Ref untuk Blue Book existing harus mengarah ke Blue Book existing")
		}
		expectedReplacesID := target.replacesBlueBookID
		if target.replacesTarget != nil && target.replacesTarget.blueBookID.Valid {
			expectedReplacesID = target.replacesTarget.blueBookID
		}
		if !blueBookOptionalUUIDEqual(existing.ReplacesBlueBookID, expectedReplacesID) {
			target.addError("Replaces Blue Book Ref tidak sama dengan Blue Book existing")
		}
	}

	return targets, targetsByKey, nil
}

func (s *BlueBookService) ensureBlueBookImportTarget(ctx context.Context, qtx *queries.Queries, target *blueBookImportTargetDraft, creating map[string]bool) error {
	if target == nil || target.failed() || target.blueBookID.Valid {
		return nil
	}
	key := normalizeLookupKey(target.key)
	if creating[key] {
		target.addError("Replaces Blue Book Ref membentuk siklus")
		return nil
	}
	creating[key] = true
	defer delete(creating, key)

	replacesID := target.replacesBlueBookID
	if target.replacesTarget != nil {
		if err := s.ensureBlueBookImportTarget(ctx, qtx, target.replacesTarget, creating); err != nil {
			return err
		}
		if target.replacesTarget.failed() || !target.replacesTarget.blueBookID.Valid {
			target.addError("Blue Book sumber revisi gagal validasi")
			return nil
		}
		replacesID = target.replacesTarget.blueBookID
	}

	if err := s.ensureBlueBookVersionAvailable(ctx, qtx, target.periodID, target.revisionNumber, target.revisionYear, pgtype.UUID{}); err != nil {
		target.addError("Blue Book dengan Period dan versi yang sama sudah ada")
		return nil
	}
	created, err := qtx.CreateBlueBook(ctx, queries.CreateBlueBookParams{
		PeriodID:           target.periodID,
		ReplacesBlueBookID: replacesID,
		PublishDate:        target.publishDate,
		RevisionNumber:     target.revisionNumber,
		RevisionYear:       target.revisionYear,
		Status:             target.status,
	})
	if err != nil {
		return fromPg(err)
	}
	target.blueBookID = created.ID
	target.created = true
	return nil
}

func (s *BlueBookService) parseBlueBookInputRows(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups, mode blueBookImportParseMode, result *model.MasterImportSheetResult) ([]*blueBookImportProjectDraft, map[string]*blueBookImportProjectDraft, error) {
	rows, ok := workbook.importRows(blueBookImportSheetInput, mode.requiredHeaders("program_title", "bb_code", "project_name"))
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
		target, blueBookKey, targetError := mode.targetForRow(row)
		draft := &blueBookImportProjectDraft{
			row:         row.number,
			blueBookKey: blueBookKey,
			target:      target,
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

		if targetError != "" {
			draft.addError(targetError)
		}
		if target != nil && target.failed() {
			draft.addError("Blue Book terkait gagal validasi")
		}

		if draft.bbCode == "" {
			draft.addError("BB Code wajib diisi")
			continue
		}
		codeKey := blueBookProjectLookupKey(blueBookKey, draft.bbCode)
		if _, exists := seenCodes[codeKey]; exists {
			draft.addError("BB Code duplikat di workbook untuk Blue Book Key yang sama")
			continue
		}
		seenCodes[codeKey] = struct{}{}
		projectsByCode[codeKey] = draft
		if target == nil || target.failed() || !target.blueBookID.Valid {
			continue
		}

		existing, err := qtx.GetBBProjectByBlueBookAndCode(ctx, queries.GetBBProjectByBlueBookAndCodeParams{BlueBookID: target.blueBookID, Lower: draft.bbCode})
		if err != nil && err != pgx.ErrNoRows {
			return nil, nil, apperrors.Internal("Gagal memeriksa BB Code")
		}
		if err == nil && existing.ID.Valid {
			draft.skipExisting = true
			continue
		}
		previous, err := qtx.FindPreviousBBProjectByCodeForBlueBook(ctx, queries.FindPreviousBBProjectByCodeForBlueBookParams{ID: target.blueBookID, Lower: draft.bbCode})
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

func (s *BlueBookService) parseBlueBookInstitutionRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, mode blueBookImportParseMode, sheetName, nameHeader, role string) []blueBookImportRelationRow {
	result := relationResults[sheetName]
	rows, ok := workbook.importRows(sheetName, mode.requiredHeaders("bb_code", nameHeader))
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
		blueBookKey := row.value("blue_book_key")
		code := row.value("bb_code")
		name := row.value(nameHeader)
		label := blueBookRelationLabel(blueBookKey, code, name)
		relation := blueBookImportRelationRow{row: row.number, sheet: sheetName, bbKey: blueBookKey, code: code, label: label}
		draft, targetError := blueBookDraftForRelation(mode, projectsByCode, row, code)
		relation.draft = draft
		if blueBookRequiredMissing(mode, blueBookKey) || code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = blueBookMissingMessage(mode, "BB Code dan nama institution wajib diisi")
			if draft != nil {
				draft.addError(fmt.Sprintf("%s baris %d tidak lengkap", sheetName, row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if targetError != "" {
			relation.status = masterImportStatusFailed
			relation.message = targetError
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
		key := blueBookProjectLookupKey(blueBookKey, code) + "|" + model.UUIDToString(institution.ID) + "|" + role
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

func (s *BlueBookService) parseBlueBookLocationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, mode blueBookImportParseMode) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetLocations]
	rows, ok := workbook.importRows(blueBookImportSheetLocations, mode.requiredHeaders("bb_code", "location_name"))
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
		blueBookKey := row.value("blue_book_key")
		code := row.value("bb_code")
		name := row.value("location_name")
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetLocations, bbKey: blueBookKey, code: code, label: blueBookRelationLabel(blueBookKey, code, name)}
		draft, targetError := blueBookDraftForRelation(mode, projectsByCode, row, code)
		relation.draft = draft
		if blueBookRequiredMissing(mode, blueBookKey) || code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = blueBookMissingMessage(mode, "BB Code dan Location Name wajib diisi")
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Locations baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if targetError != "" {
			relation.status = masterImportStatusFailed
			relation.message = targetError
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
		key := blueBookProjectLookupKey(blueBookKey, code) + "|" + model.UUIDToString(region.ID)
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

func (s *BlueBookService) parseBlueBookNationalPriorityRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, mode blueBookImportParseMode) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetNationalPriority]
	rows, ok := workbook.importRows(blueBookImportSheetNationalPriority, mode.requiredHeaders("bb_code", "national_priority_name"))
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
		blueBookKey := row.value("blue_book_key")
		code := row.value("bb_code")
		name := row.value("national_priority_name")
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetNationalPriority, bbKey: blueBookKey, code: code, label: blueBookRelationLabel(blueBookKey, code, name)}
		draft, targetError := blueBookDraftForRelation(mode, projectsByCode, row, code)
		relation.draft = draft
		if blueBookRequiredMissing(mode, blueBookKey) || code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = blueBookMissingMessage(mode, "BB Code dan National Priority Name wajib diisi")
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - National Priority baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if targetError != "" {
			relation.status = masterImportStatusFailed
			relation.message = targetError
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
		key := blueBookProjectLookupKey(blueBookKey, code) + "|" + model.UUIDToString(priority.ID)
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

func (s *BlueBookService) parseBlueBookProjectCostRelation(workbook *xlsxWorkbook, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, mode blueBookImportParseMode) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetProjectCost]
	rows, ok := workbook.importRows(blueBookImportSheetProjectCost, mode.requiredHeaders("bb_code", "funding_type", "funding_category"))
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
		blueBookKey := row.value("blue_book_key")
		code := row.value("bb_code")
		fundingTypeRaw := row.value("funding_type")
		category := row.value("funding_category")
		label := blueBookRelationLabel(blueBookKey, code, fmt.Sprintf("%s/%s", fundingTypeRaw, category))
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetProjectCost, bbKey: blueBookKey, code: code, label: label}
		draft, targetError := blueBookDraftForRelation(mode, projectsByCode, row, code)
		relation.draft = draft
		if blueBookRequiredMissing(mode, blueBookKey) || code == "" || fundingTypeRaw == "" || category == "" {
			relation.status = masterImportStatusFailed
			relation.message = blueBookMissingMessage(mode, "BB Code, Funding Type, dan Funding Category wajib diisi")
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Project Cost baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if targetError != "" {
			relation.status = masterImportStatusFailed
			relation.message = targetError
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

func (s *BlueBookService) parseBlueBookLenderIndicationRelation(workbook *xlsxWorkbook, lookups *masterImportLookups, projectsByCode map[string]*blueBookImportProjectDraft, relationResults map[string]model.MasterImportSheetResult, mode blueBookImportParseMode) []blueBookImportRelationRow {
	result := relationResults[blueBookImportSheetLenderIndication]
	rows, ok := workbook.importRows(blueBookImportSheetLenderIndication, mode.requiredHeaders("bb_code", "lender_name"))
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
		blueBookKey := row.value("blue_book_key")
		code := row.value("bb_code")
		name := row.value("lender_name")
		relation := blueBookImportRelationRow{row: row.number, sheet: blueBookImportSheetLenderIndication, bbKey: blueBookKey, code: code, label: blueBookRelationLabel(blueBookKey, code, name)}
		draft, targetError := blueBookDraftForRelation(mode, projectsByCode, row, code)
		relation.draft = draft
		if blueBookRequiredMissing(mode, blueBookKey) || code == "" || name == "" {
			relation.status = masterImportStatusFailed
			relation.message = blueBookMissingMessage(mode, "BB Code dan Lender Name wajib diisi")
			if draft != nil {
				draft.addError(fmt.Sprintf("Relasi - Lender Indication baris %d tidak lengkap", row.number))
			}
			relations = append(relations, relation)
			continue
		}
		if targetError != "" {
			relation.status = masterImportStatusFailed
			relation.message = targetError
			relations = append(relations, relation)
			continue
		}
		if draft == nil {
			relation.status = masterImportStatusFailed
			relation.message = "BB Code tidak ada di sheet Input Data"
			relations = append(relations, relation)
			continue
		}
		lender, exists, ambiguous := lookups.lookupLenderReference(name)
		if ambiguous {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q ambigu karena cocok dengan lebih dari satu short_name di master data", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		if !exists {
			relation.status = masterImportStatusFailed
			relation.message = fmt.Sprintf("Lender %q belum ada di master data", name)
			draft.addError(relation.message)
			relations = append(relations, relation)
			continue
		}
		key := blueBookProjectLookupKey(blueBookKey, code) + "|" + model.UUIDToString(lender.ID)
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

func (mode blueBookImportParseMode) requiredHeaders(headers ...string) []string {
	if !mode.requireBlueBookKey {
		return headers
	}
	result := make([]string, 0, len(headers)+1)
	result = append(result, "blue_book_key")
	result = append(result, headers...)
	return result
}

func (mode blueBookImportParseMode) targetForRow(row importRow) (*blueBookImportTargetDraft, string, string) {
	if !mode.requireBlueBookKey {
		if mode.singleTarget == nil {
			return nil, "", "Blue Book target tidak tersedia"
		}
		return mode.singleTarget, "", ""
	}
	blueBookKey := strings.TrimSpace(row.value("blue_book_key"))
	if blueBookKey == "" {
		return nil, "", "Blue Book Key wajib diisi"
	}
	target := mode.targetsByKey[normalizeLookupKey(blueBookKey)]
	if target == nil {
		return nil, blueBookKey, fmt.Sprintf("Blue Book Key %q tidak ada di sheet Blue Book", blueBookKey)
	}
	return target, blueBookKey, ""
}

func blueBookDraftForRelation(mode blueBookImportParseMode, projectsByCode map[string]*blueBookImportProjectDraft, row importRow, code string) (*blueBookImportProjectDraft, string) {
	_, blueBookKey, targetError := mode.targetForRow(row)
	if targetError != "" {
		return nil, targetError
	}
	if strings.TrimSpace(code) == "" {
		return nil, ""
	}
	return projectsByCode[blueBookProjectLookupKey(blueBookKey, code)], ""
}

func blueBookRequiredMissing(mode blueBookImportParseMode, blueBookKey string) bool {
	return mode.requireBlueBookKey && strings.TrimSpace(blueBookKey) == ""
}

func blueBookMissingMessage(mode blueBookImportParseMode, message string) string {
	if mode.requireBlueBookKey {
		return "Blue Book Key, " + message
	}
	return message
}

func blueBookProjectLookupKey(blueBookKey, code string) string {
	if strings.TrimSpace(blueBookKey) == "" {
		return normalizeLookupKey(code)
	}
	return normalizeLookupKey(blueBookKey) + "|" + normalizeLookupKey(code)
}

func blueBookVersionImportKey(periodID pgtype.UUID, revisionNumber int32, revisionYear pgtype.Int4) string {
	year := "null"
	if revisionYear.Valid {
		year = strconv.FormatInt(int64(revisionYear.Int32), 10)
	}
	return model.UUIDToString(periodID) + "|" + strconv.FormatInt(int64(revisionNumber), 10) + "|" + year
}

func blueBookRelationLabel(blueBookKey string, parts ...string) string {
	values := make([]string, 0, len(parts)+1)
	if strings.TrimSpace(blueBookKey) != "" {
		values = append(values, strings.TrimSpace(blueBookKey))
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return strings.Join(values, " - ")
}

func blueBookProjectImportLabel(draft *blueBookImportProjectDraft) string {
	if draft == nil {
		return "BB Project"
	}
	return blueBookRelationLabel(draft.blueBookKey, draft.bbCode, draft.projectName)
}

func parseBlueBookImportStatus(value string) (string, error) {
	status := normalizeLookupKey(value)
	if status == "" || status == "berlaku" || status == "active" {
		return "active", nil
	}
	if status == "tidak berlaku" || status == "superseded" {
		return "superseded", nil
	}
	return "", fmt.Errorf("Status harus Berlaku atau Tidak Berlaku")
}

func blueBookDateEqual(a, b pgtype.Date) bool {
	if a.Valid != b.Valid {
		return false
	}
	if !a.Valid {
		return true
	}
	return a.Time.Equal(b.Time)
}

func blueBookOptionalUUIDEqual(a, b pgtype.UUID) bool {
	if a.Valid != b.Valid {
		return false
	}
	if !a.Valid {
		return true
	}
	return model.UUIDToString(a) == model.UUIDToString(b)
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
