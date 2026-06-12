<template>
  <div class="bp-body">
    <!-- loading -->
    <div v-if="loading" class="bp-loading">
      <div class="bp-spinner" />
      {{ t('common.loading') }}
    </div>

    <template v-else>
      <!-- Rule Cards -->
      <div
        v-for="rule in rules"
        :key="rule.beta_token"
        class="bp-rule-card"
      >
        <!-- card header -->
        <div class="bp-rule-head">
          <span class="bp-rule-name">{{ getBetaDisplayName(rule.beta_token) }}</span>
          <span class="bp-rule-token">{{ rule.beta_token }}</span>
        </div>

        <div class="bp-grid-2">
          <!-- Action -->
          <div class="bp-field">
            <label class="bp-field-label">{{ t('admin.settings.betaPolicy.action') }}</label>
            <Select
              :modelValue="rule.action"
              @update:modelValue="rule.action = $event as any"
              :options="actionOptions"
            />
          </div>

          <!-- Scope -->
          <div class="bp-field">
            <label class="bp-field-label">{{ t('admin.settings.betaPolicy.scope') }}</label>
            <Select
              :modelValue="rule.scope"
              @update:modelValue="rule.scope = $event as any"
              :options="scopeOptions"
            />
          </div>
        </div>

        <!-- Error Message (only when action=block) -->
        <div v-if="rule.action === 'block'" class="bp-field bp-field--mt">
          <label class="bp-field-label">{{ t('admin.settings.betaPolicy.errorMessage') }}</label>
          <input
            v-model="rule.error_message"
            type="text"
            class="bp-input"
            :placeholder="t('admin.settings.betaPolicy.errorMessagePlaceholder')"
          />
          <p class="bp-field-hint">{{ t('admin.settings.betaPolicy.errorMessageHint') }}</p>
        </div>

        <!-- Quick Presets (only for tokens with presets) -->
        <div v-if="betaPresets[rule.beta_token]?.length" class="bp-field bp-field--mt">
          <label class="bp-field-label">{{ t('admin.settings.betaPolicy.quickPresets') }}</label>
          <div class="bp-presets">
            <button
              v-for="preset in betaPresets[rule.beta_token]"
              :key="preset.label"
              type="button"
              class="bp-preset-btn"
              @click="applyPreset(rule, preset)"
              :title="preset.description"
            >
              {{ preset.label }}
            </button>
          </div>
        </div>

        <!-- Model Whitelist -->
        <div class="bp-field bp-field--mt">
          <label class="bp-field-label">{{ t('admin.settings.betaPolicy.modelWhitelist') }}</label>
          <p class="bp-field-hint bp-field-hint--top">{{ t('admin.settings.betaPolicy.modelWhitelistHint') }}</p>
          <!-- existing patterns -->
          <div
            v-for="(_, index) in rule.model_whitelist || []"
            :key="index"
            class="bp-pattern-row"
          >
            <input
              v-model="rule.model_whitelist![index]"
              type="text"
              class="bp-input bp-input--sm bp-input--mono"
              :placeholder="t('admin.settings.betaPolicy.modelPatternPlaceholder')"
            />
            <button
              type="button"
              class="bp-icon-btn bp-icon-btn--danger"
              @click="rule.model_whitelist!.splice(index, 1)"
              :aria-label="t('common.remove')"
            >
              <svg class="bp-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <!-- add pattern -->
          <button
            type="button"
            class="bp-add-btn"
            @click="addModelPattern(rule)"
          >
            <svg class="bp-add-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            {{ t('admin.settings.betaPolicy.addModelPattern') }}
          </button>
          <!-- Common pattern chips -->
          <div class="bp-chips">
            <span class="bp-chips-label">{{ t('admin.settings.betaPolicy.commonPatterns') }}:</span>
            <button
              v-for="pattern in commonModelPatterns"
              :key="pattern"
              type="button"
              class="bp-chip"
              @click="addQuickPattern(rule, pattern)"
            >
              {{ pattern }}
            </button>
          </div>
        </div>

        <!-- Fallback Action (only when model_whitelist is non-empty) -->
        <div
          v-if="rule.model_whitelist && rule.model_whitelist.length > 0"
          class="bp-field bp-field--mt"
        >
          <label class="bp-field-label">{{ t('admin.settings.betaPolicy.fallbackAction') }}</label>
          <Select
            :modelValue="rule.fallback_action || 'pass'"
            @update:modelValue="rule.fallback_action = $event as any"
            :options="actionOptions"
          />
          <p class="bp-field-hint">{{ t('admin.settings.betaPolicy.fallbackActionHint') }}</p>
          <!-- Fallback Error Message (only when fallback_action=block) -->
          <div v-if="rule.fallback_action === 'block'" class="bp-fallback-err">
            <input
              v-model="rule.fallback_error_message"
              type="text"
              class="bp-input"
              :placeholder="t('admin.settings.betaPolicy.fallbackErrorMessagePlaceholder')"
            />
            <p class="bp-field-hint">{{ t('admin.settings.betaPolicy.errorMessageHint') }}</p>
          </div>
        </div>
      </div>

      <!-- Save button -->
      <div class="bp-footer">
        <button
          type="button"
          class="bp-save-btn"
          :disabled="saving"
          @click="save"
        >
          <svg v-if="saving" class="bp-spin" fill="none" viewBox="0 0 24 24">
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
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { BetaPolicyRule } from '@/api/admin/settings'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)

