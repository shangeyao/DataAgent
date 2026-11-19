import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { useMediaQuery } from '@vueuse/core'
import theme from './theme.json'

export const useAppStore = defineStore('app', () => {
  const dark = ref(localStorage.getItem('odb.dark') === 'true')
  const collapsed = ref(false)
  const mobile = useMediaQuery('(max-width: 760px)')
  const mobileMenu = ref(false)
  const themeOverrides = computed(() => ({
    ...theme,
    common: { ...theme.common, bodyColor: dark.value ? '#101014' : '#f5f7f9' },
  }))
  watch(dark, (value) => {
    localStorage.setItem('odb.dark', String(value))
    document.documentElement.style.colorScheme = value ? 'dark' : 'light'
  }, { immediate: true })
  return { dark, collapsed, mobile, mobileMenu, themeOverrides }
})
