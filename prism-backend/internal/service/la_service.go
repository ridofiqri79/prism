package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/sse"
)

type LAService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	broker  *sse.Broker
}

func NewLAService(db *pgxpool.Pool, queries *queries.Queries, broker *sse.Broker) *LAService {
	return &LAService{db: db, queries: queries, broker: broker}
}

func (s *LAService) ListLoanAgreements(ctx context.Context, filter model.LoanAgreementListFilter, params model.PaginationParams) (*model.ListResponse[model.LoanAgreementResponse], error) {
	page, limit, offset := normalizeList(params)
	queryParams, err := buildLoanAgreementListParams(filter, params, limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListLoanAgreements(ctx, queryParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil Loan Agreement")
	}
	total, err := s.queries.CountLoanAgreements(ctx, queries.CountLoanAgreementsParams{
		PeriodIds:         queryParams.PeriodIds,
		Search:            queryParams.Search,
		LenderID:          queryParams.LenderID,
		IsExtended:        queryParams.IsExtended,
		ClosingDateBefore: queryParams.ClosingDateBefore,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung Loan Agreement")
	}
	data := make([]model.LoanAgreementResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, laListResponse(row))
	}
	data, err = s.attachLoanAgreementDKProjects(ctx, data)
	if err != nil {
		return nil, err
	}
	return listResponse(data, page, limit, total), nil
}

func buildLoanAgreementListParams(filter model.LoanAgreementListFilter, params model.PaginationParams, limit, offset int) (queries.ListLoanAgreementsParams, error) {
	periodIDs, err := parseUUIDList(filter.PeriodIDs, "period_ids")
	if err != nil {
		return queries.ListLoanAgreementsParams{}, err
	}

	lenderID, err := parseOptionalUUID(filter.LenderID, "lender_id")
	if err != nil {
		return queries.ListLoanAgreementsParams{}, err
	}
	isExtended := pgtype.Bool{}
	if filter.IsExtended != nil && strings.TrimSpace(*filter.IsExtended) != "" {
		switch strings.ToLower(strings.TrimSpace(*filter.IsExtended)) {
		case "true":
			isExtended = pgtype.Bool{Bool: true, Valid: true}
		case "false":
			isExtended = pgtype.Bool{Bool: false, Valid: true}
		default:
			return queries.ListLoanAgreementsParams{}, validation("is_extended", "harus true atau false")
		}
	}
	closingDateBefore, err := optionalDate(filter.ClosingDateBefore, "closing_date_before")
	if err != nil {
		return queries.ListLoanAgreementsParams{}, err
	}
	sortField, sortOrder, err := normalizeListSort(params.Sort, params.Order, "created_at", "desc", map[string]struct{}{
		"loan_code": {}, "lender": {}, "effective_date": {}, "closing_date": {}, "currency": {}, "amount_usd": {}, "cumulative_disbursement_usd": {}, "disbursement_ratio": {}, "estimated_time_ratio": {}, "performance_value": {}, "performance_status": {}, "status": {}, "created_at": {},
	})
	if err != nil {
		return queries.ListLoanAgreementsParams{}, err
	}
	return queries.ListLoanAgreementsParams{
		PeriodIds:         periodIDs,
		Search:            nullableText(params.Search),
		LenderID:          lenderID,
		IsExtended:        isExtended,
		ClosingDateBefore: closingDateBefore,
		SortField:         sortField,
		SortOrder:         sortOrder,
		Limit:             int32(limit),
		Offset:            int32(offset),
	}, nil
}

func (s *LAService) GetLoanAgreement(ctx context.Context, id pgtype.UUID) (*model.LoanAgreementResponse, error) {
	row, err := s.queries.GetLoanAgreement(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Loan Agreement tidak ditemukan")
	}
	res := laGetResponse(row)
	attached, err := s.attachLoanAgreementDKProjects(ctx, []model.LoanAgreementResponse{res})
	if err != nil {
		return nil, err
	}
	if len(attached) == 1 {
		res = attached[0]
	}
	return &res, nil
}

func (s *LAService) CreateLoanAgreement(ctx context.Context, req model.LoanAgreementRequest) (*model.LoanAgreementResponse, error) {
	parsed, err := parseLoanAgreementRequest(req)
	if err != nil {
		return nil, err
	}
	var created queries.LoanAgreement
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if err := validateLADKProjects(ctx, qtx, parsed.DKProjectIDs()); err != nil {
			return err
		}
		if err := validateLALenderForProjects(ctx, qtx, parsed.DKProjectIDs(), parsed.LenderID); err != nil {
			return err
		}
		if err := validateActiveCurrency(ctx, qtx, "currency", parsed.Currency); err != nil {
			return err
		}
		amountUSD, err := calculateLoanAgreementAmountUSD(ctx, qtx, parsed.Currency, parsed.AmountOriginalValue)
		if err != nil {
			return err
		}
		allocationRows, err := buildLoanAgreementAllocationRows(ctx, qtx, parsed, amountUSD)
		if err != nil {
			return err
		}
		row, err := qtx.CreateLoanAgreement(ctx, queries.CreateLoanAgreementParams{
			DkProjectID:            parsed.PrimaryDKProjectID(),
			LenderID:               parsed.LenderID,
			LoanCode:               parsed.LoanCode,
			AgreementDate:          parsed.AgreementDate,
			EffectiveDate:          parsed.EffectiveDate,
			OriginalClosingDate:    parsed.OriginalClosingDate,
			ClosingDate:            parsed.ClosingDate,
			Currency:               parsed.Currency,
			AmountOriginal:         parsed.AmountOriginal,
			AmountUsd:              numericFromFloat(amountUSD),
			CumulativeDisbursement: parsed.CumulativeDisbursement,
		})
		if err != nil {
			return err
		}
		for _, allocation := range allocationRows {
			if err := qtx.AddLoanAgreementDKProject(ctx, queries.AddLoanAgreementDKProjectParams{
				LoanAgreementID:    row.ID,
				DkProjectID:        allocation.DKProjectID,
				AllocationOriginal: numericFromFloat(allocation.AllocationOriginal),
				AllocationUsd:      numericFromFloat(allocation.AllocationUSD),
			}); err != nil {
				return err
			}
		}
		created = row
		return nil
	}); err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.Publish("loan_agreement.created", map[string]string{"id": model.UUIDToString(created.ID)})
	}
	return s.GetLoanAgreement(ctx, created.ID)
}

