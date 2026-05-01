package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

const (
	monitoringImportSheetInput      = "Monitoring Disbursement"
	monitoringImportSheetComponents = "Relasi - Komponen"
)

type monitoringImportLoanAgreementLookup struct {
	byID    map[string]queries.ListMonitoringImportLoanAgreementReferencesRow
	byLabel map[string]queries.ListMonitoringImportLoanAgreementReferencesRow
}

type monitoringImportDraft struct {
	row                int
	loanAgreementID    pgtype.UUID
	loanCode           string
	budgetYear         int32
	quarter            string
	exchangeRateUSDIDR pgtype.Numeric
	exchangeRateLAIDR  pgtype.Numeric
	plannedLA          pgtype.Numeric
	plannedUSD         pgtype.Numeric
	plannedIDR         pgtype.Numeric
	realizedLA         pgtype.Numeric
	realizedUSD        pgtype.Numeric
	realizedIDR        pgtype.Numeric
	components         []model.MonitoringKomponenItem
	skip               bool
	skipMessage        string
	errors             []string
	createdID          pgtype.UUID
}

func (d *monitoringImportDraft) failed() bool {
	return len(d.errors) > 0
}

func (s *MonitoringService) PreviewMonitoringImport(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processMonitoringWorkbook(ctx, fileName, reader, size, false)
}

func (s *MonitoringService) ImportMonitoring(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processMonitoringWorkbook(ctx, fileName, reader, size, true)
}

func (s *MonitoringService) processMonitoringWorkbook(ctx context.Context, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
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
	response, created, err := s.buildMonitoringImportPreview(ctx, qtx, workbook, fileName)
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
		return nil, apperrors.Internal("Gagal menyimpan hasil import Monitoring Disbursement")
	}

	if s.broker != nil {
		for _, item := range created {
			s.broker.Publish("monitoring.created", map[string]string{
				"id":                model.UUIDToString(item.createdID),
				"loan_agreement_id": model.UUIDToString(item.loanAgreementID),
				"quarter":           item.quarter,
			})
		}
	}

	return response, nil
}

func (s *MonitoringService) buildMonitoringImportPreview(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, fileName string) (*model.MasterImportResponse, []*monitoringImportDraft, error) {
	inputResult := model.MasterImportSheetResult{Sheet: monitoringImportSheetInput}
	componentResult := model.MasterImportSheetResult{Sheet: monitoringImportSheetComponents}

	refs, err := qtx.ListMonitoringImportLoanAgreementReferences(ctx)
	if err != nil {
		return nil, nil, apperrors.Internal("Gagal membaca referensi Loan Agreement")
	}
	lookup := buildMonitoringImportLoanAgreementLookup(refs)

	drafts, draftsByKey, err := s.parseMonitoringImportRows(ctx, qtx, workbook, lookup, &inputResult)
	if err != nil {
		return nil, nil, err
	}
	componentRows := s.parseMonitoringImportComponents(workbook, lookup, draftsByKey, &componentResult)

	created := make([]*monitoringImportDraft, 0)
	for _, draft := range drafts {
		label := monitoringImportLabel(draft.loanCode, draft.budgetYear, draft.quarter)
		switch {
		case draft.skip:
			inputResult.Skipped++
			addImportRow(&inputResult, draft.row, masterImportStatusSkip, label, draft.skipMessage)
		case draft.failed():
			addImportError(&inputResult, draft.row, strings.Join(draft.errors, "; "))
		default:
			row, err := qtx.CreateMonitoring(ctx, monitoringCreateParams(draft.loanAgreementID, parsedMonitoringRequest{
				BudgetYear:         draft.budgetYear,
				Quarter:            draft.quarter,
				ExchangeRateUSDIDR: draft.exchangeRateUSDIDR,
				ExchangeRateLAIDR:  draft.exchangeRateLAIDR,
				PlannedLA:          draft.plannedLA,
				PlannedUSD:         draft.plannedUSD,
				PlannedIDR:         draft.plannedIDR,
				RealizedLA:         draft.realizedLA,
				RealizedUSD:        draft.realizedUSD,
				RealizedIDR:        draft.realizedIDR,
				Komponen:           draft.components,
			}))
			if err != nil {
				return nil, nil, fromPg(err)
			}
			draft.createdID = row.ID
			if err := createMonitoringKomponen(ctx, qtx, row.ID, draft.components); err != nil {
				return nil, nil, fromPg(err)
			}
			created = append(created, draft)
			addImportCreated(&inputResult, draft.row, label)
		}
	}

	for _, componentRow := range componentRows {
		label := componentRow.label()
		switch {
		case componentRow.draft == nil:
			addImportError(&componentResult, componentRow.row, "Monitoring terkait tidak ada di sheet Monitoring Disbursement")
		case componentRow.draft.skip:
			componentResult.Skipped++
			addImportRow(&componentResult, componentRow.row, masterImportStatusSkip, label, "Monitoring sudah ada, komponen dilewati")
		case componentRow.draft.failed():
			addImportError(&componentResult, componentRow.row, "Monitoring terkait gagal validasi")
		default:
			addImportCreated(&componentResult, componentRow.row, label)
		}
	}

	response := &model.MasterImportResponse{
		FileName: fileName,
		Sheets:   []model.MasterImportSheetResult{inputResult, componentResult},
	}
	recalculateImportTotals(response)
	return response, created, nil
}

