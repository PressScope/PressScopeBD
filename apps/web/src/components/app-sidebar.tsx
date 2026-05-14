import * as React from "react";
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
} from "@PressScopeBd/ui/components/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@PressScopeBd/ui/components/dropdown-menu";
import {
  Home,
  Settings,
  Users,
  FileText,
  Calendar,
  BarChart3,
  HelpCircle,
  ChevronUp,
} from "lucide-react";
import { NavUser } from "./nav-user";

export function AppSidebar() {
  const [activeItem, setActiveItem] = React.useState("Home");

  const mainNavItems = [
    { id: "Home", label: "Home", icon: Home },
    { id: "Dashboard", label: "Dashboard", icon: BarChart3 },
    { id: "Projects", label: "Projects", icon: FileText },
    { id: "Team", label: "Team", icon: Users },
    { id: "Calendar", label: "Calendar", icon: Calendar },
  ];

  const settingsItems = [
    { id: "Settings", label: "Settings", icon: Settings },
    { id: "Help", label: "Help & Support", icon: HelpCircle },
  ];

  return (
    <Sidebar collapsible="icon">
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
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {mainNavItems.map((item) => {
                const Icon = item.icon;
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
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel className="flex items-center justify-between">
            <span>Settings</span>
            <DropdownMenu>
              <DropdownMenuTrigger>
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
                const Icon = item.icon;
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
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  );
}
