<template>
  <AppLayout>
    <div class="space-y-4">
      <!-- 页面头部：标题 + 操作 -->
      <div class="flex items-center justify-between gap-4">
        <div>
          <h1 class="text-xl font-semibold" :style="{ color: 'var(--ink-0)' }">PayGo 计价台</h1>
          <p class="mt-0.5 text-sm" :style="{ color: 'var(--ink-2)' }">
            渠道售价 × 分组倍率 → 用户实际单价矩阵
          </p>
        </div>
        <div class="flex items-center gap-2">
          <!-- 刷新 -->
          <button
            class="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm transition-colors"
            :style="{
              background: 'var(--bg-2)',
              border: '1px solid var(--line-1)',
              color: 'var(--ink-1)',
              boxShadow: 'var(--edge-hi)'
            }"
            :disabled="loading"
            @click="fetchAll"
          >
            <RefreshCwIcon class="h-4 w-4" :class="loading ? 'animate-spin' : ''" />
            刷新
          </button>

          <!-- 价格模拟器 -->
          <button
            class="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
            :style="{
              background: 'var(--metal-raised)',
              border: '1px solid var(--line-1)',
              color: 'var(--ink-0)',
              boxShadow: 'var(--edge-hi)'
            }"
            @click="simulatorVisible = true"
          >
            <CalculatorIcon class="h-4 w-4" :style="{ color: 'var(--azure)' }" />
            价格模拟器
          </button>
        </div>
      </div>

      <!-- 错误提示 -->
      <div
        v-if="error"
        class="rounded-lg px-4 py-3 text-sm"
        :style="{ background: 'var(--bad-dim)', border: '1px solid var(--bad)', color: 'var(--bad)' }"
      >
        加载失败：{{ error }}
      </div>

      <!-- 矩阵表格 -->
      <MatrixTable
        :loading="loading"
        :platforms="platforms"
        :active-groups="activeGroups"
        :matrix="matrix"
        :official-pricing-cache="officialPricingCache as Record<string, OfficialPricing | 'loading'>"
        @hover-model="ensureOfficialPricing"
        @update-multiplier="handleUpdateMultiplier"
      />

      <!-- 价格模拟器抽屉 -->
      <PriceSimulator
        :visible="simulatorVisible"
        :platforms="platforms"
        :matrix="matrix"
        :active-groups="activeGroups"
        :official-pricing-cache="officialPricingCache as Record<string, OfficialPricing | 'loading'>"
        @close="simulatorVisible = false"
        @need-official-pricing="ensureOfficialPricing"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RefreshCwIcon, CalculatorIcon } from 'lucide-vue-next'
import AppLayout from '@/components/layout/AppLayout.vue'
import MatrixTable from './MatrixTable.vue'
import PriceSimulator from './PriceSimulator.vue'
import { usePricingMatrix } from './usePricingMatrix'
import type { OfficialPricing } from './usePricingMatrix'

const {
  loading,
  error,
  matrix,
  platforms,
  activeGroups,
  officialPricingCache,
  fetchAll,
  ensureOfficialPricing,
  updateGroupMultiplier
} = usePricingMatrix()

const simulatorVisible = ref(false)

onMounted(() => fetchAll())

async function handleUpdateMultiplier(groupId: number, value: number) {
  try {
    await updateGroupMultiplier(groupId, value)
  } catch (e) {
    // 错误已在 composable 内回滚，这里可选 toast
    console.error('更新倍率失败', e)
  }
}
</script>
