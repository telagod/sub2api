<template>
  <div class="flex items-center gap-1">
    <span
      :class="['w-[28px] shrink-0 rounded px-1 text-center text-[10px] font-medium', labelClass]"
    >{{ label }}</span>
    <span :class="['text-[11px] font-mono font-medium tabular-nums', percentClass]">{{ displayPercent }}</span>
    <span v-if="shouldShowResetTime" class="text-[10px] text-muted-foreground">{{ formatResetTime }}</span>
    <template v-if="windowStats && (windowStats.requests > 0 || windowStats.tokens > 0)">
      <span class="text-[9px] text-muted-foreground/60">·</span>
      <span class="text-[9px] text-muted-foreground tabular-nums">{{ formatRequests }}r</span>
      <span class="text-[9px] text-muted-foreground tabular-nums">{{ formatTokens }}t</span>
      <span class="text-[9px] text-emerald-400/80 tabular-nums">${{ formatAccountCost }}</span>
      <span v-if="windowStats.user_cost != null" class="text-[9px] text-muted-foreground tabular-nums">u${{ formatUserCost }}</span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import type { WindowStats } from '@/types'
import { formatCompactNumber } from '@/utils/format'

const props = defineProps<{
  label: string
  utilization: number
  resetsAt?: string | null
  color: 'indigo' | 'emerald' | 'purple' | 'amber'
  windowStats?: WindowStats | null
  showNowWhenIdle?: boolean
}>()

const { t } = useI18n()

const now = ref(new Date())
const { pause: pauseClock, resume: resumeClock } = useIntervalFn(
  () => { now.value = new Date() },
  60_000,
  { immediate: false },
)
if (props.resetsAt) resumeClock()
watch(
  () => props.resetsAt,
  (val) => { if (val) { now.value = new Date(); resumeClock() } else { pauseClock() } },
)

const labelClass = computed(() => {
  const colors = {
    indigo: 'bg-indigo-900/40 text-indigo-300',
    emerald: 'bg-emerald-900/40 text-emerald-400',
    purple: 'bg-purple-900/40 text-purple-300',
    amber: 'bg-amber-900/40 text-amber-400'
  }
  return colors[props.color]
})

const percentClass = computed(() => {
  if (props.utilization >= 100) return 'text-red-400'
  if (props.utilization >= 80) return 'text-amber-400'
  if (props.utilization >= 50) return 'text-foreground/85'
  return 'text-muted-foreground'
})

const displayPercent = computed(() => {
  const p = Math.round(props.utilization)
  return p > 999 ? '>999%' : `${p}%`
})

const shouldShowResetTime = computed(() => {
  if (props.resetsAt) return true
  return Boolean(props.showNowWhenIdle && props.utilization <= 0)
})

const formatResetTime = computed(() => {
  if (props.showNowWhenIdle && props.utilization <= 0) return t('usage.resetNow')
  if (!props.resetsAt) return '-'
  const date = new Date(props.resetsAt)
  const diffMs = date.getTime() - now.value.getTime()
  if (diffMs <= 0) return props.utilization > 0 ? t('usage.resetPending') : t('usage.resetNow')
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
  const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
  if (diffHours >= 24) return `${Math.floor(diffHours / 24)}d${diffHours % 24}h`
  if (diffHours > 0) return `${diffHours}h${diffMins}m`
  return `${diffMins}m`
})

const formatRequests = computed(() => {
  if (!props.windowStats) return ''
  return formatCompactNumber(props.windowStats.requests, { allowBillions: false })
})

const formatTokens = computed(() => {
  if (!props.windowStats) return ''
  return formatCompactNumber(props.windowStats.tokens)
})

const formatAccountCost = computed(() => {
  if (!props.windowStats) return '0'
  return props.windowStats.cost.toFixed(2)
})

const formatUserCost = computed(() => {
  if (!props.windowStats?.user_cost) return '0'
  return props.windowStats.user_cost.toFixed(2)
})
</script>
