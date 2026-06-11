<template>
  <div class="dq-alert-row">
    <!-- 账号池异常 -->
    <div class="dq-alert-card rise" :style="`animation-delay:${baseDelay ?? 0}s`" role="button" tabindex="0" @click="$emit('nav', '/admin/accounts')" @keyup.enter="$emit('nav', '/admin/accounts')">
      <div class="dq-ac-head">
        <span class="dq-ac-title">账号池</span>
        <span class="dq-ac-more">查看 →</span>
      </div>
      <div class="dq-ac-nums">
        <div class="dq-ac-num">
          <span class="sdot" :class="(errorAccounts ?? 0) > 0 ? 'bad' : 'ok'"></span>
          <span class="dq-ac-val" :class="(errorAccounts ?? 0) > 0 ? 'dq-bad' : ''">{{ errorAccounts ?? 0 }}</span>
          <span class="dq-ac-lbl">错误</span>
        </div>
        <div class="dq-ac-num">
          <span class="sdot" :class="(ratelimitAccounts ?? 0) > 0 ? 'warn' : 'ok'"></span>
          <span class="dq-ac-val" :class="(ratelimitAccounts ?? 0) > 0 ? 'dq-warn' : ''">{{ ratelimitAccounts ?? 0 }}</span>
          <span class="dq-ac-lbl">限流</span>
        </div>
        <div class="dq-ac-num">
          <span class="sdot ok"></span>
          <span class="dq-ac-val dq-ok">{{ normalAccounts ?? 0 }}</span>
          <span class="dq-ac-lbl">正常</span>
        </div>
      </div>
      <div class="dq-ac-total">共 {{ totalAccounts ?? 0 }} 个账号</div>
    </div>

    <!-- 未处理告警 -->
    <div class="dq-alert-card rise" :style="`animation-delay:${(baseDelay ?? 0) + 0.04}s`" role="button" tabindex="0" @click="$emit('nav', '/admin/ops')" @keyup.enter="$emit('nav', '/admin/ops')">
      <div class="dq-ac-head">
        <span class="dq-ac-title">告警事件</span>
        <span class="dq-ac-more">查看 →</span>
      </div>
      <div v-if="alertsLoading" class="dq-ac-spin"><LoadingSpinner size="sm" /></div>
      <template v-else>
        <div class="dq-ac-nums">
          <div class="dq-ac-num">
            <span class="sdot" :class="(firingCount ?? 0) > 0 ? 'bad' : 'ok'"></span>
            <span class="dq-ac-val" :class="(firingCount ?? 0) > 0 ? 'dq-bad' : ''">{{ firingCount ?? 0 }}</span>
            <span class="dq-ac-lbl">告警中</span>
          </div>
          <div class="dq-ac-num">
            <span class="sdot dim"></span>
            <span class="dq-ac-val">{{ resolvedCount }}</span>
            <span class="dq-ac-lbl">已解决</span>
          </div>
        </div>
        <div v-if="latestTitle" class="dq-ac-latest">
          <span class="dq-al-sev" :class="`sev-${latestSeverity}`">{{ latestSeverity }}</span>
          <span class="dq-al-title">{{ latestTitle }}</span>
        </div>
        <div v-else class="dq-ac-ok">无未处理告警</div>
      </template>
    </div>

    <!-- API Keys -->
    <div class="dq-alert-card rise" :style="`animation-delay:${(baseDelay ?? 0) + 0.08}s`" role="button" tabindex="0" @click="$emit('nav', '/admin/users')" @keyup.enter="$emit('nav', '/admin/users')">
      <div class="dq-ac-head">
        <span class="dq-ac-title">API Keys</span>
        <span class="dq-ac-more">查看 →</span>
      </div>
      <div class="dq-ac-nums">
        <div class="dq-ac-num">
          <span class="sdot ok"></span>
          <span class="dq-ac-val dq-ok">{{ activeKeys ?? 0 }}</span>
          <span class="dq-ac-lbl">活跃</span>
        </div>
        <div class="dq-ac-num">
          <span class="sdot dim"></span>
          <span class="dq-ac-val">{{ inactiveKeys }}</span>
          <span class="dq-ac-lbl">非活跃</span>
        </div>
      </div>
      <div class="dq-ac-total">共 {{ totalKeys ?? 0 }} 个密钥</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const props = defineProps<{
  baseDelay?: number
  errorAccounts?: number
  ratelimitAccounts?: number
  normalAccounts?: number
  totalAccounts?: number
  activeKeys?: number
  totalKeys?: number
  alertsLoading?: boolean
  firingCount?: number
  resolvedCount?: number
  latestTitle?: string | null
  latestSeverity?: string | null
}>()

