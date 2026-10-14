import { ThemeProvider } from '@/components/theme-provider';
import type { UserResponse } from '@/service/auth_service';
import { createRootRouteWithContext, Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'
import { TooltipProvider } from "@/components/ui/tooltip"
interface AuthState {
  isAuthenticated: boolean
  user: UserResponse | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

interface MyRouterContext {
  auth: AuthState
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  component: () => (
    <ThemeProvider defaultTheme='dark'>
      <TooltipProvider>
      <Outlet />
      <TanStackRouterDevtools />
      </TooltipProvider>
    </ThemeProvider>
  ),
})
