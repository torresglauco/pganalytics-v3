import React, { useState } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

interface LocalLoginFormProps {
  onSuccess?: () => void;
}

export const LocalLoginForm: React.FC<LocalLoginFormProps> = ({ onSuccess }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [mfaCode, setMfaCode] = useState('');
  const [showMFA, setShowMFA] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  const { loginLocal, error: contextError, clearError } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError(null);

    if (!username || !password) {
      setLocalError('Username and password are required');
      return;
    }

    try {
      setIsLoading(true);
      await loginLocal({
        username,
        password,
        mfa_code: mfaCode || undefined,
      });
      onSuccess?.();
    } catch (err) {
      // Error is set in context, but check if MFA is needed
      const errorMessage = err instanceof Error ? err.message : 'Login failed';
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
      <h3 className="text-xl font-semibold text-gray-900 mb-6">Sign in with your credentials</h3>

      {/* Error Message */}
      {displayError && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
          {displayError}
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
          placeholder="Enter your username"
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
          autoFocus
        />
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
            placeholder="Enter your password"
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
            Enter the code from your authenticator app or SMS
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

      {/* Forgot Password Link */}
      <div className="text-center mt-4">
        <a href="/forgot-password" className="text-sm text-blue-600 hover:text-blue-700">
          Forgot your password?
        </a>
      </div>
    </form>
  );
};

export default LocalLoginForm;
