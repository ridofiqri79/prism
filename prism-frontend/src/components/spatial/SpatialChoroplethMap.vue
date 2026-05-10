<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type {
  SpatialDistributionLevel,
  SpatialDistributionMetric,
  SpatialDistributionRegionMetric,
  SpatialMapFeature,
  SpatialMapFeatureCollection,
} from '@/types/spatial-distribution.types'

type GeoCoordinate = [number, number]
type GeoRing = GeoCoordinate[]
type GeoPolygon = GeoRing[]
type GeoMultiPolygon = GeoPolygon[]

interface MapPath {
  key: string
  code: string
  name: string
  d: string
  metric: SpatialDistributionRegionMetric | null
}

const props = defineProps<{
  level: SpatialDistributionLevel
  metric: SpatialDistributionMetric
  provinceCode?: string
  provinceName?: string
  regions: SpatialDistributionRegionMetric[]
  selectedRegionCode?: string | null
  loading?: boolean
  compact?: boolean
}>()

const emit = defineEmits<{
  select: [region: SpatialDistributionRegionMetric]
}>()

const geoJson = ref<SpatialMapFeatureCollection | null>(null)
const mapLoading = ref(false)
const mapError = ref<string | null>(null)
const hoveredPath = ref<MapPath | null>(null)
let mapRequestId = 0
let mapAbortController: AbortController | null = null

const metricLabel = computed(() =>
  props.metric === 'count' ? 'Jumlah proyek' : 'Nilai pinjaman USD',
)

const hoverDetail = computed(() => {
  const path = hoveredPath.value
  if (!path) return null

  const region = path.metric
  const projectCount = region?.project_count ?? 0
  const totalLoanUsd = region?.total_loan_usd ?? 0

  return {
    title: region?.region_name ?? path.name,
    typeLabel: region ? regionTypeLabel(region.region_type) : 'Wilayah',
    projectLabel: `${new Intl.NumberFormat('id-ID').format(projectCount)} proyek`,
    loanLabel: formatUsd(totalLoanUsd),
    activeMetricLabel:
      props.metric === 'count'
        ? `${new Intl.NumberFormat('id-ID').format(projectCount)} proyek`
        : formatUsd(totalLoanUsd),
    hasData: projectCount > 0 || totalLoanUsd > 0,
  }
})

const regionByCode = computed(() => new Map(props.regions.map((item) => [item.region_code, item])))
const regionByName = computed(() => new Map(props.regions.map((item) => [item.region_name, item])))

const rawMaxValue = computed(() =>
  props.metric === 'count'
    ? Math.max(...props.regions.map((item) => item.project_count), 0)
    : Math.max(...props.regions.map((item) => item.total_loan_usd), 0),
)

const maxValue = computed(() => Math.max(rawMaxValue.value, 1))

const legendMinLabel = computed(() => formatLegendValue(0))
const legendMaxLabel = computed(() => formatLegendValue(rawMaxValue.value))
const legendRangeLabel = computed(() => `${legendMinLabel.value} - ${legendMaxLabel.value}`)

const mapShellSizeClass = computed(() =>
  props.compact ? 'min-h-[22rem] lg:min-h-[23rem] xl:min-h-[24rem]' : 'min-h-[31rem]',
)

const mapCanvasPaddingClass = computed(() => (props.compact ? 'px-3 py-5' : 'px-4 py-8'))

const mapBounds = computed(() => {
  const features = geoJson.value?.features ?? []
  let minLon = Infinity
  let minLat = Infinity
  let maxLon = -Infinity
  let maxLat = -Infinity

  for (const feature of features) {
    for (const polygon of getGeometryPolygons(feature)) {
      for (const ring of polygon) {
        for (const [lon, lat] of ring) {
          minLon = Math.min(minLon, lon)
          minLat = Math.min(minLat, lat)
          maxLon = Math.max(maxLon, lon)
          maxLat = Math.max(maxLat, lat)
        }
      }
    }
  }

  if (![minLon, minLat, maxLon, maxLat].every(Number.isFinite)) {
    return null
  }

  const width = Math.max(maxLon - minLon, 1)
  const height = Math.max(maxLat - minLat, 1)
  const pad = Math.max(width, height) * 0.035

  return {
    minX: minLon - pad,
    minY: -maxLat - pad,
    width: width + pad * 2,
    height: height + pad * 2,
  }
})

