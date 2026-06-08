<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-center gap-3">
      <div class="flex-1 min-w-0">
        <slot name="search" />
      </div>

      <div class="flex items-center gap-2">
        <button
          v-if="hasFilters"
          type="button"
          class="btn btn-secondary gap-1.5"
          :class="expanded && 'bg-accent'"
          @click="toggle"
        >
          <Icon name="filter" size="sm" />
          <span class="hidden sm:inline">{{ t('common.filters') }}</span>
          <span
            v-if="activeCount > 0"
            class="inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-primary-400 px-1 text-[10px] font-bold text-background"
          >
            {{ activeCount }}
          </span>
          <Icon
            name="chevronDown"
            size="xs"
            class="transition-transform duration-150"
            :class="expanded && 'rotate-180'"
          />
        </button>

        <slot name="actions" />
      </div>
    </div>

    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      leave-active-class="transition-all duration-150 ease-in"
      enter-from-class="opacity-0 -translate-y-1 max-h-0"
      enter-to-class="opacity-100 translate-y-0 max-h-40"
      leave-from-class="opacity-100 translate-y-0 max-h-40"
      leave-to-class="opacity-0 -translate-y-1 max-h-0"
    >
      <div v-show="expanded" class="overflow-hidden">
        <div class="flex flex-wrap items-center gap-2 rounded-lg border border-border bg-card/50 px-3 py-2.5">
          <slot name="filters" />
          <button
            v-if="activeCount > 0"
            type="button"
            class="ml-auto text-xs text-muted-foreground transition-colors hover:text-foreground"
            @click="$emit('clear')"
          >
            {{ t('common.clearFilters') }}
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useSlots } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const slots = useSlots()

interface Props {
  activeCount?: number
  storageKey?: string
}

const props = withDefaults(defineProps<Props>(), {
  activeCount: 0
})

defineEmits<{
  (e: 'clear'): void
}>()

const hasFilters = computed(() => !!slots.filters)

const stored = props.storageKey
  ? localStorage.getItem(`filter-expanded:${props.storageKey}`)
  : null
const expanded = ref(stored === 'true')

const toggle = () => {
  expanded.value = !expanded.value
  if (props.storageKey) {
    localStorage.setItem(`filter-expanded:${props.storageKey}`, String(expanded.value))
  }
}
</script>
