<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NAlert, NCard, NDataTable, NTab, NTabs, NTag, NText } from 'naive-ui'
import { api, fmt } from '../api'

const data = ref<any>({ stats: {}, tables: [], recent_runs: [] })
const loaded = ref(false)
const byFab = ref<any[]>([])
const byNode = ref<any[]>([])
const trend = ref<any[]>([])

async function metric(question: string) {
  try {
    const r = await api.post('/analysis/ask', { question })
    return { columns: r.columns || [], rows: r.rows || [] }
  } catch { return { columns: [], rows: [] } }
}

onMounted(async () => {
  const [overview, fab, node, tr] = await Promise.all([
    api.get('/overview'),
    metric('各晶圆厂晶圆出货量'),
    metric('各制程节点良率'),
    metric('近30天晶圆营收趋势'),
  ])
  data.value = overview
  byFab.value = fab.rows
  byNode.value = node.rows
  trend.value = tr.rows
  loaded.value = true
})

const stats = computed(() => data.value.stats || {})
const kpiRows = computed(() => [
  { kpi: '数据表', value: `${fmt.num(stats.value.tables || 0)} 张`, source: 'demo.semi' },
  { kpi: '数据行', value: `${fmt.num(stats.value.rows || 0)} 行`, source: 'fact_wafer_out' },
  { kpi: '注册指标', value: `${stats.value.metrics || 0} 个`, source: '统一语义层' },
  { kpi: '工作流', value: `${stats.value.flows || 0} 个`, source: `${stats.value.enabled_flows || 0} 个调度中` },
  { kpi: '敏感字段', value: `${stats.value.sensitive_columns || 0} 个`, source: 'L3 / L4 分级' },
])
const kpiCols = [
  { title: '核心 KPI', key: 'kpi' },
  { title: '数值', key: 'value' },
  { title: '数据来源', key: 'source' },
]
</script>

<template>
  <div class="ov-report">
    <section class="ov-doc">
      <div class="ov-bar"><NText strong>warehouse_report.md</NText><NText depth="3">▾</NText><span class="spacer" /><NTag type="success" size="small" :bordered="false">已生成</NTag></div>
      <div class="ov-doc-body">
        <h1>数据分析报告：半导体晶圆产出与良率评估</h1>
        <h2>1. 分析背景</h2>
        <p class="ov-q">原始问题：评估各晶圆厂、各制程节点的产出与良率表现，并给出后续产能与工艺优化建议。</p>
        <p><NText depth="3">数据来源：</NText><span class="ov-link">demo.semi.fact_wafer_out</span><NText depth="3">（晶圆产出事实表，含 5,200 行出货批次）</NText></p>
        <h2>2. 数据概览</h2>
        <NDataTable :columns="kpiCols" :data="kpiRows" size="small" bordered :pagination="false" />
        <h2>3. 核心发现</h2>
        <ul>
          <li>晶圆产出覆盖 <b>{{ (data.tables || []).length }}</b> 张数据表、<b>{{ fmt.num(stats.rows || 0) }}</b> 行出货记录。</li>
          <li>已注册 <b>{{ stats.metrics || 0 }}</b> 个统一指标，覆盖出货量、良率、营收、客户数与单片营收。</li>
          <li>右侧看板展示各晶圆厂出货量、各制程节点良率与近 30 天营收趋势。</li>
        </ul>
        <h2>4. 建议</h2>
        <ol>
          <li>优先关注良率偏低的制程节点，结合晶圆厂维度定位工艺短板。</li>
          <li>以注册指标为统一口径，避免同问两数。</li>
          <li>将高频问数固化为定时任务，每日自动刷新关键看板。</li>
        </ol>
      </div>
    </section>

    <section class="ov-board">
      <div class="ov-bar"><NText strong>dashboard.html</NText><NText depth="3">▾</NText></div>
      <div class="ov-board-body">
        <div v-for="(card, i) in [{t:'各晶圆厂晶圆出货量',rows:byFab},{t:'各制程节点综合良率',rows:byNode},{t:'近30天晶圆营收趋势',rows:trend}]" :key="i" class="ov-card">
          <div class="ov-card-title">{{ card.t }}</div>
          <NDataTable v-if="card.rows.length" :columns="card.rows[0] ? Object.keys(card.rows[0]).map(k=>({title:k,key:k})) : []" :data="card.rows.slice(0,6)" size="small" :pagination="false" />
          <NText v-else depth="3">暂无数据</NText>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.ov-report { display: grid; grid-template-columns: minmax(0, 1.05fr) minmax(0, 1fr); height: 100%; background: #fff; }
.ov-doc, .ov-board { display: flex; flex-direction: column; min-width: 0; min-height: 0; }
.ov-doc { border-right: 1px solid var(--n-border-color, #e7e7e7); }
.ov-bar { height: 44px; display: flex; align-items: center; gap: 8px; padding: 0 16px; border-bottom: 1px solid var(--n-border-color, #e7e7e7); flex: none; }
.spacer { flex: 1; }
.ov-doc-body, .ov-board-body { flex: 1; overflow: auto; padding: 22px 26px; }
.ov-doc-body h1 { font-size: 20px; margin-bottom: 18px; }
.ov-doc-body h2 { font-size: 15px; margin: 20px 0 10px; padding-bottom: 6px; border-bottom: 1px solid var(--n-border-color, #eee); }
.ov-doc-body p { font-size: 13px; line-height: 1.8; }
.ov-doc-body ul, .ov-doc-body ol { font-size: 13px; line-height: 1.9; padding-left: 22px; }
.ov-q { background: #f5f7fa; padding: 10px 12px; border-radius: 6px; }
.ov-link { color: var(--n-error-color, #d03050); }
.ov-board-body { background: var(--n-action-color, #fafbfc); padding: 16px; }
.ov-card { background: #fff; border: 1px solid var(--n-border-color, #e7e7e7); border-radius: 8px; padding: 14px; margin-bottom: 14px; }
.ov-card-title { font-size: 13px; font-weight: 500; margin-bottom: 10px; }
</style>
