import { api, setToken, clearSession } from './api'

export interface LoginResult {
  token: string
  user: { id: string; name: string; email: string; role: string }
  workspaces: { id: string; name: string; description: string; role: string }[]
}

export async function login(email: string, password: string): Promise<LoginResult> {
  const out = await api.post('/auth/login', { email, password })
  setToken(out.token)
  return out
}

export async function logout() {
  try { await api.post('/auth/logout', {}) } catch { /* ignore */ }
  clearSession()
}
