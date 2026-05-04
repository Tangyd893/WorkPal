export const AUTH_STORAGE_KEY = 'workpal-auth'

export interface PersistedAuthState {
  token: string | null
  userId: number | null
  username: string | null
}

export const emptyAuthState: PersistedAuthState = {
  token: null,
  userId: null,
  username: null,
}

function getStorage(): Storage | null {
  try {
    return globalThis.localStorage
  } catch {
    return null
  }
}

function asNullableString(value: unknown): string | null {
  return typeof value === 'string' ? value : null
}

function asNullableNumber(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}

export function loadAuthState(): PersistedAuthState {
  const storage = getStorage()
  if (!storage) {
    return { ...emptyAuthState }
  }

  try {
    const raw = storage.getItem(AUTH_STORAGE_KEY)
    if (!raw) {
      return { ...emptyAuthState }
    }

    const parsed = JSON.parse(raw) as Partial<PersistedAuthState>
    return {
      token: asNullableString(parsed.token),
      userId: asNullableNumber(parsed.userId),
      username: asNullableString(parsed.username),
    }
  } catch {
    return { ...emptyAuthState }
  }
}

export function saveAuthState(state: PersistedAuthState): void {
  const storage = getStorage()
  if (!storage) {
    return
  }

  storage.setItem(AUTH_STORAGE_KEY, JSON.stringify(state))
}

export function clearAuthState(): void {
  const storage = getStorage()
  if (!storage) {
    return
  }

  storage.removeItem(AUTH_STORAGE_KEY)
}

export function getStoredToken(): string | null {
  return loadAuthState().token
}
