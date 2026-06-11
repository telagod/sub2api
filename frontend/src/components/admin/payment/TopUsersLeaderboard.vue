<template>
  <div class="oq-chart-card">
    <h3 class="oq-chart-title">{{ t('payment.admin.topUsers') }}</h3>
    <div v-if="!users?.length" class="oq-no-data" style="min-height:120px">
      {{ t('payment.admin.noData') }}
    </div>
    <div v-else>
      <div v-for="(user, idx) in users" :key="user.user_id" class="oq-rank-row">
        <div class="oq-rank-left">
          <span :class="['oq-rank-num', rankClass(idx)]">{{ idx + 1 }}</span>
          <span class="oq-rank-email">{{ user.email }}</span>
        </div>
        <span class="oq-rank-amt">${{ user.amount.toFixed(2) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps<{
  users: { user_id: number; email: string; amount: number }[]
}>()

// ── QUENCH 排名配色（取自 tokens.css）──────────────────────────────────
// rank 1/3: --warn-dim + --warn（琥珀金奖牌）
// rank 2:   --bg-2 + --ink-1（钢银）
// rest:     transparent + --ink-2（弱化）
function rankClass(idx: number): string {
  if (idx === 0) return 'oq-rank-gold'
  if (idx === 1) return 'oq-rank-silver'
  if (idx === 2) return 'oq-rank-gold'
  return 'oq-rank-dim'
}
</script>
