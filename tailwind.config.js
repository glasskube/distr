/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./frontend/ui/src/**/*.{html,ts}'],
  darkMode: 'selector',
  theme: {
    extend: {
      spacing: {
        108: '27rem',
        128: '32rem',
        144: '36rem',
        160: '40rem',
        180: '45rem',
        200: '50rem',
        256: '64rem',
      },
      colors: {
        primary: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
          950: '#172554',
        },
      },
    },
    fontFamily: {
      sans: ['Inter', 'system-ui', 'sans-serif'],
      display: ['Poppins', 'Inter', 'system-ui', 'sans-serif'],
      mono: ['monospace'],
    },
  },
  safelist: [
    'text-5xl',
    'my-5',
    'text-4xl',
    'my-4',
    'text-3xl',
    'my-3',
    'text-2xl',
    'my-2',
    'text-xl',
    'my-1',
    'text-lg',
  ],
  plugins: [require('flowbite/plugin')],
};
