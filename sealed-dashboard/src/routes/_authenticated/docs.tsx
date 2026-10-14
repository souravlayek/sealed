import { DocsPage } from '@/components/docs/docs-page'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/docs')({
  component: DocsPage,
})