defineEmits<{ nav: [path: string] }>()

const inactiveKeys = computed(() => (props.totalKeys ?? 0) - (props.activeKeys ?? 0))
</script>

<style scoped>
.dq-alert-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; }
@media (max-width: 900px) { .dq-alert-row { grid-template-columns: 1fr 1fr; } }
@media (max-width: 540px) { .dq-alert-row { grid-template-columns: 1fr; } }

.dq-alert-card {
  background: var(--card, #111318); border: 1px solid var(--border, #252830); border-radius: var(--q-radius, 12px);
  padding: 14px 16px; cursor: pointer; transition: border-color .15s;
}
.dq-alert-card:hover { border-color: rgba(92,168,255,.45); }
.dq-alert-card:focus-visible { outline: 2px solid rgba(92,168,255,.6); outline-offset: 2px; }

.dq-ac-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.dq-ac-title { font-size: 11px; font-weight: 600; letter-spacing: .08em; text-transform: uppercase; color: var(--muted-foreground); }
.dq-ac-more { font-size: 11px; color: var(--muted-foreground); opacity: 0; transition: opacity .15s; }
.dq-alert-card:hover .dq-ac-more { opacity: 1; }

.dq-ac-nums { display: flex; gap: 18px; }
.dq-ac-num { display: flex; align-items: center; gap: 6px; }
.dq-ac-val { font-family: var(--font-mono, monospace); font-size: 18px; font-weight: 700; font-variant-numeric: tabular-nums; color: var(--foreground); }
.dq-ac-lbl { font-size: 11px; color: var(--muted-foreground); }
.dq-ac-total { font-size: 11px; color: var(--muted-foreground); margin-top: 8px; }
.dq-ac-ok { font-size: 12px; color: var(--ok); margin-top: 8px; }
.dq-ac-spin { display: flex; justify-content: center; padding: 8px 0; }
.dq-ac-latest { display: flex; align-items: center; gap: 7px; margin-top: 8px; overflow: hidden; }
.dq-al-sev { font-size: 9.5px; font-weight: 700; letter-spacing: .06em; text-transform: uppercase; padding: 1px 6px; border-radius: 5px; flex-shrink: 0; }
.sev-critical { background: color-mix(in srgb, var(--bad) 18%, transparent); color: var(--bad); }
.sev-warning  { background: color-mix(in srgb, var(--warn) 18%, transparent); color: var(--warn); }
.sev-info     { background: color-mix(in srgb, var(--azure) 15%, transparent); color: var(--azure); }
.dq-al-title { font-size: 11.5px; color: var(--foreground); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.dq-ok  { color: var(--ok); }
.dq-bad { color: var(--bad); }
.dq-warn { color: var(--warn); }

.sdot { display: inline-block; width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.sdot.ok   { background: var(--ok); }
.sdot.warn { background: var(--warn); animation: pulse-w 2s infinite; }
.sdot.bad  { background: var(--bad); }
.sdot.dim  { background: var(--muted-foreground); }
@keyframes pulse-w { 0%,100%{ box-shadow:0 0 0 0 rgba(224,179,78,.5);} 50%{ box-shadow:0 0 0 5px rgba(224,179,78,0);} }

.rise { opacity: 0; transform: translateY(10px); animation: rise .5s cubic-bezier(.22,.68,0,1.2) forwards; }
@keyframes rise { to { opacity: 1; transform: none; } }
@media (prefers-reduced-motion: reduce) { .rise { animation: none; opacity: 1; transform: none; } .sdot.warn { animation: none; } }
</style>
