/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        forest: {
          50:  '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
          950: '#052e16'
        },
        earth: {
          50:  '#fdfbf5',
          100: '#fbf4de',
          200: '#f7e8bc',
          300: '#f1d892',
          400: '#e7c15e',
          500: '#d5a337',
          600: '#b4832a',
          700: '#8a6222',
          800: '#854d0e',
          900: '#5a3610',
          950: '#341e07'
        },
        cream: {
          50:  '#fffef9',
          100: '#fefae0',
          200: '#fdf4c9',
          300: '#faeca8'
        }
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto', 'sans-serif']
      }
    }
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
};
