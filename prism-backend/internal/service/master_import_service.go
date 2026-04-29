package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

const (
	maxMasterImportFileSize  = 20 << 20
	masterImportListLimit    = 100000
	masterImportStatusCreate = "create"
	masterImportStatusSkip   = "skip"
	masterImportStatusFailed = "failed"
)

type xlsxWorkbook struct {
	sheets map[string][]xlsxRow
}

type xlsxRow struct {
	number int
	values []string
}

type importRow struct {
	number int
	values map[string]string
}

type masterImportLenderRef struct {
	ID        pgtype.UUID
	Name      string
	Type      string
	ShortName pgtype.Text
}

type workbookXML struct {
	Sheets []workbookSheetXML `xml:"sheets>sheet"`
}

type workbookSheetXML struct {
	Name string `xml:"name,attr"`
	RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type relationshipsXML struct {
	Relationships []relationshipXML `xml:"Relationship"`
}

type relationshipXML struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type sharedStringsXML struct {
	Items []sharedStringItemXML `xml:"si"`
}

type sharedStringItemXML struct {
	Text string           `xml:"t"`
	Runs []richTextRunXML `xml:"r"`
}

type richTextRunXML struct {
	Text string `xml:"t"`
}

type worksheetXML struct {
	Rows []worksheetRowXML `xml:"sheetData>row"`
}

type worksheetRowXML struct {
	Number int                `xml:"r,attr"`
	Cells  []worksheetCellXML `xml:"c"`
}

type worksheetCellXML struct {
	Ref    string          `xml:"r,attr"`
	Type   string          `xml:"t,attr"`
	Value  string          `xml:"v"`
	Inline inlineStringXML `xml:"is"`
}

type inlineStringXML struct {
	Text string           `xml:"t"`
	Runs []richTextRunXML `xml:"r"`
}

type masterImportLookups struct {
	countriesByCode           map[string]queries.Country
	countriesByName           map[string]queries.Country
	lendersByName             map[string]masterImportLenderRef
	institutionsByName        map[string]queries.Institution
	regionsByCode             map[string]queries.Region
	regionsByName             map[string]queries.Region
	programTitlesByTitle      map[string]queries.ProgramTitle
	bappenasPartnersByName    map[string]queries.BappenasPartner
	periodsByName             map[string]queries.Period
	periodsByYears            map[string]queries.Period
	nationalPriorityKeys      map[string]struct{}
	nationalPrioritiesByTitle map[string]queries.ListNationalPrioritiesRow
	bbProjectsByCode          map[string]queries.ListActiveBBProjectReferencesRow
	gbProjectsByCode          map[string]queries.ListActiveGBProjectReferencesRow
}

func (s *MasterService) ImportMasterData(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processMasterDataWorkbook(ctx, fileName, reader, size, true)
}

func (s *MasterService) PreviewMasterData(ctx context.Context, fileName string, reader io.Reader, size int64) (*model.MasterImportResponse, error) {
	return s.processMasterDataWorkbook(ctx, fileName, reader, size, false)
}

func (s *MasterService) processMasterDataWorkbook(ctx context.Context, fileName string, reader io.Reader, size int64, shouldCommit bool) (*model.MasterImportResponse, error) {
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
	lookups, err := s.loadMasterImportLookups(ctx, qtx)
	if err != nil {
		return nil, err
	}

	importers := []func(context.Context, *queries.Queries, *xlsxWorkbook, *masterImportLookups) (model.MasterImportSheetResult, error){
		s.importProgramTitles,
		s.importBappenasPartners,
		s.importInstitutions,
		s.importRegions,
		s.importPeriods,
		s.importNationalPriorities,
		s.importLenders,
	}

	response := &model.MasterImportResponse{FileName: fileName, Sheets: make([]model.MasterImportSheetResult, 0, len(importers))}
	for _, importSheet := range importers {
		result, err := importSheet(ctx, qtx, workbook, lookups)
		if err != nil {
			return nil, err
		}
		response.TotalInserted += result.Inserted
		response.TotalSkipped += result.Skipped
		response.TotalFailed += result.Failed
		response.Sheets = append(response.Sheets, result)
	}

	if !shouldCommit {
		return response, nil
	}
	if response.TotalFailed > 0 {
		return nil, validation("file", "Perbaiki error preview sebelum eksekusi import")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.Internal("Gagal menyimpan hasil import")
	}

	return response, nil
}

func (s *MasterService) loadMasterImportLookups(ctx context.Context, qtx *queries.Queries) (*masterImportLookups, error) {
	lookups := &masterImportLookups{
		countriesByCode:           map[string]queries.Country{},
		countriesByName:           map[string]queries.Country{},
		lendersByName:             map[string]masterImportLenderRef{},
		institutionsByName:        map[string]queries.Institution{},
		regionsByCode:             map[string]queries.Region{},
		regionsByName:             map[string]queries.Region{},
		programTitlesByTitle:      map[string]queries.ProgramTitle{},
		bappenasPartnersByName:    map[string]queries.BappenasPartner{},
		periodsByName:             map[string]queries.Period{},
		periodsByYears:            map[string]queries.Period{},
		nationalPriorityKeys:      map[string]struct{}{},
		nationalPrioritiesByTitle: map[string]queries.ListNationalPrioritiesRow{},
		bbProjectsByCode:          map[string]queries.ListActiveBBProjectReferencesRow{},
		gbProjectsByCode:          map[string]queries.ListActiveGBProjectReferencesRow{},
	}

	countries, err := qtx.ListCountries(ctx, queries.ListCountriesParams{Limit: masterImportListLimit, Offset: 0, Search: pgtype.Text{}, SortField: "name", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi country")
	}
	for _, country := range countries {
		lookups.addCountry(country)
	}

	lenders, err := qtx.ListLenders(ctx, queries.ListLendersParams{Limit: masterImportListLimit, Offset: 0, TypeFilters: nil, Search: pgtype.Text{}, SortField: "name", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi lender")
	}
	for _, lender := range lenders {
		lookups.addLender(lender.ID, lender.Name, lender.Type, lender.ShortName)
	}

	institutions, err := qtx.ListInstitutions(ctx, queries.ListInstitutionsParams{Limit: masterImportListLimit, Offset: 0, LevelFilters: nil, ParentIDFilter: pgtype.UUID{}, Search: pgtype.Text{}, SortField: "level", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi institution")
	}
	for _, institution := range institutions {
		lookups.addInstitution(institution)
	}

	regions, err := qtx.ListRegions(ctx, queries.ListRegionsParams{Limit: masterImportListLimit, Offset: 0, TypeFilters: nil, ParentCodeFilter: pgtype.Text{}, Search: pgtype.Text{}, SortField: "type", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi region")
	}
	for _, region := range regions {
		lookups.addRegion(region)
	}

	programTitles, err := qtx.ListProgramTitles(ctx, queries.ListProgramTitlesParams{Limit: masterImportListLimit, Offset: 0, Search: pgtype.Text{}, SortField: "title", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi program title")
	}
	for _, programTitle := range programTitles {
		lookups.addProgramTitle(programTitle)
	}

	partners, err := qtx.ListBappenasPartners(ctx, queries.ListBappenasPartnersParams{Limit: masterImportListLimit, Offset: 0, LevelFilters: nil, Search: pgtype.Text{}, SortField: "level", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi bappenas partner")
	}
	for _, partner := range partners {
		lookups.addBappenasPartner(partner)
	}

	periods, err := qtx.ListPeriods(ctx, queries.ListPeriodsParams{Limit: masterImportListLimit, Offset: 0, SortField: "year_start", SortOrder: "desc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi period")
	}
	for _, period := range periods {
		lookups.addPeriod(period)
	}

	priorities, err := qtx.ListNationalPriorities(ctx, queries.ListNationalPrioritiesParams{Limit: masterImportListLimit, Offset: 0, PeriodIDFilters: nil, Search: pgtype.Text{}, SortField: "title", SortOrder: "asc"})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi national priority")
	}
	for _, priority := range priorities {
		lookups.nationalPriorityKeys[nationalPriorityKey(model.UUIDToString(priority.PeriodID), priority.Title)] = struct{}{}
		titleKey := normalizeLookupKey(priority.Title)
		if _, exists := lookups.nationalPrioritiesByTitle[titleKey]; !exists {
			lookups.nationalPrioritiesByTitle[titleKey] = priority
		}
	}

	bbProjects, err := qtx.ListActiveBBProjectReferences(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi BB Project")
	}
	for _, project := range bbProjects {
		lookups.addBBProjectReference(project)
	}

	gbProjects, err := qtx.ListActiveGBProjectReferences(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca referensi GB Project")
	}
	for _, project := range gbProjects {
		lookups.addGBProjectReference(project)
	}

	return lookups, nil
}

func (s *MasterService) importProgramTitles(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "Program Titles"}
	rows, ok := workbook.importRows("Program Titles", []string{"title"})
	if !ok {
		addImportError(&result, 0, "Sheet Program Titles tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	for pass := 0; pass < 2; pass++ {
		for _, row := range rows {
			parentTitle := row.value("parent_title")
			if (pass == 0 && parentTitle != "") || (pass == 1 && parentTitle == "") {
				continue
			}

			title := row.value("title")
			if title == "" {
				addImportError(&result, row.number, "Title wajib diisi")
				continue
			}
			if _, exists := lookups.programTitlesByTitle[normalizeLookupKey(title)]; exists {
				addImportSkipped(&result, row.number, title)
				continue
			}

			parentID := pgtype.UUID{}
			if parentTitle != "" {
				parent, exists := lookups.programTitlesByTitle[normalizeLookupKey(parentTitle)]
				if !exists {
					addImportError(&result, row.number, fmt.Sprintf("Parent Title %q belum ada", parentTitle))
					continue
				}
				parentID = parent.ID
			}

			created, err := qtx.CreateProgramTitle(ctx, queries.CreateProgramTitleParams{ParentID: parentID, Title: title})
			if err != nil {
				return result, fromPg(err)
			}
			lookups.addProgramTitle(created)
			addImportCreated(&result, row.number, created.Title)
		}
	}

	return result, nil
}

func (s *MasterService) importBappenasPartners(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "Bappenas Partners"}
	rows, ok := workbook.importRows("Bappenas Partners", []string{"name", "level"})
	if !ok {
		addImportError(&result, 0, "Sheet Bappenas Partners tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	for pass := 0; pass < 2; pass++ {
		for _, row := range rows {
			parentName := row.value("parent_name")
			if (pass == 0 && parentName != "") || (pass == 1 && parentName == "") {
				continue
			}

			name := row.value("name")
			level, ok := normalizeBappenasPartnerLevel(row.value("level"))
			if name == "" {
				addImportError(&result, row.number, "Name wajib diisi")
				continue
			}
			if !ok {
				addImportError(&result, row.number, "Level harus Eselon I atau Eselon II")
				continue
			}
			if _, exists := lookups.bappenasPartnersByName[normalizeLookupKey(name)]; exists {
				addImportSkipped(&result, row.number, fmt.Sprintf("%s (%s)", name, level))
				continue
			}

			parentID := pgtype.UUID{}
			if parentName != "" {
				parent, exists := lookups.bappenasPartnersByName[normalizeLookupKey(parentName)]
				if !exists {
					addImportError(&result, row.number, fmt.Sprintf("Parent Name %q belum ada", parentName))
					continue
				}
				parentID = parent.ID
			} else if level == "Eselon II" {
				addImportError(&result, row.number, "Eselon II wajib memiliki Parent Name")
				continue
			}

			created, err := qtx.CreateBappenasPartner(ctx, queries.CreateBappenasPartnerParams{ParentID: parentID, Name: name, Level: level})
			if err != nil {
				return result, fromPg(err)
			}
			lookups.addBappenasPartner(created)
			addImportCreated(&result, row.number, fmt.Sprintf("%s (%s)", created.Name, created.Level))
		}
	}

	return result, nil
}

func (s *MasterService) importInstitutions(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "Institutions"}
	rows, ok := workbook.importRows("Institutions", []string{"name", "level"})
	if !ok {
		addImportError(&result, 0, "Sheet Institutions tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	for pass := 0; pass < 2; pass++ {
		for _, row := range rows {
			parentName := row.value("parent_name")
			if (pass == 0 && parentName != "") || (pass == 1 && parentName == "") {
				continue
			}

			name := row.value("name")
			level, ok := normalizeInstitutionLevel(row.value("level"))
			if name == "" {
				addImportError(&result, row.number, "Name wajib diisi")
				continue
			}
			if !ok {
				addImportError(&result, row.number, "Level institution tidak valid")
				continue
			}
			if _, exists := lookups.institutionsByName[normalizeLookupKey(name)]; exists {
				addImportSkipped(&result, row.number, fmt.Sprintf("%s (%s)", name, level))
				continue
			}

			parentID := pgtype.UUID{}
			if parentName != "" {
				parent, exists := lookups.institutionsByName[normalizeLookupKey(parentName)]
				if !exists {
					addImportError(&result, row.number, fmt.Sprintf("Parent Name %q belum ada", parentName))
					continue
				}
				parentID = parent.ID
			}

			created, err := qtx.CreateInstitution(ctx, queries.CreateInstitutionParams{
				ParentID:  parentID,
				Name:      name,
				ShortName: nullableTextPtr(row.optionalString("short_name")),
				Level:     level,
			})
			if err != nil {
				return result, fromPg(err)
			}
			lookups.addInstitution(created)
			addImportCreated(&result, row.number, fmt.Sprintf("%s (%s)", created.Name, created.Level))
		}
	}

	return result, nil
}

func (s *MasterService) importRegions(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "Regions"}
	rows, ok := workbook.importRows("Regions", []string{"code", "name", "level"})
	if !ok {
		addImportError(&result, 0, "Sheet Regions tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	if shouldEnsureIndonesiaRegion(rows, lookups) {
		created, err := qtx.CreateRegion(ctx, queries.CreateRegionParams{Code: "ID", Name: "Indonesia", Type: "COUNTRY", ParentCode: pgtype.Text{}})
		if err != nil {
			return result, fromPg(err)
		}
		lookups.addRegion(created)
		addImportCreated(&result, 0, "ID - Indonesia")
	}

	for _, regionType := range []string{"COUNTRY", "PROVINCE", "CITY"} {
		for _, row := range rows {
			normalizedType, ok := normalizeRegionType(row.value("level"))
			if !ok || normalizedType != regionType {
				if !ok && regionType == "COUNTRY" {
					addImportError(&result, row.number, "Level region harus COUNTRY, PROVINCE, CITY, Provinsi, atau Kota/Kabupaten")
				}
				continue
			}

			code := strings.ToUpper(row.value("code"))
			name := row.value("name")
			parentCode := strings.ToUpper(row.value("parent_code"))
			if code == "" || name == "" {
				addImportError(&result, row.number, "Code dan Name wajib diisi")
				continue
			}
			if _, exists := lookups.regionsByCode[code]; exists {
				addImportSkipped(&result, row.number, fmt.Sprintf("%s - %s", code, name))
				continue
			}
			if parentCode != "" {
				if _, exists := lookups.regionsByCode[parentCode]; !exists {
					addImportError(&result, row.number, fmt.Sprintf("Parent Code %q belum ada", parentCode))
					continue
				}
			}

			created, err := qtx.CreateRegion(ctx, queries.CreateRegionParams{
				Code:       code,
				Name:       name,
				Type:       normalizedType,
				ParentCode: nullableTextPtr(emptyStringToNil(parentCode)),
			})
			if err != nil {
				return result, fromPg(err)
			}
			lookups.addRegion(created)
			addImportCreated(&result, row.number, fmt.Sprintf("%s - %s", created.Code, created.Name))
		}
	}

	return result, nil
}

func (s *MasterService) importPeriods(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "Periods"}
	rows, ok := workbook.importRows("Periods", []string{"name", "year_start", "year_end"})
	if !ok {
		addImportError(&result, 0, "Sheet Periods tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	for _, row := range rows {
		name := row.value("name")
		yearStart, err := parseImportInt(row.value("year_start"))
		if err != nil {
			addImportError(&result, row.number, "Year Start wajib berupa angka")
			continue
		}
		yearEnd, err := parseImportInt(row.value("year_end"))
		if err != nil {
			addImportError(&result, row.number, "Year End wajib berupa angka")
			continue
		}
		if name == "" {
			addImportError(&result, row.number, "Name wajib diisi")
			continue
		}
		if yearEnd <= yearStart {
			addImportError(&result, row.number, "Year End harus lebih besar dari Year Start")
			continue
		}

		if existing, exists := lookups.periodsByName[normalizeLookupKey(name)]; exists {
			lookups.addPeriodAlias(name, existing)
			addImportSkipped(&result, row.number, fmt.Sprintf("%s (%d-%d)", name, yearStart, yearEnd))
			continue
		}
		if existing, exists := lookups.periodsByYears[periodYearsKey(int32(yearStart), int32(yearEnd))]; exists {
			lookups.addPeriodAlias(name, existing)
			addImportSkipped(&result, row.number, fmt.Sprintf("%s (%d-%d)", name, yearStart, yearEnd))
			continue
		}

		created, err := qtx.CreatePeriod(ctx, queries.CreatePeriodParams{Name: name, YearStart: int32(yearStart), YearEnd: int32(yearEnd)})
		if err != nil {
			return result, fromPg(err)
		}
		lookups.addPeriod(created)
		addImportCreated(&result, row.number, fmt.Sprintf("%s (%d-%d)", created.Name, created.YearStart, created.YearEnd))
	}

	return result, nil
}

func (s *MasterService) importNationalPriorities(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "National Priorities"}
	rows, ok := workbook.importRows("National Priorities", []string{"period_name", "title"})
	if !ok {
		addImportError(&result, 0, "Sheet National Priorities tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	for _, row := range rows {
		periodName := row.value("period_name")
		title := row.value("title")
		if periodName == "" || title == "" {
			addImportError(&result, row.number, "Period Name dan Title wajib diisi")
			continue
		}

		period, exists := lookups.periodByLabel(periodName)
		if !exists {
			addImportError(&result, row.number, fmt.Sprintf("Period Name %q belum ada", periodName))
			continue
		}

		key := nationalPriorityKey(model.UUIDToString(period.ID), title)
		if _, exists := lookups.nationalPriorityKeys[key]; exists {
			addImportSkipped(&result, row.number, fmt.Sprintf("%s - %s", periodName, title))
			continue
		}

		created, err := qtx.CreateNationalPriority(ctx, queries.CreateNationalPriorityParams{PeriodID: period.ID, Title: title})
		if err != nil {
			return result, fromPg(err)
		}
		lookups.nationalPriorityKeys[nationalPriorityKey(model.UUIDToString(created.PeriodID), created.Title)] = struct{}{}
		addImportCreated(&result, row.number, fmt.Sprintf("%s - %s", periodName, created.Title))
	}

	return result, nil
}

func (s *MasterService) importLenders(ctx context.Context, qtx *queries.Queries, workbook *xlsxWorkbook, lookups *masterImportLookups) (model.MasterImportSheetResult, error) {
	result := model.MasterImportSheetResult{Sheet: "Lenders"}
	rows, ok := workbook.importRows("Lenders", []string{"name", "type"})
	if !ok {
		addImportError(&result, 0, "Sheet Lenders tidak ditemukan")
		return result, nil
	}
	if hasImportHeaderError(&result, rows) {
		return result, nil
	}

	for _, row := range rows {
		name := row.value("name")
		lenderType, ok := normalizeLenderType(row.value("type"))
		if name == "" {
			addImportError(&result, row.number, "Name wajib diisi")
			continue
		}
		if !ok {
			addImportError(&result, row.number, "Type harus Bilateral, Multilateral, atau KSA")
			continue
		}
		if _, exists := lookups.lendersByName[normalizeLookupKey(name)]; exists {
			addImportSkipped(&result, row.number, fmt.Sprintf("%s (%s)", name, lenderType))
			continue
		}

		countryID := pgtype.UUID{}
		countryName := row.value("country_name")
		if lenderType != "Multilateral" {
			if countryName == "" {
				addImportError(&result, row.number, "Country Name wajib diisi untuk Bilateral dan KSA")
				continue
			}
			country, exists := lookups.countryByNameOrCode(countryName)
			if !exists {
				addImportError(&result, row.number, fmt.Sprintf("Country Name %q belum ada di master country", countryName))
				continue
			}
			countryID = country.ID
		}

		created, err := qtx.CreateLender(ctx, queries.CreateLenderParams{
			CountryID: countryID,
			Name:      name,
			ShortName: nullableTextPtr(row.optionalString("short_name")),
			Type:      lenderType,
		})
		if err != nil {
			return result, fromPg(err)
		}
		lookups.addLender(created.ID, created.Name, created.Type, created.ShortName)
		addImportCreated(&result, row.number, fmt.Sprintf("%s (%s)", created.Name, created.Type))
	}

	return result, nil
}

func readXLSXWorkbook(data []byte) (*xlsxWorkbook, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	files := make(map[string]*zip.File, len(reader.File))
	for _, file := range reader.File {
		files[file.Name] = file
	}

	sharedStrings, err := readSharedStrings(files)
	if err != nil {
		return nil, err
	}

	workbookBytes, err := readZipFile(files, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	var workbook workbookXML
	if err := xml.Unmarshal(workbookBytes, &workbook); err != nil {
		return nil, err
	}

	relsBytes, err := readZipFile(files, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}
	var rels relationshipsXML
	if err := xml.Unmarshal(relsBytes, &rels); err != nil {
		return nil, err
	}
	relTargets := map[string]string{}
	for _, rel := range rels.Relationships {
		relTargets[rel.ID] = rel.Target
	}

	result := &xlsxWorkbook{sheets: map[string][]xlsxRow{}}
	for _, sheet := range workbook.Sheets {
		target, exists := relTargets[sheet.RID]
		if !exists {
			continue
		}
		rows, err := readWorksheetRows(files, workbookTargetPath(target), sharedStrings)
		if err != nil {
			return nil, err
		}
		result.sheets[normalizeSheetName(sheet.Name)] = rows
	}

	return result, nil
}

func readSharedStrings(files map[string]*zip.File) ([]string, error) {
	file, exists := files["xl/sharedStrings.xml"]
	if !exists {
		return nil, nil
	}
	data, err := readOpenedZipFile(file)
	if err != nil {
		return nil, err
	}
	var shared sharedStringsXML
	if err := xml.Unmarshal(data, &shared); err != nil {
		return nil, err
	}
	values := make([]string, 0, len(shared.Items))
	for _, item := range shared.Items {
		if item.Text != "" {
			values = append(values, item.Text)
			continue
		}
		var builder strings.Builder
		for _, run := range item.Runs {
			builder.WriteString(run.Text)
		}
		values = append(values, builder.String())
	}
	return values, nil
}

func readWorksheetRows(files map[string]*zip.File, sheetPath string, sharedStrings []string) ([]xlsxRow, error) {
	data, err := readZipFile(files, sheetPath)
	if err != nil {
		return nil, err
	}
	var worksheet worksheetXML
	if err := xml.Unmarshal(data, &worksheet); err != nil {
		return nil, err
	}

	rows := make([]xlsxRow, 0, len(worksheet.Rows))
	for rowIndex, row := range worksheet.Rows {
		valuesByIndex := map[int]string{}
		maxIndex := -1
		for _, cell := range row.Cells {
			index := columnIndexFromCellRef(cell.Ref)
			if index > maxIndex {
				maxIndex = index
			}
			valuesByIndex[index] = cellStringValue(cell, sharedStrings)
		}
		if maxIndex < 0 {
			continue
		}
		values := make([]string, maxIndex+1)
		for index, value := range valuesByIndex {
			values[index] = strings.TrimSpace(value)
		}
		number := row.Number
		if number == 0 {
			number = rowIndex + 1
		}
		rows = append(rows, xlsxRow{number: number, values: values})
	}
	return rows, nil
}

func readZipFile(files map[string]*zip.File, name string) ([]byte, error) {
	file, exists := files[name]
	if !exists {
		return nil, fmt.Errorf("xlsx part %s not found", name)
	}
	return readOpenedZipFile(file)
}

func readOpenedZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func workbookTargetPath(target string) string {
	target = strings.TrimPrefix(target, "/")
	if strings.HasPrefix(target, "xl/") {
		return path.Clean(target)
	}
	return path.Clean("xl/" + target)
}

func cellStringValue(cell worksheetCellXML, sharedStrings []string) string {
	switch cell.Type {
	case "s":
		index, err := strconv.Atoi(strings.TrimSpace(cell.Value))
		if err != nil || index < 0 || index >= len(sharedStrings) {
			return cell.Value
		}
		return sharedStrings[index]
	case "inlineStr":
		if cell.Inline.Text != "" {
			return cell.Inline.Text
		}
		var builder strings.Builder
		for _, run := range cell.Inline.Runs {
			builder.WriteString(run.Text)
		}
		return builder.String()
	case "b":
		if strings.TrimSpace(cell.Value) == "1" {
			return "TRUE"
		}
		return "FALSE"
	default:
		return cell.Value
	}
}

func columnIndexFromCellRef(ref string) int {
	index := 0
	hasColumn := false
	for _, char := range ref {
		if char < 'A' || char > 'Z' {
			break
		}
		hasColumn = true
		index = index*26 + int(char-'A'+1)
	}
	if !hasColumn {
		return 0
	}
	return index - 1
}

func (w *xlsxWorkbook) importRows(sheetName string, requiredHeaders []string) ([]importRow, bool) {
	rows, exists := w.sheets[normalizeSheetName(sheetName)]
	if !exists {
		return nil, false
	}

	headerIndex := -1
	headerMap := map[int]string{}
	for i, row := range rows {
		if isBlankRow(row.values) {
			continue
		}
		for column, value := range row.values {
			normalized := normalizeHeader(value)
			if normalized != "" {
				headerMap[column] = normalized
			}
		}
		headerIndex = i
		break
	}
	if headerIndex < 0 {
		return nil, true
	}

	available := map[string]struct{}{}
	for _, header := range headerMap {
		available[header] = struct{}{}
	}
	for _, required := range requiredHeaders {
		if _, exists := available[required]; !exists {
			return []importRow{{number: rows[headerIndex].number, values: map[string]string{"_error": fmt.Sprintf("Kolom %s tidak ditemukan", required)}}}, true
		}
	}

	result := make([]importRow, 0, len(rows)-headerIndex-1)
	for _, row := range rows[headerIndex+1:] {
		if isBlankRow(row.values) {
			continue
		}
		values := map[string]string{}
		for column, header := range headerMap {
			if column < len(row.values) {
				values[header] = strings.TrimSpace(row.values[column])
			}
		}
		result = append(result, importRow{number: row.number, values: values})
	}
	return result, true
}

func (r importRow) value(key string) string {
	if r.values == nil {
		return ""
	}
	return strings.TrimSpace(r.values[key])
}

func (r importRow) optionalString(key string) *string {
	value := r.value(key)
	if value == "" {
		return nil
	}
	return &value
}

func isBlankRow(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func normalizeSheetName(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeHeader(value string) string {
	value = strings.ToLower(strings.ReplaceAll(value, "(*)", ""))
	var builder strings.Builder
	lastUnderscore := false
	for _, char := range value {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteRune('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

func addImportCreated(result *model.MasterImportSheetResult, row int, label string) {
	result.Inserted++
	addImportRow(result, row, masterImportStatusCreate, label, "")
}

func addImportCreatedWithMessage(result *model.MasterImportSheetResult, row int, label, message string) {
	result.Inserted++
	addImportRow(result, row, masterImportStatusCreate, label, message)
}

func addImportSkipped(result *model.MasterImportSheetResult, row int, label string) {
	result.Skipped++
	addImportRow(result, row, masterImportStatusSkip, label, "Data sudah ada, dilewati")
}

func addImportError(result *model.MasterImportSheetResult, row int, message string) {
	result.Failed++
	result.Errors = append(result.Errors, model.MasterImportRowError{Row: row, Message: message})
	addImportRow(result, row, masterImportStatusFailed, "", message)
}

func addImportRow(result *model.MasterImportSheetResult, row int, status, label, message string) {
	label = strings.TrimSpace(label)
	if label == "" {
		if row > 0 {
			label = fmt.Sprintf("Baris %d", row)
		} else {
			label = "Workbook"
		}
	}

	result.Rows = append(result.Rows, model.MasterImportRowResult{
		Row:     row,
		Status:  status,
		Label:   label,
		Message: strings.TrimSpace(message),
	})
}

func hasImportHeaderError(result *model.MasterImportSheetResult, rows []importRow) bool {
	if len(rows) != 1 {
		return false
	}
	message := rows[0].value("_error")
	if message == "" {
		return false
	}
	addImportError(result, rows[0].number, message)
	return true
}

func normalizeLookupKey(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

func normalizeLenderType(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "bilateral":
		return "Bilateral", true
	case "multilateral":
		return "Multilateral", true
	case "ksa":
		return "KSA", true
	default:
		return "", false
	}
}

func normalizeBappenasPartnerLevel(value string) (string, bool) {
	switch normalizeLookupKey(value) {
	case "eselon i":
		return "Eselon I", true
	case "eselon ii":
		return "Eselon II", true
	default:
		return "", false
	}
}

func normalizeInstitutionLevel(value string) (string, bool) {
	switch normalizeLookupKey(value) {
	case "kementerian/badan/lembaga", "kementerian", "kementerian lembaga":
		return "Kementerian/Badan/Lembaga", true
	case "eselon i":
		return "Eselon I", true
	case "bumn":
		return "BUMN", true
	case "pemerintah daerah":
		return "Pemerintah Daerah", true
	case "bumd":
		return "BUMD", true
	case "lainnya":
		return "Lainnya", true
	default:
		return "", false
	}
}

func normalizeRegionType(value string) (string, bool) {
	switch normalizeLookupKey(value) {
	case "country", "nasional":
		return "COUNTRY", true
	case "province", "provinsi":
		return "PROVINCE", true
	case "city", "kota/kabupaten", "kabupaten/kota", "kota", "kabupaten":
		return "CITY", true
	default:
		return "", false
	}
}

func parseImportInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty number")
	}
	if strings.Contains(value, ".") {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		}
		return int(parsed), nil
	}
	return strconv.Atoi(value)
}

func parseImportOptionalPositiveInt32(value string) (*int32, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := parseImportInt(value)
	if err != nil {
		return nil, err
	}
	if parsed <= 0 || parsed > 2147483647 {
		return nil, fmt.Errorf("invalid positive int32")
	}
	result := int32(parsed)
	return &result, nil
}

func shouldEnsureIndonesiaRegion(rows []importRow, lookups *masterImportLookups) bool {
	if _, exists := lookups.regionsByCode["ID"]; exists {
		return false
	}
	for _, row := range rows {
		if strings.ToUpper(row.value("code")) == "ID" {
			return false
		}
		if strings.ToUpper(row.value("parent_code")) == "ID" {
			return true
		}
	}
	return false
}

func emptyStringToNil(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func periodYearsKey(yearStart, yearEnd int32) string {
	return fmt.Sprintf("%d-%d", yearStart, yearEnd)
}

func nationalPriorityKey(periodID, title string) string {
	return periodID + "|" + normalizeLookupKey(title)
}

func (l *masterImportLookups) addCountry(country queries.Country) {
	l.countriesByCode[strings.ToUpper(country.Code)] = country
	l.countriesByName[normalizeLookupKey(country.Name)] = country
}

func (l *masterImportLookups) addInstitution(institution queries.Institution) {
	l.institutionsByName[normalizeLookupKey(institution.Name)] = institution
}

func (l *masterImportLookups) addRegion(region queries.Region) {
	l.regionsByCode[strings.ToUpper(region.Code)] = region
	l.regionsByName[normalizeLookupKey(region.Name)] = region
}

func (l *masterImportLookups) addProgramTitle(programTitle queries.ProgramTitle) {
	l.programTitlesByTitle[normalizeLookupKey(programTitle.Title)] = programTitle
}

func (l *masterImportLookups) addBappenasPartner(partner queries.BappenasPartner) {
	l.bappenasPartnersByName[normalizeLookupKey(partner.Name)] = partner
}

func (l *masterImportLookups) addPeriod(period queries.Period) {
	l.periodsByName[normalizeLookupKey(period.Name)] = period
	l.periodsByYears[periodYearsKey(period.YearStart, period.YearEnd)] = period
}

func (l *masterImportLookups) addPeriodAlias(name string, period queries.Period) {
	l.periodsByName[normalizeLookupKey(name)] = period
}

func (l *masterImportLookups) periodByLabel(label string) (queries.Period, bool) {
	if period, exists := l.periodsByName[normalizeLookupKey(label)]; exists {
		return period, true
	}

	parts := strings.Split(strings.TrimSpace(label), "-")
	if len(parts) != 2 {
		return queries.Period{}, false
	}
	start, err := parsePeriodLabelYear(parts[0])
	if err != nil {
		return queries.Period{}, false
	}
	end, err := parsePeriodLabelYear(parts[1])
	if err != nil {
		return queries.Period{}, false
	}
	period, exists := l.periodsByYears[periodYearsKey(int32(start), int32(end))]
	return period, exists
}

func (l *masterImportLookups) countryByNameOrCode(value string) (queries.Country, bool) {
	if country, exists := l.countriesByName[normalizeLookupKey(value)]; exists {
		return country, true
	}
	country, exists := l.countriesByCode[strings.ToUpper(strings.TrimSpace(value))]
	return country, exists
}

func (l *masterImportLookups) addLender(id pgtype.UUID, name, lenderType string, shortName pgtype.Text) {
	l.lendersByName[normalizeLookupKey(name)] = masterImportLenderRef{
		ID:        id,
		Name:      name,
		Type:      lenderType,
		ShortName: shortName,
	}
}

func (l *masterImportLookups) regionByNameOrCode(value string) (queries.Region, bool) {
	if region, exists := l.regionsByCode[strings.ToUpper(strings.TrimSpace(value))]; exists {
		return region, true
	}
	region, exists := l.regionsByName[normalizeLookupKey(value)]
	return region, exists
}

func (l *masterImportLookups) nationalPriorityByTitle(title string) (queries.ListNationalPrioritiesRow, bool) {
	priority, exists := l.nationalPrioritiesByTitle[normalizeLookupKey(title)]
	return priority, exists
}

func (l *masterImportLookups) addBBProjectReference(project queries.ListActiveBBProjectReferencesRow) {
	l.bbProjectsByCode[normalizeLookupKey(project.BbCode)] = project
}

func (l *masterImportLookups) bbProjectByCode(code string) (queries.ListActiveBBProjectReferencesRow, bool) {
	project, exists := l.bbProjectsByCode[normalizeLookupKey(code)]
	return project, exists
}

func (l *masterImportLookups) addGBProjectReference(project queries.ListActiveGBProjectReferencesRow) {
	l.gbProjectsByCode[normalizeLookupKey(project.GbCode)] = project
}

func (l *masterImportLookups) gbProjectByCode(code string) (queries.ListActiveGBProjectReferencesRow, bool) {
	project, exists := l.gbProjectsByCode[normalizeLookupKey(code)]
	return project, exists
}

func parsePeriodLabelYear(value string) (int, error) {
	year, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	if year < 100 {
		return 2000 + year, nil
	}
	return year, nil
}
