// src/routes/_authenticated/device.tsx
import { createFileRoute } from "@tanstack/react-router"
import { DeviceValidate } from "@/components/device/device-validate"
import { z } from 'zod'
import { DeviceRouteError } from "@/components/device/device-route-error"
const validateSearchParams = z.object({
  device_code: z.string().nonempty(),
})
export const Route = createFileRoute("/_authenticated/device")({
  validateSearch: validateSearchParams,
    errorComponent: DeviceRouteError,
  component: DeviceValidatePage,
})

function DeviceValidatePage() {
  const { device_code } = Route.useSearch()
  return <DeviceValidate initialDeviceCode={device_code} />
}
