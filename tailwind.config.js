const defaultTheme = require("tailwindcss/defaultTheme");

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./internal/**/*.templ",
    "./internal/static/**/*.js",
    "./internal/components/**/*.go",
    "./internal/static/src/**/*.{ts,tsx,html,css}",
    "./internal/static/src/*.{ts,tsx,html,css}",
  ],
  theme: {
    screens: {
      xs: "475px",
      ...defaultTheme.screens,
    },
    extend: {},
  },
  plugins: [require("@tailwindcss/typography"), require("tailwindcss-animate")],
};
