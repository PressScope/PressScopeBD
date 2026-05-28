// routes/__root.tsx

import { Toaster } from "@PressScopeBd/ui/components/sonner";
import {
  HeadContent,
  Outlet,
  createRootRouteWithContext,
  redirect,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

import { SidebarProvider } from "@PressScopeBd/ui/components/sidebar";

import { AppSidebar } from "@/components/app-sidebar";
import Header from "@/components/header";
import { ThemeProvider } from "@/components/theme-provider";

import { authClient } from "@/lib/auth-client";

import "../index.css";

type Session = Awaited<ReturnType<typeof authClient.getSession>>;

export interface RouterAppContext {
  session: Session;
}

const publicRoutes = ["/login", "/signup"];

export const Route = createRootRouteWithContext<RouterAppContext>()({
  beforeLoad: async ({ location }) => {
    const session = await authClient.getSession();

    const isPublicRoute = publicRoutes.includes(location.pathname);

    if (!session.data && !isPublicRoute) {
      throw redirect({
        to: "/login",
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
        content: "PressScopeBd dashboard",
      },
    ],
  }),

  component: RootComponent,
});

function RootComponent() {
  const pathname = window.location.pathname;

  const isAuthPage = pathname === "/login";

  return (
    <>
      <HeadContent />

      <ThemeProvider attribute="class" defaultTheme="dark" enableSystem={false}>
        {isAuthPage ? <AuthLayout /> : <DashboardLayout />}

        <Toaster richColors />
      </ThemeProvider>

      <TanStackRouterDevtools position="bottom-right" />
    </>
  );
}

function AuthLayout() {
  return (
    <div className="flex h-screen flex-col">
      <Header />

      <main className="flex-1 px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}

function DashboardLayout() {
  return (
    <SidebarProvider>
      <div className="flex h-screen w-full">
        <AppSidebar />

        <div className="flex flex-1 flex-col">
          <Header />

          <main className="flex-1 px-4 py-6">
            <Outlet />
          </main>
        </div>
      </div>
    </SidebarProvider>
  );
}
