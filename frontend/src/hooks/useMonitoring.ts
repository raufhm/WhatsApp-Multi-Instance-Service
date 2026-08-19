import { useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import apiClient from '@/lib/apiClient'
import type {
  InstanceLogEvent,
  InstanceMonitoring,
  MessageMetrics,
  QueueDepthSample,
  StatusEvent,
} from '@/types'

export const monitoringKeys = {
  status: (host: string) => ['monitoring', 'status', host] as const,
  statusEvents: (host: string) => ['monitoring', 'status-events', host] as const,
  metrics: (host: string, window: string) => ['monitoring', 'metrics', host, window] as const,
  events: (host: string) => ['monitoring', 'events', host] as const,
  queueDepth: (host: string, minutes: number) => ['monitoring', 'queue-depth', host, minutes] as const,
}

export function useMonitoringStatus(hostId: string | null) {
  return useQuery({
    queryKey: monitoringKeys.status(hostId ?? ''),
    queryFn: async () => {
      const { data } = await apiClient.get<InstanceMonitoring>('/dashboard/api/monitoring/status', {
        params: { host: hostId },
      })
      return data
    },
    enabled: !!hostId,
    refetchInterval: 10_000,
  })
}

export function useMonitoringStatusEvents(hostId: string | null, limit = 50) {
  return useQuery({
    queryKey: monitoringKeys.statusEvents(hostId ?? ''),
    queryFn: async () => {
      const { data } = await apiClient.get<StatusEvent[]>('/dashboard/api/monitoring/status-events', {
        params: { host: hostId, limit },
      })
      return data ?? []
    },
    enabled: !!hostId,
    refetchInterval: 30_000,
  })
}

export function useMonitoringMetrics(hostId: string | null, window: string) {
  return useQuery({
    queryKey: monitoringKeys.metrics(hostId ?? '', window),
    queryFn: async () => {
      const { data } = await apiClient.get<MessageMetrics>('/dashboard/api/monitoring/metrics', {
        params: { host: hostId, window },
      })
      return data
    },
    enabled: !!hostId,
    refetchInterval: 60_000,
  })
}

export function useMonitoringEvents(hostId: string | null, limit = 50) {
  return useQuery({
    queryKey: monitoringKeys.events(hostId ?? ''),
    queryFn: async () => {
      const { data } = await apiClient.get<InstanceLogEvent[]>('/dashboard/api/monitoring/events', {
        params: { host: hostId, type: 'ERROR', limit },
      })
      return data ?? []
    },
    enabled: !!hostId,
    refetchInterval: 30_000,
  })
}

export function useMonitoringQueueDepth(hostId: string | null, minutes = 60) {
  return useQuery({
    queryKey: monitoringKeys.queueDepth(hostId ?? '', minutes),
    queryFn: async () => {
      const { data } = await apiClient.get<QueueDepthSample[]>('/dashboard/api/monitoring/queue-depth', {
        params: { host: hostId, minutes },
      })
      return data ?? []
    },
    enabled: !!hostId,
    refetchInterval: 15_000,
  })
}

const MAX_TAIL_EVENTS = 500

/**
 * useMonitoringStream opens an SSE connection to the live per-account event
 * tail. It keeps at most MAX_TAIL_EVENTS events in memory and automatically
 * reconnects (with a since cursor backfill) via the browser's EventSource.
 */
export function useMonitoringStream(hostId: string | null) {
  const [events, setEvents] = useState<InstanceLogEvent[]>([])
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const lastEventId = useRef<number>(0)
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    if (!hostId) {
      setEvents([])
      setConnected(false)
      return
    }

    let cancelled = false
    setError(null)

    const connect = () => {
      if (cancelled) return
      const params = new URLSearchParams({ host: hostId })
      if (lastEventId.current > 0) {
        params.set('since', String(lastEventId.current))
      }
      const es = new EventSource(`/dashboard/api/monitoring/stream?${params.toString()}`)
      esRef.current = es

      es.onopen = () => {
        if (!cancelled) setConnected(true)
      }

      es.onmessage = (msg) => {
        if (cancelled) return
        try {
          const ev = JSON.parse(msg.data as string) as InstanceLogEvent
          lastEventId.current = ev.id
          setEvents((prev) => {
            const next = [...prev, ev]
            return next.length > MAX_TAIL_EVENTS ? next.slice(next.length - MAX_TAIL_EVENTS) : next
          })
        } catch {
          // ignore malformed frames
        }
      }

      es.onerror = () => {
        if (cancelled) return
        setConnected(false)
        setError('Connection lost — retrying...')
        // EventSource reconnects automatically; the server sends backfilled
        // events after the since cursor on each reconnect.
      }
    }

    connect()

    return () => {
      cancelled = true
      if (esRef.current) {
        esRef.current.close()
        esRef.current = null
      }
    }
  }, [hostId])

  const clear = () => setEvents([])

  return { events, connected, error, clear }
}
