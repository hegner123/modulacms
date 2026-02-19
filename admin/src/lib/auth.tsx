import { createContext, useContext, useCallback, useMemo } from 'react'
import type { ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import type { MeResponse, Email } from '@modulacms/admin-sdk'
import { sdk, sessionSdk, activateSessionAuth, activateApiKeyAuth } from '@/lib/sdk'
import { queryKeys } from '@/lib/query-keys'
import { useMe } from '@/queries/auth'

export type AuthUser = {
  user_id: string
  email: string
  username: string
  name: string
  role: string
}

type AuthContextValue = {
  user: AuthUser | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

const hasApiKey = !!import.meta.env.VITE_CMS_API_KEY

function toAuthUser(me: MeResponse): AuthUser {
  return {
    user_id: me.user_id,
    email: me.email,
    username: me.username,
    name: me.name,
    role: me.role,
  }
}

function useApiKeyUser(enabled: boolean) {
  return useQuery({
    queryKey: ['apiKeyUser'],
    queryFn: async () => {
      const users = await sdk.users.list()
      const first = users[0]
      if (!first) return null
      return {
        user_id: first.user_id,
        email: first.email,
        username: first.username,
        name: first.name,
        role: first.role,
      } satisfies AuthUser
    },
    enabled,
    staleTime: Infinity,
  })
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const { data: me, isLoading: meLoading } = useMe()

  // Switch active SDK client based on session state
  if (me) {
    activateSessionAuth()
  } else if (!meLoading) {
    activateApiKeyAuth()
  }

  // Fall back to API key user only when session auth failed and an API key exists
  const useApiFallback = hasApiKey && !meLoading && !me
  const { data: apiUser, isLoading: apiUserLoading } = useApiKeyUser(useApiFallback)

  const isLoading = meLoading || (useApiFallback && apiUserLoading)
  const user = me ? toAuthUser(me) : (apiUser ?? null)

  const login = useCallback(async (email: string, password: string) => {
    await sessionSdk.auth.login({ email: email as Email, password })
    activateSessionAuth()
    await queryClient.invalidateQueries({ queryKey: queryKeys.auth.me() })
  }, [queryClient])

  const logout = useCallback(async () => {
    await sessionSdk.auth.logout()
    activateApiKeyAuth()
    queryClient.setQueryData(queryKeys.auth.me(), null)
    queryClient.clear()
  }, [queryClient])

  const value = useMemo<AuthContextValue>(() => ({
    user,
    isLoading,
    login,
    logout,
  }), [user, isLoading, login, logout])

  return (
    <AuthContext value={value}>
      {children}
    </AuthContext>
  )
}

export function useAuthContext(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuthContext must be used within an AuthProvider')
  }
  return context
}