func (s *LAService) UpdateLoanAgreement(ctx context.Context, id pgtype.UUID, req model.LoanAgreementRequest) (*model.LoanAgreementResponse, error) {
	parsed, err := parseLoanAgreementRequest(req)
	if err != nil {
		return nil, err
	}
	var updated queries.LoanAgreement
	var publishExtended bool
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		current, err := qtx.GetLoanAgreement(ctx, id)
		if err != nil {
			return mapNotFound(err, "Loan Agreement tidak ditemukan")
		}
		if err := validateLADKProjects(ctx, qtx, parsed.DKProjectIDs()); err != nil {
			return err
		}
		if err := validateLALenderForProjects(ctx, qtx, parsed.DKProjectIDs(), parsed.LenderID); err != nil {
			return err
		}
		if err := validateActiveCurrency(ctx, qtx, "currency", parsed.Currency); err != nil {
			return err
		}
		amountUSD, err := calculateLoanAgreementAmountUSD(ctx, qtx, parsed.Currency, parsed.AmountOriginalValue)
		if err != nil {
			return err
		}
		allocationRows, err := buildLoanAgreementAllocationRows(ctx, qtx, parsed, amountUSD)
		if err != nil {
			return err
		}
		row, err := qtx.UpdateLoanAgreement(ctx, queries.UpdateLoanAgreementParams{
			ID:                     id,
			DkProjectID:            parsed.PrimaryDKProjectID(),
			LenderID:               parsed.LenderID,
			LoanCode:               parsed.LoanCode,
			AgreementDate:          parsed.AgreementDate,
			EffectiveDate:          parsed.EffectiveDate,
			OriginalClosingDate:    parsed.OriginalClosingDate,
			ClosingDate:            parsed.ClosingDate,
			Currency:               parsed.Currency,
			AmountOriginal:         parsed.AmountOriginal,
			AmountUsd:              numericFromFloat(amountUSD),
			CumulativeDisbursement: parsed.CumulativeDisbursement,
		})
		if err != nil {
			return err
		}
		if err := qtx.DeleteLoanAgreementDKProjects(ctx, id); err != nil {
			return err
		}
		for _, allocation := range allocationRows {
			if err := qtx.AddLoanAgreementDKProject(ctx, queries.AddLoanAgreementDKProjectParams{
				LoanAgreementID:    id,
				DkProjectID:        allocation.DKProjectID,
				AllocationOriginal: numericFromFloat(allocation.AllocationOriginal),
				AllocationUsd:      numericFromFloat(allocation.AllocationUSD),
			}); err != nil {
				return err
			}
		}
		updated = row
		publishExtended = !sameDate(current.ClosingDate, row.ClosingDate) && isExtended(row.OriginalClosingDate, row.ClosingDate)
		return nil
	}); err != nil {
		return nil, err
	}
	if publishExtended && s.broker != nil {
		s.broker.Publish("loan_agreement.extended", map[string]string{"id": model.UUIDToString(updated.ID)})
	}
	return s.GetLoanAgreement(ctx, updated.ID)
}

