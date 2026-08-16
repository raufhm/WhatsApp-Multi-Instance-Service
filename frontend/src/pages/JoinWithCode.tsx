import React, { useState } from 'react'
import { useNavigate, Link } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { onboardingApi } from '@/lib/apiClient'
import type { InvitationDetails } from '@/types'
import {
  KeyRound,
  UserCheck,
  ShieldCheck,
  AlertCircle,
  Loader2,
  ArrowRight,
  MessageCircle,
} from 'lucide-react'

export const JoinWithCode: React.FC = () => {
  const [code, setCode] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [invitationInfo, setInvitationInfo] = useState<InvitationDetails | null>(null)

  const navigate = useNavigate()

  const handleValidate = async (e: React.FormEvent) => {
    e.preventDefault()
    const trimmed = code.trim()
    if (!trimmed) {
      setError('Please enter your invitation code')
      return
    }

    setError(null)
    setIsLoading(true)

    try {
      const details = await onboardingApi.getInvitationDetails(trimmed)
      setInvitationInfo(details)
    } catch (err: any) {
      // If error from backend
      setError(
        err?.response?.data?.error?.message ||
          err?.message ||
          'Invalid or expired invitation code. Please check and try again.'
      )
      setInvitationInfo(null)
    } finally {
      setIsLoading(false)
    }
  }

  const handleProceed = () => {
    const trimmed = code.trim()
    if (trimmed) {
      navigate({ to: `/invitation/${encodeURIComponent(trimmed)}` as any })
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center">
        <div className="inline-flex items-center justify-center h-14 w-14 rounded-2xl bg-primary-600 text-white shadow-lg mb-4">
          <MessageCircle className="h-8 w-8" />
        </div>
        <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">
          Join with Invitation Code
        </h1>
        <p className="mt-2 text-sm text-gray-600 max-w-sm mx-auto">
          Enter the code you received via WhatsApp or Email to set up your TOTP-secured operator account.
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md px-4">
        <Card className="p-6 sm:p-8 shadow-sm">
          {!invitationInfo ? (
            <form onSubmit={handleValidate} className="space-y-5">
              {error && (
                <div className="p-4 bg-red-50 border border-red-200 rounded-xl flex items-center gap-3 text-xs text-red-700 animate-fadeIn">
                  <AlertCircle className="h-5 w-5 text-red-600 flex-shrink-0" />
                  <span>{error}</span>
                </div>
              )}

              <div>
                <Label htmlFor="invitationCode" className="text-sm font-semibold text-gray-700">
                  Invitation Code or Token
                </Label>
                <div className="mt-1 relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-gray-400">
                    <KeyRound className="h-5 w-5" />
                  </div>
                  <Input
                    id="invitationCode"
                    type="text"
                    value={code}
                    onChange={(e) => {
                      setCode(e.target.value)
                      if (error) setError(null)
                    }}
                    placeholder="Enter or paste your invitation code..."
                    className="pl-10 font-mono text-sm"
                    required
                    autoFocus
                  />
                </div>
                <p className="mt-1.5 text-xs text-gray-500">
                  You can find this code in your WhatsApp invitation message.
                </p>
              </div>

              <Button
                type="submit"
                variant="primary"
                size="md"
                disabled={isLoading || !code.trim()}
                className="w-full justify-center"
              >
                {isLoading ? (
                  <>
                    <Loader2 className="animate-spin h-4 w-4 mr-2" />
                    Validating Code...
                  </>
                ) : (
                  <>
                    <span>Validate Invitation</span>
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </>
                )}
              </Button>
            </form>
          ) : (
            <div className="space-y-5 animate-fadeIn">
              <div className="p-4 bg-green-50 border border-green-200 rounded-xl space-y-2">
                <div className="flex items-center gap-2 text-green-800 font-semibold text-sm">
                  <UserCheck className="h-5 w-5 text-green-600" />
                  <span>Invitation Verified!</span>
                </div>
                <div className="text-xs text-green-700 space-y-1 pl-7">
                  <p>
                    <strong>Workspace:</strong> {invitationInfo.tenant_name || 'Organization'}
                  </p>
                  <p>
                    <strong>Role:</strong> {invitationInfo.role || 'Operator'}
                  </p>
                  {(invitationInfo.whatsapp_number || invitationInfo.identifier) && (
                    <p>
                      <strong>Recipient:</strong> {invitationInfo.whatsapp_number || invitationInfo.identifier}
                    </p>
                  )}
                </div>
              </div>

              <p className="text-xs text-gray-600">
                Click below to complete your profile and set up your personal Two-Factor Authentication (TOTP).
              </p>

              <div className="flex flex-col gap-2">
                <Button
                  type="button"
                  variant="primary"
                  size="md"
                  onClick={handleProceed}
                  className="w-full justify-center"
                >
                  <span>Continue to Setup</span>
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>

                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setInvitationInfo(null)}
                  className="w-full justify-center text-gray-500"
                >
                  Enter a different code
                </Button>
              </div>
            </div>
          )}

          {/* Security Note */}
          <div className="mt-6 pt-5 border-t border-gray-100 flex items-center gap-2.5 text-xs text-gray-500">
            <ShieldCheck className="h-4 w-4 text-green-600 flex-shrink-0" />
            <span>Passwordless TOTP security guarantees direct, safe access.</span>
          </div>
        </Card>

        {/* Back and alternative options */}
        <div className="mt-6 flex flex-col items-center gap-2 text-sm text-gray-600 text-center">
          <div>
            Already completed setup?{' '}
            <Link to="/login" className="font-semibold text-primary-600 hover:text-primary-500 underline">
              Sign in here
            </Link>
          </div>
          <div>
            Need a new workspace?{' '}
            <Link to="/signup/tenant" className="font-semibold text-primary-600 hover:text-primary-500 underline">
              Register organization
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}

export default JoinWithCode
