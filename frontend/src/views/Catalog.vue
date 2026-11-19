<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue'
import { NButton, NDataTable, NInput, NSpin, NTag, NText, type DataTableColumns } from 'naive-ui'
import { api, fmt } from '../api'

const tables = ref<any[]>([])
const loading = ref(true)
const keyword = ref('')
const active = ref<any>(null)
const preview = ref<{ columns: string[]; rows: any[] }>({ columns: [], rows: [] })

onMounted(async () => {
  try { const out = await api.get('/catalog/tables'); tables.value = out.tables || [] } catch { /* ignore */ }
  loading.value = false
})

const filtered = computed(() => tables.value.filter((t) =>
  !keyword.value || (t.ref || '').toLowerCase().includes(keyword.value.toLowerCase())
))

const columns: DataTableColumns<any> = [
  { title: '表名', key: 'ref', render: (r) => h('span', { class: 'mono' }, { default: () => r.ref }) },
  { title: 'Schema', key: 'schema_name' },
  { title: '行数', key: 'row_count', align: 'right', render: (r) => fmt.num(r.row_count) },
  { title: '描述', key: 'description', ellipsis: { tooltip: true }, render: (r) => r.description || '-' },
  { title: '操作', key: 'ops', width: 100, render: (r) => h(NButton, { size: 'tiny', text: true, type: 'primary', onClick: () => openTable(r) }, { default: () => '预览' }) },
]

async function openTable(t: any) {
  active.value = t
  try {
    const out = await api.get(`/catalog/preview?ref=${encodeURIComponent(t.ref)}`)
    preview.value = { columns: out.columns || [], rows: out.rows || [] }
  } catch { preview.value = { columns: [], rows: [] } }
}
</script>

<template>
  <div>
    <div class="cat-bar">
      <NInput v-model:value="keyword" placeholder="搜索表名…" clearable style="width: 260px" />
      <NTag :bordered="false" type="info">{{ filtered.length }} 张表</NTag>
    </div>
    <NSpin v-if="loading" description="加载中…" />
    <div v-else class="cat-grid">
      <div class="cat-list">
        <NDataTable :columns="columns" :data="filtered" size="small" bordered :pagination="{ pageSize: 15 }" />
      </div>
      <div v-if="active" class="cat-preview">
        <div class="cat-preview-head">
          <NText strong>{{ active.ref }}</NText>
          <NText depth="3" class="tiny">{{ fmt.num(active.row_count) }} 行</NText>
        </div>
        <NText depth="3" class="tiny">{{ active.description || '暂无描述' }}</NText>
        <div class="cat-preview-table">
          <NDataTable v-if="preview.rows.length" :columns="preview.columns.map(c=>({title:c,key:c}))" :data="preview.rows" size="small" :pagination="false" :scroll-x="600" />
          <NText v-else depth="3">暂无预览数据</NText>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.cat-bar { display: flex; align-items: center; gap: 12px; margin-bottom: 14px; }
.cat-grid { display: grid; grid-template-columns: minmax(0, 1fr) 420px; gap: 16px; }
.cat-list { min-width: 0; }
.cat-preview { background: #fff; border: 1px solid var(--n-border-color, #e7e7e7); border-radius: 6px; padding: 14px; align-self: start; }
.cat-preview-head { display: flex; justify-content: space-between; margin-bottom: 6px; }
.cat-preview-table { margin-top: 12px; max-height: 400px; overflow: auto; }
.mono { font-family: ui-monospace, Consolas, monospace; }
.tiny { font-size: 12px; }
</style>