func (s *LAService) DeleteLoanAgreement(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		return qtx.DeleteLoanAgreement(ctx, id)
	})
}

func (s *LAService) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
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

type parsedLoanAgreementRequest struct {
	Allocations            []parsedLoanAgreementAllocation
	LenderID               pgtype.UUID
	LoanCode               string
	AgreementDate          pgtype.Date
	EffectiveDate          pgtype.Date
	OriginalClosingDate    pgtype.Date
	ClosingDate            pgtype.Date
	Currency               string
	AmountOriginalValue    float64
	AmountOriginal         pgtype.Numeric
	CumulativeDisbursement pgtype.Numeric
}

type parsedLoanAgreementAllocation struct {
	DKProjectID        pgtype.UUID
	AllocationOriginal float64
}

type loanAgreementAllocationRow struct {
	DKProjectID        pgtype.UUID
	AllocationOriginal float64
	AllocationUSD      float64
}

func (p parsedLoanAgreementRequest) DKProjectIDs() []pgtype.UUID {
	ids := make([]pgtype.UUID, 0, len(p.Allocations))
	for _, allocation := range p.Allocations {
		ids = append(ids, allocation.DKProjectID)
	}
	return ids
}

func (p parsedLoanAgreementRequest) PrimaryDKProjectID() pgtype.UUID {
	if len(p.Allocations) == 0 {
		return pgtype.UUID{}
	}
	return p.Allocations[0].DKProjectID
}

func parseLoanAgreementRequest(req model.LoanAgreementRequest) (parsedLoanAgreementRequest, error) {
	lenderID, err := model.ParseUUID(req.LenderID)
	if err != nil {
		return parsedLoanAgreementRequest{}, validation("lender_id", "UUID tidak valid")
	}
	agreementDate, err := parseDate(req.AgreementDate, "agreement_date")
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	effectiveDate, err := parseDate(req.EffectiveDate, "effective_date")
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	originalClosingDate, err := parseOptionalLoanAgreementDate(req.OriginalClosingDate, "original_closing_date")
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	closingDate, err := parseDate(req.ClosingDate, "closing_date")
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	if originalClosingDate.Valid && closingDate.Time.Before(originalClosingDate.Time) {
		return parsedLoanAgreementRequest{}, validation("closing_date", "tidak boleh lebih awal dari tanggal closing awal")
	}
	if strings.TrimSpace(req.LoanCode) == "" {
		return parsedLoanAgreementRequest{}, validation("loan_code", "wajib diisi")
	}
	if strings.TrimSpace(req.Currency) == "" {
		return parsedLoanAgreementRequest{}, validation("currency", "wajib diisi")
	}
	currency := normalizeCurrency(req.Currency)
	if req.AmountOriginal <= 0 {
		return parsedLoanAgreementRequest{}, validation("amount_original", "wajib lebih dari 0")
	}
	if req.CumulativeDisbursement < 0 {
		return parsedLoanAgreementRequest{}, validation("cumulative_disbursement", "tidak boleh negatif")
	}
	allocations, err := parseLoanAgreementAllocations(req)
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	if !sameCurrencyAmount(totalLoanAgreementAllocation(allocations), req.AmountOriginal) {
		return parsedLoanAgreementRequest{}, validation("dk_project_allocations", "total alokasi harus sama dengan nilai Loan Agreement")
	}
	return parsedLoanAgreementRequest{
		Allocations:            allocations,
		LenderID:               lenderID,
		LoanCode:               strings.TrimSpace(req.LoanCode),
		AgreementDate:          agreementDate,
		EffectiveDate:          effectiveDate,
		OriginalClosingDate:    originalClosingDate,
		ClosingDate:            closingDate,
		Currency:               currency,
		AmountOriginalValue:    req.AmountOriginal,
		AmountOriginal:         numericFromFloat(req.AmountOriginal),
		CumulativeDisbursement: numericFromFloat(req.CumulativeDisbursement),
	}, nil
}

