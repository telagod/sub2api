<template>
  <div class="qn-body">
    <!-- enabled switch -->
    <div class="qn-row-switch">
      <label class="qn-label">{{ t('admin.settings.quotaNotify.enabled') }}</label>
      <Toggle v-model="localEnabled" />
    </div>

    <!-- email list -->
    <template v-if="localEnabled">
      <div class="qn-email-list">
        <div
          v-for="(entry, index) in localEmails"
          :key="index"
          class="qn-email-row"
        >
          <!-- per-item mini toggle -->
          <button
            type="button"
            class="qn-mini-toggle"
            :class="{ on: !entry.disabled }"
            :aria-checked="!entry.disabled"
            role="switch"
            @click="toggleEntryDisabled(index)"
          >
            <span class="qn-mini-thumb" />
          </button>
          <input
            v-model="entry.email"
            type="email"
            class="qn-input"
            :placeholder="t('admin.settings.quotaNotify.emailPlaceholder')"
            @input="emitEmails"
          />
          <button
            type="button"
            class="qn-icon-btn"
            :aria-label="t('common.remove')"
            @click="removeEmail(index)"
          >
            <Icon name="x" size="xs" class="qn-icon" />
          </button>
        </div>

        <button
          type="button"
          class="qn-add-btn"
          @click="addEmail"
        >
          + {{ t('admin.settings.quotaNotify.addEmail') }}
        </button>
      </div>

      <p class="qn-hint">{{ t('admin.settings.quotaNotify.emailsHint') }}</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'

interface QuotaEmail {
  email: string
  disabled?: boolean
}

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

const { t } = useI18n()

// ── helpers ────────────────────────────────────────────────────────────────────
const activeSource = computed(() => props.formValues ?? props.settings)

function cloneEmails(src: Record<string, unknown>): QuotaEmail[] {
  const raw = src['account_quota_notify_emails']
  if (!Array.isArray(raw)) return []
  return raw.map((e) => ({ ...(e as QuotaEmail) }))
}

// ── local state ────────────────────────────────────────────────────────────────
const localEnabled = ref<boolean>(!!activeSource.value['account_quota_notify_enabled'])
const localEmails = ref<QuotaEmail[]>(cloneEmails(activeSource.value))

// Sync enabled when parent resets
watch(
  () => activeSource.value['account_quota_notify_enabled'],
  (v) => { localEnabled.value = !!v },
)

// Sync emails when parent resets
watch(
  () => activeSource.value['account_quota_notify_emails'],
  (incoming) => {
    if (JSON.stringify(incoming) !== JSON.stringify(localEmails.value)) {
      localEmails.value = cloneEmails(activeSource.value)
    }
  },
  { deep: true },
)

// Propagate enabled changes up
watch(localEnabled, (v) => {
  emit('update:field', 'account_quota_notify_enabled', v)
})

function emitEmails() {
  emit('update:field', 'account_quota_notify_emails', localEmails.value.map((e) => ({ ...e })))
}

function toggleEntryDisabled(index: number) {
  localEmails.value = localEmails.value.map((e, i) =>
    i === index ? { ...e, disabled: !e.disabled } : e,
  )
  emitEmails()
}

function addEmail() {
  localEmails.value = [...localEmails.value, { email: '', disabled: false }]
  emitEmails()
}

function removeEmail(index: number) {
  localEmails.value = localEmails.value.filter((_, i) => i !== index)
  emitEmails()
}
</script>

<style scoped>
.qn-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

/* switch row */
.qn-row-switch { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.qn-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }

/* email list */
.qn-email-list { display: flex; flex-direction: column; gap: 8px; }
.qn-email-row { display: flex; align-items: center; gap: 8px; }

/* mini toggle (per-item) */
.qn-mini-toggle {
  flex-shrink: 0; width: 32px; height: 18px; border-radius: 9px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-2, #171A20);
  cursor: pointer; position: relative; transition: background .15s, border-color .15s; padding: 0; outline: none;
}
.qn-mini-toggle:focus-visible { box-shadow: 0 0 0 3px rgba(92,168,255,.3); border-color: var(--azure, #5CA8FF); }
.qn-mini-toggle.on { background: var(--azure, #5CA8FF); border-color: var(--azure, #5CA8FF); }
.qn-mini-thumb {
  position: absolute; top: 2px; left: 2px; width: 12px; height: 12px;
  border-radius: 50%; background: var(--ink-2, #5C6470);
  transition: transform .15s, background .15s;
}
.qn-mini-toggle.on .qn-mini-thumb { transform: translateX(14px); background: var(--bg-0, #0C0E12); }

/* input */
.qn-input {
  flex: 1; min-width: 0; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.qn-input:focus, .qn-input:focus-visible { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }

/* icon remove button */
.qn-icon-btn {
  flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 6px;
  border: 1px solid transparent; background: transparent;
  color: var(--ink-2, #5C6470); cursor: pointer; transition: color .12s, background .12s, border-color .12s;
}
.qn-icon-btn:hover { color: var(--bad, #F25C69); background: rgba(242,92,105,.1); border-color: rgba(242,92,105,.25); }
.qn-icon-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.qn-icon { width: 14px; height: 14px; }

/* add button */
.qn-add-btn {
  align-self: flex-start; padding: 6px 12px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: transparent;
  color: var(--ink-1, #97A0AF); font-size: 12.5px; font-weight: 500;
  cursor: pointer; font-family: inherit; transition: border-color .15s, color .15s, background .15s;
}
.qn-add-btn:hover { border-color: var(--azure, #5CA8FF); color: var(--azure, #5CA8FF); background: rgba(92,168,255,.07); }
.qn-add-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

/* hint */
.qn-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }

@media (prefers-reduced-motion: reduce) {
  .qn-mini-toggle, .qn-mini-thumb { transition: none; }
}
</style>
