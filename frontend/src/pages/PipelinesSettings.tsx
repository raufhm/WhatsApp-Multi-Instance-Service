import React, { useState } from 'react'
import {
  usePipelines,
  useDealStages,
  useCreateDealStage,
  useUpdateDealStage,
  useDeleteDealStage,
  useCreatePipeline,
  useUpdatePipeline,
  useDeletePipeline,
} from '@/hooks/useInbox'
import type { DealStage, Pipeline } from '@/types'
import { StageIcon } from '@/components/DealPipelineTracker'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  GitBranch,
  Plus,
  Edit2,
  Trash2,
  Check,
  X,
  Trophy,
  AlertOctagon,
  Layers,
  Loader2,
  Star,
} from 'lucide-react'

const PRESET_COLORS = [
  '#3b82f6', // blue
  '#6366f1', // indigo
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#f59e0b', // amber
  '#10b981', // emerald
  '#ef4444', // red
  '#14b8a6', // teal
  '#64748b', // slate
  '#0284c7', // light blue
]

const AVAILABLE_ICONS = [
  { value: 'user-plus', label: 'User / Lead' },
  { value: 'calendar', label: 'Calendar / Scheduled' },
  { value: 'flame', label: 'Flame / Hot' },
  { value: 'snowflake', label: 'Snowflake / Cold' },
  { value: 'spinner', label: 'In Progress / Spinner' },
  { value: 'trophy', label: 'Trophy / Won' },
  { value: 'x-circle', label: 'Cross / Lost' },
]

