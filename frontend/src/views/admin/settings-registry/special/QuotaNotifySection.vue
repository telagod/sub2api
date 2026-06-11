<template>
  <div class="px-5 py-4 space-y-4">
    <!-- enabled switch -->
    <div class="flex items-center justify-between">
      <label class="text-sm font-medium text-foreground/85">
        {{ t('admin.settings.quotaNotify.enabled') }}
      </label>
      <Toggle v-model="localEnabled" />
    </div>

    <!-- email list -->
    <template v-if="localEnabled">
      <div class="space-y-2">
        <div
          v-for="(entry, index) in localEmails"
          :key="index"
          class="flex items-center gap-2"
        >
          <!-- per-item enabled toggle -->
          <label class="relative inline-flex shrink-0 cursor-pointer items-center">
            <input
              type="checkbox"
              :checked="!entry.disabled"
              class="sr-only peer"
              @change="toggleEntryDisabled(index)"
            />
            <div
              class="h-5 w-9 rounded-full bg-accent peer-checked:bg-primary-600 after:absolute after:left-[2px] after:top-[2px] after:h-4 after:w-4 after:rounded-full after:border after:border-border after:bg-white after:transition-all after:content-[''] peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:outline-none"
            ></div>
          </label>
          <input
            v-model="entry.email"
            type="email"
            class="input flex-1"
            :placeholder="t('admin.settings.quotaNotify.emailPlaceholder')"
            @input="emitEmails"
          />
          <button
            type="button"
            class="btn btn-secondary px-2"
            @click="removeEmail(index)"
          >
            <Icon name="x" size="xs" class="h-4 w-4" />
          </button>
        </div>

        <button
          type="button"
          class="btn btn-secondary btn-sm"
          @click="addEmail"
        >
          + {{ t('admin.settings.quotaNotify.addEmail') }}
        </button>
      </div>

      <p class="text-xs text-muted-foreground">
        {{ t('admin.settings.quotaNotify.emailsHint') }}
      </p>
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
