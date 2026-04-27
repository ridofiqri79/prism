package service

import (
	"context"
	"math"
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

type MonitoringService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	broker  *sse.Broker
}

func NewMonitoringService(db *pgxpool.Pool, queries *queries.Queries, broker *sse.Broker) *MonitoringService {
	return &MonitoringService{db: db, queries: queries, broker: broker}
}

func (s *MonitoringService) ListMonitoring(ctx context.Context, laID pgtype.UUID, params model.PaginationParams) (*model.ListResponse[model.MonitoringResponse], error) {
	if _, err := s.queries.GetLoanAgreement(ctx, laID); err != nil {
		return nil, mapNotFound(err, "Loan Agreement tidak ditemukan")
	}
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListMonitoringByLA(ctx, queries.ListMonitoringByLAParams{LoanAgreementID: laID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil monitoring")
	}
	total, err := s.queries.CountMonitoringByLA(ctx, laID)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung monitoring")
	}
	data := make([]model.MonitoringResponse, 0, len(rows))
	for _, row := range rows {
		komponen, err := s.queries.GetKomponenByMonitoring(ctx, row.ID)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil komponen monitoring")
		}
		data = append(data, toMonitoringResponse(row, komponen))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *MonitoringService) GetMonitoring(ctx context.Context, laID, id pgtype.UUID) (*model.MonitoringResponse, error) {
	row, err := s.queries.GetMonitoringByLA(ctx, queries.GetMonitoringByLAParams{ID: id, LoanAgreementID: laID})
	if err != nil {
		return nil, mapNotFound(err, "Monitoring tidak ditemukan")
	}
	komponen, err := s.queries.GetKomponenByMonitoring(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil komponen monitoring")
	}
	res := toMonitoringResponse(row, komponen)
	return &res, nil
}

func (s *MonitoringService) CreateMonitoring(ctx context.Context, laID pgtype.UUID, req model.MonitoringRequest) (*model.MonitoringResponse, error) {
	parsed, err := parseMonitoringRequest(req)
	if err != nil {
		return nil, err
	}
	la, err := s.queries.GetLoanAgreement(ctx, laID)
	if err != nil {
		return nil, mapNotFound(err, "Loan Agreement tidak ditemukan")
	}
	if la.EffectiveDate.Valid && la.EffectiveDate.Time.After(time.Now()) {
		return nil, apperrors.BusinessRule("Monitoring hanya bisa dibuat setelah Loan Agreement efektif")
	}
	if _, err := s.queries.GetMonitoringByLAAndPeriod(ctx, queries.GetMonitoringByLAAndPeriodParams{LoanAgreementID: laID, BudgetYear: parsed.BudgetYear, Quarter: parsed.Quarter}); err == nil {
		return nil, apperrors.Conflict("Monitoring untuk periode tersebut sudah ada")
	} else if err != pgx.ErrNoRows {
		return nil, apperrors.Internal("Gagal memeriksa duplikasi monitoring")
	}

	var created queries.MonitoringDisbursement
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.CreateMonitoring(ctx, monitoringCreateParams(laID, parsed))
		if err != nil {
			return err
		}
		created = row
		return createMonitoringKomponen(ctx, qtx, row.ID, parsed.Komponen)
	}); err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.Publish("monitoring.created", map[string]string{"id": model.UUIDToString(created.ID), "loan_agreement_id": model.UUIDToString(laID), "quarter": created.Quarter})
	}
	return s.GetMonitoring(ctx, laID, created.ID)
}

