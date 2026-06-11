<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { opsAPI, type OpsAccountAvailabilityStatsResponse, type OpsConcurrencyStatsResponse, type OpsUserConcurrencyStatsResponse } from '@/api/admin/ops'

interface Props {
  platformFilter?: string
  groupIdFilter?: number | null
  refreshToken: number
}

const props = withDefaults(defineProps<Props>(), {
  platformFilter: '',
  groupIdFilter: null
})

const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const concurrency = ref<OpsConcurrencyStatsResponse | null>(null)
const availability = ref<OpsAccountAvailabilityStatsResponse | null>(null)
const userConcurrency = ref<OpsUserConcurrencyStatsResponse | null>(null)

// 用户视图开关
const showByUser = ref(false)

const realtimeEnabled = computed(() => {
  return (concurrency.value?.enabled ?? true) && (availability.value?.enabled ?? true)
})

function safeNumber(n: unknown): number {
  return typeof n === 'number' && Number.isFinite(n) ? n : 0
}

// 计算显示维度
const displayDimension = computed<'platform' | 'group' | 'account' | 'user'>(() => {
  if (showByUser.value) {
    return 'user'
  }
  if (typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0) {
    return 'account'
  }
  if (props.platformFilter) {
    return 'group'
  }
  return 'platform'
})

// 平台/分组汇总行数据
interface SummaryRow {
  key: string
  name: string
  platform?: string
  // 账号统计
  total_accounts: number
  available_accounts: number
  rate_limited_accounts: number
  error_accounts: number
  // 并发统计
  total_concurrency: number
  used_concurrency: number
  waiting_in_queue: number
  // 计算字段
  availability_percentage: number
  concurrency_percentage: number
}

// 账号详细行数据
interface AccountRow {
  key: string
  name: string
  platform: string
  group_name: string
  // 并发
  current_in_use: number
  max_capacity: number
  waiting_in_queue: number
  load_percentage: number
  // 状态
  is_available: boolean
  is_rate_limited: boolean
  rate_limit_remaining_sec?: number
  is_overloaded: boolean
  overload_remaining_sec?: number
  has_error: boolean
  error_message?: string
}

// 用户行数据
interface UserRow {
  key: string
  user_id: number
  user_email: string
  username: string
  current_in_use: number
  max_capacity: number
  waiting_in_queue: number
  load_percentage: number
}

// 平台维度汇总
const platformRows = computed((): SummaryRow[] => {
  const concStats = concurrency.value?.platform || {}
  const availStats = availability.value?.platform || {}

  const platforms = new Set([...Object.keys(concStats), ...Object.keys(availStats)])

  return Array.from(platforms).map(platform => {
    const conc = concStats[platform] || {}
    const avail = availStats[platform] || {}

    const totalAccounts = safeNumber(avail.total_accounts)
    const availableAccounts = safeNumber(avail.available_count)
    const totalConcurrency = safeNumber(conc.max_capacity)
    const usedConcurrency = safeNumber(conc.current_in_use)

    return {
      key: platform,
      name: platform.toUpperCase(),
      total_accounts: totalAccounts,
      available_accounts: availableAccounts,
      rate_limited_accounts: safeNumber(avail.rate_limit_count),

      error_accounts: safeNumber(avail.error_count),
      total_concurrency: totalConcurrency,
      used_concurrency: usedConcurrency,
      waiting_in_queue: safeNumber(conc.waiting_in_queue),
      availability_percentage: totalAccounts > 0 ? Math.round((availableAccounts / totalAccounts) * 100) : 0,
      concurrency_percentage: totalConcurrency > 0 ? Math.round((usedConcurrency / totalConcurrency) * 100) : 0
    }
  }).sort((a, b) => b.concurrency_percentage - a.concurrency_percentage)
})

