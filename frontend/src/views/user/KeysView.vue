<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-col gap-3">
          <div class="flex flex-wrap items-center gap-3">
            <SearchInput
              v-model="filterSearch"
              :placeholder="t('keys.searchPlaceholder')"
              class="w-full sm:w-64"
              @search="onFilterChange"
            />
            <Select
              :model-value="filterGroupId"
              class="w-40"
              :options="groupFilterOptions"
              @update:model-value="onGroupFilterChange"
            />
            <Select
              :model-value="filterStatus"
              class="w-40"
              :options="statusFilterOptions"
              @update:model-value="onStatusFilterChange"
            />
          </div>
          <EndpointPopover
            v-if="publicSettings?.api_base_url || (publicSettings?.custom_endpoints?.length ?? 0) > 0"
            :api-base-url="publicSettings?.api_base_url || ''"
            :custom-endpoints="publicSettings?.custom_endpoints || []"
          />
        </div>
      </template>

      <template #actions>
        <div class="flex justify-end gap-3">
          <Button
            variant="secondary"
            size="icon"
            @click="loadApiKeys"
            :disabled="loading"
            :title="t('common.refresh')"
          >
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </Button>
          <Button @click="showCreateModal = true" data-tour="keys-create-btn">
            <Icon name="plus" size="md" class="mr-2" />
            {{ t('keys.createKey') }}
          </Button>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="apiKeys"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          @sort="handleSort"
        >
          <template #cell-key="{ value, row }">
            <div class="flex items-center gap-2">
              <code class="code text-xs">
                {{ maskApiKey(value) }}
              </code>
              <Button
                variant="ghost"
                size="sm"
                @click="copyToClipboard(value, row.id)"
                class="h-auto p-1"
                :class="
                  copiedKeyId === row.id
                    ? 'text-emerald-400'
                    : 'text-muted-foreground hover:text-foreground'
                "
                :title="copiedKeyId === row.id ? t('keys.copied') : t('keys.copyToClipboard')"
              >
                <Icon
                  v-if="copiedKeyId === row.id"
                  name="check"
                  size="sm"
                  :stroke-width="2"
                />
                <Icon v-else name="clipboard" size="sm" />
              </Button>
            </div>
          </template>

          <template #cell-name="{ value, row }">
            <div class="flex items-center gap-1.5">
              <span class="font-medium text-foreground">{{ value }}</span>
              <Icon
                v-if="row.ip_whitelist?.length > 0 || row.ip_blacklist?.length > 0"
                name="shield"
                size="sm"
                class="text-primary-200"
                :title="t('keys.ipRestrictionEnabled')"
              />
            </div>
          </template>

          <template #cell-group="{ row }">
            <div class="group/dropdown relative">
              <button
                :ref="(el) => setGroupButtonRef(row.id, el)"
                @click="openGroupSelector(row)"
                class="-mx-2 -my-1 flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1 transition-all duration-200 hover:bg-muted dark:hover:bg-dark-700"
                :title="t('keys.clickToChangeGroup')"
              >
                <GroupBadge
                  v-if="row.group"
                  :name="row.group.name"
                  :platform="row.group.platform"
                  :subscription-type="row.group.subscription_type"
                  :rate-multiplier="row.group.rate_multiplier"
                  :user-rate-multiplier="userGroupRates[row.group.id]"
                />
                <span v-else class="text-sm text-muted-foreground dark:text-dark-500">{{
                  t('keys.noGroup')
                }}</span>
                <span class="text-xs text-muted-foreground">{{ t('keys.selectGroup') }}</span>
                <svg
                  class="h-3.5 w-3.5 text-muted-foreground opacity-60 transition-opacity group-hover/dropdown:opacity-100"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  stroke-width="2"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M8.25 15L12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9"
                  />
                </svg>
              </button>
            </div>
          </template>

          <template #cell-usage="{ row }">
            <div class="text-sm">
              <div class="flex items-center gap-1.5">
                <span class="text-muted-foreground">{{ t('keys.today') }}:</span>
                <span class="font-medium text-foreground dark:text-white">
                  ${{ (usageStats[row.id]?.today_actual_cost ?? 0).toFixed(4) }}
                </span>
              </div>
              <div class="mt-0.5 flex items-center gap-1.5">
                <span class="text-muted-foreground">{{ t('keys.total') }}:</span>
                <span class="font-medium text-foreground dark:text-white">
                  ${{ (usageStats[row.id]?.total_actual_cost ?? 0).toFixed(4) }}
                </span>
              </div>
              <div v-if="row.quota > 0" class="mt-1.5">
                <div class="flex items-center gap-1.5">
                  <span class="text-muted-foreground">{{ t('keys.quota') }}:</span>
                  <span :class="['font-medium', row.quota_used >= row.quota ? 'text-red-400' : row.quota_used >= row.quota * 0.8 ? 'text-yellow-500' : 'text-foreground dark:text-white']">
                    ${{ row.quota_used?.toFixed(2) || '0.00' }} / ${{ row.quota?.toFixed(2) }}
                  </span>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', row.quota_used >= row.quota ? 'bg-red-500' : row.quota_used >= row.quota * 0.8 ? 'bg-yellow-500' : 'bg-primary-500']" :style="{ width: Math.min((row.quota_used / row.quota) * 100, 100) + '%' }" />
                </div>
              </div>
            </div>
          </template>

          <template #cell-rate_limit="{ row }">
            <div v-if="row.rate_limit_5h > 0 || row.rate_limit_1d > 0 || row.rate_limit_7d > 0" class="space-y-1.5 min-w-[140px]">
              <div v-if="row.rate_limit_5h > 0">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-muted-foreground">5h</span>
                  <span :class="['font-medium tabular-nums', row.usage_5h >= row.rate_limit_5h ? 'text-red-400' : row.usage_5h >= row.rate_limit_5h * 0.8 ? 'text-yellow-500' : 'text-foreground/85']">
                    ${{ row.usage_5h?.toFixed(2) || '0.00' }}/${{ row.rate_limit_5h?.toFixed(2) }}
                  </span>
                </div>
                <div class="h-1 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', row.usage_5h >= row.rate_limit_5h ? 'bg-red-500' : row.usage_5h >= row.rate_limit_5h * 0.8 ? 'bg-yellow-500' : 'bg-emerald-500']" :style="{ width: Math.min((row.usage_5h / row.rate_limit_5h) * 100, 100) + '%' }" />
                </div>
                <div v-if="row.reset_5h_at && formatResetTime(row.reset_5h_at)" class="text-[10px] text-muted-foreground tabular-nums">&#x27F3; {{ formatResetTime(row.reset_5h_at) }}</div>
              </div>
              <div v-if="row.rate_limit_1d > 0">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-muted-foreground">1d</span>
                  <span :class="['font-medium tabular-nums', row.usage_1d >= row.rate_limit_1d ? 'text-red-400' : row.usage_1d >= row.rate_limit_1d * 0.8 ? 'text-yellow-500' : 'text-foreground/85']">
                    ${{ row.usage_1d?.toFixed(2) || '0.00' }}/${{ row.rate_limit_1d?.toFixed(2) }}
                  </span>
                </div>
                <div class="h-1 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', row.usage_1d >= row.rate_limit_1d ? 'bg-red-500' : row.usage_1d >= row.rate_limit_1d * 0.8 ? 'bg-yellow-500' : 'bg-emerald-500']" :style="{ width: Math.min((row.usage_1d / row.rate_limit_1d) * 100, 100) + '%' }" />
                </div>
                <div v-if="row.reset_1d_at && formatResetTime(row.reset_1d_at)" class="text-[10px] text-muted-foreground tabular-nums">&#x27F3; {{ formatResetTime(row.reset_1d_at) }}</div>
              </div>
              <div v-if="row.rate_limit_7d > 0">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-muted-foreground">7d</span>
                  <span :class="['font-medium tabular-nums', row.usage_7d >= row.rate_limit_7d ? 'text-red-400' : row.usage_7d >= row.rate_limit_7d * 0.8 ? 'text-yellow-500' : 'text-foreground/85']">
                    ${{ row.usage_7d?.toFixed(2) || '0.00' }}/${{ row.rate_limit_7d?.toFixed(2) }}
                  </span>
                </div>
                <div class="h-1 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', row.usage_7d >= row.rate_limit_7d ? 'bg-red-500' : row.usage_7d >= row.rate_limit_7d * 0.8 ? 'bg-yellow-500' : 'bg-emerald-500']" :style="{ width: Math.min((row.usage_7d / row.rate_limit_7d) * 100, 100) + '%' }" />
                </div>
                <div v-if="row.reset_7d_at && formatResetTime(row.reset_7d_at)" class="text-[10px] text-muted-foreground tabular-nums">&#x27F3; {{ formatResetTime(row.reset_7d_at) }}</div>
              </div>
              <Button v-if="row.usage_5h > 0 || row.usage_1d > 0 || row.usage_7d > 0" variant="ghost" size="sm" @click.stop="confirmResetRateLimitFromTable(row)" class="mt-0.5 h-auto gap-1 px-1.5 py-0.5 text-xs text-muted-foreground hover:text-primary-600 dark:hover:text-primary-400" :title="t('keys.resetRateLimitUsage')">
                <Icon name="refresh" size="xs" />
                {{ t('keys.resetUsage') }}
              </Button>
            </div>
            <span v-else class="text-sm text-muted-foreground dark:text-dark-500">-</span>
          </template>

          <template #cell-expires_at="{ value }">
            <span v-if="value" :class="['text-sm', new Date(value) < new Date() ? 'text-red-400' : 'text-muted-foreground dark:text-dark-400']">{{ formatDateTime(value) }}</span>
            <span v-else class="text-sm text-muted-foreground dark:text-dark-500">{{ t('keys.noExpiration') }}</span>
          </template>

          <template #cell-status="{ value }">
            <Badge :variant="value === 'active' ? 'default' : value === 'quota_exhausted' ? 'secondary' : value === 'expired' ? 'destructive' : 'outline'" :class="value === 'active' ? 'bg-emerald-600 text-white border-emerald-600' : value === 'quota_exhausted' ? 'bg-yellow-600 text-white border-yellow-600' : ''">
              {{ t('keys.status.' + value) }}
            </Badge>
          </template>

          <template #cell-last_used_at="{ value }">
            <span v-if="value" class="text-sm text-muted-foreground dark:text-dark-400">{{ formatDateTime(value) }}</span>
            <span v-else class="text-sm text-muted-foreground dark:text-dark-500">-</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-muted-foreground dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <Button variant="ghost" size="sm" @click="openUseKeyModal(row)" class="flex flex-col items-center gap-0.5 h-auto px-1.5 py-1.5 text-muted-foreground hover:bg-green-50 hover:text-emerald-400 dark:hover:bg-green-900/20 dark:hover:text-green-400">
                <Icon name="terminal" size="sm" />
                <span class="text-xs">{{ t('keys.useKey') }}</span>
              </Button>
              <Button v-if="!publicSettings?.hide_ccs_import_button" variant="ghost" size="sm" @click="importToCcswitch(row)" class="flex flex-col items-center gap-0.5 h-auto px-1.5 py-1.5 text-muted-foreground hover:bg-blue-50 hover:text-sky-400 dark:hover:bg-blue-900/20 dark:hover:text-blue-400">
                <Icon name="upload" size="sm" />
                <span class="text-xs">{{ t('keys.importToCcSwitch') }}</span>
              </Button>
              <Button variant="ghost" size="sm" @click="toggleKeyStatus(row)" :class="'flex flex-col items-center gap-0.5 h-auto px-1.5 py-1.5 ' + (row.status === 'active' ? 'text-muted-foreground hover:bg-yellow-50 hover:text-amber-400 dark:hover:bg-yellow-900/20 dark:hover:text-yellow-400' : 'text-muted-foreground hover:bg-green-50 hover:text-emerald-400 dark:hover:bg-green-900/20 dark:hover:text-green-400')">
                <Icon v-if="row.status === 'active'" name="ban" size="sm" />
                <Icon v-else name="checkCircle" size="sm" />
                <span class="text-xs">{{ row.status === 'active' ? t('keys.disable') : t('keys.enable') }}</span>
              </Button>
              <Button variant="ghost" size="sm" @click="editKey(row)" class="flex flex-col items-center gap-0.5 h-auto px-1.5 py-1.5 text-muted-foreground hover:bg-muted hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400">
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit') }}</span>
              </Button>
              <Button variant="ghost" size="sm" @click="confirmDelete(row)" class="flex flex-col items-center gap-0.5 h-auto px-1.5 py-1.5 text-muted-foreground hover:bg-red-50 hover:text-red-400 dark:hover:bg-red-900/20 dark:hover:text-red-400">
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('common.delete') }}</span>
              </Button>
            </div>
          </template>

          <template #empty>
            <EmptyState :title="t('keys.noKeysYet')" :description="t('keys.createFirstKey')" :action-text="t('keys.createKey')" @action="showCreateModal = true" />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination v-if="pagination.total > 0" :page="pagination.page" :total="pagination.total" :page-size="pagination.page_size" @update:page="handlePageChange" @update:pageSize="handlePageSizeChange" />
      </template>
    </TablePageLayout>

    <!-- Create/Edit Modal -->
    <BaseDialog :show="showCreateModal || showEditModal" :title="showEditModal ? t('keys.editKey') : t('keys.createKey')" width="normal" @close="closeModals">
      <form id="key-form" @submit.prevent="handleSubmit" class="space-y-5">
        <div class="space-y-2">
          <Label>{{ t('keys.nameLabel') }}</Label>
          <Input v-model="formData.name" type="text" required :placeholder="t('keys.namePlaceholder')" data-tour="key-form-name" />
        </div>

        <div class="space-y-2">
          <Label>{{ t('keys.groupLabel') }}</Label>
          <Select v-model="formData.group_id" :options="groupOptions" :placeholder="t('keys.selectGroup')" :searchable="true" :search-placeholder="t('keys.searchGroup')" data-tour="key-form-group">
            <template #selected="{ option }">
              <GroupBadge v-if="option" :name="(option as unknown as GroupOption).label" :platform="(option as unknown as GroupOption).platform" :subscription-type="(option as unknown as GroupOption).subscriptionType" :rate-multiplier="(option as unknown as GroupOption).rate" :user-rate-multiplier="(option as unknown as GroupOption).userRate" />
              <span v-else class="text-muted-foreground">{{ t('keys.selectGroup') }}</span>
            </template>
            <template #option="{ option, selected }">
              <GroupOptionItem :name="(option as unknown as GroupOption).label" :platform="(option as unknown as GroupOption).platform" :subscription-type="(option as unknown as GroupOption).subscriptionType" :rate-multiplier="(option as unknown as GroupOption).rate" :user-rate-multiplier="(option as unknown as GroupOption).userRate" :description="(option as unknown as GroupOption).description" :selected="selected" />
            </template>
          </Select>
        </div>

        <div v-if="!showEditModal" class="space-y-3">
          <div class="flex items-center justify-between">
            <Label class="mb-0">{{ t('keys.customKeyLabel') }}</Label>
            <Switch :checked="formData.use_custom_key" @update:checked="formData.use_custom_key = $event" />
          </div>
          <div v-if="formData.use_custom_key">
            <Input v-model="formData.custom_key" type="text" class="font-mono" :placeholder="t('keys.customKeyPlaceholder')" :class="{ 'border-red-500 dark:border-red-500': customKeyError }" />
            <p v-if="customKeyError" class="mt-1 text-sm text-red-400">{{ customKeyError }}</p>
            <p v-else class="input-hint">{{ t('keys.customKeyHint') }}</p>
          </div>
        </div>

        <div v-if="showEditModal" class="space-y-2">
          <Label>{{ t('keys.statusLabel') }}</Label>
          <Select v-model="formData.status" :options="statusOptions" :placeholder="t('keys.selectStatus')" />
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label class="mb-0">{{ t('keys.ipRestriction') }}</Label>
            <Switch :checked="formData.enable_ip_restriction" @update:checked="formData.enable_ip_restriction = $event" />
          </div>
          <div v-if="formData.enable_ip_restriction" class="space-y-4 pt-2">
            <div class="space-y-2">
              <Label>{{ t('keys.ipWhitelist') }}</Label>
              <Textarea v-model="formData.ip_whitelist" rows="3" class="font-mono text-sm" :placeholder="t('keys.ipWhitelistPlaceholder')" />
              <p class="input-hint">{{ t('keys.ipWhitelistHint') }}</p>
            </div>
            <div class="space-y-2">
              <Label>{{ t('keys.ipBlacklist') }}</Label>
              <Textarea v-model="formData.ip_blacklist" rows="3" class="font-mono text-sm" :placeholder="t('keys.ipBlacklistPlaceholder')" />
              <p class="input-hint">{{ t('keys.ipBlacklistHint') }}</p>
            </div>
          </div>
        </div>

        <div class="space-y-3">
          <Label>{{ t('keys.quotaLimit') }}</Label>
          <div class="space-y-4">
            <div>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                <Input :model-value="formData.quota ?? ''" @update:model-value="formData.quota = $event === '' ? null : Number($event)" type="number" step="0.01" min="0" class="pl-7" :placeholder="t('keys.quotaAmountPlaceholder')" />
              </div>
              <p class="input-hint">{{ t('keys.quotaAmountHint') }}</p>
            </div>
            <div v-if="showEditModal && selectedKey && selectedKey.quota > 0" class="space-y-2">
              <Label>{{ t('keys.quotaUsed') }}</Label>
              <div class="flex items-center gap-2">
                <div class="flex-1 rounded-lg bg-muted px-3 py-2 dark:bg-dark-700">
                  <span class="font-medium text-foreground dark:text-white">${{ selectedKey.quota_used?.toFixed(4) || '0.0000' }}</span>
                  <span class="mx-2 text-muted-foreground">/</span>
                  <span class="text-muted-foreground">${{ selectedKey.quota?.toFixed(2) || '0.00' }}</span>
                </div>
                <Button type="button" variant="secondary" size="sm" @click="confirmResetQuota" :title="t('keys.resetQuotaUsed')">{{ t('keys.reset') }}</Button>
              </div>
            </div>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label class="mb-0">{{ t('keys.rateLimitSection') }}</Label>
            <Switch :checked="formData.enable_rate_limit" @update:checked="formData.enable_rate_limit = $event" />
          </div>
          <div v-if="formData.enable_rate_limit" class="space-y-4 pt-2">
            <p class="input-hint -mt-2">{{ t('keys.rateLimitHint') }}</p>
            <div>
              <Label>{{ t('keys.rateLimit5h') }}</Label>
              <div class="relative mt-2">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                <Input :model-value="formData.rate_limit_5h ?? ''" @update:model-value="formData.rate_limit_5h = $event === '' ? null : Number($event)" type="number" step="0.01" min="0" class="pl-7" placeholder="0" />
              </div>
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_5h > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-muted px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="['font-medium', selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'text-red-400' : selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'text-yellow-500' : 'text-foreground dark:text-white']">${{ selectedKey.usage_5h?.toFixed(4) || '0.0000' }}</span>
                    <span class="mx-2 text-muted-foreground">/</span>
                    <span class="text-muted-foreground">${{ selectedKey.rate_limit_5h?.toFixed(2) || '0.00' }}</span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'bg-red-500' : selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'bg-yellow-500' : 'bg-green-500']" :style="{ width: Math.min((selectedKey.usage_5h / selectedKey.rate_limit_5h) * 100, 100) + '%' }" />
                </div>
              </div>
            </div>
            <div>
              <Label>{{ t('keys.rateLimit1d') }}</Label>
              <div class="relative mt-2">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                <Input :model-value="formData.rate_limit_1d ?? ''" @update:model-value="formData.rate_limit_1d = $event === '' ? null : Number($event)" type="number" step="0.01" min="0" class="pl-7" placeholder="0" />
              </div>
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_1d > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-muted px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="['font-medium', selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'text-red-400' : selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'text-yellow-500' : 'text-foreground dark:text-white']">${{ selectedKey.usage_1d?.toFixed(4) || '0.0000' }}</span>
                    <span class="mx-2 text-muted-foreground">/</span>
                    <span class="text-muted-foreground">${{ selectedKey.rate_limit_1d?.toFixed(2) || '0.00' }}</span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'bg-red-500' : selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'bg-yellow-500' : 'bg-green-500']" :style="{ width: Math.min((selectedKey.usage_1d / selectedKey.rate_limit_1d) * 100, 100) + '%' }" />
                </div>
              </div>
            </div>
            <div>
              <Label>{{ t('keys.rateLimit7d') }}</Label>
              <div class="relative mt-2">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">$</span>
                <Input :model-value="formData.rate_limit_7d ?? ''" @update:model-value="formData.rate_limit_7d = $event === '' ? null : Number($event)" type="number" step="0.01" min="0" class="pl-7" placeholder="0" />
              </div>
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_7d > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-muted px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="['font-medium', selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'text-red-400' : selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'text-yellow-500' : 'text-foreground dark:text-white']">${{ selectedKey.usage_7d?.toFixed(4) || '0.0000' }}</span>
                    <span class="mx-2 text-muted-foreground">/</span>
                    <span class="text-muted-foreground">${{ selectedKey.rate_limit_7d?.toFixed(2) || '0.00' }}</span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-accent dark:bg-dark-600">
                  <div :class="['h-full rounded-full transition-all', selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'bg-red-500' : selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'bg-yellow-500' : 'bg-green-500']" :style="{ width: Math.min((selectedKey.usage_7d / selectedKey.rate_limit_7d) * 100, 100) + '%' }" />
                </div>
              </div>
            </div>
            <div v-if="showEditModal && selectedKey && (selectedKey.rate_limit_5h > 0 || selectedKey.rate_limit_1d > 0 || selectedKey.rate_limit_7d > 0)">
              <Button type="button" variant="secondary" size="sm" @click="confirmResetRateLimit">{{ t('keys.resetRateLimitUsage') }}</Button>
            </div>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <Label class="mb-0">{{ t('keys.expiration') }}</Label>
            <Switch :checked="formData.enable_expiration" @update:checked="formData.enable_expiration = $event" />
          </div>
          <div v-if="formData.enable_expiration" class="space-y-4 pt-2">
            <div class="flex flex-wrap gap-2">
              <Button v-for="days in ['7', '30', '90']" :key="days" type="button" :variant="formData.expiration_preset === days ? 'default' : 'outline'" size="sm" @click="setExpirationDays(parseInt(days))">
                {{ showEditModal ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days }) }}
              </Button>
              <Button type="button" :variant="formData.expiration_preset === 'custom' ? 'default' : 'outline'" size="sm" @click="formData.expiration_preset = 'custom'">{{ t('keys.customDate') }}</Button>
            </div>
            <div class="space-y-2">
              <Label>{{ t('keys.expirationDate') }}</Label>
              <Input v-model="formData.expiration_date" type="datetime-local" />
              <p class="input-hint">{{ t('keys.expirationDateHint') }}</p>
            </div>
            <div v-if="showEditModal && selectedKey?.expires_at" class="text-sm">
              <span class="text-muted-foreground">{{ t('keys.currentExpiration') }}: </span>
              <span class="font-medium text-foreground dark:text-white">{{ formatDateTime(selectedKey.expires_at) }}</span>
            </div>
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button variant="secondary" @click="closeModals" type="button">{{ t('common.cancel') }}</Button>
          <Button form="key-form" type="submit" :disabled="submitting" data-tour="key-form-submit">
            <svg v-if="submitting" class="-ml-1 mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            {{ submitting ? t('keys.saving') : showEditModal ? t('common.update') : t('common.create') }}
          </Button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog :show="showDeleteDialog" :title="t('keys.deleteKey')" :message="t('keys.deleteConfirmMessage', { name: selectedKey?.name })" :confirm-text="t('common.delete')" :cancel-text="t('common.cancel')" :danger="true" @confirm="handleDelete" @cancel="showDeleteDialog = false" />
    <ConfirmDialog :show="showResetQuotaDialog" :title="t('keys.resetQuotaTitle')" :message="t('keys.resetQuotaConfirmMessage', { name: selectedKey?.name, used: selectedKey?.quota_used?.toFixed(4) })" :confirm-text="t('keys.reset')" :cancel-text="t('common.cancel')" :danger="true" @confirm="resetQuotaUsed" @cancel="showResetQuotaDialog = false" />
    <ConfirmDialog :show="showResetRateLimitDialog" :title="t('keys.resetRateLimitTitle')" :message="t('keys.resetRateLimitConfirmMessage', { name: selectedKey?.name })" :confirm-text="t('keys.reset')" :cancel-text="t('common.cancel')" :danger="true" @confirm="resetRateLimitUsage" @cancel="showResetRateLimitDialog = false" />

    <UseKeyModal :show="showUseKeyModal" :api-key="selectedKey?.key || ''" :base-url="publicSettings?.api_base_url || ''" :platform="selectedKey?.group?.platform || null" :allow-messages-dispatch="selectedKey?.group?.allow_messages_dispatch || false" @close="closeUseKeyModal" />

    <!-- CCS Client Selection Dialog -->
    <Dialog :open="showCcsClientSelect" @update:open="(v: boolean) => { if (!v) closeCcsClientSelect() }">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{{ t('keys.ccsClientSelect.title') }}</DialogTitle>
          <DialogDescription>{{ t('keys.ccsClientSelect.description') }}</DialogDescription>
        </DialogHeader>
        <div class="grid grid-cols-2 gap-3">
          <Button variant="outline" @click="handleCcsClientSelect('claude')" class="flex flex-col items-center gap-2 h-auto p-4 rounded-xl border-2 hover:border-primary-500 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20">
            <Icon name="terminal" size="xl" class="text-foreground/75" />
            <span class="font-medium text-foreground dark:text-white">{{ t('keys.ccsClientSelect.claudeCode') }}</span>
            <span class="text-xs text-muted-foreground">{{ t('keys.ccsClientSelect.claudeCodeDesc') }}</span>
          </Button>
          <Button variant="outline" @click="handleCcsClientSelect('gemini')" class="flex flex-col items-center gap-2 h-auto p-4 rounded-xl border-2 hover:border-primary-500 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20">
            <Icon name="sparkles" size="xl" class="text-foreground/75" />
            <span class="font-medium text-foreground dark:text-white">{{ t('keys.ccsClientSelect.geminiCli') }}</span>
            <span class="text-xs text-muted-foreground">{{ t('keys.ccsClientSelect.geminiCliDesc') }}</span>
          </Button>
        </div>
        <DialogFooter>
          <Button variant="secondary" @click="closeCcsClientSelect">{{ t('common.cancel') }}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Group Selector Dropdown -->
    <Teleport to="body">
      <div v-if="groupSelectorKeyId !== null && dropdownPosition" ref="dropdownRef" class="animate-in fade-in slide-in-from-top-2 fixed z-[100000020] w-max min-w-[380px] overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 duration-200 dark:bg-dark-800 dark:ring-white/10" style="pointer-events: auto !important;" :style="{ top: dropdownPosition.top !== undefined ? dropdownPosition.top + 'px' : undefined, bottom: dropdownPosition.bottom !== undefined ? dropdownPosition.bottom + 'px' : undefined, left: dropdownPosition.left + 'px' }">
        <div class="border-b border-gray-100 p-2 dark:border-dark-700">
          <div class="relative">
            <svg class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <Input v-model="groupSearchQuery" type="text" class="pl-8" :placeholder="t('keys.searchGroup')" @click.stop />
          </div>
        </div>
        <div class="max-h-80 overflow-y-auto p-1.5">
          <button v-for="option in filteredGroupOptions" :key="option.value ?? 'null'" @click="changeGroup(selectedKeyForGroup!, option.value)" :class="['flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-sm transition-colors', 'border-b border-gray-100 last:border-0 dark:border-dark-700', selectedKeyForGroup?.group_id === option.value || (!selectedKeyForGroup?.group_id && option.value === null) ? 'bg-primary-50 dark:bg-primary-900/20' : 'hover:bg-muted dark:hover:bg-dark-700']" :title="option.description || undefined">
            <GroupOptionItem :name="option.label" :platform="option.platform" :subscription-type="option.subscriptionType" :rate-multiplier="option.rate" :user-rate-multiplier="option.userRate" :description="option.description" :selected="selectedKeyForGroup?.group_id === option.value || (!selectedKeyForGroup?.group_id && option.value === null)" />
          </button>
          <div v-if="filteredGroupOptions.length === 0" class="py-4 text-center text-sm text-muted-foreground">{{ t('keys.noGroupFound') }}</div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, type ComponentPublicInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useOnboardingStore } from '@/stores/onboarding'
