<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="工作流编排" endpoint="/flows" list-key="flows" description="可视化 DAG、依赖与调度配置" />
</template>
