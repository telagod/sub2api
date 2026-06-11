<template>
  <AppLayout>
    <div class="srg-root">
      <!-- Loading splash -->
      <div v-if="loading" class="srg-loading">
        <div class="srg-spinner" />
      </div>

      <template v-else>
        <div class="srg-layout">
          <!-- Left anchor nav -->
          <aside class="srg-nav">
            <div class="srg-search-wrap">
              <input
                v-model="searchQuery"
                type="search"
                class="srg-search"
                placeholder="搜索设置..."
                @input="onSearch"
              />
            </div>
            <nav class="srg-toc">
              <template v-for="[tab, sections] in visibleSectionsByTab" :key="tab">
                <div class="srg-toc-group">
                  <span class="srg-toc-tab">{{ tabLabel(tab) }}</span>
                  <a
                    v-for="section in sections"
                    :key="section.id"
                    :href="`#sr-section-${section.id}`"
                    class="srg-toc-item"
                    :class="{ active: activeSection === section.id, highlight: matchingSections.has(section.id) }"
                    @click.prevent="scrollToSection(section.id)"
                  >{{ resolveLabel(section.title) }}</a>
                </div>
              </template>
            </nav>
          </aside>

          <!-- Right scroll area -->
          <main class="srg-main" ref="mainEl" @scroll="onMainScroll">
            <!-- 未迁移配置的 legacy 入口（迁移完成后移除） -->
            <div class="srg-legacy-note">
              <span>
                部分高级配置（微信/钉钉/OIDC 接入、网关冷却与限流策略、自定义菜单、Admin API Key 等）尚未迁入新设置中心。
              </span>
              <router-link to="/admin/settings/legacy" class="srg-legacy-link">前往旧版设置 →</router-link>
            </div>
            <template v-for="[, sections] in visibleSectionsByTab" :key="sections[0]?.tab">
              <SectionRenderer
                v-for="section in sections"
                :key="section.id"
                :section="section"
                :form="form"
                :settings="savedSettings"
                class="srg-section"
                :class="{ 'srg-highlight': matchingSections.has(section.id) }"
                @update:field="onFieldUpdate"
              />
            </template>
          </main>
        </div>

        <!-- Sticky save bar -->
        <Transition name="srg-bar">
          <div v-if="dirtyCount > 0" class="srg-save-bar">
            <span class="srg-dirty-count">已修改 <b class="srg-mono">{{ dirtyCount }}</b> 项</span>
            <div class="srg-bar-acts">
              <button class="srg-btn" :disabled="saving" @click="discardChanges">放弃</button>
              <button class="srg-btn srg-btn-metal" :disabled="saving" @click="saveChanges">
                <span v-if="saving" class="srg-spinner srg-spinner-sm" />
                {{ saving ? '保存中…' : '保存更改' }}
              </button>
            </div>
          </div>
        </Transition>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api'
import type { TabId } from './types'
import { allSections, getSectionsByTab } from './registry'
import SectionRenderer from './SectionRenderer.vue'

const { t } = useI18n()
const appStore = useAppStore()

// ── state ──────────────────────────────────────────────────────────────────
const loading = ref(true)
const saving = ref(false)
const savedSettings = ref<Record<string, unknown>>({})
const form = ref<Record<string, unknown>>({})
const searchQuery = ref('')
const activeSection = ref('')
const mainEl = ref<HTMLElement | null>(null)
const matchingSections = ref<Set<string>>(new Set())

// ── load ───────────────────────────────────────────────────────────────────
async function loadSettings() {
  loading.value = true
  try {
    const data = await adminAPI.settings.getSettings() as unknown as Record<string, unknown>
    savedSettings.value = { ...data }
    form.value = { ...data }
  } catch (err) {
    appStore.showError(String(err))
  } finally {
    loading.value = false
  }
}

onMounted(loadSettings)

// ── dirty tracking ─────────────────────────────────────────────────────────
const dirtyCount = computed(() => {
  let count = 0
  for (const key of Object.keys(form.value)) {
    if (JSON.stringify(form.value[key]) !== JSON.stringify(savedSettings.value[key])) {
      count++
    }
  }
  return count
})

