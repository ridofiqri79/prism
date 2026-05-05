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
import { resolveDefaultAuthenticatedRoute } from '@/utils/default-route'

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

      const distance = fromNode.x * toNode.x + fromNode.y * toNode.y + fromNode.z * toNode.z

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
    body.addColorStop(0, 'rgba(255,212,90,0.16)')
    body.addColorStop(0.3, 'rgba(31,181,178,0.24)')
    body.addColorStop(0.72, 'rgba(38,183,165,0.1)')
    body.addColorStop(1, 'rgba(11,111,115,0)')
    ctx.beginPath()
    ctx.arc(centerX, centerY, radius, 0, Math.PI * 2)
    ctx.fillStyle = body
    ctx.fill()

    const rim = ctx.createLinearGradient(
      centerX - radius,
      centerY - radius,
      centerX + radius,
      centerY + radius,
    )
    rim.addColorStop(0, 'rgba(255,212,90,0.26)')
    rim.addColorStop(0.44, 'rgba(31,181,178,0.3)')
    rim.addColorStop(1, 'rgba(21,126,92,0.16)')
    ctx.beginPath()
    ctx.arc(centerX, centerY, radius * 0.995, 0, Math.PI * 2)
    ctx.strokeStyle = rim
    ctx.lineWidth = 1.3
    ctx.stroke()

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

      const opacity = Math.max(0, (averageDepth + 1) / 2) * 0.55 + 0.12
      ctx.beginPath()
      ctx.moveTo(a.sx, a.sy)
      ctx.lineTo(b.sx, b.sy)
      ctx.strokeStyle = `rgba(31,181,178,${opacity * 0.46})`
      ctx.lineWidth = averageDepth > 0.42 ? 1.45 : 1.05
      ctx.stroke()
    }

    ctx.restore()

    for (const point of projectedNodes) {
      if (point.depth < -0.65) continue

      const depthRatio = (point.depth + 1) / 2
      const opacity = depthRatio * 0.78 + 0.14
      const nodeSize = depthRatio * 3.8 + 1.1

      if (depthRatio > 0.58) {
        const glow = ctx.createRadialGradient(
          point.sx,
          point.sy,
          0,
          point.sx,
          point.sy,
          nodeSize * 4,
        )
        glow.addColorStop(0, `rgba(255,212,90,${opacity * 0.32})`)
        glow.addColorStop(0.42, `rgba(31,181,178,${opacity * 0.12})`)
        glow.addColorStop(1, 'rgba(31,181,178,0)')
        ctx.beginPath()
        ctx.arc(point.sx, point.sy, nodeSize * 4, 0, Math.PI * 2)
        ctx.fillStyle = glow
        ctx.fill()
      }

      ctx.beginPath()
      ctx.arc(point.sx, point.sy, nodeSize, 0, Math.PI * 2)
      ctx.fillStyle =
        depthRatio > 0.72
          ? `rgba(253,184,19,${opacity * 0.9})`
          : `rgba(31,181,178,${opacity * 0.88})`
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

  if (
    !target ||
    !target.startsWith('/') ||
    target.startsWith('//') ||
    target.startsWith('/login')
  ) {
    return null
  }

  return target
})

const onSubmit = handleSubmit(async (values) => {
  loginError.value = null

  try {
    await auth.login(values)
    await router.push(
      safeRedirectTarget.value ??
        resolveDefaultAuthenticatedRoute({
          user: auth.user,
          permissions: auth.permissions,
        }),
    )
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
  <main class="relative flex min-h-screen flex-col overflow-hidden bg-[#e8f7f7] px-5 sm:px-8">
    <div class="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden="true">
      <div
        class="absolute left-1/2 top-1/2 h-[460px] w-[460px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary-50/80 blur-3xl lg:h-[680px] lg:w-[680px]"
      />
      <canvas
        ref="globeCanvas"
        class="absolute left-1/2 top-1/2 h-[880px] w-[880px] -translate-x-1/2 -translate-y-1/2 opacity-20 lg:h-[1240px] lg:w-[1240px] lg:opacity-25"
      />
    </div>

    <div class="relative z-10 mx-auto flex w-full max-w-6xl flex-1 items-center justify-center py-10">
      <section
        class="w-full max-w-[448px] rounded-2xl border border-white/70 bg-white px-10 py-10 shadow-[0_24px_70px_rgba(11,111,115,0.08)] sm:px-10 sm:py-10"
        aria-labelledby="login-title"
      >
        <div class="text-prism-teal-dark">
          <div class="mb-8 border-b border-surface-200 pb-8 text-center">
            <div
              class="mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-full border border-prism-teal/18 bg-prism-teal/10 shadow-[0_10px_28px_rgba(31,181,178,0.12)]"
            >
              <img src="/prism-logo.png" alt="Logo PRISM" class="h-11 w-11 object-contain" />
            </div>
            <p class="text-[1.55rem] font-extrabold leading-none tracking-wide text-surface-900">PRISM</p>
            <h1 class="mt-3 text-[0.62rem] font-semibold uppercase tracking-[0.32em] text-surface-700">
              Project Loan Integrated Monitoring System
            </h1>
          </div>

          <div class="mb-7">
            <div class="flex items-center justify-between gap-4">
              <h2 id="login-title" class="text-xl font-semibold leading-tight text-surface-900">Login</h2>
              <span class="rounded-md bg-surface-100 px-3 py-1.5 text-xs font-medium text-surface-600">
                Akses Internal
              </span>
            </div>
            <p class="mt-4 text-sm leading-6 text-surface-600">Silakan masukkan kredensial Anda.</p>
          </div>

          <form class="space-y-5" @submit.prevent="onSubmit">
            <Message v-if="loginError" severity="error" size="small" :closable="false">
              {{ loginError }}
            </Message>

            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-900">Username</span>
              <InputText
                v-model="username"
                class="h-[46px] w-full border-surface-300 bg-surface-50 px-4 text-surface-950 placeholder:text-surface-400"
                autocomplete="username"
                placeholder="Masukkan username"
                :invalid="Boolean(errors.username)"
                @input="loginError = null"
              />
              <small v-if="errors.username" class="text-red-600">{{ errors.username }}</small>
            </label>

            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-900">Password</span>
              <Password
                v-model="password"
                class="w-full"
                input-class="h-[46px] w-full border-surface-300 bg-surface-50 px-4 text-surface-950 placeholder:text-surface-400"
                :input-props="{ autocomplete: 'current-password', placeholder: 'Masukkan password' }"
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
              icon="pi pi-arrow-right"
              icon-pos="right"
              class="mt-3 h-11 w-full border-prism-teal-deep bg-prism-teal-deep font-semibold text-white hover:border-prism-teal-dark hover:bg-prism-teal-dark"
              :loading="auth.loading"
            />
          </form>
        </div>
      </section>
    </div>
  </main>
</template>
