<template>
  <AppLayout>
    <div class="uq-root">
      <!-- 页头 -->
      <div class="uq-head">
        <div>
          <h1 class="uq-title">{{ t('admin.usersQuench.title') }}</h1>
          <p class="uq-desc">{{ t('admin.usersQuench.desc') }}</p>
        </div>
        <div class="uq-head-acts">
          <button class="uq-btn" @click="loadUsers">{{ t('admin.usersQuench.refresh') }}</button>
          <button class="uq-btn uq-btn-metal" @click="openCreateDrawer">{{ t('admin.usersQuench.createBtn') }}</button>
        </div>
      </div>

      <!-- 视图页签 -->
      <SavedViewTabs storage-key="admin_users" :current-state="savedViewState" :total-count="pagination.total" @apply="onApplyView" />

      <!-- 快速内置视图 -->
      <div class="uq-qtabs">
        <button v-for="qv in QUICK_VIEWS" :key="qv.id" class="uq-qtab" :class="{ on: activeQuickView === qv.id }" @click="applyQuickView(qv as any)">{{ qv.label }}</button>
      </div>

      <!-- 筛选栏 -->
      <UsersFilterBar
        v-model:search="searchInput"
        v-model:role="filterRole"
        v-model:status="filterStatus"
        v-model:density="density"
        @commit-search="commitSearch"
        @clear="clearFilters"
      />

      <!-- 表格卡片 -->
      <div class="uq-card">
        <DataTableV2
          :columns="(COLUMNS as any)"
          :rows="(users as unknown as Record<string, unknown>[])"
          :total="pagination.total"
          :loading="loading"
          :selectable="true"
          row-key="id"
          :density="density"
          :page="state.page"
          :page-size="state.pageSize"
          :sort="state.sort"
          :order="state.order"
          @row-click="onRowClick"
          @update:selected="onSelectionChange"
          @update:page="p => { state.page = p }"
          @update:sort="s => { state.sort = s; state.page = 1 }"
          @update:order="o => { state.order = o; state.page = 1 }"
        >
          <template #cell-email="{ row }">
            <div class="uq-cell-user">
              <div class="uq-av" :style="{ background: avatarColor(String(row.email)) }">{{ String(row.email).charAt(0).toUpperCase() }}</div>
              <div>
                <div class="uq-email">{{ row.email }}</div>
                <div class="uq-uname"><span v-if="row.username">@{{ row.username }}</span><span v-else class="uq-muted">—</span><span class="uq-uid"> · #{{ String(row.id).padStart(4, '0') }}</span></div>
              </div>
            </div>
          </template>

          <template #cell-role="{ value }">
            <span :class="['uq-badge', value === 'admin' ? 'uq-badge-azure' : 'uq-badge-dim']">{{ value === 'admin' ? t('admin.usersQuench.roleAdmin') : t('admin.usersQuench.roleUser') }}</span>
          </template>

          <template #cell-balance="{ row }">
            <div class="uq-bal">
              <span class="uq-money" :class="{ 'c-bad': Number(row.balance) < 1, 'c-warn': Number(row.balance) >= 1 && Number(row.balance) < 5 }">${{ Number(row.balance).toFixed(2) }}</span>
              <div class="uq-meter"><i :style="{ width: Math.min(100, Math.max(0, Number(row.balance))) + '%' }" :class="{ 'c-bad': Number(row.balance) < 1, 'c-warn': Number(row.balance) >= 1 && Number(row.balance) < 5 }"></i></div>
            </div>
          </template>

          <template #cell-concurrency="{ row }">
            <span class="uq-mono uq-muted">{{ row.current_concurrency ?? 0 }}/{{ row.concurrency }}</span>
          </template>

          <template #cell-status="{ value }">
            <span class="uq-status"><span class="uq-dot" :class="value === 'active' ? 'ok' : 'bad'"></span>{{ value === 'active' ? t('admin.usersQuench.statusActive') : t('admin.usersQuench.statusDisabled') }}</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="uq-mono uq-muted uq-xs">{{ fmtDate(String(value)) }}</span>
          </template>

          <template #cell-_actions="{ row }">
            <div class="uq-acts">
              <button class="uq-ib" :title="t('admin.usersQuench.actionAdjustBalance')" @click.stop="openBalance(row as unknown as AdminUser, 'add')">
                <svg width="13" height="13" viewBox="0 0 13 13" fill="none"><circle cx="6.5" cy="6.5" r="5.5" stroke="currentColor" stroke-width="1.2"/><path d="M6.5 4v5M4 6.5h5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
              </button>
              <button class="uq-ib" :title="t('admin.usersQuench.actionEdit')" @click.stop="openEditDrawer(row as unknown as AdminUser)">
                <svg width="13" height="13" viewBox="0 0 13 13" fill="none"><path d="M9.5 2L11 3.5L5 9.5H3.5V8L9.5 2Z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </button>
              <button v-if="(row as unknown as AdminUser).role !== 'admin'" class="uq-ib" :class="(row as unknown as AdminUser).status === 'active' ? 'ib-warn' : 'ib-ok'" :title="(row as unknown as AdminUser).status === 'active' ? t('admin.usersQuench.actionDisable') : t('admin.usersQuench.actionEnable')" @click.stop="toggleStatus(row as unknown as AdminUser)">
                <svg v-if="(row as unknown as AdminUser).status === 'active'" width="13" height="13" viewBox="0 0 13 13" fill="none"><circle cx="6.5" cy="6.5" r="5.5" stroke="currentColor" stroke-width="1.2"/><path d="M4.5 4.5L8.5 8.5M8.5 4.5L4.5 8.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>
                <svg v-else width="13" height="13" viewBox="0 0 13 13" fill="none"><circle cx="6.5" cy="6.5" r="5.5" stroke="currentColor" stroke-width="1.2"/><path d="M4.5 6.5L5.8 7.8L8.5 5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
              </button>
            </div>
          </template>
        </DataTableV2>
      </div>

      <!-- 批量操作条 -->
      <BulkBar :count="selected.length" @clear="selected = []">
        <button @click="bulkEnable">{{ t('admin.usersQuench.bulkEnable') }}</button>
        <button @click="bulkDisable">{{ t('admin.usersQuench.bulkDisable') }}</button>
        <button class="q-btn-danger" @click="showBulkDel = true">{{ t('admin.usersQuench.bulkDelete') }}</button>
      </BulkBar>

      <!-- 批量删除确认 -->
      <Teleport to="body">
        <div v-if="showBulkDel" class="uq-overlay" @click.self="showBulkDel = false">
          <div class="uq-dialog">
            <div class="uq-dlg-title">{{ t('admin.usersQuench.bulkDeleteTitle') }}</div>
            <p class="uq-dlg-body">{{ t('admin.usersQuench.bulkDeleteConfirm', { n: selected.length }) }}</p>
            <div class="uq-dlg-foot">
              <button class="uq-btn" @click="showBulkDel = false">{{ t('admin.usersQuench.bulkDeleteCancel') }}</button>
              <button class="uq-btn uq-btn-danger" :disabled="bulkDeleting" @click="doBulkDelete">{{ bulkDeleting ? t('admin.usersQuench.bulkDeletingProgress', { current: bulkDelProg, total: selected.length }) : t('admin.usersQuench.bulkDeleteConfirmBtn') }}</button>
            </div>
          </div>
        </div>
      </Teleport>

      <!-- 用户 360 抽屉 (U2 契约) -->
      <UserDetailDrawer :userId="drawerUserId" :open="drawerOpen" @close="drawerOpen = false" @updated="loadUsers" />

      <!-- 创建/编辑抽屉 -->
      <UserFormDrawer :open="formOpen" :user="editingUser" @close="formOpen = false" @success="onFormSuccess" />

      <!-- 调余额 Modal -->
      <UserBalanceModal :show="showBalance" :user="balanceUser" :operation="balanceOp" @close="showBalance = false" @success="loadUsers" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted, defineAsyncComponent } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { DataTableV2, SavedViewTabs, BulkBar, useTableUrlState } from '@/components/datatable'
