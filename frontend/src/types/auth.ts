// Basic User and Auth types as per task requirements
export interface User {
  id: string
  email: string
  name: string
  role: 'admin' | 'editor' | 'viewer'
  organization_id: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface LoginRequest {
  email: string
  password: string
}

export interface SignupRequest {
  email: string
  password: string
  name: string
  organization_name?: string
}

export interface MFAVerificationRequest {
  code: string
  session_id: string
}

export interface PasswordResetRequest {
  email: string
}

export interface PasswordResetConfirmRequest {
  token: string
  password: string
}

// Extended auth types for existing code compatibility
export type AuthMethod = 'local' | 'ldap' | 'saml' | 'oauth'
export type MFAMethod = 'totp' | 'sms' | 'backup_code'
export type OAuthProvider = 'google' | 'azure_ad' | 'github' | 'custom_oidc'

// Local login request
export interface LocalLoginRequest {
  username: string
  password: string
  mfa_code?: string
}

// LDAP login request
export interface LDAPLoginRequest {
  username: string
  password: string
  ldap_server?: string
  mfa_code?: string
}

// SAML responses
export interface SAMLInitiateResponse {
  redirect_url: string
  state: string
}

// OAuth responses
export interface OAuthInitiateResponse {
  authorization_url: string
  state: string
}

export interface OAuthCallbackRequest {
  code: string
  state: string
  provider: OAuthProvider
}

// MFA Setup/Verification
export interface MFASetupRequest {
  method: MFAMethod
  delivery_method?: 'email' | 'sms'
}

export interface MFASetupResponse {
  setup_id: string
  method: MFAMethod
  secret?: string
  qr_code?: string
  backup_codes?: string[]
}

export interface MFAVerificationResponse {
  verified: true
  backup_codes?: string[]
}

export interface MFAChallengeRequest {
  username: string
  mfa_method: MFAMethod
}

export interface MFAChallengeResponse {
  challenge_id: string
  mfa_method: MFAMethod
  expires_at: number
}

// Session type
export interface Session {
  access_token: string
  refresh_token?: string
  expires_at: number
  user: User
}

// Auth Context Type for app state
export interface AuthContextType {
  user: User | null
  session: Session | null
  isLoading: boolean
  error: string | null
  isAuthenticated: boolean
  authMethod: AuthMethod | null

  loginLocal: (credentials: LocalLoginRequest) => Promise<void>
  loginLDAP: (credentials: LDAPLoginRequest) => Promise<void>
  initiateSAML: () => Promise<string>
  handleSAMLCallback: (code: string, state: string) => Promise<void>
  initiateOAuth: (provider: OAuthProvider) => Promise<string>
  handleOAuthCallback: (request: OAuthCallbackRequest) => Promise<void>

  setupMFA: (method: MFAMethod) => Promise<MFASetupResponse>
  verifyMFA: (request: MFAVerificationRequest) => Promise<void>
  completeMFASetup: (request: MFAVerificationRequest) => Promise<void>
  disableMFA: (method: MFAMethod) => Promise<void>

  logout: () => Promise<void>
  refreshSession: () => Promise<void>
  validateSession: () => Promise<boolean>

  clearError: () => void
}

// Login Page State
export interface LoginPageState {
  selectedMethod: AuthMethod
  showMFAChallenge: boolean
  mfaChallengeId?: string
  mfaMethod?: MFAMethod
  isLoading: boolean
  error?: string
}
