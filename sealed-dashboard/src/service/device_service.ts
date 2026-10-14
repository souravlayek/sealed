// src/service/device_service.ts
import api from "."

export type DeviceInfo = {
  id: number
  device_name: string
  device_code: string
  created_at: string
  last_used_at: string | null
  revoked_at: string | null
}

export type ValidateDeviceInput = {
  device_code: string
  user_code: string
}

export const validateDevice = (body: ValidateDeviceInput) =>
  api
    .post<ServerResponse<string>>("/device/validate", body)
    .then((res) => res.data.data)
// Not in the API yet. Paths are placeholders until the backend adds them.
export const getDevices = () =>
  api.get<ServerResponse<DeviceInfo[]>>("/device/").then((res) => res.data.data)

export const revokeDevice = (id: number) =>
  api.post(`/device/${id}/revoke`)
