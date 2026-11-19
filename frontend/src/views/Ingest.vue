<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="数据接入" endpoint="/catalog/tables" list-key="tables" description="文件到表、离线/实时同步与数据源接入" />
</template>
