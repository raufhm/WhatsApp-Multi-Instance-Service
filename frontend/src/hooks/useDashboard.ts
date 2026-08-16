import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/apiClient'
import type { BotRuleSet, UploadJob } from '@/types'

export const dashboardKeys = {
  accounts: ['dashboard', 'accounts'] as const,
  botRules: ['dashboard', 'bot-rules'] as const,
  uploadJobs: (status?: string) => ['dashboard', 'upload-jobs', status] as const,
}

export function useDashboardAccounts() {
  return useQuery({
    queryKey: dashboardKeys.accounts,
    queryFn: async () => {
      const { data } = await apiClient.get('/dashboard/api/accounts')
      return data
    },
    refetchInterval: 10_000,
  })
}

export function useDashboardBotRules() {
  return useQuery({
    queryKey: dashboardKeys.botRules,
    queryFn: async () => {
      const { data } = await apiClient.get<BotRuleSet[]>('/dashboard/api/bot-rules')
      return data
    },
    refetchInterval: 30_000,
  })
}

export function useCreateBotRuleSet() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (rules: any[]) => {
      const { data } = await apiClient.post('/dashboard/api/bot-rules', { rules })
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.botRules })
    },
  })
}

export function useActivateBotRuleSet() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: async (version: number) => {
      const { data } = await apiClient.post(`/dashboard/api/bot-rules/activate?version=${version}`)
      return data
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: dashboardKeys.botRules })
    },
  })
}

export function useUploadJobs(status?: string) {
  return useQuery({
    queryKey: dashboardKeys.uploadJobs(status),
    queryFn: async () => {
      const params = status ? `?status=${status}` : ''
      const { data } = await apiClient.get<UploadJob[]>(`/dashboard/api/upload-jobs${params}`)
      return data
    },
    refetchInterval: 15_000,
  })
}
