import React, { useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  usePipelines,
  useDealStages,
  useDealStageHistory,
  useMoveContactToStage,
} from '@/hooks/useInbox'
import type { DealStage, DealStageTransition } from '@/types'
import {
  GitBranch,
  Check,
  ChevronRight,
  Settings,
  Clock,
  AlertCircle,
  Loader2,
  Tag,
} from 'lucide-react'

// ── Inline SVG icons for stage representations ──────────────────
const STAGE_ICONS: Record<string, JSX.Element> = {
  'user-plus': (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" />
    </svg>
  ),
  calendar: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <rect x="3" y="4" width="18" height="18" rx="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" />
    </svg>
  ),
  flame: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <path d="M8.5 14.5A2.5 2.5 0 0011 12c0-1.38-.5-2-1-3-1.07-2.14 0-5.5 3-7-.5 2.5 1.5 4 2.5 5.5A4.5 4.5 0 0114 20a6 6 0 01-6-4c0-1.5.5-2 1.5-2.5z" />
    </svg>
  ),
  snowflake: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <line x1="12" y1="2" x2="12" y2="22" /><line x1="2" y1="12" x2="22" y2="12" /><line x1="4.93" y1="4.93" x2="19.07" y2="19.07" /><line x1="19.07" y1="4.93" x2="4.93" y2="19.07" />
    </svg>
  ),
  spinner: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <path d="M21 12a9 9 0 11-6.22-8.56" />
    </svg>
  ),
  trophy: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <path d="M6 9H4.5a2.5 2.5 0 010-5H6" /><path d="M18 9h1.5a2.5 2.5 0 000-5H18" /><path d="M4 22h16" /><path d="M10 14.66V17c0 .55-.47.98-.97 1.21C7.85 18.75 7 20 7 22" /><path d="M14 14.66V17c0 .55.47.98.97 1.21C16.15 18.75 17 20 17 22" /><path d="M18 2H6v7a6 6 0 0012 0V2z" />
    </svg>
  ),
  'x-circle': (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-3.5 h-3.5">
      <circle cx="12" cy="12" r="10" /><line x1="15" y1="9" x2="9" y2="15" /><line x1="9" y1="9" x2="15" y2="15" />
    </svg>
  ),
}

export function StageIcon({ icon, className }: { icon: string; className?: string }) {
  return STAGE_ICONS[icon] ?? <span className={className || 'w-3.5 h-3.5 inline-block text-center font-bold'}>●</span>
}

interface DealPipelineTrackerProps {
  contactId: string
  currentStageKey?: string | null
  currentStageId?: string | null
}

