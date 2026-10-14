import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AppSidebar } from "./app-sidebar"

export const AuthenticatedLayout: React.FC<React.PropsWithChildren> = ({ children }) => {
  return (<SidebarProvider>
    <AppSidebar />
        <main className="flex-1 min-w-0">
          <SidebarTrigger />
          {children}
        </main>
      </SidebarProvider>)
}
