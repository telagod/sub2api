<template>
  <!-- 空态 -->
  <div
    v-if="!loading && platforms.length === 0"
    class="flex flex-col items-center justify-center gap-4 rounded-xl border border-dashed py-20"
    :style="{ borderColor: 'var(--line-1)', background: 'var(--bg-1)' }"
  >
    <div class="text-5xl opacity-30">💰</div>
    <p class="text-base" :style="{ color: 'var(--ink-1)' }">{{ t('admin.pricingDesk.noData') }}</p>
    <p class="text-sm" :style="{ color: 'var(--ink-2)' }">{{ t('admin.pricingDesk.noDataHint') }}</p>
    <RouterLink
      to="/admin/channels/pricing"
      class="mt-2 inline-flex items-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium transition-colors"
      :style="{ background: 'var(--metal-raised)', color: 'var(--ink-0)', border: '1px solid var(--line-1)', boxShadow: 'var(--edge-hi)' }"
    >
      <CalculatorIcon class="h-4 w-4" />
      {{ t('admin.pricingDesk.goConfigBtn') }}
    </RouterLink>
  </div>

  <!-- 矩阵表 -->
  <div v-else class="overflow-x-auto rounded-xl" :style="{ background: 'var(--bg-1)', border: '1px solid var(--line-0)' }">
    <table class="w-full text-sm border-collapse">
      <!-- 表头 -->
      <thead>
        <tr :style="{ background: 'var(--bg-2)', borderBottom: '1px solid var(--line-0)' }">
          <th
            class="sticky left-0 z-10 px-4 py-3 text-left font-medium"
            :style="{ background: 'var(--bg-2)', color: 'var(--ink-1)', minWidth: '200px' }"
          >
            {{ t('admin.pricingDesk.colModel') }}
          </th>
          <th
            v-for="group in activeGroups"
            :key="group.id"
            class="px-3 py-3 text-center font-medium whitespace-nowrap"
            :style="{ color: 'var(--ink-1)', minWidth: '140px' }"
          >
            <div class="flex flex-col items-center gap-1">
              <span>{{ group.name }}</span>
              <!-- 列头倍率就地编辑 -->
              <div
                class="flex items-center gap-1"
                :style="{ color: 'var(--azure)' }"
              >
                <span v-if="editingGroupId !== group.id" class="text-xs cursor-pointer hover:underline" @dblclick="startEditMultiplier(group)">
                  ×{{ group.rate_multiplier.toFixed(2) }}
                </span>
                <template v-else>
                  <input
                    :ref="(el) => { if (group.id === editingGroupId) { multiplierInputRef = el as HTMLInputElement } }"
                    v-model.number="editingMultiplierValue"
                    type="number"
                    step="0.01"
                    min="0"
                    class="w-16 rounded px-1 py-0.5 text-xs text-center"
                    :style="{
                      background: 'var(--bg-0)',
                      border: '1px solid var(--azure)',
                      color: 'var(--ink-0)',
                      outline: 'none'
                    }"
                    @keydown.enter="commitMultiplier(group.id)"
                    @keydown.esc="cancelEditMultiplier"
                    @blur="commitMultiplier(group.id)"
                  />
                </template>
              </div>
            </div>
          </th>
        </tr>
      </thead>

      <!-- 按 platform 分组折叠 -->
      <tbody>
        <template v-for="platform in platforms" :key="platform">
          <!-- platform 行组标题 -->
          <tr
            class="cursor-pointer select-none"
            :style="{ background: 'var(--bg-2)', borderTop: '1px solid var(--line-0)' }"
            @click="togglePlatform(platform)"
          >
            <td
              :colspan="activeGroups.length + 1"
              class="px-4 py-2"
              :style="{ color: 'var(--ink-1)' }"
            >
              <div class="flex items-center gap-2">
                <ChevronDownIcon
                  class="h-4 w-4 transition-transform"
                  :class="collapsedPlatforms.has(platform) ? '-rotate-90' : ''"
                />
                <span class="text-xs font-semibold uppercase tracking-wider">{{ platform }}</span>
                <span class="text-xs" :style="{ color: 'var(--ink-2)' }">
                  {{ t('admin.pricingDesk.modelCount', { n: rowsByPlatform[platform]?.length ?? 0 }) }}
                </span>
              </div>
            </td>
          </tr>

          <!-- 模型行 -->
          <template v-if="!collapsedPlatforms.has(platform)">
            <tr
              v-for="row in rowsByPlatform[platform]"
              :key="row.model"
              class="group border-b transition-colors"
              :style="{ borderColor: 'var(--line-0)' }"
              @mouseenter="onRowHover(row.model)"
            >
              <!-- 模型名列 -->
              <td
                class="sticky left-0 z-10 px-4 py-2 font-mono text-xs"
                :style="{ background: 'var(--bg-1)', color: 'var(--ink-0)' }"
              >
                {{ row.model }}
              </td>

              <!-- 每个分组的单元格 -->
              <td
                v-for="group in activeGroups"
                :key="group.id"
                class="px-3 py-2 text-center align-top"
              >
                <MatrixCell
                  v-if="row.cells[group.id]"
                  :cell="row.cells[group.id]"
                  :model="row.model"
                  :official-pricing="officialPricingCache[row.model]"
                />
                <span v-else class="text-xs" :style="{ color: 'var(--ink-2)' }">—</span>
              </td>
            </tr>
          </template>
        </template>
      </tbody>
    </table>

    <!-- 加载骨架 -->
    <div v-if="loading" class="space-y-2 p-4">
      <div
        v-for="i in 6"
        :key="i"
        class="h-10 animate-pulse rounded"
        :style="{ background: 'var(--bg-2)' }"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import { ChevronDownIcon, CalculatorIcon } from 'lucide-vue-next'
