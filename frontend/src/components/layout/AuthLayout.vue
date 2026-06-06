<template>
  <div class="relative flex min-h-screen items-center justify-center overflow-hidden p-4">
    <!-- 背景：炭黑 / 冷白 -->
    <div
      class="absolute inset-0 bg-gradient-to-br from-gray-100 via-gray-50 to-gray-200 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950"
    ></div>

    <!-- 冷钢装饰层 -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- 顶部中心银色径向高光 -->
      <div
        class="absolute left-1/2 top-[-10%] h-[36rem] w-[36rem] -translate-x-1/2 rounded-full bg-white/[0.06] blur-3xl dark:bg-white/[0.05]"
      ></div>
      <!-- 底部钢灰微光 -->
      <div
        class="absolute -bottom-40 left-1/2 h-80 w-[40rem] -translate-x-1/2 rounded-full bg-primary-300/[0.04] blur-3xl"
      ></div>
      <!-- 金属网格（银线） -->
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.025)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.025)_1px,transparent_1px)] bg-[size:56px_56px] [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)]"
      ></div>
    </div>

    <!-- 内容容器 -->
    <div class="relative z-10 w-full max-w-md">
      <!-- 品牌 -->
      <div class="mb-8 text-center">
        <template v-if="settingsLoaded">
          <div
            class="mb-4 inline-flex h-16 w-16 items-center justify-center overflow-hidden rounded-xl border border-border bg-metal-raised p-2 shadow-metal-edge"
          >
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <h1 class="text-metal mb-2 text-3xl font-bold tracking-tight">
            {{ siteName }}
          </h1>
          <p class="text-sm text-muted-foreground">
            {{ siteSubtitle }}
          </p>
        </template>
      </div>

      <!-- 金属卡片（gunmetal + 边缘高光 + 拉丝扫光） -->
      <div class="metal-sheen rounded-lg border border-border bg-metal-surface p-8 shadow-metal-lg metal-card">
        <slot />
      </div>

      <!-- 页脚链接 -->
      <div class="mt-6 text-center text-sm">
        <slot name="footer" />
      </div>

      <!-- 版权 -->
      <div class="mt-8 text-center text-xs text-muted-foreground/70">
        &copy; {{ currentYear }} {{ siteName }}. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'subme')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription to API Conversion Platform')
const settingsLoaded = computed(() => appStore.publicSettingsLoaded)

const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
/* 卡片顶部金属高光边 */
.metal-card {
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.07),
    0 2px 4px rgba(0, 0, 0, 0.6),
    0 24px 64px rgba(0, 0, 0, 0.5);
}
</style>
