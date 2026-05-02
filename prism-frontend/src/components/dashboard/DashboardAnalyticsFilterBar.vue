<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import type {
  DashboardAnalyticsFilterState,
  DashboardAnalyticsPipelineStage,
  DashboardAnalyticsProjectStatus,
  DashboardQuarter,
} from '@/types/dashboard.types'
import type { Institution, Lender, LenderType, ProgramTitle, Region } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: DashboardAnalyticsFilterState
    lenders: Lender[]
    institutions: Institution[]
    regions: Region[]
    programTitles: ProgramTitle[]
    loading?: boolean
    loadingOptions?: boolean
  }>(),
  {
    loading: false,
    loadingOptions: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [DashboardAnalyticsFilterState]
  apply: []
  reset: []
}>()

const showAdvanced = ref(false)
const currentYear = new Date().getFullYear()
const yearOptions = Array.from({ length: 11 }, (_, index) => {
  const year = currentYear - 5 + index
  return { label: String(year), value: year }
})
const quarterOptions: Array<{ label: string; value: DashboardQuarter | null }> = [
  { label: 'Semua Triwulan', value: null },
  { label: 'TW1 (Apr-Jun)', value: 'TW1' },
  { label: 'TW2 (Jul-Sep)', value: 'TW2' },
  { label: 'TW3 (Okt-Des)', value: 'TW3' },
  { label: 'TW4 (Jan-Mar)', value: 'TW4' },
]
const lenderTypeOptions: Array<{ label: string; value: LenderType }> = [
  { label: 'Bilateral', value: 'Bilateral' },
  { label: 'Multilateral', value: 'Multilateral' },
  { label: 'KSA', value: 'KSA' },
]
const pipelineStatusOptions: Array<{ label: string; value: DashboardAnalyticsPipelineStage }> = [
  { label: 'Blue Book', value: 'BB' },
  { label: 'Green Book', value: 'GB' },
  { label: 'Daftar Kegiatan', value: 'DK' },
  { label: 'Loan Agreement', value: 'LA' },
  { label: 'Monitoring', value: 'Monitoring' },
]
const projectStatusOptions: Array<{ label: string; value: DashboardAnalyticsProjectStatus }> = [
  { label: 'Pipeline', value: 'Pipeline' },
  { label: 'Ongoing', value: 'Ongoing' },
]

const localFilters = reactive<DashboardAnalyticsFilterState>(cloneFilters(props.modelValue))

const lenderOptions = computed(() =>
  props.lenders.map((lender) => ({
    ...lender,
    label: lender.short_name ? `${lender.name} (${lender.short_name})` : lender.name,
    value: lender.id,
  })),
)
const institutionOptions = computed(() =>
  props.institutions.map((institution) => ({
    ...institution,
    label: institution.short_name
      ? `${institution.name} (${institution.short_name})`
      : institution.name,
    value: institution.id,
  })),
)
const programTitleOptions = computed(() =>
  props.programTitles.map((programTitle) => ({
    ...programTitle,
    label: formatProgramTitle(programTitle),
    value: programTitle.id,
  })),
)
const selectedCountryCodes = computed(() => {
  const selected = new Set(localFilters.region_ids)

  return props.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})
const regionOptions = computed(() =>
  props.regions.map((region) => ({
    ...region,
    label: formatRegion(region),
    value: region.id,
    disabled:
      selectedCountryCodes.value.length > 0 &&
      region.type !== 'COUNTRY' &&
      isCoveredBySelectedCountry(region),
  })),
)

function cloneFilters(filters: DashboardAnalyticsFilterState): DashboardAnalyticsFilterState {
  return {
    budget_year: filters.budget_year,
    quarter: filters.quarter,
    lender_ids: [...filters.lender_ids],
    lender_types: [...filters.lender_types],
    institution_ids: [...filters.institution_ids],
    pipeline_statuses: [...filters.pipeline_statuses],
    project_statuses: [...filters.project_statuses],
    region_ids: [...filters.region_ids],
    program_title_ids: [...filters.program_title_ids],
    foreign_loan_min: filters.foreign_loan_min,
    foreign_loan_max: filters.foreign_loan_max,
    include_history: filters.include_history,
  }
}

function assignFilters(
  target: DashboardAnalyticsFilterState,
  source: DashboardAnalyticsFilterState,
) {
  target.budget_year = source.budget_year
  target.quarter = source.quarter
  target.lender_ids = [...source.lender_ids]
  target.lender_types = [...source.lender_types]
  target.institution_ids = [...source.institution_ids]
  target.pipeline_statuses = [...source.pipeline_statuses]
  target.project_statuses = [...source.project_statuses]
  target.region_ids = [...source.region_ids]
  target.program_title_ids = [...source.program_title_ids]
  target.foreign_loan_min = source.foreign_loan_min
  target.foreign_loan_max = source.foreign_loan_max
  target.include_history = source.include_history
}

function formatProgramTitle(programTitle: ProgramTitle) {
  const parent = props.programTitles.find((item) => item.id === programTitle.parent_id)

  return parent ? `${parent.title} / ${programTitle.title}` : programTitle.title
}

function formatRegion(region: Region) {
  const levelLabel: Record<Region['type'], string> = {
    COUNTRY: 'Region',
    PROVINCE: 'Provinsi',
    CITY: 'Kab/Kota',
  }

  if (region.type === 'COUNTRY') return `${region.name} (${levelLabel[region.type]})`
  if (region.type === 'CITY') return `-- ${region.name} (${levelLabel[region.type]})`

  return `- ${region.name} (${levelLabel[region.type]})`
}

