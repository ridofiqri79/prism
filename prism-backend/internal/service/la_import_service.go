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
	laImportSheetInput     = "Loan Agreement"
	laImportSheetRelations = "Relasi - DK Project"
)

type laImportDKProjectLookup struct {
	byID    map[string]queries.ListLoanAgreementImportDKProjectReferencesRow
	byLabel map[string]queries.ListLoanAgreementImportDKProjectReferencesRow
}

type parsedLoanAgreementImportAllocation struct {
	DKProjectID        pgtype.UUID
	DKProjectName      string
	AllocationOriginal float64
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
		"lender_name",
		"loan_code",
		"agreement_date",
		"effective_date",
		"original_closing_date",
		"closing_date",
		"currency",
		"amount_original",
		"cumulative_disbursement",
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

	relationResult := model.MasterImportSheetResult{Sheet: laImportSheetRelations}
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
	relationAllocations, hasRelationSheet := parseLoanAgreementImportRelations(workbook, lookup, &relationResult)

	masterSvc := &MasterService{db: s.db, queries: s.queries}
	lookups, err := masterSvc.loadMasterImportLookups(ctx, qtx)
	if err != nil {
		return nil, nil, err
	}

	createdIDs := make([]string, 0)
	seenLoanCodes := map[string]struct{}{}

	for _, row := range rows {
		label := strings.TrimSpace(row.value("loan_code"))
		if label == "" {
			label = row.value("dk_project_ref")
		}

		parsed, skip, messages := s.parseLoanAgreementImportRow(ctx, qtx, row, lookup, allowedByDK, lookups, seenLoanCodes, relationAllocations, hasRelationSheet)
		switch {
		case skip:
			result.Skipped++
			addImportRow(&result, row.number, masterImportStatusSkip, label, strings.Join(messages, "; "))
		case len(messages) > 0:
			addImportError(&result, row.number, strings.Join(messages, "; "))
		default:
			allocationRows, err := buildLoanAgreementImportAllocationRows(ctx, qtx, parsed)
			if err != nil {
				return nil, nil, err
			}
			created, err := qtx.CreateLoanAgreement(ctx, queries.CreateLoanAgreementParams{
				DkProjectID:            parsed.PrimaryDKProjectID(),
				LenderID:               parsed.LenderID,
				LoanCode:               parsed.LoanCode,
				AgreementDate:          parsed.AgreementDate,
				EffectiveDate:          parsed.EffectiveDate,
				OriginalClosingDate:    parsed.OriginalClosingDate,
				ClosingDate:            parsed.ClosingDate,
				Currency:               parsed.Currency,
				AmountOriginal:         parsed.AmountOriginal,
				AmountUsd:              parsed.AmountUsd,
				CumulativeDisbursement: parsed.CumulativeDisbursement,
			})
			if err != nil {
				return nil, nil, fromPg(err)
			}
			for _, allocation := range allocationRows {
				if err := qtx.AddLoanAgreementDKProject(ctx, queries.AddLoanAgreementDKProjectParams{
					LoanAgreementID:    created.ID,
					DkProjectID:        allocation.DKProjectID,
					AllocationOriginal: numericFromFloat(allocation.AllocationOriginal),
					AllocationUsd:      numericFromFloat(allocation.AllocationUSD),
				}); err != nil {
					return nil, nil, fromPg(err)
				}
			}
			createdIDs = append(createdIDs, model.UUIDToString(created.ID))
			addImportCreated(&result, row.number, fmt.Sprintf("%s - %s", parsed.LoanCode, parsed.ProjectNames()))
		}
	}

	sheets := []model.MasterImportSheetResult{result}
	if hasRelationSheet {
		sheets = append(sheets, relationResult)
	}
	response := &model.MasterImportResponse{
		FileName: fileName,
		Sheets:   sheets,
	}
	recalculateImportTotals(response)
	return response, createdIDs, nil
}

type parsedLoanAgreementImportRow struct {
	Allocations            []parsedLoanAgreementImportAllocation
	LenderID               pgtype.UUID
	LoanCode               string
	AgreementDate          pgtype.Date
	EffectiveDate          pgtype.Date
	OriginalClosingDate    pgtype.Date
	ClosingDate            pgtype.Date
	Currency               string
	AmountOriginalValue    float64
	AmountUSDValue         float64
	AmountOriginal         pgtype.Numeric
	AmountUsd              pgtype.Numeric
	CumulativeDisbursement pgtype.Numeric
}