func (s *MonitoringService) parseMonitoringImportRows(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookup monitoringImportLoanAgreementLookup, result *model.MasterImportSheetResult) ([]*monitoringImportDraft, map[string]*monitoringImportDraft, error) {
	rows, ok := workbook.importRows(monitoringImportSheetInput, []string{
		"loan_agreement_ref",
		"budget_year",
		"quarter",
		"exchange_rate_usd_idr",
		"exchange_rate_loan_agreement_idr",
	})
	if !ok {
		addImportError(result, 0, "Sheet Monitoring Disbursement tidak ditemukan")
		return nil, map[string]*monitoringImportDraft{}, nil
	}
	if hasImportHeaderError(result, rows) {
		return nil, map[string]*monitoringImportDraft{}, nil
	}

	drafts := make([]*monitoringImportDraft, 0, len(rows))
	draftsByKey := map[string]*monitoringImportDraft{}
	seenNewPeriods := map[string]struct{}{}

	for _, row := range rows {
		draft := s.parseMonitoringImportRow(ctx, qtx, row, lookup, seenNewPeriods)
		drafts = append(drafts, draft)
		if draft.loanAgreementID.Valid && draft.budgetYear > 0 && draft.quarter != "" {
			key := monitoringImportKey(draft.loanAgreementID, draft.budgetYear, draft.quarter)
			if _, exists := draftsByKey[key]; !exists {
				draftsByKey[key] = draft
			}
		}
	}

	return drafts, draftsByKey, nil
}

func (s *MonitoringService) parseMonitoringImportRow(ctx context.Context, qtx *queries.Queries, row importRow, lookup monitoringImportLoanAgreementLookup, seenNewPeriods map[string]struct{}) *monitoringImportDraft {
	draft := &monitoringImportDraft{row: row.number}
	addError := func(message string) {
		if strings.TrimSpace(message) != "" {
			draft.errors = append(draft.errors, strings.TrimSpace(message))
		}
	}

	refValue := row.value("loan_agreement_ref")
	ref, exists := resolveMonitoringImportLoanAgreementRef(refValue, lookup)
	if refValue == "" {
		addError("Loan Agreement Ref wajib diisi")
	} else if !exists {
		addError("Loan Agreement Ref tidak ditemukan di snapshot Master Data")
	} else {
		draft.loanAgreementID = ref.ID
		draft.loanCode = ref.LoanCode
		if !ref.IsEffective {
			addError("Loan Agreement belum efektif")
		}
	}

	budgetYear, err := parseMonitoringImportBudgetYear(row.value("budget_year"))
	if err != nil {
		addError(err.Error())
	} else {
		draft.budgetYear = budgetYear
	}

	quarter, err := parseMonitoringImportQuarter(row.value("quarter"))
	if err != nil {
		addError(err.Error())
	} else {
		draft.quarter = quarter
	}

	if draft.loanAgreementID.Valid && draft.budgetYear > 0 && draft.quarter != "" {
		key := monitoringImportKey(draft.loanAgreementID, draft.budgetYear, draft.quarter)
		if _, err := qtx.GetMonitoringByLAAndPeriod(ctx, queries.GetMonitoringByLAAndPeriodParams{
			LoanAgreementID: draft.loanAgreementID,
			BudgetYear:      draft.budgetYear,
			Quarter:         draft.quarter,
		}); err == nil {
			draft.skip = true
			draft.skipMessage = "Monitoring untuk periode tersebut sudah ada, dilewati"
			return draft
		} else if err != pgx.ErrNoRows {
			addError("Gagal memeriksa duplikasi monitoring")
		}
		if _, seen := seenNewPeriods[key]; seen {
			addError("Periode duplikat di workbook untuk Loan Agreement yang sama")
		} else {
			seenNewPeriods[key] = struct{}{}
		}
	}

	draft.exchangeRateUSDIDR = numericFromFloat(parseMonitoringImportNumber(row.value("exchange_rate_usd_idr"), "Exchange Rate USD/IDR", true, true, addError))
	draft.exchangeRateLAIDR = numericFromFloat(parseMonitoringImportNumber(row.value("exchange_rate_loan_agreement_idr"), "Exchange Rate Loan Agreement/IDR", true, true, addError))
	draft.plannedLA = numericFromFloat(parseMonitoringImportNumber(row.value("planned_loan_agreement"), "Planned Loan Agreement", false, false, addError))
	draft.plannedUSD = numericFromFloat(parseMonitoringImportNumber(row.value("planned_usd"), "Planned USD", false, false, addError))
	draft.plannedIDR = numericFromFloat(parseMonitoringImportNumber(row.value("planned_idr"), "Planned IDR", false, false, addError))
	draft.realizedLA = numericFromFloat(parseMonitoringImportNumber(row.value("realized_loan_agreement"), "Realized Loan Agreement", false, false, addError))
	draft.realizedUSD = numericFromFloat(parseMonitoringImportNumber(row.value("realized_usd"), "Realized USD", false, false, addError))
	draft.realizedIDR = numericFromFloat(parseMonitoringImportNumber(row.value("realized_idr"), "Realized IDR", false, false, addError))

	return draft
}

