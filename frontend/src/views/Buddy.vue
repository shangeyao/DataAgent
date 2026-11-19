<script setup lang="ts">
import { computed, h, nextTick, onMounted, ref } from 'vue'
import {
  NAlert, NAvatar, NButton, NDataTable, NDropdown, NFlex, NIcon, NInput, NModal,
  NScrollbar, NSelect, NSpin, NTag, NText, NTooltip,
} from 'naive-ui'
import {
  SparklesOutline, FolderOutline, CloudUploadOutline, SettingsOutline,
  SendOutline, ChevronUpOutline, PeopleOutline, AnalyticsOutline, ShieldCheckmarkOutline,
  CheckmarkCircleOutline, CreateOutline, RefreshOutline,
} from '@vicons/ionicons5'
import { api, fmt } from '../api'
import { usePortalStore } from '../store/portal'
import TeamHero from '../components/TeamHero.vue'

interface Artifact { type: string; title?: string; [k: string]: any }
interface Msg { role: string; content: string; created_at?: string; meta?: any }

const portal = usePortalStore()
const experts = [
  { id: 'engineering', label: '数据工程专家', desc: '数据接入、ETL 开发、工作流编排、运维诊断', tags: ['接入', 'ETL', '编排', '诊断'], prompt: '我刚接手这个数仓项目，帮我梳理一下整体结构和待办' },
  { id: 'analysis', label: '数据分析专家', desc: '问数、异动归因、报告生成、仪表盘搭建', tags: ['问数', '归因', '报告', '看板'], prompt: '近 30 天综合良率的趋势如何？帮我找一下拐点' },
  { id: 'governance', label: '数据治理专家', desc: '数据建模、指标注册、知识库探查', tags: ['建模', '指标', '质量', '知识库'], prompt: '帮我输出一份当前数仓的数据治理健康度分析报告' },
]
const bubbles = [
  '我刚接手这个数仓项目，帮我梳理一下整体结构',
  '我想看各晶圆厂近 30 天的良率对比',
  '帮我写一个 SQL，把产出表和客户表关联',
  '帮我把这个 SQL 任务串成工作流',
]

const expert = ref('analysis')
const chatMode = ref<'agent' | 'ask'>('agent')
const model = ref('')
const models = ref<string[]>([])
const input = ref('')
const messages = ref<Msg[]>([])
const artifacts = ref<Artifact[]>([])
const convs = ref<any[]>([])
const convId = ref<string | null>(null)
const sending = ref(false)
const previewOpen = ref(false)
const previewIndex = ref(0)
const skillsOpen = ref(false)
const skillQuery = ref('')
const skills = ref<string[]>([])
const workdir = ref('工作空间根目录')
const bodyRef = ref<any>(null)

const skillCatalog = [
  { name: 'data-engineering', title: '数据工程专家', desc: '负责数仓建设全生命周期，覆盖技术方案设计、离线/实时数据接入、ETL 开发、工作流编排调度、调度运维诊断、数据源管理、项目规范管理、文件上传建表等 10 大能力。', keys: '技术方案、搭建数仓、同步数据、写 ETL SQL、编排工作流、运维诊断' },
  { name: 'data-model-designer', title: '数据模型设计', desc: '从需求出发设计数据模型：定义实体、字段、关系，生成 ER 图（Mermaid）。', keys: '建模型、ER 图、实体关系、维度建模' },
  { name: 'data-analysis', title: '数据分析专家', desc: '自然语言问数、异动归因、多维拆解、报告与仪表盘生成。', keys: '问数、查数、归因、趋势、报告、看板' },
  { name: 'data-governance', title: '数据治理专家', desc: '元数据盘点、敏感字段识别与分级、质量规则推荐、指标口径注册。', keys: '治理、敏感、分级、质量、指标' },
]
const filteredSkills = computed(() => {
  const q = skillQuery.value.trim().toLowerCase()
  if (!q) return skillCatalog
  return skillCatalog.filter((s) => (s.name + s.title + s.desc + s.keys).toLowerCase().includes(q))
})

