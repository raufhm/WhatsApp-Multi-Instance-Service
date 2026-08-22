import React, { useState } from 'react'
import Button from '@/components/ui/button'
import AssignmentModal from '@/components/inbox/AssignmentModal'
import { statusColors, formatDateTime } from './constants'
import type { Conversation } from '@/types'
import {
  useHandoffConversation,
  useCloseConversation,
  useReopenConversation,
} from '@/hooks/useInbox'
import { UserPlus, Phone, CheckCircle2, RotateCcw, Loader2, Users } from 'lucide-react'

interface ConversationDetailsProps {
  conversation: Conversation | null
  channelLabel?: string | null
}

const ConversationDetails: React.FC<ConversationDetailsProps> = ({ conversation, channelLabel }) => {
  const handoff = useHandoffConversation()
  const close = useCloseConversation()
  const reopen = useReopenConversation()
  const [isAssignOpen, setIsAssignOpen] = useState(false)

  if (!conversation) {
    return (
      <div className="h-full flex items-center justify-center p-6 text-gray-500 text-sm">
        Select a conversation to view details
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b border-gray-200 bg-gradient-to-b from-white to-gray-50/70">
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Conversation</h3>
          <div className="space-y-3 text-sm">
            {((conversation as any).contact_name || (conversation as any).contact_number) && (
              <div className="flex items-center justify-between">
                <span className="text-gray-500">{(conversation as any).is_group ? 'Group' : 'Contact'}</span>
                <span className="font-medium text-gray-900 text-right truncate max-w-[160px] flex items-center justify-end gap-1">
                {(conversation as any).is_group && <Users className="h-3.5 w-3.5 text-amber-600 shrink-0" />}
                <span className="truncate">{(conversation as any).contact_name || (conversation as any).contact_number}</span>
                </span>
              </div>
            )}
            {channelLabel && (
              <div className="flex items-center justify-between">
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

      <div className="p-4 border-b border-gray-200">
        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">Actions</h4>
        <div className="space-y-2">
          <Button
            variant="primary"
            size="sm"
            className="w-full justify-center"
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
                className="w-full justify-center"
                onClick={() => handoff.mutate({ id: conversation.id })}
                disabled={handoff.isPending}
              >
                {handoff.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <Phone className="h-4 w-4 mr-1" />}
                Handoff
              </Button>
              <Button
                variant="danger"
                size="sm"
                className="w-full justify-center"
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
              className="w-full justify-center"
              onClick={() => reopen.mutate({ id: conversation.id })}
              disabled={reopen.isPending}
            >
              {reopen.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1" /> : <RotateCcw className="h-4 w-4 mr-1" />}
              Reopen
            </Button>
          )}
        </div>
      </div>

      <AssignmentModal
        isOpen={isAssignOpen}
        conversation={conversation}
        onClose={() => setIsAssignOpen(false)}
      />

      <div className="p-4">
        <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">Timestamps</h4>
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
