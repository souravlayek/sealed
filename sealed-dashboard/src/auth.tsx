import React, { createContext, useContext, useState, useEffect } from 'react'
import { getUserDetails, loginUser, type UserResponse } from './service/auth_service'

interface AuthState {
  isAuthenticated: boolean
  user: UserResponse | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthState | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserResponse | null>(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const getAndStoreUserDetails = async () => {
    const token = localStorage.getItem('access_token')
    if (token) {
      try {
      const userDetails = await getUserDetails()
      setUser(userDetails)
      setIsAuthenticated(true)
      setIsLoading(false)

      } catch (error) {
        console.error(error)
        setIsAuthenticated(false)
        setUser(null)
        setIsLoading(false)
      }
    } else {
      setIsLoading(false)
    }
  }
  // Restore auth state on app load
  useEffect(() => {
    getAndStoreUserDetails()
  }, [])

  // Show loading state while checking auth
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        Loading...
      </div>
    )
  }

  const login = async (username: string, password: string) => {
    const response = await loginUser(username, password)
    localStorage.setItem("access_token", response.access_token)
    localStorage.setItem("refresh_token", response.refresh_token)
    const user = await getUserDetails()
    setIsAuthenticated(true)
    setUser(user)
  }

  const logout = () => {
    setUser(null)
    setIsAuthenticated(false)
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  return (
    <AuthContext.Provider value={{ isAuthenticated, user, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
