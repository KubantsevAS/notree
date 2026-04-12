import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import reactX from 'eslint-plugin-react-x';
import prettierConfig from 'eslint-config-prettier';
import { defineConfig } from 'eslint/config';

export default defineConfig([
  {
    ignores: ['**/.react-router/**'],
  },
  eslint.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{js,jsx,ts,tsx}'],
    plugins: {
      'react-x': reactX,
    },
    rules: {
      ...reactX.configs.recommended.rules,
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
  },
  prettierConfig,
]);
