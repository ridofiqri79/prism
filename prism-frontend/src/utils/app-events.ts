export const TOAST_EVENT_NAME = 'prism:toast'
export const UNAUTHORIZED_EVENT_NAME = 'prism:unauthorized'
export const LOGIN_REDIRECT_EVENT_NAME = 'prism:login-redirect'

export interface AppToastMessage {
  severity: 'success' | 'info' | 'warn' | 'error'
  summary: string
  detail?: string
  life?: number
}

export function emitToast(message: AppToastMessage) {
  window.dispatchEvent(new CustomEvent<AppToastMessage>(TOAST_EVENT_NAME, { detail: message }))
}

export function emitUnauthorized() {
  window.dispatchEvent(new Event(UNAUTHORIZED_EVENT_NAME))
}

export function emitLoginRedirect() {
  window.dispatchEvent(new Event(LOGIN_REDIRECT_EVENT_NAME))
}
