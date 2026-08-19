import React, { useState } from 'react'
import { useParams, Link } from '@tanstack/react-router'
import {
  useContact,
  useCreateContactActivity,
  useUpdateContact,
  useContactConversations,
  useContactFieldDefinitions,
  useDealStages,
  usePipelines,
} from '@/hooks/useInbox'
import { DealPipelineTracker, StageIcon } from '@/components/DealPipelineTracker'
import { SummaryFeed } from '@/components/SummaryFeed'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Loader2,
  ArrowLeft,
  Users,
  AlertCircle,
  Phone,
  Tag,
  Plus,
  Activity as ActivityIcon,
  Edit3,
  X,
  Check,
  Trash2,
  MessageSquare,
  CheckSquare,
  ExternalLink,
  Calendar,
  Layers,
  Inbox,
  User,
} from 'lucide-react'
import type { Conversation } from '@/types'

const priorityBadgeStyles: Record<string, string> = {
  HIGH: 'bg-red-50 text-red-700 border border-red-200',
  NORMAL: 'bg-amber-50 text-amber-700 border border-amber-200',
  LOW: 'bg-emerald-50 text-emerald-700 border border-emerald-200',
}

const statusBadgeStyles: Record<string, string> = {
  OPEN: 'bg-blue-50 text-blue-700 border border-blue-200',
  BOT_ACTIVE: 'bg-purple-50 text-purple-700 border border-purple-200',
  WAITING: 'bg-amber-50 text-amber-700 border border-amber-200',
  HANDED_OFF: 'bg-orange-50 text-orange-700 border border-orange-200',
  CLOSED: 'bg-gray-100 text-gray-700 border border-gray-200',
}

const formatDate = (iso: string | null | undefined) =>
  iso ? new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }) : '-'

const formatDateOnly = (iso: string | null | undefined) =>
  iso ? new Date(iso).toLocaleDateString(undefined, { dateStyle: 'medium' }) : '-'

// ── Local task entry type (manual-only) ──────────────────────────────
interface TaskEntry {
  id: string
  summary: string
  type: 'task' | 'follow_up'
  priority: 'HIGH' | 'NORMAL' | 'LOW'
  status: string
  due_at: string | null
  created_at: string
}

// ── TaskEntryRow ─────────────────────────────────────────────────────

