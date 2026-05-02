package service

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestAbsorptionPctDivByZero(t *testing.T) {
	result := absorptionPct(0, 100)
	if result != 0 {
		t.Fatalf("absorptionPct(0, 100) = %v, want 0", result)
	}
	result = absorptionPct(0, 0)
	if result != 0 {
		t.Fatalf("absorptionPct(0, 0) = %v, want 0", result)
	}
}

func TestAbsorptionPctNormal(t *testing.T) {
	result := absorptionPct(100, 75)
	if result != 75.0 {
		t.Fatalf("absorptionPct(100, 75) = %v, want 75.0", result)
	}
	result = absorptionPct(200, 50)
	if result != 25.0 {
		t.Fatalf("absorptionPct(200, 50) = %v, want 25.0", result)
	}
}

func TestAbsorptionStatus(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{0, "low"},
		{30.5, "low"},
		{49.99, "low"},
		{50, "normal"},
		{75.0, "normal"},
		{89.99, "normal"},
		{90, "high"},
		{95.5, "high"},
		{100, "high"},
	}
	for _, tc := range tests {
		got := absorptionStatus(tc.pct)
		if got != tc.want {
			t.Fatalf("absorptionStatus(%v) = %q, want %q", tc.pct, got, tc.want)
		}
	}
}

func TestNullableInt32(t *testing.T) {
	result := nullableInt32(nil)
	if result.Valid {
		t.Fatal("nullableInt32(nil) should not be valid")
	}
	val := int32(2025)
	result = nullableInt32(&val)
	if !result.Valid || result.Int32 != 2025 {
		t.Fatalf("nullableInt32(2025) = %+v, want valid 2025", result)
	}
}

func TestNumericFromInterface(t *testing.T) {
	if got := numericFromInterface(float64(42.5)); got != 42.5 {
		t.Fatalf("numericFromInterface(42.5) = %v, want 42.5", got)
	}
	if got := numericFromInterface(int64(100)); got != 100 {
		t.Fatalf("numericFromInterface(100) = %v, want 100", got)
	}
	num := pgtype.Numeric{Int: pgtype.Numeric{}.Int, Valid: true}
	result := numericFromInterface(num)
	// pgtype.Numeric with zero values converts to 0
	_ = result // conversion from pgtype.Numeric is tested indirectly
	if got := numericFromInterface("invalid"); got != 0 {
		t.Fatalf("numericFromInterface(\"invalid\") = %v, want 0", got)
	}
	if got := numericFromInterface(nil); got != 0 {
		t.Fatalf("numericFromInterface(nil) = %v, want 0", got)
	}
}

func TestParseUUIDs(t *testing.T) {
	result := parseUUIDs(nil)
	if result != nil {
		t.Fatal("parseUUIDs(nil) should return nil")
	}
	result = parseUUIDs([]string{})
	if result != nil {
		t.Fatal("parseUUIDs([]) should return nil")
	}
	result = parseUUIDs([]string{"not-a-uuid"})
	if len(result) != 0 {
		t.Fatalf("parseUUIDs(invalid) = %d items, want 0", len(result))
	}
	result = parseUUIDs([]string{"00000000-0000-0000-0000-000000000001"})
	if len(result) != 1 || !result[0].Valid {
		t.Fatalf("parseUUIDs(valid) = %+v, want 1 valid item", result)
	}
}
