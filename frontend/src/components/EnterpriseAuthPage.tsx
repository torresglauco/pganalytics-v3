import React, { useEffect, useState } from 'react';
import { AlertCircle, ArrowRight } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import LDAPLoginForm from './LDAPLoginForm';
import OAuthProviderSelector from './OAuthProviderSelector';
import type { AuthMethod } from '../types/auth';
import LocalLoginForm from './LocalLoginForm';

interface EnterpriseAuthPageProps {
  availableMethods?: AuthMethod[];
}

export const EnterpriseAuthPage: React.FC<EnterpriseAuthPageProps> = ({
  availableMethods = ['local', 'ldap', 'oauth'],
}) => {
  const [selectedMethod, setSelectedMethod] = useState<AuthMethod | null>(null);
  const [showOAuthSelector, setShowOAuthSelector] = useState(false);
  const { error, clearError } = useAuth();

  // Set default method if only one is available
  useEffect(() => {
    if (!selectedMethod && availableMethods.length === 1) {
      setSelectedMethod(availableMethods[0]);
    }
  }, [availableMethods, selectedMethod]);

  const handleClearError = () => {
    clearError();
  };

  // Render selected auth method form
  const renderAuthForm = () => {
    switch (selectedMethod) {
      case 'local':
        return <LocalLoginForm onSuccess={() => {}} />;
      case 'ldap':
        return <LDAPLoginForm onSuccess={() => {}} />;
      case 'oauth':
        return showOAuthSelector ? (
          <OAuthProviderSelector onBack={() => setShowOAuthSelector(false)} />
        ) : (
          <div className="text-center">
            <p className="text-gray-600 mb-4">Select OAuth Provider</p>
            <button
              onClick={() => setShowOAuthSelector(true)}
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Continue with OAuth <ArrowRight className="inline ml-2" size={16} />
            </button>
          </div>
        );
      case 'saml':
        return (
          <div className="text-center py-8">
            <p className="text-gray-600 mb-4">Redirecting to your identity provider...</p>
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center px-4 py-12">
      <div className="w-full max-w-md">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">pgAnalytics</h1>
          <p className="text-gray-600">PostgreSQL Performance Analytics</p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4 flex gap-3">
            <AlertCircle className="text-red-600 flex-shrink-0 mt-0.5" size={20} />
            <div className="flex-1">
              <h3 className="font-medium text-red-900">Authentication Error</h3>
              <p className="text-sm text-red-700 mt-1">{error}</p>
            </div>
            <button
              onClick={handleClearError}
              className="text-red-600 hover:text-red-700 font-bold text-lg"
            >
              ×
            </button>
          </div>
        )}

        {/* Auth Card */}
        <div className="bg-white rounded-lg shadow-lg p-8">
          {!selectedMethod ? (
            <>
              {/* Method Selector */}
              <h2 className="text-2xl font-bold text-gray-900 mb-6">Sign In</h2>

              {availableMethods.length > 1 && (
                <p className="text-sm text-gray-600 mb-6">Choose your authentication method</p>
              )}

              <div className="space-y-3">
                {/* Local */}
                {availableMethods.includes('local') && (
                  <button
                    onClick={() => setSelectedMethod('local')}
                    className="w-full p-4 border-2 border-gray-200 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition text-left"
                  >
                    <div className="font-semibold text-gray-900">Username & Password</div>
                    <div className="text-sm text-gray-600">Local account</div>
                  </button>
                )}

                {/* LDAP */}
                {availableMethods.includes('ldap') && (
                  <button
                    onClick={() => setSelectedMethod('ldap')}
                    className="w-full p-4 border-2 border-gray-200 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition text-left"
                  >
                    <div className="font-semibold text-gray-900">LDAP / Active Directory</div>
                    <div className="text-sm text-gray-600">Enterprise directory</div>
                  </button>
                )}

                {/* SAML */}
                {availableMethods.includes('saml') && (
                  <button
                    onClick={() => setSelectedMethod('saml')}
                    className="w-full p-4 border-2 border-gray-200 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition text-left"
                  >
                    <div className="font-semibold text-gray-900">SAML 2.0</div>
                    <div className="text-sm text-gray-600">Single sign-on</div>
                  </button>
                )}

                {/* OAuth */}
                {availableMethods.includes('oauth') && (
                  <button
                    onClick={() => setSelectedMethod('oauth')}
                    className="w-full p-4 border-2 border-gray-200 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition text-left"
                  >
                    <div className="font-semibold text-gray-900">OAuth 2.0 / OIDC</div>
                    <div className="text-sm text-gray-600">Google, Azure, GitHub</div>
                  </button>
                )}
              </div>

              {/* Info */}
              <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
                <p className="text-sm text-blue-900">
                  <strong>First time?</strong> Contact your administrator if you don't have an account.
                </p>
              </div>
            </>
          ) : (
            <>
              {/* Auth Form */}
              <div className="mb-4">
                <button
                  onClick={() => {
                    setSelectedMethod(null);
                    setShowOAuthSelector(false);
                  }}
                  className="text-blue-600 hover:text-blue-700 text-sm font-medium"
                >
                  ← Back to method selection
                </button>
              </div>

              {renderAuthForm()}
            </>
          )}
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-sm text-gray-600">
          <p>© 2026 pgAnalytics. All rights reserved.</p>
          <p className="mt-1">v3.5.0 - PostgreSQL Performance Analytics Platform</p>
        </div>
      </div>
    </div>
  );
};

export default EnterpriseAuthPage;
