import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import apiClient, { mediaApi } from '@/lib/apiClient'
import type { Conversation, ConversationSummary, ConversationMessage, Activity, Contact, Paginated, ContactFieldDefinition, DealStage, DealStageTransition, Pipeline } from '@/types'

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
      const { data } = await apiClient.get<ConversationSummary[]>(`/api/v1/conversations?${params}`)
      return data ?? []
    },
    refetchInterval: 30_000,
  })
}

export function useConversation(id: string) {
  const qc = useQueryClient()
  return useQuery({
    queryKey: inboxKeys.detail(id),
    queryFn: async () => {
      const { data } = await apiClient.get<{ conversation: ConversationSummary; messages: ConversationMessage[] }>(
        `/api/v1/conversations/${id}`
      )
      if (!data) return data
      // Preserve any pending optimistic messages that are not yet returned in the server response
      const currentCache = qc.getQueryData<{ conversation: ConversationSummary; messages: ConversationMessage[] }>(
        inboxKeys.detail(id)
      )
      if (currentCache?.messages && Array.isArray(data.messages)) {
        const pending = currentCache.messages.filter(
          (m) =>
            m.id.startsWith('temp-') &&
            !data.messages.some(
              (sm) =>
                sm.direction === m.direction &&
                sm.content === m.content &&
                Math.abs(new Date(sm.provider_timestamp || sm.created_at).getTime() - new Date(m.provider_timestamp || m.created_at).getTime()) < 60000
            )
        )
        if (pending.length > 0) {
          return {
            ...data,
            messages: [...data.messages, ...pending],
          }
        }
      }
      return data
    },
    enabled: !!id,
    refetchInterval: 3000,
  })
}

export function useActivities() {
  return useQuery({
    queryKey: inboxKeys.activities,
    queryFn: async () => {
      const { data } = await apiClient.get<Activity[]>('/api/v1/activities')
      return data ?? []
    },
    refetchInterval: 30_000,
  })
}

export function useContacts(filters: { limit?: number; offset?: number; q?: string } = {}) {
  return useQuery({
    queryKey: inboxKeys.contacts(filters),
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filters.limit) params.set('limit', String(filters.limit))
      if (filters.offset) params.set('offset', String(filters.offset))
      if (filters.q) params.set('q', filters.q)
      const { data } = await apiClient.get<Paginated<Contact>>(`/api/v1/contacts?${params}`)
      return data ?? { items: [], total: 0, limit: 20, offset: 0 }
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
      return data ?? []
    },
    enabled: !!contactId,
  })
}

export function useUpdateContact(contactId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      name?: string
      email?: string
      tags?: string[]
      custom_values?: Record<string, unknown>
      deal_stage_key?: string | null
      deal_stage_id?: string | null
    }) => {
      const { data } = await apiClient.patch<Contact>(`/api/v1/contacts/${contactId}`, input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['contacts', contactId] })
      qc.invalidateQueries({ queryKey: [inboxKeys.contacts()] })
    },
  })
}