const viewBox = computed(() => {
  if (!mapBounds.value) return '0 0 100 60'
  const { minX, minY, width, height } = mapBounds.value
  return `${minX} ${minY} ${width} ${height}`
})

const mapPaths = computed<MapPath[]>(() =>
  (geoJson.value?.features ?? [])
    .map((feature, index) => {
      const code = readFeatureCode(feature)
      const name = readFeatureName(feature)
      const metric = code
        ? (regionByCode.value.get(code) ?? null)
        : (regionByName.value.get(name) ?? null)
      const d = getGeometryPolygons(feature).map(polygonToPath).join(' ')

      return {
        key: `${code || name || 'region'}-${index}`,
        code,
        name,
        d,
        metric,
      }
    })
    .filter((item) => item.d.length > 0),
)

async function loadMapAsset() {
  hoveredPath.value = null
  const code = props.provinceCode
  const assetPath =
    props.level === 'city' && code ? `/maps/cities/${code}.json` : '/maps/indonesia-provinces.json'
  const requestId = ++mapRequestId
  mapAbortController?.abort()
  const controller = new AbortController()
  mapAbortController = controller

  mapLoading.value = true
  mapError.value = null
  try {
    const response = await fetch(assetPath, { signal: controller.signal })
    if (!response.ok) {
      throw new Error(`Failed to load ${assetPath}`)
    }

    const nextGeoJson = (await response.json()) as SpatialMapFeatureCollection
    if (requestId === mapRequestId) {
      geoJson.value = nextGeoJson
    }
  } catch (err) {
    if ((err as { name?: string }).name === 'AbortError' || requestId !== mapRequestId) return
    geoJson.value = null
    mapError.value = 'Asset peta tidak dapat dimuat.'
  } finally {
    if (requestId === mapRequestId) {
      mapLoading.value = false
      if (mapAbortController === controller) {
        mapAbortController = null
      }
    }
  }
}

function getGeometryPolygons(feature: SpatialMapFeature): GeoMultiPolygon {
  if (!feature.geometry || typeof feature.geometry !== 'object') return []

  const geometry = feature.geometry as {
    type?: string
    coordinates?: unknown
  }

  if (geometry.type === 'Polygon' && Array.isArray(geometry.coordinates)) {
    return [geometry.coordinates as GeoPolygon]
  }

  if (geometry.type === 'MultiPolygon' && Array.isArray(geometry.coordinates)) {
    return geometry.coordinates as GeoMultiPolygon
  }

  return []
}

function polygonToPath(polygon: GeoPolygon) {
  return polygon
    .map((ring) => {
      const points = ring
        .filter((point): point is GeoCoordinate => Array.isArray(point) && point.length >= 2)
        .map(([lon, lat], index) => `${index === 0 ? 'M' : 'L'} ${lon} ${-lat}`)

      return points.length ? `${points.join(' ')} Z` : ''
    })
    .join(' ')
}

function readFeatureCode(feature: SpatialMapFeature) {
  const properties = feature.properties ?? {}
  const code = properties.REGION_CODE ?? properties.PROVINCE_CODE
  return typeof code === 'string' ? code : String(feature.id ?? '')
}

function readFeatureName(feature: SpatialMapFeature) {
  const name = feature.properties?.name
  return typeof name === 'string' ? name : String(feature.id ?? 'Wilayah')
}

function metricValue(region: SpatialDistributionRegionMetric | null) {
  if (!region) return 0
  return props.metric === 'count' ? region.project_count : region.total_loan_usd
}

function pathFill(path: MapPath) {
  const value = metricValue(path.metric)
  if (path.code === props.selectedRegionCode) return '#fdb813'
  if (value <= 0) return '#e8eef3'

  return interpolateColor('#cfeff0', '#0b6f73', Math.min(value / maxValue.value, 1))
}

