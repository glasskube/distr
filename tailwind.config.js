/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./frontend/cloud-ui/src/**/*.{html,ts}', './frontend/cloud-ui/node_modules/flowbite/**/*.js'],
  darkMode: 'selector',
  theme: {
    extend: {
      spacing: {
        144: '36rem',
      },
      /*colors: {
        primary: {"50":"#eff6ff","100":"#dbeafe","200":"#bfdbfe","300":"#93c5fd","400":"#60a5fa","500":"#3b82f6","600":"#2563eb","700":"#1d4ed8","800":"#1e40af","900":"#1e3a8a","950":"#172554"}
      }*/
    },
    fontFamily: {
      sans: ['Inter', 'system-ui', 'sans-serif'],
      display: ['Poppins', 'Inter', 'system-ui', 'sans-serif'],
    },
  },
  plugins: [require('flowbite/plugin')],
};
