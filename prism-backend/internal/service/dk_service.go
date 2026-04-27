package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/sse"
)

type DKService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	broker  *sse.Broker
}

func NewDKService(db *pgxpool.Pool, queries *queries.Queries, broker *sse.Broker) *DKService {
	return &DKService{db: db, queries: queries, broker: broker}
}

func (s *DKService) ListDaftarKegiatan(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.DaftarKegiatanResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListDaftarKegiatan(ctx, queries.ListDaftarKegiatanParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar kegiatan")
	}
	total, err := s.queries.CountDaftarKegiatan(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung daftar kegiatan")
	}
	data := make([]model.DaftarKegiatanResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, daftarKegiatanResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *DKService) GetDaftarKegiatan(ctx context.Context, id pgtype.UUID) (*model.DaftarKegiatanResponse, error) {
	row, err := s.queries.GetDaftarKegiatan(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Daftar Kegiatan tidak ditemukan")
	}
	res := daftarKegiatanResponse(row)
	return &res, nil
}

func (s *DKService) CreateDaftarKegiatan(ctx context.Context, req model.DaftarKegiatanRequest) (*model.DaftarKegiatanResponse, error) {
	date, err := parseDKDate(req)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Subject) == "" {
		return nil, validation("subject", "wajib diisi")
	}
	var created queries.DaftarKegiatan
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateDaftarKegiatan(ctx, queries.CreateDaftarKegiatanParams{
			LetterNumber: nullableTextPtr(req.LetterNumber),
			Subject:      strings.TrimSpace(req.Subject),
			Date:         date,
		})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.Publish("daftar_kegiatan.created", map[string]string{"id": model.UUIDToString(created.ID)})
	}
	res := daftarKegiatanResponse(created)
	return &res, nil
}

func (s *DKService) UpdateDaftarKegiatan(ctx context.Context, id pgtype.UUID, req model.DaftarKegiatanRequest) (*model.DaftarKegiatanResponse, error) {
	date, err := parseDKDate(req)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Subject) == "" {
		return nil, validation("subject", "wajib diisi")
	}
	var updated queries.DaftarKegiatan
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateDaftarKegiatan(ctx, queries.UpdateDaftarKegiatanParams{
			ID:           id,
			LetterNumber: nullableTextPtr(req.LetterNumber),
			Subject:      strings.TrimSpace(req.Subject),
			Date:         date,
		})
		if err != nil {
			return mapNotFound(err, "Daftar Kegiatan tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	res := daftarKegiatanResponse(updated)
	return &res, nil
}

func (s *DKService) DeleteDaftarKegiatan(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return qtx.DeleteDaftarKegiatan(ctx, id)
	})
}

func (s *DKService) ListDKProjects(ctx context.Context, dkID pgtype.UUID, params model.PaginationParams) (*model.ListResponse[model.DKProjectResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListDKProjectsByDK(ctx, queries.ListDKProjectsByDKParams{DkID: dkID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil DK Project")
	}
	total, err := s.queries.CountDKProjectsByDK(ctx, dkID)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung DK Project")
	}
	data := make([]model.DKProjectResponse, 0, len(rows))
	for _, row := range rows {
		res, err := s.buildDKProjectResponse(ctx, row)
		if err != nil {
			return nil, err
		}
		data = append(data, *res)
	}
	return listResponse(data, page, limit, total), nil
}

func (s *DKService) GetDKProject(ctx context.Context, dkID, id pgtype.UUID) (*model.DKProjectResponse, error) {
	row, err := s.queries.GetDKProjectByDK(ctx, queries.GetDKProjectByDKParams{DkID: dkID, ID: id})
	if err != nil {
		return nil, mapNotFound(err, "DK Project tidak ditemukan")
	}
	return s.buildDKProjectResponse(ctx, row)
}

func (s *DKService) CreateDKProject(ctx context.Context, dkID pgtype.UUID, req model.CreateDKProjectRequest) (*model.DKProjectResponse, error) {
	if err := validateDKProjectRequest(req); err != nil {
		return nil, err
	}
	var created queries.DkProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetDaftarKegiatan(ctx, dkID); err != nil {
			return mapNotFound(err, "Daftar Kegiatan tidak ditemukan")
		}
		project, err := qtx.CreateDKProject(ctx, queries.CreateDKProjectParams{
			DkID:           dkID,
			ProgramTitleID: uuidOrInvalid(req.ProgramTitleID),
			InstitutionID:  uuidOrInvalid(req.InstitutionID),
			Duration:       nullableTextPtr(req.Duration),
			Objectives:     nullableTextPtr(req.Objectives),
		})
		if err != nil {
			return err
		}
		if err := s.replaceDKProjectChildren(ctx, qtx, project.ID, req); err != nil {
			return err
		}
		created = project
		return nil
	}); err != nil {
		return nil, err
	}
	return s.buildDKProjectResponse(ctx, created)
}

