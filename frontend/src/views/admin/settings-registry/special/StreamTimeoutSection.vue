<template>
  <div class="st-body">
    <!-- loading -->
    <div v-if="loading" class="st-loading">
      <div class="st-spinner" />
      {{ t('common.loading') }}
    </div>

    <template v-else>
      <!-- enabled switch -->
      <div class="st-row-between">
        <div>
          <label class="st-label">{{ t('admin.settings.streamTimeout.enabled') }}</label>
          <p class="st-hint">{{ t('admin.settings.streamTimeout.enabledHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <!-- expanded fields — only when enabled -->
      <div v-if="form.enabled" class="st-expanded">

        <!-- action select -->
        <div class="st-field">
          <label class="st-field-label">{{ t('admin.settings.streamTimeout.action') }}</label>
          <select v-model="form.action" class="st-select">
            <option value="temp_unsched">{{ t('admin.settings.streamTimeout.actionTempUnsched') }}</option>
            <option value="error">{{ t('admin.settings.streamTimeout.actionError') }}</option>
            <option value="none">{{ t('admin.settings.streamTimeout.actionNone') }}</option>
          </select>
          <p class="st-field-hint">{{ t('admin.settings.streamTimeout.actionHint') }}</p>
        </div>

        <!-- temp_unsched_minutes — gated on action -->
        <div v-if="form.action === 'temp_unsched'" class="st-field">
          <label class="st-field-label">{{ t('admin.settings.streamTimeout.tempUnschedMinutes') }}</label>
          <input
            v-model.number="form.temp_unsched_minutes"
            type="number"
            min="1"
            max="60"
            class="st-input st-input--mono"
          />
          <p class="st-field-hint">{{ t('admin.settings.streamTimeout.tempUnschedMinutesHint') }}</p>
        </div>

        <!-- threshold_count -->
        <div class="st-field">
          <label class="st-field-label">{{ t('admin.settings.streamTimeout.thresholdCount') }}</label>
          <input
            v-model.number="form.threshold_count"
            type="number"
            min="1"
            max="10"
            class="st-input st-input--mono"
          />
          <p class="st-field-hint">{{ t('admin.settings.streamTimeout.thresholdCountHint') }}</p>
        </div>

        <!-- threshold_window_minutes -->
        <div class="st-field">
          <label class="st-field-label">{{ t('admin.settings.streamTimeout.thresholdWindowMinutes') }}</label>
          <input
            v-model.number="form.threshold_window_minutes"
            type="number"
            min="1"
            max="60"
            class="st-input st-input--mono"
          />
          <p class="st-field-hint">{{ t('admin.settings.streamTimeout.thresholdWindowMinutesHint') }}</p>
        </div>
      </div>

      <!-- save -->
      <div class="st-footer">
        <button
          type="button"
          class="st-save-btn"
          :disabled="saving"
          @click="save"
        >
          <svg
            v-if="saving"
            class="st-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)

const form = reactive({
  enabled: true,
  action: 'temp_unsched' as 'temp_unsched' | 'error' | 'none',
  temp_unsched_minutes: 5,
  threshold_count: 3,
  threshold_window_minutes: 10,
})

onMounted(async () => {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getStreamTimeoutSettings()
    Object.assign(form, settings)
  } catch {
    // silent — form stays at defaults
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  try {
    const updated = await adminAPI.settings.updateStreamTimeoutSettings({
      enabled: form.enabled,
      action: form.action,
      temp_unsched_minutes: form.temp_unsched_minutes,
      threshold_count: form.threshold_count,
      threshold_window_minutes: form.threshold_window_minutes,
    })
    Object.assign(form, updated)
    appStore.showSuccess(t('admin.settings.streamTimeout.saved'))
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t('admin.settings.streamTimeout.saveFailed')),
    )
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.st-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.st-loading { display: flex; align-items: center; gap: 8px; color: var(--ink-2, #5C6470); font-size: 13px; }
.st-spinner {
  width: 16px; height: 16px; border-radius: 50%;
  border: 2px solid var(--line-1, #2F3540);
  border-bottom-color: var(--azure, #5CA8FF);
  animation: st-spin .7s linear infinite; flex-shrink: 0;
}
@keyframes st-spin { to { transform: rotate(360deg); } }

.st-row-between { display: flex; align-items: center; justify-content: space-between; gap: 16px; }

.st-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.st-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.st-expanded { border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px; display: flex; flex-direction: column; gap: 14px; }

.st-field { display: flex; flex-direction: column; gap: 4px; }
.st-field-label { font-size: 12px; font-weight: 500; color: var(--ink-1, #97A0AF); }
.st-field-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.st-input {
  width: 8rem; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.st-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.st-input--mono { font-family: var(--font-mono, ui-monospace, monospace); }

.st-select {
  width: 16rem; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  cursor: pointer; transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%235C6470' d='M2 4l4 4 4-4'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
}
.st-select:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }

.st-footer {
  display: flex; justify-content: flex-end;
  border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px;
}

.st-save-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 18px; border-radius: 8px; font-size: 13px; font-weight: 600;
  font-family: inherit; cursor: pointer; user-select: none;
  border: 1px solid var(--azure, #5CA8FF);
  background: linear-gradient(180deg, rgba(92,168,255,.18) 0%, rgba(92,168,255,.08) 100%);
  color: var(--azure, #5CA8FF);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.08), 0 2px 6px rgba(92,168,255,.15);
  transition: background .15s, box-shadow .15s, opacity .15s;
}
.st-save-btn:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(92,168,255,.28) 0%, rgba(92,168,255,.14) 100%);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.1), 0 3px 10px rgba(92,168,255,.25);
}
.st-save-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.st-save-btn:disabled { opacity: .55; cursor: not-allowed; }
.st-spin { width: 14px; height: 14px; animation: st-spin .7s linear infinite; }
</style>
