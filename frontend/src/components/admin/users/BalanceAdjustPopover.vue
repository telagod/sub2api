<template>
  <!-- 小弹窗：调余额 -->
  <Teleport to="body">
    <div v-if="open" class="bal-backdrop" @click.self="$emit('close')" role="dialog" aria-modal="true" :aria-label="t('admin.balanceAdjustPopover.title')">
      <div class="bal-panel" @keydown.esc.stop="$emit('close')">
        <div class="bal-header">
          <span class="bal-title">{{ t('admin.balanceAdjustPopover.title') }}</span>
          <button class="bal-close" @click="$emit('close')" :aria-label="t('admin.balanceAdjustPopover.ariaClose')">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none"><path d="M2 2L12 12M12 2L2 12" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/></svg>
          </button>
        </div>
        <div class="bal-body">
          <div class="bal-cur">
            {{ t('admin.balanceAdjustPopover.currentBalance') }}<span class="q-money">${{ fmtBal(currentBalance) }}</span>
          </div>
          <div class="bal-field">
            <label class="bal-label">{{ t('admin.balanceAdjustPopover.operationLabel') }}</label>
            <div class="bal-ops">
              <button
                v-for="op in ops"
                :key="op.value"
                class="bal-op-btn"
                :class="{ 'bal-op-active': form.operation === op.value }"
                @click="form.operation = op.value"
              >{{ op.label }}</button>
            </div>
          </div>
          <div class="bal-field">
            <label class="bal-label">{{ t('admin.balanceAdjustPopover.amountLabel') }}</label>
            <div class="bal-input-wrap">
              <span class="bal-prefix">$</span>
              <input
                ref="amountRef"
                v-model.number="form.amount"
                type="number"
                step="any"
                min="0"
                class="bal-input q-focus-glow"
                placeholder="0.00"
              />
            </div>
          </div>
          <div class="bal-field" v-if="form.operation !== 'set'">
            <div class="bal-preview">
              → <span class="q-money">${{ fmtBal(previewBalance) }}</span>
            </div>
          </div>
          <div class="bal-field">
            <label class="bal-label">{{ t('admin.balanceAdjustPopover.notesLabel') }}</label>
            <textarea v-model="form.notes" rows="2" class="bal-textarea q-focus-glow" :placeholder="t('admin.balanceAdjustPopover.notesPlaceholder')"></textarea>
          </div>
        </div>
        <div class="bal-footer">
          <button class="bal-btn bal-btn-ghost" @click="$emit('close')">{{ t('admin.balanceAdjustPopover.cancelBtn') }}</button>
          <button
            class="bal-btn bal-btn-primary"
            :disabled="submitting || !form.amount"
            @click="submit"
          >{{ submitting ? t('admin.balanceAdjustPopover.submitting') : t('admin.balanceAdjustPopover.confirmBtn') }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { reactive, ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'

const props = defineProps<{
  open: boolean
  userId: number
  currentBalance: number
}>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'updated'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()
const amountRef = ref<HTMLInputElement | null>(null)
const submitting = ref(false)
const form = reactive({ amount: 0, operation: 'add' as 'add' | 'subtract' | 'set', notes: '' })
const ops = computed<{ value: 'add' | 'subtract' | 'set'; label: string }[]>(() => [
  { value: 'add', label: t('admin.balanceAdjustPopover.opAdd') },
  { value: 'subtract', label: t('admin.balanceAdjustPopover.opSubtract') },
  { value: 'set', label: t('admin.balanceAdjustPopover.opSet') },
])

const previewBalance = computed(() => {
  const a = form.amount || 0
  if (form.operation === 'add') return props.currentBalance + a
  if (form.operation === 'subtract') return props.currentBalance - a
  return a
})

function fmtBal(v: number) {
  if (!v && v !== 0) return '0.00'
  const s = v.toFixed(8).replace(/\.?0+$/, '')
  const parts = s.split('.')
  if (parts.length === 1) return s + '.00'
  if (parts[1].length < 2) return s + '0'
  return s
}

watch(() => props.open, async (v) => {
  if (v) {
    form.amount = 0; form.operation = 'add'; form.notes = ''
    await nextTick()
    amountRef.value?.focus()
  }
})

async function submit() {
  if (!form.amount || form.amount <= 0) {
    appStore.showError(t('admin.balanceAdjustPopover.invalidAmount'))
    return
  }
  submitting.value = true
  try {
    await adminAPI.users.updateBalance(props.userId, form.amount, form.operation, form.notes)
    appStore.showSuccess(t('admin.balanceAdjustPopover.adjusted'))
    emit('updated')
    emit('close')
  } catch (e: any) {
    appStore.showError(e?.response?.data?.detail || t('admin.balanceAdjustPopover.operationFailed'))
  } finally { submitting.value = false }
}
</script>

<style scoped>
.bal-backdrop {
  position: fixed; inset: 0; z-index: 99999;
  background: rgba(0, 0, 0, 0.55);
  display: flex; align-items: center; justify-content: center;
}
@media (prefers-reduced-motion: no-preference) {
  .bal-panel { animation: bal-in 0.18s ease; }
}
@keyframes bal-in {
  from { opacity: 0; transform: scale(0.96) translateY(4px); }
  to { opacity: 1; transform: none; }
}
.bal-panel {
  width: 340px;
  background: var(--bg-1);
  border: 1px solid var(--line-1);
  border-radius: var(--q-radius);
  box-shadow: 0 24px 64px rgba(0,0,0,.5), var(--edge-hi);
  display: flex; flex-direction: column;
  font-size: 13px; font-family: var(--font-ui, sans-serif);
  color: var(--ink-0);
}
.bal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 18px 12px;
  border-bottom: 1px solid var(--line-0);
}
.bal-title { font-weight: 700; font-size: 14px; }
.bal-close {
  background: none; border: none; cursor: pointer;
  color: var(--ink-2); padding: 4px;
  border-radius: 6px; display: flex; align-items: center; justify-content: center;
  transition: background 0.12s;
}
.bal-close:hover { background: var(--bg-2); color: var(--ink-0); }
.bal-body { padding: 16px 18px; display: flex; flex-direction: column; gap: 14px; }
.bal-cur { font-size: 12.5px; color: var(--ink-1); }
.bal-field { display: flex; flex-direction: column; gap: 6px; }
.bal-label { font-size: 11.5px; color: var(--ink-2); }
.bal-ops { display: flex; gap: 6px; }
.bal-op-btn {
  flex: 1; padding: 6px 10px; border-radius: 8px;
  border: 1px solid var(--line-1); background: var(--bg-2);
  color: var(--ink-1); font-size: 12px; cursor: pointer;
  transition: all 0.13s; font-family: inherit;
}
.bal-op-active {
  border-color: rgba(92,168,255,.6);
  background: var(--azure-dim);
  color: var(--azure);
}
.bal-input-wrap { display: flex; align-items: center; background: var(--bg-2); border: 1px solid var(--line-1); border-radius: 8px; overflow: hidden; }
.bal-prefix { padding: 0 10px; color: var(--ink-2); font-size: 13px; user-select: none; }
.bal-input {
  flex: 1; background: transparent; border: none; outline: none;
  padding: 8px 10px 8px 0; font-size: 13px; color: var(--ink-0);
  font-family: 'IBM Plex Mono', monospace;
}
.bal-input:focus { outline: none; }
.bal-input-wrap:focus-within {
  border-color: rgba(92,168,255,.55);
  box-shadow: var(--glow-focus);
}
.bal-preview { font-size: 12.5px; color: var(--ink-1); padding: 4px 0; }
.bal-textarea {
  background: var(--bg-2); border: 1px solid var(--line-1); border-radius: 8px;
  padding: 8px 12px; font-size: 12.5px; color: var(--ink-0);
  resize: none; outline: none; font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.15s;
}
.bal-footer {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 12px 18px;
  border-top: 1px solid var(--line-0);
}
.bal-btn {
  padding: 7px 16px; border-radius: 8px; font-size: 12.5px;
  font-weight: 600; cursor: pointer; border: 1px solid transparent;
  font-family: inherit; transition: all 0.13s;
}
.bal-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.bal-btn-ghost {
  background: transparent; border-color: var(--line-1); color: var(--ink-1);
}
.bal-btn-ghost:hover { background: var(--bg-2); color: var(--ink-0); }
.bal-btn-primary {
  background: var(--azure-dim); border-color: rgba(92,168,255,.5); color: var(--azure);
}
.bal-btn-primary:not(:disabled):hover {
  background: rgba(92,168,255,.25); box-shadow: 0 0 14px rgba(92,168,255,.2);
}
</style>
