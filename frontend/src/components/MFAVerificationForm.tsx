import React, { useState } from 'react';
import { AlertCircle, Eye, EyeOff } from 'lucide-react';
import type { MFAMethod } from '../types/auth';

interface MFAVerificationFormProps {
  method: MFAMethod;
  challengeId: string;
  onVerify: (code: string) => Promise<void>;
  onUseBackupCode?: () => void;
  isLoading?: boolean;
  error?: string;
}

export const MFAVerificationForm: React.FC<MFAVerificationFormProps> = ({
  method,
  challengeId: _challengeId,
  onVerify,
  onUseBackupCode: _onUseBackupCode,
  isLoading = false,
  error,
}) => {
  const [code, setCode] = useState('');
  const [showBackupCodeInput, setShowBackupCodeInput] = useState(false);
  const [backupCode, setBackupCode] = useState('');
  const [showBackupCode, setShowBackupCode] = useState(false);

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    const verificationCode = showBackupCodeInput ? backupCode : code;

    if (!verificationCode) {
      return;
    }

    try {
      await onVerify(verificationCode);
    } catch (err) {
      // Error is handled by parent component
    }
  };

  return (
    <form onSubmit={handleVerify} className="space-y-4">
      <h3 className="text-xl font-semibold text-gray-900 mb-6">Two-Factor Authentication</h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex gap-2">
          <AlertCircle size={16} className="mt-0.5 flex-shrink-0" />
          <div>{error}</div>
        </div>
      )}

      {!showBackupCodeInput ? (
        <>
          {/* Standard MFA Code */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              {method === 'totp'
                ? '6-digit code from your authenticator app'
                : '6-digit code sent to your phone'}
            </label>
            <input
              type="text"
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
              placeholder="000000"
              maxLength={6}
              disabled={isLoading}
              autoFocus
              className="w-full px-4 py-3 border border-gray-300 rounded-lg text-center text-2xl tracking-widest font-mono focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
            />
          </div>

          {/* Use Backup Code */}
          <button
            type="button"
            onClick={() => {
              setShowBackupCodeInput(true);
              setCode('');
            }}
            className="text-sm text-blue-600 hover:text-blue-700 font-medium"
          >
            Can't access your authenticator? Use a backup code
          </button>
        </>
      ) : (
        <>
          {/* Backup Code */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Backup code
            </label>
            <div className="relative">
              <input
                type={showBackupCode ? 'text' : 'password'}
                value={backupCode}
                onChange={(e) => setBackupCode(e.target.value)}
                placeholder="Enter backup code"
                disabled={isLoading}
                autoFocus
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:opacity-50"
              />
              <button
                type="button"
                onClick={() => setShowBackupCode(!showBackupCode)}
                className="absolute right-3 top-2.5 text-gray-600 hover:text-gray-700"
                tabIndex={-1}
              >
                {showBackupCode ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
            </div>
            <p className="mt-1 text-xs text-gray-600">
              Enter one of the backup codes you saved during setup
            </p>
          </div>

          {/* Back to MFA Code */}
          <button
            type="button"
            onClick={() => {
              setShowBackupCodeInput(false);
              setBackupCode('');
            }}
            className="text-sm text-blue-600 hover:text-blue-700 font-medium"
          >
            ← Back to authenticator code
          </button>
        </>
      )}

      {/* Submit Button */}
      <button
        type="submit"
        disabled={
          isLoading ||
          (showBackupCodeInput ? !backupCode : code.length !== 6)
        }
        className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg transition mt-6"
      >
        {isLoading ? (
          <span className="flex items-center justify-center gap-2">
            <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
            Verifying...
          </span>
        ) : (
          'Verify'
        )}
      </button>

      {/* Help Text */}
      <div className="text-center mt-4 text-xs text-gray-600">
        <p>Having trouble with 2FA?</p>
        <a href="/support" className="text-blue-600 hover:text-blue-700">
          Contact support
        </a>
      </div>
    </form>
  );
};

export default MFAVerificationForm;
