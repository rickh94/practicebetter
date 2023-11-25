/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./internal/**/*.templ",
    "./internal/static/**/*.js",
    "./internal/static/src/**/*.{ts,tsx,html,css}",
    "./internal/static/src/*.{ts,tsx,html,css}",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/typography")],
};
