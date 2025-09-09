/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx}",      // folder app (App Router)
    "./pages/**/*.{js,ts,jsx,tsx}",    // optional ila kayn pages dir
    "./components/**/*.{js,ts,jsx,tsx}" // components
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
