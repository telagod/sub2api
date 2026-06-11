<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page — QUENCH 淬钢 landing -->
  <div v-else class="q-home relative flex min-h-screen flex-col overflow-hidden">
    <!-- Ambient: 车间顶灯 + 坐标纸网格 -->
    <div class="q-ambient pointer-events-none absolute inset-0"></div>

    <!-- Header -->
    <header class="q-header sticky top-0 z-30">
      <nav class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <div class="flex items-center gap-3">
          <div class="h-8 w-8 overflow-hidden rounded-lg">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </div>
          <span class="q-brand">{{ siteName }}</span>
        </div>

        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="q-navlink"
          >
            <BookOpen :size="14" />
            <span class="hidden sm:inline">{{ t('home.docs') }}</span>
          </a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="q-navlink">
            <Github :size="14" />
            <span class="hidden sm:inline">GitHub</span>
          </a>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="q-btn-metal">
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
            <ArrowRight :size="13" />
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main -->
    <main class="relative z-10 flex-1 px-6">
      <div class="mx-auto max-w-6xl">
        <!-- ═══ Hero ═══ -->
        <section class="flex flex-col items-center gap-14 pb-20 pt-20 lg:flex-row lg:gap-16 lg:pt-24">
          <!-- Left -->
          <div class="q-rise flex-1 text-center lg:text-left">
            <div class="q-eyebrow">
              <i class="q-eyebrow-dot"></i>
              {{ t('home.quench.eyebrow') }}
            </div>

            <h1 class="q-hero-title">{{ siteName }}</h1>
            <p class="q-hero-sub">{{ siteSubtitle }}</p>
            <p class="q-hero-desc">{{ t('home.heroDescription') }}</p>

            <div class="mt-9 flex flex-wrap items-center justify-center gap-3 lg:justify-start">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="q-btn-metal q-btn-lg"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <ArrowRight :size="15" />
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="q-btn-ghost q-btn-lg"
              >
                {{ t('home.viewDocs') }}
              </a>
            </div>

            <!-- Stats -->
            <div class="q-stats">
              <div class="q-stat">{{ t('home.quench.stats.providers') }}</div>
              <div class="q-stat-sep"></div>
              <div class="q-stat">{{ t('home.quench.stats.protocol') }}</div>
              <div class="q-stat-sep"></div>
              <div class="q-stat">{{ t('home.quench.stats.billing') }}</div>
            </div>
          </div>

          <!-- Right: Gateway Trace -->
          <div class="q-rise flex flex-1 justify-center lg:justify-end" style="animation-delay: 0.12s">
            <div class="q-trace">
              <div class="q-trace-head">
                <div class="q-trace-dots"><i></i><i></i><i></i></div>
                <span class="q-trace-title">gateway · trace</span>
                <span class="q-trace-live"><i></i>LIVE</span>
              </div>
              <div class="q-trace-body">
                <div class="q-line q-l1">
                  <span class="q-method">POST</span>
                  <span class="q-path">/v1/messages</span>
                  <span class="q-ok">200</span>
                  <span class="q-lat">1.24s</span>
                </div>
                <div class="q-line q-l2">
                  <span class="q-tree">├─</span><span class="q-step">auth</span>
                  <span class="q-val">key sk-…f8a2</span><span class="q-check">✓</span>
                </div>
                <div class="q-line q-l3">
                  <span class="q-tree">├─</span><span class="q-step">route</span>
                  <span class="q-val">claude-pro <span class="q-arrow">→</span> pool-01</span><span class="q-check">✓</span>
                </div>
                <div class="q-line q-l4">
                  <span class="q-tree">├─</span><span class="q-step">stream</span>
                  <span class="q-val">first token <b>380ms</b></span><span class="q-check">✓</span>
                </div>
                <div class="q-line q-l5">
                  <span class="q-tree">└─</span><span class="q-step">billing</span>
                  <span class="q-val q-cost">$0.0042 deducted</span><span class="q-check">✓</span>
                </div>
                <div class="q-line q-l6">
                  <span class="q-prompt">$</span><span class="q-cursor"></span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ═══ Capabilities ═══ -->
        <section class="pb-20">
          <div class="mb-10 text-center">
            <h2 class="q-sec-title">{{ t('home.quench.capabilitiesTitle') }}</h2>
            <p class="q-sec-desc">{{ t('home.quench.capabilitiesDesc') }}</p>
          </div>

          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div v-for="(cap, i) in capabilities" :key="cap.title" class="q-card q-rise" :style="{ animationDelay: `${0.05 * i}s` }">
              <div class="q-card-icon">
                <component :is="cap.icon" :size="18" />
              </div>
              <h3 class="q-card-title">{{ cap.title }}</h3>
              <p class="q-card-desc">{{ cap.desc }}</p>
            </div>
          </div>
        </section>

        <!-- ═══ Providers ═══ -->
        <section class="pb-20">
          <div class="mb-8 text-center">
            <h2 class="q-sec-title">{{ t('home.providers.title') }}</h2>
            <p class="q-sec-desc">{{ t('home.providers.description') }}</p>
          </div>
          <div class="flex flex-wrap items-center justify-center gap-3">
            <div v-for="p in providers" :key="p.name" class="q-provider" :class="{ 'opacity-50': p.soon }">
              <span class="q-provider-mark">{{ p.mark }}</span>
              <span class="q-provider-name">{{ p.name }}</span>
              <span class="q-provider-badge" :class="p.soon ? 'is-soon' : 'is-on'">
                {{ p.soon ? t('home.providers.soon') : t('home.providers.supported') }}
              </span>
            </div>
          </div>
        </section>

        <!-- ═══ CTA band ═══ -->
        <section class="pb-24">
          <div class="q-cta">
            <div>
              <h2 class="q-cta-title">{{ t('home.quench.ctaTitle') }}</h2>
              <p class="q-cta-desc">{{ t('home.quench.ctaDesc') }}</p>
            </div>
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="q-btn-metal q-btn-lg shrink-0"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <ArrowRight :size="15" />
            </router-link>
          </div>
        </section>
      </div>
    </main>

    <!-- Footer -->
    <footer class="q-footer relative z-10 px-6 py-8">
      <div
        class="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 text-center sm:flex-row sm:text-left"
      >
        <p class="q-foot-text">
          &copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}
        </p>
        <div class="flex items-center gap-5">
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="q-foot-link"
          >
            {{ t('home.docs') }}
          </a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer" class="q-foot-link">
            GitHub
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import {
  ArrowRight,
  BookOpen,
  Github,
  Network,
  Boxes,
  Wallet,
  Activity,
  ShieldCheck,
  Route,
} from 'lucide-vue-next'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'subme')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('home.heroSubtitle'))
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const githubUrl = 'https://github.com/telagod/subme'

// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))

// Capability grid: 接入 → 调度 → 计费 → 监控 → 风控 → 会话
const capabilities = computed(() => [
  { icon: Network, title: t('home.features.unifiedGateway'), desc: t('home.features.unifiedGatewayDesc') },
  { icon: Boxes, title: t('home.features.multiAccount'), desc: t('home.features.multiAccountDesc') },
  { icon: Wallet, title: t('home.features.balanceQuota'), desc: t('home.features.balanceQuotaDesc') },
  { icon: Activity, title: t('home.quench.features.realtimeUsage'), desc: t('home.quench.features.realtimeUsageDesc') },
  { icon: ShieldCheck, title: t('home.quench.features.riskControl'), desc: t('home.quench.features.riskControlDesc') },
  { icon: Route, title: t('home.quench.features.stickySession'), desc: t('home.quench.features.stickySessionDesc') },
])

const providers = computed(() => [
  { mark: 'C', name: t('home.providers.claude'), soon: false },
  { mark: 'G', name: 'GPT', soon: false },
  { mark: 'G', name: t('home.providers.gemini'), soon: false },
  { mark: 'A', name: t('home.providers.antigravity'), soon: false },
  { mark: '+', name: t('home.providers.more'), soon: true },
])

// Current year for footer
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
/* ════ 基底与氛围 ════ */
.q-home {
  background: var(--bg-0, #08090b);
  color: var(--ink-0, #e8ebf0);
}
.q-ambient {
  background:
    radial-gradient(880px 460px at 70% -140px, rgba(92, 168, 255, 0.06), transparent 65%),
    radial-gradient(700px 420px at 8% 108%, rgba(92, 168, 255, 0.03), transparent 60%),
    linear-gradient(rgba(232, 235, 240, 0.016) 1px, transparent 1px),
    linear-gradient(90deg, rgba(232, 235, 240, 0.016) 1px, transparent 1px);
  background-size: auto, auto, 48px 48px, 48px 48px;
}

/* ════ 入场编排 ════ */
.q-rise {
  opacity: 0;
  transform: translateY(10px);
  animation: q-rise 0.55s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}
@keyframes q-rise {
  to { opacity: 1; transform: none; }
}

/* ════ Header ════ */
.q-header {
  background: rgba(10, 11, 14, 0.72);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line-0, #20242c);
}
.q-brand {
  font-weight: 700;
  font-size: 15px;
  letter-spacing: 0.02em;
}
.q-navlink {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 12.5px;
  font-weight: 500;
  color: var(--ink-1, #97a0af);
  transition: all 0.15s;
}
.q-navlink:hover {
  color: var(--ink-0, #e8ebf0);
  background: var(--bg-2, #171a20);
}

/* ════ 按钮：金属凸面 / ghost ════ */
.q-btn-metal {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 7px 15px;
  border-radius: 10px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink-0, #e8ebf0);
  background: var(--metal-raised, linear-gradient(180deg, #272d37, #14171d));
  border: 1px solid #3a4250;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06), 0 2px 10px rgba(0, 0, 0, 0.4);
  transition: all 0.18s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.q-btn-metal:hover {
  border-color: rgba(92, 168, 255, 0.55);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 0 16px rgba(92, 168, 255, 0.22),
    0 2px 10px rgba(0, 0, 0, 0.4);
}
.q-btn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 7px 15px;
  border-radius: 10px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink-1, #97a0af);
  border: 1px solid var(--line-1, #2f3540);
  transition: all 0.15s;
}
.q-btn-ghost:hover {
  color: var(--ink-0, #e8ebf0);
  background: var(--bg-2, #171a20);
}
.q-btn-lg {
  padding: 10px 20px;
  font-size: 13.5px;
  border-radius: 11px;
}

/* ════ Hero ════ */
.q-eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  border-radius: 99px;
  border: 1px solid rgba(92, 168, 255, 0.25);
  background: rgba(92, 168, 255, 0.08);
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.22em;
  color: #8cc4ff;
}
.q-eyebrow-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--azure, #5ca8ff);
  animation: q-pulse 2s infinite;
}
@keyframes q-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(92, 168, 255, 0.5); }
  50% { box-shadow: 0 0 0 6px rgba(92, 168, 255, 0); }
}
.q-hero-title {
  margin-top: 22px;
  font-size: clamp(40px, 6vw, 64px);
  font-weight: 800;
  font-stretch: 110%;
  letter-spacing: 0.01em;
  line-height: 1.05;
}
.q-hero-sub {
  margin-top: 18px;
  font-size: clamp(18px, 2.2vw, 23px);
  font-weight: 600;
  color: rgba(232, 235, 240, 0.92);
}
.q-hero-desc {
  margin-top: 10px;
  max-width: 480px;
  font-size: 14px;
  line-height: 1.7;
  color: var(--ink-1, #97a0af);
}
@media (max-width: 1023px) {
  .q-hero-desc { margin-inline: auto; }
}

.q-stats {
  margin-top: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}
@media (min-width: 1024px) {
  .q-stats { justify-content: flex-start; }
}
.q-stat {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 12px;
  color: var(--ink-1, #97a0af);
}
.q-stat-sep {
  width: 1px;
  height: 14px;
  background: var(--line-1, #2f3540);
}

/* ════ Gateway Trace ════ */
.q-trace {
  width: 440px;
  max-width: 100%;
  border-radius: 14px;
  background: linear-gradient(170deg, #14171d, #0c0e12 75%);
  border: 1px solid var(--line-1, #2f3540);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.07),
    0 30px 70px rgba(0, 0, 0, 0.55),
    0 0 50px rgba(92, 168, 255, 0.05);
  overflow: hidden;
}
.q-trace-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 16px;
  border-bottom: 1px solid var(--line-0, #20242c);
}
.q-trace-dots { display: flex; gap: 6px; }
.q-trace-dots i {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--bg-3, #1f232b);
  border: 1px solid var(--line-1, #2f3540);
}
.q-trace-title {
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 11px;
  color: var(--ink-2, #5c6470);
}
.q-trace-live {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: #8cc4ff;
}
.q-trace-live i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--azure, #5ca8ff);
  animation: q-pulse 1.8s infinite;
}
.q-trace-body {
  padding: 18px 20px;
  font-family: 'IBM Plex Mono', ui-monospace, monospace;
  font-size: 13px;
  line-height: 2.1;
}
.q-line {
  display: flex;
  align-items: center;
  gap: 9px;
  opacity: 0;
  animation: q-rise 0.45s cubic-bezier(0.2, 0.8, 0.2, 1) forwards;
}
.q-l1 { animation-delay: 0.35s; }
.q-l2 { animation-delay: 0.8s; }
.q-l3 { animation-delay: 1.15s; }
.q-l4 { animation-delay: 1.5s; }
.q-l5 { animation-delay: 1.85s; }
.q-l6 { animation-delay: 2.3s; }
.q-method { color: #8cc4ff; font-weight: 600; }
.q-path { color: var(--ink-0, #e8ebf0); }
.q-ok {
  margin-left: auto;
  padding: 1px 8px;
  border-radius: 5px;
  font-size: 11px;
  font-weight: 600;
  color: var(--ok, #46c98c);
  background: rgba(70, 201, 140, 0.12);
}
.q-lat { font-size: 11px; color: var(--ink-2, #5c6470); }
.q-tree { color: var(--ink-2, #5c6470); }
.q-step { color: var(--ink-1, #97a0af); width: 56px; }
.q-val { color: var(--ink-0, #e8ebf0); font-size: 12px; }
.q-val b { color: #8cc4ff; font-weight: 600; }
.q-arrow { color: var(--azure, #5ca8ff); }
.q-cost {
  color: #f2f5fa;
  text-shadow: 0 0 16px rgba(214, 232, 255, 0.25);
}
.q-check { margin-left: auto; color: var(--ok, #46c98c); font-size: 12px; }
.q-prompt { color: var(--azure, #5ca8ff); font-weight: 700; }
.q-cursor {
  display: inline-block;
  width: 8px;
  height: 15px;
  background: var(--azure, #5ca8ff);
  animation: q-blink 1.1s step-end infinite;
}
@keyframes q-blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* ════ Section 标题 ════ */
.q-sec-title {
  font-size: 26px;
  font-weight: 700;
  font-stretch: 106%;
}
.q-sec-desc {
  margin-top: 8px;
  font-size: 13px;
  color: var(--ink-1, #97a0af);
}

/* ════ 能力卡：锻面 ════ */
.q-card {
  padding: 22px 22px 20px;
  border-radius: 12px;
  background: var(--metal, linear-gradient(180deg, #15181e, #0e1014));
  border: 1px solid var(--line-0, #20242c);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04), 0 10px 28px rgba(0, 0, 0, 0.35);
  transition: border-color 0.18s, transform 0.18s cubic-bezier(0.2, 0.8, 0.2, 1), box-shadow 0.2s;
}
.q-card:hover {
  border-color: rgba(92, 168, 255, 0.35);
  transform: translateY(-3px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    0 14px 36px rgba(0, 0, 0, 0.45),
    0 0 20px rgba(92, 168, 255, 0.08);
}
.q-card-icon {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border-radius: 9px;
  color: #8cc4ff;
  background: rgba(92, 168, 255, 0.08);
  border: 1px solid rgba(92, 168, 255, 0.18);
}
.q-card-title {
  margin-top: 16px;
  font-size: 14.5px;
  font-weight: 650;
}
.q-card-desc {
  margin-top: 7px;
  font-size: 12.5px;
  line-height: 1.7;
  color: var(--ink-1, #97a0af);
}

/* ════ Providers ════ */
.q-provider {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  padding: 9px 14px;
  border-radius: 10px;
  background: var(--bg-1, #101216);
  border: 1px solid var(--line-0, #20242c);
}
.q-provider-mark {
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 800;
  color: #d7dee8;
  background: linear-gradient(180deg, #2a303a, #1b1f26);
  border: 1px solid #3a4250;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.07);
}
.q-provider-name {
  font-size: 13px;
  font-weight: 500;
}
.q-provider-badge {
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 10px;
  font-weight: 600;
}
.q-provider-badge.is-on {
  color: var(--ok, #46c98c);
  background: rgba(70, 201, 140, 0.12);
}
.q-provider-badge.is-soon {
  color: var(--ink-2, #5c6470);
  background: var(--bg-3, #1f232b);
}

/* ════ CTA band ════ */
.q-cta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  flex-wrap: wrap;
  padding: 34px 38px;
  border-radius: 16px;
  background: linear-gradient(170deg, #15181e, #0c0e12);
  border: 1px solid var(--line-1, #2f3540);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 20px 50px rgba(0, 0, 0, 0.4),
    0 0 40px rgba(92, 168, 255, 0.05);
}
.q-cta-title {
  font-size: 21px;
  font-weight: 700;
}
.q-cta-desc {
  margin-top: 6px;
  font-size: 13px;
  color: var(--ink-1, #97a0af);
}

/* ════ Footer ════ */
.q-footer {
  border-top: 1px solid var(--line-0, #20242c);
}
.q-foot-text,
.q-foot-link {
  font-size: 12.5px;
  color: var(--ink-2, #5c6470);
}
.q-foot-link {
  transition: color 0.15s;
}
.q-foot-link:hover {
  color: var(--ink-0, #e8ebf0);
}

/* ════ 动效降级 ════ */
@media (prefers-reduced-motion: reduce) {
  .q-rise, .q-line { opacity: 1; transform: none; animation: none; }
  .q-cursor, .q-eyebrow-dot, .q-trace-live i { animation: none; }
}
</style>
