<template>
  <AuthLayout>
    <div class="fv-body">
      <!-- 标题 -->
      <div class="fv-head">
        <h2 class="fv-title">{{ t('auth.forgotPasswordTitle') }}</h2>
        <p class="fv-sub">{{ t('auth.forgotPasswordHint') }}</p>
      </div>

      <!-- 成功状态 -->
      <div v-if="isSubmitted" class="fv-success">
        <div class="fv-success-icon-wrap">
          <Icon name="checkCircle" size="lg" class="fv-icon-ok" />
        </div>
        <div class="fv-success-text">
          <h3 class="fv-success-title">{{ t('auth.resetEmailSent') }}</h3>
          <p class="fv-success-sub">{{ t('auth.resetEmailSentHint') }}</p>
        </div>
        <router-link to="/login" class="fv-back-link">
          <Icon name="arrowLeft" size="sm" />
          {{ t('auth.backToLogin') }}
        </router-link>
      </div>

      <!-- 表单状态 -->
      <form v-else @submit.prevent="handleSubmit" class="fv-form">
        <div class="fv-field">
          <label for="email" class="fv-label">{{ t('auth.emailLabel') }}</label>
          <div class="fv-inp-wrap" :class="{ 'fv-inp-wrap--error': errors.email }">
            <Icon name="mail" size="md" class="fv-inp-icon" />
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="isLoading"
              class="fv-inp"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
        </div>

        <div v-if="turnstileEnabled && turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <button
          type="submit"
          :disabled="isLoading || (turnstileEnabled && !turnstileToken)"
          class="fv-submit"
        >
          <svg v-if="isLoading" class="fv-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <Icon v-else name="mail" size="md" />
          {{ isLoading ? t('auth.sendingResetLink') : t('auth.sendResetLink') }}
        </button>
      </form>
    </div>

    <template #footer>
      <p class="fv-footer-txt">
        {{ t('auth.rememberedPassword') }}
        <router-link to="/login" class="fv-footer-link">{{ t('auth.signIn') }}</router-link>
      </p>
    </template>
  </AuthLayout>
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAppStore } from '@/stores'
import { getPublicSettings, forgotPassword } from '@/api/auth'

const { t } = useI18n()

// ==================== Stores ====================

const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const isSubmitted = ref<boolean>(false)
const errorMessage = ref<string>('')

// Public settings
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

const formData = reactive({
  email: ''
})

const errors = reactive({
  email: '',
  turnstile: ''
})

const validationToastMessage = computed(() => errors.email || errors.turnstile || '')

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// ==================== Lifecycle ====================

onMounted(async () => {
  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})

// ==================== Turnstile Handlers ====================

function onTurnstileVerify(token: string): void {
  turnstileToken.value = token
  errors.turnstile = ''
}

function onTurnstileExpire(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileExpired')
}

function onTurnstileError(): void {
  turnstileToken.value = ''
  errors.turnstile = t('auth.turnstileFailed')
}

// ==================== Validation ====================

function validateForm(): boolean {
  errors.email = ''
  errors.turnstile = ''

  let isValid = true

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  // Turnstile validation
  if (turnstileEnabled.value && !turnstileToken.value) {
    errors.turnstile = t('auth.completeVerification')
    isValid = false
  }

  return isValid
}

// ==================== Form Handlers ====================

