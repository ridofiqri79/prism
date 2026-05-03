export interface ProjectAuditEntry {
  id: string
  section: string
  action: 'INSERT' | 'UPDATE' | 'DELETE' | string
  action_label: string
  changed_fields: string[]
  changed_field_labels: string[]
  changed_by_id?: string
  changed_by_username: string
  changed_at: string
  summary: string
}
