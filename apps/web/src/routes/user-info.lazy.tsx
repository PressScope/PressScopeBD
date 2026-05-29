import {
  Avatar,
  AvatarFallback,
  AvatarImage,
} from "@PressScopeBd/ui/components/avatar";
import { Button } from "@PressScopeBd/ui/components/button";
import { createLazyFileRoute } from "@tanstack/react-router";

export const Route = createLazyFileRoute("/user-info")({
  component: RouteComponent,
});

function RouteComponent() {
  const { session } = Route.useRouteContext();

  const user = session.data?.user;
  const initials =
    user?.name
      ?.split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase() || "U";

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6">
      <h1 className="text-3xl font-bold">Profile</h1>

      <div className="bg-card flex items-center gap-4 rounded-lg border p-6">
        <Avatar className="h-16 w-16">
          <AvatarImage
            src={user?.image || undefined}
            alt={user?.name || "User"}
          />
          <AvatarFallback className="text-lg">{initials}</AvatarFallback>
        </Avatar>
        <div className="flex flex-col gap-1">
          <h2 className="text-xl font-semibold">{user?.name}</h2>
          <p className="text-muted-foreground text-sm">{user?.email}</p>
        </div>
      </div>

      <div className="grid gap-4">
        <div className="bg-card rounded-lg border p-4">
          <h3 className="mb-2 font-medium">Account Information</h3>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Name</span>
              <p>{user?.name}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Email</span>
              <p>{user?.email}</p>
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-2">
          <Button variant="outline">Edit Profile</Button>
        </div>
      </div>
    </div>
  );
}