export function useContactConversations(contactId: string) {
  return useQuery({
    queryKey: ['contacts', contactId, 'conversations'],
    queryFn: async () => {
      const { data } = await apiClient.get<Conversation[]>(`/api/v1/contacts/${contactId}/conversations`)
      return data ?? []
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

export function useSendMessage(conversationId?: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      account,
      recipient,
      message,
      type = 'TEXT',
      media_key,
      is_group,
    }: {
      account: string
      recipient: string
      message: string
      type?: string
      media_key?: string
      is_group?: boolean
    }) => {
      const { data } = await apiClient.post(`/api/v1/accounts/${account}/messages`, {
        recipient,
        message,
        type,
        media_key,
        is_group,
      })
      return data
    },
    onMutate: async (vars) => {
      const targetId = conversationId
      const tempId = 'temp-' + Date.now() + '-' + Math.random().toString(36).substring(2, 7)
      await qc.cancelQueries({ queryKey: inboxKeys.all })
      if (targetId) {
        await qc.cancelQueries({ queryKey: inboxKeys.detail(targetId) })
      }

      const prevData = targetId
        ? qc.getQueryData<{ conversation: ConversationSummary; messages: ConversationMessage[] }>(
            inboxKeys.detail(targetId)
          )
        : undefined

      const optimisticMsg: ConversationMessage = {
        id: tempId,
        tenant_id: prevData?.conversation?.tenant_id || '',
        conversation_id: targetId || '',
        actor: 'OPERATOR',
        provider: 'whatsmeow',
        provider_message_id: '',
        direction: 'OUTGOING',
        content: vars.message,
        message_type: (vars.type as any) || 'TEXT',
        media_url: '',
        status: 'PENDING',
        provider_timestamp: new Date().toISOString(),
        is_internal: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      }

      if (targetId) {
        qc.setQueryData<{ conversation: ConversationSummary; messages: ConversationMessage[] }>(
          inboxKeys.detail(targetId),
          (old) => {
            if (!old) {
              return {
                conversation: {
                  id: targetId,
                  tenant_id: '',
                  account_id: vars.account,
                  contact_id: '',
                  ticket_number: 0,
                  status: 'OPEN',
                  bot_state: '',
                  started_at: new Date().toISOString(),
                  last_activity_at: new Date().toISOString(),
                  created_at: new Date().toISOString(),
                  updated_at: new Date().toISOString(),
                  contact_number: vars.recipient,
                } as ConversationSummary,
                messages: [optimisticMsg],
              }
            }
            return {
              ...old,
              messages: [...(old.messages || []), optimisticMsg],
            }
          }
        )
      }

      // Snapshot all conversation list queries for rollback
      const prevSummariesQueries = qc.getQueriesData<ConversationSummary[]>({ queryKey: inboxKeys.all, exact: false })

      // Optimistically update all conversation list queries
      qc.setQueriesData<ConversationSummary[]>(
        { queryKey: inboxKeys.all, exact: false },
        (oldSummaries) => {
          if (!Array.isArray(oldSummaries)) return oldSummaries
          const updated = oldSummaries.map((s) => {
            if (s.id === targetId || (s.account_id === vars.account && s.contact_number === vars.recipient)) {
              return {
                ...s,
                last_message_preview: vars.message,
                last_message_actor: 'OPERATOR' as any,
                last_activity_at: new Date().toISOString(),
              }
            }
            return s
          })
          return [...updated].sort(
            (a, b) => new Date(b.last_activity_at).getTime() - new Date(a.last_activity_at).getTime()
          )
        }
      )

      return { prevData, prevSummariesQueries, tempId }
    },
    onError: (_err, _vars, context) => {
      if (conversationId && context?.prevData) {
        qc.setQueryData(inboxKeys.detail(conversationId), context.prevData)
      }
      if (context?.prevSummariesQueries) {
        for (const [key, val] of context.prevSummariesQueries) {
          qc.setQueryData(key, val)
        }
      }
    },
    onSettled: () => {
      if (conversationId) {
        qc.invalidateQueries({ queryKey: inboxKeys.detail(conversationId) })
      }
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

export function useContactFieldDefinitions() {
  return useQuery({
    queryKey: ['contact-field-definitions'],
    queryFn: async () => {
      const { data } = await apiClient.get<ContactFieldDefinition[]>('/api/v1/contact-field-definitions')
      return data ?? []
    },
  })
}

export function useCreateContactFieldDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      key: string
      label: string
      field_type: ContactFieldDefinition['field_type']
      options?: string[]
      is_required?: boolean
      sort_order?: number
    }) => {
      const { data } = await apiClient.post<ContactFieldDefinition>('/api/v1/contact-field-definitions', input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['contact-field-definitions'] })
    },
  })
}

