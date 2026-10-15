// src/components/device/device-validate.tsx
import { useState, type FormEvent } from "react"
import { useMutation } from "@tanstack/react-query"
import { isAxiosError } from "axios"
import { CheckCircle2 } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSeparator,
  InputOTPSlot,
} from "@/components/ui/input-otp"
import { validateDevice } from "@/service/device_service"

const CODE_LENGTH = 8

export function DeviceValidate({
  initialDeviceCode,
}: {
  initialDeviceCode: string
  }) {
  const [slots, setSlots] = useState("") // 8 chars, no dash
  const userCode = slots.length > 4 ? `${slots.slice(0, 4)}-${slots.slice(4)}` : slots

  const [localError, setLocalError] = useState<string>()

  const approve = useMutation({
    mutationFn: (code: string) => validateDevice({ device_code: initialDeviceCode, user_code: code }),
  })
  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (!/^[A-Z0-9]{8}$/.test(slots)) {
      return setLocalError("Enter all 8 characters of the code.")
    }
    setLocalError(undefined)
    approve.mutate(userCode) // sends "ABCD-1234"
  }
  const serverError =
    approve.error && isAxiosError(approve.error)
      ? (approve.error.response?.data?.message as string | undefined)
      : approve.error
        ? "Couldn't authorize this device. Try again."
        : undefined

  if (approve.isSuccess) {
    return (
      <Card className="mx-auto max-w-md">
        <CardHeader className="items-center text-center">
          <CheckCircle2 className="size-10 text-primary" />
          <CardTitle>Device authorized</CardTitle>
          <CardDescription>You can return to your terminal.</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <Card className="mx-auto max-w-md">
      <CardHeader>
        <CardTitle>Authorize a device</CardTitle>
        <CardDescription>
          Enter the code shown in your terminal after running <code>sealed login</code>.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {(localError || serverError) && (
          <Alert variant="destructive">
            <AlertDescription>{localError ?? serverError}</AlertDescription>
          </Alert>
        )}

        <form onSubmit={submit} className="space-y-6">

          <div className="space-y-2">
            <Label className="block text-center">User code</Label>
            <div className="flex justify-center">
              <InputOTP
                maxLength={CODE_LENGTH}
                value={slots}
                onChange={(v) => setSlots(v.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, CODE_LENGTH))}
                pasteTransformer={(text) => text.toUpperCase().replace(/[^A-Z0-9]/g, "")}
                autoFocus
              >
                <InputOTPGroup>
                  <InputOTPSlot index={0} />
                  <InputOTPSlot index={1} />
                  <InputOTPSlot index={2} />
                  <InputOTPSlot index={3} />
                </InputOTPGroup>
                <InputOTPSeparator />
                <InputOTPGroup>
                  <InputOTPSlot index={4} />
                  <InputOTPSlot index={5} />
                  <InputOTPSlot index={6} />
                  <InputOTPSlot index={7} />
                </InputOTPGroup>
              </InputOTP>
            </div>
          </div>

          <Button type="submit" className="w-full" disabled={approve.isPending}>
            {approve.isPending ? "Authorizing…" : "Authorize device"}
          </Button>
        </form>

        <p className="text-xs text-muted-foreground">
          Only approve codes you started yourself. Anyone with these codes can attach a device to
          your account.
        </p>
      </CardContent>
    </Card>
  )
}
