<template>
  <AppLayout>
    <div class="dq-root">
      <!-- 页头 -->
      <div class="dq-head rise">
        <div>
          <h1 class="dq-title">驾驶舱</h1>
          <p class="dq-desc">运营第一屏 · 数据 5s 心跳刷新</p>
        </div>
        <button class="dq-btn" :disabled="loading" @click="reload">刷新</button>
      </div>

      <div v-if="loading && !stats" class="dq-spin">
        <LoadingSpinner size="md" />
      </div>

      <template v-else>
        <!-- ═══ 经营行 ═══ -->
        <div class="dq-kpi-row">
          <div class="dq-kpi rise" style="animation-delay:.04s">
            <div class="dq-kpi-glow"></div>
            <div class="dq-kpi-label">今日营收</div>
            <div class="dq-kpi-value q-money">${{ fmtMoney(payDash?.today_amount ?? 0) }}</div>
            <div class="dq-kpi-sub">{{ payDash?.today_count ?? 0 }} 笔订单</div>
          </div>
          <div class="dq-kpi rise" style="animation-delay:.08s">
            <div class="dq-kpi-label">今日请求</div>
            <div class="dq-kpi-value">{{ fmtNum(stats?.today_requests ?? 0) }}</div>
            <div class="dq-kpi-sub">累计 {{ fmtNum(stats?.total_requests ?? 0) }}</div>
          </div>
          <div class="dq-kpi rise" style="animation-delay:.12s">
            <div class="dq-kpi-label">新增用户</div>
            <div class="dq-kpi-value dq-ok">+{{ stats?.today_new_users ?? 0 }}</div>
            <div class="dq-kpi-sub">总计 {{ fmtNum(stats?.total_users ?? 0) }}</div>
          </div>
          <div class="dq-kpi rise" style="animation-delay:.16s">
            <div class="dq-kpi-label">今日消耗</div>
            <div class="dq-kpi-value q-money">${{ fmtMoney(stats?.today_actual_cost ?? 0) }}</div>
            <div class="dq-kpi-sub">成本 <span class="q-money">${{ fmtMoney(stats?.today_account_cost ?? 0) }}</span></div>
          </div>
        </div>

        <!-- ═══ 流量行 ═══ -->
        <div class="dq-traffic-row">
          <div class="dq-kpi rise" style="animation-delay:.20s">
            <div class="dq-kpi-label">今日 Token</div>
            <div class="dq-kpi-value">{{ fmtTok(stats?.today_tokens ?? 0) }}</div>
            <div class="dq-kpi-sub">累计 {{ fmtTok(stats?.total_tokens ?? 0) }}</div>
          </div>
          <div class="dq-kpi rise dq-kpi-azure" style="animation-delay:.24s">
            <div class="dq-kpi-label"><span class="dq-live-dot"></span>RPM</div>
            <div class="dq-kpi-value dq-azure">{{ fmtTok(stats?.rpm ?? 0) }}</div>
            <div class="dq-kpi-sub">TPM {{ fmtTok(stats?.tpm ?? 0) }}</div>
          </div>
          <div class="dq-kpi rise" style="animation-delay:.28s">
            <div class="dq-kpi-label">平均响应</div>
            <div class="dq-kpi-value">{{ fmtDur(stats?.average_duration_ms ?? 0) }}</div>
            <div class="dq-kpi-sub">{{ stats?.active_users ?? 0 }} 活跃用户</div>
          </div>
          <!-- 图表面板 -->
          <div class="dq-charts-panel rise" style="animation-delay:.30s">
            <div class="dq-chart-filters">
              <DateRangePicker v-model:start-date="startDate" v-model:end-date="endDate" @change="onRangeChange" />
              <Select v-model="granularity" :options="granOpts" class="dq-gran-sel" @change="loadCharts" />
            </div>
            <div class="dq-charts-grid">
              <ModelDistributionChart
                :model-stats="modelStats"
                :enable-ranking-view="true"
                :ranking-items="rankingItems"
                :ranking-total-actual-cost="rankingTotalActualCost"
                :ranking-total-requests="rankingTotalRequests"
                :ranking-total-tokens="rankingTotalTokens"
                :loading="chartsLoading"
                :ranking-loading="rankingLoading"
                :ranking-error="rankingError"
                :start-date="startDate"
                :end-date="endDate"
                @ranking-click="goToUsage"
              />
              <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
            </div>
          </div>
        </div>

        <!-- ═══ 异常行 ═══ -->
        <DashAnomalyRow
          :base-delay=".36"
          :error-accounts="stats?.error_accounts"
          :ratelimit-accounts="stats?.ratelimit_accounts"
          :normal-accounts="stats?.normal_accounts"
          :total-accounts="stats?.total_accounts"
          :active-keys="stats?.active_api_keys"
          :total-keys="stats?.total_api_keys"
          :alerts-loading="alertsLoading"
          :firing-count="firingAlerts"
          :resolved-count="resolvedAlerts"
          :latest-title="latestAlert?.title ?? latestAlert?.description ?? null"
          :latest-severity="latestAlert?.severity ?? null"
          @nav="goTo"
        />
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { opsAPI } from '@/api/admin/ops'
import { adminPaymentAPI } from '@/api/admin/payment'
import type { DashboardStats, TrendDataPoint, ModelStat, UserSpendingRankingItem } from '@/types'
import type { DashboardStats as PayDashboardStats } from '@/types/payment'
import type { AlertEvent } from '@/api/admin/ops'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import DashAnomalyRow from './DashAnomalyRow.vue'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false), chartsLoading = ref(false), rankingLoading = ref(false)
const rankingError = ref(false), alertsLoading = ref(false)
const stats = ref<DashboardStats | null>(null), payDash = ref<PayDashboardStats | null>(null)
const trendData = ref<TrendDataPoint[]>([]), modelStats = ref<ModelStat[]>([])
const rankingItems = ref<UserSpendingRankingItem[]>([])
const rankingTotalActualCost = ref(0), rankingTotalRequests = ref(0), rankingTotalTokens = ref(0)
const alertEvents = ref<AlertEvent[]>([])
let chartSeq = 0, rankingSeq = 0

