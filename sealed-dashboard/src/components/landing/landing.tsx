// src/components/landing/landing.tsx
import { Link } from "@tanstack/react-router"
import { Terminal, ShieldCheck, Ban } from "lucide-react"
import { Button } from "@/components/ui/button"
import Logo from "../logo"
import { Navbar } from "../layout/navbar"

const features = [
  {
    icon: Terminal,
    title: "Load into your shell",
    body: "One command puts your project's secrets into the terminal you're already using.",
  },
  {
    icon: ShieldCheck,
    title: "Signed, not shared",
    body: "Each machine signs its own requests with a key that never leaves the keychain. No tokens to steal.",
  },
  {
    icon: Ban,
    title: "Revoke in one click",
    body: "Lost a laptop? Revoke its device and it loses access on the next request.",
  },
]

export function Landing() {
  return (
    <div className="min-h-svh bg-background text-foreground">
<Navbar />

      <main className="mx-auto max-w-5xl px-6">
        <section className="flex flex-col items-center py-24 text-center sm:py-36">
          <h1 className="text-5xl font-semibold tracking-tight sm:text-7xl">
            Secrets, sealed.
          </h1>
          <p className="mt-6 max-w-xl text-lg text-muted-foreground sm:text-xl">
            Keep your environment variables out of dotfiles and chat threads.
            Load them where you work, and cut off a lost machine in one step.
          </p>
          <div className="mt-10 flex gap-3">
            <Button size="lg" render={<Link to="/signup" />}>
              Get started
            </Button>
            <Button size="lg" variant="outline" render={<Link to="/docs" />}>
              Read the docs
            </Button>
          </div>
          <code className="mt-14 rounded-full border bg-muted px-4 py-2 font-mono text-sm text-muted-foreground">
            eval "$(sealed load payments-api prod)"
          </code>
        </section>

        <section className="grid gap-12 border-t py-24 sm:grid-cols-3">
          {features.map(({ icon: Icon, title, body }) => (
            <div key={title} className="space-y-3">
              <Icon className="size-5" />
              <h2 className="font-medium">{title}</h2>
              <p className="text-sm leading-relaxed text-muted-foreground">{body}</p>
            </div>
          ))}
        </section>

        <section className="flex flex-col items-center gap-6 border-t py-24 text-center">
          <h2 className="text-3xl font-semibold tracking-tight sm:text-4xl">
            Start with one project.
          </h2>
          <Button size="lg" render={<Link to="/signup" />}>
            Get started
          </Button>
        </section>
      </main>

      <footer className="mx-auto flex max-w-5xl items-center justify-between border-t px-6 py-8 text-xs text-muted-foreground">
        <span>© {new Date().getFullYear()} Sealed</span>
        <Link to="/docs" className="hover:text-foreground">Docs</Link>
      </footer>
    </div>
  )
}
