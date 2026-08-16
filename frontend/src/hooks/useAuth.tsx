import React, { useState, useEffect, useContext, createContext, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import apiClient, { TENANT_ID_KEY, REMEMBER_ME_KEY } from '@/lib/apiClient'
import type { Operator } from '@/types'

interface AuthContextType {
  isAuthenticated: boolean
  isLoading: boolean
  user: Operator | null
  tenantId: string | null
  setupCompleted: boolean
  login: (tenantId: string, identifier: string, codeOrPassword: string, rememberMe?: boolean) => Promise<any>
  loginWithBackupCode: (tenantId: string, identifier: string, backupCode: string, rememberMe?: boolean) => Promise<any>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  setSessionUser: (user: Operator, tenantId?: string) => void
  checkSetupStatus: () => Promise<boolean>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState<Operator | null>(null)
  const [tenantId, setTenantId] = useState<string | null>(() => localStorage.getItem(TENANT_ID_KEY))
  const [setupCompleted, setSetupCompleted] = useState<boolean>(true)
  const [isLoading, setIsLoading] = useState(true)

  const { data, refetch } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const response = await apiClient.get('/dashboard/api/me')
      return response.data
    },
    enabled: false,
    retry: false,
    staleTime: 1000 * 60 * 5,
  })

  // Trigger the initial /me check once on mount.
  useEffect(() => {
    refetch()
      .then((res) => {
        if (res.data?.user) {
          setIsAuthenticated(true)
          setUser(res.data.user)
          if (res.data.user.tenant_id) {
            setTenantId(res.data.user.tenant_id)
            localStorage.setItem(TENANT_ID_KEY, res.data.user.tenant_id)
          }
        }
      })
      .catch(() => {
        setIsAuthenticated(false)
        setUser(null)
      })
      .finally(() => setIsLoading(false))
  }, [])

  // Update auth state when /me data arrives.
  useEffect(() => {
    if (data?.user) {
      setIsAuthenticated(true)
      setUser(data.user)
      if (data.user.tenant_id) {
        setTenantId(data.user.tenant_id)
        localStorage.setItem(TENANT_ID_KEY, data.user.tenant_id)
      }
    }
  }, [data])

  const checkSetupStatus = useCallback(async (): Promise<boolean> => {
    try {
      const res = await apiClient.get('/dashboard/api/tenant/setup-status')
      const isCompleted = !!res.data?.completed
      setSetupCompleted(isCompleted)
      return isCompleted
    } catch {
      // If status endpoint returns error, default to true so user isn't stuck
      setSetupCompleted(true)
      return true
    }
  }, [])

  const setSessionUser = useCallback((newUser: Operator, newTenantId?: string) => {
    const tid = newTenantId || newUser.tenant_id
    if (tid) {
      localStorage.setItem(TENANT_ID_KEY, tid)
      setTenantId(tid)
    }
    setUser(newUser)
    setIsAuthenticated(true)
  }, [])

  const login = async (
    targetTenantId: string,
    identifier: string,
    codeOrPassword: string,
    rememberMe: boolean = false
  ) => {
    localStorage.setItem(TENANT_ID_KEY, targetTenantId)
    if (rememberMe) {
      localStorage.setItem(REMEMBER_ME_KEY, 'true')
    } else {
      localStorage.removeItem(REMEMBER_ME_KEY)
    }
    setTenantId(targetTenantId)

    try {
      const isEmail = identifier.includes('@')
      const response = await apiClient.post(
        '/dashboard/api/login',
        {
          identifier,
          email: isEmail ? identifier : undefined,
          whatsapp_number: !isEmail ? identifier : undefined,
          totp_code: codeOrPassword,
          password: codeOrPassword,
          remember_me: rememberMe,
        },
        {
          headers: { 'X-Tenant': targetTenantId },
        }
      )

      setIsAuthenticated(true)
      const op = response.data.user || {
        id: response.data.operator_id || 'operator',
        name: identifier,
        email: isEmail ? identifier : null,
        whatsapp_number: !isEmail ? identifier : null,
        role: response.data.role || 'ADMIN',
        tenant_id: targetTenantId,
        is_active: true,
      }
      setUser(op)

      // Check tenant setup status if applicable
      if (response.data.setup_completed !== undefined) {
        setSetupCompleted(response.data.setup_completed)
      }

      return response.data
    } catch (error: any) {
      localStorage.removeItem(TENANT_ID_KEY)
      setTenantId(null)
      const msg = error.response?.data?.error || error.message || 'Invalid credentials or tenant'
      throw new Error(msg)
    }
  }

  const loginWithBackupCode = async (
    targetTenantId: string,
    identifier: string,
    backupCode: string,
    rememberMe: boolean = false
  ) => {
    localStorage.setItem(TENANT_ID_KEY, targetTenantId)
    if (rememberMe) {
      localStorage.setItem(REMEMBER_ME_KEY, 'true')
    } else {
      localStorage.removeItem(REMEMBER_ME_KEY)
    }
    setTenantId(targetTenantId)

    try {
      const isEmail = identifier.includes('@')
      const response = await apiClient.post(
        '/dashboard/api/login/backup-code',
        {
          identifier,
          email: isEmail ? identifier : undefined,
          whatsapp_number: !isEmail ? identifier : undefined,
          backup_code: backupCode.trim(),
          remember_me: rememberMe,
        },
        {
          headers: { 'X-Tenant': targetTenantId },
        }
      )

      setIsAuthenticated(true)
      const op = response.data.user || {
        id: response.data.operator_id || 'operator',
        name: identifier,
        email: isEmail ? identifier : null,
        whatsapp_number: !isEmail ? identifier : null,
        role: response.data.role || 'ADMIN',
        tenant_id: targetTenantId,
        is_active: true,
      }
      setUser(op)
      return response.data
    } catch (error: any) {
      localStorage.removeItem(TENANT_ID_KEY)
      setTenantId(null)
      const msg = error.response?.data?.error || error.message || 'Invalid backup code or tenant'
      throw new Error(msg)
    }
  }

  const logout = async () => {
    try {
      await apiClient.post('/dashboard/api/logout')
    } finally {
      localStorage.removeItem(TENANT_ID_KEY)
      setTenantId(null)
      setIsAuthenticated(false)
      setUser(null)
    }
  }

  const refreshUser = async () => {
    try {
      const res = await apiClient.get('/dashboard/api/me')
      if (res.data?.user) {
        setUser(res.data.user)
        setIsAuthenticated(true)
      }
    } catch {
      // ignore
    }
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isLoading,
        user,
        tenantId,
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
