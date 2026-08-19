import React, { useState, useEffect, useRef } from 'react'
import {
  useConversation,
  useSendMessage,
  useUploadMedia,
  useAddInternalNote,
} from '@/hooks/useInbox'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
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
} from 'lucide-react'

const formatTime = (iso: string) =>
  new Date(iso).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })

const resolveMediaUrl = (url: string) => {
  if (!url) return ''
  if (url.startsWith('/api/v1/media/')) {
    return `/dashboard/api/media/${url.slice('/api/v1/media/'.length)}`
  }
  return url
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

const EMOJIS = ['😀','😂','🙂','😊','😍','🤔','👍','👎','🙏','🔥','❤️','🎉','👀','✅','❌','🙌','💪','🤝','📝','📎','🎁','🚀','😭','😡']

const ConversationThread: React.FC<ConversationThreadProps> = ({ conversationId }) => {
  const { data, isLoading, isError } = useConversation(conversationId)
  const sendMessage = useSendMessage(conversationId)
  const uploadMedia = useUploadMedia()
  const addNote = useAddInternalNote()

  const [reply, setReply] = useState('')
  const [note, setNote] = useState('')
  const [showNote, setShowNote] = useState(false)
  const [showEmoji, setShowEmoji] = useState(false)
  const [attachedMedia, setAttachedMedia] = useState<AttachedMedia | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)
  const timelineRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const conversation = data?.conversation
  const messages = data?.messages || []

  useEffect(() => {
    if (timelineRef.current) {
      timelineRef.current.scrollTop = timelineRef.current.scrollHeight
    }
  }, [messages.length])

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
        is_group: !!(conversation as any).is_group,
      },
      {
        onError: (err) => {
          const e = err as any
          setReply(currentReply)
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

  const closed = conversation.status === 'CLOSED'
  const displayName = (conversation as any).contact_name || (conversation as any).contact_number || `Ticket #${conversation.ticket_number}`
  const isGroup = !!(conversation as any).is_group

  return (
    <Card className="h-full flex flex-col overflow-hidden rounded-none lg:rounded-lg">
      <div className="p-2.5 border-b border-gray-200 bg-white flex items-center justify-between shrink-0">
        <div className="flex items-center gap-2 min-w-0">
          <div className="h-7 w-7 rounded-full bg-primary-100 flex items-center justify-center shrink-0">
            <span className="text-xs font-semibold text-primary-600">
              {displayName.charAt(0).toUpperCase()}
            </span>
          </div>
          <div className="min-w-0">
            <h3 className="text-sm font-semibold text-gray-900 truncate flex items-center gap-1.5">
              {displayName}
              {isGroup && (
                <span className="inline-flex items-center gap-0.5 text-[10px] font-medium text-amber-600 bg-amber-50 px-1.5 py-0.5 rounded-full">
                  <Users className="h-3 w-3" />
                  Group
                </span>
              )}
            </h3>
            <p className="text-xs text-gray-500 truncate">
              {(conversation as any).contact_number ? (conversation as any).contact_number : `Ticket #${conversation.ticket_number}`}
            </p>
          </div>
        </div>
      </div>

      <div ref={timelineRef} className="flex-1 overflow-y-auto p-2 space-y-1">
        {messages.length === 0 && (
          <div className="text-center text-gray-400 py-8 text-sm">No messages yet</div>
        )}
        {messages.map((m) => {
          const resolvedUrl = resolveMediaUrl(m.media_url)
          return (
            <div
              key={m.id}
              className={`flex ${m.actor === 'CONTACT' || m.direction === 'INCOMING' ? 'justify-start' : 'justify-end'}`}
            >
              <div
                className={`max-w-[80%] px-3 py-1.5 rounded-lg ${
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
                {(m.actor === 'OPERATOR' || m.actor === 'BOT' || m.direction === 'OUTGOING') && !m.is_internal && (
                  <span
                    className="text-[10px] font-semibold text-primary-200 block mb-0.5"
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
                <p className="text-sm leading-snug">{m.content}</p>
                <div className="flex items-center justify-end gap-1 mt-0.5">
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
              </div>
            </div>
          )
        })}
      </div>

      {showNote && (
        <form onSubmit={handleNote} className="p-2 border-t border-gray-200">
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

      <input
        type="file"
        ref={fileInputRef}
        className="hidden"
        onChange={handleFileSelect}
        accept="image/*,video/*,audio/*,application/*,text/*"
      />

      <form onSubmit={handleSend} className="p-2 border-t border-gray-200">
        {sendError && (
          <div className="flex items-center gap-2 mb-1 text-sm text-red-600 bg-red-50 rounded px-3 py-1.5">
            <AlertCircle className="h-4 w-4 shrink-0" />
            <span>{sendError}</span>
          </div>
        )}
        <div className="flex items-end gap-1.5">
          <div className="relative flex-1">
            <textarea
              ref={textareaRef}
              className="form-control w-full resize-none text-sm py-2 px-3 min-h-[38px]"
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
          </div>
          <div className="flex gap-1 relative">
            <div className="relative">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowEmoji(!showEmoji)}
                title="Insert emoji"
                type="button"
              >
                <Smile className="h-5 w-5" />
              </Button>
              {showEmoji && (
                <div className="absolute bottom-full right-0 mb-2 w-64 p-2 bg-white border border-gray-200 rounded-lg shadow-lg grid grid-cols-8 gap-1 z-20">
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
              disabled={closed || attachedMedia?.isUploading}
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
