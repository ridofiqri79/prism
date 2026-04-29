package service

import (
	"context"
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

func (s *LAService) ListLoanAgreements(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.LoanAgreementResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListLoanAgreements(ctx, queries.ListLoanAgreementsParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil Loan Agreement")
	}
	total, err := s.queries.CountLoanAgreements(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung Loan Agreement")
	}
	data := make([]model.LoanAgreementResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, laListResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *LAService) GetLoanAgreement(ctx context.Context, id pgtype.UUID) (*model.LoanAgreementResponse, error) {
	row, err := s.queries.GetLoanAgreement(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Loan Agreement tidak ditemukan")
	}
	res := laGetResponse(row)
	return &res, nil
}

func (s *LAService) CreateLoanAgreement(ctx context.Context, req model.LoanAgreementRequest) (*model.LoanAgreementResponse, error) {
	parsed, err := parseLoanAgreementRequest(req)
	if err != nil {
		return nil, err
	}
	var created queries.LoanAgreement
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetLoanAgreementByDKProject(ctx, parsed.DKProjectID); err == nil {
			return apperrors.Conflict("DK Project sudah memiliki Loan Agreement")
		} else if err != pgx.ErrNoRows {
			return err
		}
		if err := validateLALender(ctx, qtx, parsed.DKProjectID, parsed.LenderID); err != nil {
			return err
		}
		if err := validateActiveCurrency(ctx, qtx, "currency", parsed.Currency); err != nil {
			return err
		}
		row, err := qtx.CreateLoanAgreement(ctx, queries.CreateLoanAgreementParams{
			DkProjectID:         parsed.DKProjectID,
			LenderID:            parsed.LenderID,
			LoanCode:            parsed.LoanCode,
			AgreementDate:       parsed.AgreementDate,
			EffectiveDate:       parsed.EffectiveDate,
			OriginalClosingDate: parsed.OriginalClosingDate,
			ClosingDate:         parsed.ClosingDate,
			Currency:            parsed.Currency,
			AmountOriginal:      parsed.AmountOriginal,
			AmountUsd:           parsed.AmountUsd,
		})
		created = row
		return err
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
		if err := validateLALender(ctx, qtx, current.DkProjectID, parsed.LenderID); err != nil {
			return err
		}
		if err := validateActiveCurrency(ctx, qtx, "currency", parsed.Currency); err != nil {
			return err
		}
		row, err := qtx.UpdateLoanAgreement(ctx, queries.UpdateLoanAgreementParams{
			ID:                  id,
			LenderID:            parsed.LenderID,
			LoanCode:            parsed.LoanCode,
			AgreementDate:       parsed.AgreementDate,
			EffectiveDate:       parsed.EffectiveDate,
			OriginalClosingDate: parsed.OriginalClosingDate,
			ClosingDate:         parsed.ClosingDate,
			Currency:            parsed.Currency,
			AmountOriginal:      parsed.AmountOriginal,
			AmountUsd:           parsed.AmountUsd,
		})
		if err != nil {
			return err
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
	DKProjectID         pgtype.UUID
	LenderID            pgtype.UUID
	LoanCode            string
	AgreementDate       pgtype.Date
	EffectiveDate       pgtype.Date
	OriginalClosingDate pgtype.Date
	ClosingDate         pgtype.Date
	Currency            string
	AmountOriginal      pgtype.Numeric
	AmountUsd           pgtype.Numeric
}

func parseLoanAgreementRequest(req model.LoanAgreementRequest) (parsedLoanAgreementRequest, error) {
	if strings.TrimSpace(req.DKProjectID) == "" {
		return parsedLoanAgreementRequest{}, validation("dk_project_id", "wajib diisi")
	}
	dkProjectID, err := model.ParseUUID(req.DKProjectID)
	if err != nil {
		return parsedLoanAgreementRequest{}, validation("dk_project_id", "UUID tidak valid")
	}
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
	originalClosingDate, err := parseDate(req.OriginalClosingDate, "original_closing_date")
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	closingDate, err := parseDate(req.ClosingDate, "closing_date")
	if err != nil {
		return parsedLoanAgreementRequest{}, err
	}
	if strings.TrimSpace(req.LoanCode) == "" {
		return parsedLoanAgreementRequest{}, validation("loan_code", "wajib diisi")
	}
	if strings.TrimSpace(req.Currency) == "" {
		return parsedLoanAgreementRequest{}, validation("currency", "wajib diisi")
	}
	currency := normalizeCurrency(req.Currency)
	amountOriginal, amountUSD := normalizeCurrencyAmountPair(currency, req.AmountOriginal, req.AmountUSD)
	return parsedLoanAgreementRequest{
		DKProjectID:         dkProjectID,
		LenderID:            lenderID,
		LoanCode:            strings.TrimSpace(req.LoanCode),
		AgreementDate:       agreementDate,
		EffectiveDate:       effectiveDate,
		OriginalClosingDate: originalClosingDate,
		ClosingDate:         closingDate,
		Currency:            currency,
		AmountOriginal:      numericFromFloat(amountOriginal),
		AmountUsd:           numericFromFloat(amountUSD),
	}, nil
}

func validateLALender(ctx context.Context, qtx *queries.Queries, dkProjectID, lenderID pgtype.UUID) error {
	allowed, err := qtx.GetAllowedLenderIDsForLA(ctx, dkProjectID)
	if err != nil {
		return err
	}
	if _, ok := uuidSet(allowed)[model.UUIDToString(lenderID)]; !ok {
		return apperrors.BusinessRule("Lender harus berasal dari Financing Detail DK Project terkait")
	}
	return nil
}

func laGetResponse(row queries.GetLoanAgreementRow) model.LoanAgreementResponse {
	return model.LoanAgreementResponse{
		ID:                  model.UUIDToString(row.ID),
		DKProjectID:         model.UUIDToString(row.DkProjectID),
		Lender:              model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
		LoanCode:            row.LoanCode,
		AgreementDate:       dateString(row.AgreementDate),
		EffectiveDate:       dateString(row.EffectiveDate),
		OriginalClosingDate: dateString(row.OriginalClosingDate),
		ClosingDate:         dateString(row.ClosingDate),
		IsExtended:          isExtended(row.OriginalClosingDate, row.ClosingDate),
		ExtensionDays:       extensionDays(row.OriginalClosingDate, row.ClosingDate),
		Currency:            row.Currency,
		AmountOriginal:      floatFromNumeric(row.AmountOriginal),
		AmountUSD:           floatFromNumeric(row.AmountUsd),
		CreatedAt:           formatMasterTime(row.CreatedAt),
		UpdatedAt:           formatMasterTime(row.UpdatedAt),
	}
}

func laListResponse(row queries.ListLoanAgreementsRow) model.LoanAgreementResponse {
	return model.LoanAgreementResponse{
		ID:                  model.UUIDToString(row.ID),
		DKProjectID:         model.UUIDToString(row.DkProjectID),
		Lender:              model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
		LoanCode:            row.LoanCode,
		AgreementDate:       dateString(row.AgreementDate),
		EffectiveDate:       dateString(row.EffectiveDate),
		OriginalClosingDate: dateString(row.OriginalClosingDate),
		ClosingDate:         dateString(row.ClosingDate),
		IsExtended:          isExtended(row.OriginalClosingDate, row.ClosingDate),
		ExtensionDays:       extensionDays(row.OriginalClosingDate, row.ClosingDate),
		Currency:            row.Currency,
		AmountOriginal:      floatFromNumeric(row.AmountOriginal),
		AmountUSD:           floatFromNumeric(row.AmountUsd),
		CreatedAt:           formatMasterTime(row.CreatedAt),
		UpdatedAt:           formatMasterTime(row.UpdatedAt),
	}
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
