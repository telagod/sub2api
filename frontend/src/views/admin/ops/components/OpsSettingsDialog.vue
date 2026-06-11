<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { opsAPI } from '@/api/admin/ops'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import type { OpsAlertRuntimeSettings, EmailNotificationConfig, AlertSeverity, OpsAdvancedSettings, OpsMetricThresholds } from '../types'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const loading = ref(false)
const saving = ref(false)

// 运行时设置
const runtimeSettings = ref<OpsAlertRuntimeSettings | null>(null)
// 邮件通知配置
const emailConfig = ref<EmailNotificationConfig | null>(null)
// 高级设置
const advancedSettings = ref<OpsAdvancedSettings | null>(null)
// 指标阈值配置
const metricThresholds = ref<OpsMetricThresholds>({
  sla_percent_min: 99.5,
  ttft_p99_ms_max: 500,
  request_error_rate_percent_max: 5,
  upstream_error_rate_percent_max: 5
})

// 加载所有配置
async function loadAllSettings() {
  loading.value = true
  try {
    const [runtime, email, advanced, thresholds] = await Promise.all([
      opsAPI.getAlertRuntimeSettings(),
      opsAPI.getEmailNotificationConfig(),
      opsAPI.getAdvancedSettings(),
      opsAPI.getMetricThresholds()
    ])
    runtimeSettings.value = runtime
    emailConfig.value = email
    advancedSettings.value = advanced
    // 兼容旧 payload：后端未返回该字段时补默认值，保证表单可绑定
    if (advancedSettings.value && !advancedSettings.value.openai_account_quota_auto_pause) {
      advancedSettings.value.openai_account_quota_auto_pause = { default_threshold_5h: 0, default_threshold_7d: 0 }
    }
    // 如果后端返回了阈值，使用后端的值；否则保持默认值
    if (thresholds && Object.keys(thresholds).length > 0) {
        metricThresholds.value = {
          sla_percent_min: thresholds.sla_percent_min ?? 99.5,
          ttft_p99_ms_max: thresholds.ttft_p99_ms_max ?? 500,
          request_error_rate_percent_max: thresholds.request_error_rate_percent_max ?? 5,
          upstream_error_rate_percent_max: thresholds.upstream_error_rate_percent_max ?? 5
        }
    }
  } catch (err: any) {
    console.error('[OpsSettingsDialog] Failed to load settings', err)
    appStore.showError(err?.response?.data?.detail || t('admin.ops.settings.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 监听弹窗打开
watch(() => props.show, (show) => {
  if (show) {
    loadAllSettings()
  }
})

// 邮件输入
const alertRecipientInput = ref('')
const reportRecipientInput = ref('')

// 严重级别选项
const severityOptions: Array<{ value: AlertSeverity | ''; label: string }> = [
  { value: '', label: t('admin.ops.email.minSeverityAll') },
  { value: 'critical', label: t('common.critical') },
  { value: 'warning', label: t('common.warning') },
  { value: 'info', label: t('common.info') }
]

// 验证邮箱
function isValidEmailAddress(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

// 添加收件人
function addRecipient(target: 'alert' | 'report') {
  if (!emailConfig.value) return
  const raw = (target === 'alert' ? alertRecipientInput.value : reportRecipientInput.value).trim()
  if (!raw) return

  if (!isValidEmailAddress(raw)) {
    appStore.showError(t('common.invalidEmail'))
    return
  }

  const normalized = raw.toLowerCase()
  const list = target === 'alert' ? emailConfig.value.alert.recipients : emailConfig.value.report.recipients
  if (!list.includes(normalized)) {
    list.push(normalized)
  }
  if (target === 'alert') alertRecipientInput.value = ''
  else reportRecipientInput.value = ''
}

// 移除收件人
function removeRecipient(target: 'alert' | 'report', email: string) {
  if (!emailConfig.value) return
  const list = target === 'alert' ? emailConfig.value.alert.recipients : emailConfig.value.report.recipients
  const idx = list.indexOf(email)
  if (idx >= 0) list.splice(idx, 1)
}

// OpenAI 账号配额自动暂停：后端按 0~1 分数存储，UI 按百分比(0~100)展示
const quotaAutoPause5hPercent = computed<number | null>({
  get() {
    const v = advancedSettings.value?.openai_account_quota_auto_pause?.default_threshold_5h
    return v && v > 0 ? Math.round(v * 1000) / 10 : null
  },
  set(val) {
    if (!advancedSettings.value?.openai_account_quota_auto_pause) return
    advancedSettings.value.openai_account_quota_auto_pause.default_threshold_5h = val != null && val > 0 ? val / 100 : 0
  }
})
const quotaAutoPause7dPercent = computed<number | null>({
  get() {
    const v = advancedSettings.value?.openai_account_quota_auto_pause?.default_threshold_7d
    return v && v > 0 ? Math.round(v * 1000) / 10 : null
  },
  set(val) {
    if (!advancedSettings.value?.openai_account_quota_auto_pause) return
    advancedSettings.value.openai_account_quota_auto_pause.default_threshold_7d = val != null && val > 0 ? val / 100 : 0
  }
})

// 验证
const validation = computed(() => {
  const errors: string[] = []

  // 验证运行时设置
  if (runtimeSettings.value) {
    const evalSeconds = runtimeSettings.value.evaluation_interval_seconds
    if (!Number.isFinite(evalSeconds) || evalSeconds < 1 || evalSeconds > 86400) {
      errors.push(t('admin.ops.runtime.validation.evalIntervalRange'))
    }
  }

  // 邮件配置: 启用但无收件人时不阻断保存, 保存时会自动禁用

  // 验证高级设置
  if (advancedSettings.value) {
    const { error_log_retention_days, minute_metrics_retention_days, hourly_metrics_retention_days } = advancedSettings.value.data_retention
    if (error_log_retention_days < 0 || error_log_retention_days > 365) {
      errors.push(t('admin.ops.settings.validation.retentionDaysRange'))
    }
    if (minute_metrics_retention_days < 0 || minute_metrics_retention_days > 365) {
      errors.push(t('admin.ops.settings.validation.retentionDaysRange'))
    }
    if (hourly_metrics_retention_days < 0 || hourly_metrics_retention_days > 365) {
      errors.push(t('admin.ops.settings.validation.retentionDaysRange'))
    }

    const { default_threshold_5h, default_threshold_7d } = advancedSettings.value.openai_account_quota_auto_pause
    if (default_threshold_5h < 0 || default_threshold_5h > 1 || default_threshold_7d < 0 || default_threshold_7d > 1) {
      errors.push(t('admin.ops.settings.validation.openaiQuotaAutoPauseRange'))
    }
  }

  // 验证指标阈值
  if (metricThresholds.value.sla_percent_min != null && (metricThresholds.value.sla_percent_min < 0 || metricThresholds.value.sla_percent_min > 100)) {
    errors.push(t('admin.ops.settings.validation.slaMinPercentRange'))
  }
  if (metricThresholds.value.ttft_p99_ms_max != null && metricThresholds.value.ttft_p99_ms_max < 0) {
    errors.push(t('admin.ops.settings.validation.ttftP99MaxRange'))
  }
  if (metricThresholds.value.request_error_rate_percent_max != null && (metricThresholds.value.request_error_rate_percent_max < 0 || metricThresholds.value.request_error_rate_percent_max > 100)) {
    errors.push(t('admin.ops.settings.validation.requestErrorRateMaxRange'))
  }
  if (metricThresholds.value.upstream_error_rate_percent_max != null && (metricThresholds.value.upstream_error_rate_percent_max < 0 || metricThresholds.value.upstream_error_rate_percent_max > 100)) {
    errors.push(t('admin.ops.settings.validation.upstreamErrorRateMaxRange'))
  }

  return { valid: errors.length === 0, errors }
})

// 保存所有配置
async function saveAllSettings() {
  if (!validation.value.valid) {
    appStore.showError(validation.value.errors[0])
    return
  }

  saving.value = true
  try {
    // 无收件人时自动禁用邮件通知
    if (emailConfig.value) {
      if (emailConfig.value.alert.enabled && emailConfig.value.alert.recipients.length === 0) {
        emailConfig.value.alert.enabled = false
      }
      if (emailConfig.value.report.enabled && emailConfig.value.report.recipients.length === 0) {
        emailConfig.value.report.enabled = false
      }
    }
    await Promise.all([
      runtimeSettings.value ? opsAPI.updateAlertRuntimeSettings(runtimeSettings.value) : Promise.resolve(),
      emailConfig.value ? opsAPI.updateEmailNotificationConfig(emailConfig.value) : Promise.resolve(),
      advancedSettings.value ? opsAPI.updateAdvancedSettings(advancedSettings.value) : Promise.resolve(),
      opsAPI.updateMetricThresholds(metricThresholds.value)
    ])
    appStore.showSuccess(t('admin.ops.settings.saveSuccess'))
    emit('saved')
    emit('close')
  } catch (err: any) {
    console.error('[OpsSettingsDialog] Failed to save settings', err)
    appStore.showError(err?.response?.data?.message || err?.response?.data?.detail || t('admin.ops.settings.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <BaseDialog :show="show" :title="t('admin.ops.settings.title')" width="extra-wide" @close="emit('close')">
    <div v-if="loading" style="padding:36px;text-align:center;font-size:13px;color:var(--ink-2,#5C6470);">{{ t('common.loading') }}</div>

    <div v-else-if="runtimeSettings && emailConfig && advancedSettings" style="display:flex;flex-direction:column;gap:16px;">
      <div v-if="!validation.valid" style="border-radius:8px;border:1px solid var(--ops-warn-border,rgba(224,179,78,.25));background:var(--ops-warn-dim,rgba(224,179,78,.08));padding:10px 14px;font-size:11.5px;color:var(--ops-warn,#E0B34E);">
        <div style="font-weight:700;">{{ t('admin.ops.settings.validation.title') }}</div>
        <ul style="margin-top:4px;padding-left:16px;">
          <li v-for="msg in validation.errors" :key="msg">{{ msg }}</li>
        </ul>
      </div>

      <!-- 数据采集频率 -->
      <div class="od-card" style="padding:14px;">
        <h4 style="font-size:13px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:10px;">{{ t('admin.ops.settings.dataCollection') }}</h4>
        <div>
          <label class="od-form-label">{{ t('admin.ops.settings.evaluationInterval') }}</label>
          <input v-model.number="runtimeSettings.evaluation_interval_seconds" type="number" min="1" max="86400" class="input" />
          <p style="margin-top:3px;font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.evaluationIntervalHint') }}</p>
        </div>
      </div>

      <!-- 预警配置 -->
      <div class="od-card" style="padding:14px;">
        <h4 style="font-size:13px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:12px;">{{ t('admin.ops.settings.alertConfig') }}</h4>
        <div style="display:flex;flex-direction:column;gap:12px;">
          <div style="display:flex;align-items:center;justify-content:space-between;">
            <label style="font-size:13px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.enableAlert') }}</label>
            <Toggle v-model="emailConfig.alert.enabled" />
          </div>
          <div v-if="emailConfig.alert.enabled">
            <label class="od-form-label">{{ t('admin.ops.settings.alertRecipients') }}</label>
            <div style="display:flex;gap:6px;">
              <input v-model="alertRecipientInput" type="email" class="input" :placeholder="t('admin.ops.settings.emailPlaceholder')" @keydown.enter.prevent="addRecipient('alert')" />
              <button class="od-btn" style="padding:4px 10px;font-size:11px;white-space:nowrap;" type="button" @click="addRecipient('alert')">{{ t('common.add') }}</button>
            </div>
            <div style="margin-top:6px;display:flex;flex-wrap:wrap;gap:6px;">
              <span v-for="email in emailConfig.alert.recipients" :key="email" class="od-badge od-badge-azure" style="gap:6px;">
                {{ email }}
                <button type="button" style="color:inherit;opacity:.7;background:none;border:none;cursor:pointer;padding:0;" :aria-label="`移除 ${email}`" @click="removeRecipient('alert', email)">×</button>
              </span>
            </div>
            <p style="margin-top:4px;font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.recipientsHint') }}</p>
          </div>
          <div v-if="emailConfig.alert.enabled">
            <label class="od-form-label">{{ t('admin.ops.settings.minSeverity') }}</label>
            <Select v-model="emailConfig.alert.min_severity" :options="severityOptions" />
          </div>
        </div>
      </div>

      <!-- 评估报告配置 -->
      <div class="od-card" style="padding:14px;">
        <h4 style="font-size:13px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:12px;">{{ t('admin.ops.settings.reportConfig') }}</h4>
        <div style="display:flex;flex-direction:column;gap:12px;">
          <div style="display:flex;align-items:center;justify-content:space-between;">
            <label style="font-size:13px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.enableReport') }}</label>
            <Toggle v-model="emailConfig.report.enabled" />
          </div>
          <div v-if="emailConfig.report.enabled">
            <label class="od-form-label">{{ t('admin.ops.settings.reportRecipients') }}</label>
            <div style="display:flex;gap:6px;">
              <input v-model="reportRecipientInput" type="email" class="input" :placeholder="t('admin.ops.settings.emailPlaceholder')" @keydown.enter.prevent="addRecipient('report')" />
              <button class="od-btn" style="padding:4px 10px;font-size:11px;white-space:nowrap;" type="button" @click="addRecipient('report')">{{ t('common.add') }}</button>
            </div>
            <div style="margin-top:6px;display:flex;flex-wrap:wrap;gap:6px;">
              <span v-for="email in emailConfig.report.recipients" :key="email" class="od-badge od-badge-azure" style="gap:6px;">
                {{ email }}
                <button type="button" style="color:inherit;opacity:.7;background:none;border:none;cursor:pointer;padding:0;" :aria-label="`移除 ${email}`" @click="removeRecipient('report', email)">×</button>
              </span>
            </div>
            <p style="margin-top:4px;font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.recipientsHint') }}</p>
          </div>
          <div v-if="emailConfig.report.enabled" style="display:grid;grid-template-columns:1fr 1fr;gap:10px;">
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.dailySummary') }}</label>
              <Toggle v-model="emailConfig.report.daily_summary_enabled" />
            </div>
            <div v-if="emailConfig.report.daily_summary_enabled">
              <input v-model="emailConfig.report.daily_summary_schedule" type="text" class="input" placeholder="0 9 * * *" />
            </div>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.weeklySummary') }}</label>
              <Toggle v-model="emailConfig.report.weekly_summary_enabled" />
            </div>
            <div v-if="emailConfig.report.weekly_summary_enabled">
              <input v-model="emailConfig.report.weekly_summary_schedule" type="text" class="input" placeholder="0 9 * * 1" />
            </div>
          </div>
        </div>
      </div>

      <!-- 指标阈值 -->
      <div class="od-card" style="padding:14px;">
        <h4 style="font-size:13px;font-weight:600;color:var(--ink-0,#E8EBF0);margin-bottom:4px;">{{ t('admin.ops.settings.metricThresholds') }}</h4>
        <p style="font-size:11px;color:var(--ink-2,#5C6470);margin-bottom:12px;">{{ t('admin.ops.settings.metricThresholdsHint') }}</p>
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;">
          <div>
            <label class="od-form-label">{{ t('admin.ops.settings.slaMinPercent') }}</label>
            <input v-model.number="metricThresholds.sla_percent_min" type="number" min="0" max="100" step="0.1" class="input" />
            <p style="margin-top:3px;font-size:10px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.slaMinPercentHint') }}</p>
          </div>
          <div>
            <label class="od-form-label">{{ t('admin.ops.settings.ttftP99MaxMs') }}</label>
            <input v-model.number="metricThresholds.ttft_p99_ms_max" type="number" min="0" step="50" class="input" />
            <p style="margin-top:3px;font-size:10px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.ttftP99MaxMsHint') }}</p>
          </div>
          <div>
            <label class="od-form-label">{{ t('admin.ops.settings.requestErrorRateMaxPercent') }}</label>
            <input v-model.number="metricThresholds.request_error_rate_percent_max" type="number" min="0" max="100" step="0.1" class="input" />
            <p style="margin-top:3px;font-size:10px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.requestErrorRateMaxPercentHint') }}</p>
          </div>
          <div>
            <label class="od-form-label">{{ t('admin.ops.settings.upstreamErrorRateMaxPercent') }}</label>
            <input v-model.number="metricThresholds.upstream_error_rate_percent_max" type="number" min="0" max="100" step="0.1" class="input" />
            <p style="margin-top:3px;font-size:10px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.upstreamErrorRateMaxPercentHint') }}</p>
          </div>
        </div>
      </div>

      <!-- 高级设置 -->
      <details class="od-card" style="padding:0;">
        <summary style="cursor:pointer;padding:14px;font-size:13px;font-weight:600;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.advancedSettings') }}</summary>
        <div style="padding:0 14px 14px;display:flex;flex-direction:column;gap:16px;">
          <!-- 数据保留策略 -->
          <div style="display:flex;flex-direction:column;gap:8px;">
            <h5 style="font-size:11.5px;font-weight:600;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.dataRetention') }}</h5>
            <div style="display:flex;align-items:center;justify-content:space-between;">
              <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.enableCleanup') }}</label>
              <Toggle v-model="advancedSettings.data_retention.cleanup_enabled" />
            </div>
            <div v-if="advancedSettings.data_retention.cleanup_enabled">
              <label class="od-form-label">{{ t('admin.ops.settings.cleanupSchedule') }}</label>
              <input v-model="advancedSettings.data_retention.cleanup_schedule" type="text" class="input" placeholder="0 2 * * *" />
              <p style="margin-top:3px;font-size:10px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.cleanupScheduleHint') }}</p>
            </div>
            <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:10px;">
              <div>
                <label class="od-form-label">{{ t('admin.ops.settings.errorLogRetentionDays') }}</label>
                <input v-model.number="advancedSettings.data_retention.error_log_retention_days" type="number" min="0" max="365" class="input" />
              </div>
              <div>
                <label class="od-form-label">{{ t('admin.ops.settings.minuteMetricsRetentionDays') }}</label>
                <input v-model.number="advancedSettings.data_retention.minute_metrics_retention_days" type="number" min="0" max="365" class="input" />
              </div>
              <div>
                <label class="od-form-label">{{ t('admin.ops.settings.hourlyMetricsRetentionDays') }}</label>
                <input v-model.number="advancedSettings.data_retention.hourly_metrics_retention_days" type="number" min="0" max="365" class="input" />
              </div>
            </div>
            <p style="font-size:11px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.retentionDaysHint') }}</p>
          </div>

          <!-- 预聚合 -->
          <div style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;">
            <div>
              <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.enableAggregation') }}</label>
              <p style="margin-top:2px;font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.aggregationHint') }}</p>
            </div>
            <Toggle v-model="advancedSettings.aggregation.aggregation_enabled" />
          </div>

          <!-- OpenAI 配额 -->
          <div style="display:flex;flex-direction:column;gap:8px;">
            <h5 style="font-size:11.5px;font-weight:600;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.openaiQuotaAutoPause') }}</h5>
            <p style="font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.openaiQuotaAutoPauseHint') }}</p>
            <div style="display:grid;grid-template-columns:1fr 1fr;gap:10px;">
              <div>
                <label class="od-form-label">{{ t('admin.ops.settings.openaiQuotaAutoPauseDefault5h') }}</label>
                <input v-model.number="quotaAutoPause5hPercent" type="number" min="0" max="100" step="0.1" class="input" data-testid="ops-quota-auto-pause-5h" />
              </div>
              <div>
                <label class="od-form-label">{{ t('admin.ops.settings.openaiQuotaAutoPauseDefault7d') }}</label>
                <input v-model.number="quotaAutoPause7dPercent" type="number" min="0" max="100" step="0.1" class="input" data-testid="ops-quota-auto-pause-7d" />
              </div>
            </div>
            <p style="font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.openaiQuotaAutoPauseThresholdHint') }}</p>
          </div>

          <!-- 错误过滤 -->
          <div style="display:flex;flex-direction:column;gap:10px;">
            <h5 style="font-size:11.5px;font-weight:600;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.errorFiltering') }}</h5>
            <div v-for="(key, idx) in (['ignore_count_tokens_errors','ignore_context_canceled','ignore_no_available_accounts','ignore_invalid_api_key_errors','ignore_insufficient_balance_errors'] as const)" :key="idx" style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;">
              <div>
                <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t(`admin.ops.settings.${key}`) }}</label>
                <p style="margin-top:2px;font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t(`admin.ops.settings.${key}Hint`) }}</p>
              </div>
              <Toggle v-model="(advancedSettings as any)[key]" />
            </div>
          </div>

          <!-- Auto Refresh -->
          <div style="display:flex;flex-direction:column;gap:8px;">
            <h5 style="font-size:11.5px;font-weight:600;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.autoRefresh') }}</h5>
            <div style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;">
              <div>
                <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.enableAutoRefresh') }}</label>
                <p style="margin-top:2px;font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.enableAutoRefreshHint') }}</p>
              </div>
              <Toggle v-model="advancedSettings.auto_refresh_enabled" />
            </div>
            <div v-if="advancedSettings.auto_refresh_enabled">
              <label class="od-form-label">{{ t('admin.ops.settings.refreshInterval') }}</label>
              <Select v-model="advancedSettings.auto_refresh_interval_seconds" :options="[{value:15,label:t('admin.ops.settings.refreshInterval15s')},{value:30,label:t('admin.ops.settings.refreshInterval30s')},{value:60,label:t('admin.ops.settings.refreshInterval60s')}]" />
            </div>
          </div>

          <!-- Dashboard Cards -->
          <div style="display:flex;flex-direction:column;gap:10px;">
            <h5 style="font-size:11.5px;font-weight:600;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.dashboardCards') }}</h5>
            <div style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;">
              <div>
                <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.displayAlertEvents') }}</label>
                <p style="margin-top:2px;font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.displayAlertEventsHint') }}</p>
              </div>
              <Toggle v-model="advancedSettings.display_alert_events" />
            </div>
            <div style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;">
              <div>
                <label style="font-size:12px;color:var(--ink-1,#97A0AF);">{{ t('admin.ops.settings.displayOpenAITokenStats') }}</label>
                <p style="margin-top:2px;font-size:10.5px;color:var(--ink-2,#5C6470);">{{ t('admin.ops.settings.displayOpenAITokenStatsHint') }}</p>
              </div>
              <Toggle v-model="advancedSettings.display_openai_token_stats" />
            </div>
          </div>
        </div>
      </details>
    </div>

    <template #footer>
      <div style="display:flex;justify-content:flex-end;gap:8px;">
        <button class="od-btn" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button class="od-btn od-btn-azure" :disabled="saving || !validation.valid" @click="saveAllSettings">{{ saving ? t('common.saving') : t('common.save') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<style src="../ops-quench.css"></style>
