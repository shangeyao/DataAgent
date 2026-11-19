<script setup lang="ts">
import { NConfigProvider, NMessageProvider, NDialogProvider, NGlobalStyle, NLoadingBarProvider } from 'naive-ui'
import { useAppStore } from './store/app'
import RootGate from './components/RootGate.vue'
const app = useAppStore()
</script>

<template>
  <NConfigProvider :theme="app.dark ? undefined : null" :theme-overrides="app.themeOverrides as any">
    <NGlobalStyle />
    <NLoadingBarProvider>
      <NMessageProvider>
        <NDialogProvider>
          <RootGate />
        </NDialogProvider>
      </NMessageProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>
