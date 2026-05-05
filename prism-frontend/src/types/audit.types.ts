export interface AuditFieldChange {
  field: string
  label: string
  old_value: string | null
  new_value: string | null
}

export interface ProjectAuditEntry {
  id: string
  section: string
  action: 'INSERT' | 'UPDATE' | 'DELETE' | string
  action_label: string
  changed_fields: string[]
  changed_field_labels: string[]
  field_changes: AuditFieldChange[]
  changed_by_id?: string
  changed_by_username: string
  changed_at: string
  summary: string
}
