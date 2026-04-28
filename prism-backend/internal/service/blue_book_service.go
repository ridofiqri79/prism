package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/sse"
)

const (
	roleExecutingAgency    = "Executing Agency"
	roleImplementingAgency = "Implementing Agency"
)

type BlueBookService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	broker  *sse.Broker
}

func NewBlueBookService(db *pgxpool.Pool, queries *queries.Queries, broker *sse.Broker) *BlueBookService {
	return &BlueBookService{db: db, queries: queries, broker: broker}
}

func (s *BlueBookService) ListBlueBooks(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.BlueBookResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListBlueBooks(ctx, queries.ListBlueBooksParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar Blue Book")
	}
	total, err := s.queries.CountBlueBooks(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung Blue Book")
	}
	data := make([]model.BlueBookResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, blueBookFromListRow(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *BlueBookService) GetBlueBook(ctx context.Context, id pgtype.UUID) (*model.BlueBookResponse, error) {
	row, err := s.queries.GetBlueBook(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Blue Book tidak ditemukan")
	}
	res := blueBookFromGetRow(row)
	return &res, nil
}

func (s *BlueBookService) CreateBlueBook(ctx context.Context, req model.BlueBookRequest) (*model.BlueBookResponse, error) {
	periodID, publishDate, revisionYear, err := parseBlueBookRequest(req)
	if err != nil {
		return nil, err
	}
	var created queries.BlueBook
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if err := qtx.SupersedeBlueBooksByPeriod(ctx, periodID); err != nil {
			return err
		}
		row, err := qtx.CreateBlueBook(ctx, queries.CreateBlueBookParams{
			PeriodID:       periodID,
			PublishDate:    publishDate,
			RevisionNumber: req.RevisionNumber,
			RevisionYear:   revisionYear,
		})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	return s.GetBlueBook(ctx, created.ID)
}

func (s *BlueBookService) UpdateBlueBook(ctx context.Context, id pgtype.UUID, req model.BlueBookRequest) (*model.BlueBookResponse, error) {
	_, publishDate, revisionYear, err := parseBlueBookRequest(req)
	if err != nil {
		return nil, err
	}
	var updated queries.BlueBook
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateBlueBook(ctx, queries.UpdateBlueBookParams{
			ID:             id,
			PublishDate:    publishDate,
			RevisionNumber: req.RevisionNumber,
			RevisionYear:   revisionYear,
		})
		if err != nil {
			return mapNotFound(err, "Blue Book tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	return s.GetBlueBook(ctx, updated.ID)
}

func (s *BlueBookService) DeleteBlueBook(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.SupersedeBlueBook(ctx, id); err != nil {
			return mapNotFound(err, "Blue Book tidak ditemukan")
		}
		return nil
	})
}

func (s *BlueBookService) ListBBProjects(ctx context.Context, bbID pgtype.UUID, params model.PaginationParams) (*model.ListResponse[model.BBProjectResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListBBProjectsByBlueBook(ctx, queries.ListBBProjectsByBlueBookParams{BlueBookID: bbID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar BB Project")
	}
	total, err := s.queries.CountBBProjectsByBlueBook(ctx, bbID)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung BB Project")
	}
	data := make([]model.BBProjectResponse, 0, len(rows))
	for _, row := range rows {
		res, err := s.buildBBProjectResponse(ctx, row)
		if err != nil {
			return nil, err
		}
		data = append(data, *res)
	}
	return listResponse(data, page, limit, total), nil
}

func (s *BlueBookService) GetBBProject(ctx context.Context, bbID, id pgtype.UUID) (*model.BBProjectResponse, error) {
	row, err := s.queries.GetActiveBBProjectByBlueBook(ctx, queries.GetActiveBBProjectByBlueBookParams{BlueBookID: bbID, ID: id})
	if err != nil {
		return nil, mapNotFound(err, "BB Project tidak ditemukan")
	}
	return s.buildBBProjectResponse(ctx, row)
}

func (s *BlueBookService) CreateBBProject(ctx context.Context, bbID pgtype.UUID, req model.CreateBBProjectRequest) (*model.BBProjectResponse, error) {
	if err := validateBBProjectRequest(req, true); err != nil {
		return nil, err
	}
	if err := s.ensureBBCodeAvailable(ctx, req.BBCode); err != nil {
		return nil, err
	}

	var created queries.BbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetBlueBook(ctx, bbID); err != nil {
			return mapNotFound(err, "Blue Book tidak ditemukan")
		}
		if err := s.validateNationalPriorities(ctx, qtx, req.NationalPriorityIDs); err != nil {
			return err
		}
		project, err := qtx.CreateBBProject(ctx, queries.CreateBBProjectParams{
			BlueBookID:        bbID,
			ProgramTitleID:    uuidOrInvalid(req.ProgramTitleID),
			BappenasPartnerID: uuidOrInvalid(req.BappenasPartnerID),
			BbCode:            strings.TrimSpace(req.BBCode),
			ProjectName:       strings.TrimSpace(req.ProjectName),
			Duration:          nullableTextPtr(req.Duration),
			Objective:         nullableTextPtr(req.Objective),
			ScopeOfWork:       nullableTextPtr(req.ScopeOfWork),
			Outputs:           nullableTextPtr(req.Outputs),
			Outcomes:          nullableTextPtr(req.Outcomes),
		})
		if err != nil {
			return err
		}
		if err := s.replaceBBProjectChildren(ctx, qtx, project.ID, req); err != nil {
			return err
		}
		created = project
		return nil
	}); err != nil {
		return nil, err
	}

	if s.broker != nil {
		s.broker.Publish("bb_project.created", map[string]string{"id": model.UUIDToString(created.ID)})
	}
	return s.buildBBProjectResponse(ctx, created)
}

func (s *BlueBookService) UpdateBBProject(ctx context.Context, bbID, id pgtype.UUID, req model.UpdateBBProjectRequest) (*model.BBProjectResponse, error) {
	if err := validateBBProjectRequest(req, false); err != nil {
		return nil, err
	}

	var updated queries.BbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetActiveBBProjectByBlueBook(ctx, queries.GetActiveBBProjectByBlueBookParams{BlueBookID: bbID, ID: id}); err != nil {
			return mapNotFound(err, "BB Project tidak ditemukan")
		}
		if err := s.validateNationalPriorities(ctx, qtx, req.NationalPriorityIDs); err != nil {
			return err
		}
		project, err := qtx.UpdateBBProject(ctx, queries.UpdateBBProjectParams{
			ID:                id,
			ProgramTitleID:    uuidOrInvalid(req.ProgramTitleID),
			BappenasPartnerID: uuidOrInvalid(req.BappenasPartnerID),
			ProjectName:       strings.TrimSpace(req.ProjectName),
			Duration:          nullableTextPtr(req.Duration),
			Objective:         nullableTextPtr(req.Objective),
			ScopeOfWork:       nullableTextPtr(req.ScopeOfWork),
			Outputs:           nullableTextPtr(req.Outputs),
			Outcomes:          nullableTextPtr(req.Outcomes),
		})
		if err != nil {
			return mapNotFound(err, "BB Project tidak ditemukan")
		}
		if err := s.replaceBBProjectChildren(ctx, qtx, id, req); err != nil {
			return err
		}
		updated = project
		return nil
	}); err != nil {
		return nil, err
	}

	if s.broker != nil {
		s.broker.Publish("bb_project.updated", map[string]string{"id": model.UUIDToString(updated.ID)})
	}
	return s.buildBBProjectResponse(ctx, updated)
}

