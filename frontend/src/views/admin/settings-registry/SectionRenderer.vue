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
.sr-section-card {
  background: var(--bg-1, #101216);
  border: 1px solid var(--line-0, #20242C);
  border-radius: 12px;
  overflow: hidden;
  scroll-margin-top: 6rem;
}
.sr-section-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--line-0, #20242C);
  background: linear-gradient(180deg, var(--bg-1, #101216), var(--bg-0, #0C0E12));
}
.sr-section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ink-0, #E8EBF0);
  margin: 0 0 2px;
  letter-spacing: -.2px;
}
.sr-section-desc {
  font-size: 12px;
  color: var(--ink-2, #5C6470);
  margin: 0;
  line-height: 1.5;
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
