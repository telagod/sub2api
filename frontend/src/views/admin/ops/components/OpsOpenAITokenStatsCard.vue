<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { opsAPI, type OpsOpenAITokenStatsResponse, type OpsOpenAITokenStatsTimeRange } from '@/api/admin/ops'
import { formatNumber } from '@/utils/format'

interface Props {
  platformFilter?: string
  groupIdFilter?: number | null
  refreshToken: number
}

type ViewMode = 'topn' | 'pagination'

const props = withDefaults(defineProps<Props>(), {
  platformFilter: '',
  groupIdFilter: null
})

const { t } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const response = ref<OpsOpenAITokenStatsResponse | null>(null)

const timeRange = ref<OpsOpenAITokenStatsTimeRange>('30d')
const viewMode = ref<ViewMode>('topn')
const topN = ref<number>(20)
const page = ref<number>(1)
const pageSize = ref<number>(20)

const items = computed(() => response.value?.items ?? [])
const total = computed(() => response.value?.total ?? 0)
const totalPages = computed(() => {
  if (viewMode.value !== 'pagination') return 1
  const size = pageSize.value > 0 ? pageSize.value : 20
  return Math.max(1, Math.ceil(total.value / size))
})

const timeRangeOptions = computed(() => [
  { value: '30m', label: t('admin.ops.timeRange.30m') },
  { value: '1h', label: t('admin.ops.timeRange.1h') },
  { value: '1d', label: t('admin.ops.timeRange.1d') },
  { value: '15d', label: t('admin.ops.timeRange.15d') },
  { value: '30d', label: t('admin.ops.timeRange.30d') }
])

const viewModeOptions = computed(() => [
  { value: 'topn', label: t('admin.ops.openaiTokenStats.viewModeTopN') },
  { value: 'pagination', label: t('admin.ops.openaiTokenStats.viewModePagination') }
])

const topNOptions = computed(() => [
  { value: 10, label: 'Top 10' },
  { value: 20, label: 'Top 20' },
  { value: 50, label: 'Top 50' },
  { value: 100, label: 'Top 100' }
])

const pageSizeOptions = computed(() => [
  { value: 10, label: '10' },
  { value: 20, label: '20' },
  { value: 50, label: '50' },
  { value: 100, label: '100' }
])

function formatRate(v?: number | null): string {
  if (typeof v !== 'number' || !Number.isFinite(v)) return '-'
  return v.toFixed(2)
}

function formatInt(v?: number | null): string {
  if (typeof v !== 'number' || !Number.isFinite(v)) return '-'
  return formatNumber(Math.round(v))
}

function buildParams() {
  const params: Record<string, any> = {
    time_range: timeRange.value,
    platform: props.platformFilter || undefined,
    group_id: typeof props.groupIdFilter === 'number' && props.groupIdFilter > 0 ? props.groupIdFilter : undefined
  }

  if (viewMode.value === 'topn') {
    params.top_n = topN.value
  } else {
    params.page = page.value
    params.page_size = pageSize.value
  }
  return params
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    response.value = await opsAPI.getOpenAITokenStats(buildParams())
    // 防御：若 total 变化导致当前页超出最大页，则回退到末页并重新拉取一次。
    if (viewMode.value === 'pagination' && page.value > totalPages.value) {
      page.value = totalPages.value
      response.value = await opsAPI.getOpenAITokenStats(buildParams())
    }
  } catch (err: any) {
    console.error('[OpsOpenAITokenStatsCard] Failed to load data', err)
    response.value = null
    errorMessage.value = err?.message || t('admin.ops.openaiTokenStats.failedToLoad')
  } finally {
    loading.value = false
  }
}

watch(
  () => ({
    timeRange: timeRange.value,
    viewMode: viewMode.value,
    topN: topN.value,
    page: page.value,
    pageSize: pageSize.value,
    platform: props.platformFilter,
    groupId: props.groupIdFilter,
    refreshToken: props.refreshToken
  }),
  (next, prev) => {
    // 避免“筛选变化 -> 重置页码 -> 触发两次请求”：
    // 先只重置页码，等待下一次 watch（仅 page 变化）再发起请求。
    const filtersChanged = !prev ||
      next.timeRange !== prev.timeRange ||
      next.viewMode !== prev.viewMode ||
      next.pageSize !== prev.pageSize ||
      next.platform !== prev.platform ||
      next.groupId !== prev.groupId

    if (next.viewMode === 'pagination' && filtersChanged && next.page !== 1) {
      page.value = 1
      return
    }

    void loadData()
  },
  { immediate: true }
)

