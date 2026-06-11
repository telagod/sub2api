<template>
  <Teleport to="body">
    <Transition name="cmdk-overlay">
      <div
        v-if="modelValue"
        class="cmdk-overlay"
        @click.self="close"
      >
        <div class="cmdk-panel" role="dialog" aria-modal="true" :aria-label="t('nav.quench.commandPalette')">
          <!-- Search input -->
          <div class="cmdk-input-wrap">
            <Search class="cmdk-search-icon" />
            <input
              ref="inputRef"
              v-model="query"
              class="cmdk-input"
              :placeholder="t('nav.quench.commandPalettePlaceholder')"
              @keydown="handleKeydown"
            />
            <kbd class="cmdk-esc-kbd">Esc</kbd>
          </div>

          <!-- Results -->
          <div class="cmdk-results" ref="listRef">
            <template v-if="filteredItems.length > 0">
              <button
                v-for="(item, idx) in filteredItems"
                :key="item.key"
                class="cmdk-item"
                :class="{ 'cmdk-item--active': idx === activeIndex }"
                @click="selectItem(item)"
                @mouseenter="activeIndex = idx"
              >
                <component :is="item.icon" class="cmdk-item-icon" />
                <span class="cmdk-item-label">{{ t(item.labelKey) }}</span>
                <span class="cmdk-item-group">{{ t(item.groupLabelKey) }}</span>
              </button>
            </template>
            <div v-else class="cmdk-empty">
              {{ t('nav.quench.noResults') }}
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useMagicKeys } from '@vueuse/core'
import { Search } from 'lucide-vue-next'
import { flatNavItems } from './nav'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const { t } = useI18n()
const router = useRouter()

const query = ref('')
const activeIndex = ref(0)
const inputRef = ref<HTMLInputElement | null>(null)
const listRef = ref<HTMLElement | null>(null)

const allItems = flatNavItems()

const filteredItems = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return allItems
  return allItems.filter((item) => {
    const label = t(item.labelKey).toLowerCase()
    const group = t(item.groupLabelKey).toLowerCase()
    return label.includes(q) || group.includes(q)
  })
})

watch(query, () => {
  activeIndex.value = 0
})

watch(
  () => props.modelValue,
  async (val) => {
    if (val) {
      query.value = ''
      activeIndex.value = 0
      await nextTick()
      inputRef.value?.focus()
    }
  }
)

// Global ⌘K / Ctrl+K
const { Meta_k, Ctrl_k } = useMagicKeys()
watch([Meta_k, Ctrl_k], ([mk, ck]) => {
  if (mk || ck) {
    emit('update:modelValue', !props.modelValue)
  }
})

function close() {
  emit('update:modelValue', false)
}

function selectItem(item: ReturnType<typeof flatNavItems>[number]) {
  router.push(item.path)
  close()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    close()
    return
  }
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value = Math.min(activeIndex.value + 1, filteredItems.value.length - 1)
    scrollActiveIntoView()
    return
  }
  if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value = Math.max(activeIndex.value - 1, 0)
    scrollActiveIntoView()
    return
  }
  if (e.key === 'Enter') {
    const item = filteredItems.value[activeIndex.value]
    if (item) selectItem(item)
  }
}

function scrollActiveIntoView() {
  nextTick(() => {
    const list = listRef.value
    if (!list) return
    const active = list.querySelectorAll('.cmdk-item')[activeIndex.value] as HTMLElement | undefined
    active?.scrollIntoView({ block: 'nearest' })
  })
}
</script>

<style scoped>
.cmdk-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 120px;
}

.cmdk-panel {
  width: 100%;
  max-width: 540px;
  background: #101216;
  border: 1px solid #2f3540;
  border-radius: 14px;
  box-shadow: 0 32px 80px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(92, 168, 255, 0.08);
  overflow: hidden;
}

/* Input row */
.cmdk-input-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  border-bottom: 1px solid #20242c;
}

.cmdk-search-icon {
  width: 16px;
  height: 16px;
  color: #5c6470;
  flex-shrink: 0;
}

.cmdk-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #e8ebf0;
  font-size: 14px;
  font-family: inherit;
}

.cmdk-input::placeholder {
  color: #5c6470;
}

.cmdk-esc-kbd {
  font-family: 'IBM Plex Mono', 'SFMono-Regular', monospace;
  font-size: 10px;
  color: #5c6470;
  background: #1f232b;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid #2f3540;
  flex-shrink: 0;
}

/* Results */
.cmdk-results {
  max-height: 340px;
  overflow-y: auto;
  padding: 6px;
  scrollbar-width: thin;
  scrollbar-color: #2f3540 transparent;
}

.cmdk-results::-webkit-scrollbar {
  width: 4px;
}

.cmdk-results::-webkit-scrollbar-thumb {
  background: #2f3540;
  border-radius: 4px;
}

.cmdk-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 8px 12px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #97a0af;
  font-size: 13px;
  font-family: inherit;
  cursor: pointer;
  text-align: left;
  transition: background 0.1s, color 0.1s;
}

.cmdk-item:hover,
.cmdk-item--active {
  background: rgba(92, 168, 255, 0.1);
  color: #e8ebf0;
  box-shadow: inset 0 0 0 1px rgba(92, 168, 255, 0.18);
}

/* Shadow-glow-focus style from design tokens for focused item */
.cmdk-item--active {
  box-shadow:
    inset 0 0 0 1px rgba(92, 168, 255, 0.25),
    0 0 12px rgba(92, 168, 255, 0.1);
}

.cmdk-item-icon {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  opacity: 0.8;
}

.cmdk-item-label {
  flex: 1;
  font-weight: 500;
}

.cmdk-item-group {
  font-size: 10.5px;
  color: #5c6470;
  font-family: 'IBM Plex Mono', 'SFMono-Regular', monospace;
  letter-spacing: 0.06em;
}

.cmdk-empty {
  padding: 24px;
  text-align: center;
  color: #5c6470;
  font-size: 13px;
}

/* Transition */
.cmdk-overlay-enter-active,
.cmdk-overlay-leave-active {
  transition: opacity 0.18s ease;
}

.cmdk-overlay-enter-active .cmdk-panel,
.cmdk-overlay-leave-active .cmdk-panel {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.cmdk-overlay-enter-from,
.cmdk-overlay-leave-to {
  opacity: 0;
}

.cmdk-overlay-enter-from .cmdk-panel,
.cmdk-overlay-leave-to .cmdk-panel {
  opacity: 0;
  transform: scale(0.96) translateY(-10px);
}
</style>
