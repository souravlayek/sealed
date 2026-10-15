// src/components/docs/install-section.tsx
import { Copy, Check } from "lucide-react"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  INSTALL_COMMAND,
  INSTALL_GO_COMMAND,
} from "@/components/landing/landing-config"

const QUARANTINE_COMMAND = `xattr -d com.apple.quarantine "$(which sealed)"`

export function InstallSection() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Install the CLI</CardTitle>
        <CardDescription>
          Install once per machine. Then sign in and load or run your secrets.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <Step
          number={1}
          title="Install with Homebrew"
          badge="macOS"
          hint="Recommended on macOS."
          command={INSTALL_COMMAND}
        />

        <Step
          number={2}
          title="Or install with Go"
          badge="Any platform"
          hint="Needs Go 1.22 or later. Installs to $(go env GOPATH)/bin."
          command={INSTALL_GO_COMMAND}
        />

        <Step
          number={3}
          title="Allow the first run on macOS"
          badge="macOS only"
          hint="macOS may block an unsigned binary. Run this once after installing or upgrading."
          command={QUARANTINE_COMMAND}
        />

        <Step
          number={4}
          title="Check the install"
          hint="Prints the version, commit, and build date."
          command="sealed version"
        />
      </CardContent>
    </Card>
  )
}

function Step({
  number,
  title,
  badge,
  hint,
  command,
}: {
  number: number
  title: string
  badge?: string
  hint: string
  command: string
}) {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <span className="flex size-6 items-center justify-center rounded-full bg-muted text-xs font-medium">
          {number}
        </span>
        <p className="text-sm font-medium">{title}</p>
        {badge && (
          <Badge variant="outline" className="ml-auto">
            {badge}
          </Badge>
        )}
      </div>
      <p className="text-xs text-muted-foreground">{hint}</p>
      <div className="flex items-start gap-2">
        <pre className="flex-1 overflow-x-auto rounded-md bg-muted p-3 font-mono text-sm">
          <code>{command}</code>
        </pre>
        <Button variant="outline" size="icon" aria-label={`Copy ${title}`} onClick={copy}>
          {copied ? <Check /> : <Copy />}
        </Button>
      </div>
    </div>
  )
}