// 分组维度汇总
const groupRows = computed((): SummaryRow[] => {
  const concStats = concurrency.value?.group || {}
  const availStats = availability.value?.group || {}

  const groupIds = new Set([...Object.keys(concStats), ...Object.keys(availStats)])

  const rows = Array.from(groupIds)
    .map(gid => {
      const conc = concStats[gid] || {}
      const avail = availStats[gid] || {}

      // 只显示匹配的平台
      if (props.platformFilter && conc.platform !== props.platformFilter && avail.platform !== props.platformFilter) {
        return null
      }

      const totalAccounts = safeNumber(avail.total_accounts)
      const availableAccounts = safeNumber(avail.available_count)
      const totalConcurrency = safeNumber(conc.max_capacity)
      const usedConcurrency = safeNumber(conc.current_in_use)

      return {
        key: gid,
        name: String(conc.group_name || avail.group_name || `Group ${gid}`),
        platform: String(conc.platform || avail.platform || ''),
        total_accounts: totalAccounts,
        available_accounts: availableAccounts,
        rate_limited_accounts: safeNumber(avail.rate_limit_count),
  
        error_accounts: safeNumber(avail.error_count),
        total_concurrency: totalConcurrency,
        used_concurrency: usedConcurrency,
        waiting_in_queue: safeNumber(conc.waiting_in_queue),
        availability_percentage: totalAccounts > 0 ? Math.round((availableAccounts / totalAccounts) * 100) : 0,
        concurrency_percentage: totalConcurrency > 0 ? Math.round((usedConcurrency / totalConcurrency) * 100) : 0
      }
    })
    .filter((row): row is NonNullable<typeof row> => row !== null)

  return rows.sort((a, b) => b.concurrency_percentage - a.concurrency_percentage)
})

// 账号维度详细
const accountRows = computed((): AccountRow[] => {
  const concStats = concurrency.value?.account || {}
  const availStats = availability.value?.account || {}

  const accountIds = new Set([...Object.keys(concStats), ...Object.keys(availStats)])

  const rows = Array.from(accountIds)
    .map(aid => {
      const conc = concStats[aid] || {}
      const avail = availStats[aid] || {}

      // 只显示匹配的分组
      if (typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0) {
        if (conc.group_id !== props.groupIdFilter && avail.group_id !== props.groupIdFilter) {
          return null
        }
      }

      return {
        key: aid,
        name: String(conc.account_name || avail.account_name || `Account ${aid}`),
        platform: String(conc.platform || avail.platform || ''),
        group_name: String(conc.group_name || avail.group_name || ''),
        current_in_use: safeNumber(conc.current_in_use),
        max_capacity: safeNumber(conc.max_capacity),
        waiting_in_queue: safeNumber(conc.waiting_in_queue),
        load_percentage: safeNumber(conc.load_percentage),
        is_available: avail.is_available || false,
        is_rate_limited: avail.is_rate_limited || false,
        rate_limit_remaining_sec: avail.rate_limit_remaining_sec,
        is_overloaded: avail.is_overloaded || false,
        overload_remaining_sec: avail.overload_remaining_sec,
        has_error: avail.has_error || false,
        error_message: avail.error_message || ''
      }
    })
    .filter((row): row is NonNullable<typeof row> => row !== null)

  return rows.sort((a, b) => {
    // 优先显示异常账号
    if (a.has_error !== b.has_error) return a.has_error ? -1 : 1
    if (a.is_rate_limited !== b.is_rate_limited) return a.is_rate_limited ? -1 : 1
    // 然后按负载排序
    return b.load_percentage - a.load_percentage
  })
})

// 用户维度详细
const userRows = computed((): UserRow[] => {
  const userStats = userConcurrency.value?.user || {}

  return Object.keys(userStats)
    .map(uid => {
      const u = userStats[uid] || {}
      return {
        key: uid,
        user_id: safeNumber(u.user_id),
        user_email: u.user_email || `User ${uid}`,
        username: u.username || '',
        current_in_use: safeNumber(u.current_in_use),
        max_capacity: safeNumber(u.max_capacity),
        waiting_in_queue: safeNumber(u.waiting_in_queue),
        load_percentage: safeNumber(u.load_percentage)
      }
    })
    .sort((a, b) => b.current_in_use - a.current_in_use || b.load_percentage - a.load_percentage)
})

