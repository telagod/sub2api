<template>
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
    <!-- Today Revenue -->
    <Card>
      <CardContent class="p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md border border-border bg-metal-raised p-2 shadow-metal-edge">
          <Icon name="dollar" size="md" class="text-primary-200" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted-foreground">{{ t('payment.admin.todayRevenue') }}</p>
          <p class="text-xl font-bold text-foreground">${{ formatMoney(stats.today_amount) }}</p>
          <p class="text-xs text-muted-foreground">
            {{ stats.today_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </CardContent>
    </Card>

    <!-- Total Revenue -->
    <Card>
      <CardContent class="p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md border border-border bg-metal-raised p-2 shadow-metal-edge">
          <Icon name="creditCard" size="md" class="text-primary-200" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted-foreground">{{ t('payment.admin.totalRevenue') }}</p>
          <p class="text-xl font-bold text-foreground">${{ formatMoney(stats.total_amount) }}</p>
          <p class="text-xs text-muted-foreground">
            {{ stats.total_count }} {{ t('payment.admin.orders') }}
          </p>
        </div>
      </div>
    </CardContent>
    </Card>

    <!-- Today Orders -->
    <Card>
      <CardContent class="p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md border border-border bg-metal-raised p-2 shadow-metal-edge">
          <Icon name="chart" size="md" class="text-primary-200" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted-foreground">{{ t('payment.admin.todayOrders') }}</p>
          <p class="text-xl font-bold text-foreground">{{ stats.today_count }}</p>
        </div>
      </div>
    </CardContent>
    </Card>

    <!-- Average Amount -->
    <Card>
      <CardContent class="p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md border border-border bg-metal-raised p-2 shadow-metal-edge">
          <Icon name="chart" size="md" class="text-primary-200" :stroke-width="2" />
        </div>
        <div>
          <p class="text-xs font-medium text-muted-foreground">{{ t('payment.admin.avgAmount') }}</p>
          <p class="text-xl font-bold text-foreground">${{ formatMoney(stats.avg_amount) }}</p>
        </div>
      </div>
    </CardContent>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { Card, CardContent } from '@/components/ui/card'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { DashboardStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  stats: DashboardStats
}>()

function formatMoney(value: number): string {
  return value.toFixed(2)
}
</script>
