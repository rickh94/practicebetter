import { defineConfig } from "vite";
import preact from "@preact/preset-vite";

export default defineConfig({
  plugins: [preact()],
  build: {
    outDir: "internal/static/out",
    lib: {
      entry: {
        main: "internal/static/src/main.ts",
        practice: "internal/static/src/practice.ts",
        about: "internal/static/src/about.ts",
        "notes-display": "internal/static/src/notes-display.ts",
        "practice-display": "internal/static/src/practice-display.ts",
        "practice-menu": "internal/static/src/practice-menu.ts",
        "add-spot": "internal/static/src/add-spot.ts",
        "edit-spot": "internal/static/src/edit-spot.ts",
        prompts: "internal/static/src/prompts.ts",
      },
      name: "musiclib",
      formats: ["es"],
    },
  },
  define: {
    "process.env": process.env,
  },
});
