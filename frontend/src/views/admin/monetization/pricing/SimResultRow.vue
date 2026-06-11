<template>
  <div class="flex items-center justify-between">
    <span class="text-xs" :style="{ color: 'var(--ink-1)' }">{{ label }}</span>
    <span :class="['q-money', large ? 'text-sm font-semibold' : 'text-xs']">
      {{ fmtUSD(value) }}
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value: number
  large?: boolean
}>()

function fmtUSD(v: number): string {
  return `$${v.toFixed(6)}`
}
</script>
