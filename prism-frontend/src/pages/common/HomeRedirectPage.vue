<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import { useAuthStore } from '@/stores/auth.store'
import { resolveDefaultAuthenticatedRoute } from '@/utils/default-route'

const router = useRouter()
const auth = useAuthStore()

onMounted(() => {
  void router.replace(
    resolveDefaultAuthenticatedRoute({
      user: auth.user,
      permissions: auth.permissions,
    }),
  )
})
</script>

<template>
  <section class="flex min-h-[40vh] items-center justify-center text-sm text-surface-500">
    Mengalihkan...
  </section>
</template>
