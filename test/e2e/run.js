// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Browser end-to-end test: starts the built program on a free port with a throwaway data folder and drives
// Setup, Operator and Audience pages in real Chrome. Exits non-zero if any check fails.
//
//   make build && cd test/e2e && npm install && node run.js
//
// Environment: CHROME (path to Chrome/Chromium), OC_BIN (path to the program, default ../../bin/open-caption),
// OC_SHOTS (a folder to save screenshots in).
'use strict';
const puppeteer = require('puppeteer-core');
const { spawn } = require('child_process');
const fs = require('fs'), os = require('os'), path = require('path'), net = require('net');

const CHROME = process.env.CHROME || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const BIN = process.env.OC_BIN || path.join(__dirname, '..', '..', 'bin', 'open-caption' + (process.platform === 'win32' ? '.exe' : ''));
const SHOTS = process.env.OC_SHOTS || '';
const ADMIN = 'e2e-admin-pw', OPERATOR = 'e2e-operator-pw';
// Slow computers (CI runners, a laptop on battery): OC_E2E_SLOW stretches every pause and time limit (default 2 on CI, else 1).
const SLOW = Number(process.env.OC_E2E_SLOW) || (process.env.CI ? 2 : 1);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms * SLOW));
const T = (ms) => ms * SLOW;
// Every browser started is remembered so that a stuck or failed run still closes them (no orphan Chrome or server left behind).
const launched = [];
const launch = async (opts) => { const b = await puppeteer.launch(opts); launched.push(b); return b; };
async function cleanup() {
  for (const b of launched.splice(0)) {
    const proc = b.process && b.process();
    try { await Promise.race([b.close(), new Promise((r) => setTimeout(r, 3000))]); } catch (e) { /* already gone */ }
    if (proc && proc.exitCode === null) { try { proc.kill('SIGKILL'); } catch (e) { /* ignore */ } }
  }
  await stopServer();
  if (dataDir) { try { fs.rmSync(dataDir, { recursive: true, force: true }); } catch (e) { /* the server may still hold files on Windows */ } }
}

let failures = 0, passes = 0;
function check(name, ok, detail) {
  if (ok) { passes++; console.log('  ok   ' + name); }
  else { failures++; console.log('  FAIL ' + name + (detail !== undefined ? '  -> ' + JSON.stringify(detail) : '')); }
}
let sectionName = 'start', sectionStart = Date.now();
const section = (s) => {
  const took = (Date.now() - sectionStart) / 1000;
  if (took > 20) console.log(`  (the section before took ${Math.round(took)} s)`);
  sectionName = s; sectionStart = Date.now();
  console.log('\n' + s);
};
// A section that makes no progress for this long means something hung: say which, clean up and stop, instead of waiting for the whole run to time out.
const SECTION_LIMIT_S = Number(process.env.OC_E2E_SECTION_LIMIT) || 120;
setInterval(() => {
  if (Date.now() - sectionStart > SECTION_LIMIT_S * 1000 * SLOW) {
    console.error(`\nSTUCK: the section "${sectionName}" has run for more than ${SECTION_LIMIT_S * SLOW} s; stopping.`);
    cleanup().finally(() => process.exit(2));
  }
}, 5000).unref();

function freePort() {
  return new Promise((res, rej) => { const s = net.createServer(); s.listen(0, '127.0.0.1', () => { const p = s.address().port; s.close(() => res(p)); }); s.on('error', rej); });
}

let server, dataDir, port, B;
async function startServer() {
  server = spawn(BIN, ['--data', dataDir, '--http-addr', ':' + port, '--no-browser'], { stdio: 'ignore' });
  for (let i = 0; i < 50; i++) { try { const r = await fetch(B + '/healthz'); if (r.ok) return; } catch (e) { /* not up yet */ } await sleep(100); }
  throw new Error('the program did not start');
}
const stopServer = () => new Promise((r) => { if (!server || server.exitCode !== null) return r(); server.once('exit', () => r()); server.kill('SIGTERM'); setTimeout(r, 3000); });

