/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/templates/**/*.templ",
    "./internal/templates/**/*.html",
    "./static/js/**/*.js"
  ],
  darkMode: 'media',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
        }
      }
    },
  },
  plugins: [],
}