func parseLoanAgreementAllocations(req model.LoanAgreementRequest) ([]parsedLoanAgreementAllocation, error) {
	items := req.DKProjectAllocations
	if len(items) == 0 && strings.TrimSpace(req.DKProjectID) != "" {
		items = []model.LoanAgreementDKProjectAllocation{{
			DKProjectID:        req.DKProjectID,
			AllocationOriginal: req.AmountOriginal,
		}}
	}
	if len(items) == 0 {
		return nil, validation("dk_project_allocations", "minimal satu Proyek Daftar Kegiatan")
	}
	seen := map[string]struct{}{}
	allocations := make([]parsedLoanAgreementAllocation, 0, len(items))
	for idx, item := range items {
		field := fmt.Sprintf("dk_project_allocations.%d.dk_project_id", idx)
		dkProjectID, err := model.ParseUUID(item.DKProjectID)
		if err != nil {
			return nil, validation(field, "UUID tidak valid")
		}
		key := model.UUIDToString(dkProjectID)
		if _, exists := seen[key]; exists {
			return nil, validation("dk_project_allocations", "Proyek Daftar Kegiatan tidak boleh duplikat")
		}
		if item.AllocationOriginal <= 0 {
			return nil, validation(fmt.Sprintf("dk_project_allocations.%d.allocation_original", idx), "wajib lebih dari 0")
		}
		seen[key] = struct{}{}
		allocations = append(allocations, parsedLoanAgreementAllocation{
			DKProjectID:        dkProjectID,
			AllocationOriginal: item.AllocationOriginal,
		})
	}
	return allocations, nil
}

func totalLoanAgreementAllocation(allocations []parsedLoanAgreementAllocation) float64 {
	total := 0.0
	for _, allocation := range allocations {
		total += allocation.AllocationOriginal
	}
	return roundCurrencyValue(total)
}

func roundCurrencyValue(value float64) float64 {
	return math.Round(value*100) / 100
}

func sameCurrencyAmount(left, right float64) bool {
	return math.Abs(roundCurrencyValue(left)-roundCurrencyValue(right)) < 0.005
}

func parseOptionalLoanAgreementDate(value string, field string) (pgtype.Date, error) {
	if strings.TrimSpace(value) == "" {
		return pgtype.Date{}, nil
	}
	return parseDate(value, field)
}

func validateLADKProjects(ctx context.Context, qtx *queries.Queries, dkProjectIDs []pgtype.UUID) error {
	missing, err := qtx.CountMissingDKProjectsForLA(ctx, dkProjectIDs)
	if err != nil {
		return err
	}
	if missing > 0 {
		return validation("dk_project_allocations", "memuat Proyek Daftar Kegiatan yang tidak ditemukan")
	}
	return nil
}

func validateLALenderForProjects(ctx context.Context, qtx *queries.Queries, dkProjectIDs []pgtype.UUID, lenderID pgtype.UUID) error {
	allowed, err := qtx.GetAllowedLenderIDsForLAProjects(ctx, dkProjectIDs)
	if err != nil {
		return err
	}
	if _, ok := uuidSet(allowed)[model.UUIDToString(lenderID)]; !ok {
		return apperrors.BusinessRule("Lender harus berasal dari Financing Detail semua Proyek Daftar Kegiatan terkait")
	}
	return nil
}

