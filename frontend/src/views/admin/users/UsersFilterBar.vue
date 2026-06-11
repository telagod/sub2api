<template>
  <div class="ufb-bar">
    <!-- 搜索框 -->
    <div class="ufb-search" :class="{ 'ufb-search-focus': focused }">
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
        <circle cx="6" cy="6" r="4.5" stroke="currentColor" stroke-width="1.3"/>
        <path d="M9.5 9.5L12.5 12.5" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/>
      </svg>
      <input
        :value="search"
        :placeholder="t('admin.usersQuench.searchPlaceholder')"
        class="ufb-input"
        @focus="focused = true"
        @blur="focused = false"
        @input="$emit('update:search', ($event.target as HTMLInputElement).value)"
        @keyup.enter="$emit('commit-search')"
      />
      <button v-if="search" class="ufb-clear-x" @click="$emit('update:search', ''); $emit('commit-search')" :aria-label="t('admin.usersQuench.clearSearch')">✕</button>
    </div>

    <!-- 角色筛选芯片 -->
    <div class="ufb-chip-wrap">
      <button
        class="ufb-chip"
        :class="{ 'ufb-chip-on': role }"
        @click.stop="showRole = !showRole; showStatus = false"
      >
        {{ t('admin.usersQuench.filterRole') }} <b>{{ roleLabel }}</b>
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none" aria-hidden="true">
          <path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
        </svg>
      </button>
      <div v-if="showRole" class="ufb-menu" @click.stop>
        <button
          v-for="opt in ROLE_OPTIONS"
          :key="opt.value"
          class="ufb-menu-item"
          :class="{ on: role === opt.value }"
          @click="$emit('update:role', opt.value); showRole = false"
        >{{ opt.label }}</button>
      </div>
    </div>

    <!-- 状态筛选芯片 -->
    <div class="ufb-chip-wrap">
      <button
        class="ufb-chip"
        :class="{ 'ufb-chip-on': status }"
        @click.stop="showStatus = !showStatus; showRole = false"
      >
        {{ t('admin.usersQuench.filterStatus') }} <b>{{ statusLabel }}</b>
        <svg width="10" height="10" viewBox="0 0 10 10" fill="none" aria-hidden="true">
          <path d="M2 3.5L5 6.5L8 3.5" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
        </svg>
      </button>
      <div v-if="showStatus" class="ufb-menu" @click.stop>
        <button
          v-for="opt in STATUS_OPTIONS"
          :key="opt.value"
          class="ufb-menu-item"
          :class="{ on: status === opt.value }"
          @click="$emit('update:status', opt.value); showStatus = false"
        >{{ opt.label }}</button>
      </div>
    </div>

    <!-- 清空筛选 -->
    <button v-if="hasFilters" class="ufb-clear-all" @click="$emit('clear')">{{ t('admin.usersQuench.clearFilters') }}</button>

    <!-- 密度切换 -->
    <div class="ufb-seg">
      <button :class="{ on: density === 'comfortable' }" @click="$emit('update:density', 'comfortable')">{{ t('admin.usersQuench.densityComfortable') }}</button>
      <button :class="{ on: density === 'compact' }" @click="$emit('update:density', 'compact')">{{ t('admin.usersQuench.densityCompact') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  search: string
  role: string
  status: string
  density: 'comfortable' | 'compact'
}>()

defineEmits<{
  'update:search': [val: string]
  'update:role': [val: string]
  'update:status': [val: string]
  'update:density': [val: 'comfortable' | 'compact']
  'commit-search': []
  'clear': []
}>()

const focused = ref(false)
const showRole = ref(false)
const showStatus = ref(false)

const ROLE_OPTIONS = computed(() => [
  { value: '', label: t('admin.users.allRoles') },
  { value: 'admin', label: t('admin.usersQuench.roleAdmin') },
  { value: 'user', label: t('admin.users.roles.user') },
])
const STATUS_OPTIONS = computed(() => [
  { value: '', label: t('admin.users.allStatuses') },
  { value: 'active', label: t('admin.usersQuench.statusActive') },
  { value: 'disabled', label: t('admin.usersQuench.statusDisabled') },
])

const roleLabel = computed(() => ROLE_OPTIONS.value.find(o => o.value === props.role)?.label ?? t('admin.users.allRoles'))
const statusLabel = computed(() => STATUS_OPTIONS.value.find(o => o.value === props.status)?.label ?? t('admin.users.allStatuses'))
const hasFilters = computed(() => !!(props.search || props.role || props.status))

// 点击外部关闭
function onDocClick() { showRole.value = false; showStatus.value = false }
import { onMounted, onUnmounted } from 'vue'
onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>

<style scoped>
.ufb-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
  font-family: var(--font-ui, "Archivo", "PingFang SC", sans-serif);
}

