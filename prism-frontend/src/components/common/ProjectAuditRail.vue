<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { formatDateTime } from '@/utils/formatters'
import type { AuditFieldChange, ProjectAuditEntry } from '@/types/audit.types'

interface ProjectAuditRailItem extends ProjectAuditEntry {
  snapshot_label: string
}

defineProps<{
  items: ProjectAuditRailItem[]
}>()

const isAuditTrailOpen = ref(false)

function actionSeverity(action: string) {
  if (action === 'INSERT') return 'success'
  if (action === 'DELETE') return 'danger'
  return 'info'
}

function hasValue(v: string | null | undefined): boolean {
  return v != null && v !== ''
}

/** Alokasi Funding entries pakai tampilan tabel, bukan list */
function isFundingAllocation(item: ProjectAuditRailItem): boolean {
  return item.section.startsWith('Alokasi Funding')
}

/** Nilai untuk sel tabel: INSERT pakai new_value, DELETE pakai old_value, UPDATE keduanya */
function displayOld(fc: AuditFieldChange, action: string): string | null {
  if (action === 'INSERT') return null
  return hasValue(fc.old_value) ? fc.old_value : null
}

function displayNew(fc: AuditFieldChange, action: string): string | null {
  if (action === 'DELETE') return null
  return hasValue(fc.new_value) ? fc.new_value : null
}
</script>

