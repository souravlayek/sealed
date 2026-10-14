// src/components/dashboard/dashboard.tsx
import { useMemo, useState } from "react"
import { Link } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { Folder, Layers, Plus, ShieldCheck, Clock } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { getAllProjects, type Project } from "@/service/project_service"
import { getUserDetails } from "@/service/auth_service"
import { queryKeys } from "@/lib/query-keys"
import { CreateProjectDialog } from "@/components/projects/create-project-dialog"

export function Dashboard() {
  const [creating, setCreating] = useState(false)

  const { data: user } = useQuery({
    queryKey: queryKeys.me,
    queryFn: getUserDetails,
    staleTime: 5 * 60_000,
  })

  const { data: projects = [], isLoading, isError } = useQuery({
    queryKey: queryKeys.projects,
    queryFn: getAllProjects,
  })

  const stats = useMemo(() => summarize(projects), [projects])

  return (
    <div className="mx-auto max-w-5xl space-y-6 p-6">
      <div className="flex items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">
            {user?.name ? `Welcome back, ${user.name.split(" ")[0]}` : "Dashboard"}
          </h1>
          <p className="text-sm text-muted-foreground">Your projects and environments at a glance.</p>
        </div>
        <Button onClick={() => setCreating(true)}>
          <Plus /> New project
        </Button>
      </div>

      {isError && (
        <Card>
          <CardContent className="pt-6 text-sm text-destructive">
            Couldn't load your projects. Refresh to try again.
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard icon={Folder} label="Projects" value={isLoading ? "–" : stats.total} />
        <StatCard
          icon={ShieldCheck}
          label="Admin of"
          value={isLoading ? "–" : stats.admin}
          hint={`${stats.member} as member`}
        />
        <StatCard
          icon={Layers}
          label="Environments"
          value={isLoading ? "–" : stats.environments}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-5">
        <Card className="lg:col-span-3">
          <CardHeader>
            <CardTitle className="text-base">Recent projects</CardTitle>
            <CardDescription>Newest first</CardDescription>
          </CardHeader>
          <CardContent>
            {!isLoading && projects.length === 0 ? (
              <EmptyState onCreate={() => setCreating(true)} />
            ) : (
              <ul className="divide-y">
                {stats.recent.map((p) => (
                  <li key={p.id}>
                    <Link
                      to="/projects/$projectSlug"
                      params={{ projectSlug: p.slug }}
                      className="flex items-center justify-between gap-4 py-3 transition-colors hover:text-primary"
                    >
                      <div className="min-w-0">
                        <div className="truncate font-medium">{p.name}</div>
                        <div className="truncate font-mono text-xs text-muted-foreground">{p.slug}</div>
                      </div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <Clock className="size-3" />
                        {new Date(p.created_at).toLocaleDateString()}
                      </div>
                    </Link>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Environments in use</CardTitle>
            <CardDescription>How many projects use each name</CardDescription>
          </CardHeader>
          <CardContent>
            {stats.envBreakdown.length === 0 ? (
              <p className="text-sm text-muted-foreground">No environments yet.</p>
            ) : (
              <div className="flex flex-wrap gap-2">
                {stats.envBreakdown.map(([env, count]) => (
                  <Badge key={env} variant="outline" className="font-mono">
                    {env}
                    <span className="ml-2 text-muted-foreground">{count}</span>
                  </Badge>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card className="border-dashed">
        <CardContent className="flex items-center justify-between gap-4 pt-6 text-sm">
          <p className="text-muted-foreground">
            Secret counts and device status will appear here once those endpoints are available.
          </p>
          <Button variant="outline" size="sm" render={<Link to="/docs" />}>
            Command docs
          </Button>
        </CardContent>
      </Card>

      {creating && <CreateProjectDialog onClose={() => setCreating(false)} />}
    </div>
  )
}

function StatCard({
  icon: Icon,
  label,
  value,
  hint,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: number | string
  hint?: string
}) {
  return (
    <Card>
      <CardContent className="flex items-start justify-between pt-6">
        <div>
          <p className="text-sm text-muted-foreground">{label}</p>
          <p className="mt-1 text-3xl font-semibold tracking-tight">{value}</p>
          {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
        </div>
        <Icon className="size-4 text-muted-foreground" />
      </CardContent>
    </Card>
  )
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 py-8 text-center">
      <p className="text-sm text-muted-foreground">No projects yet.</p>
      <Button size="sm" onClick={onCreate}>
        <Plus /> Create your first project
      </Button>
    </div>
  )
}

function summarize(projects: Project[]) {
  const admin = projects.filter((p) => p.role === "PROJECT_ADMIN").length

  const envCounts = new Map<string, number>()
  for (const p of projects) {
    for (const env of p.environments) {
      envCounts.set(env, (envCounts.get(env) ?? 0) + 1)
    }
  }

  const recent = [...projects]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5)

  return {
    total: projects.length,
    admin,
    member: projects.length - admin,
    environments: projects.reduce((n, p) => n + p.environments.length, 0),
    envBreakdown: [...envCounts.entries()].sort((a, b) => b[1] - a[1]),
    recent,
  }
}
