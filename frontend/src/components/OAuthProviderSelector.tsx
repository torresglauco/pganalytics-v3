import React, { useState } from 'react';
import { AlertCircle, Loader } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import type { OAuthProvider } from '../types/auth';

interface OAuthProviderSelectorProps {
  onBack?: () => void;
  onSuccess?: () => void;
}

const providers: Array<{
  id: OAuthProvider;
  name: string;
  icon: string;
  color: string;
  description: string;
}> = [
  {
    id: 'google',
    name: 'Google',
    icon: '🔍',
    color: 'hover:bg-blue-50 border-blue-200',
    description: 'Sign in with your Google account',
  },
  {
    id: 'azure_ad',
    name: 'Microsoft Azure AD',
    icon: '☁️',
    color: 'hover:bg-cyan-50 border-cyan-200',
    description: 'Sign in with your Microsoft account',
  },
  {
    id: 'github',
    name: 'GitHub',
    icon: '🐙',
    color: 'hover:bg-gray-50 border-gray-200',
    description: 'Sign in with your GitHub account',
  },
];

export const OAuthProviderSelector: React.FC<OAuthProviderSelectorProps> = ({
  onBack,
  onSuccess,
}) => {
  const [selectedProvider, setSelectedProvider] = useState<OAuthProvider | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  const { initiateOAuth, error: contextError, clearError } = useAuth();

  const handleProviderClick = async (provider: OAuthProvider) => {
    setLocalError(null);
    clearError();

    try {
      setSelectedProvider(provider);
      setIsLoading(true);

      // Get OAuth redirect URL from backend
      const redirectUrl = await initiateOAuth(provider);

      // Redirect to OAuth provider
      window.location.href = redirectUrl;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'OAuth initiation failed';
      setLocalError(errorMessage);
      setSelectedProvider(null);
    } finally {
      setIsLoading(false);
    }
  };

  const displayError = localError || contextError;

  return (
    <div className="space-y-4">
      <h3 className="text-xl font-semibold text-gray-900 mb-6">Choose your OAuth provider</h3>

      {/* Error Message */}
      {displayError && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex gap-2">
          <AlertCircle size={16} className="mt-0.5 flex-shrink-0" />
          <div>{displayError}</div>
        </div>
      )}

      {/* Provider List */}
      <div className="space-y-3">
        {providers.map((provider) => (
          <button
            key={provider.id}
            onClick={() => handleProviderClick(provider.id)}
            disabled={isLoading && selectedProvider !== provider.id}
            className={`w-full p-4 border-2 border-gray-200 rounded-lg transition text-left ${provider.color} ${
              isLoading && selectedProvider !== provider.id ? 'opacity-50 cursor-not-allowed' : ''
            }`}
          >
            <div className="flex items-center gap-3">
              {selectedProvider === provider.id && isLoading ? (
                <Loader className="h-6 w-6 animate-spin text-blue-600" />
              ) : (
                <span className="text-2xl">{provider.icon}</span>
              )}
              <div className="flex-1">
                <div className="font-semibold text-gray-900">{provider.name}</div>
                <div className="text-sm text-gray-600">{provider.description}</div>
              </div>
            </div>
          </button>
        ))}
      </div>

      {/* Custom OIDC Option */}
      <div className="mt-6 p-4 bg-gray-50 border border-gray-200 rounded-lg">
        <p className="text-sm text-gray-600 mb-2">
          <strong>Using a custom OIDC provider?</strong>
        </p>
        <p className="text-xs text-gray-600">
          Contact your administrator to configure your custom OAuth/OIDC provider.
        </p>
      </div>

      {/* Back Button */}
      {onBack && (
        <button
          onClick={onBack}
          disabled={isLoading}
          className="w-full py-2 px-4 border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium rounded-lg transition disabled:opacity-50 mt-4"
        >
          Back
        </button>
      )}
    </div>
  );
};

export default OAuthProviderSelector;
