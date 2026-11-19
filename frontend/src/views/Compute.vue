<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="计算资源" endpoint="/compute-resources" list-key="types" description="四类集群、CU 配额与网络设置" />
</template>
