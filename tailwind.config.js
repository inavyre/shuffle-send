/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["templates/*.html"],
  theme: {
    extend: {
        boxShadow: {
            "rb-lg": "10px 10px 13px -5px rgb(0 0 0 / 0.1), 0 8px 8px -4px rgb(0 0 0 / 0.1)",
        }
    },
  },
  plugins: [],
}

