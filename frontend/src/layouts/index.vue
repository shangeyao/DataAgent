<script setup lang="ts">
import { computed, h, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useFullscreen } from '@vueuse/core'
import { ProLayout, useLayoutMenu } from '../components/pro-layout'
import {
  NAvatar, NBreadcrumb, NBreadcrumbItem, NButton, NDropdown, NFlex, NIcon, NMenu,
  NSelect, NSpin, NTab, NTabs, NText, type MenuOption,
} from 'naive-ui'
import {
  GridOutline, FolderOpenOutline, CloudUploadOutline, ServerOutline, CodeSlashOutline,
  GitBranchOutline, RocketOutline, PulseOutline, FlaskOutline, CubeOutline, LibraryOutline,
  TerminalOutline, BarChartOutline, AnalyticsOutline, LayersOutline, GitNetworkOutline,
  ShieldCheckmarkOutline, ShieldOutline, PeopleOutline, ListOutline, HardwareChipOutline,
  SettingsOutline, SparklesOutline, MenuOutline, SunnyOutline, MoonOutline,
  ExpandOutline, RefreshOutline, HomeOutline,
} from '@vicons/ionicons5'
import { pages } from '../router'
import { useAppStore } from '../store/app'
import { usePortalStore } from '../store/portal'
import { logout } from '../auth'
import { icon } from '../utils/presentation'

const app = useAppStore()
const portal = usePortalStore()
const route = useRoute()
const router = useRouter()
const { toggle: toggleFullscreen } = useFullscreen()
const visited = ref<{ key: string; title: string; to: string }[]>([{ key: 'overview', title: '工作空间概览', to: '/overview' }])

const currentTab = computed(() => String(route.name ?? ''))
const iconMap: Record<string, any> = {
  sparkles: SparklesOutline, grid: GridOutline, 'cloud-upload': CloudUploadOutline, server: ServerOutline,
  code: CodeSlashOutline, 'git-branch': GitBranchOutline, rocket: RocketOutline, pulse: PulseOutline,
  flask: FlaskOutline, cube: CubeOutline, library: LibraryOutline, terminal: TerminalOutline,
  'bar-chart': BarChartOutline, analytics: AnalyticsOutline, layers: LayersOutline,
  'git-network': GitNetworkOutline, 'shield-check': ShieldCheckmarkOutline, shield: ShieldOutline,
  people: PeopleOutline, list: ListOutline, 'hard-drive': HardwareChipOutline, settings: SettingsOutline,
  'folder-open': FolderOpenOutline, home: HomeOutline,
}
const menus = computed<MenuOption[]>(() => {
  const groups = ['智能体', '门户', '数据接入', '数据工程', '数据科学', '数据分析', '数据治理', '数据应用', '平台管理', '控制台']
  const items: MenuOption[] = []
  for (const g of groups) {
    const children = pages.filter((p) => p.group === g && (!p.admin || portal.canAdmin)).map((p) => ({
      key: p.path,
      label: () => h(RouterLink, { to: `/${p.path}` }, { default: () => p.title }),
      icon: icon(iconMap[p.icon] || GridOutline),
    }))
    if (!children.length) continue
    if (g === '门户' || g === '控制台') items.push(...children)
    else items.push({ key: `g-${g}`, label: g, icon: icon(LayersOutline), children })
  }
  return items
})
const { layout, activeKey } = useLayoutMenu({ mode: 'vertical', menus, accordion: true })

watch(() => route.fullPath, () => {
  const name = String(route.name ?? '')
  activeKey.value = name
  if (pages.some((p) => p.path === name)) {
    const tab = { key: name, title: String(route.meta.title ?? name), to: route.fullPath }
    const idx = visited.value.findIndex((t) => t.key === tab.key)
    if (idx < 0) visited.value.push(tab)
    else visited.value[idx] = tab
  }
  app.mobileMenu = false
}, { immediate: true })

const workspaceOptions = computed(() => portal.workspaces.map((w) => ({ label: w.name, value: w.id })))
function chooseWorkspace(id: string) { portal.selectWorkspace(id) }
function closeTab(key: string) {
  const idx = visited.value.findIndex((t) => t.key === key)
  visited.value = visited.value.filter((t) => t.key !== key)
  if (currentTab.value === key) void router.push(visited.value[Math.max(0, idx - 1)]?.to ?? '/overview')
}
function switchTab(key: string) { const tab = visited.value.find((t) => t.key === key); if (tab) void router.push(tab.to) }
async function userAction(key: string) { if (key === 'logout') { await logout(); await router.replace('/login') } }
onMounted(() => { if (!portal.user) void portal.bootstrap() })
</script>