import type { ColumnDef, SavedView } from '@/components/datatable'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import UsersFilterBar from './UsersFilterBar.vue'
import UserFormDrawer from './UserFormDrawer.vue'

const UserDetailDrawer = defineAsyncComponent(() => import('@/components/admin/users/UserDetailDrawer.vue'))
const UserBalanceModal = defineAsyncComponent(() => import('@/components/admin/user/UserBalanceModal.vue'))

const { t } = useI18n()
const appStore = useAppStore()
const { state, reset } = useTableUrlState('u')

// ─── 快速视图定义 ───────────────────────────────────────────────
const QUICK_VIEWS = computed(() => [
  { id: 'all', label: t('admin.usersQuench.viewAll'), filters: {} as Record<string,string> },
  { id: 'admin', label: t('admin.usersQuench.viewAdmin'), filters: { role: 'admin' } },
  { id: 'disabled', label: t('admin.usersQuench.viewDisabled'), filters: { status: 'disabled' } },
])

// ─── 列定义 ─────────────────────────────────────────────────────
const COLUMNS = computed(() => [
  { key: 'email',       title: t('admin.usersQuench.colUser'),       width: '220px' },
  { key: 'role',        title: t('admin.usersQuench.colRole'),        width: '80px' },
  { key: 'balance',     title: t('admin.usersQuench.colBalance'),     align: 'right', width: '140px', sortable: true },
  { key: 'concurrency', title: t('admin.usersQuench.colConcurrency'), align: 'center', width: '80px' },
  { key: 'status',      title: t('admin.usersQuench.colStatus'),      width: '90px',  sortable: true },
  { key: 'created_at',  title: t('admin.usersQuench.colCreatedAt'),   width: '110px', sortable: true },
  { key: '_actions',    title: '',                                     width: '96px' },
] as unknown as ColumnDef<Record<string, unknown>>[])