import { useClipboard } from '@/composables/useClipboard'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { keysAPI, authAPI, usageAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Icon from '@/components/icons/Icon.vue'
import UseKeyModal from '@/components/keys/UseKeyModal.vue'
import EndpointPopover from '@/components/keys/EndpointPopover.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import type { ApiKey, Group, PublicSettings, SubscriptionType, GroupPlatform } from '@/types'
import type { Column } from '@/components/common/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'
import { formatDateTime } from '@/utils/format'
import { maskApiKey } from '@/utils/maskApiKey'
import { buildCcSwitchImportDeeplink, type CcSwitchClientType } from '@/utils/ccswitchImport'

const formatDateTimeLocal = (isoDate: string): string => {
  const date = new Date(isoDate)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

interface GroupOption {
  value: number
  label: string
  description: string | null
  rate: number
  userRate: number | null
  subscriptionType: SubscriptionType
  platform: GroupPlatform
}

const appStore = useAppStore()
const onboardingStore = useOnboardingStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('common.name'), sortable: true },
  { key: 'key', label: t('keys.apiKey'), sortable: false },
  { key: 'group', label: t('keys.group'), sortable: false },
  { key: 'usage', label: t('keys.usage'), sortable: false },
  { key: 'rate_limit', label: t('keys.rateLimitColumn'), sortable: false },
  { key: 'expires_at', label: t('keys.expiresAt'), sortable: true },
  { key: 'status', label: t('common.status'), sortable: true },
  { key: 'last_used_at', label: t('keys.lastUsedAt'), sortable: true },
  { key: 'created_at', label: t('keys.created'), sortable: true },
  { key: 'actions', label: t('common.actions'), sortable: false }
])

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const loading = ref(false)
const submitting = ref(false)
const now = ref(new Date())
let resetTimer: ReturnType<typeof setInterval> | null = null
const usageStats = ref<Record<string, BatchApiKeyUsageStats>>({})
const userGroupRates = ref<Record<number, number>>({})
const pagination = ref({ page: 1, page_size: getPersistedPageSize(), total: 0, pages: 0 })
const sortState = ref({ sort_by: 'created_at', sort_order: 'desc' as 'asc' | 'desc' })
const filterSearch = ref('')
const filterStatus = ref('')
const filterGroupId = ref<string | number>('')
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const showResetQuotaDialog = ref(false)
const showResetRateLimitDialog = ref(false)
const showUseKeyModal = ref(false)
const showCcsClientSelect = ref(false)
const pendingCcsRow = ref<ApiKey | null>(null)
const selectedKey = ref<ApiKey | null>(null)
const copiedKeyId = ref<number | null>(null)
const groupSelectorKeyId = ref<number | null>(null)
const publicSettings = ref<PublicSettings | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref<{ top?: number; bottom?: number; left: number } | null>(null)
const groupButtonRefs = ref<Map<number, HTMLElement>>(new Map())
let abortController: AbortController | null = null

