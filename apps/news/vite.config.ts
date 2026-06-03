import { defineConfig } from "vite";
import { devtools } from "@tanstack/devtools-vite";

import { tanstackStart } from "@tanstack/react-start/plugin/vite";

import viteReact from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { nitro } from "nitro/vite";

const config = defineConfig({
  resolve: { tsconfigPaths: true },
  plugins: [
    devtools(),
    nitro({ rollupConfig: { external: [/^@sentry\//] } }),
    tailwindcss(),
    tanstackStart(),
    viteReact(),
  ],
  build: {
    rollupOptions: {
      output: {
        // This splits your node_modules into a separate 'vendor' chunk
        manualChunks(id) {
          if (id.includes("node_modules")) {
            // Puts Tanstack packages in their own chunk
            if (id.includes("@tanstack")) {
              return "vendor-tanstack";
            }

            // Puts React core packages in their own chunk
            if (id.includes("react")) {
              return "vendor-react";
            }

            // Everything else in node_modules goes to standard vendor
            return "vendor";
          }
        },
      },
    },
  },
});

export default config;
