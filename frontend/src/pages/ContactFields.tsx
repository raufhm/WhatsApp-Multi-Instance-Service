import React, { useState } from 'react'
import Button from '@/components/ui/button'
import Card from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useContactFieldDefinitions, useCreateContactFieldDefinition, useUpdateContactFieldDefinition, useDeleteContactFieldDefinition } from '@/hooks/useInbox'
import { Loader2, Plus, Edit2, Trash2, Save, X } from 'lucide-react'

type FieldType = 'text' | 'number' | 'date' | 'select' | 'checkbox'

const ContactFieldsPage: React.FC = () => {
  const { data: fields = [], isLoading, refetch } = useContactFieldDefinitions()
  const createMutation = useCreateContactFieldDefinition()
  const updateMutation = useUpdateContactFieldDefinition()
  const deleteMutation = useDeleteContactFieldDefinition()

  const [editingId, setEditingId] = useState<string | null>(null)
  const [editForm, setEditForm] = useState<{ label: string; field_type: FieldType; options: string[]; is_required: boolean; sort_order: number }>({
    label: '',
    field_type: 'text',
    options: [],
    is_required: false,
    sort_order: 0,
  })
  const [newOptions, setNewOptions] = useState('')

  const handleAdd = () => {
    if (!editForm.label.trim()) return
    createMutation.mutate(
      {
        key: `field_${Date.now()}`,
        ...editForm,
      },
      { onSuccess: () => {
        setEditingId(null)
        setEditForm({ label: '', field_type: 'text', options: [], is_required: false, sort_order: 0 })
        refetch()
      }}
    )
  }

  const startEdit = (field: typeof fields[0]) => {
    setEditingId(field.id)
    setEditForm({
      label: field.label,
      field_type: field.field_type as FieldType,
      options: field.options || [],
      is_required: field.is_required,
      sort_order: field.sort_order,
    })
  }

  const cancelEdit = () => {
    setEditingId(null)
  }

  const handleSaveEdit = () => {
    if (!editingId || !editForm.label.trim()) return
    updateMutation.mutate({ id: editingId, ...editForm }, { onSuccess: () => {
      setEditingId(null)
      refetch()
    }})
  }

  const handleDelete = (id: string) => {
    if (confirm('Are you sure?')) {
      deleteMutation.mutate(id, { onSuccess: () => refetch() })
    }
  }

  const addOption = () => {
    if (!newOptions.trim()) return
    setEditForm((prev) => ({ ...prev, options: [...prev.options, newOptions.trim()] }))
    setNewOptions('')
  }

  const removeOption = (idx: number) => {
    setEditForm((prev) => ({ ...prev, options: prev.options.filter((_, i) => i !== idx) }))
  }

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Custom Contact Fields</h1>

      <Card>
        <div className="space-y-4 p-4 border-b border-gray-100">
          <Label>New Field</Label>
          <div className="grid grid-cols-2 gap-3">
            <Input
              value={editForm.label}
              onChange={(e) => setEditForm((prev) => ({ ...prev, label: e.target.value }))}
              placeholder="Field label"
            />
            <select
              value={editForm.field_type}
              onChange={(e) => setEditForm((prev) => ({ ...prev, field_type: e.target.value as FieldType }))}
              className="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm"
            >
              <option value="text">Text</option>
              <option value="number">Number</option>
              <option value="date">Date</option>
              <option value="select">Select Dropdown</option>
              <option value="checkbox">Checkbox</option>
            </select>
          </div>
          {(editForm.field_type === 'select' || editForm.field_type === 'checkbox') && (
            <>
              <div className="flex gap-2">
                <Input value={newOptions} onChange={(e) => setNewOptions(e.target.value)} placeholder="Option value" />
                <Button size="sm" onClick={addOption}>Add Option</Button>
              </div>
              <div className="flex flex-wrap gap-2">
                {editForm.options.map((opt, idx) => (
                  <span key={idx} className="px-2 py-0.5 rounded-full bg-gray-100 text-xs">{opt}
                    <button className="ml-1 text-red-500" onClick={() => removeOption(idx)}>×</button>
                  </span>
                ))}
              </div>
            </>
          )}
          <div className="flex items-center gap-4 text-sm">
            <label className="flex items-center gap-2">
              <input type="checkbox" checked={editForm.is_required} onChange={(e) => setEditForm((prev) => ({ ...prev, is_required: e.target.checked }))} />
              Required
            </label>
            <label className="flex items-center gap-2">
              Sort order:
              <Input type="number" min="0" value={editForm.sort_order.toString()} onChange={(e) => setEditForm((prev) => ({ ...prev, sort_order: Number(e.target.value) }))} className="w-20" />
            </label>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={handleAdd} disabled={createMutation.isPending || !editForm.label.trim()}>
              {createMutation.isPending ? <Loader2 className="h-4 w-4 animate-spin mr-1"/> : <Plus className="h-4 w-4 mr-1"/>}
              Add Field
            </Button>
          </div>
        </div>
      </Card>

      <Card>
        <div className="p-4">
          {isLoading ? (
            <div className="text-center py-8"><Loader2 className="h-8 w-8 animate-spin mx-auto"/></div>
          ) : fields.length === 0 ? (
            <div className="text-center py-8 text-gray-500">No custom fields defined.</div>
          ) : (
            <table className="min-w-full text-sm">
              <thead>
                <tr className="border-b border-gray-100">
                  <th className="text-left py-2 pr-4">Key</th>
                  <th className="text-left py-2 pr-4">Label</th>
                  <th className="text-left py-2 pr-4">Type</th>
                  <th className="text-left py-2 pr-4">Required</th>
                  <th className="text-left py-2 pr-4">Actions</th>
                </tr>
              </thead>
              <tbody>
                {fields.map((field) => (
                  <tr key={field.id} className="border-b border-gray-50 last:border-0">
                    <td className="py-3 pr-4 font-mono text-xs">{field.key}</td>
                    <td className="py-3 pr-4">
                      {editingId === field.id ? (
                        <Input
                          value={editForm.label}
                          onChange={(e) => setEditForm((prev) => ({ ...prev, label: e.target.value }))}
                          className="w-40"
                        />
                      ) : (
                        field.label
                      )}
                    </td>
                    <td className="py-3 pr-4">{field.field_type}</td>
                    <td className="py-3 pr-4">{field.is_required ? 'Yes' : 'No'}</td>
                    <td className="py-3 pr-4 flex items-center gap-2">
                      {editingId === field.id ? (
                        <>
                          <Button size="sm" variant="primary" onClick={handleSaveEdit}><Save className="h-3.5 w-3.5 mr-1"/>Save</Button>
                          <Button size="sm" variant="ghost" onClick={cancelEdit}><X className="h-3.5 w-3.5"/></Button>
                        </>
                      ) : (
                        <>
                          <Button size="sm" variant="ghost" onClick={() => startEdit(field)}><Edit2 className="h-3.5 w-3.5"/></Button>
                          <Button size="sm" variant="ghost" className="text-red-500" onClick={() => handleDelete(field.id)}><Trash2 className="h-3.5 w-3.5"/></Button>
                        </>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </Card>
    </div>
  )
}

export default ContactFieldsPage
