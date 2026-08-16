import React, { useState, useEffect } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import TotpQrCode from '@/components/ui/TotpQrCode'
import TotpCodeInput from '@/components/ui/TotpCodeInput'
import BackupCodesDisplay from '@/components/ui/BackupCodesDisplay'
import { onboardingApi } from '@/lib/apiClient'
import { useAuth } from '@/hooks/useAuth'
import type { TOTPSetupData } from '@/types'
import { Mail, CheckCircle2, AlertCircle, Loader2, ArrowLeft } from 'lucide-react'

export const VerifyEmail: React.FC = () => {
  const [token, setToken] = useState('')
  const [isVerifying, setIsVerifying] = useState(false)
  const [verified, setVerified] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // TOTP setup state after email verification
  const [tempToken, setTempToken] = useState('')
  const [totpData, setTotpData] = useState<TOTPSetupData | null>(null)
  const [totpCode, setTotpCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [isSettingUpTotp, setIsSettingUpTotp] = useState(false)

  const { setSessionUser } = useAuth()
  const navigate = useNavigate()

  // Extract query params if arrived via ?token=...
  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const tokenParam = urlParams.get('token')
    if (tokenParam) {
      setToken(tokenParam)
      handleVerify(tokenParam)
    }
  }, [])

  const handleVerify = async (tokenToVerify?: string) => {
    const targetToken = (tokenToVerify || token).trim()
    if (!targetToken) {
      setError('Please enter your email verification token')
      return
    }

    setError(null)
    setIsVerifying(true)

    try {
      const res = await onboardingApi.verifyEmail(targetToken)
      setVerified(true)
      const setupToken = res.setup_token || res.temp_token || ''
      if (setupToken) {
        setTempToken(setupToken)
      }

      if (res.totp_setup) {
        setTotpData(res.totp_setup)
      } else if (setupToken) {
        try {
          const setup = await onboardingApi.getTotpSetup(setupToken)
          setTotpData(setup)
        } catch (err: any) {
          throw new Error(err.response?.data?.error || err.message || 'Failed to load TOTP setup.')
        }
      } else {
        throw new Error('Missing TOTP setup data from server')
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Invalid or expired email verification token.')
    } finally {
      setIsVerifying(false)
    }
  }

  const handleVerifyTotp = async (codeToVerify?: string) => {
    const code = (codeToVerify || totpCode).trim()
    if (code.length !== 6) {
      setError('Please enter the 6-digit code from your authenticator app')
      return
    }

    if (!tempToken) {
      setError('Missing setup token. Please verify email token first.')
      return
    }

    setError(null)
    setIsSettingUpTotp(true)

    try {
      const res = await onboardingApi.verifyTotpSetup(tempToken, code)
      const codes = res.backup_codes
      if (!codes || codes.length === 0) {
        throw new Error('No backup codes returned by server')
      }
      setBackupCodes(codes)

      if (res.user) {
        setSessionUser(res.user)
      } else {
        throw new Error('No user returned by server')
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Invalid authenticator code. Please try again.')
    } finally {
      setIsSettingUpTotp(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-lg">
        <div className="text-center mb-6">
          <Link
            to="/signup"
            className="inline-flex items-center gap-1 text-xs font-medium text-gray-500 hover:text-gray-700 mb-3"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> Back
          </Link>
          <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">
            Email Verification
          </h1>
          <p className="mt-1 text-sm text-gray-600">
            {!verified
              ? 'Verify your administrator email address to continue setup'
              : backupCodes.length > 0
              ? 'Save your backup codes'
              : 'Set up your authenticator app'}
          </p>
        </div>

        <Card className="p-6 sm:p-8 shadow-md">
          {error && (
            <div className="mb-5 flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
              <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {!verified ? (
            <form
              onSubmit={(e) => {
                e.preventDefault()
                handleVerify()
              }}
              className="space-y-4"
            >
              <div className="text-center py-2">
                <div className="inline-flex p-3 bg-primary-100 text-primary-600 rounded-full mb-3">
                  <Mail className="h-8 w-8" />
                </div>
                <h2 className="text-base font-semibold text-gray-900">Enter Verification Token</h2>
                <p className="text-xs text-gray-500 mt-1">
                  Paste the verification token you received in your email inbox or click the link in
                  your email.
                </p>
              </div>

              <div>
                <Label htmlFor="token">Verification Token</Label>
                <Input
                  id="token"
                  type="text"
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  required
                  placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
                  className="mt-1 font-mono text-center tracking-wider text-sm"
                  disabled={isVerifying}
                />
              </div>

              <Button
                type="submit"
                variant="primary"
                size="lg"
                className="w-full justify-center"
                disabled={isVerifying || !token.trim()}
              >
                {isVerifying ? (
                  <span className="flex items-center gap-2">
                    <Loader2 className="h-5 w-5 animate-spin" />
                    Verifying Email...
                  </span>
                ) : (
                  'Confirm Verification Token'
                )}
              </Button>
            </form>
          ) : backupCodes.length > 0 ? (
            <BackupCodesDisplay
              codes={backupCodes}
              title="Save Your Backup Codes"
              showAcknowledgeCheckbox={true}
              acknowledgeLabel="Proceed to Setup"
              onAcknowledge={() => navigate({ to: '/setup' })}
            />
          ) : totpData ? (
            <div className="space-y-6">
              <div className="p-3 bg-green-50 border border-green-200 rounded-lg flex items-center gap-2.5 text-xs text-green-800 font-medium">
                <CheckCircle2 className="h-4 w-4 text-green-600 flex-shrink-0" />
                <span>Email verified successfully! Now configure your authenticator app.</span>
              </div>

              <TotpQrCode
                secret={totpData.secret}
                otpauthUrl={totpData.otpauth_url}
                qrSvg={totpData.qr_code_svg}
                qrDataUrl={totpData.qr_code_data_url}
              />

              <div className="pt-2 border-t border-gray-200">
                <p className="text-center text-xs font-semibold text-gray-700 mb-3">
                  Enter the 6-digit code from your authenticator app:
                </p>

                <TotpCodeInput
                  value={totpCode}
                  onChange={setTotpCode}
                  onComplete={handleVerifyTotp}
                  disabled={isSettingUpTotp}
                  error={!!error}
                />

                <div className="mt-5">
                  <Button
                    type="button"
                    variant="primary"
                    size="lg"
                    className="w-full justify-center"
                    disabled={isSettingUpTotp || totpCode.length !== 6}
                    onClick={() => handleVerifyTotp()}
                  >
                    {isSettingUpTotp ? (
                      <span className="flex items-center gap-2">
                        <Loader2 className="h-5 w-5 animate-spin" />
                        Activating Authenticator...
                      </span>
                    ) : (
                      'Verify & Reveal Backup Codes'
                    )}
                  </Button>
                </div>
              </div>
            </div>
          ) : null}
        </Card>
      </div>
    </div>
  )
}

export default VerifyEmail
