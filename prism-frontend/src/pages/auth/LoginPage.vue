<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { isAxiosError } from 'axios'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Password from 'primevue/password'
import { loginSchema, type LoginFormValues } from '@/schemas/auth.schema'
import { useAuthStore } from '@/stores/auth.store'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const globeCanvas = ref<HTMLCanvasElement | null>(null)
const loginError = ref<string | null>(null)
let stopGlobe: (() => void) | null = null

function initGlobe(canvas: HTMLCanvasElement) {
  const ctx = canvas.getContext('2d')

  if (!ctx) return () => {}

  let size = 0
  let centerX = 0
  let centerY = 0
  let radius = 0
  let animationFrameId = 0

  const nodeCount = 130
  const nodes: Array<{ x: number; y: number; z: number }> = []
  const goldenAngle = Math.PI * (1 + Math.sqrt(5))

  for (let i = 0; i < nodeCount; i += 1) {
    const phi = Math.acos(1 - (2 * (i + 0.5)) / nodeCount)
    const theta = goldenAngle * i

    nodes.push({
      x: Math.sin(phi) * Math.cos(theta),
      y: Math.sin(phi) * Math.sin(theta),
      z: Math.cos(phi),
    })
  }

  const edges: Array<[number, number]> = []

  for (let i = 0; i < nodeCount; i += 1) {
    for (let j = i + 1; j < nodeCount; j += 1) {
      const fromNode = nodes[i]
      const toNode = nodes[j]

      if (!fromNode || !toNode) continue

      const distance =
        fromNode.x * toNode.x + fromNode.y * toNode.y + fromNode.z * toNode.z

      if (distance > 0.89) {
        edges.push([i, j])
      }
    }
  }

  let angle = 0
  const tilt = 0.18
  const cosTilt = Math.cos(tilt)
  const sinTilt = Math.sin(tilt)

  const resizeCanvas = () => {
    const dpr = window.devicePixelRatio || 1
    size = canvas.offsetWidth
    centerX = size / 2
    centerY = size / 2
    radius = size * 0.46
    canvas.width = size * dpr
    canvas.height = size * dpr
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  }

  const project = (node: { x: number; y: number; z: number }) => {
    const cosAngle = Math.cos(angle)
    const sinAngle = Math.sin(angle)
    const rotatedX = node.x * cosAngle + node.z * sinAngle
    const rotatedY = node.y
    const rotatedZ = -node.x * sinAngle + node.z * cosAngle
    const finalY = rotatedY * cosTilt - rotatedZ * sinTilt
    const finalZ = rotatedY * sinTilt + rotatedZ * cosTilt
    const scale = 3 / (3 + finalZ)

    return {
      sx: centerX + rotatedX * radius * scale,
      sy: centerY + finalY * radius * scale,
      depth: finalZ,
    }
  }

  const draw = () => {
    ctx.clearRect(0, 0, size, size)

    const body = ctx.createRadialGradient(
      centerX - radius * 0.28,
      centerY - radius * 0.22,
      0,
      centerX,
      centerY,
      radius,
    )
    body.addColorStop(0, 'rgba(59,130,246,0.24)')
    body.addColorStop(1, 'rgba(15,23,42,0)')
    ctx.beginPath()
    ctx.arc(centerX, centerY, radius, 0, Math.PI * 2)
    ctx.fillStyle = body
    ctx.fill()

    const projectedNodes = nodes.map((node) => project(node))

    ctx.save()
    ctx.beginPath()
    ctx.arc(centerX, centerY, radius, 0, Math.PI * 2)
    ctx.clip()

    for (const [from, to] of edges) {
      const a = projectedNodes[from]
      const b = projectedNodes[to]

      if (!a || !b) continue

      const averageDepth = (a.depth + b.depth) / 2

      if (averageDepth < -0.5) continue

      const opacity = Math.max(0, (averageDepth + 1) / 2) * 0.7 + 0.1
      ctx.beginPath()
      ctx.moveTo(a.sx, a.sy)
      ctx.lineTo(b.sx, b.sy)
      ctx.strokeStyle = `rgba(59,130,246,${opacity * 0.5})`
      ctx.lineWidth = 1.5
      ctx.stroke()
    }

    ctx.restore()

    for (const point of projectedNodes) {
      if (point.depth < -0.65) continue

      const depthRatio = (point.depth + 1) / 2
      const opacity = depthRatio * 0.9 + 0.1
      const nodeSize = depthRatio * 4.5 + 1.2

      if (depthRatio > 0.5) {
        const glow = ctx.createRadialGradient(
          point.sx,
          point.sy,
          0,
          point.sx,
          point.sy,
          nodeSize * 4,
        )
        glow.addColorStop(0, `rgba(59,130,246,${opacity * 0.65})`)
        glow.addColorStop(1, 'rgba(96,165,250,0)')
        ctx.beginPath()
        ctx.arc(point.sx, point.sy, nodeSize * 4, 0, Math.PI * 2)
        ctx.fillStyle = glow
        ctx.fill()
      }

      ctx.beginPath()
      ctx.arc(point.sx, point.sy, nodeSize, 0, Math.PI * 2)
      ctx.fillStyle = `rgba(29,78,216,${opacity})`
      ctx.fill()
    }

    angle += 0.0025
    animationFrameId = requestAnimationFrame(draw)
  }

  resizeCanvas()
  draw()
  window.addEventListener('resize', resizeCanvas)

  return () => {
    cancelAnimationFrame(animationFrameId)
    window.removeEventListener('resize', resizeCanvas)
  }
}

