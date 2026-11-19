<script setup lang="ts">
import { computed } from 'vue'
import { NScrollbar, useThemeVars } from 'naive-ui'

type ClassValue = string | string[] | undefined

const props = withDefaults(defineProps<{
  collapsed?: boolean
  mode?: string
  showLogo?: boolean
  showSidebar?: boolean
  sidebarWidth?: number
  sidebarCollapsedWidth?: number
  showNav?: boolean
  navHeight?: number
  showFooter?: boolean
  footerHeight?: number
  showTabbar?: boolean
  tabbarHeight?: number
  isMobile?: boolean
  logoClass?: ClassValue
  asideClass?: ClassValue
  headerClass?: ClassValue
  navClass?: ClassValue
  tabbarClass?: ClassValue
  contentClass?: ClassValue
  footerClass?: ClassValue
}>(), {
  collapsed: false,
  mode: 'vertical',
  showLogo: true,
  showSidebar: true,
  sidebarWidth: 224,
  sidebarCollapsedWidth: 58,
  showNav: true,
  navHeight: 50,
  showFooter: true,
  footerHeight: 32,
  showTabbar: true,
  tabbarHeight: 38,
  isMobile: false,
})

defineOptions({ inheritAttrs: false })

const themeVars = useThemeVars()
const sidebarWidthValue = computed(() => props.isMobile || !props.showSidebar
  ? '0px'
  : `${props.collapsed ? props.sidebarCollapsedWidth : props.sidebarWidth}px`)
const rootStyle = computed(() => ({
  '--pl-nav-height': `${props.navHeight}px`,
  '--pl-tabbar-height': `${props.tabbarHeight}px`,
  '--pl-footer-height': `${props.footerHeight}px`,
  '--pl-sidebar-width': sidebarWidthValue.value,
  '--pl-bg': themeVars.value.cardColor,
  '--pl-content-bg': themeVars.value.bodyColor,
  '--pl-border': themeVars.value.dividerColor,
  '--pl-text': themeVars.value.textColor2,
}))
</script>

<template>
  <div class="pl" :style="rootStyle">
    <aside v-if="showSidebar && !isMobile" class="pl-aside" :class="asideClass">
      <div v-if="showLogo" class="pl-logo" :class="logoClass"><slot name="logo" /></div>
      <div class="pl-sidebar"><slot name="sidebar" /></div>
    </aside>
    <div class="pl-main">
      <header v-if="showNav" class="pl-header" :class="headerClass">
        <div class="pl-nav" :class="navClass">
          <div class="pl-nav-left"><slot name="nav-left" /></div>
          <div class="pl-nav-center"><slot name="nav-center" /></div>
          <div class="pl-nav-right"><slot name="nav-right" /></div>
        </div>
        <div v-if="showTabbar" class="pl-tabbar" :class="tabbarClass"><slot name="tabbar" /></div>
      </header>
      <div class="pl-scroll">
        <NScrollbar style="height: 100%" content-class="pl-scroll-content">
          <main class="pl-content" :class="contentClass"><slot /></main>
          <footer v-if="showFooter" class="pl-footer" :class="footerClass"><slot name="footer" /></footer>
        </NScrollbar>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pl { position: relative; display: flex; width: 100%; height: 100%; color: var(--pl-text); background: var(--pl-bg); }
.pl-aside { width: var(--pl-sidebar-width); height: 100%; flex: 0 0 var(--pl-sidebar-width); display: flex; flex-direction: column; overflow: hidden; background: var(--pl-bg); border-right: 1px solid var(--pl-border); transition: width .3s, flex-basis .3s, background .3s; }
.pl-logo { height: var(--pl-nav-height); flex: 0 0 var(--pl-nav-height); }
.pl-sidebar { flex: 1; min-height: 0; overflow-y: auto; }
.pl-main { flex: 1; min-width: 0; height: 100%; display: flex; flex-direction: column; }
.pl-header { flex: 0 0 auto; background: var(--pl-bg); }
.pl-nav { display: flex; align-items: center; height: var(--pl-nav-height); border-bottom: 1px solid var(--pl-border); }
.pl-nav-left { height: 100%; display: flex; align-items: center; flex: 0 0 auto; }
.pl-nav-center { height: 100%; flex: 1; min-width: 0; overflow: hidden; }
.pl-nav-right { height: 100%; display: flex; align-items: center; margin-left: auto; }
.pl-tabbar { height: var(--pl-tabbar-height); display: flex; background: var(--pl-bg); border-bottom: 1px solid var(--pl-border); }
.pl-scroll { flex: 1; min-height: 0; }
.pl-scroll :deep(.pl-scroll-content) { min-height: 100%; display: flex; flex-direction: column; }
.pl-content { flex: 1; background: var(--pl-content-bg); }
.pl-footer { height: var(--pl-footer-height); flex: 0 0 var(--pl-footer-height); background: var(--pl-bg); }
</style>
