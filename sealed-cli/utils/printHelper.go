package utils

import "fmt"

func PrintHelp() {
	fmt.Println(`sealed - secrets management for developers

Usage:
  sealed <command> [arguments]

Commands:
  login                    Authenticate via OAuth device flow
  load <project> <env>     Load secrets for a project/environment use eval to dump it directly
  run <project> <env> --   Run a command with secrets injected

Run 'sealed <command> --help' for more information on a command.`)
}

func PrintCommandHelp(cmd string) {
	switch cmd {
	case "login":
		fmt.Println(`sealed login - Authenticate via OAuth device flow

Usage:
  sealed login

Opens a browser to authorize this device. The Ed25519 private key is stored
in your OS keychain (macOS Keychain, Windows Credential Manager, or
libsecret on Linux). No tokens are written to disk.`)
	case "load":
		fmt.Println(`sealed load - Load secrets for a project/environment

Usage:
  sealed load <project> <env> [KEY...]

Arguments:
  project    Project name or ID
  env        Environment: dev, staging, or prod
  KEY...     Optional specific keys to load (default: all)

Flags:
  --allow-missing    Exit 0 even if some keys are missing (for CI)
  --format=FORMAT    Output format: export (default), json, dotenv

Output:
  Prints 'export KEY='value'' lines (POSIX shell compatible).
  Values with newlines work; single quotes escaped as '\''.
  NUL bytes are rejected at write time.`)
	case "run":
		fmt.Println(`sealed run - Run a command with secrets injected into environment

Usage:
  sealed run <project> <env> -- <command> [args...]

Arguments:
  project    Project name or ID
  env        Environment: dev, staging, or prod
  --         Separator: everything after is the command to run

This is the preferred way to use secrets - they are injected into the
child process environment and never printed to stdout/stderr.

Example:
  sealed run myapp prod -- npm start
  sealed run myapp staging -- ./deploy.sh`)
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		PrintHelp()
	}
}
