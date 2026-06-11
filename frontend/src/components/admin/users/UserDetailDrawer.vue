<template>
  <Teleport to="body">
    <!-- Scrim -->
    <Transition name="ud-scrim">
      <div
        v-if="open"
        class="ud-scrim"
        @click="handleClose"
        aria-hidden="true"
      />
    </Transition>

    <!-- Drawer -->
    <Transition name="ud-slide">
      <div
        v-if="open"
        ref="drawerRef"
        class="ud-drawer"
        role="dialog"
        aria-modal="true"
        :aria-label="user ? `用户详情 - ${user.email}` : '用户详情'"
        @keydown.esc="handleClose"
        tabindex="-1"
      >
        <!-- 加载态 -->
        <div v-if="loading" class="ud-loading-cover">
          <div class="ud-spinner"></div>
        </div>

        <!-- 错误态 -->
        <div v-else-if="loadError" class="ud-error-cover">
          <p>{{ loadError }}</p>
          <button class="ud-btn-ghost" @click="loadUser">重试</button>
        </div>

        <template v-else-if="user">
          <!-- ── 头部 ── -->
          <div class="ud-head">
            <div class="ud-avatar" :title="user.email">
              {{ user.email.charAt(0).toUpperCase() }}
            </div>
            <div class="ud-head-info">
              <div class="ud-email">{{ user.email }}</div>
              <div class="ud-head-meta">
                <span class="ud-reg-time">注册 {{ fmtDate(user.created_at) }}</span>
              </div>
            </div>
            <div class="ud-head-badges">
              <span class="ud-badge" :class="user.role === 'admin' ? 'ud-badge-azure' : 'ud-badge-gray'">
                {{ user.role === 'admin' ? '管理员' : '用户' }}
              </span>
              <span class="ud-badge" :class="user.status === 'active' ? 'ud-badge-ok' : 'ud-badge-bad'">
                {{ user.status === 'active' ? '活跃' : '禁用' }}
              </span>
            </div>
            <button class="ud-close" @click="handleClose" aria-label="关闭抽屉">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                <path d="M2 2L12 12M12 2L2 12" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
              </svg>
            </button>
          </div>

          <!-- ── KPI 三格 ── -->
          <div class="ud-kpi-strip">
            <div class="ud-kpi-item">
              <span class="ud-kpi-label">余额</span>
              <span class="ud-kpi-val q-money">${{ fmtBal(user.balance) }}</span>
            </div>
            <div class="ud-kpi-divider"></div>
            <div class="ud-kpi-item">
              <span class="ud-kpi-label">本月消耗</span>
              <span class="ud-kpi-val q-money" v-if="!monthStatsLoading">${{ fmtBal(monthStats.total_cost) }}</span>
              <span class="ud-kpi-val ud-muted" v-else>…</span>
            </div>
            <div class="ud-kpi-divider"></div>
            <div class="ud-kpi-item">
              <span class="ud-kpi-label">订阅状态</span>
              <span class="ud-kpi-val" v-if="activeSub">
                <span class="ud-badge ud-badge-ok">活跃</span>
              </span>
              <span class="ud-kpi-val ud-muted" v-else>无</span>
            </div>
          </div>

          <!-- ── 页签条 ── -->
          <div class="ud-tabs" role="tablist">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              class="ud-tab"
              :class="{ 'ud-tab-active': activeTab === tab.key }"
              role="tab"
              :aria-selected="activeTab === tab.key"
              @click="activeTab = tab.key"
            >{{ tab.label }}</button>
          </div>

          <!-- ── 页签内容 ── -->
          <div class="ud-body" role="tabpanel">
            <OverviewTab v-if="activeTab === 'overview'" :user="user" />
            <SubscriptionsTab v-else-if="activeTab === 'subscriptions'" :user="user" :active="activeTab === 'subscriptions'" />
            <KeysTab v-else-if="activeTab === 'keys'" :user="user" :active="activeTab === 'keys'" />
            <OrdersTab v-else-if="activeTab === 'orders'" :user="user" :active="activeTab === 'orders'" />
            <UsageTab v-else-if="activeTab === 'usage'" :user="user" :active="activeTab === 'usage'" />
            <RiskTab v-else-if="activeTab === 'risk'" :user="user" :active="activeTab === 'risk'" />
          </div>

          <!-- ── 底部操作栏 ── -->
          <div class="ud-footer">
            <button class="ud-foot-btn ud-foot-btn-primary" @click="showBalanceAdj = true">
              <svg width="13" height="13" viewBox="0 0 13 13" fill="none" aria-hidden="true">
                <path d="M6.5 2v9M2 6.5h9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
              </svg>
              调余额
            </button>
            <button
              class="ud-foot-btn"
              :class="user.status === 'active' ? 'ud-foot-btn-danger' : 'ud-foot-btn-ok'"
              :disabled="user.role === 'admin' && isCurrentAdmin"
              @click="toggleStatus"
            >
              {{ statusToggleLabel }}
            </button>
          </div>
        </template>
      </div>
    </Transition>

    <!-- 调余额弹窗 -->
    <BalanceAdjustPopover
      v-if="user"
      :open="showBalanceAdj"
      :user-id="user.id"
      :current-balance="user.balance"
      @close="showBalanceAdj = false"
      @updated="onUpdated"
    />
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { AdminUser } from '@/types'