function isCoveredBySelectedCountry(region: Region) {
  if (!region.parent_code) return false
  if (selectedCountryCodes.value.includes(region.parent_code)) return true

  const parent = props.regions.find((item) => item.code === region.parent_code)

  return parent?.parent_code ? selectedCountryCodes.value.includes(parent.parent_code) : false
}

watch(
  () => props.modelValue,
  (value) => assignFilters(localFilters, value),
  { deep: true },
)

watch(localFilters, () => emit('update:modelValue', cloneFilters(localFilters)), { deep: true })
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white">
    <div class="grid gap-4 p-4 lg:grid-cols-[12rem_14rem_1fr_auto]">
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Tahun Anggaran</span>
        <Select
          v-model="localFilters.budget_year"
          :options="yearOptions"
          option-label="label"
          option-value="value"
          show-clear
          class="w-full"
        />
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Triwulan</span>
        <Select
          v-model="localFilters.quarter"
          :options="quarterOptions"
          option-label="label"
          option-value="value"
          class="w-full"
        />
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Lender</span>
        <MultiSelect
          v-model="localFilters.lender_ids"
          :options="lenderOptions"
          option-label="label"
          option-value="value"
          placeholder="Semua lender"
          display="chip"
          filter
          filter-placeholder="Cari lender"
          :loading="loadingOptions"
          class="w-full"
        >
          <template #option="{ option }">
            <div class="flex w-full items-center justify-between gap-3">
              <span>{{ option.label }}</span>
              <Tag :value="option.type" severity="info" rounded />
            </div>
          </template>
        </MultiSelect>
      </label>

      <div class="flex flex-wrap items-end gap-2">
        <Button label="Terapkan" icon="pi pi-filter" :loading="loading" @click="emit('apply')" />
        <Button
          label="Reset"
          icon="pi pi-refresh"
          severity="secondary"
          outlined
          @click="emit('reset')"
        />
        <Button
          :label="showAdvanced ? 'Tutup' : 'Filter Lanjutan'"
          :icon="showAdvanced ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
          severity="secondary"
          text
          @click="showAdvanced = !showAdvanced"
        />
      </div>
    </div>

    <div
      v-if="showAdvanced"
      class="grid gap-4 border-t border-surface-200 p-4 md:grid-cols-2 xl:grid-cols-6"
    >
      <label class="block space-y-2 xl:col-span-2">
        <span class="text-sm font-medium text-surface-700">Tipe Lender</span>
        <MultiSelect
          v-model="localFilters.lender_types"
          :options="lenderTypeOptions"
          option-label="label"
          option-value="value"
          placeholder="Semua tipe"
          display="chip"
          class="w-full"
        />
      </label>

      <label class="block space-y-2 xl:col-span-2">
        <span class="text-sm font-medium text-surface-700">Kementerian/Lembaga</span>
        <MultiSelect
          v-model="localFilters.institution_ids"
          :options="institutionOptions"
          option-label="label"
          option-value="value"
          placeholder="Semua Kementerian/Lembaga"
          display="chip"
          filter
          filter-placeholder="Cari Kementerian/Lembaga"
          :loading="loadingOptions"
          class="w-full"
        />
      </label>

      <label class="block space-y-2 xl:col-span-2">
        <span class="text-sm font-medium text-surface-700">Status Pipeline</span>
        <MultiSelect
          v-model="localFilters.pipeline_statuses"
          :options="pipelineStatusOptions"
          option-label="label"
          option-value="value"
          placeholder="Semua stage"
          display="chip"
          class="w-full"
        />
      </label>

      <label class="block space-y-2 xl:col-span-2">
        <span class="text-sm font-medium text-surface-700">Status Project</span>
        <MultiSelect
          v-model="localFilters.project_statuses"
          :options="projectStatusOptions"
          option-label="label"
          option-value="value"
          placeholder="Semua status"
          display="chip"
          class="w-full"
        />
      </label>

      <label class="block space-y-2 xl:col-span-2">
        <span class="text-sm font-medium text-surface-700">Wilayah</span>
        <MultiSelect
          v-model="localFilters.region_ids"
          :options="regionOptions"
          option-label="label"
          option-value="value"
          option-disabled="disabled"
          placeholder="Semua wilayah"
          display="chip"
          filter
          filter-placeholder="Cari wilayah"
          :loading="loadingOptions"
          class="w-full"
        />
      </label>

      <label class="block space-y-2 xl:col-span-2">
        <span class="text-sm font-medium text-surface-700">Program Title</span>
        <MultiSelect
          v-model="localFilters.program_title_ids"
          :options="programTitleOptions"
          option-label="label"
          option-value="value"
          placeholder="Semua program title"
          display="chip"
          filter
          filter-placeholder="Cari program title"
          :loading="loadingOptions"
          class="w-full"
        />
      </label>

      <div class="grid gap-4 md:grid-cols-2 xl:col-span-3">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nilai Pinjaman USD Min</span>
          <InputNumber
            v-model="localFilters.foreign_loan_min"
            :min="0"
            :min-fraction-digits="0"
            :max-fraction-digits="2"
            placeholder="Minimum"
            class="w-full"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nilai Pinjaman USD Max</span>
          <InputNumber
            v-model="localFilters.foreign_loan_max"
            :min="0"
            :min-fraction-digits="0"
            :max-fraction-digits="2"
            placeholder="Maksimum"
            class="w-full"
          />
        </label>
      </div>

      <label
        class="flex items-center gap-3 rounded-lg border border-surface-200 px-3 py-2 xl:col-span-3"
      >
        <ToggleSwitch v-model="localFilters.include_history" />
        <span class="text-sm font-medium text-surface-700"
          >Tampilkan histori revisi untuk mode audit/history</span
        >
      </label>
    </div>
  </section>
</template>