const selectedKeyForGroup = computed(() => {
  if (groupSelectorKeyId.value === null) return null
  return apiKeys.value.find((k) => k.id === groupSelectorKeyId.value) || null
})

const setGroupButtonRef = (keyId: number, el: Element | ComponentPublicInstance | null) => {
  if (el instanceof HTMLElement) { groupButtonRefs.value.set(keyId, el) } else { groupButtonRefs.value.delete(keyId) }
}

const formData = ref({
  name: '', group_id: null as number | null, status: 'active' as 'active' | 'inactive',
  use_custom_key: false, custom_key: '', enable_ip_restriction: false, ip_whitelist: '', ip_blacklist: '',
  enable_quota: false, quota: null as number | null,
  enable_rate_limit: false, rate_limit_5h: null as number | null, rate_limit_1d: null as number | null, rate_limit_7d: null as number | null,
  enable_expiration: false, expiration_preset: '30' as '7' | '30' | '90' | 'custom', expiration_date: ''
})

const customKeyError = computed(() => {
  if (!formData.value.use_custom_key || !formData.value.custom_key) return ''
  const key = formData.value.custom_key
  if (key.length < 16) return t('keys.customKeyTooShort')
  if (!/^[a-zA-Z0-9_-]+$/.test(key)) return t('keys.customKeyInvalidChars')
  return ''
})

