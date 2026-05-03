package service

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

func bbProjectAuditResponses(rows []queries.ListBBProjectAuditEntriesRow) []model.ProjectAuditEntry {
	entries := make([]model.ProjectAuditEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, projectAuditEntry(
			row.ID,
			row.TableName,
			row.Action,
			row.ChangedFields,
			row.ChangedBy,
			row.ChangedByUsername,
			row.ChangedAt,
		))
	}
	return entries
}

func gbProjectAuditResponses(rows []queries.ListGBProjectAuditEntriesRow) []model.ProjectAuditEntry {
	entries := make([]model.ProjectAuditEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, projectAuditEntry(
			row.ID,
			row.TableName,
			row.Action,
			row.ChangedFields,
			row.ChangedBy,
			row.ChangedByUsername,
			row.ChangedAt,
		))
	}
	return entries
}

func projectAuditEntry(id pgtype.UUID, tableName, action string, changedFields []string, changedBy pgtype.UUID, changedByUsername string, changedAt pgtype.Timestamptz) model.ProjectAuditEntry {
	section := auditSectionLabel(tableName)
	actionLabel := auditActionLabel(action)
	fieldLabels := auditFieldLabels(tableName, changedFields)
	summary := fmt.Sprintf("%s %s", actionLabel, section)
	if len(fieldLabels) > 0 {
		summary = fmt.Sprintf("%s: %s", summary, strings.Join(fieldLabels, ", "))
	}

	return model.ProjectAuditEntry{
		ID:                 model.UUIDToString(id),
		Section:            section,
		Action:             action,
		ActionLabel:        actionLabel,
		ChangedFields:      changedFields,
		ChangedFieldLabels: fieldLabels,
		ChangedByID:        stringPtrFromUUID(changedBy),
		ChangedByUsername:  changedByUsername,
		ChangedAt:          formatMasterTime(changedAt),
		Summary:            summary,
	}
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
