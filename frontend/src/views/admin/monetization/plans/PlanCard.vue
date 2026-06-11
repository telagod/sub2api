<template>
  <div
    class="pcard"
    :class="{
      'pcard--off': !plan.for_sale,
      'pcard--missing': groupMissing,
      'pcard--has-order': plan.sort_order != null,
    }"
  >
    <!-- QUENCH accent bar: azure for platform groups, neutral for none -->
    <div
      class="pcard-accent"
      :class="group ? '' : 'pcard-accent--neutral'"
    />

    <!-- Order tag — top-left, avoids overlapping header controls -->
    <div v-if="plan.sort_order != null" class="pcard-order-tag">
      #{{ plan.sort_order + 1 }}
    </div>

    <!-- Header row: name + status badge | sort arrows -->
    <div class="pcard-hdr">
      <div class="pcard-hdr-left">
        <span class="pcard-name">{{ plan.name }}</span>
        <span class="pcard-sale-badge" :class="plan.for_sale ? 'badge-on' : 'badge-off'">
          {{ plan.for_sale ? t('admin.plansCatalog.onSale') : t('admin.plansCatalog.offSale') }}
        </span>
      </div>
      <div class="pcard-sort">
        <button
          type="button"
          class="pcard-sort-btn"
          :disabled="isFirst"
          @click="emit('move-up')"
          :title="t('admin.plansCatalog.moveUp')"
        >
          <Icon name="chevronUp" size="xs" />
        </button>
        <button
          type="button"
          class="pcard-sort-btn"
          :disabled="isLast"
          @click="emit('move-down')"
          :title="t('admin.plansCatalog.moveDown')"
        >
          <Icon name="chevronDown" size="xs" />
        </button>
      </div>
    </div>

    <!-- Price block -->
    <div class="pcard-price-row">
      <span class="q-money pcard-price">${{ plan.price.toFixed(2) }}</span>
      <span v-if="plan.original_price && plan.original_price > plan.price" class="pcard-orig-price">
        ${{ plan.original_price.toFixed(2) }}
      </span>
      <span class="pcard-period-badge">{{ periodLabel }}</span>
    </div>

    <!-- Description -->
    <p v-if="plan.description" class="pcard-desc">{{ plan.description }}</p>

    <!-- Key config chips -->
    <div class="pcard-chips">
      <template v-if="groupMissing">
        <span class="pcard-chip pcard-chip-bad">{{ t('admin.plansCatalog.groupMissingFmt', { id: plan.group_id }) }}</span>
      </template>
      <template v-else-if="group">
        <GroupBadge
          :name="group.name"
          :platform="group.platform"
          :rate-multiplier="group.rate_multiplier"
          :subscription-type="group.subscription_type"
        />
      </template>

      <span v-if="group?.daily_limit_usd != null" class="pcard-chip">
        {{ t('admin.plansCatalog.dailyLimitFmt', { v: group.daily_limit_usd }) }}
      </span>
      <span v-else-if="group" class="pcard-chip pcard-chip-ok">{{ t('admin.plansCatalog.unlimited') }}</span>
    </div>

    <!-- Features list (top 3 + overflow count) -->
    <ul v-if="plan.features?.length" class="pcard-features">
      <li v-for="(f, i) in plan.features.slice(0, 3)" :key="i" class="pcard-feature">
        <span class="pcard-feature-dot" />
        {{ f }}
      </li>
      <li v-if="plan.features.length > 3" class="pcard-feature pcard-feature-more">
        {{ t('admin.plansCatalog.moreFeaturesFmt', { n: plan.features.length - 3 }) }}
      </li>
    </ul>

    <!-- Footer: on-sale toggle | separator | edit / delete -->
    <div class="pcard-footer">
      <button
        type="button"
        role="switch"
        :aria-checked="plan.for_sale"
        class="pcard-toggle"
        :class="plan.for_sale ? 'pcard-toggle--on' : 'pcard-toggle--off'"
        @click="emit('toggle-sale')"
        :title="plan.for_sale ? t('admin.plansCatalog.toggleOnTitle') : t('admin.plansCatalog.toggleOffTitle')"
      >
        <span class="pcard-toggle-knob" />
      </button>
      <span class="pcard-toggle-lbl">
        {{ plan.for_sale ? t('admin.plansCatalog.onSale') : t('admin.plansCatalog.offSale') }}
      </span>

      <span class="pcard-acts-sep" aria-hidden="true" />

      <div class="pcard-acts">
        <button type="button" class="pcard-act" @click="emit('edit')" :title="t('common.edit')">
          <Icon name="edit" size="sm" />
          <span>{{ t('common.edit') }}</span>
        </button>
        <button type="button" class="pcard-act pcard-act-del" @click="emit('delete')" :title="t('common.delete')">
          <Icon name="trash" size="sm" />
          <span>{{ t('common.delete') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import './plan-card.css'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'

const props = defineProps<{
  plan: SubscriptionPlan
  group?: AdminGroup
  groupMissing?: boolean
  isFirst: boolean
  isLast: boolean
}>()

const emit = defineEmits<{
  'toggle-sale': []
  edit: []
  delete: []
  'move-up': []
  'move-down': []
}>()

const { t } = useI18n()

const periodLabel = computed(() => {
  const unit = props.plan.validity_unit || 'days'
  const n = props.plan.validity_days
  if (unit === 'months') return t('admin.plansCatalog.periodMonths', { n })
  if (unit === 'weeks') return t('admin.plansCatalog.periodWeeks', { n })
  return t('admin.plansCatalog.periodDays', { n })
})
</script>
