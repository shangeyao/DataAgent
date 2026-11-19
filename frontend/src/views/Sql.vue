<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { NButton, NDataTable, NSelect, NSpin, NTag, NText } from 'naive-ui'
import { api } from '../api'

const sql = ref(`SELECT
  f.fab_name,
  sum(o.lot_quantity) AS wafer_out,
  round(sum(o.revenue), 2) AS revenue
FROM demo.semi.fact_wafer_out o
JOIN demo.semi.dim_fab f ON f.fab_id = o.fab_id
GROUP BY 1
ORDER BY revenue DESC`)
const columns = ref<string[]>([])
const rows = ref<any[]>([])
const loading = ref(false)
const elapsed = ref(0)
const error = ref('')
const tables = ref<any[]>([])

onMounted(async () => {
  try { const out = await api.get('/catalog/tables'); tables.value = out.tables || [] } catch { /* ignore */ }
  run()
})

async function run() {
  loading.value = true
  error.value = ''
  try {
    const t0 = Date.now()
    const out = await api.post('/sql/run', { sql: sql.value, limit: 200 })
    columns.value = out.columns || []
    rows.value = out.rows || []
    elapsed.value = Date.now() - t0
  } catch (e) {
    error.value = (e as Error).message
    rows.value = []
  } finally {
    loading.value = false
  }
}

</script>

<template>
  <div class="sql-page">
    <div class="sql-editor">
      <div class="sql-bar">
        <NText strong>SQL 编辑器</NText>
        <NTag size="small" :bordered="false" type="info">Spark SQL 兼容</NTag>
        <span class="spacer" />
        <NText depth="3" class="tiny">{{ tables.length }} 张可用表 · 结果上限 500 行</NText>
      </div>
      <textarea v-model="sql" class="sql-input" spellcheck="false" />
      <div class="sql-actions">
        <NButton type="primary" size="small" :loading="loading" @click="run">运行</NButton>
        <NText depth="3" class="tiny">Ctrl / Cmd + Enter 运行</NText>
      </div>
    </div>
    <div class="sql-result">
      <NSpin v-if="loading" description="执行中…" />
      <NAlert v-else-if="error" type="error">{{ error }}</NAlert>
      <div v-else>
        <NText depth="3" class="tiny">{{ rows.length }} 行 · {{ elapsed }} ms</NText>
        <NDataTable v-if="rows.length" :columns="columns.map(c=>({title:c,key:c}))" :data="rows" size="small" bordered :pagination="{ pageSize: 20 }" />
        <NText v-else depth="3">点击运行查看结果</NText>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sql-page { display: grid; grid-template-columns: minmax(0, 1fr); gap: 16px; }
.sql-editor { background: #fff; border: 1px solid var(--n-border-color, #e7e7e7); border-radius: 6px; }
.sql-bar { display: flex; align-items: center; gap: 8px; padding: 10px 14px; border-bottom: 1px solid var(--n-border-color, #eee); }
.spacer { flex: 1; }
.sql-input { width: 100%; min-height: 200px; border: none; outline: none; resize: vertical; padding: 14px; font: 13px/1.6 ui-monospace, Consolas, monospace; }
.sql-actions { display: flex; align-items: center; gap: 12px; padding: 10px 14px; border-top: 1px solid var(--n-border-color, #eee); }
.sql-result { background: #fff; border: 1px solid var(--n-border-color, #e7e7e7); border-radius: 6px; padding: 14px; min-height: 200px; }
.tiny { font-size: 12px; }
</style>