import { useAuthStore } from '@/stores/auth'

// Tab components (lazy via v-if)
import OverviewTab from './tabs/OverviewTab.vue'
import SubscriptionsTab from './tabs/SubscriptionsTab.vue'
import KeysTab from './tabs/KeysTab.vue'
import OrdersTab from './tabs/OrdersTab.vue'
import UsageTab from './tabs/UsageTab.vue'
import RiskTab from './tabs/RiskTab.vue'
import BalanceAdjustPopover from './BalanceAdjustPopover.vue'

// ── Props / Emits ──────────────────────────────────────────────────────
const props = defineProps<{
  userId: number | null
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'updated'): void
}>()

// ── Store ──────────────────────────────────────────────────────────────
const appStore = useAppStore()
const authStore = useAuthStore()

// ── State ──────────────────────────────────────────────────────────────
const drawerRef = ref<HTMLElement | null>(null)
const user = ref<AdminUser | null>(null)
const loading = ref(false)
const loadError = ref<string | null>(null)
const activeTab = ref('overview')
const showBalanceAdj = ref(false)
const monthStatsLoading = ref(false)
const monthStats = ref({ total_cost: 0, total_requests: 0, total_tokens: 0 })

const tabs = [
  { key: 'overview', label: '概览' },
  { key: 'subscriptions', label: '订阅' },
  { key: 'keys', label: 'Keys' },
  { key: 'orders', label: '订单' },
  { key: 'usage', label: '用量' },
  { key: 'risk', label: '风控' },
]

// ── Computed ──────────────────────────────────────────────────────────
const activeSub = computed(() =>
  user.value?.subscriptions?.some(s => s.status === 'active') ?? false
)

const isCurrentAdmin = computed(() =>
  authStore.user?.id === user.value?.id
)

const statusToggleLabel = computed(() =>
  user.value?.status === 'active' ? '禁用账号' : '启用账号'
)

// ── Helpers ───────────────────────────────────────────────────────────
function fmtBal(v: number) {
  if (!v && v !== 0) return '0.00'
  const s = v.toFixed(8).replace(/\.?0+$/, '')
  const parts = s.split('.')
  if (parts.length === 1) return s + '.00'
  if (parts[1].length < 2) return s + '0'
  return s
}

function fmtDate(iso: string | null | undefined) {
  if (!iso) return '-'
  return new Date(iso).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

// ── Data loading ──────────────────────────────────────────────────────
async function loadUser() {
  if (!props.userId) return
  loading.value = true; loadError.value = null
  try {
    user.value = await adminAPI.users.getById(props.userId)
    activeTab.value = 'overview'
    loadMonthStats()
  } catch (e: any) {
    loadError.value = e?.response?.data?.detail || '加载用户失败'
  } finally { loading.value = false }
}

async function loadMonthStats() {
  if (!props.userId) return
  monthStatsLoading.value = true
  try {
    const res = await adminAPI.users.getUserUsageStats(props.userId, 'month')
    monthStats.value = res
  } catch { /* silent */ } finally { monthStatsLoading.value = false }
}

// ── Actions ────────────────────────────────────────────────────────────
async function toggleStatus() {
  if (!user.value) return
  if (user.value.role === 'admin' && isCurrentAdmin.value) {
    appStore.showWarning('无法禁用自己的管理员账号')
    return
  }
  const newStatus = user.value.status === 'active' ? 'disabled' : 'active'
  try {
    const updated = await adminAPI.users.toggleStatus(user.value.id, newStatus)
    user.value = updated
    appStore.showSuccess(newStatus === 'active' ? '账号已启用' : '账号已禁用')
    emit('updated')
  } catch (e: any) {
    appStore.showError(e?.response?.data?.detail || '操作失败')
  }
}

function onUpdated() {
  emit('updated')
  // 刷新用户数据（余额已变）
  loadUser()
}

function handleClose() {
  emit('close')
}

// ── Keyboard & focus ──────────────────────────────────────────────────
function onKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.open && !showBalanceAdj.value) {
    handleClose()
  }
}