function pathStroke(path: MapPath) {
  if (path.code === props.selectedRegionCode || hoveredPath.value?.key === path.key) {
    return '#0b6f73'
  }

  return '#ffffff'
}

function pathStrokeWidth(path: MapPath) {
  const width = Math.max(mapBounds.value?.width ?? 100, mapBounds.value?.height ?? 60)
  if (path.code === props.selectedRegionCode) return width * 0.0024
  if (hoveredPath.value?.key === path.key) return width * 0.0018
  return width * 0.0011
}

function interpolateColor(start: string, end: string, ratio: number) {
  const left = hexToRgb(start)
  const right = hexToRgb(end)
  const rgb: [number, number, number] = [
    Math.round(left[0] + (right[0] - left[0]) * ratio),
    Math.round(left[1] + (right[1] - left[1]) * ratio),
    Math.round(left[2] + (right[2] - left[2]) * ratio),
  ]
  return `rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`
}

function hexToRgb(hex: string): [number, number, number] {
  const normalized = hex.replace('#', '')
  return [
    Number.parseInt(normalized.slice(0, 2), 16),
    Number.parseInt(normalized.slice(2, 4), 16),
    Number.parseInt(normalized.slice(4, 6), 16),
  ]
}

function tooltipTitle(path: MapPath) {
  const region = path.metric
  if (!region) return path.name

  const valueLabel =
    props.metric === 'count'
      ? `${new Intl.NumberFormat('id-ID').format(region.project_count)} proyek`
      : formatUsd(region.total_loan_usd)

  return `${region.region_name} | ${metricLabel.value}: ${valueLabel}`
}

function regionTypeLabel(type: SpatialDistributionRegionMetric['region_type']) {
  if (type === 'COUNTRY') return 'Nasional'
  if (type === 'PROVINCE') return 'Provinsi'
  return 'Kabupaten/Kota'
}

