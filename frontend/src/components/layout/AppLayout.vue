<template>
  <!-- Admin routes: Quench Shell -->
  <QuenchShell v-if="isAdminRoute">
    <slot />
  </QuenchShell>

  <!-- Non-admin routes: original layout preserved -->
  <template v-else>
    <div class="min-h-screen bg-background">
      <!-- Background Decoration -->
      <div class="pointer-events-none fixed inset-0 bg-mesh-gradient"></div>

      <!-- Sidebar -->
      <AppSidebar />

      <!-- Main Content Area -->
      <div
        class="relative min-h-screen transition-all duration-300"
        :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
      >
        <!-- Header -->
        <AppHeader />

        <!-- Main Content -->
        <main class="p-3 sm:p-4 md:p-6 lg:p-8">
          <slot />
        </main>
      </div>
    </div>
  </template>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import QuenchShell from '@/components/shell/QuenchShell.vue'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

// Route-based: admin paths get Quench Shell
const isAdminRoute = computed(() => route.path.startsWith('/admin'))

// admin_guide_quench：QUENCH 壳重写后的新引导，换 key 让旧用户重看一次
const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide_quench' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
