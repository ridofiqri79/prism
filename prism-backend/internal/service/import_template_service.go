package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

const importTemplateContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
const importTemplateEditableRows = 5000

const (
	xlsxStyleDefault = iota
	xlsxStyleTitle
	xlsxStyleSubtitle
	xlsxStyleSection
	xlsxStyleHeader
	xlsxStyleNote
)

type importTemplateFile struct {
	FileName    string
	ContentType string
	Data        []byte
}

type simpleXLSXWorkbook struct {
	Sheets       []simpleXLSXSheet
	DefinedNames []simpleXLSXDefinedName
}

type simpleXLSXSheet struct {
	Name          string
	Rows          [][]simpleXLSXCell
	Columns       []simpleXLSXColumn
	Validations   []simpleXLSXValidation
	AutoFilter    string
	FreezeRows    int
	Hidden        bool
	ShowGridLines bool
}

type simpleXLSXCell struct {
	Value  string
	Number bool
	Style  int
}

type simpleXLSXColumn struct {
	Min   int
	Max   int
	Width float64
}

type simpleXLSXValidation struct {
	Range       string
	Type        string
	Operator    string
	Formula1    string
	Formula2    string
	PromptTitle string
	Prompt      string
	ErrorTitle  string
	Error       string
	ErrorStyle  string
	AllowBlank  bool
}

type simpleXLSXDefinedName struct {
	Name string
	Ref  string
}

type importTemplateReferenceData struct {
	ProgramTitles      []queries.ProgramTitle
	BappenasPartners   []queries.BappenasPartner
	Institutions       []queries.Institution
	Regions            []queries.Region
	Periods            []queries.Period
	NationalPriorities []queries.ListNationalPrioritiesRow
	Lenders            []queries.ListLendersRow
	Countries          []queries.Country
	BBProjects         []queries.ListActiveBBProjectReferencesRow
	GBProjects         []queries.ListActiveGBProjectReferencesRow
	AllowedGBLenders   []queries.ListAllowedLenderReferencesByGBProjectRow
}

func textCell(value string) simpleXLSXCell {
	return simpleXLSXCell{Value: value}
}

func styledTextCell(value string, style int) simpleXLSXCell {
	return simpleXLSXCell{Value: value, Style: style}
}

func numberCell(value int32) simpleXLSXCell {
	return simpleXLSXCell{Value: strconv.FormatInt(int64(value), 10), Number: true}
}

func floatCell(value float64) simpleXLSXCell {
	return simpleXLSXCell{Value: strconv.FormatFloat(value, 'f', 2, 64), Number: true}
}

func textRow(values ...string) []simpleXLSXCell {
	row := make([]simpleXLSXCell, 0, len(values))
	for _, value := range values {
		row = append(row, textCell(value))
	}
	return row
}

func styledTextRow(style int, values ...string) []simpleXLSXCell {
	row := make([]simpleXLSXCell, 0, len(values))
	for _, value := range values {
		row = append(row, styledTextCell(value, style))
	}
	return row
}

func headerRow(values ...string) []simpleXLSXCell {
	return styledTextRow(xlsxStyleHeader, values...)
}

func (s *MasterService) BuildMasterImportTemplate(ctx context.Context) (*importTemplateFile, error) {
	workbook, err := s.buildMasterImportTemplateWorkbook(ctx)
	if err != nil {
		return nil, err
	}

	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat template import")
	}

	return &importTemplateFile{
		FileName:    "master_data_import_template.xlsx",
		ContentType: importTemplateContentType,
		Data:        data,
	}, nil
}

func (s *BlueBookService) BuildProjectImportTemplate(ctx context.Context, bbID pgtype.UUID) (*importTemplateFile, error) {
	blueBook, err := s.queries.GetBlueBook(ctx, bbID)
	if err != nil {
		return nil, mapNotFound(err, "Blue Book tidak ditemukan")
	}

	workbook, err := s.buildBlueBookProjectImportTemplateWorkbook(ctx, blueBook)
	if err != nil {
		return nil, err
	}

	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat template import Blue Book")
	}

	return &importTemplateFile{
		FileName:    fmt.Sprintf("blue_book_%s_import_template.xlsx", safeFileToken(blueBook.PeriodName)),
		ContentType: importTemplateContentType,
		Data:        data,
	}, nil
}

func (s *GreenBookService) BuildProjectImportTemplate(ctx context.Context, gbID pgtype.UUID) (*importTemplateFile, error) {
	greenBook, err := s.queries.GetGreenBook(ctx, gbID)
	if err != nil {
		return nil, mapNotFound(err, "Green Book tidak ditemukan")
	}

	workbook, err := s.buildGreenBookProjectImportTemplateWorkbook(ctx, greenBook)
	if err != nil {
		return nil, err
	}

	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat template import Green Book")
	}

	return &importTemplateFile{
		FileName:    fmt.Sprintf("green_book_%d_import_template.xlsx", greenBook.PublishYear),
		ContentType: importTemplateContentType,
		Data:        data,
	}, nil
}

func (s *DKService) BuildImportTemplate(ctx context.Context) (*importTemplateFile, error) {
	workbook, err := s.buildDKImportTemplateWorkbook(ctx)
	if err != nil {
		return nil, err
	}

	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat template import Daftar Kegiatan")
	}

	return &importTemplateFile{
		FileName:    "daftar_kegiatan_import_template.xlsx",
		ContentType: importTemplateContentType,
		Data:        data,
	}, nil
}

func (s *MasterService) buildMasterImportTemplateWorkbook(ctx context.Context) (simpleXLSXWorkbook, error) {
	reference, err := s.loadImportTemplateReferenceData(ctx, pgtype.UUID{})
	if err != nil {
		return simpleXLSXWorkbook{}, err
	}

	dropdowns, definedNames := buildDropdownSheet(reference)
	sheets := []simpleXLSXSheet{
		buildMasterGuideSheet(),
		templateInputSheet("Program Titles", []string{"Title (*)", "Parent Title"}, []float64{42, 42}, []simpleXLSXValidation{
			listValidation("B2:B"+inputLastRow(), "ddProgramTitles", "Parent Title", "Pilih parent dari Program Title yang sudah ada. Jika parent baru ikut diimpor, tulis parent sebagai baris tanpa Parent Title terlebih dahulu."),
		}),
		templateInputSheet("Bappenas Partners", []string{"Name (*)", "Level (*)", "Parent Name"}, []float64{42, 20, 42}, []simpleXLSXValidation{
			listValidation("B2:B"+inputLastRow(), "ddBappenasPartnerLevels", "Level", "Pilih Eselon I atau Eselon II. Eselon II wajib memiliki Parent Name."),
			listValidation("C2:C"+inputLastRow(), "ddBappenasPartnerParents", "Parent Name", "Pilih Eselon I yang sudah ada atau yang dibuat di baris sebelumnya."),
		}),
		templateInputSheet("Institutions", []string{"Name (*)", "Short Name", "Level (*)", "Parent Name"}, []float64{46, 18, 30, 46}, []simpleXLSXValidation{
			listValidation("C2:C"+inputLastRow(), "ddInstitutionLevels", "Level", "Pilih level institution yang sesuai dengan skema PRISM."),
			listValidation("D2:D"+inputLastRow(), "ddInstitutions", "Parent Name", "Pilih parent institution yang sudah ada jika level ini punya parent."),
		}),
		templateInputSheet("Regions", []string{"Code (*)", "Name (*)", "Level (*)", "Parent Code"}, []float64{18, 42, 20, 18}, []simpleXLSXValidation{
			listValidation("C2:C"+inputLastRow(), "ddRegionLevels", "Level", "Pilih COUNTRY, PROVINCE, atau CITY. Alias Indonesia seperti Nasional/Provinsi/Kota juga diterima backend."),
			listValidation("D2:D"+inputLastRow(), "ddRegionCodes", "Parent Code", "Pilih kode parent region. PROVINCE biasanya parent COUNTRY; CITY parent PROVINCE."),
		}),
		templateInputSheet("Periods", []string{"Name (*)", "Year Start (*)", "Year End (*)"}, []float64{28, 18, 18}, []simpleXLSXValidation{
			numberValidation("B2:B"+inputLastRow(), "Year Start", "Isi tahun awal dalam angka empat digit, contoh 2025."),
			numberValidation("C2:C"+inputLastRow(), "Year End", "Isi tahun akhir dalam angka empat digit dan harus lebih besar dari Year Start."),
		}),
		templateInputSheet("National Priorities", []string{"Period Name (*)", "Title (*)"}, []float64{28, 70}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddPeriods", "Period Name", "Pilih nama period yang sudah ada atau yang dibuat di sheet Periods."),
		}),
		templateInputSheet("Lenders", []string{"Name (*)", "Short Name", "Type (*)", "Country Name"}, []float64{42, 18, 22, 34}, []simpleXLSXValidation{
			listValidation("C2:C"+inputLastRow(), "ddLenderTypes", "Type", "Pilih Bilateral, Multilateral, atau KSA. Bilateral dan KSA wajib memiliki Country Name."),
			listValidation("D2:D"+inputLastRow(), "ddCountries", "Country Name", "Pilih country untuk Bilateral atau KSA. Kosongkan untuk Multilateral."),
		}),
		buildMasterDataSnapshotSheet("Master Data Snapshot", reference),
		dropdowns,
	}

	return simpleXLSXWorkbook{Sheets: sheets, DefinedNames: definedNames}, nil
}

