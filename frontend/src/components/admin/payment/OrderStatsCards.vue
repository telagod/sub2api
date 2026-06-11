<template>
  <div class="oq-stats" style="margin-bottom:16px">
    <!-- 今日收入 -->
    <div class="oq-stat-card">
      <div class="oq-stat-icon">
        <!-- dollar sign -->
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M8 2v12M5.5 5.5C5.5 4.12 6.62 3 8 3h1a2 2 0 0 1 0 4H7a2 2 0 0 0 0 4h1.5A2 2 0 0 0 10.5 9" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </div>
      <div class="oq-stat-label">{{ t('payment.admin.todayRevenue') }}</div>
      <div class="oq-stat-value">${{ formatMoney(stats.today_amount) }}</div>
      <div class="oq-stat-sub">{{ stats.today_count }} {{ t('payment.admin.orders') }}</div>
    </div>

    <!-- 总收入 -->
    <div class="oq-stat-card">
      <div class="oq-stat-icon">
        <!-- credit card -->
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><rect x="1.5" y="3.5" width="13" height="9" rx="1.5" stroke="currentColor" stroke-width="1.3"/><path d="M1.5 6.5h13" stroke="currentColor" stroke-width="1.3"/><path d="M4 9.5h3" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></svg>
      </div>
      <div class="oq-stat-label">{{ t('payment.admin.totalRevenue') }}</div>
      <div class="oq-stat-value">${{ formatMoney(stats.total_amount) }}</div>
      <div class="oq-stat-sub">{{ stats.total_count }} {{ t('payment.admin.orders') }}</div>
    </div>

    <!-- 今日订单数 -->
    <div class="oq-stat-card">
      <div class="oq-stat-icon">
        <!-- bar chart -->
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M2 12h2V7H2v5ZM7 12h2V4H7v8ZM12 12h2V9h-2v3Z" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </div>
      <div class="oq-stat-label">{{ t('payment.admin.todayOrders') }}</div>
      <div class="oq-stat-value" style="font-size:22px">{{ stats.today_count }}</div>
    </div>

    <!-- 平均客单 -->
    <div class="oq-stat-card">
      <div class="oq-stat-icon">
        <!-- trend up -->
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M2 11L6 7L9 10L14 5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M11 5h3v3" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </div>
      <div class="oq-stat-label">{{ t('payment.admin.avgAmount') }}</div>
      <div class="oq-stat-value">${{ formatMoney(stats.avg_amount) }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { DashboardStats } from '@/types/payment'

const { t } = useI18n()

defineProps<{
  stats: DashboardStats
}>()

function formatMoney(value: number): string {
  return value.toFixed(2)
}
</script>