async function main() {
  if (!fs.existsSync(BIN)) throw new Error('build the program first: make build (looked for ' + BIN + ')');
  dataDir = fs.mkdtempSync(path.join(os.tmpdir(), 'oc-e2e-'));
  fs.writeFileSync(path.join(dataDir, 'config.yaml'), `admin_password: ${ADMIN}\noperator_password: ${OPERATOR}\nmdns:\n  enabled: false\naccess:\n  local_auto_login: false\n`);
  port = await freePort(); B = 'http://localhost:' + port;
  await startServer();

  const b = await launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  const errs = [];
  const open = async (opts = {}) => {
    const p = await b.newPage(); p.setDefaultTimeout(T(8000));
    await p.setViewport(opts.viewport || { width: 1300, height: 950 });
    p.on('dialog', (d) => d.accept());
    p.on('pageerror', (e) => errs.push(e.message));
    if (opts.headers) await p.setExtraHTTPHeaders(opts.headers);
    // Some computers (CI runners on Windows and macOS) have animations turned off, which the page rightly honours; the tests
    // of how things animate need the ordinary setting.
    await p.emulateMediaFeatures([{ name: 'prefers-reduced-motion', value: 'no-preference' }]);
    return p;
  };
  const shot = (p, n) => (SHOTS ? p.screenshot({ path: path.join(SHOTS, n + '.png') }) : null);
  const click = (p, label, sel = 'button') => p.evaluate((l, s) => { const e = [...document.querySelectorAll(s)].find((x) => x.textContent.includes(l)); if (!e) throw new Error('no ' + s + ' with text ' + l); e.click(); }, label, sel);
  const step = (p, label) => click(p, label, '#steps button');
  const api = (p, method, url, body) => p.evaluate(async (m, u, bd) => { const r = await fetch(u, { method: m, headers: { 'Content-Type': 'application/json' }, body: bd === undefined ? undefined : JSON.stringify(bd) }); let j = null; try { j = await r.json(); } catch (e) { /* none */ } return { status: r.status, body: j }; }, method, url, body);

  const compat = require('child_process').spawnSync(process.execPath, [path.join(__dirname, 'check-compat.js')], { encoding: 'utf8' });
  check('the pages stay inside the browser floor (ES2018 syntax, no newer APIs or CSS)', compat.status === 0, compat.stdout);
  const i18n = require('child_process').spawnSync(process.execPath, [path.join(__dirname, 'check-i18n.js')], { encoding: 'utf8' });
  check('every sentence in Setup, the operator page and Help is translated into all five other languages', i18n.status === 0, i18n.stdout);

  try {
    // ------------------------------------------------------------------ Setup
    section('Setup: sign in, sample show, show file, languages');
    const admin = await open();
    await admin.goto(B + '/admin'); await admin.waitForSelector('#pw');
    // The form can show before the page's script has attached its handlers on a slow computer: try again.
    for (let attempt = 0; attempt < 4; attempt++) {
      await admin.evaluate(() => { document.getElementById('pw').value = ''; });
      await admin.type('#pw', 'wrong-password'); await admin.keyboard.press('Enter');
      const shown = await admin.waitForFunction(() => document.getElementById('loginerr').textContent.trim() !== '', { timeout: 4000 }).then(() => true, () => false);
      if (shown) break;
    }
    const loginErr = await admin.evaluate(() => document.getElementById('loginerr').textContent);
    check('a wrong password is refused', loginErr.includes('not accepted'), loginErr);
    await admin.evaluate(() => { document.getElementById('pw').value = ''; });
    await admin.type('#pw', ADMIN); await admin.keyboard.press('Enter');
    await admin.waitForSelector('#steps button');
    await admin.waitForFunction(() => [...document.querySelectorAll('button')].some((x) => x.textContent.includes('sample')), { timeout: 20000 });
    check('with no shows the page says Add a show and offers Import as a button like the others', await admin.evaluate(() => {
      const h = document.querySelector('#panel h2'), b = [...document.querySelectorAll('#panel button')].find((x) => x.textContent === 'Import a show…'), f = document.querySelector('#panel input[type=file]'), c = [...document.querySelectorAll('#panel button')].find((x) => x.textContent.includes('sample'));
      return h.textContent === 'Add a show' && !!b && !!f && f.hidden && b.getBoundingClientRect().height === c.getBoundingClientRect().height;
    }));
    await click(admin, 'sample'); await admin.waitForSelector('.editor textarea'); await sleep(900);
    check('the sample show file has no problems', (await admin.evaluate(() => document.querySelector('.diags').innerText)).includes('No problems') || (await admin.evaluate(() => document.getElementById('panel').innerText)).includes('No problems found'));
    const tabs = await admin.evaluate(() => [...document.querySelectorAll('.langtabs button')].map((x) => x.textContent.trim()));
    check('English (main) and Español tabs are offered', tabs.some((t) => t.startsWith('English')) && tabs.some((t) => t.startsWith('Español')), tabs);
    const showId = (await api(admin, 'GET', '/api/admin/shows')).body.shows[0].id;
    await shot(admin, '01-setup-showfile');

    // edit conflict
    await api(admin, 'PUT', `/api/admin/shows/${showId}/showfile`, { text: 'THEIR CHANGE\nline two' });
    await admin.evaluate(() => { const ta = document.querySelector('.editor textarea'); ta.focus(); ta.setRangeText('MY CHANGE\n', 0, 0, 'end'); ta.dispatchEvent(new Event('input', { bubbles: true })); });
    await sleep(200); await click(admin, 'Save'); await sleep(600);
    check('saving over someone else\'s change asks first', (await admin.evaluate(() => document.getElementById('dlg').open && document.getElementById('dlgtitle').textContent)) === 'This show file was changed by someone else');
    await admin.evaluate(() => document.getElementById('dlgextra').click()); await sleep(400);
    check('"use their version" loads theirs', (await admin.evaluate(() => document.querySelector('.editor textarea').value.split('\n')[0])) === 'THEIR CHANGE');
    // a fresh, untouched sample show for the rest of the run
    const dup = await api(admin, 'POST', `/api/admin/shows`, { sample: true });
    check('a second sample show can be created', dup.status === 201);
    const showId2 = dup.body.id;
    await api(admin, 'DELETE', `/api/admin/shows/${showId}`);

    // look
    section('Setup: look');
    await step(admin, 'Show'); await sleep(700); await click(admin, 'Work on this show'); await sleep(700);
    await step(admin, 'Look'); await sleep(900);
    await admin.evaluate(() => { const f = [...document.querySelectorAll('select')].find((s) => [...s.options].some((o) => o.value === 'dyslexic')); f.value = 'dyslexic'; f.dispatchEvent(new Event('input', { bubbles: true })); });
    await admin.evaluate(() => { const sp = [...document.querySelectorAll('select')].find((s) => [...s.options].some((o) => o.value === 'color')); sp.value = 'color'; sp.dispatchEvent(new Event('input', { bubbles: true })); });
    await admin.evaluate(() => { const sp = [...document.querySelectorAll('select')].find((s) => [...s.options].some((o) => o.value === 'beside')); sp.value = 'beside'; sp.dispatchEvent(new Event('input', { bubbles: true })); });
    check('the Look preview puts the names beside the text when asked', await admin.evaluate(() => { const e = document.querySelector('.preview.proj .spk'); return document.querySelector('.preview.proj').classList.contains('pos-beside') && !!e && getComputedStyle(e).display === 'inline'; }));
    await admin.evaluate(() => { const sp = [...document.querySelectorAll('select')].find((s) => [...s.options].some((o) => o.value === 'column')); sp.value = 'column'; sp.dispatchEvent(new Event('input', { bubbles: true })); });
    check('the Look preview can put the names in a column left of the text', await admin.evaluate(() => { const p = document.querySelector('.preview.proj p'); return getComputedStyle(p).display === 'grid'; }));
    await admin.evaluate(() => { const sp = [...document.querySelectorAll('select')].find((s) => [...s.options].some((o) => o.value === 'beside')); sp.value = 'beside'; sp.dispatchEvent(new Event('input', { bubbles: true })); });
    await click(admin, 'Save look'); await sleep(700);
    const look = (await api(admin, 'GET', '/api/admin/shows/' + showId2)).body;
    check('the look is saved with the show', look.lookSaved === true && look.theme.font === 'dyslexic' && look.theme.speakers === 'color' && look.theme.speakerPosition === 'beside', look.theme);
    check('the Look step is ticked', await admin.evaluate(() => [...document.querySelectorAll('#steps button')].find((x) => x.textContent.includes('Look')).classList.contains('done')));

    // ------------------------------------------------------------------ rehearsal
    section('Rehearsal never reaches the audience');
    await step(admin, 'Test'); await sleep(500); await click(admin, 'Start rehearsal');
    await admin.waitForSelector('.reh iframe'); await sleep(1500);
    check('the Next button turns blue as soon as a rehearsal is running', await admin.evaluate(() => { const b = document.querySelector('#nextstep button'); return !!b && b.classList.contains('primary'); }));
    const rehearsalAudience = await open();
    try {
      await rehearsalAudience.goto(B + '/?rehearsal=1');
      await rehearsalAudience.waitForSelector('#langbtn:not([hidden])');
      await rehearsalAudience.click('#langbtn');
      await Promise.all([
        rehearsalAudience.waitForNavigation({ waitUntil: 'domcontentloaded' }),
        rehearsalAudience.click('#langmenu button[lang="es"]'),
      ]);
      await rehearsalAudience.waitForFunction(() => document.documentElement.lang === 'es');
      check('the rehearsal language picker switches the private audience stream',
        await rehearsalAudience.evaluate(() => location.search.includes('lang=es') && document.documentElement.lang === 'es'));
    } finally {
      await rehearsalAudience.close();
    }
    // The rehearsal frames can reload once while the page settles: retry if one is detached.
    for (let attempt = 0; attempt < 4; attempt++) {
      try {
        const opf = admin.frames().find((f) => f.url().includes('rehearsal=1') && f.url().includes('/operator'));
        await opf.waitForSelector('#list .cue');
        await opf.evaluate(() => document.getElementById('list').click());
        for (let i = 0; i < 3; i++) { await admin.keyboard.press('Space'); await sleep(250); }
        break;
      } catch (e) { if (attempt === 3) throw e; await sleep(700); }
    }
    await api(admin, 'POST', `/api/admin/shows/${showId2}/activate`, {});
    const audIdle = await open({ viewport: { width: 420, height: 760 } }); await audIdle.goto(B + '/'); await sleep(900);
    check('the live audience screen is untouched by a rehearsal', (await audIdle.evaluate(() => document.querySelectorAll('.cap').length)) === 0);
    // Chrome allows only 6 connections per site, and every open page holds a live stream: close what is finished.
    await audIdle.close(); await admin.close();

    // ------------------------------------------------------------------ live show
    section('Live show: operator, audience, languages');
    const op = await open(); await op.goto(B + '/operator'); await op.waitForSelector('#list .cue'); await sleep(800);
    await op.evaluate(() => fetch('/api/operator/claim', { method: 'POST', body: '{"takeover":true}' }));
    const en = await open({ viewport: { width: 420, height: 760, deviceScaleFactor: 2 } }); await en.goto(B + '/'); await sleep(900);
    const es = await open({ viewport: { width: 420, height: 760, deviceScaleFactor: 2 }, headers: { 'Accept-Language': 'es-MX,es;q=0.9,en;q=0.5' } }); await es.goto(B + '/'); await sleep(900);
    // Standby: the show is loaded but not on air, so the audience sees the venue's screen and nothing of the show.
    const standbyShown = (pg) => pg.evaluate(() => { const d = document.getElementById('standby'); return document.body.classList.contains('mode-standby') && getComputedStyle(d).display !== 'none' && d.textContent.trim().length > 0 && document.getElementById('fs').hidden && document.querySelectorAll('.cap').length === 0 && getComputedStyle(document.getElementById('captions')).visibility === 'hidden'; });
    check('before the operator starts the show, the audience sees the standby screen', await standbyShown(en));
    check('with the built-in [oc] logo until the venue sets its own', await en.evaluate(() => { const i = document.getElementById('standbylogo'); return !i.hidden && /favicon\.svg/.test(i.getAttribute('src')) && i.naturalWidth > 0; }));
    check('in every language (the built-in words are translated)', await es.evaluate(() => /Los subtítulos comenzarán en breve/.test(document.getElementById('standby').textContent)));
    check('and the operator page says it is not on air, with START where NEXT will be', await op.evaluate(() => /STANDBY/.test(document.getElementById('live').textContent) && /START/.test(document.getElementById('next').textContent) && document.getElementById('offair').disabled && /STANDBY — /.test(document.getElementById('banner').textContent) && document.getElementById('stop').disabled));
    check('in standby the highlighted next line is the pre-show loop that START will run', await op.evaluate(() => { const n = document.querySelector('#list .cue.next'); return !!n && n.classList.contains('k-loop') && !n.classList.contains('sent') && !document.querySelector('#list .cue.now'); }));
    await op.keyboard.press('Backspace'); await op.keyboard.press('b'); await sleep(400);
    check('nothing but START does anything in standby', await standbyShown(en));
    check('Restart show… is disabled while the show is at the start in standby', await op.evaluate(() => document.getElementById('reset').disabled));
    await op.keyboard.press('Space'); await sleep(T(900)); // START: the pre-show loop begins
    check('START puts the show on air (ON AIR), and the pre-show appears', await op.evaluate(() => /ON AIR/.test(document.getElementById('live').textContent) && !document.getElementById('offair').disabled && /NEXT/.test(document.getElementById('next').textContent)));
    check('and Restart show… works once the show is on air', await op.evaluate(() => !document.getElementById('reset').disabled));
    check('the item of a running loop is marked with a play triangle in the operator list', await op.evaluate(() => { const l = document.querySelectorAll('.items li.playing'); return l.length === 1 && l[0].getAttribute('aria-current') === 'true' && getComputedStyle(l[0].querySelector('.pl'), '::before').content.includes('▶'); }));
    check('the pre-show loop is on the audience screens after START (fullscreen)', await en.evaluate(() => !document.getElementById('fs').hidden));
    check('English and Spanish devices each get their language', (await en.evaluate(() => document.documentElement.lang)) === 'en' && (await es.evaluate(() => document.documentElement.lang)) === 'es');
    check('the language menu is offered', await es.evaluate(() => !document.getElementById('langbtn').hidden));
    // REQ-070: the audience report counts, from START, the devices that were already waiting (both pages share one address and browser, so they are one device)
    {
      const rep = await op.evaluate(async () => { const r = await fetch('/api/operator/report'); return { status: r.status, body: await r.json() }; });
      const tt = (rep.body.report || {}).totals || {};
      check('from START the audience report is running and counts the devices that were waiting, once each', rep.status === 200 && rep.body.report.running === true && tt.devices === 1 && tt.connections >= 2 && rep.body.report.languages.en === 1 && rep.body.report.languages.es === 1, tt);
      const rp = await open(); await rp.goto(B + '/operator/report'); await rp.waitForSelector('#report:not([hidden])', { timeout: T(10000) }).catch(() => {});
      check('the report page shows the totals, the graph of devices over time (whole numbers, the start and the end on its axis), the loop under the name the show file gives it, the languages and a row for each device, and offers print and CSV', await rp.evaluate(() => document.querySelectorAll('#totals .card').length >= 8 && !!document.querySelector('#chart path.line') && [...document.querySelectorAll('#chart text.tick')].filter((t) => /^[\d.]+$/.test(t.textContent)).every((t) => /^\d+$/.test(t.textContent)) && document.querySelectorAll('#chart text.edge').length === 2 && /Pre-show/.test(document.getElementById('loops').textContent) && document.querySelectorAll('#chart .tick').length >= 4 && document.querySelectorAll('#langs tbody tr').length === 2 && document.querySelectorAll('#devices tbody tr').length === 1 && /report\.csv/.test(document.getElementById('csv').getAttribute('href')) && !!document.getElementById('print') && /on air now/.test(document.getElementById('period').textContent)));
      await rp.close();
      // ... and in Setup, as its last step
      const sp = await open(); await sp.goto(B + '/admin#report'); await sp.waitForSelector('#panel .oc-report', { timeout: T(10000) }).catch(() => {});
      check('Setup has the report as its seventh step, the same report with the graph, a menu of reports, Print and CSV', await sp.evaluate(() => { const b = [...document.querySelectorAll('#steps button')]; const cur = document.querySelector('#steps button[aria-current=step]'); return b.length === 7 && /Report/.test(b[6].textContent) && cur === b[6] && document.querySelectorAll('#panel #totals .card').length >= 8 && !!document.querySelector('#panel #chart path.line') && !!document.querySelector('#panel select') && [...document.querySelectorAll('#panel button')].some((x) => x.textContent === 'Print') && /report\.csv/.test(document.querySelector('#panel a[download]').getAttribute('href')); }));
      check('and the report is reachable with no show selected (it does not depend on one)', await sp.evaluate(() => !document.querySelectorAll('#steps button')[6].disabled));
      check('the report of the run in progress has no Delete button (it is saved as it goes)', await sp.evaluate(() => { const d = [...document.querySelectorAll('#panel button')].find((x) => /^Delete/.test(x.textContent)); return !!d && d.hidden; }));
      await sp.close();
    }
    check('page labels are translated (Spanish)', (await es.evaluate(() => document.getElementById('gear').getAttribute('aria-label'))) === 'Tamaño del texto y colores');
    const opButtons = () => op.evaluate(() => ['back', 'next', 'blank', 'clear', 'message'].map((id) => { const r = document.getElementById(id).getBoundingClientRect(); return Math.round(r.left) + ',' + Math.round(r.top); }).join(' '));
    const inLoop = { stop: await op.evaluate(() => { const b = document.getElementById('stop'); return !b.hidden && !b.disabled; }), at: await opButtons() };
    check('during a loop the STOP loop button is available', inLoop.stop);
    for (let i = 0; i < 6; i++) { await op.keyboard.press('Space'); await sleep(250); }
    const afterLoop = { stop: await op.evaluate(() => { const b = document.getElementById('stop'); return b.offsetWidth > 0 && b.disabled; }), at: await opButtons() };
    check('outside a loop it stays in place, disabled', afterLoop.stop);
    check('so Back, Next and the other buttons never move', inLoop.at === afterLoop.at, [inLoop.at, afterLoop.at]);
    check('the operator language menu is styled like the Setup one', await op.evaluate(() => { const c = getComputedStyle(document.querySelector('header select.uilang')); return c.backgroundColor !== 'rgb(255, 255, 255)' && parseFloat(c.borderTopLeftRadius) >= 6; }));
    const caps = (pg) => pg.evaluate(() => [...document.querySelectorAll('.cap')].map((c) => c.textContent.replace(/\s+/g, ' ').trim()));
    const enCaps = await caps(en), esCaps = await caps(es);
    check('one press shows the same lines in both languages', enCaps.length === esCaps.length && enCaps.length >= 3, [enCaps.length, esCaps.length]);
    check('English text is English', enCaps.join(' ').includes('You\'re late, Edwin'), enCaps);
    check('Spanish text is Spanish', esCaps.join(' ').includes('Llegas tarde, Edwin'), esCaps);
    check('the dyslexia-friendly font is applied from the show look', await en.evaluate(() => document.body.classList.contains('dys')));
    const colors = await en.evaluate(() => [...document.querySelectorAll('.spk')].map((s) => s.textContent + ':' + s.style.color));
    check('each character has their own colour', new Set(colors.filter((c) => c.startsWith('MARGARET')).map((c) => c.split(':')[1])).size === 1 && new Set(colors.map((c) => c.split(':')[1])).size >= 2, colors);
    check('the operator sees how many devices read each language', /en 1.*es 1/.test(await op.evaluate(() => document.getElementById('viewers').textContent)));
    await shot(op, '02-operator'); await shot(es, '03-audience-spanish');

    // Skip passes over the next line without sending it; Back takes it back
    {
      const cur = () => op.evaluate(async () => (await (await fetch('/api/operator/state')).json()).payload.op.cursor);
      const c0 = await cur(), n0 = (await caps(en)).length;
      await op.keyboard.press('s'); await sleep(500);
      check('Skip (S) moves past the next line without showing it, and says which', (await cur()) > c0 && (await caps(en)).length === n0 && (await caps(es)).length === n0 && /^Skipped/.test(await op.evaluate(() => document.getElementById('toast').textContent)), [c0, await cur()]);
      check('the skipped line is marked SKIPPED in the list', await op.evaluate(() => { const li = document.querySelector('#list .cue.skipped'); return !!li && /SKIPPED/.test(li.textContent); }));
      await sleep(300); await op.keyboard.press('Backspace'); await sleep(500);
      check('and the mark goes with the skip', await op.evaluate(() => !document.querySelector('#list .cue.skipped')));
      check('Back takes the skip back', (await cur()) === c0, [c0, await cur()]);
    }

    // back, blank vs clear
    const cursor = () => op.evaluate(async () => (await (await fetch('/api/operator/state')).json()).payload.op.cursor);
    const c1 = await cursor(); await op.keyboard.press('Backspace'); await sleep(400);
    check('Backspace undoes the last line in every language', (await cursor()) < c1 && (await caps(es)).length === esCaps.length - 1);
    await op.keyboard.press('b'); await sleep(400);
    check('Blank hides the screen but keeps the lines', (await en.evaluate(() => document.body.classList.contains('mode-blank'))) && (await en.evaluate(() => document.querySelectorAll('.cap').length)) > 0);
    await op.keyboard.press('b'); await sleep(300);
    await op.keyboard.press('c'); await sleep(400);
    check('Clear erases the lines', (await en.evaluate(() => document.querySelectorAll('.cap').length)) === 0);
    await op.keyboard.press('Backspace'); await sleep(400);

    // songs and voices together
    const idx = await op.evaluate(async () => (await (await fetch('/api/operator/state')).json()).cues.findIndex((c) => c.parts && c.parts.length));
    await op.evaluate(async (i) => { await fetch('/api/operator/jump', { method: 'POST', body: JSON.stringify({ index: i }) }); }, idx);
    await sleep(500); await op.keyboard.press('Space'); await sleep(500);
    check('two voices appear together from one press', (await en.evaluate(() => document.querySelectorAll('.cap.group .part').length)) === 2);
    check('sung lines have notes and italics', (await en.evaluate(() => [...document.querySelectorAll('.cap.sung, .cap.group .part.sung')].some((e) => e.textContent.includes('♪')))));
    const posRes = await en.evaluate(() => {
      const one = document.querySelector('.cap:not(.group) .spk'); if (!one) return 'no single-voice caption with a name';
      const out = [];
      document.body.classList.add('spk-beside'); out.push(getComputedStyle(one).display === 'inline');
      document.body.classList.remove('spk-beside'); document.body.classList.add('spk-column'); out.push(getComputedStyle(one.parentNode).display === 'grid' && getComputedStyle(document.querySelector('.cap.group .spk')).display === 'inline');
      document.body.classList.remove('spk-column'); out.push(getComputedStyle(one).display === 'block');
      return out.every(Boolean) || out;
    });
    check('the audience page places names beside the text, or in a column, by the look', posRes === true, posRes);
    {
      // a shared screen never scrolls: a caption taller than the screen is shrunk until it fits (here a small window)
      const proj = await open({ viewport: { width: 420, height: 260 } }); await proj.goto(B + '/?display=projector'); await sleep(1200);
      const r = await proj.evaluate(() => { const c = document.getElementById('captions'), l = c.lastElementChild, cr = c.getBoundingClientRect(), lr = l.getBoundingClientRect(); return { fit: parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--fit')), top: lr.top >= cr.top - 1, bottom: lr.bottom <= cr.bottom + 1 }; });
      check('on a shared screen a caption taller than the screen is shrunk to fit', r.fit < 1 && r.top && r.bottom, r);
      await proj.close();
    }

    // help pauses keys
    await op.keyboard.press('?'); await sleep(1200);
    const before = await cursor(); await op.keyboard.press('Space'); await op.keyboard.press('Space'); await sleep(400);
    check('operator keys are paused while help is open', (await cursor()) === before);
    await op.keyboard.press('Escape'); await sleep(400);
    await op.keyboard.press('Space'); await sleep(400);
    check('keys work again after closing help', (await cursor()) > before);

    // language menu
    await es.evaluate(() => document.getElementById('langbtn').click());
    await es.evaluate(() => [...document.querySelectorAll('#langmenu button')].find((x) => x.textContent === 'English').click()); await sleep(1200);
    check('the language menu switches this device', (await es.evaluate(() => location.search)) === '?lang=en' && (await es.evaluate(() => document.documentElement.lang)) === 'en');

    // scroll back (a plain 40-line show so there is plenty to scroll)
    const long = Array.from({ length: 40 }, (_, i) => `Line ${i + 1}: the quick brown fox jumps over the lazy dog`).join('\n');
    const sid = (await api(op, 'POST', '/api/admin/shows', { name: 'Scroll test' })).body.id;
    await api(op, 'PUT', `/api/admin/shows/${sid}/showfile`, { text: long });
    await api(op, 'POST', `/api/admin/shows/${sid}/activate`, { force: true, start: true }); // the first show is live
    await sleep(600);
    const nextN = (n) => op.evaluate(async (n) => { for (let i = 0; i < n; i++) await fetch('/api/operator/next', { method: 'POST', body: '{}' }); }, n);
    await nextN(25); await sleep(500);
    await en.bringToFront(); // background tabs do not deliver scroll events
    const pos = () => en.evaluate(() => { const c = document.getElementById('captions'); return { top: Math.round(c.scrollTop), max: c.scrollHeight - c.clientHeight, btn: document.getElementById('live').hidden ? '' : document.getElementById('live').textContent }; });
    check('following the live captions, the button is hidden', (await pos()).btn === '' && (await pos()).max > 100, await pos());
    await en.evaluate(() => { const c = document.getElementById('captions'); c.dispatchEvent(new Event('wheel')); c.scrollTop = 120; });
    await en.waitForFunction(() => { const c = document.getElementById('captions'); return c.scrollTop <= 130 && !document.getElementById('live').hidden; });
    const up = await pos();
    check('scrolling up shows "back to live"', up.btn.includes('Back to live'), up);
    await nextN(2); await sleep(500);
    const later = await pos();
    check('new lines do not move a reader, and are counted', later.top === up.top && later.btn.includes('(2)'), later);
    await en.evaluate(() => document.getElementById('live').click()); await sleep(T(900));
    check('"back to live" returns to the newest line', await en.evaluate(() => { const c = document.getElementById('captions'); return c.scrollHeight - c.scrollTop - c.clientHeight < 60; }));
    // A new line glides into place instead of jumping.
    await en.evaluate(() => { const c = document.getElementById('captions'); window.__seen = []; const f = () => { window.__seen.push(c.scrollTop); window.__raf = requestAnimationFrame(f); }; f(); });
    await nextN(1); await sleep(T(900));
    const seen = await en.evaluate(() => { cancelAnimationFrame(window.__raf); return window.__seen; });
    const lo = seen[0], hi = seen[seen.length - 1];
    check('a new line scrolls into view smoothly rather than jumping', hi > lo && seen.some((v) => v > lo + 1 && v < hi - 1), seen.slice(0, 40).join(','));
    // The new line starts out of sight below the bottom edge, and comes up with the others (not popping on).
    const arrival = async (label) => {
      await en.evaluate(() => { const c = document.getElementById('captions'); window.__arr = []; const f = () => { const caps = c.querySelectorAll('.cap'), last = caps[caps.length - 1], r = c.getBoundingClientRect(); window.__arr.push({ n: caps.length, top: last ? last.getBoundingClientRect().top : 0, edge: r.bottom - parseFloat(getComputedStyle(c).paddingBottom) }); window.__raf2 = requestAnimationFrame(f); }; f(); });
      await nextN(1); await sleep(T(900));
      const arr = await en.evaluate(() => { cancelAnimationFrame(window.__raf2); return window.__arr; });
      const k = arr.findIndex((x, i) => i > 0 && x.n > arr[0].n); // the first frame with the new line
      const first = arr[k], last = arr[arr.length - 1];
      check(label + ': the new line starts below the visible edge', k > 0 && first.top >= first.edge - 2 && (await en.evaluate(() => /inset/.test(getComputedStyle(document.getElementById('captions')).clipPath))), first);
      check(label + ': and comes up with the rest', k > 0 && last.top < first.top - 5 && arr.slice(k).some((x) => x.top < first.top - 1 && x.top > last.top + 1), arr.slice(k, k + 12).map((x) => Math.round(x.top)).join(','));
    };
    await arrival('with a full screen');
    check('and the reader is still following the live captions', await en.evaluate(() => document.getElementById('live').hidden));
    await shot(en, '04-audience-english');
    // With only a few lines on screen (they sit at the bottom), the older lines slide up instead of jumping.
    const fid = (await api(op, 'POST', '/api/admin/shows', { name: 'Few lines' })).body.id;
    await api(op, 'PUT', `/api/admin/shows/${fid}/showfile`, { text: 'One\n~~Two~~\n***Three***\nFour' });
    await api(op, 'POST', `/api/admin/shows/${fid}/activate`, { force: true, start: true }); await sleep(T(800));
    await nextN(2); await sleep(T(700));
    await en.evaluate(() => { const first = document.querySelector('#captions .cap'); window.__seen = []; const f = () => { window.__seen.push(Math.round(first.getBoundingClientRect().top)); window.__raf = requestAnimationFrame(f); }; f(); });
    await nextN(1); await sleep(T(900));
    const few = await en.evaluate(() => { cancelAnimationFrame(window.__raf); return window.__seen; });
    const f0 = few[0], f1 = few[few.length - 1];
    check('with few lines on screen, the older ones slide up instead of jumping', f1 < f0 && few.some((v) => v < f0 - 1 && v > f1 + 1), few.slice(0, 40).join(','));
    await arrival('with few lines');
    check('~~strikethrough~~ and ***bold italic*** are drawn on the audience screen', await en.evaluate(() => { const d = document.querySelector('#captions .cap del'), b = document.querySelector('#captions .cap strong em'); return !!d && /line-through/.test(getComputedStyle(d).textDecorationLine) && d.textContent === 'Two' && !!b && b.textContent === 'Three'; }));

    // A device keeps as many lines as the show's scrollback setting, however long it stays connected.
    const lrid = (await api(op, 'POST', '/api/admin/shows', { name: 'Long run' })).body.id;
    await api(op, 'PUT', `/api/admin/shows/${lrid}/showfile`, { text: '[settings]\nscrollback lines: 20\n[/settings]\n' + Array.from({ length: 30 }, (_, i) => 'Line ' + (i + 1)).join('\n') });
    await api(op, 'POST', `/api/admin/shows/${lrid}/activate`, { force: true, start: true }); await sleep(T(800));
    await nextN(30); await sleep(T(700));
    const kept = await en.evaluate(() => { const c = document.querySelectorAll('#captions .cap'); return { n: c.length, last: c[c.length - 1].textContent }; });
    check('a device keeps the show\'s scrollback lines (20), not a fixed number', kept.n === 20 && /Line 30/.test(kept.last), JSON.stringify(kept));

    // ------------------------------------------------------------------ settings & security
    // An intermission timer: the loop's {timer} counts down live on the audience screen, from the moment the loop starts.
    // A show file that puts the names beside the text reaches the audience; a shared screen shrinks a crowded caption to fit
    {
      const crowd = '[settings]\nspeaker position: beside\n[/settings]\nALICE: Hello\n[together]\n' + 'ABCDEFGH'.split('').map((c, i) => `${c}${i}: voice number ${i + 1} sings along`).join('\n') + '\n[/together]\nALICE: Goodbye\n';
      const pid = (await api(op, 'POST', '/api/admin/shows', { name: 'Position test' })).body.id;
      await api(op, 'PUT', `/api/admin/shows/${pid}/showfile`, { text: crowd });
      await api(op, 'POST', `/api/admin/shows/${pid}/activate`, { force: true, start: true }); await sleep(T(800));
      const aud = await open({ viewport: { width: 420, height: 300 } }); await aud.goto(B + '/'); await sleep(T(900));
      const proj = await open({ viewport: { width: 420, height: 300 } }); await proj.goto(B + '/?display=projector'); await sleep(T(900));
      await nextN(1); await sleep(T(700));
      check('a show file that puts names beside the text reaches the audience page', await aud.evaluate(() => { const s = document.querySelector('.cap .spk'); return document.body.classList.contains('spk-beside') && !!s && getComputedStyle(s).display === 'inline'; }));
      await nextN(1); await sleep(T(900));
      const crowded = await proj.evaluate(() => { const c = document.getElementById('captions'), l = c.lastElementChild, cr = c.getBoundingClientRect(), lr = l.getBoundingClientRect(); return { voices: l.querySelectorAll('.part').length, fit: parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--fit')), top: lr.top >= cr.top - 1, bottom: lr.bottom <= cr.bottom + 1 }; });
      check('a shared screen shows all eight voices of a crowded caption, shrunk to fit', crowded.voices === 8 && crowded.fit < 1 && crowded.fit >= 0.4 && crowded.top && crowded.bottom, crowded);
      await nextN(1); await sleep(T(900));
      check('and the text is full size again for the next ordinary caption', await proj.evaluate(() => parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--fit')) === 1));
      await aud.close(); await proj.close(); await en.bringToFront(); // a background tab is throttled, which the timer checks below would feel
    }
    const tid = (await api(op, 'POST', '/api/admin/shows', { name: 'Timer test' })).body.id;
    await api(op, 'PUT', `/api/admin/shows/${tid}/showfile`, { text: '[loop: timer 15 minutes]\n**INTERMISSION**\nBack in {timer}\n[/loop]\nWelcome back' });
    await api(op, 'POST', `/api/admin/shows/${tid}/activate`, { force: true, start: true });
    const timerText = () => en.evaluate(() => { const t = document.querySelector('#fs [data-timer]'); return t ? t.textContent : null; });
    await en.waitForFunction(() => { const t = document.querySelector('#fs [data-timer]'); return t && /\d/.test(t.textContent); }, { timeout: T(25000) }).catch(() => {});
    const t1 = await timerText();
    const mins = (x) => { const m = /(\d+)\s*(?:h|hr)\b\s*(?:(\d+))?/.exec(x || ''); return m ? Number(m[1]) * 60 + Number(m[2] || 0) : Number((/(\d+)/.exec(x || '') || [0, -1])[1]); };
    check('the intermission timer shows the time left in whole minutes (15 min)', /^\d+ ?min/i.test(t1 || '') && mins(t1) === 15, t1);
    check('and has no seconds', !/\d:\d\d/.test(t1 || ''), t1);
    check('the operator sees the countdown too', await op.evaluate(() => /left on the timer/.test((document.getElementById('banner') || {}).textContent || '')));
    // a device whose clock is wrong still shows the right time (the server's clock is used)
    await en.evaluate(() => { const real = Date.now; Date.now = () => real() + 3600e3; }); await sleep(T(900));
    // (the page measured the difference at its snapshot, so reload with the fake clock in place)
    await en.evaluateOnNewDocument(() => { const real = Date.now; Date.now = () => real() + 3600e3; });
    await en.reload(); await en.waitForFunction(() => { const t = document.querySelector('#fs [data-timer]'); return t && /\d/.test(t.textContent); }, { timeout: T(15000) }).catch(() => {});
    const t3 = await timerText();
    check('a device with its clock an hour wrong shows the right time left', mins(t3) >= 14 && mins(t3) <= 15, t3);
    // it counts down: a very short timer (0.05 minutes, 3 seconds) reaches 0 min
    await api(op, 'PUT', `/api/admin/shows/${tid}/showfile`, { text: '[loop: timer 0.05 minutes]\nBack in {timer}\n[/loop]\nWelcome back', confirm: 'Untitled show' });
    await api(op, 'POST', `/api/admin/shows/${tid}/activate`, { force: true, start: true });
    await en.waitForFunction(() => { const t = document.querySelector('#fs [data-timer]'); return t && /^1 ?min/i.test(t.textContent.trim()); }, { timeout: T(15000) }).catch(() => {});
    const first = await timerText();
    await en.waitForFunction(() => { const t = document.querySelector('#fs [data-timer]'); return t && /^0 ?min/i.test(t.textContent.trim()); }, { timeout: T(15000) }).catch(() => {});
    check('the figure counts down to 0 min', /^1 ?min/i.test(first || '') && /^0 ?min/i.test((await timerText()) || ''), [first, await timerText()]);

    // Going off air: the audience gets the standby screen at once, the show keeps its place, START puts it back.
    await op.bringToFront(); // a page in the background does not get its dialog's close event
    await op.evaluate(() => document.getElementById('offair').click());
    await op.waitForFunction(() => document.getElementById('offdlg').open, { timeout: T(5000) }).catch(() => {});
    check('Go to standby asks first, and says what happens', await op.evaluate(() => /standby screen/.test(document.getElementById('offdlg').innerText)));
    await op.evaluate(() => document.getElementById('offgo').click()); await sleep(T(900));
    check('off air: the audience sees the standby screen again, nothing of the show', await standbyShown(en));
    check('and the operator page shows STANDBY with START', await op.evaluate(() => /STANDBY/.test(document.getElementById('live').textContent) && /START/.test(document.getElementById('next').textContent)));
    check('the show can be deleted while off air (it is not live)', await (async () => { const sh = (await api(op, 'GET', '/api/admin/shows')).body.shows.find((x) => x.active); return sh && sh.live === false; })());
    await op.keyboard.press('Space'); await sleep(T(900));
    check('START puts it back on air where it was (the pre-show timer is showing again)', await en.evaluate(() => !document.body.classList.contains('mode-standby') && !!document.querySelector('#fs [data-timer]')));

    // A picture with no words fills the whole screen; with a caption it keeps its margin.
    {
      const sample = (await api(op, 'POST', '/api/admin/shows', { sample: true })).body.id;
      const showImage = async (text) => {
        await api(op, 'PUT', `/api/admin/shows/${sample}/showfile`, { text: '[settings]\ntitle: Pic\n[/settings]\n' + text, confirm: 'Pic' });
        await api(op, 'POST', `/api/admin/shows/${sample}/activate`, { force: true, start: true }); await sleep(T(700));
        await nextN(1); await sleep(T(900));
        return en.evaluate(() => { const i = document.querySelector('#fs img'); if (!i) return null; const r = i.getBoundingClientRect(); return { l: Math.round(r.left), t: Math.round(r.top), w: Math.round(r.width), h: Math.round(r.height), vw: innerWidth, vh: innerHeight, loaded: i.naturalWidth > 0 }; });
      };
      const bare = await showImage('[image: welcome.png]\nHello');
      check('a picture with no caption fills the whole screen', !!bare && bare.loaded && bare.l === 0 && bare.t === 0 && bare.w === bare.vw && bare.h === bare.vh, bare);
      const bare2 = await showImage('![](welcome.png)\nHello');
      check('also when written the Markdown way with no description', !!bare2 && bare2.w === bare2.vw && bare2.h === bare2.vh, bare2);
      const capt = await showImage('[image: welcome.png | The garden]\nHello');
      check('a picture with a caption keeps its margin around it', !!capt && capt.l > 0 && capt.w < capt.vw, capt);
    }
    section('The operator page and Setup');
    // The right side of the header is in the same order on both pages (devices, language, the page's own actions, the other page, Help, Sign out, connection).
    const headerOrder = (pg, sels) => pg.evaluate((sels) => { const pts = sels.map((q) => { const e = document.querySelector(q); if (!e || e.hidden || !e.offsetWidth) return null; const r = e.getBoundingClientRect(); return [Math.round(r.top / 40), r.left]; }); const seen = pts.filter(Boolean); return { ok: seen.every((x, i) => i === 0 || x[0] > seen[i - 1][0] || (x[0] === seen[i - 1][0] && x[1] > seen[i - 1][1])), pts: pts.map((x, i) => (x ? sels[i] : '(' + sels[i] + ' hidden)')) }; }, sels);
    const headerLook = (pg) => pg.evaluate(() => { const g = (q) => { const e = document.querySelector(q); if (!e || !e.offsetWidth) return null; const c = getComputedStyle(e); return [c.fontSize, c.fontFamily, c.fontWeight, Math.round(e.getBoundingClientRect().height)].join('|'); }; const h = document.querySelector('header'); return { header: getComputedStyle(h).fontSize + '|' + getComputedStyle(h).paddingTop + '|' + Math.round(h.getBoundingClientRect().height), title: g('header h1, header #show').split('|').slice(0, 3).join('|'), pill: g('#viewers'), help: g('#helpbtn'), lang: g('header select.uilang') }; });
    const opLook = await headerLook(op);
    const liveRed = (pg) => pg.evaluate(() => { const e = document.createElement('span'); e.className = 'pill live'; document.querySelector('header').appendChild(e); const c = getComputedStyle(e).backgroundColor; e.remove(); return c; });
    const opRed = await liveRed(op);
    const countText = (pg) => pg.evaluate(() => [OCClients.label({ audience: 3, languages: [{ code: 'en' }, { code: 'es' }], audienceByLang: { en: 3 } }), OCClients.label({ audience: 0, languages: [{ code: 'en' }, { code: 'es' }], audienceByLang: {} }), OCClients.label({ audience: 5, languages: [{ code: 'en' }, { code: 'es' }], audienceByLang: { en: 3, es: 2 } }), OCClients.label({ audience: 2, languages: [{ code: 'en' }], audienceByLang: { en: 2 } })]);
    check('the audience count leaves out a language nobody is using', JSON.stringify(await countText(op)) === JSON.stringify(['👥 3 (en 3)', '👥 0', '👥 5 (en 3 · es 2)', '👥 2']));
    check('the count in the operator header shows no language with 0', !/ 0\b/.test(await op.evaluate(() => document.getElementById('viewers').textContent.replace(/^👥 0$/, ''))));
    // the count opens the list of connected devices: device, address, language
    await op.evaluate(() => document.getElementById('viewers').click()); await sleep(700);
    check('clicking the count on the operator page lists the connected devices (device, address, language)', await op.evaluate(() => { const d = document.getElementById('clientsdlg'); const rows = [...d.querySelectorAll('table.clients tr')].slice(1).map((r) => [...r.children].map((c) => c.textContent)); const head = [...d.querySelectorAll('table.clients th')].map((x) => x.textContent).join('|'); return d.open && head === 'Device|IP address|Language' && rows.length >= 1 && rows.every((r) => /^(Phone|Tablet|Computer|Projector screen|Unknown device)/.test(r[0]) && /^[0-9a-f.:]+$/i.test(r[1]) && r[2].length > 0) && rows.some((r) => r[2] === 'English') && rows.length === Number((document.getElementById('viewers').textContent.match(/\d+/) || [0])[0]); }));
    const dlgLook = (pg, sel) => pg.evaluate((q) => { const d = document.querySelector(q), c = (e) => { const x = getComputedStyle(e); return [x.fontSize, x.fontFamily, x.fontWeight, x.lineHeight].join('|'); }; return { width: Math.round(d.getBoundingClientRect().width), title: c(d.querySelector('h2')), th: c(d.querySelector('th')), td: c(d.querySelector('td')), pad: getComputedStyle(d.querySelector('td')).padding, cols: [...d.querySelectorAll('th')].map((x) => Math.round(x.getBoundingClientRect().width)).join(',') }; }, sel);
    const opDlg = await dlgLook(op, '#clientsdlg');
    await op.evaluate(() => document.querySelector('#clientsdlg button').click()); await sleep(300);
    {
      const oh = await headerOrder(op, ['#viewers', 'header select.uilang', '#offair', '#reset', '#setuplink', '#helpbtn', '#logout']);
      check('operator header, right side: devices, language, Go to standby, Restart, Setup, Help, Sign out, connection', oh.ok && oh.pts.filter((x) => x.includes('hidden')).length === 0, oh);
    }
    // (the sign-in rate limit is five a minute, so the role is faked in the page's own state request instead of logging in again)
    for (const [role, linked] of [['operator', false], ['admin', true]]) {
      const pg = await open();
      await pg.evaluateOnNewDocument((r) => { const f = window.fetch; window.fetch = (u, o) => f(u, o).then((resp) => (String(u).includes('/api/operator/state') ? resp.json().then((j) => { j.role = r; return new Response(JSON.stringify(j), { status: 200, headers: { 'Content-Type': 'application/json' } }); }) : resp)); }, role);
      await pg.goto(B + '/operator'); await pg.waitForSelector('#list .cue');
      check('on the operator page a session with the ' + role + ' role ' + (linked ? 'has a link back to Setup at the top' : 'has no link to Setup'), await pg.evaluate((want) => { const l = document.getElementById('setuplink'); return want ? (!l.hidden && l.offsetWidth > 0 && l.getAttribute('href') === '/admin' && l.getBoundingClientRect().top < 120 && l.textContent === 'Setup') : (l.hidden && l.offsetWidth === 0); }, linked));
      await pg.close();
    }
    // Moving back and forth between Setup and the operator page must not hang: a browser keeps six connections to one
    // address, and a page that was left used to keep its live stream open until all six were taken.
    {
      const nav = await open(); nav.setDefaultNavigationTimeout(T(8000));
      await nav.goto(B + '/operator'); await nav.waitForSelector('#list .cue');
      let stuck = 0, done = 0;
      for (let i = 0; i < 16 && stuck < 2; i++) {
        const toOp = i % 2 === 1;
        try {
          await Promise.all([nav.waitForNavigation({ waitUntil: 'domcontentloaded', timeout: T(8000) }), nav.evaluate((sel) => document.querySelector(sel).click(), toOp ? 'header a[href="/operator"]' : '#setuplink')]);
          await nav.waitForSelector(toOp ? '#list .cue' : '#steps button', { timeout: T(8000) }); done++;
        } catch (e) { stuck++; await nav.goto(B + (toOp ? '/admin' : '/operator')).catch(() => {}); }
      }
      check('moving quickly back and forth between the operator page and Setup never hangs', stuck === 0 && done === 16, { stuck, done });
      await nav.close();
    }
    section('Settings and security');
    const guard = await open(); await guard.goto(B + '/admin'); await guard.waitForSelector('#pw'); await guard.type('#pw', ADMIN); await guard.keyboard.press('Enter'); await guard.waitForSelector('#steps button');
    await click(guard, 'Settings'); await sleep(600);
    const tab = async (id) => { await guard.evaluate((i) => document.getElementById('settab-' + i).click(), id); await sleep(150); };
    check('Settings has five tabs, with Close in view', await guard.evaluate(() => { const t = [...document.querySelectorAll('#dlg [role=tab]')].map((x) => x.textContent); const c = document.getElementById('dlgok').getBoundingClientRect(); return t.join('|') === 'Audience screen|Network|Security|Certificate|About' && c.bottom <= innerHeight && c.top >= 0; }));
    await tab('abt');
    {
      const sizes = [];
      for (const id of ['aud', 'net', 'sec', 'abt']) { await tab(id); sizes.push(await guard.evaluate(() => { const b = document.getElementById('dlgbody'); return { dlg: Math.round(document.getElementById('dlg').getBoundingClientRect().height), scrolls: b.scrollHeight > b.clientHeight + 1 }; })); }
      check('the Settings window is as tall as its tallest tab, so no tab scrolls and it never changes size', sizes.every((z) => !z.scrolls) && sizes.every((z) => z.dlg === sizes[0].dlg), sizes);
    }
    await guard.focus('#settab-abt'); await guard.keyboard.press('ArrowLeft'); await sleep(150);
    check('the arrow keys move between the tabs (and the selected tab says so)', await guard.evaluate(() => document.activeElement.id === 'settab-cert' && document.activeElement.getAttribute('aria-selected') === 'true' && document.getElementById('setpanel-cert').classList.contains('inactive') === false && document.getElementById('setpanel-abt').classList.contains('inactive')));
    await tab('abt');
    check('Settings tells you Setup is open to the whole network', (await guard.evaluate(() => document.getElementById('dlg').innerText)).includes('any computer on the network'));
    await tab('aud');
    check('the Audience screen tab has no network name (it is in the Network tab)', await guard.evaluate(() => !/Name on the network/.test(document.getElementById('setpanel-aud').innerText) && !document.querySelector('#setpanel-aud input[aria-label="Announce this computer by name"]')));
    check('Settings has a Standby screen section with a message and a logo', await guard.evaluate(() => { const d = document.getElementById('dlg').innerText; return /Standby screen/.test(d) && !!document.querySelector('#dlg input[aria-label="Standby message"]') && !!document.querySelector('#dlg input[type=file]'); }));
    await guard.type('#dlg input[aria-label="Standby message"]', 'Welcome to the Little Theatre');
    await guard.evaluate(() => [...document.querySelectorAll('#dlg button')].find((x) => x.textContent === 'Save message').click()); await sleep(600);
    const png = path.join(os.tmpdir(), 'oc-logo-' + process.pid + '.png');
    fs.writeFileSync(png, Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==', 'base64'));
    await (await guard.$('#dlg input[type=file]')).uploadFile(png); await sleep(900);
    try { fs.unlinkSync(png); } catch (e) { /* not important */ }
    const sbs = (await api(guard, 'GET', '/api/admin/settings')).body.standby;
    check('the standby message and logo are saved (venue-level)', sbs.message === 'Welcome to the Little Theatre' && /^\/standby-logo\?v=/.test(sbs.logo || ''), sbs);
    check('and the logo is public, for the audience page', await guard.evaluate(async (u) => { const r = await fetch(u, { credentials: 'omit' }); return r.ok && /image\/png/.test(r.headers.get('Content-Type')); }, sbs.logo));
    await tab('abt');
    check('Settings offers the log detail setting', await guard.evaluate(() => !!document.querySelector('#dlg select[aria-label="Log detail"]')));
    check('the Settings window has no Recent log button (it is in the Troubleshoot panel)', await guard.evaluate(() => !document.getElementById('dlgextra') || document.getElementById('dlgextra').hidden));
    check('and its only footer button just closes it', await guard.evaluate(() => document.getElementById('dlgok').textContent.trim() === 'Close' && document.getElementById('dlgcancel').hidden));
    // the operator password can be shown and hidden again
    await tab('sec');
    await guard.evaluate(() => [...document.querySelectorAll('#dlg button')].find((x) => x.textContent.includes('Show the current operator password')).click()); await sleep(500);
    check('showing the operator password turns the button into Hide', await guard.evaluate(() => /Operator password: /.test(document.getElementById('dlg').innerText) && [...document.querySelectorAll('#dlg button')].some((x) => x.textContent.includes('Hide the operator password'))));
    await guard.evaluate(() => [...document.querySelectorAll('#dlg button')].find((x) => x.textContent.includes('Hide the operator password')).click()); await sleep(300);
    check('and Hide hides it again', await guard.evaluate(() => !/Operator password: /.test(document.getElementById('dlg').innerText)));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    await step(guard, 'Share'); await sleep(900);
    check('the Share step lists the .local name only as a plain link row, with no on/off box or Rename button of its own', await guard.evaluate(() => !document.querySelector('.urls input[aria-label="Announce this computer by name"]') && ![...document.querySelectorAll('.urls button')].some((b) => /Rename/.test(b.textContent))));
    await guard.evaluate(() => [...document.querySelectorAll('#panel a')].find((a) => /Choose which network interfaces are used/.test(a.textContent)).click()); await guard.waitForSelector('#setpanel-net:not(.inactive)'); await sleep(200);
    check('the Network tab has the .local name: a Use a name box, the name and an Apply button', await guard.evaluate(() => { const p = document.getElementById('setpanel-net'); return /Name on the network/.test(p.innerText) && !!p.querySelector('input[aria-label="Announce this computer by name"]') && !!p.querySelector('input[aria-label="Name"]') && [...p.querySelectorAll('button')].some((b) => b.textContent === 'Apply'); }));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    check('the addresses sit to the right of the QR code, and the projector link lines up with them', await guard.evaluate(() => { const q = document.querySelector('.qrrow .qr').getBoundingClientRect(), u = document.querySelector('.urls').getBoundingClientRect(); const links = [...document.querySelectorAll('.urls a')]; const lefts = links.map((a) => Math.round(a.getBoundingClientRect().left)); return u.left >= q.right - 1 && Math.abs(u.top - q.top) < 40 && links.length >= 2 && Math.max(...lefts) - Math.min(...lefts) <= 1; }));
    check('on the Share step each address is a link that opens a new window', await guard.evaluate(() => { const l = [...document.querySelectorAll('.urls label a')]; return l.length > 0 && l.every((x) => x.target === '_blank' && /noopener/.test(x.rel) && /^https?:\/\//.test(x.href)); }));
    check('the Share step has only a Troubleshoot link, not the log, support bundle and help links (they are in the Troubleshoot panel)', await guard.evaluate(() => { const t = document.getElementById('panel').innerText; return /Something not working\? Troubleshoot/.test(t) && !/Recent log|Download a support bundle|Troubleshooting help/.test(t); }));
    await click(guard, 'Troubleshoot'); await guard.waitForSelector('dialog[open] ul.checks li'); await sleep(300);
    await guard.evaluate(() => [...document.querySelectorAll('dialog[open] button')].find((x) => x.textContent.includes('Open the full log')).click()); await guard.waitForFunction(() => { const p = document.querySelector('pre.logview'); return p && p.textContent.includes('signed in'); }, { timeout: 8000 }).catch(() => {});
    check('the recent log can be read in Setup', await guard.evaluate(() => /INFO .*signed in/.test(document.querySelector('pre.logview').textContent)));
    await guard.select('dialog select[aria-label="Show"]', 'warn'); await sleep(600);
    check('and narrowed to problems', await guard.evaluate(() => { const t = document.querySelector('pre.logview').textContent; return /WARN/.test(t) && !/INFO/.test(t); }));
    check('it never contains a password', !(await guard.evaluate(() => document.querySelector('pre.logview').textContent)).includes(ADMIN));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    // Troubleshoot: a header button with the checks, the recent problems, and a count of the new ones
    await click(guard, 'Troubleshoot'); await guard.waitForSelector('dialog[open] ul.checks li'); await sleep(300);
    const trouble = await guard.evaluate(() => ({ n: document.querySelectorAll('dialog[open] ul.checks li').length, text: document.querySelector('dialog[open] .trouble').innerText, log: document.querySelector('dialog[open] pre.logview').textContent, badge: document.getElementById('troublebadge').hidden }));
    check('the Troubleshoot button opens a list of checks with the facts the program knows', trouble.n >= 6 && /Setup can save its settings/.test(trouble.text) && /connected right now/.test(trouble.text) && /network address/.test(trouble.text), trouble);
    check('it lists the recent warnings and errors, and has the log and the support bundle one click away', /WARN/.test(trouble.log) && /Open the full log/.test(trouble.text) && /Download a support bundle/.test(trouble.text), trouble.log.slice(0, 200));
    check('opening it clears the count on the button', trouble.badge === true);
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    // (the program's own count is checked in the Go tests; here the page is told that more problems were logged)
    await guard.evaluateOnNewDocument(() => { const real = window.fetch; window.fetch = (u, o) => (/troubleshoot\?brief=1/.test(String(u)) ? Promise.resolve(new Response(JSON.stringify({ problems: 4000 }), { status: 200 })) : real(u, o)); });
    await guard.reload(); await guard.waitForSelector('#steps button');
    await guard.waitForFunction(() => !document.getElementById('troublebadge').hidden, { timeout: T(8000) }).catch(() => {});
    check('a new warning puts a count on the Troubleshoot button', await guard.evaluate(() => { const b = document.getElementById('troublebadge'); return !b.hidden && /^\d+$/.test(b.textContent) && /new warning/.test(document.getElementById('troublebtn').getAttribute('aria-label')); }));
    await step(guard, 'Share'); await sleep(700);
    // Working on a show is not the same as it being the active show
    {
      const list = (await api(guard, 'GET', '/api/admin/shows')).body, act = list.shows.find((x) => x.id === list.activeId), other = list.shows.find((x) => x.id !== list.activeId && !x.archived);
      if (act && other) {
        const work = async (name) => { await step(guard, 'Shows'); await sleep(500); await guard.evaluate((n) => { const c = [...document.querySelectorAll('.card.show')].find((x) => x.querySelector('.name').firstChild.textContent.trim() === n); [...c.querySelectorAll('button')].find((b) => /Work on this show|Continue with this show/.test(b.textContent)).click(); }, name); await sleep(600); await step(guard, 'Share'); await sleep(900); };
        await work(other.name);
        const pill = await guard.evaluate(() => document.getElementById('working').textContent);
        check('the header names the show being worked on and the active one when they differ', pill.includes('Working on: ' + other.name) && pill.includes('Active: ' + act.name), pill);
        const box = await guard.evaluate(() => ({ btn: document.querySelector('.startbox .primary').textContent, hint: document.querySelector('.startbox .hint').textContent }));
        check('the Share step says the button makes this show the active one, and what the audience gets until then', /^Open the Run page and make .+ the active show$/.test(box.btn) && box.hint.includes(act.name) && /active show|on air now/.test(box.hint), box);
        await work(act.name);
        const pill2 = await guard.evaluate(() => document.getElementById('working').textContent);
        check('and it says so when the show being worked on is the active one', pill2.includes('Working on: ' + act.name) && pill2.includes('(the active show)') && !pill2.includes('Active:'), pill2);
      }
    }
    await click(guard, 'Settings'); await sleep(600);
    await guard.evaluate(() => document.getElementById('settab-sec').click()); await sleep(150);
    check('Settings → Security has a Certificate section that says where HTTPS gets its certificate', await guard.evaluate(() => { const b = document.querySelector('.certbox'); return !!b && /HTTPS/.test(b.textContent) && b.closest('section').querySelector('h3').textContent === 'Certificate'; }));
    const inputs = await guard.$$('#dlg input[type=password]'); await inputs[1].type('new-operator-pw');
    await guard.evaluate(() => [...document.querySelectorAll('#dlg button')].find((x) => x.textContent.trim() === 'Change passwords').click()); await sleep(700);
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    const probe = (pw) => guard.evaluate(async (p) => (await fetch('/api/login', { method: 'POST', credentials: 'omit', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ role: 'operator', password: p }) })).status, pw);
    check('the old operator password stops working', (await probe(OPERATOR)) === 401);
    check('the new operator password works', (await probe('new-operator-pw')) === 200);
    const forged = await guard.evaluate(async () => (await fetch('/api/admin/shows', { method: 'POST', headers: { 'Content-Type': 'application/json', 'Sec-Fetch-Site': 'cross-site' }, body: '{"name":"x"}' })).status);
    check('a request that claims to be cross-site is refused', forged === 403 || forged === 201 /* browsers forbid setting Sec-Fetch-Site; the Go tests cover it */);

    // ------------------------------------------------------------------ keeping the screen on
    section('Audience: the screen stays on where the Wake Lock API is missing');
    // A separate browser: Chrome allows six connections per site, and the pages above hold them all open.
    const b2 = await launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
    const nolock = await b2.newPage(); nolock.setDefaultTimeout(T(8000));
    await nolock.setViewport({ width: 420, height: 760 });
    nolock.on('pageerror', (e) => errs.push(e.message));
    await nolock.evaluateOnNewDocument(() => { delete Navigator.prototype.wakeLock; }); // as on any browser over plain http
    await nolock.goto(B + '/', { waitUntil: 'domcontentloaded' }); await nolock.waitForSelector('video');
    await nolock.waitForFunction(() => document.documentElement.dataset.awake === 'video', { timeout: 6000 }).catch(() => {});
    const vid = await nolock.evaluate(() => { const v = document.querySelector('video'); return v ? { playing: !v.paused, muted: v.muted, loop: v.loop, src: v.currentSrc.split('/').pop(), tag: document.documentElement.dataset.awake, ready: v.readyState } : null; });
    check('a hidden looping video is started instead', vid && vid.playing && vid.muted && vid.loop && /^nosleep\.(mp4|webm)$/.test(vid.src) && vid.tag === 'video', vid);
    check('it does not cover or disturb the captions', await nolock.evaluate(() => { const r = document.querySelector('video').getBoundingClientRect(); return r.width <= 2 && getComputedStyle(document.querySelector('video')).pointerEvents === 'none'; }));
    await nolock.click('#gear'); await nolock.click('#awake');
    await nolock.waitForFunction(() => document.documentElement.dataset.awake === 'off', { timeout: 4000 }).catch(() => {});
    check('a device can turn "screen stays on" off', await nolock.evaluate(() => document.querySelector('video').paused && document.getElementById('awake').getAttribute('aria-pressed') === 'false'));
    await nolock.reload(); await nolock.waitForSelector('#awake'); await sleep(600);
    check('and the choice is remembered', await nolock.evaluate(() => document.documentElement.dataset.awake === 'off' && !document.querySelector('video')));
    await nolock.click('#gear'); await nolock.click('#awake');
    await nolock.waitForFunction(() => document.documentElement.dataset.awake === 'video', { timeout: 6000 }).catch(() => {});
    check('turning it back on restarts the video', await nolock.evaluate(() => document.documentElement.dataset.awake === 'video' && !document.querySelector('video').paused));
    await nolock.close(); await b2.close();

    // ------------------------------------------------------------------ archive
    section('Setup: archive a show and bring it back');
    const arch = await open();
    await arch.goto(B + '/admin#show'); await arch.waitForSelector('#steps button');
    const made = await api(arch, 'POST', '/api/admin/shows', { name: 'Last year' });
    {
      const ah = await headerOrder(arch, ['#viewers', 'header select.uilang', '#settingsbtn', 'header a[href="/operator"]', '#helpbtn', '#logout']);
      check('Setup header, right side: devices, language, Settings, Run, Help, Sign out, connection', ah.ok && ah.pts.filter((x) => x.includes('hidden')).length === 0, ah);
      const adLook = await headerLook(arch);
      check('the ON AIR pill is the same red on Setup and the operator page', (await liveRed(arch)) === opRed && opRed === 'rgb(179, 38, 30)', { setup: await liveRed(arch), operator: opRed });
      check('the audience count leaves out a language nobody is using (Setup too)', JSON.stringify(await countText(arch)) === JSON.stringify(['👥 3 (en 3)', '👥 0', '👥 5 (en 3 · es 2)', '👥 2']));
      await arch.evaluate(() => document.getElementById('viewers').click()); await arch.waitForSelector('dialog[open] table.clients', { timeout: T(5000) }).catch(() => {});
      check('clicking the count on Setup lists the connected devices too', await arch.evaluate(() => { const d = document.getElementById('dlg'); const rows = [...d.querySelectorAll('table.clients tr')].slice(1); return d.open && /Connected devices/.test(d.innerText) && rows.length >= 2; }));
      const adDlg = await dlgLook(arch, '#dlg');
      const colsAlike = adDlg.cols.split(',').every((w, k) => Math.abs(Number(w) - Number(opDlg.cols.split(',')[k])) <= 4); // (Setup keeps a few pixels for a scrollbar)
      check('the Connected devices window looks the same on both pages (size, type, spacing)', JSON.stringify({ ...adDlg, cols: '' }) === JSON.stringify({ ...opDlg, cols: '' }) && colsAlike, { operator: opDlg, setup: adDlg });
      await arch.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
      check('the two headers use the same type, spacing and control sizes', JSON.stringify(adLook) === JSON.stringify(opLook), { operator: opLook, setup: adLook });
      check('neither header has a status light: the red bar is the only connection indicator', await arch.evaluate(() => !document.getElementById('dot')) && await op.evaluate(() => !document.getElementById('dot')));
    }
    check('the ON AIR badge in the show list is the same red as the ON AIR pills', await arch.evaluate(() => { const e = document.createElement('span'); e.className = 'badge live'; document.querySelector('.card.show .name').appendChild(e); const c = getComputedStyle(e).backgroundColor; e.remove(); return c === 'rgb(179, 38, 30)'; }));
    check('the name box under Add another show is a little larger than a normal box', await arch.evaluate(() => { const i = document.querySelector('input.newname'), r = i.getBoundingClientRect(); return parseFloat(getComputedStyle(i).fontSize) >= 17 && r.height >= 44 && r.width >= 300; }));
    // the "More ▾" menu on a show card
    const menuState = () => arch.evaluate(() => { const m = document.querySelector('.card.show .moremenu .menu'); return { open: !m.hidden, focus: document.activeElement && document.activeElement.textContent.trim().slice(0, 12), expanded: m.parentNode.querySelector('button.more').getAttribute('aria-expanded') }; });
    await arch.focus('.card.show button.more'); await arch.keyboard.press('Enter'); await sleep(150);
    let ms = await menuState();
    check('More opens a menu from the keyboard, with the first action focused', ms.open && ms.expanded === 'true' && /^Rename/.test(ms.focus), ms);
    await arch.keyboard.press('ArrowDown'); await arch.keyboard.press('ArrowDown'); await sleep(100);
    check('the arrow keys move through the actions', /^Export/.test((await menuState()).focus), await menuState());
    await arch.keyboard.press('Escape'); await sleep(150); ms = await menuState();
    check('Esc closes the menu and returns to the More button', !ms.open && ms.expanded === 'false' && await arch.evaluate(() => document.activeElement.classList.contains('more')), ms);
    await arch.click('.card.show button.more'); await sleep(200);
    const rp = await arch.evaluate(() => { const e = [...document.querySelectorAll('.card.show .menu .item')].find((x) => /^Rename/.test(x.textContent)), r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2, w: r.width }; });
    await arch.mouse.click(rp.x, rp.y); await sleep(300);
    check('with the mouse, More shows Rename… on screen and clicking it opens its dialog', rp.w > 50 && await arch.evaluate(() => document.getElementById('dlg').hasAttribute('open') && document.querySelector('.card.show .menu').hidden));
    await arch.keyboard.press('Escape'); await sleep(250);
    await arch.click('.card.show button.more'); await sleep(150);
    await arch.click('#panel h2'); await sleep(150);
    check('a click elsewhere closes the menu', await arch.evaluate(() => document.querySelector('.card.show .menu').hidden));
    check('the menu lists Rename, Duplicate, Export, Archive and Delete, Delete last and in red', await arch.evaluate(() => { const c = document.querySelector('.card.show'), it = [...c.querySelectorAll('.menu .item')].map((e) => e.textContent.trim()), d = c.querySelector('.menu .item.danger'); return it.length === 5 && /^Rename/.test(it[0]) && /^Delete/.test(it[4]) && !!d && getComputedStyle(d).color !== getComputedStyle(document.body).color; }));
    // Export is fetched and saved by the page, so a failure can be explained (a plain link leaves the browser's own "download failed").
    const clickExport = () => arch.evaluate(() => { const c = document.querySelector('.card.show'); c.querySelector('button.more').click(); [...c.querySelectorAll('.menu .item')].find((x) => /^Export/.test(x.textContent)).click(); });
    const toastText = () => arch.evaluate(() => document.getElementById('toast').textContent);
    await arch.evaluate(() => { window.__saved = []; const click = HTMLAnchorElement.prototype.click; HTMLAnchorElement.prototype.click = function () { if (this.download) { window.__saved.push(this.download); return; } return click.call(this); }; });
    await clickExport(); await arch.waitForFunction(() => window.__saved.length === 1, { timeout: 10000 }).catch(() => {});
    check('Export saves the show as a .zip', await arch.evaluate(() => window.__saved.length === 1 && /\.zip$/.test(window.__saved[0])), await arch.evaluate(() => window.__saved));
    await arch.evaluate(() => { window.__fetch = window.fetch; window.fetch = () => Promise.reject(new TypeError('offline')); });
    await clickExport(); await sleep(400);
    check('Export while the program does not answer says so', /Could not reach the program/.test(await toastText()), await toastText());
    await arch.evaluate(() => { window.fetch = () => Promise.resolve(new Response(JSON.stringify({ error: 'The show could not be exported. Try again; if it keeps happening, look at the log.' }), { status: 500, headers: { 'Content-Type': 'application/json' } })); });
    await clickExport(); await sleep(400);
    check('Export that fails on the program says why', /could not be exported/.test(await toastText()), await toastText());
    await arch.evaluate(() => { window.fetch = window.__fetch; });
    await arch.evaluate(() => { const m = document.querySelector('.card.show .menu'); if (!m.hidden) document.querySelector('.card.show button.more').click(); });
    // A step that cannot reach the program says so (api() no longer leaves a blank panel).
    await arch.evaluate(() => { window.__fetch2 = window.fetch; window.fetch = () => Promise.reject(new TypeError('offline')); document.getElementById('toast').textContent = ''; document.querySelectorAll('#steps button')[1].click(); });
    await sleep(500);
    check('a step opened while the program does not answer says so', /Could not reach the program/.test(await toastText()), await toastText());
    await arch.evaluate(() => { window.fetch = window.__fetch2; document.querySelectorAll('#steps button')[0].click(); });
    await sleep(500);
    check('an action that cannot be done is listed with its reason', await arch.evaluate(() => [...document.querySelectorAll('.card.show .menu .item[aria-disabled=true] small')].every((x) => x.textContent.length > 10)));
    check('Duplicate… is named for the dialog it opens', await arch.evaluate(() => [...document.querySelectorAll('.card.show button')].some((b) => b.textContent === 'Duplicate…') && ![...document.querySelectorAll('.card.show button')].some((b) => b.textContent === 'Duplicate')));
    check('a show to archive is created', made.status === 201 || made.status === 200, made);
    await arch.reload(); await arch.waitForFunction(() => [...document.querySelectorAll('.card.show')].some((x) => x.textContent.includes('Last year')), { timeout: 15000 });
    const onCard = (label) => arch.evaluate((l) => { const c = [...document.querySelectorAll('.card.show')].find((x) => x.textContent.includes('Last year')); [...c.querySelectorAll('button')].find((b) => b.textContent.startsWith(l)).click(); }, label);
    // Earlier versions: every save keeps the one it replaced.
    const lastId = (await api(arch, 'GET', '/api/admin/shows')).body.shows.find((x) => x.name === 'Last year').id;
    await api(arch, 'PUT', '/api/admin/shows/' + lastId + '/showfile', { text: '[settings]\ntitle: Last year\n[/settings]\nFirst draft' });
    await api(arch, 'PUT', '/api/admin/shows/' + lastId + '/showfile', { text: '[settings]\ntitle: Last year\n[/settings]\nSecond draft' });
    await onCard('Work on this show'); await arch.waitForSelector('.editor textarea');
    await click(arch, 'Earlier versions'); await arch.waitForSelector('ul.versions li button');
    await arch.evaluate(() => document.querySelector('ul.versions li button').click());
    await arch.waitForFunction(() => document.querySelector('.editor textarea').value.includes('First draft'), { timeout: 8000 }).catch(() => {});
    check('an earlier version can be loaded into the editor', await arch.evaluate(() => document.querySelector('.editor textarea').value.includes('First draft')));
    await click(arch, 'Earlier versions'); await arch.waitForSelector('dialog[open] ul.versions li button');
    await arch.evaluate(() => [...document.querySelectorAll('ul.versions li button')].filter((x) => x.textContent === 'Compare').pop().click());
    await arch.waitForSelector('dialog[open] pre.diff', { timeout: T(8000) }).catch(() => {});
    check('Compare shows what loading a version would change', await arch.evaluate(() => { const d = document.querySelector('pre.diff'); return !!d && /First draft/.test((d.querySelector('.d-del') || {}).textContent || ''); }));
    await arch.type('dialog[open] ul.versions input.vname', 'Dress rehearsal'); await arch.keyboard.press('Tab'); await sleep(500);
    const hist = JSON.stringify((await api(arch, 'GET', '/api/admin/shows/' + lastId + '/showfile/history')).body); check('a version can be given a name', hist.includes('Dress rehearsal'), hist);
    await arch.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    // Characters: every speaker name, most used first, with a misspelling marked.
    await arch.evaluate(() => { const t = document.querySelector('.editor textarea'); t.value = 'ALICE: One\nALICE & BOB: Two\nALICE: Three\nALCE: Four\nBOB: Five'; t.dispatchEvent(new Event('input')); });
    await sleep(900);
    await click(arch, 'Characters'); await arch.waitForSelector('dialog[open] ul.chars li');
    const chars = await arch.evaluate(() => [...document.querySelectorAll('dialog[open] ul.chars li')].map((li) => li.querySelector('.charname').textContent + '|' + li.textContent));
    check('Characters lists each name once, the most used first, splitting a group', chars.length === 3 && chars[0].startsWith('ALICE|') && chars[1].startsWith('BOB|') && chars[2].startsWith('ALCE|'), chars);
    check('Characters marks a name one letter off from a more used name', /Did you mean ALICE/.test(chars[2]) && !/Did you mean/.test(chars[0]), chars);
    await arch.evaluate(() => [...document.querySelectorAll('dialog[open] ul.chars .charname')].pop().click()); await sleep(300);
    check('clicking a character goes to its first line in the editor', await arch.evaluate(() => { const t = document.querySelector('.editor textarea'); return !document.getElementById('dlg').open && t.value.slice(t.selectionStart, t.selectionEnd) === 'ALCE: Four'; }));
    check('loading it changes nothing until Save', (await arch.evaluate(() => document.querySelector('.editor textarea').value)) !== (await api(arch, 'GET', '/api/admin/shows/' + lastId + '/showfile')).body.text);
    // The Share step no longer puts a show in front of the audience: it opens the operator page, with the show in standby.
    await step(arch, 'Share'); await sleep(T(900));
    check('the Share step ends its list of addresses with a projector link, in a new window, by numeric address', await arch.evaluate(() => { const rows = [...document.querySelectorAll('.urls > .row')]; const p = document.querySelector('.urls a[href*="display=projector"]'); return rows.length >= 2 && !!p && p.target === '_blank' && /\/\?display=projector$/.test(p.getAttribute('href')) && rows[rows.length - 1].contains(p) && !/\.local/.test(p.getAttribute('href')); }));
    check('the Share step offers Open the Run page (the show waits in standby), not Start show', await arch.evaluate(() => { const t = document.getElementById('panel').innerText; return /Open the Run page/.test(t) && !/Start show/.test(t) && /standby|on air now/.test(t); }));
    check('buttons that are links (Open the Run page, Print a QR sign) are not underlined', await arch.evaluate(() => [...document.querySelectorAll('#panel a.btn')].length >= 1 && [...document.querySelectorAll('#panel a.btn')].every((e) => getComputedStyle(e).textDecorationLine === 'none')));
    check('the Print QR sign button sits right beside Copy address', await arch.evaluate(() => { const c = [...document.querySelectorAll('#panel button')].find((b) => b.textContent === 'Copy address'), p = [...document.querySelectorAll('#panel a.btn')].find((b) => /QR sign/.test(b.textContent)); return !!c && !!p && p.getBoundingClientRect().left - c.getBoundingClientRect().right < 40; }));
    check('the Print QR sign link follows the address that is chosen', await arch.evaluate(() => { const r = [...document.querySelectorAll('.urls input[type=radio]')]; const last = r[r.length - 1]; last.click(); const a = document.getElementById('printsign'); const want = last.closest('.row').querySelector('a').href; return r.length >= 2 && new URL(a.href).searchParams.get('url') === want; }));
    check('the projector link is a numeric address with its own radio button, so it can be copied', await arch.evaluate(() => { const r = document.querySelector('.projrow input[type=radio]'), a = document.querySelector('.projrow a'); return !!r && !!a && /display=projector/.test(a.href) && !/\.local/.test(a.href); }));
    // REQ-069: choosing which network interfaces are used (a setting of this computer: Settings > Network, not the Share step)
    {
      const netOpen = async () => { await arch.evaluate(() => [...document.querySelectorAll('#panel a')].find((a) => /Choose which network interfaces are used/.test(a.textContent)).click()); await arch.waitForFunction(() => document.getElementById('dlg').hasAttribute('open') && document.querySelector('#setpanel-net:not(.inactive) .ifrow')); };
      const netClose = async () => { await arch.click('#dlgok'); await sleep(T(900)); };
      check('the Share step has no network controls of its own, only a link to Settings > Network', await arch.evaluate(() => [...document.querySelectorAll('#panel button')].map((b) => b.textContent).filter((t) => /interface|Choose addresses/i.test(t)).length === 0 && [...document.querySelectorAll('#panel a')].some((a) => /Choose which network interfaces are used/.test(a.textContent))));
      await netOpen();
      check('the link opens Settings on the Network tab, which lists this computer\'s interfaces with a tick each (all ticked until a choice is made), their addresses and whether they are used', await arch.evaluate(() => { const r = [...document.querySelectorAll('#setpanel-net .ifrow')]; return document.getElementById('settab-net').getAttribute('aria-selected') === 'true' && r.length > 0 && r.every((x) => x.querySelector('input').checked && /\d+\.\d+\.\d+\.\d+/.test(x.textContent)); }));
      check('there is a box to refuse connections to interfaces that are not ticked, off by default', await arch.evaluate(() => { const b = document.querySelector('.addrrestrict input'); return !!b && !b.checked; }));
      await netClose();
    }
    // REQ-069: two interfaces on one network. The program's answers are faked in the page (this computer may have one interface): a cable and Wi-Fi on 192.168.1.x.
    {
      await arch.evaluate(() => {
        const orig = window.fetch; window.__realFetch = orig; let ticked = ['Wi-Fi', 'Ethernet'], restrict = false;
        const list = () => { const cable = ticked.includes('Ethernet'), wifi = ticked.includes('Wi-Fi'); return [
          { name: 'Wi-Fi', kind: 'wifi', addresses: ['192.168.1.138'], selected: wifi, used: wifi && !cable, instead: wifi && cable ? 'Ethernet' : '', insteadIp: wifi && cable ? '192.168.1.185' : '', current: false },
          { name: 'Ethernet', kind: 'ethernet', addresses: ['192.168.1.185'], selected: cable, used: cable, instead: '', insteadIp: '', current: false }]; };
        const json = (o) => new Response(JSON.stringify(o), { status: 200, headers: { 'Content-Type': 'application/json' } });
        window.fetch = async function (url, opt) {
          const u = String(url), m = (opt && opt.method) || 'GET';
          if (u === '/api/admin/settings' && m === 'GET') { const j = await (await orig.call(window, url, opt)).json(); j.interfaces = list(); j.network = { interfaces: [], restrictListening: restrict }; return json(j); }
          if (u === '/api/admin/settings/network' && m === 'PUT') { const b = JSON.parse(opt.body); ticked = b.interfaces; restrict = b.restrictListening; window.__lastPut = b; return json({ saved: true, interfaces: list() }); }
          return orig.call(window, url, opt);
        };
      });
      await arch.evaluate(() => document.getElementById('settingsbtn').click()); await arch.waitForSelector('#settab-net');
      await arch.evaluate(() => document.getElementById('settab-net').click()); await sleep(200);
      const rows = () => arch.evaluate(() => [...document.querySelectorAll('#setpanel-net .ifrow')].map((r) => (r.querySelector('input').checked ? '[x] ' : '[ ] ') + r.textContent.replace(/\s+/g, ' ').trim()));
      const tick = async (name) => { await arch.evaluate((n) => [...document.querySelectorAll('#setpanel-net .ifrow')].find((r) => r.querySelector('b').textContent === n).querySelector('input').click(), name); await sleep(T(500)); };
      let r = await rows();
      check('two interfaces on one network: both ticked, the cable is used and Wi-Fi says it is not and why', /^\[x\] Wi-Fi \(Wi-Fi\)192\.168\.1\.138 — not used: same network as Ethernet \(192\.168\.1\.185\)$/.test(r[0]) && /^\[x\] Ethernet \(Ethernet\)192\.168\.1\.185 — used$/.test(r[1]), r);
      await tick('Ethernet'); r = await rows();
      check('unticking the cable saves at once and Wi-Fi is then the one used', /^\[x\] Wi-Fi .*— used$/.test(r[0]) && /^\[ \] Ethernet .*— not used$/.test(r[1]) && (await arch.evaluate(() => window.__lastPut.interfaces.join())) === 'Wi-Fi', r);
      await tick('Ethernet'); r = await rows();
      check('ticking it again toggles back', /not used: same network as Ethernet/.test(r[0]) && /^\[x\] Ethernet .*— used$/.test(r[1]), r);
      await arch.evaluate(() => document.querySelector('.addrrestrict input').click()); await sleep(T(500));
      check('the refuse-connections box saves at once, with the ticks', await arch.evaluate(() => window.__lastPut.restrictListening === true && window.__lastPut.interfaces.length === 2));
      await arch.evaluate(() => document.querySelector('.addrrestrict input').click()); await sleep(T(300));
      await arch.click('#dlgok'); await sleep(T(500));
      await arch.evaluate(() => { window.fetch = window.__realFetch; });
    }
    await step(arch, 'Show file'); await sleep(T(600));
    check('the Show file step has no second Help button (the header Help opens the same topic)', await arch.evaluate(() => ![...document.querySelectorAll('#panel button')].some((b) => /help/i.test(b.textContent))));
    // The editor grows with the window (no maximum width), and has a minimum.
    {
      const edW = () => arch.evaluate(() => { const e = document.querySelector('.editor'), r = e.getBoundingClientRect(); return { w: Math.round(r.width), h: Math.round(r.height), vw: innerWidth, scroll: document.documentElement.scrollWidth }; });
      await arch.setViewport({ width: 1300, height: 950 }); await sleep(300); const w1300 = await edW();
      await arch.setViewport({ width: 2000, height: 1100 }); await sleep(300); const w2000 = await edW();
      await arch.setViewport({ width: 1000, height: 700 }); await sleep(300); const w1000 = await edW();
      await arch.setViewport({ width: 1300, height: 950 }); await sleep(300);
      check('the Show file editor grows as the window gets wider (no maximum)', w2000.w >= w1300.w + 600 && w1300.w > 800, { w1000, w1300, w2000 });
      check('and taller', w2000.h > w1300.h, { w1300, w2000 });
      check('with a minimum, and the page never scrolls sideways', w1000.w >= 440 && w1000.scroll <= w1000.vw && w2000.scroll <= w2000.vw, { w1000, w2000 });
    }
    // Offline editing: download the plain show file, edit it elsewhere, open it again and save.
    check('the Show file step offers the show file as a plain download', await arch.evaluate(() => { const l = [...document.querySelectorAll('a')].find((x) => x.textContent === 'Download the show file'); return !!l && /\/showfile\/download$/.test(l.getAttribute('href')) && /\.md$/.test(l.getAttribute('download')); }));
    const dl = await arch.evaluate(async (id) => { const r = await fetch('/api/admin/shows/' + id + '/showfile/download'); const bytes = new Uint8Array(await r.clone().arrayBuffer()); const t = await r.text(); return { ok: r.ok, disp: r.headers.get('Content-Disposition'), bom: bytes[0] === 0xEF && bytes[1] === 0xBB && bytes[2] === 0xBF, crlf: t.includes('\r\n'), text: t }; }, lastId);
    check('the download is a plain .md file that Windows editors open correctly', dl.ok && /attachment; filename=".*\.md"/.test(dl.disp) && dl.bom && dl.crlf && dl.text.includes('Second draft'), dl.disp);
    const offline = path.join(os.tmpdir(), 'oc-offline-' + process.pid + '.txt');
    fs.writeFileSync(offline, Buffer.from('[settings]\r\ntitle: Last year\r\n[/settings]\r\nCAF\xc9 edited offline\r\n', 'latin1'));
    check('the file chooser is behind an Upload button like the other buttons, with the save state and Save together', await arch.evaluate(() => {
      const b = [...document.querySelectorAll('#panel button')].find((x) => x.textContent === 'Upload a show file…'), f = document.querySelector('#panel input[type=file]');
      const st = [...document.querySelectorAll('#panel .savebar span')][0], sv = [...document.querySelectorAll('#panel .savebar button')][0];
      return !!b && f.hidden && b.getBoundingClientRect().height === document.querySelector('#panel a.btn').getBoundingClientRect().height && !!st && !!sv && Math.abs(st.getBoundingClientRect().top - sv.getBoundingClientRect().top) < 30;
    }));
    await (await arch.$('input[type=file]')).uploadFile(offline);
    await arch.waitForFunction(() => document.querySelector('.editor textarea').value.includes('edited offline'), { timeout: T(8000) }).catch(() => {});
    check('a file edited elsewhere (even saved as Windows ANSI) opens in the editor with its accents', await arch.evaluate(() => document.querySelector('.editor textarea').value.includes('CAFÉ edited offline')));
    await click(arch, 'Save'); await sleep(900);
    const saved = (await api(arch, 'GET', '/api/admin/shows/' + lastId + '/showfile')).body.text;
    check('saving makes it the show file', saved.includes('CAFÉ edited offline'), saved);
    try { fs.unlinkSync(offline); } catch (e) { /* not important */ }
    await step(arch, 'Shows'); await arch.waitForSelector('.card.show');
    await onCard('Archive'); await sleep(700);
    check('an archived show moves to the archive list', (await arch.evaluate(() => document.getElementById('panel').innerText)).includes('Archived shows (1)'));
    check('the archive cannot be activated from the list (it offers Restore)', await arch.evaluate(() => [...document.querySelectorAll('details.archive button')].some((b) => b.textContent === 'Restore')));
    await arch.evaluate(() => document.querySelector('details.archive').open = true);
    await onCard('Restore'); await sleep(700);
    check('a restored show is back in the main list', !(await arch.evaluate(() => document.getElementById('panel').innerText)).includes('Archived shows'));
    await onCard('Delete'); await sleep(400);
    await arch.evaluate(() => document.getElementById('dlgok').click()); await sleep(700);
    check('a show can be deleted', !(await arch.evaluate(() => document.getElementById('panel').innerText)).includes('Last year'));
    await arch.close();

    // ------------------------------------------------------------------ older browsers
    section('Older browsers: dialogs without <dialog>, and a plain message when a page cannot run');
    const b3 = await launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
    const old = async (script, url) => {
      const p = await b3.newPage(); p.setDefaultTimeout(T(30000)); await p.setViewport({ width: 1100, height: 800 });
      p.on('pageerror', (e) => errs.push(e.message));
      await p.evaluateOnNewDocument(script); await p.goto(B + url, { waitUntil: 'domcontentloaded' }); return p;
    };
    // Typing the password before the page's script has attached its handlers does nothing on a slow computer: retry.
    const signIn = async (p, ready) => {
      for (let attempt = 0; attempt < 4; attempt++) {
        await p.waitForSelector('#pw'); await p.evaluate(() => { document.getElementById('pw').value = ''; });
        await p.type('#pw', ADMIN); await p.keyboard.press('Enter');
        if (await p.waitForSelector(ready, { timeout: 5000 }).then(() => true, () => false)) return;
      }
      throw new Error('could not sign in to ' + ready);
    };
    const noDialog = () => { delete HTMLDialogElement.prototype.showModal; delete Element.prototype.replaceChildren; };
    const oldAdmin = await old(noDialog, '/admin'); await signIn(oldAdmin, '#steps button');
    // Setup may redraw this button while loading; find and click it in the same browser turn.
    const clickOldRename = () => oldAdmin.waitForFunction(() => {
      const button = [...document.querySelectorAll('button')].find((x) => x.textContent.includes('Rename'));
      if (button) button.click();
      return !!button;
    }, { timeout: 15000 });
    await clickOldRename(); await oldAdmin.waitForFunction(() => document.getElementById('dlg').hasAttribute('open'));
    const shown = await oldAdmin.evaluate(() => { const d = document.getElementById('dlg'), r = d.getBoundingClientRect(); return { fixed: getComputedStyle(d).position === 'fixed', w: r.width, onScreen: r.left >= 0 && r.top >= 0 && r.right <= innerWidth && r.bottom <= innerHeight, hasShowModal: typeof HTMLDialogElement !== 'undefined' && !!HTMLDialogElement.prototype.showModal }; });
    check('without <dialog> support the dialog still appears, centred over the page', shown.fixed && shown.w > 200 && shown.onScreen && !shown.hasShowModal, shown);
    await oldAdmin.keyboard.press('Escape'); await sleep(200);
    check('Escape closes it', await oldAdmin.evaluate(() => !document.getElementById('dlg').hasAttribute('open')));
    await clickOldRename(); await oldAdmin.waitForFunction(() => document.getElementById('dlg').hasAttribute('open'));
    await oldAdmin.evaluate(() => { const i = document.querySelector('#dlgbody input'); i.value = 'Renamed in an old browser'; });
    await oldAdmin.click('#dlgok'); await oldAdmin.waitForFunction(() => document.getElementById('panel').innerText.includes('Renamed in an old browser'), { timeout: 8000 }).catch(() => {});
    check('and its buttons work (OK renames the show)', (await oldAdmin.evaluate(() => document.getElementById('panel').innerText)).includes('Renamed in an old browser'));
    await oldAdmin.close();
    const oldOp = await old(noDialog, '/operator'); await signIn(oldOp, '#message');
    await oldOp.waitForSelector('#message', { visible: true });
    await api(oldOp, 'POST', '/api/operator/claim', { takeover: true }); await sleep(800);
    await oldOp.click('#message'); await oldOp.waitForFunction(() => document.getElementById('msgdlg').hasAttribute('open'));
    await oldOp.type('#msgtext', 'Back soon'); await oldOp.click('#msgsend'); await sleep(500);
    check('the operator page works without <dialog> too (Send closes the dialog)', await oldOp.evaluate(() => !document.getElementById('msgdlg').hasAttribute('open') && document.getElementById('msgtext').value === 'Back soon'));
    await oldOp.close();
    await api(op, 'POST', '/api/operator/claim', { takeover: true }); await sleep(500); // give control back to the operator page used below
    const fine = await old(() => {}, '/'); await sleep(500);
    check('a current browser is never shown the "too old" message', !(await fine.evaluate(() => !!document.querySelector('[role=alert]'))));
    await fine.close();
    const tooOld = await old(() => { delete window.EventSource; }, '/');
    check('a browser that cannot run the audience page is told so in plain words', (await tooOld.evaluate(() => (document.querySelector('[role=alert]') || {}).textContent || '')).includes('too old'));
    await tooOld.close();
    const tooOldSetup = await old(() => { CSS.supports = function (a) { return a !== 'inset'; }; }, '/admin');
    check('Setup says so when the browser lacks what it needs', (await tooOldSetup.evaluate(() => (document.querySelector('[role=alert]') || {}).textContent || '')).includes('Safari 14.1'));
    const stillAudience = await old(() => { CSS.supports = function (a) { return a !== 'inset'; }; }, '/');
    await sleep(600);
    check('the audience page does not need that (it has fallbacks)', !(await stillAudience.evaluate(() => !!document.querySelector('[role=alert]'))));
    await tooOldSetup.close(); await stillAudience.close(); await b3.close();

    // ------------------------------------------------------------------ interface languages
    section('Interface languages: Setup, the operator page and Help in the six audience languages');
    const b4 = await launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
    // Visible text that is still English although the language has a translation for it.
    const leaksOf = (p) => p.evaluate(() => {
      const dict = new Set(OCI18n.keys(OCI18n.lang)), out = [];
      const visible = (el) => !!(el.offsetWidth || el.offsetHeight || el.getClientRects().length);
      const check = (text, where) => { const t = text.replace(/\s+/g, ' ').trim(); if (t.length > 2 && dict.has(t) && OCI18n.t(t) !== t) out.push(where + ': ' + t.slice(0, 60)); };
      const w = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
      for (let n = w.nextNode(); n; n = w.nextNode()) { if (n.parentElement && visible(n.parentElement) && !/^(SCRIPT|STYLE|OPTION)$/.test(n.parentElement.tagName)) check(n.nodeValue, 'text'); }
      document.querySelectorAll('option').forEach((o) => check(o.textContent, 'option'));
      document.querySelectorAll('[placeholder],[title],[aria-label]').forEach((el) => ['placeholder', 'title', 'aria-label'].forEach((a) => { const v = el.getAttribute(a); if (v) check(v, a); }));
      return [...new Set(out)];
    });
    const uiPage = async (url, lang) => {
      const p = await b4.newPage(); p.setDefaultTimeout(T(30000)); await p.setViewport({ width: 1300, height: 950 }); // generous: a slow runner loads six languages' pages here
      p.on('pageerror', (e) => errs.push(e.message)); await p.bringToFront();
      await p.goto(B + url + (url.includes('?') ? '&' : '?') + 'ui=' + lang, { waitUntil: 'domcontentloaded' }); return p;
    };
    const HELP_TITLE = { es: 'Primeros pasos', fr: 'Premiers pas', de: 'Erste Schritte', it: 'Primi passi', pt: 'Primeiros passos' };
    const SEEN = { es: 'Espectáculos', fr: 'Spectacles', de: 'Shows', it: 'Spettacoli', pt: 'Espetáculos' };
    for (const lang of ['es', 'fr', 'de', 'it', 'pt']) {
      const a = await uiPage('/admin', lang); await a.waitForSelector('#pw');
      const loginText = await a.evaluate(() => document.getElementById('login').innerText);
      check(lang + ': the sign-in page is translated', !/Sign in|Admin password/.test(loginText) && (await leaksOf(a)).length === 0, [loginText.slice(0, 120), await leaksOf(a)]);
      await signIn(a, '#steps button');
      check(lang + ': the Setup steps are translated', (await a.evaluate(() => document.getElementById('steps').innerText)).includes(SEEN[lang]) && !(await a.evaluate(() => document.documentElement.lang !== '' && false)));
      const leaks = [], stray = [];
      for (const step of ['Shows', 'Show file', 'Images', 'Look', 'Test', 'Share', 'Report']) {
        await a.evaluate((i) => document.querySelectorAll('#steps button')[i].click(), ['Shows', 'Show file', 'Images', 'Look', 'Test', 'Share', 'Report'].indexOf(step));
        await sleep(700); leaks.push(...(await leaksOf(a)));
        stray.push(...(await a.evaluate(() => [...document.getElementById('panel').innerText.matchAll(/\b(null|undefined|NaN|\[object Object\])\b/g)].map((m) => m[0]))));
        stray.push(...(await a.evaluate(() => [...document.querySelectorAll('input[type=file]')].filter((f) => !f.hidden).map(() => 'a raw file chooser'))));
      }
      check(lang + ': no stray "null" or "undefined" is printed on any step', stray.length === 0, stray);
      await a.evaluate(() => document.getElementById('settingsbtn').click()); await sleep(700); leaks.push(...(await leaksOf(a)));
      check(lang + ': no English is left in Setup (steps, Settings dialog)', leaks.length === 0, [...new Set(leaks)].slice(0, 8));
      const hp = await uiPage('/help?role=admin', lang);
      await hp.waitForFunction(() => { const h = document.querySelector('#doc h1'); return h && h.textContent.trim().length > 0; }, { timeout: 8000 }).catch(() => {});
      const helpState = await hp.evaluate(() => ({ title: (document.querySelector('#doc h1') || {}).textContent, topics: document.getElementById('topics').innerText.slice(0, 200), heads: document.querySelectorAll('#topics h2').length, leaks: [] }));
      check(lang + ': Help shows its topics and pages in the language', helpState.title === HELP_TITLE[lang] && !/Setting up|For the operator|Guides and reference/.test(helpState.topics) && helpState.heads === 3, helpState);
      await hp.close();
      await a.close();
    }
    const opEs = await uiPage('/operator', 'es'); await signIn(opEs, '#message'); await sleep(800);
    const opLeaks = await leaksOf(opEs);
    check('the key hints show the keycap symbols with the translated key names', await opEs.evaluate(() => { const t = (id) => document.querySelector('#' + id + ' small').textContent.replace(/\s+/g, ' ').trim(); return t('next') === '␣ Espacio' && t('back') === '⌫ Retroceso'; }));
    check('the operator page is translated (Spanish)', opLeaks.length === 0 && (await opEs.evaluate(() => document.getElementById('next').innerText)).includes('SIGUIENTE'), opLeaks.slice(0, 8));
    await opEs.close();
    // The language menu: choosing a language reloads the page in it and is remembered.
    const menu = await uiPage('/admin', 'en'); await signIn(menu, '#steps button');
    await menu.select('header select.uilang', 'fr'); await menu.waitForFunction(() => document.documentElement.lang === 'fr' && document.readyState === 'complete' && window.OCI18n && OCI18n.keys('fr').length > 0, { timeout: T(15000) }).catch(() => {});
    check('the language menu switches the page and remembers the choice', await menu.evaluate(() => document.documentElement.lang === 'fr' && localStorage.getItem('oc.ui.lang') === 'fr'));
    // The address now ends in a step (#show): going to the same address again is not a reload, so the menu has to reload on purpose.
    await menu.evaluate(() => { location.hash = '#show'; });
    check('the page has a #step in its address, as it does once the user has moved between steps', await menu.evaluate(() => location.hash === '#show'));
    await menu.select('header select.uilang', 'de'); await menu.waitForFunction(() => document.documentElement.lang === 'de' && document.readyState === 'complete' && window.OCI18n && OCI18n.keys('de').length > 0, { timeout: T(15000) }).catch(() => {});
    check('the language menu still works once the address ends in a step', await menu.evaluate(() => document.documentElement.lang === 'de' && localStorage.getItem('oc.ui.lang') === 'de' && document.getElementById('settingsbtn').textContent === 'Einstellungen'));
    await menu.select('header select.uilang', 'fr'); await menu.waitForFunction(() => document.documentElement.lang === 'fr' && document.readyState === 'complete' && window.OCI18n && OCI18n.keys('fr').length > 0, { timeout: T(15000) }).catch(() => {});
    check('an unknown or missing translation falls back to English', await menu.evaluate(() => OCI18n.t('A sentence nobody translated') === 'A sentence nobody translated' && OCI18n.t('Image "a.png" not found. Upload it in the Images step.').includes('introuvable')));
    await menu.evaluate(() => localStorage.removeItem('oc.ui.lang'));
    await menu.close(); await b4.close();

    // ------------------------------------------------------------------ restart
    section('Restart: logins, control and undo survive');
    const before2 = await cursor();
    check('the operator status strip shows how far through the show it is', /^(Item \d+ of \d+|\d+ items)$/.test(await op.evaluate(() => document.getElementById('progress').textContent)));
    await guard.evaluate(() => { location.hash = 'images'; }); await sleep(900);
    check('Setup offers the next step under the current one', await guard.evaluate(() => { const b = document.querySelector('#nextstep button'); return !!b && b.offsetWidth > 0 && /^Next: /.test(b.textContent); }));
    check('the operator page has a title with the show and its state', /^Run — .+ \((ON AIR|STANDBY)\)$/.test(await op.title()), await op.title());
    await stopServer(); await sleep(T(3500));
    const lost = (pg) => pg.evaluate(() => !document.getElementById('connbar').hidden && /lost/.test(document.getElementById('connbar').textContent));
    check('when the program cannot be reached, the operator page and Setup both say so', (await lost(op)) && (await lost(guard)));
    await startServer(); await sleep(T(5000));
    const back = (pg) => pg.evaluate(() => document.getElementById('connbar').hidden);
    check('and the notice goes away by itself once it is back', (await back(op)) && (await back(guard)));
    check('the operator page is still signed in and in control', (await op.evaluate(() => document.getElementById('ctrl').textContent)).includes('in control'), await op.evaluate(() => document.getElementById('ctrl').textContent));
    check('the position is kept', (await cursor()) === before2);
    await op.keyboard.press('Space'); await sleep(500);
    check('Space works after a restart without signing in again', (await cursor()) === before2 + 1);
    await op.keyboard.press('Backspace'); await sleep(500);
    check('Backspace still undoes a line sent before the restart', (await cursor()) === before2);
    await op.keyboard.press('Space'); await sleep(500); // the show is live again
    await op.evaluate(() => fetch('/api/operator/message', { method: 'POST', body: JSON.stringify({ text: 'Back soon' }) })); await sleep(700);
    check('an announcement on the screens is quoted in a banner on the operator page', await op.evaluate(() => { const b = document.getElementById('banner'); return !b.hidden && b.classList.contains('announce') && /Back soon/.test(b.textContent); }));
    await op.evaluate(() => fetch('/api/operator/back', { method: 'POST', body: '{}' })); await sleep(700);
    check('and the banner goes when it is undone', await op.evaluate(() => !document.getElementById('banner').classList.contains('announce')));
    await op.evaluate(() => fetch('/api/operator/message', { method: 'POST', body: JSON.stringify({ text: 'Please remain seated, the show will resume shortly. '.repeat(9) }) })); await sleep(700);
    check('a very long announcement is cut short in the banner (which stays a few lines)', await op.evaluate(() => { const b = document.getElementById('banner'); return !b.hidden && b.classList.contains('announce') && b.textContent.includes('…') && b.textContent.length < 260 && b.getBoundingClientRect().height < 130; }));
    await op.evaluate(() => fetch('/api/operator/back', { method: 'POST', body: '{}' })); await sleep(700);
    check('choosing a section gives the keyboard back to the page (Enter and Esc work at once)', await op.evaluate(() => { const sel = document.getElementById('sections'); if (sel.hidden || sel.options.length < 2) return true; sel.focus(); sel.value = sel.options[1].value; sel.dispatchEvent(new Event('change')); const ok = document.activeElement !== sel; document.getElementById('jumpcancel').click(); return ok; }));
    // Deleting the show that is live is refused, in a window that cannot be missed.
    await guard.goto(B + '/admin#show'); await guard.reload(); await guard.waitForSelector('.card.show'); await sleep(500);
    await guard.evaluate(() => { const c = [...document.querySelectorAll('.card.show')].find((x) => /ON AIR/.test(x.textContent)); [...c.querySelectorAll('button')].find((b) => b.textContent.startsWith('Delete')).click(); });
    await guard.waitForFunction(() => document.getElementById('dlg').open, { timeout: T(5000) }).catch(() => {});
    check('deleting the live show says why it cannot be done', await guard.evaluate(() => { const d = document.getElementById('dlg'); return d.open && /This show is on air/.test(d.innerText) && d.querySelector('#dlgcancel').hidden; }));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    check('and the show is still there', await guard.evaluate(() => [...document.querySelectorAll('.card.show')].some((x) => /ON AIR/.test(x.textContent))));

    // ------------------------------------------------------------------ this computer
    section('This computer: no password for Setup and the operator page');
    await stopServer(); await sleep(400);
    fs.writeFileSync(path.join(dataDir, 'config.yaml'), `admin_password: ${ADMIN}\noperator_password: ${OPERATOR}\nmdns:\n  enabled: false\naccess:\n  local_auto_login: true\n`);
    await startServer(); await sleep(T(1500));
    const ctx = await b.createBrowserContext();
    const lp = await ctx.newPage(); lp.setDefaultTimeout(T(8000)); lp.on('dialog', (d) => d.accept());
    await lp.goto(B + '/operator'); await lp.waitForSelector('#list .cue', { timeout: T(15000) }).catch(() => {});
    check('on localhost the operator page opens without a password', await lp.evaluate(() => document.getElementById('login').hidden && !document.getElementById('app').hidden));
    check('and links back to Setup (it is signed in as admin), with no Sign out', await lp.evaluate(() => !document.getElementById('setuplink').hidden && document.getElementById('logout').offsetWidth === 0));
    await lp.goto(B + '/admin'); await lp.waitForSelector('#steps button', { timeout: T(15000) }).catch(() => {});
    check('and Setup opens without a password, with no Sign out', await lp.evaluate(() => document.getElementById('login').hidden && !!document.querySelector('#steps button') && document.getElementById('logout').offsetWidth === 0));
    {
      // REQ-070: at the end of the show the banner has a link to the audience report
      const j = (m, u, b) => lp.evaluate(async (m2, u2, b2) => { const r = await fetch(u2, { method: m2, headers: { 'Content-Type': 'application/json' }, body: b2 === undefined ? undefined : JSON.stringify(b2) }); return r.json().catch(() => ({})); }, m, u, b);
      await lp.goto(B + '/admin#show'); await lp.waitForSelector('#steps button', { timeout: T(15000) }).catch(() => {});
      const tiny = await j('POST', '/api/admin/shows', { name: 'Tiny' });
      await j('PUT', '/api/admin/shows/' + tiny.id + '/showfile', { text: 'One\nTwo\n' });
      await j('POST', '/api/admin/shows/' + tiny.id + '/activate', { start: true });
      await lp.goto(B + '/operator'); await lp.waitForSelector('#list .cue', { timeout: T(15000) }).catch(() => {});
      await j('POST', '/api/operator/claim', { takeover: true });
      for (let i = 0; i < 3; i++) { await lp.keyboard.press('Space'); await sleep(300); }
      await lp.waitForFunction(() => /End of the show/.test(document.getElementById('banner').textContent), { timeout: T(5000) }).catch(() => {});
      check('at the end of the show the banner says so and links to the audience report', await lp.evaluate(() => { const b = document.getElementById('banner'); const a = b.querySelector('a.reportlink'); return !b.hidden && /End of the show/.test(b.textContent) && !!a && a.getAttribute('href') === '/operator/report' && a.target === '_blank'; }));
    }
    await lp.close(); await ctx.close();
    // the address of this computer on the network is not localhost, so a password is asked for there
    const lan = require('os').networkInterfaces();
    const ip = Object.values(lan).flat().find((i) => i && i.family === 'IPv4' && !i.internal);
    if (ip) {
      const ctx2 = await b.createBrowserContext(); const lp2 = await ctx2.newPage(); lp2.setDefaultTimeout(T(8000));
      await lp2.goto('http://' + ip.address + ':' + port + '/admin'); await lp2.waitForSelector('#pw', { visible: true }).catch(() => {});
      check('reaching the same computer by its network address asks for the password', await lp2.evaluate(() => !document.getElementById('login').hidden));
      await lp2.close(); await ctx2.close();
    }
  } finally {
    await b.close();
  }
  check('no JavaScript errors on any page', errs.length === 0, errs);
}

main().catch((e) => { console.error('\nERROR (in the section "' + sectionName + '"):', e.stack || e.message); failures++; }).finally(async () => {
  await cleanup();
  console.log(`\n${passes} passed, ${failures} failed`);
  process.exit(failures ? 1 : 0);
});
setTimeout(() => { console.error(`timed out in the section "${sectionName}"`); cleanup().finally(() => process.exit(2)); }, (Number(process.env.OC_E2E_TIMEOUT) || 900) * 1000).unref();
for (const sig of ['SIGINT', 'SIGTERM']) process.on(sig, () => cleanup().finally(() => process.exit(130)));
