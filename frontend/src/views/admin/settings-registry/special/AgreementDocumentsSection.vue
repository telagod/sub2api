<template>
  <div class="agr-body">
    <!-- header row -->
    <div class="agr-header">
      <p class="agr-hint">{{ t('admin.settings.agreement.docsHint') }}</p>
      <button type="button" class="agr-add-btn" @click="addDocument">
        <Icon name="plus" size="sm" class="agr-btn-icon" />
        {{ t('admin.settings.agreement.addDoc') }}
      </button>
    </div>

    <!-- document cards -->
    <div class="agr-list">
      <div
        v-for="(doc, index) in localDocs"
        :key="doc.id || index"
        class="agr-doc-card"
      >
        <!-- card header -->
        <div class="agr-doc-head">
          <div class="agr-doc-identity">
            <span class="agr-doc-icon">
              <Icon :name="index === 1 ? 'shield' : index === 2 ? 'globe' : index === 3 ? 'cog' : 'document'" size="sm" />
            </span>
            <div class="agr-doc-meta">
              <p class="agr-doc-title">{{ doc.title || t('admin.settings.agreement.unnamedDoc') }}</p>
              <p class="agr-doc-path">/legal/{{ doc.id || '…' }}</p>
            </div>
          </div>
          <button
            type="button"
            class="agr-del-btn"
            :disabled="agreementEnabled && localDocs.length <= 1"
            @click="removeDocument(index)"
          >
            <Icon name="trash" size="sm" />
          </button>
        </div>

        <!-- fields grid -->
        <div class="agr-fields">
          <div>
            <label class="agr-field-label">{{ t('admin.settings.agreement.docTitle') }}</label>
            <input
              v-model="doc.title"
              type="text"
              class="agr-input"
              :placeholder="t('admin.settings.agreement.docTitlePlaceholder')"
              @input="emitUpdate"
            />
          </div>
          <div>
            <label class="agr-field-label">{{ t('admin.settings.agreement.docSlug') }}</label>
            <div class="agr-slug-wrap">
              <span class="agr-slug-prefix">/legal/</span>
              <input
                v-model="doc.id"
                type="text"
                class="agr-slug-input"
                placeholder="usage-policy"
                @input="emitUpdate"
              />
            </div>
          </div>
        </div>

        <!-- content textarea -->
        <div class="agr-content-wrap">
          <label class="agr-field-label">{{ t('admin.settings.agreement.docContent') }}</label>
          <textarea
            v-model="doc.content_md"
            rows="8"
            class="agr-textarea"
            :placeholder="t('admin.settings.agreement.docContentPlaceholder')"
            @input="emitUpdate"
          />
        </div>
      </div>
    </div>

    <p v-if="localDocs.length === 0" class="agr-empty">{{ t('admin.settings.agreement.noDocs') }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

interface AgreementDoc {
  id: string
  title: string
  content_md: string
}

const props = defineProps<{
  settings: Record<string, unknown>
  formValues?: Record<string, unknown>
}>()

const emit = defineEmits<{
  'update:field': [key: string, value: unknown]
}>()

// ── helpers ────────────────────────────────────────────────────────────────────
/** Prefer formValues (current dirty state) over savedSettings */
const activeSource = computed(() => props.formValues ?? props.settings)

function cloneDocs(src: Record<string, unknown>): AgreementDoc[] {
  const raw = src['login_agreement_documents']
  if (!Array.isArray(raw)) return []
  return raw.map((d) => ({ ...(d as AgreementDoc) }))
}

// ── local state — single source of truth for this component ───────────────────
const localDocs = ref<AgreementDoc[]>(cloneDocs(activeSource.value))

// Re-sync when parent resets (discard) or initial load completes
watch(
  () => activeSource.value['login_agreement_documents'],
  (incoming) => {
    if (JSON.stringify(incoming) !== JSON.stringify(localDocs.value)) {
      localDocs.value = cloneDocs(activeSource.value)
    }
  },
  { deep: true },
)

const agreementEnabled = computed(() => !!activeSource.value['login_agreement_enabled'])

function emitUpdate() {
  emit('update:field', 'login_agreement_documents', localDocs.value.map((d) => ({ ...d })))
}

function addDocument() {
  localDocs.value = [...localDocs.value, { id: '', title: '', content_md: '' }]
  emitUpdate()
}

function removeDocument(index: number) {
  localDocs.value = localDocs.value.filter((_, i) => i !== index)
  emitUpdate()
}
</script>

<style scoped>
.agr-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 16px; }

