<template>
  <!-- 表格模式面板，对齐旧视图全量列能力 -->
  <div class="apt-wrap">
    <!-- 批量操作 Bar（进度条 + 快捷按钮） -->
    <div
      v-if="selectedIds.length > 0 || bulkDeleteProgress"
      class="apt-bulk"
    >
      <div class="apt-bulk-info">
        {{ t('admin.accountTablePanel.bulkSelected', { n: selectedIds.length }) }}
        <button class="apt-bulk-link" @click="$emit('select-page')">{{ t('admin.accountTablePanel.bulkSelectPage') }}</button>
        <button class="apt-bulk-link" @click="$emit('clear-selection')">{{ t('admin.accountTablePanel.bulkClear') }}</button>
      </div>
      <div class="apt-bulk-actions">
        <button
          class="apt-btn apt-btn-danger"
          :disabled="!!bulkDeleteProgress"
          @click="$emit('bulk-delete')"
        >{{ t('admin.accountTablePanel.bulkDelete') }}</button>
        <button
          class="apt-btn"
          :disabled="!!bulkDeleteProgress"
          @click="$emit('bulk-reset-status')"
        >{{ t('admin.accountTablePanel.bulkResetStatus') }}</button>
        <button
          class="apt-btn"
          :disabled="!!bulkDeleteProgress"
          @click="$emit('bulk-refresh-token')"
        >{{ t('admin.accountTablePanel.bulkRefreshToken') }}</button>
        <button
          class="apt-btn"
          :disabled="!!bulkDeleteProgress"
          @click="$emit('bulk-toggle-schedulable', true)"
        >{{ t('admin.accountTablePanel.bulkEnableSchedule') }}</button>
        <button
          class="apt-btn"
          :disabled="!!bulkDeleteProgress"
          @click="$emit('bulk-toggle-schedulable', false)"
        >{{ t('admin.accountTablePanel.bulkDisableSchedule') }}</button>
        <button
          class="apt-btn apt-btn-primary"
          :disabled="!!bulkDeleteProgress"
          @click="$emit('bulk-edit-selected')"
        >{{ t('admin.accountTablePanel.bulkEdit') }}</button>
      </div>
      <!-- 进度条 -->
      <div v-if="bulkDeleteProgress" class="apt-progress">
        <div class="apt-progress-bg">
          <div
            class="apt-progress-fill"
            :style="{ width: progressPercent + '%' }"
          ></div>
        </div>
        <span class="apt-progress-label">{{ bulkDeleteProgress.current }}/{{ bulkDeleteProgress.total }}</span>
      </div>
    </div>

    <!-- DataTableV2 -->
    <DataTableV2
      :columns="(columns as any)"
      :rows="(accounts as any[])"
      :total="total"
      :loading="loading"
      :selectable="true"
      row-key="id"
      :page="page"
      :page-size="pageSize"
      :sort="sortBy"
      :order="sortOrder"
      @update:page="$emit('update:page', $event)"
      @update:sort="$emit('update:sort', $event)"
      @update:order="$emit('update:order', $event)"
      @update:selected="onSelected"
    >
      <!-- 账号名 + 平台 chip -->
      <template #cell-name="{ row }">
        <div class="apt-name-cell">
          <div class="apt-name-row">
            <span class="apt-name">{{ row.name }}</span>
            <PlatformTypeBadge
              :platform="(row as any).platform"
              :type="(row as any).type"
              :plan-type="(row as any).credentials?.plan_type"
              :privacy-mode="(row as any).extra?.privacy_mode"
              :compact="true"
            />
          </div>
          <span v-if="accountEmail(row)" class="apt-email">{{ accountEmail(row) }}</span>
        </div>
      </template>

      <!-- 容量 -->
      <template #cell-capacity="{ row }">
        <AccountCapacityCell :account="(row as any)" />
      </template>

      <!-- 状态 -->
      <template #cell-status="{ row }">
        <AccountStatusIndicator
          :account="(row as any)"
          @show-temp-unsched="$emit('show-temp-unsched', $event)"
        />
      </template>

      <!-- 调度开关 -->
      <template #cell-schedulable="{ row }">
        <button
          class="apt-toggle"
          :class="row.schedulable ? 'apt-toggle-on' : 'apt-toggle-off'"
          :title="row.schedulable ? t('admin.accountTablePanel.scheduleEnabled') : t('admin.accountTablePanel.scheduleDisabled')"
          :disabled="togglingSchedulable === row.id"
          @click="$emit('toggle-schedulable', row)"
        >
          <span class="apt-toggle-thumb" :class="row.schedulable ? 'apt-toggle-thumb-on' : ''"></span>
        </button>
      </template>

      <!-- 分组 -->
      <template #cell-groups="{ row }">
        <AccountGroupsCell :groups="(row as any).groups" :max-display="3" />
      </template>

      <!-- 用量窗口 -->
      <template #cell-usage="{ row }">
        <AccountUsageCell
          :account="(row as any)"
          :today-stats="todayStatsByAccountId[String(row.id)] ?? null"
          :today-stats-loading="todayStatsLoading"
          :manual-refresh-token="manualRefreshToken"
        />
      </template>

      <!-- 今日统计 -->
      <template #cell-today_stats="{ row }">
        <AccountTodayStatsCell
          :stats="todayStatsByAccountId[String(row.id)] ?? null"
          :loading="todayStatsLoading"
          :error="null"
        />
      </template>

      <!-- 代理 -->
      <template #cell-proxy="{ row }">
        <div v-if="row.proxy" class="apt-proxy">
          <span>{{ (row.proxy as any).name }}</span>
          <span v-if="(row.proxy as any).country_code" class="apt-muted">({{ (row.proxy as any).country_code }})</span>
        </div>
        <span v-else class="apt-muted">-</span>
      </template>

      <!-- 优先级 -->
      <template #cell-priority="{ value }">
        <span class="apt-mono">{{ value }}</span>
      </template>

      <!-- 倍率 -->
      <template #cell-rate_multiplier="{ value }">
        <span class="apt-mono">{{ ((value as number | null) ?? 1).toFixed(2) }}x</span>
      </template>

      <!-- 最后使用 -->
      <template #cell-last_used_at="{ value }">
        <span class="apt-muted apt-sm">{{ formatRelativeTime(value as string | null) }}</span>
      </template>

      <!-- 创建时间 -->
      <template #cell-created_at="{ value }">
        <span class="apt-muted apt-sm">{{ formatDateTime(value as string | Date | null | undefined) }}</span>
      </template>

      <!-- 操作列 -->
      <template #cell-actions="{ row }">
        <div class="apt-row-actions">
          <button class="apt-icon-btn" :title="t('admin.accountTablePanel.editBtn')" :aria-label="t('admin.accountTablePanel.editBtn')" @click="$emit('edit', row)">
            <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931z"/></svg>
          </button>
          <button class="apt-icon-btn apt-icon-btn-danger" :title="t('admin.accountTablePanel.deleteBtn')" :aria-label="t('admin.accountTablePanel.deleteBtn')" @click="$emit('delete', row)">
            <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"/></svg>
          </button>
          <button class="apt-icon-btn" :title="t('admin.accountTablePanel.moreBtn')" :aria-label="t('admin.accountTablePanel.moreBtn')" @click="$emit('more', row, $event)">
            <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM12.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0zM18.75 12a.75.75 0 11-1.5 0 .75.75 0 011.5 0z"/></svg>
          </button>
        </div>
      </template>
    </DataTableV2>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account, AdminGroup, WindowStats } from '@/types'