const statusOptions = computed(() => [
  { value: 'active', label: t('common.active') },
  { value: 'inactive', label: t('common.inactive') }
])

const groupFilterOptions = computed(() => [
  { value: '', label: t('keys.allGroups') },
  { value: 0, label: t('keys.noGroup') },
  ...groups.value.map((g) => ({ value: g.id, label: g.name }))
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('keys.allStatus') },
  { value: 'active', label: t('keys.status.active') },
  { value: 'inactive', label: t('keys.status.inactive') },
  { value: 'quota_exhausted', label: t('keys.status.quota_exhausted') },
  { value: 'expired', label: t('keys.status.expired') }
])

const onFilterChange = () => { pagination.value.page = 1; loadApiKeys() }
const onGroupFilterChange = (value: string | number | boolean | null) => { filterGroupId.value = value as string | number; onFilterChange() }
const onStatusFilterChange = (value: string | number | boolean | null) => { filterStatus.value = value as string; onFilterChange() }

const groupOptions = computed(() =>
  groups.value.map((group) => ({
    value: group.id, label: group.name, description: group.description, rate: group.rate_multiplier,
    userRate: userGroupRates.value[group.id] ?? null, subscriptionType: group.subscription_type, platform: group.platform
  }))
)

const groupSearchQuery = ref('')
const filteredGroupOptions = computed(() => {
  const query = groupSearchQuery.value.trim().toLowerCase()
  if (!query) return groupOptions.value
  return groupOptions.value.filter((opt) => opt.label.toLowerCase().includes(query) || (opt.description && opt.description.toLowerCase().includes(query)))
})

