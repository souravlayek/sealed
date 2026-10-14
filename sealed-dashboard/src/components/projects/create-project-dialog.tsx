// src/components/projects/create-project-dialog.tsx
import { useState, type FormEvent } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { createProject } from "@/service/project_service"
import { queryKeys } from "@/lib/query-keys"

const ENV_RE = /^[a-z0-9-]+$/

export function CreateProjectDialog({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient()
  const navigate = useNavigate()
  const [name, setName] = useState("")
  const [envs, setEnvs] = useState("dev, staging, prod")
  const [error, setError] = useState<string>()

  const create = useMutation({
    mutationFn: createProject,
    onSuccess: async (project) => {
      await qc.invalidateQueries({ queryKey: queryKeys.projects })
      onClose()
      if (project?.slug) {
        navigate({ to: "/projects/$projectSlug", params: { projectSlug: project.slug } })
      }
    },
    onError: () => setError("Could not create the project. Try again."),
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    const environments = envs
      .split(",")
      .map((s) => s.trim().toLowerCase())
      .filter(Boolean)

    if (!name.trim()) return setError("Project name is required.")
    if (environments.some((env) => !ENV_RE.test(env))) {
      return setError("Environments: lowercase letters, numbers and hyphens only.")
    }
    setError(undefined)
    create.mutate({ name: name.trim(), environments })
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <form onSubmit={submit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>New project</DialogTitle>
            <DialogDescription>You become the admin of this project.</DialogDescription>
          </DialogHeader>

          <div className="space-y-2">
            <Label htmlFor="project-name">Name</Label>
            <Input
              id="project-name"
              placeholder="Payments API"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="project-envs">Environments (comma separated)</Label>
            <Input
              id="project-envs"
              className="font-mono"
              value={envs}
              onChange={(e) => setEnvs(e.target.value)}
            />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={create.isPending}>
              {create.isPending ? "Creating…" : "Create"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