// 根据维度选择数据
const displayRows = computed(() => {
  if (displayDimension.value === 'user') return userRows.value
  if (displayDimension.value === 'account') return accountRows.value
  if (displayDimension.value === 'group') return groupRows.value
  return platformRows.value
})

const displayTitle = computed(() => {
  if (displayDimension.value === 'user') return t('admin.ops.concurrency.byUser')
  if (displayDimension.value === 'account') return t('admin.ops.concurrency.byAccount')
  if (displayDimension.value === 'group') return t('admin.ops.concurrency.byGroup')
  return t('admin.ops.concurrency.byPlatform')
})

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    if (showByUser.value) {
      // 用户视图模式只加载用户并发数据
      const userData = await opsAPI.getUserConcurrencyStats()
      userConcurrency.value = userData
    } else {
      // 常规模式加载账号/平台/分组数据
      const [concData, availData] = await Promise.all([
        opsAPI.getConcurrencyStats(props.platformFilter, props.groupIdFilter),
        opsAPI.getAccountAvailabilityStats(props.platformFilter, props.groupIdFilter)
      ])
      concurrency.value = concData
      availability.value = availData
    }
  } catch (err: any) {
    console.error('[OpsConcurrencyCard] Failed to load data', err)
    errorMessage.value = err?.response?.data?.detail || t('admin.ops.concurrency.loadFailed')
  } finally {
    loading.value = false
  }
}

// 刷新节奏由父组件统一控制（OpsDashboard Header 的刷新状态/倒计时）
watch(
  () => props.refreshToken,
  () => {
    if (!realtimeEnabled.value) return
    loadData()
  }
)

// 切换用户视图时重新加载数据
watch(
  () => showByUser.value,
  () => {
    loadData()
  }
)

function getLoadBarClass(loadPct: number): string {
  if (loadPct >= 90) return 'od-progress-fill bad'
  if (loadPct >= 70) return 'od-progress-fill warn'
  if (loadPct >= 50) return 'od-progress-fill warn'
  return 'od-progress-fill ok'
}

function getLoadBarStyle(loadPct: number): string {
  return `width: ${Math.min(100, Math.max(0, loadPct))}%`
}

function getLoadTextClass(loadPct: number): string {
  if (loadPct >= 90) return 'od-c-bad'
  if (loadPct >= 70) return 'od-c-warn'
  if (loadPct >= 50) return 'od-c-warn'
  return 'od-c-ok'
}