const fmtLocalDate = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
const last24h = () => {
  const end = new Date()
  return { start: fmtLocalDate(new Date(end.getTime() - 86400000)), end: fmtLocalDate(end) }
}
const granularity = ref<'day' | 'hour'>('hour')
const r0 = last24h()
const startDate = ref(r0.start)
const endDate = ref(r0.end)

const granOpts = computed(() => [
  { value: 'hour', label: t('admin.dashboard.hour') },
  { value: 'day', label: t('admin.dashboard.day') }
])
const firingAlerts = computed(() => alertEvents.value.filter(e => e.status === 'firing').length)
const resolvedAlerts = computed(() => alertEvents.value.filter(e => e.status !== 'firing').length)
const latestAlert = computed(() => alertEvents.value.find(e => e.status === 'firing') ?? null)

function fmtNum(v: number) { return v.toLocaleString() }
function fmtMoney(v: number) {
  if (v >= 1000) return (v / 1000).toFixed(2) + 'K'
  if (v >= 1) return v.toFixed(2)
  if (v >= 0.01) return v.toFixed(3)
  return v.toFixed(4)
}
function fmtTok(v: number) {
  if (v >= 1e9) return (v / 1e9).toFixed(2) + 'B'
  if (v >= 1e6) return (v / 1e6).toFixed(2) + 'M'
  if (v >= 1e3) return (v / 1e3).toFixed(2) + 'K'
  return v.toLocaleString()
}
function fmtDur(ms: number) { return ms >= 1000 ? `${(ms / 1000).toFixed(2)}s` : `${Math.round(ms)}ms` }
function goTo(path: string) { void router.push(path) }
function goToUsage(item: UserSpendingRankingItem) {
  void router.push({ path: '/admin/usage', query: { user_id: String(item.user_id), start_date: startDate.value, end_date: endDate.value } })
}

async function loadSnapshot(withStats: boolean) {
  const seq = ++chartSeq
  if (withStats && !stats.value) loading.value = true
  chartsLoading.value = true
  try {
    const res = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value, end_date: endDate.value, granularity: granularity.value,
      include_stats: withStats, include_trend: true, include_model_stats: true
    })
    if (seq !== chartSeq) return
    if (withStats && res.stats) stats.value = res.stats
    trendData.value = res.trend ?? []
    modelStats.value = res.models ?? []
  } catch (e) {
    if (seq !== chartSeq) return
    appStore.showError('驾驶舱数据加载失败')
  } finally {
    if (seq === chartSeq) { loading.value = false; chartsLoading.value = false }
  }
}

async function loadRanking() {
  const seq = ++rankingSeq
  rankingLoading.value = true; rankingError.value = false
  try {
    const res = await adminAPI.dashboard.getUserSpendingRanking({ start_date: startDate.value, end_date: endDate.value, limit: 12 })
    if (seq !== rankingSeq) return
    rankingItems.value = res.ranking ?? []
    rankingTotalActualCost.value = res.total_actual_cost ?? 0
    rankingTotalRequests.value = res.total_requests ?? 0
    rankingTotalTokens.value = res.total_tokens ?? 0
  } catch { if (seq !== rankingSeq) return; rankingError.value = true }
  finally { if (seq === rankingSeq) rankingLoading.value = false }
}