import MatrixCell from './MatrixCell.vue'
import type { MatrixRow, OfficialPricing } from './usePricingMatrix'
import type { AdminGroup } from '@/types'

const props = defineProps<{
  loading: boolean
  platforms: string[]
  activeGroups: AdminGroup[]
  matrix: MatrixRow[]
  officialPricingCache: Record<string, OfficialPricing | 'loading'>
}>()

const emit = defineEmits<{
  (e: 'hover-model', model: string): void
  (e: 'update-multiplier', groupId: number, value: number): void
}>()

const { t } = useI18n()

// platform 折叠状态
const collapsedPlatforms = ref(new Set<string>())
function togglePlatform(p: string) {
  if (collapsedPlatforms.value.has(p)) {
    collapsedPlatforms.value.delete(p)
  } else {
    collapsedPlatforms.value.add(p)
  }
}

const rowsByPlatform = computed(() => {
  const map: Record<string, MatrixRow[]> = {}
  for (const row of props.matrix) {
    if (!map[row.platform]) map[row.platform] = []
    map[row.platform].push(row)
  }
  return map
})

// hover 触发官方价懒加载
function onRowHover(model: string) {
  emit('hover-model', model)
}

// 倍率就地编辑
// Major fix: multiplierInputRef cannot be a template ref inside v-for (Vue 3 collects into array).
// Use a plain variable and a function ref binding (:ref="...") in the template instead.
const editingGroupId = ref<number | null>(null)
const editingMultiplierValue = ref(1)
const multiplierInputRef = ref<HTMLInputElement | null>(null)

function startEditMultiplier(group: AdminGroup) {
  editingGroupId.value = group.id
  editingMultiplierValue.value = group.rate_multiplier
  nextTick(() => multiplierInputRef.value?.select())
}

function cancelEditMultiplier() {
  editingGroupId.value = null
}

async function commitMultiplier(groupId: number) {
  if (editingGroupId.value !== groupId) return
  editingGroupId.value = null
  const v = Number(editingMultiplierValue.value)
  if (!isNaN(v) && v > 0) {
    emit('update-multiplier', groupId, v)
  }
}
</script>
