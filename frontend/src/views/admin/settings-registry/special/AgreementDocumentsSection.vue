<template>
  <div class="p-5 space-y-4">
    <!-- Document list -->
    <div class="flex items-center justify-between">
      <p class="text-xs text-muted-foreground">
        {{ localText(
          '可参考：服务条款、使用政策、支持的国家和地区、服务特定条款。',
          'Example documents: Terms of Service, Usage Policy, Supported Regions, Service Terms.'
        ) }}
      </p>
      <button
        type="button"
        class="btn btn-primary btn-sm inline-flex items-center gap-1.5"
        @click="addDocument"
      >
        <Icon name="plus" size="sm" />
        {{ localText('添加文档', 'Add document') }}
      </button>
    </div>

    <div class="space-y-3">
      <div
        v-for="(doc, index) in localDocs"
        :key="doc.id || index"
        class="rounded-lg border border-border bg-card p-4"
      >
        <div class="mb-3 flex items-center justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <span class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-md bg-accent text-foreground/85">
              <Icon
                :name="index === 1 ? 'shield' : index === 2 ? 'globe' : index === 3 ? 'cog' : 'document'"
                size="sm"
              />
            </span>
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold text-foreground">
                {{ doc.title || localText('未命名文档', 'Untitled document') }}
              </p>
              <p class="truncate text-xs text-muted-foreground">/legal/{{ doc.id || '…' }}</p>
            </div>
          </div>
          <button
            type="button"
            class="rounded-md p-2 text-red-400 transition hover:bg-red-500/10 hover:text-red-300 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="agreementEnabled && localDocs.length <= 1"
            @click="removeDocument(index)"
          >
            <Icon name="trash" size="sm" />
          </button>
        </div>

        <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
          <div>
            <label class="mb-1 block text-xs font-medium text-muted-foreground">
              {{ localText('文档名称', 'Document title') }}
            </label>
            <input
              v-model="doc.title"
              type="text"
              class="input text-sm"
              :placeholder="localText('例如：服务条款', 'Example: Terms of Service')"
              @input="emitUpdate"
            />
          </div>
          <div>
            <label class="mb-1 block text-xs font-medium text-muted-foreground">
              {{ localText('路由标识', 'Route slug') }}
            </label>
            <div class="flex overflow-hidden rounded-lg border border-border bg-card focus-within:border-ring focus-within:ring-1 focus-within:ring-ring">
              <span class="inline-flex flex-shrink-0 items-center border-r border-border bg-muted px-3 text-sm text-muted-foreground">
                /legal/
              </span>
              <input
                v-model="doc.id"
                type="text"
                class="min-w-0 flex-1 border-0 bg-transparent px-3 py-2 text-sm text-foreground outline-none placeholder:text-muted-foreground focus:ring-0"
                placeholder="usage-policy"
                @input="emitUpdate"
              />
            </div>
          </div>
        </div>
        <div class="mt-3">
          <label class="mb-1 block text-xs font-medium text-muted-foreground">
            {{ localText('Markdown 内容', 'Markdown content') }}
          </label>
          <textarea
            v-model="doc.content_md"
            rows="8"
            class="input font-mono text-sm"
            :placeholder="localText('在这里填写正式 Markdown 内容。', 'Write the final Markdown content here.')"
            @input="emitUpdate"
          ></textarea>
        </div>
      </div>
    </div>

    <p v-if="localDocs.length === 0" class="text-sm text-muted-foreground">
      {{ localText('暂无协议文档，点击「添加文档」创建第一个。', 'No documents yet. Click "Add document" to create one.') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Icon from '@/components/icons/Icon.vue'

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

function localText(zh: string, _en: string): string {
  return zh
}

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
