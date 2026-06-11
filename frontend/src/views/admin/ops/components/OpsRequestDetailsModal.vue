<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsRequestDetailsParams, type OpsRequestDetail } from '@/api/admin/ops'
import { parseTimeRangeMinutes, formatDateTime } from '../utils/opsFormatters'

export interface OpsRequestDetailsPreset {
  title: string
  kind?: OpsRequestDetailsParams['kind']
  sort?: OpsRequestDetailsParams['sort']
  min_duration_ms?: number
  max_duration_ms?: number
}

interface Props {
  modelValue: boolean
  timeRange: string
  preset: OpsRequestDetailsPreset
  platform?: string
  groupId?: number | null
}

const props = defineProps<Props>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'openErrorDetail', errorId: number): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const items = ref<OpsRequestDetail[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

const close = () => emit('update:modelValue', false)

const rangeLabel = computed(() => {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  if (minutes >= 60) return t('admin.ops.requestDetails.rangeHours', { n: Math.round(minutes / 60) })
  return t('admin.ops.requestDetails.rangeMinutes', { n: minutes })
})

function buildTimeParams(): Pick<OpsRequestDetailsParams, 'start_time' | 'end_time'> {
  const minutes = parseTimeRangeMinutes(props.timeRange)
  const endTime = new Date()
  const startTime = new Date(endTime.getTime() - minutes * 60 * 1000)
  return {
    start_time: startTime.toISOString(),
    end_time: endTime.toISOString()
  }
}

const fetchData = async () => {
  if (!props.modelValue) return
  loading.value = true
  try {
    const params: OpsRequestDetailsParams = {
      ...buildTimeParams(),
      page: page.value,
      page_size: pageSize.value,
      kind: props.preset.kind ?? 'all',
      sort: props.preset.sort ?? 'created_at_desc'
    }

    const platform = (props.platform || '').trim()
    if (platform) params.platform = platform
    if (typeof props.groupId === 'number' && props.groupId > 0) params.group_id = props.groupId

    if (typeof props.preset.min_duration_ms === 'number') params.min_duration_ms = props.preset.min_duration_ms
    if (typeof props.preset.max_duration_ms === 'number') params.max_duration_ms = props.preset.max_duration_ms

    const res = await opsAPI.listRequestDetails(params)
    items.value = res.items || []
    total.value = res.total || 0
  } catch (e: any) {
    console.error('[OpsRequestDetailsModal] Failed to fetch request details', e)
    appStore.showError(e?.message || t('admin.ops.requestDetails.failedToLoad'))
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      page.value = 1
      pageSize.value = 10
      fetchData()
    }
  }
)

watch(
  () => [
    props.timeRange,
    props.platform,
    props.groupId,
    props.preset.kind,
    props.preset.sort,
    props.preset.min_duration_ms,
    props.preset.max_duration_ms
  ],
  () => {
    if (!props.modelValue) return
    page.value = 1
    fetchData()
  }
)

function handlePageChange(next: number) {
  page.value = next
  fetchData()
}

function handlePageSizeChange(next: number) {
  pageSize.value = next
  page.value = 1
  fetchData()
}

async function handleCopyRequestId(requestId: string) {
  const ok = await copyToClipboard(requestId, t('admin.ops.requestDetails.requestIdCopied'))
  if (ok) return
  // `useClipboard` already shows toast on failure; this keeps UX consistent with older ops modal.
  appStore.showWarning(t('admin.ops.requestDetails.copyFailed'))
}

function openErrorDetail(errorId: number | null | undefined) {
  if (!errorId) return
  close()
  emit('openErrorDetail', errorId)
}

const kindBadgeClass = (kind: string) => {
  if (kind === 'error') return 'od-badge od-badge-bad'
  return 'od-badge od-badge-ok'
}
</script>