func (s *BlueBookService) DeleteBBProject(ctx context.Context, bbID, id pgtype.UUID) error {
	var deleted queries.BbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetActiveBBProjectByBlueBook(ctx, queries.GetActiveBBProjectByBlueBookParams{BlueBookID: bbID, ID: id}); err != nil {
			return mapNotFound(err, "BB Project tidak ditemukan")
		}
		row, err := qtx.SoftDeleteBBProject(ctx, id)
		if err != nil {
			return mapNotFound(err, "BB Project tidak ditemukan")
		}
		deleted = row
		return nil
	}); err != nil {
		return err
	}
	if s.broker != nil {
		s.broker.Publish("bb_project.deleted", map[string]string{"id": model.UUIDToString(deleted.ID)})
	}
	return nil
}

func (s *BlueBookService) ListLoI(ctx context.Context, bbProjectID pgtype.UUID) ([]model.LoIResponse, error) {
	rows, err := s.queries.GetLoIsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar LoI")
	}
	return loiResponses(rows), nil
}

func (s *BlueBookService) CreateLoI(ctx context.Context, bbProjectID pgtype.UUID, req model.LoIRequest) (*model.LoIResponse, error) {
	lenderID, date, err := parseLoIRequest(req)
	if err != nil {
		return nil, err
	}
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetBBProject(ctx, bbProjectID); err != nil {
			return mapNotFound(err, "BB Project tidak ditemukan")
		}
		_, err := qtx.CreateLoI(ctx, queries.CreateLoIParams{
			BbProjectID:  bbProjectID,
			LenderID:     lenderID,
			Subject:      strings.TrimSpace(req.Subject),
			Date:         date,
			LetterNumber: nullableTextPtr(req.LetterNumber),
		})
		return err
	}); err != nil {
		return nil, err
	}
	rows, err := s.queries.GetLoIsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil LoI")
	}
	responses := loiResponses(rows)
	if len(responses) == 0 {
		return nil, apperrors.Internal("Gagal mengambil LoI")
	}
	return &responses[0], nil
}