export function DealPipelineTracker({ contactId, currentStageKey, currentStageId }: DealPipelineTrackerProps) {
  const { data: pipelines, isLoading: isPipelinesLoading } = usePipelines()
  const { data: stages, isLoading: isStagesLoading } = useDealStages()
  const { data: history } = useDealStageHistory(contactId)
  const moveStage = useMoveContactToStage(contactId)

  const [selectedPipelineKey, setSelectedPipelineKey] = useState<string>('sales')
  const [selectedTargetStage, setSelectedTargetStage] = useState<string>(currentStageKey || '')
  const [selectedTargetStageId, setSelectedTargetStageId] = useState<string>(currentStageId || '')
  const [moveNote, setMoveNote] = useState('')
  const [showNoteField, setShowNoteField] = useState(false)

  const activeStages = (stages ?? []).filter((s: DealStage) => s.is_active)
  const currentStage = activeStages.find((s: DealStage) => s.key === currentStageKey)
  const lastTransition: DealStageTransition | undefined = history?.[0]

  // Synchronize internal state when currentStageKey/id prop changes
  React.useEffect(() => {
    setSelectedTargetStage(currentStageKey || '')
  }, [currentStageKey])

  React.useEffect(() => {
    setSelectedTargetStageId(currentStageId || '')
  }, [currentStageId])

  const handleExecuteMove = (targetKey: string, targetId?: string, note?: string) => {
    if (!targetKey) return
    setSelectedTargetStage(targetKey)
    if (targetId) setSelectedTargetStageId(targetId)
    moveStage.mutate(
      { stageKey: targetKey, stageId: targetId, dealStageId: targetId, note: note !== undefined ? note : moveNote },
      {
        onSuccess: () => {
          setMoveNote('')
          setShowNoteField(false)
        },
      }
    )
  }

  const handleDropdownStageChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newStageKey = e.target.value
    const stage = activeStages.find((s: DealStage) => s.key === newStageKey)
    setSelectedTargetStage(newStageKey)
    if (stage && newStageKey && newStageKey !== currentStageKey) {
      handleExecuteMove(newStageKey, stage.id)
    }
  }

  if (isPipelinesLoading || isStagesLoading) {
    return (
      <div className="border border-gray-200/90 rounded-xl bg-white p-4 shadow-sm flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Loader2 className="h-4 w-4 text-primary-600 animate-spin" />
          <span className="text-xs text-gray-500 font-medium">Loading pipeline & deal stages…</span>
        </div>
      </div>
    )
  }

  if (!stages || stages.length === 0) {
    return (
      <div className="border border-dashed border-gray-300 rounded-xl bg-gray-50/70 p-4 text-center">
        <div className="flex flex-col items-center justify-center gap-1.5">
          <GitBranch className="h-5 w-5 text-gray-400" />
          <p className="text-xs font-semibold text-gray-700">No Deal Stages Configured</p>
          <p className="text-[11px] text-gray-500 max-w-sm">
            Set up your sales pipeline and deal stages in settings to start tracking lead progression.
          </p>
          <Link
            to="/settings/pipelines"
            className="mt-2 inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold rounded-lg bg-primary-600 text-white hover:bg-primary-700 transition-colors shadow-sm"
          >
            <Settings className="h-3.5 w-3.5" />
            Configure Deal Stages
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="border border-gray-200/90 rounded-xl bg-white p-4 shadow-sm space-y-3.5 transition-all">
      {/* 1. Header with Pipeline and Stage Dropdowns */}
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 pb-3">
        <div className="flex flex-wrap items-center gap-2.5 sm:gap-3">
          {/* Pipeline Dropdown */}
          <div className="flex items-center gap-1.5">
            <span className="text-[11px] font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-1">
              <GitBranch className="h-3.5 w-3.5 text-primary-600" />
              Pipeline:
            </span>
            <select
              id="pipeline-dropdown"
              aria-label="Pipeline"
              value={selectedPipelineKey}
              onChange={(e) => setSelectedPipelineKey(e.target.value)}
              className="rounded-md border border-gray-200 bg-gray-50/50 px-2.5 py-1 text-xs font-semibold text-gray-800 focus:border-primary-500 focus:bg-white focus:outline-none focus:ring-1 focus:ring-primary-500 cursor-pointer shadow-2xs"
            >
              {(pipelines || [{ id: 'sales', key: 'sales', name: 'Sales Pipeline' }]).map((pipe) => (
                <option key={pipe.key} value={pipe.key}>
                  {pipe.name}
                </option>
              ))}
            </select>
          </div>

          <span className="text-gray-300 hidden sm:inline">|</span>

          {/* Stage Dropdown */}
          <div className="flex items-center gap-1.5">
            <span className="text-[11px] font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-1">
              <Tag className="h-3.5 w-3.5 text-indigo-600" />
              Stage:
            </span>
            <div className="relative inline-flex items-center">
              <select
                id="stage-dropdown"
                aria-label="Deal Stage"
                value={selectedTargetStage}
                disabled={moveStage.isPending}
                onChange={handleDropdownStageChange}
                className="rounded-md border border-gray-200 bg-white px-2.5 py-1 text-xs font-semibold text-gray-800 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 cursor-pointer shadow-2xs pr-7 disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <option value="">-- Select Stage --</option>
                {activeStages.map((stage: DealStage) => (
                  <option key={stage.key} value={stage.key}>
                    {stage.label} {stage.is_won ? '🏆 (Won)' : stage.is_lost ? '❌ (Lost)' : ''}
                  </option>
                ))}
              </select>
              {moveStage.isPending && (
                <Loader2 className="h-3.5 w-3.5 text-primary-600 animate-spin absolute right-2 pointer-events-none" />
              )}
            </div>
          </div>

          {/* Note Toggle / Quick Note Trigger */}
          <button
            type="button"
            onClick={() => setShowNoteField((prev) => !prev)}
            className="text-[11px] text-gray-500 hover:text-primary-600 underline-offset-2 hover:underline transition-colors ml-1"
          >
            {showNoteField ? 'Hide note' : '+ Add transition note'}
          </button>
        </div>

        {/* Settings Shortcut Link */}
        <div className="flex items-center gap-2">
          <Link
            to="/settings/pipelines"
            className="inline-flex items-center gap-1 text-[11px] font-medium text-gray-500 hover:text-gray-800 transition-colors p-1 rounded hover:bg-gray-100"
            title="Configure Pipelines & Stages in Settings"
          >
            <Settings className="h-3 w-3 text-gray-400" />
            <span className="hidden md:inline">Settings</span>
          </Link>
        </div>
      </div>

      {/* Optional Note Field */}
      {showNoteField && (
        <div className="flex items-center gap-2 bg-gray-50 p-2 rounded-lg border border-gray-200/80 transition-all">
          <input
            type="text"
            value={moveNote}
            onChange={(e) => setMoveNote(e.target.value)}
            placeholder="Reason or note for stage change (e.g., Demo completed, Budget approved)..."
            className="flex-1 text-xs rounded border border-gray-200 px-2.5 py-1 bg-white focus:outline-none focus:ring-1 focus:ring-primary-500"
            onKeyDown={(e) => {
              if (e.key === 'Enter' && selectedTargetStage && selectedTargetStageId) {
                handleExecuteMove(selectedTargetStage, selectedTargetStageId, moveNote)
              }
            }}
          />
          <button
            type="button"
            disabled={!selectedTargetStage || moveStage.isPending}
            onClick={() => handleExecuteMove(selectedTargetStage, selectedTargetStageId, moveNote)}
            className="px-2.5 py-1 bg-primary-600 text-white rounded text-xs font-semibold hover:bg-primary-700 disabled:opacity-50 transition-colors flex items-center gap-1 shadow-2xs cursor-pointer"
          >
            {moveStage.isPending ? <Loader2 className="h-3 w-3 animate-spin" /> : <Check className="h-3 w-3" />}
            Save Note
          </button>
        </div>
      )}

      {/* 2. Interactive Clickable Visual Stepper */}
      <div className="pt-0.5">
        <div className="flex items-center gap-1 overflow-x-auto pb-1.5 pt-0.5 no-scrollbar">
          {activeStages.map((stage: DealStage, idx: number) => {
            const isCurrent = stage.key === currentStageKey
            const isTargetPending = moveStage.isPending && selectedTargetStage === stage.key
            const isPast =
              currentStage &&
              !stage.is_won &&
              !stage.is_lost &&
              stage.sort_order < currentStage.sort_order &&
              !currentStage.is_won &&
              !currentStage.is_lost

            let badgeStyle: React.CSSProperties = {}
            let badgeClass =
              'border bg-white text-gray-700 hover:bg-gray-50 border-gray-200 hover:border-gray-300 shadow-2xs'

            if (isCurrent) {
              badgeClass = 'text-white border-transparent ring-2 ring-offset-1 ring-primary-400 font-semibold shadow-sm'
              badgeStyle = { backgroundColor: stage.color || '#3b82f6' }
            } else if (isPast) {
              badgeClass = 'bg-emerald-50 text-emerald-800 border-emerald-200 hover:bg-emerald-100 font-medium'
            } else if (stage.is_won) {
              badgeClass = 'bg-emerald-50/60 text-emerald-700 border-emerald-200 hover:bg-emerald-100 font-medium'
            } else if (stage.is_lost) {
              badgeClass = 'bg-red-50/60 text-red-700 border-red-200 hover:bg-red-100 font-medium'
            }

            return (
              <div key={stage.id} className="flex items-center shrink-0">
                <button
                  type="button"
                  id={`stage-button-${stage.key}`}
                  disabled={moveStage.isPending}
                  onClick={() => {
                    const s = activeStages.find((st: DealStage) => st.key === stage.key)
                    handleExecuteMove(stage.key, s?.id)
                  }}
                  style={badgeStyle}
                  className={`
                    group relative flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs
                    cursor-pointer transition-all duration-150 select-none whitespace-nowrap
                    focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-1
                    disabled:cursor-not-allowed disabled:opacity-75
                    ${badgeClass}
                  `}
                  title={`Click to set stage to ${stage.label}`}
                >
                  {isTargetPending ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  ) : isPast ? (
                    <Check className="w-3.5 h-3.5 text-emerald-600" />
                  ) : (
                    <StageIcon icon={stage.icon} />
                  )}
                  <span>{stage.label}</span>
                </button>

                {idx < activeStages.length - 1 && (
                  <ChevronRight className="w-3.5 h-3.5 text-gray-300 mx-0.5 shrink-0" />
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* 3. Current Stage Summary & Transition Audit info */}
      <div className="flex flex-wrap items-center justify-between gap-2 pt-1 text-xs text-gray-500">
        <div className="flex items-center gap-2">
          {currentStage ? (
            <div className="flex items-center gap-1.5">
              <span className="text-gray-400 font-medium text-[11px]">Current Stage:</span>
              <span
                className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-white text-[11px] font-semibold shadow-2xs"
                style={{ backgroundColor: currentStage.color || '#3b82f6' }}
              >
                <StageIcon icon={currentStage.icon} />
                {currentStage.label}
              </span>
            </div>
          ) : (
            <div className="flex items-center gap-1.5 text-amber-700 bg-amber-50 px-2.5 py-0.5 rounded-full border border-amber-200 text-[11px] font-medium">
              <AlertCircle className="h-3 w-3 text-amber-600" />
              <span>No stage assigned. Select a stage from the dropdown or click a step above.</span>
            </div>
          )}

          {lastTransition && (
            <div className="hidden lg:flex items-center gap-1 text-gray-500 text-[11px] ml-2 pl-2 border-l border-gray-200">
              <Clock className="h-3 w-3 text-gray-400" />
              <span>
                {lastTransition.from_stage ? `Moved from ${lastTransition.from_stage}` : 'Assigned'}{' '}
                {lastTransition.note && <span className="text-gray-700 font-medium italic">"{lastTransition.note}"</span>}
                <span className="text-gray-400 ml-1">({new Date(lastTransition.created_at).toLocaleDateString()})</span>
              </span>
            </div>
          )}
        </div>

        {moveStage.isPending && (
          <div className="flex items-center gap-1.5 text-primary-600 text-xs font-medium animate-pulse">
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
            <span>Updating deal stage…</span>
          </div>
        )}
      </div>
    </div>
  )
}
