// Authentication types and interfaces

export type AuthMethod = 'local' | 'ldap' | 'saml' | 'oauth';
export type MFAMethod = 'totp' | 'sms' | 'backup_code';
export type OAuthProvider = 'google' | 'azure_ad' | 'github' | 'custom_oidc';

// User and Session
export interface User {
  id: string;
  username: string;
  email: string;
  full_name?: string;
  avatar_url?: string;
  auth_method: AuthMethod;
  mfa_enabled: boolean;
  mfa_methods: MFAMethod[];
  created_at: string;
  updated_at: string;
  last_login?: string;
}

export interface Session {
  access_token: string;
  refresh_token?: string;
  expires_at: number;
  user: User;
}

// Login/Auth Requests
export interface LocalLoginRequest {
  username: string;
  password: string;
  mfa_code?: string;
}

export interface LDAPLoginRequest {
  username: string;
  password: string;
  ldap_server?: string;
  mfa_code?: string;
}

export interface SAMLInitiateResponse {
  redirect_url: string;
  state: string;
}

export interface OAuthInitiateResponse {
  authorization_url: string;
  state: string;
}

export interface OAuthCallbackRequest {
  code: string;
  state: string;
  provider: OAuthProvider;
}

// MFA Setup/Verification
export interface MFASetupRequest {
  method: MFAMethod;
  delivery_method?: 'email' | 'sms'; // for SMS/Email
}

export interface MFASetupResponse {
  setup_id: string;
  method: MFAMethod;
  secret?: string; // for TOTP QR code generation
  qr_code?: string; // base64 encoded QR code
  backup_codes?: string[]; // for display after setup
}

export interface MFAVerificationRequest {
  setup_id?: string;
  code: string;
  method: MFAMethod;
  remember_device?: boolean;
}

export interface MFAVerificationResponse {
  verified: true;
  backup_codes?: string[]; // returned after successful setup
}

export interface MFAChallengeRequest {
  username: string;
  mfa_method: MFAMethod;
}

export interface MFAChallengeResponse {
  challenge_id: string;
  mfa_method: MFAMethod;
  expires_at: number;
}

// Auth Context
export interface AuthContextType {
  user: User | null;
  session: Session | null;
  isLoading: boolean;
  error: string | null;
  isAuthenticated: boolean;
  authMethod: AuthMethod | null;

  // Auth methods
  loginLocal: (credentials: LocalLoginRequest) => Promise<void>;
  loginLDAP: (credentials: LDAPLoginRequest) => Promise<void>;
  initiateSAML: () => Promise<string>; // returns redirect URL
  handleSAMLCallback: (code: string, state: string) => Promise<void>;
  initiateOAuth: (provider: OAuthProvider) => Promise<string>; // returns redirect URL
  handleOAuthCallback: (request: OAuthCallbackRequest) => Promise<void>;

  // MFA methods
  setupMFA: (method: MFAMethod) => Promise<MFASetupResponse>;
  verifyMFA: (request: MFAVerificationRequest) => Promise<void>;
  completeMFASetup: (request: MFAVerificationRequest) => Promise<void>;
  disableMFA: (method: MFAMethod) => Promise<void>;

  // Session methods
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
  validateSession: () => Promise<boolean>;

  // Utility
  clearError: () => void;
}

// Login Page State
export interface LoginPageState {
  selectedMethod: AuthMethod;
  showMFAChallenge: boolean;
  mfaChallengeId?: string;
  mfaMethod?: MFAMethod;
  isLoading: boolean;
  error?: string;
}
