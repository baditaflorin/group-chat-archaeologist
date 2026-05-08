import type { Config } from 'tailwindcss';

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: '#16251f',
        paper: '#f7f5ef',
        moss: '#2f5f53',
        rust: '#9a5038',
        wheat: '#d8be7f',
        lake: '#2f6f89'
      },
      boxShadow: {
        soft: '0 18px 60px rgba(22, 37, 31, 0.12)'
      }
    }
  },
  plugins: []
} satisfies Config;
