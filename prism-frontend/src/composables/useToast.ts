import { useToast as usePrimeToast } from 'primevue/usetoast'

export function useToast() {
  const toast = usePrimeToast()

  function success(summary: string, detail?: string, life = 3000) {
    toast.add({ severity: 'success', summary, detail, life })
  }

  function error(summary: string, detail?: string, life = 5000) {
    toast.add({ severity: 'error', summary, detail, life })
  }

  function warn(summary: string, detail?: string, life = 5000) {
    toast.add({ severity: 'warn', summary, detail, life })
  }

  function info(summary: string, detail?: string, life = 3000) {
    toast.add({ severity: 'info', summary, detail, life })
  }

  return {
    success,
    error,
    warn,
    info,
  }
}
