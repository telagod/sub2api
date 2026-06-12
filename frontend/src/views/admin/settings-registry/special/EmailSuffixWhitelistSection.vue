<template>
  <div class="esw-body">
    <!-- Tag chip container -->
    <div class="esw-chip-area">
      <div class="esw-chips">
        <span
          v-for="suffix in localTags"
          :key="suffix"
          class="esw-chip"
        >
          <span class="esw-chip-text">{{ suffix }}</span>
          <button
            type="button"
            class="esw-chip-del"
            @click="removeTag(suffix)"
          >
            <svg class="esw-chip-x" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </span>

        <!-- inline input -->
        <div class="esw-input-wrap">
          <input
            v-model="draft"
            type="text"
            class="esw-input"
            :placeholder="localTags.length === 0 ? t('admin.settings.registration.emailSuffixWhitelistPlaceholder') : ''"
            @input="onDraftInput"
            @keydown="onDraftKeydown"
            @blur="commitDraft"
            @paste.prevent="onPaste"
          />
        </div>
      </div>
    </div>

    <p class="esw-hint">{{ t('admin.settings.registration.emailSuffixWhitelistInputHint') }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  isRegistrationEmailSuffixDomainValid,
  normalizeRegistrationEmailSuffixDomain,
  normalizeRegistrationEmailSuffixDomains,
  parseRegistrationEmailSuffixWhitelistInput,
} from '@/utils/registrationEmailPolicy'

const { t } = useI18n()

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

// Prefer dirty form state over saved settings
const activeSource = computed(() => props.formValues ?? props.settings)

// ── Separator keys ─────────────────────────────────────────────────────────────
const separatorKeys = new Set([' ', ',', '，', 'Enter', 'Tab'])

// ── Helpers ────────────────────────────────────────────────────────────────────

function parseTagsFromSource(src: Record<string, unknown>): string[] {
  const raw = src['registration_email_suffix_whitelist']
  if (!Array.isArray(raw)) return []
  return normalizeRegistrationEmailSuffixDomains(raw as string[])
}

// ── Local state ────────────────────────────────────────────────────────────────

const localTags = ref<string[]>(parseTagsFromSource(activeSource.value))
const draft = ref('')

// Re-sync when parent resets (discard) or initial settings load
watch(
  () => activeSource.value['registration_email_suffix_whitelist'],
  (incoming) => {
    const next = normalizeRegistrationEmailSuffixDomains(
      Array.isArray(incoming) ? (incoming as string[]) : [],
    )
    if (JSON.stringify(next) !== JSON.stringify(localTags.value)) {
      localTags.value = next
    }
  },
  { deep: true },
)

// ── Emit ───────────────────────────────────────────────────────────────────────

function emitTags() {
  // Mirror the save transform in SettingsView: wildcard domains are kept as-is,
  // exact domains get @-prefixed so the backend's existing format is preserved.
  const canonical = localTags.value.map((suffix) =>
    suffix.startsWith('*.') ? suffix : `@${suffix}`,
  )
  emit('update:field', 'registration_email_suffix_whitelist', canonical)
}

// ── Tag management ─────────────────────────────────────────────────────────────

function addTag(raw: string) {
  const suffix = normalizeRegistrationEmailSuffixDomain(raw)
  if (!isRegistrationEmailSuffixDomainValid(suffix) || localTags.value.includes(suffix)) return
  localTags.value = [...localTags.value, suffix]
  emitTags()
}

function removeTag(suffix: string) {
  localTags.value = localTags.value.filter((t) => t !== suffix)
  emitTags()
}

function commitDraft() {
  if (!draft.value) return
  addTag(draft.value)
  draft.value = ''
}

// ── Event handlers ─────────────────────────────────────────────────────────────

function onDraftInput() {
  draft.value = normalizeRegistrationEmailSuffixDomain(draft.value)
}

function onDraftKeydown(event: KeyboardEvent) {
  if (event.isComposing) return

  if (separatorKeys.has(event.key)) {
    event.preventDefault()
    commitDraft()
    return
  }

  if (event.key === 'Backspace' && !draft.value && localTags.value.length > 0) {
    localTags.value = localTags.value.slice(0, -1)
    emitTags()
  }
}

function onPaste(event: ClipboardEvent) {
  const text = event.clipboardData?.getData('text') || ''
  if (!text.trim()) return
  const tokens = parseRegistrationEmailSuffixWhitelistInput(text)
  for (const token of tokens) {
    addTag(token)
  }
}
</script>

<style scoped>
.esw-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* chip area container */
.esw-chip-area {
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 10px;
  background: var(--bg-0, #0C0E12);
  padding: 8px 10px;
  transition: border-color .15s, box-shadow .15s;
}
.esw-chip-area:focus-within {
  border-color: var(--azure, #5CA8FF);
  box-shadow: 0 0 0 3px rgba(92,168,255,.14);
}

/* chips flex row */
.esw-chips {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

/* individual chip */
.esw-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px 3px 10px;
  border-radius: 6px;
  background: var(--bg-2, #171A20);
  border: 1px solid var(--line-1, #2F3540);
}

.esw-chip-text {
  font-size: 12px;
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  color: var(--ink-0, #E8EBF0);
  white-space: nowrap;
}

/* chip delete button */
.esw-chip-del {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 4px;
  border: none;
  background: transparent;
  color: var(--ink-2, #5C6470);
  cursor: pointer;
  padding: 0;
  transition: color .1s, background .1s;
}
.esw-chip-del:hover {
  color: var(--bad, #F25C69);
  background: rgba(242,92,105,.12);
}
.esw-chip-del:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 1px; }

.esw-chip-x { width: 10px; height: 10px; }

/* inline input */
.esw-input-wrap {
  display: flex;
  flex: 1;
  min-width: 180px;
}

.esw-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--ink-0, #E8EBF0);
  font-size: 12.5px;
  font-family: var(--font-mono, "IBM Plex Mono", monospace);
  outline: none;
  padding: 3px 4px;
}
.esw-input::placeholder { color: var(--ink-2, #5C6470); }

/* hint */
.esw-hint {
  font-size: 11px;
  color: var(--ink-2, #5C6470);
  line-height: 1.55;
  margin: 0;
}
</style>
