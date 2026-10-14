// src/components/docs/command-builder.ts
export type CommandId = "login" | "load" | "run"

export type BuilderInput = {
  project: string
  environment: string
  keys: string[]
  allowMissing: boolean
  shell: "bash" | "zsh" | "fish"
  runCommand: string
}

export function buildCommand(id: CommandId, input: BuilderInput): string {
  const project = input.project || "<project>"
  const env = input.environment || "<env>"

  switch (id) {
    case "login":
      return "sealed login"

    case "load": {
      const keys = input.keys.filter(Boolean).map((k) => ` ${k}`).join("")
      const flags = input.allowMissing ? " --allow-missing" : ""
      const inner = `sealed load ${project} ${env}${keys}${flags}`
      // POSIX shells: eval "$(...)". Fish uses a pipe into source instead.
      return input.shell === "fish"
        ? `${inner} | source`
        : `eval "$(${inner})"`
    }

    case "run": {
      const project = input.project || "<project>"
      const env = input.environment || "<env>"
      return `sealed run ${project} ${env} -- ${input.runCommand || "<command>"}`
    }
  }
}
