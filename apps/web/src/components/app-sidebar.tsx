import * as React from "react"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "@PressScopeBd/ui/components/sidebar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@PressScopeBd/ui/components/dropdown-menu"
import {
  Home,
  Settings,
  Users,
  FileText,
  Calendar,
  BarChart3,
  HelpCircle,
  ChevronUp
} from "lucide-react"

export function AppSidebar() {
  const [activeItem, setActiveItem] = React.useState("Home")

  const mainNavItems = [
    { id: "Home", label: "Home", icon: Home },
    { id: "Dashboard", label: "Dashboard", icon: BarChart3 },
    { id: "Projects", label: "Projects", icon: FileText },
    { id: "Team", label: "Team", icon: Users },
    { id: "Calendar", label: "Calendar", icon: Calendar },
  ]

  const settingsItems = [
    { id: "Settings", label: "Settings", icon: Settings },
    { id: "Help", label: "Help & Support", icon: HelpCircle },
  ]

  return (
    <Sidebar>
      <SidebarHeader>
        <SidebarGroup>
          <SidebarGroupContent className="flex flex-row items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                <FileText className="size-4" />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">PressScopeBd</span>
                <span className="truncate text-xs text-muted-foreground">
                  Enterprise
                </span>
              </div>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-9 w-9">
                  <span className="sr-only">Toggle user menu</span>
                  <div className="flex h-6 w-6 items-center justify-center rounded-full bg-sidebar-primary text-sidebar-primary-foreground text-xs font-medium">
                    JD
                  </div>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <div className="px-2 py-1.5 text-sm font-semibold">My Account</div>
                <DropdownMenuSeparator />
                <DropdownMenuItem>Profile</DropdownMenuItem>
                <DropdownMenuItem>Billing</DropdownMenuItem>
                <DropdownMenuItem>Settings</DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem>Log out</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainNavItems.map((item) => {
                const Icon = item.icon
                return (
                  <SidebarMenuItem key={item.id}>
                    <SidebarMenuButton
                      isActive={activeItem === item.id}
                      onClick={() => setActiveItem(item.id)}
                      tooltip={item.label}
                    >
                      <Icon />
                      <span>{item.label}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel className="flex items-center justify-between">
            <span>Settings</span>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button className="flex items-center justify-center rounded-sm hover:bg-sidebar-accent h-5 w-5">
                  <ChevronUp className="h-3 w-3" />
                  <span className="sr-only">Toggle settings</span>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {settingsItems.map((item) => (
                  <DropdownMenuItem key={item.id}>
                    <item.icon className="mr-2 h-4 w-4" />
                    <span>{item.label}</span>
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {settingsItems.map((item) => {
                const Icon = item.icon
                return (
                  <SidebarMenuItem key={item.id}>
                    <SidebarMenuButton
                      isActive={activeItem === item.id}
                      onClick={() => setActiveItem(item.id)}
                      tooltip={item.label}
                    >
                      <Icon />
                      <span>{item.label}</span>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarGroup>
          <SidebarGroupContent>
            <div className="px-2 py-2">
              <div className="flex items-center gap-2 px-2 py-1.5 text-xs text-muted-foreground">
                <span>© 2026 PressScopeBd</span>
                <span className="ml-auto">v1.0.0</span>
              </div>
            </div>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarFooter>
    </Sidebar>
  )
}
