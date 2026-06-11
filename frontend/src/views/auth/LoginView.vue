<template>
  <AuthLayout>
    <div class="lv-body">
      <!-- 表单标题 -->
      <div class="lv-head">
        <h2 class="lv-title">{{ t('auth.welcomeBack') }}</h2>
        <p class="lv-sub">{{ t('auth.signInToAccount') }}</p>
      </div>

      <!-- 登录表单 -->
      <form @submit.prevent="handleLogin" class="lv-form">
        <!-- Email -->
        <div class="lv-field">
          <label for="email" class="lv-label">{{ t('auth.emailLabel') }}</label>
          <div class="lv-inp-wrap" :class="{ 'lv-inp-wrap--error': errors.email }">
            <Icon name="mail" size="md" class="lv-inp-icon" />
            <input
              id="email"
              v-model="formData.email"
              type="email"
              required
              autofocus
              autocomplete="email"
              :disabled="authActionDisabled"
              class="lv-inp"
              :placeholder="t('auth.emailPlaceholder')"
            />
          </div>
        </div>

        <!-- Password -->
        <div class="lv-field">
          <label for="password" class="lv-label">{{ t('auth.passwordLabel') }}</label>
          <div class="lv-inp-wrap" :class="{ 'lv-inp-wrap--error': errors.password }">
            <Icon name="lock" size="md" class="lv-inp-icon" />
            <input
              id="password"
              v-model="formData.password"
              :type="showPassword ? 'text' : 'password'"
              required
              autocomplete="current-password"
              :disabled="authActionDisabled"
              class="lv-inp"
              :placeholder="t('auth.passwordPlaceholder')"
            />
            <button
              type="button"
              @click="showPassword = !showPassword"
              :disabled="authActionDisabled"
              class="lv-eye"
              :aria-label="showPassword ? '隐藏密码' : '显示密码'"
            >
              <Icon v-if="showPassword" name="eyeOff" size="md" />
              <Icon v-else name="eye" size="md" />
            </button>
          </div>
          <!-- 忘记密码 -->
          <div class="lv-pwd-row">
            <span></span>
            <router-link
              v-if="passwordResetEnabled && !backendModeEnabled"
              to="/forgot-password"
              class="lv-link"
            >
              {{ t('auth.forgotPassword') }}
            </router-link>
          </div>
        </div>

        <!-- Turnstile -->
        <div v-if="turnstileEnabled && turnstileSiteKey">
          <TurnstileWidget
            ref="turnstileRef"
            :site-key="turnstileSiteKey"
            @verify="onTurnstileVerify"
            @expire="onTurnstileExpire"
            @error="onTurnstileError"
          />
        </div>

        <!-- 主按钮 -->
        <button
          type="submit"
          :disabled="authActionDisabled || (turnstileEnabled && !turnstileToken)"
          class="lv-submit"
        >
          <svg
            v-if="isLoading"
            class="lv-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          <Icon v-else name="login" size="md" />
          {{ isLoading ? t('auth.signingIn') : t('auth.signIn') }}
        </button>

        <!-- 登录协议 -->
        <LoginAgreementPrompt
          v-if="loginAgreementEnabled"
          :accepted="agreementAccepted"
          :documents="loginAgreementDocuments"
          :mode="loginAgreementMode"
          :updated-at="loginAgreementUpdatedAt"
          :visible="showAgreementModal"
          @accept="acceptLoginAgreement"
          @reject="rejectLoginAgreement"
          @open="showAgreementModal = true"
        />

        <!-- OAuth 区 -->
        <div v-if="showOAuthLogin" class="lv-oauth">
          <div class="lv-divider">
            <span class="lv-divider-line"></span>
            <span class="lv-divider-txt">{{ t('auth.oauthOrContinue') }}</span>
            <span class="lv-divider-line"></span>
          </div>

          <EmailOAuthButtons
            :disabled="authActionDisabled"
            :github-enabled="githubOAuthEnabled"
            :google-enabled="googleOAuthEnabled"
            :show-divider="false"
          />
          <LinuxDoOAuthSection
            v-if="linuxdoOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
          />
          <DingTalkOAuthSection
            v-if="dingtalkOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
          />
          <WechatOAuthSection
            v-if="wechatOAuthEnabled"
            :disabled="authActionDisabled"
            :show-divider="false"
          />
          <OidcOAuthSection
            v-if="oidcOAuthEnabled"
            :disabled="authActionDisabled"
            :provider-name="oidcOAuthProviderName"
            :show-divider="false"
          />
        </div>
      </form>
    </div>

    <!-- 页脚 -->
    <template v-if="!backendModeEnabled" #footer>
      <p class="lv-footer-txt">
        {{ t('auth.dontHaveAccount') }}
        <router-link to="/register" class="lv-footer-link">{{ t('auth.signUp') }}</router-link>
      </p>
    </template>
  </AuthLayout>

  <!-- 2FA Modal -->
  <TotpLoginModal
    v-if="show2FAModal"
    ref="totpModalRef"
    :temp-token="totpTempToken"
    :user-email-masked="totpUserEmailMasked"
    @verify="handle2FAVerify"
    @cancel="handle2FACancel"
  />
