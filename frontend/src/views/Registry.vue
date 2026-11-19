<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="模型管理" endpoint="/registry/models" list-key="models" description="模型注册、版本与血缘" />
</template>
