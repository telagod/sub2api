<template>
  <div class="cms-body">
    <!-- ── Custom Endpoints ─────────────────────────────────────────── -->
    <div class="cms-block">
      <p class="cms-block-hint">{{ t('admin.settings.site.customEndpoints.description') }}</p>

      <div v-if="localEndpoints.length > 0" class="cms-list">
        <div
          v-for="(ep, index) in localEndpoints"
          :key="index"
          class="cms-item-card"
        >
          <div class="cms-item-head">
            <span class="cms-item-label">
              {{ t('admin.settings.site.customEndpoints.itemLabel', { n: index + 1 }) }}
            </span>
            <button
              type="button"
              class="cms-del-btn"
              @click="removeEndpoint(index)"
            >
              <svg class="cms-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
          <div class="cms-grid-2">
            <div>
              <label class="cms-field-label">{{ t('admin.settings.site.customEndpoints.name') }}</label>
              <input
                v-model="ep.name"
                type="text"
                class="cms-input"
                :placeholder="t('admin.settings.site.customEndpoints.namePlaceholder')"
                @input="emitEndpoints"
              />
            </div>
            <div>
              <label class="cms-field-label">{{ t('admin.settings.site.customEndpoints.endpointUrl') }}</label>
              <input
                v-model="ep.endpoint"
                type="url"
                class="cms-input cms-mono"
                :placeholder="t('admin.settings.site.customEndpoints.endpointUrlPlaceholder')"
                @input="emitEndpoints"
              />
            </div>
            <div class="cms-span-2">
              <label class="cms-field-label">{{ t('admin.settings.site.customEndpoints.descriptionLabel') }}</label>
              <input
                v-model="ep.description"
                type="text"
                class="cms-input"
                :placeholder="t('admin.settings.site.customEndpoints.descriptionPlaceholder')"
                @input="emitEndpoints"
              />
            </div>
          </div>
        </div>
      </div>

      <button type="button" class="cms-add-btn" @click="addEndpoint">
        <svg class="cms-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        {{ t('admin.settings.site.customEndpoints.add') }}
      </button>
    </div>

    <!-- ── Custom Menu Items ────────────────────────────────────────── -->
    <div class="cms-block">
      <p class="cms-block-hint">{{ t('admin.settings.customMenu.description') }}</p>

      <div v-if="localMenuItems.length > 0" class="cms-list">
        <div
          v-for="(item, index) in localMenuItems"
          :key="item.id || index"
          class="cms-item-card"
        >
          <div class="cms-item-head">
            <span class="cms-item-label">
              {{ t('admin.settings.customMenu.itemLabel', { n: index + 1 }) }}
            </span>
            <div class="cms-item-actions">
              <!-- Move up -->
              <button
                v-if="index > 0"
                type="button"
                class="cms-order-btn"
                :title="t('admin.settings.customMenu.moveUp')"
                @click="moveMenuItem(index, -1)"
              >
                <svg class="cms-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 15l7-7 7 7" />
                </svg>
              </button>
              <!-- Move down -->
              <button
                v-if="index < localMenuItems.length - 1"
                type="button"
                class="cms-order-btn"
                :title="t('admin.settings.customMenu.moveDown')"
                @click="moveMenuItem(index, 1)"
              >
                <svg class="cms-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              <!-- Delete -->
              <button
                type="button"
                class="cms-del-btn"
                :title="t('admin.settings.customMenu.remove')"
                @click="removeMenuItem(index)"
              >
                <svg class="cms-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </div>

          <div class="cms-grid-2">
            <!-- Label -->
            <div>
              <label class="cms-field-label">{{ t('admin.settings.customMenu.name') }}</label>
              <input
                v-model="item.label"
                type="text"
                class="cms-input"
                :placeholder="t('admin.settings.customMenu.namePlaceholder')"
                @input="emitMenuItems"
              />
            </div>

            <!-- Visibility -->
            <div>
              <label class="cms-field-label">{{ t('admin.settings.customMenu.visibility') }}</label>
              <select v-model="item.visibility" class="cms-input" @change="emitMenuItems">
                <option value="user">{{ t('admin.settings.customMenu.visibilityUser') }}</option>
                <option value="admin">{{ t('admin.settings.customMenu.visibilityAdmin') }}</option>
              </select>
            </div>

            <!-- URL (full width) -->
            <div class="cms-span-2">
              <label class="cms-field-label">{{ t('admin.settings.customMenu.url') }}</label>
              <input
                v-model="item.url"
                type="url"
                class="cms-input cms-mono"
                :placeholder="t('admin.settings.customMenu.urlPlaceholder')"
                @input="emitMenuItems"
              />
            </div>

            <!-- SVG Icon (full width) -->
            <div class="cms-span-2">
              <label class="cms-field-label">{{ t('admin.settings.customMenu.iconSvg') }}</label>
              <ImageUpload
                :model-value="item.icon_svg"
                mode="svg"
                size="sm"
                :upload-label="t('admin.settings.customMenu.uploadSvg')"
                :remove-label="t('admin.settings.customMenu.removeSvg')"
                @update:model-value="(v: string) => { item.icon_svg = v; emitMenuItems() }"
              />
            </div>
          </div>
        </div>
      </div>

      <button type="button" class="cms-add-btn" @click="addMenuItem">
        <svg class="cms-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
        </svg>
        {{ t('admin.settings.customMenu.add') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, defineAsyncComponent } from 'vue'
import { useI18n } from 'vue-i18n'

const ImageUpload = defineAsyncComponent(() => import('@/components/common/ImageUpload.vue'))

const { t } = useI18n()

interface MenuItem {
  id: string
  label: string
  icon_svg: string
  url: string
  visibility: 'user' | 'admin'
  sort_order: number
}

interface CustomEndpoint {
  name: string
  endpoint: string
  description: string
}

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

// Prefer dirty form state over saved settings
const activeSource = computed(() => props.formValues ?? props.settings)

// ── Helpers ────────────────────────────────────────────────────────────────────

function cloneMenuItems(src: Record<string, unknown>): MenuItem[] {
  const raw = src['custom_menu_items']
  if (!Array.isArray(raw)) return []
  return raw.map((item) => ({ ...(item as MenuItem) }))
}

function cloneEndpoints(src: Record<string, unknown>): CustomEndpoint[] {
  const raw = src['custom_endpoints']
  if (!Array.isArray(raw)) return []
  return raw.map((ep) => ({ ...(ep as CustomEndpoint) }))
}

// ── Local state ────────────────────────────────────────────────────────────────

const localMenuItems = ref<MenuItem[]>(cloneMenuItems(activeSource.value))
const localEndpoints = ref<CustomEndpoint[]>(cloneEndpoints(activeSource.value))

// Re-sync when parent resets (discard) or initial load completes
watch(
  () => activeSource.value['custom_menu_items'],
  (incoming) => {
    if (JSON.stringify(incoming) !== JSON.stringify(localMenuItems.value)) {
      localMenuItems.value = cloneMenuItems(activeSource.value)
    }
  },
  { deep: true },
)

watch(
  () => activeSource.value['custom_endpoints'],
  (incoming) => {
    if (JSON.stringify(incoming) !== JSON.stringify(localEndpoints.value)) {
      localEndpoints.value = cloneEndpoints(activeSource.value)
    }
  },
  { deep: true },
)

// ── Emit helpers ───────────────────────────────────────────────────────────────

function emitMenuItems() {
  emit('update:field', 'custom_menu_items', localMenuItems.value.map((item) => ({ ...item })))
}

function emitEndpoints() {
  emit('update:field', 'custom_endpoints', localEndpoints.value.map((ep) => ({ ...ep })))
}

// ── Menu item CRUD + reorder ───────────────────────────────────────────────────

function addMenuItem() {
  localMenuItems.value = [
    ...localMenuItems.value,
    {
      id: '',
      label: '',
      icon_svg: '',
      url: '',
      visibility: 'user',
      sort_order: localMenuItems.value.length,
    },
  ]
  emitMenuItems()
}

function removeMenuItem(index: number) {
  localMenuItems.value = localMenuItems.value
    .filter((_, i) => i !== index)
    .map((item, i) => ({ ...item, sort_order: i }))
  emitMenuItems()
}

function moveMenuItem(index: number, direction: -1 | 1) {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= localMenuItems.value.length) return
  const items = localMenuItems.value.map((item) => ({ ...item }))
  const temp = items[index]
  items[index] = items[targetIndex]
  items[targetIndex] = temp
  items.forEach((item, i) => { item.sort_order = i })
  localMenuItems.value = items
  emitMenuItems()
}

