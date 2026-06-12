<template>
  <div class="ofp-body">
    <!-- Empty state -->
    <div
      v-if="rules.length === 0"
      class="ofp-empty"
    >
      {{ t('admin.settings.openaiFastPolicy.empty') }}
    </div>

    <!-- Rule Cards -->
    <div
      v-for="(rule, ruleIndex) in rules"
      :key="ruleIndex"
      class="ofp-rule-card"
    >
      <div class="ofp-rule-head">
        <span class="ofp-rule-title">
          {{ t('admin.settings.openaiFastPolicy.ruleHeader', { index: ruleIndex + 1 }) }}
        </span>
        <button
          type="button"
          class="ofp-icon-btn ofp-icon-btn--danger"
          @click="removeRule(ruleIndex)"
          :title="t('admin.settings.openaiFastPolicy.removeRule')"
        >
          <svg class="ofp-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="ofp-grid-3">
        <!-- Service Tier -->
        <div class="ofp-field">
          <label class="ofp-field-label">{{ t('admin.settings.openaiFastPolicy.serviceTier') }}</label>
          <Select
            :modelValue="rule.service_tier"
            @update:modelValue="rule.service_tier = $event as 'all' | 'priority' | 'flex'"
            :options="tierOptions"
          />
        </div>

        <!-- Action -->
        <div class="ofp-field">
          <label class="ofp-field-label">{{ t('admin.settings.openaiFastPolicy.action') }}</label>
          <Select
            :modelValue="rule.action"
            @update:modelValue="rule.action = $event as 'pass' | 'filter' | 'block'"
            :options="actionOptions"
          />
        </div>

        <!-- Scope -->
        <div class="ofp-field">
          <label class="ofp-field-label">{{ t('admin.settings.openaiFastPolicy.scope') }}</label>
          <Select
            :modelValue="rule.scope"
            @update:modelValue="rule.scope = $event as 'all' | 'oauth' | 'apikey' | 'bedrock'"
            :options="scopeOptions"
          />
        </div>
      </div>

      <!-- Error Message (only when action=block) -->
      <div v-if="rule.action === 'block'" class="ofp-field ofp-field--mt">
        <label class="ofp-field-label">{{ t('admin.settings.openaiFastPolicy.errorMessage') }}</label>
        <input
          v-model="rule.error_message"
          type="text"
          class="ofp-input"
          :placeholder="t('admin.settings.openaiFastPolicy.errorMessagePlaceholder')"
        />
        <p class="ofp-field-hint">{{ t('admin.settings.openaiFastPolicy.errorMessageHint') }}</p>
      </div>

      <!-- Model Whitelist -->
      <div class="ofp-field ofp-field--mt">
        <label class="ofp-field-label">{{ t('admin.settings.openaiFastPolicy.modelWhitelist') }}</label>
        <p class="ofp-field-hint ofp-field-hint--top">{{ t('admin.settings.openaiFastPolicy.modelWhitelistHint') }}</p>
        <div
          v-for="(_, patternIdx) in rule.model_whitelist || []"
          :key="patternIdx"
          class="ofp-pattern-row"
        >
          <input
            v-model="rule.model_whitelist![patternIdx]"
            type="text"
            class="ofp-input ofp-input--sm ofp-input--mono"
            :placeholder="t('admin.settings.openaiFastPolicy.modelPatternPlaceholder')"
          />
          <button
            type="button"
            class="ofp-icon-btn ofp-icon-btn--danger"
            @click="removeModelPattern(rule, patternIdx)"
          >
            <svg class="ofp-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <button
          type="button"
          class="ofp-add-btn"
          @click="addModelPattern(rule)"
        >
          <svg class="ofp-add-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          {{ t('admin.settings.openaiFastPolicy.addModelPattern') }}
        </button>
      </div>

      <!-- Fallback Action (only when model_whitelist is non-empty) -->
      <div
        v-if="rule.model_whitelist && rule.model_whitelist.length > 0"
        class="ofp-field ofp-field--mt"
      >
        <label class="ofp-field-label">{{ t('admin.settings.openaiFastPolicy.fallbackAction') }}</label>
        <Select
          :modelValue="rule.fallback_action || 'pass'"
          @update:modelValue="rule.fallback_action = $event as 'pass' | 'filter' | 'block'"
          :options="actionOptions"
        />
        <p class="ofp-field-hint">{{ t('admin.settings.openaiFastPolicy.fallbackActionHint') }}</p>
        <div v-if="rule.fallback_action === 'block'" class="ofp-fallback-err">
          <input
            v-model="rule.fallback_error_message"
            type="text"
            class="ofp-input"
            :placeholder="t('admin.settings.openaiFastPolicy.fallbackErrorMessagePlaceholder')"
          />
        </div>
      </div>
    </div>

    <!-- Add Rule Button + save hint -->
    <div class="ofp-add-rule-wrap">
      <button
        type="button"
        class="ofp-add-rule-btn"
        @click="addRule"
      >
        <svg class="ofp-add-rule-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        {{ t('admin.settings.openaiFastPolicy.addRule') }}
      </button>
      <p class="ofp-save-hint">{{ t('admin.settings.openaiFastPolicy.saveHint') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { OpenAIFastPolicyRule } from '@/api/admin/settings'

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

const { t } = useI18n()

// ── Options ────────────────────────────────────────────────────────────────
const tierOptions = computed(() => [
  { value: 'all', label: t('admin.settings.openaiFastPolicy.tierAll') },
  { value: 'priority', label: t('admin.settings.openaiFastPolicy.tierPriority') },
  { value: 'flex', label: t('admin.settings.openaiFastPolicy.tierFlex') },
])

const actionOptions = computed(() => [
  { value: 'pass', label: t('admin.settings.openaiFastPolicy.actionPass') },
  { value: 'filter', label: t('admin.settings.openaiFastPolicy.actionFilter') },
  { value: 'block', label: t('admin.settings.openaiFastPolicy.actionBlock') },
])

const scopeOptions = computed(() => [
  { value: 'all', label: t('admin.settings.openaiFastPolicy.scopeAll') },
  { value: 'oauth', label: t('admin.settings.openaiFastPolicy.scopeOAuth') },
  { value: 'apikey', label: t('admin.settings.openaiFastPolicy.scopeAPIKey') },
  { value: 'bedrock', label: t('admin.settings.openaiFastPolicy.scopeBedrock') },
])

// ── Local state (mirrors global settings key) ──────────────────────────────
const activeSource = computed(() => props.formValues ?? props.settings)

function parseRulesFromSource(src: Record<string, unknown>): OpenAIFastPolicyRule[] {
  const raw = src['openai_fast_policy_settings']
  if (raw && typeof raw === 'object' && !Array.isArray(raw)) {
    const obj = raw as Record<string, unknown>
    if (Array.isArray(obj['rules'])) {
      return (obj['rules'] as OpenAIFastPolicyRule[]).map((rule) => ({
        ...rule,
        model_whitelist: rule.model_whitelist ? [...rule.model_whitelist] : [],
      }))
    }
  }
  return []
}

const rules = ref<OpenAIFastPolicyRule[]>(parseRulesFromSource(activeSource.value))

// Re-sync when parent resets (e.g., after global save)
watch(
  () => activeSource.value['openai_fast_policy_settings'],
  (incoming) => {
    if (incoming && typeof incoming === 'object' && !Array.isArray(incoming)) {
      const obj = incoming as Record<string, unknown>
      if (Array.isArray(obj['rules'])) {
        rules.value = (obj['rules'] as OpenAIFastPolicyRule[]).map((rule) => ({
          ...rule,
          model_whitelist: rule.model_whitelist ? [...rule.model_whitelist] : [],
        }))
      }
    }
  },
  { deep: true },
)

// Emit cleaned rules up whenever local state changes
function emitRules() {
  const cleaned = rules.value.map((rule) => {
    const whitelist = (rule.model_whitelist || [])
      .map((p) => p.trim())
      .filter((p) => p !== '')
    const hasWhitelist = whitelist.length > 0
    return {
      service_tier: rule.service_tier,
      action: rule.action,
      scope: rule.scope,
      error_message: rule.action === 'block' ? rule.error_message : undefined,
      model_whitelist: hasWhitelist ? whitelist : undefined,
      fallback_action: hasWhitelist ? rule.fallback_action || 'pass' : undefined,
      fallback_error_message:
        hasWhitelist && rule.fallback_action === 'block'
          ? rule.fallback_error_message
          : undefined,
    } as OpenAIFastPolicyRule
  })
  emit('update:field', 'openai_fast_policy_settings', { rules: cleaned })
}

// ── Mutations ──────────────────────────────────────────────────────────────
function addRule() {
  rules.value.push({
    service_tier: 'priority',
    action: 'filter',
    scope: 'all',
    error_message: '',
    model_whitelist: [],
    fallback_action: 'pass',
    fallback_error_message: '',
  })
  emitRules()
}

function removeRule(index: number) {
  rules.value.splice(index, 1)
  emitRules()
}

function addModelPattern(rule: OpenAIFastPolicyRule) {
  if (!rule.model_whitelist) rule.model_whitelist = []
  rule.model_whitelist.push('')
  emitRules()
}

function removeModelPattern(rule: OpenAIFastPolicyRule, idx: number) {
  rule.model_whitelist?.splice(idx, 1)
  emitRules()
}

// Emit on any inline edit (Select uses update:modelValue, inputs use v-model which
// mutates the reactive ref — we watch the whole rules array for deep changes)
watch(rules, emitRules, { deep: true })
</script>

<style scoped>
.ofp-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

.ofp-empty {
  border: 1px dashed var(--line-1, #2F3540); border-radius: 10px;
  padding: 24px 16px; text-align: center; font-size: 13px; color: var(--ink-2, #5C6470);
}

/* Rule card */
.ofp-rule-card {
  border: 1px solid var(--line-1, #2F3540); border-radius: 10px; padding: 16px;
  display: flex; flex-direction: column; gap: 12px;
}

.ofp-rule-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.ofp-rule-title { font-size: 13px; font-weight: 600; color: var(--ink-0, #E8EBF0); }

.ofp-grid-3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; }
@media (max-width: 600px) { .ofp-grid-3 { grid-template-columns: 1fr; } }

.ofp-field { display: flex; flex-direction: column; gap: 4px; }
.ofp-field--mt { margin-top: 4px; }
.ofp-field-label { font-size: 11.5px; font-weight: 500; color: var(--ink-1, #97A0AF); }
.ofp-field-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 2px 0 0; }
.ofp-field-hint--top { margin: 0 0 6px; }

.ofp-input {
  width: 100%; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.ofp-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.ofp-input--sm { padding: 5px 9px; font-size: 12px; }
.ofp-input--mono { font-family: var(--font-mono, ui-monospace, monospace); }

.ofp-pattern-row { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }

.ofp-icon-btn {
  flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid transparent; background: transparent; cursor: pointer;
  transition: color .12s, background .12s, border-color .12s;
}
.ofp-icon-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.ofp-icon-btn--danger { color: var(--ink-2, #5C6470); }
.ofp-icon-btn--danger:hover { color: var(--bad, #F25C69); background: rgba(242,92,105,.1); border-color: rgba(242,92,105,.25); }
.ofp-icon { width: 14px; height: 14px; }

.ofp-add-btn {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 0; background: none; border: none; cursor: pointer; font-family: inherit;
  font-size: 12px; color: var(--azure, #5CA8FF); transition: color .15s;
}
.ofp-add-btn:hover { color: var(--ink-0, #E8EBF0); }
.ofp-add-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.ofp-add-icon { width: 13px; height: 13px; }

.ofp-fallback-err { margin-top: 8px; }

.ofp-add-rule-wrap { display: flex; flex-direction: column; gap: 6px; }

.ofp-add-rule-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 14px; border-radius: 8px; font-size: 12.5px; font-weight: 500;
  font-family: inherit; cursor: pointer;
  border: 1px solid var(--line-1, #2F3540); background: transparent;
  color: var(--ink-1, #97A0AF); transition: border-color .15s, color .15s, background .15s;
}
.ofp-add-rule-btn:hover { border-color: var(--azure, #5CA8FF); color: var(--azure, #5CA8FF); background: rgba(92,168,255,.07); }
.ofp-add-rule-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.ofp-add-rule-icon { width: 14px; height: 14px; }

.ofp-save-hint { font-size: 11px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }
</style>
