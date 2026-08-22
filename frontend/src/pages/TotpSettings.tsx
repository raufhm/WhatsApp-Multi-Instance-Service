import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import TotpCodeInput from '@/components/ui/TotpCodeInput'
import BackupCodesDisplay from '@/components/ui/BackupCodesDisplay'
import { totpApi } from '@/lib/apiClient'
import type { TOTPStatus } from '@/types'
import {
  ShieldCheck,
  Key,
  RefreshCw,
  AlertTriangle,
  CheckCircle2,
  Loader2,
  X,
} from 'lucide-react'

export const TotpSettings: React.FC = () => {
  const [status, setStatus] = useState<TOTPStatus | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Modal for Regenerating Backup Codes
  const [showRegenModal, setShowRegenModal] = useState(false)
  const [regenTotpCode, setRegenTotpCode] = useState('')
  const [isRegenerating, setIsRegenerating] = useState(false)
  const [regenError, setRegenError] = useState<string | null>(null)

  // Displaying newly regenerated backup codes
  const [newBackupCodes, setNewBackupCodes] = useState<string[] | null>(null)

  useEffect(() => {
    fetchTotpStatus()
  }, [])

  const fetchTotpStatus = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const data = await totpApi.getStatus()
      setStatus(data)
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to load TOTP status')
    } finally {
      setIsLoading(false)
    }
  }

  const handleRegenerateCodes = async (codeToUse?: string) => {
    const code = (codeToUse || regenTotpCode).trim()
    if (code.length !== 6) {
      setRegenError('Please enter your 6-digit authenticator code')
      return
    }

    setRegenError(null)
    setIsRegenerating(true)

    try {
      const res = await totpApi.regenerateBackupCodes(code)
      const codes = res.backup_codes
      if (!codes || codes.length === 0) {
        throw new Error('No backup codes returned by server')
      }
      setNewBackupCodes(codes)
      setShowRegenModal(false)
      setRegenTotpCode('')
      setStatus((prev) => (prev ? { ...prev, backup_codes_remaining: codes.length } : prev))
    } catch (err: any) {
      setRegenError(err.response?.data?.error || err.message || 'Failed to regenerate backup codes. Invalid code.')
    } finally {
      setIsRegenerating(false)
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-5">
      <div>
        <h1 className="text-xl font-semibold text-gray-900 flex items-center gap-2">
          <ShieldCheck className="h-6 w-6 text-primary-600" />
          Security & Authenticator Settings
        </h1>
        <p className="text-[13px] text-gray-600 mt-1">
          Manage your Time-based One-Time Password (TOTP) two-factor authentication and recovery backup
          codes.
        </p>
      </div>

      {error && (
        <div className="flex items-start justify-between gap-4 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
          <span className="flex-1">{error}</span>
          <Button variant="primary" size="sm" type="button" onClick={fetchTotpStatus} disabled={isLoading}>
            Retry
          </Button>
        </div>
      )}

      {isLoading && (
        <div className="flex items-center justify-center py-6">
          <Loader2 className="h-6 w-6 animate-spin text-primary-600" />
        </div>
      )}

      {newBackupCodes ? (
        <Card className="p-5 sm:p-6">
          <div className="mb-4 flex items-center justify-between">
            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-green-100 text-green-800">
              <CheckCircle2 className="h-4 w-4 text-green-600" />
              New Backup Codes Generated
            </span>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setNewBackupCodes(null)}
            >
              Done
            </Button>
          </div>
          <BackupCodesDisplay
            codes={newBackupCodes}
            title="Your New 10 Backup Codes"
            description="All previous backup codes have been invalidated. Save these new recovery codes securely."
            onAcknowledge={() => setNewBackupCodes(null)}
            acknowledgeLabel="I have saved my new backup codes"
          />
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {/* Card 1: TOTP Authenticator Status */}
          <Card className="p-5 space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-lg font-bold text-gray-900">Authenticator (TOTP)</h2>
                <p className="text-xs text-gray-500 mt-0.5">Primary login factor</p>
              </div>
              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-cyan-100 text-cyan-800">
                <CheckCircle2 className="h-3.5 w-3.5" />
                Active
              </span>
            </div>

            <div className="p-4 bg-gray-50 rounded-xl space-y-3 text-xs text-gray-700 border border-gray-100">
              <div className="flex items-center justify-between">
                <span className="text-gray-500">Method:</span>
                <span className="font-semibold">6-digit TOTP (30s)</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-500">Protection:</span>
                <span className="font-semibold text-cyan-700">AES-256-GCM Encrypted</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-gray-500">Verified On:</span>
                <span className="font-semibold">
                  {status?.verified_at
                    ? new Date(status.verified_at).toLocaleDateString()
                    : 'Active'}
                </span>
              </div>
            </div>

            <p className="text-xs text-gray-500 leading-relaxed">
              Your account is protected with passwordless two-factor authentication. To change devices,
              contact your organization administrator.
            </p>
          </Card>

          {/* Card 2: Backup Codes Status & Regeneration */}
          <Card className="p-5 space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-lg font-bold text-gray-900">Recovery Backup Codes</h2>
                <p className="text-xs text-gray-500 mt-0.5">Emergency single-use recovery</p>
              </div>
              <span
                className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold ${
                  (status?.backup_codes_remaining ?? 10) > 3
                    ? 'bg-cyan-100 text-cyan-800'
                    : 'bg-amber-100 text-amber-800'
                }`}
              >
                <Key className="h-3.5 w-3.5" />
                {status?.backup_codes_remaining ?? 10} Remaining
              </span>
            </div>

            <div className="p-4 bg-gray-50 rounded-xl space-y-2 text-xs text-gray-700 border border-gray-100">
              <p className="leading-relaxed">
                Backup codes allow you to log in if you lose access to your authenticator app. Each code can
                only be used once.
              </p>
              {(status?.backup_codes_remaining ?? 10) <= 3 && (
                <p className="text-amber-700 font-semibold flex items-center gap-1 mt-2">
                  <AlertTriangle className="h-3.5 w-3.5 text-amber-600" />
                  Running low on backup codes. Please generate new ones.
                </p>
              )}
            </div>

            <div className="pt-2">
              <Button
                type="button"
                variant="secondary"
                size="md"
                onClick={() => {
                  setRegenError(null)
                  setRegenTotpCode('')
                  setShowRegenModal(true)
                }}
                className="w-full justify-center flex items-center gap-2"
              >
                <RefreshCw className="h-4 w-4" />
                <span>Regenerate 10 Backup Codes</span>
              </Button>
            </div>
          </Card>
        </div>
      )}

      {/* Confirmation Modal for Regenerating Backup Codes */}
      {showRegenModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50 animate-fadeIn">
          <div className="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl relative space-y-5">
            <button
              type="button"
              onClick={() => setShowRegenModal(false)}
              className="absolute top-4 right-4 p-1 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100"
            >
              <X className="h-5 w-5" />
            </button>

            <div className="flex items-center gap-3">
              <div className="p-3 bg-amber-100 text-amber-700 rounded-xl">
                <AlertTriangle className="h-6 w-6" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">Regenerate Backup Codes</h3>
                <p className="text-xs text-gray-500">Security verification required</p>
              </div>
            </div>

            <div className="text-xs text-gray-600 space-y-2">
              <p>
                Regenerating backup codes will <strong>permanently invalidate</strong> all existing unused
                backup codes.
              </p>
              <p>Please enter your current 6-digit TOTP code to confirm this action:</p>
            </div>

            {regenError && (
              <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-xs text-red-700">
                {regenError}
              </div>
            )}

            <div className="py-2 flex justify-center">
              <TotpCodeInput
                id="regen-totp"
                value={regenTotpCode}
                onChange={setRegenTotpCode}
                onComplete={(code) => handleRegenerateCodes(code)}
                disabled={isRegenerating}
                error={!!regenError}
              />
            </div>

            <div className="flex items-center justify-end gap-3 pt-2">
              <Button
                type="button"
                variant="ghost"
                size="md"
                onClick={() => setShowRegenModal(false)}
                disabled={isRegenerating}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="primary"
                size="md"
                onClick={() => handleRegenerateCodes()}
                disabled={isRegenerating || regenTotpCode.length !== 6}
              >
                {isRegenerating ? (
                  <span className="flex items-center gap-2">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Regenerating...
                  </span>
                ) : (
                  'Confirm & Generate'
                )}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default TotpSettings
