// src/components/settings/devices-section.tsx
import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { ConfirmDialog } from "@/components/confirm-dialog"
import { getDevices, revokeDevice, type DeviceInfo } from "@/service/device_service"
import { queryKeys } from "@/lib/query-keys"

const fmt = (iso: string | null) => (iso ? new Date(iso).toLocaleString() : "Never")

export function DevicesSection() {
  const qc = useQueryClient()
  const [target, setTarget] = useState<DeviceInfo | null>(null)

  const { data: devices = [], isLoading, isError } = useQuery({
    queryKey: queryKeys.devices,
    queryFn: getDevices,
  })

  const revoke = useMutation({
    mutationFn: (id: number) => revokeDevice(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: queryKeys.devices })
      setTarget(null)
    },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Devices</CardTitle>
        <CardDescription>
          Machines signed in with the CLI. Revoking one cuts its access on the next request.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isError && (
          <Alert>
            <AlertTitle>Devices aren't available yet</AlertTitle>
            <AlertDescription>The devices endpoint isn't implemented on the server.</AlertDescription>
          </Alert>
        )}

        {!isError && (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Last used</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground">Loading…</TableCell>
                </TableRow>
              )}
              {!isLoading && devices.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground">
                    No devices yet. Run <code>sealed login</code> on a machine.
                  </TableCell>
                </TableRow>
              )}
              {devices.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-medium">{d.device_name}</TableCell>
                  <TableCell className="text-muted-foreground">{fmt(d.created_at)}</TableCell>
                  <TableCell className="text-muted-foreground">{fmt(d.last_used_at)}</TableCell>
                  <TableCell>
                    {d.revoked_at ? (
                      <Badge variant="destructive">Revoked</Badge>
                    ) : (
                      <Badge variant="outline">Active</Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="ghost"
                      size="sm"
                      disabled={!!d.revoked_at}
                      onClick={() => setTarget(d)}
                    >
                      Revoke
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}

        {target && (
          <ConfirmDialog
            title={`Revoke ${target.device_name}?`}
            description="This machine loses access on its next request. It can sign in again with sealed login."
            confirmLabel="Revoke"
            pending={revoke.isPending}
            onConfirm={() => revoke.mutate(target.id)}
            onClose={() => setTarget(null)}
          />
        )}
      </CardContent>
    </Card>
  )
}
