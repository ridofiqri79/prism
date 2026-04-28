import type { RouteLocationNormalizedLoaded } from 'vue-router'

export type RouteTitle = string | ((route: RouteLocationNormalizedLoaded) => string)

export const DEFAULT_ROUTE_TITLE = 'PRISM'

export function resolveRouteTitle(route: RouteLocationNormalizedLoaded): string {
  const matchedRoute = [...route.matched].reverse().find((record) => record.meta.title)
  const title = matchedRoute?.meta.title ?? route.meta.title

  if (typeof title === 'function') {
    return title(route).trim() || DEFAULT_ROUTE_TITLE
  }

  if (typeof title === 'string') {
    return title.trim() || DEFAULT_ROUTE_TITLE
  }

  return DEFAULT_ROUTE_TITLE
}
