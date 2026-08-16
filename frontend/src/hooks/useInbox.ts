import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import apiClient, { mediaApi } from '@/lib/apiClient'
import type { Conversation, ConversationMessage, Activity, Contact } from '@/types'

export interface ConversationFilters {
  status?: string
  assignee?: string
  account?: string
  priority?: string
  limit?: number
  offset?: number
}

export const inboxKeys = {
  all: ['conversations'] as const,
  list: (filters: ConversationFilters) => ['conversations', filters] as const,
  detail: (id: string) => ['conversations', id] as const,
  messages: (id: string) => ['conversations', id, 'messages'] as const,
  activities: ['activities'] as const,
  contacts: (filters?: Record<string, unknown>) => ['contacts', filters] as const,
}

export function useInbox(filters: ConversationFilters = {}) {
  return useQuery({
    queryKey: inboxKeys.list(filters),
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filters.status) params.set('status', filters.status)
      if (filters.assignee) params.set('assignee', filters.assignee)
      if (filters.limit) params.set('limit', String(filters.limit))
      if (filters.offset) params.set('offset', String(filters.offset))
      const { data } = await apiClient.get<Conversation[]>(`/api/v1/conversations?${params}`)
      return data
    },
    refetchInterval: 30_000,
  })
}

export function useConversation(id: string) {
  return useQuery({
    queryKey: inboxKeys.detail(id),
    queryFn: async () => {
      const { data } = await apiClient.get<{ conversation: Conversation; messages: ConversationMessage[] }>(
        `/api/v1/conversations/${id}`
      )
      return data
    },
    enabled: !!id,
  })
}

export function useActivities() {
  return useQuery({
    queryKey: inboxKeys.activities,
    queryFn: async () => {
      const { data } = await apiClient.get<Activity[]>('/api/v1/activities')
      return data
    },
    refetchInterval: 30_000,
  })
}

export function useContacts() {
  return useQuery({
    queryKey: inboxKeys.contacts(),
    queryFn: async () => {
      const { data } = await apiClient.get<Contact[]>('/api/v1/contacts')
      return data
    },
  })
}

export function useContact(id: string) {
  return useQuery({
    queryKey: ['contacts', id],
    queryFn: async () => {
      const { data } = await apiClient.get<Contact>(`/api/v1/contacts/${id}`)
      return data
    },
    enabled: !!id,
  })
}

export function useContactActivities(contactId: string) {
  return useQuery({
    queryKey: ['contacts', contactId, 'activities'],
    queryFn: async () => {
      const { data } = await apiClient.get<Activity[]>(`/api/v1/contacts/${contactId}/activities`)
      return data
    },
    enabled: !!contactId,
  })
}

export function useCreateContactActivity(contactId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      type?: string
      summary: string
      next_action?: string
      priority?: string
      due_at?: string | null
    }) => {
      const { data } = await apiClient.post<Activity>(`/api/v1/contacts/${contactId}/activities`, {
        type: input.type || 'FOLLOW_UP',
        summary: input.summary,
        next_action: input.next_action,
        priority: input.priority,
        due_at: input.due_at || undefined,
      })
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['contacts', contactId, 'activities'] })
    },
  })
}

export function useSendMessage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      account,
      recipient,
      message,
      type = 'TEXT',
      media_key,
    }: {
      account: string
      recipient: string
      message: string
      type?: string
      media_key?: string
    }) => {
      const { data } = await apiClient.post(`/api/v1/accounts/${account}/messages`, {
        recipient,
        message,
        type,
        media_key,
      })
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: inboxKeys.all })
    },
  })
}

export function useUploadMedia() {
  return useMutation({
    mutationFn: async ({ file, onProgress }: { file: File; onProgress?: (pct: number) => void }) => {
      return await mediaApi.uploadMedia(file, onProgress)
    },
  })
}

export function useAssignConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, assignee }: { id: string; assignee: string }) => {
      const { data } = await apiClient.post(`/api/v1/operator/assign?id=${id}`, { assignee })
      return data
    },
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: inboxKeys.detail(id) })
      qc.invalidateQueries({ queryKey: inboxKeys.all })
    },
  })
}

export function useHandoffConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason?: string }) => {
      const { data } = await apiClient.post(`/api/v1/operator/handoff?id=${id}`, { reason })
      return data
    },
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: inboxKeys.detail(id) })
      qc.invalidateQueries({ queryKey: inboxKeys.all })
    },
  })
}

export function useCloseConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason?: string }) => {
      const { data } = await apiClient.post(`/api/v1/operator/close?id=${id}`, { reason })
      return data
    },
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: inboxKeys.detail(id) })
      qc.invalidateQueries({ queryKey: inboxKeys.all })
    },
  })
}

export function useReopenConversation() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id }: { id: string }) => {
      const { data } = await apiClient.post(`/api/v1/operator/reopen?id=${id}`)
      return data
    },
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: inboxKeys.detail(id) })
      qc.invalidateQueries({ queryKey: inboxKeys.all })
    },
  })
}

export function useAddInternalNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({ id, content }: { id: string; content: string }) => {
      const { data } = await apiClient.post(`/api/v1/notes?id=${id}`, {
        content,
        actor: 'OPERATOR',
      })
      return data
    },
    onSuccess: (_data, { id }) => {
      qc.invalidateQueries({ queryKey: inboxKeys.messages(id) })
    },
  })
}

export function useAcknowledgeActivity() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await apiClient.post(`/api/v1/activities/${id}/acknowledge`)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: inboxKeys.activities })
    },
  })
}