function onPrevPage() {
  if (viewMode.value !== 'pagination') return
  if (page.value > 1) page.value -= 1
}

function onNextPage() {
  if (viewMode.value !== 'pagination') return
  if (page.value < totalPages.value) page.value += 1
}
</script>

<template>
  <section class="od-card od-card-pad-sm">
    <div style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:10px;margin-bottom:14px;">
      <h3 class="od-chart-title">{{ t('admin.ops.openaiTokenStats.title') }}</h3>
      <div style="display:flex;flex-wrap:wrap;align-items:center;gap:6px;">
        <div style="width:136px;"><Select v-model="timeRange" :options="timeRangeOptions" /></div>
        <div style="width:136px;"><Select v-model="viewMode" :options="viewModeOptions" /></div>
        <div v-if="viewMode === 'topn'" style="width:100px;"><Select v-model="topN" :options="topNOptions" /></div>
        <template v-else>
          <div style="width:80px;"><Select v-model="pageSize" :options="pageSizeOptions" /></div>
          <button class="od-btn" style="padding:4px 10px;font-size:11px;" :disabled="loading || page <= 1" @click="onPrevPage">{{ t('admin.ops.openaiTokenStats.prevPage') }}</button>
          <button class="od-btn" style="padding:4px 10px;font-size:11px;" :disabled="loading || page >= totalPages" @click="onNextPage">{{ t('admin.ops.openaiTokenStats.nextPage') }}</button>
          <span style="font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.pageInfo', { page, total: totalPages }) }}</span>
        </template>
      </div>
    </div>

    <div v-if="errorMessage" style="margin-bottom:12px;border-radius:8px;border:1px solid var(--ops-bad-border);background:var(--ops-bad-dim);padding:8px 12px;font-size:11.5px;color:var(--ops-bad);">
      {{ errorMessage }}
    </div>

    <div v-if="loading" style="padding:28px 0;text-align:center;font-size:13px;color:var(--ink-2,#5C6470);">
      {{ t('admin.ops.loadingText') }}
    </div>

    <EmptyState v-else-if="items.length === 0" :title="t('common.noData')" :description="t('admin.ops.openaiTokenStats.empty')" />

    <div v-else>
      <div class="od-table-card">
        <div style="max-height:420px;overflow:auto;">
          <table style="min-width:100%;text-align:left;font-size:12px;border-collapse:collapse;">
            <thead class="od-table-head-row" style="position:sticky;top:0;z-index:10;">
              <tr>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.model') }}</th>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.requestCount') }}</th>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.avgTokensPerSec') }}</th>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.avgFirstTokenMs') }}</th>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.totalOutputTokens') }}</th>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.avgDurationMs') }}</th>
                <th style="padding:7px 10px;font-weight:600;color:var(--ink-2,#5C6470);">{{ t('admin.ops.openaiTokenStats.table.requestsWithFirstToken') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in items" :key="row.model" class="od-tr-border">
                <td style="padding:7px 10px;font-weight:500;color:var(--ink-0,#E8EBF0);">{{ row.model }}</td>
                <td style="padding:7px 10px;color:var(--ink-1,#97A0AF);">{{ formatInt(row.request_count) }}</td>
                <td style="padding:7px 10px;color:var(--ink-1,#97A0AF);">{{ formatRate(row.avg_tokens_per_sec) }}</td>
                <td style="padding:7px 10px;color:var(--ink-1,#97A0AF);">{{ formatRate(row.avg_first_token_ms) }}</td>
                <td style="padding:7px 10px;color:var(--ink-1,#97A0AF);">{{ formatInt(row.total_output_tokens) }}</td>
                <td style="padding:7px 10px;color:var(--ink-1,#97A0AF);">{{ formatInt(row.avg_duration_ms) }}</td>
                <td style="padding:7px 10px;color:var(--ink-1,#97A0AF);">{{ formatInt(row.requests_with_first_token) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div v-if="viewMode === 'topn'" style="margin-top:10px;font-size:11px;color:var(--ink-2,#5C6470);">
        {{ t('admin.ops.openaiTokenStats.totalModels', { total }) }}
      </div>
    </div>
  </section>
</template>

<style src="../ops-quench.css"></style>
