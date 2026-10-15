// src/components/docs/docs-page.tsx
import { useEffect, useMemo, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { Copy, Check } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import { Badge } from "@/components/ui/badge"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { getAllProjects } from "@/service/project_service"
import { queryKeys } from "@/lib/query-keys"
import { buildCommand, type CommandId } from "./command-builder"
import { InstallSection } from "./install-section"

const COMMANDS: Record<CommandId, { title: string; summary: string; detail: string }> = {
  login: {
    title: "sealed login",
    summary: "Authorize this machine.",
    detail:
      "Starts the OAuth device flow. Approve the device in your browser and the CLI stores an Ed25519 private key in your OS keychain. No tokens are written to disk.",
  },
  load: {
    title: "sealed load",
    summary: "Print export lines for a project's environment.",
    detail:
      "Prints one `export KEY='value'` line per secret, single-quoted with POSIX escaping. eval runs those lines in your current shell. Exit code 3 means keys are missing; --allow-missing treats that as success.",
  },
  run: {
    title: "sealed run",
    summary: "Run a command with secrets injected.",
    detail:
      "Starts the command with the secrets in its environment and never prints them. Use this when the variables don't need to live in your shell.",
  },
}

export function DocsPage() {
  const { data: projects = [], isLoading } = useQuery({
    queryKey: queryKeys.projects,
    queryFn: getAllProjects,
  })

  const [active, setActive] = useState<CommandId>("load")
  const [projectSlug, setProjectSlug] = useState<string>("")
  const [environment, setEnvironment] = useState<string>("")
  const [keysText, setKeysText] = useState("")
  const [allowMissing, setAllowMissing] = useState(false)
  const [shell, setShell] = useState<"bash" | "zsh" | "fish">("zsh")
  const [runCommand, setRunCommand] = useState("npm start")
  const [copied, setCopied] = useState(false)

  // Default to the first project once the list loads
  useEffect(() => {
    if (!projectSlug && projects.length > 0) setProjectSlug(projects[0].slug)
  }, [projects, projectSlug])

  const selectedProject = useMemo(
    () => projects.find((p) => p.slug === projectSlug),
    [projects, projectSlug]
  )

  // Reset the environment whenever the project changes
  useEffect(() => {
    const envs = selectedProject?.environments ?? []
    if (!envs.includes(environment)) setEnvironment(envs[0] ?? "")
  }, [selectedProject, environment])

  const keys = keysText.split(/[\s,]+/).filter(Boolean)
  const command = buildCommand(active, {
    project: projectSlug,
    environment,
    keys,
    allowMissing,
    shell,
    runCommand,
  })

  const copy = async () => {
    await navigator.clipboard.writeText(command)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6 p-6">
      <div>
        <h1 className="text-2xl font-semibold">Docs</h1>
        <p className="text-sm text-muted-foreground">
          Pick a project and environment, then copy the command.
        </p>
      </div>
<InstallSection />
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Target</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>Project</Label>
            <Select
              value={projectSlug}
              onValueChange={(v) => setProjectSlug(v ?? "")}
              disabled={isLoading || projects.length === 0}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder={isLoading ? "Loading…" : "No projects"} />
              </SelectTrigger>
              <SelectContent>
                {projects.map((p) => (
                  <SelectItem key={p.id} value={p.slug}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Environment</Label>
            <Select
              value={environment}
              onValueChange={(v) => setEnvironment(v ?? "")}
              disabled={!selectedProject || selectedProject.environments.length === 0}
            >
              <SelectTrigger className="w-full">
                <SelectValue placeholder="No environments" />
              </SelectTrigger>
              <SelectContent>
                {(selectedProject?.environments ?? []).map((env) => (
                  <SelectItem key={env} value={env}>
                    {env}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      <Tabs value={active} onValueChange={(v) => setActive(v as CommandId)}>
        <TabsList>
          <TabsTrigger value="login">login</TabsTrigger>
          <TabsTrigger value="load">load</TabsTrigger>
          <TabsTrigger value="run">run</TabsTrigger>
        </TabsList>

        {(Object.keys(COMMANDS) as CommandId[]).map((id) => (
          <TabsContent key={id} value={id} className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="font-mono">{COMMANDS[id].title}</CardTitle>
                <CardDescription>{COMMANDS[id].summary}</CardDescription>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                {COMMANDS[id].detail}
              </CardContent>
            </Card>

            {id === "load" && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">Options</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-2 sm:col-span-2">
                    <Label htmlFor="keys">Only these keys (optional)</Label>
                    <Input
                      id="keys"
                      className="font-mono"
                      placeholder="DATABASE_URL STRIPE_KEY"
                      value={keysText}
                      onChange={(e) => setKeysText(e.target.value)}
                    />
                  </div>
                  <div className="flex items-center justify-between sm:col-span-2">
                    <Label htmlFor="allow-missing">Allow missing keys (--allow-missing)</Label>
                    <Switch
                      id="allow-missing"
                      checked={allowMissing}
                      onCheckedChange={setAllowMissing}
                    />
                  </div>
                  <div className="space-y-2 sm:col-span-2">
                    <Label>Shell</Label>
                    <div className="flex gap-2">
                      {(["bash", "zsh", "fish"] as const).map((s) => (
                        <Button
                          key={s}
                          size="sm"
                          variant={shell === s ? "default" : "outline"}
                          onClick={() => setShell(s)}
                        >
                          {s}
                        </Button>
                      ))}
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}

            {id === "run" && (
              <Card>
                <CardContent className="space-y-2 pt-6">
                  <Label htmlFor="run-cmd">Command to run</Label>
                  <Input
                    id="run-cmd"
                    className="font-mono"
                    value={runCommand}
                    onChange={(e) => setRunCommand(e.target.value)}
                  />
                </CardContent>
              </Card>
            )}

            <Card>
              <CardHeader className="flex-row items-center justify-between space-y-0">
                <CardTitle className="text-base">Command</CardTitle>
                <Badge variant="outline">{id === "load" ? `${shell} syntax` : "copy & run"}</Badge>
              </CardHeader>
              <CardContent className="space-y-3">
                <pre className="overflow-x-auto rounded-md bg-muted p-4 font-mono text-sm">
                  <code>{command}</code>
                </pre>
                <Button variant="outline" size="sm" onClick={copy}>
                  {copied ? <Check /> : <Copy />}
                  {copied ? "Copied" : "Copy"}
                </Button>
              </CardContent>
            </Card>
          </TabsContent>
        ))}
      </Tabs>
    </div>
  )
}
