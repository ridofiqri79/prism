package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

// uuidPattern mendeteksi UUID v4 yang valid.
var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// auditNameMap adalah peta UUID (lowercase) → nama tampilan.
type auditNameMap map[string]string

// buildAuditNameMap melakukan batch lookup UUID → nama dari semua tabel master.
func buildAuditNameMap(ctx context.Context, q *queries.Queries, rows []queries.ListBBProjectAuditEntriesRow) auditNameMap {
	return buildNameMapFromJSONRows(ctx, q, extractUUIDsFromBBRows(rows))
}

func buildAuditNameMapGB(ctx context.Context, q *queries.Queries, rows []queries.ListGBProjectAuditEntriesRow) auditNameMap {
	return buildNameMapFromJSONRows(ctx, q, extractUUIDsFromGBRows(rows))
}

func buildNameMapFromJSONRows(ctx context.Context, q *queries.Queries, uuids []pgtype.UUID) auditNameMap {
	if len(uuids) == 0 {
		return auditNameMap{}
	}
	resolved, err := q.ResolveAuditNames(ctx, uuids)
	if err != nil {
		return auditNameMap{}
	}
	m := make(auditNameMap, len(resolved))
	for _, r := range resolved {
		m[strings.ToLower(r.ID)] = r.DisplayName
	}
	return m
}

func extractUUIDsFromBBRows(rows []queries.ListBBProjectAuditEntriesRow) []pgtype.UUID {
	seen := map[string]struct{}{}
	var result []pgtype.UUID
	for _, row := range rows {
		extractUUIDsFromJSON(row.OldData, &result, seen)
		extractUUIDsFromJSON(row.NewData, &result, seen)
	}
	return result
}

func extractUUIDsFromGBRows(rows []queries.ListGBProjectAuditEntriesRow) []pgtype.UUID {
	seen := map[string]struct{}{}
	var result []pgtype.UUID
	for _, row := range rows {
		extractUUIDsFromJSON(row.OldData, &result, seen)
		extractUUIDsFromJSON(row.NewData, &result, seen)
	}
	return result
}

// extractUUIDsFromJSON mengambil semua nilai string yang berbentuk UUID dari JSONB.
func extractUUIDsFromJSON(data []byte, result *[]pgtype.UUID, seen map[string]struct{}) {
	if len(data) == 0 {
		return
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return
	}
	for _, raw := range m {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		lower := strings.ToLower(s)
		if !uuidPattern.MatchString(lower) {
			continue
		}
		if _, exists := seen[lower]; exists {
			continue
		}
		seen[lower] = struct{}{}
		parsed, err := model.ParseUUID(s)
		if err == nil {
			*result = append(*result, parsed)
		}
	}
}

// ── Public response builders ──────────────────────────────────────────────────

func bbProjectAuditResponses(ctx context.Context, q *queries.Queries, rows []queries.ListBBProjectAuditEntriesRow) []model.ProjectAuditEntry {
	nameMap := buildAuditNameMap(ctx, q, rows)
	entries := make([]model.ProjectAuditEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, projectAuditEntry(
			row.ID,
			row.TableName,
			row.Action,
			row.ChangedFields,
			row.OldData,
			row.NewData,
			nameMap,
			row.ChangedBy,
			row.ChangedByUsername,
			row.ChangedAt,
		))
	}
	return entries
}

func gbProjectAuditResponses(ctx context.Context, q *queries.Queries, rows []queries.ListGBProjectAuditEntriesRow) []model.ProjectAuditEntry {
	nameMap := buildAuditNameMapGB(ctx, q, rows)
	entries := make([]model.ProjectAuditEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, projectAuditEntry(
			row.ID,
			row.TableName,
			row.Action,
			row.ChangedFields,
			row.OldData,
			row.NewData,
			nameMap,
			row.ChangedBy,
			row.ChangedByUsername,
			row.ChangedAt,
		))
	}
	return entries
}

func projectAuditEntry(
	id pgtype.UUID,
	tableName, action string,
	changedFields []string,
	oldData, newData []byte,
	nameMap auditNameMap,
	changedBy pgtype.UUID,
	changedByUsername string,
	changedAt pgtype.Timestamptz,
) model.ProjectAuditEntry {
	section := auditSectionLabel(tableName)

	// Untuk gb_funding_allocation: tambahkan nama activity ke section label.
	if tableName == "gb_funding_allocation" {
		activityID := extractActivityIDFromJSON(oldData, newData)
		if activityID != "" {
			if name, ok := nameMap[strings.ToLower(activityID)]; ok {
				section = fmt.Sprintf("%s — %s", section, name)
			}
		}
	}

	actionLabel := auditActionLabel(action)
	fieldLabels := auditFieldLabels(tableName, changedFields)
	fieldChanges := buildFieldChanges(tableName, action, changedFields, oldData, newData, nameMap)

	summary := fmt.Sprintf("%s %s", actionLabel, section)
	if len(fieldLabels) > 0 {
		summary = fmt.Sprintf("%s: %s", summary, strings.Join(fieldLabels, ", "))
	}

	return model.ProjectAuditEntry{
		ID:                 model.UUIDToString(id),
		Section:            section,
		Action:             action,
		ActionLabel:        actionLabel,
		ChangedFields:      nonNilStrings(changedFields),
		ChangedFieldLabels: nonNilStrings(fieldLabels),
		FieldChanges:       fieldChanges,
		ChangedByID:        stringPtrFromUUID(changedBy),
		ChangedByUsername:  changedByUsername,
		ChangedAt:          formatMasterTime(changedAt),
		Summary:            summary,
	}
}