func (s *BlueBookService) buildBlueBookProjectImportTemplateWorkbook(ctx context.Context, blueBook queries.GetBlueBookRow) (simpleXLSXWorkbook, error) {
	masterSvc := &MasterService{db: s.db, queries: s.queries}
	reference, err := masterSvc.loadImportTemplateReferenceData(ctx, pgtype.UUID{})
	if err != nil {
		return simpleXLSXWorkbook{}, err
	}

	dropdowns, definedNames := buildDropdownSheet(reference)
	definedNames = append(definedNames, simpleXLSXDefinedName{
		Name: "ddInputBBCodes",
		Ref:  xlsxRangeRef(blueBookImportSheetInput, 3, 2, importTemplateEditableRows+1),
	})

	sheets := []simpleXLSXSheet{
		buildBlueBookGuideSheet(blueBook),
		buildMasterDataSnapshotSheet("Master Data", reference),
		templateInputSheet(blueBookImportSheetInput, []string{"Program Title (*)", "Bappenas Partner", "BB Code (*)", "Project Name (*)", "Duration", "Objective", "Scope of Work", "Outputs", "Outcomes"}, []float64{38, 38, 22, 54, 18, 54, 54, 44, 44}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddProgramTitles", "Program Title", "Pilih Program Title dari master data."),
			listValidation("B2:B"+inputLastRow(), "ddBappenasPartnersEselonII", "Bappenas Partner", "Pilih Bappenas Partner Eselon II bila ada. Kolom ini opsional."),
		}),
		templateInputSheet(blueBookImportSheetEA, []string{"BB Code (*)", "Executing Agency Name (*)"}, []float64{22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputBBCodes", "BB Code", "Pilih BB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddInstitutions", "Executing Agency", "Pilih institution dari master data. Institution yang sama boleh dipakai sebagai EA dan IA bila memang sesuai data proyek."),
		}),
		templateInputSheet(blueBookImportSheetIA, []string{"BB Code (*)", "Implementing Agency Name (*)"}, []float64{22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputBBCodes", "BB Code", "Pilih BB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddInstitutions", "Implementing Agency", "Pilih institution dari master data. Boleh sama dengan EA untuk BB Code yang sama."),
		}),
		templateInputSheet(blueBookImportSheetLocations, []string{"BB Code (*)", "Location Name (*)"}, []float64{22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputBBCodes", "BB Code", "Pilih BB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddRegionNames", "Location Name", "Pilih region dari master data. Backend juga menerima kode region bila diketik manual."),
		}),
		templateInputSheet(blueBookImportSheetNationalPriority, []string{"BB Code (*)", "National Priority Name (*)"}, []float64{22, 64}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputBBCodes", "BB Code", "Pilih BB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddNationalPriorities", "National Priority", "Pilih National Priority dari seluruh master data, tidak dibatasi period Blue Book target."),
		}),
		templateInputSheet(blueBookImportSheetProjectCost, []string{"BB Code (*)", "Funding Type (*)", "Funding Category (*)", "Amount USD"}, []float64{22, 22, 32, 18}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputBBCodes", "BB Code", "Pilih BB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddFundingTypes", "Funding Type", "Pilih Foreign atau Counterpart."),
			listValidation("C2:C"+inputLastRow(), "ddFundingCategories", "Funding Category", "Foreign biasanya Loan/Grant; Counterpart biasanya Central Government/Regional Government/State-Owned Enterprise/Others."),
			decimalValidation("D2:D"+inputLastRow(), "Amount USD", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
		}),
		templateInputSheet(blueBookImportSheetLenderIndication, []string{"BB Code (*)", "Lender Name (*)", "Keterangan"}, []float64{22, 42, 54}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputBBCodes", "BB Code", "Pilih BB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddLenders", "Lender Name", "Pilih lender dari master data."),
		}),
		dropdowns,
	}

	return simpleXLSXWorkbook{Sheets: sheets, DefinedNames: definedNames}, nil
}

