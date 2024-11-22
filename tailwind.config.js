/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./frontend/cloud-ui/src/**/*.{html,ts}', './frontend/cloud-ui/node_modules/flowbite/**/*.js'],
  darkMode: 'selector',
  theme: {
    extend: {},
  },
  plugins: [require('flowbite/plugin')],
};