export function useUpdateContactFieldDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      id,
      ...input
    }: {
      id: string
      label: string
      field_type: ContactFieldDefinition['field_type']
      options?: string[]
      is_required?: boolean
      sort_order?: number
      is_active?: boolean
    }) => {
      const { data } = await apiClient.patch<ContactFieldDefinition>(`/api/v1/contact-field-definitions/${id}`, input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['contact-field-definitions'] })
    },
  })
}

export function useDeleteContactFieldDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/api/v1/contact-field-definitions/${id}`)
      return id
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['contact-field-definitions'] })
    },
  })
}

// ── Deal Stage / CRM Pipeline hooks ──────────────────────────────────

export function usePipelines(isActive?: boolean) {
  return useQuery({
    queryKey: ['pipelines', isActive],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (isActive !== undefined) params.set('is_active', String(isActive))
      const { data } = await apiClient.get<Pipeline[]>(`/api/v1/pipelines?${params}`)
      return data ?? []
    },
    staleTime: 60 * 60_000,
  })
}

export function useCreatePipeline() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: { name: string; description?: string; is_default?: boolean }) => {
      const { data } = await apiClient.post<Pipeline>('/api/v1/pipelines', input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

export function useUpdatePipeline() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      id,
      ...input
    }: {
      id: string
      name?: string
      description?: string
      is_default?: boolean
      is_active?: boolean
    }) => {
      const { data } = await apiClient.patch<Pipeline>(`/api/v1/pipelines/${id}`, input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

export function useDeletePipeline() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/api/v1/pipelines/${id}`)
      return id
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['pipelines'] })
    },
  })
}

export function useDealStages(pipelineId?: string) {
  return useQuery({
    queryKey: ['deal-stages', pipelineId],
    queryFn: async () => {
      const url = pipelineId ? `/api/v1/deal-stages?pipeline_id=${pipelineId}` : '/api/v1/deal-stages'
      const { data } = await apiClient.get<DealStage[]>(url)
      return data ?? []
    },
    staleTime: 5 * 60_000,
  })
}

export function useDealStageHistory(contactId: string) {
  return useQuery({
    queryKey: ['contacts', contactId, 'deal-history'],
    queryFn: async () => {
      const { data } = await apiClient.get<DealStageTransition[]>(`/api/v1/contacts/${contactId}/deal-history`)
      return data ?? []
    },
    enabled: !!contactId,
  })
}

export function useMoveContactToStage(contactId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      stageKey,
      stageId,
      dealStageId,
      note,
    }: {
      stageKey?: string
      stageId?: string
      dealStageId?: string
      note?: string
    }) => {
      const { data } = await apiClient.post<DealStageTransition>(`/api/v1/contacts/${contactId}/move-stage`, {
        stage_key: stageKey,
        stage_id: stageId || dealStageId,
        deal_stage_id: dealStageId || stageId,
        note: note || '',
      })
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['contacts', contactId] })
      qc.invalidateQueries({ queryKey: ['contacts', contactId, 'deal-history'] })
      qc.invalidateQueries({ queryKey: [inboxKeys.contacts()] })
    },
  })
}

export function useCreateDealStage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (input: {
      pipeline_id?: string
      key: string
      label: string
      color: string
      icon: string
      sort_order: number
      is_won?: boolean
      is_lost?: boolean
    }) => {
      const { data } = await apiClient.post<DealStage>('/api/v1/deal-stages', input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['deal-stages'] })
    },
  })
}

export function useUpdateDealStage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async ({
      id,
      ...input
    }: {
      id: string
      pipeline_id?: string
      label?: string
      color?: string
      icon?: string
      sort_order?: number
      is_active?: boolean
      is_won?: boolean
      is_lost?: boolean
    }) => {
      const { data } = await apiClient.patch<DealStage>(`/api/v1/deal-stages/${id}`, input)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['deal-stages'] })
    },
  })
}

export function useDeleteDealStage() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (id: string) => {
      await apiClient.delete(`/api/v1/deal-stages/${id}`)
      return id
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['deal-stages'] })
    },
  })
}
