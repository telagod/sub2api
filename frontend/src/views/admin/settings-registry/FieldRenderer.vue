<template>
  <div v-if="visible" class="sr-field">
    <!-- switch -->
    <template v-if="field.type === 'switch'">
      <div class="sr-field-row">
        <div class="sr-field-info">
          <label class="sr-label">{{ resolveLabel(field.label) }}</label>
          <p v-if="field.help" class="sr-help">{{ resolveLabel(field.help) }}</p>
        </div>
        <button
          type="button"
          class="sr-toggle"
          :class="{ on: !!modelValue }"
          :aria-checked="!!modelValue"
          role="switch"
          @click="emit('update:modelValue', !modelValue)"
        >
          <span class="sr-toggle-thumb" />
        </button>
      </div>
    </template>

    <!-- select -->
    <template v-else-if="field.type === 'select'">
      <label class="sr-label mb-1 block">{{ resolveLabel(field.label) }}</label>
      <select class="sr-input" :value="modelValue" @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)">
        <option v-for="opt in field.options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option>
      </select>
      <p v-if="field.help" class="sr-help mt-1">{{ resolveLabel(field.help) }}</p>
    </template>

    <!-- password -->
    <template v-else-if="field.type === 'password'">
      <label class="sr-label mb-1 block">{{ resolveLabel(field.label) }}</label>
      <div class="sr-pw-wrap">
        <input
          :type="showPassword ? 'text' : 'password'"
          class="sr-input sr-input-pw"
          :value="modelValue ?? ''"
          :placeholder="passwordPlaceholder"
          autocomplete="new-password"
          @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        />
        <button type="button" class="sr-pw-eye" :aria-label="showPassword ? 'Hide' : 'Show'" @click="showPassword = !showPassword">
          <svg v-if="showPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="sr-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 4.411m0 0L21 21" /></svg>
          <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="sr-icon"><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
        </button>
      </div>
      <p v-if="field.help" class="sr-help mt-1">{{ resolveLabel(field.help) }}</p>
    </template>

    <!-- image -->
    <template v-else-if="field.type === 'image'">
      <label class="sr-label mb-1 block">{{ resolveLabel(field.label) }}</label>
      <ImageUpload
        :model-value="String(modelValue ?? '')"
        mode="image"
        upload-label="Upload"
        remove-label="Remove"
        :hint="field.help ? resolveLabel(field.help) : ''"
        @update:model-value="emit('update:modelValue', $event)"
      />
    </template>

    <!-- json / textarea -->
    <template v-else-if="field.type === 'json' || field.type === 'textarea'">
      <label class="sr-label mb-1 block">{{ resolveLabel(field.label) }}</label>
      <textarea
        class="sr-input sr-mono"
        rows="4"
        :value="jsonDisplay"
        :placeholder="field.placeholder ? resolveLabel(field.placeholder) : ''"
        @change="handleJsonChange(($event.target as HTMLTextAreaElement).value)"
      />
      <p v-if="jsonError" class="sr-error mt-1">{{ jsonError }}</p>
      <p v-else-if="field.help" class="sr-help mt-1">{{ resolveLabel(field.help) }}</p>
    </template>

    <!-- number -->
    <template v-else-if="field.type === 'number'">
      <label class="sr-label mb-1 block">{{ resolveLabel(field.label) }}</label>
      <input
        type="number"
        class="sr-input sr-input-num"
        :value="modelValue ?? ''"
        :placeholder="field.placeholder ? resolveLabel(field.placeholder) : ''"
        @input="emit('update:modelValue', Number(($event.target as HTMLInputElement).value))"
      />
      <p v-if="field.help" class="sr-help mt-1">{{ resolveLabel(field.help) }}</p>
    </template>

    <!-- text (default) -->
    <template v-else>
      <label class="sr-label mb-1 block">{{ resolveLabel(field.label) }}</label>
      <input
        type="text"
        class="sr-input"
        :value="modelValue ?? ''"
        :placeholder="field.placeholder ? resolveLabel(field.placeholder) : ''"
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <p v-if="field.help" class="sr-help mt-1">{{ resolveLabel(field.help) }}</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import ImageUpload from '@/components/common/ImageUpload.vue'
import type { SettingsField } from './types'

const props = defineProps<{
  field: SettingsField
  modelValue: unknown
  formValues: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:modelValue': [value: unknown]
}>()

