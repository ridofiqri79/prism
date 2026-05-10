import type { LocationQuery } from 'vue-router'

function rawQueryValues(query: LocationQuery, keys: string[]) {
  return keys.flatMap((key) => {
    const value = query[key]
    if (Array.isArray(value)) return value
    return value === undefined ? [] : [value]
  })
}

export function hasQueryParam(query: LocationQuery, ...keys: string[]) {
  return keys.some((key) => Object.prototype.hasOwnProperty.call(query, key))
}

export function queryStringValues(query: LocationQuery, ...keys: string[]) {
  const values = rawQueryValues(query, keys)
    .filter((value): value is string => typeof value === 'string')
    .flatMap((value) => value.split(','))
    .map((value) => value.trim())
    .filter(Boolean)

  return [...new Set(values)]
}

export function queryString(query: LocationQuery, ...keys: string[]) {
  return queryStringValues(query, ...keys)[0]
}

export function queryBoolean(query: LocationQuery, ...keys: string[]) {
  const value = queryString(query, ...keys)?.toLowerCase()
  if (value === 'true' || value === '1' || value === 'yes') return true
  if (value === 'false' || value === '0' || value === 'no') return false
  return null
}

export function queryNumber(query: LocationQuery, ...keys: string[]) {
  const value = queryString(query, ...keys)
  if (!value) return undefined

  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

export function queryEnum<T extends string>(
  query: LocationQuery,
  allowedValues: readonly T[],
  ...keys: string[]
) {
  const value = queryString(query, ...keys)
  return allowedValues.find((item) => item === value)
}

export function queryEnumArray<T extends string>(
  query: LocationQuery,
  allowedValues: readonly T[],
  ...keys: string[]
) {
  const allowed = new Set<string>(allowedValues)
  return queryStringValues(query, ...keys).filter((value): value is T => allowed.has(value))
}
