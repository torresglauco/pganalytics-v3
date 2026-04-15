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

  // ✅ UPDATED: Add CSRF token for state-changing requests
  if (['POST', 'PUT', 'DELETE', 'PATCH'].includes(method.toUpperCase())) {
    const csrfToken = getCsrfTokenFromCookie();
    if (csrfToken) {
      defaultHeaders['X-CSRF-Token'] = csrfToken;
    }
  }

  // ❌ REMOVED: No longer read from localStorage
  // The auth_token is now sent automatically via cookies (credentials: 'include')

  const response = await fetch(url, {
    method,
    headers: { ...defaultHeaders, ...headers },
    body: body ? JSON.stringify(body) : undefined,
    credentials: 'include',  // ✅ This sends cookies automatically
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Unknown error' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
};

// ✅ NEW: Helper to extract CSRF token from cookie
const getCsrfTokenFromCookie = (): string | null => {
  const name = 'csrf_token=';
  const decodedCookie = decodeURIComponent(document.cookie);
  const cookieArray = decodedCookie.split(';');

  for (let cookie of cookieArray) {
    cookie = cookie.trim();
    if (cookie.indexOf(name) === 0) {
      return cookie.substring(name.length, cookie.length);
    }
  }
  return null;
};

// Auth API calls
export const authApi = {
  // Local authentication
  loginLocal: async (credentials: LocalLoginRequest): Promise<Session> => {
    // ✅ UPDATED: Response now includes user and csrf_token (token is in httpOnly cookie)
    const response = await apiCall<{ message: string; user: User; csrf_token: string; expires_at: string }>(
      'POST',
      '/auth/login',
      credentials
    );

    // Return session-like object
    return {
      user: response.user,
      token: '', // Token is in httpOnly cookie, not returned here
      csrfToken: response.csrf_token,
      expiresAt: new Date(response.expires_at),
    } as Session;
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
    // ✅ UPDATED: Response now includes user and csrf_token
    const response = await apiCall<{ message: string; user: User; csrf_token: string; expires_at: string }>(
      'POST',
      '/auth/refresh'
    );

    return {
      user: response.user,
      token: '', // Token is in httpOnly cookie, not returned here
      csrfToken: response.csrf_token,
      expiresAt: new Date(response.expires_at),
    } as Session;
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
