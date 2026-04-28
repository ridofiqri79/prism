package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

func TestBuildSimpleXLSXTemplateMetadata(t *testing.T) {
	workbook := simpleXLSXWorkbook{
		Sheets: []simpleXLSXSheet{
			templateInputSheet("Input Data", []string{"Name (*)", "Type (*)"}, []float64{28, 20}, []simpleXLSXValidation{
				listValidation("B2:B"+inputLastRow(), "ddTypes", "Type", "Pilih type dari dropdown."),
			}),
			{
				Name:          "_Dropdowns",
				Rows:          [][]simpleXLSXCell{headerRow("Types"), textRow("A"), textRow("B")},
				Columns:       columns(24),
				Hidden:        true,
				ShowGridLines: false,
			},
		},
		DefinedNames: []simpleXLSXDefinedName{
			{Name: "ddTypes", Ref: xlsxRangeRef("_Dropdowns", 1, 2, 3)},
		},
	}

	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		t.Fatalf("buildSimpleXLSX() error = %v", err)
	}
	assertAllXMLPartsParse(t, data)

	workbookXML := readXLSXPart(t, data, "xl/workbook.xml")
	if !strings.Contains(workbookXML, `state="hidden"`) {
		t.Fatalf("workbook.xml does not mark dropdown sheet hidden:\n%s", workbookXML)
	}
	if !strings.Contains(workbookXML, `<definedName name="ddTypes">&#39;_Dropdowns&#39;!$A$2:$A$3</definedName>`) {
		t.Fatalf("workbook.xml missing ddTypes defined name:\n%s", workbookXML)
	}

	sheetXML := readXLSXPart(t, data, "xl/worksheets/sheet1.xml")
	for _, expected := range []string{`<dataValidations count="1">`, `errorStyle="warning"`, `<formula1>ddTypes</formula1>`, `sqref="B2:B5001"`, `<cols>`} {
		if !strings.Contains(sheetXML, expected) {
			t.Fatalf("sheet1.xml missing %q:\n%s", expected, sheetXML)
		}
	}

	parsed, err := readXLSXWorkbook(data)
	if err != nil {
		t.Fatalf("generated workbook cannot be parsed by importer: %v", err)
	}
	rows, ok := parsed.importRows("Input Data", []string{"name", "type"})
	if !ok {
		t.Fatal("Input Data sheet was not found")
	}
	if len(rows) != 0 {
		t.Fatalf("expected empty template data rows, got %d", len(rows))
	}
}

func assertAllXMLPartsParse(t *testing.T, data []byte) {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid xlsx zip: %v", err)
	}
	for _, file := range reader.File {
		if !strings.HasSuffix(file.Name, ".xml") {
			continue
		}
		content := readZipTestFile(t, file)
		decoder := xml.NewDecoder(bytes.NewReader(content))
		for {
			if _, err := decoder.Token(); err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("%s is not valid XML: %v", file.Name, err)
			}
		}
	}
}

func readXLSXPart(t *testing.T, data []byte, name string) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid xlsx zip: %v", err)
	}
	for _, file := range reader.File {
		if file.Name == name {
			return string(readZipTestFile(t, file))
		}
	}
	t.Fatalf("xlsx part %s not found", name)
	return ""
}

func readZipTestFile(t *testing.T, file *zip.File) []byte {
	t.Helper()
	reader, err := file.Open()
	if err != nil {
		t.Fatalf("open zip file %s: %v", file.Name, err)
	}
	defer reader.Close()
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read zip file %s: %v", file.Name, err)
	}
	return content
}