func buildLoanAgreementAllocationRows(ctx context.Context, qtx *queries.Queries, parsed parsedLoanAgreementRequest, totalAmountUSD float64) ([]loanAgreementAllocationRow, error) {
	rows := make([]loanAgreementAllocationRow, 0, len(parsed.Allocations))
	totalAmountUSD = roundCurrencyValue(totalAmountUSD)
	usdTotalSoFar := 0.0
	for idx, allocation := range parsed.Allocations {
		allocationUSD := 0.0
		if idx == len(parsed.Allocations)-1 {
			allocationUSD = roundCurrencyValue(totalAmountUSD - usdTotalSoFar)
		} else {
			calculated, err := calculateLoanAgreementAmountUSD(ctx, qtx, parsed.Currency, allocation.AllocationOriginal)
			if err != nil {
				return nil, err
			}
			allocationUSD = roundCurrencyValue(calculated)
			usdTotalSoFar += allocationUSD
		}
		if allocationUSD < 0 {
			return nil, validation("dk_project_allocations", "alokasi USD tidak valid")
		}
		rows = append(rows, loanAgreementAllocationRow{
			DKProjectID:        allocation.DKProjectID,
			AllocationOriginal: allocation.AllocationOriginal,
			AllocationUSD:      allocationUSD,
		})
	}
	return rows, nil
}

func calculateLoanAgreementAmountUSD(ctx context.Context, qtx *queries.Queries, currency string, amountOriginal float64) (float64, error) {
	if normalizeCurrency(currency) == "USD" {
		return amountOriginal, nil
	}
	kurs, err := qtx.GetLatestKursTengahByCurrencyCode(ctx, normalizeCurrency(currency))
	if err == pgx.ErrNoRows {
		return 0, validation("currency", fmt.Sprintf("Kurs Tengah BI untuk %s belum tersedia", normalizeCurrency(currency)))
	}
	if err != nil {
		return 0, apperrors.Internal("Gagal membaca Kurs Tengah BI")
	}
	rate := floatFromNumeric(kurs.KursTengahBi)
	if rate <= 0 {
		return 0, validation("currency", fmt.Sprintf("Kurs Tengah BI untuk %s tidak valid", normalizeCurrency(currency)))
	}
	return amountOriginal / rate, nil
}

func laGetResponse(row queries.GetLoanAgreementRow) model.LoanAgreementResponse {
	kursCutOffDate := dateStringPtr(row.KursCutOffDate)
	return model.LoanAgreementResponse{
		ID:                        model.UUIDToString(row.ID),
		DKProjectID:               model.UUIDToString(row.DkProjectID),
		DKProjects:                []model.LoanAgreementDKProjectResponse{},
		Lender:                    model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
		LoanCode:                  row.LoanCode,
		AgreementDate:             dateString(row.AgreementDate),
		EffectiveDate:             dateString(row.EffectiveDate),
		OriginalClosingDate:       dateString(row.OriginalClosingDate),
		ClosingDate:               dateString(row.ClosingDate),
		IsExtended:                isExtended(row.OriginalClosingDate, row.ClosingDate),
		ExtensionDays:             extensionDays(row.OriginalClosingDate, row.ClosingDate),
		Currency:                  row.Currency,
		AmountOriginal:            floatFromNumeric(row.AmountOriginal),
		AmountUSD:                 floatFromNumeric(row.AmountUsd),
		CumulativeDisbursement:    floatFromNumeric(row.CumulativeDisbursement),
		CumulativeDisbursementUSD: floatPtrFromNumeric(row.CumulativeDisbursementUsd),
		DisbursementRatio:         floatPtrFromNumeric(row.DisbursementRatio),
		EstimatedTimeRatio:        floatPtrFromNumeric(row.EstimatedTimeRatio),
		PerformanceValue:          floatPtrFromNumeric(row.PerformanceValue),
		PerformanceStatus:         stringPtrFromAny(row.PerformanceStatus),
		KursTengahBI:              floatPtrFromNumeric(row.KursTengahBi),
		KursCutOffDate:            kursCutOffDate,
		CreatedAt:                 formatMasterTime(row.CreatedAt),
		UpdatedAt:                 formatMasterTime(row.UpdatedAt),
	}
}