const currentArtifact = computed(() => artifacts.value[previewIndex.value] || null)
const artifactName = (a: Artifact, i: number) => {
  const base: Record<string, string> = { query: 'query_result', table_list: 'assets', table_schema: 'schema', sensitive_scan: 'pii_scan', quality_results: 'quality', dashboard: 'dashboard', knowledge: 'knowledge' }
  const ext = a.type === 'dashboard' ? 'html' : a.type === 'query' ? 'csv' : 'md'
  return `${base[a.type] || a.type || 'artifact'}_${i + 1}.${ext}`
}

async function loadSettings() {
  try {
    const s = await api.get('/settings')
    const set = new Set<string>()
    ;(s.providers || []).forEach((p: any) => p.model && set.add(p.model))
    if (s.llm?.model) set.add(s.llm.model)
    models.value = [...set]
    model.value = s.llm?.model || ''
  } catch { /* ignore */ }
}
async function loadConvs() {
  try { const out = await api.get('/conversations'); convs.value = out.conversations || [] } catch { convs.value = [] }
}
async function openConv(id: string) {
  try {
    const out = await api.get(`/conversations/${id}`)
    convId.value = id
    messages.value = out.messages || []
    expert.value = out.conversation.mode || 'analysis'
    chatMode.value = out.conversation.chat_mode || 'agent'
    collectArtifacts()
    if (artifacts.value.length) { previewIndex.value = artifacts.value.length - 1; previewOpen.value = true }
  } catch { /* ignore */ }
}
function collectArtifacts() {
  const all: Artifact[] = []
  for (const m of messages.value) for (const a of (m.meta?.artifacts || [])) all.push(a)
  artifacts.value = all
}
function newConv() {
  convId.value = null; messages.value = []; artifacts.value = []; previewOpen.value = false; input.value = ''
}

function fillPrompt(text: string) { input.value = text }

async function send() {
  const text = input.value.trim()
  if (!text || sending.value) return
  if (!convId.value) {
    try {
      const out = await api.post('/conversations', { mode: expert.value, chat_mode: chatMode.value })
      convId.value = out.conversation.id
    } catch (e) { return }
  }
  input.value = ''
  messages.value.push({ role: 'user', content: text, created_at: new Date().toISOString() })
  sending.value = true
  await scrollBottom()
  try {
    const out = await api.post(`/conversations/${convId.value}/messages`, { content: text, mode: expert.value, chat_mode: chatMode.value, model: model.value || null })
    sending.value = false
    messages.value.push({ role: 'assistant', content: out.content, meta: out.meta, created_at: new Date().toISOString() })
    collectArtifacts()
    if (artifacts.value.length) { previewIndex.value = artifacts.value.length - 1; previewOpen.value = true }
    await loadConvs()
    await scrollBottom()
  } catch (e) {
    sending.value = false
    messages.value.push({ role: 'assistant', content: `请求失败：${(e as Error).message}`, meta: {}, created_at: new Date().toISOString() })
  }
}
async function scrollBottom() { await nextTick(); bodyRef.value?.scrollTo({ top: 999999 }) }

function renderArtifact(a: Artifact) {
  if (a.type === 'query' && a.rows) {
    const cols = (a.columns || []).map((c: string) => ({ title: c, key: c }))
    return h(NDataTable, { columns: cols, data: a.rows.slice(0, 8), size: 'small', bordered: true, pagination: false })
  }
  if ((a.type === 'table_list' || a.type === 'knowledge') && a.items) {
    return h('div', a.items.map((it: any) => h('div', { style: 'padding:8px 0;border-bottom:1px solid #eee' }, [
      h(NText, { strong: true }, { default: () => it.ref || it.title || '-' }),
      h(NText, { depth: 3, style: 'margin-left:8px;font-size:12px' }, { default: () => it.description || it.snippet || '' }),
    ])))
  }
  if (a.type === 'table_schema' && a.columns) {
    const cols = [{ title: '字段', key: 'name' }, { title: '类型', key: 'data_type' }, { title: '说明', key: 'description' }]
    return h(NDataTable, { columns: cols, data: a.columns, size: 'small', bordered: true, pagination: false })
  }
  return h(NText, { depth: 3 }, { default: () => JSON.stringify(a).slice(0, 400) })
}

onMounted(async () => {
  await Promise.all([loadSettings(), loadConvs()])
  const deep = location.hash.split('/')[2]
  if (deep) await openConv(deep)
})
</script>

