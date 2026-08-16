import axios from 'axios'
import type {
  BackupCodesResponse,
  Invitation,
  InvitationDetails,
  MediaUploadResponse,
  Operator,
  PairingSnapshot,
  Tenant,
  TenantSetupStatus,
  TenantSetupUpdate,
  TOTPSetupData,
  TOTPStatus,
  TOTPVerifyResult,
} from '@/types'

export const TENANT_ID_KEY = 'whatsapp_dashboard_tenant_id'
export const REMEMBER_ME_KEY = 'whatsapp_dashboard_remember_me'

const apiClient = axios.create({
  baseURL: import.meta.env.DEV ? 'http://localhost:8080' : '',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
})

apiClient.interceptors.request.use(config => {
  const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content')
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken
  }
  const tenantId = localStorage.getItem(TENANT_ID_KEY)
  if (tenantId && !config.headers['X-Tenant']) {
    config.headers['X-Tenant'] = tenantId
  }
  return config
})

const PUBLIC_PATH_PREFIXES = [
  '/dashboard/login',
  '/dashboard/signup',
  '/dashboard/verify-email',
  '/dashboard/invitation',
  '/dashboard/recovery',
  '/login',
  '/signup',
  '/verify-email',
  '/invitation',
  '/recovery',
]

apiClient.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      const currentPath = window.location.pathname
      const isPublicPath = PUBLIC_PATH_PREFIXES.some(prefix => currentPath.startsWith(prefix))
      if (!isPublicPath) {
        localStorage.removeItem(TENANT_ID_KEY)
        window.location.href = '/dashboard/login'
      }
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  login: async (tenantId: string, identifier: string, totpCode: string, rememberMe: boolean = false) => {
    localStorage.setItem(TENANT_ID_KEY, tenantId)
    const response = await apiClient.post(
      '/dashboard/api/login',
      {
        identifier,
        email: identifier.includes('@') ? identifier : undefined,
        whatsapp_number: !identifier.includes('@') ? identifier : undefined,
        totp_code: totpCode,
        remember_me: rememberMe,
      },
      { headers: { 'X-Tenant': tenantId } }
    )
    return response.data
  },

  loginWithBackupCode: async (tenantId: string, identifier: string, backupCode: string, rememberMe: boolean = false) => {
    localStorage.setItem(TENANT_ID_KEY, tenantId)
    const response = await apiClient.post(
      '/dashboard/api/login/backup-code',
      {
        identifier,
        email: identifier.includes('@') ? identifier : undefined,
        whatsapp_number: !identifier.includes('@') ? identifier : undefined,
        backup_code: backupCode.trim(),
        remember_me: rememberMe,
      },
      { headers: { 'X-Tenant': tenantId } }
    )
    return response.data
  },

  logout: async () => {
    const response = await apiClient.post('/dashboard/api/logout')
    return response.data
  },

  getMe: async () => {
    const response = await apiClient.get('/dashboard/api/me')
    return response.data
  },
}

export const onboardingApi = {
  signupTenant: async (data: {
    org_name: string
    admin_name: string
    admin_email: string
    whatsapp_number: string
  }) => {
    const response = await apiClient.post('/dashboard/api/signup/tenant', {
      org_name: data.org_name,
      tenant_name: data.org_name,
      name: data.admin_name,
      admin_name: data.admin_name,
      email: data.admin_email,
      admin_email: data.admin_email,
      whatsapp_number: data.whatsapp_number,
      admin_whatsapp: data.whatsapp_number,
    })
    return response.data as {
      tenant_id?: string
      tenant?: Tenant
      operator?: Operator
      operator_id?: string
      verification_token?: string
      temp_token?: string
      setup_token?: string
      totp_setup?: TOTPSetupData
      email_verification_required?: boolean
      setup_url?: string
      message?: string
    }
  },

  verifyEmail: async (token: string) => {
    const response = await apiClient.post('/dashboard/api/verify-email', { token })
    return response.data as {
      status?: string
      verified: boolean
      tenant?: Tenant
      tenant_id?: string
      operator?: Operator
      operator_id?: string
      setup_token?: string
      temp_token?: string
      totp_setup?: TOTPSetupData
      setup_url?: string
      message?: string
    }
  },

  getTotpSetup: async (token: string) => {
    const response = await apiClient.get(`/dashboard/api/totp/setup/${encodeURIComponent(token)}`)
    return response.data as TOTPSetupData
  },

  verifyTotpSetup: async (token: string, code: string) => {
    const response = await apiClient.post('/dashboard/api/totp/verify-setup', {
      token,
      setup_token: token,
      code,
      totp_code: code,
    })
    return response.data as TOTPVerifyResult
  },

  getInvitationDetails: async (token: string) => {
    const response = await apiClient.get(`/dashboard/api/invitations/accept/${encodeURIComponent(token)}`)
    return response.data as InvitationDetails
  },

  signupOperator: async (data: {
    token: string
    name: string
    whatsapp_number?: string
    email?: string
  }) => {
    const response = await apiClient.post('/dashboard/api/signup/operator', {
      token: data.token,
      invitation_token: data.token,
      name: data.name,
      whatsapp_number: data.whatsapp_number,
      email: data.email,
    })
    return response.data as {
      status: string
      operator: Operator
      operator_id?: string
      setup_token: string
      temp_token?: string
      setup_url?: string
    }
  },

  acceptOperatorInvitation: async (data: {
    token: string
    name: string
    email?: string
    whatsapp_number?: string
    totp_code: string
  }) => {
    const signupRes = await onboardingApi.signupOperator({
      token: data.token,
      name: data.name,
      email: data.email,
      whatsapp_number: data.whatsapp_number,
    })
    const setupToken = signupRes.setup_token || signupRes.temp_token || ''
    return await onboardingApi.verifyTotpSetup(setupToken, data.totp_code)
  },

  requestRecovery: async (data: { tenant_id: string; identifier: string; channel?: 'whatsapp' | 'email' }) => {
    const response = await apiClient.post('/dashboard/api/recovery/request', data)
    return response.data as { success: boolean; message?: string; instructions?: string }
  },
}

