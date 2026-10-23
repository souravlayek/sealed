import { DocsPage } from '@/components/docs/docs-page'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { Navbar } from '@/components/layout/navbar'
import { createFileRoute } from '@tanstack/react-router'

const DocsRoute = () => {
  const { auth } = Route.useRouteContext()

   if (auth.isAuthenticated) {
     return (
       <AuthenticatedLayout>
         <DocsPage isAuthenticated={true} />
       </AuthenticatedLayout>
     )
   }

  return <>
    <Navbar />
    <DocsPage isAuthenticated={false} />
   </>
 }

export const Route = createFileRoute('/docs')({
  component: DocsRoute,
})
