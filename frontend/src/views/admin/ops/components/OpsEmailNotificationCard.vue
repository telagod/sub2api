<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsAPI } from '@/api/admin/ops'
import type { EmailNotificationConfig, AlertSeverity } from '../types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const config = ref<EmailNotificationConfig | null>(null)

const showEditor = ref(false)
const saving = ref(false)
const draft = ref<EmailNotificationConfig | null>(null)
const alertRecipientInput = ref('')
const reportRecipientInput = ref('')
const alertRecipientError = ref('')
const reportRecipientError = ref('')

const severityOptions: Array<{ value: AlertSeverity | ''; label: string }> = [
  { value: '', label: t('admin.ops.email.minSeverityAll') },
  { value: 'critical', label: t('common.critical') },
  { value: 'warning', label: t('common.warning') },
  { value: 'info', label: t('common.info') }
]

async function loadConfig() {
  loading.value = true
  try {
    const data = await opsAPI.getEmailNotificationConfig()
    config.value = data
  } catch (err: any) {
    console.error('[OpsEmailNotificationCard] Failed to load config', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.email.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  if (!draft.value) return
  if (!editorValidation.value.valid) {
    appStore.showError(editorValidation.value.errors[0] || t('admin.ops.email.validation.invalid'))
    return
  }
  saving.value = true
  try {
    config.value = await opsAPI.updateEmailNotificationConfig(draft.value)
    showEditor.value = false
    appStore.showSuccess(t('admin.ops.email.saveSuccess'))
  } catch (err: any) {
    console.error('[OpsEmailNotificationCard] Failed to save config', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.email.saveFailed'))
  } finally {
    saving.value = false
  }
}

function openEditor() {
  if (!config.value) return
  draft.value = JSON.parse(JSON.stringify(config.value))
  alertRecipientInput.value = ''
  reportRecipientInput.value = ''
  alertRecipientError.value = ''
  reportRecipientError.value = ''
  showEditor.value = true
}

function isValidEmailAddress(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

function isNonNegativeNumber(value: unknown): boolean {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0
}

function validateCronField(enabled: boolean, cron: string): string | null {
  if (!enabled) return null
  if (!cron || !cron.trim()) return t('admin.ops.email.validation.cronRequired')
  if (cron.trim().split(/\s+/).length < 5) return t('admin.ops.email.validation.cronFormat')
  return null
}

const editorValidation = computed(() => {
  const errors: string[] = []
  if (!draft.value) return { valid: true, errors }

  if (draft.value.alert.enabled && draft.value.alert.recipients.length === 0) {
    errors.push(t('admin.ops.email.validation.alertRecipientsRequired'))
  }
  if (draft.value.report.enabled && draft.value.report.recipients.length === 0) {
    errors.push(t('admin.ops.email.validation.reportRecipientsRequired'))
  }

  const invalidAlertRecipients = draft.value.alert.recipients.filter((e) => !isValidEmailAddress(e))
  if (invalidAlertRecipients.length > 0) errors.push(t('admin.ops.email.validation.invalidRecipients'))

  const invalidReportRecipients = draft.value.report.recipients.filter((e) => !isValidEmailAddress(e))
  if (invalidReportRecipients.length > 0) errors.push(t('admin.ops.email.validation.invalidRecipients'))

  if (!isNonNegativeNumber(draft.value.alert.rate_limit_per_hour)) {
    errors.push(t('admin.ops.email.validation.rateLimitRange'))
  }
  if (
    !isNonNegativeNumber(draft.value.alert.batching_window_seconds) ||
    draft.value.alert.batching_window_seconds > 86400
  ) {
    errors.push(t('admin.ops.email.validation.batchWindowRange'))
  }

  const dailyErr = validateCronField(
    draft.value.report.daily_summary_enabled,
    draft.value.report.daily_summary_schedule
  )
  if (dailyErr) errors.push(dailyErr)
  const weeklyErr = validateCronField(
    draft.value.report.weekly_summary_enabled,
    draft.value.report.weekly_summary_schedule
  )
  if (weeklyErr) errors.push(weeklyErr)
  const digestErr = validateCronField(
    draft.value.report.error_digest_enabled,
    draft.value.report.error_digest_schedule
  )
  if (digestErr) errors.push(digestErr)
  const accErr = validateCronField(
    draft.value.report.account_health_enabled,
    draft.value.report.account_health_schedule
  )
  if (accErr) errors.push(accErr)

  if (!isNonNegativeNumber(draft.value.report.error_digest_min_count)) {
    errors.push(t('admin.ops.email.validation.digestMinCountRange'))
  }

  const thr = draft.value.report.account_health_error_rate_threshold
  if (!(typeof thr === 'number' && Number.isFinite(thr) && thr >= 0 && thr <= 100)) {
    errors.push(t('admin.ops.email.validation.accountHealthThresholdRange'))
  }

  return { valid: errors.length === 0, errors }
})

function addRecipient(target: 'alert' | 'report') {
  if (!draft.value) return
  const raw = (target === 'alert' ? alertRecipientInput.value : reportRecipientInput.value).trim()
  if (!raw) return

  if (!isValidEmailAddress(raw)) {
    const msg = t('common.invalidEmail')
    if (target === 'alert') alertRecipientError.value = msg
    else reportRecipientError.value = msg
    return
  }

  const normalized = raw.toLowerCase()
  const list = target === 'alert' ? draft.value.alert.recipients : draft.value.report.recipients
  if (!list.includes(normalized)) {
    list.push(normalized)
  }
  if (target === 'alert') alertRecipientInput.value = ''
  else reportRecipientInput.value = ''
  if (target === 'alert') alertRecipientError.value = ''
  else reportRecipientError.value = ''
}

function removeRecipient(target: 'alert' | 'report', email: string) {
  if (!draft.value) return
  const list = target === 'alert' ? draft.value.alert.recipients : draft.value.report.recipients
  const idx = list.indexOf(email)
  if (idx >= 0) list.splice(idx, 1)
}

onMounted(() => {
  loadConfig()
})
</script>

<template>
  <div class="od-card od-card-pad">
    <div style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;margin-bottom:14px;">
      <div>
        <h3 class="od-chart-title">{{ t('admin.ops.email.title') }}</h3>
        <p style="margin-top:3px;font-size:11.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.email.description') }}</p>
      </div>
      <div style="display:flex;align-items:center;gap:6px;">
        <button class="od-btn od-btn-icon" :disabled="loading" @click="loadConfig">
          <svg width="13" height="13" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
          {{ t('common.refresh') }}
        </button>
        <button class="od-btn" style="padding:4px 10px;font-size:11px;" :disabled="!config" @click="openEditor">{{ t('common.edit') }}</button>
      </div>
    </div>

    <div v-if="!config" style="font-size:13px;color:var(--ink-2,#5C6470);">
      <span v-if="loading">{{ t('admin.ops.email.loading') }}</span>
      <span v-else>{{ t('admin.ops.email.noData') }}</span>
    </div>

    <div v-else style="display:flex;flex-direction:column;gap:12px;">
      <div class="od-sys-card">
        <div style="font-size:12px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:8px;">{{ t('admin.ops.email.alertTitle') }}</div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">
          <div class="od-sys-card">
            <div class="od-sys-label">{{ t('common.enabled') }}</div>
            <div class="od-sys-val">{{ config.alert.enabled ? t('common.enabled') : t('common.disabled') }}</div>
          </div>
          <div class="od-sys-card">
            <div class="od-sys-label">{{ t('admin.ops.email.recipients') }}</div>
            <div class="od-sys-val">{{ config.alert.recipients.length }}</div>
          </div>
          <div class="od-sys-card">
            <div class="od-sys-label">{{ t('admin.ops.email.minSeverity') }}</div>
            <div class="od-sys-val">{{ config.alert.min_severity || t('admin.ops.email.minSeverityAll') }}</div>
          </div>
          <div class="od-sys-card">
            <div class="od-sys-label">{{ t('admin.ops.email.rateLimitPerHour') }}</div>
            <div class="od-sys-val">{{ config.alert.rate_limit_per_hour }}</div>
          </div>
        </div>
      </div>
      <div class="od-sys-card">
        <div style="font-size:12px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:8px;">{{ t('admin.ops.email.reportTitle') }}</div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">
          <div class="od-sys-card">
            <div class="od-sys-label">{{ t('common.enabled') }}</div>
            <div class="od-sys-val">{{ config.report.enabled ? t('common.enabled') : t('common.disabled') }}</div>
          </div>
          <div class="od-sys-card">
            <div class="od-sys-label">{{ t('admin.ops.email.recipients') }}</div>
            <div class="od-sys-val">{{ config.report.recipients.length }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <BaseDialog :show="showEditor" :title="t('admin.ops.email.title')" width="extra-wide" @close="showEditor = false">
    <div v-if="draft" style="display:flex;flex-direction:column;gap:16px;">
      <div v-if="!editorValidation.valid" style="border-radius:8px;border:1px solid var(--ops-warn-border,rgba(224,179,78,.25));background:var(--ops-warn-dim,rgba(224,179,78,.08));padding:10px 14px;font-size:11.5px;color:var(--ops-warn,#E0B34E);">
        <div style="font-weight:700;">{{ t('admin.ops.email.validation.title') }}</div>
        <ul style="margin-top:4px;padding-left:16px;">
          <li v-for="msg in editorValidation.errors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <div class="od-card" style="padding:14px;background:var(--bg-2,#171A20);">
        <div style="font-size:13px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:12px;">{{ t('admin.ops.email.alertTitle') }}</div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
          <div>
            <div class="od-form-label">{{ t('common.enabled') }}</div>
            <label style="display:inline-flex;align-items:center;gap:6px;font-size:12px;color:var(--ink-1,#97A0AF);">
              <input v-model="draft.alert.enabled" type="checkbox" />
              {{ draft.alert.enabled ? t('common.enabled') : t('common.disabled') }}
            </label>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.minSeverity') }}</div>
            <Select v-model="draft.alert.min_severity" :options="severityOptions" />
          </div>
          <div style="grid-column:span 2;">
            <div class="od-form-label">{{ t('admin.ops.email.recipients') }}</div>
            <div style="display:flex;gap:6px;">
              <input v-model="alertRecipientInput" type="email" class="input" :placeholder="t('admin.ops.email.recipients')" @keydown.enter.prevent="addRecipient('alert')" />
              <button class="od-btn" style="padding:4px 10px;font-size:11px;white-space:nowrap;" type="button" @click="addRecipient('alert')">{{ t('common.add') }}</button>
            </div>
            <p v-if="alertRecipientError" style="margin-top:3px;font-size:11px;color:var(--ops-bad,#F25C69);">{{ alertRecipientError }}</p>
            <div style="margin-top:6px;display:flex;flex-wrap:wrap;gap:6px;">
              <span v-for="email in draft.alert.recipients" :key="email" class="od-badge od-badge-azure" style="gap:6px;">
                {{ email }}
                <button type="button" style="color:inherit;opacity:.7;background:none;border:none;cursor:pointer;padding:0;line-height:1;" @click="removeRecipient('alert', email)">×</button>
              </span>
            </div>
            <div style="margin-top:4px;font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.email.recipientsHint') }}</div>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.rateLimitPerHour') }}</div>
            <input v-model.number="draft.alert.rate_limit_per_hour" type="number" min="0" max="100000" class="input" />
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.batchWindowSeconds') }}</div>
            <input v-model.number="draft.alert.batching_window_seconds" type="number" min="0" max="86400" class="input" />
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.includeResolved') }}</div>
            <label style="display:inline-flex;align-items:center;gap:6px;font-size:12px;color:var(--ink-1,#97A0AF);">
              <input v-model="draft.alert.include_resolved_alerts" type="checkbox" />
              {{ draft.alert.include_resolved_alerts ? t('common.enabled') : t('common.disabled') }}
            </label>
          </div>
        </div>
      </div>

      <div class="od-card" style="padding:14px;background:var(--bg-2,#171A20);">
        <div style="font-size:13px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:12px;">{{ t('admin.ops.email.reportTitle') }}</div>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
          <div style="grid-column:span 2;">
            <div class="od-form-label">{{ t('common.enabled') }}</div>
            <label style="display:inline-flex;align-items:center;gap:6px;font-size:12px;color:var(--ink-1,#97A0AF);">
              <input v-model="draft.report.enabled" type="checkbox" />
              {{ draft.report.enabled ? t('common.enabled') : t('common.disabled') }}
            </label>
          </div>
          <div style="grid-column:span 2;">
            <div class="od-form-label">{{ t('admin.ops.email.recipients') }}</div>
            <div style="display:flex;gap:6px;">
              <input v-model="reportRecipientInput" type="email" class="input" :placeholder="t('admin.ops.email.recipients')" @keydown.enter.prevent="addRecipient('report')" />
              <button class="od-btn" style="padding:4px 10px;font-size:11px;white-space:nowrap;" type="button" @click="addRecipient('report')">{{ t('common.add') }}</button>
            </div>
            <p v-if="reportRecipientError" style="margin-top:3px;font-size:11px;color:var(--ops-bad,#F25C69);">{{ reportRecipientError }}</p>
            <div style="margin-top:6px;display:flex;flex-wrap:wrap;gap:6px;">
              <span v-for="email in draft.report.recipients" :key="email" class="od-badge od-badge-azure" style="gap:6px;">
                {{ email }}
                <button type="button" style="color:inherit;opacity:.7;background:none;border:none;cursor:pointer;padding:0;line-height:1;" @click="removeRecipient('report', email)">×</button>
              </span>
            </div>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.dailySummary') }}</div>
            <div style="display:flex;align-items:center;gap:8px;">
              <input v-model="draft.report.daily_summary_enabled" type="checkbox" />
              <input v-model="draft.report.daily_summary_schedule" type="text" class="input" :placeholder="t('admin.ops.email.cronPlaceholder')" />
            </div>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.weeklySummary') }}</div>
            <div style="display:flex;align-items:center;gap:8px;">
              <input v-model="draft.report.weekly_summary_enabled" type="checkbox" />
              <input v-model="draft.report.weekly_summary_schedule" type="text" class="input" :placeholder="t('admin.ops.email.cronPlaceholder')" />
            </div>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.errorDigest') }}</div>
            <div style="display:flex;align-items:center;gap:8px;">
              <input v-model="draft.report.error_digest_enabled" type="checkbox" />
              <input v-model="draft.report.error_digest_schedule" type="text" class="input" :placeholder="t('admin.ops.email.cronPlaceholder')" />
            </div>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.errorDigestMinCount') }}</div>
            <input v-model.number="draft.report.error_digest_min_count" type="number" min="0" max="1000000" class="input" />
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.accountHealth') }}</div>
            <div style="display:flex;align-items:center;gap:8px;">
              <input v-model="draft.report.account_health_enabled" type="checkbox" />
              <input v-model="draft.report.account_health_schedule" type="text" class="input" :placeholder="t('admin.ops.email.cronPlaceholder')" />
            </div>
          </div>
          <div>
            <div class="od-form-label">{{ t('admin.ops.email.accountHealthThreshold') }}</div>
            <input v-model.number="draft.report.account_health_error_rate_threshold" type="number" min="0" max="100" step="0.1" class="input" />
          </div>
          <div style="grid-column:span 2;font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.email.reportHint') }}</div>
        </div>
      </div>
    </div>
    <template #footer>
      <div style="display:flex;justify-content:flex-end;gap:8px;">
        <button class="od-btn" @click="showEditor = false">{{ t('common.cancel') }}</button>
        <button class="od-btn od-btn-azure" :disabled="saving || !editorValidation.valid" @click="saveConfig">{{ saving ? t('common.saving') : t('common.save') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<style src="../ops-quench.css"></style>
