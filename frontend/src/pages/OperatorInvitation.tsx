import React, { useState, useEffect } from 'react'
import { useNavigate, useParams, Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import TotpQrCode from '@/components/ui/TotpQrCode'
import TotpCodeInput from '@/components/ui/TotpCodeInput'
import BackupCodesDisplay from '@/components/ui/BackupCodesDisplay'
import { onboardingApi } from '@/lib/apiClient'
import { useAuth } from '@/hooks/useAuth'
import type { InvitationDetails } from '@/types'
import {
  UserPlus,
  Building2,
  ShieldCheck,
  AlertCircle,
  Loader2,
  ArrowLeft,
  Phone,
} from 'lucide-react'

export const OperatorInvitation: React.FC = () => {
  // Can get token from URL param or search param or input
  const params = useParams({ strict: false }) as { token?: string }
  const [token, setToken] = useState<string>(params?.token || '')
  const [invitation, setInvitation] = useState<InvitationDetails | null>(null)

  const [operatorName, setOperatorName] = useState('')
  const [operatorEmail, setOperatorEmail] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])

  const [step, setStep] = useState<'INPUT_TOKEN' | 'DETAILS' | 'TOTP_SETUP' | 'BACKUP_CODES' | 'INVALID'>('INPUT_TOKEN')
  const [setupToken, setSetupToken] = useState<string>('')
  const [totpData, setTotpData] = useState<{
    secret: string
    otpauth_url: string
    qr_code_data_url?: string
    qr_code?: string
    qr_code_svg?: string
  } | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { setSessionUser } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    // Check if token was provided in URL query or path
    const urlParams = new URLSearchParams(window.location.search)
    const tokenFromQuery = urlParams.get('token')
    const initialToken = params?.token || tokenFromQuery

    if (initialToken) {
      setToken(initialToken)
      fetchInvitation(initialToken)
    }
  }, [params?.token])

  const fetchInvitation = async (invToken: string) => {
    setIsLoading(true)
    setError(null)
    try {
      const details = await onboardingApi.getInvitationDetails(invToken)
      setInvitation(details)
      if (details.email) setOperatorEmail(details.email)
      setStep('DETAILS')
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to load invitation')
      setStep('INVALID')
    } finally {
      setIsLoading(false)
    }
  }

  const handleDetailsSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!operatorName.trim()) {
      setError('Please enter your full name')
      return
    }

    setError(null)
    setIsSubmitting(true)

    try {
      const res = await onboardingApi.signupOperator({
        token,
        name: operatorName.trim(),
        email: operatorEmail.trim() || undefined,
      })

      const targetSetupToken = res.setup_token || res.temp_token || ''
      setSetupToken(targetSetupToken)

      if (!targetSetupToken) {
        throw new Error('Missing TOTP setup token from server')
      }

      try {
        const setup = await onboardingApi.getTotpSetup(targetSetupToken)
        setTotpData(setup)
        setStep('TOTP_SETUP')
      } catch (err: any) {
        throw new Error(err.response?.data?.error || err.message || 'Failed to load TOTP setup.')
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to submit details. Please try again.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleAcceptSubmit = async (codeToUse?: string) => {
    const code = (codeToUse || totpCode).trim()
    if (code.length !== 6) {
      setError('Please enter the 6-digit code from your authenticator app')
      return
    }

    if (!setupToken) {
      setError('Missing setup token. Please enter your details again.')
      return
    }

    setError(null)
    setIsSubmitting(true)

    try {
      const res = await onboardingApi.verifyTotpSetup(setupToken, code)

      const codes = res.backup_codes
      if (!codes || codes.length === 0) {
        throw new Error('No backup codes returned by server')
      }
      setBackupCodes(codes)

      const currentUser = res.user
      if (!currentUser) {
        throw new Error('No user returned by server')
      }

      setSessionUser(currentUser, currentUser.tenant_id)
      setStep('BACKUP_CODES')
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to verify code. Please check your authenticator code.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleFinish = () => {
    navigate({ to: '/' })
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
          <div className="inline-flex items-center justify-center h-12 w-12 rounded-xl bg-green-600 text-white shadow mb-3">
            <UserPlus className="h-6 w-6" />
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-gray-900 tracking-tight">
            Accept Operator Invitation
          </h1>
          <p className="mt-1 text-sm text-gray-600">
            {step === 'INPUT_TOKEN' && 'Enter your invitation code to join your organization'}
            {step === 'DETAILS' && 'Configure your profile to get started'}
            {step === 'TOTP_SETUP' && 'Set up your two-factor authenticator'}
            {step === 'BACKUP_CODES' && 'Save your emergency recovery backup codes'}
            {step === 'INVALID' && 'This invitation is invalid or expired'}
          </p>
        </div>

        <Card className="p-6 sm:p-8 shadow-md">
          {error && (
            <div className="mb-5 flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
              <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* STEP 1: Enter Token (if no token in URL) */}
          {step === 'INPUT_TOKEN' && (
            <form
              onSubmit={(e) => {
                e.preventDefault()
                fetchInvitation(token)
              }}
              className="space-y-4"
            >
              <div>
                <Label htmlFor="invitationToken">Invitation Token / Code</Label>
                <Input
                  id="invitationToken"
                  type="text"
                  value={token}
                  onChange={(e) => setToken(e.target.value)}
                  required
                  placeholder="Paste invitation code from WhatsApp or Email"
                  className="mt-1 font-mono text-center tracking-wider text-sm"
                  disabled={isLoading}
                />
              </div>

              <Button
                type="submit"
                variant="primary"
                size="lg"
                className="w-full justify-center"
                disabled={isLoading || !token.trim()}
              >
                {isLoading ? (
                  <span className="flex items-center gap-2">
                    <Loader2 className="h-5 w-5 animate-spin" />
                    Finding Invitation...
                  </span>
                ) : (
                  'Continue with Invitation'
                )}
              </Button>
            </form>
          )}

          {/* STEP 2: Name + Details */}
          {step === 'DETAILS' && invitation && (
            <form onSubmit={handleDetailsSubmit} className="space-y-6">
              {/* Invitation metadata banner */}
              <div className="p-4 bg-primary-50 border border-primary-200 rounded-xl">
                <div className="flex items-center justify-between text-xs text-primary-900 mb-1">
                  <span className="font-semibold flex items-center gap-1.5">
                    <Building2 className="h-4 w-4 text-primary-600" />
                    {invitation.tenant_name || 'Your Organization'}
                  </span>
                  <span className="bg-primary-200/70 font-mono px-2 py-0.5 rounded text-[11px] uppercase">
                    Role: {invitation.role}
                  </span>
                </div>
                <p className="text-xs text-primary-700 flex items-center gap-1 mt-1">
                  <Phone className="h-3.5 w-3.5" /> Invited Identifier:{' '}
                  <strong>{invitation.identifier || invitation.whatsapp_number}</strong>
                </p>
              </div>

              {/* Profile fields */}
              <div className="space-y-4">
                <div>
                  <Label htmlFor="opName">Your Full Name *</Label>
                  <Input
                    id="opName"
                    type="text"
                    value={operatorName}
                    onChange={(e) => setOperatorName(e.target.value)}
                    required
                    placeholder="Jane Operator"
                    className="mt-1"
                    disabled={isSubmitting}
                  />
                </div>

                <div>
                  <Label htmlFor="opEmail">Email Address (Optional)</Label>
                  <Input
                    id="opEmail"
                    type="email"
                    value={operatorEmail}
                    onChange={(e) => setOperatorEmail(e.target.value)}
                    placeholder="jane@example.com"
                    className="mt-1"
                    disabled={isSubmitting}
                  />
                </div>
              </div>

              <Button
                type="submit"
                variant="primary"
                size="lg"
                className="w-full justify-center"
                disabled={isSubmitting || !operatorName.trim()}
              >
                {isSubmitting ? (
                  <span className="flex items-center gap-2">
                    <Loader2 className="h-5 w-5 animate-spin" />
                    Saving details...
                  </span>
                ) : (
                  'Continue to Authenticator Setup'
                )}
              </Button>
            </form>
          )}

          {/* STEP 3: TOTP Setup */}
          {step === 'TOTP_SETUP' && totpData && (
            <div className="space-y-6">
              <div className="text-center mb-4">
                <h3 className="text-base font-bold text-gray-900 flex items-center justify-center gap-2">
                  <ShieldCheck className="h-5 w-5 text-primary-600" />
                  Scan Authenticator QR Code
                </h3>
                <p className="text-xs text-gray-500 mt-0.5">
                  Scan with Google Authenticator, Authy, or 1Password.
                </p>
              </div>

              <TotpQrCode
                secret={totpData.secret}
                otpauthUrl={totpData.otpauth_url}
                qrSvg={totpData.qr_code_svg}
                qrDataUrl={totpData.qr_code_data_url || totpData.qr_code}
                issuer={invitation?.tenant_name}
                accountName={operatorEmail || operatorName || invitation?.identifier}
              />

              <div className="mt-5 pt-4 border-t border-gray-200">
                <p className="text-center text-xs font-semibold text-gray-700 mb-3">
                  Enter the 6-digit code from your authenticator:
                </p>

                <TotpCodeInput
                  value={totpCode}
                  onChange={setTotpCode}
                  onComplete={(code) => handleAcceptSubmit(code)}
                  disabled={isSubmitting}
                  error={!!error}
                />

                <div className="mt-5">
                  <Button
                    type="button"
                    variant="primary"
                    size="lg"
                    className="w-full justify-center"
                    disabled={isSubmitting || totpCode.length !== 6}
                    onClick={() => handleAcceptSubmit()}
                  >
                    {isSubmitting ? (
                      <span className="flex items-center gap-2">
                        <Loader2 className="h-5 w-5 animate-spin" />
                        Verifying Code...
                      </span>
                    ) : (
                      'Activate & Join Workspace'
                    )}
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* STEP 4: Backup Codes */}
          {step === 'BACKUP_CODES' && (
            <BackupCodesDisplay
              codes={backupCodes}
              title="Your 10 Recovery Backup Codes"
              description="Save these codes securely. You can use any of them once if you lose access to your phone or authenticator app."
              showAcknowledgeCheckbox={true}
              acknowledgeLabel="Launch Dashboard"
              onAcknowledge={handleFinish}
            />
          )}
        </Card>
      </div>
    </div>
  )
}

export default OperatorInvitation