func extractActivityIDFromJSON(oldData, newData []byte) string {
	for _, data := range [][]byte{newData, oldData} {
		if len(data) == 0 {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}
		if raw, ok := m["gb_activity_id"]; ok {
			var s string
			if err := json.Unmarshal(raw, &s); err == nil && s != "" {
				return s
			}
		}
	}
	return ""
}

// buildFieldChanges membuat list perubahan per field dengan nilai lama dan baru.
// UUID di-resolve ke nama menggunakan nameMap.
func buildFieldChanges(tableName, action string, changedFields []string, oldData, newData []byte, nameMap auditNameMap) []model.AuditFieldChange {
	if len(changedFields) == 0 {
		return []model.AuditFieldChange{}
	}

	var oldMap, newMap map[string]json.RawMessage
	if len(oldData) > 0 {
		_ = json.Unmarshal(oldData, &oldMap)
	}
	if len(newData) > 0 {
		_ = json.Unmarshal(newData, &newMap)
	}

	seen := map[string]struct{}{}
	changes := make([]model.AuditFieldChange, 0, len(changedFields))

	for _, field := range changedFields {
		if shouldSkipAuditField(tableName, field) {
			continue
		}
		label := auditFieldLabel(field)
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}

		var oldVal, newVal *string

		if action != "INSERT" && oldMap != nil {
			if raw, ok := oldMap[field]; ok {
				s := resolveFieldValue(raw, nameMap)
				oldVal = &s
			}
		}
		if action != "DELETE" && newMap != nil {
			if raw, ok := newMap[field]; ok {
				s := resolveFieldValue(raw, nameMap)
				newVal = &s
			}
		}

		changes = append(changes, model.AuditFieldChange{
			Field:    field,
			Label:    label,
			OldValue: oldVal,
			NewValue: newVal,
		})
	}
	return changes
}

// resolveFieldValue mengubah JSON raw value menjadi string yang dapat dibaca manusia.
// - null → ""
// - string UUID → nama dari nameMap jika tersedia, fallback ke UUID
// - string biasa → tanpa quote
// - angka → diformat dengan pemisah ribuan (1,234,567.89)
// - bool/object/array → representasi JSON
func resolveFieldValue(raw json.RawMessage, nameMap auditNameMap) string {
	if raw == nil || string(raw) == "null" {
		return ""
	}

	// Coba sebagai string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		// UUID: resolve ke nama
		if uuidPattern.MatchString(strings.ToLower(s)) {
			if name, ok := nameMap[strings.ToLower(s)]; ok {
				return name
			}
		}
		return s
	}

	// Coba sebagai angka (float64) dan format
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return formatAuditNumber(f)
	}

	// Fallback: representasi JSON apa adanya
	return string(raw)
}

// formatAuditNumber memformat angka dengan pemisah ribuan.
// Contoh: 1234567.89 → "1,234,567.89", 1234 → "1,234"
func formatAuditNumber(f float64) string {
	if f == math.Trunc(f) {
		// Bilangan bulat
		n := int64(f)
		return formatInt(n)
	}
	// Float: format dengan 2 desimal
	formatted := strconv.FormatFloat(f, 'f', 2, 64)
	parts := strings.SplitN(formatted, ".", 2)
	return formatInt64String(parts[0]) + "." + parts[1]
}

func formatInt(n int64) string {
	return formatInt64String(strconv.FormatInt(n, 10))
}

func formatInt64String(s string) string {
	if len(s) == 0 {
		return s
	}
	// Tangani tanda negatif
	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}
	// Sisipkan koma setiap 3 digit dari kanan
	var result []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, c)
	}
	if negative {
		return "-" + string(result)
	}
	return string(result)
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func applyLatestAuditSummary(changedBy, changedAt, summary **string, entries []model.ProjectAuditEntry) {
	if len(entries) == 0 {
		return
	}
	latest := entries[0]
	latestChangedBy := latest.ChangedByUsername
	latestChangedAt := latest.ChangedAt
	latestSummary := latest.Summary
	*changedBy = &latestChangedBy
	*changedAt = &latestChangedAt
	*summary = &latestSummary
}