func laListResponse(row queries.ListLoanAgreementsRow) model.LoanAgreementResponse {
	kursCutOffDate := dateStringPtr(row.KursCutOffDate)
	return model.LoanAgreementResponse{
		ID:                        model.UUIDToString(row.ID),
		DKProjectID:               model.UUIDToString(row.DkProjectID),
		DKProjects:                []model.LoanAgreementDKProjectResponse{},
		Lender:                    model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
		LoanCode:                  row.LoanCode,
		AgreementDate:             dateString(row.AgreementDate),
		EffectiveDate:             dateString(row.EffectiveDate),
		OriginalClosingDate:       dateString(row.OriginalClosingDate),
		ClosingDate:               dateString(row.ClosingDate),
		IsExtended:                isExtended(row.OriginalClosingDate, row.ClosingDate),
		ExtensionDays:             extensionDays(row.OriginalClosingDate, row.ClosingDate),
		Currency:                  row.Currency,
		AmountOriginal:            floatFromNumeric(row.AmountOriginal),
		AmountUSD:                 floatFromNumeric(row.AmountUsd),
		CumulativeDisbursement:    floatFromNumeric(row.CumulativeDisbursement),
		CumulativeDisbursementUSD: floatPtrFromNumeric(row.CumulativeDisbursementUsd),
		DisbursementRatio:         floatPtrFromNumeric(row.DisbursementRatio),
		EstimatedTimeRatio:        floatPtrFromNumeric(row.EstimatedTimeRatio),
		PerformanceValue:          floatPtrFromNumeric(row.PerformanceValue),
		PerformanceStatus:         stringPtrFromAny(row.PerformanceStatus),
		KursTengahBI:              floatPtrFromNumeric(row.KursTengahBi),
		KursCutOffDate:            kursCutOffDate,
		CreatedAt:                 formatMasterTime(row.CreatedAt),
		UpdatedAt:                 formatMasterTime(row.UpdatedAt),
	}
}

func (s *LAService) attachLoanAgreementDKProjects(ctx context.Context, items []model.LoanAgreementResponse) ([]model.LoanAgreementResponse, error) {
	if len(items) == 0 {
		return items, nil
	}
	ids := make([]pgtype.UUID, 0, len(items))
	indexByID := map[string]int{}
	for idx, item := range items {
		id, err := model.ParseUUID(item.ID)
		if err != nil {
			return nil, validation("id", "UUID tidak valid")
		}
		ids = append(ids, id)
		indexByID[item.ID] = idx
	}
	rows, err := s.queries.ListLoanAgreementDKProjectsByLoanAgreements(ctx, ids)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil relasi Proyek Daftar Kegiatan Loan Agreement")
	}
	for _, row := range rows {
		loanAgreementID := model.UUIDToString(row.LoanAgreementID)
		idx, ok := indexByID[loanAgreementID]
		if !ok {
			continue
		}
		items[idx].DKProjects = append(items[idx].DKProjects, model.LoanAgreementDKProjectResponse{
			ID:          model.UUIDToString(row.DkProjectID),
			DKID:        model.UUIDToString(row.DkID),
			ProjectName: row.ProjectName,
			Objectives:  stringPtrFromText(row.Objectives),
			GBCodes:     row.GbCodes,
			DaftarKegiatan: model.LoanAgreementDKHeaderInfo{
				ID:           model.UUIDToString(row.DkID),
				Subject:      row.DkSubject,
				Date:         dateString(row.DkDate),
				LetterNumber: stringPtrFromText(row.DkLetterNumber),
			},
			AllocationOriginal: floatFromNumeric(row.AllocationOriginal),
			AllocationUSD:      floatFromNumeric(row.AllocationUsd),
		})
	}
	return items, nil
}

func floatPtrFromNumeric(value pgtype.Numeric) *float64 {
	if !value.Valid {
		return nil
	}
	result := floatFromNumeric(value)
	return &result
}

func dateStringPtr(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	result := dateString(value)
	return &result
}

func stringPtrFromAny(value interface{}) *string {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		return stringPtrFromValue(typed)
	case []byte:
		return stringPtrFromValue(string(typed))
	case pgtype.Text:
		return stringPtrFromText(typed)
	default:
		return stringPtrFromValue(fmt.Sprint(typed))
	}
}

func stringPtrFromValue(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func isExtended(original, closing pgtype.Date) bool {
	return original.Valid && closing.Valid && !sameDate(original, closing)
}

func extensionDays(original, closing pgtype.Date) int {
	if !original.Valid || !closing.Valid {
		return 0
	}
	return int(closing.Time.Sub(original.Time).Hours() / 24)
}

func sameDate(a, b pgtype.Date) bool {
	if !a.Valid || !b.Valid {
		return a.Valid == b.Valid
	}
	return a.Time.Equal(b.Time)
}