func (s *DKService) UpdateDKProject(ctx context.Context, dkID, id pgtype.UUID, req model.UpdateDKProjectRequest) (*model.DKProjectResponse, error) {
	if err := validateDKProjectRequest(req); err != nil {
		return nil, err
	}
	var updated queries.DkProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetDKProjectByDK(ctx, queries.GetDKProjectByDKParams{DkID: dkID, ID: id}); err != nil {
			return mapNotFound(err, "DK Project tidak ditemukan")
		}
		project, err := qtx.UpdateDKProject(ctx, queries.UpdateDKProjectParams{
			ID:             id,
			ProgramTitleID: uuidOrInvalid(req.ProgramTitleID),
			InstitutionID:  uuidOrInvalid(req.InstitutionID),
			Duration:       nullableTextPtr(req.Duration),
			Objectives:     nullableTextPtr(req.Objectives),
		})
		if err != nil {
			return err
		}
		if err := s.replaceDKProjectChildren(ctx, qtx, id, req); err != nil {
			return err
		}
		updated = project
		return nil
	}); err != nil {
		return nil, err
	}
	return s.buildDKProjectResponse(ctx, updated)
}

func (s *DKService) DeleteDKProject(ctx context.Context, dkID, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetDKProjectByDK(ctx, queries.GetDKProjectByDKParams{DkID: dkID, ID: id}); err != nil {
			return mapNotFound(err, "DK Project tidak ditemukan")
		}
		return qtx.DeleteDKProject(ctx, id)
	})
}