import { DataTableV2 } from '@/components/datatable'
import type { ColumnDef } from '@/components/datatable'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import AccountCapacityCell from '@/components/account/AccountCapacityCell.vue'
import AccountStatusIndicator from '@/components/account/AccountStatusIndicator.vue'
import AccountUsageCell from '@/components/account/AccountUsageCell.vue'
import AccountTodayStatsCell from '@/components/account/AccountTodayStatsCell.vue'
import AccountGroupsCell from '@/components/account/AccountGroupsCell.vue'
import { formatDateTime, formatRelativeTime } from '@/utils/format'

const props = defineProps<{
  accounts: Account[]
  groups: AdminGroup[]
  total: number
  loading?: boolean
  page: number
  pageSize: number
  sortBy: string
  sortOrder: 'asc' | 'desc'
  selectedIds: number[]
  bulkDeleteProgress: { current: number; total: number } | null
  todayStatsByAccountId: Record<string, WindowStats>
  todayStatsLoading?: boolean
  manualRefreshToken?: number
  togglingSchedulable: number | null
}>()

const emit = defineEmits<{
  'edit': [account: any]
  'delete': [account: any]
  'more': [account: any, event: MouseEvent]
  'show-temp-unsched': [account: Account]
  'toggle-schedulable': [account: any]
  'update:selectedIds': [ids: number[]]
  'update:page': [page: number]
  'update:sort': [sort: string]
  'update:order': [order: 'asc' | 'desc']
  'bulk-delete': []
  'bulk-reset-status': []
  'bulk-refresh-token': []
  'bulk-toggle-schedulable': [schedulable: boolean]
  'bulk-edit-selected': []
  'select-page': []
  'clear-selection': []
}>()

const { t } = useI18n()