func (s *GreenBookService) buildGreenBookProjectImportTemplateWorkbook(ctx context.Context, greenBook queries.GreenBook) (simpleXLSXWorkbook, error) {
	masterSvc := &MasterService{db: s.db, queries: s.queries}
	reference, err := masterSvc.loadImportTemplateReferenceData(ctx, pgtype.UUID{})
	if err != nil {
		return simpleXLSXWorkbook{}, err
	}

	dropdowns, definedNames := buildDropdownSheet(reference)
	definedNames = append(definedNames, simpleXLSXDefinedName{
		Name: "ddInputGBCodes",
		Ref:  xlsxRangeRef(greenBookImportSheetInput, 2, 2, importTemplateEditableRows+1),
	}, simpleXLSXDefinedName{
		Name: "ddInputActivityNos",
		Ref:  xlsxRangeRef(greenBookImportSheetActivities, 2, 2, importTemplateEditableRows+1),
	})

	sheets := []simpleXLSXSheet{
		buildGreenBookGuideSheet(greenBook),
		buildGreenBookMasterDataSnapshotSheet("Master Data", reference),
		templateInputSheet(greenBookImportSheetInput, []string{"Program Title (*)", "GB Code (*)", "Project Name (*)", "Duration", "Objective", "Scope of Project"}, []float64{38, 22, 54, 18, 54, 54}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddProgramTitles", "Program Title", "Pilih Program Title dari master data."),
		}),
		templateInputSheet(greenBookImportSheetBBProject, []string{"GB Code (*)", "BB Code (*)"}, []float64{22, 22}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddBBProjectCodes", "BB Code", "Pilih BB Project active dari database."),
		}),
		templateInputSheet(greenBookImportSheetEA, []string{"GB Code (*)", "Executing Agency Name (*)"}, []float64{22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddInstitutions", "Executing Agency", "Pilih institution dari master data."),
		}),
		templateInputSheet(greenBookImportSheetIA, []string{"GB Code (*)", "Implementing Agency Name (*)"}, []float64{22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddInstitutions", "Implementing Agency", "Pilih institution dari master data."),
		}),
		templateInputSheet(greenBookImportSheetLocations, []string{"GB Code (*)", "Location Name (*)"}, []float64{22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddRegionNames", "Location Name", "Pilih region dari master data. Backend juga menerima kode region bila diketik manual."),
		}),
		templateInputSheet(greenBookImportSheetActivities, []string{"GB Code (*)", "Activity No (*)", "Activity Name (*)", "Implementation Location", "PIU", "Sort Order"}, []float64{22, 18, 54, 42, 42, 16}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			integerValidation("B2:B"+inputLastRow(), "Activity No", "Isi nomor aktivitas yang unik per GB Code. Nomor ini dipakai oleh Funding Allocation."),
			integerValidation("F2:F"+inputLastRow(), "Sort Order", "Isi angka urutan tampilan. Kosong akan mengikuti urutan baris."),
		}),
		templateInputSheet(greenBookImportSheetFundingSource, []string{"GB Code (*)", "Lender Name (*)", "Institution Name", "Loan USD", "Grant USD", "Local USD"}, []float64{22, 42, 48, 18, 18, 18}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddLenders", "Lender Name", "Pilih lender dari master data."),
			listValidation("C2:C"+inputLastRow(), "ddInstitutions", "Institution Name", "Pilih institution terkait funding source bila ada."),
			decimalValidation("D2:D"+inputLastRow(), "Loan USD", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
			decimalValidation("E2:E"+inputLastRow(), "Grant USD", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
			decimalValidation("F2:F"+inputLastRow(), "Local USD", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
		}),
		templateInputSheet(greenBookImportSheetDisbursementPlan, []string{"GB Code (*)", "Year (*)", "Amount USD"}, []float64{22, 18, 20}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			numberValidation("B2:B"+inputLastRow(), "Year", "Isi tahun anggaran dalam angka empat digit. Tahun harus unik per GB Code."),
			decimalValidation("C2:C"+inputLastRow(), "Amount USD", "Isi total rencana penarikan proyek per tahun, bukan per lender."),
		}),
		templateInputSheet(greenBookImportSheetFundingAllocation, []string{"GB Code (*)", "Activity No (*)", "Services", "Constructions", "Goods", "Trainings", "Other"}, []float64{22, 18, 18, 20, 18, 18, 18}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddInputGBCodes", "GB Code", "Pilih GB Code dari sheet Input Data."),
			listValidation("B2:B"+inputLastRow(), "ddInputActivityNos", "Activity No", "Pilih Activity No dari sheet Relasi - Activities."),
			decimalValidation("C2:C"+inputLastRow(), "Services", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
			decimalValidation("D2:D"+inputLastRow(), "Constructions", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
			decimalValidation("E2:E"+inputLastRow(), "Goods", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
			decimalValidation("F2:F"+inputLastRow(), "Trainings", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
			decimalValidation("G2:G"+inputLastRow(), "Other", "Isi angka dalam USD. Kosong akan dianggap 0 oleh backend."),
		}),
		dropdowns,
	}

	return simpleXLSXWorkbook{Sheets: sheets, DefinedNames: definedNames}, nil
}

func (s *DKService) buildDKImportTemplateWorkbook(ctx context.Context) (simpleXLSXWorkbook, error) {
	masterSvc := &MasterService{db: s.db, queries: s.queries}
	reference, err := masterSvc.loadImportTemplateReferenceData(ctx, pgtype.UUID{})
	if err != nil {
		return simpleXLSXWorkbook{}, err
	}

	dropdowns, definedNames := buildDropdownSheet(reference)
	definedNames = append(definedNames, simpleXLSXDefinedName{
		Name: "ddDKKeys",
		Ref:  xlsxRangeRef(dkImportSheetHeader, 1, 2, importTemplateEditableRows+1),
	}, simpleXLSXDefinedName{
		Name: "ddProjectKeys",
		Ref:  xlsxRangeRef(dkImportSheetInput, 2, 2, importTemplateEditableRows+1),
	})

	sheets := []simpleXLSXSheet{
		buildDKGuideSheet(),
		buildDKMasterDataSnapshotSheet("Master Data", reference),
		templateInputSheet(dkImportSheetHeader, []string{"DK Key (*)", "Letter Number (*)", "Subject (*)", "Date (*)"}, []float64{22, 34, 70, 18}, nil),
		templateInputSheet(dkImportSheetInput, []string{"DK Key (*)", "Project Key (*)", "Program Title", "Executing Agency Name (*)", "Duration", "Objectives"}, []float64{22, 24, 38, 48, 18, 64}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddDKKeys", "DK Key", "Pilih DK Key dari sheet Daftar Kegiatan."),
			listValidation("C2:C"+inputLastRow(), "ddProgramTitles", "Program Title", "Opsional. Jika diisi harus sesuai master Program Title."),
			listValidation("D2:D"+inputLastRow(), "ddInstitutions", "Executing Agency", "Pilih institution dari master data."),
		}),
		templateInputSheet(dkImportSheetGBProject, []string{"DK Key (*)", "Project Key (*)", "GB Code (*)"}, []float64{22, 24, 22}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddDKKeys", "DK Key", "Pilih DK Key dari sheet Daftar Kegiatan."),
			listValidation("B2:B"+inputLastRow(), "ddProjectKeys", "Project Key", "Pilih Project Key dari sheet Input Data."),
			listValidation("C2:C"+inputLastRow(), "ddGBProjectCodes", "GB Code", "Pilih GB Project active dari database."),
		}),
		templateInputSheet(dkImportSheetLocations, []string{"DK Key (*)", "Project Key (*)", "Location Name (*)"}, []float64{22, 24, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddDKKeys", "DK Key", "Pilih DK Key dari sheet Daftar Kegiatan."),
			listValidation("B2:B"+inputLastRow(), "ddProjectKeys", "Project Key", "Pilih Project Key dari sheet Input Data."),
			listValidation("C2:C"+inputLastRow(), "ddRegionNames", "Location Name", "Pilih region dari master data. Backend juga menerima kode region bila diketik manual."),
		}),
		templateInputSheet(dkImportSheetFinancingDetail, []string{"DK Key (*)", "Project Key (*)", "Lender Name (*)", "Currency", "Amount Original", "Grant Original", "Counterpart Original", "Amount USD", "Grant USD", "Counterpart USD", "Remarks"}, []float64{22, 24, 42, 14, 18, 18, 22, 18, 18, 22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddDKKeys", "DK Key", "Pilih DK Key dari sheet Daftar Kegiatan."),
			listValidation("B2:B"+inputLastRow(), "ddProjectKeys", "Project Key", "Pilih Project Key dari sheet Input Data."),
			listValidation("C2:C"+inputLastRow(), "ddLenders", "Lender Name", "Pilih lender dari master data. Backend memvalidasi lender harus berasal dari GB Project terkait."),
			listValidation("D2:D"+inputLastRow(), "ddCurrencies", "Currency", "Kosong akan dianggap USD. Jika diketik manual, gunakan kode 3 huruf."),
			decimalValidation("E2:J"+inputLastRow(), "Amount", "Isi angka 0 atau lebih. Kosong akan dianggap 0 oleh backend."),
		}),
		templateInputSheet(dkImportSheetLoanAllocation, []string{"DK Key (*)", "Project Key (*)", "Institution Name (*)", "Currency", "Amount Original", "Grant Original", "Counterpart Original", "Amount USD", "Grant USD", "Counterpart USD", "Remarks"}, []float64{22, 24, 48, 14, 18, 18, 22, 18, 18, 22, 48}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddDKKeys", "DK Key", "Pilih DK Key dari sheet Daftar Kegiatan."),
			listValidation("B2:B"+inputLastRow(), "ddProjectKeys", "Project Key", "Pilih Project Key dari sheet Input Data."),
			listValidation("C2:C"+inputLastRow(), "ddInstitutions", "Institution Name", "Pilih institution dari master data."),
			listValidation("D2:D"+inputLastRow(), "ddCurrencies", "Currency", "Kosong akan dianggap USD. Jika diketik manual, gunakan kode 3 huruf."),
			decimalValidation("E2:J"+inputLastRow(), "Amount", "Isi angka 0 atau lebih. Kosong akan dianggap 0 oleh backend."),
		}),
		templateInputSheet(dkImportSheetActivityDetail, []string{"DK Key (*)", "Project Key (*)", "Activity No (*)", "Activity Name (*)"}, []float64{22, 24, 18, 64}, []simpleXLSXValidation{
			listValidation("A2:A"+inputLastRow(), "ddDKKeys", "DK Key", "Pilih DK Key dari sheet Daftar Kegiatan."),
			listValidation("B2:B"+inputLastRow(), "ddProjectKeys", "Project Key", "Pilih Project Key dari sheet Input Data."),
			integerValidation("C2:C"+inputLastRow(), "Activity No", "Isi nomor aktivitas yang unik per Project Key."),
		}),
		dropdowns,
	}

	return simpleXLSXWorkbook{Sheets: sheets, DefinedNames: definedNames}, nil
}