func (s *MonitoringService) UpdateMonitoring(ctx context.Context, laID, id pgtype.UUID, req model.MonitoringRequest) (*model.MonitoringResponse, error) {
	parsed, err := parseMonitoringRequest(req)
	if err != nil {
		return nil, err
	}
	var updated queries.MonitoringDisbursement
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetMonitoringByLA(ctx, queries.GetMonitoringByLAParams{ID: id, LoanAgreementID: laID}); err != nil {
			return mapNotFound(err, "Monitoring tidak ditemukan")
		}
		row, err := qtx.UpdateMonitoring(ctx, queries.UpdateMonitoringParams{
			ID:                 id,
			LoanAgreementID:    laID,
			ExchangeRateUsdIdr: parsed.ExchangeRateUSDIDR,
			ExchangeRateLaIdr:  parsed.ExchangeRateLAIDR,
			PlannedLa:          parsed.PlannedLA,
			PlannedUsd:         parsed.PlannedUSD,
			PlannedIdr:         parsed.PlannedIDR,
			RealizedLa:         parsed.RealizedLA,
			RealizedUsd:        parsed.RealizedUSD,
			RealizedIdr:        parsed.RealizedIDR,
		})
		if err != nil {
			return err
		}
		updated = row
		if err := qtx.DeleteKomponenByMonitoring(ctx, id); err != nil {
			return err
		}
		return createMonitoringKomponen(ctx, qtx, id, parsed.Komponen)
	}); err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.Publish("monitoring.updated", map[string]string{"id": model.UUIDToString(updated.ID), "loan_agreement_id": model.UUIDToString(laID), "quarter": updated.Quarter})
	}
	return s.GetMonitoring(ctx, laID, updated.ID)
}

func (s *MonitoringService) DeleteMonitoring(ctx context.Context, laID, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetMonitoringByLA(ctx, queries.GetMonitoringByLAParams{ID: id, LoanAgreementID: laID}); err != nil {
			return mapNotFound(err, "Monitoring tidak ditemukan")
		}
		return qtx.DeleteMonitoring(ctx, queries.DeleteMonitoringParams{ID: id, LoanAgreementID: laID})
	})
}

func (s *MonitoringService) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
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
		return apperrors.Internal("Gagal menyimpan monitoring")
	}
	return nil
}

type parsedMonitoringRequest struct {
	BudgetYear         int32
	Quarter            string
	ExchangeRateUSDIDR pgtype.Numeric
	ExchangeRateLAIDR  pgtype.Numeric
	PlannedLA          pgtype.Numeric
	PlannedUSD         pgtype.Numeric
	PlannedIDR         pgtype.Numeric
	RealizedLA         pgtype.Numeric
	RealizedUSD        pgtype.Numeric
	RealizedIDR        pgtype.Numeric
	Komponen           []model.MonitoringKomponenItem
}

func parseMonitoringRequest(req model.MonitoringRequest) (parsedMonitoringRequest, error) {
	quarter := strings.ToUpper(strings.TrimSpace(req.Quarter))
	if req.BudgetYear == 0 {
		return parsedMonitoringRequest{}, validation("budget_year", "wajib diisi")
	}
	if quarter != "TW1" && quarter != "TW2" && quarter != "TW3" && quarter != "TW4" {
		return parsedMonitoringRequest{}, validation("quarter", "harus TW1, TW2, TW3, atau TW4")
	}
	komponen := req.Komponen
	if len(komponen) == 0 && len(req.Components) > 0 {
		komponen = req.Components
	}
	for i, item := range komponen {
		if strings.TrimSpace(item.ComponentName) == "" {
			return parsedMonitoringRequest{}, validation("komponen", "component_name wajib diisi pada komponen ke-"+strconv.Itoa(i+1))
		}
	}
	return parsedMonitoringRequest{
		BudgetYear:         req.BudgetYear,
		Quarter:            quarter,
		ExchangeRateUSDIDR: numericFromFloat(req.ExchangeRateUSDIDR),
		ExchangeRateLAIDR:  numericFromFloat(req.ExchangeRateLAIDR),
		PlannedLA:          numericFromFloat(req.PlannedLA),
		PlannedUSD:         numericFromFloat(req.PlannedUSD),
		PlannedIDR:         numericFromFloat(req.PlannedIDR),
		RealizedLA:         numericFromFloat(req.RealizedLA),
		RealizedUSD:        numericFromFloat(req.RealizedUSD),
		RealizedIDR:        numericFromFloat(req.RealizedIDR),
		Komponen:           komponen,
	}, nil
}

