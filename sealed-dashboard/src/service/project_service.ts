// src/service/project_service.ts
import api from "."

type ProjectRole = "PROJECT_ADMIN" | "PROJECT_MEMBER"

export type Project = {
  created_at: string
  environments: Array<string>
  id: number
  name: string
  role: ProjectRole
  slug: string
}

export type SecretMeta = {
  key_name: string
  rotation_days: number | null
  value_changed_at: string | null
  created_at?: string
}

export type ProjectDetails = {
  created_at: string
  environments: Array<string>
  id: number
  my_role: ProjectRole
  name: string
  secrets: Record<string, SecretMeta[]> // environment -> secrets
  slug: string
}

export type SecretInput = {
  environment: string
  key_name: string
  secret: string
  rotation_days: number
}
export type CreateProjectInput = {
  name: string
  environments: string[]
}

export const createProject = (body: CreateProjectInput) =>
  api
    .post<ServerResponse<Project>>("/projects/", body)
    .then((res) => res.data.data)

export const getAllProjects = async () =>
  api.get<ServerResponse<Array<Project>>>("/projects/").then((res) => res.data.data)

export const getProjectDetails = async (slug: string) =>
  api
    .get<ServerResponse<ProjectDetails>>(`/projects/${slug}`)
    .then((res) => res.data.data)

export const createSecret = (projectId: number, body: SecretInput) =>
  api.post(`/projects/${projectId}/secret`, body)

export const deleteSecret = (projectId: number, environment: string, key: string) =>
  api.delete(
    `/projects/${projectId}/secret/${encodeURIComponent(environment)}/${encodeURIComponent(key)}`
  )

export const addEnvironment = (projectId: number, environment: string) =>
  api.put(`/projects/${projectId}/environment`, { environment })

/** Next rotation due date and days remaining, or null when rotation is off. */
export function getRotation(s: SecretMeta) {
  if (!s.rotation_days || s.rotation_days <= 0) return null
  const base = new Date(s.value_changed_at ?? s.created_at ?? Date.now())
  const due = new Date(base.getTime() + s.rotation_days * 864e5)
  const daysLeft = Math.ceil((due.getTime() - Date.now()) / 864e5)
  return { due, daysLeft }
}
