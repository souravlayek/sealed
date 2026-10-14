import { ProjectDetails } from '@/components/projects/project-details'
import { getProjectDetails } from '@/service/project_service'
import { createFileRoute, useLoaderData } from '@tanstack/react-router'

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
