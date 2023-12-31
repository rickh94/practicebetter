const { addDynamicIconSelectors } = require("@iconify/tailwind");
const defaultTheme = require("tailwindcss/defaultTheme");

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./internal/**/*.templ",
    "./internal/static/**/*.js",
    "./internal/components/**/*.go",
    "./internal/static/src/**/*.{ts,tsx,html,css}",
    "**/*.tsx",
    "./internal/static/src/*.{ts,tsx,html,css}",
  ],
  theme: {
    screens: {
      xs: "475px",
      ...defaultTheme.screens,
    },
    extend: {},
  },
  plugins: [
    require("@tailwindcss/typography"),
    require("tailwindcss-animate"),

    addDynamicIconSelectors({
      iconSets: {
        custom: "./internal/static/icons.json",
      },
    }),
  ],
};