/* header row */
.agr-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; }
.agr-hint { font-size: 11.5px; color: var(--ink-2, #5C6470); line-height: 1.5; margin: 0; }

/* add button — metal raised */
.agr-add-btn {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 6px 14px; border-radius: 8px;
  background: var(--metal-raised, linear-gradient(180deg,#272D37,#14171D));
  border: 1px solid rgba(255,255,255,.1); color: var(--ink-0, #E8EBF0);
  font-size: 12.5px; font-weight: 500; font-family: inherit;
  cursor: pointer; box-shadow: inset 0 1px 0 rgba(255,255,255,.06);
  transition: border-color .15s, box-shadow .15s;
}
.agr-add-btn:hover { border-color: rgba(92,168,255,.4); box-shadow: inset 0 1px 0 rgba(255,255,255,.06), 0 0 10px rgba(92,168,255,.14); }
.agr-add-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }
.agr-btn-icon { width: 14px; height: 14px; }

/* document list */
.agr-list { display: flex; flex-direction: column; gap: 12px; }

/* document card — nested metal surface */
.agr-doc-card {
  border: 1px solid var(--line-0, #20242C); border-radius: 10px;
  background: var(--bg-1, #101216);
  overflow: hidden;
}

/* card head */
.agr-doc-head {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--line-0, #20242C);
  background: linear-gradient(180deg, rgba(255,255,255,.018) 0%, transparent 100%);
}
.agr-doc-identity { display: flex; align-items: center; gap: 10px; min-width: 0; }
.agr-doc-icon {
  flex-shrink: 0; width: 34px; height: 34px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-2, #171A20); border: 1px solid var(--line-0, #20242C);
  color: var(--ink-1, #97A0AF);
}
.agr-doc-meta { min-width: 0; }
.agr-doc-title { font-size: 13px; font-weight: 600; color: var(--ink-0, #E8EBF0); margin: 0 0 1px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.agr-doc-path  { font-size: 11px; color: var(--ink-2, #5C6470); margin: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* delete button */
.agr-del-btn {
  flex-shrink: 0; display: inline-flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: 6px;
  border: 1px solid transparent; background: transparent;
  color: var(--ink-2, #5C6470); cursor: pointer;
  transition: color .12s, background .12s, border-color .12s;
}
.agr-del-btn:hover:not(:disabled) { color: var(--bad, #F25C69); background: rgba(242,92,105,.1); border-color: rgba(242,92,105,.25); }
.agr-del-btn:disabled { opacity: .35; cursor: not-allowed; }
.agr-del-btn:focus-visible { outline: 2px solid var(--azure, #5CA8FF); outline-offset: 2px; }

/* fields grid */
.agr-fields { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; padding: 14px; }
@media (max-width: 600px) { .agr-fields { grid-template-columns: 1fr; } }

/* content wrap */
.agr-content-wrap { padding: 0 14px 14px; display: flex; flex-direction: column; gap: 4px; }

/* labels */
.agr-field-label { display: block; font-size: 11.5px; font-weight: 500; color: var(--ink-2, #5C6470); margin-bottom: 4px; }

/* shared input style */
.agr-input, .agr-textarea {
  width: 100%; padding: 7px 11px; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
  transition: border-color .15s, box-shadow .15s; box-sizing: border-box;
}
.agr-input:focus, .agr-input:focus-visible,
.agr-textarea:focus, .agr-textarea:focus-visible {
  border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14);
}
.agr-textarea { font-family: var(--font-mono, "IBM Plex Mono", monospace); font-size: 12px; resize: vertical; }

/* slug compound input */
.agr-slug-wrap {
  display: flex; overflow: hidden; border-radius: 8px;
  border: 1px solid var(--line-1, #2F3540); background: var(--bg-0, #0C0E12);
  transition: border-color .15s, box-shadow .15s;
}
.agr-slug-wrap:focus-within { border-color: var(--azure, #5CA8FF); box-shadow: 0 0 0 3px rgba(92,168,255,.14); }
.agr-slug-prefix {
  flex-shrink: 0; display: inline-flex; align-items: center;
  padding: 0 10px; border-right: 1px solid var(--line-1, #2F3540);
  background: var(--bg-2, #171A20); color: var(--ink-2, #5C6470);
  font-size: 12.5px; white-space: nowrap;
}
.agr-slug-input {
  flex: 1; min-width: 0; border: none; background: transparent;
  padding: 7px 11px; color: var(--ink-0, #E8EBF0); font-size: 13px; font-family: inherit; outline: none;
}

/* empty state */
.agr-empty { font-size: 13px; color: var(--ink-2, #5C6470); margin: 0; }
</style>
