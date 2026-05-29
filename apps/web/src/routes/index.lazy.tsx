// routes/index.lazy.tsx

import { createLazyFileRoute } from "@tanstack/react-router";

export const Route = createLazyFileRoute("/")({
  component: HomePage,
});

function HomePage() {
  const { session } = Route.useRouteContext();

  return (
    <div className="space-y-2">
      <h1 className="text-3xl font-bold">Dashboard</h1>

      <p className="text-muted-foreground">
        Welcome back, {session.data?.user?.role}
      </p>
    </div>
  );
}