const copyToClipboard = async (text: string, keyId: number) => {
  const success = await clipboardCopy(text, t('keys.copied'))
  if (success) { copiedKeyId.value = keyId; setTimeout(() => { copiedKeyId.value = null }, 800) }
}

const isAbortError = (error: unknown) => {
  if (!error || typeof error !== 'object') return false
  const { name, code } = error as { name?: string; code?: string }
  return name === 'AbortError' || code === 'ERR_CANCELED'
}

const loadApiKeys = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  const { signal } = controller
  loading.value = true
  try {
    const filters: { search?: string; status?: string; group_id?: number | string; sort_by?: string; sort_order?: 'asc' | 'desc' } = {}
    if (filterSearch.value) filters.search = filterSearch.value
    if (filterStatus.value) filters.status = filterStatus.value
    if (filterGroupId.value !== '') filters.group_id = filterGroupId.value
    filters.sort_by = sortState.value.sort_by
    filters.sort_order = sortState.value.sort_order
    const response = await keysAPI.list(pagination.value.page, pagination.value.page_size, filters, { signal })
    if (signal.aborted) return
    apiKeys.value = response.items
    pagination.value.total = response.total
    pagination.value.pages = response.pages
    if (response.items.length > 0) {
      const keyIds = response.items.map((k) => k.id)
      try {
        const usageResponse = await usageAPI.getDashboardApiKeysUsage(keyIds, { signal })
        if (signal.aborted) return
        usageStats.value = usageResponse.stats
      } catch (e) { if (!isAbortError(e)) console.error('Failed to load usage stats:', e) }
    }
  } catch (error) {
    if (isAbortError(error)) return
    appStore.showError(t('keys.failedToLoad'))
  } finally {
    if (abortController === controller) loading.value = false
  }
}

