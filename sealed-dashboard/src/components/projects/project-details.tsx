// src/components/projects/project-details.tsx
import { useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { AlertTriangle, Pencil, Plus, Trash2 } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { getProjectDetails, getRotation, type SecretMeta } from "@/service/project_service"
import {
  AddEnvironmentDialog,
  DeleteSecretDialog,
  SecretDialog,
} from "./project-dialogs"

type DialogState =
  | { type: "add-secret" }
  | { type: "edit-secret"; secret: SecretMeta }
  | { type: "delete-secret"; secret: SecretMeta }
  | { type: "add-env" }
  | null

export function ProjectDetails({ slug }: { slug: string }) {
  const { data: project, isLoading, isError } = useQuery({
    queryKey: ["project", slug],
    queryFn: () => getProjectDetails(slug),
  })
  const [activeEnv, setActiveEnv] = useState<string>()
  const [dialog, setDialog] = useState<DialogState>(null)

  if (isLoading) return <p className="text-muted-foreground">Loading project…</p>
  if (isError || !project) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Project not found</AlertTitle>
        <AlertDescription>It may not exist or you may not have access.</AlertDescription>
      </Alert>
    )
  }

  const env = activeEnv ?? project.environments[0]
  const secrets = env ? project.secrets[env] ?? [] : []

  // Key parity: keys defined in another environment but missing here
  const allKeys = new Set(
    project.environments.flatMap((e) => (project.secrets[e] ?? []).map((s) => s.key_name))
  )
  const currentKeys = new Set(secrets.map((s) => s.key_name))
  const missing = [...allKeys].filter((k) => !currentKeys.has(k))

  const close = () => setDialog(null)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{project.name}</h1>
        <p className="text-sm text-muted-foreground">
          {project.slug} · your role: {project.my_role}
        </p>
      </div>

      {project.environments.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-10">
            <p className="text-muted-foreground">No environments yet.</p>
            <Button onClick={() => setDialog({ type: "add-env" })}>
              <Plus /> Add environment
            </Button>
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="flex items-center justify-between gap-2">
            <Tabs value={env} onValueChange={(v) => setActiveEnv(v as string)}>
              <TabsList>
                {project.environments.map((e) => (
                  <TabsTrigger key={e} value={e}>
                    {e}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
            <Button variant="outline" size="sm" onClick={() => setDialog({ type: "add-env" })}>
              <Plus /> Environment
            </Button>
          </div>

          <Card>
            <CardHeader className="flex-row items-center justify-between space-y-0">
              <div>
                <CardTitle className="font-mono">{env}</CardTitle>
                <CardDescription>{secrets.length} secrets</CardDescription>
              </div>
              <Button size="sm" onClick={() => setDialog({ type: "add-secret" })}>
                <Plus /> Add secret
              </Button>
            </CardHeader>

            <CardContent className="space-y-4">
              {missing.length > 0 && (
                <Alert>
                  <AlertTriangle />
                  <AlertTitle>Keys missing in {env}</AlertTitle>
                  <AlertDescription className="font-mono">{missing.join(", ")}</AlertDescription>
                </Alert>
              )}

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Key</TableHead>
                    <TableHead>Rotation</TableHead>
                    <TableHead>Last changed</TableHead>
                    <TableHead className="text-right">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {secrets.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center text-muted-foreground">
                        No secrets in {env}.
                      </TableCell>
                    </TableRow>
                  )}
                  {secrets.map((s) => (
                    <TableRow key={s.key_name}>
                      <TableCell className="font-mono">{s.key_name}</TableCell>
                      <TableCell>
                        <RotationBadge secret={s} />
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {s.value_changed_at
                          ? new Date(s.value_changed_at).toLocaleDateString()
                          : "—"}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label={`Edit ${s.key_name}`}
                          onClick={() => setDialog({ type: "edit-secret", secret: s })}
                        >
                          <Pencil />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          aria-label={`Delete ${s.key_name}`}
                          onClick={() => setDialog({ type: "delete-secret", secret: s })}
                        >
                          <Trash2 />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </>
      )}

      {dialog?.type === "add-secret" && (
        <SecretDialog
          projectId={project.id}
          slug={slug}
          environment={env}
          onClose={close}
        />
      )}
      {dialog?.type === "edit-secret" && (
        <SecretDialog
          projectId={project.id}
          slug={slug}
          environment={env}
          existing={dialog.secret}
          onClose={close}
        />
      )}
      {dialog?.type === "delete-secret" && (
        <DeleteSecretDialog
          projectId={project.id}
          slug={slug}
          environment={env}
          secretKey={dialog.secret.key_name}
          onClose={close}
        />
      )}
      {dialog?.type === "add-env" && (
        <AddEnvironmentDialog projectId={project.id} slug={slug} onClose={close} />
      )}
    </div>
  )
}

function RotationBadge({ secret }: { secret: SecretMeta }) {
  const r = getRotation(secret)
  if (!r) return <Badge variant="outline">Off</Badge>
  if (r.daysLeft < 0) return <Badge variant="destructive">Overdue {-r.daysLeft}d</Badge>
  if (r.daysLeft <= 7) return <Badge variant="secondary">Due in {r.daysLeft}d</Badge>
  return <Badge variant="outline">Due in {r.daysLeft}d</Badge>
}