func (s *MasterService) loadImportTemplateReferenceData(ctx context.Context, periodID pgtype.UUID) (*importTemplateReferenceData, error) {
	programTitles, err := s.queries.ListProgramTitles(ctx, queries.ListProgramTitlesParams{Limit: masterImportListLimit, Offset: 0})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot program title")
	}

	partners, err := s.queries.ListBappenasPartners(ctx, queries.ListBappenasPartnersParams{Limit: masterImportListLimit, Offset: 0})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot bappenas partner")
	}

	institutions, err := s.queries.ListInstitutions(ctx, queries.ListInstitutionsParams{Limit: masterImportListLimit, Offset: 0, LevelFilter: pgtype.Text{}, ParentIDFilter: pgtype.UUID{}})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot institution")
	}

	regions, err := s.queries.ListRegions(ctx, queries.ListRegionsParams{Limit: masterImportListLimit, Offset: 0, TypeFilter: pgtype.Text{}, ParentCodeFilter: pgtype.Text{}})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot region")
	}

	periods, err := s.queries.ListPeriods(ctx, queries.ListPeriodsParams{Limit: masterImportListLimit, Offset: 0})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot period")
	}

	priorities, err := s.queries.ListNationalPriorities(ctx, queries.ListNationalPrioritiesParams{Limit: masterImportListLimit, Offset: 0, PeriodIDFilter: periodID})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot national priority")
	}

	lenders, err := s.queries.ListLenders(ctx, queries.ListLendersParams{Limit: masterImportListLimit, Offset: 0, TypeFilter: pgtype.Text{}})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot lender")
	}

	countries, err := s.queries.ListCountries(ctx, queries.ListCountriesParams{Limit: masterImportListLimit, Offset: 0})
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot country")
	}

	bbProjects, err := s.queries.ListActiveBBProjectReferences(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot BB Project")
	}

	gbProjects, err := s.queries.ListActiveGBProjectReferences(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal membaca snapshot GB Project")
	}

	allowedGBLenders := make([]queries.ListAllowedLenderReferencesByGBProjectRow, 0)
	for _, project := range gbProjects {
		items, err := s.queries.ListAllowedLenderReferencesByGBProject(ctx, project.ID)
		if err != nil {
			return nil, apperrors.Internal("Gagal membaca snapshot allowed lender GB Project")
		}
		allowedGBLenders = append(allowedGBLenders, items...)
	}

	return &importTemplateReferenceData{
		ProgramTitles:      programTitles,
		BappenasPartners:   partners,
		Institutions:       institutions,
		Regions:            regions,
		Periods:            periods,
		NationalPriorities: priorities,
		Lenders:            lenders,
		Countries:          countries,
		BBProjects:         bbProjects,
		GBProjects:         gbProjects,
		AllowedGBLenders:   allowedGBLenders,
	}, nil
}

func buildMasterDataSnapshotSheet(name string, reference *importTemplateReferenceData) simpleXLSXSheet {
	rows := [][]simpleXLSXCell{
		styledTextRow(xlsxStyleTitle, name),
		styledTextRow(xlsxStyleSubtitle, "Snapshot ini diambil dari database saat template diunduh. Gunakan kolom Name/Title atau Code sebagai referensi pengisian dropdown dan kolom relasi."),
		textRow(""),
		headerRow("Entity", "ID/Code", "Name/Title", "Level/Type", "Parent/Period", "Extra"),
	}

	for _, item := range reference.ProgramTitles {
		rows = append(rows, textRow("Program Title", model.UUIDToString(item.ID), item.Title, "", uuidText(item.ParentID), ""))
	}
	for _, item := range reference.BappenasPartners {
		rows = append(rows, textRow("Bappenas Partner", model.UUIDToString(item.ID), item.Name, item.Level, uuidText(item.ParentID), ""))
	}
	for _, item := range reference.Institutions {
		rows = append(rows, textRow("Institution", model.UUIDToString(item.ID), item.Name, item.Level, uuidText(item.ParentID), textFromPg(item.ShortName)))
	}
	for _, item := range reference.Regions {
		rows = append(rows, textRow("Region", item.Code, item.Name, item.Type, textFromPg(item.ParentCode), ""))
	}
	for _, item := range reference.Periods {
		rows = append(rows, []simpleXLSXCell{
			textCell("Period"),
			textCell(model.UUIDToString(item.ID)),
			textCell(item.Name),
			numberCell(item.YearStart),
			numberCell(item.YearEnd),
			textCell(""),
		})
	}
	for _, item := range reference.NationalPriorities {
		rows = append(rows, textRow("National Priority", model.UUIDToString(item.ID), item.Title, "", item.PeriodName, ""))
	}
	for _, item := range reference.Lenders {
		rows = append(rows, textRow("Lender", model.UUIDToString(item.ID), item.Name, item.Type, textFromPg(item.CountryName), textFromPg(item.ShortName)))
	}
	for _, item := range reference.Countries {
		rows = append(rows, textRow("Country", item.Code, item.Name, "", "", ""))
	}

	return simpleXLSXSheet{
		Name:          name,
		Rows:          rows,
		Columns:       columns(18, 42, 52, 24, 36, 36),
		AutoFilter:    fmt.Sprintf("A4:F%d", len(rows)),
		FreezeRows:    4,
		ShowGridLines: false,
	}
}

func buildGreenBookMasterDataSnapshotSheet(name string, reference *importTemplateReferenceData) simpleXLSXSheet {
	sheet := buildMasterDataSnapshotSheet(name, reference)
	for _, item := range reference.BBProjects {
		sheet.Rows = append(sheet.Rows, textRow("BB Project", item.BbCode, item.ProjectName, "Active", item.PeriodName, model.UUIDToString(item.ID)))
	}
	sheet.AutoFilter = fmt.Sprintf("A4:F%d", len(sheet.Rows))
	return sheet
}

func buildDKMasterDataSnapshotSheet(name string, reference *importTemplateReferenceData) simpleXLSXSheet {
	sheet := buildMasterDataSnapshotSheet(name, reference)
	gbCodesByID := map[string]string{}
	for _, item := range reference.GBProjects {
		id := model.UUIDToString(item.ID)
		gbCodesByID[id] = item.GbCode
		sheet.Rows = append(sheet.Rows, textRow("GB Project", item.GbCode, item.ProjectName, "Active", fmt.Sprintf("GB %d Revisi %d", item.PublishYear, item.RevisionNumber), id))
	}
	for _, item := range reference.AllowedGBLenders {
		gbCode := gbCodesByID[model.UUIDToString(item.GbProjectID)]
		sheet.Rows = append(sheet.Rows, textRow("Allowed Lender", gbCode, item.LenderName, item.LenderType, "", model.UUIDToString(item.LenderID)))
	}
	sheet.AutoFilter = fmt.Sprintf("A4:F%d", len(sheet.Rows))
	return sheet
}