</template>

<script setup lang="ts">
import { computed, ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import DingTalkOAuthSection from '@/components/auth/DingTalkOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import WechatOAuthSection from '@/components/auth/WechatOAuthSection.vue'
import EmailOAuthButtons from '@/components/auth/EmailOAuthButtons.vue'
import LoginAgreementPrompt from '@/components/auth/LoginAgreementPrompt.vue'
import TotpLoginModal from '@/components/auth/TotpLoginModal.vue'
import Icon from '@/components/icons/Icon.vue'
import TurnstileWidget from '@/components/TurnstileWidget.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { getPublicSettings, isTotp2FARequired, isWeChatWebOAuthEnabled } from '@/api/auth'
import type { LoginAgreementDocument, TotpLoginResponse } from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { clearAllAffiliateReferralCodes } from '@/utils/oauthAffiliate'

const { t } = useI18n()
const LOGIN_AGREEMENT_STORAGE_KEY = 'sub2api_login_agreement_consent'

// ==================== Router & Stores ====================

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

// ==================== State ====================

const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')
const showPassword = ref<boolean>(false)
const publicSettingsLoaded = ref<boolean>(false)

// Public settings
const turnstileEnabled = ref<boolean>(false)
const turnstileSiteKey = ref<string>('')
const linuxdoOAuthEnabled = ref<boolean>(false)
const dingtalkOAuthEnabled = ref<boolean>(false)
const wechatOAuthEnabled = ref<boolean>(false)
const backendModeEnabled = ref<boolean>(false)
const oidcOAuthEnabled = ref<boolean>(false)
const oidcOAuthProviderName = ref<string>('OIDC')
const githubOAuthEnabled = ref<boolean>(false)
const googleOAuthEnabled = ref<boolean>(false)
const passwordResetEnabled = ref<boolean>(false)
const loginAgreementEnabled = ref<boolean>(false)
const loginAgreementMode = ref<'modal' | 'checkbox' | string>('modal')
const loginAgreementUpdatedAt = ref<string>('')
const loginAgreementRevision = ref<string>('')
const loginAgreementDocuments = ref<LoginAgreementDocument[]>([])
const agreementAccepted = ref<boolean>(false)
const showAgreementModal = ref<boolean>(false)

// Turnstile
const turnstileRef = ref<InstanceType<typeof TurnstileWidget> | null>(null)
const turnstileToken = ref<string>('')

// 2FA state
const show2FAModal = ref<boolean>(false)
const totpTempToken = ref<string>('')
const totpUserEmailMasked = ref<string>('')
const totpModalRef = ref<InstanceType<typeof TotpLoginModal> | null>(null)

const formData = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: '',
  turnstile: ''
})

const validationToastMessage = computed(
  () => errors.email || errors.password || errors.turnstile || ''
)

const agreementGateActive = computed(
  () => loginAgreementEnabled.value && !agreementAccepted.value
)

const authActionDisabled = computed(
  () => isLoading.value || !publicSettingsLoaded.value || agreementGateActive.value
)

