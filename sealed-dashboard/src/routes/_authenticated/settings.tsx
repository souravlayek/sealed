// src/routes/_authenticated/settings.tsx
import { createFileRoute } from "@tanstack/react-router"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { DevicesSection } from "@/components/settings/devices-section"

export const Route = createFileRoute("/_authenticated/settings")({
  component: SettingsPage,
})

function SettingsPage() {
  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <div>
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="text-sm text-muted-foreground">Manage your devices and account.</p>
      </div>

      <DevicesSection />

      <Card className="opacity-60">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            Security <Badge variant="secondary">WIP</Badge>
          </CardTitle>
          <CardDescription>Audit log and sessions are coming later.</CardDescription>
        </CardHeader>
        <CardContent />
      </Card>
    </div>
  )
}
