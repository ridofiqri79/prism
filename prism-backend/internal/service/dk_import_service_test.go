package service

import "testing"

func TestParseDKImportDate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "text date", value: "2026-04-28", want: "2026-04-28"},
		{name: "excel serial date", value: "46140", want: "2026-04-28"},
		{name: "excel serial date with time fraction", value: "46140.75", want: "2026-04-28"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDKImportDate(tt.value)
			if err != nil {
				t.Fatalf("parseDKImportDate(%q) error = %v", tt.value, err)
			}
			if !got.Valid {
				t.Fatalf("parseDKImportDate(%q) returned invalid date", tt.value)
			}
			if got.Time.Format("2006-01-02") != tt.want {
				t.Fatalf("parseDKImportDate(%q) = %s, want %s", tt.value, got.Time.Format("2006-01-02"), tt.want)
			}
		})
	}
}

func TestParseDKImportDateRejectsInvalidExcelSerial(t *testing.T) {
	if _, err := parseDKImportDate("60"); err == nil {
		t.Fatal("parseDKImportDate(\"60\") expected error")
	}
}
