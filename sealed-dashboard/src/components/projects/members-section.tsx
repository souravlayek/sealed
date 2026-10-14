// src/components/projects/members-section.tsx
import { useState, type FormEvent } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { isAxiosError } from "axios"
import { CheckCircle2 } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { inviteMember } from "@/service/project_service"

export type Member = {
  user_id: number
  name: string
  email: string
  role: "PROJECT_ADMIN" | "PROJECT_MEMBER"
  joined_at?: string
}

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function MembersSection({
  projectId,
  slug,
  myRole,
  members,
}: {
  projectId: number
  slug: string
  myRole: "PROJECT_ADMIN" | "PROJECT_MEMBER"
  members: Member[] | undefined
}) {
  const isAdmin = myRole === "PROJECT_ADMIN"

  return (
    <div className="space-y-6">
      {isAdmin && <InviteForm projectId={projectId} slug={slug} />}
      <MembersTable members={members ?? []} />
    </div>
  )
}

function InviteForm({ projectId }: { projectId: number; slug: string }) {
  const [email, setEmail] = useState("")
  const [error, setError] = useState<string>()
  const [sentTo, setSentTo] = useState<string>()

  const invite = useMutation({
    mutationFn: (value: string) => inviteMember(projectId, { email: value }),
    onSuccess: (_data, value) => {
      setSentTo(value)
      setEmail("")
    },
    onError: (err) => {
      setSentTo(undefined)
      if (isAxiosError(err)) {
        if (err.response?.status === 404) {
          return setError("Project not found, or you don't have access to invite people.")
        }
        const msg = err.response?.data?.message as string | undefined
        return setError(msg ?? "Couldn't send the invite. Try again.")
      }
      setError("Couldn't send the invite. Try again.")
    },
  })

  const submit = (e: FormEvent) => {
    e.preventDefault()
    const normalized = email.trim().toLowerCase()
    if (!EMAIL_RE.test(normalized)) {
      setSentTo(undefined)
      return setError("Enter a valid email address.")
    }
    setError(undefined)
    setSentTo(undefined)
    invite.mutate(normalized)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Invite a collaborator</CardTitle>
        <CardDescription>
          They join as a member. Invites are sent to their email address.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        {sentTo && (
          <Alert>
            <CheckCircle2 className="size-4" />
            <AlertDescription>Invite sent to {sentTo}.</AlertDescription>
          </Alert>
        )}

        <form onSubmit={submit} className="flex flex-col gap-2 sm:flex-row sm:items-end">
          <div className="flex-1 space-y-2">
            <Label htmlFor="invite-email">Email</Label>
            <Input
              id="invite-email"
              type="email"
              placeholder="teammate@company.com"
              autoComplete="off"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <Button type="submit" disabled={invite.isPending}>
            {invite.isPending ? "Sending…" : "Invite"}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}

function MembersTable({ members }: { members: Member[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Collaborators</CardTitle>
        <CardDescription>
          {members.length} {members.length === 1 ? "person" : "people"} with access
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Member</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Joined</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {members.length === 0 && (
              <TableRow>
                <TableCell colSpan={3} className="text-center text-muted-foreground">
                  No collaborators to show yet.
                </TableCell>
              </TableRow>
            )}
            {members.map((m) => (
              <TableRow key={m.user_id}>
                <TableCell>
                  <div className="font-medium">{m.name}</div>
                  <div className="text-xs text-muted-foreground">{m.email}</div>
                </TableCell>
                <TableCell>
                  <Badge variant={m.role === "PROJECT_ADMIN" ? "default" : "outline"}>
                    {m.role === "PROJECT_ADMIN" ? "Admin" : "Member"}
                  </Badge>
                </TableCell>
                <TableCell className="text-muted-foreground">
                  {m.joined_at ? new Date(m.joined_at).toLocaleDateString() : "—"}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
