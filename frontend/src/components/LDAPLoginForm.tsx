import React, { useState } from 'react';
import { Eye, EyeOff, AlertCircle } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

interface LDAPLoginFormProps {
  onSuccess?: () => void;
  ldapServerUrl?: string;
}

export const LDAPLoginForm: React.FC<LDAPLoginFormProps> = ({
  onSuccess,
  ldapServerUrl = 'ldap://directory.company.com',
}) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [mfaCode, setMfaCode] = useState('');
  const [showMFA, setShowMFA] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  const { loginLDAP, error: contextError, clearError } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null);

    if (!username || !password) {
      setLocalError('Username and password are required');
      return;
    }

    try {
      setIsLoading(true);
      await loginLDAP({
        username,
        password,
        ldap_server: ldapServerUrl,
        mfa_code: mfaCode || undefined,
      });
      onSuccess?.();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'LDAP login failed';
      if (errorMessage.includes('MFA') || errorMessage.includes('2FA')) {
        setShowMFA(true);
        setPassword(''); // Clear password for security
      }
    } finally {
      setIsLoading(false);
    }
  };

  const _clearErrors = () => {
    setLocalError(null);
    clearError();
  };

  const displayError = localError || contextError;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h3 className="text-xl font-semibold text-gray-900 mb-6">Sign in with LDAP</h3>

      {/* Server Info */}
      <div className="p-3 bg-blue-50 border border-blue-200 rounded-lg">
        <p className="text-sm text-blue-900">
          <strong>LDAP Server:</strong> {ldapServerUrl}
        </p>
      </div>

      {/* Error Message */}
      {displayError && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex gap-2">
          <AlertCircle size={16} className="mt-0.5 flex-shrink-0" />
          <div>{displayError}</div>
        </div>
      )}

      {/* Username */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Username
        </label>
        <input
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          disabled={isLoading}
          placeholder="e.g., john.doe or user@company.com"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
          autoFocus
        />
        <p className="mt-1 text-xs text-gray-600">
          Enter your LDAP username or email address
        </p>
      </div>

      {/* Password */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Password
        </label>
        <div className="relative">
          <input
            type={showPassword ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={isLoading}
            placeholder="Enter your LDAP password"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-2.5 text-gray-600 hover:text-gray-700"
            tabIndex={-1}
          >
            {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
          </button>
        </div>
      </div>

      {/* MFA Code (if needed) */}
      {showMFA && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            MFA Code
          </label>
          <input
            type="text"
            value={mfaCode}
            onChange={(e) => setMfaCode(e.target.value)}
            disabled={isLoading}
            placeholder="Enter your 6-digit code"
            maxLength={6}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50 text-center text-2xl tracking-widest"
            autoFocus
          />
          <p className="mt-2 text-xs text-gray-600">
            Your LDAP account has MFA enabled. Enter the code from your authenticator.
          </p>
        </div>
      )}

      {/* Submit Button */}
      <button
        type="submit"
        disabled={isLoading}
        className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg transition mt-6"
      >
        {isLoading ? (
          <span className="flex items-center justify-center gap-2">
            <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            Signing in...
          </span>
        ) : (
          'Sign In'
        )}
      </button>

      {/* Help Text */}
      <div className="text-center mt-4 text-xs text-gray-600">
        <p>Problems signing in?</p>
        <p>
          Contact <a href="mailto:support@company.com" className="text-blue-600 hover:text-blue-700">support</a>
        </p>
      </div>
    </form>
  );
};

export default LDAPLoginForm;
