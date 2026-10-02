import { defineConfig } from "astro/config";

// Mojang APIs send no CORS headers, so the browser calls them same-origin.
// Production: vercel.json rewrites. `astro dev`: this proxy mirrors them.
const mojang = (target) => ({
  target,
  changeOrigin: true,
  rewrite: (path) => path.replace(/^\/mojang\/\w+/, ""),
});

export default defineConfig({
  vite: {
    server: {
      proxy: {
        "/mojang/api": mojang("https://api.mojang.com"),
        "/mojang/session": mojang("https://sessionserver.mojang.com"),
      },
    },
  },
});
