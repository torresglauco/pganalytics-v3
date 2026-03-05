import React, { useState } from 'react';
import { AlertCircle, CheckCircle, Copy, Download } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import type { MFAMethod, MFASetupResponse } from '../types/auth';

interface MFASetupWizardProps {
  onComplete?: () => void;
  onCancel?: () => void;
  method: MFAMethod;
}

export const MFASetupWizard: React.FC<MFASetupWizardProps> = ({
  onComplete,
  onCancel,
  method,
}) => {
  const [step, setStep] = useState<'setup' | 'verify' | 'complete'>('setup');
  const [setupData, setSetupData] = useState<MFASetupResponse | null>(null);
  const [verificationCode, setVerificationCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copiedBackupCodes, setCopiedBackupCodes] = useState(false);

  const { setupMFA, completeMFASetup } = useAuth();

  // Step 1: Setup
  const handleSetup = async () => {
    try {
      setError(null);
      setIsLoading(true);
      const data = await setupMFA(method);
      setSetupData(data);
      setStep('verify');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Setup failed');
    } finally {
      setIsLoading(false);
    }
  };

  // Step 2: Verify
  const handleVerify = async () => {
    if (!verificationCode || !setupData) {
      setError('Verification code is required');
      return;
    }

    try {
      setError(null);
      setIsLoading(true);
      await completeMFASetup({
        setup_id: setupData.setup_id,
        code: verificationCode,
        method,
      });
      setStep('complete');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setIsLoading(false);
    }
  };

  // Download backup codes
  const downloadBackupCodes = () => {
    if (!setupData?.backup_codes) return;

    const content = setupData.backup_codes.join('\n');
    const element = document.createElement('a');
    element.setAttribute(
      'href',
      `data:text/plain;charset=utf-8,${encodeURIComponent(content)}`
    );
    element.setAttribute('download', `pganalytics-backup-codes-${Date.now()}.txt`);
    element.style.display = 'none';
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
  };

  // Copy backup codes to clipboard
  const copyBackupCodes = () => {
    if (!setupData?.backup_codes) return;
    navigator.clipboard.writeText(setupData.backup_codes.join('\n'));
    setCopiedBackupCodes(true);
    setTimeout(() => setCopiedBackupCodes(false), 2000);
  };

  return (
    <div className="space-y-6">
      <h3 className="text-xl font-semibold text-gray-900">
        Set up {method === 'totp' ? 'Authenticator App' : 'Two-Factor Authentication'}
      </h3>

      {error && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700 flex gap-2">
          <AlertCircle size={16} className="mt-0.5 flex-shrink-0" />
          <div>{error}</div>
        </div>
      )}

      {/* STEP 1: Setup */}
      {step === 'setup' && (
        <div className="space-y-4">
          <p className="text-gray-600">
            {method === 'totp'
              ? 'Scan the QR code with your authenticator app (Google Authenticator, Authy, Microsoft Authenticator, etc.)'
              : 'A verification code will be sent to your phone'}
          </p>

          {method === 'totp' && setupData?.qr_code && (
            <div className="bg-white p-4 border border-gray-200 rounded-lg">
              <img src={setupData.qr_code} alt="QR Code" className="mx-auto h-64 w-64" />
            </div>
          )}

          {setupData?.secret && (
            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
              <p className="text-xs text-gray-600 mb-2">Can't scan? Enter this code instead:</p>
              <code className="text-sm font-mono text-gray-900 break-all">{setupData.secret}</code>
            </div>
          )}

          <button
            onClick={handleSetup}
            disabled={isLoading}
            className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg transition"
          >
            {isLoading ? 'Setting up...' : 'Setup Complete, Continue'}
          </button>
        </div>
      )}

      {/* STEP 2: Verify */}
      {step === 'verify' && (
        <div className="space-y-4">
          <p className="text-gray-600">
            {method === 'totp'
              ? 'Enter the 6-digit code from your authenticator app'
              : 'Enter the code sent to your phone'}
          </p>

          <input
            type="text"
            value={verificationCode}
            onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="000000"
            maxLength={6}
            disabled={isLoading}
            className="w-full px-4 py-3 border border-gray-300 rounded-lg text-center text-2xl tracking-widest font-mono focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />

          <button
            onClick={handleVerify}
            disabled={isLoading || verificationCode.length !== 6}
            className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white font-medium rounded-lg transition"
          >
            {isLoading ? 'Verifying...' : 'Verify'}
          </button>

          <button
            onClick={() => setStep('setup')}
            className="w-full py-2 px-4 border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium rounded-lg transition"
          >
            Back
          </button>
        </div>
      )}

      {/* STEP 3: Complete */}
      {step === 'complete' && setupData?.backup_codes && (
        <div className="space-y-4">
          <div className="p-4 bg-green-50 border border-green-200 rounded-lg flex gap-3">
            <CheckCircle className="text-green-600 flex-shrink-0 mt-0.5" size={20} />
            <div>
              <h4 className="font-semibold text-green-900">Setup Complete!</h4>
              <p className="text-sm text-green-700 mt-1">
                Your two-factor authentication is now enabled.
              </p>
            </div>
          </div>

          <div className="space-y-3">
            <h4 className="font-semibold text-gray-900">Save your backup codes</h4>
            <p className="text-sm text-gray-600">
              Save these codes in a safe place. You can use them if you lose access to your authenticator.
            </p>

            <div className="bg-gray-50 p-4 rounded-lg border border-gray-200 font-mono text-sm space-y-1">
              {setupData.backup_codes.map((code, idx) => (
                <div key={idx} className="text-gray-900">
                  {code}
                </div>
              ))}
            </div>

            <div className="flex gap-2">
              <button
                onClick={copyBackupCodes}
                className="flex-1 py-2 px-4 border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium rounded-lg transition flex items-center justify-center gap-2"
              >
                <Copy size={16} />
                {copiedBackupCodes ? 'Copied!' : 'Copy'}
              </button>
              <button
                onClick={downloadBackupCodes}
                className="flex-1 py-2 px-4 border border-gray-300 text-gray-700 hover:bg-gray-50 font-medium rounded-lg transition flex items-center justify-center gap-2"
              >
                <Download size={16} />
                Download
              </button>
            </div>
          </div>

          <button
            onClick={onComplete}
            className="w-full py-2 px-4 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition"
          >
            Done
          </button>
        </div>
      )}
    </div>
  );
};

export default MFASetupWizard;