const loadGroups = async () => { try { groups.value = await userGroupsAPI.getAvailable() } catch (error) { console.error('Failed to load groups:', error) } }
const loadUserGroupRates = async () => { try { userGroupRates.value = await userGroupsAPI.getUserGroupRates() } catch (error) { console.error('Failed to load user group rates:', error) } }
const loadPublicSettings = async () => { try { publicSettings.value = await authAPI.getPublicSettings() } catch (error) { console.error('Failed to load public settings:', error) } }

const openUseKeyModal = (key: ApiKey) => { selectedKey.value = key; showUseKeyModal.value = true }
const closeUseKeyModal = () => { showUseKeyModal.value = false; selectedKey.value = null }
const handlePageChange = (page: number) => { pagination.value.page = page; loadApiKeys() }
const handlePageSizeChange = (pageSize: number) => { pagination.value.page_size = pageSize; pagination.value.page = 1; loadApiKeys() }
const handleSort = (key: string, order: 'asc' | 'desc') => { sortState.value.sort_by = key; sortState.value.sort_order = order; pagination.value.page = 1; loadApiKeys() }

const editKey = (key: ApiKey) => {
  selectedKey.value = key
  const hasIPRestriction = (key.ip_whitelist?.length > 0) || (key.ip_blacklist?.length > 0)
  const hasExpiration = !!key.expires_at
  formData.value = {
    name: key.name, group_id: key.group_id,
    status: key.status === 'quota_exhausted' || key.status === 'expired' ? 'inactive' : key.status,
    use_custom_key: false, custom_key: '', enable_ip_restriction: hasIPRestriction,
    ip_whitelist: (key.ip_whitelist || []).join('\n'), ip_blacklist: (key.ip_blacklist || []).join('\n'),
    enable_quota: key.quota > 0, quota: key.quota > 0 ? key.quota : null,
    enable_rate_limit: (key.rate_limit_5h > 0) || (key.rate_limit_1d > 0) || (key.rate_limit_7d > 0),
    rate_limit_5h: key.rate_limit_5h || null, rate_limit_1d: key.rate_limit_1d || null, rate_limit_7d: key.rate_limit_7d || null,
    enable_expiration: hasExpiration, expiration_preset: 'custom',
    expiration_date: key.expires_at ? formatDateTimeLocal(key.expires_at) : ''
  }
  showEditModal.value = true
}

