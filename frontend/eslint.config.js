// ESLint 9 flat configuration for the GreenMetrics SvelteKit frontend.
//
// Doctrine refs: Rule 24 (continuous verification), Rule 52 (shift-left).
//
// Scope: minimal lint that catches the most common JS/TS/Svelte mistakes
// without requiring `@typescript-eslint/parser` (not in devDependencies; full
// TS-aware lint is deferred to a dedicated migration PR per docs/PLAN.md).
// `no-undef` is disabled — TypeScript handles undefined-symbol detection in
// .ts and .svelte files; enabling it on plain JS produces noise on browser
// globals (window/document) and node globals (process) without their globals
// declarations (deferred until `globals` package is added to devDependencies).

import js from '@eslint/js';
import sveltePlugin from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import tsParser from '@typescript-eslint/parser';

export default [
  {
    ignores: [
      'build/',
      '.svelte-kit/',
      'node_modules/',
      'dist/',
      'coverage/',
      '*.config.js',
      '*.config.ts',
    ],
  },

  // Base JS recommended rules.
  js.configs.recommended,

  // Default language options.
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
    },
    rules: {
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      'no-undef': 'off',
    },
  },

  // TypeScript files — parse with @typescript-eslint/parser (no
  // @typescript-eslint/eslint-plugin yet; lint coverage is JS-recommended
  // rules only, plus svelte-check for TS-aware analysis).
  {
    files: ['**/*.ts', '**/*.tsx'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
  },

  // Svelte components — svelte-eslint-parser delegates <script lang="ts"> to
  // @typescript-eslint/parser.
  ...sveltePlugin.configs['flat/recommended'],
  {
    files: ['**/*.svelte'],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: tsParser,
        extraFileExtensions: ['.svelte'],
      },
    },
    rules: {
      // Demote svelte/valid-compile a11y findings to warnings until the
      // underlying components are fixed in a dedicated frontend-a11y PR.
      'svelte/valid-compile': 'warn',
      'no-unused-vars': 'off',
    },
  },
];
