/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}", // Include all JS, JSX, TS, and TSX files in the src directory
  ],
  theme: {
    extend: {
      colors: {
        primary: '#1DA1F2',
        secondary: '#14171A',
      },
    },
  },
  
  plugins: [],
}