<template>
  <div class="bd-shell" :class="{ 'preview-open': previewOpen }">
    <aside class="bd-side">
      <div class="bd-side-head">
        <div class="bd-logo"><span class="bd-logo-mark">B</span><span>Buddy</span></div>
      </div>
      <div class="bd-side-scroll">
        <NButton block secondary type="primary" class="bd-new" @click="newConv"><template #icon><NIcon :component="CreateOutline" /></template>新建任务</NButton>
        <div class="bd-label">专家场景</div>
        <button v-for="e in experts" :key="e.id" class="bd-expert-mini" :class="{ active: expert === e.id }" @click="expert = e.id">
          <NIcon :component="e.id === 'engineering' ? SettingsOutline : e.id === 'analysis' ? AnalyticsOutline : ShieldCheckmarkOutline" :size="15" />
          <span>{{ e.label }}</span>
        </button>
        <div class="bd-label">历史会话</div>
        <div v-if="!convs.length" class="bd-empty">暂无历史会话</div>
        <button v-for="c in convs" :key="c.id" class="bd-conv" :class="{ active: convId === c.id }" @click="openConv(c.id)">{{ c.title || '新任务' }}</button>
      </div>
    </aside>

    <section class="bd-main">
      <header class="bd-head">
        <NFlex align="center" :size="8">
          <NIcon :component="expert === 'engineering' ? SettingsOutline : expert === 'analysis' ? AnalyticsOutline : ShieldCheckmarkOutline" :size="17" />
          <NText strong>{{ experts.find(x => x.id === expert)?.label }}</NText>
        </NFlex>
        <NFlex align="center" :size="8">
          <NButton quaternary size="tiny" @click="previewOpen = !previewOpen"><template #icon><NIcon :component="CreateOutline" /></template>产物</NButton>
        </NFlex>
      </header>

      <div ref="bodyRef" class="bd-body">
        <div v-if="!messages.length && !sending" class="bd-welcome">
          <div class="bd-welcome-head">
            <h1>DataBuddy <span class="bd-star">✦</span> <span class="bd-slogan">数据工程、治理专家团随时待命</span></h1>
            <p class="bd-welcome-sub">用对话完成取数、开发、治理与分析</p>
          </div>
          <TeamHero class="bd-hero" compact />
          <div class="bd-experts">
            <button v-for="e in experts" :key="e.id" class="bd-ecard" :class="{ active: expert === e.id }" @click="expert = e.id; fillPrompt(e.prompt)">
              <div class="bd-ec-head"><NIcon :component="e.id === 'engineering' ? SettingsOutline : e.id === 'analysis' ? AnalyticsOutline : ShieldCheckmarkOutline" :size="16" /><strong>{{ e.label }}</strong></div>
              <p class="bd-ec-desc">{{ e.desc }}</p>
              <div class="bd-ec-tags"><NTag v-for="t in e.tags" :key="t" size="small" :bordered="false">{{ t }}</NTag></div>
            </button>
          </div>
          <div class="bd-bubbles">
            <button v-for="b in bubbles" :key="b" class="bd-bubble" @click="fillPrompt(b)">{{ b }} <span>›</span></button>
          </div>
          <div class="bd-foot">内容由 AI 生成，仅供参考</div>
        </div>

        <div v-else class="bd-msgs">
          <div v-for="(m, i) in messages" :key="i" class="bd-msg" :class="m.role === 'user' ? 'user' : 'bot'">
            <template v-if="m.role === 'user'">
              <div class="bd-bubble-user">{{ m.content }}</div>
              <NAvatar round size="small">{{ (portal.user?.name || 'U').slice(0, 1) }}</NAvatar>
            </template>
            <template v-else>
              <NAvatar round size="small" class="bd-bot-avatar">B</NAvatar>
              <div class="bd-msg-body">
                <div class="bd-msg-head"><NText strong>Buddy</NText><NText depth="3" class="tiny">{{ fmt.ago(m.created_at) }}</NText></div>
                <div class="bd-md">{{ m.content }}</div>
                <div v-for="(a, ai) in (m.meta?.artifacts || [])" :key="ai" class="bd-art-inline" @click="previewIndex = artifacts.indexOf(a); previewOpen = true">
                  <component :is="renderArtifact(a)" />
                </div>
              </div>
            </template>
          </div>
          <div v-if="sending" class="bd-msg bot"><NAvatar round size="small" class="bd-bot-avatar">B</NAvatar><NSpin size="small" description="正在思考并调用工具…" /></div>
        </div>
      </div>

      <div class="bd-composer">
        <div class="bd-compose-box">
          <NInput v-model:value="input" type="textarea" :autosize="{ minRows: 2, maxRows: 6 }" placeholder="提出你的需求，专家团帮你解决" :disabled="sending" @keydown.enter.exact.prevent="send" />
          <div class="bd-tools">
            <NTooltip><template #trigger><NButton quaternary circle size="small"><template #icon><NIcon :component="FolderOutline" /></template></NButton></template>工作目录：{{ workdir }}</NTooltip>
            <NTooltip><template #trigger><NButton quaternary circle size="small"><template #icon><NIcon :component="CloudUploadOutline" /></template></NButton></template>上传文件</NTooltip>
            <NTooltip><template #trigger><NButton quaternary circle size="small"><template #icon><NIcon :component="SparklesOutline" /></template></NButton></template>@ 引用</NTooltip>
            <span class="bd-sep" />
            <NButton size="tiny" :type="chatMode === 'agent' ? 'primary' : 'default'" @click="chatMode = chatMode === 'agent' ? 'ask' : 'agent'">{{ chatMode === 'agent' ? 'Agent' : 'Ask' }}</NButton>
            <NSelect v-model:value="model" :options="models.map(m => ({ label: m || '默认模型', value: m }))" size="tiny" style="width: 140px" placeholder="选择模型" />
            <NButton size="tiny" @click="skillsOpen = true"><template #icon><NIcon :component="SparklesOutline" /></template>Skills</NButton>
            <span class="spacer" />
            <NButton type="primary" circle :loading="sending" :disabled="!input.trim()" @click="send"><template #icon><NIcon :component="SendOutline" /></template></NButton>
          </div>
        </div>
        <div class="bd-compose-foot"><NText depth="3" class="tiny">Enter 发送 · Shift + Enter 换行</NText></div>
      </div>
    </section>

    <aside class="bd-preview">
      <div class="bd-preview-head">
        <NButton quaternary size="tiny" @click="previewOpen = false">‹</NButton>
        <NIcon :component="CreateOutline" :size="14" />
        <NText strong>{{ currentArtifact ? artifactName(currentArtifact, previewIndex) : '产物预览' }}</NText>
        <span class="spacer" />
        <NText depth="3">▾</NText>
      </div>
      <div class="bd-preview-body">
        <div v-if="!artifacts.length" class="bd-preview-empty">对话产生的报告、看板、表格会显示在这里</div>
        <template v-else>
          <div class="bd-file-tabs">
            <button v-for="(a, i) in artifacts" :key="i" class="bd-file-tab" :class="{ on: i === previewIndex }" @click="previewIndex = i">{{ artifactName(a, i) }}</button>
          </div>
          <div v-if="currentArtifact" class="bd-doc">
            <h3>{{ currentArtifact.title || '产物详情' }}</h3>
            <component :is="renderArtifact(currentArtifact)" />
          </div>
        </template>
      </div>
    </aside>

    <NModal v-model:show="skillsOpen" preset="card" title="Skills" style="width: min(720px, 92vw)">
      <NInput v-model:value="skillQuery" placeholder="搜索 Skills" clearable>
        <template #prefix><NIcon :component="SparklesOutline" /></template>
      </NInput>
      <div class="bd-skill-list">
        <label v-for="s in filteredSkills" :key="s.name" class="bd-skill-item">
          <input type="checkbox" v-model="skills" :value="s.name" />
          <div>
            <NText strong class="mono">{{ s.name }}</NText>
            <p>{{ s.desc }}</p>
            <em>触发关键词：{{ s.keys }}</em>
          </div>
        </label>
        <div v-if="!filteredSkills.length" class="bd-empty">没有匹配的 Skills</div>
      </div>
    </NModal>
  </div>
