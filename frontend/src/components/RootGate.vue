<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { NAlert, NButton, NFormItem, NInput, NSpin } from 'naive-ui'
import { getToken } from '../api'
import { login } from '../auth'
import { usePortalStore } from '../store/portal'
import TeamHero from './TeamHero.vue'

const portal = usePortalStore()
const router = useRouter()
const initializing = ref(true)
const signingIn = ref(false)
const email = ref('admin@local')
const password = ref('admin123')
const loginError = ref('')
const authenticated = () => !!getToken() && !!portal.user

onMounted(() => { void bootstrap() })

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
  <div v-if="initializing" class="init-screen"><NSpin size="large" description="正在加载工作空间…" /></div>
  <RouterView v-else-if="authenticated()" />

  <div v-else class="auth-shell">
    <!-- 左侧品牌展示区 -->
    <section class="auth-brand">
      <div class="auth-brand-glow glow-a" />
      <div class="auth-brand-glow glow-b" />
      <div class="auth-brand-dots" />
      <div class="auth-brand-inner">
        <div class="auth-logo"><span class="auth-logo-mark">B</span></div>
        <h1 class="auth-title">
          数据工程、治理专家团<span class="auth-star">✦</span><br />随时待命
        </h1>
        <p class="auth-lead">用对话完成取数、开发、治理与分析</p>
        <TeamHero class="auth-hero" />
        <div class="auth-features">
          <span><i>✦</i> 对话即开发</span>
          <span><i>✦</i> 语义层统一口径</span>
          <span><i>✦</i> 数据治理内建</span>
        </div>
      </div>
    </section>

    <!-- 右侧登录面板 -->
    <section class="auth-panel">
      <div class="auth-card">
        <div class="auth-card-head">
          <h2>欢迎回来</h2>
          <p>登录你的数据智能体工作台</p>
        </div>
        <form class="auth-form" @submit.prevent="signIn">
          <NFormItem label="邮箱 / 账号">
            <NInput v-model:value="email" placeholder="admin@local" size="large" :disabled="signingIn" />
          </NFormItem>
          <NFormItem label="密码">
            <NInput v-model:value="password" type="password" show-password-on="click" placeholder="admin123" size="large" :disabled="signingIn" />
          </NFormItem>
          <NAlert v-if="loginError" type="error" class="auth-error">{{ loginError }}</NAlert>
          <NButton type="primary" size="large" block attr-type="submit" :loading="signingIn" class="auth-submit">登 录</NButton>
        </form>
        <div class="auth-hint">
          <span>首次启动默认账号 <code>admin@local</code> / <code>admin123</code></span>
          <span>MIT License · 数据完全落在本机</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.init-screen { min-height: 100vh; display: grid; place-items: center; }

.auth-shell { min-height: 100vh; display: grid; grid-template-columns: minmax(0, 1.15fr) minmax(420px, 0.85fr); background: #fff; }
.auth-brand { position: relative; overflow: hidden; background: linear-gradient(155deg, #f4f9ff 0%, #eef5ff 42%, #f7f4ff 100%); display: flex; flex-direction: column; }
.auth-brand-glow { position: absolute; border-radius: 50%; filter: blur(70px); pointer-events: none; }
.glow-a { width: 620px; height: 620px; background: radial-gradient(circle, rgba(0,162,179,.22), rgba(0,162,179,0) 70%); left: -140px; top: -160px; }
.glow-b { width: 560px; height: 560px; background: radial-gradient(circle, rgba(96,165,250,.2), rgba(96,165,250,0) 70%); right: -120px; bottom: -180px; }
.auth-brand-dots { position: absolute; inset: 0; background-image: radial-gradient(rgba(80,130,190,.14) 1px, transparent 1px); background-size: 24px 24px; mask-image: linear-gradient(to bottom, #000, transparent 85%); pointer-events: none; }
.auth-brand-inner { position: relative; z-index: 1; flex: 1; display: flex; flex-direction: column; justify-content: center; padding: 40px clamp(32px, 6vw, 88px); }
.auth-logo { display: flex; align-items: center; gap: 12px; font-size: 22px; font-weight: 700; color: #1f2937; margin-bottom: 10px; letter-spacing: .3px; }
.auth-logo-mark { width: 42px; height: 42px; border-radius: 11px; display: grid; place-items: center; color: #fff; font-size: 22px; background: linear-gradient(135deg, #00c2d1, #0089a8); box-shadow: 0 8px 20px rgba(0,162,179,.28); }
.auth-title { font-family: "PingFang SC", "HarmonyOS Sans SC", "Microsoft YaHei", v-sans, system-ui, sans-serif; font-size: clamp(26px, 3vw, 40px); font-weight: 600; color: #111827; line-height: 1.4; letter-spacing: 1px; margin: 6px 0 12px; }
.auth-star { background: linear-gradient(135deg,#60a5fa,#34d399); -webkit-background-clip: text; background-clip: text; color: transparent; font-size: .8em; vertical-align: middle; margin-left: 6px; }
.auth-lead { font-size: clamp(14px, 1.4vw, 17px); color: #6b7280; font-weight: 400; }
.auth-hero { margin-top: 6px; max-height: 330px; }
.auth-features { display: flex; gap: 28px; flex-wrap: wrap; margin-top: 14px; color: #4b5563; font-size: 14px; font-weight: 500; }
.auth-features i { color: #00a2b3; font-style: normal; margin-right: 6px; }
.auth-brand-foot { position: relative; z-index: 1; padding: 0 clamp(32px, 6vw, 88px) 26px; font-size: 12px; color: #9ca3af; }

.auth-panel { display: flex; align-items: center; justify-content: center; padding: 40px 32px; background: #fff; }
.auth-card { width: min(400px, 100%); }
.auth-card-head { margin-bottom: 30px; }
.auth-card-head h2 { font-size: 28px; font-weight: 700; color: #111827; }
.auth-card-head p { color: #6b7280; font-size: 14px; margin-top: 8px; }
.auth-form :deep(.n-form-item-label__text) { font-weight: 500; color: #374151; }
.auth-form :deep(.n-input) { --n-height: 48px; --n-border: 1px solid #e5e7eb; --n-border-hover: 1px solid #b9dfe4; --n-border-focus: 1px solid #008c8a; --n-box-shadow-focus: 0 0 0 3px rgba(0,162,179,.12); }
.auth-form :deep(.n-input__password) { color: #6b7280; }
.auth-form :deep(.n-base-selection), .auth-form :deep(.n-input__input-el) { padding-top: 12px; padding-bottom: 12px; }
.auth-form :deep(.n-form-item) { margin-bottom: 20px; }
.auth-submit { margin-top: 10px; height: 48px; font-size: 15px; font-weight: 600; letter-spacing: 2px; border-radius: 10px; background: linear-gradient(135deg, #00a2b3, #0089a8) !important; border: none; box-shadow: 0 8px 20px -6px rgba(0,162,179,.45); }
.auth-submit:hover { background: linear-gradient(135deg, #00b3c4, #0097b8) !important; }
.auth-error { margin-bottom: 12px; }
.auth-hint { display: flex; flex-direction: column; gap: 6px; margin-top: 26px; padding-top: 20px; border-top: 1px solid #eef0f3; font-size: 12px; color: #9ca3af; }
.auth-hint code { background: #f3f5f7; color: #4b5563; padding: 1px 5px; border-radius: 3px; }

@media (max-width: 960px) {
  .auth-shell { grid-template-columns: 1fr; }
  .auth-brand { display: none; }
}
</style>
