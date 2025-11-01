/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './cmd/api-server/templates/**/*.html',
  ],
  theme: {
    extend: {
      colors: {
        // Custom colors can be added here if needed
      },
    },
  },
  plugins: [],
  darkMode: 'class',
}