func buildMasterGuideSheet() simpleXLSXSheet {
	rows := [][]simpleXLSXCell{
		styledTextRow(xlsxStyleTitle, "Panduan Import Master Data"),
		styledTextRow(xlsxStyleSubtitle, "Gunakan template ini untuk menambah master data PRISM secara massal. Jalankan Preview terlebih dahulu; eksekusi hanya dilakukan setelah tidak ada baris failed."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Alur Aman", "Deskripsi", "Catatan"),
		textRow("1. Unduh template terbaru", "Template membawa snapshot master data dan dropdown terbaru dari database.", "Jangan gunakan template lama jika master data sering berubah."),
		textRow("2. Isi sheet import", "Isi hanya sheet yang sesuai kebutuhan. Header di baris pertama jangan diubah nama atau urutannya.", "Kolom bertanda (*) wajib diisi pada baris yang digunakan."),
		textRow("3. Gunakan dropdown", "Kolom relasi memakai dropdown dari master data existing agar penulisan nama konsisten.", "Jika perlu membuat parent baru, buat parent pada sheet terkait sebelum child."),
		textRow("4. Preview", "Upload workbook lalu klik Preview untuk melihat create, skip, dan failed.", "Preview tidak menyimpan data."),
		textRow("5. Eksekusi", "Klik Eksekusi hanya setelah failed = 0.", "Baris skip tidak dibuat ulang karena sudah ada di database."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Sheet", "Kolom Wajib", "Panduan Pengisian"),
		textRow("Program Titles", "Title (*)", "Parent Title opsional. Pilih dari dropdown bila parent sudah ada atau dibuat sebagai baris parent tanpa Parent Title."),
		textRow("Bappenas Partners", "Name (*), Level (*)", "Level hanya Eselon I/Eselon II. Eselon II wajib memilih Parent Name Eselon I."),
		textRow("Institutions", "Name (*), Level (*)", "Short Name opsional. Parent Name dipakai untuk struktur berjenjang."),
		textRow("Regions", "Code (*), Name (*), Level (*)", "Level gunakan COUNTRY, PROVINCE, atau CITY. Parent Code wajib jika region punya parent."),
		textRow("Periods", "Name (*), Year Start (*), Year End (*)", "Year End harus lebih besar dari Year Start."),
		textRow("National Priorities", "Period Name (*), Title (*)", "Period Name pilih dari sheet Periods atau period yang sudah ada di database."),
		textRow("Lenders", "Name (*), Type (*)", "Bilateral dan KSA wajib Country Name; Multilateral harus dikosongkan Country Name."),
		textRow(""),
		styledTextRow(xlsxStyleNote, "Catatan", "Sheet Master Data Snapshot hanya referensi. Sheet _Dropdowns disembunyikan dan dipakai Excel untuk pilihan dropdown.", ""),
	}
	return simpleXLSXSheet{
		Name:          "Panduan",
		Rows:          rows,
		Columns:       columns(28, 72, 72),
		FreezeRows:    0,
		ShowGridLines: false,
	}
}

func buildBlueBookGuideSheet(blueBook queries.GetBlueBookRow) simpleXLSXSheet {
	rows := [][]simpleXLSXCell{
		styledTextRow(xlsxStyleTitle, "Panduan Import Proyek Blue Book"),
		styledTextRow(xlsxStyleSubtitle, "Target Blue Book", blueBook.PeriodName, "National Priority pada template ini menampilkan seluruh master data dan tidak dibatasi period target."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Alur Aman", "Deskripsi", "Catatan"),
		textRow("1. Isi Input Data", "Satu baris mewakili satu proyek. BB Code menjadi kunci penghubung ke semua sheet relasi.", "BB Code yang sudah ada di database akan masuk status skip."),
		textRow("2. Isi sheet relasi", "Gunakan BB Code yang sama dengan Input Data. Ulangi BB Code di beberapa baris untuk multi EA/IA/lokasi/prioritas/lender.", "BB Code pada sheet relasi punya dropdown dari Input Data."),
		textRow("3. Gunakan Master Data", "Kolom Program Title, Bappenas Partner, Institution, Location, National Priority, dan Lender memakai dropdown dari master data.", "Jika pilihan tidak ada, tambahkan dulu melalui menu Master Data atau import Master Data."),
		textRow("4. Preview", "Upload workbook lalu klik Preview untuk memisahkan create, skip, dan failed.", "Preview tidak menyimpan data."),
		textRow("5. Eksekusi", "Klik Eksekusi hanya jika tidak ada failed.", "Data dibuat dalam satu transaksi."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Sheet", "Kolom Wajib", "Panduan Pengisian"),
		textRow("Input Data", "Program Title (*), BB Code (*), Project Name (*)", "Duration dan uraian proyek opsional. Bappenas Partner pilih Eselon II bila tersedia."),
		textRow("Relasi - EA", "BB Code (*), Executing Agency Name (*)", "Minimal satu EA wajib untuk proyek baru."),
		textRow("Relasi - IA", "BB Code (*), Implementing Agency Name (*)", "Minimal satu IA wajib. Institution yang sama boleh dipakai sebagai EA dan IA pada proyek yang sama."),
		textRow("Relasi - Locations", "BB Code (*), Location Name (*)", "Minimal satu lokasi wajib. Pilih nama region dari dropdown."),
		textRow("Relasi - National Priority", "BB Code (*), National Priority Name (*)", "Pilih national priority dari seluruh master data, tidak dibatasi period Blue Book target."),
		textRow("Relasi - Project Cost", "BB Code (*), Funding Type (*), Funding Category (*)", "Amount USD angka. Funding Type: Foreign/Counterpart."),
		textRow("Relasi - Lender Indication", "BB Code (*), Lender Name (*)", "Keterangan opsional dan akan disimpan sebagai remarks lender indication."),
		textRow(""),
		styledTextRow(xlsxStyleNote, "Catatan", "Sheet Master Data adalah snapshot referensi. Sheet _Dropdowns disembunyikan dan dipakai Excel untuk pilihan dropdown.", ""),
	}
	return simpleXLSXSheet{
		Name:          "Panduan",
		Rows:          rows,
		Columns:       columns(30, 72, 74),
		ShowGridLines: false,
	}
}

func buildGreenBookGuideSheet(greenBook queries.GreenBook) simpleXLSXSheet {
	rows := [][]simpleXLSXCell{
		styledTextRow(xlsxStyleTitle, "Panduan Import Proyek Green Book"),
		styledTextRow(xlsxStyleSubtitle, "Target Green Book", fmt.Sprintf("GB %d Revisi %d", greenBook.PublishYear, greenBook.RevisionNumber), "Workbook ini menambah GB Project ke target Green Book yang dipilih."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Alur Aman", "Deskripsi", "Catatan"),
		textRow("1. Isi Input Data", "Satu baris mewakili satu GB Project. GB Code menjadi kunci penghubung ke semua sheet relasi.", "GB Code yang sudah ada di database akan masuk status skip."),
		textRow("2. Isi sheet relasi", "Gunakan GB Code yang sama dengan Input Data. Ulangi GB Code untuk multi BB Project, EA/IA, lokasi, lender, dan rencana tahunan.", "GB Code pada sheet relasi punya dropdown dari Input Data."),
		textRow("3. Activities dan allocation", "Activity No wajib unik per GB Code dan dipakai oleh sheet Funding Allocation.", "Activity tanpa Funding Allocation eksplisit tetap dibuat dengan allocation 0 agar sinkron."),
		textRow("4. Preview", "Upload workbook lalu klik Preview untuk memisahkan create, skip, dan failed.", "Preview tidak menyimpan data."),
		textRow("5. Eksekusi", "Klik Eksekusi hanya jika tidak ada failed.", "Data dibuat dalam satu transaksi."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Sheet", "Kolom Wajib", "Panduan Pengisian"),
		textRow("Input Data", "Program Title (*), GB Code (*), Project Name (*)", "Duration dan uraian proyek opsional."),
		textRow("Relasi - BB Project", "GB Code (*), BB Code (*)", "Minimal satu BB Project active wajib untuk proyek baru."),
		textRow("Relasi - EA", "GB Code (*), Executing Agency Name (*)", "Minimal satu EA wajib."),
		textRow("Relasi - IA", "GB Code (*), Implementing Agency Name (*)", "Minimal satu IA wajib."),
		textRow("Relasi - Locations", "GB Code (*), Location Name (*)", "Minimal satu lokasi wajib. Pilih nama region atau ketik kode region."),
		textRow("Relasi - Activities", "GB Code (*), Activity No (*), Activity Name (*)", "Activity No unik per GB Code. Sort Order opsional."),
		textRow("Relasi - Funding Source", "GB Code (*), Lender Name (*)", "Loan/Grant/Local USD kosong dianggap 0."),
		textRow("Relasi - Disbursement Plan", "GB Code (*), Year (*)", "Year harus unik per GB Code. Amount USD adalah total proyek per tahun."),
		textRow("Relasi - Funding Allocation", "GB Code (*), Activity No (*)", "Isi breakdown per Activity No. Jika tidak diisi, allocation dibuat 0."),
		textRow(""),
		styledTextRow(xlsxStyleNote, "Catatan", "Sheet Master Data adalah snapshot referensi. Sheet _Dropdowns disembunyikan dan dipakai Excel untuk pilihan dropdown.", ""),
	}
	return simpleXLSXSheet{
		Name:          "Panduan",
		Rows:          rows,
		Columns:       columns(30, 72, 74),
		ShowGridLines: false,
	}
}

func buildDKGuideSheet() simpleXLSXSheet {
	rows := [][]simpleXLSXCell{
		styledTextRow(xlsxStyleTitle, "Panduan Import Daftar Kegiatan"),
		styledTextRow(xlsxStyleSubtitle, "Workbook ini membuat header Daftar Kegiatan baru beserta DK Project dan seluruh relasinya. Letter Number wajib untuk import dan menjadi kunci skip jika sudah ada di database."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Alur Aman", "Deskripsi", "Catatan"),
		textRow("1. Isi Daftar Kegiatan", "Satu baris mewakili satu header DK. DK Key adalah kunci sementara workbook untuk menghubungkan project dan relasi.", "DK Key wajib unik. Letter Number duplikat di workbook gagal; Letter Number existing di DB masuk skip."),
		textRow("2. Isi Input Data", "Satu baris mewakili satu DK Project. Project Key wajib unik di dalam DK Key yang sama.", "Program Title opsional, tetapi jika diisi harus ada di master data."),
		textRow("3. Isi sheet relasi", "Gunakan DK Key dan Project Key yang sama dengan Input Data untuk GB Project, lokasi, pembiayaan, alokasi, dan aktivitas.", "Project baru wajib memiliki minimal satu baris di setiap sheet relasi."),
		textRow("4. Lender", "Financing Detail hanya boleh memakai lender yang berasal dari GB Funding Source atau BB Lender Indication pada GB Project terkait.", "Cek referensi Allowed Lender di sheet Master Data."),
		textRow("5. Preview dan Eksekusi", "Upload workbook lalu Preview untuk melihat create, skip, dan failed. Eksekusi hanya jika failed = 0.", "Preview tidak menyimpan data; eksekusi menyimpan dalam satu transaksi."),
		textRow(""),
		styledTextRow(xlsxStyleSection, "Sheet", "Kolom Wajib", "Panduan Pengisian"),
		textRow("Daftar Kegiatan", "DK Key (*), Letter Number (*), Subject (*), Date (*)", "Date isi format YYYY-MM-DD. Header existing by Letter Number dilewati."),
		textRow("Input Data", "DK Key (*), Project Key (*), Executing Agency Name (*)", "Duration dan Objectives opsional. Project Key hanya disimpan sebagai kunci workbook."),
		textRow("Relasi - GB Project", "DK Key (*), Project Key (*), GB Code (*)", "Minimal satu GB Project active wajib untuk project baru."),
		textRow("Relasi - Locations", "DK Key (*), Project Key (*), Location Name (*)", "Minimal satu lokasi wajib. Pilih nama region atau ketik kode region."),
		textRow("Relasi - Financing Detail", "DK Key (*), Project Key (*), Lender Name (*)", "Currency kosong dianggap USD. Amount kosong dianggap 0 dan tidak boleh negatif."),
		textRow("Relasi - Loan Allocation", "DK Key (*), Project Key (*), Institution Name (*)", "Currency kosong dianggap USD. Amount kosong dianggap 0 dan tidak boleh negatif."),
		textRow("Relasi - Activity Detail", "DK Key (*), Project Key (*), Activity No (*), Activity Name (*)", "Activity No wajib unik per project."),
		textRow(""),
		styledTextRow(xlsxStyleNote, "Catatan", "Sheet Master Data adalah snapshot referensi. Sheet _Dropdowns disembunyikan dan dipakai Excel untuk pilihan dropdown.", ""),
	}
	return simpleXLSXSheet{
		Name:          "Panduan",
		Rows:          rows,
		Columns:       columns(30, 78, 78),
		ShowGridLines: false,
	}
}

func templateInputSheet(name string, headers []string, widths []float64, validations []simpleXLSXValidation) simpleXLSXSheet {
	return simpleXLSXSheet{
		Name:          name,
		Rows:          [][]simpleXLSXCell{headerRow(headers...)},
		Columns:       columns(widths...),
		Validations:   validations,
		AutoFilter:    fmt.Sprintf("A1:%s1", xlsxColumnName(len(headers)-1)),
		FreezeRows:    1,
		ShowGridLines: false,
	}
}

func buildDropdownSheet(reference *importTemplateReferenceData) (simpleXLSXSheet, []simpleXLSXDefinedName) {
	type dropdownColumn struct {
		Name   string
		Header string
		Values []string
	}

	columnsData := []dropdownColumn{
		{Name: "ddProgramTitles", Header: "Program Titles", Values: uniqueSorted(programTitleValues(reference.ProgramTitles))},
		{Name: "ddBappenasPartners", Header: "Bappenas Partners", Values: uniqueSorted(bappenasPartnerValues(reference.BappenasPartners, ""))},
		{Name: "ddBappenasPartnerParents", Header: "Bappenas Partner Parents", Values: uniqueSorted(bappenasPartnerValues(reference.BappenasPartners, "Eselon I"))},
		{Name: "ddBappenasPartnersEselonII", Header: "Bappenas Partner Eselon II", Values: uniqueSorted(bappenasPartnerValues(reference.BappenasPartners, "Eselon II"))},
		{Name: "ddInstitutions", Header: "Institutions", Values: uniqueSorted(institutionValues(reference.Institutions))},
		{Name: "ddRegionNames", Header: "Region Names", Values: uniqueSorted(regionNameValues(reference.Regions))},
		{Name: "ddRegionCodes", Header: "Region Codes", Values: uniqueSorted(regionCodeValues(reference.Regions))},
		{Name: "ddPeriods", Header: "Periods", Values: uniqueSorted(periodValues(reference.Periods))},
		{Name: "ddNationalPriorities", Header: "National Priorities", Values: uniqueSorted(nationalPriorityValues(reference.NationalPriorities))},
		{Name: "ddLenders", Header: "Lenders", Values: uniqueSorted(lenderValues(reference.Lenders))},
		{Name: "ddCountries", Header: "Countries", Values: uniqueSorted(countryValues(reference.Countries))},
		{Name: "ddBBProjectCodes", Header: "BB Project Codes", Values: uniqueSorted(bbProjectCodeValues(reference.BBProjects))},
		{Name: "ddGBProjectCodes", Header: "GB Project Codes", Values: uniqueSorted(gbProjectCodeValues(reference.GBProjects))},
		{Name: "ddBappenasPartnerLevels", Header: "Bappenas Partner Levels", Values: []string{"Eselon I", "Eselon II"}},
		{Name: "ddInstitutionLevels", Header: "Institution Levels", Values: []string{"Kementerian/Badan/Lembaga", "Eselon I", "BUMN", "Pemerintah Daerah", "BUMD", "Lainnya"}},
		{Name: "ddRegionLevels", Header: "Region Levels", Values: []string{"COUNTRY", "PROVINCE", "CITY"}},
		{Name: "ddLenderTypes", Header: "Lender Types", Values: []string{"Bilateral", "Multilateral", "KSA"}},
		{Name: "ddCurrencies", Header: "Currencies", Values: []string{"USD", "IDR", "JPY", "EUR", "AUD", "GBP", "CNY", "KRW"}},
		{Name: "ddFundingTypes", Header: "Funding Types", Values: []string{"Foreign", "Counterpart"}},
		{Name: "ddFundingCategories", Header: "Funding Categories", Values: []string{"Loan", "Grant", "Central Government", "Regional Government", "State-Owned Enterprise", "Others"}},
	}

	maxRows := 1
	for _, column := range columnsData {
		if len(column.Values)+1 > maxRows {
			maxRows = len(column.Values) + 1
		}
	}

	rows := make([][]simpleXLSXCell, maxRows)
	header := make([]simpleXLSXCell, 0, len(columnsData))
	for _, column := range columnsData {
		header = append(header, styledTextCell(column.Header, xlsxStyleHeader))
	}
	rows[0] = header

	for rowIndex := 1; rowIndex < maxRows; rowIndex++ {
		row := make([]simpleXLSXCell, len(columnsData))
		for colIndex, column := range columnsData {
			valueIndex := rowIndex - 1
			if valueIndex < len(column.Values) {
				row[colIndex] = textCell(column.Values[valueIndex])
			}
		}
		rows[rowIndex] = row
	}

	definedNames := make([]simpleXLSXDefinedName, 0, len(columnsData))
	for index, column := range columnsData {
		endRow := len(column.Values) + 1
		if endRow < 2 {
			endRow = 2
		}
		definedNames = append(definedNames, simpleXLSXDefinedName{
			Name: column.Name,
			Ref:  xlsxRangeRef("_Dropdowns", index+1, 2, endRow),
		})
	}

	return simpleXLSXSheet{
		Name:          "_Dropdowns",
		Rows:          rows,
		Columns:       repeatedColumns(len(columnsData), 28),
		FreezeRows:    1,
		Hidden:        true,
		ShowGridLines: false,
	}, definedNames
}

func listValidation(cellRange, definedName, title, prompt string) simpleXLSXValidation {
	return simpleXLSXValidation{
		Range:       cellRange,
		Type:        "list",
		Formula1:    definedName,
		PromptTitle: title,
		Prompt:      prompt,
		ErrorTitle:  "Pilihan tidak valid",
		Error:       "Pilih nilai dari dropdown atau kosongkan seluruh baris jika tidak digunakan.",
		ErrorStyle:  "warning",
		AllowBlank:  true,
	}
}

func numberValidation(cellRange, title, prompt string) simpleXLSXValidation {
	return simpleXLSXValidation{
		Range:       cellRange,
		Type:        "whole",
		Operator:    "between",
		Formula1:    "1900",
		Formula2:    "2200",
		PromptTitle: title,
		Prompt:      prompt,
		ErrorTitle:  "Angka tidak valid",
		Error:       "Isi angka tahun empat digit.",
		ErrorStyle:  "stop",
		AllowBlank:  true,
	}
}

func integerValidation(cellRange, title, prompt string) simpleXLSXValidation {
	return simpleXLSXValidation{
		Range:       cellRange,
		Type:        "whole",
		Operator:    "greaterThanOrEqual",
		Formula1:    "0",
		PromptTitle: title,
		Prompt:      prompt,
		ErrorTitle:  "Angka tidak valid",
		Error:       "Isi angka 0 atau lebih besar.",
		ErrorStyle:  "stop",
		AllowBlank:  true,
	}
}

func decimalValidation(cellRange, title, prompt string) simpleXLSXValidation {
	return simpleXLSXValidation{
		Range:       cellRange,
		Type:        "decimal",
		Operator:    "greaterThanOrEqual",
		Formula1:    "0",
		PromptTitle: title,
		Prompt:      prompt,
		ErrorTitle:  "Amount tidak valid",
		Error:       "Isi angka 0 atau lebih besar.",
		ErrorStyle:  "stop",
		AllowBlank:  true,
	}
}

func inputLastRow() string {
	return strconv.Itoa(importTemplateEditableRows + 1)
}

func columns(widths ...float64) []simpleXLSXColumn {
	result := make([]simpleXLSXColumn, 0, len(widths))
	for index, width := range widths {
		result = append(result, simpleXLSXColumn{Min: index + 1, Max: index + 1, Width: width})
	}
	return result
}

func repeatedColumns(count int, width float64) []simpleXLSXColumn {
	result := make([]simpleXLSXColumn, 0, count)
	for index := 0; index < count; index++ {
		result = append(result, simpleXLSXColumn{Min: index + 1, Max: index + 1, Width: width})
	}
	return result
}

func uniqueSorted(values []string) []string {
	seen := map[string]string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := normalizeLookupKey(trimmed)
		if _, exists := seen[key]; !exists {
			seen[key] = trimmed
		}
	}
	result := make([]string, 0, len(seen))
	for _, value := range seen {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i]) < strings.ToLower(result[j])
	})
	return result
}

func programTitleValues(items []queries.ProgramTitle) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Title)
	}
	return values
}