const { t } = useI18n()

const showPassword = ref(false)
const jsonError = ref('')

/**
 * For sensitive password fields: if the value is empty/null but the backend
 * reports <key>_configured = true, show a hint instead of a blank placeholder.
 */
const passwordPlaceholder = computed(() => {
  if (props.field.type !== 'password') return ''
  if (props.modelValue != null && props.modelValue !== '') {
    return props.field.placeholder ? resolveLabel(props.field.placeholder) : ''
  }
  if (props.field.sensitive) {
    const configuredKey = `${props.field.key}_configured`
    if (props.formValues[configuredKey]) return t('admin.settingsRegistry.fieldPasswordConfigured')
  }
  return props.field.placeholder ? resolveLabel(props.field.placeholder) : ''
})

/** Show only when showWhen predicate passes (or absent) */
const visible = computed(() =>
  props.field.showWhen ? props.field.showWhen(props.formValues) : true,
)

/** Resolve i18n key or literal string */
function resolveLabel(key: string): string {
  // If the key exists in i18n, translate; else return as-is
  try {
    const result = t(key)
    return result === key ? key : result
  } catch {
    return key
  }
}

/** JSON/textarea display */
const jsonDisplay = computed(() => {
  const v = props.modelValue
  if (v === undefined || v === null) return ''
  if (typeof v === 'string') return v
  return JSON.stringify(v, null, 2)
})

function handleJsonChange(raw: string) {
  jsonError.value = ''
  if (!raw.trim()) {
    emit('update:modelValue', raw)
    return
  }
  // Try to parse as JSON for array/object fields
  try {
    const parsed = JSON.parse(raw)
    emit('update:modelValue', parsed)
  } catch {
    jsonError.value = 'Invalid JSON'
    emit('update:modelValue', raw)
  }
}
</script>

<style scoped>
.sr-field { display: flex; flex-direction: column; }

/* switch row */
.sr-field-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; }
.sr-field-info { flex: 1; min-width: 0; }

/* labels + help */
.sr-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.sr-help  { font-size: 11.5px; color: var(--ink-2, #5C6470); margin-top: 3px; line-height: 1.5; }
.sr-error { font-size: 11.5px; color: var(--bad, #F25C69); margin-top: 3px; }

/* input */
.sr-input {
  width: 100%;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color .15s, box-shadow .15s;
  box-sizing: border-box;
}
.sr-input:focus,
.sr-input:focus-visible { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.sr-input-num { width: 120px; }
.sr-mono { font-family: var(--font-mono, "IBM Plex Mono", monospace); font-size: 12px; resize: vertical; }

/* password wrapper */
.sr-pw-wrap { position: relative; }
.sr-input-pw { padding-right: 40px; }
.sr-pw-eye {
  position: absolute; right: 10px; top: 50%; transform: translateY(-50%);
  background: none; border: none; cursor: pointer;
  color: var(--ink-2, #5C6470); padding: 2px; border-radius: 4px;
  transition: color .12s;
}
.sr-pw-eye:hover { color: var(--ink-0, #E8EBF0); }
.sr-pw-eye:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.sr-icon { width: 16px; height: 16px; }

/* toggle */
.sr-toggle {
  flex-shrink: 0;
  width: 40px; height: 22px;
  border-radius: 11px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-2, #171A20);
  cursor: pointer;
  position: relative;
  transition: background .15s, border-color .15s, box-shadow .15s;
  padding: 0;
  outline: none;
}
.sr-toggle:focus-visible {
  box-shadow: 0 0 0 3px rgba(92,168,255,.3);
  border-color: var(--azure, #5CA8FF);
}
.sr-toggle.on {
  background: var(--azure, #5CA8FF);
  border-color: var(--azure, #5CA8FF);
}
.sr-toggle-thumb {
  position: absolute;
  top: 2px; left: 2px;
  width: 16px; height: 16px;
  border-radius: 50%;
  background: var(--ink-2, #5C6470);
  transition: transform .15s, background .15s;
}
.sr-toggle.on .sr-toggle-thumb {
  transform: translateX(18px);
  background: var(--bg-00, var(--bg-0, #0E1014));
}
@media (prefers-reduced-motion: reduce) {
  .sr-toggle, .sr-toggle-thumb { transition: none; }
}
</style>
