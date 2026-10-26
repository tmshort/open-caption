// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Browser coverage of the page scripts and styles, from Chrome's own coverage (no extra packages). The browser tests turn
// it on with COVERAGE=1 (make cover-js): every page they open records which parts of the JavaScript and CSS ran, the
// results of all pages and of both test scripts are merged, and `node coverage.js report` writes coverage/js.txt (a
// summary) and coverage/js.html (the lines no test ran).
//
// What is measured is Chrome's block coverage, not statements: a line counts as run when some code on it ran. CSS
// "run" means the rule was used by some page the tests opened (so dark-mode and print rules show as unused unless a
// test turns them on). It is a guide to what the tests never touch, not a pass mark.
'use strict';
const fs = require('fs'), path = require('path');

const OUT = path.join(__dirname, '..', '..', 'coverage');
const files = new Map(); // url -> { kind, text, ranges: [[start, end], ...] }
const open = new Set();

function add(kind, e) {
  let u;
  try { u = new URL(e.url); } catch (x) { return; }
  if (!/^https?:$/.test(u.protocol) || !/\.(js|css)$/.test(u.pathname)) return;
  const f = files.get(u.pathname) || { kind, text: e.text, ranges: [] };
  for (const r of e.ranges) f.ranges.push([r.start, r.end]);
  files.set(u.pathname, f);
}

async function collect(page) {
  if (!open.delete(page)) return;
  try {
    const [js, css] = await Promise.all([page.coverage.stopJSCoverage(), page.coverage.stopCSSCoverage()]);
    js.forEach((e) => add('js', e));
    css.forEach((e) => add('css', e));
  } catch (e) { /* the page is already gone */ }
}

async function started(page) {
  const keep = { resetOnNavigation: false }; // pages are reloaded and navigated all the time: keep what ran before
  await Promise.all([page.coverage.startJSCoverage(keep), page.coverage.startCSSCoverage(keep)]);
  open.add(page);
  const close = page.close.bind(page);
  page.close = async (...a) => { await collect(page); return close(...a); };
  return page;
}

// Make every page this browser (and its separate contexts) opens record coverage.
function instrument(browser) {
  const newPage = browser.newPage.bind(browser);
  browser.newPage = async (...a) => started(await newPage(...a));
  const createContext = browser.createBrowserContext.bind(browser);
  browser.createBrowserContext = async (...a) => {
    const ctx = await createContext(...a);
    const np = ctx.newPage.bind(ctx);
    ctx.newPage = async (...b) => started(await np(...b));
    return ctx;
  };
  const close = browser.close.bind(browser);
  browser.close = async (...a) => { for (const p of [...open]) await collect(p); return close(...a); };
  return browser;
}

// Write what was collected, to be merged with the other test script's by report().
function save(name) {
  fs.mkdirSync(OUT, { recursive: true });
  const list = [...files].map(([url, f]) => ({ url, kind: f.kind, text: f.text, ranges: f.ranges }));
  fs.writeFileSync(path.join(OUT, `js-raw-${name}.json`), JSON.stringify(list));
}

function merge(ranges) {
  const r = ranges.slice().sort((a, b) => a[0] - b[0]), out = [];
  for (const x of r) {
    const last = out[out.length - 1];
    if (last && x[0] <= last[1]) last[1] = Math.max(last[1], x[1]); else out.push([x[0], x[1]]);
  }
  return out;
}

// For each line of text: null for a blank line, else whether any of its code ran.
function lineState(text, ranges) {
  const merged = merge(ranges);
  const states = [];
  let pos = 0, k = 0;
  for (const line of text.split('\n')) {
    const end = pos + line.length;
    let code = false, ran = false;
    for (let i = 0; i < line.length && !ran; i++) {
      if (/\s/.test(line[i])) continue;
      code = true;
      const at = pos + i;
      while (k < merged.length && merged[k][1] <= at) k++;
      if (k < merged.length && merged[k][0] <= at) ran = true;
    }
    states.push(code ? ran : null);
    pos = end + 1;
    k = 0; // lines are few; a rescan keeps the logic plain
  }
  return states;
}

// The translation tables are data that always "runs"; they are listed but left out of the totals.
const DATA = /\/i18n\/[a-z]{2}\.js$/;
const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;');

function report() {
  const byUrl = new Map();
  for (const f of fs.readdirSync(OUT).filter((n) => /^js-raw-.*\.json$/.test(n))) {
    for (const e of JSON.parse(fs.readFileSync(path.join(OUT, f), 'utf8'))) {
      const cur = byUrl.get(e.url) || { kind: e.kind, text: e.text, ranges: [] };
      cur.ranges.push(...e.ranges);
      byUrl.set(e.url, cur);
    }
  }
  if (!byUrl.size) throw new Error('no coverage collected: run the browser tests with COVERAGE=1 first (make cover-js)');
  const rows = [], totals = { js: [0, 0], css: [0, 0] };
  let html = '<!doctype html><meta charset="utf-8"><title>Page coverage</title><style>body{font:14px system-ui;margin:1.5rem}pre{margin:0;font:12px/1.4 ui-monospace,monospace}' +
    '.n{color:#888;display:inline-block;width:4em;text-align:right;margin-right:1em}.miss{background:#fcd9d9}details{margin:.5rem 0}summary{cursor:pointer}</style><h1>Page coverage</h1>' +
    '<p>Lines in red had no code that ran in any browser test. Chrome block coverage, so a guide, not a pass mark.</p>';
  for (const [url, f] of [...byUrl].sort((a, b) => a[0].localeCompare(b[0]))) {
    const st = lineState(f.text, merge(f.ranges));
    const code = st.filter((s) => s !== null).length, ran = st.filter((s) => s === true).length;
    if (!DATA.test(url)) { totals[f.kind][0] += ran; totals[f.kind][1] += code; }
    rows.push({ url, kind: f.kind, code, ran });
    const lines = f.text.split('\n');
    html += `<details><summary><b>${esc(url)}</b> ${code ? (100 * ran / code).toFixed(1) : '100.0'}% (${ran} of ${code} lines)</summary><pre>` +
      lines.map((l, i) => `<span class="${st[i] === false ? 'miss' : ''}"><span class="n">${i + 1}</span>${esc(l)}</span>`).join('\n') + '</pre></details>';
  }
  const pct = (a, b) => (b ? (100 * a / b).toFixed(1) : '100.0') + '%';
  let txt = '';
  for (const r of rows) txt += `${pct(r.ran, r.code).padStart(7)}  ${String(r.code).padStart(5)} lines  ${r.url}\n`;
  txt += `\n${pct(totals.js[0], totals.js[1]).padStart(7)}  ${String(totals.js[1]).padStart(5)} lines  JavaScript total (translation tables not counted)\n`;
  txt += `${pct(totals.css[0], totals.css[1]).padStart(7)}  ${String(totals.css[1]).padStart(5)} lines  CSS total\n`;
  fs.writeFileSync(path.join(OUT, 'js.txt'), txt);
  fs.writeFileSync(path.join(OUT, 'js.html'), html);
  console.log(txt + '\nLines no test ran: open coverage/js.html');
}

module.exports = { instrument, save, report };
if (require.main === module) report();
