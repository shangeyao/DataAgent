// OpenDataBuddy API 封装（Bearer token）+ 会话与格式化工具
const TOKEN_KEY = 'odb.token'
const WS_KEY = 'odb.workspace'

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

let token = localStorage.getItem(TOKEN_KEY) || ''
let workspaceId = localStorage.getItem(WS_KEY) || ''

export function setToken(value: string) {
  token = value || ''
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}
export function setWorkspaceId(value: string) {
  workspaceId = value || ''
  if (workspaceId) localStorage.setItem(WS_KEY, workspaceId)
  else localStorage.removeItem(WS_KEY)
}
export function getToken() { return token }
export function getWorkspaceId() { return workspaceId }
export function clearSession() {
  setToken('')
  setWorkspaceId('')
}

export async function request(path: string, opts: { method?: string; body?: unknown; form?: FormData; raw?: boolean } = {}) {
  const headers: Record<string, string> = {}
  if (token) headers.Authorization = `Bearer ${token}`
  if (workspaceId) headers['X-Workspace-Id'] = workspaceId
  let payload: BodyInit | undefined
  if (opts.form) payload = opts.form
  else if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(opts.body)
  }
  const resp = await fetch(`/api${path}`, { method: opts.method || 'GET', headers, body: payload })
  if (resp.status === 401) {
    clearSession()
    window.dispatchEvent(new CustomEvent('odb:logout'))
    throw new ApiError('会话已过期，请重新登录', 401)
  }
  if (opts.raw) return resp
  const text = await resp.text()
  let data: any = null
  try { data = text ? JSON.parse(text) : null } catch { data = { detail: text } }
  if (!resp.ok) {
    const detail = data && (data.detail || data.message) || `请求失败 (${resp.status})`
    throw new ApiError(typeof detail === 'string' ? detail : JSON.stringify(detail), resp.status)
  }
  return data
}

export const api = {
  get: (p: string) => request(p),
  post: (p: string, body?: unknown) => request(p, { method: 'POST', body }),
  patch: (p: string, body?: unknown) => request(p, { method: 'PATCH', body }),
  put: (p: string, body?: unknown) => request(p, { method: 'PUT', body }),
  del: (p: string) => request(p, { method: 'DELETE' }),
  upload: (p: string, form: FormData) => request(p, { method: 'POST', form }),
}

export const fmt = {
  num(v: unknown, digits = 0) {
    if (v === null || v === undefined || v === '') return '-'
    const n = Number(v)
    if (Number.isNaN(n)) return String(v)
    return n.toLocaleString('zh-CN', { minimumFractionDigits: digits, maximumFractionDigits: digits })
  },
  money(v: unknown) {
    if (v === null || v === undefined || v === '') return '-'
    const n = Number(v)
    if (Number.isNaN(n)) return String(v)
    if (Math.abs(n) >= 1e8) return `${(n / 1e8).toFixed(2)} 亿`
    if (Math.abs(n) >= 1e4) return `${(n / 1e4).toFixed(2)} 万`
    return n.toLocaleString('zh-CN', { maximumFractionDigits: 2 })
  },
  pct(v: unknown) { if (v === null || v === undefined) return '-'; return `${(Number(v) * 100).toFixed(1)}%` },
  dt(s?: string) { if (!s) return '-'; return String(s).replace('T', ' ').slice(0, 19) },
  ago(s?: string) {
    if (!s) return '-'
    const diff = (Date.now() - new Date(s).getTime()) / 1000
    if (diff < 60) return '刚刚'
    if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
    if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
    if (diff < 86400 * 30) return `${Math.floor(diff / 86400)} 天前`
    return String(s).slice(0, 10)
  },
  bytes(n?: number) {
    if (!n) return '0 B'
    if (n < 1024) return `${n} B`
    if (n < 1024 ** 2) return `${(n / 1024).toFixed(1)} KB`
    return `${(n / 1024 ** 2).toFixed(1)} MB`
  },
}
