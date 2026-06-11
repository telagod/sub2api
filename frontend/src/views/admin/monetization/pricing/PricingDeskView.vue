<template>
  <AppLayout>
    <div class="pd-root">
      <!-- 页头 -->
      <div class="pd-head rise">
        <div>
          <h1 class="pd-title">{{ t('admin.pricingDesk.title') }}</h1>
          <p class="pd-desc">{{ t('admin.pricingDesk.desc') }}</p>
        </div>
        <div class="pd-actions">
          <!-- 刷新 -->
          <button
            class="pd-btn"
            :disabled="loading"
            @click="fetchAll"
          >
            <RefreshCwIcon class="pd-btn-ico" :class="loading ? 'pd-spinning' : ''" />
            {{ t('admin.pricingDesk.refresh') }}
          </button>

          <!-- 价格模拟器 -->
          <button
            class="pd-btn pd-btn-primary"
            @click="simulatorVisible = true"
          >
            <CalculatorIcon class="pd-btn-ico pd-ico-azure" />
            {{ t('admin.pricingDesk.simulatorBtn') }}
          </button>
        </div>
      </div>

      <!-- 同步成功 toast -->
      <Transition name="pd-toast">
        <div v-if="syncToast" class="pd-toast rise">
          {{ t('admin.pricingDesk.syncSuccess', { n: syncToast }) }}
        </div>
      </Transition>

      <!-- 错误提示 -->
      <div v-if="error" class="pd-error rise">
        {{ t('admin.pricingDesk.loadFailed') }}{{ error }}
      </div>

      <!-- 矩阵表格 -->
      <MatrixTable
        :loading="loading"
        :platforms="platforms"
        :active-groups="activeGroups"
        :matrix="matrix"
        :official-pricing-cache="officialPricingCache as Record<string, OfficialPricing | 'loading'>"
        :sync-loading="syncLoading"
        @hover-model="ensureOfficialPricing"
        @update-multiplier="handleUpdateMultiplier"
        @sync-catalog="handleSyncCatalog"
        @open-detail="handleOpenDetail"
      />

      <!-- 价格模拟器抽屉 -->
      <PriceSimulator
        :visible="simulatorVisible"
        :platforms="platforms"
        :matrix="matrix"
        :active-groups="activeGroups"
        :official-pricing-cache="officialPricingCache as Record<string, OfficialPricing | 'loading'>"
        @close="simulatorVisible = false"
        @need-official-pricing="ensureOfficialPricing"
      />

      <!-- 供应商核对抽屉 -->
      <ProviderVerifyDrawer
        :open="detailDrawerVisible"
        :slug="detailSlug"
        :model-name="detailModel"
        @close="detailDrawerVisible = false"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { RefreshCwIcon, CalculatorIcon } from 'lucide-vue-next'
import AppLayout from '@/components/layout/AppLayout.vue'
import MatrixTable from './MatrixTable.vue'
import PriceSimulator from './PriceSimulator.vue'
import ProviderVerifyDrawer from './ProviderVerifyDrawer.vue'
import { usePricingMatrix } from './usePricingMatrix'
import type { OfficialPricing } from './usePricingMatrix'

const { t } = useI18n()
const {
  loading,
  error,
  matrix,
  platforms,
  activeGroups,
  officialPricingCache,
  fetchAll,
  ensureOfficialPricing,
  updateGroupMultiplier,
  syncCatalog
} = usePricingMatrix()

const simulatorVisible = ref(false)

// 供应商核对抽屉状态
const detailDrawerVisible = ref(false)
const detailSlug = ref('')
const detailModel = ref('')

// 同步目录状态
const syncLoading = ref(false)
const syncToast = ref<number | null>(null)
let syncToastTimer: ReturnType<typeof setTimeout> | null = null

onMounted(() => fetchAll())

async function handleUpdateMultiplier(groupId: number, value: number) {
  try {
    await updateGroupMultiplier(groupId, value)
  } catch (e) {
    console.error('更新倍率失败', e)
  }
}

async function handleSyncCatalog() {
  if (syncLoading.value) return
  syncLoading.value = true
  try {
    const result = await syncCatalog()
    // 展示 toast
    syncToast.value = result.synced
    if (syncToastTimer) clearTimeout(syncToastTimer)
    syncToastTimer = setTimeout(() => { syncToast.value = null }, 3500)
  } catch (e) {
    console.error('同步目录失败', e)
  } finally {
    syncLoading.value = false
  }
}

function handleOpenDetail(payload: { slug: string; model: string }) {
  detailSlug.value = payload.slug
  detailModel.value = payload.model
  detailDrawerVisible.value = true
}
</script>

<style scoped>
.pd-root { display: flex; flex-direction: column; gap: 14px; }

.rise { opacity: 0; transform: translateY(8px); animation: rise .45s cubic-bezier(.22,.68,0,1.2) forwards; }
@keyframes rise { to { opacity: 1; transform: none; } }
@media (prefers-reduced-motion: reduce) { .rise { animation: none; opacity: 1; transform: none; } .pd-spinning { animation: none; } }

/* ── 页头 ── */
.pd-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
.pd-title { font-size: 21px; font-weight: 700; letter-spacing: .01em; color: var(--ink-0); margin: 0; }
.pd-desc { font-size: 12px; color: var(--ink-2); margin: 4px 0 0; }

.pd-actions { display: flex; align-items: center; gap: 8px; }

.pd-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 7px 15px; border-radius: 10px;
  font-size: 12.5px; font-weight: 600;
  background: var(--bg-2);
  border: 1px solid var(--line-1); color: var(--ink-1);
  box-shadow: var(--edge-hi, inset 0 1px 0 rgba(255,255,255,.06));
  cursor: pointer; transition: border-color .18s, box-shadow .18s, color .18s;
}
.pd-btn:hover:not(:disabled) { border-color: var(--line-0); color: var(--ink-0); }
.pd-btn:disabled { opacity: .5; cursor: default; }
.pd-btn:focus-visible { outline: none; box-shadow: var(--glow-focus); }

.pd-btn-primary {
  background: var(--metal-raised, linear-gradient(180deg,#272D37,#14171D));
  color: var(--ink-0);
  box-shadow: var(--edge-hi, inset 0 1px 0 rgba(255,255,255,.06)), 0 2px 8px rgba(0,0,0,.3);
}
.pd-btn-primary:hover:not(:disabled) { border-color: rgba(92,168,255,.45); box-shadow: var(--edge-hi), 0 0 12px rgba(92,168,255,.18); }

.pd-btn-ico { width: 14px; height: 14px; flex-shrink: 0; }
.pd-ico-azure { color: var(--azure); }
.pd-spinning { animation: pd-spin 1s linear infinite; }
@keyframes pd-spin { to { transform: rotate(360deg); } }

/* ── 错误条 ── */
.pd-error {
  padding: 10px 14px; border-radius: 10px; font-size: 12.5px;
  background: var(--bad-dim); border: 1px solid var(--bad); color: var(--bad);
}

/* ── 同步成功 toast ── */
.pd-toast {
  padding: 9px 14px; border-radius: 10px; font-size: 12.5px; font-weight: 600;
  background: var(--ok-dim, rgba(52,199,89,.12)); border: 1px solid var(--ok, #34c759);
  color: var(--ok, #34c759);
}
.pd-toast-enter-active { transition: opacity .22s, transform .22s; }
.pd-toast-leave-active { transition: opacity .35s; }
.pd-toast-enter-from { opacity: 0; transform: translateY(-6px); }
.pd-toast-leave-to { opacity: 0; }
</style>
