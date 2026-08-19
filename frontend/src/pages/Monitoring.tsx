import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import Card from '@/components/ui/card'
import Button from '@/components/ui/button'
import {
  useDashboardAccounts,
} from '@/hooks/useDashboard'
import {
  useMonitoringStatus,
  useMonitoringStatusEvents,
  useMonitoringMetrics,
  useMonitoringEvents,
  useMonitoringQueueDepth,
  useMonitoringStream,
} from '@/hooks/useMonitoring'
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  ChevronDown,
  Database,
  Loader2,
  Pause,
  Play,
  RefreshCw,
  ShieldAlert,
  Signal,
  Timer,
  Trash2,
  Wifi,
  WifiOff,
} from 'lucide-react'

interface AccountWithHealth {
  id: string
  host_id: string
  display_name: string
  health: 'healthy' | 'disconnected' | 'unknown'
  is_connected: boolean
  queue_size: number
}

const statusBadgeCls = (status: string) => {
  switch (status) {
    case 'ONLINE':
      return 'bg-green-100 text-green-800'
    case 'OFFLINE':
      return 'bg-red-100 text-red-800'
    case 'ERROR':
      return 'bg-amber-100 text-amber-800'
    default:
      return 'bg-gray-100 text-gray-700'
  }
}

const eventTypeColor = (eventType: string) => {
  if (eventType === 'MESSAGE_IN') return 'text-blue-600'
  if (eventType === 'MESSAGE_OUT') return 'text-emerald-600'
  if (eventType === 'STATUS') return 'text-purple-600'
  if (eventType === 'RECEIPT') return 'text-cyan-600'
  if (eventType === 'QUEUE_DEPTH') return 'text-gray-600'
  return 'text-red-600'
}

const eventLabel = (eventType: string, payload: Record<string, unknown> | null): string => {
  const p = payload ?? {}
  switch (eventType) {
    case 'MESSAGE_IN':
      return `IN ${p.type ?? 'text'}${p.status ? ` · ${p.status}` : ''}`
    case 'MESSAGE_OUT':
      return `OUT ${p.type ?? 'text'}${p.status ? ` · ${p.status}` : ''}`
    case 'STATUS':
      return `STATUS ${p.status ?? ''}${p.message ? ` · ${p.message}` : ''}`
    case 'RECEIPT':
      return `RECEIPT ${p.status ?? ''}`
    case 'QUEUE_DEPTH':
      return `QUEUE_DEPTH ${p.queue_size ?? 0}`
    case 'SEND_ERROR':
    case 'MEDIA_ERROR':
    case 'UPLOAD_FAILED':
    case 'PROJECTION_FAILED':
    case 'LOGGED_OUT':
      return `${eventType}${p.error ? ` · ${String(p.error)}` : ''}`
    default:
      return eventType
  }
}