func (s *BlueBookService) DeleteLoI(ctx context.Context, bbProjectID, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return qtx.DeleteLoI(ctx, queries.DeleteLoIParams{ID: id, BbProjectID: bbProjectID})
	})
}

func (s *BlueBookService) replaceBBProjectChildren(ctx context.Context, qtx *queries.Queries, projectID pgtype.UUID, req model.CreateBBProjectRequest) error {
	if err := qtx.DeleteBBProjectInstitutions(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteBBProjectLocations(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteBBProjectNationalPriorities(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteBBProjectCosts(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteLenderIndications(ctx, projectID); err != nil {
		return err
	}

	for _, id := range req.ExecutingAgencyIDs {
		institutionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("executing_agency_ids", "UUID tidak valid")
		}
		if err := qtx.AddBBProjectInstitution(ctx, queries.AddBBProjectInstitutionParams{BbProjectID: projectID, InstitutionID: institutionID, Role: roleExecutingAgency}); err != nil {
			return err
		}
	}
	for _, id := range req.ImplementingAgencyIDs {
		institutionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("implementing_agency_ids", "UUID tidak valid")
		}
		if err := qtx.AddBBProjectInstitution(ctx, queries.AddBBProjectInstitutionParams{BbProjectID: projectID, InstitutionID: institutionID, Role: roleImplementingAgency}); err != nil {
			return err
		}
	}
	for _, id := range req.LocationIDs {
		regionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("location_ids", "UUID tidak valid")
		}
		if err := qtx.AddBBProjectLocation(ctx, queries.AddBBProjectLocationParams{BbProjectID: projectID, RegionID: regionID}); err != nil {
			return err
		}
	}
	for _, id := range req.NationalPriorityIDs {
		priorityID, err := model.ParseUUID(id)
		if err != nil {
			return validation("national_priority_ids", "UUID tidak valid")
		}
		if err := qtx.AddBBProjectNationalPriority(ctx, queries.AddBBProjectNationalPriorityParams{BbProjectID: projectID, NationalPriorityID: priorityID}); err != nil {
			return err
		}
	}
	for _, item := range req.ProjectCosts {
		if _, err := qtx.CreateBBProjectCost(ctx, queries.CreateBBProjectCostParams{
			BbProjectID:     projectID,
			FundingType:     item.FundingType,
			FundingCategory: item.FundingCategory,
			AmountUsd:       numericFromFloat(item.AmountUSD),
		}); err != nil {
			return err
		}
	}
	for _, item := range req.LenderIndications {
		lenderID, err := model.ParseUUID(item.LenderID)
		if err != nil {
			return validation("lender_indications.lender_id", "UUID tidak valid")
		}
		if _, err := qtx.CreateLenderIndication(ctx, queries.CreateLenderIndicationParams{
			BbProjectID: projectID,
			LenderID:    lenderID,
			Remarks:     nullableTextPtr(item.Remarks),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *BlueBookService) validateNationalPriorities(ctx context.Context, qtx *queries.Queries, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		priorityID, err := model.ParseUUID(id)
		if err != nil {
			return validation("national_priority_ids", "UUID tidak valid")
		}
		if _, err := qtx.GetNationalPriority(ctx, priorityID); err != nil {
			return mapNotFound(err, "National Priority tidak ditemukan")
		}
	}
	return nil
}

func (s *BlueBookService) ensureBBCodeAvailable(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return validation("bb_code", "wajib diisi")
	}
	_, err := s.queries.GetBBProjectByCode(ctx, strings.TrimSpace(code))
	if err == nil {
		return apperrors.Conflict("BB Code sudah digunakan")
	}
	if err == pgx.ErrNoRows {
		return nil
	}
	return apperrors.Internal("Gagal memeriksa BB Code")
}

func (s *BlueBookService) buildBBProjectResponse(ctx context.Context, row queries.BbProject) (*model.BBProjectResponse, error) {
	institutions, err := s.queries.GetBBProjectInstitutions(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil institution BB Project")
	}
	locations, err := s.queries.GetBBProjectLocations(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil lokasi BB Project")
	}
	priorities, err := s.queries.GetBBProjectNationalPriorities(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil prioritas nasional BB Project")
	}
	costs, err := s.queries.GetBBProjectCosts(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil biaya BB Project")
	}
	lenders, err := s.queries.GetLenderIndications(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil lender indication BB Project")
	}

	res := model.BBProjectResponse{
		ID:                 model.UUIDToString(row.ID),
		BlueBookID:         model.UUIDToString(row.BlueBookID),
		ProgramTitleID:     stringPtrFromUUID(row.ProgramTitleID),
		BappenasPartnerID:  stringPtrFromUUID(row.BappenasPartnerID),
		BBCode:             row.BbCode,
		ProjectName:        row.ProjectName,
		Duration:           stringPtrFromText(row.Duration),
		Objective:          stringPtrFromText(row.Objective),
		ScopeOfWork:        stringPtrFromText(row.ScopeOfWork),
		Outputs:            stringPtrFromText(row.Outputs),
		Outcomes:           stringPtrFromText(row.Outcomes),
		Locations:          make([]model.RegionResponse, 0, len(locations)),
		NationalPriorities: make([]model.NationalPriorityResponse, 0, len(priorities)),
		ProjectCosts:       make([]model.ProjectCostResponse, 0, len(costs)),
		LenderIndications:  lenderIndicationResponses(lenders),
		Status:             row.Status,
		CreatedAt:          formatMasterTime(row.CreatedAt),
		UpdatedAt:          formatMasterTime(row.UpdatedAt),
	}
	for _, item := range institutions {
		institution := model.InstitutionResponse{ID: model.UUIDToString(item.ID), ParentID: stringPtrFromUUID(item.ParentID), Name: item.Name, ShortName: stringPtrFromText(item.ShortName), Level: item.Level, CreatedAt: formatMasterTime(item.CreatedAt), UpdatedAt: formatMasterTime(item.UpdatedAt)}
		if item.Role == roleExecutingAgency {
			res.ExecutingAgencies = append(res.ExecutingAgencies, institution)
		}
		if item.Role == roleImplementingAgency {
			res.ImplementingAgencies = append(res.ImplementingAgencies, institution)
		}
	}
	for _, item := range locations {
		res.Locations = append(res.Locations, toRegionResponse(item))
	}
	for _, item := range priorities {
		res.NationalPriorities = append(res.NationalPriorities, model.NationalPriorityResponse{ID: model.UUIDToString(item.ID), PeriodID: model.UUIDToString(item.PeriodID), Title: item.Title, CreatedAt: formatMasterTime(item.CreatedAt), UpdatedAt: formatMasterTime(item.UpdatedAt)})
	}
	for _, item := range costs {
		res.ProjectCosts = append(res.ProjectCosts, model.ProjectCostResponse{ID: model.UUIDToString(item.ID), FundingType: item.FundingType, FundingCategory: item.FundingCategory, AmountUSD: floatFromNumeric(item.AmountUsd)})
	}
	return &res, nil
}

func validateBBProjectRequest(req model.CreateBBProjectRequest, validateCode bool) error {
	if validateCode && strings.TrimSpace(req.BBCode) == "" {
		return validation("bb_code", "wajib diisi")
	}
	if strings.TrimSpace(req.ProjectName) == "" {
		return validation("project_name", "wajib diisi")
	}
	if len(req.ExecutingAgencyIDs) == 0 {
		return validation("executing_agency_ids", "minimal satu")
	}
	if len(req.ImplementingAgencyIDs) == 0 {
		return validation("implementing_agency_ids", "minimal satu")
	}
	if len(req.LocationIDs) == 0 {
		return validation("location_ids", "minimal satu")
	}
	return nil
}

func parseBlueBookRequest(req model.BlueBookRequest) (pgtype.UUID, pgtype.Date, pgtype.Int4, error) {
	if strings.TrimSpace(req.PeriodID) == "" {
		return pgtype.UUID{}, pgtype.Date{}, pgtype.Int4{}, validation("period_id", "wajib diisi")
	}
	periodID, err := model.ParseUUID(req.PeriodID)
	if err != nil {
		return pgtype.UUID{}, pgtype.Date{}, pgtype.Int4{}, validation("period_id", "UUID tidak valid")
	}
	publishDate, err := parseDate(req.PublishDate, "publish_date")
	if err != nil {
		return pgtype.UUID{}, pgtype.Date{}, pgtype.Int4{}, err
	}
	return periodID, publishDate, int4Ptr(req.RevisionYear), nil
}

func parseLoIRequest(req model.LoIRequest) (pgtype.UUID, pgtype.Date, error) {
	if strings.TrimSpace(req.Subject) == "" {
		return pgtype.UUID{}, pgtype.Date{}, validation("subject", "wajib diisi")
	}
	lenderID, err := model.ParseUUID(req.LenderID)
	if err != nil {
		return pgtype.UUID{}, pgtype.Date{}, validation("lender_id", "UUID tidak valid")
	}
	date, err := parseDate(req.Date, "date")
	if err != nil {
		return pgtype.UUID{}, pgtype.Date{}, err
	}
	return lenderID, date, nil
}

func parseDate(value, field string) (pgtype.Date, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return pgtype.Date{}, validation(field, "format harus YYYY-MM-DD")
	}
	return pgtype.Date{Time: parsed, Valid: true}, nil
}

func (s *BlueBookService) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.Internal("Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)
	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return apperrors.Internal("Gagal menyiapkan audit user")
	}
	if err := fn(s.queries.WithTx(tx)); err != nil {
		if _, ok := err.(*apperrors.AppError); ok {
			return err
		}
		return apperrors.FromPgError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return apperrors.Internal("Gagal menyimpan data")
	}
	return nil
}