const showOAuthLogin = computed(
  () =>
    !backendModeEnabled.value &&
    (linuxdoOAuthEnabled.value ||
      dingtalkOAuthEnabled.value ||
      wechatOAuthEnabled.value ||
      oidcOAuthEnabled.value ||
      githubOAuthEnabled.value ||
      googleOAuthEnabled.value)
)

watch(validationToastMessage, (value, previousValue) => {
  if (value && value !== previousValue) {
    appStore.showError(value)
  }
})

// ==================== Lifecycle ====================

onMounted(async () => {
  const expiredFlag = sessionStorage.getItem('auth_expired')
  if (expiredFlag) {
    sessionStorage.removeItem('auth_expired')
    const message = t('auth.reloginRequired')
    errorMessage.value = message
    appStore.showWarning(message)
  }

  try {
    const settings = await getPublicSettings()
    turnstileEnabled.value = settings.turnstile_enabled
    turnstileSiteKey.value = settings.turnstile_site_key || ''
    linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled
    dingtalkOAuthEnabled.value = settings.dingtalk_oauth_enabled ?? false
    wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
    backendModeEnabled.value = settings.backend_mode_enabled
    oidcOAuthEnabled.value = settings.oidc_oauth_enabled
    oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    githubOAuthEnabled.value = settings.github_oauth_enabled
    googleOAuthEnabled.value = settings.google_oauth_enabled
    backendModeEnabled.value = settings.backend_mode_enabled
    passwordResetEnabled.value = settings.password_reset_enabled
    applyLoginAgreementSettings(settings)
  } catch (error) {
    console.error('Failed to load public settings:', error)
    loginAgreementEnabled.value = false
    agreementAccepted.value = true
  } finally {
    publicSettingsLoaded.value = true
  }
})

// ==================== Login Agreement ====================

function applyLoginAgreementSettings(settings: {
  login_agreement_enabled?: boolean
  login_agreement_mode?: string
  login_agreement_updated_at?: string
  login_agreement_revision?: string
  login_agreement_documents?: LoginAgreementDocument[]
}): void {
  const documents = Array.isArray(settings.login_agreement_documents)
    ? settings.login_agreement_documents.filter((doc) => doc.title?.trim())
    : []
  loginAgreementDocuments.value = documents
  loginAgreementEnabled.value = settings.login_agreement_enabled === true && documents.length > 0
  loginAgreementMode.value = settings.login_agreement_mode === 'checkbox' ? 'checkbox' : 'modal'
  loginAgreementUpdatedAt.value = settings.login_agreement_updated_at || ''
  loginAgreementRevision.value =
    settings.login_agreement_revision ||
    `${loginAgreementUpdatedAt.value}:${documents.map((doc) => `${doc.id}:${doc.title}`).join('|')}`

  agreementAccepted.value = !loginAgreementEnabled.value || hasAcceptedLoginAgreement(loginAgreementRevision.value)
  showAgreementModal.value =
    loginAgreementEnabled.value && !agreementAccepted.value && loginAgreementMode.value !== 'checkbox'
}

function hasAcceptedLoginAgreement(revision: string): boolean {
  if (!revision) {
    return false
  }
  try {
    const raw = localStorage.getItem(LOGIN_AGREEMENT_STORAGE_KEY)
    if (!raw) {
      return false
    }
    const parsed = JSON.parse(raw) as { revision?: string }
    return parsed.revision === revision
  } catch {
    return false
  }
}

function acceptLoginAgreement(): void {
  if (loginAgreementRevision.value) {
    localStorage.setItem(
      LOGIN_AGREEMENT_STORAGE_KEY,
      JSON.stringify({
        revision: loginAgreementRevision.value,
        accepted_at: new Date().toISOString()
      })
    )
  }
  agreementAccepted.value = true
  showAgreementModal.value = false
}

