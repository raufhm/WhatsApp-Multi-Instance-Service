import React, { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import Button from '@/components/ui/button'
import AssignmentModal from '@/components/inbox/AssignmentModal'
import { statusColors, formatDateTime } from './constants'
import type { Conversation } from '@/types'
import {
  useHandoffConversation,
  useCloseConversation,
  useReopenConversation,
  useDeleteConversation,
} from '@/hooks/useInbox'
import {
  UserPlus,
  Phone,
  CheckCircle2,
  RotateCcw,
  Loader2,
  Users,
  Mail,
  MapPin,
  Image as ImageIcon,
  FileText,
  MoreHorizontal,
  Trash2,
} from 'lucide-react'

interface ConversationDetailsProps {
  conversation: Conversation | null
  channelLabel?: string | null
}

const ConversationDetails: React.FC<ConversationDetailsProps> = ({ conversation, channelLabel }) => {
  const handoff = useHandoffConversation()
  const close = useCloseConversation()
  const reopen = useReopenConversation()
  const deleteConversation = useDeleteConversation()
  const navigate = useNavigate()
  const [isAssignOpen, setIsAssignOpen] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  if (!conversation) {
    return (
      <div className="h-full flex items-center justify-center p-6 text-gray-500 text-sm bg-white">
        Select a conversation to view details
      </div>
    )
  }

  const displayName = (conversation as any).contact_name || (conversation as any).contact_number || `Ticket #${conversation.ticket_number}`
  const initials = displayName.charAt(0).toUpperCase()
  const contactEmail = (conversation as any).contact_email || 'No email captured'
  const contactLocation = (conversation as any).contact_location || 'Location unknown'
  const attachmentExamples: { name: string; size: string; kind: string }[] = []

  return (
    <div className="h-full flex flex-col bg-white">
      <div className="flex h-16 shrink-0 items-center justify-between border-b border-gray-200/80 px-4">
        <button type="button" className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-white text-gray-800 shadow-sm" aria-label="Contact profile">
          <Users className="h-5 w-5" />
        </button>
        <button type="button" className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-white text-gray-800 shadow-sm" aria-label="More customer actions">
          <MoreHorizontal className="h-5 w-5" />
        </button>
      </div>

      <div className="border-b border-gray-200/80 px-5 py-7 text-center">
        <div className="mx-auto mb-4 flex h-24 w-24 items-center justify-center rounded-full bg-gradient-to-br from-cyan-200 via-orange-100 to-orange-500 text-3xl font-bold text-gray-950 shadow-sm">
          {initials}
        </div>
        <h3 className="truncate text-lg font-bold text-gray-950">{displayName}</h3>
        <div className="mt-2 space-y-1 text-xs text-gray-600">
          <p className="flex items-center justify-center gap-1.5">
            <Mail className="h-3.5 w-3.5" />
            <span className="truncate">{contactEmail}</span>
          </p>
          <p className="flex items-center justify-center gap-1.5">
            <MapPin className="h-3.5 w-3.5" />
            <span className="truncate">{contactLocation}</span>
          </p>
        </div>
      </div>

      <div className="border-b border-gray-200/80 p-4">
        <h4 className="mb-3 text-xs font-semibold uppercase tracking-wide text-gray-500">Conversation</h4>
        <div className="space-y-3 text-sm">
          {((conversation as any).contact_name || (conversation as any).contact_number) && (
            <div className="flex items-center justify-between gap-3">
              <span className="text-gray-500">{(conversation as any).is_group ? 'Group' : 'Contact'}</span>
              <span className="font-medium text-gray-900 text-right truncate max-w-[160px] flex items-center justify-end gap-1">
                {(conversation as any).is_group && <Users className="h-3.5 w-3.5 text-amber-600 shrink-0" />}
                <span className="truncate">{(conversation as any).contact_name || (conversation as any).contact_number}</span>
              </span>
            </div>
          )}
          {channelLabel && (
            <div className="flex items-center justify-between gap-3">
              <span className="text-gray-500">Channel</span>
              <span className="font-medium text-gray-900 text-right truncate max-w-[160px]">{channelLabel}</span>
            </div>
          )}
          <div className="flex items-center justify-between">
            <span className="text-gray-500">Ticket</span>
            <span className="font-medium text-gray-900">#{conversation.ticket_number}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-gray-500">Status</span>
            <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${statusColors[conversation.status] || 'bg-gray-100 text-gray-800'}`}>
              {conversation.status}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-gray-500">Assignee</span>
            <span className="font-medium text-gray-900">{conversation.assignee || 'Unassigned'}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-gray-500">Bot state</span>
            <span className="font-medium text-gray-900">{conversation.bot_state || 'None'}</span>
          </div>
        </div>
      </div>

      <div className="border-b border-gray-200/80 p-4">
        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">Actions</h4>
        <div className="grid grid-cols-1 gap-2">
          <Button
            variant="primary"
            size="sm"
            className="w-full justify-center rounded-full bg-orange-600 hover:bg-orange-700"
            onClick={() => setIsAssignOpen(true)}
          >
            <UserPlus className="h-4 w-4 mr-1" />
            Assign
          </Button>

          {conversation.status !== 'CLOSED' ? (
            <>
              <Button
                variant="secondary"
                size="sm"
                className="w-full justify-center rounded-full"
                onClick={() => handoff.mutate({ id: conversation.id })}
                disabled={handoff.isPending}
              >
                {handoff.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <Phone className="h-4 w-4 mr-1" />}
                Handoff
              </Button>
              <Button
                variant="danger"
                size="sm"
                className="w-full justify-center rounded-full"
                onClick={() => close.mutate({ id: conversation.id })}
                disabled={close.isPending}
              >
                {close.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <CheckCircle2 className="h-4 w-4 mr-1" />}
                Close
              </Button>
            </>
          ) : (
            <Button
              variant="primary"
              size="sm"
              className="w-full justify-center rounded-full bg-orange-600 hover:bg-orange-700"
              onClick={() => reopen.mutate({ id: conversation.id })}
              disabled={reopen.isPending}
            >
              {reopen.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <RotateCcw className="h-4 w-4 mr-1" />}
              Reopen
            </Button>
          )}
          <Button
            variant="danger"
            size="sm"
            className="w-full justify-center rounded-full"
            onClick={() => setShowDeleteConfirm(true)}
            disabled={deleteConversation.isPending}
          >
            {deleteConversation.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <Trash2 className="h-4 w-4 mr-1" />}
            Delete conversation
          </Button>
        </div>
      </div>

      <AssignmentModal
        isOpen={isAssignOpen}
        conversation={conversation}
        onClose={() => setIsAssignOpen(false)}
      />

      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm">
          <div className="w-full max-w-sm rounded-2xl bg-white p-5 shadow-2xl">
            <div className="mb-4 flex h-11 w-11 items-center justify-center rounded-full bg-red-100 text-red-600">
              <Trash2 className="h-5 w-5" />
            </div>
            <h3 className="text-base font-bold text-gray-900">Delete conversation?</h3>
            <p className="mt-2 text-sm leading-relaxed text-gray-600">
              This removes the conversation and its messages permanently. Operators submit a deletion request for admin approval.
            </p>
            <div className="mt-5 flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                size="sm"
                onClick={() => setShowDeleteConfirm(false)}
                disabled={deleteConversation.isPending}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant="danger"
                size="sm"
                onClick={() => {
                  deleteConversation.mutate({ id: conversation.id }, {
                    onSuccess: (result) => {
                      setShowDeleteConfirm(false)
                      if (result?.deleted) {
                        navigate({ to: '/inbox' })
                      } else {
                        window.alert('Deletion requested. An admin must approve and complete it.')
                      }
                    },
                  })
                }}
                disabled={deleteConversation.isPending}
              >
                {deleteConversation.isPending ? 'Submitting…' : 'Confirm'}
              </Button>
            </div>
          </div>
        </div>
      )}

      <div className="border-b border-gray-200/80 p-4">
        <div className="mb-3 flex items-center justify-between">
          <h4 className="flex items-center gap-2 text-sm font-semibold text-gray-700">
            <ImageIcon className="h-4 w-4" />
            Images
          </h4>
          <span className="text-xs text-gray-400">0</span>
        </div>
        <div className="grid grid-cols-3 gap-2">
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              key={index}
              className="aspect-square rounded-xl border border-dashed border-gray-200 bg-[#f8f4ef]"
            />
          ))}
        </div>
      </div>

      <div className="border-b border-gray-200/80 p-4">
        <div className="mb-3 flex items-center justify-between">
          <h4 className="flex items-center gap-2 text-sm font-semibold text-gray-700">
            <FileText className="h-4 w-4" />
            Files
          </h4>
          <span className="text-xs text-gray-400">{attachmentExamples.length}</span>
        </div>
        <div className="space-y-2">
          {attachmentExamples.length === 0 ? (
            <div className="rounded-xl border border-dashed border-gray-200 bg-[#f8f4ef] px-3 py-4 text-center text-xs text-gray-500">
              No files yet
            </div>
          ) : (
            attachmentExamples.map((file) => (
              <div key={file.name} className="grid grid-cols-[2rem_1fr_auto] items-center gap-2 rounded-xl bg-[#f0edea] px-2 py-2">
                <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-white text-[10px] font-bold text-orange-600">
                  {file.kind}
                </span>
                <span className="min-w-0">
                  <span className="block truncate text-xs font-semibold text-gray-800">{file.name}</span>
                  <span className="block text-[11px] text-gray-500">{file.size}</span>
                </span>
                <MoreHorizontal className="h-4 w-4 text-gray-400" />
              </div>
            ))
          )}
        </div>
      </div>

      <div className="p-4">
        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">Timestamps</h4>
        <div className="space-y-2 text-sm">
          <div className="flex items-center justify-between">
            <span className="text-gray-500">Started</span>
            <span className="text-gray-900">{formatDateTime(conversation.started_at)}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-gray-500">Last activity</span>
            <span className="text-gray-900">{formatDateTime(conversation.last_activity_at)}</span>
          </div>
          {conversation.closed_at && (
            <div className="flex items-center justify-between">
              <span className="text-gray-500">Closed</span>
              <span className="text-gray-900">{formatDateTime(conversation.closed_at)}</span>
            </div>
          )}
          {conversation.handoff_at && (
            <div className="flex items-center justify-between">
              <span className="text-gray-500">Handed off</span>
              <span className="text-gray-900">{formatDateTime(conversation.handoff_at)}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ConversationDetails
