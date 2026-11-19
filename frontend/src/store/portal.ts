import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { api, setWorkspaceId, getWorkspaceId } from '../api'

export interface Workspace { id: string; name: string; description?: string; role?: string }
export interface User { id: string; name: string; email: string; role: string }

export const usePortalStore = defineStore('portal', () => {
  const user = ref<User | null>(null)
  const workspaces = ref<Workspace[]>([])
  const workspaceId = ref(getWorkspaceId())
  const loading = ref(false)
  const error = ref('')

  const workspace = computed(() => workspaces.value.find((w) => w.id === workspaceId.value) || workspaces.value[0] || null)
  const canAdmin = computed(() => user.value?.role === 'admin')

  function applyBootstrap(data: { user?: User; workspaces?: Workspace[] }) {
    if (data.user) user.value = data.user
    workspaces.value = data.workspaces || []
    if (!workspaceId.value && workspaces.value[0]) workspaceId.value = workspaces.value[0].id
  }

  async function bootstrap() {
    loading.value = true
    error.value = ''
    try {
      const me = await api.get('/auth/me')
      applyBootstrap(me)
    } finally {
      loading.value = false
    }
  }

  function selectWorkspace(id: string) {
    workspaceId.value = id
    setWorkspaceId(id)
  }

  async function refresh() { await bootstrap() }

  return { user, workspaces, workspaceId, workspace, loading, error, canAdmin, bootstrap, applyBootstrap, selectWorkspace, refresh }
})