const rules = ref<BetaPolicyRule[]>([])

// ── Options ────────────────────────────────────────────────────────────────
const actionOptions = computed(() => [
  { value: 'pass', label: t('admin.settings.betaPolicy.actionPass') },
  { value: 'filter', label: t('admin.settings.betaPolicy.actionFilter') },
  { value: 'block', label: t('admin.settings.betaPolicy.actionBlock') },
])

const scopeOptions = computed(() => [
  { value: 'all', label: t('admin.settings.betaPolicy.scopeAll') },
  { value: 'oauth', label: t('admin.settings.betaPolicy.scopeOAuth') },
  { value: 'apikey', label: t('admin.settings.betaPolicy.scopeAPIKey') },
  { value: 'bedrock', label: t('admin.settings.betaPolicy.scopeBedrock') },
])

// ── Display names ──────────────────────────────────────────────────────────
const betaDisplayNames: Record<string, string> = {
  'fast-mode-2026-02-01': 'Fast Mode',
  'context-1m-2025-08-07': 'Context 1M',
}

function getBetaDisplayName(token: string): string {
  return betaDisplayNames[token] || token
}

// ── Quick presets ──────────────────────────────────────────────────────────
const betaPresets: Record<
  string,
  Array<{
    label: string
    description: string
    action: 'pass' | 'filter' | 'block'
    model_whitelist: string[]
    fallback_action: 'pass' | 'filter' | 'block'
  }>
> = {
  'context-1m-2025-08-07': [
    {
      label: t('admin.settings.betaPolicy.presetOpusOnly'),
      description: t('admin.settings.betaPolicy.presetOpusOnlyDesc'),
      action: 'pass',
      model_whitelist: ['claude-opus-4-6'],
      fallback_action: 'filter',
    },
  ],
}

// ── Common patterns ────────────────────────────────────────────────────────
const commonModelPatterns = [
  'claude-opus-4-6',
  'claude-sonnet-4-6',
  'claude-opus-*',
  'claude-sonnet-*',
]

function applyPreset(
  rule: BetaPolicyRule,
  preset: {
    action: 'pass' | 'filter' | 'block'
    model_whitelist: string[]
    fallback_action: 'pass' | 'filter' | 'block'
  },
) {
  rule.action = preset.action
  rule.model_whitelist = [...preset.model_whitelist]
  rule.fallback_action = preset.fallback_action
}

function addModelPattern(rule: BetaPolicyRule) {
  if (!rule.model_whitelist) rule.model_whitelist = []
  rule.model_whitelist.push('')
}

function addQuickPattern(rule: BetaPolicyRule, pattern: string) {
  if (!rule.model_whitelist) rule.model_whitelist = []
  if (!rule.model_whitelist.includes(pattern)) {
    rule.model_whitelist.push(pattern)
  }
}

// ── Load ───────────────────────────────────────────────────────────────────
onMounted(async () => {
  loading.value = true
  try {
    const settings = await adminAPI.settings.getBetaPolicySettings()
    rules.value = settings.rules
  } catch {
    // silent — form stays at defaults
  } finally {
    loading.value = false
  }
})

