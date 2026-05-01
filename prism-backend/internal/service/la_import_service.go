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

const laImportSheetInput = "Loan Agreement"

type laImportDKProjectLookup struct {
	byID    map[string]queries.ListLoanAgreementImportDKProjectReferencesRow
	byLabel map[string]queries.ListLoanAgreementImportDKProjectReferencesRow
}

func (s *LAService) PreviewLoanAgreementImport(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processLoanAgreementWorkbook(ctx, fileName, reader, size, false)
}

func (s *LAService) ImportLoanAgreement(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processLoanAgreementWorkbook(ctx, fileName, reader, size, true)
}

func (s *LAService) processLoanAgreementWorkbook(ctx context.Context, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
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
	response, createdIDs, err := s.buildLoanAgreementImportPreview(ctx, qtx, workbook, fileName)
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
		return nil, apperrors.Internal("Gagal menyimpan hasil import Loan Agreement")
	}

	if s.broker != nil {
		for _, id := range createdIDs {
			s.broker.Publish("loan_agreement.created", map[string]string{"id": id})
		}
	}

	return response, nil
}

func (s *LAService) buildLoanAgreementImportPreview(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, fileName string) (*model.MasterImportResponse, []string, error) {
	result := model.MasterImportSheetResult{Sheet: laImportSheetInput}
	rows, ok := workbook.importRows(laImportSheetInput, []string{
		"dk_project_ref",
		"lender_name",
		"loan_code",
		"agreement_date",
		"effective_date",
		"original_closing_date",
		"closing_date",
		"currency",
		"amount_original",
		"amount_usd",
	})
	if !ok {
		addImportError(&result, 0, "Sheet Loan Agreement tidak ditemukan")
		response := &model.MasterImportResponse{FileName: fileName, Sheets: []model.MasterImportSheetResult{result}}
		recalculateImportTotals(response)
		return response, nil, nil
	}
	if hasImportHeaderError(&result, rows) {
		response := &model.MasterImportResponse{FileName: fileName, Sheets: []model.MasterImportSheetResult{result}}
		recalculateImportTotals(response)
		return response, nil, nil
	}

	dkRefs, err := qtx.ListLoanAgreementImportDKProjectReferences(ctx)
	if err != nil {
		return nil, nil, apperrors.Internal("Gagal membaca referensi DK Project")
	}
	allowedLenders, err := qtx.ListLoanAgreementAllowedLenderReferences(ctx)
	if err != nil {
		return nil, nil, apperrors.Internal("Gagal membaca referensi lender Loan Agreement")
	}
	lookup := buildLAImportDKProjectLookup(dkRefs)
	allowedByDK := buildLAImportAllowedLenderMap(allowedLenders)

	masterSvc := &MasterService{db: s.db, queries: s.queries}
	lookups, err := masterSvc.loadMasterImportLookups(ctx, qtx)
	if err != nil {
		return nil, nil, err
	}

	createdIDs := make([]string, 0)
	seenDKProjects := map[string]struct{}{}
	seenLoanCodes := map[string]struct{}{}

	for _, row := range rows {
		label := strings.TrimSpace(row.value("loan_code"))
		if label == "" {
			label = row.value("dk_project_ref")
		}

		parsed, skip, messages := s.parseLoanAgreementImportRow(ctx, qtx, row, lookup, allowedByDK, lookups, seenDKProjects, seenLoanCodes)
		switch {
		case skip:
			result.Skipped++
			addImportRow(&result, row.number, masterImportStatusSkip, label, strings.Join(messages, "; "))
		case len(messages) > 0:
			addImportError(&result, row.number, strings.Join(messages, "; "))
		default:
			created, err := qtx.CreateLoanAgreement(ctx, queries.CreateLoanAgreementParams{
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
			if err != nil {
				return nil, nil, fromPg(err)
			}
			createdIDs = append(createdIDs, model.UUIDToString(created.ID))
			addImportCreated(&result, row.number, fmt.Sprintf("%s - %s", parsed.LoanCode, parsed.DKProjectName))
		}
	}

	response := &model.MasterImportResponse{
		FileName: fileName,
		Sheets:   []model.MasterImportSheetResult{result},
	}
	recalculateImportTotals(response)
	return response, createdIDs, nil
}

type parsedLoanAgreementImportRow struct {
	DKProjectID         pgtype.UUID
	DKProjectName       string
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

func (s *LAService) parseLoanAgreementImportRow(ctx context.Context, qtx *queries.Queries, row importRow, dkLookup laImportDKProjectLookup, allowedByDK map[string]map[string]struct{}, lookups *masterImportLookups, seenDKProjects map[string]struct{}, seenLoanCodes map[string]struct{}) (parsedLoanAgreementImportRow, bool, []string) {
	var parsed parsedLoanAgreementImportRow
	messages := make([]string, 0)
	addMessage := func(message string) {
		message = strings.TrimSpace(message)
		if message != "" {
			messages = append(messages, message)
		}
	}

	dkProjectRef := row.value("dk_project_ref")
	dkProject, exists := resolveLAImportDKProjectRef(dkProjectRef, dkLookup)
	if dkProjectRef == "" {
		addMessage("DK Project Ref wajib diisi")
	} else if !exists {
		addMessage("DK Project Ref tidak ditemukan di snapshot Master Data")
	} else {
		parsed.DKProjectID = dkProject.ID
		parsed.DKProjectName = dkProject.ProjectName
		dkProjectID := model.UUIDToString(dkProject.ID)
		if dkProject.ExistingLoanAgreementID.Valid {
			return parsed, true, []string{"DK Project sudah memiliki Loan Agreement, dilewati"}
		}
		if !dkProject.HasFinancingDetail {
			addMessage("DK Project belum memiliki Financing Detail")
		}
		if _, seen := seenDKProjects[dkProjectID]; seen {
			addMessage("DK Project duplikat di workbook")
		} else {
			seenDKProjects[dkProjectID] = struct{}{}
		}
	}

	loanCode := strings.TrimSpace(row.value("loan_code"))
	if loanCode == "" {
		addMessage("Loan Code wajib diisi")
	} else {
		loanCodeKey := normalizeLookupKey(loanCode)
		if _, seen := seenLoanCodes[loanCodeKey]; seen {
			addMessage("Loan Code duplikat di workbook")
		} else {
			seenLoanCodes[loanCodeKey] = struct{}{}
		}
		if _, err := qtx.GetLoanAgreementByLoanCode(ctx, loanCode); err == nil {
			addMessage("Loan Code sudah digunakan")
		} else if err != pgx.ErrNoRows {
			addMessage("Gagal memeriksa Loan Code")
		}
		parsed.LoanCode = loanCode
	}

	lenderName := row.value("lender_name")
	if lenderName == "" {
		addMessage("Lender Name wajib diisi")
	} else {
		lender, exists, ambiguous := lookups.lookupLenderReference(lenderName)
		switch {
		case ambiguous:
			addMessage(fmt.Sprintf("Lender %q ambigu karena short_name dipakai lebih dari satu lender", lenderName))
		case !exists:
			addMessage(fmt.Sprintf("Lender %q belum ada di master data", lenderName))
		default:
			parsed.LenderID = lender.ID
			if parsed.DKProjectID.Valid {
				dkAllowed := allowedByDK[model.UUIDToString(parsed.DKProjectID)]
				if _, ok := dkAllowed[model.UUIDToString(lender.ID)]; !ok {
					addMessage("Lender harus berasal dari Financing Detail DK Project terkait")
				}
			}
		}
	}

	agreementDate, err := parseLAImportDate(row.value("agreement_date"), "Agreement Date")
	if err != nil {
		addMessage(err.Error())
	} else {
		parsed.AgreementDate = agreementDate
	}
	effectiveDate, err := parseLAImportDate(row.value("effective_date"), "Effective Date")
	if err != nil {
		addMessage(err.Error())
	} else {
		parsed.EffectiveDate = effectiveDate
	}
	originalClosingDate, err := parseLAImportDate(row.value("original_closing_date"), "Original Closing Date")
	if err != nil {
		addMessage(err.Error())
	} else {
		parsed.OriginalClosingDate = originalClosingDate
	}
	closingDate, err := parseLAImportDate(row.value("closing_date"), "Closing Date")
	if err != nil {
		addMessage(err.Error())
	} else {
		parsed.ClosingDate = closingDate
	}
	if parsed.OriginalClosingDate.Valid && parsed.ClosingDate.Valid && parsed.ClosingDate.Time.Before(parsed.OriginalClosingDate.Time) {
		addMessage("Closing Date tidak boleh lebih awal dari Original Closing Date")
	}

	currency := normalizeCurrency(row.value("currency"))
	if strings.TrimSpace(row.value("currency")) == "" {
		addMessage("Currency wajib diisi")
	} else if err := validateLAImportCurrency(ctx, qtx, currency); err != nil {
		addMessage(err.Error())
	}
	parsed.Currency = currency

	amountOriginal, err := parseLAImportAmount(row.value("amount_original"), "Amount Original", true)
	if err != nil {
		addMessage(err.Error())
	}
	amountUSD, err := parseLAImportAmount(row.value("amount_usd"), "Amount USD", currency != "USD")
	if err != nil {
		addMessage(err.Error())
	}
	amountOriginal, amountUSD = normalizeCurrencyAmountPair(currency, amountOriginal, amountUSD)
	parsed.AmountOriginal = numericFromFloat(amountOriginal)
	parsed.AmountUsd = numericFromFloat(amountUSD)

	return parsed, false, messages
}

func buildLAImportDKProjectLookup(items []queries.ListLoanAgreementImportDKProjectReferencesRow) laImportDKProjectLookup {
	lookup := laImportDKProjectLookup{
		byID:    map[string]queries.ListLoanAgreementImportDKProjectReferencesRow{},
		byLabel: map[string]queries.ListLoanAgreementImportDKProjectReferencesRow{},
	}
	for _, item := range items {
		id := model.UUIDToString(item.ID)
		lookup.byID[id] = item
		lookup.byLabel[normalizeLookupKey(laDKProjectReferenceLabel(item))] = item
	}
	return lookup
}

func resolveLAImportDKProjectRef(value string, lookup laImportDKProjectLookup) (queries.ListLoanAgreementImportDKProjectReferencesRow, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return queries.ListLoanAgreementImportDKProjectReferencesRow{}, false
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

func buildLAImportAllowedLenderMap(items []queries.ListLoanAgreementAllowedLenderReferencesRow) map[string]map[string]struct{} {
	result := map[string]map[string]struct{}{}
	for _, item := range items {
		dkProjectID := model.UUIDToString(item.DkProjectID)
		lenderID := model.UUIDToString(item.LenderID)
		if result[dkProjectID] == nil {
			result[dkProjectID] = map[string]struct{}{}
		}
		result[dkProjectID][lenderID] = struct{}{}
	}
	return result
}

func parseLAImportDate(value, label string) (pgtype.Date, error) {
	if strings.TrimSpace(value) == "" {
		return pgtype.Date{}, fmt.Errorf("%s wajib diisi", label)
	}
	date, err := parseDKImportDate(value)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("%s harus tanggal valid", label)
	}
	return date, nil
}

func parseLAImportAmount(value, label string, required bool) (float64, error) {
	if strings.TrimSpace(value) == "" && !required {
		return 0, nil
	}
	amount, err := parseDKImportAmount(value, label)
	if err != nil {
		return 0, err
	}
	if required && amount <= 0 {
		return 0, fmt.Errorf("%s wajib lebih dari 0", label)
	}
	return amount, nil
}

func validateLAImportCurrency(ctx context.Context, qtx *queries.Queries, code string) error {
	currency, err := qtx.GetCurrencyByCode(ctx, code)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("Currency harus terdaftar di Master Currency")
	}
	if err != nil {
		return fmt.Errorf("Gagal memeriksa Currency")
	}
	if !currency.IsActive {
		return fmt.Errorf("Currency tidak aktif di Master Currency")
	}
	return nil
}

func laDKProjectReferenceLabel(item queries.ListLoanAgreementImportDKProjectReferencesRow) string {
	id := model.UUIDToString(item.ID)
	projectName := strings.TrimSpace(item.ProjectName)
	if projectName == "" {
		projectName = "Tanpa nama proyek"
	}
	context := laDKProjectContextLabel(item)
	gbCodes := strings.TrimSpace(item.GbCodes)
	if gbCodes == "" {
		return fmt.Sprintf("%s | %s | %s", context, projectName, id)
	}
	return fmt.Sprintf("%s | %s | %s | %s", context, projectName, gbCodes, id)
}

func laDKProjectContextLabel(item queries.ListLoanAgreementImportDKProjectReferencesRow) string {
	if strings.TrimSpace(item.LetterNumber) != "" {
		return strings.TrimSpace(item.LetterNumber)
	}
	if strings.TrimSpace(item.Subject) != "" {
		return strings.TrimSpace(item.Subject)
	}
	return "Daftar Kegiatan"
}