func blueBookFromListRow(row queries.ListBlueBooksRow) model.BlueBookResponse {
	return model.BlueBookResponse{ID: model.UUIDToString(row.ID), Period: model.PeriodInfo{ID: model.UUIDToString(row.PeriodID), Name: row.PeriodName, YearStart: row.YearStart, YearEnd: row.YearEnd}, PublishDate: dateString(row.PublishDate), RevisionNumber: row.RevisionNumber, RevisionYear: int32PtrFromInt4(row.RevisionYear), Status: row.Status, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func blueBookFromGetRow(row queries.GetBlueBookRow) model.BlueBookResponse {
	return model.BlueBookResponse{ID: model.UUIDToString(row.ID), Period: model.PeriodInfo{ID: model.UUIDToString(row.PeriodID), Name: row.PeriodName, YearStart: row.YearStart, YearEnd: row.YearEnd}, PublishDate: dateString(row.PublishDate), RevisionNumber: row.RevisionNumber, RevisionYear: int32PtrFromInt4(row.RevisionYear), Status: row.Status, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func lenderIndicationResponses(rows []queries.GetLenderIndicationsRow) []model.LenderIndicationResponse {
	res := make([]model.LenderIndicationResponse, 0, len(rows))
	for _, row := range rows {
		res = append(res, model.LenderIndicationResponse{ID: model.UUIDToString(row.ID), Lender: model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType}, Remarks: stringPtrFromText(row.Remarks)})
	}
	return res
}

func loiResponses(rows []queries.GetLoIsByBBProjectRow) []model.LoIResponse {
	res := make([]model.LoIResponse, 0, len(rows))
	for _, row := range rows {
		res = append(res, model.LoIResponse{ID: model.UUIDToString(row.ID), BBProjectID: model.UUIDToString(row.BbProjectID), Lender: model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType}, Subject: row.Subject, Date: dateString(row.Date), LetterNumber: stringPtrFromText(row.LetterNumber), CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)})
	}
	return res
}

func uuidOrInvalid(value *string) pgtype.UUID {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.UUID{}
	}
	parsed, err := model.ParseUUID(*value)
	if err != nil {
		return pgtype.UUID{}
	}
	return parsed
}

func int4Ptr(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func int32PtrFromInt4(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

func dateString(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func numericFromFloat(value float64) pgtype.Numeric {
	var numeric pgtype.Numeric
	_ = numeric.Scan(strconv.FormatFloat(value, 'f', 2, 64))
	return numeric
}

func floatFromNumeric(value pgtype.Numeric) float64 {
	raw, err := value.Value()
	if err != nil || raw == nil {
		return 0
	}
	parsed, err := strconv.ParseFloat(fmt.Sprint(raw), 64)
	if err != nil {
		return 0
	}
	return parsed
}