async function loadPayDash() {
  try { const res = await adminPaymentAPI.getDashboard(1); payDash.value = res.data } catch { /* 支付未启用时静默降级 */ }
}

async function loadAlerts() {
  alertsLoading.value = true
  try {
    const [firing, resolved] = await Promise.all([
      opsAPI.listAlertEvents({ limit: 20, status: 'firing' }),
      opsAPI.listAlertEvents({ limit: 5, status: 'resolved' })
    ])
    alertEvents.value = [...firing, ...resolved]
  } catch { alertEvents.value = [] }
  finally { alertsLoading.value = false }
}

function onRangeChange(r: { startDate: string; endDate: string; preset: string | null }) {
  const diff = Math.ceil((new Date(r.endDate).getTime() - new Date(r.startDate).getTime()) / 86400000)
  granularity.value = diff <= 1 ? 'hour' : 'day'
  loadCharts()
}
async function loadCharts() { await Promise.all([loadSnapshot(false), loadRanking()]) }
async function reload() { await Promise.all([loadSnapshot(true), loadRanking(), loadPayDash(), loadAlerts()]) }

let heartbeat: ReturnType<typeof setInterval>
onMounted(() => { void reload(); heartbeat = setInterval(() => { void reload() }, 5000) })
onUnmounted(() => clearInterval(heartbeat))
</script>

<style scoped>
.dq-root { display: flex; flex-direction: column; gap: 14px; }

.rise { opacity: 0; transform: translateY(10px); animation: rise .5s cubic-bezier(.22,.68,0,1.2) forwards; }
@keyframes rise { to { opacity: 1; transform: none; } }
@media (prefers-reduced-motion: reduce) { .rise { animation: none; opacity: 1; transform: none; } .dq-live-dot { animation: none; } }

.dq-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 12px; }
.dq-title { font-size: 18px; font-weight: 700; color: var(--foreground); margin: 0; }
.dq-desc { font-size: 12px; color: var(--muted-foreground); margin: 3px 0 0; }
.dq-btn {
  padding: 6px 14px; border-radius: 9px; font-size: 12px; font-weight: 500;
  background: var(--secondary); border: 1px solid var(--border); color: var(--foreground);
  cursor: pointer; transition: background .15s;
}
.dq-btn:hover { background: var(--accent); }
.dq-btn:disabled { opacity: .5; cursor: default; }

.dq-kpi { @apply card p-4; position: relative; overflow: hidden; }
.dq-kpi-glow {
  position: absolute; right: -24px; top: -24px; width: 80px; height: 80px; border-radius: 50%;
  background: radial-gradient(circle, rgba(92,168,255,.12), transparent 70%); pointer-events: none;
}
.dq-kpi-label {
  font-size: 10.5px; font-weight: 600; letter-spacing: .08em; text-transform: uppercase;
  color: var(--muted-foreground); margin-bottom: 6px; display: flex; align-items: center; gap: 6px;
}
.dq-kpi-value { font-size: 24px; font-weight: 700; font-variant-numeric: tabular-nums; color: var(--foreground); line-height: 1.1; }
.dq-kpi-sub { font-size: 11px; color: var(--muted-foreground); margin-top: 5px; }
.dq-ok    { color: var(--ok); }
.dq-azure { color: var(--azure); }
.dq-kpi-azure { border-color: rgba(92,168,255,.25); }

.dq-live-dot {
  display: inline-block; width: 7px; height: 7px; border-radius: 50%;
  background: var(--azure); animation: pulse-b 1.8s infinite;
}
@keyframes pulse-b { 0%,100%{ box-shadow:0 0 0 0 rgba(92,168,255,.55);} 50%{ box-shadow:0 0 0 5px rgba(92,168,255,0);} }

.dq-kpi-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; }
@media (max-width: 900px) { .dq-kpi-row { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 540px) { .dq-kpi-row { grid-template-columns: 1fr; } }

.dq-traffic-row { display: grid; grid-template-columns: 1fr 1fr 1fr 2.2fr; gap: 12px; }
@media (max-width: 1100px) { .dq-traffic-row { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 1100px) { .dq-charts-panel { grid-column: 1 / -1; } }
@media (max-width: 540px) { .dq-traffic-row { grid-template-columns: 1fr; } }

.dq-charts-panel { @apply card; padding: 14px; display: flex; flex-direction: column; gap: 12px; }
.dq-chart-filters { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.dq-gran-sel { width: 90px; }
.dq-charts-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
@media (max-width: 800px) { .dq-charts-grid { grid-template-columns: 1fr; } }

.dq-spin { display: flex; justify-content: center; padding: 48px 0; }
</style>