const { defineField, errors, handleSubmit } = useForm<LoginFormValues>({
  validationSchema: toTypedSchema(loginSchema),
  initialValues: {
    username: '',
    password: '',
  },
})

const [username] = defineField('username')
const [password] = defineField('password')

const safeRedirectTarget = computed(() => {
  const redirect = route.query.redirect
  const target = Array.isArray(redirect) ? redirect[0] : redirect

  if (!target || !target.startsWith('/') || target.startsWith('//') || target.startsWith('/login')) {
    return null
  }

  return target
})

const onSubmit = handleSubmit(async (values) => {
  loginError.value = null

  try {
    await auth.login(values)
    await router.push(safeRedirectTarget.value ?? { name: 'dashboard' })
  } catch (err) {
    if (isAxiosError(err) && err.response?.status === 401) {
      loginError.value = 'Username atau password salah'
      return
    }

    loginError.value = 'Login gagal. Silakan coba lagi.'
  }
})

onMounted(() => {
  if (globeCanvas.value) {
    stopGlobe = initGlobe(globeCanvas.value)
  }
})

onBeforeUnmount(() => {
  stopGlobe?.()
})
</script>

<template>
  <main class="relative min-h-screen overflow-x-hidden px-5 py-8 sm:px-8 lg:px-12">
    <div class="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      <div
        class="absolute left-1/2 top-1/2 h-[460px] w-[460px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary-50/80 blur-3xl lg:left-[62%] lg:h-[680px] lg:w-[680px]"
      />
      <canvas
        ref="globeCanvas"
        class="absolute left-1/2 top-1/2 h-[880px] w-[880px] -translate-x-1/2 -translate-y-1/2 opacity-60 lg:left-[62%] lg:h-[1240px] lg:w-[1240px] lg:opacity-70"
      />
    </div>

    <div class="relative z-10 mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl items-center">
      <div class="grid w-full gap-10 lg:grid-cols-[minmax(0,1fr)_440px] lg:items-center lg:gap-16">
        <section class="space-y-10">
          <div class="max-w-2xl space-y-6">
            <div class="flex items-center gap-3">
              <div
                class="flex h-11 w-11 items-center justify-center rounded-md border border-primary-200 bg-primary-50 text-sm font-semibold text-primary-700"
                aria-hidden="true"
              >
                PR
              </div>
              <div>
                <p class="text-sm font-semibold uppercase tracking-[0.2em] text-primary">PRISM</p>
                <p class="text-sm text-surface-500">Internal monitoring system</p>
              </div>
            </div>

            <div class="space-y-4">
              <h1 class="max-w-xl text-4xl font-semibold leading-tight text-surface-950 sm:text-5xl">
                Project Loan Integrated Monitoring System
              </h1>
              <p class="max-w-2xl text-base leading-7 text-surface-600">
                Kelola alur pinjaman luar negeri dari perencanaan, perjanjian, sampai monitoring
                disbursement triwulanan dalam satu ruang kerja.
              </p>
            </div>
          </div>
        </section>

        <section
          class="w-full rounded-lg border border-surface-200 bg-white p-6 shadow-sm sm:p-8"
          aria-labelledby="login-title"
        >
          <div class="mb-8 space-y-2">
            <p class="text-sm font-semibold text-primary">Akses internal</p>
            <h2 id="login-title" class="text-2xl font-semibold text-surface-950">Masuk ke PRISM</h2>
            <p class="text-sm leading-6 text-surface-500">Gunakan akun yang sudah diberikan admin.</p>
          </div>

          <form class="space-y-5" @submit.prevent="onSubmit">
            <Message v-if="loginError" severity="error" size="small" :closable="false">
              {{ loginError }}
            </Message>

            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-700">Username</span>
              <InputText
                v-model="username"
                class="w-full"
                autocomplete="username"
                :invalid="Boolean(errors.username)"
                @input="loginError = null"
              />
              <small v-if="errors.username" class="text-red-600">{{ errors.username }}</small>
            </label>

            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-700">Password</span>
              <Password
                v-model="password"
                class="w-full"
                input-class="w-full"
                :input-props="{ autocomplete: 'current-password' }"
                :feedback="false"
                :invalid="Boolean(errors.password)"
                toggle-mask
                @input="loginError = null"
              />
              <small v-if="errors.password" class="text-red-600">{{ errors.password }}</small>
            </label>

            <Button
              type="submit"
              label="Masuk"
              icon="pi pi-sign-in"
              class="w-full"
              :loading="auth.loading"
            />
          </form>

          <div class="mt-8 border-t border-surface-200 pt-5">
            <p class="text-sm leading-6 text-surface-500">
              Akses modul dikelola oleh admin. Hubungi admin jika menu kerja belum tersedia setelah
              login.
            </p>
          </div>
        </section>
      </div>
    </div>
  </main>
</template>
