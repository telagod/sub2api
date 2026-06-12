/**
 * Security tab — Generic OIDC Connect section.
 *
 * Keys covered (all form bindings from SettingsView.vue oidc block):
 *   oidc_connect_enabled
 *   oidc_connect_provider_name
 *   oidc_connect_client_id
 *   oidc_connect_client_secret              (masked / configured sentinel)
 *   oidc_connect_client_secret_configured
 *   oidc_connect_issuer_url
 *   oidc_connect_discovery_url
 *   oidc_connect_authorize_url
 *   oidc_connect_token_url
 *   oidc_connect_userinfo_url
 *   oidc_connect_jwks_url
 *   oidc_connect_scopes
 *   oidc_connect_redirect_url               (+ quick-set/copy button)
 *   oidc_connect_frontend_redirect_url
 *   oidc_connect_token_auth_method          (select: client_secret_post | client_secret_basic | none)
 *   oidc_connect_use_pkce
 *   oidc_connect_validate_id_token
 *   oidc_connect_allowed_signing_algs
 *   oidc_connect_clock_skew_seconds
 *   oidc_connect_require_email_verified
 *   oidc_connect_userinfo_email_path
 *   oidc_connect_userinfo_id_path
 *   oidc_connect_userinfo_username_path
 *
 * The section uses OidcConnectSection.vue as its custom component because of:
 *   - Masked client_secret with configured-sentinel logic
 *   - Quick-set/copy button for redirect URL (derives suggestion from window.location)
 *   - Multi-column grid layout (lg:grid-cols-2 and lg:grid-cols-3 rows)
 *   - Three toggle-card booleans (PKCE, validate_id_token, require_email_verified)
 * All state flows through emit('update:field'); fields[] is empty.
 */
import { defineAsyncComponent } from 'vue'
import type { SettingsSection } from '../types'

const oidcConnect: SettingsSection = {
  id: 'security.oidc',
  tab: 'security',
  title: 'admin.settings.oidc.title',
  description: 'admin.settings.oidc.description',
  fields: [],
  component: defineAsyncComponent(
    () => import('../special/OidcConnectSection.vue'),
  ),
}

export default oidcConnect
