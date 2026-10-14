import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { createFileRoute, redirect, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: ({ context, location }) => {
    if (!context.auth.isAuthenticated) {
      throw redirect({
        to: '/login',
        search: {
          // Save current location for redirect after login
          redirect: location.href,
        },
      })
    }
  },

  component: () => (<AuthenticatedLayout>
    <Outlet />
  </AuthenticatedLayout>),
})