onMounted(() => document.addEventListener('keydown', onKeyDown))
onUnmounted(() => document.removeEventListener('keydown', onKeyDown))

// ── Watchers ──────────────────────────────────────────────────────────
watch(
  () => ({ uid: props.userId, open: props.open }),
  async ({ uid, open }, prev) => {
    if (open && uid) {
      // 仅在 open 从 false→true 或 userId 变化时重新加载
      if (!prev?.open || prev?.uid !== uid) {
        await loadUser()
        await nextTick()
        drawerRef.value?.focus()
      }
    } else if (!open) {
      user.value = null
      loadError.value = null
    }
  },
  { immediate: true }
)
</script>

<style scoped>
/* ── Scrim ── */
.ud-scrim {
  position: fixed; inset: 0; z-index: 9990;
  background: rgba(0, 0, 0, 0.52);
}
.ud-scrim-enter-active, .ud-scrim-leave-active { transition: opacity 0.24s ease; }
.ud-scrim-enter-from, .ud-scrim-leave-to { opacity: 0; }

/* ── Drawer ── */
.ud-drawer {
  position: fixed; top: 0; right: 0; bottom: 0; z-index: 9991;
  width: 560px;
  background: var(--bg-1);
  border-left: 1px solid var(--line-1);
  box-shadow: -24px 0 64px rgba(0,0,0,.45), var(--edge-hi);
  display: flex; flex-direction: column;
  font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif);
  font-size: 13px;
  color: var(--ink-0);
  outline: none;
  overflow: hidden;
}

/* 滑入动画（prefers-reduced-motion 降级） */
@media (prefers-reduced-motion: no-preference) {
  .ud-slide-enter-active { transition: transform 0.24s cubic-bezier(0.22, 1, 0.36, 1); }
  .ud-slide-leave-active { transition: transform 0.2s cubic-bezier(0.55, 0, 1, 0.45); }
}
@media (prefers-reduced-motion: reduce) {
  .ud-slide-enter-active { transition: opacity 0.15s; }
  .ud-slide-leave-active { transition: opacity 0.12s; }
  .ud-slide-enter-from, .ud-slide-leave-to { opacity: 0; transform: none !important; }
}
.ud-slide-enter-from { transform: translateX(100%); }
.ud-slide-leave-to  { transform: translateX(100%); }

