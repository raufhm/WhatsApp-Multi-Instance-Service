import React, { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/hooks/useAuth'
import { onboardingApi, TENANT_ID_KEY } from '@/lib/apiClient'
import {
  KeyRound,
  ShieldAlert,
  Building2,
  User,
  Key,
  MessageSquare,
  AlertCircle,
  CheckCircle2,
  Loader2,
  ArrowLeft,
  LifeBuoy,
  Phone,
} from 'lucide-react'

type RecoveryTab = 'BACKUP_CODE' | 'REQUEST_RESET'

export const Recovery: React.FC = () => {
  const [tab, setTab] = useState<RecoveryTab>('BACKUP_CODE')
  const [tenantId, setTenantId] = useState(() => localStorage.getItem(TENANT_ID_KEY) || '')
  const [identifier, setIdentifier] = useState('')
  const [backupCode, setBackupCode] = useState('')

  // Request recovery state
  const [channel, setChannel] = useState<'whatsapp' | 'email'>('whatsapp')
  const [isRequested, setIsRequested] = useState(false)
  const [recoveryInstructions, setRecoveryInstructions] = useState<string | null>(null)

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { loginWithBackupCode } = useAuth()
  const navigate = useNavigate()

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
      setError('Backup code is required')
      setIsLoading(false)
      return
    }

    try {
      await loginWithBackupCode(tenantId.trim(), identifier.trim(), backupCode.trim())
      navigate({ to: '/' })
    } catch (err: any) {
      setError(err?.message || 'Invalid backup code or tenant identifier.')
    } finally {
      setIsLoading(false)
    }
  }

  const handleRequestRecoverySubmit = async (e: React.FormEvent) => {
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

    try {
      const res = await onboardingApi.requestRecovery({
        tenant_id: tenantId.trim(),
        identifier: identifier.trim(),
        channel,
      })
      setIsRequested(true)
      setRecoveryInstructions(
        res.instructions ||
          'Recovery alert sent. Your Tenant Administrator has been notified to reset your authenticator credentials via Team settings.'
      )
    } catch (err: any) {
      setError(err.response?.data?.error || err.message || 'Failed to request recovery. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md">
        <div className="text-center mb-6">
          <Link
            to="/login"
            className="inline-flex items-center gap-1 text-xs font-medium text-gray-500 hover:text-gray-700 mb-3"
          >
            <ArrowLeft className="h-3.5 w-3.5" /> Back to Sign In
          </Link>
          <div className="inline-flex items-center justify-center h-12 w-12 rounded-2xl bg-amber-500 text-white shadow-md mb-3">
            <KeyRound className="h-6 w-6" />
          </div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-gray-900 tracking-tight">
            Account Recovery
          </h1>
          <p className="mt-1 text-sm text-gray-600">
            Lost access to your authenticator app? Use a backup code or request an admin reset.
          </p>
        </div>

        {/* Tab Selection */}
        <div className="flex bg-gray-200/80 p-1 rounded-xl mb-6 text-xs font-medium">
          <button
            type="button"
            onClick={() => {
              setError(null)
              setTab('BACKUP_CODE')
            }}
            className={`flex-1 py-2 rounded-lg transition-all flex items-center justify-center gap-1.5 ${
              tab === 'BACKUP_CODE' ? 'bg-white text-gray-900 shadow-sm font-bold' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <Key className="h-4 w-4" />
            1. Use Backup Code
          </button>
          <button
            type="button"
            onClick={() => {
              setError(null)
              setTab('REQUEST_RESET')
            }}
            className={`flex-1 py-2 rounded-lg transition-all flex items-center justify-center gap-1.5 ${
              tab === 'REQUEST_RESET' ? 'bg-white text-gray-900 shadow-sm font-bold' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <LifeBuoy className="h-4 w-4" />
            2. Request Admin Reset
          </button>
        </div>

        <Card className="p-6 sm:p-8 shadow-md">
          {error && (
            <div className="mb-5 flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
              <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {tab === 'BACKUP_CODE' ? (
            <form onSubmit={handleBackupCodeSubmit} className="space-y-4">
              <div>
                <Label htmlFor="recTenantId">
                  <span className="flex items-center gap-1.5">
                    <Building2 className="h-4 w-4 text-gray-500" />
                    Tenant ID *
                  </span>
                </Label>
                <Input
                  id="recTenantId"
                  type="text"
                  value={tenantId}
                  onChange={(e) => setTenantId(e.target.value)}
                  required
                  placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
                  className="mt-1 font-mono text-sm"
                  disabled={isLoading}
                />
              </div>

              <div>
                <Label htmlFor="recIdentifier">
                  <span className="flex items-center gap-1.5">
                    <User className="h-4 w-4 text-gray-500" />
                    Email or WhatsApp Number *
                  </span>
                </Label>
                <Input
                  id="recIdentifier"
                  type="text"
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  required
                  placeholder="operator@example.com or +14155552671"
                  className="mt-1"
                  disabled={isLoading}
                />
              </div>

              <div>
                <Label htmlFor="recBackupCode">
                  <span className="flex items-center gap-1.5">
                    <Key className="h-4 w-4 text-gray-500" />
                    Single-Use Backup Code *
                  </span>
                </Label>
                <Input
                  id="recBackupCode"
                  type="text"
                  value={backupCode}
                  onChange={(e) => setBackupCode(e.target.value.toUpperCase())}
                  required
                  placeholder="e.g. A7B9-C2D4"
                  className="mt-1 font-mono text-center tracking-widest uppercase text-base"
                  disabled={isLoading}
                />
                <p className="text-xs text-gray-500 mt-1">
                  Enter any of the 10 backup codes generated when you joined.
                </p>
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
                    Logging in...
                  </span>
                ) : (
                  'Recover & Log In'
                )}
              </Button>
            </form>
          ) : isRequested ? (
            <div className="text-center py-4 space-y-4">
              <div className="inline-flex p-3 bg-green-100 text-green-700 rounded-full">
                <CheckCircle2 className="h-8 w-8" />
              </div>
              <h3 className="text-lg font-bold text-gray-900">Recovery Request Sent</h3>
              <p className="text-xs text-gray-600 leading-relaxed max-w-sm mx-auto">
                {recoveryInstructions}
              </p>

              <div className="p-4 bg-gray-50 border border-gray-200 rounded-lg text-left text-xs text-gray-600 space-y-1">
                <p className="font-semibold text-gray-800">What happens next?</p>
                <p>1. Your tenant administrator receives a reset notification.</p>
                <p>2. Once approved, you will receive a new TOTP setup link via WhatsApp.</p>
              </div>

              <Button
                type="button"
                variant="secondary"
                size="md"
                onClick={() => navigate({ to: '/login' })}
                className="w-full"
              >
                Return to Login
              </Button>
            </div>
          ) : (
            <form onSubmit={handleRequestRecoverySubmit} className="space-y-4">
              <div>
                <Label htmlFor="reqTenantId">
                  <span className="flex items-center gap-1.5">
                    <Building2 className="h-4 w-4 text-gray-500" />
                    Tenant ID *
                  </span>
                </Label>
                <Input
                  id="reqTenantId"
                  type="text"
                  value={tenantId}
                  onChange={(e) => setTenantId(e.target.value)}
                  required
                  placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
                  className="mt-1 font-mono text-sm"
                  disabled={isLoading}
                />
              </div>

              <div>
                <Label htmlFor="reqIdentifier">
                  <span className="flex items-center gap-1.5">
                    <User className="h-4 w-4 text-gray-500" />
                    Registered WhatsApp Number or Email *
                  </span>
                </Label>
                <Input
                  id="reqIdentifier"
                  type="text"
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  required
                  placeholder="+14155552671 or operator@example.com"
                  className="mt-1"
                  disabled={isLoading}
                />
              </div>

              <div>
                <Label>Notification Channel</Label>
                <div className="mt-1.5 grid grid-cols-2 gap-3">
                  <label
                    className={`flex items-center justify-center gap-2 p-2.5 rounded-lg border text-xs font-medium cursor-pointer transition-all ${
                      channel === 'whatsapp'
                        ? 'border-primary-600 bg-primary-50 text-primary-900 font-bold'
                        : 'border-gray-200 text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    <input
                      type="radio"
                      name="channel"
                      value="whatsapp"
                      checked={channel === 'whatsapp'}
                      onChange={() => setChannel('whatsapp')}
                      className="sr-only"
                    />
                    <Phone className="h-4 w-4" /> WhatsApp (Instant)
                  </label>

                  <label
                    className={`flex items-center justify-center gap-2 p-2.5 rounded-lg border text-xs font-medium cursor-pointer transition-all ${
                      channel === 'email'
                        ? 'border-primary-600 bg-primary-50 text-primary-900 font-bold'
                        : 'border-gray-200 text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    <input
                      type="radio"
                      name="channel"
                      value="email"
                      checked={channel === 'email'}
                      onChange={() => setChannel('email')}
                      className="sr-only"
                    />
                    <MessageSquare className="h-4 w-4" /> Email Fallback
                  </label>
                </div>
              </div>

              <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg text-xs text-amber-800 flex items-start gap-2">
                <ShieldAlert className="h-4 w-4 text-amber-600 flex-shrink-0 mt-0.5" />
                <span>
                  Tenant admins can immediately reset your TOTP in the Admin Team settings.
                </span>
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
                    Submitting Request...
                  </span>
                ) : (
                  'Request TOTP Reset'
                )}
              </Button>
            </form>
          )}
        </Card>
      </div>
    </div>
  )
}

export default Recovery
