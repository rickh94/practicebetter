import { defineConfig } from "vite";
import preact from "@preact/preset-vite";

export default defineConfig({
  plugins: [preact()],
  build: {
    outDir: "internal/static/dist",
    lib: {
      entry: {
        main: "internal/static/src/main.ts",
        practice: "internal/static/src/practice.ts",
        about: "internal/static/src/about.ts",
      },
      name: "musiclib",
      formats: ["es"],
    },
  },
  define: {
    "process.env": process.env,
  },
});
