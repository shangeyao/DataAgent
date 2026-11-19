<script setup lang="ts">
import GenericView from '../components/GenericView.vue'
</script>

<template>
  <GenericView title="用户管理" endpoint="/users" list-key="users" description="控制台用户与全局角色" />
</template>
