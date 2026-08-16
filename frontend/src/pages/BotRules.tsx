import React, { useState } from 'react'
import { useDashboardBotRules, useCreateBotRuleSet, useActivateBotRuleSet } from '@/hooks/useDashboard'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import { Loader2, Bot, CheckCircle2, AlertCircle, Save, Rocket, RotateCcw } from 'lucide-react'

const defaultRule = {
  name: '',
  pattern: '',
  match: 'CONTAINS',
  response: '',
  terminal: false,
  handoff: false,
  enabled: true,
}

const BotRules: React.FC = () => {
  const { data: ruleSets = [], isLoading, isError, refetch } = useDashboardBotRules()
  const create = useCreateBotRuleSet()
  const activate = useActivateBotRuleSet()
  const [draft, setDraft] = useState([{ ...defaultRule }])
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const handleAddRule = () => setDraft([...draft, { ...defaultRule }])
  const handleUpdate = (idx: number, field: string, value: any) => {
    const next = [...draft]
    next[idx] = { ...next[idx], [field]: value }
    setDraft(next)
  }
  const handleRemove = (idx: number) => {
    const next = [...draft]
    next.splice(idx, 1)
    setDraft(next)
  }

  const handleSave = () => {
    setError(null)
    setSuccess(null)
    for (let i = 0; i < draft.length; i++) {
      const r = draft[i]
      if (!r.name || !r.pattern || !r.response) {
        setError(`Rule ${i + 1}: name, pattern, and response are required`)
        return
      }
      if (!['CONTAINS', 'EXACT', 'PREFIX'].includes(r.match)) {
        setError(`Rule ${i + 1}: match must be CONTAINS, EXACT, or PREFIX`)
        return
      }
    }
    create.mutate(draft, {
      onSuccess: () => {
        setSuccess('Ruleset saved as a new version.')
        setDraft([{ ...defaultRule }])
      },
      onError: (e: any) => setError(e?.message || 'Failed to save ruleset'),
    })
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">Bot Rules</h1>
          <p className="text-sm text-gray-600">Manage conversation bot rules and versions</p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => refetch()}>
          <RotateCcw className="h-4 w-4 mr-1" /> Refresh
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
        </div>
      ) : isError ? (
        <div className="text-center py-12 text-red-500">
          <AlertCircle className="h-12 w-12 mx-auto mb-3" />
          <p>Failed to load bot rules.</p>
          <Button variant="primary" size="sm" className="mt-4" onClick={() => refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <div className="space-y-6">
          <Card className="p-4">
            <h2 className="text-lg font-medium text-gray-900 mb-3 flex items-center gap-2">
              <Bot className="h-5 w-5" /> Versions
            </h2>
            {ruleSets.length === 0 ? (
              <p className="text-gray-500">No rulesets yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Version</th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Rules</th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Active</th>
                      <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {ruleSets.map((s) => (
                      <tr key={s.id}>
                        <td className="px-4 py-2 text-sm text-gray-900">{s.version}</td>
                        <td className="px-4 py-2 text-sm text-gray-500">{s.rules.length}</td>
                        <td className="px-4 py-2">
                          {s.is_active ? (
                            <span className="flex items-center text-green-600 text-sm font-medium">
                              <CheckCircle2 className="h-4 w-4 mr-1" /> Active
                            </span>
                          ) : (
                            <span className="text-gray-400 text-sm">Inactive</span>
                          )}
                        </td>
                        <td className="px-4 py-2">
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => activate.mutate(s.version)}
                            disabled={s.is_active}
                          >
                            <Rocket className="h-4 w-4 mr-1" /> Activate
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>

          <Card className="p-4">
            <h2 className="text-lg font-medium text-gray-900 mb-3">Draft New Version</h2>
            {error && <div className="mb-4 p-2 bg-red-50 text-red-700 rounded-md text-sm">{error}</div>}
            {success && <div className="mb-4 p-2 bg-green-50 text-green-700 rounded-md text-sm">{success}</div>}
            <div className="space-y-4">
              {draft.map((rule, idx) => (
                <div key={idx} className="grid grid-cols-1 md:grid-cols-5 gap-3 p-3 border border-gray-200 rounded-md">
                  <input
                    className="form-control"
                    placeholder="Name"
                    value={rule.name}
                    onChange={(e) => handleUpdate(idx, 'name', e.target.value)}
                  />
                  <input
                    className="form-control"
                    placeholder="Pattern"
                    value={rule.pattern}
                    onChange={(e) => handleUpdate(idx, 'pattern', e.target.value)}
                  />
                  <select
                    className="form-control"
                    value={rule.match}
                    onChange={(e) => handleUpdate(idx, 'match', e.target.value)}
                  >
                    <option value="CONTAINS">CONTAINS</option>
                    <option value="EXACT">EXACT</option>
                    <option value="PREFIX">PREFIX</option>
                  </select>
                  <input
                    className="form-control"
                    placeholder="Response"
                    value={rule.response}
                    onChange={(e) => handleUpdate(idx, 'response', e.target.value)}
                  />
                  <div className="flex items-center gap-2">
                    <label className="flex items-center text-sm">
                      <input
                        type="checkbox"
                        className="mr-1"
                        checked={rule.terminal}
                        onChange={(e) => handleUpdate(idx, 'terminal', e.target.checked)}
                      />
                      Terminal
                    </label>
                    <button className="text-red-500 text-sm" onClick={() => handleRemove(idx)}>
                      Remove
                    </button>
                  </div>
                </div>
              ))}
            </div>
            <div className="mt-4 flex gap-2">
              <Button variant="secondary" size="sm" onClick={handleAddRule}>
                Add Rule
              </Button>
              <Button variant="primary" size="sm" onClick={handleSave} disabled={create.isPending}>
                {create.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4 mr-1" />}
                Save Draft as New Version
              </Button>
            </div>
          </Card>
        </div>
      )}
    </div>
  )
}

export default BotRules