// ─── 本地状态 ────────────────────────────────────────────────────
const users = ref<AdminUser[]>([])
const loading = ref(false)
const pagination = reactive({ total: 0, pages: 0 })
const density = ref<'comfortable' | 'compact'>('comfortable')
const selected = ref<AdminUser[]>([])
const activeQuickView = ref('all')

// 筛选状态（与 state.filters 双绑）
const searchInput = ref(state.q)
const filterRole = ref(state.filters.role as string ?? '')
const filterStatus = ref(state.filters.status as string ?? '')
let searchTimer: ReturnType<typeof setTimeout> | null = null

// 抽屉
const drawerOpen = ref(false)
const drawerUserId = ref<number | null>(null)
const formOpen = ref(false)
const editingUser = ref<AdminUser | null>(null)
// 余额
const showBalance = ref(false)
const balanceUser = ref<AdminUser | null>(null)
const balanceOp = ref<'add' | 'subtract'>('add')
// 批量删除
const showBulkDel = ref(false)
const bulkDeleting = ref(false)
const bulkDelProg = ref(0)

// AbortController
let abortCtrl: AbortController | null = null

// ─── 计算属性 ────────────────────────────────────────────────────
const savedViewState = computed(() => ({ page: state.page, pageSize: state.pageSize, sort: state.sort, order: state.order, q: state.q, filters: { ...state.filters } }))

// ─── 工具函数 ────────────────────────────────────────────────────
const PALETTE = ['#B9C7E8','#E8B9C2','#9BC4E8','#A3E0C8','#D6DCE6','#E8D5B9','#C4B9E8','#B9E8D5']
function avatarColor(email: string) {
  let h = 0; for (let i = 0; i < email.length; i++) h = (h * 31 + email.charCodeAt(i)) & 0xFFFFFFFF
  return PALETTE[Math.abs(h) % PALETTE.length]
}
function fmtDate(iso: string) { return iso ? formatDateTime(iso) : '-' }

// ─── 数据加载 ────────────────────────────────────────────────────
async function loadUsers() {
  abortCtrl?.abort(); abortCtrl = new AbortController()
  const { signal } = abortCtrl
  loading.value = true
  try {
    const res = await adminAPI.users.list(state.page, state.pageSize, {
      role: (state.filters.role as any) || undefined,
      status: (state.filters.status as any) || undefined,
      search: state.q || undefined,
      sort_by: state.sort || 'created_at',
      sort_order: (state.order as any) || 'desc',
      include_subscriptions: true,
    }, { signal })
    if (signal.aborted) return
    users.value = res.items; pagination.total = res.total; pagination.pages = res.pages
  } catch (e: any) {
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(e?.response?.data?.detail || t('admin.usersQuench.loadFailed'))
  } finally {
    if (!abortCtrl?.signal.aborted) loading.value = false
  }
}

