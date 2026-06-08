<template>
  <Card>
    <CardContent class="p-4">
    <h3 class="mb-4 text-sm font-semibold text-foreground">
      {{ t('payment.admin.topUsers') }}
    </h3>
    <div
      v-if="!users?.length"
      class="flex h-32 items-center justify-center text-sm text-muted-foreground"
    >
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="(user, idx) in users"
        :key="user.user_id"
        class="flex items-center justify-between rounded-lg px-3 py-2 hover:bg-accent"
      >
        <div class="flex items-center gap-3">
          <span
            :class="[
              'flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold',
              rankClass(idx),
            ]"
          >
            {{ idx + 1 }}
          </span>
          <span class="text-sm text-foreground/85">{{ user.email }}</span>
        </div>
        <span class="text-sm font-medium text-foreground">
          ${{ user.amount.toFixed(2) }}
        </span>
      </div>
    </div>
  </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  users: { user_id: number; email: string; amount: number }[]
}>()

function rankClass(idx: number): string {
  if (idx === 0) return 'bg-amber-500/10 text-amber-400'
  if (idx === 1) return 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
  if (idx === 2) return 'bg-amber-500/10 text-amber-400'
  return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
}
</script>