const toggleKeyStatus = async (key: ApiKey) => {
  const newStatus = key.status === 'active' ? 'inactive' : 'active'
  try { await keysAPI.toggleStatus(key.id, newStatus); appStore.showSuccess(newStatus === 'active' ? t('keys.keyEnabledSuccess') : t('keys.keyDisabledSuccess')); loadApiKeys() }
  catch (error) { appStore.showError(t('keys.failedToUpdateStatus')) }
}

const openGroupSelector = (key: ApiKey) => {
  if (groupSelectorKeyId.value === key.id) { groupSelectorKeyId.value = null; dropdownPosition.value = null }
  else {
    const buttonEl = groupButtonRefs.value.get(key.id)
    if (buttonEl) {
      const rect = buttonEl.getBoundingClientRect()
      const dropdownEstHeight = 400
      const spaceBelow = window.innerHeight - rect.bottom
      const spaceAbove = rect.top
      if (spaceBelow < dropdownEstHeight && spaceAbove > spaceBelow) { dropdownPosition.value = { bottom: window.innerHeight - rect.top + 4, left: rect.left } }
      else { dropdownPosition.value = { top: rect.bottom + 4, left: rect.left } }
    }
    groupSelectorKeyId.value = key.id
    groupSearchQuery.value = ''
  }
}

const changeGroup = async (key: ApiKey, newGroupId: number | null) => {
  groupSelectorKeyId.value = null; dropdownPosition.value = null
  if (key.group_id === newGroupId) return
  try { await keysAPI.update(key.id, { group_id: newGroupId }); appStore.showSuccess(t('keys.groupChangedSuccess')); loadApiKeys() }
  catch (error) { appStore.showError(t('keys.failedToChangeGroup')) }
}

const closeGroupSelector = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.group\\/dropdown') && !dropdownRef.value?.contains(target)) { groupSelectorKeyId.value = null; dropdownPosition.value = null }
}

const confirmDelete = (key: ApiKey) => { selectedKey.value = key; showDeleteDialog.value = true }

