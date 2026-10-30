// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// ESLint for the page scripts (web/), the browser tests (test/e2e) and tools/. The linter and its plugins are installed
// with the browser tests' packages (test/e2e/package.json), so this file loads them from there. Run it with `make lint-js`.
import { createRequire } from 'node:module';

const require = createRequire(new URL('./test/e2e/', import.meta.url));
const js = require('@eslint/js');
const globals = require('globals');

// What the page scripts share through the window (see the header of each file).
const shared = { OCI18n: 'writable', OCReport: 'writable', OCClients: 'writable', OCTour: 'writable', OCSpeakers: 'writable', __: 'writable' };

export default [
  { ignores: ['test/e2e/node_modules/**', 'bin/**', 'dist/**', 'web/sample/**', 'web/shared/fonts/**'] },
  js.configs.recommended,
  {
    // The pages: plain scripts, no build step, the browser floor is ES2018 (test/e2e/check-compat.js checks the syntax).
    files: ['web/**/*.js'],
    languageOptions: { ecmaVersion: 2018, sourceType: 'script', globals: { ...globals.browser, ...shared } },
    rules: { 'no-unused-vars': ['error', { args: 'none', caughtErrors: 'none' }], 'no-empty': ['error', { allowEmptyCatch: true }] },
  },
  {
    // The tests and tools run in Node (some of their code runs inside the page, through puppeteer's evaluate).
    files: ['test/**/*.js', 'tools/**/*.js'],
    languageOptions: { ecmaVersion: 2022, sourceType: 'commonjs', globals: { ...globals.node, ...globals.browser, ...shared, axe: 'readonly' } },
    rules: { 'no-unused-vars': ['error', { args: 'none', caughtErrors: 'none' }], 'no-empty': ['error', { allowEmptyCatch: true }] },
  },
  { files: ['eslint.config.mjs'], languageOptions: { ecmaVersion: 2022, sourceType: 'module', globals: globals.node } },
];
