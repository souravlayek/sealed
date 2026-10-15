// src/components/projects/cli-commands-modal.tsx
import { Link } from "@tanstack/react-router"
import { BookOpen, Copy } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { buildCommand } from "@/components/docs/command-builder"

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  projectSlug: string
  environment: string
}

export function CliCommandsModal({ open, onOpenChange, projectSlug, environment }: Props) {
  const load = buildCommand("load", {
    project: projectSlug,
    environment,
    keys: [],
    allowMissing: false,
    shell: "zsh",
    runCommand: "",
  })

  const run = buildCommand("run", {
    project: projectSlug,
    environment,
    keys: [],
    allowMissing: false,
    shell: "zsh",
    runCommand: "npm start",
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>CLI commands</DialogTitle>
          <DialogDescription>
            <span className="font-mono">{projectSlug}</span> ·{" "}
            <span className="font-mono">{environment}</span>
          </DialogDescription>
          <Link
            to="/docs"
            onClick={() => onOpenChange(false)}
            className="inline-flex w-fit items-center gap-1.5 text-xs text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
          >
            <BookOpen className="size-3.5" />
            Not installed yet? See install steps in the docs
          </Link>
        </DialogHeader>

        <CommandBlock
          label="Load into this shell"
          hint="Sets the variables in your current terminal until you close it."
          command={load}
        />
        <CommandBlock
          label="Run a single command"
          hint="Injects the secrets into one process without printing them. Replace npm start with your command."
          command={run}
        />
      </DialogContent>
    </Dialog>
  )
}

function CommandBlock({
  label,
  hint,
  command,
}: {
  label: string
  hint: string
  command: string
}) {
  return (
    <div className="space-y-2">
      <div>
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{hint}</p>
      </div>
      <div className="flex items-start gap-2">
        <pre className="flex-1 overflow-x-auto rounded-md bg-muted p-3 font-mono text-sm">
          <code>{command}</code>
        </pre>
        <Button
          variant="outline"
          size="icon"
          aria-label={`Copy ${label}`}
          onClick={() => navigator.clipboard.writeText(command)}
        >
          <Copy />
        </Button>
      </div>
    </div>
  )
}