const handleSubmit = async () => {
  if (formData.value.group_id === null) { appStore.showError(t('keys.groupRequired')); return }
  if (!showEditModal.value && formData.value.use_custom_key) {
    if (!formData.value.custom_key) { appStore.showError(t('keys.customKeyRequired')); return }
    if (customKeyError.value) { appStore.showError(customKeyError.value); return }
  }
  const parseIPList = (text: string): string[] => text.split('\n').map(ip => ip.trim()).filter(ip => ip.length > 0)
  const ipWhitelist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_whitelist) : []
  const ipBlacklist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_blacklist) : []
  const quota = formData.value.quota && formData.value.quota > 0 ? formData.value.quota : 0
  let expiresInDays: number | undefined
  let expiresAt: string | null | undefined
  if (formData.value.enable_expiration && formData.value.expiration_date) {
    if (!showEditModal.value) {
      const expDate = new Date(formData.value.expiration_date); const now = new Date()
      const diffDays = Math.ceil((expDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      expiresInDays = diffDays > 0 ? diffDays : 1
    } else { expiresAt = new Date(formData.value.expiration_date).toISOString() }
  } else if (showEditModal.value) { expiresAt = '' }
  const rateLimitData = formData.value.enable_rate_limit ? {
    rate_limit_5h: formData.value.rate_limit_5h && formData.value.rate_limit_5h > 0 ? formData.value.rate_limit_5h : 0,
    rate_limit_1d: formData.value.rate_limit_1d && formData.value.rate_limit_1d > 0 ? formData.value.rate_limit_1d : 0,
    rate_limit_7d: formData.value.rate_limit_7d && formData.value.rate_limit_7d > 0 ? formData.value.rate_limit_7d : 0,
  } : { rate_limit_5h: 0, rate_limit_1d: 0, rate_limit_7d: 0 }
  submitting.value = true
  try {
    if (showEditModal.value && selectedKey.value) {
      await keysAPI.update(selectedKey.value.id, {
        name: formData.value.name, group_id: formData.value.group_id, status: formData.value.status,
        ip_whitelist: ipWhitelist, ip_blacklist: ipBlacklist, quota: quota, expires_at: expiresAt,
        rate_limit_5h: rateLimitData.rate_limit_5h, rate_limit_1d: rateLimitData.rate_limit_1d, rate_limit_7d: rateLimitData.rate_limit_7d,
      })
      appStore.showSuccess(t('keys.keyUpdatedSuccess'))
    } else {
      const customKey = formData.value.use_custom_key ? formData.value.custom_key : undefined
      await keysAPI.create(formData.value.name, formData.value.group_id, customKey, ipWhitelist, ipBlacklist, quota, expiresInDays, rateLimitData)
      appStore.showSuccess(t('keys.keyCreatedSuccess'))
      if (onboardingStore.isCurrentStep('[data-tour="key-form-submit"]')) { onboardingStore.nextStep(500) }
    }
    closeModals(); loadApiKeys()
  } catch (error: any) { appStore.showError(error.response?.data?.detail || t('keys.failedToSave')) }
  finally { submitting.value = false }
}

const handleDelete = async () => {
  if (!selectedKey.value) return
  try { await keysAPI.delete(selectedKey.value.id); appStore.showSuccess(t('keys.keyDeletedSuccess')); showDeleteDialog.value = false; loadApiKeys() }
  catch (error: any) { appStore.showError(error?.message || t('keys.failedToDelete')) }
}

const closeModals = () => {
  showCreateModal.value = false; showEditModal.value = false; selectedKey.value = null
  formData.value = {
    name: '', group_id: null, status: 'active', use_custom_key: false, custom_key: '',
    enable_ip_restriction: false, ip_whitelist: '', ip_blacklist: '', enable_quota: false, quota: null,
    enable_rate_limit: false, rate_limit_5h: null, rate_limit_1d: null, rate_limit_7d: null,
    enable_expiration: false, expiration_preset: '30', expiration_date: ''
  }
}

const confirmResetQuota = () => { showResetQuotaDialog.value = true }

const setExpirationDays = (days: number) => {
  formData.value.expiration_preset = days.toString() as '7' | '30' | '90'
  const expDate = new Date(); expDate.setDate(expDate.getDate() + days)
  formData.value.expiration_date = formatDateTimeLocal(expDate.toISOString())
}

const resetQuotaUsed = async () => {
  if (!selectedKey.value) return
  showResetQuotaDialog.value = false
  try { await keysAPI.update(selectedKey.value.id, { reset_quota: true }); appStore.showSuccess(t('keys.quotaResetSuccess')); if (selectedKey.value) selectedKey.value.quota_used = 0 }
  catch (error: any) { appStore.showError(error.response?.data?.detail || t('keys.failedToResetQuota')) }
}

const confirmResetRateLimit = () => { showResetRateLimitDialog.value = true }
const confirmResetRateLimitFromTable = (row: ApiKey) => { selectedKey.value = row; showResetRateLimitDialog.value = true }

const resetRateLimitUsage = async () => {
  if (!selectedKey.value) return
  showResetRateLimitDialog.value = false
  try {
    await keysAPI.update(selectedKey.value.id, { reset_rate_limit_usage: true }); appStore.showSuccess(t('keys.rateLimitResetSuccess'))
    await loadApiKeys()
    const refreshedKey = apiKeys.value.find(k => k.id === selectedKey.value!.id)
    if (refreshedKey) selectedKey.value = refreshedKey
  } catch (error: any) { appStore.showError(error.response?.data?.detail || t('keys.failedToResetRateLimit')) }
}

const importToCcswitch = (row: ApiKey) => {
  const platform = row.group?.platform || 'anthropic'
  if (platform === 'antigravity') { pendingCcsRow.value = row; showCcsClientSelect.value = true; return }
  executeCcsImport(row, platform === 'gemini' ? 'gemini' : 'claude')
}

const executeCcsImport = (row: ApiKey, clientType: CcSwitchClientType) => {
  const baseUrl = publicSettings.value?.api_base_url || window.location.origin
  const platform = row.group?.platform || 'anthropic'
  const usageScript = `({ request: { url: "{{baseUrl}}/v1/usage", method: "GET", headers: { "Authorization": "Bearer {{apiKey}}" } }, extractor: function(response) { const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance; const unit = response?.unit ?? response?.quota?.unit ?? "USD"; return { isValid: response?.is_active ?? response?.isValid ?? true, remaining, unit }; } })`
  const providerName = (publicSettings.value?.site_name || 'sub2api').trim() || 'sub2api'
  const deeplink = buildCcSwitchImportDeeplink({ baseUrl, platform, clientType, providerName, apiKey: row.key, usageScript })
  try { window.open(deeplink, '_self'); setTimeout(() => { if (document.hasFocus()) appStore.showError(t('keys.ccSwitchNotInstalled')) }, 100) }
  catch (error) { appStore.showError(t('keys.ccSwitchNotInstalled')) }
}

const handleCcsClientSelect = (clientType: CcSwitchClientType) => {
  if (pendingCcsRow.value) executeCcsImport(pendingCcsRow.value, clientType)
  showCcsClientSelect.value = false; pendingCcsRow.value = null
}

const closeCcsClientSelect = () => { showCcsClientSelect.value = false; pendingCcsRow.value = null }

function formatResetTime(resetAt: string | null): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - now.value.getTime()
  if (diff <= 0) return t('keys.resetNow')
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

// Suppress unused type import warnings
void (undefined as unknown as Group)

onMounted(() => { loadApiKeys(); loadGroups(); loadUserGroupRates(); loadPublicSettings(); document.addEventListener('click', closeGroupSelector); resetTimer = setInterval(() => { now.value = new Date() }, 60000) })
onUnmounted(() => { document.removeEventListener('click', closeGroupSelector); if (resetTimer) clearInterval(resetTimer) })
</script>
