// src/components/auth/auth-shell.tsx
import type { PropsWithChildren } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import Logo from "../logo";
import { Link } from "@tanstack/react-router";

export function AuthShell({
  title,
  description,
  children,
}: PropsWithChildren<{ title: string; description: string }>) {
  return (
    <div className="flex min-h-svh items-center justify-center bg-muted/40 p-4">
      <div className="w-full max-w-sm space-y-6">
        <div className="flex flex-col items-center gap-3 text-center">
          <Link to="/">
          <div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary-foreground">
            <Logo className="size-6" />
            </div>
          </Link>
          <div>
            <h1 className="text-xl font-semibold tracking-tight">Sealed</h1>
          </div>
        </div>
        <Card>
          <CardHeader className="space-y-1">
            <CardTitle className="text-2xl">{title}</CardTitle>
            <CardDescription>{description}</CardDescription>
          </CardHeader>
          <CardContent>{children}</CardContent>
        </Card>
      </div>
    </div>
  )
}
