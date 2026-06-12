<template>
  <div class="rc-body">
    <!-- loading -->
    <div v-if="loading" class="rc-loading">
      <div class="rc-spinner" />
      {{ t('common.loading') }}
    </div>

    <template v-else>
      <!-- Master enabled switch -->
      <div class="rc-row-between">
        <div>
          <label class="rc-label">{{ t('admin.settings.rectifier.enabled') }}</label>
          <p class="rc-hint">{{ t('admin.settings.rectifier.enabledHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <!-- Sub-toggles (only show when master is enabled) -->
      <div v-if="form.enabled" class="rc-expanded">
        <!-- Thinking Signature Rectifier -->
        <div class="rc-row-between">
          <div>
            <label class="rc-sub-label">{{ t('admin.settings.rectifier.thinkingSignature') }}</label>
            <p class="rc-sub-hint">{{ t('admin.settings.rectifier.thinkingSignatureHint') }}</p>
          </div>
          <Toggle v-model="form.thinking_signature_enabled" />
        </div>

        <!-- Thinking Budget Rectifier -->
        <div class="rc-row-between">
          <div>
            <label class="rc-sub-label">{{ t('admin.settings.rectifier.thinkingBudget') }}</label>
            <p class="rc-sub-hint">{{ t('admin.settings.rectifier.thinkingBudgetHint') }}</p>
          </div>
          <Toggle v-model="form.thinking_budget_enabled" />
        </div>

        <!-- API Key Signature Rectifier -->
        <div class="rc-row-between">
          <div>
            <label class="rc-sub-label">{{ t('admin.settings.rectifier.apikeySignature') }}</label>
            <p class="rc-sub-hint">{{ t('admin.settings.rectifier.apikeySignatureHint') }}</p>
          </div>
          <Toggle v-model="form.apikey_signature_enabled" />
        </div>

        <!-- Custom Patterns (only when apikey_signature_enabled) -->
        <div v-if="form.apikey_signature_enabled" class="rc-patterns-block">
          <div>
            <label class="rc-sub-label">{{ t('admin.settings.rectifier.apikeyPatterns') }}</label>
            <p class="rc-sub-hint">{{ t('admin.settings.rectifier.apikeyPatternsHint') }}</p>
          </div>
          <div
            v-for="(_, index) in form.apikey_signature_patterns"
            :key="index"
            class="rc-pattern-row"
          >
            <input
              v-model="form.apikey_signature_patterns[index]"
              type="text"
              class="rc-input rc-input--mono"
              :placeholder="t('admin.settings.rectifier.apikeyPatternPlaceholder')"
            />
            <button
              type="button"
              class="rc-icon-btn rc-icon-btn--danger"
              @click="form.apikey_signature_patterns.splice(index, 1)"
              :aria-label="t('common.remove')"
            >
              <svg class="rc-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <button
            type="button"
            class="rc-add-btn"
            @click="form.apikey_signature_patterns.push('')"
          >
            + {{ t('admin.settings.rectifier.addPattern') }}
          </button>
        </div>
      </div>

      <!-- Save button -->
      <div class="rc-footer">
        <button
          type="button"
          class="rc-save-btn"
          :disabled="saving"
          @click="save"
        >
          <svg v-if="saving" class="rc-spin" fill="none" viewBox="0 0 24 24">
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
  thinking_signature_enabled: true,
  thinking_budget_enabled: true,
  apikey_signature_enabled: false,
  apikey_signature_patterns: [] as string[],
})

onMounted(async () => {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getRectifierSettings()
    Object.assign(form, settings)
    if (!Array.isArray(form.apikey_signature_patterns)) {
      form.apikey_signature_patterns = []
    }
  } catch {
    // silent — form stays at defaults
  } finally {
    loading.value = false
  }
})

async function save() {
  saving.value = true
  try {
    const updated = await adminAPI.settings.updateRectifierSettings({
      enabled: form.enabled,
      thinking_signature_enabled: form.thinking_signature_enabled,
      thinking_budget_enabled: form.thinking_budget_enabled,
      apikey_signature_enabled: form.apikey_signature_enabled,
      apikey_signature_patterns: form.apikey_signature_patterns.filter((p) => p.trim() !== ''),
    })
    Object.assign(form, updated)
    if (!Array.isArray(form.apikey_signature_patterns)) {
      form.apikey_signature_patterns = []
    }
    appStore.showSuccess(t('admin.settings.rectifier.saved'))
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t('admin.settings.rectifier.saveFailed')),
    )
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.rc-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.rc-loading { display: flex; align-items: center; gap: 8px; color: var(--ink-2, #5C6470); font-size: 13px; }
.rc-spinner {
  width: 16px; height: 16px; border-radius: 50%;
  border: 2px solid var(--line-1, #2F3540);
  border-bottom-color: var(--azure, #5CA8FF);
  animation: rc-spin .7s linear infinite; flex-shrink: 0;
}
@keyframes rc-spin { to { transform: rotate(360deg); } }

.rc-row-between { display: flex; align-items: center; justify-content: space-between; gap: 16px; }

.rc-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.rc-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.rc-sub-label { font-size: 12.5px; font-weight: 500; color: var(--ink-0, #E8EBF0); opacity: .85; }
.rc-sub-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }

.rc-expanded {
  border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px;
  display: flex; flex-direction: column; gap: 16px;
}

.rc-patterns-block {
  margin-left: 16px; padding-left: 16px;
  border-left: 2px solid var(--line-1, #2F3540);
  display: flex; flex-direction: column; gap: 8px;
}

.rc-pattern-row { display: flex; align-items: center; gap: 8px; }

.rc-input {
  flex: 1; min-width: 0; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.rc-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.rc-input--mono { font-family: var(--font-mono, ui-monospace, monospace); }

.rc-icon-btn {
  flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid transparent; background: transparent; cursor: pointer;
  transition: color .12s, background .12s, border-color .12s;
}
.rc-icon-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.rc-icon-btn--danger { color: var(--ink-2, #5C6470); }
.rc-icon-btn--danger:hover { color: var(--bad, #F25C69); background: rgba(242,92,105,.1); border-color: rgba(242,92,105,.25); }
.rc-icon { width: 14px; height: 14px; }

.rc-add-btn {
  align-self: flex-start; padding: 5px 11px; border-radius: 7px;
  border: 1px solid var(--line-1, #2F3540); background: transparent;
  color: var(--azure, #5CA8FF); font-size: 12px; font-weight: 500;
  cursor: pointer; font-family: inherit; transition: border-color .15s, background .15s;
}
.rc-add-btn:hover { border-color: var(--azure, #5CA8FF); background: rgba(92,168,255,.07); }
.rc-add-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

.rc-footer {
  display: flex; justify-content: flex-end;
  border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px;
}

.rc-save-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 18px; border-radius: 8px; font-size: 13px; font-weight: 600;
  font-family: inherit; cursor: pointer; user-select: none;
  border: 1px solid var(--azure, #5CA8FF);
  background: linear-gradient(180deg, rgba(92,168,255,.18) 0%, rgba(92,168,255,.08) 100%);
  color: var(--azure, #5CA8FF);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.08), 0 2px 6px rgba(92,168,255,.15);
  transition: background .15s, box-shadow .15s, opacity .15s;
}
.rc-save-btn:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(92,168,255,.28) 0%, rgba(92,168,255,.14) 100%);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.1), 0 3px 10px rgba(92,168,255,.25);
}
.rc-save-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.rc-save-btn:disabled { opacity: .55; cursor: not-allowed; }
.rc-spin { width: 14px; height: 14px; animation: rc-spin .7s linear infinite; }
</style>
