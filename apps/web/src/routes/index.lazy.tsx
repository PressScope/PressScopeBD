import { createLazyFileRoute, redirect } from "@tanstack/react-router";
import { useHotkey } from "@tanstack/react-hotkeys";

export const Route = createLazyFileRoute("/")({
  component: RouteComponent,
});

function RouteComponent() {
  const { session } = Route.useRouteContext();
  useHotkey("Mod+K", () => {
    console.log("Redirecting to dashboard");
  });
  return (
    <div>
      <h1>Dashboard</h1>
      <p>Welcome {session.data?.user.name}</p>
    </div>
  );
}