export default function PipelinesSettings() {
  const { data: pipelines } = usePipelines()

  // Pipeline selection must come before stage query
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingStage, setEditingStage] = useState<DealStage | null>(null)
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null)
  const [selectedPipelineId, setSelectedPipelineId] = useState<string>('')
  const [pipelineModalOpen, setPipelineModalOpen] = useState(false)
  const [editingPipeline, setEditingPipeline] = useState<Pipeline | null>(null)
  const [pipelineDeleteId, setPipelineDeleteId] = useState<string | null>(null)

  const defaultPipeline = pipelines?.find((p) => p.is_default)
  const activePipeline = pipelines?.find((p) => p.id === selectedPipelineId) || defaultPipeline || pipelines?.[0]

  const { data: stages, isLoading, isError, refetch } = useDealStages(selectedPipelineId)
  const createStage = useCreateDealStage()
  const updateStage = useUpdateDealStage()
  const deleteStage = useDeleteDealStage()
  const createPipeline = useCreatePipeline()
  const updatePipeline = useUpdatePipeline()
  const deletePipeline = useDeletePipeline()

  React.useEffect(() => {
    if (!selectedPipelineId && activePipeline?.id) {
      setSelectedPipelineId(activePipeline.id)
    }
  }, [pipelines, selectedPipelineId, activePipeline?.id])

  // Pipeline form state
  const [pipelineForm, setPipelineForm] = useState({
    name: '',
    description: '',
    is_default: false,
    is_active: true,
  })

  const handleOpenPipelineCreate = () => {
    setEditingPipeline(null)
    setPipelineForm({ name: '', description: '', is_default: false, is_active: true })
    setPipelineModalOpen(true)
  }

  const handleOpenPipelineEdit = (pipeline: Pipeline) => {
    setEditingPipeline(pipeline)
    setPipelineForm({
      name: pipeline.name,
      description: pipeline.description || '',
      is_default: pipeline.is_default,
      is_active: pipeline.is_active ?? true,
    })
    setPipelineModalOpen(true)
  }

  const handleSavePipeline = (e: React.FormEvent) => {
    e.preventDefault()
    if (!pipelineForm.name.trim()) return
    if (editingPipeline) {
      updatePipeline.mutate(
        { id: editingPipeline.id, ...pipelineForm },
        { onSuccess: () => setPipelineModalOpen(false) }
      )
    } else {
      createPipeline.mutate(
        { name: pipelineForm.name.trim(), description: pipelineForm.description, is_default: pipelineForm.is_default },
        { onSuccess: () => setPipelineModalOpen(false) }
      )
    }
  }

  const handleDeletePipeline = (id: string) => {
    deletePipeline.mutate(id, {
      onSuccess: () => {
        setPipelineDeleteId(null)
        if (selectedPipelineId === id) setSelectedPipelineId('')
      },
    })
  }

  // Modal form state
  const [formKey, setFormKey] = useState('')
  const [formLabel, setFormLabel] = useState('')
  const [formColor, setFormColor] = useState('#3b82f6')
  const [formIcon, setFormIcon] = useState('user-plus')
  const [formSortOrder, setFormSortOrder] = useState(1)
  const [formIsWon, setFormIsWon] = useState(false)
  const [formIsLost, setFormIsLost] = useState(false)
  const [formIsActive, setFormIsActive] = useState(true)
  const [formError, setFormError] = useState<string | null>(null)

  const handleOpenCreate = () => {
    setEditingStage(null)
    setFormKey('')
    setFormLabel('')
    setFormColor('#3b82f6')
    setFormIcon('user-plus')
    setFormSortOrder((stages?.length || 0) + 1)
    setFormIsWon(false)
    setFormIsLost(false)
    setFormIsActive(true)
    setFormError(null)
    setIsModalOpen(true)
  }

  const handleOpenEdit = (stage: DealStage) => {
    setEditingStage(stage)
    setFormKey(stage.key)
    setFormLabel(stage.label)
    setFormColor(stage.color || '#3b82f6')
    setFormIcon(stage.icon || 'user-plus')
    setFormSortOrder(stage.sort_order)
    setFormIsWon(stage.is_won)
    setFormIsLost(stage.is_lost)
    setFormIsActive(stage.is_active)
    setFormError(null)
    setIsModalOpen(true)
  }

  const handleSaveStage = (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)

    if (!formLabel.trim()) {
      setFormError('Stage label is required.')
      return
    }

    if (!editingStage && !formKey.trim()) {
      setFormError('Stage key is required.')
      return
    }

    const formattedKey = formKey
      .trim()
      .toUpperCase()
      .replace(/[^A-Z0-9_]/g, '_')

    const pipelineId = selectedPipelineId || activePipeline?.id
    if (!pipelineId) {
      setFormError('Please select or create a pipeline first.')
      return
    }

    if (editingStage) {
      updateStage.mutate(
        {
          id: editingStage.id,
          pipeline_id: pipelineId,
          label: formLabel.trim(),
          color: formColor,
          icon: formIcon,
          sort_order: Number(formSortOrder),
          is_active: formIsActive,
          is_won: formIsWon,
          is_lost: formIsLost,
        },
        {
          onSuccess: () => {
            setIsModalOpen(false)
          },
          onError: (err: any) => {
            setFormError(err?.response?.data?.message || 'Failed to update stage.')
          },
        }
      )
    } else {
      createStage.mutate(
        {
          pipeline_id: pipelineId,
          key: formattedKey,
          label: formLabel.trim(),
          color: formColor,
          icon: formIcon,
          sort_order: Number(formSortOrder),
          is_won: formIsWon,
          is_lost: formIsLost,
        },
        {
          onSuccess: () => {
            setIsModalOpen(false)
          },
          onError: (err: any) => {
            setFormError(err?.response?.data?.message || 'Failed to create stage.')
          },
        }
      )
    }
  }

  const handleDelete = (id: string) => {
    deleteStage.mutate(id, {
      onSuccess: () => {
        setDeleteConfirmId(null)
      },
    })
  }

  const sortedStages = [...(stages || [])].sort((a, b) => a.sort_order - b.sort_order)

  return (
    <div className="space-y-6 pb-12">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-gray-200 pb-5">
        <div>
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-lg bg-primary-50 text-primary-600 border border-primary-200">
              <GitBranch className="h-5 w-5" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-gray-900">Pipelines & Deal Stages</h1>
              <p className="text-xs text-gray-500 mt-0.5">
                Configure your sales pipelines, progression stages, custom branding colors, and conversion rules.
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="primary" size="sm" onClick={handleOpenCreate} className="shadow-sm">
            <Plus className="h-4 w-4 mr-1.5" />
            Add Stage
          </Button>
        </div>
      </div>

      {/* Pipeline Management */}
      <Card className="p-4 border border-gray-200/80 shadow-xs bg-white">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-3">
          <div>
            <h3 className="text-sm font-semibold text-gray-900">Pipelines</h3>
            <p className="text-xs text-gray-500 mt-0.5">Select a pipeline to manage its stages.</p>
          </div>
          <Button variant="primary" size="sm" onClick={handleOpenPipelineCreate} className="shadow-sm">
            <Plus className="h-4 w-4 mr-1.5" />
            Add Pipeline
          </Button>
        </div>

        {pipelines && pipelines.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-gray-50/80 text-gray-500 uppercase tracking-wider text-[11px] border-b border-gray-100">
                <tr>
                  <th className="py-2 px-3 font-semibold">Name</th>
                  <th className="py-2 px-3 font-semibold">Description</th>
                  <th className="py-2 px-3 font-semibold">Default</th>
                  <th className="py-2 px-3 font-semibold">Active</th>
                  <th className="py-2 px-3 font-semibold text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {pipelines.map((pipeline) => (
                  <tr
                    key={pipeline.id}
                    className={`${selectedPipelineId === pipeline.id ? 'bg-primary-50/60' : 'hover:bg-gray-50/60'} transition-colors`}
                  >
                    <td className="py-2 px-3">
                      <button
                        type="button"
                        onClick={() => setSelectedPipelineId(pipeline.id)}
                        className="text-left font-medium text-gray-900 hover:text-primary-600"
                      >
                        {pipeline.name}
                      </button>
                    </td>
                    <td className="py-2 px-3 text-gray-500 max-w-xs truncate">{pipeline.description}</td>
                    <td className="py-2 px-3">
                      {pipeline.is_default ? (
                        <span className="inline-flex items-center gap-1 text-amber-600 text-[11px] font-medium">
                          <Star className="h-3 w-3" /> Default
                        </span>
                      ) : (
                        <span className="text-gray-400 text-[11px]">-</span>
                      )}
                    </td>
                    <td className="py-2 px-3">
                      {pipeline.is_active ? (
                        <span className="inline-flex items-center gap-1 text-emerald-600 text-[11px] font-medium">
                          <Check className="h-3 w-3" /> Active
                        </span>
                      ) : (
                        <span className="text-gray-400 text-[11px]">Inactive</span>
                      )}
                    </td>
                    <td className="py-2 px-3 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleOpenPipelineEdit(pipeline)}
                          className="h-7 px-2 text-gray-600 hover:text-gray-900"
                          title="Edit Pipeline"
                        >
                          <Edit2 className="h-3.5 w-3.5" />
                        </Button>
                        {pipelineDeleteId === pipeline.id ? (
                          <div className="flex items-center gap-1">
                            <Button
                              variant="danger"
                              size="sm"
                              onClick={() => handleDeletePipeline(pipeline.id)}
                              disabled={deletePipeline.isPending}
                              className="h-7 px-2 text-[11px]"
                            >
                              Confirm
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setPipelineDeleteId(null)}
                              className="h-7 px-1 text-gray-400 hover:text-gray-600"
                            >
                              <X className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setPipelineDeleteId(pipeline.id)}
                            className="h-7 px-2 text-gray-400 hover:text-red-600"
                            title="Delete Pipeline"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <div className="text-center py-6 text-gray-500 text-xs">
            No pipelines found. Create one to get started.
          </div>
        )}
      </Card>

      {/* Active Pipeline Card */}
      <Card className="p-4 border border-gray-200/80 shadow-xs bg-white">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-lg bg-indigo-50 text-indigo-600 border border-indigo-100">
              <Layers className="h-5 w-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-gray-900">
                  {activePipeline?.name || 'Sales Pipeline'}
                </h3>
                {activePipeline?.is_default && (
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                    Default Active Pipeline
                  </span>
                )}
              </div>
              <p className="text-xs text-gray-500 mt-0.5">
                {activePipeline?.description || 'Standard customer sales and deal qualification pipeline'}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-3 text-xs text-gray-600 bg-gray-50 px-3 py-1.5 rounded-lg border border-gray-200/60">
            <div>
              <span className="font-semibold text-gray-900">{stages?.length || 0}</span> Total Stages
            </div>
            <span className="text-gray-300">•</span>
            <div>
              <span className="font-semibold text-emerald-600">
                {stages?.filter((s) => s.is_won).length || 0}
              </span> Won
            </div>
            <span className="text-gray-300">•</span>
            <div>
              <span className="font-semibold text-red-600">
                {stages?.filter((s) => s.is_lost).length || 0}
              </span> Lost
            </div>
          </div>
        </div>
      </Card>

      {/* Stages Table List */}
      <Card className="border border-gray-200/80 shadow-xs overflow-hidden bg-white">
        <div className="px-4 py-3 border-b border-gray-100 bg-gray-50/50 flex items-center justify-between">
          <h3 className="text-xs font-bold uppercase tracking-wider text-gray-700">
            Configured Deal Stages
          </h3>
          <span className="text-xs text-gray-500">
            Stages appear in order of Sort Order in the contact dropdown and tracker.
          </span>
        </div>

        {isLoading ? (
          <div className="p-8 text-center">
            <Loader2 className="h-6 w-6 text-primary-600 animate-spin mx-auto mb-2" />
            <p className="text-xs text-gray-500">Loading deal stages…</p>
          </div>
        ) : isError ? (
          <div className="p-8 text-center text-red-600">
            <p className="text-xs font-semibold">Failed to load deal stages.</p>
            <Button variant="default" size="sm" onClick={() => refetch()} className="mt-2 text-xs">
              Retry
            </Button>
          </div>
        ) : sortedStages.length === 0 ? (
          <div className="p-8 text-center">
            <GitBranch className="h-8 w-8 text-gray-300 mx-auto mb-2" />
            <p className="text-xs font-semibold text-gray-700">No Deal Stages Found</p>
            <p className="text-xs text-gray-500 mt-1">Get started by creating your first stage.</p>
            <Button variant="primary" size="sm" onClick={handleOpenCreate} className="mt-3 text-xs">
              <Plus className="h-3.5 w-3.5 mr-1" /> Add Stage
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-gray-50/80 text-gray-500 uppercase tracking-wider text-[11px] border-b border-gray-100">
                <tr>
                  <th className="py-2.5 px-4 font-semibold w-16">Order</th>
                  <th className="py-2.5 px-4 font-semibold">Stage & Color</th>
                  <th className="py-2.5 px-4 font-semibold">Stage Key</th>
                  <th className="py-2.5 px-4 font-semibold">Stage Type</th>
                  <th className="py-2.5 px-4 font-semibold">Status</th>
                  <th className="py-2.5 px-4 font-semibold text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
                {sortedStages.map((stage) => (
                  <tr key={stage.id} className="hover:bg-gray-50/60 transition-colors">
                    {/* Sort Order */}
                    <td className="py-3 px-4 font-mono font-bold text-gray-500">
                      #{stage.sort_order}
                    </td>

                    {/* Stage Label & Color */}
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        <span
                          className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-white font-medium text-xs shadow-2xs"
                          style={{ backgroundColor: stage.color || '#3b82f6' }}
                        >
                          <StageIcon icon={stage.icon} />
                          <span>{stage.label}</span>
                        </span>
                        <span className="text-[11px] text-gray-400 font-mono">
                          {stage.color}
                        </span>
                      </div>
                    </td>

                    {/* Stage Key */}
                    <td className="py-3 px-4">
                      <span className="font-mono text-[11px] bg-gray-100 text-gray-700 px-2 py-0.5 rounded border border-gray-200">
                        {stage.key}
                      </span>
                    </td>

                    {/* Stage Nature */}
                    <td className="py-3 px-4">
                      {stage.is_won ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                          <Trophy className="h-3 w-3" /> Closed Won
                        </span>
                      ) : stage.is_lost ? (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-semibold bg-red-50 text-red-700 border border-red-200">
                          <AlertOctagon className="h-3 w-3" /> Closed Lost
                        </span>
                      ) : (
                        <span className="text-gray-500 text-[11px]">In Progress</span>
                      )}
                    </td>

                    {/* Status */}
                    <td className="py-3 px-4">
                      {stage.is_active ? (
                        <span className="inline-flex items-center gap-1 text-emerald-600 font-medium text-[11px]">
                          <Check className="h-3.5 w-3.5" /> Active
                        </span>
                      ) : (
                        <span className="text-gray-400 text-[11px]">Inactive</span>
                      )}
                    </td>

                    {/* Actions */}
                    <td className="py-3 px-4 text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleOpenEdit(stage)}
                          className="h-7 px-2 text-gray-600 hover:text-gray-900"
                          title="Edit Stage"
                        >
                          <Edit2 className="h-3.5 w-3.5" />
                        </Button>

                        {deleteConfirmId === stage.id ? (
                          <div className="flex items-center gap-1">
                            <Button
                              variant="danger"
                              size="sm"
                              onClick={() => handleDelete(stage.id)}
                              disabled={deleteStage.isPending}
                              className="h-7 px-2 text-[11px]"
                            >
                              Confirm
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setDeleteConfirmId(null)}
                              className="h-7 px-1 text-gray-400 hover:text-gray-600"
                            >
                              <X className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        ) : (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setDeleteConfirmId(stage.id)}
                            className="h-7 px-2 text-gray-400 hover:text-red-600"
                            title="Delete Stage"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Add / Edit Stage Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-xs">
          <div className="bg-white rounded-xl shadow-xl border border-gray-200 max-w-md w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
              <h3 className="text-sm font-bold text-gray-900">
                {editingStage ? 'Edit Deal Stage' : 'Add New Deal Stage'}
              </h3>
              <button
                type="button"
                onClick={() => setIsModalOpen(false)}
                className="text-gray-400 hover:text-gray-600 p-1 rounded-lg"
              >
                <X className="h-4 w-4" />
              </button>
            </div>

            <form onSubmit={handleSaveStage} className="p-5 space-y-4">
              {formError && (
                <div className="p-2.5 rounded-lg bg-red-50 border border-red-200 text-xs text-red-700">
                  {formError}
                </div>
              )}

              {/* Stage Key */}
              <div>
                <Label htmlFor="stage-key" className="text-xs font-semibold text-gray-700">
                  Stage Key (Unique Code)
                </Label>
                <Input
                  id="stage-key"
                  value={formKey}
                  disabled={!!editingStage}
                  onChange={(e) => setFormKey(e.target.value)}
                  placeholder="e.g. DEMO_COMPLETED"
                  className="mt-1 font-mono text-xs uppercase"
                  required
                />
                <p className="text-[10px] text-gray-400 mt-1">
                  Unique machine key (e.g. NEW_LEAD, DEMO_DONE). Cannot be changed after creation.
                </p>
              </div>

              {/* Stage Label */}
              <div>
                <Label htmlFor="stage-label" className="text-xs font-semibold text-gray-700">
                  Stage Display Label
                </Label>
                <Input
                  id="stage-label"
                  value={formLabel}
                  onChange={(e) => setFormLabel(e.target.value)}
                  placeholder="e.g. Demo Completed"
                  className="mt-1 text-xs"
                  required
                />
              </div>

              {/* Color Picker & Swatches */}
              <div>
                <Label className="text-xs font-semibold text-gray-700">Stage Badge Color</Label>
                <div className="flex items-center gap-2 mt-1">
                  <input
                    type="color"
                    value={formColor}
                    onChange={(e) => setFormColor(e.target.value)}
                    className="h-8 w-8 rounded border border-gray-300 cursor-pointer p-0.5"
                  />
                  <Input
                    value={formColor}
                    onChange={(e) => setFormColor(e.target.value)}
                    className="w-28 text-xs font-mono"
                    placeholder="#3b82f6"
                  />
                  <div className="flex items-center gap-1 flex-wrap">
                    {PRESET_COLORS.map((c) => (
                      <button
                        key={c}
                        type="button"
                        onClick={() => setFormColor(c)}
                        className={`h-5 w-5 rounded-full transition-transform ${
                          formColor === c ? 'scale-125 ring-2 ring-offset-1 ring-gray-400' : ''
                        }`}
                        style={{ backgroundColor: c }}
                      />
                    ))}
                  </div>
                </div>
              </div>

              {/* Icon Selector */}
              <div>
                <Label htmlFor="stage-icon" className="text-xs font-semibold text-gray-700">
                  Icon
                </Label>
                <select
                  id="stage-icon"
                  value={formIcon}
                  onChange={(e) => setFormIcon(e.target.value)}
                  className="mt-1 w-full rounded-md border border-gray-200 bg-white px-3 py-1.5 text-xs text-gray-800 focus:outline-none focus:ring-1 focus:ring-primary-500"
                >
                  {AVAILABLE_ICONS.map((ico) => (
                    <option key={ico.value} value={ico.value}>
                      {ico.label} ({ico.value})
                    </option>
                  ))}
                </select>
              </div>

              {/* Sort Order */}
              <div>
                <Label htmlFor="stage-order" className="text-xs font-semibold text-gray-700">
                  Sort Order / Progression Sequence
                </Label>
                <Input
                  id="stage-order"
                  type="number"
                  value={formSortOrder}
                  onChange={(e) => setFormSortOrder(Number(e.target.value))}
                  min={1}
                  className="mt-1 text-xs"
                  required
                />
              </div>

              {/* Outcome Checkboxes */}
              <div className="space-y-2 pt-2 border-t border-gray-100">
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={formIsWon}
                    onChange={(e) => {
                      setFormIsWon(e.target.checked)
                      if (e.target.checked) setFormIsLost(false)
                    }}
                    className="h-4 w-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500"
                  />
                  <span className="text-xs font-medium text-gray-700">
                    Mark as Closed Won 🏆 (Positive conversion)
                  </span>
                </label>

                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={formIsLost}
                    onChange={(e) => {
                      setFormIsLost(e.target.checked)
                      if (e.target.checked) setFormIsWon(false)
                    }}
                    className="h-4 w-4 rounded border-gray-300 text-red-600 focus:ring-red-500"
                  />
                  <span className="text-xs font-medium text-gray-700">
                    Mark as Closed Lost ❌ (Terminal deal loss)
                  </span>
                </label>

                {editingStage && (
                  <label className="flex items-center gap-2 cursor-pointer pt-1">
                    <input
                      type="checkbox"
                      checked={formIsActive}
                      onChange={(e) => setFormIsActive(e.target.checked)}
                      className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-xs font-medium text-gray-700">
                      Active Stage (Visible in dropdowns)
                    </span>
                  </label>
                )}
              </div>

              {/* Modal Actions */}
              <div className="flex items-center justify-end gap-2 pt-4 border-t border-gray-100">
                <Button variant="ghost" size="sm" type="button" onClick={() => setIsModalOpen(false)}>
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  type="submit"
                  disabled={createStage.isPending || updateStage.isPending}
                >
                  {createStage.isPending || updateStage.isPending ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />
                  ) : (
                    <Check className="h-3.5 w-3.5 mr-1.5" />
                  )}
                  {editingStage ? 'Save Changes' : 'Create Stage'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
      {/* Pipeline Create / Edit Modal */}
      {pipelineModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-xs">
          <div className="bg-white rounded-xl shadow-xl border border-gray-200 max-w-md w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
              <h3 className="text-sm font-bold text-gray-900">
                {editingPipeline ? 'Edit Pipeline' : 'Add New Pipeline'}
              </h3>
              <button
                type="button"
                onClick={() => setPipelineModalOpen(false)}
                className="text-gray-400 hover:text-gray-600 p-1 rounded-lg"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
            <form
              onSubmit={handleSavePipeline}
              className="p-5 space-y-4"
            >
              <div>
                <Label htmlFor="pipeline-name" className="text-xs font-semibold text-gray-700">
                  Pipeline Name
                </Label>
                <Input
                  id="pipeline-name"
                  value={pipelineForm.name}
                  onChange={(e) => setPipelineForm((prev) => ({ ...prev, name: e.target.value }))}
                  placeholder="e.g. Support Pipeline"
                  className="mt-1 text-xs"
                  required
                />
              </div>
              <div>
                <Label htmlFor="pipeline-description" className="text-xs font-semibold text-gray-700">
                  Description
                </Label>
                <Input
                  id="pipeline-description"
                  value={pipelineForm.description}
                  onChange={(e) => setPipelineForm((prev) => ({ ...prev, description: e.target.value }))}
                  placeholder="Short description..."
                  className="mt-1 text-xs"
                />
              </div>
              <div className="flex items-center gap-2 pt-2">
                <input
                  id="pipeline-default"
                  type="checkbox"
                  checked={pipelineForm.is_default}
                  onChange={(e) =>
                    setPipelineForm((prev) => ({ ...prev, is_default: e.target.checked }))
                  }
                  className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                <Label htmlFor="pipeline-default" className="text-xs font-medium text-gray-700">
                  Set as default pipeline
                </Label>
              </div>
              {editingPipeline && (
                <div className="flex items-center gap-2">
                  <input
                    id="pipeline-active"
                    type="checkbox"
                    checked={pipelineForm.is_active}
                    onChange={(e) =>
                      setPipelineForm((prev) => ({ ...prev, is_active: e.target.checked }))
                    }
                    className="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                  <Label htmlFor="pipeline-active" className="text-xs font-medium text-gray-700">
                    Active
                  </Label>
                </div>
              )}
              <div className="flex items-center justify-end gap-2 pt-4 border-t border-gray-100">
                <Button
                  variant="ghost"
                  size="sm"
                  type="button"
                  onClick={() => setPipelineModalOpen(false)}
                >
                  Cancel
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  type="submit"
                  disabled={createPipeline.isPending || updatePipeline.isPending}
                >
                  {createPipeline.isPending || updatePipeline.isPending ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />
                  ) : (
                    <Check className="h-3.5 w-3.5 mr-1.5" />
                  )}
                  {editingPipeline ? 'Save Changes' : 'Create Pipeline'}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