func bappenasPartnerValues(items []queries.BappenasPartner, level string) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		if level != "" && item.Level != level {
			continue
		}
		values = append(values, item.Name)
	}
	return values
}

func institutionValues(items []queries.Institution) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name)
	}
	return values
}

func regionNameValues(items []queries.Region) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name)
	}
	return values
}

func regionCodeValues(items []queries.Region) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Code)
	}
	return values
}

func periodValues(items []queries.Period) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name)
	}
	return values
}

func nationalPriorityValues(items []queries.ListNationalPrioritiesRow) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Title)
	}
	return values
}

func lenderValues(items []queries.ListLendersRow) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name)
	}
	return values
}

func countryValues(items []queries.Country) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.Name)
	}
	return values
}

func bbProjectCodeValues(items []queries.ListActiveBBProjectReferencesRow) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.BbCode)
	}
	return values
}

func gbProjectCodeValues(items []queries.ListActiveGBProjectReferencesRow) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, item.GbCode)
	}
	return values
}

func buildSimpleXLSX(workbook simpleXLSXWorkbook) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	if err := writeZipString(writer, "[Content_Types].xml", xlsxContentTypes(workbook.Sheets)); err != nil {
		return nil, err
	}
	if err := writeZipString(writer, "_rels/.rels", xlsxRootRels()); err != nil {
		return nil, err
	}
	if err := writeZipString(writer, "xl/workbook.xml", xlsxWorkbookXML(workbook)); err != nil {
		return nil, err
	}
	if err := writeZipString(writer, "xl/_rels/workbook.xml.rels", xlsxWorkbookRels(workbook.Sheets)); err != nil {
		return nil, err
	}
	if err := writeZipString(writer, "xl/styles.xml", xlsxStylesXML()); err != nil {
		return nil, err
	}
	for index, sheet := range workbook.Sheets {
		if err := writeZipString(writer, fmt.Sprintf("xl/worksheets/sheet%d.xml", index+1), xlsxWorksheetXML(sheet)); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func writeZipString(writer *zip.Writer, name, content string) error {
	file, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(content))
	return err
}

