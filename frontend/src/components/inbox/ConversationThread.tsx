import React, { useState, useEffect, useMemo, useRef, useLayoutEffect } from 'react'
import {
  useConversation,
  useSendMessage,
  useUploadMedia,
  useAddInternalNote,
  useOperators,
} from '@/hooks/useInbox'
import { useDashboardAccounts } from '@/hooks/useDashboard'
import { useAuth } from '@/hooks/useAuth'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
import AssignmentModal from '@/components/inbox/AssignmentModal'
import {
  Loader2,
  Send,
  StickyNote,
  Paperclip,
  X,
  FileText,
  AlertCircle,
  Check,
  CheckCheck,
  Clock,
  Users,
  Smile,
  Reply,
  UserCircle,
  ChevronDown,
} from 'lucide-react'
import { statusColors } from './constants'
import type { ConversationMessage } from '@/types'

const formatTime = (iso: string) =>
  new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })

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

const renderMessageContent = (content: string, outgoing: boolean) => {
  const parts = content.split(/(https?:\/\/[^\s<]+)/gi)
  return parts.map((part, index) => {
    if (!/^https?:\/\//i.test(part)) return <React.Fragment key={index}>{part}</React.Fragment>
    const trailing = part.match(/[),.!?;:]+$/)?.[0] || ''
    const href = trailing ? part.slice(0, -trailing.length) : part
    return (
      <React.Fragment key={index}>
        <a
          href={href}
          target="_blank"
          rel="noreferrer noopener"
          className={`underline break-all ${outgoing ? 'text-white' : 'text-primary-600'}`}
        >
          {href}
        </a>
        {trailing}
      </React.Fragment>
    )
  })
}

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

interface ConversationThreadProps {
  conversationId: string
}

const EMOJIS = [
  '😀','😃','😄','😁','😆','😂','🤣','🙂',
  '😊','😇','😍','🥰','😘','😎','🤔','🫡',
  '😐','😕','🙁','😢','😭','😤','😡','😅',
  '👍','👎','👏','🙌','🙏','🤝','💪','👀',
  '✅','☑️','❌','⚠️','🔥','⭐','💯','🎉',
  '❤️','💙','💚','💛','🧡','💜','🖤','🤍',
  '📌','📝','📎','📞','⏰','🚀','🎁','💡',
]
const QUICK_REACTIONS = ['✅', '👀', '🙏', '🔥', '👏', '⭐']
const getMessageExcerpt = (message: ConversationMessage) => {
  const text = message.content?.trim()
  if (text) return text.length > 90 ? `${text.slice(0, 87)}...` : text
  return message.message_type === 'TEXT' ? 'Message' : `${message.message_type.toLowerCase()} attachment`
}

const MESSAGE_BATCH_SIZE = 500