func monitoringCreateParams(laID pgtype.UUID, parsed parsedMonitoringRequest) queries.CreateMonitoringParams {
	return queries.CreateMonitoringParams{
		LoanAgreementID:    laID,
		BudgetYear:         parsed.BudgetYear,
		Quarter:            parsed.Quarter,
		ExchangeRateUsdIdr: parsed.ExchangeRateUSDIDR,
		ExchangeRateLaIdr:  parsed.ExchangeRateLAIDR,
		PlannedLa:          parsed.PlannedLA,
		PlannedUsd:         parsed.PlannedUSD,
		PlannedIdr:         parsed.PlannedIDR,
		RealizedLa:         parsed.RealizedLA,
		RealizedUsd:        parsed.RealizedUSD,
		RealizedIdr:        parsed.RealizedIDR,
	}
}

func createMonitoringKomponen(ctx context.Context, qtx *queries.Queries, monitoringID pgtype.UUID, komponen []model.MonitoringKomponenItem) error {
	for _, item := range komponen {
		if _, err := qtx.CreateKomponen(ctx, queries.CreateKomponenParams{
			MonitoringDisbursementID: monitoringID,
			ComponentName:            strings.TrimSpace(item.ComponentName),
			PlannedLa:                numericFromFloat(item.PlannedLA),
			PlannedUsd:               numericFromFloat(item.PlannedUSD),
			PlannedIdr:               numericFromFloat(item.PlannedIDR),
			RealizedLa:               numericFromFloat(item.RealizedLA),
			RealizedUsd:              numericFromFloat(item.RealizedUSD),
			RealizedIdr:              numericFromFloat(item.RealizedIDR),
		}); err != nil {
			return err
		}
	}
	return nil
}

func toMonitoringResponse(row queries.MonitoringDisbursement, komponen []queries.MonitoringKomponen) model.MonitoringResponse {
	plannedUSD := floatFromNumeric(row.PlannedUsd)
	realizedUSD := floatFromNumeric(row.RealizedUsd)
	return model.MonitoringResponse{
		ID:                 model.UUIDToString(row.ID),
		LoanAgreementID:    model.UUIDToString(row.LoanAgreementID),
		BudgetYear:         row.BudgetYear,
		Quarter:            row.Quarter,
		ExchangeRateUSDIDR: floatFromNumeric(row.ExchangeRateUsdIdr),
		ExchangeRateLAIDR:  floatFromNumeric(row.ExchangeRateLaIdr),
		PlannedLA:          floatFromNumeric(row.PlannedLa),
		PlannedUSD:         plannedUSD,
		PlannedIDR:         floatFromNumeric(row.PlannedIdr),
		RealizedLA:         floatFromNumeric(row.RealizedLa),
		RealizedUSD:        realizedUSD,
		RealizedIDR:        floatFromNumeric(row.RealizedIdr),
		AbsorptionPct:      absorptionPct(plannedUSD, realizedUSD),
		Komponen:           toMonitoringKomponenResponses(komponen),
		CreatedAt:          formatMasterTime(row.CreatedAt),
		UpdatedAt:          formatMasterTime(row.UpdatedAt),
	}
}

func toMonitoringKomponenResponses(rows []queries.MonitoringKomponen) []model.MonitoringKomponenResponse {
	data := make([]model.MonitoringKomponenResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, model.MonitoringKomponenResponse{
			ID:            model.UUIDToString(row.ID),
			ComponentName: row.ComponentName,
			PlannedLA:     floatFromNumeric(row.PlannedLa),
			PlannedUSD:    floatFromNumeric(row.PlannedUsd),
			PlannedIDR:    floatFromNumeric(row.PlannedIdr),
			RealizedLA:    floatFromNumeric(row.RealizedLa),
			RealizedUSD:   floatFromNumeric(row.RealizedUsd),
			RealizedIDR:   floatFromNumeric(row.RealizedIdr),
		})
	}
	return data
}

func absorptionPct(planned, realized float64) float64 {
	if planned <= 0 {
		return 0
	}
	return math.Round(realized/planned*1000) / 10
}
