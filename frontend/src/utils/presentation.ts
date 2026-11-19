import { h, type Component } from 'vue'
import { NButton, NIcon } from 'naive-ui'

export const icon = (component: Component) => () => h(NIcon, null, { default: () => h(component) })
export function action(label: string, onClick: () => void, options: { disabled?: boolean; type?: 'primary' | 'error' | 'warning' } = {}) {
  return h(NButton, { text: true, size: 'small', ...options, onClick }, { default: () => label })
}
export const displayTime = (value?: string) => value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '-'

export const statusType = (status?: string): 'default' | 'info' | 'success' | 'warning' | 'error' => {
  switch ((status || '').toLowerCase()) {
    case 'success': case 'pass': case 'ok': case 'running': case 'published': return 'success'
    case 'fail': case 'failed': case 'error': case 'blocked': return 'error'
    case 'warning': case 'warn': case 'pending': case 'waiting': return 'warning'
    case 'info': return 'info'
    default: return 'default'
  }
}