<template>
  <aside v-if="items.length" class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <div class="flex items-center justify-between gap-3 px-5 py-4">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-lg font-semibold text-surface-950">Riwayat Perubahan</h2>
        <Tag :value="`${items.length} event`" severity="secondary" rounded />
      </div>
      <Button
        :label="isAuditTrailOpen ? 'Tutup' : 'Detail'"
        :icon="isAuditTrailOpen ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
        severity="secondary"
        size="small"
        outlined
        @click="isAuditTrailOpen = !isAuditTrailOpen"
      />
    </div>

    <ol v-if="isAuditTrailOpen" class="divide-y divide-surface-100">
      <li v-for="item in items" :key="item.id" class="flex gap-4 px-5 py-4">
        <!-- Timeline dot -->
        <div class="mt-1 flex shrink-0 flex-col items-center">
          <span
            class="h-2.5 w-2.5 rounded-full"
            :class="{
              'bg-emerald-500': item.action === 'INSERT',
              'bg-red-500': item.action === 'DELETE',
              'bg-primary': item.action === 'UPDATE',
            }"
          />
          <span class="mt-1 w-px grow bg-surface-100" />
        </div>

        <!-- Content -->
        <div class="min-w-0 flex-1 space-y-2 pb-1">
          <!-- Header: action tag + timestamp + user -->
          <div class="flex flex-wrap items-center gap-2">
            <Tag :value="item.action_label" :severity="actionSeverity(item.action)" rounded />
            <span class="text-xs font-medium text-surface-500">{{ formatDateTime(item.changed_at) }}</span>
            <span class="text-xs text-surface-400">·</span>
            <span class="text-xs font-medium text-surface-700">{{ item.changed_by_username }}</span>
            <span v-if="item.snapshot_label" class="text-xs text-surface-400">·</span>
            <span v-if="item.snapshot_label" class="text-xs text-surface-500">{{ item.snapshot_label }}</span>
          </div>

          <!-- Section label -->
          <p class="text-sm font-semibold text-surface-900">{{ item.section }}</p>

          <!-- ── Alokasi Funding: tampilkan sebagai tabel ringkas ── -->
          <template v-if="isFundingAllocation(item) && item.field_changes?.length">
            <div class="overflow-hidden rounded border border-surface-200">
              <table class="w-full border-collapse text-xs">
                <thead class="bg-surface-50 text-xs font-semibold uppercase tracking-wide text-surface-500">
                  <tr>
                    <th class="px-3 py-2 text-left font-semibold uppercase tracking-wide text-surface-500">
                      Komponen
                    </th>
                    <!-- INSERT: hanya kolom "Nilai" -->
                    <template v-if="item.action === 'INSERT'">
                      <th class="px-3 py-2 text-right font-semibold uppercase tracking-wide text-emerald-600">
                        Nilai (USD)
                      </th>
                    </template>
                    <!-- DELETE: hanya kolom "Dihapus" -->
                    <template v-else-if="item.action === 'DELETE'">
                      <th class="px-3 py-2 text-right font-semibold uppercase tracking-wide text-red-500">
                        Dihapus (USD)
                      </th>
                    </template>
                    <!-- UPDATE: dua kolom lama → baru -->
                    <template v-else>
                      <th class="px-3 py-2 text-right font-semibold uppercase tracking-wide text-surface-500">
                        Sebelum (USD)
                      </th>
                      <th class="px-3 py-2 text-right font-semibold uppercase tracking-wide text-primary">
                        Sesudah (USD)
                      </th>
                    </template>
                  </tr>
                </thead>
                <tbody class="divide-y divide-surface-100">
                  <tr
                    v-for="fc in item.field_changes"
                    :key="fc.field"
                    class="hover:bg-surface-50"
                  >
                    <td class="px-3 py-2 font-medium text-surface-800">{{ fc.label }}</td>
                    <!-- INSERT -->
                    <template v-if="item.action === 'INSERT'">
                      <td class="px-3 py-2 text-right font-mono text-emerald-700">
                        {{ displayNew(fc, item.action) ?? '0' }}
                      </td>
                    </template>
                    <!-- DELETE -->
                    <template v-else-if="item.action === 'DELETE'">
                      <td class="px-3 py-2 text-right font-mono text-red-600 line-through">
                        {{ displayOld(fc, item.action) ?? '0' }}
                      </td>
                    </template>
                    <!-- UPDATE -->
                    <template v-else>
                      <td class="px-3 py-2 text-right font-mono text-surface-500">
                        {{ displayOld(fc, item.action) ?? '0' }}
                      </td>
                      <td
                        class="px-3 py-2 text-right font-mono font-semibold"
                        :class="
                          displayNew(fc, item.action) !== displayOld(fc, item.action)
                            ? 'text-primary-700'
                            : 'text-surface-500'
                        "
                      >
                        {{ displayNew(fc, item.action) ?? '0' }}
                      </td>
                    </template>
                  </tr>
                </tbody>
              </table>
            </div>
          </template>

          <!-- ── Lainnya: daftar deskriptif field: oldVal → newVal ── -->
          <template v-else-if="item.field_changes?.length">
            <ul class="space-y-1">
              <li
                v-for="fc in item.field_changes"
                :key="fc.field"
                class="flex flex-wrap items-baseline gap-1 text-xs text-surface-700"
              >
                <span class="font-medium text-surface-900">{{ fc.label }}:</span>
                <!-- INSERT -->
                <template v-if="item.action === 'INSERT'">
                  <code
                    v-if="hasValue(fc.new_value)"
                    class="rounded bg-emerald-50 px-1.5 py-0.5 font-mono text-emerald-700"
                  >{{ fc.new_value }}</code>
                  <span v-else class="italic text-surface-400">(kosong)</span>
                </template>
                <!-- DELETE -->
                <template v-else-if="item.action === 'DELETE'">
                  <code
                    v-if="hasValue(fc.old_value)"
                    class="rounded bg-red-50 px-1.5 py-0.5 font-mono text-red-700 line-through"
                  >{{ fc.old_value }}</code>
                  <span v-else class="italic text-surface-400">(kosong)</span>
                </template>
                <!-- UPDATE -->
                <template v-else>
                  <code
                    v-if="hasValue(fc.old_value)"
                    class="rounded bg-surface-100 px-1.5 py-0.5 font-mono text-surface-600 line-through"
                  >{{ fc.old_value }}</code>
                  <span v-else class="italic text-surface-400">(kosong)</span>
                  <span class="text-surface-400">→</span>
                  <code
                    v-if="hasValue(fc.new_value)"
                    class="rounded bg-primary-50 px-1.5 py-0.5 font-mono text-primary-700"
                  >{{ fc.new_value }}</code>
                  <span v-else class="italic text-surface-400">(kosong)</span>
                </template>
              </li>
            </ul>
          </template>

          <!-- Fallback: chip list -->
          <div
            v-else-if="(item.changed_field_labels?.length ?? 0) > 0"
            class="flex flex-wrap gap-1.5"
          >
            <span
              v-for="field in item.changed_field_labels ?? []"
              :key="`${item.id}-${field}`"
              class="rounded-md bg-surface-100 px-2 py-1 text-xs font-medium text-surface-700"
            >
              {{ field }}
            </span>
          </div>

          <p v-else class="text-xs text-surface-400">Tidak ada detail field yang berubah.</p>
        </div>
      </li>
    </ol>
  </aside>
</template>