function onFieldUpdate(key: string, value: unknown) {
  form.value = { ...form.value, [key]: value }
}

// ── save / discard ─────────────────────────────────────────────────────────
async function saveChanges() {
  const patch: Record<string, unknown> = {}
  for (const key of Object.keys(form.value)) {
    if (JSON.stringify(form.value[key]) !== JSON.stringify(savedSettings.value[key])) {
      patch[key] = form.value[key]
    }
  }
  saving.value = true
  try {
    await adminAPI.settings.updateSettings(patch as Parameters<typeof adminAPI.settings.updateSettings>[0])
    savedSettings.value = { ...form.value }
    appStore.showSuccess(t('common.saved'))
  } catch (err) {
    appStore.showError(String(err))
  } finally {
    saving.value = false
  }
}

function discardChanges() {
  form.value = { ...savedSettings.value }
}

// ── search ─────────────────────────────────────────────────────────────────
function onSearch() {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) {
    matchingSections.value = new Set()
    return
  }
  const hits = new Set<string>()
  for (const section of allSections) {
    const titleMatch = resolveLabel(section.title).toLowerCase().includes(q)
    const fieldMatch = section.fields.some(
      (f) =>
        f.key.toLowerCase().includes(q) ||
        resolveLabel(f.label).toLowerCase().includes(q),
    )
    if (titleMatch || fieldMatch) hits.add(section.id)
  }
  matchingSections.value = hits
}

// ── visible sections (filtered by search) ──────────────────────────────────
const visibleSectionsByTab = computed<Map<TabId, typeof allSections>>(() => {
  const byTab = getSectionsByTab()
  if (matchingSections.value.size === 0) return byTab
  const filtered = new Map<TabId, typeof allSections>()
  for (const [tab, sections] of byTab) {
    const visible = sections.filter((s) => matchingSections.value.has(s.id))
    if (visible.length > 0) filtered.set(tab, visible)
  }
  return filtered
})

// ── scroll tracking (throttled via rAF) ────────────────────────────────────────
let scrollRafId: number | null = null

function onMainScroll() {
  if (scrollRafId !== null) return
  scrollRafId = requestAnimationFrame(() => {
    scrollRafId = null
    if (!mainEl.value) return
    const cards = mainEl.value.querySelectorAll<HTMLElement>('[id^="sr-section-"]')
    for (const card of cards) {
      const rect = card.getBoundingClientRect()
      if (rect.top <= 160) activeSection.value = card.id.replace('sr-section-', '')
    }
  })
}

onUnmounted(() => {
  if (scrollRafId !== null) cancelAnimationFrame(scrollRafId)
})

async function scrollToSection(id: string) {
  await nextTick()
  const el = document.getElementById(`sr-section-${id}`)
  el?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  activeSection.value = id
}

// ── label helpers ──────────────────────────────────────────────────────────
function resolveLabel(key: string): string {
  try {
    const r = t(key)
    return r === key ? key : r
  } catch { return key }
}

const TAB_LABELS: Record<TabId, string> = {
  general: '通用',
  agreement: '协议',
  features: '功能',
  security: '安全',
  users: '用户',
  gateway: '网关',
  payment: '支付',
  email: '邮件',
  backup: '备份',
}

function tabLabel(tab: string): string {
  return TAB_LABELS[tab as TabId] ?? tab
}
</script>

<style scoped>
.srg-root { padding: 24px 28px 120px; font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif); color: var(--ink-0, #E8EBF0); position: relative; }

/* layout */
.srg-layout { display: flex; gap: 24px; align-items: flex-start; }
.srg-nav { width: 200px; flex-shrink: 0; position: sticky; top: 5rem; max-height: calc(100vh - 7rem); overflow-y: auto; scrollbar-width: none; }
.srg-nav::-webkit-scrollbar { display: none; }
.srg-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 16px; }

