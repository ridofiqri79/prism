import { useConfirm as usePrimeConfirm } from 'primevue/useconfirm'

export function useConfirm() {
  const confirm = usePrimeConfirm()

  function confirmDelete(label: string, onAccept: () => void) {
    confirm.require({
      header: 'Konfirmasi Hapus',
      message: `Hapus ${label}? Data yang dihapus tidak dapat dikembalikan.`,
      icon: 'pi pi-exclamation-triangle',
      rejectLabel: 'Batal',
      acceptLabel: 'Hapus',
      acceptClass: 'p-button-danger',
      accept: onAccept,
    })
  }

  return { confirmDelete }
}
