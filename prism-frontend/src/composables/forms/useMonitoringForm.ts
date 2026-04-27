import { computed, reactive, ref } from 'vue'
import type { ZodError } from 'zod'
import { monitoringSchema } from '@/schemas/monitoring.schema'
import type {
  MonitoringDisbursement,
  MonitoringKomponen,
  MonitoringPayload,
  Quarter,
} from '@/types/monitoring.types'

export type MonitoringFormErrors = Partial<Record<keyof MonitoringPayload, string>>

function defaultValues(): MonitoringPayload {
  return {
    budget_year: new Date().getFullYear(),
    quarter: 'TW1',
    exchange_rate_usd_idr: 0,
    exchange_rate_la_idr: 0,
    planned_la: 0,
    planned_usd: 0,
    planned_idr: 0,
    realized_la: 0,
    realized_usd: 0,
    realized_idr: 0,
    komponen: [],
  }
}

function emptyKomponen(): MonitoringKomponen {
  return {
    component_name: '',
    planned_la: 0,
    planned_usd: 0,
    planned_idr: 0,
    realized_la: 0,
    realized_usd: 0,
    realized_idr: 0,
  }
}

function assignErrors(target: MonitoringFormErrors, error: ZodError) {
  Object.keys(target).forEach((key) => {
    delete target[key as keyof MonitoringFormErrors]
  })

  for (const issue of error.issues) {
    const field = String(issue.path[0]) as keyof MonitoringPayload
    if (!target[field]) {
      target[field] = issue.message
    }
  }
}

function toPayload(data: MonitoringDisbursement): MonitoringPayload {
  return {
    budget_year: data.budget_year,
    quarter: data.quarter,
    exchange_rate_usd_idr: data.exchange_rate_usd_idr,
    exchange_rate_la_idr: data.exchange_rate_la_idr,
    planned_la: data.planned_la,
    planned_usd: data.planned_usd,
    planned_idr: data.planned_idr,
    realized_la: data.realized_la,
    realized_usd: data.realized_usd,
    realized_idr: data.realized_idr,
    komponen: data.komponen ?? [],
  }
}

export function useMonitoringForm(initialData?: MonitoringDisbursement | null) {
  const values = reactive<MonitoringPayload>({
    ...defaultValues(),
    ...(initialData ? toPayload(initialData) : {}),
  })
  const errors = reactive<MonitoringFormErrors>({})
  const showKomponen = ref(Boolean(initialData?.komponen?.length))
  const komponen = ref<MonitoringKomponen[]>(initialData?.komponen ? [...initialData.komponen] : [])

  const absorptionPct = computed(() => {
    const planned = values.planned_usd ?? 0
    const realized = values.realized_usd ?? 0

    if (planned === 0) return 0

    return Math.round((realized / planned) * 1000) / 10
  })

  function addKomponen() {
    showKomponen.value = true
    komponen.value.push(emptyKomponen())
  }

  function removeKomponen(index: number) {
    komponen.value.splice(index, 1)
  }

  function applyMonitoring(data: MonitoringDisbursement) {
    Object.assign(values, toPayload(data))
    komponen.value = data.komponen ? [...data.komponen] : []
    showKomponen.value = komponen.value.length > 0
  }

  function setQuarter(value: Quarter) {
    values.quarter = value
  }

  function submit(callback: (payload: MonitoringPayload) => unknown | Promise<unknown>) {
    return async () => {
      const payload: MonitoringPayload = {
        ...values,
        komponen: showKomponen.value ? komponen.value : [],
      }
      const parsed = monitoringSchema.safeParse(payload)

      if (!parsed.success) {
        assignErrors(errors, parsed.error)
        return
      }

      Object.keys(errors).forEach((key) => {
        delete errors[key as keyof MonitoringFormErrors]
      })
      await callback(parsed.data)
    }
  }

  return {
    values,
    errors,
    showKomponen,
    komponen,
    absorptionPct,
    addKomponen,
    removeKomponen,
    applyMonitoring,
    setQuarter,
    submit,
  }
}
