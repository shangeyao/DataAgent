<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NAlert, NButton, NCard, NFlex, NFormItem, NInput, NResult, NSpin, NText } from 'naive-ui'
import { getToken } from '../api'
import { login } from '../auth'
import { usePortalStore } from '../store/portal'

const portal = usePortalStore()
const router = useRouter()
const initializing = ref(true)
const signingIn = ref(false)
const email = ref('admin@local')
const password = ref('admin123')
const loginError = ref('')
const authenticated = computed(() => !!getToken() && !!portal.user)

onMounted(async () => {
  await bootstrap()
})

async function bootstrap() {
  initializing.value = true
  if (!getToken()) {
    initializing.value = false
    if (router.currentRoute.value.path !== '/login') await router.replace('/login')
    return
  }
  try {
    await portal.bootstrap()
    if (router.currentRoute.value.path === '/login') await router.replace('/overview')
  } catch {
    if (router.currentRoute.value.path !== '/login') await router.replace('/login')
  } finally {
    initializing.value = false
  }
}

async function signIn() {
  if (signingIn.value) return
  signingIn.value = true
  loginError.value = ''
  try {
    const out = await login(email.value.trim(), password.value)
    portal.applyBootstrap(out)
    password.value = ''
    if (router.currentRoute.value.path === '/login') await router.replace('/overview')
  } catch (error) {
    password.value = ''
    loginError.value = (error as Error).message
  } finally {
    signingIn.value = false
  }
}
</script>

<template>
  <div v-if="initializing" class="startup"><NSpin size="large" description="正在加载工作空间…" /></div>
  <RouterView v-else-if="authenticated" />
  <div v-else class="startup login-page">
    <NCard class="login-card" :bordered="false">
      <NFlex vertical :size="24">
        <NFlex align="center" :size="12">
          <span class="brand-mark-big">B</span>
          <div>
            <h1>大数据智能体工作台</h1>
            <NText depth="3">OpenDataBuddy · 开源版</NText>
          </div>
        </NFlex>
        <form class="login-form" @submit.prevent="signIn">
          <NFormItem label="邮箱 / 账号">
            <NInput v-model:value="email" placeholder="admin@local" size="large" :disabled="signingIn" />
          </NFormItem>
          <NFormItem label="密码">
            <NInput v-model:value="password" type="password" show-password-on="click" placeholder="admin123" size="large" :disabled="signingIn" />
          </NFormItem>
          <NAlert v-if="loginError" type="error" class="section-gap">{{ loginError }}</NAlert>
          <NButton type="primary" size="large" block attr-type="submit" :loading="signingIn">登录</NButton>
        </form>
        <NText depth="3">首次启动默认账号 admin@local / admin123 · 数据完全落在本机</NText>
      </NFlex>
    </NCard>
  </div>
</template>
