<template>
  <div class="sr-row">
    <span class="sr-label">{{ label }}</span>
    <span class="q-money sr-value" :class="large ? 'sr-large' : ''">
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

<style scoped>
.sr-row { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.sr-label { font-size: 11px; color: var(--ink-1); }
.sr-value { font-size: 11px; }
.sr-large { font-size: 13px; font-weight: 600; }
</style>
