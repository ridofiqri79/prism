<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'

const router = useRouter()
</script>

<template>
  <section class="mx-auto flex max-w-xl flex-col items-center justify-center py-20 text-center">
    <i class="pi pi-lock text-5xl text-primary" />
    <h1 class="mt-6 text-3xl font-semibold text-surface-950">403 - Anda tidak memiliki izin mengakses halaman ini</h1>
    <Button class="mt-6" label="Kembali ke Dashboard" icon="pi pi-arrow-left" @click="router.push({ name: 'dashboard' })" />
  </section>
</template>
