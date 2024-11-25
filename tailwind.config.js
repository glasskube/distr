/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./frontend/cloud-ui/src/**/*.{html,ts}', './frontend/cloud-ui/node_modules/flowbite/**/*.js'],
  darkMode: 'selector',
  theme: {
    extend: {},
    fontFamily: {
      sans: ['Inter', 'system-ui', 'sans-serif'],
      display: ['Poppins', 'Inter', 'system-ui', 'sans-serif'],
    },
  },
  plugins: [require('flowbite/plugin')],
};