func (p parsedLoanAgreementImportRow) PrimaryDKProjectID() pgtype.UUID {
	if len(p.Allocations) == 0 {
		return pgtype.UUID{}
	}
	return p.Allocations[0].DKProjectID
}

func (p parsedLoanAgreementImportRow) ProjectNames() string {
	names := make([]string, 0, len(p.Allocations))
	seen := map[string]struct{}{}
	for _, allocation := range p.Allocations {
		name := strings.TrimSpace(allocation.DKProjectName)
		if name == "" {
			name = model.UUIDToString(allocation.DKProjectID)
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func (p parsedLoanAgreementImportRow) RequestAllocations() []parsedLoanAgreementAllocation {
	allocations := make([]parsedLoanAgreementAllocation, 0, len(p.Allocations))
	for _, allocation := range p.Allocations {
		allocations = append(allocations, parsedLoanAgreementAllocation{
			DKProjectID:        allocation.DKProjectID,
			AllocationOriginal: allocation.AllocationOriginal,
		})
	}
	return allocations
}

func (s *LAService) parseLoanAgreementImportRow(ctx context.Context, qtx *queries.Queries, row importRow, dkLookup laImportDKProjectLookup, allowedByDK map[string]map[string]struct{}, lookups *masterImportLookups, seenLoanCodes map[string]struct{}, relationAllocations map[string][]parsedLoanAgreementImportAllocation, hasRelationSheet bool) (parsedLoanAgreementImportRow, bool, []string) {
	var parsed parsedLoanAgreementImportRow
	messages := make([]string, 0)
	addMessage := func(message string) {
		message = strings.TrimSpace(message)
		if message != "" {
			messages = append(messages, message)
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

	if hasRelationSheet {
		allocations := relationAllocations[normalizeLookupKey(loanCode)]
		if loanCode != "" && len(allocations) == 0 {
			addMessage("Relasi DK Project untuk Loan Code belum diisi")
		}
		parsed.Allocations = allocations
	} else {
		dkProjectRef := row.value("dk_project_ref")
		dkProject, exists := resolveLAImportDKProjectRef(dkProjectRef, dkLookup)
		if dkProjectRef == "" {
			addMessage("DK Project Ref wajib diisi")
		} else if !exists {
			addMessage("DK Project Ref tidak ditemukan di snapshot Master Data")
		} else {
			parsed.Allocations = []parsedLoanAgreementImportAllocation{{
				DKProjectID:        dkProject.ID,
				DKProjectName:      dkProject.ProjectName,
				AllocationOriginal: 0,
			}}
			if !dkProject.HasFinancingDetail {
				addMessage("DK Project belum memiliki Financing Detail")
			}
		}
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
			for _, allocation := range parsed.Allocations {
				dkAllowed := allowedByDK[model.UUIDToString(allocation.DKProjectID)]
				if _, ok := dkAllowed[model.UUIDToString(lender.ID)]; !ok {
					addMessage(fmt.Sprintf("Lender harus berasal dari Financing Detail DK Project %s", allocation.DKProjectName))
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
	originalClosingDate, err := parseLAImportOptionalDate(row.value("original_closing_date"), "Original Closing Date")
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
	cumulativeDisbursement, err := parseLAImportAmount(row.value("cumulative_disbursement"), "Cumulative Disbursement", false)
	if err != nil {
		addMessage(err.Error())
	}
	if !hasRelationSheet && len(parsed.Allocations) == 1 {
		parsed.Allocations[0].AllocationOriginal = amountOriginal
	}
	if hasRelationSheet && len(parsed.Allocations) > 0 && !sameLoanAgreementImportAllocationTotal(parsed.Allocations, amountOriginal) {
		addMessage("Total Allocation Original pada sheet Relasi - DK Project harus sama dengan Amount Original")
	}
	amountUSD, err := calculateLoanAgreementAmountUSD(ctx, qtx, currency, amountOriginal)
	if err != nil {
		addMessage(err.Error())
	}
	parsed.AmountOriginalValue = amountOriginal
	parsed.AmountUSDValue = amountUSD
	parsed.AmountOriginal = numericFromFloat(amountOriginal)
	parsed.AmountUsd = numericFromFloat(amountUSD)
	parsed.CumulativeDisbursement = numericFromFloat(cumulativeDisbursement)

	return parsed, false, messages
}

func parseLoanAgreementImportRelations(workbook *xlsxWorkbook, dkLookup laImportDKProjectLookup, result *model.MasterImportSheetResult) (map[string][]parsedLoanAgreementImportAllocation, bool) {
	rows, ok := workbook.importRows(laImportSheetRelations, []string{
		"loan_code",
		"dk_project_ref",
		"allocation_original",
	})
	if !ok {
		return map[string][]parsedLoanAgreementImportAllocation{}, false
	}
	if hasImportHeaderError(result, rows) {
		return map[string][]parsedLoanAgreementImportAllocation{}, true
	}
	allocationsByLoanCode := map[string][]parsedLoanAgreementImportAllocation{}
	seenByLoanCode := map[string]map[string]struct{}{}
	for _, row := range rows {
		messages := make([]string, 0)
		addMessage := func(message string) {
			if strings.TrimSpace(message) != "" {
				messages = append(messages, strings.TrimSpace(message))
			}
		}
		loanCode := strings.TrimSpace(row.value("loan_code"))
		if loanCode == "" {
			addMessage("Loan Code wajib diisi")
		}
		dkProjectRef := row.value("dk_project_ref")
		dkProject, exists := resolveLAImportDKProjectRef(dkProjectRef, dkLookup)
		if dkProjectRef == "" {
			addMessage("DK Project Ref wajib diisi")
		} else if !exists {
			addMessage("DK Project Ref tidak ditemukan di snapshot Master Data")
		} else if !dkProject.HasFinancingDetail {
			addMessage("DK Project belum memiliki Financing Detail")
		}
		allocationOriginal, err := parseLAImportAmount(row.value("allocation_original"), "Allocation Original", true)
		if err != nil {
			addMessage(err.Error())
		}
		loanCodeKey := normalizeLookupKey(loanCode)
		if loanCodeKey != "" && exists {
			if seenByLoanCode[loanCodeKey] == nil {
				seenByLoanCode[loanCodeKey] = map[string]struct{}{}
			}
			dkProjectID := model.UUIDToString(dkProject.ID)
			if _, duplicate := seenByLoanCode[loanCodeKey][dkProjectID]; duplicate {
				addMessage("DK Project duplikat untuk Loan Code yang sama")
			}
			seenByLoanCode[loanCodeKey][dkProjectID] = struct{}{}
		}
		label := strings.TrimSpace(loanCode)
		if exists {
			label = fmt.Sprintf("%s - %s", loanCode, dkProject.ProjectName)
		}
		if len(messages) > 0 {
			addImportError(result, row.number, strings.Join(messages, "; "))
			continue
		}
		allocationsByLoanCode[loanCodeKey] = append(allocationsByLoanCode[loanCodeKey], parsedLoanAgreementImportAllocation{
			DKProjectID:        dkProject.ID,
			DKProjectName:      dkProject.ProjectName,
			AllocationOriginal: allocationOriginal,
		})
		addImportRow(result, row.number, masterImportStatusCreate, label, "Relasi valid")
	}
	return allocationsByLoanCode, true
}

func sameLoanAgreementImportAllocationTotal(allocations []parsedLoanAgreementImportAllocation, amountOriginal float64) bool {
	total := 0.0
	for _, allocation := range allocations {
		total += allocation.AllocationOriginal
	}
	return sameCurrencyAmount(total, amountOriginal)
}

func buildLoanAgreementImportAllocationRows(ctx context.Context, qtx *queries.Queries, parsed parsedLoanAgreementImportRow) ([]loanAgreementAllocationRow, error) {
	request := parsedLoanAgreementRequest{
		Allocations:         parsed.RequestAllocations(),
		Currency:            parsed.Currency,
		AmountOriginalValue: parsed.AmountOriginalValue,
		AmountOriginal:      parsed.AmountOriginal,
	}
	return buildLoanAgreementAllocationRows(ctx, qtx, request, parsed.AmountUSDValue)
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

func parseLAImportOptionalDate(value, label string) (pgtype.Date, error) {
	if strings.TrimSpace(value) == "" {
		return pgtype.Date{}, nil
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