function TaskEntryRow({
  task,
  onUpdate,
  onDelete,
}: {
  task: TaskEntry
  onUpdate: (id: string, updates: Partial<TaskEntry>) => void
  onDelete: (id: string) => void
}) {
  const [isEditing, setIsEditing] = useState(false)
  const [editSummary, setEditSummary] = useState(task.summary)
  const [editType, setEditType] = useState(task.type)
  const [editPriority, setEditPriority] = useState(task.priority)
  const [editDueAt, setEditDueAt] = useState(task.due_at ? task.due_at.slice(0, 16) : '')

  const handleSave = () => {
    if (!editSummary.trim()) return
    onUpdate(task.id, {
      summary: editSummary.trim(),
      type: editType,
      priority: editPriority,
      due_at: editDueAt ? new Date(editDueAt).toISOString() : null,
    })
    setIsEditing(false)
  }

  const handleCancel = () => {
    setEditSummary(task.summary)
    setEditType(task.type)
    setEditPriority(task.priority)
    setEditDueAt(task.due_at ? task.due_at.slice(0, 16) : '')
    setIsEditing(false)
  }

  const handleDelete = () => {
    onDelete(task.id)
  }

  if (isEditing) {
    return (
      <div className="p-2.5 rounded-lg bg-gray-50 border border-gray-200 space-y-2">
        <Input
          type="text"
          value={editSummary}
          onChange={(e) => setEditSummary(e.target.value)}
          placeholder="Task summary..."
          className="w-full text-xs"
        />
        <div className="flex flex-wrap items-center gap-3">
          <div className="flex items-center gap-2 min-w-[110px] flex-1">
            <Label className="text-[11px] text-gray-500 shrink-0">Type:</Label>
            <select
              value={editType}
              onChange={(e) => setEditType(e.target.value as 'task' | 'follow_up')}
              className="w-full rounded-lg border border-gray-200 bg-white px-2 py-1 text-[11px] text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="follow_up">Follow-up</option>
              <option value="task">Task</option>
            </select>
          </div>
          <div className="flex items-center gap-2 min-w-[120px] flex-1">
            <Label className="text-[11px] text-gray-500 shrink-0">Priority:</Label>
            <select
              value={editPriority}
              onChange={(e) => setEditPriority(e.target.value as 'HIGH' | 'NORMAL' | 'LOW')}
              className="w-full rounded-lg border border-gray-200 bg-white px-2 py-1 text-[11px] text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="HIGH">High</option>
              <option value="NORMAL">Normal</option>
              <option value="LOW">Low</option>
            </select>
          </div>
          <div className="flex items-center gap-2 flex-1 min-w-[160px]">
            <Label className="text-[11px] text-gray-500 shrink-0">Due:</Label>
            <Input
              type="datetime-local"
              value={editDueAt}
              onChange={(e) => setEditDueAt(e.target.value)}
              className="w-full text-xs py-1"
            />
          </div>
          <div className="flex items-center gap-1.5 shrink-0">
            <button
              type="button"
              onClick={handleCancel}
              className="p-1 rounded hover:bg-gray-200 text-gray-500 hover:text-gray-700 transition-colors"
              title="Cancel"
            >
              <X className="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              onClick={handleSave}
              className="p-1 rounded hover:bg-primary-100 text-primary-600 hover:text-primary-700 transition-colors"
              title="Save"
            >
              <Check className="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex gap-2.5 p-2.5 rounded-lg hover:bg-gray-50 transition-colors group">
      <div className="flex-shrink-0 mt-0.5">
        <span
          className={`inline-block w-2.5 h-2.5 rounded-full mt-0.5 ${
            task.status === 'OPEN' ? 'bg-blue-400' : 'bg-gray-300'
          }`
}
        />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-xs font-medium text-gray-800 leading-snug">{task.summary}</p>
        <div className="flex items-center gap-2 mt-0.5">
          <span
            className={`text-[10px] font-medium px-1.5 py-0.5 rounded-full ${
              priorityBadgeStyles[task.priority] || 'bg-gray-100 text-gray-700'
            }`
}
          >
            {task.priority}
          </span>
          <span className="text-[10px] text-gray-500">
            {task.type === 'follow_up' ? 'Follow-up' : 'Task'}
          </span>
          {task.due_at && (
            <span className="inline-flex items-center gap-0.5 text-[10px] text-amber-700 font-medium">
              <Calendar className="h-2.5 w-2.5" />
              {formatDate(task.due_at)}
            </span>
          )}
        </div>
        <p className="text-[10px] text-gray-400 mt-0.5">{formatDate(task.created_at)}</p>
      </div>
      <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-0.5 shrink-0 mt-0.5">
        <button
          type="button"
          onClick={() => setIsEditing(true)}
          className="p-1 rounded hover:bg-gray-200 text-gray-400 hover:text-gray-700 transition-colors"
          title="Edit"
        >
          <Edit3 className="h-3 w-3" />
        </button>
        <button
          type="button"
          onClick={handleDelete}
          className="p-1 rounded hover:bg-red-50 text-gray-400 hover:text-red-600 transition-colors"
          title="Delete"
        >
          <Trash2 className="h-3 w-3" />
        </button>
      </div>
    </div>
  )
}

type HubTabKey = 'summary' | 'tasks' | 'tickets'

export const ContactDetail: React.FC = () => {
  const { id } = useParams({ strict: false })
  const contactId = id || ''
  const {
    data: contact,
    isLoading: contactLoading,
    isError: contactError,
    refetch: refetchContact,
  } = useContact(contactId)
  const {
    data: conversations = [],
    isLoading: conversationsLoading,
    isError: conversationsError,
    refetch: refetchConversations,
  } = useContactConversations(contactId)
  const createActivity = useCreateContactActivity(contactId)
  const updateContact = useUpdateContact(contactId)
  const { data: fieldDefinitions = [] } = useContactFieldDefinitions()
  const { data: stages = [] } = useDealStages()
  const { data: pipelines = [] } = usePipelines()

  const [activeTab, setActiveTab] = useState<HubTabKey>('tasks')
  const [isEditing, setIsEditing] = useState(false)
  const [editName, setEditName] = useState('')
  const [editEmail, setEditEmail] = useState('')
  const [editTags, setEditTags] = useState('')
  const [editPipelineKey, setEditPipelineKey] = useState('sales')
  const [editDealStageKey, setEditDealStageKey] = useState('')
  const [editDealStageId, setEditDealStageId] = useState('')
  const [customValues, setCustomValues] = useState<Record<string, unknown>>({})

  React.useEffect(() => {
    if (contact) {
      setEditName(contact.name || '')
      setEditEmail(contact.email || '')
      setEditTags((contact.tags || []).join(', '))
      setEditDealStageKey(contact.deal_stage?.key || '')
      setEditDealStageId(contact.deal_stage?.id || '')
      setCustomValues(contact.custom_values || {})
    }
  }, [contact])

  // ── Manual tasks state ─────────────────────────────────────────────
  const [taskEntries, setTaskEntries] = useState<TaskEntry[]>([])
  const [taskSummary, setTaskSummary] = useState('')
  const [taskType, setTaskType] = useState<'task' | 'follow_up'>('follow_up')
  const [taskPriority, setTaskPriority] = useState<'HIGH' | 'NORMAL' | 'LOW'>('NORMAL')
  const [taskDueAt, setTaskDueAt] = useState('')
  const [taskFormError, setTaskFormError] = useState<string | null>(null)

  const handleAddTask = async (e: React.FormEvent) => {
    e.preventDefault()
    setTaskFormError(null)
    if (!taskSummary.trim()) {
      setTaskFormError('Please enter a task summary')
      return
    }
    const newTask: TaskEntry = {
      id: `local-${Date.now()}`,
      summary: taskSummary.trim(),
      type: taskType,
      priority: taskPriority,
      status: 'OPEN',
      due_at: taskDueAt ? new Date(taskDueAt).toISOString() : null,
      created_at: new Date().toISOString(),
    }
    createActivity.mutate(
      {
        type: taskType === 'follow_up' ? 'FOLLOW_UP' : 'NOTE',
        summary: taskSummary.trim(),
        priority: taskPriority,
        due_at: taskDueAt ? new Date(taskDueAt).toISOString() : null,
      },
      {
        onError: (err: any) =>
          setTaskFormError(
            err?.response?.data?.error || err?.message || 'Failed to create task'
          ),
        onSuccess: () => {
          setTaskEntries((prev) => [newTask, ...prev])
          setTaskSummary('')
          setTaskDueAt('')
          setTaskPriority('NORMAL')
          setTaskType('follow_up')
        },
      }
    )
  }

  const handleUpdateTask = (id: string, updates: Partial<TaskEntry>) => {
    setTaskEntries((prev) =>
      prev.map((t) => (t.id === id ? { ...t, ...updates } : t))
    )
  }

  const handleDeleteTask = (id: string) => {
    setTaskEntries((prev) => prev.filter((t) => t.id !== id))
  }

  const handleSaveProperties = async () => {
    updateContact.mutate(
      {
        name: editName.trim(),
        email: editEmail.trim(),
        tags: editTags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean),
        custom_values: customValues,
        deal_stage_key: editDealStageKey || undefined,
        deal_stage_id: editDealStageId || undefined,
      },
      {
        onSuccess: () => {
          setIsEditing(false)
        },
      }
    )
  }

  if (contactLoading) {
    return (
      <div className="flex justify-center py-16">
        <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
      </div>
    )
  }

  if (contactError || !contact) {
    return (
      <div className="text-center py-16">
        <AlertCircle className="h-12 w-12 text-red-400 mx-auto mb-3" />
        <p className="text-gray-600 font-medium">Failed to load contact.</p>
        <Button variant="primary" size="sm" className="mt-4" onClick={() => refetchContact()}>
          Retry
        </Button>
      </div>
    )
  }

  const displayName = contact.name || contact.number
  const initials = (contact.name || contact.number || '?').charAt(0).toUpperCase()
  const openTaskCount = taskEntries.filter((t) => t.status === 'OPEN').length
  const latestConversation = conversations.length > 0 ? conversations[0] : null

  // ── Task form renderer ─────────────────────────────────────────────
  const renderTaskForm = () => (
    <form onSubmit={handleAddTask} className="space-y-2.5">
      <div className="flex items-center gap-2 mb-1">
        <Plus className="h-3.5 w-3.5 text-primary-600" />
        <span className="text-[11px] font-semibold uppercase tracking-wider text-gray-600">
          Add Task / Follow-up
        </span>
      </div>
      {taskFormError && (
        <div className="flex items-start gap-2 p-2 rounded-lg bg-red-50 border border-red-200 text-[11px] text-red-700">
          <AlertCircle className="h-3.5 w-3.5 text-red-500 shrink-0 mt-0.5" />
          <span>{taskFormError}</span>
        </div>
      )}
      <div>
        <Input
          type="text"
          value={taskSummary}
          onChange={(e) => setTaskSummary(e.target.value)}
          placeholder="Enter task summary..."
          className="w-full text-xs"
        />
      </div>
      <div className="space-y-2">
        <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2 min-w-[120px] flex-1">
          <Label className="text-[11px] text-gray-500 shrink-0">Type:</Label>
          <select
            value={taskType}
            onChange={(e) => setTaskType(e.target.value as 'task' | 'follow_up')}
            className="w-full rounded-lg border border-gray-200 bg-white px-2 py-1 text-[11px] text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="follow_up">Follow-up</option>
            <option value="task">Task</option>
          </select>
        </div>
        <div className="flex items-center gap-2 min-w-[130px] flex-1">
          <Label className="text-[11px] text-gray-500 shrink-0">Priority:</Label>
          <select
            value={taskPriority}
            onChange={(e) => setTaskPriority(e.target.value as 'HIGH' | 'NORMAL' | 'LOW')}
            className="w-full rounded-lg border border-gray-200 bg-white px-2 py-1 text-[11px] text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="HIGH">High</option>
            <option value="NORMAL">Normal</option>
            <option value="LOW">Low</option>
          </select>
        </div>
        </div>
        <div className="flex flex-wrap items-center gap-3">
        <div className="flex items-center gap-2 flex-1 min-w-[200px]">
          <Label className="text-[11px] text-gray-500 shrink-0">Due:</Label>
          <Input
            type="datetime-local"
            value={taskDueAt}
            onChange={(e) => setTaskDueAt(e.target.value)}
            className="w-full text-xs py-1"
          />
        </div>
        <Button
          type="submit"
          variant="primary"
          size="sm"
          disabled={createActivity.isPending}
          className="shrink-0 gap-1 text-[11px]"
        >
          {createActivity.isPending ? (
            <>
              <Loader2 className="h-3 w-3 animate-spin" />
              <span>Adding...</span>
            </>
          ) : (
            <>
              <Plus className="h-3 w-3" />
              <span>Add {taskType === 'follow_up' ? 'Follow-up' : 'Task'}</span>
            </>
          )}
        </Button>
        </div>
      </div>
    </form>
  )

  return (
    <div className="max-w-7xl mx-auto space-y-5">
      {/* Top Breadcrumb & Action Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-1 border-b border-gray-200">
        <div className="flex items-center gap-3">
          <Link
            to="/contacts"
            className="inline-flex items-center gap-1.5 text-xs font-medium text-gray-500 hover:text-gray-900 transition-colors p-1.5 rounded-md hover:bg-gray-100"
          >
            <ArrowLeft className="h-4 w-4" />
            <span>Contacts</span>
          </Link>
          <span className="text-gray-300">/</span>
          <div className="flex items-center gap-2">
            <h1 className="text-lg font-semibold text-gray-900 truncate">{displayName}</h1>
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                contact.is_group
                  ? 'bg-amber-50 text-amber-700 border border-amber-200'
                  : 'bg-blue-50 text-blue-700 border border-blue-200'
              }`}
            >
              {contact.is_group ? <Users className="h-3 w-3" /> : <User className="h-3 w-3" />}
              {contact.is_group ? 'Group' : 'Personal'}
            </span>
          </div>
        </div>

        {latestConversation ? (
          <Link
            to="/conversations/$id"
            params={{ id: latestConversation.id }}
            className="inline-flex items-center gap-1.5 text-xs font-medium text-primary-600 hover:text-primary-700 bg-primary-50 hover:bg-primary-100 border border-primary-200 px-3 py-1.5 rounded-lg transition-colors"
          >
            <Inbox className="h-3.5 w-3.5" />
            <span>Open in Inbox</span>
          </Link>
        ) : (
          <Link
            to="/inbox"
            className="inline-flex items-center gap-1.5 text-xs font-medium text-gray-600 hover:text-gray-900 bg-gray-50 hover:bg-gray-100 border border-gray-200 px-3 py-1.5 rounded-lg transition-colors"
          >
            <Inbox className="h-3.5 w-3.5" />
            <span>Go to Inbox</span>
          </Link>
        )}
      </div>

      {/* Deal Pipeline Stage Banner */}
      <DealPipelineTracker contactId={contactId} currentStageKey={contact.deal_stage?.key ?? null} currentStageId={contact.deal_stage?.id ?? null} />

      {/* Main 2-Column Responsive Split Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
        {/* LEFT PANEL */}
        <div className="lg:col-span-5 space-y-4">
          {/* Identity & Profile Card */}
          <Card className="p-5 border border-gray-200/80 shadow-sm">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-center gap-3.5 min-w-0">
                <div className="h-12 w-12 rounded-full bg-primary-100 text-primary-700 border border-primary-200 flex items-center justify-center font-bold text-lg shrink-0 shadow-inner">
                  {initials}
                </div>
                <div className="min-w-0">
                  <h2 className="text-base font-semibold text-gray-900 truncate">{displayName}</h2>
                  <div className="flex items-center gap-2 text-xs text-gray-500 mt-0.5">
                    <span className="inline-flex items-center gap-1 truncate font-mono">
                      <Phone className="h-3 w-3 text-gray-400 shrink-0" />
                      {contact.number}
                    </span>
                  </div>
                </div>
              </div>

              {!isEditing && (
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => setIsEditing(true)}
                  className="gap-1 text-xs shrink-0"
                >
                  <Edit3 className="h-3.5 w-3.5" />
                  <span>Edit</span>
                </Button>
              )}
            </div>

            {/* CRM Snapshot Strip */}
            <div className="mt-4 grid grid-cols-3 divide-x divide-gray-200 rounded-lg bg-gray-50/80 border border-gray-200/80 p-2.5 text-center">
              <div>
                <span className="block text-xs text-gray-500 font-medium">Tickets</span>
                <span className="block text-sm font-semibold text-gray-900 mt-0.5">
                  {conversations.length}
                </span>
              </div>
              <div>
                <span className="block text-xs text-gray-500 font-medium">Open Tasks</span>
                <span className="block text-sm font-semibold text-gray-900 mt-0.5">
                  {openTaskCount}
                </span>
              </div>
              <div>
                <span className="block text-xs text-gray-500 font-medium">Joined</span>
                <span className="block text-xs font-semibold text-gray-800 mt-1 truncate px-1">
                  {formatDateOnly(contact.created_at)}
                </span>
              </div>
            </div>
          </Card>

          {/* Properties Inspector / Edit Form */}
          <Card className="p-0 border border-gray-200/80 shadow-sm overflow-hidden">
            <div className="px-4 py-3 bg-gray-50/80 border-b border-gray-200 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Layers className="h-4 w-4 text-gray-500" />
                <h3 className="text-xs font-semibold uppercase tracking-wider text-gray-700">
                  {isEditing ? 'Edit Contact Properties' : 'Contact Properties'}
                </h3>
              </div>
              {isEditing && (
                <button
                  type="button"
                  onClick={() => setIsEditing(false)}
                  className="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-200/60"
                >
                  <X className="h-4 w-4" />
                </button>
              )}
            </div>

            {isEditing ? (
              <div className="p-4 space-y-3.5">
                <div>
                  <Label htmlFor="editName" className="text-xs font-medium text-gray-700">Display Name</Label>
                  <Input id="editName" value={editName} onChange={(e) => setEditName(e.target.value)} className="mt-1 text-xs" placeholder="e.g. John Doe" />
                </div>
                <div>
                  <Label htmlFor="editEmail" className="text-xs font-medium text-gray-700">Email Address</Label>
                  <Input id="editEmail" type="email" value={editEmail} onChange={(e) => setEditEmail(e.target.value)} className="mt-1 text-xs" placeholder="e.g. john@example.com" />
                </div>
                <div>
                  <Label htmlFor="editTags" className="text-xs font-medium text-gray-700">Tags (comma-separated)</Label>
                  <Input id="editTags" value={editTags} onChange={(e) => setEditTags(e.target.value)} className="mt-1 text-xs" placeholder="VIP, Wholesale, Lead" />
                </div>
                <div>
                  <Label htmlFor="editPipeline" className="text-xs font-medium text-gray-700">Pipeline</Label>
                  <select id="editPipeline" value={editPipelineKey} onChange={(e) => setEditPipelineKey(e.target.value)} className="mt-1 w-full rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500">
                    {(pipelines || [{ id: 'sales', key: 'sales', name: 'Sales Pipeline' }]).map((p) => (
                      <option key={p.key} value={p.key}>{p.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <Label htmlFor="editDealStage" className="text-xs font-medium text-gray-700">Deal Stage</Label>
                  <select id="editDealStage" value={editDealStageId} onChange={(e) => { const stageId = e.target.value; setEditDealStageId(stageId); const stage = stages.find((s) => s.id === stageId); setEditDealStageKey(stage?.key || '') }} className="mt-1 w-full rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500">
                    <option value="">-- No Deal Stage --</option>
                    {stages.filter((s) => s.is_active).map((stage) => (
                      <option key={stage.id} value={stage.id}>{stage.label} {stage.is_won ? '🏆 (Won)' : stage.is_lost ? '❌ (Lost)' : ''}</option>
                    ))}
                  </select>
                </div>

                {fieldDefinitions.length > 0 && (
                  <div className="pt-3 border-t border-gray-100 space-y-3">
                    <p className="text-xs font-semibold text-gray-700">Tenant Custom Fields</p>
                    {fieldDefinitions.map((field) => {
                      const value = customValues[field.key]
                      const setValue = (val: unknown) => { setCustomValues((prev) => ({ ...prev, [field.key]: val })) }
                      return (
                        <div key={field.key}>
                          <Label htmlFor={`custom-${field.key}`} className="text-xs text-gray-600">{field.label}</Label>
                          {field.field_type === 'text' && <Input id={`custom-${field.key}`} value={String(value || '')} onChange={(e) => setValue(e.target.value)} className="mt-1 text-xs" />}
                          {field.field_type === 'number' && <Input id={`custom-${field.key}`} type="number" value={String(value || '')} onChange={(e) => setValue(e.target.value === '' ? '' : Number(e.target.value))} className="mt-1 text-xs" />}
                          {field.field_type === 'date' && <Input id={`custom-${field.key}`} type="date" value={String(value || '')} onChange={(e) => setValue(e.target.value)} className="mt-1 text-xs" />}
                          {field.field_type === 'select' && (
                            <select id={`custom-${field.key}`} value={String(value || '')} onChange={(e) => setValue(e.target.value)} className="mt-1 w-full rounded-lg border border-gray-200 bg-white px-2.5 py-1.5 text-xs text-gray-800 focus:outline-none focus:ring-2 focus:ring-primary-500">
                              <option value="">Select an option...</option>
                              {(field.options || []).map((opt: string) => <option key={opt} value={opt}>{opt}</option>)}
                            </select>
                          )}
                          {field.field_type === 'checkbox' && (
                            <label className="mt-1 flex items-center gap-2 cursor-pointer">
                              <input id={`custom-${field.key}`} type="checkbox" checked={!!value} onChange={(e) => setValue(e.target.checked)} className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                              <span className="text-xs text-gray-700">{field.label}</span>
                            </label>
                          )}
                        </div>
                      )
                    })}
                  </div>
                )}

                <div className="flex items-center justify-end gap-2 pt-3 border-t border-gray-100">
                  <Button variant="ghost" size="sm" onClick={() => setIsEditing(false)}>
                    <X className="h-3.5 w-3.5 mr-1" /> Cancel
                  </Button>
                  <Button variant="primary" size="sm" onClick={handleSaveProperties} disabled={updateContact.isPending}>
                    {updateContact.isPending ? <Loader2 className="h-3.5 w-3.5 animate-spin mr-1" /> : <Check className="h-3.5 w-3.5 mr-1" />}
                    Save Properties
                  </Button>
                </div>
              </div>
            ) : (
              <div className="divide-y divide-gray-100">
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">WhatsApp Number</span>
                  <span className="text-gray-900 font-mono text-xs select-all">{contact.number}</span>
                </div>
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Email</span>
                  <span className="text-gray-900 text-xs">
                    {contact.email ? <a href={`mailto:${contact.email}`} className="text-primary-600 hover:underline">{contact.email}</a> : <span className="text-gray-400">-</span>}
                  </span>
                </div>
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Pipeline</span>
                  <span className="text-gray-900 text-xs font-medium">Sales Pipeline</span>
                </div>
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Deal Stage</span>
                  <span className="text-gray-900 text-xs">
                    {contact.deal_stage ? (
                      <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-white text-xs font-semibold shadow-2xs" style={{ backgroundColor: contact.deal_stage.color || '#3b82f6' }}>
                        <StageIcon icon={contact.deal_stage.icon} />{contact.deal_stage.label}
                      </span>
                    ) : <span className="text-gray-400">Unassigned</span>}
                  </span>
                </div>
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Contact Type</span>
                  <span className="text-gray-900 text-xs font-medium">{contact.is_group ? 'WhatsApp Group' : 'Individual Contact'}</span>
                </div>
                <div className="px-4 py-2.5 flex flex-col gap-1.5 hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Tags</span>
                  {contact.tags && contact.tags.length > 0 ? (
                    <div className="flex flex-wrap gap-1.5 mt-0.5">
                      {contact.tags.map((t) => (
                        <span key={t} className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-gray-100 text-gray-700 text-xs font-medium border border-gray-200"><Tag className="h-3 w-3 text-gray-400" />{t}</span>
                      ))}
                    </div>
                  ) : <span className="text-gray-400 text-xs">No tags assigned</span>}
                </div>
                {fieldDefinitions.length > 0 && (
                  <>
                    <div className="px-4 py-2 bg-gray-50/50 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Custom Fields</div>
                    {fieldDefinitions.map((field) => {
                      const val = contact.custom_values?.[field.key]
                      return (
                        <div key={field.key} className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                          <span className="text-gray-500 font-medium text-xs">{field.label}</span>
                          <span className="text-gray-900 text-xs font-medium">
                            {field.field_type === 'checkbox' ? (
                              val ? <span className="inline-flex items-center gap-1 text-emerald-600 font-medium"><Check className="h-3.5 w-3.5" /> Yes</span> : <span className="text-gray-400">No</span>
                            ) : val !== undefined && val !== null && String(val).trim() !== '' ? String(val) : <span className="text-gray-400">-</span>}
                          </span>
                        </div>
                      )
                    })}
                  </>
                )}
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Created At</span>
                  <span className="text-gray-600 text-xs">{formatDate(contact.created_at)}</span>
                </div>
                <div className="px-4 py-2.5 flex items-center justify-between hover:bg-gray-50/50">
                  <span className="text-gray-500 font-medium text-xs">Last Updated</span>
                  <span className="text-gray-600 text-xs">{formatDate(contact.updated_at)}</span>
                </div>
              </div>
            )}
          </Card>
        </div>

        {/* RIGHT PANEL */}
        <div className="lg:col-span-7 space-y-4">
          {/* CRM Hub View Switcher Card */}
          <Card className="p-0 border border-gray-200/80 shadow-sm overflow-hidden">
            <div className="border-b border-gray-200 bg-gray-50/70 px-4">
              <nav className="-mb-px flex gap-6" aria-label="CRM Hub Tabs">
                {[
                  { key: 'summary', label: 'Summary Feed', icon: ActivityIcon, count: 0 },
                  { key: 'tasks', label: 'Tasks & Follow-ups', icon: CheckSquare, count: openTaskCount },
                  { key: 'tickets', label: 'Tickets & Conversations', icon: MessageSquare, count: conversations.length },
                ].map(({ key, label, icon: Icon, count }) => (
                  <button
                    key={key}
                    onClick={() => setActiveTab(key as HubTabKey)}
                    className={`inline-flex items-center gap-2 py-3 px-1 text-xs font-semibold border-b-2 transition-colors ${
                      activeTab === key ? 'border-primary-600 text-primary-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                    }`}
                  >
                    <Icon className="h-4 w-4" />
                    <span>{label}</span>
                    <span className={`px-1.5 py-0.2 rounded-full text-[10px] font-bold ${activeTab === key ? 'bg-primary-100 text-primary-700' : 'bg-gray-200/80 text-gray-600'}`}>{count}</span>
                  </button>
                ))}
              </nav>
            </div>

            {/* TAB 1: Summary Feed (manual-only) */}
            {activeTab === 'summary' && <SummaryFeed contactId={contactId} />}

            {/* TAB 2: Tasks & Follow-ups (manual-only) */}
            {activeTab === 'tasks' && (
              <div>
                {taskEntries.length === 0 ? (
                  <div className="p-4">
                    <div className="text-center py-8 px-4 text-gray-400">
                      <div className="w-10 h-10 rounded-full bg-gray-50 border border-gray-100 flex items-center justify-center mx-auto mb-3 text-gray-400">
                        <CheckSquare className="h-5 w-5 text-gray-400 stroke-[1.5]" />
                      </div>
                      <p className="text-xs font-medium text-gray-700">No tasks or follow-ups yet</p>
                      <p className="text-[11px] text-gray-400 mt-1 max-w-sm mx-auto">
                        Manually add tasks and follow-ups for this contact using the form below.
                      </p>
                    </div>
                    <div className="border-t border-gray-100 pt-4">{renderTaskForm()}</div>
                  </div>
                ) : (
                  <div className="p-4 space-y-3">
                    {renderTaskForm()}
                    <div className="border-t border-gray-100 pt-3 space-y-0.5">
                      {taskEntries.map((task) => <TaskEntryRow key={task.id} task={task} onUpdate={handleUpdateTask} onDelete={handleDeleteTask} />)}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* TAB 3: Structured Conversation Tickets Table */}
            {activeTab === 'tickets' && (
              <div>
                {conversationsError ? (
                  <div className="text-center py-8">
                    <AlertCircle className="h-8 w-8 text-red-400 mx-auto mb-2" />
                    <p className="text-xs text-gray-600 font-medium">Failed to load conversation history.</p>
                    <Button variant="secondary" size="sm" className="mt-3 text-xs" onClick={() => refetchConversations()}>Retry</Button>
                  </div>
                ) : conversationsLoading ? (
                  <div className="flex justify-center py-10"><Loader2 className="h-6 w-6 animate-spin text-primary-500" /></div>
                ) : conversations.length === 0 ? (
                  <div className="text-center py-12 px-4 text-gray-400">
                    <div className="w-10 h-10 rounded-full bg-gray-50 border border-gray-100 flex items-center justify-center mx-auto mb-3 text-gray-400">
                      <MessageSquare className="h-5 w-5 text-gray-400 stroke-[1.5]" />
                    </div>
                    <p className="text-xs font-medium text-gray-700">No conversations recorded yet</p>
                    <p className="text-[11px] text-gray-400 mt-1 max-w-sm mx-auto">Incoming and outgoing WhatsApp messaging tickets will appear here when conversations start.</p>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200/80">
                      <thead className="bg-gray-50/70">
                        <tr>
                          <th className="px-4 py-2 text-left text-[10px] font-semibold text-gray-500 uppercase tracking-wider">Ticket #</th>
                          <th className="px-4 py-2 text-left text-[10px] font-semibold text-gray-500 uppercase tracking-wider">Status</th>
                          <th className="px-4 py-2 text-left text-[10px] font-semibold text-gray-500 uppercase tracking-wider">Assignee</th>
                          <th className="px-4 py-2 text-left text-[10px] font-semibold text-gray-500 uppercase tracking-wider">Started</th>
                          <th className="px-4 py-2 text-left text-[10px] font-semibold text-gray-500 uppercase tracking-wider">Last Active</th>
                          <th className="px-4 py-2 text-right text-[10px] font-semibold text-gray-500 uppercase tracking-wider">Actions</th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200/80">
                        {conversations.map((c: Conversation) => (
                          <tr key={c.id} className="hover:bg-gray-50/70 transition-colors">
                            <td className="px-4 py-2.5 whitespace-nowrap text-xs font-semibold text-gray-900">#{c.ticket_number}</td>
                            <td className="px-4 py-2.5 whitespace-nowrap text-xs">
                              <span className={`px-2 py-0.5 rounded-full text-[10px] font-semibold ${statusBadgeStyles[c.status] || 'bg-gray-100 text-gray-700'}`}>{c.status.replace(/_/g, ' ')}</span>
                            </td>
                            <td className="px-4 py-2.5 whitespace-nowrap text-xs text-gray-600">
                              {c.assignee || <span className="text-gray-400">Unassigned</span>}
                            </td>
                            <td className="px-4 py-2.5 whitespace-nowrap text-[11px] text-gray-500">{formatDate(c.created_at)}</td>
                            <td className="px-4 py-2.5 whitespace-nowrap text-[11px] text-gray-500">{formatDate(c.last_activity_at)}</td>
                            <td className="px-4 py-2.5 whitespace-nowrap text-right">
                              <Link
                                to="/conversations/$id"
                                params={{ id: c.id }}
                                className="inline-flex items-center gap-1 text-[11px] font-medium text-primary-600 hover:text-primary-700"
                              >
                                <ExternalLink className="h-3 w-3" />
                                Open
                              </Link>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            )}
          </Card>
        </div>
      </div>
    </div>
  )
}

export default ContactDetail
