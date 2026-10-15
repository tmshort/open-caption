// Checks that the browser pages stay inside the documented browser floor (docs/05): the JavaScript must parse as
// ES2018 (async functions and object spread, so Safari 11.1 and Chrome 64), and none of the newer APIs or CSS
// that older browsers lack may creep in without a fallback. Runs at the start of the end-to-end test.
'use strict';
const acorn = require('acorn');
const fs = require('fs'), path = require('path');
const web = path.join(__dirname, '..', '..', 'web');

const FORBIDDEN_JS = [
  [/\.replaceAll\(/, 'String.replaceAll (Safari 13.1)'], [/\.at\(-?\d/, 'Array.at (Safari 15.4)'], [/structuredClone/, 'structuredClone (Safari 15.4)'],
  [/ResizeObserver/, 'ResizeObserver (Safari 13.1)'], [/\.matchAll\(/, 'matchAll (Safari 13)'], [/AbortSignal\.timeout/, 'AbortSignal.timeout'],
  [/\.showPopover|popover=/, 'popover API'], [/Object\.hasOwn\(/, 'Object.hasOwn (Safari 15.4)'], [/\.findLast\(/, 'findLast'],
];
const FORBIDDEN_CSS = [
  [/:is\(|:where\(|:has\(/, ':is/:where/:has selectors'], [/\b[sd]vh\b|\blvh\b/, 'dvh/svh units'], [/color-mix\(/, 'color-mix'],
  [/@container/, 'container queries'], [/accent-color/, 'accent-color'], [/\binset\s*:/, 'inset (Safari 14.1) -- use top/right/bottom/left on the audience page'],
];
let bad = 0;
const fail = (f, msg) => { bad++; console.log('  FAIL ' + path.relative(web, f) + ': ' + msg); };
function walk(dir, out = []) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) { if (!['fonts', 'sample', 'help'].includes(e.name)) walk(p, out); } else out.push(p);
  }
  return out;
}
for (const f of walk(web)) {
  const text = fs.readFileSync(f, 'utf8');
  const audienceOrShared = /[\\/](audience|shared)[\\/]/.test(f);
  if (f.endsWith('.js')) {
    try { acorn.parse(text, { ecmaVersion: f.endsWith('compat.js') ? 5 : 2018 }); } catch (e) { fail(f, 'does not parse as ' + (f.endsWith('compat.js') ? 'ES5' : 'ES2018') + ': ' + e.message); }
    for (const [re, what] of FORBIDDEN_JS) if (re.test(text)) fail(f, 'uses ' + what);
  } else if (f.endsWith('.css')) {
    for (const [re, what] of FORBIDDEN_CSS) {
      if (re.test(text) && !(what.startsWith('inset') && !audienceOrShared)) fail(f, 'uses ' + what);
    }
  }
}
console.log(bad ? `\n${bad} compatibility problem(s)` : '  ok   the pages stay inside the browser floor (ES2018 syntax, no newer APIs or CSS)');
process.exit(bad ? 1 : 0);
