<template>
  <div class="ppl-body">
    <!-- payment_enabled_types badge toggles -->
    <div class="ppl-type-section">
      <label class="ppl-label">{{ t('admin.settings.payment.enabledPaymentTypes') }}</label>
      <div class="ppl-badges">
        <button
          v-for="pt in allPaymentTypes"
          :key="pt.value"
          type="button"
          class="ppl-badge"
          :class="{ active: isEnabled(pt.value) }"
          @click="toggleType(pt.value)"
        >
          {{ pt.label }}
        </button>
      </div>
      <p class="ppl-hint">{{ t('admin.settings.payment.enabledPaymentTypesHint') }}</p>
    </div>

    <!-- Existing PaymentProviderList (reused as-is) -->
    <PaymentProviderList
      v-if="paymentEnabled"
      :providers="providers"
      :loading="loading"
      :can-create="enabledTypes.length > 0"
      :enabled-payment-types="enabledTypes"
      :all-payment-types="allPaymentTypes"
      :redirect-label="t('admin.settings.payment.easypayRedirect')"
      @refresh="load"
      @create="onCreateRequest"
      @edit="onEditRequest"
      @delete="onDeleteRequest"
      @toggle-field="onToggleField"
      @toggle-type="onToggleProviderType"
      @reorder="onReorder"
    />

    <!-- Provider dialog rendered via async component to keep bundle lean -->
    <component
      :is="ProviderDialog"
      v-if="dialog.open"
      ref="dialogRef"
      :show="dialog.open"
      :saving="saving"
      :editing="dialog.provider"
      :all-key-options="providerKeyOptions"
      :enabled-key-options="enabledProviderKeyOptions"
      :all-payment-types="allPaymentTypes"
      :redirect-label="t('admin.settings.payment.easypayRedirect')"
      @close="dialog.open = false"
      @save="onSave"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, defineAsyncComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import PaymentProviderList from '@/components/payment/PaymentProviderList.vue'
import adminAPI from '@/api/admin'
import type { ProviderInstance } from '@/types/payment'
import type { TypeOption } from '@/components/payment/providerConfig'

const ProviderDialog = defineAsyncComponent(
  () => import('@/components/payment/PaymentProviderDialog.vue'),
)

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

// ─── payment_enabled_types (local mirror for badge toggles) ───────────────────

const allPaymentTypes: TypeOption[] = [
  { value: 'sub2apipay', label: 'sub2apipay' },
  { value: 'epay', label: 'epay' },
  { value: 'stripe', label: 'Stripe' },
  { value: 'alipay', label: 'Alipay' },
  { value: 'wechat', label: 'WeChat Pay' },
]

function parseEnabledTypes(src: Record<string, unknown>): string[] {
  const raw = src['payment_enabled_types']
  if (Array.isArray(raw)) return raw as string[]
  if (typeof raw === 'string' && raw)
    return raw.split(',').map((s) => s.trim()).filter(Boolean)
  return []
}

const localEnabledTypes = ref<string[]>(parseEnabledTypes(activeSource.value))

// Re-sync when parent resets
watch(
  () => activeSource.value['payment_enabled_types'],
  (incoming) => {
    const parsed = Array.isArray(incoming)
      ? (incoming as string[])
      : typeof incoming === 'string' && incoming
        ? incoming.split(',').map((s: string) => s.trim()).filter(Boolean)
        : []
    if (JSON.stringify(parsed) !== JSON.stringify(localEnabledTypes.value)) {
      localEnabledTypes.value = parsed
    }
  },
  { deep: true },
)

const enabledTypes = computed(() => localEnabledTypes.value)
const paymentEnabled = computed(() => !!activeSource.value['payment_enabled'])

function isEnabled(type: string): boolean {
  return localEnabledTypes.value.includes(type)
}

function toggleType(type: string) {
  const next = isEnabled(type)
    ? localEnabledTypes.value.filter((t) => t !== type)
    : [...localEnabledTypes.value, type]
  localEnabledTypes.value = next
  // Propagate to parent form — no server reload needed
  emit('update:field', 'payment_enabled_types', next)
}

