import React, { useState, useRef, useLayoutEffect, useMemo } from 'react'
import { useParams, useNavigate } from '@tanstack/react-router'
import {
  useConversation,
  useSendMessage,
  useUploadMedia,
  useHandoffConversation,
  useCloseConversation,
  useReopenConversation,
  useAddInternalNote,
} from '@/hooks/useInbox'
import { useDashboardAccounts } from '@/hooks/useDashboard'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
import AssignmentModal from '@/components/inbox/AssignmentModal'
import {
  Loader2,
  ArrowLeft,
  Send,
  UserPlus,
  Phone,
  CheckCircle2,
  RotateCcw,
  StickyNote,
  Paperclip,
  X,
  FileText,
  AlertCircle,
  Check,
  CheckCheck,
  Clock,
  Users,
  Reply,
} from 'lucide-react'

const statusColors: Record<string, string> = {
  OPEN: 'bg-green-100 text-green-800',
  BOT_ACTIVE: 'bg-blue-100 text-blue-800',
  WAITING: 'bg-yellow-100 text-yellow-800',
  HANDED_OFF: 'bg-purple-100 text-purple-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

const formatTime = (iso: string) =>
  new Date(iso).toLocaleString(undefined, { hour: '2-digit', minute: '2-digit' })

const formatSenderPhone = (address?: string | null) => {
  if (!address) return ''
  const value = address.split('@')[0].split(':')[0]
  return value ? `+${value.replace(/^\+/, '')}` : ''
}

const resolveMediaUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('/api/v1/media/')) {
    return `/dashboard/api/media/${url.slice('/api/v1/media/'.length)}`
  }
  return url
}

const MESSAGE_BATCH_SIZE = 50

interface AttachedMedia {
  file: File
  media_key?: string
  media_url?: string
  mime_type: string
  previewUrl?: string
  progress: number
  isUploading: boolean
  error?: string
}

