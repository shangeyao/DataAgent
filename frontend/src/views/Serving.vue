<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="模型服务" endpoint="/serving/services" list-key="services" description="在线服务、副本与启停" />
</template>