async function handleSubmit(): Promise<void> {
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    await forgotPassword({
      email: formData.email,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    isSubmitted.value = true
    appStore.showSuccess(t('auth.resetEmailSent'))
  } catch (error: unknown) {
    // Reset Turnstile on error
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    const err = error as { message?: string; response?: { data?: { detail?: string } } }

    if (err.response?.data?.detail) {
      errorMessage.value = err.response.data.detail
    } else if (err.message) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = t('auth.sendResetLinkFailed')
    }

    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
.fv-body { display: flex; flex-direction: column; gap: 0; }

.fv-head { margin-bottom: 24px; text-align: center; }
.fv-title { font-size: 17px; font-weight: 700; letter-spacing: 0.04em; color: var(--ink-0); margin-bottom: 4px; }
.fv-sub { font-size: 12px; color: var(--ink-2); }

/* 成功状态 */
.fv-success {
  display: flex; flex-direction: column; align-items: center; gap: 16px;
  border-radius: 12px; border: 1px solid rgba(70,201,140,.25);
  background: rgba(70,201,140,.07); padding: 28px 20px; text-align: center;
}
.fv-success-icon-wrap {
  width: 48px; height: 48px; border-radius: 50%;
  border: 1px solid rgba(70,201,140,.3);
  background: rgba(70,201,140,.12);
  display: grid; place-items: center;
}
.fv-icon-ok { color: #46C98C; }
.fv-success-title { font-size: 15px; font-weight: 600; color: var(--ink-0); margin-bottom: 6px; }
.fv-success-sub { font-size: 12px; color: var(--ink-2); }
.fv-back-link {
  display: inline-flex; align-items: center; gap: 6px;
  font-size: 13px; font-weight: 500; color: var(--ink-1); text-decoration: none;
  transition: color 0.15s ease;
}
.fv-back-link:hover { color: var(--azure); }
.fv-back-link:focus-visible { outline: 1.5px solid var(--azure); outline-offset: 2px; border-radius: 4px; }

/* 表单 */
.fv-form { display: flex; flex-direction: column; gap: 0; }
.fv-field { margin-bottom: 18px; }
.fv-label { display: block; font-size: 12px; color: var(--ink-1); margin-bottom: 7px; }

.fv-inp-wrap {
  display: flex; align-items: center;
  background: #0a0c0f; border: 1px solid var(--line-1);
  border-radius: 12px; padding: 0 14px; height: 46px;
  transition: box-shadow 0.25s ease, border-color 0.25s ease; gap: 10px;
}
.fv-inp-wrap:focus-within { border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus); }
.fv-inp-wrap--error { border-color: rgba(242,92,105,.6); }
.fv-inp-wrap--error:focus-within { border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus); }

.fv-inp-icon { flex: none; color: var(--ink-2); }
.fv-inp {
  flex: 1; min-width: 0; background: none; border: none; outline: none;
  color: var(--ink-0); font: inherit; font-size: 13.5px;
}
.fv-inp::placeholder { color: var(--ink-2); }
.fv-inp:disabled { opacity: 0.5; cursor: not-allowed; }

/* 按钮 */
.fv-submit {
  width: 100%; height: 46px; border-radius: 12px;
  border: 1px solid #3a4250; background: var(--metal-raised);
  color: var(--ink-0); font: inherit; font-size: 14px; font-weight: 600;
  letter-spacing: 0.2em; cursor: pointer;
  display: inline-flex; align-items: center; justify-content: center; gap: 8px;
  box-shadow: var(--edge-hi), 0 2px 10px rgba(0,0,0,.4);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
  margin-bottom: 0;
}
.fv-submit:hover:not(:disabled) {
  border-color: rgba(92,168,255,.55);
  box-shadow: var(--edge-hi), 0 0 16px rgba(92,168,255,.22), 0 2px 10px rgba(0,0,0,.4);
}
.fv-submit:focus-visible { outline: none; border-color: rgba(92,168,255,.75); box-shadow: var(--glow-focus), 0 2px 10px rgba(0,0,0,.4); }
.fv-submit:disabled { opacity: 0.45; cursor: not-allowed; }
.fv-submit:active:not(:disabled) { transform: scale(0.985); }

.fv-spin { width: 16px; height: 16px; animation: fv-spin 0.8s linear infinite; flex: none; }
@keyframes fv-spin { to { transform: rotate(360deg); } }

/* 页脚 */
.fv-footer-txt { color: var(--ink-2); font-size: 13px; }
.fv-footer-link { color: var(--ink-1); text-decoration: none; margin-left: 4px; transition: color 0.15s ease; }
.fv-footer-link:hover { color: var(--azure); }

@media (prefers-reduced-motion: reduce) {
  .fv-spin { animation: none; }
  .fv-inp-wrap, .fv-submit { transition: none; }
}
</style>
