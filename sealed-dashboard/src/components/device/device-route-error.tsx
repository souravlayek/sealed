// src/components/device/device-route-error.tsx
import { Link, useRouter } from "@tanstack/react-router"
import { AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export function DeviceRouteError({ error  }: { error: Error;  }) {
  const router = useRouter()
  const isMissingCode = error.name === "ZodError" || /device_code/i.test(error.message)

  return (
    <div className="flex min-h-[60svh] items-center justify-center p-6">
      <Card className="w-full max-w-md">
        <CardHeader className="items-center text-center">
          <AlertTriangle className="size-10 text-destructive" />
          <CardTitle>
            {isMissingCode ? "This link is incomplete" : "Something went wrong"}
          </CardTitle>
          <CardDescription>
            {isMissingCode
              ? "The device code is missing from the link. Run sealed login in your terminal and open the link it prints."
              : "We couldn't load this page. Try again, or go back to your projects."}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center gap-2">
          <Button render={<Link to="/dashboard" />}>Go to dashboard</Button>
        </CardContent>
      </Card>
    </div>
  )
}