/* search */
.srg-search-wrap { margin-bottom: 12px; }
.srg-search {
  width: 100%; padding: 6px 10px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 12.5px; font-family: inherit; outline: none;
  transition: border-color .15s; box-sizing: border-box;
}
.srg-search:focus { border-color: var(--azure, #5CA8FF); }

/* toc */
.srg-toc { display: flex; flex-direction: column; gap: 4px; }
.srg-toc-group { display: flex; flex-direction: column; gap: 1px; margin-bottom: 8px; }
.srg-toc-tab { font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: .08em; color: var(--ink-2, #5C6470); padding: 0 8px 4px; }
.srg-toc-item {
  display: block; padding: 5px 8px; border-radius: 6px;
  font-size: 12px; color: var(--ink-1, #97A0AF); text-decoration: none;
  transition: background .1s, color .1s; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.srg-toc-item:hover { background: var(--bg-2, #171A20); color: var(--ink-0, #E8EBF0); }
.srg-toc-item.active { background: var(--bg-2, #171A20); color: var(--ink-0, #E8EBF0); }
.srg-toc-item.highlight { color: var(--azure, #5CA8FF); }

/* section highlight on search */
.srg-section { transition: outline .15s; }
.srg-highlight { outline: 1px solid rgba(92,168,255,.35); }

/* loading */
.srg-loading { display: flex; align-items: center; justify-content: center; height: 60vh; }
.srg-spinner { width: 28px; height: 28px; border-radius: 50%; border: 2px solid var(--line-1, #2F3540); border-top-color: var(--azure, #5CA8FF); animation: srg-spin .7s linear infinite; }
.srg-spinner-sm { display: inline-block; width: 14px; height: 14px; border-width: 2px; vertical-align: -3px; margin-right: 6px; }
@keyframes srg-spin { to { transform: rotate(360deg); } }

/* save bar */
.srg-save-bar {
  position: fixed; bottom: 20px; left: 50%; transform: translateX(-50%);
  display: flex; align-items: center; gap: 16px;
  background: var(--bg-1, #101216); border: 1px solid var(--line-1, #2F3540);
  border-radius: 12px; padding: 10px 16px;
  box-shadow: 0 8px 32px rgba(0,0,0,.6), inset 0 1px 0 rgba(255,255,255,.06);
  z-index: 50; white-space: nowrap;
}
.srg-dirty-count { font-size: 12.5px; color: var(--ink-1, #97A0AF); }
.srg-mono { font-family: var(--font-mono, "IBM Plex Mono", monospace); color: var(--azure, #5CA8FF); }
.srg-bar-acts { display: flex; gap: 8px; }
.srg-btn {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 6px 14px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-2, #171A20);
  color: var(--ink-1, #97A0AF); font-size: 12.5px; font-weight: 500;
  cursor: pointer; font-family: inherit; transition: border-color .15s, color .15s;
}
.srg-btn:hover:not(:disabled) { border-color: #3D4554; color: var(--ink-0, #E8EBF0); }
.srg-btn:disabled { opacity: .5; cursor: not-allowed; }
.srg-btn-metal {
  background: var(--metal-raised, linear-gradient(180deg,#272D37,#14171D));
  border-color: rgba(255,255,255,.1); color: var(--ink-0, #E8EBF0);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.06);
}
.srg-btn-metal:hover:not(:disabled) { border-color: rgba(92,168,255,.4); }

/* bar transition */
.srg-bar-enter-active, .srg-bar-leave-active { transition: opacity .2s, transform .2s; }
.srg-bar-enter-from, .srg-bar-leave-to { opacity: 0; transform: translateX(-50%) translateY(8px); }

/* legacy 入口横幅（迁移完成后随入口一并移除） */
.srg-legacy-note {
  display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap;
  padding: 10px 16px; margin-bottom: 16px; border-radius: 10px;
  font-size: 12.5px; color: var(--ink-1, #97A0AF);
  background: var(--info-dim, rgba(92,168,255,.12));
  border: 1px solid rgba(92,168,255,.25);
}
.srg-legacy-link {
  flex: none; font-weight: 600; color: var(--azure, #5CA8FF); text-decoration: none;
}
.srg-legacy-link:hover { color: #8CC4FF; }
</style>
