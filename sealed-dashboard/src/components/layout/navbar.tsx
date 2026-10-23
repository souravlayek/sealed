import { Link } from "@tanstack/react-router"
import Logo from "../logo"
import { Button } from "../ui/button"

export const Navbar = () => {
  return (
    <header className="mx-auto flex max-w-5xl items-center justify-between px-6 py-6">
      <Link to="/">
      <div className="flex items-center gap-2 font-semibold tracking-tight">
        <Logo className="size-5" />
        Sealed
        </div>
      </Link>
      <nav className="flex items-center gap-3">
        <Button variant="ghost" size="sm" render={<Link to="/login" search={{redirect: "/dashboard"}} />}>
          Sign in
        </Button>
        <Button size="sm" render={<Link to="/signup" />}>
          Get started
        </Button>
      </nav>
    </header>
  )
}