function formatRelative(iso: string | undefined | null): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (diff < 0) return 'now'
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}s ago`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  const d = Math.floor(h / 24)
  return `${d}d ago`
}

function formatDateTime(iso: string | undefined | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

const WINDOWS = [
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
]

// ── Account Dropdown ────────────────────────────────────────────────────────

function AccountDropdown({
  accounts,
  selectedHost,
  onSelect,
}: {
  accounts: AccountWithHealth[]
  selectedHost: string | null
  onSelect: (hostId: string) => void
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const selected = accounts.find(a => a.host_id === selectedHost)

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(v => !v)}
        className="inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium bg-white border border-gray-200 rounded-lg shadow-sm hover:bg-gray-50 hover:border-gray-300 transition-colors max-w-[280px]"
      >
        <span
          className={`shrink-0 h-2 w-2 rounded-full ${
            selected?.is_connected ? 'bg-green-500' : 'bg-gray-300'
          }`}
        />
        <span className="truncate">
          {selected?.display_name || selected?.host_id || 'Select account'}
        </span>
        <ChevronDown className={`h-3.5 w-3.5 text-gray-400 shrink-0 transition-transform ${open ? 'rotate-180' : ''}`} />
      </button>

      {open && (
        <div className="absolute right-0 mt-1 w-72 bg-white border border-gray-200 rounded-lg shadow-lg z-20 py-1 max-h-72 overflow-y-auto">
          {accounts.length === 0 ? (
            <p className="px-3 py-2 text-sm text-gray-500">No accounts</p>
          ) : (
            accounts.map(a => (
              <button
                key={a.id}
                type="button"
                onClick={() => {
                  onSelect(a.host_id)
                  setOpen(false)
                }}
                className={`w-full flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-50 transition-colors ${
                  a.host_id === selectedHost ? 'bg-primary-50 text-primary-700' : 'text-gray-700'
                }`}
              >
                <span
                  className={`shrink-0 h-2 w-2 rounded-full ${
                    a.is_connected ? 'bg-green-500' : 'bg-gray-300'
                  }`}
                />
                <span className="truncate flex-1 text-left">
                  {a.display_name || a.host_id}
                </span>
                {a.is_connected ? (
                  <span className="shrink-0 text-[10px] font-medium text-green-600 bg-green-50 px-1.5 py-0.5 rounded">ONLINE</span>
                ) : (
                  <span className="shrink-0 text-[10px] font-medium text-gray-400 bg-gray-50 px-1.5 py-0.5 rounded">OFFLINE</span>
                )}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  )
}

// ── Main Monitoring Page ────────────────────────────────────────────────────

const Monitoring: React.FC = () => {
  const {
    data: accounts = [],
    isLoading: loadingAccounts,
    isError: accountsError,
    refetch: refetchAccounts,
  } = useDashboardAccounts()

  const [selectedHost, setSelectedHost] = useState<string | null>(null)
  const [windowValue, setWindowValue] = useState('24h')
  const [filter, setFilter] = useState<Set<string>>(new Set())
  const [paused, setPaused] = useState(false)

  const accountList = accounts as AccountWithHealth[]

  // Initialise selectedHost to the first connected account (or first account)
  // once the account list arrives.
  const initialisedRef = useRef(false)
  useEffect(() => {
    if (initialisedRef.current || accountList.length === 0) return
    const connected = accountList.find(a => a.is_connected)
    setSelectedHost(connected?.host_id ?? accountList[0]?.host_id ?? null)
    initialisedRef.current = true
  }, [accountList])

  const { data: status } = useMonitoringStatus(selectedHost)
  const { data: statusEvents = [], isLoading: loadingStatusEvents } = useMonitoringStatusEvents(selectedHost)
  const { data: metrics, isLoading: loadingMetrics } = useMonitoringMetrics(selectedHost, windowValue)
  const { data: errors = [], isLoading: loadingErrors } = useMonitoringEvents(selectedHost)
  const { data: queueDepth = [], isLoading: loadingQueue } = useMonitoringQueueDepth(selectedHost)
  const { events: tailEvents, connected: tailConnected, error: tailError, clear: clearTail } = useMonitoringStream(selectedHost)

  const NON_ERROR_TAIL_TYPES = ['MESSAGE_IN', 'MESSAGE_OUT', 'RECEIPT', 'STATUS', 'QUEUE_DEPTH'] as const

  const visibleTailEvents = useMemo(() => {
    if (filter.size === 0) return tailEvents
    return tailEvents.filter(ev => {
      if (filter.has('ERROR') && !(NON_ERROR_TAIL_TYPES as readonly string[]).includes(ev.event_type)) {
        return true
      }
      return filter.has(ev.event_type)
    })
  }, [tailEvents, filter])

  const toggleFilter = useCallback((eventType: string) => {
    setFilter(prev => {
      const next = new Set(prev)
      if (next.has(eventType)) next.delete(eventType)
      else next.add(eventType)
      return next
    })
  }, [])

  const chartMax = useMemo(() => {
    const all = (metrics?.buckets ?? []).flatMap(b => [b.inbound, b.outbound])
    return Math.max(1, ...all)
  }, [metrics])

  const queueMax = useMemo(() => {
    return Math.max(1, ...queueDepth.map(s => s.queue_size))
  }, [queueDepth])

  const isForbidden = (accountsError as unknown as { response?: { status?: number } } | null)?.response?.status === 403

  // Count online accounts for the overview
  const onlineCount = accountList.filter(a => a.is_connected).length
  const offlineCount = accountList.length - onlineCount

  if (loadingAccounts) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
      </div>
    )
  }

  if (accountsError) {
    return isForbidden ? (
      <Card className="p-12 text-center max-w-lg mx-auto">
        <ShieldAlert className="h-12 w-12 text-amber-500 mx-auto mb-3" />
        <h3 className="text-lg font-semibold text-gray-900">You don't have permission to view monitoring</h3>
        <p className="text-sm text-gray-500 mt-1 max-w-sm mx-auto">
          Contact an administrator to request access.
        </p>
      </Card>
    ) : (
      <div className="text-center py-12 text-red-500">
        <AlertCircle className="h-12 w-12 mx-auto mb-3" />
        <p>Failed to load accounts.</p>
        <Button variant="primary" size="sm" className="mt-4" onClick={() => refetchAccounts()}>
          Retry
        </Button>
      </div>
    )
  }

  if (accountList.length === 0) {
    return (
      <Card className="p-12 text-center max-w-lg mx-auto">
        <Database className="h-12 w-12 text-gray-300 mx-auto mb-3" />
        <h3 className="text-lg font-semibold text-gray-900">No WhatsApp accounts linked</h3>
        <p className="text-sm text-gray-500 mt-1">
          Pair a WhatsApp number to start monitoring its activity.
        </p>
      </Card>
    )
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-5">
        <div>
          <h1 className="text-xl font-semibold text-gray-900">Monitoring</h1>
          <p className="text-[13px] text-gray-600">Live tail, status history, and metrics per WhatsApp account</p>
        </div>
        <AccountDropdown
          accounts={accountList}
          selectedHost={selectedHost}
          onSelect={setSelectedHost}
        />
      </div>

      {/* Multi-channel overview strip */}
      <div className="mb-5">
        <div className="flex items-center gap-2 mb-2">
          <h2 className="text-sm font-semibold text-gray-900">Channels</h2>
          <span className="text-xs text-gray-500">{onlineCount} online · {offlineCount} offline</span>
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
          {accountList.map(a => (
            <button
              key={a.id}
              type="button"
              onClick={() => setSelectedHost(a.host_id)}
              className={`text-left rounded-lg border px-3 py-2 transition-colors ${
                a.host_id === selectedHost
                  ? 'border-primary-300 bg-primary-50/60 ring-1 ring-primary-200'
                  : 'border-gray-200 bg-white hover:bg-gray-50'
              }`}
            >
              <div className="flex items-center gap-2">
                <span className={`shrink-0 h-2 w-2 rounded-full ${a.is_connected ? 'bg-green-500' : 'bg-gray-400'}`} />
                <span className="truncate flex-1 text-sm font-medium text-gray-800">
                  {a.display_name || a.host_id}
                </span>
                {a.is_connected ? (
                  <span className="shrink-0 text-[10px] font-semibold text-green-700 bg-green-100 px-1.5 py-0.5 rounded">ONLINE</span>
                ) : (
                  <span className="shrink-0 text-[10px] font-semibold text-gray-500 bg-gray-100 px-1.5 py-0.5 rounded">OFFLINE</span>
                )}
              </div>
              {a.queue_size > 0 && (
                <p className="text-[11px] text-amber-600 mt-0.5 ml-4">Queue: {a.queue_size}</p>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Status overview cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-5">
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-500 mb-2">
            <Activity className="h-4 w-4" />
            Status
          </div>
          <span className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${statusBadgeCls(status?.status ?? 'OFFLINE')}`}>
            {status?.status ?? 'OFFLINE'}
          </span>
          {status?.is_connected ? (
            <p className="text-xs text-green-600 mt-2 flex items-center gap-1">
              <Wifi className="h-3.5 w-3.5" /> Connected
            </p>
          ) : (
            <p className="text-xs text-red-500 mt-2 flex items-center gap-1">
              <WifiOff className="h-3.5 w-3.5" /> Disconnected
            </p>
          )}
        </Card>

        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-500 mb-2">
            <Timer className="h-4 w-4" />
            Uptime
          </div>
          <p className="text-lg font-semibold text-gray-900">{status?.is_connected ? (status?.uptime || '0s') : '—'}</p>
        </Card>

        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-500 mb-2">
            <CheckCircle2 className="h-4 w-4" />
            Last Online
          </div>
          <p className="text-sm font-medium text-gray-900" title={formatDateTime(status?.last_connected_at)}>
            {formatRelative(status?.last_connected_at)}
          </p>
        </Card>

        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-500 mb-2">
            <Signal className="h-4 w-4" />
            Last Offline
          </div>
          <p className="text-sm font-medium text-gray-900" title={formatDateTime(status?.last_disconnected_at)}>
            {formatRelative(status?.last_disconnected_at)}
          </p>
        </Card>
      </div>

      {/* Status history + metrics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-5">
        <Card className="p-4">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">Status History</h3>
          {loadingStatusEvents ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-primary-500" />
            </div>
          ) : statusEvents.length === 0 ? (
            <p className="text-sm text-gray-500 py-6 text-center">No status transitions recorded yet.</p>
          ) : (
            <ul className="space-y-2 max-h-64 overflow-y-auto">
              {statusEvents.map((ev) => (
                <li key={ev.id} className="flex items-center justify-between gap-2 text-sm">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className={`shrink-0 px-2 py-0.5 rounded-full text-[11px] font-medium ${statusBadgeCls(ev.status)}`}>
                      {ev.status}
                    </span>
                    {ev.message && <span className="text-gray-500 truncate text-xs">{ev.message}</span>}
                  </div>
                  <span className="text-xs text-gray-400 shrink-0" title={formatDateTime(ev.occurred_at)}>
                    {formatRelative(ev.occurred_at)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Card>

        <Card className="p-4">
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-semibold text-gray-900">Message Volume</h3>
            <div className="flex gap-1">
              {WINDOWS.map((w) => (
                <button
                  key={w.value}
                  onClick={() => setWindowValue(w.value)}
                  className={`px-2 py-1 text-[11px] rounded-md font-medium ${
                    windowValue === w.value ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                  }`}
                >
                  {w.label}
                </button>
              ))}
            </div>
          </div>

          {loadingMetrics ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-primary-500" />
            </div>
          ) : (
            <>
              <div className="grid grid-cols-3 gap-3 mb-4">
                <div className="text-center">
                  <p className="text-xl font-semibold text-blue-600">{metrics?.inbound ?? 0}</p>
                  <p className="text-[11px] text-gray-500">Inbound</p>
                </div>
                <div className="text-center">
                  <p className="text-xl font-semibold text-emerald-600">{metrics?.outbound ?? 0}</p>
                  <p className="text-[11px] text-gray-500">Outbound</p>
                </div>
                <div className="text-center">
                  <p className="text-xl font-semibold text-red-600">{metrics?.failed ?? 0}</p>
                  <p className="text-[11px] text-gray-500">Failed</p>
                </div>
              </div>

              {(metrics?.buckets?.length ?? 0) > 0 ? (
                <div className="flex items-end gap-1 h-20 mb-4">
                  {metrics!.buckets.map((b, i) => (
                    <div key={i} className="flex-1 flex items-end gap-px h-full" title={`${new Date(b.start).toLocaleString()}`}>
                      <div className="flex-1 bg-blue-400/70" style={{ height: `${Math.max(2, (b.inbound / chartMax) * 100)}%` }} />
                      <div className="flex-1 bg-emerald-400/70" style={{ height: `${Math.max(2, (b.outbound / chartMax) * 100)}%` }} />
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-sm text-gray-500 py-6 text-center">No messages in this window.</p>
              )}

              {metrics && Object.keys(metrics.status_breakdown).length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {Object.entries(metrics.status_breakdown).map(([k, v]) => (
                    <span key={k} className="px-2 py-1 rounded-full text-[11px] font-medium bg-gray-100 text-gray-700">
                      {k}: {v}
                    </span>
                  ))}
                </div>
              )}
            </>
          )}
        </Card>
      </div>

      {/* Queue depth + errors */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-5">
        <Card className="p-4">
          <h3 className="text-sm font-semibold text-gray-900 mb-1">Queue Depth</h3>
          <p className="text-xs text-gray-500 mb-3">Current pending: <b className="text-gray-800">{status?.queue_size ?? 0}</b></p>
          {loadingQueue ? (
            <div className="flex justify-center py-6">
              <Loader2 className="h-5 w-5 animate-spin text-primary-500" />
            </div>
          ) : queueDepth.length === 0 ? (
            <p className="text-sm text-gray-500 py-4 text-center">No queue samples yet.</p>
          ) : (
            <div className="flex items-end gap-0.5 h-16">
              {queueDepth.map((s) => (
                <div
                  key={s.id}
                  className="flex-1 bg-gray-400/60"
                  style={{ height: `${Math.max(2, (s.queue_size / queueMax) * 100)}%` }}
                  title={`${formatDateTime(s.timestamp)} · q=${s.queue_size}`}
                />
              ))}
            </div>
          )}
        </Card>

        <Card className="p-4">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">Errors & Warnings</h3>
          {loadingErrors ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-primary-500" />
            </div>
          ) : errors.length === 0 ? (
            <p className="text-sm text-gray-500 py-6 text-center">No errors or warnings recorded.</p>
          ) : (
            <ul className="space-y-2 max-h-64 overflow-y-auto">
              {errors.map((ev) => (
                <li key={ev.id} className="flex items-start justify-between gap-2 text-sm">
                  <span className="truncate text-xs text-red-600">{eventLabel(ev.event_type, ev.payload)}</span>
                  <span className="text-xs text-gray-400 shrink-0" title={formatDateTime(ev.occurred_at)}>
                    {formatRelative(ev.occurred_at)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>

      {/* Live event tail */}
      <Card className="p-4">
        <div className="flex flex-wrap items-center justify-between gap-2 mb-3">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-gray-900">Live Event Tail</h3>
            <span
              className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium ${
                tailConnected ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
              }`}
            >
              {tailConnected ? (
                <>
                  <Wifi className="h-3 w-3" /> Live
                </>
              ) : (
                <>
                  <WifiOff className="h-3 w-3" /> Reconnecting
                </>
              )}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="secondary" size="sm" onClick={() => setPaused(p => !p)}>
              {paused ? <Play className="h-3.5 w-3.5" /> : <Pause className="h-3.5 w-3.5" />}
              {paused ? 'Resume' : 'Pause'}
            </Button>
            <Button variant="ghost" size="sm" onClick={clearTail}>
              <Trash2 className="h-3.5 w-3.5" />
              Clear
            </Button>
          </div>
        </div>

        {tailError && (
          <div className="mb-3 text-xs text-amber-600 bg-amber-50 border border-amber-200 rounded-md px-3 py-2 flex items-center gap-2">
            <RefreshCw className="h-3.5 w-3.5 animate-spin" />
            {tailError}
          </div>
        )}

        <div className="flex flex-wrap gap-2 mb-3">
          {(['MESSAGE_IN', 'MESSAGE_OUT', 'RECEIPT', 'STATUS', 'QUEUE_DEPTH', 'ERROR'] as const).map((t) => (
            <button
              key={t}
              onClick={() => toggleFilter(t)}
              className={`px-2 py-1 text-[11px] rounded-md font-medium ${
                filter.has(t) ? 'bg-gray-700 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {t === 'ERROR' ? 'Errors' : t}
            </button>
          ))}
        </div>

        <div className="bg-gray-950 rounded-lg p-3 h-80 overflow-y-auto font-mono text-[12px] leading-6">
          {visibleTailEvents.length === 0 ? (
            <p className="text-gray-500 text-center py-10">Waiting for events…</p>
          ) : (
            visibleTailEvents.map((ev) => (
              <div key={`${ev.id}-${ev.occurred_at}`} className="flex gap-2 items-baseline">
                <span className="text-gray-500 shrink-0">
                  {new Date(ev.occurred_at).toLocaleTimeString(navigator.language || 'en-US', { hour12: false })}
                </span>
                <span className={`font-semibold shrink-0 ${eventTypeColor(ev.event_type)}`}>{ev.event_type}</span>
                <span className="text-gray-300 truncate">{eventLabel(ev.event_type, ev.payload)}</span>
              </div>
            ))
          )}
        </div>
      </Card>
    </div>
  )
}

export default Monitoring