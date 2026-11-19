<script setup lang="ts">
import { NResult, NButton } from 'naive-ui'
import { useRouter } from 'vue-router'
const router = useRouter()
</script>
<template>
  <NResult status="404" title="页面不存在" description="你访问的页面不存在或已移除">
    <template #footer><NButton type="primary" @click="router.replace('/overview')">返回工作台</NButton></template>
  </NResult>
</template>
