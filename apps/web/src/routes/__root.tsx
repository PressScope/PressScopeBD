import { SidebarProvider } from "@PressScopeBd/ui/components/sidebar";
import { Toaster } from "@PressScopeBd/ui/components/sonner";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  HeadContent,
  Outlet,
  createRootRouteWithContext,
  redirect,
  useRouterState,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

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
    if (session.data && isPublicRoute) {
      throw redirect({
        to: "/",
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

const queryClient = new QueryClient();

function RootComponent() {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const isAuthPage = ["/login", "/signup"].includes(pathname);

  return (
    <QueryClientProvider client={queryClient}>
      <HeadContent />

      <ThemeProvider attribute="class" defaultTheme="dark" enableSystem={false}>
        {isAuthPage ? <AuthLayout /> : <DashboardLayout />}

        <Toaster richColors />
      </ThemeProvider>

      <TanStackRouterDevtools position="bottom-right" />
    </QueryClientProvider>
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
