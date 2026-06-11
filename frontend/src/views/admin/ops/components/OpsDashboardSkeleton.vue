<script setup lang="ts">
interface Props {
  fullscreen?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  fullscreen: false
})
</script>

<template>
  <div style="display:flex;flex-direction:column;gap:18px;">
    <!-- Header skeleton -->
    <div class="od-card" :style="props.fullscreen ? 'padding:28px;' : 'padding:18px;'">
      <div style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:12px;border-bottom:1px solid var(--line-0,#20242C);padding-bottom:14px;">
        <div style="display:flex;flex-direction:column;gap:6px;">
          <div style="height:20px;width:160px;border-radius:4px;background:var(--bg-2,#171A20);" class="animate-pulse"></div>
          <div style="height:10px;width:280px;border-radius:4px;background:var(--bg-2,#171A20);opacity:.7;" class="animate-pulse"></div>
        </div>
        <div v-if="!props.fullscreen" style="display:flex;flex-wrap:wrap;gap:8px;">
          <div v-for="w in [120,140,120,32,96,96,32]" :key="w" :style="`height:32px;width:${w}px;border-radius:6px;background:var(--bg-2,#171A20);`" class="animate-pulse"></div>
        </div>
      </div>
      <div style="margin-top:18px;display:grid;grid-template-columns:5fr 7fr;gap:16px;">
        <div class="od-card" style="padding:14px;background:var(--bg-2,#171A20);">
          <div style="display:grid;grid-template-columns:180px 1fr;gap:16px;align-items:center;">
            <div style="height:96px;border-radius:8px;background:var(--bg-3,#0E1014);" class="animate-pulse"></div>
            <div style="display:flex;flex-direction:column;gap:10px;">
              <div style="height:14px;width:110px;border-radius:4px;background:var(--bg-3,#0E1014);" class="animate-pulse"></div>
              <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">
                <div v-for="i in 4" :key="i" style="height:50px;border-radius:6px;background:var(--bg-3,#0E1014);" class="animate-pulse"></div>
              </div>
            </div>
          </div>
        </div>
        <div style="display:grid;grid-template-columns:repeat(3,1fr);gap:10px;align-content:center;">
          <div v-for="i in 6" :key="i" class="od-card animate-pulse" style="height:72px;background:var(--bg-2,#171A20);"></div>
        </div>
      </div>
    </div>

    <!-- Row: Concurrency + Throughput -->
    <div style="display:grid;grid-template-columns:1fr 1fr 2fr;gap:16px;">
      <div v-for="(_span,i) in [1,1,2]" :key="i" class="od-chart-card animate-pulse" style="min-height:320px;">
        <div style="height:14px;width:140px;border-radius:4px;background:var(--bg-2,#171A20);margin-bottom:16px;"></div>
        <div style="flex:1;border-radius:6px;background:var(--bg-2,#171A20);"></div>
      </div>
    </div>

    <!-- Row: Visual Analysis 3-up -->
    <div style="display:grid;grid-template-columns:repeat(3,1fr);gap:16px;">
      <div v-for="i in 3" :key="i" class="od-chart-card animate-pulse" style="min-height:240px;">
        <div style="height:14px;width:140px;border-radius:4px;background:var(--bg-2,#171A20);margin-bottom:16px;"></div>
        <div style="flex:1;border-radius:6px;background:var(--bg-2,#171A20);"></div>
      </div>
    </div>

    <!-- Alert Events skeleton -->
    <div class="od-card" :style="props.fullscreen ? 'padding:28px;' : 'padding:18px;'">
      <div style="display:flex;flex-wrap:wrap;align-items:center;justify-content:space-between;gap:10px;margin-bottom:16px;">
        <div style="height:14px;width:180px;border-radius:4px;background:var(--bg-2,#171A20);" class="animate-pulse"></div>
        <div v-if="!props.fullscreen" style="display:flex;gap:8px;">
          <div v-for="w in [120,96,96]" :key="w" :style="`height:32px;width:${w}px;border-radius:6px;background:var(--bg-2,#171A20);`" class="animate-pulse"></div>
        </div>
      </div>
      <div style="display:flex;flex-direction:column;gap:8px;">
        <div v-for="i in 5" :key="i" class="od-card animate-pulse" style="padding:12px;display:flex;align-items:center;justify-content:space-between;gap:12px;background:var(--bg-2,#171A20);">
          <div style="display:flex;flex-direction:column;gap:6px;flex:1;">
            <div style="height:10px;width:200px;border-radius:3px;background:var(--bg-3,#0E1014);"></div>
            <div style="height:10px;width:280px;border-radius:3px;background:var(--bg-3,#0E1014);opacity:.7;"></div>
          </div>
          <div style="height:26px;width:72px;border-radius:6px;background:var(--bg-3,#0E1014);"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style src="../ops-quench.css"></style>
