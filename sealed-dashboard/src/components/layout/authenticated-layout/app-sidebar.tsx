// src/components/app-sidebar.tsx
import { useState } from "react"
import { Link } from "@tanstack/react-router"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { ChevronsUpDown, Lock, LogOut, Plus, Settings } from "lucide-react"
import { useAuth } from "@/auth"
import { getAllProjects } from "@/service/project_service"
import { getUserDetails } from "@/service/auth_service"
import { queryKeys } from "@/lib/query-keys"
import { CreateProjectDialog } from "@/components/projects/create-project-dialog"
import { BookOpen } from "lucide-react"

export function AppSidebar() {
  const auth = useAuth()
  const qc = useQueryClient()
  const [creating, setCreating] = useState(false)

  const { data: user } = useQuery({
    queryKey: queryKeys.me,
    queryFn: getUserDetails,
    staleTime: 5 * 60_000,
  })

  const { data: projects = [], isLoading } = useQuery({
    queryKey: queryKeys.projects,
    queryFn: getAllProjects,
  })

  const initials = user?.name
    ? user.name.split(" ").map((p) => p[0]).slice(0, 2).join("").toUpperCase()
    : "?"

  const logout = () => {
    qc.clear() // drop cached data so the next user doesn't see it
    auth.logout()
  }

  return (
    <Sidebar>
      <SidebarHeader>
        <div className="flex items-center gap-3 px-2 py-2">
          <div className="flex size-9 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Lock className="size-4" />
          </div>
          <div className="flex flex-col leading-tight">
            <span className="text-sm font-semibold">Sealed</span>
            <span className="text-xs text-muted-foreground">Secrets manager</span>
          </div>
        </div>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Projects</SidebarGroupLabel>
          <SidebarGroupAction title="New project" onClick={() => setCreating(true)}>
            <Plus />
            <span className="sr-only">New project</span>
          </SidebarGroupAction>
          <SidebarGroupContent>
            <SidebarMenu>
              {isLoading && (
                <SidebarMenuItem>
                  <span className="px-2 text-xs text-muted-foreground">Loading…</span>
                </SidebarMenuItem>
              )}

              {!isLoading && projects.length === 0 && (
                <SidebarMenuItem>
                  <button
                    onClick={() => setCreating(true)}
                    className="flex w-full items-center gap-2 px-2 py-1.5 text-sm text-muted-foreground hover:text-foreground"
                  >
                    <Plus className="size-4" /> Create your first project
                  </button>
                </SidebarMenuItem>
              )}

              {projects.map((project) => (
                <SidebarMenuItem key={project.id}>
                  <SidebarMenuButton
                    render={
                      <Link
                        to="/projects/$projectSlug"
                        params={{ projectSlug: project.slug }}
                      />
                    }
                  >
                    <span>{project.name}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton
                  render={<Link to="/docs" />}
                >
                  <BookOpen />
                  <span>Docs</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <SidebarMenuButton size="lg" className="data-[popup-open]:bg-sidebar-accent" />
                }
              >
                <Avatar className="size-8 rounded-lg">
                  <AvatarFallback className="rounded-lg">{initials}</AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">{user?.name ?? "Loading..."}</span>
                  <span className="truncate text-xs text-muted-foreground">{user?.email ?? ""}</span>
                </div>
                <ChevronsUpDown className="ml-auto size-4" />
              </DropdownMenuTrigger>

              <DropdownMenuContent side="top" align="start" className="min-w-56">
                <DropdownMenuGroup>
                  <DropdownMenuLabel className="font-normal">
                    <div className="truncate text-sm font-medium">{user?.name}</div>
                    <div className="truncate text-xs text-muted-foreground">{user?.email}</div>
                  </DropdownMenuLabel>
                </DropdownMenuGroup>

                <DropdownMenuSeparator />

                <DropdownMenuGroup>
                  <DropdownMenuItem render={<Link to="/settings" />}>
                    <Settings className="mr-2 size-4" />
                    Settings
                  </DropdownMenuItem>
                </DropdownMenuGroup>

                <DropdownMenuSeparator />

                <DropdownMenuGroup>
                  <DropdownMenuItem
                    onClick={logout}
                    className="text-destructive focus:text-destructive"
                  >
                    <LogOut className="mr-2 size-4" />
                    Log out
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>

      {creating && <CreateProjectDialog onClose={() => setCreating(false)} />}
    </Sidebar>
  )
}
