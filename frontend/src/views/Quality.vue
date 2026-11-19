<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="数据质量" endpoint="/quality/rules" list-key="rules" description="规则引擎、执行结果与异常检测" />
</template>