// ─── 筛选 ────────────────────────────────────────────────────────
function commitSearch() { state.q = searchInput.value; state.page = 1 }
function clearFilters() { searchInput.value = ''; filterRole.value = ''; filterStatus.value = ''; state.q = ''; state.page = 1; activeQuickView.value = 'all' }
function applyQuickView(qv: typeof QUICK_VIEWS.value[0]) {
  activeQuickView.value = qv.id; filterRole.value = qv.filters.role ?? ''; filterStatus.value = qv.filters.status ?? ''; searchInput.value = ''; state.q = ''; state.page = 1
}
function onApplyView(view: SavedView | null) {
  if (!view) { reset(); searchInput.value = ''; filterRole.value = ''; filterStatus.value = ''; activeQuickView.value = 'all'; return }
  if (view.state.q != null) { state.q = view.state.q; searchInput.value = view.state.q }
  if (view.state.sort) state.sort = view.state.sort
  if (view.state.order) state.order = view.state.order
  if (view.state.page) state.page = view.state.page
  if (view.state.pageSize) state.pageSize = view.state.pageSize
  if (view.state.filters) { Object.assign(state.filters, view.state.filters); filterRole.value = (state.filters.role as string) ?? ''; filterStatus.value = (state.filters.status as string) ?? '' }
  activeQuickView.value = ''
}

// ─── 行操作 ──────────────────────────────────────────────────────
function onRowClick(row: Record<string, unknown>) { drawerUserId.value = (row as unknown as AdminUser).id; drawerOpen.value = true }
function onSelectionChange(rows: Record<string, unknown>[]) { selected.value = rows as unknown as AdminUser[] }
function openCreateDrawer() { editingUser.value = null; formOpen.value = true }
function openEditDrawer(user: AdminUser) { editingUser.value = user; formOpen.value = true }
function onFormSuccess() { formOpen.value = false; loadUsers() }
function openBalance(user: AdminUser, op: 'add' | 'subtract') { balanceUser.value = user; balanceOp.value = op; showBalance.value = true }

async function toggleStatus(user: AdminUser) {
  const ns = user.status === 'active' ? 'disabled' : 'active'
  try { await adminAPI.users.toggleStatus(user.id, ns); appStore.showSuccess(ns === 'active' ? t('admin.usersQuench.enabled') : t('admin.usersQuench.disabled')); loadUsers() }
  catch (e: any) { appStore.showError(e?.response?.data?.detail || t('admin.usersQuench.operationFailed')) }
}

// ─── 批量操作 ────────────────────────────────────────────────────
async function bulkEnable() {
  const targets = selected.value.filter(u => u.status !== 'active')
  if (!targets.length) { appStore.showError(t('admin.usersQuench.noEnableTarget')); return }
  await Promise.allSettled(targets.map(u => adminAPI.users.toggleStatus(u.id, 'active')))
  appStore.showSuccess(t('admin.usersQuench.bulkEnabledSuccess', { n: targets.length })); selected.value = []; loadUsers()
}
async function bulkDisable() {
  const targets = selected.value.filter(u => u.role !== 'admin' && u.status !== 'disabled')
  if (!targets.length) { appStore.showError(t('admin.usersQuench.noDisableTarget')); return }
  await Promise.allSettled(targets.map(u => adminAPI.users.toggleStatus(u.id, 'disabled')))
  appStore.showSuccess(t('admin.usersQuench.bulkDisabledSuccess', { n: targets.length })); selected.value = []; loadUsers()
}
async function doBulkDelete() {
  const targets = selected.value.filter(u => u.role !== 'admin')
  bulkDeleting.value = true; bulkDelProg.value = 0; let done = 0
  for (const u of targets) { try { await adminAPI.users.delete(u.id) } catch { /* continue */ } bulkDelProg.value = ++done }
  bulkDeleting.value = false; showBulkDel.value = false
  appStore.showSuccess(t('admin.usersQuench.bulkDeletedSuccess', { n: done })); selected.value = []; loadUsers()
}

// ─── 同步 filterRole/filterStatus → state.filters ────────────────
watch(filterRole, v => { state.filters.role = v; state.page = 1 })
watch(filterStatus, v => { state.filters.status = v; state.page = 1 })
watch(searchInput, () => { if (searchTimer) clearTimeout(searchTimer); searchTimer = setTimeout(commitSearch, 350) })

// ─── state 变化 → loadUsers ──────────────────────────────────────
watch(() => [state.page, state.pageSize, state.sort, state.order, state.q, JSON.stringify(state.filters)], loadUsers, { flush: 'post' })

onMounted(loadUsers)
onUnmounted(() => { abortCtrl?.abort(); if (searchTimer) clearTimeout(searchTimer) })
</script>

<style src="./users-quench.css"></style>
<style scoped>
/* :deep 规则必须在 scoped 块里才能穿透子组件 */
:deep(.q-tr:hover) .uq-acts,
:deep(.q-tr:focus-visible) .uq-acts { opacity: 1; }
</style>
