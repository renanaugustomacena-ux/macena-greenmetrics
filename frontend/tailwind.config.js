/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        // Italian-flagship palette: regulator-grade, calm, slightly saturated.
        // forest = primary action, brand. earth = warm secondary. cream =
        // background. saffron + rosso provide warning/danger semantics.
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
          300: '#faeca8',
          400: '#f6dc7a'
        },
        saffron: {
          50:  '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          500: '#f59e0b',
          700: '#b45309',
          900: '#78350f'
        },
        rosso: {
          50:  '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          500: '#ef4444',
          700: '#b91c1c',
          900: '#7f1d1d'
        }
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        card: '0 1px 2px 0 rgba(20, 83, 45, 0.04), 0 1px 3px 0 rgba(20, 83, 45, 0.06)',
        elevated: '0 4px 12px -2px rgba(20, 83, 45, 0.10), 0 2px 4px -2px rgba(20, 83, 45, 0.08)'
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite'
      }
    }
  },
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')]
};
