import type {
  LocalLoginRequest,
  LDAPLoginRequest,
  SAMLInitiateResponse,
  OAuthInitiateResponse,
  OAuthCallbackRequest,
  MFASetupRequest,
  MFASetupResponse,
  MFAVerificationRequest,
  MFAVerificationResponse,
  MFAChallengeRequest,
  MFAChallengeResponse,
  Session,
  User,
} from '../types/auth';

const API_BASE = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

// Helper for API calls
const apiCall = async <T,>(
  method: string,
  endpoint: string,
  body?: unknown,
  headers?: Record<string, string>
): Promise<T> => {
  const url = `${API_BASE}${endpoint}`;
  const defaultHeaders: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  const token = localStorage.getItem('access_token');
  if (token) {
    defaultHeaders['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, {
    method,
    headers: { ...defaultHeaders, ...headers },
    body: body ? JSON.stringify(body) : undefined,
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
};

// Auth API calls
export const authApi = {
  // Local authentication
  loginLocal: async (credentials: LocalLoginRequest): Promise<Session> => {
    const response = await apiCall<{ session: Session }>(
      'POST',
      '/auth/login',
      credentials
    );
    return response.session;
  },

  // LDAP authentication
  loginLDAP: async (credentials: LDAPLoginRequest): Promise<Session> => {
    const response = await apiCall<{ session: Session }>(
      'POST',
      '/auth/ldap/login',
      credentials
    );
    return response.session;
  },

  // SAML
  initiateSAML: async (): Promise<SAMLInitiateResponse> => {
    return apiCall<SAMLInitiateResponse>(
      'GET',
      '/auth/saml/initiate'
    );
  },

  handleSAMLCallback: async (code: string, state: string): Promise<Session> => {
    const response = await apiCall<{ session: Session }>(
      'POST',
      '/auth/saml/callback',
      { code, state }
    );
    return response.session;
  },

  // OAuth
  initiateOAuth: async (provider: string): Promise<OAuthInitiateResponse> => {
    return apiCall<OAuthInitiateResponse>(
      'GET',
      `/auth/oauth/${provider}/initiate`
    );
  },

  handleOAuthCallback: async (request: OAuthCallbackRequest): Promise<Session> => {
    const response = await apiCall<{ session: Session }>(
      'POST',
      `/auth/oauth/callback`,
      request
    );
    return response.session;
  },

  // MFA Setup
  setupMFA: async (request: MFASetupRequest): Promise<MFASetupResponse> => {
    return apiCall<MFASetupResponse>(
      'POST',
      '/auth/mfa/setup',
      request
    );
  },

  // MFA Verification (during setup)
  verifyMFASetup: async (request: MFAVerificationRequest): Promise<MFAVerificationResponse> => {
    return apiCall<MFAVerificationResponse>(
      'POST',
      '/auth/mfa/verify-setup',
      request
    );
  },

  // MFA Challenge (during login)
  requestMFAChallenge: async (request: MFAChallengeRequest): Promise<MFAChallengeResponse> => {
    return apiCall<MFAChallengeResponse>(
      'POST',
      '/auth/mfa/challenge',
      request
    );
  },

  // MFA Challenge Verification (during login)
  verifyMFAChallenge: async (challengeId: string, code: string): Promise<Session> => {
    const response = await apiCall<{ session: Session }>(
      'POST',
      `/auth/mfa/challenge/${challengeId}/verify`,
      { code }
    );
    return response.session;
  },

  // MFA Disable
  disableMFA: async (method: string): Promise<void> => {
    await apiCall<void>(
      'DELETE',
      `/auth/mfa/${method}`
    );
  },

  // Session
  refreshSession: async (): Promise<Session> => {
    const response = await apiCall<{ session: Session }>(
      'POST',
      '/auth/refresh'
    );
    return response.session;
  },

  validateSession: async (): Promise<User> => {
    return apiCall<User>(
      'GET',
      '/auth/me'
    );
  },

  logout: async (): Promise<void> => {
    await apiCall<void>(
      'POST',
      '/auth/logout'
    );
  },

  // Get authentication methods available
  getAuthMethods: async (): Promise<{ methods: string[] }> => {
    return apiCall<{ methods: string[] }>(
      'GET',
      '/auth/methods'
    );
  },
};

export default authApi;