</template>

<style scoped>
.bd-shell { display: grid; grid-template-columns: 240px minmax(0, 1fr) 0; height: 100%; background: #fff; overflow: hidden; }
.bd-shell.preview-open { grid-template-columns: 240px minmax(0, 1fr) 420px; }
.bd-side { border-right: 1px solid var(--n-border-color, #e7e7e7); display: flex; flex-direction: column; min-height: 0; }
.bd-side-head { height: 56px; display: flex; align-items: center; padding: 0 16px; border-bottom: 1px solid var(--n-border-color, #eee); }
.bd-logo { display: flex; align-items: center; gap: 8px; font-size: 15px; font-weight: 600; }
.bd-logo-mark { width: 28px; height: 28px; border-radius: 7px; display: grid; place-items: center; color: #fff; background: linear-gradient(135deg,#00c2d1,#0089a8); font-size: 15px; }
.bd-side-scroll { flex: 1; overflow: auto; padding: 12px 10px; }
.bd-new { margin-bottom: 10px; }
.bd-label { font-size: 11px; font-weight: 500; letter-spacing: .4px; color: #a1aab5; padding: 14px 8px 6px; }
.bd-expert-mini, .bd-conv { display: flex; align-items: center; gap: 9px; width: 100%; min-height: 38px; padding: 7px 10px; margin: 2px 0; border: none; border-radius: 8px; background: transparent; cursor: pointer; font-size: 13px; line-height: 1.5; text-align: left; color: #4b5563; transition: .12s; }
.bd-conv { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: block; }
.bd-expert-mini:hover, .bd-conv:hover { background: #f3f6f9; color: #111827; }
.bd-expert-mini.active, .bd-conv.active { background: #e4f7f9; color: #007c8a; font-weight: 500; }
.bd-empty { font-size: 12px; color: #9ca3af; padding: 8px 10px; }
.bd-main { display: flex; flex-direction: column; min-width: 0; min-height: 0; }
.bd-head { height: 56px; flex: none; display: flex; align-items: center; justify-content: space-between; padding: 0 18px; border-bottom: 1px solid var(--n-border-color, #eee); }
.bd-body { flex: 1; overflow: auto; padding: 20px 26px; }
.bd-welcome { max-width: 880px; margin: 0 auto; }
.bd-welcome-head { text-align: center; margin-bottom: 6px; }
.bd-welcome-head h1 { font-family: "PingFang SC", "HarmonyOS Sans SC", "Microsoft YaHei", v-sans, system-ui, sans-serif; font-size: clamp(22px, 2.4vw, 30px); font-weight: 600; color: #111827; letter-spacing: 1px; line-height: 1.4; }
.bd-star { background: linear-gradient(135deg,#60a5fa,#34d399); -webkit-background-clip: text; background-clip: text; color: transparent; font-size: .85em; }
.bd-slogan { font-weight: 500; color: #374151; }
.bd-welcome-sub { color: #9ca3af; font-size: 13px; margin-top: 8px; }
.bd-hero { width: 100%; display: block; margin: 6px 0 8px; }
.bd-experts { display: grid; grid-template-columns: repeat(3, minmax(0,1fr)); gap: 10px; }
.bd-ecard { position: relative; text-align: left; padding: 16px 16px 14px; border: 1px solid transparent; border-radius: 14px; background: linear-gradient(160deg, #ffffff, #fbfcfe); cursor: pointer; box-shadow: 0 2px 10px -4px rgba(20,50,80,.08), 0 0 0 1px rgba(20,50,80,.045); transition: .18s; }
.bd-ecard:hover, .bd-ecard.active { border-color: transparent; background: linear-gradient(150deg, #f2fcfd, #eaf8fa); box-shadow: 0 14px 30px -12px rgba(0,162,179,.28), 0 0 0 1px rgba(0,162,179,.22); transform: translateY(-2px); }
.bd-ec-head { display: flex; align-items: center; gap: 7px; color: #007c8a; font-size: 14px; font-weight: 600; }
.bd-ec-desc { font-size: 12px; color: #6b7280; margin: 9px 0; min-height: 38px; line-height: 1.65; }
.bd-ec-tags { display: flex; gap: 5px; flex-wrap: wrap; }
.bd-bubbles { display: grid; grid-template-columns: repeat(2, minmax(0,1fr)); gap: 9px; margin-top: 16px; }
.bd-bubble { display: flex; justify-content: space-between; gap: 8px; padding: 10px 14px; border: 1px solid #eceef2; border-radius: 9px; background: #fbfcfd; color: #4b5563; font-size: 12.5px; cursor: pointer; text-align: left; transition: .15s; }
.bd-bubble:hover { border-color: #7fd3dc; color: #007c8a; background: #f2fbfc; box-shadow: 0 3px 10px rgba(0,162,179,.1); }
.bd-bubble span { color: #b6c0cb; }
.bd-foot { text-align: center; font-size: 11.5px; color: #aaa; margin-top: 16px; }
.bd-msgs { display: flex; flex-direction: column; gap: 20px; }
.bd-msg { display: flex; gap: 10px; align-items: flex-start; }
.bd-msg.user { justify-content: flex-end; }
.bd-bubble-user { background: #f0f2f5; padding: 10px 14px; border-radius: 10px; max-width: 76%; white-space: pre-wrap; font-size: 13.5px; }
.bd-bot-avatar { background: #e4f7f9; color: #00a2b3; }
.bd-msg-body { min-width: 0; flex: 1; }
.bd-msg-head { display: flex; gap: 8px; align-items: center; margin-bottom: 6px; }
.bd-md { font-size: 13.5px; line-height: 1.8; white-space: pre-wrap; }
.bd-art-inline { margin-top: 10px; border: 1px solid var(--n-border-color, #eee); border-radius: 6px; padding: 10px; cursor: pointer; max-height: 280px; overflow: auto; }
.bd-composer { flex: none; padding: 10px 22px 14px; border-top: 1px solid var(--n-border-color, #eee); }
.bd-compose-box { border: 1px solid var(--n-border-color-strong, #dcdcdc); border-radius: 12px; padding: 10px 12px 8px; }
.bd-compose-box:focus-within { border-color: #00a2b3; box-shadow: 0 0 0 3px #00a2b31a; }
.bd-tools { display: flex; align-items: center; gap: 6px; margin-top: 8px; flex-wrap: wrap; }
.bd-sep { width: 1px; height: 16px; background: #ddd; }
.bd-compose-foot { text-align: right; margin-top: 6px; }
.bd-preview { border-left: 1px solid var(--n-border-color, #e7e7e7); display: none; flex-direction: column; min-height: 0; }
.bd-shell.preview-open .bd-preview { display: flex; }
.bd-preview-head { height: 48px; display: flex; align-items: center; gap: 6px; padding: 0 12px; border-bottom: 1px solid var(--n-border-color, #eee); }
.bd-preview-body { flex: 1; overflow: auto; padding: 14px; background: #fafbfc; }
.bd-preview-empty { text-align: center; color: #aaa; font-size: 13px; padding-top: 60px; }
.bd-file-tabs { display: flex; gap: 4px; flex-wrap: wrap; margin-bottom: 12px; }
.bd-file-tab { border: 1px solid var(--n-border-color, #e7e7e7); background: #fff; border-radius: 14px; padding: 4px 10px; font-size: 11.5px; color: #666; cursor: pointer; }
.bd-file-tab.on { border-color: #00a2b3; color: #008a99; background: #f0fbfc; }
.bd-doc { background: #fff; border-radius: 8px; padding: 14px; }
.bd-doc h3 { font-size: 15px; margin-bottom: 12px; }
.bd-skill-list { max-height: 420px; overflow: auto; margin-top: 12px; }
.bd-skill-item { display: flex; gap: 10px; padding: 10px 8px; border-radius: 6px; cursor: pointer; }
.bd-skill-item:hover { background: #f5f7fa; }
.bd-skill-item p { font-size: 12px; color: #666; margin: 4px 0; line-height: 1.6; }
.bd-skill-item em { font-style: normal; font-size: 11.5px; color: #999; }
.mono { font-family: ui-monospace, Consolas, monospace; }
.tiny { font-size: 12px; }
.spacer { flex: 1; }
@media (max-width: 1200px) { .bd-shell.preview-open { grid-template-columns: 200px minmax(0,1fr) 340px; } }
@media (max-width: 900px) { .bd-shell, .bd-shell.preview-open { grid-template-columns: 0 minmax(0,1fr) 0; } .bd-side { display: none; } .bd-experts, .bd-bubbles { grid-template-columns: 1fr; } }
</style>