// ─── Provider key options ─────────────────────────────────────────────────────

const providerKeyOptions = computed<TypeOption[]>(() => [
  { value: 'easypay', label: t('admin.settings.payment.providerEasypay') },
  { value: 'alipay', label: t('admin.settings.payment.providerAlipay') },
  { value: 'wxpay', label: t('admin.settings.payment.providerWxpay') },
  { value: 'stripe', label: t('admin.settings.payment.providerStripe') },
  { value: 'airwallex', label: t('admin.settings.payment.providerAirwallex') },
])

const enabledProviderKeyOptions = computed<TypeOption[]>(() => {
  return providerKeyOptions.value.filter((opt) => enabledTypes.value.includes(opt.value))
})

// ─── Provider list ────────────────────────────────────────────────────────────

const providers = ref<ProviderInstance[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogRef = ref<{ loadProvider: (p: ProviderInstance) => void; reset: (key: string) => void } | null>(null)

const dialog = ref<{
  open: boolean
  provider: ProviderInstance | null
}>({ open: false, provider: null })

async function load() {
  loading.value = true
  try {
    const res = await adminAPI.payment.getProviders()
    providers.value = (res.data ?? []) as ProviderInstance[]
  } finally {
    loading.value = false
  }
}

function onCreateRequest() {
  dialog.value = { open: true, provider: null }
}

function onEditRequest(p: ProviderInstance) {
  dialog.value = { open: true, provider: p }
}

async function onDeleteRequest(p: ProviderInstance) {
  if (!confirm(t('common.confirmDelete'))) return
  await adminAPI.payment.deleteProvider(p.id)
  await load()
}

async function onToggleField(
  p: ProviderInstance,
  field: 'enabled' | 'refund_enabled' | 'allow_user_refund',
) {
  const payload: Partial<ProviderInstance> = { [field]: !p[field] }
  if (field === 'refund_enabled' && !payload[field]) {
    payload['allow_user_refund'] = false
  }
  await adminAPI.payment.updateProvider(p.id, payload)
  await load()
}

async function onToggleProviderType(p: ProviderInstance, type: string) {
  const current = p.supported_types ?? []
  const updated = current.includes(type)
    ? current.filter((t) => t !== type)
    : [...current, type]
  await adminAPI.payment.updateProvider(p.id, { supported_types: updated })
  await load()
}

async function onReorder(updates: { id: number; sort_order: number }[]) {
  await Promise.all(
    updates.map((u) =>
      adminAPI.payment.updateProvider(u.id, { sort_order: u.sort_order }),
    ),
  )
  await load()
}

async function onSave(payload: Partial<ProviderInstance>) {
  saving.value = true
  try {
    if (dialog.value.provider) {
      await adminAPI.payment.updateProvider(dialog.value.provider.id, payload)
    } else {
      await adminAPI.payment.createProvider(payload)
    }
    dialog.value.open = false
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  if (paymentEnabled.value) load()
})
</script>

<style scoped>
.ppl-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

/* type section */
.ppl-type-section { display: flex; flex-direction: column; gap: 8px; }
.ppl-label { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.ppl-badges { display: flex; flex-wrap: wrap; gap: 8px; }

/* badge toggle buttons */
.ppl-badge {
  padding: 5px 12px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-1, #101216);
  color: var(--ink-1, #97A0AF);
  font-size: 12.5px; font-weight: 500; font-family: inherit;
  cursor: pointer; transition: border-color .15s, color .15s, background .15s, box-shadow .15s;
}
.ppl-badge:hover:not(.active) {
  border-color: rgba(92,168,255,.35); color: var(--ink-0, #E8EBF0); background: var(--bg-2, #171A20);
}
.ppl-badge.active {
  border-color: var(--azure, #5CA8FF);
  background: rgba(92,168,255,.12);
  color: var(--azure, #5CA8FF);
  box-shadow: 0 0 0 1px rgba(92,168,255,.2);
}
.ppl-badge:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

/* hint */
.ppl-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }
</style>
