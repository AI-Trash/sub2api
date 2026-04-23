import type { LocationQuery, LocationQueryValue } from 'vue-router'

export const INVITATION_LINK_QUERY_KEY = 'invite'
export const INVITATION_LINK_STORAGE_KEY = 'auth_invitation_code'

const INVITATION_QUERY_KEYS = [
  INVITATION_LINK_QUERY_KEY,
  'invitation_code',
  'invitationCode'
] as const

function normalizeInvitationValue(
  value: LocationQueryValue | LocationQueryValue[] | string | string[] | null | undefined
): string {
  if (Array.isArray(value)) {
    for (const entry of value) {
      const normalizedEntry = normalizeInvitationValue(entry)
      if (normalizedEntry) {
        return normalizedEntry
      }
    }
    return ''
  }

  return typeof value === 'string' ? value.trim() : ''
}

function normalizeBasePath(basePath: string | undefined): string {
  const normalized = (basePath || '/').trim()
  if (!normalized || normalized === '/') {
    return '/'
  }
  return `/${normalized.replace(/^\/+|\/+$/g, '')}/`
}

function safeSessionStorage(): Storage | null {
  if (typeof sessionStorage === 'undefined') {
    return null
  }

  try {
    return sessionStorage
  } catch {
    return null
  }
}

export function extractInvitationCodeFromQuery(
  query: LocationQuery | URLSearchParams | Record<string, unknown> | undefined
): string {
  if (!query) {
    return ''
  }

  if (query instanceof URLSearchParams) {
    for (const key of INVITATION_QUERY_KEYS) {
      const code = query.get(key)?.trim() || ''
      if (code) {
        return code
      }
    }
    return ''
  }

  const queryRecord = query as Record<string, unknown>
  for (const key of INVITATION_QUERY_KEYS) {
    const code = normalizeInvitationValue(
      queryRecord[key] as LocationQueryValue | LocationQueryValue[] | string | string[] | null | undefined
    )
    if (code) {
      return code
    }
  }

  return ''
}

export function getPersistedInvitationCode(): string {
  const storage = safeSessionStorage()
  if (!storage) {
    return ''
  }

  return storage.getItem(INVITATION_LINK_STORAGE_KEY)?.trim() || ''
}

export function persistInvitationCode(code: string | null | undefined): string {
  const normalized = (code || '').trim()
  const storage = safeSessionStorage()

  if (!storage) {
    return normalized
  }

  if (normalized) {
    storage.setItem(INVITATION_LINK_STORAGE_KEY, normalized)
  } else {
    storage.removeItem(INVITATION_LINK_STORAGE_KEY)
  }

  return normalized
}

export function clearPersistedInvitationCode(): void {
  persistInvitationCode('')
}

export function resolveInvitationCode(
  query: LocationQuery | URLSearchParams | Record<string, unknown> | undefined
): string {
  const queryCode = extractInvitationCodeFromQuery(query)
  if (queryCode) {
    return persistInvitationCode(queryCode)
  }

  return getPersistedInvitationCode()
}

export function buildInvitationLink(
  invitationCode: string,
  options?: {
    origin?: string
    basePath?: string
    redirect?: string
  }
): string {
  const normalizedCode = invitationCode.trim()
  if (!normalizedCode) {
    return ''
  }

  const normalizedBasePath = normalizeBasePath(
    options?.basePath || (import.meta.env.BASE_URL as string | undefined)
  )
  const pathname = `${normalizedBasePath === '/' ? '' : normalizedBasePath.slice(0, -1)}/register`
  const params = new URLSearchParams()
  params.set(INVITATION_LINK_QUERY_KEY, normalizedCode)

  const redirect = options?.redirect?.trim()
  if (redirect) {
    params.set('redirect', redirect)
  }

  const relativeURL = `${pathname}?${params.toString()}`
  const origin = options?.origin?.trim() || (typeof window !== 'undefined' ? window.location.origin : '')

  return origin ? new URL(relativeURL, origin).toString() : relativeURL
}