const columns = computed<ColumnDef<Record<string, unknown>>[]>(() => [
  { key: 'name',            title: t('admin.accountTablePanel.colName'),        sortable: true  },
  { key: 'capacity',        title: t('admin.accountTablePanel.colCapacity')                     },
  { key: 'status',          title: t('admin.accountTablePanel.colStatus'),      sortable: true,  width: '90px' },
  { key: 'schedulable',     title: '⚙',                                         sortable: true,  width: '44px' },
  { key: 'groups',          title: t('admin.accountTablePanel.colGroups')                       },
  { key: 'usage',           title: t('admin.accountTablePanel.colUsage'),                       width: '280px' },
  { key: 'today_stats',     title: t('admin.accountTablePanel.colTodayStats')                   },
  { key: 'proxy',           title: t('admin.accountTablePanel.colProxy')                        },
  { key: 'priority',        title: t('admin.accountTablePanel.colPriority'),    sortable: true,  width: '70px', align: 'right' },
  { key: 'rate_multiplier', title: t('admin.accountTablePanel.colMultiplier'),  sortable: true,  width: '70px', align: 'right' },
  { key: 'last_used_at',    title: t('admin.accountTablePanel.colLastUsed'),    sortable: true,  width: '110px' },
  { key: 'created_at',      title: t('admin.accountTablePanel.colCreatedAt'),   sortable: true,  width: '110px' },
  { key: 'actions',         title: '',                                                           width: '88px' },
])

const progressPercent = computed(() => {
  if (!props.bulkDeleteProgress) return 0
  const { current, total } = props.bulkDeleteProgress
  if (total <= 0) return 0
  return Math.min(100, Math.round((current / total) * 100))
})

function accountEmail(row: any): string {
  return row.extra?.email_address || row.extra?.email || row.credentials?.email || ''
}

function onSelected(rows: Record<string, unknown>[]) {
  const ids = rows.map(r => r['id'] as number)
  emit('update:selectedIds', ids)
}
</script>

<style scoped>
.apt-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 批量操作 */
.apt-bulk {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: var(--bg-1);
  border: 1px solid var(--line-0);
  border-radius: 10px;
}

.apt-bulk-info {
  font-size: 13px;
  color: var(--ink-0);
  display: flex;
  align-items: center;
  gap: 8px;
}

.apt-bulk-link {
  font-size: 12px;
  color: var(--azure);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.apt-bulk-link:hover { text-decoration: underline; }

.apt-bulk-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-left: auto;
}

.apt-btn {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 6px;
  border: 1px solid var(--line-0);
  background: var(--bg-2);
  color: var(--ink-0);
  cursor: pointer;
  transition: background 0.12s;
}

.apt-btn:hover { background: var(--line-0); }
.apt-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.apt-btn-danger  { color: var(--bad); border-color: var(--bad-dim); }
.apt-btn-danger:hover  { background: var(--bad-dim); }
.apt-btn-primary { color: var(--azure); border-color: var(--azure-dim); }
.apt-btn-primary:hover { background: var(--azure-dim); }

/* 进度条 */
.apt-progress {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  margin-top: 2px;
}

.apt-progress-bg {
  flex: 1;
  height: 4px;
  border-radius: 2px;
  background: var(--bg-2);
  overflow: hidden;
}

.apt-progress-fill {
  height: 100%;
  border-radius: 2px;
  background: var(--bad);
  transition: width 0.3s;
}

.apt-progress-label {
  font-size: 11px;
  font-family: monospace;
  color: var(--ink-2);
  white-space: nowrap;
}

/* 单元格内样式 */
.apt-name-cell { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.apt-name-row  { display: flex; align-items: center; gap: 4px; min-width: 0; }
.apt-name      { font-size: 13px; font-weight: 500; color: var(--ink-0); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.apt-email     { font-size: 10px; color: var(--ink-2); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.apt-muted     { color: var(--ink-2); font-size: 12px; }
.apt-sm        { font-size: 11px; }
.apt-mono      { font-family: monospace; font-size: 12px; color: var(--ink-1); }
.apt-proxy     { display: flex; align-items: center; gap: 4px; font-size: 12px; color: var(--ink-1); }

/* 调度开关 */
.apt-toggle {
  position: relative;
  display: inline-flex;
  width: 36px;
  height: 20px;
  border-radius: 10px;
  border: 2px solid transparent;
  cursor: pointer;
  transition: background 0.2s;
  flex-shrink: 0;
}

.apt-toggle-on  { background: var(--ok); }
.apt-toggle-off { background: var(--bg-2); border-color: var(--line-1); }
.apt-toggle:disabled { opacity: 0.5; cursor: not-allowed; }

.apt-toggle-thumb {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: white;
  transition: transform 0.2s;
}

.apt-toggle-thumb-on { transform: translateX(16px); }

/* 行内 icon 按钮 */
.apt-row-actions { display: flex; align-items: center; gap: 2px; }

.apt-icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--ink-2);
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}

.apt-icon-btn:hover { background: var(--bg-2); color: var(--ink-0); }
.apt-icon-btn-danger:hover { background: var(--bad-dim); color: var(--bad); }
</style>
