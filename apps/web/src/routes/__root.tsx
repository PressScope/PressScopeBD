import { Toaster } from "@PressScopeBd/ui/components/sonner";
import {
  HeadContent,
  Outlet,
  createRootRouteWithContext,
  redirect,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

import Header from "@/components/header";
import { ThemeProvider } from "@/components/theme-provider";
import { SidebarProvider, SidebarTrigger } from "@PressScopeBd/ui/components/sidebar";
import { AppSidebar } from "@/components/app-sidebar";

import "../index.css";
import { authClient } from "@/lib/auth-client";
export interface RouterAppContext {}

export const Route = createRootRouteWithContext<RouterAppContext>()({
  component: RootComponent,
  beforeLoad: async () => {
    const session = await authClient.getSession();
    if (window.location.pathname === "/login") {
      console.log("Already on login page");
    } else if (!session.data) {
      redirect({
        to: "/login",
        throw: true,
      });
    }
    return { session };
  },
  head: () => ({
    meta: [
      {
        title: "PressScopeBd",
      },
      {
        name: "description",
        content: "PressScopeBd is a web application",
      },
    ],
    links: [
      {
        rel: "icon",
        href: "/favicon.ico",
      },
    ],
  }),
});

function RootComponent() {
  return (
    <>
      <HeadContent />
      <ThemeProvider
        attribute="class"
        defaultTheme="dark"
        disableTransitionOnChange
        storageKey="vite-ui-theme"
      >
        <SidebarProvider>
          <div className="grid grid-cols-[auto_1fr] h-svh">
            <AppSidebar />
            <div className="flex flex-col">
              <Header />
              <main className="flex-1 px-4 py-6">
                <Outlet />
              </main>
            </div>
          </div>
        </SidebarProvider>
        <Toaster richColors />
      </ThemeProvider>
      <TanStackRouterDevtools position="bottom-left" />
    </>
  );
}
