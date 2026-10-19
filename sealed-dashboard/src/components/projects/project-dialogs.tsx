// src/components/projects/project-dialogs.tsx
import { useState, type FormEvent } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
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
import {
  addEnvironment,
  createSecret,
  deleteSecret,
  type SecretMeta,
} from "@/service/project_service"
import { useNavigate } from "@tanstack/react-router"
import { ConfirmDialog } from "@/components/confirm-dialog"
import { deleteProject, removeEnvironment } from "@/service/project_service"
import { queryKeys } from "@/lib/query-keys"


const KEY_RE = /^[A-Za-z_][A-Za-z0-9_]*$/
const ENV_RE = /^[a-z0-9-]+$/

type Common = { projectId: number; slug: string; onClose: () => void }

function useRefreshProject(slug: string) {
  const qc = useQueryClient()
  return () => qc.invalidateQueries({ queryKey: ["project", slug] })
}

export function SecretDialog({
  projectId,
  slug,
  environment,
  existing,
  onClose,
}: Common & { environment: string; existing?: SecretMeta }) {
  const refresh = useRefreshProject(slug)
  const [key, setKey] = useState(existing?.key_name ?? "")
  const [value, setValue] = useState("")
  const [rotationDays, setRotationDays] = useState(String(existing?.rotation_days ?? 0))
  const [error, setError] = useState<string>()

  const save = useMutation({
    mutationFn: () =>
      createSecret(projectId, {
        environment,
        key_name: key,
        secret: value,
        rotation_days: Number(rotationDays) || 0,
      }),
    onSuccess: async () => {
      await refresh()
      onClose()
    },
    onError: () => setError("Could not save the secret. Try again."),
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (!KEY_RE.test(key)) {
      setError("Use letters, numbers and underscores. Don't start with a number.")
      return
    }
    if (!value) {
      setError("Value is required.")
      return
    }
    setError(undefined)
    save.mutate()
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <form onSubmit={submit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>{existing ? "Edit secret" : "Add secret"}</DialogTitle>
            <DialogDescription>
              Environment: <span className="font-mono">{environment}</span>
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-2">
            <Label htmlFor="secret-key">Key</Label>
            <Input
              id="secret-key"
              className="font-mono"
              placeholder="DATABASE_URL"
              value={key}
              disabled={!!existing}
              onChange={(e) => setKey(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="secret-value">Value</Label>
            <Input
              id="secret-value"
              type="password"
              autoComplete="off"
              placeholder={existing ? "Enter new value" : ""}
              value={value}
              onChange={(e) => setValue(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="secret-rotation">Rotate every (days, 0 = off)</Label>
            <Input
              id="secret-rotation"
              type="number"
              min={0}
              value={rotationDays}
              onChange={(e) => setRotationDays(e.target.value)}
            />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={save.isPending}>
              {save.isPending ? "Saving…" : existing ? "Save" : "Add"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export function DeleteSecretDialog({
  projectId,
  slug,
  environment,
  secretKey,
  onClose,
}: Common & { environment: string; secretKey: string }) {
  const refresh = useRefreshProject(slug)
  const remove = useMutation({
    mutationFn: () => deleteSecret(projectId, environment, secretKey),
    onSuccess: async () => {
      await refresh()
      onClose()
    },
  })

  return (
    <AlertDialog open onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete {secretKey}?</AlertDialogTitle>
          <AlertDialogDescription>
            This removes the key from {environment}. Values already loaded onto machines are not
            affected. Rotate them at the provider if they need to change.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <Button variant="destructive" disabled={remove.isPending} onClick={() => remove.mutate()}>
            {remove.isPending ? "Deleting…" : "Delete"}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

export function AddEnvironmentDialog({ projectId, slug, onClose }: Common) {
  const refresh = useRefreshProject(slug)
  const [name, setName] = useState("")
  const [error, setError] = useState<string>()

  const add = useMutation({
    mutationFn: () => addEnvironment(projectId, name),
    onSuccess: async () => {
      await refresh()
      onClose()
    },
    onError: () => setError("Could not add the environment. It may already exist."),
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    const normalized = name.trim().toLowerCase()
    if (!ENV_RE.test(normalized)) {
      setError("Use lowercase letters, numbers and hyphens.")
      return
    }
    setName(normalized)
    setError(undefined)
    add.mutate()
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <form onSubmit={submit} className="space-y-4">
          <DialogHeader>
            <DialogTitle>Add environment</DialogTitle>
            <DialogDescription>Example: dev, staging, prod</DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="env-name">Name</Label>
            <Input
              id="env-name"
              className="font-mono"
              placeholder="staging"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={add.isPending}>
              {add.isPending ? "Adding…" : "Add"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
export function DeleteEnvironmentDialog({
  projectId,
  slug,
  environment,
  secretCount,
  onClose,
}: Common & { environment: string; secretCount: number }) {
  const qc = useQueryClient()
  const remove = useMutation({
    mutationFn: () => removeEnvironment(projectId, environment),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["project", slug] })
      await qc.invalidateQueries({ queryKey: queryKeys.projects })
      onClose()
    },
  })

  return (
    <ConfirmDialog
      title={`Delete environment ${environment}?`}
      description={
        secretCount > 0
          ? `This also deletes ${secretCount} secret(s) in ${environment}. This can't be undone.`
          : `${environment} has no secrets. It will be removed from the project.`
      }
      pending={remove.isPending}
      error={remove.isError ? "Could not delete the environment." : undefined}
      onConfirm={() => remove.mutate()}
      onClose={onClose}
    />
  )
}

export function DeleteProjectDialog({
  projectId,
  name,
  onClose,
}: { projectId: number; name: string; onClose: () => void }) {
  const qc = useQueryClient()
  const navigate = useNavigate()
  const remove = useMutation({
    mutationFn: () => deleteProject(projectId),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: queryKeys.projects })
      onClose()
      navigate({ to: "/dashboard" }) // change to your dashboard route
    },
  })

  return (
    <ConfirmDialog
      title={`Delete ${name}?`}
      description="This deletes the project, its secrets, and its member access. Rotate any values already loaded onto machines."
      pending={remove.isPending}
      error={remove.isError ? "Could not delete the project." : undefined}
      onConfirm={() => remove.mutate()}
      onClose={onClose}
    />
  )
}
