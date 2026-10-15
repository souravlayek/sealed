import { ProjectDetails } from '@/components/projects/project-details'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute(
  '/_authenticated/projects/$projectSlug',
)({
  component: RouteComponent,
})

function RouteComponent() {
    const { projectSlug } = Route.useParams()
  return (
  <div className="p-6">
    <ProjectDetails slug={projectSlug} />
  </div>)
}
