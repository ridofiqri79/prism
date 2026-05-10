<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { isNavigationFailure, useRoute, useRouter } from 'vue-router'
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

const workflowStages = [
  {
    step: 'Tahap 01',
    title: 'Blue Book',
    description: 'Usulan dan indikasi awal proyek.',
    image: '/landing/blue-book.png',
    markerClass: 'bg-blue-500',
  },
  {
    step: 'Tahap 02',
    title: 'Green Book',
    description: 'Prioritas pendanaan dan kegiatan.',
    image: '/landing/green-book.png',
    markerClass: 'bg-emerald-500',
  },
  {
    step: 'Tahap 03',
    title: 'Daftar Kegiatan',
    description: 'Surat kegiatan dan pembiayaan.',
    image: '/landing/daftar-kegiatan.png',
    markerClass: 'bg-amber-500',
  },
  {
    step: 'Tahap 04',
    title: 'Perjanjian Pinjaman',
    description: 'Komitmen legal dan kinerja pinjaman.',
    image: '/landing/loan-agreements.png',
    markerClass: 'bg-violet-500',
  },
]

const reviewModules = [
  {
    title: 'Dashboard',
    icon: 'pi pi-chart-line',
    description:
      'Membaca funnel tahap, nilai portofolio, pemberi pinjaman, instansi, wilayah, dan program.',
  },
  {
    title: 'Proyek',
    icon: 'pi pi-table',
    description:
      'Mencari proyek lintas dokumen dengan filter aktif dan hasil ekspor sesuai konteks peninjauan.',
  },
  {
    title: 'Perjalanan Proyek',
    icon: 'pi pi-sitemap',
    description: 'Menelusuri riwayat satu proyek dari Blue Book sampai Perjanjian Pinjaman.',
  },
  {
    title: 'Sebaran Wilayah',
    icon: 'pi pi-map',
    description: 'Menganalisis konsentrasi proyek menurut provinsi atau kabupaten/kota.',
  },
]

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
    target === '/' ||
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
  } catch (err) {
    if (isAxiosError(err) && err.response?.status === 401) {
      loginError.value = 'Username atau password salah'
      return
    }

    loginError.value = 'Login gagal. Silakan coba lagi.'
    return
  }

  try {
    await router.replace(
      safeRedirectTarget.value ??
        resolveDefaultAuthenticatedRoute({
          user: auth.user,
          permissions: auth.permissions,
        }),
    )
  } catch (err) {
    if (!isNavigationFailure(err)) {
      throw err
    }
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
  <main
    class="relative min-h-screen overflow-hidden bg-[#f2fbfb] text-surface-950 [&_*]:box-border"
  >
    <div
      class="pointer-events-none absolute inset-x-0 top-0 h-[860px] overflow-hidden"
      aria-hidden="true"
    >
      <div
        class="absolute inset-0 bg-[radial-gradient(circle_at_18%_18%,rgba(253,184,19,0.16),transparent_26%),radial-gradient(circle_at_78%_8%,rgba(31,181,178,0.22),transparent_30%),linear-gradient(180deg,rgba(255,255,255,0.74),rgba(242,251,251,0))]"
      />
      <canvas
        ref="globeCanvas"
        class="absolute -right-48 top-8 h-[760px] w-[760px] opacity-20 lg:-right-36 lg:h-[980px] lg:w-[980px] lg:opacity-25"
      />
    </div>

    <section
      class="relative z-10 mx-auto grid min-h-screen w-full max-w-7xl items-center gap-8 px-5 py-12 sm:px-8 lg:grid-cols-[minmax(0,1.08fr)_448px] lg:gap-12 lg:py-16"
      aria-labelledby="landing-title"
    >
      <div class="w-full min-w-0 max-w-[350px] sm:max-w-none">
        <p class="mb-4 text-xs font-semibold uppercase text-prism-teal-dark">
          Project Loan Integrated Monitoring System
        </p>
        <h1
          id="landing-title"
          class="max-w-3xl text-5xl font-extrabold leading-[1.02] text-surface-950"
        >
          PRISM
        </h1>
        <p class="mt-6 max-w-2xl text-lg leading-8 text-surface-700">
          Mengelola informasi pinjaman luar negeri dari usulan awal, pematangan prioritas, penetapan
          kegiatan, sampai menjadi perjanjian pinjaman.
        </p>

        <div class="mt-8 flex flex-col gap-3 sm:flex-row">
          <a
            href="#workflow"
            class="inline-flex h-12 w-full items-center justify-center gap-2 rounded-lg border border-prism-teal/26 bg-white/78 px-6 text-sm font-semibold text-prism-teal-dark transition hover:border-prism-teal hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-prism-gold sm:w-auto"
          >
            Lihat alur kerja
            <i class="pi pi-compass text-xs" aria-hidden="true" />
          </a>
        </div>

        <figure
          class="mt-10 overflow-hidden rounded-lg border border-white/80 bg-white shadow-[0_24px_70px_rgba(11,111,115,0.12)]"
        >
          <img
            src="/landing/dashboard.png"
            alt="Dashboard PRISM"
            class="aspect-[16/9] w-full object-cover object-left-top"
          />
        </figure>
      </div>

      <section
        id="login-form"
        class="w-full max-w-[350px] rounded-lg border border-white/80 bg-white px-6 py-7 shadow-[0_24px_70px_rgba(11,111,115,0.12)] sm:max-w-none sm:px-8 sm:py-8 lg:sticky lg:top-6"
        aria-labelledby="login-title"
      >
        <div class="text-prism-teal-dark">
          <div class="mb-8 border-b border-surface-200 pb-8 text-center">
            <div
              class="mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-lg border border-prism-teal/18 bg-prism-teal/10 shadow-[0_10px_28px_rgba(31,181,178,0.12)]"
            >
              <img src="/prism-logo.png" alt="Logo PRISM" class="h-11 w-11 object-contain" />
            </div>
            <p class="text-2xl font-extrabold leading-none text-surface-900">PRISM</p>
            <h1 class="mt-3 text-xs font-semibold uppercase leading-5 text-surface-700">
              Project Loan Integrated Monitoring System
            </h1>
          </div>

          <div class="mb-7">
            <div class="flex items-center justify-between gap-4">
              <h2 id="login-title" class="text-xl font-semibold leading-tight text-surface-900">
                Masuk
              </h2>
              <span
                class="rounded-md bg-prism-teal/10 px-3 py-1.5 text-xs font-medium text-prism-teal-dark"
              >
                Akses Internal
              </span>
            </div>
            <p class="mt-4 text-sm leading-6 text-surface-600">
              Masukkan username dan kata sandi untuk membuka ruang kerja sesuai hak akses.
            </p>
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
                :input-props="{
                  autocomplete: 'current-password',
                  placeholder: 'Masukkan password',
                }"
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

          <p class="mt-5 text-xs leading-5 text-surface-500">
            Jika menu yang dibutuhkan belum muncul, hubungi ADMIN untuk pengecekan hak akses modul.
          </p>
        </div>
      </section>
    </section>

    <section
      id="workflow"
      class="relative z-10 border-y border-prism-teal/12 bg-white/72 py-16 sm:py-20"
    >
      <div class="mx-auto w-full max-w-7xl px-5 sm:px-8">
        <div class="max-w-3xl">
          <p class="text-xs font-semibold uppercase text-prism-teal-dark">
            Alur dokumen perencanaan
          </p>
          <h2 class="mt-3 text-3xl font-bold leading-tight text-surface-950">
            Satu jejak proyek dari usulan sampai komitmen legal.
          </h2>
          <p class="mt-4 text-base leading-7 text-surface-700">
            PRISM menjaga hubungan antar dokumen agar perubahan versi, status tahap, nilai, lender,
            instansi, dan wilayah tetap dapat ditinjau dalam satu konteks.
          </p>
        </div>

        <div class="mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <article
            v-for="stage in workflowStages"
            :key="stage.title"
            class="group overflow-hidden rounded-lg border border-surface-200 bg-white shadow-[0_14px_34px_rgba(15,143,140,0.07)] transition duration-300 hover:-translate-y-1 hover:border-prism-teal/40"
          >
            <img
              :src="stage.image"
              :alt="`Tampilan ${stage.title}`"
              class="h-36 w-full object-cover object-left-top"
            />
            <div class="p-5">
              <div class="flex items-center gap-2 text-xs font-semibold uppercase text-surface-500">
                <span class="h-2.5 w-2.5 rounded-full" :class="stage.markerClass" />
                {{ stage.step }}
              </div>
              <h3 class="mt-4 text-lg font-semibold text-surface-950">{{ stage.title }}</h3>
              <p class="mt-2 text-sm leading-6 text-surface-600">{{ stage.description }}</p>
            </div>
          </article>
        </div>
      </div>
    </section>

    <section class="relative z-10 bg-[#f8fdfd] py-16 sm:py-20">
      <div
        class="mx-auto grid w-full max-w-7xl gap-10 px-5 sm:px-8 lg:grid-cols-[minmax(0,0.95fr)_minmax(360px,1fr)] lg:items-center"
      >
        <div>
          <p class="text-xs font-semibold uppercase text-prism-teal-dark">Peninjauan portofolio</p>
          <h2 class="mt-3 text-3xl font-bold leading-tight text-surface-950">
            Dashboard, proyek, perjalanan, dan wilayah berada dalam satu ruang kerja.
          </h2>
          <div class="mt-8 grid gap-4">
            <article
              v-for="module in reviewModules"
              :key="module.title"
              class="flex gap-4 rounded-lg border border-surface-200 bg-white p-4 transition duration-300 hover:border-prism-teal/40 hover:shadow-[0_12px_30px_rgba(15,143,140,0.08)]"
            >
              <span
                class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-prism-teal/10 text-prism-teal-dark"
              >
                <i :class="module.icon" aria-hidden="true" />
              </span>
              <span>
                <strong class="block text-sm font-semibold text-surface-950">{{
                  module.title
                }}</strong>
                <span class="mt-1 block text-sm leading-6 text-surface-600">{{
                  module.description
                }}</span>
              </span>
            </article>
          </div>
        </div>

        <div class="grid gap-4">
          <figure
            class="overflow-hidden rounded-lg border border-white/80 bg-white shadow-[0_24px_70px_rgba(11,111,115,0.1)]"
          >
            <img
              src="/landing/project-master.png"
              alt="Halaman Proyek PRISM"
              class="aspect-[16/10] w-full object-cover object-left-top"
            />
          </figure>
          <figure
            class="overflow-hidden rounded-lg border border-white/80 bg-white shadow-[0_16px_42px_rgba(11,111,115,0.08)]"
          >
            <img
              src="/landing/spatial.png"
              alt="Sebaran Wilayah PRISM"
              class="aspect-[16/7] w-full object-cover object-left-top"
            />
          </figure>
        </div>
      </div>
    </section>
  </main>
</template>
