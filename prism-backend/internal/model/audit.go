package model

type ProjectAuditEntry struct {
	ID                 string   `json:"id"`
	Section            string   `json:"section"`
	Action             string   `json:"action"`
	ActionLabel        string   `json:"action_label"`
	ChangedFields      []string `json:"changed_fields"`
	ChangedFieldLabels []string `json:"changed_field_labels"`
	ChangedByID        *string  `json:"changed_by_id,omitempty"`
	ChangedByUsername  string   `json:"changed_by_username"`
	ChangedAt          string   `json:"changed_at"`
	Summary            string   `json:"summary"`
}