func (s *DKService) replaceDKProjectChildren(ctx context.Context, qtx *queries.Queries, projectID pgtype.UUID, req model.CreateDKProjectRequest) error {
	if err := qtx.DeleteDKActivityDetails(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteDKLoanAllocations(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteDKFinancingDetails(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteDKProjectLocations(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteDKProjectGBProjects(ctx, projectID); err != nil {
		return err
	}
	for _, id := range req.GBProjectIDs {
		gbProjectID, err := model.ParseUUID(id)
		if err != nil {
			return validation("gb_project_ids", "UUID tidak valid")
		}
		if err := qtx.AddDKProjectGBProject(ctx, queries.AddDKProjectGBProjectParams{DkProjectID: projectID, GbProjectID: gbProjectID}); err != nil {
			return err
		}
	}
	for _, id := range req.LocationIDs {
		regionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("location_ids", "UUID tidak valid")
		}
		if err := qtx.AddDKProjectLocation(ctx, queries.AddDKProjectLocationParams{DkProjectID: projectID, RegionID: regionID}); err != nil {
			return err
		}
	}
	allowed, err := qtx.GetAllowedLenderIDsForDK(ctx, projectID)
	if err != nil {
		return err
	}
	allowedSet := uuidSet(allowed)
	for _, item := range req.FinancingDetails {
		lenderID, err := parseOptionalUUID(item.LenderID, "financing_details.lender_id")
		if err != nil {
			return err
		}
		if lenderID.Valid {
			if _, ok := allowedSet[model.UUIDToString(lenderID)]; !ok {
				return apperrors.BusinessRule("Lender tidak terdaftar di GB atau BB terkait")
			}
		}
		if _, err := qtx.CreateDKFinancingDetail(ctx, queries.CreateDKFinancingDetailParams{
			DkProjectID:         projectID,
			LenderID:            lenderID,
			Currency:            normalizeCurrency(item.Currency),
			AmountOriginal:      numericFromFloat(item.AmountOriginal),
			GrantOriginal:       numericFromFloat(item.GrantOriginal),
			CounterpartOriginal: numericFromFloat(item.CounterpartOriginal),
			AmountUsd:           numericFromFloat(item.AmountUSD),
			GrantUsd:            numericFromFloat(item.GrantUSD),
			CounterpartUsd:      numericFromFloat(item.CounterpartUSD),
			Remarks:             nullableTextPtr(item.Remarks),
		}); err != nil {
			return err
		}
	}
	for _, item := range req.LoanAllocations {
		institutionID, err := parseOptionalUUID(item.InstitutionID, "loan_allocations.institution_id")
		if err != nil {
			return err
		}
		if _, err := qtx.CreateDKLoanAllocation(ctx, queries.CreateDKLoanAllocationParams{
			DkProjectID:         projectID,
			InstitutionID:       institutionID,
			Currency:            normalizeCurrency(item.Currency),
			AmountOriginal:      numericFromFloat(item.AmountOriginal),
			GrantOriginal:       numericFromFloat(item.GrantOriginal),
			CounterpartOriginal: numericFromFloat(item.CounterpartOriginal),
			AmountUsd:           numericFromFloat(item.AmountUSD),
			GrantUsd:            numericFromFloat(item.GrantUSD),
			CounterpartUsd:      numericFromFloat(item.CounterpartUSD),
			Remarks:             nullableTextPtr(item.Remarks),
		}); err != nil {
			return err
		}
	}
	for _, item := range req.ActivityDetails {
		if _, err := qtx.CreateDKActivityDetail(ctx, queries.CreateDKActivityDetailParams{
			DkProjectID:    projectID,
			ActivityNumber: item.ActivityNumber,
			ActivityName:   strings.TrimSpace(item.ActivityName),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *DKService) buildDKProjectResponse(ctx context.Context, row queries.DkProject) (*model.DKProjectResponse, error) {
	gbProjects, err := s.queries.GetDKProjectGBProjects(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil relasi GB Project")
	}
	locations, err := s.queries.GetDKProjectLocations(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil lokasi DK Project")
	}
	financingDetails, err := s.queries.GetDKFinancingDetails(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil financing detail")
	}
	loanAllocations, err := s.queries.GetDKLoanAllocations(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil loan allocation")
	}
	activityDetails, err := s.queries.GetDKActivityDetails(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil activity detail")
	}
	res := model.DKProjectResponse{
		ID:               model.UUIDToString(row.ID),
		DKID:             model.UUIDToString(row.DkID),
		ProgramTitleID:   stringPtrFromUUID(row.ProgramTitleID),
		InstitutionID:    stringPtrFromUUID(row.InstitutionID),
		Duration:         stringPtrFromText(row.Duration),
		Objectives:       stringPtrFromText(row.Objectives),
		GBProjects:       make([]model.GBProjectSummary, 0, len(gbProjects)),
		Locations:        make([]model.RegionResponse, 0, len(locations)),
		FinancingDetails: make([]model.DKFinancingDetailResponse, 0, len(financingDetails)),
		LoanAllocations:  make([]model.DKLoanAllocationResponse, 0, len(loanAllocations)),
		ActivityDetails:  make([]model.DKActivityDetailResponse, 0, len(activityDetails)),
		CreatedAt:        formatMasterTime(row.CreatedAt),
		UpdatedAt:        formatMasterTime(row.UpdatedAt),
	}
	for _, item := range gbProjects {
		res.GBProjects = append(res.GBProjects, model.GBProjectSummary{ID: model.UUIDToString(item.ID), GBCode: item.GbCode, ProjectName: item.ProjectName})
	}
	for _, item := range locations {
		res.Locations = append(res.Locations, toRegionResponse(item))
	}
	for _, item := range financingDetails {
		res.FinancingDetails = append(res.FinancingDetails, dkFinancingDetailResponse(item))
	}
	for _, item := range loanAllocations {
		res.LoanAllocations = append(res.LoanAllocations, dkLoanAllocationResponse(item))
	}
	for _, item := range activityDetails {
		res.ActivityDetails = append(res.ActivityDetails, model.DKActivityDetailResponse{ID: model.UUIDToString(item.ID), ActivityNumber: item.ActivityNumber, ActivityName: item.ActivityName})
	}
	return &res, nil
}

func (s *DKService) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
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

func validateDKProjectRequest(req model.CreateDKProjectRequest) error {
	if len(req.GBProjectIDs) == 0 {
		return validation("gb_project_ids", "Minimal 1 GB Project")
	}
	for _, item := range req.ActivityDetails {
		if item.ActivityNumber <= 0 {
			return validation("activity_details.activity_number", "harus lebih dari 0")
		}
		if strings.TrimSpace(item.ActivityName) == "" {
			return validation("activity_details.activity_name", "wajib diisi")
		}
	}
	return nil
}

func parseDKDate(req model.DaftarKegiatanRequest) (pgtype.Date, error) {
	value := req.Date
	if strings.TrimSpace(value) == "" {
		value = req.Tanggal
	}
	if strings.TrimSpace(value) == "" {
		return pgtype.Date{}, validation("date", "wajib diisi")
	}
	return parseDate(value, "date")
}

func daftarKegiatanResponse(row queries.DaftarKegiatan) model.DaftarKegiatanResponse {
	return model.DaftarKegiatanResponse{ID: model.UUIDToString(row.ID), LetterNumber: stringPtrFromText(row.LetterNumber), Subject: row.Subject, Date: dateString(row.Date), CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func dkFinancingDetailResponse(row queries.GetDKFinancingDetailsRow) model.DKFinancingDetailResponse {
	res := model.DKFinancingDetailResponse{ID: model.UUIDToString(row.ID), Currency: row.Currency, AmountOriginal: floatFromNumeric(row.AmountOriginal), GrantOriginal: floatFromNumeric(row.GrantOriginal), CounterpartOriginal: floatFromNumeric(row.CounterpartOriginal), AmountUSD: floatFromNumeric(row.AmountUsd), GrantUSD: floatFromNumeric(row.GrantUsd), CounterpartUSD: floatFromNumeric(row.CounterpartUsd), Remarks: stringPtrFromText(row.Remarks)}
	if row.LenderID.Valid {
		res.Lender = &model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName.String, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType.String}
	}
	return res
}

func dkLoanAllocationResponse(row queries.GetDKLoanAllocationsRow) model.DKLoanAllocationResponse {
	res := model.DKLoanAllocationResponse{ID: model.UUIDToString(row.ID), Currency: row.Currency, AmountOriginal: floatFromNumeric(row.AmountOriginal), GrantOriginal: floatFromNumeric(row.GrantOriginal), CounterpartOriginal: floatFromNumeric(row.CounterpartOriginal), AmountUSD: floatFromNumeric(row.AmountUsd), GrantUSD: floatFromNumeric(row.GrantUsd), CounterpartUSD: floatFromNumeric(row.CounterpartUsd), Remarks: stringPtrFromText(row.Remarks)}
	if row.InstitutionID.Valid {
		res.Institution = &model.InstitutionInfo{ID: model.UUIDToString(row.InstitutionID), Name: row.InstitutionName.String, ShortName: stringPtrFromText(row.InstitutionShortName), Level: row.InstitutionLevel.String}
	}
	return res
}

func uuidSet(ids []pgtype.UUID) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id.Valid {
			set[model.UUIDToString(id)] = struct{}{}
		}
	}
	return set
}

func normalizeCurrency(value string) string {
	if strings.TrimSpace(value) == "" {
		return "USD"
	}
	return strings.ToUpper(strings.TrimSpace(value))
}
