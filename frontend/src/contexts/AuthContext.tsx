import React, { createContext, useCallback, useEffect, useState } from 'react';
import authApi from '../api/authApi';
import type {
  AuthContextType,
  AuthMethod,
  LocalLoginRequest,
  LDAPLoginRequest,
  MFASetupRequest,
  MFASetupResponse,
  MFAVerificationRequest,
  OAuthCallbackRequest,
  OAuthProvider,
  Session,
  User,
} from '../types/auth';

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: React.ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [authMethod, setAuthMethod] = useState<AuthMethod | null>(null);

  // Initialize auth from stored session
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        const token = localStorage.getItem('access_token');
        if (!token) {
          setIsLoading(false);
          return;
        }

        // Validate stored session with backend
        const userData = await authApi.validateSession();
        setUser(userData);
        setAuthMethod(userData.auth_method);

        // Session is valid, keep the stored token
        setIsLoading(false);
      } catch (err) {
        // Clear invalid session
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        setUser(null);
        setSession(null);
        setIsLoading(false);
      }
    };

    initializeAuth();
  }, []);

  // Store session in state and localStorage
  const storeSession = useCallback((newSession: Session) => {
    setSession(newSession);
    setUser(newSession.user);
    setAuthMethod(newSession.user.auth_method);
    localStorage.setItem('access_token', newSession.access_token);
    if (newSession.refresh_token) {
      localStorage.setItem('refresh_token', newSession.refresh_token);
    }
  }, []);

  // Local login
  const loginLocal = useCallback(async (credentials: LocalLoginRequest) => {
    try {
      setError(null);
      setIsLoading(true);
      const newSession = await authApi.loginLocal(credentials);
      storeSession(newSession);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Login failed';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [storeSession]);

  // LDAP login
  const loginLDAP = useCallback(async (credentials: LDAPLoginRequest) => {
    try {
      setError(null);
      setIsLoading(true);
      const newSession = await authApi.loginLDAP(credentials);
      storeSession(newSession);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'LDAP login failed';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [storeSession]);

  // SAML
  const initiateSAML = useCallback(async (): Promise<string> => {
    try {
      setError(null);
      const response = await authApi.initiateSAML();
      return response.redirect_url;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'SAML initiation failed';
      setError(message);
      throw err;
    }
  }, []);

  const handleSAMLCallback = useCallback(async (code: string, state: string) => {
    try {
      setError(null);
      setIsLoading(true);
      const newSession = await authApi.handleSAMLCallback(code, state);
      storeSession(newSession);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'SAML callback failed';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [storeSession]);

  // OAuth
  const initiateOAuth = useCallback(async (provider: OAuthProvider): Promise<string> => {
    try {
      setError(null);
      const response = await authApi.initiateOAuth(provider);
      return response.authorization_url;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'OAuth initiation failed';
      setError(message);
      throw err;
    }
  }, []);

  const handleOAuthCallback = useCallback(async (request: OAuthCallbackRequest) => {
    try {
      setError(null);
      setIsLoading(true);
      const newSession = await authApi.handleOAuthCallback(request);
      storeSession(newSession);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'OAuth callback failed';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [storeSession]);

  // MFA Setup
  const setupMFA = useCallback(async (method: MFASetupRequest['method']): Promise<MFASetupResponse> => {
    try {
      setError(null);
      return await authApi.setupMFA({ method });
    } catch (err) {
      const message = err instanceof Error ? err.message : 'MFA setup failed';
      setError(message);
      throw err;
    }
  }, []);

  // MFA Verification (for setup)
  const verifyMFA = useCallback(async (request: MFAVerificationRequest) => {
    try {
      setError(null);
      await authApi.verifyMFASetup(request);
      // Update user MFA status
      if (user) {
        const updatedUser = {
          ...user,
          mfa_enabled: true,
          mfa_methods: [...user.mfa_methods, request.method],
        };
        setUser(updatedUser);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'MFA verification failed';
      setError(message);
      throw err;
    }
  }, [user]);

  // Complete MFA Setup (includes verification)
  const completeMFASetup = useCallback(async (request: MFAVerificationRequest) => {
    try {
      setError(null);
      await authApi.verifyMFASetup(request);
      // Update user to reflect MFA is now enabled
      if (user) {
        const updatedUser = {
          ...user,
          mfa_enabled: true,
          mfa_methods: [...new Set([...user.mfa_methods, request.method])],
        };
        setUser(updatedUser);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'MFA setup completion failed';
      setError(message);
      throw err;
    }
  }, [user]);

  // Disable MFA
  const disableMFA = useCallback(async (method: string) => {
    try {
      setError(null);
      await authApi.disableMFA(method);
      // Update user MFA status
      if (user) {
        const updatedUser = {
          ...user,
          mfa_methods: user.mfa_methods.filter(m => m !== method),
          mfa_enabled: user.mfa_methods.length > 1,
        };
        setUser(updatedUser);
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'MFA disable failed';
      setError(message);
      throw err;
    }
  }, [user]);

  // Logout
  const logout = useCallback(async () => {
    try {
      setError(null);
      setIsLoading(true);
      await authApi.logout();
    } catch (err) {
      // Still clear session even if logout API fails
      console.error('Logout error:', err);
    } finally {
      setUser(null);
      setSession(null);
      setAuthMethod(null);
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      setIsLoading(false);
    }
  }, []);

  // Refresh Session
  const refreshSession = useCallback(async () => {
    try {
      setError(null);
      const newSession = await authApi.refreshSession();
      storeSession(newSession);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Session refresh failed';
      setError(message);
      throw err;
    }
  }, [storeSession]);

  // Validate Session
  const validateSession = useCallback(async (): Promise<boolean> => {
    try {
      const userData = await authApi.validateSession();
      setUser(userData);
      return true;
    } catch (err) {
      return false;
    }
  }, []);

  // Clear Error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const value: AuthContextType = {
    user,
    session,
    isLoading,
    error,
    isAuthenticated: !!user,
    authMethod,
    loginLocal,
    loginLDAP,
    initiateSAML,
    handleSAMLCallback,
    initiateOAuth,
    handleOAuthCallback,
    setupMFA,
    verifyMFA,
    completeMFASetup,
    disableMFA,
    logout,
    refreshSession,
    validateSession,
    clearError,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContext;
