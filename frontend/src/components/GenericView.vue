<script setup lang="ts">
// 通用数据页：根据配置的 API 路径拉取列表，渲染成表格 + 统计条。
// 用于把大量管理类页面快速接入真实数据，同时保留统一视觉。
import { computed, onMounted, ref } from 'vue'
import { NCard, NDataTable, NFlex, NStatistic, NSpin, NEmpty, NTag, type DataTableColumns } from 'naive-ui'
import { api, fmt } from '../api'

const props = defineProps<{
  title: string
  endpoint: string
  listKey?: string
  description?: string
}>()

const rows = ref<any[]>([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const out = await api.get(props.endpoint)
    const key = props.listKey || Object.keys(out || {}).find((k) => Array.isArray(out[k])) || ''
    rows.value = key ? (out[key] || []) : (Array.isArray(out) ? out : [])
  } catch (e) {
    error.value = (e as Error).message
    rows.value = []
  } finally {
    loading.value = false
  }
}
onMounted(load)
defineExpose({ load })

const columns = computed<DataTableColumns<any>>(() => {
  if (!rows.value.length) return []
  const sample = rows.value[0]
  const exclude = new Set(['meta', 'chart', 'columns', 'rows', 'config', 'yaml', 'sql', 'definition'])
  return Object.keys(sample)
    .filter((k) => !exclude.has(k) && typeof sample[k] !== 'object')
    .slice(0, 8)
    .map((k) => ({
      title: k.replace(/_/g, ' '),
      key: k,
      ellipsis: { tooltip: true },
      width: k.length > 16 ? 220 : undefined,
      render: (row: any) => {
        const v = row[k]
        if (v === null || v === undefined || v === '') return '-'
        if (k.includes('rate') || k.includes('pct')) return fmt.pct(v)
        if (typeof v === 'number' && k.includes('row')) return fmt.num(v)
        return String(v)
      },
    }))
})
</script>

<template>
  <div>
    <NFlex justify="space-between" align="center" class="section-gap">
      <div>
        <h2>{{ title }}</h2>
        <NText v-if="description" depth="3">{{ description }}</NText>
      </div>
      <NTag :bordered="false" type="info">{{ rows.length }} 条</NTag>
    </NFlex>
    <NSpin v-if="loading" description="加载中…" />
    <NCard v-else-if="error" :bordered="false"><NEmpty description="加载失败">{{ error }}</NEmpty></NCard>
    <NCard v-else :bordered="false">
      <NDataTable v-if="rows.length && columns.length" :columns="columns" :data="rows" :pagination="{ pageSize: 12 }" size="small" bordered />
      <NEmpty v-else description="暂无数据" />
    </NCard>
  </div>
</template>