/* ── 头部 ── */
.ud-head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 22px 16px;
  border-bottom: 1px solid var(--line-0);
  flex-shrink: 0;
}
.ud-avatar {
  width: 40px; height: 40px; border-radius: 50%;
  background: var(--azure-dim);
  border: 1px solid rgba(92,168,255,.35);
  display: flex; align-items: center; justify-content: center;
  font-size: 17px; font-weight: 700; color: var(--azure);
  flex-shrink: 0;
}
.ud-head-info { flex: 1; min-width: 0; }
.ud-email {
  font-size: 14px; font-weight: 600; color: var(--ink-0);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.ud-head-meta { display: flex; align-items: center; gap: 8px; margin-top: 3px; }
.ud-reg-time { font-size: 11px; color: var(--ink-2); }
.ud-head-badges { display: flex; gap: 5px; flex-shrink: 0; }
.ud-close {
  background: none; border: none; cursor: pointer;
  color: var(--ink-2); padding: 5px;
  border-radius: 7px; display: flex; align-items: center;
  transition: background 0.12s;
  flex-shrink: 0;
}
.ud-close:hover { background: var(--bg-2); color: var(--ink-0); }

/* Badges */
.ud-badge {
  font-size: 10.5px; font-weight: 600; padding: 2px 8px;
  border-radius: 5px; letter-spacing: 0.04em; border: 1px solid transparent;
}
.ud-badge-azure { background: var(--azure-dim); color: var(--azure); border-color: rgba(92,168,255,.3); }
.ud-badge-gray  { background: var(--bg-2); color: var(--ink-2); border-color: var(--line-1); }
.ud-badge-ok    { background: var(--ok-dim); color: var(--ok); border-color: rgba(70,201,140,.3); }
.ud-badge-bad   { background: var(--bad-dim); color: var(--bad); border-color: rgba(242,92,105,.3); }

/* ── KPI 条 ── */
.ud-kpi-strip {
  display: flex;
  align-items: stretch;
  padding: 14px 22px;
  border-bottom: 1px solid var(--line-0);
  gap: 0;
  flex-shrink: 0;
}
.ud-kpi-item {
  flex: 1; display: flex; flex-direction: column; gap: 3px;
  padding: 0 14px;
}
.ud-kpi-item:first-child { padding-left: 0; }
.ud-kpi-item:last-child { padding-right: 0; }
.ud-kpi-label { font-size: 10.5px; color: var(--ink-2); }
.ud-kpi-val { font-size: 15px; font-weight: 700; color: var(--ink-0); }
.ud-kpi-divider { width: 1px; background: var(--line-0); margin: 2px 0; }

/* ── 页签 ── */
.ud-tabs {
  display: flex;
  padding: 0 22px;
  border-bottom: 1px solid var(--line-0);
  gap: 0;
  flex-shrink: 0;
  overflow-x: auto;
}
.ud-tabs::-webkit-scrollbar { display: none; }
.ud-tab {
  padding: 10px 14px;
  font-size: 12.5px; font-weight: 500;
  color: var(--ink-2); background: none; border: none;
  cursor: pointer; white-space: nowrap;
  border-bottom: 2px solid transparent;
  transition: color 0.15s, border-color 0.15s;
  margin-bottom: -1px;
  font-family: inherit;
}
.ud-tab:hover { color: var(--ink-0); }
.ud-tab-active { color: var(--azure); border-bottom-color: var(--azure); }

/* ── 内容区 ── */
.ud-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px 22px;
  scrollbar-width: thin;
  scrollbar-color: var(--line-1) transparent;
}
.ud-body::-webkit-scrollbar { width: 5px; }
.ud-body::-webkit-scrollbar-track { background: transparent; }
.ud-body::-webkit-scrollbar-thumb { background: var(--line-1); border-radius: 3px; }

/* ── 底部操作栏 ── */
.ud-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 22px;
  border-top: 1px solid var(--line-0);
  flex-shrink: 0;
}
.ud-foot-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: 8px; font-size: 12.5px;
  font-weight: 600; cursor: pointer; border: 1px solid transparent;
  font-family: inherit; transition: all 0.13s;
}
.ud-foot-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.ud-foot-btn-primary {
  background: var(--azure-dim); border-color: rgba(92,168,255,.4); color: var(--azure);
}
.ud-foot-btn-primary:hover {
  background: rgba(92,168,255,.22); box-shadow: 0 0 14px rgba(92,168,255,.18);
}
.ud-foot-btn-danger {
  background: var(--bad-dim); border-color: rgba(242,92,105,.4); color: var(--bad);
}
.ud-foot-btn-danger:not(:disabled):hover {
  background: rgba(242,92,105,.2);
}
.ud-foot-btn-ok {
  background: var(--ok-dim); border-color: rgba(70,201,140,.4); color: var(--ok);
}
.ud-foot-btn-ok:not(:disabled):hover { background: rgba(70,201,140,.2); }

/* ── 覆盖态 ── */
.ud-loading-cover, .ud-error-cover {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 12px;
  color: var(--ink-2); font-size: 12.5px;
}
.ud-spinner {
  width: 28px; height: 28px; border-radius: 50%;
  border: 2px solid var(--line-1);
  border-top-color: var(--azure);
  animation: ud-spin 0.7s linear infinite;
}
@keyframes ud-spin { to { transform: rotate(360deg); } }
.ud-btn-ghost {
  padding: 6px 14px; border-radius: 8px; border: 1px solid var(--line-1);
  background: transparent; color: var(--ink-1); cursor: pointer;
  font-size: 12.5px; font-family: inherit; transition: background 0.12s;
}
.ud-btn-ghost:hover { background: var(--bg-2); }
.ud-muted { color: var(--ink-2); }
</style>
