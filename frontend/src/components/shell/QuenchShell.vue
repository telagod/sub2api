<template>
  <div class="quench-shell">
    <!-- Sidebar -->
    <QuenchSidebar />

    <!-- Main area -->
    <div class="quench-shell__main">
      <!-- Topbar -->
      <QuenchTopbar @open-command-palette="commandPaletteOpen = true" />

      <!-- Page content -->
      <main class="quench-shell__content">
        <slot />
      </main>
    </div>

    <!-- Command Palette (global) -->
    <CommandPalette v-model="commandPaletteOpen" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import QuenchSidebar from './QuenchSidebar.vue'
import QuenchTopbar from './QuenchTopbar.vue'
import CommandPalette from './CommandPalette.vue'

const commandPaletteOpen = ref(false)
</script>

<style scoped>
.quench-shell {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: #08090b;
  /* Ambient background effects matching the mockup */
  position: relative;
}

.quench-shell::before {
  content: '';
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background:
    radial-gradient(900px 480px at 72% -120px, rgba(92, 168, 255, 0.055), transparent 65%),
    radial-gradient(700px 400px at 10% 110%, rgba(92, 168, 255, 0.025), transparent 60%),
    linear-gradient(rgba(232, 235, 240, 0.016) 1px, transparent 1px),
    linear-gradient(90deg, rgba(232, 235, 240, 0.016) 1px, transparent 1px);
  background-size: auto, auto, 48px 48px, 48px 48px;
}

.quench-shell__main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
}

.quench-shell__content {
  flex: 1;
  overflow-y: auto;
  padding: 24px 26px 80px;
  scrollbar-width: thin;
  scrollbar-color: #2f3540 transparent;
}

.quench-shell__content::-webkit-scrollbar {
  width: 8px;
}

.quench-shell__content::-webkit-scrollbar-thumb {
  background: #2f3540;
  border-radius: 6px;
  border: 2px solid #08090b;
}

.quench-shell__content::-webkit-scrollbar-track {
  background: transparent;
}
</style>