type monitoringImportComponentRow struct {
	row     int
	draft   *monitoringImportDraft
	name    string
	year    int32
	quarter string
}

func (r monitoringImportComponentRow) label() string {
	if r.draft == nil {
		return r.name
	}
	return fmt.Sprintf("%s - %s", monitoringImportLabel(r.draft.loanCode, r.year, r.quarter), r.name)
}

func (s *MonitoringService) parseMonitoringImportComponents(workbook *xlsxWorkbook, lookup monitoringImportLoanAgreementLookup, draftsByKey map[string]*monitoringImportDraft, result *model.MasterImportSheetResult) []monitoringImportComponentRow {
	rows, ok := workbook.importRows(monitoringImportSheetComponents, []string{
		"loan_agreement_ref",
		"budget_year",
		"quarter",
		"component_name",
	})
	if !ok {
		return nil
	}
	if hasImportHeaderError(result, rows) {
		return nil
	}

	componentRows := make([]monitoringImportComponentRow, 0, len(rows))
	for _, row := range rows {
		componentRow, item, err := parseMonitoringImportComponentRow(row, lookup, draftsByKey)
		if err != nil {
			addImportError(result, row.number, err.Error())
			continue
		}
		if componentRow.draft != nil {
			componentRow.draft.components = append(componentRow.draft.components, item)
		}
		componentRows = append(componentRows, componentRow)
	}

	return componentRows
}

func parseMonitoringImportComponentRow(row importRow, lookup monitoringImportLoanAgreementLookup, draftsByKey map[string]*monitoringImportDraft) (monitoringImportComponentRow, model.MonitoringKomponenItem, error) {
	refValue := row.value("loan_agreement_ref")
	ref, exists := resolveMonitoringImportLoanAgreementRef(refValue, lookup)
	if refValue == "" {
		return monitoringImportComponentRow{}, model.MonitoringKomponenItem{}, fmt.Errorf("Loan Agreement Ref wajib diisi")
	}
	if !exists {
		return monitoringImportComponentRow{}, model.MonitoringKomponenItem{}, fmt.Errorf("Loan Agreement Ref tidak ditemukan di snapshot Master Data")
	}
	year, err := parseMonitoringImportBudgetYear(row.value("budget_year"))
	if err != nil {
		return monitoringImportComponentRow{}, model.MonitoringKomponenItem{}, err
	}
	quarter, err := parseMonitoringImportQuarter(row.value("quarter"))
	if err != nil {
		return monitoringImportComponentRow{}, model.MonitoringKomponenItem{}, err
	}
	name := strings.TrimSpace(row.value("component_name"))
	if name == "" {
		return monitoringImportComponentRow{}, model.MonitoringKomponenItem{}, fmt.Errorf("Component Name wajib diisi")
	}

	addError := func(message string) {
		err = fmt.Errorf("%s", message)
	}
	item := model.MonitoringKomponenItem{
		ComponentName: name,
		PlannedLA:     parseMonitoringImportNumber(row.value("planned_loan_agreement"), "Planned Loan Agreement", false, false, addError),
		PlannedUSD:    parseMonitoringImportNumber(row.value("planned_usd"), "Planned USD", false, false, addError),
		PlannedIDR:    parseMonitoringImportNumber(row.value("planned_idr"), "Planned IDR", false, false, addError),
		RealizedLA:    parseMonitoringImportNumber(row.value("realized_loan_agreement"), "Realized Loan Agreement", false, false, addError),
		RealizedUSD:   parseMonitoringImportNumber(row.value("realized_usd"), "Realized USD", false, false, addError),
		RealizedIDR:   parseMonitoringImportNumber(row.value("realized_idr"), "Realized IDR", false, false, addError),
	}
	if err != nil {
		return monitoringImportComponentRow{}, model.MonitoringKomponenItem{}, err
	}

	draft := draftsByKey[monitoringImportKey(ref.ID, year, quarter)]
	return monitoringImportComponentRow{
		row:     row.number,
		draft:   draft,
		name:    name,
		year:    year,
		quarter: quarter,
	}, item, nil
}

