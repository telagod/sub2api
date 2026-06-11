<template>
  <div :id="`sr-section-${section.id}`" class="sr-section-card">
    <!-- card header -->
    <div class="sr-section-header">
      <h2 class="sr-section-title">{{ resolveLabel(section.title) }}</h2>
      <p v-if="section.description" class="sr-section-desc">{{ resolveLabel(section.description) }}</p>
    </div>

    <!-- custom component escape hatch -->
    <component
      v-if="section.component"
      :is="section.component"
      :settings="settings"
      :form-values="form"
      @update:field="onFieldUpdate"
    />

    <!-- field grid -->
    <div v-else class="sr-section-body">
      <FieldRenderer
        v-for="field in section.fields"
        :key="field.key"
        :field="field"
        :model-value="form[field.key]"
        :form-values="form"
        class="sr-field-item"
        @update:model-value="onFieldUpdate(field.key, $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import FieldRenderer from './FieldRenderer.vue'
import type { SettingsSection } from './types'

defineProps<{
  section: SettingsSection
  form: Record<string, unknown>
  settings: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

const { t } = useI18n()

function resolveLabel(key: string): string {
  try {
    const result = t(key)
    return result === key ? key : result
  } catch {
    return key
  }
}

function onFieldUpdate(key: string, value: unknown) {
  emit('update:field', key, value)
}
</script>

<style scoped>
/* QUENCH metal surface — same surface grammar as DashboardQuenchView */
.sr-section-card {
  background: var(--metal, linear-gradient(180deg,#15181E,#0E1014));
  border: 1px solid var(--line-0, #20242C);
  border-radius: var(--q-radius, 12px);
  overflow: hidden;
  scroll-margin-top: 6rem;
  box-shadow: var(--edge-hi, inset 0 1px 0 rgba(255,255,255,.04)), 0 4px 16px rgba(0,0,0,.25);
}
.sr-section-header {
  padding: 14px 20px;
  border-bottom: 1px solid var(--line-0, #20242C);
  /* top-lit edge gives the header a slight raised feel against the card body */
  background: linear-gradient(180deg, rgba(255,255,255,.025) 0%, transparent 100%);
}
.sr-section-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--ink-0, #E8EBF0);
  margin: 0 0 2px;
  letter-spacing: -.15px;
}
.sr-section-desc {
  font-size: 11.5px;
  color: var(--ink-2, #5C6470);
  margin: 0;
  line-height: 1.55;
}
.sr-section-body {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  padding: 20px;
}
.sr-field-item {
  /* individual field cells — can span full width if needed via a future :class binding */
}
</style>
