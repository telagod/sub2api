<template>
  <header class="quench-topbar">
    <!-- Breadcrumb -->
    <div class="quench-topbar__crumb">
      <span v-if="currentGroup" class="quench-topbar__crumb-group">
        {{ t(currentGroup.labelKey) }}
      </span>
      <span v-if="currentGroup && currentItem" class="quench-topbar__crumb-sep">›</span>
      <span v-if="currentItem" class="quench-topbar__crumb-page">
        {{ t(currentItem.labelKey) }}
      </span>
    </div>

    <!-- ⌘K pill -->
    <button
      class="quench-topbar__cmdk"
      data-tour="cmdk"
      @click="emit('openCommandPalette')"
      :title="t('nav.quench.openCommandPalette')"
    >
      <Search class="quench-topbar__cmdk-icon" />
      <span class="quench-topbar__cmdk-text">{{ t('nav.quench.searchPlaceholder') }}</span>
      <kbd class="quench-topbar__cmdk-kbd">⌘K</kbd>
    </button>

    <!-- Right section: theme + locale + user -->
    <div class="quench-topbar__right">
      <!-- Theme Toggle -->
      <button
        class="quench-topbar__icon-btn"
        @click="toggleTheme"
        :title="isDark ? t('nav.lightMode') : t('nav.darkMode')"
      >
        <Sun v-if="isDark" class="quench-topbar__btn-icon" />
        <Moon v-else class="quench-topbar__btn-icon" />
      </button>

      <!-- Locale Switcher -->
      <LocaleSwitcher />

      <!-- User avatar + dropdown -->
      <div class="quench-topbar__user" ref="dropdownRef">
        <button
          class="quench-topbar__avatar"
          @click="toggleDropdown"
          :title="displayName"
        >
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            :alt="displayName"
            class="h-full w-full object-cover"
          />
          <span v-else>{{ userInitials }}</span>
        </button>

        <transition name="quench-dropdown">
          <div v-if="dropdownOpen" class="quench-topbar__dropdown">
            <div class="quench-topbar__dropdown-header">
              <div class="quench-topbar__dropdown-name">{{ displayName }}</div>
              <div class="quench-topbar__dropdown-email">{{ user?.email }}</div>
            </div>
            <div class="quench-topbar__dropdown-body">
              <router-link
                to="/profile"
                class="quench-topbar__dropdown-item"
                @click="closeDropdown"
              >
                <UserIcon class="quench-topbar__dropdown-item-icon" />
                {{ t('nav.profile') }}
              </router-link>
            </div>
            <div class="quench-topbar__dropdown-footer">
              <button
                class="quench-topbar__dropdown-item quench-topbar__dropdown-item--danger"
                @click="handleLogout"
              >
                <LogOut class="quench-topbar__dropdown-item-icon" />
                {{ t('nav.logout') }}
              </button>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Search, Sun, Moon, User as UserIcon, LogOut } from 'lucide-vue-next'
import { useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import { resolveNavItem } from './nav'

const emit = defineEmits<{
  openCommandPalette: []
}>()

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.user)
const dropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const resolved = computed(() => resolveNavItem(route.path))
const currentGroup = computed(() => resolved.value?.group ?? null)
const currentItem = computed(() => resolved.value?.item ?? null)

const avatarUrl = computed(() => user.value?.avatar_url?.trim() || '')
const userInitials = computed(() => {
  if (!user.value) return ''
  if (user.value.username) return user.value.username.substring(0, 2).toUpperCase()
  if (user.value.email) return user.value.email.split('@')[0].substring(0, 2).toUpperCase()
  return ''
})
const displayName = computed(() => {
  if (!user.value) return ''
  return user.value.username || user.value.email?.split('@')[0] || ''
})

// Theme
const isDark = computed(() => {
  return document.documentElement.getAttribute('data-theme') !== 'light'
})

function toggleTheme() {
  const html = document.documentElement
  const current = html.getAttribute('data-theme')
  html.setAttribute('data-theme', current === 'light' ? 'dark' : 'light')
  try {
    localStorage.setItem('theme', current === 'light' ? 'dark' : 'light')
  } catch {}
}

function toggleDropdown() {
  dropdownOpen.value = !dropdownOpen.value
}

function closeDropdown() {
  dropdownOpen.value = false
}

async function handleLogout() {
  closeDropdown()
  try {
    await authStore.logout()
  } catch {}
  await router.push('/login')
}

