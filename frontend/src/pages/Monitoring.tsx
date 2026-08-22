import React, { useEffect, useMemo, useRef, useState } from 'react'
import Card from '@/components/ui/card'
import {
  useDashboardAccounts,
} from '@/hooks/useDashboard'
import {
  useMonitoringStatus,
  useMonitoringMetrics,
  useMonitoringQueueDepth,
} from '@/hooks/useMonitoring'
import {
  Activity,
  CheckCircle2,
  ChevronDown,
  Loader2,
  Signal,
  Smartphone,
  Timer,
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

const WINDOWS = [
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
]

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

// ── Main Monitoring Page ────────────────────────────────────────────────────────

const Monitoring: React.FC = () => {
  const {
    data: accounts = [],
    isLoading: loadingAccounts,
    isError: accountsError,
    refetch: refetchAccounts,
  } = useDashboardAccounts()

  const [selectedHost, setSelectedHost] = useState<string | null>(null)
  const [windowValue, setWindowValue] = useState('24h')

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
  const { data: metrics, isLoading: loadingMetrics } = useMonitoringMetrics(selectedHost, windowValue)
  const { data: queueDepth = [], isLoading: loadingQueue } = useMonitoringQueueDepth(selectedHost)

  const chartMax = useMemo(() => {
    const all = (metrics?.buckets ?? []).flatMap(b => [b.inbound, b.outbound])
    return Math.max(1, ...all)
  }, [metrics])

  const queueMax = useMemo(() => {
    return Math.max(1, ...queueDepth.map(s => s.queue_size))
  }, [queueDepth])

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
    return (
      <div className="text-center py-12 text-red-500">
        <p>Failed to load accounts.</p>
        <button
          type="button"
          onClick={() => refetchAccounts()}
          className="mt-4 px-3 py-1.5 text-sm bg-primary-600 text-white rounded-md hover:bg-primary-700"
        >
          Retry
        </button>
      </div>
    )
  }

  if (accountList.length === 0) {
    return (
      <Card className="p-12 text-center max-w-lg mx-auto">
        <Smartphone className="h-12 w-12 text-gray-300 mx-auto mb-3" />
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
          <p className="text-[13px] text-gray-600">Live metrics per WhatsApp account</p>
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

      {/* Status summary cards */}
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

      {/* Metrics graphs */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
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
      </div>
    </div>
  )
}

export default Monitoring