<template>
  <BaseDialog :show="modelValue" :title="props.preset.title || t('admin.ops.requestDetails.title')" width="full" @close="close">
    <template #default>
      <div style="display:flex;height:100%;min-height:0;flex-direction:column;">
        <div style="margin-bottom:12px;flex-shrink:0;display:flex;align-items:center;justify-content:space-between;">
          <div style="font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.requestDetails.rangeLabel', { range: rangeLabel }) }}</div>
          <button type="button" class="od-btn" style="padding:4px 10px;font-size:11px;" @click="fetchData">{{ t('common.refresh') }}</button>
        </div>

        <!-- Loading -->
        <div v-if="loading" style="display:flex;flex:1;align-items:center;justify-content:center;padding:48px 0;flex-direction:column;gap:10px;" role="status" :aria-label="t('common.loading')">
          <svg width="24" height="24" class="animate-spin" fill="none" viewBox="0 0 24 24" style="color:var(--ops-azure,#5CA8FF);" aria-hidden="true"><circle opacity=".25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path opacity=".75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
          <span style="font-size:13px;color:var(--ink-2,#5C6470);" aria-hidden="true">{{ t('common.loading') }}</span>
        </div>

        <!-- Table -->
        <div v-else style="display:flex;min-height:0;flex:1;flex-direction:column;">
          <div v-if="items.length === 0" style="border-radius:8px;border:1px dashed var(--line-0,#20242C);padding:36px;text-align:center;">
            <div style="font-size:13px;font-weight:500;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.requestDetails.empty') }}</div>
            <div style="margin-top:4px;font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.requestDetails.emptyHint') }}</div>
          </div>

          <div v-else class="od-table-card" style="display:flex;min-height:0;flex:1;flex-direction:column;overflow:hidden;">
            <div style="min-height:0;flex:1;overflow:auto;">
              <table style="min-width:100%;border-collapse:collapse;font-size:11.5px;">
                <thead class="od-table-head-row" style="position:sticky;top:0;z-index:10;">
                  <tr>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.time') }}</th>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.kind') }}</th>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.platform') }}</th>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.model') }}</th>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.duration') }}</th>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.status') }}</th>
                    <th style="padding:8px 14px;text-align:left;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.requestId') }}</th>
                    <th style="padding:8px 14px;text-align:right;" class="od-sys-label">{{ t('admin.ops.requestDetails.table.actions') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in items" :key="idx" class="od-tr-border">
                    <td style="padding:8px 14px;white-space:nowrap;color:var(--ink-1,#97A0AF);">{{ formatDateTime(row.created_at) }}</td>
                    <td style="padding:8px 14px;white-space:nowrap;">
                      <span :class="kindBadgeClass(row.kind)">{{ row.kind === 'error' ? t('admin.ops.requestDetails.kind.error') : t('admin.ops.requestDetails.kind.success') }}</span>
                    </td>
                    <td style="padding:8px 14px;white-space:nowrap;font-weight:500;color:var(--ink-1,#97A0AF);">{{ (row.platform || 'unknown').toUpperCase() }}</td>
                    <td style="padding:8px 14px;max-width:240px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--ink-1,#97A0AF);" :title="row.model || ''">{{ row.model || '-' }}</td>
                    <td style="padding:8px 14px;white-space:nowrap;color:var(--ink-1,#97A0AF);">{{ typeof row.duration_ms === 'number' ? `${row.duration_ms} ms` : '-' }}</td>
                    <td style="padding:8px 14px;white-space:nowrap;color:var(--ink-1,#97A0AF);">{{ row.status_code ?? '-' }}</td>
                    <td style="padding:8px 14px;">
                      <div v-if="row.request_id" style="display:flex;align-items:center;gap:6px;">
                        <span class="od-mono" style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:10.5px;color:var(--ink-1,#97A0AF);" :title="row.request_id">{{ row.request_id }}</span>
                        <button class="od-btn" style="padding:2px 7px;font-size:10px;" @click="handleCopyRequestId(row.request_id)">{{ t('admin.ops.requestDetails.copy') }}</button>
                      </div>
                      <span v-else style="font-size:11px;color:var(--ink-2,#5C6470);">-</span>
                    </td>
                    <td style="padding:8px 14px;white-space:nowrap;text-align:right;">
                      <button v-if="row.kind === 'error' && row.error_id" class="od-btn" style="padding:3px 9px;font-size:11px;color:var(--ops-bad,#F25C69);border-color:var(--ops-bad-border,rgba(242,92,105,.25));" @click="openErrorDetail(row.error_id)">{{ t('admin.ops.requestDetails.viewError') }}</button>
                      <span v-else style="font-size:11px;color:var(--ink-2,#5C6470);">-</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <Pagination :total="total" :page="page" :page-size="pageSize" @update:page="handlePageChange" @update:pageSize="handlePageSizeChange" />
          </div>
        </div>
      </div>
    </template>
  </BaseDialog>
</template>

<style src="../ops-quench.css"></style>
