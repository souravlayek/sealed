// src/components/landing/landing-config.ts
export const REPO_URL = "https://github.com/souravlayek/sealed"
export const CLI_REPO_URL = "https://github.com/souravlayek/sealed-cli"
export const INSTALL_DOCS_URL = `${CLI_REPO_URL}#install`
export const DEPLOY_DOCS_URL = `${REPO_URL}/tree/main/deploy`

// Homebrew first, Go toolchain as fallback
export const INSTALL_COMMAND = "brew install --cask souravlayek/tap/sealed"
export const INSTALL_GO_COMMAND =
  "go install github.com/souravlayek/sealed-cli@latest"

export const FIRST_RUN = [
  "sealed login",
  `eval "$(sealed load payments-api dev)"`,
  "sealed run payments-api dev -- npm start",
]