// ── Save ───────────────────────────────────────────────────────────────────
async function save() {
  saving.value = true
  try {
    const cleanedRules = rules.value.map((rule) => {
      const whitelist = rule.model_whitelist?.filter((p) => p.trim() !== '')
      const hasWhitelist = whitelist && whitelist.length > 0
      return {
        beta_token: rule.beta_token,
        action: rule.action,
        scope: rule.scope,
        error_message: rule.error_message,
        model_whitelist: hasWhitelist ? whitelist : undefined,
        fallback_action: hasWhitelist ? rule.fallback_action || 'pass' : undefined,
        fallback_error_message:
          hasWhitelist && rule.fallback_action === 'block'
            ? rule.fallback_error_message
            : undefined,
      } as BetaPolicyRule
    })
    const updated = await adminAPI.settings.updateBetaPolicySettings({ rules: cleanedRules })
    rules.value = updated.rules
    appStore.showSuccess(t('admin.settings.betaPolicy.saved'))
  } catch (error: unknown) {
    appStore.showError(
      extractApiErrorMessage(error, t('admin.settings.betaPolicy.saveFailed')),
    )
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.bp-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.bp-loading { display: flex; align-items: center; gap: 8px; color: var(--ink-2, #5C6470); font-size: 13px; }
.bp-spinner {
  width: 16px; height: 16px; border-radius: 50%;
  border: 2px solid var(--line-1, #2F3540);
  border-bottom-color: var(--azure, #5CA8FF);
  animation: bp-spin .7s linear infinite; flex-shrink: 0;
}
@keyframes bp-spin { to { transform: rotate(360deg); } }

/* Rule card */
.bp-rule-card {
  border: 1px solid var(--line-1, #2F3540); border-radius: 10px; padding: 16px;
  display: flex; flex-direction: column; gap: 12px;
}

.bp-rule-head { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.bp-rule-name { font-size: 13px; font-weight: 600; color: var(--ink-0, #E8EBF0); }
.bp-rule-token {
  font-size: 11px; font-family: var(--font-mono, ui-monospace, monospace);
  padding: 2px 7px; border-radius: 5px;
  background: var(--bg-1, #13161B); color: var(--ink-2, #5C6470);
  border: 1px solid var(--line-1, #2F3540);
}

.bp-grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
@media (max-width: 480px) { .bp-grid-2 { grid-template-columns: 1fr; } }

.bp-field { display: flex; flex-direction: column; gap: 4px; }
.bp-field--mt { margin-top: 4px; }
.bp-field-label { font-size: 11.5px; font-weight: 500; color: var(--ink-1, #97A0AF); }
.bp-field-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }
.bp-field-hint--top { margin: 0 0 6px; }

.bp-input {
  width: 100%; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.bp-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.bp-input--sm { padding: 5px 9px; font-size: 12px; }
.bp-input--mono { font-family: var(--font-mono, ui-monospace, monospace); }

.bp-pattern-row { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }

.bp-icon-btn {
  flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid transparent; background: transparent; cursor: pointer;
  transition: color .12s, background .12s, border-color .12s;
}
.bp-icon-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.bp-icon-btn--danger { color: var(--ink-2, #5C6470); }
.bp-icon-btn--danger:hover { color: var(--bad, #F25C69); background: rgba(242,92,105,.1); border-color: rgba(242,92,105,.25); }
.bp-icon { width: 14px; height: 14px; }

.bp-add-btn {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 0; background: none; border: none; cursor: pointer; font-family: inherit;
  font-size: 12px; color: var(--azure, #5CA8FF); transition: color .15s;
}
.bp-add-btn:hover { color: var(--ink-0, #E8EBF0); }
.bp-add-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.bp-add-icon { width: 13px; height: 13px; }

.bp-chips { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin-top: 4px; }
.bp-chips-label { font-size: 11px; color: var(--ink-2, #5C6470); }
.bp-chip {
  padding: 2px 9px; border-radius: 5px; font-size: 11.5px; font-family: var(--font-mono, ui-monospace, monospace);
  border: 1px solid var(--line-1, #2F3540); background: transparent;
  color: var(--ink-1, #97A0AF); cursor: pointer;
  transition: border-color .12s, background .12s, color .12s;
}
.bp-chip:hover { border-color: var(--azure, #5CA8FF); background: rgba(92,168,255,.07); color: var(--azure, #5CA8FF); }
.bp-chip:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

.bp-presets { display: flex; flex-wrap: wrap; gap: 6px; }
.bp-preset-btn {
  display: inline-flex; align-items: center; gap: 4px; padding: 4px 11px; border-radius: 7px;
  border: 1px solid var(--azure-dim, #2A4D7A); background: rgba(92,168,255,.08);
  color: var(--azure, #5CA8FF); font-size: 12px; font-weight: 500; font-family: inherit;
  cursor: pointer; transition: background .12s, box-shadow .12s;
}
.bp-preset-btn:hover { background: rgba(92,168,255,.18); box-shadow: 0 0 0 1px var(--azure, #5CA8FF); }
.bp-preset-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

.bp-fallback-err { margin-top: 8px; display: flex; flex-direction: column; gap: 4px; }

.bp-footer {
  display: flex; justify-content: flex-end;
  border-top: 1px solid var(--line-1, #2F3540); padding-top: 16px;
}

.bp-save-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 18px; border-radius: 8px; font-size: 13px; font-weight: 600;
  font-family: inherit; cursor: pointer; user-select: none;
  border: 1px solid var(--azure, #5CA8FF);
  background: linear-gradient(180deg, rgba(92,168,255,.18) 0%, rgba(92,168,255,.08) 100%);
  color: var(--azure, #5CA8FF);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.08), 0 2px 6px rgba(92,168,255,.15);
  transition: background .15s, box-shadow .15s, opacity .15s;
}
.bp-save-btn:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(92,168,255,.28) 0%, rgba(92,168,255,.14) 100%);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.1), 0 3px 10px rgba(92,168,255,.25);
}
.bp-save-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.bp-save-btn:disabled { opacity: .55; cursor: not-allowed; }
.bp-spin { width: 14px; height: 14px; animation: bp-spin .7s linear infinite; }
</style>
