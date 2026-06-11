<template>
  <div style="display:flex;height:100%;min-height:0;flex-direction:column;background:var(--bg-1,#13161C);">
    <div v-if="loading" style="display:flex;flex:1;align-items:center;justify-content:center;padding:28px 0;" role="status" aria-label="加载中">
      <div style="width:24px;height:24px;border-radius:50%;border-bottom:2px solid var(--ops-azure,#5CA8FF);" class="animate-spin" aria-hidden="true"></div>
    </div>

    <div v-else style="display:flex;min-height:0;flex:1;flex-direction:column;">
      <div style="min-height:0;flex:1;overflow:auto;border-bottom:1px solid var(--line-0,#20242C);">
        <table style="width:100%;border-collapse:collapse;font-size:11.5px;">
          <thead class="od-table-head-row" style="position:sticky;top:0;z-index:10;">
            <tr>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.time') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.type') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.endpoint') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.platform') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.model') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.group') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.user') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.apiKey') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.account') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.status') }}</th>
              <th style="padding:7px 12px;text-align:left;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.message') }}</th>
              <th style="padding:7px 12px;text-align:right;border-bottom:1px solid var(--line-0,#20242C);" class="od-sys-label">{{ t('admin.ops.errorLog.action') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="rows.length === 0">
              <td colspan="12" style="padding:36px;text-align:center;font-size:13px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.errorLog.noErrors') }}</td>
            </tr>
            <tr v-for="log in rows" :key="log.id" class="od-tr-border" style="cursor:pointer;" @click="emit('openErrorDetail', log.id)">
              <!-- Time -->
              <td style="padding:7px 12px;white-space:nowrap;">
                <el-tooltip :content="log.request_id || log.client_request_id" placement="top" :show-after="500">
                  <span class="od-mono" style="font-size:11px;color:var(--ink-0,#E8EBF0);">{{ formatDateTime(log.created_at).split(' ')[1] }}</span>
                </el-tooltip>
              </td>
              <!-- Type -->
              <td style="padding:7px 12px;white-space:nowrap;">
                <span :class="getTypeBadge(log).className">{{ getTypeBadge(log).label }}</span>
              </td>
              <!-- Endpoint -->
              <td style="padding:7px 12px;">
                <div style="max-width:160px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">
                  <el-tooltip v-if="log.inbound_endpoint" :content="formatEndpointTooltip(log)" placement="top" :show-after="500">
                    <span class="od-mono" style="font-size:10.5px;color:var(--ink-1,#97A0AF);">{{ log.inbound_endpoint }}</span>
                  </el-tooltip>
                  <span v-else style="color:var(--ink-2,#5C6470);">-</span>
                </div>
              </td>
              <!-- Platform -->
              <td style="padding:7px 12px;white-space:nowrap;">
                <span class="od-badge od-badge-dim" style="text-transform:uppercase;font-size:9.5px;">{{ log.platform || '-' }}</span>
              </td>
              <!-- Model -->
              <td style="padding:7px 12px;">
                <div style="max-width:160px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">
                  <template v-if="hasModelMapping(log)">
                    <el-tooltip :content="modelMappingTooltip(log)" placement="top" :show-after="500">
                      <span class="od-mono" style="font-size:10.5px;color:var(--ink-1,#97A0AF);display:flex;align-items:center;gap:3px;">
                        <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">{{ log.requested_model }}</span>
                        <span style="color:var(--ink-2,#5C6470);flex-shrink:0;">→</span>
                        <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--ops-azure,#5CA8FF);">{{ log.upstream_model }}</span>
                      </span>
                    </el-tooltip>
                  </template>
                  <template v-else>
                    <span v-if="displayModel(log)" class="od-mono" style="font-size:10.5px;color:var(--ink-1,#97A0AF);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;display:block;" :title="displayModel(log)">{{ displayModel(log) }}</span>
                    <span v-else style="color:var(--ink-2,#5C6470);">-</span>
                  </template>
                </div>
              </td>
              <!-- Group -->
              <td style="padding:7px 12px;">
                <el-tooltip v-if="log.group_id" :content="t('admin.ops.errorLog.id') + ' ' + log.group_id" placement="top" :show-after="500">
                  <span style="max-width:100px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;display:block;font-size:11.5px;font-weight:500;color:var(--ink-0,#E8EBF0);">{{ log.group_name || '-' }}</span>
                </el-tooltip>
                <span v-else style="color:var(--ink-2,#5C6470);">-</span>
              </td>
              <!-- User -->
              <td style="padding:7px 12px;">
                <el-tooltip v-if="log.user_id" :content="t('admin.ops.errorLog.userId') + ' ' + log.user_id" placement="top" :show-after="500">
                  <span style="display:block;max-width:140px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:11.5px;font-weight:500;color:var(--ink-0,#E8EBF0);">{{ log.user_email || '-' }}</span>
                </el-tooltip>
                <span v-else style="color:var(--ink-2,#5C6470);">-</span>
              </td>
              <!-- API Key -->
              <td style="padding:7px 12px;">
                <div v-if="log.api_key_id || log.api_key_name" style="display:flex;max-width:140px;align-items:center;gap:4px;">
                  <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:11.5px;font-weight:500;color:var(--ink-0,#E8EBF0);" :title="log.api_key_name || ('#' + log.api_key_id)">{{ log.api_key_name || ('#' + log.api_key_id) }}</span>
                  <span v-if="log.api_key_deleted" class="od-badge od-badge-bad" style="flex-shrink:0;font-size:9px;">{{ t('admin.ops.errorLog.keyDeletedBadge') }}</span>
                </div>
                <span v-else style="color:var(--ink-2,#5C6470);">-</span>
              </td>
              <!-- Account -->
              <td style="padding:7px 12px;">
                <el-tooltip v-if="log.account_id" :content="t('admin.ops.errorLog.accountId') + ' ' + log.account_id" placement="top" :show-after="500">
                  <span style="display:block;max-width:120px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:11.5px;font-weight:500;color:var(--ink-0,#E8EBF0);">{{ log.account_name || '-' }}</span>
                </el-tooltip>
                <span v-else style="color:var(--ink-2,#5C6470);">-</span>
              </td>
              <!-- Status -->
              <td style="padding:7px 12px;white-space:nowrap;">
                <div style="display:flex;align-items:center;gap:4px;">
                  <span :class="getStatusClass(log.status_code)">{{ log.status_code }}</span>
                  <span v-if="log.severity" :class="['od-badge', getSeverityClass(log.severity)]">{{ log.severity }}</span>
                  <span v-if="log.request_type != null && log.request_type > 0" class="od-badge od-badge-dim">{{ formatRequestType(log.request_type) }}</span>
                </div>
              </td>
              <!-- Message -->
              <td style="padding:7px 12px;">
                <div style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">
                  <p style="font-size:10.5px;font-weight:500;color:var(--ink-2,#5C6470);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;" :title="log.message">{{ formatSmartMessage(log.message) || '-' }}</p>
                </div>
              </td>
              <!-- Actions -->
              <td style="padding:7px 12px;white-space:nowrap;text-align:right;" @click.stop>
                <button type="button" class="od-btn" style="padding:2px 8px;font-size:10px;color:var(--ops-azure,#5CA8FF);" @click="emit('openErrorDetail', log.id)">{{ t('admin.ops.errorLog.details') }}</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div style="background:var(--bg-2,#171A20);">
        <Pagination v-if="total > 0" :total="total" :page="page" :page-size="pageSize" @update:page="emit('update:page', $event)" @update:pageSize="emit('update:pageSize', $event)" />
      </div>
    </div>
  </div>
</template>

<style src="../ops-quench.css"></style>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Pagination from '@/components/common/Pagination.vue'
import type { OpsErrorLog } from '@/api/admin/ops'
import { getSeverityClass, formatDateTime } from '../utils/opsFormatters'

const { t } = useI18n()

function isUpstreamRow(log: OpsErrorLog): boolean {
  const phase = String(log.phase || '').toLowerCase()
  const owner = String(log.error_owner || '').toLowerCase()
  return phase === 'upstream' && owner === 'provider'
}

function formatEndpointTooltip(log: OpsErrorLog): string {
  const parts: string[] = []
  if (log.inbound_endpoint) parts.push(`Inbound: ${log.inbound_endpoint}`)
  if (log.upstream_endpoint) parts.push(`Upstream: ${log.upstream_endpoint}`)
  return parts.join('\n') || ''
}

function hasModelMapping(log: OpsErrorLog): boolean {
  const requested = String(log.requested_model || '').trim()
  const upstream = String(log.upstream_model || '').trim()
  return !!requested && !!upstream && requested !== upstream
}

function modelMappingTooltip(log: OpsErrorLog): string {
  const requested = String(log.requested_model || '').trim()
  const upstream = String(log.upstream_model || '').trim()
  if (!requested && !upstream) return ''
  if (requested && upstream) return `${requested} → ${upstream}`
  return upstream || requested
}

function displayModel(log: OpsErrorLog): string {
  const upstream = String(log.upstream_model || '').trim()
  if (upstream) return upstream
  const requested = String(log.requested_model || '').trim()
  if (requested) return requested
  return String(log.model || '').trim()
}

function formatRequestType(type: number | null | undefined): string {
  switch (type) {
    case 1: return t('admin.ops.errorLog.requestTypeSync')
    case 2: return t('admin.ops.errorLog.requestTypeStream')
    case 3: return t('admin.ops.errorLog.requestTypeWs')
    default: return ''
  }
}

function getTypeBadge(log: OpsErrorLog): { label: string; className: string } {
  const phase = String(log.phase || '').toLowerCase()
  const owner = String(log.error_owner || '').toLowerCase()

  if (isUpstreamRow(log)) {
    return { label: t('admin.ops.errorLog.typeUpstream'), className: 'od-badge od-badge-bad' }
  }
  if (phase === 'request' && owner === 'client') {
    return { label: t('admin.ops.errorLog.typeRequest'), className: 'od-badge od-badge-warn' }
  }
  if (phase === 'auth' && owner === 'client') {
    return { label: t('admin.ops.errorLog.typeAuth'), className: 'od-badge od-badge-azure' }
  }
  if (phase === 'routing' && owner === 'platform') {
    return { label: t('admin.ops.errorLog.typeRouting'), className: 'od-badge od-badge-dim' }
  }
  if (phase === 'internal' && owner === 'platform') {
    return { label: t('admin.ops.errorLog.typeInternal'), className: 'od-badge od-badge-dim' }
  }

  const fallback = phase || owner || t('common.unknown')
  return { label: fallback, className: 'od-badge od-badge-dim' }
}

interface Props {
  rows: OpsErrorLog[]
  total: number
  loading: boolean
  page: number
  pageSize: number
}

interface Emits {
  (e: 'openErrorDetail', id: number): void
  (e: 'update:page', value: number): void
  (e: 'update:pageSize', value: number): void
}

defineProps<Props>()
const emit = defineEmits<Emits>()

function getStatusClass(code: number): string {
  if (code >= 500) return 'od-badge od-badge-bad'
  if (code === 429) return 'od-badge od-badge-warn'
  if (code >= 400) return 'od-badge od-badge-warn'
  return 'od-badge od-badge-dim'
}

function formatSmartMessage(msg: string): string {
  if (!msg) return ''

  if (msg.startsWith('{') || msg.startsWith('[')) {
    try {
      const obj = JSON.parse(msg)
      if (obj?.error?.message) return String(obj.error.message)
      if (obj?.message) return String(obj.message)
      if (obj?.detail) return String(obj.detail)
      if (typeof obj === 'object') return JSON.stringify(obj).substring(0, 150)
    } catch {
      // ignore parse error
    }
  }

  if (msg.includes('context deadline exceeded')) return t('admin.ops.errorLog.commonErrors.contextDeadlineExceeded')
  if (msg.includes('connection refused')) return t('admin.ops.errorLog.commonErrors.connectionRefused')
  if (msg.toLowerCase().includes('rate limit')) return t('admin.ops.errorLog.commonErrors.rateLimit')

  return msg.length > 200 ? msg.substring(0, 200) + '...' : msg

}
</script>
