package model

// AuditFieldChange menyimpan perubahan nilai satu field.
type AuditFieldChange struct {
	Field    string  `json:"field"`
	Label    string  `json:"label"`
	OldValue *string `json:"old_value"`
	NewValue *string `json:"new_value"`
}

type ProjectAuditEntry struct {
	ID                 string             `json:"id"`
	Section            string             `json:"section"`
	Action             string             `json:"action"`
	ActionLabel        string             `json:"action_label"`
	ChangedFields      []string           `json:"changed_fields"`
	ChangedFieldLabels []string           `json:"changed_field_labels"`
	FieldChanges       []AuditFieldChange `json:"field_changes"`
	ChangedByID        *string            `json:"changed_by_id,omitempty"`
	ChangedByUsername  string             `json:"changed_by_username"`
	ChangedAt          string             `json:"changed_at"`
	Summary            string             `json:"summary"`
}
