import { defineConfig } from "vite";
import autoprefixer from "autoprefixer";
import tailwindcss from "tailwindcss";
import preact from "@preact/preset-vite";

export default defineConfig({
  plugins: [preact()],
  base: "/static/dist",
  root: "internal/static",
  build: {
    outDir: "dist",
    lib: {
      entry: {
        main: "src/main.ts",
        practice: "src/practice.ts",
        about: "src/about.ts",
        "notes-display": "src/notes-display.ts",
        "practice-menu": "src/practice-menu.ts",
        "add-spot": "src/add-spot.ts",
        "edit-spot": "src/edit-spot.ts",
        prompts: "src/prompts.ts",
        "spot-breakdown": "src/spot-breakdown.ts",
      },
      formats: ["es"],
    },
  },
  css: {
    postcss: {
      plugins: [autoprefixer, tailwindcss],
    },
  },
  define: {
    "process.env": process.env,
  },
});