const ConversationThread: React.FC<ConversationThreadProps> = ({ conversationId }) => {
  const [messageLimit, setMessageLimit] = useState(MESSAGE_BATCH_SIZE)
  const { data, isLoading, isError, isFetching } = useConversation(conversationId, {
    limit: messageLimit,
  })
  const sendMessage = useSendMessage(conversationId)
  const uploadMedia = useUploadMedia()
  const addNote = useAddInternalNote()
  const { user } = useAuth()
  const { data: operators = [] } = useOperators()
  const { data: accountsData } = useDashboardAccounts()

  const [reply, setReply] = useState('')
  const [note, setNote] = useState('')
  const [showNote, setShowNote] = useState(false)
  const [showEmoji, setShowEmoji] = useState(false)
  const [attachedMedia, setAttachedMedia] = useState<AttachedMedia | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)
  const [isAssignOpen, setIsAssignOpen] = useState(false)
  const [replyTarget, setReplyTarget] = useState<ConversationMessage | null>(null)
  const [localReactions, setLocalReactions] = useState<Record<string, string>>({})
  const [sendAsOperatorId, setSendAsOperatorId] = useState('')
  const [showSendAsMenu, setShowSendAsMenu] = useState(false)
  const timelineRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const preserveScrollRef = useRef<{ top: number; height: number } | null>(null)

  const conversation = data?.conversation
  const messages = data?.messages || []
  const accounts = Array.isArray(accountsData) ? accountsData : []
  const channelLabel = useMemo(() => {
    if (!conversation?.account_id) return null
    return accounts.find((account: any) => account.id === conversation.account_id)?.display_name || conversation.account_id
  }, [accounts, conversation?.account_id])

  const { visibleMessages, reactionsByTarget } = useMemo(() => {
    const reactions = new Map<string, string>()
    const visible: ConversationMessage[] = []
    for (const message of messages) {
      if (message.message_type === 'REACTION' && message.reaction_target) {
        if (message.content) reactions.set(message.reaction_target, message.content)
        else reactions.delete(message.reaction_target)
      } else {
        visible.push(message)
      }
    }
    return { visibleMessages: visible, reactionsByTarget: reactions }
  }, [messages])

  const activeOperators = useMemo(() => {
    const list = Array.isArray(operators) ? operators : []
    return list
      .filter((operator) => operator.is_active !== false)
      .sort((a, b) => a.name.localeCompare(b.name))
  }, [operators])
  const selectedSendAsOperator = useMemo(
    () => activeOperators.find((operator) => operator.id === sendAsOperatorId) || null,
    [activeOperators, sendAsOperatorId]
  )

  useEffect(() => {
    if (!sendAsOperatorId && user?.id) {
      setSendAsOperatorId(user.id)
    }
  }, [sendAsOperatorId, user?.id])

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
  }, [visibleMessages.length])

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
      <div className="flex items-center justify-center h-full">
        <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
      </div>
    )
  }

  if (isError || !conversation) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-center">
        <AlertCircle className="h-12 w-12 text-red-400 mb-3" />
        <p className="text-gray-600">Conversation not found.</p>
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
    const currentReplyTarget = replyTarget
    const currentAttachment = attachedMedia
    const outgoingText = currentReplyTarget
      ? `Replying to "${getMessageExcerpt(currentReplyTarget)}"\n\n${text}`
      : text
    setReply('')
    setReplyTarget(null)
    setShowSendAsMenu(false)
    handleRemoveAttachment()
    setSendError(null)

    sendMessage.mutate(
      {
        account: conversation.account_id,
        recipient: conversation.contact_number || '',
        message: outgoingText || (currentAttachment ? currentAttachment.file.name : ''),
        type: msgType,
        media_key: mediaKey,
        is_group: !!(conversation as any).is_group,
        on_behalf_operator_id: sendAsOperatorId && sendAsOperatorId !== user?.id ? sendAsOperatorId : undefined,
        on_behalf_operator_name: selectedSendAsOperator?.name,
      },
      {
        onError: (err) => {
          const e = err as any
          setReply(currentReply)
          setReplyTarget(currentReplyTarget)
          setSendError(e?.response?.data?.error || 'Failed to send message')
        },
      }
    )
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

  const insertEmoji = (emoji: string) => {
    const textarea = textareaRef.current
    if (!textarea) return
    const start = textarea.selectionStart || 0
    const end = textarea.selectionEnd || 0
    const newValue = reply.slice(0, start) + emoji + reply.slice(end)
    setReply(newValue)
    requestAnimationFrame(() => {
      textarea.focus()
      const pos = start + emoji.length
      textarea.setSelectionRange(pos, pos)
    })
    setShowEmoji(false)
  }

  const handleReplyToMessage = (message: ConversationMessage) => {
    setReplyTarget(message)
    requestAnimationFrame(() => textareaRef.current?.focus())
  }

  const handleReactToMessage = (message: ConversationMessage, emoji: string) => {
    setLocalReactions((previous) => {
      const next = { ...previous }
      if (next[message.id] === emoji) delete next[message.id]
      else next[message.id] = emoji
      return next
    })
  }

  const closed = conversation.status === 'CLOSED'
  const displayName = (conversation as any).contact_name || (conversation as any).contact_number || `Ticket #${conversation.ticket_number}`
  const isGroup = !!(conversation as any).is_group
  const assigneeLabel = conversation.assignee || 'Unassigned'
  const hasOlderMessages = messages.length === messageLimit

  return (
    <Card className="h-full flex flex-col overflow-hidden rounded-none lg:rounded-2xl border-gray-200/80 shadow-[0_20px_70px_-35px_rgba(15,23,42,0.35)] bg-white/95 backdrop-blur">
      <div className="px-4 py-3 border-b border-gray-200 bg-gradient-to-r from-slate-50 via-white to-sky-50 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-3 min-w-0">
          <div className="h-10 w-10 rounded-2xl bg-gradient-to-br from-primary-600 to-cyan-500 text-white flex items-center justify-center shrink-0 shadow-lg shadow-primary-600/20">
            <span className="text-sm font-semibold">
              {displayName.charAt(0).toUpperCase()}
            </span>
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-sm font-semibold text-gray-900 truncate">{displayName}</h3>
              {isGroup && (
                <span className="inline-flex items-center gap-1 text-[10px] font-medium text-amber-700 bg-amber-50 border border-amber-200 px-2 py-0.5 rounded-full">
                  <Users className="h-3 w-3" />
                  Group
                </span>
              )}
              {channelLabel && (
                <span className="inline-flex items-center gap-1 text-[10px] font-medium text-sky-700 bg-sky-50 border border-sky-200 px-2 py-0.5 rounded-full">
                  <Clock className="h-3 w-3" />
                  {channelLabel}
                </span>
              )}
              <span className={`inline-flex items-center text-[10px] font-semibold px-2 py-0.5 rounded-full ${statusColors[conversation.status] || 'bg-gray-100 text-gray-800'}`}>
                {conversation.status.replace('_', ' ')}
              </span>
            </div>
            <p className="text-xs text-gray-500 truncate">
              {(conversation as any).contact_number ? (conversation as any).contact_number : `Ticket #${conversation.ticket_number}`} • Assignee: {assigneeLabel}
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => setIsAssignOpen(true)}
          className="inline-flex items-center gap-1.5 rounded-full border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 shadow-sm hover:bg-gray-50 transition-colors"
        >
          <Users className="h-3.5 w-3.5 text-gray-500" />
          Assign
        </button>
      </div>

      <div ref={timelineRef} onScroll={handleTimelineScroll} className="flex-1 overflow-y-auto p-4 space-y-2 bg-[linear-gradient(to_bottom,rgba(248,250,252,0.92),rgba(255,255,255,1))]">
        {hasOlderMessages && (
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
        {visibleMessages.length === 0 && (
          <div className="text-center text-gray-400 py-8 text-sm rounded-2xl border border-dashed border-gray-200 bg-white/80">
            No messages yet
          </div>
        )}
        {visibleMessages.map((m) => {
          const resolvedUrl = resolveMediaUrl(m.media_url)
          const messageReaction = localReactions[m.id] || reactionsByTarget.get(m.provider_message_id)
          const isReplyTarget = replyTarget?.id === m.id
          return (
            <div
              key={m.id}
              className={`group flex ${m.actor === 'CONTACT' || m.direction === 'INCOMING' ? 'justify-start' : 'justify-end'}`}
            >
              <div className={`flex items-end gap-2 max-w-[96%] ${m.actor === 'CONTACT' || m.direction === 'INCOMING' ? 'flex-row' : 'flex-row-reverse'}`}>
              <div
                className={`px-3.5 py-2 rounded-2xl shadow-sm transition-transform ${
                  isReplyTarget ? 'translate-x-2 ring-2 ring-primary-300 ring-offset-2' : ''
                } ${
                  m.is_internal
                    ? 'bg-amber-50 text-amber-900 border border-amber-200 border-dashed'
                    : m.actor === 'CONTACT' || m.direction === 'INCOMING'
                      ? 'bg-white border border-gray-200 text-gray-900'
                      : 'bg-gradient-to-br from-primary-600 to-cyan-600 text-white'
                }`}
                role={m.is_internal ? 'note' : undefined}
              >
                {m.is_internal && (
                  <p className="text-[10px] uppercase tracking-wide text-amber-700 mb-1">Internal note</p>
                )}
                {isGroup && (m.actor === 'CONTACT' || m.direction === 'INCOMING') && m.sender_address && !m.is_internal && (
                  <span className="text-[10px] font-semibold block mb-0.5 text-amber-700">
                    {formatSenderPhone(m.sender_address)}
                  </span>
                )}
                {(m.actor === 'OPERATOR' || m.actor === 'BOT' || m.direction === 'OUTGOING') && !m.is_internal && (
                  <span
                    className={`text-[10px] font-semibold block mb-0.5 ${
                      m.actor === 'OPERATOR' || m.direction === 'OUTGOING' ? 'text-primary-100' : 'text-gray-500'
                    }`}
                    title={m.operator_name ? `Sent by ${m.operator_name}` : `Sent by ${m.actor === 'BOT' ? 'Bot' : 'Operator'}`}
                  >
                    {m.operator_name || (m.actor === 'BOT' ? 'Bot' : 'Operator')}
                  </span>
                )}
                {m.message_type === 'IMAGE' && resolvedUrl ? (
                  <div className="mb-1.5">
                    <img
                      src={resolvedUrl}
                      alt="Attachment"
                      className="rounded-md max-h-60 max-w-full object-contain cursor-pointer"
                      loading="lazy"
                      onClick={() => window.open(resolvedUrl, '_blank')}
                    />
                  </div>
                ) : m.message_type === 'VIDEO' && resolvedUrl ? (
                  <div className="mb-1.5">
                    <video src={resolvedUrl} controls className="rounded-md max-h-60 max-w-full" />
                  </div>
                ) : m.message_type === 'AUDIO' && resolvedUrl ? (
                  <div className="mb-1.5">
                    <audio src={resolvedUrl} controls className="max-w-full my-1" />
                  </div>
                ) : resolvedUrl ? (
                  <div className="mb-1.5">
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
                <p className="text-sm leading-snug whitespace-pre-wrap break-words">
                  {renderMessageContent(m.content, m.actor === 'OPERATOR' || m.direction === 'OUTGOING')}
                </p>
                <div className="flex items-center justify-end gap-1 mt-1">
                  <span className={`text-[10px] ${m.actor === 'OPERATOR' || (m.direction === 'OUTGOING' && !m.is_internal) ? 'text-primary-100' : 'text-gray-400'}`}>
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
                {messageReaction && (
                  <span className="mt-1 inline-flex rounded-full border border-gray-200 bg-white px-2 py-0.5 text-[11px] shadow-sm">
                    {messageReaction}
                  </span>
                )}
              </div>
              {!isGroup && !m.is_internal && (
                <div className="mb-1 flex items-center gap-1 rounded-full border border-gray-200 bg-white/95 p-1 opacity-0 shadow-sm transition-opacity group-hover:opacity-100 group-focus-within:opacity-100">
                  <button
                    type="button"
                    onClick={() => handleReplyToMessage(m)}
                    className="inline-flex h-7 w-7 items-center justify-center rounded-full text-gray-500 hover:bg-gray-100 hover:text-gray-800"
                    title="Reply with context"
                    aria-label="Reply with context"
                  >
                    <Reply className="h-3.5 w-3.5" />
                  </button>
                  {QUICK_REACTIONS.map((emoji) => (
                    <button
                      key={emoji}
                      type="button"
                      onClick={() => handleReactToMessage(m, emoji)}
                      className="inline-flex h-7 w-7 items-center justify-center rounded-full text-sm hover:bg-gray-100"
                      title={`Mark ${emoji}`}
                      aria-label={`Mark ${emoji}`}
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              )}
              </div>
            </div>
          )
        })}
      </div>

      {showNote && (
        <form onSubmit={handleNote} className="p-3 border-t border-gray-200 bg-white">
          <textarea
            className="form-control w-full mb-2 text-sm py-2 px-3"
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

      {attachedMedia && (
        <div className="p-3 px-4 border-t border-gray-200 bg-slate-50 flex items-center justify-between gap-3 text-sm">
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

      <AssignmentModal
        isOpen={isAssignOpen}
        conversation={conversation}
        onClose={() => setIsAssignOpen(false)}
      />

      <input
        type="file"
        ref={fileInputRef}
        className="hidden"
        onChange={handleFileSelect}
        accept="image/*,video/*,audio/*,application/*,text/*"
      />

      <form onSubmit={handleSend} className="p-3 border-t border-gray-200 bg-white">
        {replyTarget && (
          <div className="mb-2 flex items-start justify-between gap-3 rounded-xl border border-primary-100 bg-primary-50/70 px-3 py-2">
            <div className="min-w-0 border-l-2 border-primary-500 pl-2">
              <p className="text-[11px] font-semibold uppercase tracking-wide text-primary-700">
                Replying to {replyTarget.actor === 'CONTACT' || replyTarget.direction === 'INCOMING' ? displayName : replyTarget.operator_name || 'team message'}
              </p>
              <p className="truncate text-xs text-gray-600">{getMessageExcerpt(replyTarget)}</p>
            </div>
            <button
              type="button"
              onClick={() => setReplyTarget(null)}
              className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-gray-500 hover:bg-white hover:text-gray-800"
              title="Cancel reply"
              aria-label="Cancel reply"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        )}
        {sendError && (
          <div className="flex items-center gap-2 mb-1 text-sm text-red-600 bg-red-50 rounded px-3 py-1.5">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{sendError}</span>
          </div>
        )}
        <div className="flex items-end gap-2">
          <div className="relative flex-1">
            <textarea
              ref={textareaRef}
              className="form-control w-full resize-none text-sm py-2.5 px-3.5 min-h-[44px] rounded-xl border-gray-300 focus:border-primary-400 focus:ring-primary-500/20"
              placeholder="Type a reply... (Shift+Enter for newline, Ctrl+Enter to send)"
              rows={1}
              value={reply}
              onChange={(e) => { setReply(e.target.value); if (sendError) setSendError(null) }}
              onKeyDown={(e) => {
                if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
                  e.preventDefault()
                  handleSend(e)
                }
              }}
              disabled={closed}
            />
            <div className="mt-1.5 flex items-center gap-2">
            {activeOperators.length > 0 && (
              <div className="relative inline-flex">
                <button
                  type="button"
                  onClick={() => setShowSendAsMenu((open) => !open)}
                  disabled={closed || sendMessage.isPending}
                  className="inline-flex max-w-[14rem] items-center gap-1.5 rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 text-[11px] text-gray-600 hover:bg-white hover:text-gray-800 disabled:cursor-not-allowed disabled:opacity-60"
                  title="Choose sender"
                  aria-label="Choose sender"
                  aria-haspopup="menu"
                  aria-expanded={showSendAsMenu}
                >
                  <UserCircle className="h-3.5 w-3.5 shrink-0 text-gray-400" />
                  <span className="shrink-0 text-gray-400">as</span>
                  <span className="truncate">
                    {selectedSendAsOperator?.name || user?.name || 'Me'}
                    {selectedSendAsOperator?.id === user?.id ? ' (me)' : ''}
                  </span>
                  <ChevronDown className="h-3 w-3 shrink-0 text-gray-400" />
                </button>
                {showSendAsMenu && (
                  <div
                    className="absolute bottom-full left-0 z-30 mb-1 w-max min-w-[10rem] max-w-[16rem] overflow-hidden rounded-lg border border-gray-200 bg-white py-1 shadow-lg"
                    role="menu"
                  >
                    {user && !activeOperators.some((operator) => operator.id === user.id) && (
                      <button
                        type="button"
                        className="flex w-full items-center justify-between gap-3 px-2.5 py-1.5 text-left text-xs text-gray-700 hover:bg-gray-50"
                        onClick={() => {
                          setSendAsOperatorId(user.id)
                          setShowSendAsMenu(false)
                        }}
                        role="menuitem"
                      >
                        <span className="truncate">{user.name || 'Me'}</span>
                        {sendAsOperatorId === user.id && <Check className="h-3.5 w-3.5 text-primary-600" />}
                      </button>
                    )}
                    {activeOperators.map((operator) => (
                      <button
                        key={operator.id}
                        type="button"
                        className="flex w-full items-center justify-between gap-3 px-2.5 py-1.5 text-left text-xs text-gray-700 hover:bg-gray-50"
                        onClick={() => {
                          setSendAsOperatorId(operator.id)
                          setShowSendAsMenu(false)
                        }}
                        role="menuitem"
                      >
                        <span className="truncate">
                          {operator.name}{operator.id === user?.id ? ' (me)' : ''}
                        </span>
                        {sendAsOperatorId === operator.id && <Check className="h-3.5 w-3.5 text-primary-600" />}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowNote(!showNote)}
              title="Add internal note"
              type="button"
              className="h-7 rounded-full px-2 text-[11px] text-gray-600 hover:text-amber-700"
            >
              <StickyNote className="h-3.5 w-3.5 mr-1" />
              Note
            </Button>
            </div>
          </div>
          <div className="flex gap-1 relative">
            <div className="relative">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowEmoji(!showEmoji)}
                title="Insert emoji"
                type="button"
                className="rounded-xl"
              >
                <Smile className="h-5 w-5" />
              </Button>
              {showEmoji && (
                <div className="absolute bottom-full right-0 mb-2 w-64 p-2 bg-white border border-gray-200 rounded-2xl shadow-xl grid grid-cols-8 gap-1 z-20">
                  {EMOJIS.map((emoji) => (
                    <button
                      key={emoji}
                      type="button"
                      onClick={() => insertEmoji(emoji)}
                      className="text-lg hover:bg-gray-100 rounded p-1"
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <Button
              variant="ghost"
              size="sm"
              title="Attach media"
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={closed || attachedMedia?.isUploading}
              className="rounded-xl"
            >
              <Paperclip className="h-5 w-5" />
            </Button>
            <Button
              variant="primary"
              size="sm"
              type="submit"
              disabled={
                (!reply.trim() && !attachedMedia?.media_key) ||
                closed ||
                sendMessage.isPending ||
                attachedMedia?.isUploading
              }
              className="rounded-xl px-4"
            >
              {sendMessage.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
            </Button>
          </div>
        </div>
      </form>
    </Card>
  )
}

export default ConversationThread
