import { createRouter, createWebHistory } from 'vue-router'

// 页面清单：group 用于侧栏菜单分组，icon 用 ionicons 名称（在 layouts 里映射）
export const pages = [
  { path: 'buddy', title: 'Buddy', group: '智能体', icon: 'sparkles', full: true },
  { path: 'overview', title: '工作空间概览', group: '门户', icon: 'grid' },
  { path: 'ingest', title: '数据接入', group: '数据接入', icon: 'cloud-upload' },
  { path: 'sources', title: '数据源管理', group: '数据接入', icon: 'server' },
  { path: 'studio', title: '开发工作台', group: '数据工程', icon: 'code' },
  { path: 'flows', title: '工作流编排', group: '数据工程', icon: 'git-branch' },
  { path: 'cicd', title: '发布管理', group: '数据工程', icon: 'rocket' },
  { path: 'ops', title: '运维监控', group: '数据工程', icon: 'pulse' },
  { path: 'exp', title: '模型实验', group: '数据科学', icon: 'flask' },
  { path: 'features', title: '特征管理', group: '数据科学', icon: 'cube' },
  { path: 'registry', title: '模型管理', group: '数据科学', icon: 'library' },
  { path: 'serving', title: '模型服务', group: '数据科学', icon: 'cube' },
  { path: 'sql', title: '数据探索', group: '数据分析', icon: 'terminal' },
  { path: 'dashboards', title: '仪表盘', group: '数据分析', icon: 'bar-chart' },
  { path: 'analysis', title: '分析空间', group: '数据分析', icon: 'analytics' },
  { path: 'semantic', title: '统一语义层', group: '数据治理', icon: 'layers' },
  { path: 'ontology', title: '本体建模', group: '数据治理', icon: 'git-network' },
  { path: 'quality', title: '数据质量', group: '数据治理', icon: 'shield-check' },
  { path: 'catalog', title: '数据目录', group: '数据治理', icon: 'folder-open' },
  { path: 'security', title: 'Agent 安全防护', group: '数据治理', icon: 'shield' },
  { path: 'agents', title: '智能体', group: '数据应用', icon: 'people' },
  { path: 'mcp', title: 'MCP 服务', group: '数据应用', icon: 'cube' },
  { path: 'modellist', title: '模型列表', group: '数据应用', icon: 'list' },
  { path: 'compute', title: '计算资源', group: '平台管理', icon: 'hard-drive' },
  { path: 'members', title: '成员与权限', group: '平台管理', icon: 'people' },
  { path: 'settings', title: '工作空间设置', group: '平台管理', icon: 'settings' },
  { path: 'workspaces', title: '工作空间', group: '控制台', icon: 'grid', admin: true },
  { path: 'console-users', title: '用户管理', group: '控制台', icon: 'people', admin: true },
  { path: 'platform', title: '平台设置', group: '控制台', icon: 'settings', admin: true },
]

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('../views/Login.vue'), meta: { title: '登录', public: true } },
    { path: '/', component: () => import('../layouts/index.vue'), children: [
      { path: '', redirect: '/overview' },
      ...pages.map((page) => ({
        path: page.path,
        name: page.path,
        component: () => import(`../views/${viewFile(page.path)}.vue`),
        meta: { title: page.title, group: page.group, full: page.full, admin: page.admin },
      })),
    ] },
    { path: '/:pathMatch(.*)*', component: () => import('../views/NotFound.vue'), meta: { title: '页面不存在' } },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

// 路由 -> 组件文件名映射（多个路由共用一个组件时在这里指定）
function viewFile(path: string): string {
  const map: Record<string, string> = {
    sources: 'Sources', exp: 'Experiments', features: 'Features', registry: 'Registry', serving: 'Serving',
    modellist: 'ModelList', semantic: 'Semantic', security: 'Security', 'console-users': 'ConsoleUsers',
    platform: 'Platform', workspaces: 'Workspaces',
  }
  return map[path] || path.charAt(0).toUpperCase() + path.slice(1)
}

router.afterEach((to) => { document.title = `${String(to.meta.title ?? '工作台')} · DataAgent` })
export default router