func xlsxContentTypes(sheets []simpleXLSXSheet) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	builder.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	builder.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	builder.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	builder.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for index := range sheets {
		builder.WriteString(fmt.Sprintf(`<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, index+1))
	}
	builder.WriteString(`</Types>`)
	return builder.String()
}

func xlsxRootRels() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`
}

func xlsxWorkbookXML(workbook simpleXLSXWorkbook) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for index, sheet := range workbook.Sheets {
		state := ""
		if sheet.Hidden {
			state = ` state="hidden"`
		}
		builder.WriteString(fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"%s/>`, xmlAttr(sheet.Name), index+1, index+1, state))
	}
	builder.WriteString(`</sheets>`)
	if len(workbook.DefinedNames) > 0 {
		builder.WriteString(`<definedNames>`)
		for _, name := range workbook.DefinedNames {
			builder.WriteString(fmt.Sprintf(`<definedName name="%s">%s</definedName>`, xmlAttr(name.Name), xmlText(name.Ref)))
		}
		builder.WriteString(`</definedNames>`)
	}
	builder.WriteString(`</workbook>`)
	return builder.String()
}

func xlsxWorkbookRels(sheets []simpleXLSXSheet) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for index := range sheets {
		builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, index+1, index+1))
	}
	builder.WriteString(fmt.Sprintf(`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, len(sheets)+1))
	builder.WriteString(`</Relationships>`)
	return builder.String()
}

func xlsxStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="4">
    <font><sz val="11"/><color theme="1"/><name val="Calibri"/><family val="2"/></font>
    <font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Calibri"/><family val="2"/></font>
    <font><b/><sz val="16"/><color rgb="FF0F172A"/><name val="Calibri"/><family val="2"/></font>
    <font><i/><sz val="10"/><color rgb="FF475569"/><name val="Calibri"/><family val="2"/></font>
  </fonts>
  <fills count="6">
    <fill><patternFill patternType="none"/></fill>
    <fill><patternFill patternType="gray125"/></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF0F766E"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FF164E63"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFE0F2FE"/><bgColor indexed="64"/></patternFill></fill>
    <fill><patternFill patternType="solid"><fgColor rgb="FFFEF9C3"/><bgColor indexed="64"/></patternFill></fill>
  </fills>
  <borders count="2">
    <border><left/><right/><top/><bottom/><diagonal/></border>
    <border><left style="thin"><color rgb="FFCBD5E1"/></left><right style="thin"><color rgb="FFCBD5E1"/></right><top style="thin"><color rgb="FFCBD5E1"/></top><bottom style="thin"><color rgb="FFCBD5E1"/></bottom><diagonal/></border>
  </borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="6">
    <xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
    <xf numFmtId="0" fontId="2" fillId="0" borderId="0" xfId="0" applyFont="1"><alignment vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="3" fillId="0" borderId="0" xfId="0" applyFont="1"><alignment vertical="top" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="1" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
    <xf numFmtId="0" fontId="0" fillId="5" borderId="1" xfId="0" applyFill="1" applyBorder="1"><alignment vertical="top" wrapText="1"/></xf>
  </cellXfs>
  <cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
  <dxfs count="0"/>
  <tableStyles count="0" defaultTableStyle="TableStyleMedium2" defaultPivotStyle="PivotStyleLight16"/>
</styleSheet>`
}

func xlsxWorksheetXML(sheet simpleXLSXSheet) string {
	maxCols := 1
	for _, row := range sheet.Rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	maxRows := len(sheet.Rows)
	if maxRows == 0 {
		maxRows = 1
	}

	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	builder.WriteString(fmt.Sprintf(`<dimension ref="A1:%s%d"/>`, xlsxColumnName(maxCols-1), maxRows))
	showGridLines := "0"
	if sheet.ShowGridLines {
		showGridLines = "1"
	}
	builder.WriteString(fmt.Sprintf(`<sheetViews><sheetView showGridLines="%s" workbookViewId="0">`, showGridLines))
	if sheet.FreezeRows > 0 {
		topLeftCell := fmt.Sprintf("A%d", sheet.FreezeRows+1)
		builder.WriteString(fmt.Sprintf(`<pane ySplit="%d" topLeftCell="%s" activePane="bottomLeft" state="frozen"/><selection pane="bottomLeft"/>`, sheet.FreezeRows, topLeftCell))
	}
	builder.WriteString(`</sheetView></sheetViews>`)
	builder.WriteString(`<sheetFormatPr defaultRowHeight="18"/>`)
	if len(sheet.Columns) > 0 {
		builder.WriteString(`<cols>`)
		for _, column := range sheet.Columns {
			builder.WriteString(fmt.Sprintf(`<col min="%d" max="%d" width="%.2f" customWidth="1"/>`, column.Min, column.Max, column.Width))
		}
		builder.WriteString(`</cols>`)
	}
	builder.WriteString(`<sheetData>`)
	for rowIndex, row := range sheet.Rows {
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowIndex+1))
		for colIndex, cell := range row {
			if cell.Value == "" && !cell.Number && cell.Style == xlsxStyleDefault {
				continue
			}
			ref := fmt.Sprintf("%s%d", xlsxColumnName(colIndex), rowIndex+1)
			style := ""
			if cell.Style > xlsxStyleDefault {
				style = fmt.Sprintf(` s="%d"`, cell.Style)
			}
			if cell.Number {
				builder.WriteString(fmt.Sprintf(`<c r="%s"%s><v>%s</v></c>`, ref, style, xmlText(cell.Value)))
				continue
			}
			builder.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr"%s><is><t xml:space="preserve">%s</t></is></c>`, ref, style, xmlText(cell.Value)))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData>`)
	if sheet.AutoFilter != "" {
		builder.WriteString(fmt.Sprintf(`<autoFilter ref="%s"/>`, xmlAttr(sheet.AutoFilter)))
	}
	if len(sheet.Validations) > 0 {
		builder.WriteString(fmt.Sprintf(`<dataValidations count="%d">`, len(sheet.Validations)))
		for _, validation := range sheet.Validations {
			builder.WriteString(xlsxDataValidationXML(validation))
		}
		builder.WriteString(`</dataValidations>`)
	}
	builder.WriteString(`</worksheet>`)
	return builder.String()
}

func xlsxDataValidationXML(validation simpleXLSXValidation) string {
	validationType := validation.Type
	if validationType == "" {
		validationType = "list"
	}
	allowBlank := "0"
	if validation.AllowBlank {
		allowBlank = "1"
	}
	operator := ""
	if validation.Operator != "" {
		operator = fmt.Sprintf(` operator="%s"`, xmlAttr(validation.Operator))
	}
	errorStyle := validation.ErrorStyle
	if errorStyle == "" {
		errorStyle = "stop"
	}
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(`<dataValidation type="%s"%s allowBlank="%s" showInputMessage="1" showErrorMessage="1" errorStyle="%s" sqref="%s"`, xmlAttr(validationType), operator, allowBlank, xmlAttr(errorStyle), xmlAttr(validation.Range)))
	if validation.PromptTitle != "" {
		builder.WriteString(fmt.Sprintf(` promptTitle="%s"`, xmlAttr(validation.PromptTitle)))
	}
	if validation.Prompt != "" {
		builder.WriteString(fmt.Sprintf(` prompt="%s"`, xmlAttr(validation.Prompt)))
	}
	if validation.ErrorTitle != "" {
		builder.WriteString(fmt.Sprintf(` errorTitle="%s"`, xmlAttr(validation.ErrorTitle)))
	}
	if validation.Error != "" {
		builder.WriteString(fmt.Sprintf(` error="%s"`, xmlAttr(validation.Error)))
	}
	builder.WriteString(`>`)
	if validation.Formula1 != "" {
		builder.WriteString(fmt.Sprintf(`<formula1>%s</formula1>`, xmlText(validation.Formula1)))
	}
	if validation.Formula2 != "" {
		builder.WriteString(fmt.Sprintf(`<formula2>%s</formula2>`, xmlText(validation.Formula2)))
	}
	builder.WriteString(`</dataValidation>`)
	return builder.String()
}

func xlsxColumnName(index int) string {
	name := ""
	for index >= 0 {
		name = string(rune('A'+(index%26))) + name
		index = index/26 - 1
	}
	return name
}

func xlsxRangeRef(sheetName string, column, startRow, endRow int) string {
	columnName := xlsxColumnName(column - 1)
	return fmt.Sprintf("%s!$%s$%d:$%s$%d", xlsxQuotedSheetName(sheetName), columnName, startRow, columnName, endRow)
}

func xlsxQuotedSheetName(sheetName string) string {
	return "'" + strings.ReplaceAll(sheetName, "'", "''") + "'"
}

func xmlText(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

func xmlAttr(value string) string {
	return strings.ReplaceAll(xmlText(value), `"`, "&quot;")
}

func textFromPg(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func uuidText(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return model.UUIDToString(value)
}

func safeFileToken(value string) string {
	token := strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, char := range token {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "blue-book"
	}
	return result
}