<template>
  <ProLayout
    v-model:collapsed="app.collapsed" mode="vertical" :is-mobile="app.mobile"
    :show-logo="!app.mobile" :show-sidebar="!app.mobile" show-footer show-tabbar nav-fixed
    :nav-height="56" :tabbar-height="40" :footer-height="36" :sidebar-width="240" :sidebar-collapsed-width="64"
    content-class="portal-content"
  >
    <template #logo>
      <RouterLink to="/overview" class="brand">
        <span v-if="app.collapsed" class="brand-mark">B</span>
        <span v-else class="brand-full"><span class="brand-mark">B</span> DataBuddy</span>
      </RouterLink>
    </template>
    <template #nav-left>
      <NFlex align="center" :size="14">
        <NButton quaternary circle aria-label="展开或收起导航" @click="app.mobile ? app.mobileMenu = !app.mobileMenu : app.collapsed = !app.collapsed">
          <template #icon><NIcon :component="MenuOutline" /></template>
        </NButton>
        <NBreadcrumb v-if="!app.mobile">
          <NBreadcrumbItem>工作空间</NBreadcrumbItem>
          <NBreadcrumbItem>{{ portal.workspace?.name ?? '未选择空间' }}</NBreadcrumbItem>
          <NBreadcrumbItem>{{ route.meta.title }}</NBreadcrumbItem>
        </NBreadcrumb>
        <NText v-else strong>DataBuddy</NText>
      </NFlex>
    </template>
    <template #nav-right>
      <NFlex align="center" :size="10" :wrap="false">
        <NSelect :value="portal.workspaceId || null" :options="workspaceOptions" :style="{ width: app.mobile ? '120px' : '180px' }" placeholder="选择工作空间" @update:value="chooseWorkspace" />
        <NButton quaternary circle :aria-label="app.dark ? '浅色主题' : '深色主题'" @click="app.dark = !app.dark">
          <template #icon><NIcon :component="app.dark ? SunnyOutline : MoonOutline" /></template>
        </NButton>
        <NButton v-if="!app.mobile" quaternary circle aria-label="全屏" @click="toggleFullscreen">
          <template #icon><NIcon :component="ExpandOutline" /></template>
        </NButton>
        <NDropdown :options="[{ key: 'identity', label: `账号：${portal.user?.email ?? ''}`, disabled: true }, { key: 'logout', label: '退出登录' }]" @select="userAction">
          <NButton quaternary><NFlex align="center" :size="8"><NAvatar round size="small">{{ (portal.user?.name || 'U').slice(0, 1) }}</NAvatar><span v-if="!app.mobile">{{ portal.user?.name }}</span></NFlex></NButton>
        </NDropdown>
      </NFlex>
    </template>
    <template #sidebar>
      <NMenu v-bind="layout.verticalMenuProps" :collapsed="app.collapsed" :collapsed-width="64" :collapsed-icon-size="22" />
    </template>
    <template #tabbar>
      <NTabs type="card" :value="currentTab" class="workspace-tabs" @update:value="switchTab" @close="closeTab">
        <NTab v-for="tab in visited" :key="tab.key" :name="tab.key" :closable="tab.key !== 'overview'">{{ tab.title }}</NTab>
      </NTabs>
    </template>
    <template #footer><div class="footer">DataAgent · Powered by DataBuddy</div></template>

    <main v-if="route.meta.full" class="full-page"><RouterView v-slot="{ Component }"><component :is="Component" /></RouterView></main>
    <template v-else>
      <NFlex justify="space-between" align="center" class="page-heading">
        <div><h1>{{ route.meta.title }}</h1></div>
        <NButton quaternary circle aria-label="刷新" @click="portal.refresh"><template #icon><NIcon :component="RefreshOutline" /></template></NButton>
      </NFlex>
      <NSpin v-if="portal.loading" description="加载中…" />
      <RouterView v-else v-slot="{ Component }"><component :is="Component" :key="route.fullPath" /></RouterView>
    </template>
  </ProLayout>
</template>
