import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import Tooltip from 'primevue/tooltip'
import App from './App.vue'
import { prismPreset } from '@/assets/styles/theme'
import router from '@/router'
import { useAuthStore } from '@/stores/auth.store'
import { UNAUTHORIZED_EVENT_NAME } from '@/utils/app-events'
import '@/assets/styles/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: prismPreset,
    options: {
      darkModeSelector: '.dark',
      cssLayer: {
        name: 'primevue',
        order: 'theme, base, primevue',
      },
    },
  },
})
app.use(ToastService)
app.use(ConfirmationService)
app.directive('tooltip', Tooltip)

const auth = useAuthStore(pinia)

window.addEventListener(UNAUTHORIZED_EVENT_NAME, () => {
  auth.clearSession()
})

app.mount('#app')
