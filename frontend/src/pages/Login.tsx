import React, { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { useAuth } from '@/hooks/useAuth'
import Button from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import TotpCodeInput from '@/components/ui/TotpCodeInput'
import { TENANT_ID_KEY, TENANT_SLUG_KEY, REMEMBER_ME_KEY } from '@/lib/apiClient'
import {
  Loader2,
  AlertCircle,
  Building2,
  User,
  QrCode,
  Key,
  HelpCircle,
  Lock,
  Check,
} from 'lucide-react'

type LoginMode = 'TOTP' | 'BACKUP_CODE'

interface LoginProps {
  /** When provided, renders compact form only (for use inside modals) */
  close?: () => void
}

export const Login: React.FC<LoginProps> = ({ close }) => {
  const [mode, setMode] = useState<LoginMode>('TOTP')
  const [tenantOrCompany, setTenantOrCompany] = useState(() => localStorage.getItem(TENANT_SLUG_KEY) || localStorage.getItem(TENANT_ID_KEY) || '')
  const [identifier, setIdentifier] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [backupCode, setBackupCode] = useState('')
  const [rememberMe, setRememberMe] = useState(() => localStorage.getItem(REMEMBER_ME_KEY) === 'true')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const { login, loginWithBackupCode } = useAuth()
  const navigate = useNavigate()

  const isModal = !!close

  const handleTotpSubmit = async (codeToUse?: string) => {
    const code = (codeToUse || totpCode).trim()
    setError(null)
    setIsLoading(true)

    if (!tenantOrCompany.trim()) {
      setError('Company or workspace name is required')
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
      await login(tenantOrCompany.trim(), identifier.trim(), code, rememberMe)
      if (close) {
        close()
      }
      navigate({ to: '/inbox' })
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

    if (!tenantOrCompany.trim()) {
      setError('Company or workspace name is required')
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
      await loginWithBackupCode(tenantOrCompany.trim(), identifier.trim(), backupCode.trim(), rememberMe)
      if (close) {
        close()
      }
      navigate({ to: '/inbox' })
    } catch (err: any) {
      setError(err?.message || 'Invalid backup code or tenant credentials.')
    } finally {
      setIsLoading(false)
    }
  }

  // ─── Login form (shared between standalone and modal) ──────────────
  const loginForm = (
    <div className={isModal ? 'space-y-5' : 'w-full max-w-md space-y-6'}>
      {!isModal && (
        <div className="lg:hidden text-center mb-6">
          <div className="inline-flex items-center justify-center h-12 w-12 rounded-2xl bg-primary-600 text-white shadow-lg mb-3">
            <span className="text-white text-2xl font-bold">w</span>
          </div>
          <h2 className="text-2xl font-bold text-gray-900">whops</h2>
          <p className="text-sm text-gray-600 mt-1">WhatsApp for your team</p>
        </div>
      )}

      <div>
        <h2 className={isModal ? 'text-xl font-bold text-gray-900 tracking-tight' : 'text-2xl sm:text-3xl font-extrabold text-gray-900 tracking-tight'}>
          {mode === 'TOTP' ? 'Sign in with Authenticator' : 'Sign in with Backup Code'}
        </h2>
        <p className="mt-1 text-sm text-gray-600">
          Welcome back! Sign in to access your team's WhatsApp workspace.
        </p>
      </div>

      <Card className={`${isModal ? 'p-5' : 'p-6 sm:p-8'} shadow-md`}>
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
              <Label htmlFor="tenantOrCompany">
                <span className="flex items-center gap-1.5">
                  <Building2 className="h-4 w-4 text-gray-500" />
                  Company / Workspace Name *
                </span>
              </Label>
              <Input
                id="tenantOrCompany"
                type="text"
                value={tenantOrCompany}
                onChange={(e) => setTenantOrCompany(e.target.value)}
                required
                disabled={isLoading}
                placeholder="e.g. acme-corp or Acme Corp"
                className="mt-1 text-sm"
              />
              <p className="text-xs text-gray-500 mt-1">
                Enter your organization's company name or workspace slug.
              </p>
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
                  autoFocus={!isModal}
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
              <Label htmlFor="backupTenantOrCompany">
                <span className="flex items-center gap-1.5">
                  <Building2 className="h-4 w-4 text-gray-500" />
                  Company / Workspace Name *
                </span>
              </Label>
              <Input
                id="backupTenantOrCompany"
                type="text"
                value={tenantOrCompany}
                onChange={(e) => setTenantOrCompany(e.target.value)}
                required
                disabled={isLoading}
                placeholder="e.g. acme-corp or Acme Corp"
                className="mt-1 text-sm"
              />
              <p className="text-xs text-gray-500 mt-1">
                Enter your organization's company name or workspace slug.
              </p>
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
  )

  // ─── Standalone full-page layout (for backward compatibility) ──────
  if (!isModal) {
    return (
      <div className="min-h-screen flex">
        {/* Left side - branding */}
        <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary-700 via-primary-800 to-primary-950 text-white flex-col justify-between p-12">
          <div>
            <div className="flex items-center space-x-3">
              <div className="h-11 w-11 rounded-2xl bg-white/20 flex items-center justify-center backdrop-blur-md shadow-inner">
                <span className="text-white text-2xl font-bold">w</span>
              </div>
              <div>
                <span className="text-xl font-bold tracking-tight">whops</span>
                <span className="block text-xs text-primary-200">WhatsApp for your team</span>
              </div>
            </div>
          </div>

          <div className="space-y-6 max-w-lg">
            <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/10 text-xs font-medium backdrop-blur-sm text-primary-100 border border-white/10">
              <QrCode className="h-4 w-4" />
              Scan QR code to connect
            </div>
            <h1 className="text-4xl font-extrabold leading-tight tracking-tight">
              Share your WhatsApp with your team. As simple as scanning a QR code.
            </h1>
            <p className="text-primary-100 text-base leading-relaxed">
              No Meta Business API setup required. Use your existing WhatsApp account and get built-in customer tracking, notes, and follow-ups — no separate CRM needed.
            </p>
            <ul className="space-y-2 mt-2">
              <li className="flex items-start gap-2 text-sm text-primary-50">
                <Check className="h-5 w-5 flex-shrink-0 mt-0.5" />
                <span>Scan QR once — entire team accesses the same WhatsApp number</span>
              </li>
              <li className="flex items-start gap-2 text-sm text-primary-50">
                <Check className="h-5 w-5 flex-shrink-0 mt-0.5" />
                <span>Track customer conversations with internal notes and follow-ups</span>
              </li>
              <li className="flex items-start gap-2 text-sm text-primary-50">
                <Check className="h-5 w-5 flex-shrink-0 mt-0.5" />
                <span>No developer or technical setup required</span>
              </li>
            </ul>
          </div>

          <div className="text-xs text-primary-300/80 border-t border-primary-700/50 pt-4">
            <p className="leading-relaxed">
              Using linked WhatsApp accounts may violate WhatsApp's Terms of Service. Run at your own risk.
              Not affiliated with or endorsed by Meta/WhatsApp.
            </p>
            <p className="mt-2">© {new Date().getFullYear()} whops</p>
          </div>
        </div>

        {/* Right side - login form */}
        <div className="flex-1 flex items-center justify-center p-4 sm:p-8 bg-gray-50">
          {loginForm}
        </div>
      </div>
    )
  }

  // ─── Modal / compact form only ────────────────────────────────────
  return loginForm
}

export default Login
