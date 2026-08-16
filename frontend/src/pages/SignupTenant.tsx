import React, { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import PhoneInput from '@/components/ui/PhoneInput'
import TotpQrCode from '@/components/ui/TotpQrCode'
import TotpCodeInput from '@/components/ui/TotpCodeInput'
import BackupCodesDisplay from '@/components/ui/BackupCodesDisplay'
import { onboardingApi } from '@/lib/apiClient'
import { useAuth } from '@/hooks/useAuth'
import type { TOTPSetupData } from '@/types'
import {
  Building2,
  User,
  Mail,
  Loader2,
  AlertCircle,
  CheckCircle2,
  Shield,
  ArrowLeft,
} from 'lucide-react'

type SignupStep = 'FORM' | 'VERIFY_EMAIL' | 'TOTP_SETUP' | 'BACKUP_CODES'

export const SignupTenant: React.FC = () => {
  const [step, setStep] = useState<SignupStep>('FORM')
  const [orgName, setOrgName] = useState('')
  const [adminName, setAdminName] = useState('')
  const [adminEmail, setAdminEmail] = useState('')
  const [whatsappNumber, setWhatsappNumber] = useState('')

  // Verification & TOTP state
  const [emailToken, setEmailToken] = useState('')
  const [tempToken, setTempToken] = useState('')
  const [tenantId, setTenantId] = useState('')
  const [totpData, setTotpData] = useState<TOTPSetupData | null>(null)
  const [totpCode, setTotpCode] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { setSessionUser } = useAuth()
  const navigate = useNavigate()

  const handleRegisterSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    try {
      const res = await onboardingApi.signupTenant({
        org_name: orgName.trim(),
        admin_name: adminName.trim(),
        admin_email: adminEmail.trim(),
        whatsapp_number: whatsappNumber.trim(),
      })

      const tId = res.tenant?.id || res.tenant_id || ''
      setTenantId(tId)
      const token = res.verification_token || res.temp_token || res.setup_token || ''
      if (token) {
        setEmailToken(token)
        setTempToken(token)
      }

      if (res.email_verification_required || (adminEmail.trim() && res.verification_token)) {
        setStep('VERIFY_EMAIL')
      } else if (res.totp_setup) {
        setTotpData(res.totp_setup)
        setStep('TOTP_SETUP')
      } else if (res.setup_token || res.temp_token) {
        const setupToken = res.setup_token || res.temp_token || ''
        setTempToken(setupToken)
        try {
          const setup = await onboardingApi.getTotpSetup(setupToken)
          setTotpData(setup)
          setStep('TOTP_SETUP')
        } catch (err: any) {
          setError(err.response?.data?.error || err.message || 'Failed to load TOTP setup. Please try again.')
        }
      } else {
        setStep('VERIFY_EMAIL')
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Registration failed. Please check your inputs.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleVerifyEmail = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!emailToken.trim()) {
      setError('Please enter the verification code sent to your email')
      return
    }

    setError(null)
    setIsLoading(true)

    try {
      const res = await onboardingApi.verifyEmail(emailToken.trim())
      const setupToken = res.setup_token || res.temp_token || ''
      if (setupToken) {
        setTempToken(setupToken)
      }
      if (res.totp_setup) {
        setTotpData(res.totp_setup)
        setStep('TOTP_SETUP')
      } else if (setupToken) {
        try {
          const setup = await onboardingApi.getTotpSetup(setupToken)
          setTotpData(setup)
          setStep('TOTP_SETUP')
        } catch (err: any) {
          setError(err.response?.data?.error || err.message || 'Failed to load TOTP setup. Please try again.')
        }
      } else {
        setStep('TOTP_SETUP')
      }
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Invalid or expired verification token.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleVerifyTotp = async (codeToVerify?: string) => {
    const code = (codeToVerify || totpCode).trim()
    if (code.length !== 6) {
      setError('Please enter the full 6-digit code from your authenticator app')
      return
    }

    if (!tempToken) {
      setError('Missing setup token. Please complete email verification first.')
      return
    }

    setError(null)
    setIsLoading(true)

    try {
      const res = await onboardingApi.verifyTotpSetup(tempToken, code)
      const codes = res.backup_codes
      if (!codes || codes.length === 0) {
        throw new Error('No backup codes returned by server')
      }
      setBackupCodes(codes)

      const currentUser = res.user
      if (!currentUser) {
        throw new Error('No user returned by server')
      }

      setSessionUser(currentUser, tenantId)
      setStep('BACKUP_CODES')
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Invalid authenticator code. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleFinishOnboarding = () => {
    navigate({ to: '/setup' })
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-lg">
        {/* Header Branding */}
        <div className="text-center mb-6">
          <Link
            to="/signup"
            className="inline-flex items-center gap-1 text-xs font-medium text-gray-500 hover:text-gray-700 mb-3"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> Back to options
          </Link>
          <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">
            Register Organization
          </h1>
          <p className="mt-1 text-sm text-gray-600">
            {step === 'FORM' && 'Create your tenant workspace with passwordless TOTP security'}
            {step === 'VERIFY_EMAIL' && 'Verify your administrator email address'}
            {step === 'TOTP_SETUP' && 'Set up two-factor authenticator app'}
            {step === 'BACKUP_CODES' && 'Download your emergency recovery backup codes'}
          </p>
        </div>

        {/* Step Indicator */}
        <div className="mb-6 flex items-center justify-between max-w-xs mx-auto">
          {[
            { id: 'FORM', label: '1. Details' },
            { id: 'TOTP_SETUP', label: '2. Authenticator' },
            { id: 'BACKUP_CODES', label: '3. Backup Codes' },
          ].map((s, idx) => {
            const isDone =
              (s.id === 'FORM' && step !== 'FORM') ||
              (s.id === 'TOTP_SETUP' && step === 'BACKUP_CODES')
            const isCurrent =
              (s.id === 'FORM' && (step === 'FORM' || step === 'VERIFY_EMAIL')) ||
              (s.id === 'TOTP_SETUP' && step === 'TOTP_SETUP') ||
              (s.id === 'BACKUP_CODES' && step === 'BACKUP_CODES')

            return (
              <div key={idx} className="flex items-center text-xs font-medium">
                <span
                  className={`flex items-center justify-center w-6 h-6 rounded-full mr-1.5 ${
                    isDone
                      ? 'bg-green-600 text-white'
                      : isCurrent
                      ? 'bg-primary-600 text-white font-bold'
                      : 'bg-gray-200 text-gray-600'
                  }`}
                >
                  {isDone ? <CheckCircle2 className="h-4 w-4" /> : idx + 1}
                </span>
                <span className={isCurrent ? 'text-primary-700 font-bold' : 'text-gray-500'}>
                  {s.label.split('. ')[1]}
                </span>
              </div>
            )
          })}
        </div>

        <Card className="p-6 sm:p-8 shadow-md">
          {error && (
            <div className="mb-5 flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
              <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* STEP 1: Registration Form */}
          {step === 'FORM' && (
            <form onSubmit={handleRegisterSubmit} className="space-y-4">
              <div>
                <Label htmlFor="orgName">
                  <span className="flex items-center gap-1.5">
                    <Building2 className="h-4 w-4 text-gray-500" />
                    Organization Name *
                  </span>
                </Label>
                <Input
                  id="orgName"
                  type="text"
                  value={orgName}
                  onChange={(e) => setOrgName(e.target.value)}
                  required
                  disabled={isLoading}
                  placeholder="Acme Corp"
                  className="mt-1"
                />
              </div>

              <div>
                <Label htmlFor="adminName">
                  <span className="flex items-center gap-1.5">
                    <User className="h-4 w-4 text-gray-500" />
                    Admin Full Name *
                  </span>
                </Label>
                <Input
                  id="adminName"
                  type="text"
                  value={adminName}
                  onChange={(e) => setAdminName(e.target.value)}
                  required
                  disabled={isLoading}
                  placeholder="Jane Doe"
                  className="mt-1"
                />
              </div>

              <div>
                <Label htmlFor="adminEmail">
                  <span className="flex items-center gap-1.5">
                    <Mail className="h-4 w-4 text-gray-500" />
                    Admin Email Address *
                  </span>
                </Label>
                <Input
                  id="adminEmail"
                  type="email"
                  value={adminEmail}
                  onChange={(e) => setAdminEmail(e.target.value)}
                  required
                  disabled={isLoading}
                  placeholder="admin@example.com"
                  className="mt-1"
                />
                <p className="text-xs text-gray-500 mt-1">
                  Used for initial verification and critical security alerts.
                </p>
              </div>

              <div>
                <PhoneInput
                  id="whatsappNumber"
                  label="WhatsApp Phone Number *"
                  value={whatsappNumber}
                  onChange={setWhatsappNumber}
                  required
                  disabled={isLoading}
                  hint="Primary WhatsApp number for alerts & notifications."
                />
              </div>

              <div className="pt-2">
                <Button
                  type="submit"
                  variant="primary"
                  size="lg"
                  className="w-full justify-center"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="h-5 w-5 animate-spin" />
                      Creating Organization...
                    </span>
                  ) : (
                    'Continue to Authenticator Setup'
                  )}
                </Button>
              </div>
            </form>
          )}

          {/* STEP 1.5: Verify Email (if required) */}
          {step === 'VERIFY_EMAIL' && (
            <form onSubmit={handleVerifyEmail} className="space-y-4">
              <div className="text-center py-2">
                <Mail className="h-10 w-10 text-primary-600 mx-auto mb-2" />
                <h3 className="text-lg font-bold text-gray-900">Check your inbox</h3>
                <p className="text-sm text-gray-600 mt-1">
                  We sent a verification token to <strong>{adminEmail}</strong>.
                </p>
              </div>

              <div>
                <Label htmlFor="emailToken">Verification Code or Token</Label>
                <Input
                  id="emailToken"
                  type="text"
                  value={emailToken}
                  onChange={(e) => setEmailToken(e.target.value)}
                  required
                  placeholder="Enter token from email"
                  className="mt-1 font-mono text-center tracking-wider text-base"
                />
              </div>

              <Button
                type="submit"
                variant="primary"
                size="lg"
                className="w-full justify-center"
                disabled={isLoading}
              >
                {isLoading ? (
                  <span className="flex items-center gap-2">
                    <Loader2 className="h-5 w-5 animate-spin" />
                    Verifying Email...
                  </span>
                ) : (
                  'Confirm Email & Setup TOTP'
                )}
              </Button>
            </form>
          )}

          {/* STEP 2: TOTP Setup */}
          {step === 'TOTP_SETUP' && totpData && (
            <div className="space-y-6">
              <div className="text-center">
                <div className="inline-flex p-2 bg-primary-100 text-primary-700 rounded-xl mb-2">
                  <Shield className="h-6 w-6" />
                </div>
                <h3 className="text-lg font-bold text-gray-900">Scan Authenticator QR Code</h3>
                <p className="text-xs text-gray-600 mt-1">
                  Open Google Authenticator, Authy, or 1Password and scan this QR code.
                </p>
              </div>

              <TotpQrCode
                secret={totpData.secret}
                otpauthUrl={totpData.otpauth_url}
                qrSvg={totpData.qr_code_svg}
                qrDataUrl={totpData.qr_code_data_url}
                issuer={orgName || 'WhatsApp Service'}
                accountName={adminEmail}
              />

              <div className="pt-2 border-t border-gray-200">
                <p className="text-center text-xs font-semibold text-gray-700 mb-3">
                  Enter the 6-digit code from your authenticator app:
                </p>

                <TotpCodeInput
                  value={totpCode}
                  onChange={setTotpCode}
                  onComplete={handleVerifyTotp}
                  disabled={isLoading}
                  error={!!error}
                />

                <div className="mt-5">
                  <Button
                    type="button"
                    variant="primary"
                    size="lg"
                    className="w-full justify-center"
                    disabled={isLoading || totpCode.length !== 6}
                    onClick={() => handleVerifyTotp()}
                  >
                    {isLoading ? (
                      <span className="flex items-center gap-2">
                        <Loader2 className="h-5 w-5 animate-spin" />
                        Verifying Code...
                      </span>
                    ) : (
                      'Verify & Activate Account'
                    )}
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* STEP 3: Backup Codes */}
          {step === 'BACKUP_CODES' && (
            <BackupCodesDisplay
              codes={backupCodes}
              title="Save Your Emergency Backup Codes"
              description="Keep these 10 one-time codes safe. You will need one if you ever lose your phone or authenticator access."
              showAcknowledgeCheckbox={true}
              acknowledgeLabel="Continue to Tenant Setup Wizard"
              onAcknowledge={handleFinishOnboarding}
            />
          )}
        </Card>
      </div>
    </div>
  )
}

export default SignupTenant
