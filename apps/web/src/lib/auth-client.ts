import { env } from "@PressScopeBd/env/web";
import { createAuthClient } from "better-auth/react";
import { adminClient } from "better-auth/client/plugins";
export const authClient = createAuthClient({
  baseURL: env.VITE_SERVER_URL,
  plugins: [adminClient()],
});
