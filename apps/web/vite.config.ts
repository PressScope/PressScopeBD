import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  server: {
    port: 3001,
  },
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [
    tailwindcss(),
    tanstackRouter({
      target: "react",
      autoCodeSplitting: true,
    }),
    react(),
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
