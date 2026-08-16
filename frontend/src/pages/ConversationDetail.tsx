import React, { useState, useEffect, useRef } from 'react'
import { useParams, useNavigate } from '@tanstack/react-router'
import {
  useConversation,
  useSendMessage,
  useUploadMedia,
  useAssignConversation,
  useHandoffConversation,
  useCloseConversation,
  useReopenConversation,
  useAddInternalNote,
} from '@/hooks/useInbox'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
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

const ConversationDetail: React.FC = () => {
  const { id } = useParams({ strict: false })
  const navigate = useNavigate()
  const { data, isLoading, isError } = useConversation(id!)
  const sendMessage = useSendMessage()
  const uploadMedia = useUploadMedia()
  const assign = useAssignConversation()
  const handoff = useHandoffConversation()
  const close = useCloseConversation()
  const reopen = useReopenConversation()
  const addNote = useAddInternalNote()

  const [reply, setReply] = useState('')
  const [note, setNote] = useState('')
  const [showNote, setShowNote] = useState(false)
  const [attachedMedia, setAttachedMedia] = useState<AttachedMedia | null>(null)
  const timelineRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const conversation = data?.conversation
  const messages = data?.messages || []

  useEffect(() => {
    if (timelineRef.current) {
      timelineRef.current.scrollTop = timelineRef.current.scrollHeight
    }
  }, [messages.length])

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
        <Button variant="secondary" className="mt-4" onClick={() => navigate({ to: '/' })}>
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

    sendMessage.mutate(
      {
        account: conversation.account_id,
        recipient: '', // recipient is the contact; backend resolves
        message: text || (attachedMedia ? attachedMedia.file.name : ''),
        type: msgType,
        media_key: mediaKey,
      },
      {
        onSuccess: () => {
          setReply('')
          handleRemoveAttachment()
        },
        onError: () => setReply(reply), // keep text on error
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

  const handleAssign = () => {
    const name = window.prompt('Assign to operator:', conversation.assignee || '')
    if (name) assign.mutate({ id: conversation.id, assignee: name })
  }

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => navigate({ to: '/' })}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-xl font-semibold text-gray-900">
              Ticket #{conversation.ticket_number}
            </h1>
            <p className="text-sm text-gray-500">
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
              <Button variant="secondary" size="sm" onClick={handleAssign}>
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

      {/* Timeline */}
      <Card className="flex-1 flex flex-col overflow-hidden">
        <div ref={timelineRef} className="flex-1 overflow-y-auto p-4 space-y-2">
          {messages.length === 0 && (
            <div className="text-center text-gray-400 py-8">No messages yet</div>
          )}
          {messages.map((m) => {
            const resolvedUrl = resolveMediaUrl(m.media_url)
            return (
              <div
                key={m.id}
                className={`flex ${m.actor === 'CONTACT' || m.direction === 'INCOMING' ? 'justify-start' : 'justify-end'}`}
              >
                <div
                  className={`max-w-[70%] px-4 py-2 rounded-lg ${
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
                  <p className={`text-[10px] mt-1 ${m.actor === 'OPERATOR' || (m.direction === 'OUTGOING' && !m.is_internal) ? 'text-primary-100' : 'text-gray-400'}`}>
                    {formatTime(m.provider_timestamp)}
                  </p>
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
          <div className="flex items-end gap-2">
            <textarea
              className="form-control flex-1 resize-none"
              placeholder="Type a reply... (Shift+Enter for newline, Ctrl+Enter to send)"
              rows={2}
              value={reply}
              onChange={(e) => setReply(e.target.value)}
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
