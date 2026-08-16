import React, { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'
import Button from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import TotpCodeInput from '@/components/ui/TotpCodeInput'
import { TENANT_ID_KEY, REMEMBER_ME_KEY } from '@/lib/apiClient'
import {
  Loader2,
  MessageCircle,
  AlertCircle,
  Building2,
  User,
  ShieldCheck,
  Key,
  HelpCircle,
  Lock,
} from 'lucide-react'

type LoginMode = 'TOTP' | 'BACKUP_CODE'

export const Login: React.FC = () => {
  const [mode, setMode] = useState<LoginMode>('TOTP')
  const [tenantId, setTenantId] = useState(() => localStorage.getItem(TENANT_ID_KEY) || '')
  const [identifier, setIdentifier] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [backupCode, setBackupCode] = useState('')
  const [rememberMe, setRememberMe] = useState(() => localStorage.getItem(REMEMBER_ME_KEY) === 'true')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const { login, loginWithBackupCode } = useAuth()
  const navigate = useNavigate()

  const handleTotpSubmit = async (codeToUse?: string) => {
    const code = (codeToUse || totpCode).trim()
    setError(null)
    setIsLoading(true)

    if (!tenantId.trim()) {
      setError('Tenant ID is required')
      setIsLoading(false)
      return
    }
    if (!identifier.trim()) {
      setError('Email or WhatsApp phone number is required')
      setIsLoading(false)
      return
    }
    if (code.length !== 6) {
      setError('Please enter the 6-digit authenticator code')
      setIsLoading(false)
      return
    }

    try {
      await login(tenantId.trim(), identifier.trim(), code, rememberMe)
      navigate({ to: '/' })
    } catch (err: any) {
      setError(err?.message || 'Invalid credentials or authenticator code.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleBackupCodeSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    if (!tenantId.trim()) {
      setError('Tenant ID is required')
      setIsLoading(false)
      return
    }
    if (!identifier.trim()) {
      setError('Email or WhatsApp phone number is required')
      setIsLoading(false)
      return
    }
    if (!backupCode.trim()) {
      setError('Single-use backup code is required')
      setIsLoading(false)
      return
    }

    try {
      await loginWithBackupCode(tenantId.trim(), identifier.trim(), backupCode.trim(), rememberMe)
      navigate({ to: '/' })
    } catch (err: any) {
      setError(err?.message || 'Invalid backup code or tenant credentials.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex">
      {/* Left side - branding */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary-700 via-primary-800 to-primary-950 text-white flex-col justify-between p-12">
        <div>
          <div className="flex items-center space-x-3">
            <div className="h-11 w-11 rounded-2xl bg-white/20 flex items-center justify-center backdrop-blur-md shadow-inner">
              <MessageCircle className="h-7 w-7 text-white" />
            </div>
            <div>
              <span className="text-xl font-bold tracking-tight">WhatsApp Operator Dashboard</span>
              <span className="block text-xs text-primary-200">Passwordless Multi-Instance Platform</span>
            </div>
          </div>
        </div>

        <div className="space-y-6 max-w-lg">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/10 text-xs font-medium backdrop-blur-sm text-primary-100 border border-white/10">
            <ShieldCheck className="h-4 w-4 text-green-300" />
            Zero-Password TOTP Authentication
          </div>
          <h1 className="text-4xl font-extrabold leading-tight tracking-tight">
            Secure, scalable WhatsApp workspace for your entire team.
          </h1>
          <p className="text-primary-100 text-base leading-relaxed">
            Manage multi-instance conversations, automated bot rules, handoffs, and operator accounts with
            industry-grade security.
          </p>
        </div>

        <div className="text-xs text-primary-300 flex items-center justify-between border-t border-primary-700/50 pt-4">
          <span>© {new Date().getFullYear()} WhatsApp Multi-Instance Service</span>
          <span>End-to-End Encrypted Session</span>
        </div>
      </div>

      {/* Right side - login form */}
      <div className="flex-1 flex items-center justify-center p-4 sm:p-8 bg-gray-50">
        <div className="w-full max-w-md space-y-6">
          <div className="lg:hidden text-center mb-6">
            <div className="inline-flex items-center justify-center h-12 w-12 rounded-2xl bg-primary-600 text-white shadow-lg mb-3">
              <MessageCircle className="h-7 w-7" />
            </div>
            <h2 className="text-2xl font-bold text-gray-900">Operator Dashboard</h2>
          </div>

          <div>
            <h2 className="text-2xl sm:text-3xl font-extrabold text-gray-900 tracking-tight">
              {mode === 'TOTP' ? 'Sign in with Authenticator' : 'Sign in with Backup Code'}
            </h2>
            <p className="mt-1.5 text-sm text-gray-600">
              {mode === 'TOTP'
                ? 'Enter your Tenant ID, identifier, and 6-digit TOTP code.'
                : 'Enter your Tenant ID, identifier, and a single-use backup code.'}
            </p>
          </div>

          <Card className="p-6 sm:p-8 shadow-md">
            {error && (
              <div className="mb-5 flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
                <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
                <span>{error}</span>
              </div>
            )}

            {mode === 'TOTP' ? (
              <form
                className="space-y-4"
                onSubmit={(e) => {
                  e.preventDefault()
                  handleTotpSubmit()
                }}
              >
                <div>
                  <Label htmlFor="tenantId">
                    <span className="flex items-center gap-1.5">
                      <Building2 className="h-4 w-4 text-gray-500" />
                      Tenant ID *
                    </span>
                  </Label>
                  <Input
                    id="tenantId"
                    type="text"
                    value={tenantId}
                    onChange={(e) => setTenantId(e.target.value)}
                    required
                    disabled={isLoading}
                    placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
                    className="mt-1 font-mono text-sm"
                  />
                </div>

                <div>
                  <Label htmlFor="identifier">
                    <span className="flex items-center gap-1.5">
                      <User className="h-4 w-4 text-gray-500" />
                      Email or WhatsApp Number *
                    </span>
                  </Label>
                  <Input
                    id="identifier"
                    type="text"
                    value={identifier}
                    onChange={(e) => setIdentifier(e.target.value)}
                    required
                    disabled={isLoading}
                    placeholder="operator@example.com or +14155552671"
                    className="mt-1"
                  />
                </div>

                <div>
                  <Label htmlFor="totp-input-0">
                    <span className="flex items-center gap-1.5">
                      <Lock className="h-4 w-4 text-gray-500" />
                      6-Digit Authenticator Code *
                    </span>
                  </Label>
                  <div className="mt-2 flex justify-center">
                    <TotpCodeInput
                      id="totp-input"
                      value={totpCode}
                      onChange={setTotpCode}
                      onComplete={(code) => handleTotpSubmit(code)}
                      disabled={isLoading}
                      error={!!error}
                      autoFocus={false}
                    />
                  </div>
                  <p className="text-xs text-gray-500 text-center mt-2">
                    Open Google Authenticator, 1Password, or Authy on your phone.
                  </p>
                </div>

                <div className="flex items-center justify-between text-sm pt-1">
                  <label className="flex items-center gap-2 cursor-pointer select-none text-xs text-gray-600">
                    <input
                      type="checkbox"
                      checked={rememberMe}
                      onChange={(e) => setRememberMe(e.target.checked)}
                      className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span>Remember me (30 days)</span>
                  </label>

                  <button
                    type="button"
                    onClick={() => {
                      setError(null)
                      setMode('BACKUP_CODE')
                    }}
                    className="text-xs font-semibold text-primary-600 hover:text-primary-700"
                  >
                    Use a backup code instead
                  </button>
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="lg"
                  className="w-full justify-center mt-2"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="h-5 w-5 animate-spin" />
                      Signing in...
                    </span>
                  ) : (
                    'Sign in'
                  )}
                </Button>
              </form>
            ) : (
              <form className="space-y-4" onSubmit={handleBackupCodeSubmit}>
                <div>
                  <Label htmlFor="backupTenantId">
                    <span className="flex items-center gap-1.5">
                      <Building2 className="h-4 w-4 text-gray-500" />
                      Tenant ID *
                    </span>
                  </Label>
                  <Input
                    id="backupTenantId"
                    type="text"
                    value={tenantId}
                    onChange={(e) => setTenantId(e.target.value)}
                    required
                    disabled={isLoading}
                    placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
                    className="mt-1 font-mono text-sm"
                  />
                </div>

                <div>
                  <Label htmlFor="backupIdentifier">
                    <span className="flex items-center gap-1.5">
                      <User className="h-4 w-4 text-gray-500" />
                      Email or WhatsApp Number *
                    </span>
                  </Label>
                  <Input
                    id="backupIdentifier"
                    type="text"
                    value={identifier}
                    onChange={(e) => setIdentifier(e.target.value)}
                    required
                    disabled={isLoading}
                    placeholder="operator@example.com or +14155552671"
                    className="mt-1"
                  />
                </div>

                <div>
                  <Label htmlFor="backupCode">
                    <span className="flex items-center gap-1.5">
                      <Key className="h-4 w-4 text-gray-500" />
                      Single-Use Backup Code *
                    </span>
                  </Label>
                  <Input
                    id="backupCode"
                    type="text"
                    value={backupCode}
                    onChange={(e) => setBackupCode(e.target.value.toUpperCase())}
                    required
                    disabled={isLoading}
                    placeholder="e.g. A7B9-C2D4"
                    className="mt-1 font-mono tracking-widest text-center uppercase text-base"
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    Enter one of your 10 recovery codes generated during setup.
                  </p>
                </div>

                <div className="flex items-center justify-between text-sm pt-1">
                  <label className="flex items-center gap-2 cursor-pointer select-none text-xs text-gray-600">
                    <input
                      type="checkbox"
                      checked={rememberMe}
                      onChange={(e) => setRememberMe(e.target.checked)}
                      className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span>Remember me (30 days)</span>
                  </label>

                  <button
                    type="button"
                    onClick={() => {
                      setError(null)
                      setMode('TOTP')
                    }}
                    className="text-xs font-semibold text-primary-600 hover:text-primary-700"
                  >
                    Use authenticator code instead
                  </button>
                </div>

                <Button
                  type="submit"
                  variant="primary"
                  size="lg"
                  className="w-full justify-center mt-2"
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="h-5 w-5 animate-spin" />
                      Verifying Backup Code...
                    </span>
                  ) : (
                    'Sign in with Backup Code'
                  )}
                </Button>
              </form>
            )}
          </Card>

          {/* Links and Help */}
          <div className="space-y-3 text-center text-xs text-gray-600">
            <div>
              <Link
                to="/recovery"
                className="inline-flex items-center gap-1 text-primary-600 hover:text-primary-700 font-medium"
              >
                <HelpCircle className="h-3.5 w-3.5" />
                Lost authenticator access? Recover account
              </Link>
            </div>

            <div className="pt-2 border-t border-gray-200 flex flex-col gap-1.5">
              <div>
                Have an invitation code?{' '}
                <Link to="/join" className="font-semibold text-primary-600 hover:text-primary-700 underline">
                  Join with Code
                </Link>
              </div>
              <div>
                Need a new workspace?{' '}
                <Link to="/signup" className="font-semibold text-primary-600 hover:text-primary-700">
                  Register new organization
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Login