.ufb-search {
  display: flex;
  align-items: center;
  gap: 7px;
  background: var(--bg-1, #101216);
  border: 1px solid var(--line-1, #2F3540);
  border-radius: 9px;
  padding: 5px 10px;
  min-width: 220px;
  color: var(--ink-2, #5C6470);
  transition: border-color 0.15s, box-shadow 0.15s;
}
.ufb-search-focus {
  border-color: rgba(92, 168, 255, 0.5);
  box-shadow: var(--glow-focus, 0 0 0 1.5px rgba(92,168,255,.65), 0 0 20px rgba(92,168,255,.28));
}
.ufb-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--ink-0, #E8EBF0);
  font-size: 12.5px;
  font-family: inherit;
  outline: none;
}
.ufb-input::placeholder { color: var(--ink-2, #5C6470); }
.ufb-clear-x {
  font-size: 10px; color: var(--ink-2, #5C6470); background: transparent;
  border: none; cursor: pointer; padding: 2px;
}
.ufb-clear-x:hover { color: var(--ink-0, #E8EBF0); }

.ufb-chip-wrap { position: relative; }

.ufb-chip {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 10px; border-radius: 9px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-1, #101216);
  color: var(--ink-1, #97A0AF); font-size: 12px;
  cursor: pointer; font-family: inherit; white-space: nowrap;
  transition: border-color 0.15s;
}
.ufb-chip:hover { border-color: #3D4554; }
.ufb-chip-on { border-color: rgba(92,168,255,0.4); color: var(--azure,#5CA8FF); }
.ufb-chip b { color: var(--ink-0,#E8EBF0); font-weight: 600; }

.ufb-menu {
  position: absolute; top: calc(100% + 4px); left: 0; min-width: 120px;
  background: var(--bg-2,#171A20); border: 1px solid var(--line-1,#2F3540);
  border-radius: 9px; padding: 5px; z-index: 50;
  box-shadow: 0 8px 24px rgba(0,0,0,.4);
}
.ufb-menu-item {
  display: block; width: 100%; padding: 6px 10px;
  text-align: left; border: none; background: transparent;
  color: var(--ink-1,#97A0AF); font-size: 12px; font-family: inherit;
  cursor: pointer; border-radius: 6px; transition: background 0.1s, color 0.1s;
}
.ufb-menu-item:hover { background: var(--bg-1,#101216); color: var(--ink-0,#E8EBF0); }
.ufb-menu-item.on { color: var(--azure,#5CA8FF); font-weight: 600; }

.ufb-clear-all {
  padding: 5px 10px; border-radius: 9px;
  border: 1px solid rgba(242,92,105,0.3);
  background: transparent; color: var(--bad,#F25C69);
  font-size: 12px; cursor: pointer; font-family: inherit;
}
.ufb-clear-all:hover { background: var(--bad-dim,rgba(242,92,105,.1)); }

.ufb-seg {
  display: inline-flex; border: 1px solid var(--line-1,#2F3540);
  border-radius: 8px; overflow: hidden; margin-left: auto;
}
.ufb-seg button {
  padding: 4px 10px; background: transparent; border: none;
  color: var(--ink-2,#5C6470); font-size: 11.5px; font-family: inherit;
  cursor: pointer; transition: background 0.12s, color 0.12s;
}
.ufb-seg button.on { background: var(--bg-2,#171A20); color: var(--ink-0,#E8EBF0); }

.ufb-chip:focus-visible,
.ufb-clear-all:focus-visible,
.ufb-menu-item:focus-visible,
.ufb-seg button:focus-visible {
  outline: none;
  box-shadow: var(--glow-focus, 0 0 0 1.5px rgba(92,168,255,.65), 0 0 20px rgba(92,168,255,.28));
  border-color: rgba(92,168,255,.5);
}

@media (prefers-reduced-motion: reduce) {
  .ufb-search, .ufb-chip, .ufb-menu-item, .ufb-seg button { transition: none; }
}
</style>
