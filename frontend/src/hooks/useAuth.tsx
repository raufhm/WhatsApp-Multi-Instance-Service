import React, { useCallback, useContext, useEffect, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import apiClient, { TENANT_ID_KEY, TENANT_SLUG_KEY, REMEMBER_ME_KEY } from '@/lib/apiClient'
import type { Operator } from '@/types'

interface AuthContextType {
  isAuthenticated: boolean
  isLoading: boolean
  user: Operator | null
  tenantId: string | null
  tenantSlug: string | null
  tenantName: string | null
  setupCompleted: boolean
  login: (tenantOrSlug: string, identifier: string, codeOrPassword: string, rememberMe?: boolean) => Promise<any>
  loginWithBackupCode: (tenantOrSlug: string, identifier: string, backupCode: string, rememberMe?: boolean) => Promise<any>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  setSessionUser: (user: Operator, tenantId?: string, tenantSlug?: string) => void
  checkSetupStatus: () => Promise<boolean>
}

const AuthContext = React.createContext<AuthContextType | undefined>(undefined)

const authKeys = {
  me: ['auth', 'me'] as const,
}

function setStoredTenant(tenantId: string | null, tenantSlug: string | null) {
  if (tenantId) {
    localStorage.setItem(TENANT_ID_KEY, tenantId)
  } else {
    localStorage.removeItem(TENANT_ID_KEY)
  }
  if (tenantSlug) {
    localStorage.setItem(TENANT_SLUG_KEY, tenantSlug)
  } else {
    localStorage.removeItem(TENANT_SLUG_KEY)
  }
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const queryClient = useQueryClient()
  const [tenantId, setTenantId] = useState<string | null>(() => localStorage.getItem(TENANT_ID_KEY))
  const [tenantSlug, setTenantSlug] = useState<string | null>(() => localStorage.getItem(TENANT_SLUG_KEY))
  const [tenantName, setTenantName] = useState<string | null>(null)
  const [setupCompleted, setSetupCompleted] = useState<boolean>(true)

  const {
    data: meData,
    isPending,
    refetch,
  } = useQuery({
    queryKey: authKeys.me,
    queryFn: async () => {
      const response = await apiClient.get('/dashboard/api/me')
      return response.data as { user?: Operator; tenant_id?: string; tenant_slug?: string; tenant_name?: string }
    },
    enabled: false,
    retry: false,
    staleTime: 1000 * 60 * 5,
  })

  // Trigger the initial /me check once on mount.
  useEffect(() => {
    refetch()
  }, [refetch])

  const user = meData?.user ?? null
  const isAuthenticated = !!user
  const isLoading = isPending

  // Sync tenant fields returned by /me into local state and localStorage.
  useEffect(() => {
    if (meData?.tenant_id) {
      setTenantId(meData.tenant_id)
      localStorage.setItem(TENANT_ID_KEY, meData.tenant_id)
    }
    if (meData?.tenant_slug) {
      setTenantSlug(meData.tenant_slug)
      localStorage.setItem(TENANT_SLUG_KEY, meData.tenant_slug)
    }
    if (meData?.tenant_name) {
      setTenantName(meData.tenant_name)
    }
  }, [meData])

  const checkSetupStatus = useCallback(async (): Promise<boolean> => {
    try {
      const res = await apiClient.get('/dashboard/api/tenant/setup-status')
      const isCompleted = !!res.data?.completed
      setSetupCompleted(isCompleted)
      return isCompleted
    } catch {
      setSetupCompleted(true)
      return true
    }
  }, [])

  const setSessionUser = useCallback(
    (newUser: Operator, newTenantId?: string, newTenantSlug?: string) => {
      const tid = newTenantId || newUser.tenant_id
      setStoredTenant(tid || null, newTenantSlug || null)
      if (tid) setTenantId(tid)
      if (newTenantSlug) setTenantSlug(newTenantSlug)
      queryClient.setQueryData(authKeys.me, { user: newUser, tenant_id: tid, tenant_slug: newTenantSlug })
    },
    [queryClient]
  )

  const handleLoginResponse = useCallback(
    (response: any) => {
      const data = response.data || response
      const actualTenantId = data.tenant_id || data.user?.tenant_id || tenantId
      const actualTenantSlug = data.tenant_slug || tenantSlug
      const actualTenantName = data.tenant_name || null

      setStoredTenant(actualTenantId || null, actualTenantSlug || null)
      setTenantId(actualTenantId || null)
      setTenantSlug(actualTenantSlug || null)
      if (actualTenantName) setTenantName(actualTenantName)

      if (data.setup_completed !== undefined) {
        setSetupCompleted(data.setup_completed)
      }

      const op = data.user || {
        id: data.operator_id || 'operator',
        name: data.identifier || '',
        email: data.email || null,
        whatsapp_number: data.whatsapp_number || null,
        role: (data.role || 'admin').toLowerCase(),
        tenant_id: actualTenantId,
        is_active: true,
      }

      queryClient.setQueryData(authKeys.me, {
        user: op,
        tenant_id: actualTenantId,
        tenant_slug: actualTenantSlug,
        tenant_name: actualTenantName,
      })

      return data
    },
    [queryClient, tenantId, tenantSlug]
  )

  const login = async (
    tenantOrSlug: string,
    identifier: string,
    codeOrPassword: string,
    rememberMe: boolean = false
  ) => {
    if (rememberMe) {
      localStorage.setItem(REMEMBER_ME_KEY, 'true')
    } else {
      localStorage.removeItem(REMEMBER_ME_KEY)
    }
    setStoredTenant(tenantOrSlug, tenantOrSlug)
    setTenantId(tenantOrSlug)
    setTenantSlug(tenantOrSlug)

    const isEmail = identifier.includes('@')
    const response = await apiClient.post(
      '/dashboard/api/login',
      {
        tenant_id: tenantOrSlug,
        tenant_slug: tenantOrSlug,
        identifier,
        email: isEmail ? identifier : undefined,
        whatsapp_number: !isEmail ? identifier : undefined,
        totp_code: codeOrPassword,
        password: codeOrPassword,
        remember_me: rememberMe,
      },
      {
        headers: { 'X-Tenant': tenantOrSlug, 'X-Tenant-Slug': tenantOrSlug },
      }
    )

    return handleLoginResponse(response)
  }

  const loginWithBackupCode = async (
    tenantOrSlug: string,
    identifier: string,
    backupCode: string,
    rememberMe: boolean = false
  ) => {
    if (rememberMe) {
      localStorage.setItem(REMEMBER_ME_KEY, 'true')
    } else {
      localStorage.removeItem(REMEMBER_ME_KEY)
    }
    setStoredTenant(tenantOrSlug, tenantOrSlug)
    setTenantId(tenantOrSlug)
    setTenantSlug(tenantOrSlug)

    const isEmail = identifier.includes('@')
    const response = await apiClient.post(
      '/dashboard/api/login/backup-code',
      {
        tenant_id: tenantOrSlug,
        tenant_slug: tenantOrSlug,
        identifier,
        email: isEmail ? identifier : undefined,
        whatsapp_number: !isEmail ? identifier : undefined,
        backup_code: backupCode.trim(),
        remember_me: rememberMe,
      },
      {
        headers: { 'X-Tenant': tenantOrSlug, 'X-Tenant-Slug': tenantOrSlug },
      }
    )

    return handleLoginResponse(response)
  }

  const logout = async () => {
    try {
      await apiClient.post('/dashboard/api/logout')
    } finally {
      setStoredTenant(null, null)
      setTenantId(null)
      setTenantSlug(null)
      setTenantName(null)
      queryClient.setQueryData(authKeys.me, null)
    }
  }

  const refreshUser = async () => {
    await refetch()
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoading,
        user,
        tenantId,
        tenantSlug,
        tenantName,
        setupCompleted,
        login,
        loginWithBackupCode,
        logout,
        refreshUser,
        setSessionUser,
        checkSetupStatus,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