function formatUsd(value: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

function formatLegendValue(value: number) {
  if (props.metric === 'count') {
    return `${new Intl.NumberFormat('id-ID').format(value)} proyek`
  }

  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(value)
}

function selectPath(path: MapPath) {
  const region = path.metric
  if (region) {
    emit('select', region)
  }
}

function setHoveredPath(path: MapPath) {
  hoveredPath.value = path
}

function clearHoveredPath(path: MapPath) {
  if (hoveredPath.value?.key === path.key) {
    hoveredPath.value = null
  }
}

watch(
  () => [props.level, props.provinceCode] as const,
  () => {
    void loadMapAsset()
  },
  { immediate: true },
)
</script>

<template>
  <div
    class="relative flex items-center justify-center overflow-hidden rounded-lg border border-surface-200 bg-white"
    :class="mapShellSizeClass"
  >
    <div
      class="absolute left-4 top-4 z-10 rounded-full border border-white/70 bg-white/90 px-3 py-1 text-xs font-semibold text-surface-600 shadow-sm"
    >
      {{ level === 'province' ? 'Indonesia' : provinceName }}
    </div>

    <div
      v-if="hoverDetail"
      class="pointer-events-none absolute right-4 top-4 z-20 w-64 rounded-lg border border-surface-200 bg-white/95 p-3 text-xs text-surface-600 shadow-lg"
      role="status"
      aria-live="polite"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="font-semibold text-surface-950">{{ hoverDetail.title }}</p>
          <p class="mt-0.5 text-surface-500">{{ hoverDetail.typeLabel }}</p>
        </div>
        <span
          class="shrink-0 rounded-md px-2 py-1 font-semibold"
          :class="
            hoverDetail.hasData
              ? 'bg-primary-50 text-primary-700'
              : 'bg-surface-100 text-surface-500'
          "
        >
          {{ hoverDetail.activeMetricLabel }}
        </span>
      </div>
      <dl class="mt-3 grid grid-cols-2 gap-2">
        <div class="rounded-md bg-surface-50 px-2 py-1.5">
          <dt class="text-[11px] text-surface-500">Proyek</dt>
          <dd class="mt-0.5 font-semibold text-surface-900">{{ hoverDetail.projectLabel }}</dd>
        </div>
        <div class="rounded-md bg-surface-50 px-2 py-1.5">
          <dt class="text-[11px] text-surface-500">Nilai</dt>
          <dd class="mt-0.5 font-semibold text-surface-900">{{ hoverDetail.loanLabel }}</dd>
        </div>
      </dl>
    </div>

    <div
      class="absolute bottom-4 left-4 z-10 min-w-52 rounded-lg border border-surface-200 bg-white/95 px-3 py-2 text-xs text-surface-600 shadow-sm"
      :aria-label="`Skala ${metricLabel}: ${legendRangeLabel}`"
    >
      <div class="flex items-center justify-between gap-3">
        <span class="font-semibold text-surface-700">{{ metricLabel }}</span>
        <span class="text-surface-500">{{ legendRangeLabel }}</span>
      </div>
      <div
        class="mt-2 h-2.5 rounded-full bg-gradient-to-r from-[#cfeff0] via-[#1fb5b2] to-[#0b6f73]"
      ></div>
      <div class="mt-1 flex justify-between text-[11px] text-surface-500">
        <span>{{ legendMinLabel }}</span>
        <span>{{ legendMaxLabel }}</span>
      </div>
    </div>

    <div
      v-if="!mapError && mapPaths.length"
      class="absolute inset-0 grid place-items-center"
      :class="mapCanvasPaddingClass"
    >
      <svg
        :viewBox="viewBox"
        class="block h-full w-full bg-white"
        preserveAspectRatio="xMidYMid meet"
        role="img"
        aria-label="Peta choropleth sebaran wilayah"
      >
        <path
          v-for="path in mapPaths"
          :key="path.key"
          :d="path.d"
          :fill="pathFill(path)"
          :stroke="pathStroke(path)"
          :stroke-width="pathStrokeWidth(path)"
          class="cursor-pointer outline-none transition-[fill,opacity] duration-150 hover:opacity-80 focus:outline-none focus-visible:outline-none"
          fill-rule="evenodd"
          style="outline: none"
          vector-effect="non-scaling-stroke"
          role="button"
          tabindex="0"
          focusable="true"
          :aria-label="tooltipTitle(path)"
          @pointerenter="setHoveredPath(path)"
          @pointerleave="clearHoveredPath(path)"
          @mousemove="setHoveredPath(path)"
          @mouseenter="setHoveredPath(path)"
          @mouseleave="clearHoveredPath(path)"
          @focusin="setHoveredPath(path)"
          @focusout="clearHoveredPath(path)"
          @focus="setHoveredPath(path)"
          @blur="clearHoveredPath(path)"
          @click="selectPath(path)"
          @keydown.enter.prevent="selectPath(path)"
          @keydown.space.prevent="selectPath(path)"
        >
          <title>{{ tooltipTitle(path) }}</title>
        </path>
      </svg>
    </div>

    <div
      v-if="mapLoading || loading"
      class="absolute inset-0 grid place-items-center bg-white/70 text-sm font-medium text-surface-600 backdrop-blur-sm"
    >
      Memuat peta...
    </div>

    <div
      v-if="!mapLoading && !loading && !mapError && !mapPaths.length"
      class="absolute inset-0 grid place-items-center p-6 text-center"
    >
      <div>
        <p class="text-base font-semibold text-surface-900">Peta belum memiliki geometri</p>
        <p class="mt-1 text-sm text-surface-500">
          Asset peta tersedia, tetapi tidak ada wilayah yang dapat digambar.
        </p>
      </div>
    </div>

    <div v-if="mapError" class="absolute inset-0 grid place-items-center p-6 text-center">
      <div>
        <p class="text-base font-semibold text-surface-900">Peta tidak tersedia</p>
        <p class="mt-1 text-sm text-surface-500">{{ mapError }}</p>
      </div>
    </div>
  </div>
</template>
