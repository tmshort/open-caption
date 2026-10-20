// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Checks that Setup, the operator page and Help are translated into every language the audience page has:
// every English sentence the pages can show must have an entry in each dictionary (web/shared/i18n/<code>.js),
// every dictionary must have the same entries, and placeholders ({name}) must match the English text.
//
//   node check-i18n.js          check (exit 1 on problems)
//   node check-i18n.js --list   print the English sentences that need translating
'use strict';
const acorn = require('acorn');
const fs = require('fs'), path = require('path'), vm = require('vm');
const root = path.join(__dirname, '..', '..');
const web = path.join(root, 'web');
const LANGS = ['es', 'fr', 'de', 'it', 'pt'];

// ---- what needs translating ----
const required = new Map(); // English text -> where it came from
const need = (text, where) => { text = text.trim(); if (/[A-Za-z]{2}/.test(text) && !required.has(text)) required.set(text, where); };

// 1. JavaScript: the first argument of every __('...') call (a template with {name} placeholders is kept as is).
for (const f of ['admin/admin.js', 'operator/operator.js', 'operator/report.js', 'shared/report.js', 'helpui/help.js']) {
  const src = fs.readFileSync(path.join(web, f), 'utf8');
  const ast = acorn.parse(src, { ecmaVersion: 2018, locations: true });
  (function walk(n) {
    if (!n || typeof n.type !== 'string') return;
    if (n.type === 'CallExpression' && n.callee.type === 'Identifier' && n.callee.name === '__' && n.arguments[0] && n.arguments[0].type === 'Literal' && typeof n.arguments[0].value === 'string') need(n.arguments[0].value, f + ':' + n.loc.start.line);
    for (const k in n) { const c = n[k]; if (Array.isArray(c)) c.forEach(walk); else if (c && typeof c.type === 'string') walk(c); }
  })(ast);
  // Strings handed to __() through a variable (tables of labels): listed by hand in i18n-tables.json.
}
const tables = JSON.parse(fs.readFileSync(path.join(__dirname, 'i18n-tables.json'), 'utf8'));
for (const [where, list] of Object.entries(tables)) list.forEach((t) => need(t, 'table:' + where));

// 2. Static HTML: text, placeholder/title/aria-label, and whole elements marked data-t.
for (const f of ['admin/index.html', 'operator/index.html', 'operator/report.html', 'helpui/index.html']) {
  let html = fs.readFileSync(path.join(web, f), 'utf8').replace(/<script[\s\S]*?<\/script>/g, '').replace(/<!--[\s\S]*?-->/g, '');
  for (const m of html.matchAll(/<([a-z0-9]+)\b[^>]*\bdata-t\b[^>]*>([\s\S]*?)<\/\1>/g)) { need(m[2].replace(/\s+/g, ' '), f + ' (data-t)'); html = html.replace(m[0], ''); }
  for (const m of html.matchAll(/\b(?:placeholder|title|aria-label|alt)="([^"]+)"/g)) need(m[1], f);
  for (const m of html.matchAll(/>([^<>]+)</g)) { const t = m[1].replace(/\s+/g, ' ').trim(); if (t && !/^[\W\d_]*$/.test(t)) need(t, f); }
  const title = /<title>([^<]+)<\/title>/.exec(html); if (title) need(title[1], f);
}

// 3. Messages the server sends (errors, notes about the show file): a Go format becomes a {0}, {1} template.
const unescape = (s) => s.replace(/\\"/g, '"').replace(/\\n/g, '\n').replace(/\\\\/g, '\\');
const tmpl = (fmt) => { let i = 0; return unescape(fmt).replace(/^%w: /, '').replace(/%%/g, '\u0000').replace(/%[-+# 0-9.]*[a-zA-Z]/g, () => '{' + i++ + '}').replace(/\u0000/g, '%'); };
const goFiles = (dir) => fs.readdirSync(path.join(root, dir)).filter((n) => n.endsWith('.go') && !n.endsWith('_test.go')).map((n) => path.join(root, dir, n));
const STR = '"((?:[^"\\\\]|\\\\.)*)"';
for (const f of [...goFiles('internal/httpapi'), ...goFiles('internal/showfile'), ...goFiles('internal/show')]) {
  const src = fs.readFileSync(f, 'utf8'), rel = path.relative(root, f);
  const res = [
    new RegExp('writeErr\\([^,]+,[^,]+,\\s*(?:fmt\\.Sprintf\\()?' + STR + '(\\s*\\+)?', 'g'),
    new RegExp('"(?:error|note)":\\s*' + STR, 'g'),
    new RegExp('\\.diag\\([^,]+,\\s*[^,]+,\\s*' + STR, 'g'),
    new RegExp('(?:Errorf|errors\\.New)\\(' + STR, 'g'),
    new RegExp('func \\(p \\*parser\\) diag', 'g'),
  ];
  for (const re of res.slice(0, 4)) for (const m of src.matchAll(re)) {
    let t = tmpl(m[1]);
    if (m[2] && /\+/.test(m[2])) t += '{0}'; // "text: " + err.Error()
    if (t.length > 6) need(t, rel);
  }
}
const extra = path.join(__dirname, 'i18n-extra.json');
if (fs.existsSync(extra)) for (const t of JSON.parse(fs.readFileSync(extra, 'utf8'))) need(t, 'i18n-extra.json');

const ignore = new Set(JSON.parse(fs.readFileSync(path.join(__dirname, 'i18n-ignore.json'), 'utf8')));
for (const k of [...required.keys()]) if (ignore.has(k)) required.delete(k);

if (process.argv.includes('--list')) { for (const [k, w] of required) console.log(JSON.stringify(k) + '  // ' + w); process.exit(0); }

// ---- the dictionaries ----
function load(code) {
  const file = path.join(web, 'shared', 'i18n', code + '.js');
  if (!fs.existsSync(file)) return null;
  const dicts = {};
  vm.runInNewContext(fs.readFileSync(file, 'utf8'), { OCI18n: { add: (c, d) => { dicts[c] = d; } } });
  return dicts[code] || null;
}
let bad = 0;
const fail = (m) => { bad++; console.log('  FAIL ' + m); };
const slots = (s) => (s.match(/\{\w+\}/g) || []).sort().join(',');
for (const code of LANGS) {
  const d = load(code);
  if (!d) { fail(code + ': no dictionary (web/shared/i18n/' + code + '.js)'); continue; }
  const missing = [...required.keys()].filter((k) => !(k in d));
  if (missing.length) fail(code + ': ' + missing.length + ' untranslated, for example ' + JSON.stringify(missing.slice(0, process.argv.includes('--all') ? 999 : 3)) + ' (' + required.get(missing[0]) + ')');
  for (const [k, v] of Object.entries(d)) {
    if (typeof v !== 'string' || !v.trim()) fail(code + ': empty translation for ' + JSON.stringify(k));
    else if (slots(k) !== slots(v)) fail(code + ': placeholders differ for ' + JSON.stringify(k) + ' -> ' + JSON.stringify(v));
  }
  const stale = Object.keys(d).filter((k) => !required.has(k));
  if (stale.length) console.log('  note ' + code + ': ' + stale.length + ' entries no longer used, for example ' + JSON.stringify(stale[0]));
}
console.log(bad ? `\n${bad} translation problem(s)` : `  ok   all ${required.size} sentences are translated into ${LANGS.join(', ')}`);
process.exit(bad ? 1 : 0);
