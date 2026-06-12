<template>
  <div class="oc-body">
    <!-- loading -->
    <div v-if="loading" class="oc-loading">
      <div class="oc-spinner" />
      {{ t('common.loading') }}
    </div>

    <template v-else>
      <!-- enabled switch -->
      <div class="oc-row-between">
        <div>
          <label class="oc-label">{{ t('admin.settings.overloadCooldown.enabled') }}</label>
          <p class="oc-hint">{{ t('admin.settings.overloadCooldown.enabledHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <!-- expanded fields -->
      <div v-if="form.enabled" class="oc-expanded">
        <div class="oc-field">
          <label class="oc-field-label">{{ t('admin.settings.overloadCooldown.cooldownMinutes') }}</label>
          <input
            v-model.number="form.cooldown_minutes"
            type="number"
            min="1"
            max="120"
            class="oc-input oc-input--mono"
          />
          <p class="oc-field-hint">{{ t('admin.settings.overloadCooldown.cooldownMinutesHint') }}</p>
        </div>
      </div>

      <!-- save -->
      <div class="oc-footer">
        <button
          type="button"
          class="oc-save-btn"
          :disabled="saving"
          @click="save"
        >
          <svg
            v-if="saving"
            class="oc-spin"
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
  cooldown_minutes: 10,
})

onMounted(async () => {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getOverloadCooldownSettings()
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
    const updated = await adminAPI.settings.updateOverloadCooldownSettings({
      enabled: form.enabled,
      cooldown_minutes: form.cooldown_minutes,
    })
    Object.assign(form, updated)
    appStore.showSuccess(t('admin.settings.overloadCooldown.saved'))
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t('admin.settings.overloadCooldown.saveFailed')),
    )
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.oc-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.oc-loading { display: flex; align-items: center; gap: 8px; color: var(--ink-2, #5C6470); font-size: 13px; }
.oc-spinner {
  width: 16px; height: 16px; border-radius: 50%;
  border: 2px solid var(--line-1, #2F3540);
  border-bottom-color: var(--azure, #5CA8FF);
  animation: oc-spin .7s linear infinite; flex-shrink: 0;
}
@keyframes oc-spin { to { transform: rotate(360deg); } }

.oc-row-between { display: flex; align-items: center; justify-content: space-between; gap: 16px; }

.oc-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.oc-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.oc-expanded { border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px; display: flex; flex-direction: column; gap: 12px; }

.oc-field { display: flex; flex-direction: column; gap: 4px; }
.oc-field-label { font-size: 12px; font-weight: 500; color: var(--ink-1, #97A0AF); }
.oc-field-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.oc-input {
  width: 8rem; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.oc-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.oc-input--mono { font-family: var(--font-mono, ui-monospace, monospace); }

.oc-footer {
  display: flex; justify-content: flex-end;
  border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px;
}

.oc-save-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 18px; border-radius: 8px; font-size: 13px; font-weight: 600;
  font-family: inherit; cursor: pointer; user-select: none;
  border: 1px solid var(--azure, #5CA8FF);
  background: linear-gradient(180deg, rgba(92,168,255,.18) 0%, rgba(92,168,255,.08) 100%);
  color: var(--azure, #5CA8FF);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.08), 0 2px 6px rgba(92,168,255,.15);
  transition: background .15s, box-shadow .15s, opacity .15s;
}
.oc-save-btn:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(92,168,255,.28) 0%, rgba(92,168,255,.14) 100%);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.1), 0 3px 10px rgba(92,168,255,.25);
}
.oc-save-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.oc-save-btn:disabled { opacity: .55; cursor: not-allowed; }
.oc-spin { width: 14px; height: 14px; animation: oc-spin .7s linear infinite; }
</style>