function formatDuration(seconds: number): string {
  if (seconds <= 0) return '0s'
  if (seconds < 60) return `${Math.round(seconds)}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  return `${hours}h`
}


watch(
  () => realtimeEnabled.value,
  async (enabled) => {
    if (enabled) {
      await loadData()
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="od-conc-card">
    <!-- 头部 -->
    <div class="od-conc-head">
      <h3 class="od-conc-title">
        <svg class="od-chart-icon" width="14" height="14" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        {{ t('admin.ops.concurrency.title') }}
      </h3>
      <div style="display:flex;align-items:center;gap:6px;">
        <!-- 用户视图切换按钮 -->
        <button
          class="od-btn od-btn-icon"
          :class="showByUser ? 'od-btn-azure' : ''"
          :title="showByUser ? t('admin.ops.concurrency.switchToPlatform') : t('admin.ops.concurrency.switchToUser')"
          :aria-label="showByUser ? t('admin.ops.concurrency.switchToPlatform') : t('admin.ops.concurrency.switchToUser')"
          @click="showByUser = !showByUser"
        >
          <svg width="13" height="13" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </button>
        <!-- 刷新按钮 -->
        <button class="od-btn od-btn-icon" :disabled="loading" :title="t('common.refresh')" :aria-label="t('common.refresh')" @click="loadData">
          <svg width="13" height="13" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMessage" style="margin-bottom:10px;border-radius:8px;background:var(--ops-bad-dim);border:1px solid var(--ops-bad-border);padding:9px 12px;font-size:11.5px;color:var(--ops-bad);">
      {{ errorMessage }}
    </div>

    <!-- 禁用状态 -->
    <div
      v-if="!realtimeEnabled"
      style="flex:1;display:flex;align-items:center;justify-content:center;border-radius:8px;border:1px dashed var(--line-0,#20242C);font-size:13px;color:var(--ink-2,#5C6470);"
    >
      {{ t('admin.ops.concurrency.disabledHint') }}
    </div>

    <!-- 数据展示区域 -->
    <div v-else class="od-conc-body">
      <!-- 维度标题栏 -->
      <div class="od-conc-dim-bar">
        <span class="od-conc-dim-label">{{ displayTitle }}</span>
        <span style="font-size:10px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.concurrency.totalRows', { count: displayRows.length }) }}</span>
      </div>

      <!-- 空状态 -->
      <div v-if="displayRows.length === 0" style="flex:1;display:flex;align-items:center;justify-content:center;font-size:13px;color:var(--ink-2,#5C6470);">
        {{ t('admin.ops.concurrency.empty') }}
      </div>

      <!-- 用户视图 -->
      <div v-else-if="displayDimension === 'user'" class="od-conc-scroll" style="max-height:360px;">
        <div v-for="row in (displayRows as UserRow[])" :key="row.key" class="od-conc-row">
          <div style="display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:5px;">
            <div style="display:flex;min-width:0;flex:1;align-items:center;gap:5px;">
              <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:11px;font-weight:700;color:var(--ink-0,#E8EBF0);" :title="row.username || row.user_email">{{ row.username || row.user_email }}</span>
              <span v-if="row.username" style="flex-shrink:0;font-size:10px;color:var(--ink-2,#5C6470);overflow:hidden;text-overflow:ellipsis;" :title="row.user_email">{{ row.user_email }}</span>
            </div>
            <div style="display:flex;flex-shrink:0;align-items:center;gap:6px;font-size:10px;">
              <span class="od-mono" style="font-weight:700;color:var(--ink-0,#E8EBF0);">{{ row.current_in_use }}/{{ row.max_capacity }}</span>
              <span :class="getLoadTextClass(row.load_percentage)" style="font-weight:700;">{{ Math.round(row.load_percentage) }}%</span>
            </div>
          </div>
          <div class="od-progress"><div :class="getLoadBarClass(row.load_percentage)" :style="getLoadBarStyle(row.load_percentage)"></div></div>
          <div v-if="row.waiting_in_queue > 0" style="margin-top:5px;display:flex;justify-content:flex-end;">
            <span class="od-badge od-badge-azure">{{ t('admin.ops.concurrency.queued', { count: row.waiting_in_queue }) }}</span>
          </div>
        </div>
      </div>

      <!-- 汇总视图（平台/分组） -->
      <div v-else-if="displayDimension === 'platform' || displayDimension === 'group'" class="od-conc-scroll" style="max-height:360px;">
        <div v-for="row in (displayRows as SummaryRow[])" :key="row.key" class="od-conc-row">
          <div style="display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:6px;">
            <div style="display:flex;align-items:center;gap:6px;">
              <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:11px;font-weight:700;color:var(--ink-0,#E8EBF0);" :title="row.name">{{ row.name }}</span>
              <span v-if="displayDimension === 'group' && row.platform" style="font-size:10px;color:var(--ink-2,#5C6470);">{{ row.platform.toUpperCase() }}</span>
            </div>
            <div style="display:flex;flex-shrink:0;align-items:center;gap:6px;font-size:10px;">
              <span class="od-mono" style="font-weight:700;color:var(--ink-0,#E8EBF0);">{{ row.used_concurrency }}/{{ row.total_concurrency }}</span>
              <span :class="getLoadTextClass(row.concurrency_percentage)" style="font-weight:700;">{{ row.concurrency_percentage }}%</span>
            </div>
          </div>
          <div class="od-progress" style="margin-bottom:6px;"><div :class="getLoadBarClass(row.concurrency_percentage)" :style="getLoadBarStyle(row.concurrency_percentage)"></div></div>
          <div style="display:flex;flex-wrap:wrap;align-items:center;gap:6px 10px;font-size:10px;">
            <div style="display:flex;align-items:center;gap:4px;">
              <svg width="11" height="11" fill="none" viewBox="0 0 24 24" stroke="currentColor" style="color:var(--ink-2,#5C6470);">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
              </svg>
              <span style="color:var(--ink-1,#97A0AF);"><span class="od-c-ok" style="font-weight:700;">{{ row.available_accounts }}</span>/{{ row.total_accounts }}</span>
              <span style="color:var(--ink-2,#5C6470);">{{ row.availability_percentage }}%</span>
            </div>
            <span v-if="row.rate_limited_accounts > 0" class="od-badge od-badge-warn">{{ t('admin.ops.concurrency.rateLimited', { count: row.rate_limited_accounts }) }}</span>
            <span v-if="row.error_accounts > 0" class="od-badge od-badge-bad">{{ t('admin.ops.concurrency.errorAccounts', { count: row.error_accounts }) }}</span>
            <span v-if="row.waiting_in_queue > 0" class="od-badge od-badge-azure">{{ t('admin.ops.concurrency.queued', { count: row.waiting_in_queue }) }}</span>
          </div>
        </div>
      </div>

      <!-- 账号详细视图 -->
      <div v-else class="od-conc-scroll" style="max-height:360px;">
        <div v-for="row in (displayRows as AccountRow[])" :key="row.key" class="od-conc-row">
          <div style="display:flex;align-items:center;justify-content:space-between;gap:8px;margin-bottom:5px;">
            <div style="min-width:0;flex:1;">
              <div style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:11px;font-weight:700;color:var(--ink-0,#E8EBF0);" :title="row.name">{{ row.name }}</div>
              <div style="margin-top:1px;font-size:9px;color:var(--ink-2,#5C6470);">{{ row.group_name }}</div>
            </div>
            <div style="display:flex;flex-shrink:0;align-items:center;gap:6px;">
              <span class="od-mono" style="font-size:11px;font-weight:700;color:var(--ink-0,#E8EBF0);">{{ row.current_in_use }}/{{ row.max_capacity }}</span>
              <span v-if="row.is_available" class="od-badge od-badge-ok">
                <svg width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
                {{ t('admin.ops.accountAvailability.available') }}
              </span>
              <span v-else-if="row.is_rate_limited" class="od-badge od-badge-warn">
                <svg width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                {{ formatDuration(row.rate_limit_remaining_sec || 0) }}
              </span>
              <span v-else-if="row.is_overloaded" class="od-badge od-badge-bad">
                <svg width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
                {{ formatDuration(row.overload_remaining_sec || 0) }}
              </span>
              <span v-else-if="row.has_error" class="od-badge od-badge-bad">
                <svg width="10" height="10" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                {{ t('admin.ops.accountAvailability.accountError') }}
              </span>
              <span v-else class="od-badge od-badge-dim">{{ t('admin.ops.accountAvailability.unavailable') }}</span>
            </div>
          </div>
          <div class="od-progress"><div :class="getLoadBarClass(row.load_percentage)" :style="getLoadBarStyle(row.load_percentage)"></div></div>
          <div v-if="row.waiting_in_queue > 0" style="margin-top:5px;display:flex;justify-content:flex-end;">
            <span class="od-badge od-badge-azure">{{ t('admin.ops.concurrency.queued', { count: row.waiting_in_queue }) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style src="../ops-quench.css"></style>
