import React, { useState, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import PhoneInput from '@/components/ui/PhoneInput'
import { PairingModal } from '@/components/ui/PairingModal'
import { setupApi, invitationsApi } from '@/lib/apiClient'
import { useAuth } from '@/hooks/useAuth'
import type { TenantSetupStatus } from '@/types'
import {
  Building2,
  Smartphone,
  Users,
  CheckCircle2,
  ArrowRight,
  ArrowLeft,
  Loader2,
  Sparkles,
  QrCode,
  Globe,
  Clock,
  Briefcase,
  AlertCircle,
  Send,
} from 'lucide-react'

const STEPS = [
  { id: 1, title: 'Business Profile', icon: Building2 },
  { id: 2, title: 'WhatsApp Connection', icon: Smartphone },
  { id: 3, title: 'Invite Team', icon: Users },
  { id: 4, title: 'Launch', icon: Sparkles },
]

export const SetupWizard: React.FC = () => {
  const [currentStep, setCurrentStep] = useState(1)
  const [status, setStatus] = useState<TenantSetupStatus | null>(null)

  // Step 1: Business Profile
  const [orgName, setOrgName] = useState('')
  const [businessType, setBusinessType] = useState('Customer Support')
  const [timezone, setTimezone] = useState('UTC+0 (London, GMT)')
  const [supportHours, setSupportHours] = useState('09:00 - 18:00 (Mon - Fri)')
  const [website, setWebsite] = useState('')

  // Step 2: WhatsApp Connection state
  const [pairedInstance, setPairedInstance] = useState(false)
  const [pairedPhone, setPairedPhone] = useState<string>('')
  const [isPairingOpen, setIsPairingOpen] = useState(false)

  // Step 3: Invite team state
  const [invitePhone, setInvitePhone] = useState('')
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('OPERATOR')
  const [invitesSent, setInvitesSent] = useState<string[]>([])
  const [isInviting, setIsInviting] = useState(false)

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [statusError, setStatusError] = useState<string | null>(null)

  const { user } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    fetchSetupStatus()
  }, [])

  const fetchSetupStatus = async () => {
    setStatusError(null)
    try {
      const data = await setupApi.getStatus()
      setStatus(data)
      if (data.organization_details?.name) {
        setOrgName(data.organization_details.name)
      }
      if (data.organization_details?.business_type) {
        setBusinessType(data.organization_details.business_type)
      }
      if (data.current_step && data.current_step > 1 && data.current_step <= 4) {
        setCurrentStep(data.current_step)
      }
    } catch (err: any) {
      setStatusError(err?.response?.data?.error || err?.message || 'Failed to load setup status')
    }
  }

  const handleStep1Submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    try {
      await setupApi.updateSetup({
        step: 2,
        business_type: businessType,
        timezone,
        support_hours: supportHours,
        website,
      })
      setCurrentStep(2)
    } catch (err: any) {
      setError(err?.response?.data?.error || err?.message || 'Failed to save step 1')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSendInvite = async () => {
    if (!invitePhone.trim() && !inviteEmail.trim()) {
      setError('Please provide a WhatsApp phone number or email address')
      return
    }

    setError(null)
    setIsInviting(true)

    try {
      if (invitePhone.trim()) {
        await invitationsApi.createWhatsAppInvitation(invitePhone.trim(), inviteRole)
        setInvitesSent((prev) => [...prev, `WhatsApp: ${invitePhone} (${inviteRole})`])
        setInvitePhone('')
      } else if (inviteEmail.trim()) {
        await invitationsApi.createEmailInvitation(inviteEmail.trim(), inviteRole)
        setInvitesSent((prev) => [...prev, `Email: ${inviteEmail} (${inviteRole})`])
        setInviteEmail('')
      }
    } catch (err: any) {
      setError(err?.response?.data?.error || err?.message || 'Failed to send invitation')
    } finally {
      setIsInviting(false)
    }
  }

  const handleCompleteSetup = async () => {
    setError(null)
    setIsLoading(true)

    try {
      await setupApi.completeSetup()
      navigate({ to: '/' })
    } catch (err: any) {
      setError(err?.response?.data?.error || err?.message || 'Failed to complete setup')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center h-12 w-12 rounded-2xl bg-primary-600 text-white shadow-md mb-3">
            <Sparkles className="h-6 w-6" />
          </div>
          <h1 className="text-3xl font-extrabold text-gray-900 tracking-tight">
            Tenant Setup Wizard
          </h1>
          <p className="mt-1.5 text-sm text-gray-600">
            {user ? `Welcome, ${user.name}! ` : ''}Let's get your WhatsApp Multi-Instance workspace configured and ready.
            {status?.is_setup_complete ? ' (Completed)' : ''}
          </p>
        </div>

        {/* Multi-step progress indicator */}
        <div className="mb-8">
          <div className="flex items-center justify-between relative">
            <div className="absolute left-0 top-1/2 -translate-y-1/2 h-0.5 bg-gray-200 w-full -z-0" />
            {STEPS.map((step) => {
              const StepIcon = step.icon
              const isCompleted = step.id < currentStep
              const isCurrent = step.id === currentStep

              return (
                <div key={step.id} className="relative z-10 flex flex-col items-center bg-gray-50 px-2">
                  <div
                    className={`h-10 w-10 rounded-full flex items-center justify-center transition-all ${
                      isCompleted
                        ? 'bg-green-600 text-white'
                        : isCurrent
                        ? 'bg-primary-600 text-white ring-4 ring-primary-100 font-bold'
                        : 'bg-white border-2 border-gray-300 text-gray-400'
                    }`}
                  >
                    {isCompleted ? <CheckCircle2 className="h-5 w-5" /> : <StepIcon className="h-5 w-5" />}
                  </div>
                  <span
                    className={`mt-2 text-xs font-medium ${
                      isCurrent ? 'text-primary-700 font-bold' : isCompleted ? 'text-green-700' : 'text-gray-500'
                    }`}
                  >
                    {step.title}
                  </span>
                </div>
              )
            })}
          </div>
        </div>

        <Card className="p-6 sm:p-8 shadow-md">
          {statusError && (
            <div className="mb-6 flex items-start justify-between gap-4 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
              <div className="flex items-start gap-2">
                <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
                <span>{statusError}</span>
              </div>
              <Button variant="primary" size="sm" onClick={fetchSetupStatus} disabled={isLoading}>
                Retry
              </Button>
            </div>
          )}
          {error && (
            <div className="mb-6 flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
              <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* STEP 1: Organization & Business Details */}
          {currentStep === 1 && (
            <form onSubmit={handleStep1Submit} className="space-y-5">
              <div>
                <h2 className="text-xl font-bold text-gray-900">Step 1: Business Profile Details</h2>
                <p className="text-xs text-gray-500 mt-1">
                  Configure primary operational settings for your tenant instance.
                </p>
              </div>

              <div className="space-y-4">
                <div>
                  <Label htmlFor="wizOrgName">
                    <span className="flex items-center gap-1.5">
                      <Building2 className="h-4 w-4 text-gray-500" /> Organization Name
                    </span>
                  </Label>
                  <Input
                    id="wizOrgName"
                    type="text"
                    value={orgName}
                    onChange={(e) => setOrgName(e.target.value)}
                    required
                    placeholder="Acme Global"
                    className="mt-1"
                  />
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="wizBusinessType">
                      <span className="flex items-center gap-1.5">
                        <Briefcase className="h-4 w-4 text-gray-500" /> Business Type
                      </span>
                    </Label>
                    <select
                      id="wizBusinessType"
                      value={businessType}
                      onChange={(e) => setBusinessType(e.target.value)}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-sm border-gray-300 focus:outline-none focus:ring-primary-500 focus:border-primary-500 rounded-md border"
                    >
                      <option value="Customer Support">Customer Support</option>
                      <option value="E-Commerce / Retail">E-Commerce / Retail</option>
                      <option value="Healthcare / Clinic">Healthcare / Clinic</option>
                      <option value="Financial Services">Financial Services</option>
                      <option value="Logistics & Delivery">Logistics & Delivery</option>
                      <option value="Other">Other</option>
                    </select>
                  </div>

                  <div>
                    <Label htmlFor="wizTimezone">
                      <span className="flex items-center gap-1.5">
                        <Clock className="h-4 w-4 text-gray-500" /> Timezone
                      </span>
                    </Label>
                    <select
                      id="wizTimezone"
                      value={timezone}
                      onChange={(e) => setTimezone(e.target.value)}
                      className="mt-1 block w-full pl-3 pr-10 py-2 text-sm border-gray-300 focus:outline-none focus:ring-primary-500 focus:border-primary-500 rounded-md border"
                    >
                      <option value="UTC+0 (London, GMT)">UTC+0 (London, GMT)</option>
                      <option value="UTC-5 (New York, EST)">UTC-5 (New York, EST)</option>
                      <option value="UTC-8 (San Francisco, PST)">UTC-8 (San Francisco, PST)</option>
                      <option value="UTC+1 (Berlin, CET)">UTC+1 (Berlin, CET)</option>
                      <option value="UTC+7 (Jakarta, WIB)">UTC+7 (Jakarta, WIB)</option>
                      <option value="UTC+8 (Singapore, SGT)">UTC+8 (Singapore, SGT)</option>
                    </select>
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="wizSupportHours">Support Operating Hours</Label>
                    <Input
                      id="wizSupportHours"
                      type="text"
                      value={supportHours}
                      onChange={(e) => setSupportHours(e.target.value)}
                      placeholder="09:00 - 18:00 (Mon - Fri)"
                      className="mt-1"
                    />
                  </div>

                  <div>
                    <Label htmlFor="wizWebsite">
                      <span className="flex items-center gap-1.5">
                        <Globe className="h-4 w-4 text-gray-500" /> Website (Optional)
                      </span>
                    </Label>
                    <Input
                      id="wizWebsite"
                      type="url"
                      value={website}
                      onChange={(e) => setWebsite(e.target.value)}
                      placeholder="https://example.com"
                      className="mt-1"
                    />
                  </div>
                </div>
              </div>

              <div className="pt-4 flex justify-end">
                <Button type="submit" variant="primary" size="md" disabled={isLoading} className="group">
                  <span>Save & Continue</span>
                  <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
                </Button>
              </div>
            </form>
          )}

          {/* STEP 2: WhatsApp Connection Guidance */}
          {currentStep === 2 && (
            <div className="space-y-6">
              <div>
                <h2 className="text-xl font-bold text-gray-900">Step 2: Connect WhatsApp Instance</h2>
                <p className="text-xs text-gray-500 mt-1">
                  Pair your WhatsApp Business or personal number with our multi-instance backend engine.
                </p>
              </div>

              <div className="bg-gray-50 border border-gray-200 rounded-xl p-5 space-y-4">
                <div className="flex items-start gap-4">
                  <div className="p-3 bg-green-100 text-green-700 rounded-xl">
                    <QrCode className="h-8 w-8" />
                  </div>
                  <div>
                    <h3 className="text-sm font-bold text-gray-900">How to Pair a WhatsApp Device</h3>
                    <ol className="mt-1.5 text-xs text-gray-600 space-y-1 list-decimal list-inside">
                      <li>Open WhatsApp on your mobile phone</li>
                      <li>Go to <strong>Settings</strong> &gt; <strong>Linked Devices</strong></li>
                      <li>Tap <strong>Link a Device</strong> and scan the terminal QR code or use the API</li>
                    </ol>
                  </div>
                </div>

                <div className="p-4 bg-white border border-gray-200 rounded-lg flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <div
                      className={`h-3.5 w-3.5 rounded-full flex-shrink-0 ${
                        pairedInstance ? 'bg-green-500 animate-pulse' : 'bg-amber-400'
                      }`}
                    />
                    <div>
                      <p className="text-xs font-semibold text-gray-800">
                        {pairedInstance
                          ? `WhatsApp Account Connected (${pairedPhone || 'Online'})`
                          : 'Instance Ready for Pairing'}
                      </p>
                      <p className="text-[11px] text-gray-500">
                        {pairedInstance
                          ? 'Device is active and handling incoming chats'
                          : 'Pair right now or link additional accounts anytime from the Accounts page'}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2 self-end sm:self-auto">
                    <Button
                      type="button"
                      variant="primary"
                      size="sm"
                      onClick={() => setIsPairingOpen(true)}
                      className="gap-1.5"
                    >
                      <QrCode className="h-4 w-4" />
                      <span>{pairedInstance ? 'Pair Another Device' : 'Pair Device Now'}</span>
                    </Button>
                    {!pairedInstance && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => setPairedInstance(true)}
                        className="text-xs text-gray-500"
                      >
                        Skip for now
                      </Button>
                    )}
                  </div>
                </div>
              </div>

              <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                <Button type="button" variant="ghost" size="md" onClick={() => setCurrentStep(1)}>
                  <ArrowLeft className="mr-2 h-4 w-4" /> Back
                </Button>
                <Button
                  type="button"
                  variant="primary"
                  size="md"
                  onClick={() => setCurrentStep(3)}
                  className="group"
                >
                  <span>Continue to Team Setup</span>
                  <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
                </Button>
              </div>
            </div>
          )}

          {/* STEP 3: Team Invitations */}
          {currentStep === 3 && (
            <div className="space-y-6">
              <div>
                <h2 className="text-xl font-bold text-gray-900">Step 3: Invite Team Members</h2>
                <p className="text-xs text-gray-500 mt-1">
                  Send WhatsApp or Email setup invitations with instant TOTP onboarding.
                </p>
              </div>

              <div className="bg-gray-50 border border-gray-200 rounded-xl p-5 space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <PhoneInput
                    id="wizInvitePhone"
                    label="WhatsApp Number (Recommended)"
                    value={invitePhone}
                    onChange={setInvitePhone}
                    placeholder="+14155552671"
                  />

                  <div>
                    <Label htmlFor="wizInviteEmail">Email Fallback</Label>
                    <Input
                      id="wizInviteEmail"
                      type="email"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      placeholder="operator@example.com"
                      className="mt-1.5"
                    />
                  </div>
                </div>

                <div className="flex items-center justify-between pt-2">
                  <div className="flex items-center gap-2">
                    <Label htmlFor="wizRole" className="text-xs text-gray-600">
                      Role:
                    </Label>
                    <select
                      id="wizRole"
                      value={inviteRole}
                      onChange={(e) => setInviteRole(e.target.value)}
                      className="text-xs border-gray-300 rounded border px-2 py-1 bg-white"
                    >
                      <option value="OPERATOR">Operator</option>
                      <option value="ADMIN">Admin</option>
                      <option value="VIEWER">Viewer</option>
                    </select>
                  </div>

                  <Button
                    type="button"
                    variant="secondary"
                    size="sm"
                    onClick={handleSendInvite}
                    disabled={isInviting || (!invitePhone.trim() && !inviteEmail.trim())}
                  >
                    {isInviting ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <>
                        <Send className="h-3.5 w-3.5 mr-1.5" />
                        Send Invite
                      </>
                    )}
                  </Button>
                </div>
              </div>

              {invitesSent.length > 0 && (
                <div className="p-3 bg-green-50 border border-green-200 rounded-lg text-xs space-y-1">
                  <p className="font-semibold text-green-800">Invitations dispatched:</p>
                  {invitesSent.map((inv, idx) => (
                    <div key={idx} className="text-green-700 flex items-center gap-1.5">
                      <CheckCircle2 className="h-3.5 w-3.5" /> {inv}
                    </div>
                  ))}
                </div>
              )}

              <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                <Button type="button" variant="ghost" size="md" onClick={() => setCurrentStep(2)}>
                  <ArrowLeft className="mr-2 h-4 w-4" /> Back
                </Button>
                <Button
                  type="button"
                  variant="primary"
                  size="md"
                  onClick={() => setCurrentStep(4)}
                  className="group"
                >
                  <span>Review & Complete</span>
                  <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
                </Button>
              </div>
            </div>
          )}

          {/* STEP 4: Completion & Launch */}
          {currentStep === 4 && (
            <div className="text-center py-6 space-y-6">
              <div className="inline-flex items-center justify-center h-16 w-16 rounded-full bg-green-100 text-green-600 shadow-inner">
                <CheckCircle2 className="h-10 w-10" />
              </div>

              <div>
                <h2 className="text-2xl font-extrabold text-gray-900">Setup Ready to Complete!</h2>
                <p className="text-sm text-gray-600 mt-1 max-w-md mx-auto">
                  Your organization <strong>{orgName}</strong> is configured with passwordless TOTP security
                  and is ready to manage WhatsApp traffic.
                </p>
              </div>

              <div className="bg-gray-50 border border-gray-200 rounded-xl p-5 text-left text-xs space-y-2 max-w-md mx-auto text-gray-700">
                <div className="flex justify-between">
                  <span className="text-gray-500">Business Type:</span>
                  <span className="font-semibold">{businessType}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Operating Timezone:</span>
                  <span className="font-semibold">{timezone}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Security Model:</span>
                  <span className="font-semibold text-green-700">TOTP (Zero-Password)</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-500">Team Invites:</span>
                  <span className="font-semibold">{invitesSent.length} sent</span>
                </div>
              </div>

              <div className="pt-4 max-w-md mx-auto">
                <Button
                  type="button"
                  variant="primary"
                  size="lg"
                  className="w-full justify-center shadow-md"
                  onClick={handleCompleteSetup}
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <span className="flex items-center gap-2">
                      <Loader2 className="h-5 w-5 animate-spin" />
                      Finalizing Setup...
                    </span>
                  ) : (
                    'Complete Setup & Open Dashboard'
                  )}
                </Button>
              </div>
            </div>
          )}
        </Card>
      </div>

      <PairingModal
        isOpen={isPairingOpen}
        onClose={() => setIsPairingOpen(false)}
        onSuccess={(phone) => {
          setPairedInstance(true)
          setPairedPhone(phone)
        }}
      />
    </div>
  )
}

export default SetupWizard
