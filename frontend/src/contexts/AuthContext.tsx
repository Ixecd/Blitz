import { createContext, useContext, useState, useCallback } from 'react'
import { api } from '@/api/client'

interface AuthState {
  token:    string | null
  email: string | null
  userID:   number | null
}

interface AuthContextValue extends AuthState {
  login:  (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, setState] = useState<AuthState>({
    token:    localStorage.getItem('access_token'),
    email: localStorage.getItem('email'),
    userID:   Number(localStorage.getItem('user_id')) || null,
  })

  const login = useCallback(async (email: string, password: string) => {
    const data = await api.post<{
      access_token:  string
      refresh_token: string
      email:      string
      user_id:       number
    }>('/api/v1/login', { email, password })

    localStorage.setItem('access_token',  data.access_token)
    localStorage.setItem('refresh_token', data.refresh_token)
    localStorage.setItem('email',      data.email)
    localStorage.setItem('user_id',       String(data.user_id))

    setState({ token: data.access_token, email: data.email, userID: data.user_id })
  }, [])

  const logout = useCallback(async () => {
    const rt = localStorage.getItem('refresh_token')
    if (rt) {
      await api.post('/api/v1/logout', { refresh_token: rt }).catch(() => {})
    }
    localStorage.clear()
    setState({ token: null, email: null, userID: null })
  }, [])

  return (
    <AuthContext.Provider value={{ ...state, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be inside AuthProvider')
  return ctx
}