function rejectLoginAgreement(): void {
  localStorage.removeItem(LOGIN_AGREEMENT_STORAGE_KEY)
  agreementAccepted.value = false
  showAgreementModal.value = false
  appStore.showWarning('未同意最新条款前，无法输入账号密码或使用快捷登录。')
}

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
  // Reset errors
  errors.email = ''
  errors.password = ''
  errors.turnstile = ''

  let isValid = true

  if (agreementGateActive.value) {
    appStore.showWarning('请先阅读并同意最新条款后再登录。')
    if (loginAgreementMode.value !== 'checkbox') {
      showAgreementModal.value = true
    }
    return false
  }

  // Email validation
  if (!formData.email.trim()) {
    errors.email = t('auth.emailRequired')
    isValid = false
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
    errors.email = t('auth.invalidEmail')
    isValid = false
  }

  // Password validation
  if (!formData.password) {
    errors.password = t('auth.passwordRequired')
    isValid = false
  } else if (formData.password.length < 6) {
    errors.password = t('auth.passwordMinLength')
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

async function handleLogin(): Promise<void> {
  // Clear previous error
  errorMessage.value = ''

  // Validate form
  if (!validateForm()) {
    return
  }

  isLoading.value = true

  try {
    // Call auth store login
    const response = await authStore.login({
      email: formData.email,
      password: formData.password,
      turnstile_token: turnstileEnabled.value ? turnstileToken.value : undefined
    })

    // Check if 2FA is required
    if (isTotp2FARequired(response)) {
      const totpResponse = response as TotpLoginResponse
      totpTempToken.value = totpResponse.temp_token || ''
      totpUserEmailMasked.value = totpResponse.user_email_masked || ''
      show2FAModal.value = true
      isLoading.value = false
      return
    }

    // Show success toast
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    // Reset Turnstile on error
    if (turnstileRef.value) {
      turnstileRef.value.reset()
      turnstileToken.value = ''
    }

    errorMessage.value = extractI18nErrorMessage(error, t, 'auth.errors', t('auth.loginFailed'))

    // Also show error toast
    appStore.showError(errorMessage.value)
  } finally {
    isLoading.value = false
  }
}

// ==================== 2FA Handlers ====================

async function handle2FAVerify(code: string): Promise<void> {
  if (totpModalRef.value) {
    totpModalRef.value.setVerifying(true)
  }

  try {
    await authStore.login2FA(totpTempToken.value, code)

    // Close modal and show success
    show2FAModal.value = false
    clearAllAffiliateReferralCodes()
    appStore.showSuccess(t('auth.loginSuccess'))

    // Redirect to dashboard or intended route
    const redirectTo = (router.currentRoute.value.query.redirect as string) || '/dashboard'
    await router.push(redirectTo)
  } catch (error: unknown) {
    const err = error as { message?: string; response?: { data?: { message?: string } } }
    const message = err.response?.data?.message || err.message || t('profile.totp.loginFailed')

    if (totpModalRef.value) {
      totpModalRef.value.setError(message)
      totpModalRef.value.setVerifying(false)
    }
  }
}

function handle2FACancel(): void {
  show2FAModal.value = false
  totpTempToken.value = ''
  totpUserEmailMasked.value = ''
}
</script>

<style scoped>
/* ── 外层间距 ── */
.lv-body {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* ── 标题区 ── */
.lv-head {
  margin-bottom: 24px;
  text-align: center;
}
.lv-title {
  font-size: 17px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: var(--ink-0);
  margin-bottom: 4px;
}
.lv-sub {
  font-size: 12px;
  color: var(--ink-2);
}

/* ── 表单 ── */
.lv-form {
  display: flex;
  flex-direction: column;
  gap: 0;
}

/* ── 字段 ── */
.lv-field {
  margin-bottom: 18px;
}

.lv-label {
  display: block;
  font-size: 12px;
  color: var(--ink-1);
  margin-bottom: 7px;
}

/* ── 输入包裹（mockup .inp） ── */
.lv-inp-wrap {
  display: flex;
  align-items: center;
  background: #0a0c0f;
  border: 1px solid var(--line-1);
  border-radius: 12px;
  padding: 0 14px;
  height: 46px;
  transition: box-shadow 0.25s ease, border-color 0.25s ease;
  gap: 10px;
}
.lv-inp-wrap:focus-within {
  border-color: rgba(92, 168, 255, 0.75);
  box-shadow: var(--glow-focus);
}
.lv-inp-wrap--error {
  border-color: rgba(242, 92, 105, 0.7);
}
.lv-inp-wrap--error:focus-within {
  border-color: rgba(92, 168, 255, 0.75);
  box-shadow: var(--glow-focus);
}

.lv-inp-icon {
  flex: none;
  color: var(--ink-2);
}

.lv-inp {
  flex: 1;
  background: none;
  border: none;
  outline: none;
  color: var(--ink-0);
  font: inherit;
  font-size: 13.5px;
  min-width: 0;
}
.lv-inp::placeholder {
  color: var(--ink-2);
}
.lv-inp:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 眼睛按钮 */
.lv-eye {
  flex: none;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--ink-2);
  display: flex;
  align-items: center;
  padding: 0;
  transition: color 0.15s ease;
}
.lv-eye:hover {
  color: var(--ink-0);
}
.lv-eye:focus-visible {
  outline: 1.5px solid var(--azure);
  outline-offset: 2px;
  border-radius: 4px;
}
.lv-eye:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* 忘记密码行 */
.lv-pwd-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 6px;
}
.lv-link {
  font-size: 12px;
  color: var(--ink-1);
  text-decoration: none;
  transition: color 0.15s ease;
  border-radius: 4px;
}
.lv-link:hover {
  color: var(--azure);
}
.lv-link:focus-visible {
  outline: 1.5px solid var(--azure);
  outline-offset: 2px;
  box-shadow: var(--glow-focus);
}

