<template>
  <div class="acu-body">
    <!-- Disabled state: affiliate feature not enabled -->
    <div v-if="!affiliateEnabled" class="acu-disabled">
      {{ t('admin.settings.features.affiliate.customUsers.disabledHint') }}
    </div>

    <template v-else>
    <!-- toolbar -->
    <div class="acu-toolbar">
      <input
        v-model="state.search"
        type="text"
        class="acu-search"
        :placeholder="t('admin.settings.features.affiliate.customUsers.searchPlaceholder')"
        @input="onSearchInput"
      />
      <button
        v-if="state.selected.length > 0"
        type="button"
        class="acu-btn acu-btn--secondary"
        @click="openBatchModal"
      >
        {{ t('admin.settings.features.affiliate.customUsers.batchButton', { count: state.selected.length }) }}
      </button>
      <button
        type="button"
        class="acu-btn acu-btn--primary"
        @click="openAddModal"
      >
        + {{ t('admin.settings.features.affiliate.customUsers.addButton') }}
      </button>
    </div>

    <!-- table -->
    <div class="acu-table-wrap">
      <table class="acu-table">
        <thead>
          <tr>
            <th class="acu-th acu-th--chk">
              <input
                type="checkbox"
                class="acu-checkbox"
                :checked="state.entries.length > 0 && state.selected.length === state.entries.length"
                @change="toggleSelectAll"
              />
            </th>
            <th class="acu-th">{{ t('admin.settings.features.affiliate.customUsers.col.email') }}</th>
            <th class="acu-th">{{ t('admin.settings.features.affiliate.customUsers.col.username') }}</th>
            <th class="acu-th">{{ t('admin.settings.features.affiliate.customUsers.col.code') }}</th>
            <th class="acu-th">{{ t('admin.settings.features.affiliate.customUsers.col.rate') }}</th>
            <th class="acu-th">{{ t('admin.settings.features.affiliate.customUsers.col.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="state.loading">
            <td colspan="6" class="acu-td acu-td--center">{{ t('common.loading') }}</td>
          </tr>
          <tr v-else-if="state.entries.length === 0">
            <td colspan="6" class="acu-td acu-td--center acu-td--muted">
              {{ t('admin.settings.features.affiliate.customUsers.empty') }}
            </td>
          </tr>
          <tr v-for="entry in state.entries" :key="entry.user_id" class="acu-row">
            <td class="acu-td acu-td--chk">
              <input
                type="checkbox"
                class="acu-checkbox"
                :checked="state.selected.includes(entry.user_id)"
                @change="toggleSelect(entry.user_id)"
              />
            </td>
            <td class="acu-td">{{ entry.email }}</td>
            <td class="acu-td acu-td--muted">{{ entry.username }}</td>
            <td class="acu-td acu-td--mono">
              {{ entry.aff_code }}
              <span v-if="entry.aff_code_custom" class="acu-badge">
                {{ t('admin.settings.features.affiliate.customUsers.customBadge') }}
              </span>
            </td>
            <td class="acu-td">
              <span v-if="entry.aff_rebate_rate_percent != null">{{ entry.aff_rebate_rate_percent }}%</span>
              <span v-else class="acu-td--muted">{{ t('admin.settings.features.affiliate.customUsers.useGlobal') }}</span>
            </td>
            <td class="acu-td">
              <div class="acu-actions">
                <button type="button" class="acu-link" @click="openEditModal(entry)">
                  {{ t('common.edit') }}
                </button>
                <button type="button" class="acu-link acu-link--danger" @click="askReset(entry)">
                  {{ t('common.delete') }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- pagination -->
    <div v-if="state.total > state.pageSize" class="acu-pagination">
      <span class="acu-pagination__total">
        {{ t('admin.settings.features.affiliate.customUsers.totalLabel', { total: state.total }) }}
      </span>
      <div class="acu-pagination__nav">
        <button
          type="button"
          class="acu-btn acu-btn--secondary acu-btn--sm"
          :disabled="state.page <= 1"
          @click="changePage(state.page - 1)"
        >
          {{ t('pagination.previous') }}
        </button>
        <span class="acu-pagination__page">
          {{ state.page }} / {{ Math.max(1, Math.ceil(state.total / state.pageSize)) }}
        </span>
        <button
          type="button"
          class="acu-btn acu-btn--secondary acu-btn--sm"
          :disabled="state.page >= Math.ceil(state.total / state.pageSize)"
          @click="changePage(state.page + 1)"
        >
          {{ t('pagination.next') }}
        </button>
      </div>
    </div>

    <!-- add/edit modal -->
    <div
      v-if="modal.open"
      class="acu-overlay"
      @click.self="closeModal"
    >
      <div class="acu-modal">
        <h3 class="acu-modal__title">
          {{ modal.mode === 'add'
            ? t('admin.settings.features.affiliate.modal.addTitle')
            : t('admin.settings.features.affiliate.modal.editTitle') }}
        </h3>

        <div class="acu-modal__body">
          <!-- user picker (add mode) -->
          <div v-if="modal.mode === 'add'" class="acu-field">
            <label class="acu-label">{{ t('admin.settings.features.affiliate.modal.userLabel') }}</label>
            <!-- selected chip -->
            <div v-if="modal.selectedUser" class="acu-user-chip">
              <div class="acu-user-chip__info">
                <span class="acu-user-chip__email">{{ modal.selectedUser.email }}</span>
                <span class="acu-user-chip__name">({{ modal.selectedUser.username }})</span>
              </div>
              <button
                type="button"
                class="acu-user-chip__clear"
                :title="t('admin.settings.features.affiliate.modal.changeUser')"
                @click="clearSelectedUser"
              >×</button>
            </div>
            <!-- search input + dropdown -->
            <template v-else>
              <input
                v-model="modal.userQuery"
                type="text"
                class="acu-input"
                :placeholder="t('admin.settings.features.affiliate.modal.userPlaceholder')"
                @input="onUserSearchInput"
              />
              <div v-if="modal.userResults.length > 0" class="acu-user-dropdown">
                <button
                  v-for="u in modal.userResults"
                  :key="u.id"
                  type="button"
                  class="acu-user-option"
                  @click="selectUser(u)"
                >
                  {{ u.email }}
                  <span class="acu-user-option__name">({{ u.username }})</span>
                </button>
              </div>
            </template>
          </div>

          <!-- display user (edit mode) -->
          <div v-else class="acu-field">
            <label class="acu-label">{{ t('admin.settings.features.affiliate.modal.userLabel') }}</label>
            <input
              type="text"
              class="acu-input"
              :value="modal.editingEntry?.email ?? ''"
              disabled
            />
          </div>

          <!-- invite code -->
          <div class="acu-field">
            <label class="acu-label">{{ t('admin.settings.features.affiliate.modal.codeLabel') }}</label>
            <input
              v-model="modal.code"
              type="text"
              class="acu-input acu-input--mono"
              :placeholder="t('admin.settings.features.affiliate.modal.codePlaceholder')"
              maxlength="32"
            />
            <p class="acu-hint">{{ t('admin.settings.features.affiliate.modal.codeHint') }}</p>
          </div>

          <!-- rebate rate -->
          <div class="acu-field">
            <label class="acu-label">{{ t('admin.settings.features.affiliate.modal.rateLabel') }}</label>
            <div class="acu-rate-wrap">
              <input
                v-model="modal.rate"
                type="number"
                step="0.01"
                min="0"
                max="100"
                class="acu-input acu-input--rate"
                :placeholder="t('admin.settings.features.affiliate.modal.ratePlaceholder')"
              />
              <span class="acu-rate-suffix">%</span>
            </div>
            <p class="acu-hint">{{ t('admin.settings.features.affiliate.modal.rateHint') }}</p>
          </div>
        </div>

        <div class="acu-modal__footer">
          <p v-if="!modalCanSubmit" class="acu-modal__error">
            {{ t('admin.settings.features.affiliate.modal.errorEmpty') }}
          </p>
          <span v-else />
          <div class="acu-modal__btns">
            <button type="button" class="acu-btn acu-btn--secondary" @click="closeModal">
              {{ t('common.cancel') }}
            </button>
            <button
              type="button"
              class="acu-btn acu-btn--primary"
              :disabled="modal.saving || !modalCanSubmit"
              @click="submitModal"
            >
              {{ modal.saving ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- batch rate modal -->
    <div
      v-if="batchModal.open"
      class="acu-overlay"
      @click.self="batchModal.open = false"
    >
      <div class="acu-modal">
        <h3 class="acu-modal__title">
          {{ t('admin.settings.features.affiliate.batchModal.title', { count: state.selected.length }) }}
        </h3>
        <p class="acu-modal__hint">
          {{ t('admin.settings.features.affiliate.batchModal.hint') }}
        </p>
        <div class="acu-rate-wrap">
          <input
            v-model="batchModal.rate"
            type="number"
            step="0.01"
            min="0"
            max="100"
            class="acu-input acu-input--rate"
            :placeholder="t('admin.settings.features.affiliate.batchModal.placeholder')"
          />
          <span class="acu-rate-suffix">%</span>
        </div>
        <p class="acu-hint">{{ t('admin.settings.features.affiliate.batchModal.clearHint') }}</p>
        <div class="acu-modal__footer acu-modal__footer--right">
          <button type="button" class="acu-btn acu-btn--secondary" @click="batchModal.open = false">
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="acu-btn acu-btn--primary"
            :disabled="batchModal.saving"
            @click="submitBatchModal"
          >
            {{ batchModal.saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- confirm dialog -->
    <ConfirmDialog
      :show="confirmDialog.show"
      :title="confirmDialog.title"
      :message="confirmDialog.message"
      :confirm-text="confirmDialog.confirmText"
      danger
      @confirm="handleConfirm"
      @cancel="cancelConfirm"
    />
    </template><!-- /v-else affiliate enabled -->
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { affiliatesAPI, type AffiliateAdminEntry, type SimpleUser } from '@/api/admin/affiliates'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = defineProps<{
  settings?: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const { t } = useI18n()
const appStore = useAppStore()

// Guard: mirror the SettingsView v-if="form.affiliate_enabled" condition.
// formValues reflects the current (possibly unsaved) form state; settings is the
// saved snapshot. Fall back to settings when formValues is absent.
const affiliateEnabled = computed(() => {
  const source = props.formValues ?? props.settings ?? {}
  return source['affiliate_enabled'] === true
})

// ── List state ─────────────────────────────────────────────────────────────

interface ListState {
  loading: boolean
  entries: AffiliateAdminEntry[]
  total: number
  page: number
  pageSize: number
  search: string
  selected: number[]
  searchTimer: number | null
}

const state = reactive<ListState>({
  loading: false,
  entries: [],
  total: 0,
  page: 1,
  pageSize: 20,
  search: '',
  selected: [],
  searchTimer: null,
})

// ── Add / edit modal ───────────────────────────────────────────────────────

interface ModalState {
  open: boolean
  mode: 'add' | 'edit'
  saving: boolean
  userQuery: string
  userResults: SimpleUser[]
  selectedUser: SimpleUser | null
  editingEntry: AffiliateAdminEntry | null
  code: string
  // string|number because <input type="number"> coerces v-model to Number
  rate: string | number
  searchTimer: number | null
}

const modal = reactive<ModalState>({
  open: false,
  mode: 'add',
  saving: false,
  userQuery: '',
  userResults: [],
  selectedUser: null,
  editingEntry: null,
  code: '',
  rate: '',
  searchTimer: null,
})

// ── Batch rate modal ───────────────────────────────────────────────────────

const batchModal = reactive<{
  open: boolean
  saving: boolean
  rate: string | number
}>({
  open: false,
  saving: false,
  rate: '',
})

// ── Confirm dialog ─────────────────────────────────────────────────────────

const confirmDialog = reactive<{
  show: boolean
  title: string
  message: string
  confirmText: string
  pending: (() => Promise<unknown>) | null
}>({
  show: false,
  title: '',
  message: '',
  confirmText: '',
  pending: null,
})

// ── Helpers ────────────────────────────────────────────────────────────────

function debounce(slot: { searchTimer: number | null }, ms: number, fn: () => void) {
  if (slot.searchTimer != null) window.clearTimeout(slot.searchTimer)
  slot.searchTimer = window.setTimeout(fn, ms)
}

// parseRebateRate validates 0-100 numeric input.
// Returns: parsed number on success, null when empty (caller decides semantics),
//          or undefined on bad input (toast already shown).
function parseRebateRate(raw: unknown): number | null | undefined {
  const s = String(raw ?? '').trim()
  if (s === '') return null
  const v = Number(s)
  if (Number.isNaN(v) || v < 0 || v > 100) {
    appStore.showError(t('admin.settings.features.affiliate.modal.errorBadRate'))
    return undefined
  }
  return v
}

// ── Load ───────────────────────────────────────────────────────────────────

async function load() {
  state.loading = true
  try {
    const res = await affiliatesAPI.listUsers({
      page: state.page,
      page_size: state.pageSize,
      search: state.search,
    })
    state.entries = res.items ?? []
    state.total = res.total ?? 0
    // Drop selections that are no longer visible
    const visible = new Set(state.entries.map((e) => e.user_id))
    state.selected = state.selected.filter((id) => visible.has(id))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    state.loading = false
  }
}

// Load on mount only when affiliate is already enabled; also react to the
// toggle turning on (mirrors the watch in SettingsView.vue ~9633-9640).
onMounted(() => {
  if (affiliateEnabled.value) load()
})

watch(affiliateEnabled, (enabled, prev) => {
  if (enabled && !prev) load()
})

// ── Table interactions ─────────────────────────────────────────────────────

function onSearchInput() {
  debounce(state, 300, () => {
    state.page = 1
    load()
  })
}

function changePage(page: number) {
  if (page < 1) return
  state.page = page
  load()
}

function toggleSelectAll(e: Event) {
  const checked = (e.target as HTMLInputElement).checked
  state.selected = checked ? state.entries.map((e) => e.user_id) : []
}

function toggleSelect(userId: number) {
  const idx = state.selected.indexOf(userId)
  if (idx >= 0) state.selected.splice(idx, 1)
  else state.selected.push(userId)
}

// ── Add/edit modal ─────────────────────────────────────────────────────────

function openAddModal() {
  modal.open = true
  modal.mode = 'add'
  modal.saving = false
  modal.userQuery = ''
  modal.userResults = []
  modal.selectedUser = null
  modal.editingEntry = null
  modal.code = ''
  modal.rate = ''
}

function openEditModal(entry: AffiliateAdminEntry) {
  modal.open = true
  modal.mode = 'edit'
  modal.saving = false
  modal.userQuery = ''
  modal.userResults = []
  modal.selectedUser = null
  modal.editingEntry = entry
  modal.code = entry.aff_code_custom ? entry.aff_code : ''
  modal.rate = entry.aff_rebate_rate_percent != null ? String(entry.aff_rebate_rate_percent) : ''
}

function closeModal() {
  modal.open = false
  if (modal.searchTimer != null) {
    window.clearTimeout(modal.searchTimer)
    modal.searchTimer = null
  }
}

function onUserSearchInput() {
  const q = modal.userQuery.trim()
  if (!q) {
    modal.userResults = []
    return
  }
  debounce(modal, 300, async () => {
    try {
      modal.userResults = await affiliatesAPI.lookupUsers(q)
    } catch (err) {
      appStore.showError(extractApiErrorMessage(err, t('common.error')))
    }
  })
}

function selectUser(user: SimpleUser) {
  modal.selectedUser = user
  modal.userQuery = ''
  modal.userResults = []
}

function clearSelectedUser() {
  modal.selectedUser = null
}

// Guard: add mode needs a user picked; at least one field must be filled.
// Edit mode with empty rate field = clear the custom rate (valid if one exists).
const modalCanSubmit = computed(() => {
  if (modal.mode === 'add') {
    if (!modal.selectedUser) return false
  } else if (!modal.editingEntry) {
    return false
  }
  const codeFilled = modal.code.trim() !== ''
  const rateFilled = String(modal.rate ?? '').trim() !== ''
  if (codeFilled || rateFilled) return true
  return modal.mode === 'edit' && modal.editingEntry?.aff_rebate_rate_percent != null
})

async function submitModal() {
  if (!modalCanSubmit.value) {
    appStore.showError(t('admin.settings.features.affiliate.modal.errorEmpty'))
    return
  }

  const userId = modal.mode === 'add' ? modal.selectedUser!.id : modal.editingEntry!.user_id
  const payload: Parameters<typeof affiliatesAPI.updateUserSettings>[1] = {}

  const codeRaw = modal.code.trim()
  if (codeRaw) payload.aff_code = codeRaw.toUpperCase()

  const rateInput = parseRebateRate(modal.rate)
  if (rateInput === undefined) return // toast shown
  if (rateInput === null) {
    if (modal.mode === 'edit' && modal.editingEntry?.aff_rebate_rate_percent != null) {
      payload.clear_rebate_rate = true
    }
  } else {
    payload.aff_rebate_rate_percent = rateInput
  }

  modal.saving = true
  try {
    await affiliatesAPI.updateUserSettings(userId, payload)
    appStore.showSuccess(t('common.saved'))
    closeModal()
    state.page = 1
    await load()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    modal.saving = false
  }
}

// ── Batch modal ────────────────────────────────────────────────────────────

function openBatchModal() {
  if (state.selected.length === 0) return
  batchModal.open = true
  batchModal.rate = ''
}

async function submitBatchModal() {
  const rateInput = parseRebateRate(batchModal.rate)
  if (rateInput === undefined) return
  const userIDs = [...state.selected]
  const payload: Parameters<typeof affiliatesAPI.batchSetRate>[0] =
    rateInput === null
      ? { user_ids: userIDs, clear: true }
      : { user_ids: userIDs, aff_rebate_rate_percent: rateInput }

  batchModal.saving = true
  try {
    await affiliatesAPI.batchSetRate(payload)
    appStore.showSuccess(t('common.saved'))
    batchModal.open = false
    state.selected = []
    await load()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  } finally {
    batchModal.saving = false
  }
}

// ── Confirm / reset ────────────────────────────────────────────────────────

function askReset(entry: AffiliateAdminEntry) {
  confirmDialog.title = t('admin.settings.features.affiliate.customUsers.resetTitle')
  confirmDialog.message = t('admin.settings.features.affiliate.customUsers.resetMessage', {
    email: entry.email || `#${entry.user_id}`,
  })
  confirmDialog.confirmText = t('common.delete')
  confirmDialog.pending = () => affiliatesAPI.clearUserSettings(entry.user_id)
  confirmDialog.show = true
}

async function handleConfirm() {
  const fn = confirmDialog.pending
  confirmDialog.show = false
  confirmDialog.pending = null
  if (!fn) return
  try {
    await fn()
    appStore.showSuccess(t('common.saved'))
    await load()
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('common.error')))
  }
}

function cancelConfirm() {
  confirmDialog.show = false
  confirmDialog.pending = null
}
</script>

<style scoped>
/* ── Layout ──────────────────────────────────────────────────────────────── */
.acu-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* ── Disabled state ──────────────────────────────────────────────────────── */
.acu-disabled {
  padding: 12px 16px;
  font-size: 12.5px;
  color: var(--ink-2, #5C6470);
  border: 1px dashed var(--line-1, #2F3540);
  border-radius: 8px;
}

/* ── Toolbar ─────────────────────────────────────────────────────────────── */
.acu-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.acu-search {
  flex: 1;
  min-width: 160px;
  padding: 7px 11px;
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
.acu-search:focus {
  border-color: var(--azure, #5CA8FF);
  box-shadow: 0 0 0 3px rgba(92,168,255,.14);
}
.acu-search::placeholder { color: var(--ink-2, #5C6470); }

/* ── Buttons ─────────────────────────────────────────────────────────────── */
.acu-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 7px 14px;
  border-radius: 8px;
  font-size: 12.5px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  user-select: none;
  border: 1px solid transparent;
  transition: background .15s, box-shadow .15s, opacity .15s;
  white-space: nowrap;
}
.acu-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.acu-btn:disabled { opacity: .55; cursor: not-allowed; }

.acu-btn--primary {
  border-color: var(--azure, #5CA8FF);
  background: linear-gradient(180deg, rgba(92,168,255,.18) 0%, rgba(92,168,255,.08) 100%);
  color: var(--azure, #5CA8FF);
  box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 2px 6px rgba(92,168,255,.12);
}
.acu-btn--primary:hover:not(:disabled) {
  background: linear-gradient(180deg, rgba(92,168,255,.28) 0%, rgba(92,168,255,.14) 100%);
}

.acu-btn--secondary {
  border-color: var(--line-1, #2F3540);
  background: transparent;
  color: var(--ink-1, #97A0AF);
}
.acu-btn--secondary:hover:not(:disabled) {
  background: rgba(255,255,255,.04);
  color: var(--ink-0, #E8EBF0);
}

.acu-btn--sm { padding: 5px 11px; font-size: 12px; }

/* ── Table ───────────────────────────────────────────────────────────────── */
.acu-table-wrap {
  overflow: hidden;
  border-radius: 10px;
  border: 1px solid var(--line-1, #2F3540);
}

.acu-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: auto;
}

.acu-th {
  padding: 8px 12px;
  text-align: left;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: .04em;
  text-transform: uppercase;
  color: var(--ink-2, #5C6470);
  background: var(--bg-1, #13161B);
  border-bottom: 1px solid var(--line-1, #2F3540);
}
.acu-th--chk { width: 36px; }

.acu-row { transition: background .1s; }
.acu-row:hover { background: rgba(255,255,255,.025); }

.acu-td {
  padding: 8px 12px;
  font-size: 13px;
  color: var(--ink-0, #E8EBF0);
  border-bottom: 1px solid var(--line-0, #20242C);
}
.acu-row:last-child .acu-td { border-bottom: none; }

.acu-td--chk { width: 36px; }
.acu-td--center { text-align: center; padding: 24px 12px; }
.acu-td--muted { color: var(--ink-2, #5C6470); }
.acu-td--mono { font-family: var(--font-mono, ui-monospace, monospace); font-size: 12px; }

.acu-checkbox { accent-color: var(--azure, #5CA8FF); cursor: pointer; }

.acu-badge {
  display: inline-block;
  margin-left: 4px;
  padding: 1px 5px;
  border-radius: 4px;
  background: rgba(92,168,255,.15);
  border: 1px solid rgba(92,168,255,.3);
  font-size: 10px;
  font-weight: 500;
  color: var(--azure, #5CA8FF);
  vertical-align: middle;
}

.acu-actions { display: flex; align-items: center; gap: 10px; }

.acu-link {
  background: none;
  border: none;
  padding: 0;
  font-size: 12.5px;
  font-family: inherit;
  cursor: pointer;
  color: var(--azure, #5CA8FF);
  transition: opacity .12s;
}
.acu-link:hover { opacity: .75; }
.acu-link:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.acu-link--danger { color: var(--bad, #F25C69); }

/* ── Pagination ──────────────────────────────────────────────────────────── */
.acu-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
}

.acu-pagination__total { font-size: 12px; color: var(--ink-2, #5C6470); }
.acu-pagination__nav { display: flex; align-items: center; gap: 8px; }
.acu-pagination__page { font-size: 12px; color: var(--ink-2, #5C6470); min-width: 48px; text-align: center; }

/* ── Overlay / modal ─────────────────────────────────────────────────────── */
.acu-overlay {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,.55);
  padding: 16px;
}

.acu-modal {
  width: 100%;
  max-width: 440px;
  border-radius: 12px;
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
  border: 1px solid var(--line-0, #20242C);
  box-shadow: 0 20px 60px rgba(0,0,0,.55);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.acu-modal__title {
  font-size: 15px;
  font-weight: 600;
  color: var(--ink-0, #E8EBF0);
  margin: 0;
}

.acu-modal__hint {
  font-size: 12.5px;
  color: var(--ink-2, #5C6470);
  margin: 0;
  line-height: 1.55;
}

.acu-modal__body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.acu-modal__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  padding-top: 4px;
  border-top: 1px solid var(--line-0, #20242C);
}
.acu-modal__footer--right { justify-content: flex-end; }

.acu-modal__btns { display: flex; gap: 8px; }

.acu-modal__error {
  font-size: 11.5px;
  color: var(--bad, #F25C69);
  margin: 0;
}

/* ── Form elements ───────────────────────────────────────────────────────── */
.acu-field { display: flex; flex-direction: column; gap: 5px; }

.acu-label {
  font-size: 11.5px;
  font-weight: 500;
  color: var(--ink-1, #97A0AF);
}

.acu-input {
  width: 100%;
  padding: 7px 11px;
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
.acu-input:focus { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.acu-input:disabled { opacity: .5; cursor: not-allowed; }
.acu-input--mono { font-family: var(--font-mono, ui-monospace, monospace); }
.acu-input--rate { padding-right: 32px; }

.acu-rate-wrap { position: relative; }
.acu-rate-suffix {
  position: absolute;
  right: 11px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 13px;
  color: var(--ink-2, #5C6470);
  pointer-events: none;
}

.acu-hint {
  font-size: 11px;
  color: var(--ink-2, #5C6470);
  line-height: 1.5;
  margin: 0;
}

/* ── User picker ─────────────────────────────────────────────────────────── */
.acu-user-chip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 11px;
  border-radius: 8px;
  border: 1px solid rgba(92,168,255,.3);
  background: rgba(92,168,255,.08);
}

.acu-user-chip__info { display: flex; align-items: baseline; gap: 4px; }
.acu-user-chip__email { font-size: 13px; font-weight: 500; color: var(--ink-0, #E8EBF0); }
.acu-user-chip__name { font-size: 11.5px; color: var(--ink-2, #5C6470); }

.acu-user-chip__clear {
  background: none;
  border: none;
  font-size: 18px;
  line-height: 1;
  color: var(--ink-2, #5C6470);
  cursor: pointer;
  padding: 0 2px;
  transition: color .12s;
}
.acu-user-chip__clear:hover { color: var(--bad, #F25C69); }

.acu-user-dropdown {
  margin-top: 4px;
  max-height: 160px;
  overflow-y: auto;
  border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-1, #13161B);
}

.acu-user-option {
  display: block;
  width: 100%;
  text-align: left;
  padding: 7px 11px;
  font-size: 13px;
  font-family: inherit;
  color: var(--ink-0, #E8EBF0);
  background: none;
  border: none;
  cursor: pointer;
  transition: background .1s;
}
.acu-user-option:hover { background: rgba(255,255,255,.05); }
.acu-user-option__name { font-size: 11.5px; color: var(--ink-2, #5C6470); }
</style>