// ── Endpoint CRUD ──────────────────────────────────────────────────────────────

function addEndpoint() {
  localEndpoints.value = [...localEndpoints.value, { name: '', endpoint: '', description: '' }]
  emitEndpoints()
}

function removeEndpoint(index: number) {
  localEndpoints.value = localEndpoints.value.filter((_, i) => i !== index)
  emitEndpoints()
}
</script>

<style scoped>
.cms-body {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.cms-block {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cms-block-hint {
  font-size: 11.5px;
  color: var(--ink-2, #5C6470);
  line-height: 1.55;
  margin: 0;
}

.cms-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

/* Item card */
.cms-item-card {
  border: 1px solid var(--line-0, #20242C);
  border-radius: 10px;
  background: var(--bg-1, #101216);
  overflow: hidden;
}

/* Card head row */
.cms-item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 9px 14px;
  border-bottom: 1px solid var(--line-0, #20242C);
  background: linear-gradient(180deg, rgba(255,255,255,.018) 0%, transparent 100%);
}

.cms-item-label {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--ink-0, #E8EBF0);
}

.cms-item-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

/* Order buttons */
.cms-order-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--ink-2, #5C6470);
  cursor: pointer;
  transition: color .12s, background .12s, border-color .12s;
}
.cms-order-btn:hover {
  color: var(--ink-0, #E8EBF0);
  background: var(--bg-2, #171A20);
  border-color: var(--line-1, #2F3540);
}
.cms-order-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

/* Delete button */
.cms-del-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--bad, #F25C69);
  cursor: pointer;
  transition: color .12s, background .12s, border-color .12s;
}
.cms-del-btn:hover {
  background: rgba(242,92,105,.1);
  border-color: rgba(242,92,105,.25);
}
.cms-del-btn:focus-visible { outline: 2px solid var(--bad, #F25C69); outline-offset: 2px; }

/* Field grid */
.cms-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  padding: 14px;
}
@media (max-width: 600px) { .cms-grid-2 { grid-template-columns: 1fr; } }

.cms-span-2 { grid-column: 1 / -1; }

.cms-field-label {
  display: block;
  font-size: 11.5px;
  font-weight: 500;
  color: var(--ink-2, #5C6470);
  margin-bottom: 4px;
}

.cms-input {
  width: 100%;
  padding: 7px 11px;
  border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540);
  background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color .15s, box-shadow .15s;
  box-sizing: border-box;
}
.cms-input:focus, .cms-input:focus-visible {
  border-color: var(--azure, #5CA8FF);
  box-shadow: 0 0 0 3px rgba(92,168,255,.14);
}

.cms-mono { font-family: var(--font-mono, "IBM Plex Mono", monospace); font-size: 12px; }

/* Add button */
.cms-add-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 9px 14px;
  border-radius: 8px;
  border: 2px dashed var(--line-1, #2F3540);
  background: transparent;
  color: var(--ink-2, #5C6470);
  font-size: 12.5px;
  font-weight: 500;
  font-family: inherit;
  cursor: pointer;
  transition: border-color .15s, color .15s;
}
.cms-add-btn:hover {
  border-color: var(--azure, #5CA8FF);
  color: var(--azure, #5CA8FF);
}
.cms-add-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

.cms-icon { width: 14px; height: 14px; flex-shrink: 0; }

/* divider between the two blocks */
.cms-block + .cms-block {
  border-top: 1px solid var(--line-0, #20242C);
  padding-top: 20px;
}
</style>
