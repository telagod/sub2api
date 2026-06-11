<template>
  <aside class="quench-sidebar hidden lg:flex">
    <!-- Brand -->
    <div class="quench-sidebar__brand">
      <div class="quench-sidebar__brand-mark">
        <img
          v-if="siteLogo"
          :src="siteLogo"
          alt="logo"
          class="h-full w-full object-contain"
        />
        <span v-else class="quench-sidebar__brand-initial">S</span>
      </div>
      <div class="quench-sidebar__brand-text">
        <span class="quench-sidebar__brand-name">{{ siteName }}</span>
        <span class="quench-sidebar__brand-sub">ADMIN</span>
      </div>
    </div>

    <!-- Nav groups -->
    <nav class="quench-sidebar__nav" data-tour="nav-root">
      <div
        v-for="group in navGroups"
        :key="group.key"
        class="quench-sidebar__group"
        :data-tour="`nav-group-${group.key}`"
      >
        <div class="quench-sidebar__group-label">
          {{ t(group.labelKey) }}
        </div>
        <router-link
          v-for="item in group.items"
          :key="item.key"
          :to="item.path"
          class="quench-sidebar__item"
          :class="{ 'quench-sidebar__item--active': isActive(item.path) }"
          :aria-current="isActive(item.path) ? 'page' : undefined"
          :data-tour="`nav-${item.key}`"
        >
          <component :is="item.icon" class="quench-sidebar__item-icon" />
          <span class="quench-sidebar__item-label">{{ t(item.labelKey) }}</span>
        </router-link>
      </div>
    </nav>

    <!-- Footer: env chip -->
    <div class="quench-sidebar__foot">
      <div class="quench-sidebar__env-chip">
        <i class="quench-sidebar__env-dot"></i>
        <span>v{{ siteVersion || '—' }}</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { navGroups } from './nav'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()

const siteName = computed(() => appStore.siteName)
const siteLogo = computed(() => appStore.siteLogo)
const siteVersion = computed(() => appStore.siteVersion)

function isActive(path: string): boolean {
  if (path === '/admin/dashboard') {
    return route.path === path
  }
  return route.path === path || route.path.startsWith(path + '/')
}
</script>

<style scoped>
.quench-sidebar {
  width: 228px;
  flex-shrink: 0;
  flex-direction: column;
  background: linear-gradient(180deg, #0f1115, #0b0c0f);
  border-right: 1px solid #20242c;
  height: 100vh;
  position: sticky;
  top: 0;
  overflow: hidden;
}

/* Brand */
.quench-sidebar__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 18px 16px;
  border-bottom: 1px solid #20242c;
  flex-shrink: 0;
}

.quench-sidebar__brand-mark {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  flex-shrink: 0;
  background: linear-gradient(160deg, #e4e9f0, #8b95a5 55%, #5c6573);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.5), 0 2px 10px rgba(140, 196, 255, 0.18);
  display: grid;
  place-items: center;
  overflow: hidden;
}

.quench-sidebar__brand-initial {
  color: #14171d;
  font-weight: 800;
  font-size: 14px;
}

.quench-sidebar__brand-text {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.quench-sidebar__brand-name {
  font-weight: 700;
  font-size: 14px;
  letter-spacing: 0.04em;
  color: #e8ebf0;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.quench-sidebar__brand-sub {
  font-size: 9.5px;
  color: #5c6470;
  letter-spacing: 0.22em;
  font-family: 'IBM Plex Mono', 'SFMono-Regular', monospace;
}

/* Nav */
.quench-sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
  scrollbar-width: thin;
  scrollbar-color: #2f3540 transparent;
}

.quench-sidebar__nav::-webkit-scrollbar {
  width: 4px;
}

.quench-sidebar__nav::-webkit-scrollbar-thumb {
  background: #2f3540;
  border-radius: 4px;
}

.quench-sidebar__group {
  margin-bottom: 14px;
}

.quench-sidebar__group-label {
  font-size: 10px;
  letter-spacing: 0.18em;
  color: #5c6470;
  font-weight: 600;
  padding: 6px 10px 5px;
  text-transform: uppercase;
}

.quench-sidebar__item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 7px 10px;
  border-radius: 8px;
  color: #97a0af;
  cursor: pointer;
  user-select: none;
  position: relative;
  text-decoration: none;
  font-weight: 500;
  font-size: 13px;
  transition: background 0.15s ease, color 0.15s ease;
}

.quench-sidebar__item:hover {
  background: #171a20;
  color: #e8ebf0;
}

.quench-sidebar__item--active {
  background: rgba(92, 168, 255, 0.12);
  color: #8cc4ff;
  box-shadow: inset 0 0 0 1px rgba(92, 168, 255, 0.14);
}

.quench-sidebar__item--active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 7px;
  bottom: 7px;
  width: 2px;
  background: #5ca8ff;
  border-radius: 2px;
  box-shadow: 0 0 8px rgba(92, 168, 255, 0.6);
}

.quench-sidebar__item-icon {
  width: 15px;
  height: 15px;
  flex-shrink: 0;
  opacity: 0.9;
}

.quench-sidebar__item-label {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 键盘焦点：azure glow，不被 outline:none 干掉 */
.quench-sidebar__item:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px rgba(92, 168, 255, 0.65), 0 0 16px rgba(92, 168, 255, 0.2);
}

@media (prefers-reduced-motion: reduce) {
  .quench-sidebar__item { transition: none; }
}

/* Footer */
.quench-sidebar__foot {
  padding: 10px 16px 12px;
  border-top: 1px solid #20242c;
  flex-shrink: 0;
}

.quench-sidebar__env-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: 'IBM Plex Mono', 'SFMono-Regular', monospace;
  font-size: 10px;
  color: #97a0af;
  background: #171a20;
  border: 1px solid #20242c;
  padding: 4px 9px;
  border-radius: 99px;
}

.quench-sidebar__env-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #46c98c;
  display: inline-block;
}
</style>
