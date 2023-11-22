/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./**/*.{templ,go,html}",
    "./internal/static/**/*.js",
    "./internal/static/*.js",
    "./internal/static/input.css",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
