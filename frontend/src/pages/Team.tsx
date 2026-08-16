import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import PhoneInput from '@/components/ui/PhoneInput'
import { invitationsApi } from '@/lib/apiClient'
import { useAuth } from '@/hooks/useAuth'
import type { Invitation, Operator } from '@/types'
import {
  Users,
  UserPlus,
  Send,
  Trash2,
  RefreshCw,
  ShieldAlert,
  ShieldCheck,
  CheckCircle2,
  Clock,
  Mail,
  Phone,
  Copy,
  Check,
  Loader2,
  X,
  AlertCircle,
} from 'lucide-react'

export const Team: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'OPERATORS' | 'INVITATIONS'>('OPERATORS')

  // Operators state
  const [operators, setOperators] = useState<Operator[]>([])
  const [isLoadingOperators, setIsLoadingOperators] = useState(true)
  const [operatorsError, setOperatorsError] = useState<string | null>(null)

  // Invitations state
  const [invitations, setInvitations] = useState<Invitation[]>([])
  const [isLoadingInvitations, setIsLoadingInvitations] = useState(true)
  const [invitationsError, setInvitationsError] = useState<string | null>(null)

  // Inline action error (revoke invitation, etc.)
  const [actionError, setActionError] = useState<string | null>(null)

  // Invite modal / form state
  const [showInviteModal, setShowInviteModal] = useState(false)
  const [inviteChannel, setInviteChannel] = useState<'WHATSAPP' | 'EMAIL'>('WHATSAPP')
  const [invitePhone, setInvitePhone] = useState('')
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('OPERATOR')
  const [isSendingInvite, setIsSendingInvite] = useState(false)
  const [inviteError, setInviteError] = useState<string | null>(null)
  const [createdInvitation, setCreatedInvitation] = useState<Invitation | null>(null)
  const [copiedLink, setCopiedLink] = useState(false)

  // Reset TOTP Modal
  const [targetOperatorForReset, setTargetOperatorForReset] = useState<Operator | null>(null)
  const [isResettingTotp, setIsResettingTotp] = useState(false)
  const [resetSuccess, setResetSuccess] = useState<string | null>(null)
  const [resetError, setResetError] = useState<string | null>(null)

  const { user } = useAuth()
  const isAdmin = user?.role?.toUpperCase() === 'ADMIN'

  useEffect(() => {
    fetchOperators()
    fetchInvitations()
  }, [])

  const fetchOperators = async () => {
    setIsLoadingOperators(true)
    setOperatorsError(null)
    try {
      const data = await invitationsApi.listOperators()
      setOperators(data)
    } catch (err: any) {
      setOperatorsError(err?.response?.data?.error || err?.message || 'Failed to load operators')
      setOperators([])
    } finally {
      setIsLoadingOperators(false)
    }
  }

  const fetchInvitations = async () => {
    setIsLoadingInvitations(true)
    setInvitationsError(null)
    try {
      const data = await invitationsApi.listInvitations()
      setInvitations(data)
    } catch (err: any) {
      setInvitationsError(err?.response?.data?.error || err?.message || 'Failed to load invitations')
      setInvitations([])
    } finally {
      setIsLoadingInvitations(false)
    }
  }

  const handleSendInvite = async (e: React.FormEvent) => {
    e.preventDefault()
    setInviteError(null)
    setIsSendingInvite(true)

    try {
      let inv: Invitation
      if (inviteChannel === 'WHATSAPP') {
        if (!invitePhone.trim()) {
          setInviteError('WhatsApp phone number is required')
          setIsSendingInvite(false)
          return
        }
        inv = await invitationsApi.createWhatsAppInvitation(invitePhone.trim(), inviteRole)
      } else {
        if (!inviteEmail.trim()) {
          setInviteError('Email address is required')
          setIsSendingInvite(false)
          return
        }
        inv = await invitationsApi.createEmailInvitation(inviteEmail.trim(), inviteRole)
      }

      setCreatedInvitation(inv)
      setInvitations((prev) => [inv, ...prev])
      setInvitePhone('')
      setInviteEmail('')
    } catch (err: any) {
      setInviteError(err?.response?.data?.error || err?.message || 'Failed to send invitation')
    } finally {
      setIsSendingInvite(false)
    }
  }

  const handleRevokeInvitation = async (id: string) => {
    setActionError(null)
    try {
      await invitationsApi.revokeInvitation(id)
      setInvitations((prev) => prev.filter((i) => i.id !== id))
    } catch (err: any) {
      setActionError(err?.response?.data?.error || err?.message || 'Failed to revoke invitation')
    }
  }

  const handleResetTotp = async (operator: Operator) => {
    setIsResettingTotp(true)
    setResetError(null)
    try {
      await invitationsApi.resetOperatorTotp(operator.id)
      setResetSuccess(`TOTP reset successfully for ${operator.name}. They will receive a new setup link.`)
      setOperators((prev) =>
        prev.map((op) => (op.id === operator.id ? { ...op, totp_setup_required: true } : op))
      )
      setTimeout(() => {
        setTargetOperatorForReset(null)
        setResetSuccess(null)
      }, 2500)
    } catch (err: any) {
      setResetError(err?.response?.data?.error || err?.message || 'Failed to reset TOTP for this operator')
    } finally {
      setIsResettingTotp(false)
    }
  }

  const handleCopyLink = (url: string) => {
    navigator.clipboard.writeText(url)
    setCopiedLink(true)
    setTimeout(() => setCopiedLink(false), 2000)
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Users className="h-7 w-7 text-primary-600" />
            Team & Operator Management
          </h1>
          <p className="text-sm text-gray-600 mt-0.5">
            Manage operators, dispatch WhatsApp invitations, and oversee TOTP security status.
          </p>
        </div>

        <Button
          type="button"
          variant="primary"
          size="md"
          onClick={() => {
            setCreatedInvitation(null)
            setInviteError(null)
            setShowInviteModal(true)
          }}
          className="flex items-center gap-2 self-start sm:self-auto"
        >
          <UserPlus className="h-4 w-4" />
          <span>Invite Operator</span>
        </Button>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-6 text-sm font-medium">
          <button
            type="button"
            onClick={() => setActiveTab('OPERATORS')}
            className={`py-3 border-b-2 flex items-center gap-2 ${
              activeTab === 'OPERATORS'
                ? 'border-primary-600 text-primary-600 font-bold'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Users className="h-4 w-4" />
            Active Team Operators ({operators.length})
          </button>

          <button
            type="button"
            onClick={() => setActiveTab('INVITATIONS')}
            className={`py-3 border-b-2 flex items-center gap-2 ${
              activeTab === 'INVITATIONS'
                ? 'border-primary-600 text-primary-600 font-bold'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <Send className="h-4 w-4" />
            Invitations ({invitations.filter((i) => i.status === 'PENDING').length})
          </button>
        </nav>
      </div>

      {/* Inline action errors (revoke invite, etc.) */}
      {actionError && (
        <div className="flex items-start gap-2 p-3.5 rounded-lg bg-red-50 border border-red-200 text-sm text-red-700">
          <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
          <span>{actionError}</span>
        </div>
      )}

      {/* TAB 1: Operators */}
      {activeTab === 'OPERATORS' && (
        <Card className="overflow-hidden">
          {isLoadingOperators ? (
            <div className="py-12 flex justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
            </div>
          ) : operatorsError ? (
            <div className="text-center py-12 text-red-500">
              <AlertCircle className="h-12 w-12 mx-auto mb-3" />
              <p>{operatorsError}</p>
              <Button variant="primary" size="sm" className="mt-4" onClick={fetchOperators}>
                Retry
              </Button>
            </div>
          ) : operators.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <Users className="h-12 w-12 mx-auto text-gray-300 mb-2" />
              <p>No operators registered yet.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-sm">
                <thead className="bg-gray-50 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left">Operator</th>
                    <th scope="col" className="px-6 py-3 text-left">Role</th>
                    <th scope="col" className="px-6 py-3 text-left">TOTP Security</th>
                    <th scope="col" className="px-6 py-3 text-left">Last Login</th>
                    <th scope="col" className="px-6 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {operators.map((op) => (
                    <tr key={op.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-3">
                          <div className="h-9 w-9 rounded-full bg-primary-100 text-primary-700 font-bold flex items-center justify-center text-sm">
                            {op.name?.charAt(0) || 'U'}
                          </div>
                          <div>
                            <p className="font-semibold text-gray-900">{op.name}</p>
                            <p className="text-xs text-gray-500">{op.email || op.whatsapp_number}</p>
                          </div>
                        </div>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                          {op.role}
                        </span>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap">
                        {op.totp_setup_required ? (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-100 text-amber-800">
                            <Clock className="h-3 w-3" /> Setup Required
                          </span>
                        ) : op.totp_verified_at ? (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                            <ShieldCheck className="h-3 w-3" /> TOTP Active
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
                            Not configured
                          </span>
                        )}
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                        {op.last_login_at
                          ? new Date(op.last_login_at).toLocaleString()
                          : 'Never'}
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        {isAdmin && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => {
                              setResetError(null)
                              setResetSuccess(null)
                              setTargetOperatorForReset(op)
                            }}
                            className="text-xs text-amber-600 hover:text-amber-800 hover:bg-amber-50"
                          >
                            <RefreshCw className="h-3.5 w-3.5 mr-1" />
                            Reset TOTP
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {/* TAB 2: Invitations */}
      {activeTab === 'INVITATIONS' && (
        <Card className="overflow-hidden">
          {isLoadingInvitations ? (
            <div className="py-12 flex justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
            </div>
          ) : invitationsError ? (
            <div className="text-center py-12 text-red-500">
              <AlertCircle className="h-12 w-12 mx-auto mb-3" />
              <p>{invitationsError}</p>
              <Button variant="primary" size="sm" className="mt-4" onClick={fetchInvitations}>
                Retry
              </Button>
            </div>
          ) : invitations.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              <Send className="h-12 w-12 mx-auto text-gray-300 mb-2" />
              <p>No active or pending invitations.</p>
              <Button
                variant="secondary"
                size="sm"
                className="mt-3"
                onClick={() => setShowInviteModal(true)}
              >
                Send first invitation
              </Button>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-sm">
                <thead className="bg-gray-50 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                  <tr>
                    <th scope="col" className="px-6 py-3 text-left">Recipient</th>
                    <th scope="col" className="px-6 py-3 text-left">Channel</th>
                    <th scope="col" className="px-6 py-3 text-left">Role</th>
                    <th scope="col" className="px-6 py-3 text-left">Status</th>
                    <th scope="col" className="px-6 py-3 text-left">Created</th>
                    <th scope="col" className="px-6 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {invitations.map((inv) => (
                    <tr key={inv.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-2 font-mono text-xs font-semibold text-gray-900">
                          {inv.invitation_channel === 'WHATSAPP' ? (
                            <Phone className="h-4 w-4 text-green-600" />
                          ) : (
                            <Mail className="h-4 w-4 text-blue-600" />
                          )}
                          <span>{inv.identifier || inv.whatsapp_number || inv.email}</span>
                        </div>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap text-xs">
                        <span className="font-semibold text-gray-600">{inv.invitation_channel}</span>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700">
                          {inv.role}
                        </span>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap">
                        <span
                          className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            inv.status === 'PENDING'
                              ? 'bg-amber-100 text-amber-800'
                              : inv.status === 'ACCEPTED'
                              ? 'bg-green-100 text-green-800'
                              : 'bg-gray-100 text-gray-600'
                          }`}
                        >
                          {inv.status}
                          {inv.delivery_status && ` (${inv.delivery_status})`}
                        </span>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap text-xs text-gray-500">
                        {new Date(inv.created_at).toLocaleDateString()}
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap text-right">
                        {inv.status === 'PENDING' && (
                          <Button
                            type="button"
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRevokeInvitation(inv.id)}
                            className="text-xs text-red-600 hover:text-red-800 hover:bg-red-50"
                          >
                            <Trash2 className="h-3.5 w-3.5 mr-1" /> Revoke
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {/* MODAL: Invite Operator */}
      {showInviteModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50 animate-fadeIn">
          <div className="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl relative space-y-5">
            <button
              type="button"
              onClick={() => setShowInviteModal(false)}
              className="absolute top-4 right-4 p-1 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100"
            >
              <X className="h-5 w-5" />
            </button>

            <div className="flex items-center gap-3">
              <div className="p-3 bg-primary-100 text-primary-700 rounded-xl">
                <UserPlus className="h-6 w-6" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">Invite Team Operator</h3>
                <p className="text-xs text-gray-500">Zero-password WhatsApp or Email onboarding</p>
              </div>
            </div>

            {createdInvitation ? (
              <div className="space-y-4 py-2">
                <div className="p-4 bg-green-50 border border-green-200 rounded-xl flex items-center gap-3 text-xs text-green-800">
                  <CheckCircle2 className="h-5 w-5 text-green-600 flex-shrink-0" />
                  <span>
                    Invitation created for <strong>{createdInvitation.identifier}</strong>!
                  </span>
                </div>

                {createdInvitation.setup_url && (
                  <div>
                    <Label className="text-xs text-gray-600">Onboarding Setup Link:</Label>
                    <div className="mt-1 flex items-center gap-2">
                      <Input
                        readOnly
                        value={createdInvitation.setup_url}
                        className="font-mono text-xs bg-gray-50"
                      />
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        onClick={() => handleCopyLink(createdInvitation.setup_url!)}
                      >
                        {copiedLink ? <Check className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
                      </Button>
                    </div>
                  </div>
                )}

                <div className="pt-2 flex justify-end">
                  <Button
                    type="button"
                    variant="primary"
                    size="md"
                    onClick={() => {
                      setCreatedInvitation(null)
                      setShowInviteModal(false)
                    }}
                  >
                    Done
                  </Button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSendInvite} className="space-y-4">
                {inviteError && (
                  <div className="p-3 bg-red-50 border border-red-200 rounded-lg text-xs text-red-700">
                    {inviteError}
                  </div>
                )}

                {/* Channel tabs */}
                <div className="grid grid-cols-2 gap-2 p-1 bg-gray-100 rounded-lg text-xs font-medium">
                  <button
                    type="button"
                    onClick={() => setInviteChannel('WHATSAPP')}
                    className={`py-2 rounded-md flex items-center justify-center gap-1.5 transition-all ${
                      inviteChannel === 'WHATSAPP'
                        ? 'bg-white text-gray-900 shadow-sm font-bold'
                        : 'text-gray-600'
                    }`}
                  >
                    <Phone className="h-3.5 w-3.5 text-green-600" /> WhatsApp (Instant)
                  </button>

                  <button
                    type="button"
                    onClick={() => setInviteChannel('EMAIL')}
                    className={`py-2 rounded-md flex items-center justify-center gap-1.5 transition-all ${
                      inviteChannel === 'EMAIL'
                        ? 'bg-white text-gray-900 shadow-sm font-bold'
                        : 'text-gray-600'
                    }`}
                  >
                    <Mail className="h-3.5 w-3.5 text-blue-600" /> Email Fallback
                  </button>
                </div>

                {inviteChannel === 'WHATSAPP' ? (
                  <div>
                    <PhoneInput
                      id="teamInvitePhone"
                      label="WhatsApp Phone Number *"
                      value={invitePhone}
                      onChange={setInvitePhone}
                      required
                      placeholder="+14155552671"
                      hint="An invitation with TOTP setup link will be sent via WhatsApp."
                    />
                  </div>
                ) : (
                  <div>
                    <Label htmlFor="teamInviteEmail">Email Address *</Label>
                    <Input
                      id="teamInviteEmail"
                      type="email"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      required
                      placeholder="operator@company.com"
                      className="mt-1"
                    />
                  </div>
                )}

                <div>
                  <Label htmlFor="teamInviteRole">Operator Role</Label>
                  <select
                    id="teamInviteRole"
                    value={inviteRole}
                    onChange={(e) => setInviteRole(e.target.value)}
                    className="mt-1 block w-full pl-3 pr-10 py-2 text-sm border-gray-300 focus:outline-none focus:ring-primary-500 focus:border-primary-500 rounded-md border"
                  >
                    <option value="OPERATOR">Operator (Handle tickets & chats)</option>
                    <option value="ADMIN">Administrator (Full settings access)</option>
                    <option value="VIEWER">Viewer (Read-only)</option>
                  </select>
                </div>

                <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-200">
                  <Button
                    type="button"
                    variant="ghost"
                    size="md"
                    onClick={() => setShowInviteModal(false)}
                    disabled={isSendingInvite}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    variant="primary"
                    size="md"
                    disabled={isSendingInvite}
                    className="flex items-center gap-2"
                  >
                    {isSendingInvite ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Send className="h-4 w-4" />
                    )}
                    <span>Dispatch Invitation</span>
                  </Button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}

      {/* MODAL: Reset TOTP Confirmation */}
      {targetOperatorForReset && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50 animate-fadeIn">
          <div className="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl relative space-y-4">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-amber-100 text-amber-700 rounded-xl">
                <ShieldAlert className="h-6 w-6" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-gray-900">Reset Operator Authenticator</h3>
                <p className="text-xs text-gray-500">{targetOperatorForReset.name}</p>
              </div>
            </div>

            {resetSuccess ? (
              <div className="p-4 bg-green-50 border border-green-200 rounded-xl text-xs text-green-800 flex items-center gap-2 font-medium">
                <CheckCircle2 className="h-5 w-5 text-green-600 flex-shrink-0" />
                <span>{resetSuccess}</span>
              </div>
            ) : resetError ? (
              <div className="p-4 bg-red-50 border border-red-200 rounded-xl text-xs text-red-700 flex items-start gap-2">
                <AlertCircle className="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
                <span>{resetError}</span>
              </div>
            ) : (
              <>
                <p className="text-xs text-gray-600 leading-relaxed">
                  Are you sure you want to reset TOTP authentication for{' '}
                  <strong>{targetOperatorForReset.name}</strong>? Their current authenticator app and
                  backup codes will be revoked, and they will receive a new onboarding link via WhatsApp.
                </p>

                <div className="flex items-center justify-end gap-3 pt-3 border-t border-gray-200">
                  <Button
                    type="button"
                    variant="ghost"
                    size="md"
                    onClick={() => setTargetOperatorForReset(null)}
                    disabled={isResettingTotp}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="button"
                    variant="primary"
                    size="md"
                    onClick={() => handleResetTotp(targetOperatorForReset)}
                    disabled={isResettingTotp}
                    className="bg-amber-600 hover:bg-amber-700 text-white"
                  >
                    {isResettingTotp ? (
                      <span className="flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Resetting TOTP...
                      </span>
                    ) : (
                      'Confirm TOTP Reset'
                    )}
                  </Button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default Team
