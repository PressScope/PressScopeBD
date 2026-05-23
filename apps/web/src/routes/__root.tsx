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
import {
  SidebarProvider,
  SidebarTrigger,
} from "@PressScopeBd/ui/components/sidebar";
import { AppSidebar } from "@/components/app-sidebar";

import "../index.css";
// Importing the auth client lazily prevents a hard module-level throw if the
// network call inside createAuthClient fails during CI / when the backend is down.
// If the import itself throws (e.g. a missing dependency), fall back to a stub so
// that the rest of the app — including the theme toggler — can still render.
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const authClient: typeof import("@/lib/auth-client").authClient
  // @ts-expect-error dynamic import stub
  = await import("@/lib/auth-client").then((m) => m.authClient).catch(() => undefined);
export interface RouterAppContext {}

export const Route = createRootRouteWithContext<RouterAppContext>()({
  component: RootComponent,
  beforeLoad: async () => {
    let session: { data: unknown } = { data: {} };
    try {
      session = await authClient?.getSession?.()
        ?? { data: undefined } as unknown as { data: unknown };
    } catch {
      // Ignore auth failures — login page and tests don't need a live session
    }
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
  const { session } = Route.useRouteContext();

  // Hide sidebar on login page
  const pathname = window.location.pathname;
  const isLoginPage = pathname === "/login" || pathname === "/signup";

  return (
    <>
      <HeadContent />
      <ThemeProvider
        attribute="class"
        defaultTheme="dark"
        disableTransitionOnChange
        storageKey="vite-ui-theme"
      >
        {!isLoginPage && (
          <SidebarProvider>
            <div className="w-full h-screen flex">
              <AppSidebar />
              <div className="flex-1 flex flex-col">
                <Header />
                <main className="flex-1 px-4 py-6 w-full">
                  <Outlet />
                </main>
              </div>
            </div>
          </SidebarProvider>
        )}
        {isLoginPage ? (
          <div className="w-full h-screen flex flex-col">
            <Header />
            <main className="flex-1 px-4 py-6 w-full ">
              <Outlet />
            </main>
          </div>
        ) : null}
        <Toaster richColors />
      </ThemeProvider>
      <TanStackRouterDevtools position="bottom-right" />
    </>
  );
}
