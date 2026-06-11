<template>
  <AppLayout>
    <div class="oq-root">
      <!-- 页头 -->
      <div class="oq-head">
        <div>
          <h1 class="oq-title">收入看板</h1>
          <p class="oq-desc">收入域 · 实时统计 · 选择时间范围</p>
        </div>
        <div class="oq-head-acts">
          <!-- 日期段选择 -->
          <div class="oq-days-seg">
            <button v-for="d in DAYS_OPTIONS" :key="d" :class="{ on: days === d }" @click="days = d">
              {{ d }}{{ t('payment.admin.daySuffix') }}
            </button>
          </div>
          <button class="oq-btn" :disabled="loading" @click="loadDashboard">
            <svg width="13" height="13" viewBox="0 0 13 13" fill="none" :class="loading ? 'oq-spin-icon' : ''"><path d="M11 6.5A4.5 4.5 0 1 1 6.5 2a4.5 4.5 0 0 1 3.18 1.32" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/><path d="M11 2v2.5H8.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="oq-loading">
        <div class="oq-spinner"></div>
      </div>

      <template v-else-if="stats">
        <!-- 统计卡片 -->
        <OrderStatsCards :stats="stats" />

        <!-- 日收入折线图 -->
        <DailyRevenueChart :data="stats.daily_series || []" :loading="loading" style="margin-bottom:16px" />

        <!-- 支付方式 + Top 用户 -->
        <div class="oq-grid-2">
          <PaymentMethodChart :methods="stats.payment_methods || []" />
          <TopUsersLeaderboard :users="stats.top_users || []" />
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { DashboardStats } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import OrderStatsCards from '@/components/admin/payment/OrderStatsCards.vue'
import DailyRevenueChart from '@/components/admin/payment/DailyRevenueChart.vue'
import PaymentMethodChart from '@/components/admin/payment/PaymentMethodChart.vue'
import TopUsersLeaderboard from '@/components/admin/payment/TopUsersLeaderboard.vue'

const { t } = useI18n()
const appStore = useAppStore()

const DAYS_OPTIONS = [7, 30, 90] as const
const days = ref<number>(30)
const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

async function loadDashboard() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getDashboard(days.value)
    stats.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

watch(days, () => loadDashboard())
onMounted(() => loadDashboard())
</script>

<style src="./orders-quench.css"></style>
<style scoped>
.oq-spin-icon { animation: oq-icon-spin .7s linear infinite; }
@keyframes oq-icon-spin { to { transform: rotate(360deg); } }
</style>
