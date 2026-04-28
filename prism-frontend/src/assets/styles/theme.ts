import { definePreset } from '@primevue/themes'
import Aura from '@primevue/themes/aura'

export const prismPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#EAFBFA',
      100: '#D8F5F4',
      200: '#B7ECE9',
      300: '#78DCD8',
      400: '#26B7A5',
      500: '#1FB5B2',
      600: '#17A2A4',
      700: '#0F8F8C',
      800: '#0B6F73',
      900: '#075A5F',
      950: '#043F43',
    },
  },
})