const ConversationDetail: React.FC = () => {
  const { id } = useParams({ strict: false })
  const navigate = useNavigate()
  const [messageLimit, setMessageLimit] = useState(MESSAGE_BATCH_SIZE)
  const { data, isLoading, isError, isFetching } = useConversation(id!, { limit: messageLimit })
  const sendMessage = useSendMessage(id)
  const uploadMedia = useUploadMedia()
  const handoff = useHandoffConversation()
  const close = useCloseConversation()
  const reopen = useReopenConversation()
  const addNote = useAddInternalNote()
  const { data: accountsData } = useDashboardAccounts()

  const [reply, setReply] = useState('')
  const [note, setNote] = useState('')
  const [showNote, setShowNote] = useState(false)
  const [isAssignOpen, setIsAssignOpen] = useState(false)
  const [attachedMedia, setAttachedMedia] = useState<AttachedMedia | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)
  const [replyTo, setReplyTo] = useState<{ id: string; content: string } | null>(null)
  const [reactions, setReactions] = useState<Record<string, string>>({})
  const timelineRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const preserveScrollRef = useRef<{ top: number; height: number } | null>(null)

  const conversation = data?.conversation
  const messages = data?.messages || []
  const accounts = Array.isArray(accountsData) ? accountsData : []
  const channelLabel = useMemo(() => {
    if (!conversation?.account_id) return null
    return accounts.find((account: any) => account.id === conversation.account_id)?.display_name || conversation.account_id
  }, [accounts, conversation?.account_id])

  useLayoutEffect(() => {
    const el = timelineRef.current
    if (!el) return

    const snapshot = preserveScrollRef.current
    if (snapshot) {
      const delta = el.scrollHeight - snapshot.height
      el.scrollTop = snapshot.top + delta
      preserveScrollRef.current = null
      return
    }

    el.scrollTop = el.scrollHeight
  }, [messages.length])

  const loadOlderMessages = () => {
    if (isFetching || messages.length < messageLimit) return
    const el = timelineRef.current
    if (!el) return
    preserveScrollRef.current = { top: el.scrollTop, height: el.scrollHeight }
    setMessageLimit((current) => current + MESSAGE_BATCH_SIZE)
  }

  const handleTimelineScroll = () => {
    const el = timelineRef.current
    if (!el) return
    if (el.scrollTop <= 32) {
      loadOlderMessages()
    }
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
      </div>
    )
  }

  if (isError || !conversation) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-600">Conversation not found.</p>
        <Button variant="secondary" className="mt-4" onClick={() => navigate({ to: '/inbox' })}>
          Back to inbox
        </Button>
      </div>
    )
  }

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    let previewUrl = ''
    if (file.type.startsWith('image/')) {
      previewUrl = URL.createObjectURL(file)
    }

    const initialAttachment: AttachedMedia = {
      file,
      mime_type: file.type || 'application/octet-stream',
      previewUrl,
      progress: 0,
      isUploading: true,
    }
    setAttachedMedia(initialAttachment)

    try {
      const res = await uploadMedia.mutateAsync({
        file,
        onProgress: (pct) => {
          setAttachedMedia((prev) => (prev ? { ...prev, progress: pct } : null))
        },
      })
      setAttachedMedia((prev) =>
        prev
          ? {
              ...prev,
              media_key: res.media_key,
              media_url: res.media_url,
              mime_type: res.mime_type || prev.mime_type,
              progress: 100,
              isUploading: false,
            }
          : null
      )
    } catch (err: any) {
      setAttachedMedia((prev) =>
        prev ? { ...prev, isUploading: false, error: err?.response?.data?.error || 'Failed to upload media' } : null
      )
    } finally {
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  const handleRemoveAttachment = () => {
    if (attachedMedia?.previewUrl) {
      URL.revokeObjectURL(attachedMedia.previewUrl)
    }
    setAttachedMedia(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleReaction = (message: (typeof messages)[number], emoji: string) => {
    const target = message.provider_message_id
    if (!target || conversation.is_group || sendMessage.isPending) return
    const current = reactions[target]
    const next = current === emoji ? '' : emoji
    setReactions((previous) => {
      const updated = { ...previous }
      if (next) updated[target] = next
      else delete updated[target]
      return updated
    })
    sendMessage.mutate({
      account: conversation.account_id,
      recipient: conversation.contact_number || '',
      message: next,
      type: 'REACTION',
      reaction_target: target,
      is_group: false,
    }, {
      onError: () => setReactions((previous) => {
        const updated = { ...previous }
        if (current) updated[target] = current
        else delete updated[target]
        return updated
      }),
    })
  }

  const handleReply = (message: (typeof messages)[number]) => {
    setReplyTo({ id: message.id, content: message.content })
    document.getElementById('conversation-reply')?.focus()
  }

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault()
    const text = reply.trim()
    if (!text && !attachedMedia?.media_key) return
    if (attachedMedia?.isUploading) return

    let msgType = 'TEXT'
    let mediaKey = undefined
    if (attachedMedia?.media_key) {
      mediaKey = attachedMedia.media_key
      const mime = attachedMedia.mime_type
      if (mime.startsWith('image/')) msgType = 'IMAGE'
      else if (mime.startsWith('video/')) msgType = 'VIDEO'
      else if (mime.startsWith('audio/')) msgType = 'AUDIO'
      else msgType = 'DOCUMENT'
    }

    const currentReply = reply
    const currentAttachment = attachedMedia
    setReply('')
    handleRemoveAttachment()
    setSendError(null)

    sendMessage.mutate(
      {
        account: conversation.account_id,
        recipient: conversation.contact_number || '',
        message: text || (currentAttachment ? currentAttachment.file.name : ''),
        type: msgType,
        media_key: mediaKey,
        is_group: !!conversation.is_group,
      },
      {
        onError: (err) => {
          const e = err as any
          setReply(currentReply)
          setSendError(e?.response?.data?.error || 'Failed to send message')
        },
      }
    )
    setReplyTo(null)
  }

  const handleNote = (e: React.FormEvent) => {
    e.preventDefault()
    const text = note.trim()
    if (!text) return
    addNote.mutate(
      { id: conversation.id, content: text },
      { onSuccess: () => { setNote(''); setShowNote(false) } }
    )
  }

  const displayName = conversation.contact_name || conversation.contact_number || `Ticket #${conversation.ticket_number}`

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => navigate({ to: '/inbox' })}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-xl font-semibold text-gray-900 flex items-center gap-2">
              {displayName}
              {conversation.is_group && (
                <span className="inline-flex items-center gap-0.5 text-xs font-medium text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full">
                  <Users className="h-3.5 w-3.5" />
                  Group
                </span>
              )}
              {channelLabel && (
                <span className="inline-flex items-center gap-0.5 text-xs font-medium text-sky-600 bg-sky-50 px-2 py-0.5 rounded-full">
                  <Phone className="h-3.5 w-3.5" />
                  {channelLabel}
                </span>
              )}
            </h1>
            <p className="text-[13px] text-gray-500">
              {conversation.contact_number ? `${conversation.contact_number} • ` : ''}
              {conversation.assignee ? `Assigned to ${conversation.assignee}` : 'Unassigned'}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[conversation.status] || 'bg-gray-100 text-gray-800'}`}>
            {conversation.status}
          </span>
          {conversation.status !== 'CLOSED' ? (
            <>
              <Button variant="secondary" size="sm" onClick={() => setIsAssignOpen(true)}>
                <UserPlus className="h-4 w-4 mr-1" /> Assign
              </Button>
              <Button variant="secondary" size="sm" onClick={() => handoff.mutate({ id: conversation.id })}>
                <Phone className="h-4 w-4 mr-1" /> Handoff
              </Button>
              <Button variant="danger" size="sm" onClick={() => close.mutate({ id: conversation.id })}>
                <CheckCircle2 className="h-4 w-4 mr-1" /> Close
              </Button>
            </>
          ) : (
            <Button variant="primary" size="sm" onClick={() => reopen.mutate({ id: conversation.id })}>
              <RotateCcw className="h-4 w-4 mr-1" /> Reopen
            </Button>
          )}
        </div>
      </div>

      <AssignmentModal
        isOpen={isAssignOpen}
        conversation={conversation}
        onClose={() => setIsAssignOpen(false)}
      />

      {/* Timeline */}
      <Card className="flex-1 flex flex-col overflow-hidden">
        <div ref={timelineRef} onScroll={handleTimelineScroll} className="flex-1 overflow-y-auto p-4 space-y-2">
          {messages.length === messageLimit && (
            <div className="flex justify-center py-1">
              <button
                type="button"
                onClick={loadOlderMessages}
                className="text-[11px] font-medium text-primary-600 hover:text-primary-700"
                disabled={isFetching}
              >
                {isFetching ? 'Loading older messages...' : 'Load older messages'}
              </button>
            </div>
          )}
          {messages.length === 0 && (
            <div className="text-center text-gray-400 py-8">No messages yet</div>
          )}
          {messages.map((m) => {
            const resolvedUrl = resolveMediaUrl(m.media_url)
            return (
              <div
                key={m.id}
                className={`group/message flex ${m.actor === 'CONTACT' || m.direction === 'INCOMING' ? 'justify-start' : 'justify-end'}`}
              >
                <div
                  className={`relative max-w-[70%] px-4 py-2 rounded-lg ${
                    m.is_internal
                      ? 'bg-gray-100 text-gray-600 border border-dashed border-gray-300'
                      : m.actor === 'CONTACT' || m.direction === 'INCOMING'
                        ? 'bg-white border border-gray-200'
                        : 'bg-primary-600 text-white'
                  }`}
                  role={m.is_internal ? 'note' : undefined}
                >
                  {m.is_internal && (
                    <p className="text-[10px] uppercase tracking-wide text-gray-400 mb-1">Internal note</p>
                  )}
                  {conversation.is_group && (m.sender_address || m.direction === 'INCOMING') && !m.is_internal && (
                    <p className="text-[10px] font-semibold text-amber-700 mb-0.5">
                      {formatSenderPhone(m.sender_address) || 'Unknown sender'}
                    </p>
                  )}
                  {(m.actor === 'OPERATOR' || m.actor === 'BOT' || m.direction === 'OUTGOING') && !m.is_internal && (
                    <span
                      className="text-[10px] font-semibold text-primary-200 block mb-0.5"
                      title={m.operator_name ? `Sent by ${m.operator_name}` : `Sent by ${m.actor === 'BOT' ? 'Bot' : 'Operator'}`}
                    >
                      {m.operator_name || (m.actor === 'BOT' ? 'Bot' : 'Operator')}
                    </span>
                  )}
                  {m.message_type === 'IMAGE' && resolvedUrl ? (
                    <div className="mb-2">
                      <img
                        src={resolvedUrl}
                        alt="Attachment"
                        className="rounded-md max-h-60 max-w-full object-contain cursor-pointer"
                        loading="lazy"
                        onClick={() => window.open(resolvedUrl, '_blank')}
                      />
                    </div>
                  ) : m.message_type === 'VIDEO' && resolvedUrl ? (
                    <div className="mb-2">
                      <video
                        src={resolvedUrl}
                        controls
                        className="rounded-md max-h-60 max-w-full"
                      />
                    </div>
                  ) : m.message_type === 'AUDIO' && resolvedUrl ? (
                    <div className="mb-2">
                      <audio
                        src={resolvedUrl}
                        controls
                        className="max-w-full my-1"
                      />
                    </div>
                  ) : resolvedUrl ? (
                    <div className="mb-2">
                      <a
                        href={resolvedUrl}
                        target="_blank"
                        rel="noreferrer"
                        className={`inline-flex items-center gap-1.5 text-sm font-medium underline ${
                          m.actor === 'OPERATOR' || m.direction === 'OUTGOING' ? 'text-white' : 'text-primary-600'
                        }`}
                      >
                        <FileText className="h-4 w-4" />
                        View attachment
                      </a>
                    </div>
                  ) : null}
                  <p className="text-sm">{m.content}</p>
                  {reactions[m.provider_message_id] && (
                    <span className="absolute -bottom-2 left-2 rounded-full bg-white border border-gray-200 px-1.5 py-0.5 text-xs shadow-sm">
                      {reactions[m.provider_message_id]}
                    </span>
                  )}
                  {!conversation.is_group && !m.is_internal && m.provider_message_id && (
                    <div className="absolute -top-9 right-1 z-10 hidden group-hover/message:flex items-center gap-0.5 rounded-lg border border-gray-200 bg-white p-1 shadow-md">
                      {['👍', '❤️', '😂', '😮', '😢', '🙏'].map((emoji) => (
                        <button
                          key={emoji}
                          type="button"
                          className="rounded p-1 text-sm hover:bg-gray-100"
                          title={reactions[m.provider_message_id] === emoji ? 'Remove reaction' : `React ${emoji}`}
                          onClick={() => handleReaction(m, emoji)}
                        >
                          {emoji}
                        </button>
                      ))}
                      <button
                        type="button"
                        className="rounded p-1 text-gray-500 hover:bg-gray-100"
                        title="Reply"
                        onClick={() => handleReply(m)}
                      >
                        <Reply className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  )}
                  <div className="flex items-center justify-end gap-1 mt-1">
                    <span
                      className={`text-[10px] ${
                        m.actor === 'OPERATOR' || (m.direction === 'OUTGOING' && !m.is_internal)
                          ? 'text-primary-100'
                          : 'text-gray-400'
                      }`}
                    >
                      {formatTime(m.provider_timestamp || m.created_at)}
                    </span>
                    {(m.actor === 'OPERATOR' || m.direction === 'OUTGOING') && !m.is_internal && (
                      <span
                        className="inline-flex items-center ml-0.5"
                        title={
                          m.status === 'PENDING'
                            ? 'Pending'
                            : m.status === 'READ'
                              ? 'Read'
                              : m.status === 'DELIVERED'
                                ? 'Delivered'
                                : m.status === 'FAILED'
                                  ? 'Failed'
                                  : 'Sent'
                        }
                      >
                        {m.status === 'PENDING' ? (
                          <Clock className="h-3 w-3 text-primary-200" />
                        ) : m.status === 'READ' ? (
                          <CheckCheck className="h-3.5 w-3.5 text-sky-300" />
                        ) : m.status === 'DELIVERED' ? (
                          <CheckCheck className="h-3.5 w-3.5 text-primary-200" />
                        ) : m.status === 'FAILED' ? (
                          <AlertCircle className="h-3 w-3 text-red-300" />
                        ) : (
                          <Check className="h-3.5 w-3.5 text-primary-200" />
                        )}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            )
          })}
        </div>

        {/* Internal note composer */}
        {showNote && (
          <form onSubmit={handleNote} className="p-3 border-t border-gray-200">
            <textarea
              className="form-control w-full mb-2"
              placeholder="Internal note (not sent to customer)"
              rows={2}
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
            <div className="flex justify-end gap-2">
              <Button variant="secondary" size="sm" onClick={() => setShowNote(false)}>
                Cancel
              </Button>
              <Button variant="secondary" size="sm" type="submit" disabled={!note.trim()}>
                Save note
              </Button>
            </div>
          </form>
        )}

        {/* Attached media preview bar */}
        {attachedMedia && (
          <div className="p-2 px-3 border-t border-gray-200 bg-gray-50 flex items-center justify-between gap-3 text-sm">
            <div className="flex items-center gap-2 min-w-0">
              {attachedMedia.previewUrl ? (
                <img src={attachedMedia.previewUrl} alt="Preview" className="h-10 w-10 object-cover rounded shrink-0" />
              ) : (
                <Paperclip className="h-5 w-5 text-gray-500 shrink-0" />
              )}
              <div className="truncate">
                <p className="font-medium text-gray-800 truncate">{attachedMedia.file.name}</p>
                <p className="text-xs text-gray-500">
                  {(attachedMedia.file.size / 1024 / 1024).toFixed(2)} MB
                  {attachedMedia.isUploading && ` • Uploading ${attachedMedia.progress}%`}
                  {attachedMedia.error && <span className="text-red-500 font-semibold ml-1">({attachedMedia.error})</span>}
                </p>
                {attachedMedia.isUploading && (
                  <div className="w-32 bg-gray-200 rounded-full h-1.5 mt-1 overflow-hidden">
                    <div
                      className="bg-primary-600 h-1.5 rounded-full transition-all duration-300"
                      style={{ width: `${attachedMedia.progress}%` }}
                    />
                  </div>
                )}
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleRemoveAttachment}
              title="Remove attachment"
              type="button"
              className="text-gray-400 hover:text-gray-600 p-1"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        )}

        {/* Hidden file input for attachments */}
        <input
          type="file"
          ref={fileInputRef}
          className="hidden"
          onChange={handleFileSelect}
          accept="image/*,video/*,audio/*,application/*,text/*"
        />

        {/* Reply composer */}
        <form onSubmit={handleSend} className="p-3 border-t border-gray-200">
          {replyTo && (
            <div className="mb-2 flex items-start justify-between gap-2 rounded-md border-l-2 border-primary-500 bg-primary-50 px-2.5 py-1.5 text-xs text-gray-600">
              <span className="truncate"><strong>Replying to:</strong> {replyTo.content || 'Attachment'}</span>
              <button type="button" title="Cancel reply" onClick={() => setReplyTo(null)} className="text-gray-400 hover:text-gray-700">×</button>
            </div>
          )}
          {sendError && (
            <div className="flex items-center gap-2 mb-2 text-sm text-red-600 bg-red-50 rounded px-3 py-1.5">
              <AlertCircle className="h-4 w-4 shrink-0" />
              <span>{sendError}</span>
            </div>
          )}
          <div className="flex items-end gap-2">
            <textarea
              id="conversation-reply"
              className="form-control flex-1 resize-none"
              placeholder="Type a reply... (Shift+Enter for newline, Ctrl+Enter to send)"
              rows={2}
              value={reply}
              onChange={(e) => { setReply(e.target.value); if (sendError) setSendError(null) }}
              onKeyDown={(e) => {
                if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                  e.preventDefault()
                  handleSend(e)
                }
              }}
              disabled={conversation.status === 'CLOSED'}
            />
            <div className="flex gap-1">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowNote(!showNote)}
                title="Add internal note"
                type="button"
              >
                <StickyNote className="h-5 w-5" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                title="Attach media"
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={conversation.status === 'CLOSED' || attachedMedia?.isUploading}
              >
                <Paperclip className="h-5 w-5" />
              </Button>
              <Button
                variant="primary"
                size="sm"
                type="submit"
                disabled={
                  (!reply.trim() && !attachedMedia?.media_key) ||
                  conversation.status === 'CLOSED' ||
                  sendMessage.isPending ||
                  attachedMedia?.isUploading
                }
              >
                {sendMessage.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
              </Button>
            </div>
          </div>
        </form>
      </Card>
    </div>
  )
}

export default ConversationDetail