/* ── 主按钮（锻面凸面，mockup .btn-metal） ── */
.lv-submit {
  width: 100%;
  height: 46px;
  border-radius: 12px;
  border: 1px solid var(--line-1);
  background: var(--metal-raised);
  color: var(--ink-0);
  font: inherit;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.3em;
  text-indent: 0.3em;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: var(--edge-hi), 0 2px 10px rgba(0, 0, 0, 0.4);
  transition: border-color 0.18s ease, box-shadow 0.18s ease;
  margin-bottom: 16px;
}
.lv-submit:hover:not(:disabled) {
  border-color: rgba(92, 168, 255, 0.55);
  box-shadow: var(--edge-hi), 0 0 16px rgba(92, 168, 255, 0.22), 0 2px 10px rgba(0, 0, 0, 0.4);
}
.lv-submit:focus-visible {
  outline: none;
  border-color: rgba(92, 168, 255, 0.75);
  box-shadow: var(--glow-focus), 0 2px 10px rgba(0, 0, 0, 0.4);
}
.lv-submit:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.lv-submit:active:not(:disabled) {
  transform: scale(0.985);
}

/* loading 转圈 */
.lv-spin {
  width: 16px;
  height: 16px;
  animation: lv-spin 0.8s linear infinite;
}
@keyframes lv-spin {
  to { transform: rotate(360deg); }
}

/* ── OAuth 分隔区 ── */
.lv-oauth {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 4px;
}

.lv-divider {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}
.lv-divider-line {
  flex: 1;
  height: 1px;
  background: var(--line-0);
}
.lv-divider-txt {
  font-size: 11px;
  color: var(--ink-2);
  white-space: nowrap;
}

/* ── 页脚 ── */
.lv-footer-txt {
  color: var(--ink-2);
  font-size: 13px;
}
.lv-footer-link {
  color: var(--ink-1);
  text-decoration: none;
  margin-left: 4px;
  transition: color 0.15s ease;
  border-radius: 4px;
}
.lv-footer-link:hover {
  color: var(--azure);
}
.lv-footer-link:focus-visible {
  outline: 1.5px solid var(--azure);
  outline-offset: 2px;
  box-shadow: var(--glow-focus);
}

/* ── a11y ── */
@media (prefers-reduced-motion: reduce) {
  .lv-spin {
    animation: none;
  }
  .lv-inp-wrap,
  .lv-submit {
    transition: none;
  }
}
</style>