func buildMonitoringImportLoanAgreementLookup(items []queries.ListMonitoringImportLoanAgreementReferencesRow) monitoringImportLoanAgreementLookup {
	lookup := monitoringImportLoanAgreementLookup{
		byID:    map[string]queries.ListMonitoringImportLoanAgreementReferencesRow{},
		byLabel: map[string]queries.ListMonitoringImportLoanAgreementReferencesRow{},
	}
	for _, item := range items {
		id := model.UUIDToString(item.ID)
		lookup.byID[id] = item
		lookup.byLabel[normalizeLookupKey(monitoringLoanAgreementReferenceLabel(item))] = item
	}
	return lookup
}

func resolveMonitoringImportLoanAgreementRef(value string, lookup monitoringImportLoanAgreementLookup) (queries.ListMonitoringImportLoanAgreementReferencesRow, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return queries.ListMonitoringImportLoanAgreementReferencesRow{}, false
	}
	if id, err := model.ParseUUID(value); err == nil {
		item, exists := lookup.byID[model.UUIDToString(id)]
		return item, exists
	}
	if parts := strings.Split(value, "|"); len(parts) > 1 {
		trailingID := strings.TrimSpace(parts[len(parts)-1])
		if id, err := model.ParseUUID(trailingID); err == nil {
			item, exists := lookup.byID[model.UUIDToString(id)]
			return item, exists
		}
	}
	item, exists := lookup.byLabel[normalizeLookupKey(value)]
	return item, exists
}

func parseMonitoringImportBudgetYear(value string) (int32, error) {
	parsed, err := parseImportInt(value)
	if err != nil {
		return 0, fmt.Errorf("Budget Year wajib berupa angka")
	}
	if parsed < 2000 || parsed > 2200 {
		return 0, fmt.Errorf("Budget Year harus antara 2000 dan 2200")
	}
	return int32(parsed), nil
}

func parseMonitoringImportQuarter(value string) (string, error) {
	quarter := strings.ToUpper(strings.TrimSpace(value))
	if quarter != "TW1" && quarter != "TW2" && quarter != "TW3" && quarter != "TW4" {
		return "", fmt.Errorf("Quarter harus TW1, TW2, TW3, atau TW4")
	}
	return quarter, nil
}

func parseMonitoringImportNumber(value, label string, required, positive bool, addError func(string)) float64 {
	if strings.TrimSpace(value) == "" && !required {
		return 0
	}
	if strings.TrimSpace(value) == "" {
		addError(label + " wajib diisi")
		return 0
	}
	amount, err := parseImportFloat(value)
	if err != nil {
		addError(label + " wajib berupa angka")
		return 0
	}
	if amount < 0 {
		addError(label + " tidak boleh negatif")
		return 0
	}
	if positive && amount <= 0 {
		addError(label + " wajib lebih dari 0")
		return 0
	}
	return amount
}

func monitoringImportKey(loanAgreementID pgtype.UUID, budgetYear int32, quarter string) string {
	return model.UUIDToString(loanAgreementID) + "|" + fmt.Sprint(budgetYear) + "|" + strings.ToUpper(strings.TrimSpace(quarter))
}

func monitoringImportLabel(loanCode string, budgetYear int32, quarter string) string {
	loanCode = strings.TrimSpace(loanCode)
	if loanCode == "" {
		loanCode = "Loan Agreement"
	}
	if budgetYear == 0 || strings.TrimSpace(quarter) == "" {
		return loanCode
	}
	return fmt.Sprintf("%s - %d %s", loanCode, budgetYear, quarter)
}
