// ESLint flat configuration for pgAnalytics frontend
// @see https://eslint.org/docs/latest/use/configure/configuration-files-new

import js from '@eslint/js'
import typescriptParser from '@typescript-eslint/parser'
import typescript from '@typescript-eslint/eslint-plugin'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'

export default [
  // Base JavaScript recommended rules
  js.configs.recommended,

  // Main configuration for TypeScript/React files
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: typescriptParser,
      parserOptions: {
        ecmaVersion: 2020,
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true,
        },
      },
      globals: {
        // Browser globals
        window: 'readonly',
        document: 'readonly',
        navigator: 'readonly',
        console: 'readonly',
        setTimeout: 'readonly',
        clearTimeout: 'readonly',
        setInterval: 'readonly',
        clearInterval: 'readonly',
        fetch: 'readonly',
        URL: 'readonly',
        URLSearchParams: 'readonly',
        localStorage: 'readonly',
        sessionStorage: 'readonly',

        // Node.js globals for config files and test environment
        process: 'readonly',
        module: 'readonly',
        require: 'readonly',
        __dirname: 'readonly',
        __filename: 'readonly',
        global: 'readonly',

        // DOM Events
        Event: 'readonly',
        MessageEvent: 'readonly',
        CloseEvent: 'readonly',
        EventListener: 'readonly',
        MouseEvent: 'readonly',
        KeyboardEvent: 'readonly',

        // WebSocket API
        WebSocket: 'readonly',

        // File/Blob APIs
        Blob: 'readonly',
        File: 'readonly',
        FormData: 'readonly',

        // Fetch API types
        RequestInit: 'readonly',
        Request: 'readonly',
        Response: 'readonly',
        Headers: 'readonly',

        // DOM Element types (for TypeScript casts)
        HTMLInputElement: 'readonly',
        HTMLSelectElement: 'readonly',
        HTMLTextAreaElement: 'readonly',
        HTMLFormElement: 'readonly',
        HTMLElement: 'readonly',
        Element: 'readonly',
        EventTarget: 'readonly',
        HTMLDivElement: 'readonly',
        HTMLButtonElement: 'readonly',
        HTMLSpanElement: 'readonly',
        HTMLHeadingElement: 'readonly',
        Node: 'readonly',

        // Browser dialogs
        confirm: 'readonly',
        prompt: 'readonly',
        alert: 'readonly',

        // Other common globals
        AbortController: 'readonly',
        AbortSignal: 'readonly',
        ReadableStream: 'readonly',
        WritableStream: 'readonly',

        // Node.js types
        NodeJS: 'readonly',

        // React (for JSX)
        React: 'readonly',
      },
    },
    plugins: {
      '@typescript-eslint': typescript,
      'react': react,
      'react-hooks': reactHooks,
    },
    rules: {
      // TypeScript strict rules
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/no-explicit-any': 'warn',
      'no-unused-vars': 'off', // Use TypeScript version instead

      // React rules
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      'react/react-in-jsx-scope': 'off', // Not needed in React 17+
      'react/prop-types': 'off', // TypeScript handles prop types

      // General best practices
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'prefer-const': 'error',
      'no-var': 'error',
    },
    settings: {
      react: {
        version: 'detect',
      },
    },
  },

  // Test files configuration
  {
    files: ['**/*.test.{ts,tsx}', '**/*.spec.{ts,tsx}', '**/test/**/*.ts', '**/tests/**/*.ts'],
    languageOptions: {
      globals: {
        // Vitest/Jest test globals
        describe: 'readonly',
        it: 'readonly',
        test: 'readonly',
        expect: 'readonly',
        beforeEach: 'readonly',
        afterEach: 'readonly',
        beforeAll: 'readonly',
        afterAll: 'readonly',
        vi: 'readonly',
        jest: 'readonly',
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'no-console': 'off',
    },
  },

  // Configuration files
  {
    files: ['*.config.{js,mjs,ts}', '*.config.*.{js,mjs,ts}'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'no-console': 'off',
    },
  },

  // Ignored paths
  {
    ignores: [
      'dist/**',
      'node_modules/**',
      'coverage/**',
      'e2e/**',
      '*.min.js',
      'public/**',
    ],
  },
]