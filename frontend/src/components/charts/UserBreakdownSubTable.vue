<template>
  <div class="bg-muted/40">
    <div v-if="loading" class="flex items-center justify-center py-3">
      <LoadingSpinner />
    </div>
    <div v-else-if="items.length === 0" class="py-2 text-center text-xs text-muted-foreground">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
    <table v-else class="w-full text-xs">
      <tbody>
        <tr
          v-for="user in items"
          :key="user.user_id"
          class="border-t border-border/50"
        >
          <td class="max-w-[120px] truncate py-1 pl-6 text-foreground/85" :title="user.email">
            {{ user.email || `User #${user.user_id}` }}
          </td>
          <td class="py-1 text-right text-muted-foreground">
            {{ user.requests.toLocaleString() }}
          </td>
          <td class="py-1 text-right text-muted-foreground">
            {{ formatTokens(user.total_tokens) }}
          </td>
          <td class="py-1 text-right font-medium text-foreground">
            ${{ formatCost(user.actual_cost) }}
          </td>
          <td class="py-1 text-right text-amber-400">
            ${{ formatCost(user.account_cost) }}
          </td>
          <td class="py-1 pr-1 text-right text-muted-foreground">
            ${{ formatCost(user.cost) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { UserBreakdownItem } from '@/types'

const { t } = useI18n()

defineProps<{
  items: UserBreakdownItem[]
  loading?: boolean
}>()

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`
  return value.toLocaleString()
}

const formatCost = (value: number): string => {
  if (value >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (value >= 1) return value.toFixed(2)
  if (value >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}
</script>
