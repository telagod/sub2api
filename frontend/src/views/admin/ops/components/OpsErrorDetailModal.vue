<template>
  <BaseDialog :show="show" :title="title" width="full" :close-on-click-outside="true" @close="close">
    <div v-if="loading" style="display:flex;align-items:center;justify-content:center;padding:56px 0;flex-direction:column;gap:10px;">
      <div style="width:28px;height:28px;border-radius:50%;border-bottom:2px solid var(--ops-azure,#5CA8FF);" class="animate-spin"></div>
      <div style="font-size:13px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.errorDetail.loading') }}</div>
    </div>

    <div v-else-if="!detail" style="padding:36px 0;text-align:center;font-size:13px;color:var(--ink-2,#5C6470);">{{ emptyText }}</div>

    <div v-else style="padding:20px;display:flex;flex-direction:column;gap:16px;">
      <!-- Summary grid -->
      <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:10px;">
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.requestId') }}</div>
          <div class="od-sys-val od-mono" style="word-break:break-all;">{{ requestId || '—' }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.time') }}</div>
          <div class="od-sys-val">{{ formatDateTime(detail.created_at) }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ isUpstreamError(detail) ? t('admin.ops.errorDetail.account') : t('admin.ops.errorDetail.user') }}</div>
          <div class="od-sys-val">
            <template v-if="isUpstreamError(detail)">{{ detail.account_name || (detail.account_id != null ? String(detail.account_id) : '—') }}</template>
            <template v-else>{{ detail.user_email || (detail.user_id != null ? String(detail.user_id) : '—') }}</template>
          </div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.platform') }}</div>
          <div class="od-sys-val">{{ detail.platform || '—' }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.group') }}</div>
          <div class="od-sys-val">{{ detail.group_name || (detail.group_id != null ? String(detail.group_id) : '—') }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.model') }}</div>
          <div class="od-sys-val">
            <template v-if="hasModelMapping(detail)">
              <span class="od-mono">{{ detail.requested_model }}</span>
              <span style="margin:0 4px;color:var(--ink-2,#5C6470);">→</span>
              <span class="od-mono" style="color:var(--ops-azure,#5CA8FF);">{{ detail.upstream_model }}</span>
            </template>
            <template v-else>{{ displayModel(detail) || '—' }}</template>
          </div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.inboundEndpoint') }}</div>
          <div class="od-sys-val od-mono" style="word-break:break-all;">{{ detail.inbound_endpoint || '—' }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.upstreamEndpoint') }}</div>
          <div class="od-sys-val od-mono" style="word-break:break-all;">{{ detail.upstream_endpoint || '—' }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.status') }}</div>
          <div style="margin-top:4px;">
            <span :class="['od-badge', statusClass]">{{ detail.status_code }}</span>
          </div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.requestType') }}</div>
          <div class="od-sys-val">{{ formatRequestTypeLabel(detail.request_type) }}</div>
        </div>
        <div class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.message') }}</div>
          <div class="od-sys-val" style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;" :title="detail.message">{{ detail.message || '—' }}</div>
        </div>
        <div v-if="detail.api_key_prefix" class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.apiKeyPrefix') }}</div>
          <div class="od-sys-val od-mono">{{ detail.api_key_prefix }}</div>
        </div>
        <div v-if="detail.attempted_key_prefix" class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.attemptedKeyPrefix') }}</div>
          <div class="od-sys-val od-mono">{{ detail.attempted_key_prefix }}</div>
        </div>
        <div v-if="detail.deleted_key_owner_email" class="od-sys-card">
          <div class="od-sys-label">{{ t('admin.ops.errorDetail.deletedKeyOwner') }}</div>
          <div class="od-sys-val">
            {{ detail.deleted_key_owner_email }}
            <span v-if="detail.deleted_key_name" style="margin-left:4px;font-size:11px;color:var(--ink-2,#5C6470);">({{ detail.deleted_key_name }})</span>
            <span class="od-badge od-badge-bad" style="margin-left:6px;font-size:10px;">{{ t('admin.ops.errorDetail.keyDeletedBadge') }}</span>
          </div>
        </div>
      </div>

      <!-- Response Body -->
      <div class="od-card" style="padding:16px;">
        <h3 style="font-size:11.5px;font-weight:700;text-transform:uppercase;letter-spacing:.05em;color:var(--ink-0,#E8EBF0);">{{ t('admin.ops.errorDetail.responseBody') }}</h3>
        <pre style="margin-top:12px;max-height:480px;overflow:auto;border-radius:6px;border:1px solid var(--line-0,#20242C);background:var(--bg-3,#0E1014);padding:14px;font-size:11.5px;color:var(--ink-1,#97A0AF);"><code>{{ prettyJSON(primaryResponseBody || '') }}</code></pre>
      </div>

      <!-- Upstream errors -->
      <div v-if="showUpstreamList" class="od-card" style="padding:16px;">
        <div style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:8px;">
          <h3 style="font-size:11.5px;font-weight:700;text-transform:uppercase;letter-spacing:.05em;color:var(--ink-0,#E8EBF0);">{{ t('admin.ops.errorDetails.upstreamErrors') }}</h3>
          <div v-if="correlatedUpstreamLoading" style="font-size:11px;color:var(--ink-2,#5C6470);">{{ t('common.loading') }}</div>
        </div>
        <div v-if="!correlatedUpstreamLoading && !correlatedUpstreamErrors.length" style="margin-top:10px;font-size:13px;color:var(--ink-2,#5C6470);">{{ t('common.noData') }}</div>
        <div v-else style="margin-top:12px;display:flex;flex-direction:column;gap:10px;">
          <div v-for="(ev, idx) in correlatedUpstreamErrors" :key="ev.id" class="od-card" style="padding:14px;">
            <div style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:8px;">
              <div style="font-size:12px;font-weight:700;color:var(--ink-0,#E8EBF0);">
                #{{ idx + 1 }}
                <span v-if="ev.type" class="od-badge od-badge-dim od-mono" style="margin-left:6px;font-size:10px;">{{ ev.type }}</span>
              </div>
              <div style="display:flex;align-items:center;gap:8px;">
                <div class="od-mono" style="font-size:11.5px;color:var(--ink-2,#5C6470);">{{ ev.status_code ?? '—' }}</div>
                <button type="button" class="od-btn" style="padding:2px 8px;font-size:10px;display:inline-flex;align-items:center;gap:4px;" :disabled="!getUpstreamResponsePreview(ev)" :title="getUpstreamResponsePreview(ev) ? '' : t('common.noData')" @click="toggleUpstreamDetail(ev.id)">
                  <Icon :name="expandedUpstreamDetailIds.has(ev.id) ? 'chevronDown' : 'chevronRight'" size="xs" :stroke-width="2" />
                  {{ expandedUpstreamDetailIds.has(ev.id) ? t('admin.ops.errorDetail.responsePreview.collapse') : t('admin.ops.errorDetail.responsePreview.expand') }}
                </button>
              </div>
            </div>
            <div style="margin-top:10px;display:grid;grid-template-columns:1fr 1fr;gap:6px;font-size:11.5px;color:var(--ink-1,#97A0AF);">
              <div><span style="color:var(--ink-2,#5C6470);">{{ t('admin.ops.errorDetail.upstreamEvent.status') }}:</span><span class="od-mono" style="margin-left:4px;">{{ ev.status_code ?? '—' }}</span></div>
              <div><span style="color:var(--ink-2,#5C6470);">{{ t('admin.ops.errorDetail.upstreamEvent.requestId') }}:</span><span class="od-mono" style="margin-left:4px;">{{ ev.request_id || ev.client_request_id || '—' }}</span></div>
            </div>
            <div v-if="ev.message" style="margin-top:8px;word-break:break-word;font-size:13px;font-weight:500;color:var(--ink-0,#E8EBF0);">{{ ev.message }}</div>
            <pre v-if="expandedUpstreamDetailIds.has(ev.id)" style="margin-top:10px;max-height:200px;overflow:auto;border-radius:6px;border:1px solid var(--line-0,#20242C);background:var(--bg-2,#171A20);padding:10px;font-size:11px;color:var(--ink-1,#97A0AF);"><code>{{ prettyJSON(getUpstreamResponsePreview(ev)) }}</code></pre>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>
</template>

<style src="../ops-quench.css"></style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { opsAPI, type OpsErrorDetail } from '@/api/admin/ops'
import { formatDateTime } from '@/utils/format'
import { resolvePrimaryResponseBody, resolveUpstreamPayload } from '../utils/errorDetailResponse'

interface Props {
  show: boolean
  errorId: number | null
  errorType?: 'request' | 'upstream'
}

interface Emits {
  (e: 'update:show', value: boolean): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const detail = ref<OpsErrorDetail | null>(null)

const showUpstreamList = computed(() => props.errorType === 'request')

const requestId = computed(() => detail.value?.request_id || detail.value?.client_request_id || '')

const primaryResponseBody = computed(() => {
  return resolvePrimaryResponseBody(detail.value, props.errorType)
})




const title = computed(() => {
  if (!props.errorId) return t('admin.ops.errorDetail.title')
  return t('admin.ops.errorDetail.titleWithId', { id: String(props.errorId) })
})

const emptyText = computed(() => t('admin.ops.errorDetail.noErrorSelected'))

function isUpstreamError(d: OpsErrorDetail | null): boolean {
  if (!d) return false
  const phase = String(d.phase || '').toLowerCase()
  const owner = String(d.error_owner || '').toLowerCase()
  return phase === 'upstream' && owner === 'provider'
}

function formatRequestTypeLabel(type: number | null | undefined): string {
  switch (type) {
    case 1: return t('admin.ops.errorDetail.requestTypeSync')
    case 2: return t('admin.ops.errorDetail.requestTypeStream')
    case 3: return t('admin.ops.errorDetail.requestTypeWs')
    default: return t('admin.ops.errorDetail.requestTypeUnknown')
  }
}

function hasModelMapping(d: OpsErrorDetail | null): boolean {
  if (!d) return false
  const requested = String(d.requested_model || '').trim()
  const upstream = String(d.upstream_model || '').trim()
  return !!requested && !!upstream && requested !== upstream
}

function displayModel(d: OpsErrorDetail | null): string {
  if (!d) return ''
  const upstream = String(d.upstream_model || '').trim()
  if (upstream) return upstream
  const requested = String(d.requested_model || '').trim()
  if (requested) return requested
  return String(d.model || '').trim()
}

const correlatedUpstream = ref<OpsErrorDetail[]>([])
const correlatedUpstreamLoading = ref(false)

const correlatedUpstreamErrors = computed<OpsErrorDetail[]>(() => correlatedUpstream.value)

const expandedUpstreamDetailIds = ref(new Set<number>())

function getUpstreamResponsePreview(ev: OpsErrorDetail): string {
  const upstreamPayload = resolveUpstreamPayload(ev)
  if (upstreamPayload) return upstreamPayload
  return String(ev.error_body || '').trim()
}

function toggleUpstreamDetail(id: number) {
  const next = new Set(expandedUpstreamDetailIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expandedUpstreamDetailIds.value = next
}

async function fetchCorrelatedUpstreamErrors(requestErrorId: number) {
  correlatedUpstreamLoading.value = true
  try {
    const res = await opsAPI.listRequestErrorUpstreamErrors(
      requestErrorId,
      { page: 1, page_size: 100, view: 'all' },
      { include_detail: true }
    )
    correlatedUpstream.value = res.items || []
  } catch (err) {
    console.error('[OpsErrorDetailModal] Failed to load correlated upstream errors', err)
    correlatedUpstream.value = []
  } finally {
    correlatedUpstreamLoading.value = false
  }
}

function close() {
  emit('update:show', false)
}

function prettyJSON(raw?: string): string {
  if (!raw) return 'N/A'
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

async function fetchDetail(id: number) {
  loading.value = true
  try {
    const kind = props.errorType || (detail.value?.phase === 'upstream' ? 'upstream' : 'request')
    const d = kind === 'upstream' ? await opsAPI.getUpstreamErrorDetail(id) : await opsAPI.getRequestErrorDetail(id)
    detail.value = d
  } catch (err: any) {
    detail.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadErrorDetail'))
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.show, props.errorId] as const,
  ([show, id]) => {
    if (!show) {
      detail.value = null
      return
    }
    if (typeof id === 'number' && id > 0) {
      expandedUpstreamDetailIds.value = new Set()
      fetchDetail(id)
      if (props.errorType === 'request') {
        fetchCorrelatedUpstreamErrors(id)
      } else {
        correlatedUpstream.value = []
      }
    }
  },
  { immediate: true }
)

const statusClass = computed(() => {
  const code = detail.value?.status_code ?? 0
  if (code >= 500) return 'od-badge-bad'
  if (code === 429) return 'od-badge-warn'
  if (code >= 400) return 'od-badge-warn'
  return 'od-badge-dim'
})

</script>