func auditSectionLabel(tableName string) string {
	labels := map[string]string{
		"bb_project":                   "Informasi proyek",
		"bb_project_institution":       "Executing/Implementing Agency",
		"bb_project_bappenas_partner":  "Mitra Kerja Bappenas",
		"bb_project_location":          "Lokasi",
		"bb_project_national_priority": "Prioritas Nasional",
		"bb_project_cost":              "Biaya proyek",
		"lender_indication":            "Indikasi lender",
		"loi":                          "Letter of Intent",
		"gb_project":                   "Informasi proyek",
		"gb_project_bb_project":        "Referensi Blue Book",
		"gb_project_bappenas_partner":  "Mitra Kerja Bappenas",
		"gb_project_institution":       "Executing/Implementing Agency",
		"gb_project_location":          "Lokasi",
		"gb_activity":                  "Activities",
		"gb_funding_source":            "Funding Source",
		"gb_disbursement_plan":         "Rencana Disbursement",
		"gb_funding_allocation":        "Alokasi Funding",
	}
	if label, ok := labels[tableName]; ok {
		return label
	}
	return tableName
}

func auditActionLabel(action string) string {
	switch action {
	case "INSERT":
		return "Menambah"
	case "UPDATE":
		return "Mengubah"
	case "DELETE":
		return "Menghapus"
	default:
		return action
	}
}

func auditFieldLabels(tableName string, fields []string) []string {
	if len(fields) == 0 {
		return nil
	}
	labels := make([]string, 0, len(fields))
	seen := map[string]struct{}{}
	for _, field := range fields {
		if shouldSkipAuditField(tableName, field) {
			continue
		}
		label := auditFieldLabel(field)
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}
		labels = append(labels, label)
	}
	return labels
}

func shouldSkipAuditField(tableName, field string) bool {
	switch tableName {
	case "bb_project_institution", "bb_project_bappenas_partner", "bb_project_location", "bb_project_national_priority":
		return field == "bb_project_id"
	case "bb_project_cost", "lender_indication", "loi":
		return field == "bb_project_id"
	case "gb_activity", "gb_funding_source", "gb_disbursement_plan":
		return field == "gb_project_id"
	case "gb_funding_allocation":
		return field == "gb_activity_id"
	default:
		return false
	}
}

func auditFieldLabel(field string) string {
	labels := map[string]string{
		"blue_book_id":            "Dokumen Blue Book",
		"project_identity_id":     "Identitas logical proyek",
		"program_title_id":        "Judul program",
		"bb_code":                 "Kode Blue Book",
		"project_name":            "Nama proyek",
		"duration":                "Durasi",
		"objective":               "Objective",
		"objectives":              "Objectives",
		"scope_of_work":           "Scope of Work",
		"scope_of_project":        "Scope of Project",
		"outputs":                 "Outputs",
		"outcomes":                "Outcomes",
		"status":                  "Status",
		"bb_project_id":           "Proyek Blue Book",
		"gb_project_id":           "Proyek Green Book",
		"gb_project_identity_id":  "Identitas logical Green Book",
		"green_book_id":           "Dokumen Green Book",
		"gb_code":                 "Kode Green Book",
		"institution_id":          "Institution",
		"role":                    "Peran institution",
		"bappenas_partner_id":     "Mitra Kerja Bappenas",
		"region_id":               "Wilayah",
		"national_priority_id":    "Prioritas Nasional",
		"funding_type":            "Jenis pendanaan",
		"funding_category":        "Kategori pendanaan",
		"amount_usd":              "Nilai USD",
		"lender_id":               "Lender",
		"remarks":                 "Catatan",
		"subject":                 "Subject",
		"date":                    "Tanggal",
		"letter_number":           "Letter Number",
		"activity_name":           "Nama activity",
		"implementation_location": "Lokasi implementasi",
		"piu":                     "PIU",
		"sort_order":              "Urutan",
		"currency":                "Currency",
		"loan_original":           "Pinjaman original",
		"grant_original":          "Hibah original",
		"local_original":          "Dana lokal original",
		"loan_usd":                "Pinjaman USD",
		"grant_usd":               "Hibah USD",
		"local_usd":               "Dana lokal USD",
		"year":                    "Tahun",
		"gb_activity_id":          "Activity Green Book",
		"services":                "Services",
		"constructions":           "Constructions",
		"goods":                   "Goods",
		"trainings":               "Trainings",
		"other":                   "Other",
	}
	if label, ok := labels[field]; ok {
		return label
	}
	return strings.ReplaceAll(field, "_", " ")
}