function handleClickOutside(event: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.quench-topbar {
  height: 54px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 22px;
  border-bottom: 1px solid #20242c;
  background: rgba(13, 14, 18, 0.72);
  backdrop-filter: blur(8px);
  position: sticky;
  top: 0;
  z-index: 20;
}

/* Breadcrumb */
.quench-topbar__crumb {
  display: flex;
  align-items: center;
  gap: 0;
  font-size: 12.5px;
  min-width: 0;
}

.quench-topbar__crumb-group {
  color: #5c6470;
}

.quench-topbar__crumb-sep {
  color: #5c6470;
  margin: 0 7px;
  opacity: 0.5;
}

.quench-topbar__crumb-page {
  color: #e8ebf0;
  font-weight: 600;
}

/* ⌘K pill */
.quench-topbar__cmdk {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 220px;
  background: #171a20;
  border: 1px solid #20242c;
  border-radius: 9px;
  padding: 6px 11px;
  color: #5c6470;
  cursor: pointer;
  font-size: 12px;
  font-family: inherit;
  transition: border-color 0.15s, box-shadow 0.2s;
}

.quench-topbar__cmdk:hover {
  border-color: rgba(92, 168, 255, 0.4);
  box-shadow: 0 0 14px rgba(92, 168, 255, 0.12);
}

.quench-topbar__cmdk-icon {
  width: 13px;
  height: 13px;
  flex-shrink: 0;
}

.quench-topbar__cmdk-text {
  flex: 1;
  text-align: left;
}

.quench-topbar__cmdk-kbd {
  margin-left: auto;
  font-family: 'IBM Plex Mono', 'SFMono-Regular', monospace;
  font-size: 10px;
  color: #97a0af;
  background: #1f232b;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid #2f3540;
}

/* Right */
.quench-topbar__right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.quench-topbar__icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: transparent;
  border: none;
  cursor: pointer;
  color: #97a0af;
  transition: background 0.15s, color 0.15s;
}

.quench-topbar__icon-btn:hover {
  background: #171a20;
  color: #e8ebf0;
}

.quench-topbar__btn-icon {
  width: 15px;
  height: 15px;
}

/* User avatar */
.quench-topbar__user {
  position: relative;
}

.quench-topbar__avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  flex-shrink: 0;
  background: linear-gradient(160deg, #d7dee8, #79808e);
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 12px;
  color: #14171d;
  cursor: pointer;
  border: none;
  overflow: hidden;
}

/* Dropdown */
.quench-topbar__dropdown {
  position: absolute;
  right: 0;
  top: calc(100% + 6px);
  width: 200px;
  background: #101216;
  border: 1px solid #20242c;
  border-radius: 10px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.6);
  z-index: 50;
  overflow: hidden;
}

.quench-topbar__dropdown-header {
  padding: 12px 14px 10px;
  border-bottom: 1px solid #20242c;
}

.quench-topbar__dropdown-name {
  font-size: 13px;
  font-weight: 600;
  color: #e8ebf0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.quench-topbar__dropdown-email {
  font-size: 11px;
  color: #5c6470;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 2px;
}

.quench-topbar__dropdown-body {
  padding: 6px 4px;
}

.quench-topbar__dropdown-footer {
  padding: 6px 4px;
  border-top: 1px solid #20242c;
}

.quench-topbar__dropdown-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 7px 10px;
  border-radius: 7px;
  font-size: 12.5px;
  color: #97a0af;
  cursor: pointer;
  text-decoration: none;
  background: transparent;
  border: none;
  width: 100%;
  font-family: inherit;
  transition: background 0.12s, color 0.12s;
}

.quench-topbar__dropdown-item:hover {
  background: #171a20;
  color: #e8ebf0;
}

.quench-topbar__dropdown-item--danger {
  color: #f25c69;
}

.quench-topbar__dropdown-item--danger:hover {
  background: rgba(242, 92, 105, 0.1);
  color: #f25c69;
}

.quench-topbar__dropdown-item-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  opacity: 0.85;
}

/* Transition */
.quench-dropdown-enter-active,
.quench-dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.quench-dropdown-enter-from,
.quench-dropdown-leave-to {
  opacity: 0;
  transform: scale(0.96) translateY(-4px);
}

/* 键盘焦点：topbar 所有交互按钮 */
.quench-topbar__cmdk:focus-visible,
.quench-topbar__icon-btn:focus-visible,
.quench-topbar__avatar:focus-visible {
  outline: none;
  box-shadow: 0 0 0 1.5px rgba(92, 168, 255, 0.65), 0 0 16px rgba(92, 168, 255, 0.25);
}

.quench-topbar__dropdown-item:focus-visible {
  outline: none;
  box-shadow: inset 0 0 0 1.5px rgba(92, 168, 255, 0.5);
  border-radius: 7px;
}

@media (prefers-reduced-motion: reduce) {
  .quench-topbar__cmdk,
  .quench-topbar__icon-btn,
  .quench-topbar__dropdown-item { transition: none; }
  .quench-dropdown-enter-active,
  .quench-dropdown-leave-active { transition: none; }
  .quench-dropdown-enter-from,
  .quench-dropdown-leave-to { transform: none; }
}
</style>
