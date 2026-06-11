<template>
  <!-- 全屏冷黑底 -->
  <div class="auth-root">
    <!-- 顶部 azure 双径向环境光晕 -->
    <div class="auth-ambient" aria-hidden="true"></div>
    <!-- 坐标纸网格 -->
    <div class="auth-grid" aria-hidden="true"></div>

    <!-- 居中内容容器 -->
    <div class="auth-wrap">
      <!-- 品牌 Logo 区 -->
      <div class="auth-brand" v-if="settingsLoaded">
        <div class="auth-logo-ring">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="auth-logo-img" />
        </div>
        <h1 class="auth-site-name">{{ siteName }}</h1>
        <p class="auth-site-sub">登录到控制台</p>
      </div>
      <!-- 骨架占位，避免布局抖动 -->
      <div class="auth-brand" v-else aria-hidden="true">
        <div class="auth-logo-ring auth-logo-ring--ghost"></div>
        <div class="auth-ghost-line auth-ghost-line--title"></div>
        <div class="auth-ghost-line auth-ghost-line--sub"></div>
      </div>

      <!-- 锻面主卡 -->
      <div class="auth-card">
        <slot />
      </div>

      <!-- 页脚链接 -->
      <div class="auth-footer">
        <slot name="footer" />
      </div>

      <!-- 版权 -->
      <p class="auth-copy">&copy; {{ currentYear }} {{ siteName }}. All rights reserved.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'subme')
const siteLogo = computed(() =>
  sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true })
)
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
/* ── 全屏底层 ── */
.auth-root {
  position: relative;
  display: flex;
  min-height: 100vh;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 24px 16px;
  background: var(--bg-0);
}

/* ── 顶部 azure 双径向环境光晕（mockup .login-demo::before） ── */
.auth-ambient {
  pointer-events: none;
  position: absolute;
  inset: 0;
  background:
    radial-gradient(820px 460px at 50% -140px, rgba(92, 168, 255, 0.07), transparent 65%),
    radial-gradient(1200px 800px at 50% 120%, rgba(92, 168, 255, 0.03), transparent 60%);
}

/* ── 坐标纸网格（mockup body::before） ── */
.auth-grid {
  pointer-events: none;
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(232, 235, 240, 0.018) 1px, transparent 1px),
    linear-gradient(90deg, rgba(232, 235, 240, 0.018) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: radial-gradient(ellipse at center, black 20%, transparent 80%);
}

/* ── 内容容器 ── */
.auth-wrap {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
}

/* ── 品牌区 ── */
.auth-brand {
  margin-bottom: 28px;
  text-align: center;
}

/* Logo 环（mockup .login-logo） */
.auth-logo-ring {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  margin: 0 auto 16px;
  display: grid;
  place-items: center;
  overflow: hidden;
  background: linear-gradient(160deg, #e4e9f0, #8b95a5 55%, #5c6573);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.55),
    0 6px 24px rgba(140, 196, 255, 0.22);
}

.auth-logo-ring--ghost {
  background: var(--bg-2);
  box-shadow: none;
}

.auth-logo-img {
  width: 70%;
  height: 70%;
  object-fit: contain;
}

/* 站名（mockup .login-title） */
.auth-site-name {
  font-size: 19px;
  font-weight: 700;
  letter-spacing: 0.12em;
  color: var(--ink-0);
  margin-bottom: 6px;
}

/* 副标（mockup .login-sub） */
.auth-site-sub {
  font-size: 12px;
  color: var(--ink-2);
}

/* 骨架占位线 */
.auth-ghost-line {
  border-radius: 4px;
  background: var(--bg-2);
  margin: 0 auto;
  animation: auth-pulse 1.6s ease-in-out infinite;
}
.auth-ghost-line--title {
  width: 120px;
  height: 18px;
  margin-bottom: 10px;
}
.auth-ghost-line--sub {
  width: 80px;
  height: 12px;
}

@keyframes auth-pulse {
  0%, 100% { opacity: 0.5; }
  50%       { opacity: 1; }
}

/* ── 锻面主卡（mockup .login-card） ── */
.auth-card {
  position: relative;
  border-radius: 20px;
  padding: 40px 40px 32px;
  background: linear-gradient(165deg, #181c23, #0d0f13 70%);
  border: 1px solid var(--line-1);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.07),
    0 30px 80px rgba(0, 0, 0, 0.65),
    0 0 60px rgba(92, 168, 255, 0.05);
  animation: auth-card-in 0.5s cubic-bezier(0.2, 0.8, 0.2, 1) both;
}

@keyframes auth-card-in {
  from {
    opacity: 0;
    transform: translateY(16px) scale(0.985);
  }
}

/* ── 页脚 ── */
.auth-footer {
  margin-top: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--ink-2);
}

.auth-footer :deep(a),
.auth-footer :deep(button),
.auth-footer :deep(router-link) {
  color: var(--ink-1);
  text-decoration: none;
  transition: color 0.15s ease;
  background: none;
  border: none;
  cursor: pointer;
  font-size: inherit;
}

.auth-footer :deep(a:hover),
.auth-footer :deep(button:hover) {
  color: var(--azure);
}

/* ── 版权 ── */
.auth-copy {
  margin-top: 20px;
  text-align: center;
  font-size: 11px;
  color: var(--ink-2);
  opacity: 0.55;
}

/* ── a11y：减弱动效 ── */
@media (prefers-reduced-motion: reduce) {
  .auth-card {
    animation: none;
  }
  .auth-ghost-line {
    animation: none;
    opacity: 0.6;
  }
}

/* ── 移动端缩小内边距 ── */
@media (max-width: 480px) {
  .auth-card {
    padding: 28px 24px 24px;
  }
}
</style>