export const totpApi = {
  getStatus: async () => {
    const response = await apiClient.get('/dashboard/api/account/totp')
    return response.data as TOTPStatus
  },

  regenerateBackupCodes: async (totpCode: string) => {
    const response = await apiClient.post('/dashboard/api/account/totp/regenerate-backup-codes', {
      totp_code: totpCode,
    })
    return response.data as BackupCodesResponse
  },
}

export const invitationsApi = {
  listInvitations: async () => {
    const response = await apiClient.get('/dashboard/api/invitations')
    return (Array.isArray(response.data) ? response.data : response.data.items || []) as Invitation[]
  },

  createWhatsAppInvitation: async (whatsappNumber: string, role: string) => {
    const response = await apiClient.post('/dashboard/api/invitations/whatsapp', {
      whatsapp_number: whatsappNumber,
      role,
    })
    return response.data as Invitation
  },

  createEmailInvitation: async (email: string, role: string) => {
    const response = await apiClient.post('/dashboard/api/invitations/email', {
      email,
      role,
    })
    return response.data as Invitation
  },

  revokeInvitation: async (id: string) => {
    const response = await apiClient.delete(`/dashboard/api/invitations/${encodeURIComponent(id)}`)
    return response.data
  },

  listOperators: async () => {
    const response = await apiClient.get('/dashboard/api/operators')
    return (Array.isArray(response.data) ? response.data : response.data.items || []) as Operator[]
  },

  resetOperatorTotp: async (operatorId: string) => {
    const response = await apiClient.post(`/dashboard/api/operators/${encodeURIComponent(operatorId)}/totp-reset`)
    return response.data as { success: boolean; reset_token?: string; message?: string }
  },

  getOperatorTotpStatus: async (operatorId: string) => {
    const response = await apiClient.get(`/dashboard/api/operators/${encodeURIComponent(operatorId)}/totp-status`)
    return response.data as TOTPStatus
  },
}

export const setupApi = {
  getStatus: async () => {
    const response = await apiClient.get('/dashboard/api/tenant/setup-status')
    return response.data as TenantSetupStatus
  },

  updateSetup: async (data: TenantSetupUpdate) => {
    const response = await apiClient.put('/dashboard/api/tenant/setup', data)
    return response.data as TenantSetupStatus
  },

  completeSetup: async () => {
    const response = await apiClient.post('/dashboard/api/tenant/complete-setup')
    return response.data as { success: boolean; completed: boolean }
  },
}

export const pairingApi = {
  start: async (displayName?: string) => {
    const response = await apiClient.post<{ pairing_id: string }>('/dashboard/api/pairing', {
      display_name: displayName,
    })
    return response.data
  },

  get: async (id: string) => {
    const response = await apiClient.get<PairingSnapshot>(`/dashboard/api/pairing/${encodeURIComponent(id)}`)
    return response.data
  },

  cancel: async (id: string) => {
    const response = await apiClient.post<{ status: string }>(`/dashboard/api/pairing/${encodeURIComponent(id)}/cancel`)
    return response.data
  },
}

export const mediaApi = {
  uploadMedia: async (file: File, onProgress?: (percent: number) => void) => {
    const formData = new FormData()
    formData.append('file', file)
    const response = await apiClient.post<MediaUploadResponse>('/dashboard/api/media', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: progressEvent => {
        if (progressEvent.total && onProgress) {
          const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(percent)
        }
      },
    })
    return response.data
  },
}

export default apiClient
