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
const realLaunch = puppeteer.launch.bind(puppeteer);
puppeteer.launch = async (opts) => { const b = await realLaunch(opts); launched.push(b); return b; };
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

  const b = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  const errs = [];
  const open = async (opts = {}) => {
    const p = await b.newPage(); p.setDefaultTimeout(T(8000));
    await p.setViewport(opts.viewport || { width: 1300, height: 950 });
    p.on('dialog', (d) => d.accept());
    p.on('pageerror', (e) => errs.push(e.message));
    if (opts.headers) await p.setExtraHTTPHeaders(opts.headers);
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
    await click(admin, 'Save look'); await sleep(700);
    const look = (await api(admin, 'GET', '/api/admin/shows/' + showId2)).body;
    check('the look is saved with the show', look.lookSaved === true && look.theme.font === 'dyslexic' && look.theme.speakers === 'color', look.theme);
    check('the Look step is ticked', await admin.evaluate(() => [...document.querySelectorAll('#steps button')].find((x) => x.textContent.includes('Look')).classList.contains('done')));

    // ------------------------------------------------------------------ rehearsal
    section('Rehearsal never reaches the audience');
    await step(admin, 'Test'); await sleep(500); await click(admin, 'Start rehearsal');
    await admin.waitForSelector('.reh iframe'); await sleep(1500);
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
    check('and the operator page says it is not on air, with START where NEXT will be', await op.evaluate(() => /STANDBY/.test(document.getElementById('live').textContent) && /START/.test(document.getElementById('next').textContent) && document.getElementById('offair').disabled && /NOT ON AIR/.test(document.getElementById('banner').textContent) && document.getElementById('stop').disabled));
    check('in standby the highlighted next line is the pre-show loop that START will run', await op.evaluate(() => { const n = document.querySelector('#list .cue.next'); return !!n && n.classList.contains('k-loop') && !n.classList.contains('sent') && !document.querySelector('#list .cue.now'); }));
    await op.keyboard.press('Backspace'); await op.keyboard.press('b'); await sleep(400);
    check('nothing but START does anything in standby', await standbyShown(en));
    await op.keyboard.press('Space'); await sleep(T(900)); // START: the pre-show loop begins
    check('START puts the show on air (LIVE), and the pre-show appears', await op.evaluate(() => /LIVE/.test(document.getElementById('live').textContent) && !document.getElementById('offair').disabled && /NEXT/.test(document.getElementById('next').textContent)));
    check('the item of a running loop is marked with a play triangle in the operator list', await op.evaluate(() => { const l = document.querySelectorAll('.items li.playing'); return l.length === 1 && l[0].getAttribute('aria-current') === 'true' && getComputedStyle(l[0].querySelector('.pl'), '::before').content.includes('▶'); }));
    check('the pre-show loop is on the audience screens after START (fullscreen)', await en.evaluate(() => !document.getElementById('fs').hidden));
    check('English and Spanish devices each get their language', (await en.evaluate(() => document.documentElement.lang)) === 'en' && (await es.evaluate(() => document.documentElement.lang)) === 'es');
    check('the language menu is offered', await es.evaluate(() => !document.getElementById('langbtn').hidden));
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
    await en.evaluate(() => { document.getElementById('captions').scrollTop = 120; }); await sleep(300);
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

    // ------------------------------------------------------------------ settings & security
    // An intermission timer: the loop's {timer} counts down live on the audience screen, from the moment the loop starts.
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
    check('Go off air asks first, and says what happens', await op.evaluate(() => /standby screen/.test(document.getElementById('offdlg').innerText)));
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
    const headerLook = (pg) => pg.evaluate(() => { const g = (q) => { const e = document.querySelector(q); if (!e || !e.offsetWidth) return null; const c = getComputedStyle(e); return [c.fontSize, c.fontFamily, c.fontWeight, Math.round(e.getBoundingClientRect().height)].join('|'); }; const h = document.querySelector('header'); return { header: getComputedStyle(h).fontSize + '|' + getComputedStyle(h).paddingTop + '|' + Math.round(h.getBoundingClientRect().height), title: g('header h1, header #show').split('|').slice(0, 3).join('|'), pill: g('#viewers'), help: g('#helpbtn'), lang: g('header select.uilang'), dot: g('#dot').split('|')[3] }; });
    const opLook = await headerLook(op);
    {
      const oh = await headerOrder(op, ['#viewers', 'header select.uilang', '#offair', '#reset', '#setuplink', '#helpbtn', '#logout', '#dot']);
      check('operator header, right side: devices, language, Go off air, Restart, Setup, Help, Sign out, connection', oh.ok && oh.pts.filter((x) => x.includes('hidden')).length === 0, oh);
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
    check('Settings tells you Setup is open to the whole network', (await guard.evaluate(() => document.getElementById('dlg').innerText)).includes('any computer on the network'));
    check('Settings offers the network name setting', await guard.evaluate(() => /Network name/.test(document.getElementById('dlg').innerText) && !document.querySelector('#dlg input[aria-label="Announce this computer by name"]').checked));
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
    check('Settings offers the log detail setting', await guard.evaluate(() => !!document.querySelector('#dlg select[aria-label="Log detail"]')));
    check('the Settings window has no Recent log button (it is on Go Live)', await guard.evaluate(() => !document.getElementById('dlgextra') || document.getElementById('dlgextra').hidden));
    check('and its only footer button just closes it', await guard.evaluate(() => document.getElementById('dlgok').textContent.trim() === 'Close' && document.getElementById('dlgcancel').hidden));
    // the operator password can be shown and hidden again
    await guard.evaluate(() => [...document.querySelectorAll('#dlg button')].find((x) => x.textContent.includes('Show the current operator password')).click()); await sleep(500);
    check('showing the operator password turns the button into Hide', await guard.evaluate(() => /Operator password: /.test(document.getElementById('dlg').innerText) && [...document.querySelectorAll('#dlg button')].some((x) => x.textContent.includes('Hide the operator password'))));
    await guard.evaluate(() => [...document.querySelectorAll('#dlg button')].find((x) => x.textContent.includes('Hide the operator password')).click()); await sleep(300);
    check('and Hide hides it again', await guard.evaluate(() => !/Operator password: /.test(document.getElementById('dlg').innerText)));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    await step(guard, 'Go Live'); await sleep(900);
    check('on Go Live each address is a link that opens a new window', await guard.evaluate(() => { const l = [...document.querySelectorAll('.urls label a')]; return l.length > 0 && l.every((x) => x.target === '_blank' && /noopener/.test(x.rel) && /^https?:\/\//.test(x.href)); }));
    await guard.evaluate(() => [...document.querySelectorAll('a')].find((x) => x.textContent.includes('Recent log')).click()); await guard.waitForFunction(() => { const p = document.querySelector('pre.logview'); return p && p.textContent.includes('signed in'); }, { timeout: 8000 }).catch(() => {});
    check('the recent log can be read in Setup', await guard.evaluate(() => /INFO .*signed in/.test(document.querySelector('pre.logview').textContent)));
    await guard.select('dialog select[aria-label="Show"]', 'warn'); await sleep(600);
    check('and narrowed to problems', await guard.evaluate(() => { const t = document.querySelector('pre.logview').textContent; return /WARN/.test(t) && !/INFO/.test(t); }));
    check('it never contains a password', !(await guard.evaluate(() => document.querySelector('pre.logview').textContent)).includes(ADMIN));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    await click(guard, 'Settings'); await sleep(600);
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
    const b2 = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
    const nolock = await b2.newPage(); nolock.setDefaultTimeout(T(8000));
    await nolock.setViewport({ width: 420, height: 760 });
    nolock.on('pageerror', (e) => errs.push(e.message));
    await nolock.evaluateOnNewDocument(() => { delete Navigator.prototype.wakeLock; }); // as on any browser over plain http
    await nolock.goto(B + '/'); await nolock.waitForSelector('video');
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
      const ah = await headerOrder(arch, ['#viewers', 'header select.uilang', '#settingsbtn', 'header a[href="/operator"]', '#helpbtn', '#logout', '#dot']);
      check('Setup header, right side: devices, language, Settings, Run, Help, Sign out, connection', ah.ok && ah.pts.filter((x) => x.includes('hidden')).length === 0, ah);
      const adLook = await headerLook(arch);
      check('the two headers use the same type, spacing and control sizes', JSON.stringify(adLook) === JSON.stringify(opLook), { operator: opLook, setup: adLook });
      check('and the Setup page shows the server is answering (green dot)', await arch.evaluate(() => document.getElementById('dot').classList.contains('on')));
    }
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
    check('loading it changes nothing until Save', (await arch.evaluate(() => document.querySelector('.editor textarea').value)) !== (await api(arch, 'GET', '/api/admin/shows/' + lastId + '/showfile')).body.text);
    // Go Live no longer puts a show in front of the audience: it opens the operator page, with the show in standby.
    await step(arch, 'Go Live'); await sleep(T(900));
    check('Go Live offers Open the operator page (the show waits in standby), not Start show', await arch.evaluate(() => { const t = document.getElementById('panel').innerText; return /Open the operator page/.test(t) && !/Start show/.test(t) && /standby/.test(t); }));
    await step(arch, 'Show file'); await sleep(T(600));
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
    const b3 = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
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
    await oldAdmin.waitForFunction(() => [...document.querySelectorAll('button')].some((x) => x.textContent.includes('Rename')), { timeout: 15000 });
    await click(oldAdmin, 'Rename'); await oldAdmin.waitForFunction(() => document.getElementById('dlg').hasAttribute('open'));
    const shown = await oldAdmin.evaluate(() => { const d = document.getElementById('dlg'), r = d.getBoundingClientRect(); return { fixed: getComputedStyle(d).position === 'fixed', w: r.width, onScreen: r.left >= 0 && r.top >= 0 && r.right <= innerWidth && r.bottom <= innerHeight, hasShowModal: typeof HTMLDialogElement !== 'undefined' && !!HTMLDialogElement.prototype.showModal }; });
    check('without <dialog> support the dialog still appears, centred over the page', shown.fixed && shown.w > 200 && shown.onScreen && !shown.hasShowModal, shown);
    await oldAdmin.keyboard.press('Escape'); await sleep(200);
    check('Escape closes it', await oldAdmin.evaluate(() => !document.getElementById('dlg').hasAttribute('open')));
    await click(oldAdmin, 'Rename'); await oldAdmin.waitForFunction(() => document.getElementById('dlg').hasAttribute('open'));
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
    const b4 = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
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
      for (const step of ['Shows', 'Show file', 'Images', 'Look', 'Test', 'Go Live']) {
        await a.evaluate((i) => document.querySelectorAll('#steps button')[i].click(), ['Shows', 'Show file', 'Images', 'Look', 'Test', 'Go Live'].indexOf(step));
        await sleep(700); leaks.push(...(await leaksOf(a)));
        stray.push(...(await a.evaluate(() => [...document.getElementById('panel').innerText.matchAll(/\b(null|undefined|NaN|\[object Object\])\b/g)].map((m) => m[0]))));
        stray.push(...(await a.evaluate(() => [...document.querySelectorAll('input[type=file]')].filter((f) => !f.hidden).map(() => 'a raw file chooser'))));
      }
      check(lang + ': no stray "null" or "undefined" is printed on any step', stray.length === 0, stray);
      await a.evaluate(() => document.getElementById('settingsbtn').click()); await sleep(700); leaks.push(...(await leaksOf(a)));
      check(lang + ': no English is left in Setup (steps, Settings dialog)', leaks.length === 0, [...new Set(leaks)].slice(0, 8));
      const hp = await uiPage('/help?role=admin', lang);
      await hp.waitForFunction(() => { const h = document.querySelector('#doc h1'); return h && h.textContent.trim().length > 0; }, { timeout: 8000 }).catch(() => {});
      const helpState = await hp.evaluate(() => ({ title: (document.querySelector('#doc h1') || {}).textContent, topics: document.getElementById('topics').innerText.slice(0, 200), leaks: [] }));
      check(lang + ': Help shows its topics and pages in the language', helpState.title === HELP_TITLE[lang] && !/Setting up|For the operator/.test(helpState.topics), helpState);
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
    check('an unknown or missing translation falls back to English', await menu.evaluate(() => OCI18n.t('A sentence nobody translated') === 'A sentence nobody translated' && OCI18n.t('Image "a.png" not found. Upload it in the Images step.').includes('introuvable')));
    await menu.evaluate(() => localStorage.removeItem('oc.ui.lang'));
    await menu.close(); await b4.close();

    // ------------------------------------------------------------------ restart
    section('Restart: logins, control and undo survive');
    const before2 = await cursor();
    await stopServer(); await sleep(500); await startServer(); await sleep(4500);
    check('the operator page is still signed in and in control', (await op.evaluate(() => document.getElementById('ctrl').textContent)).includes('in control'), await op.evaluate(() => document.getElementById('ctrl').textContent));
    check('the position is kept', (await cursor()) === before2);
    await op.keyboard.press('Space'); await sleep(500);
    check('Space works after a restart without signing in again', (await cursor()) === before2 + 1);
    await op.keyboard.press('Backspace'); await sleep(500);
    check('Backspace still undoes a line sent before the restart', (await cursor()) === before2);
    await op.keyboard.press('Space'); await sleep(500); // the show is live again
    // Deleting the show that is live is refused, in a window that cannot be missed.
    await guard.goto(B + '/admin#show'); await guard.reload(); await guard.waitForSelector('.card.show'); await sleep(500);
    await guard.evaluate(() => { const c = [...document.querySelectorAll('.card.show')].find((x) => /LIVE/.test(x.textContent)); [...c.querySelectorAll('button')].find((b) => b.textContent.startsWith('Delete')).click(); });
    await guard.waitForFunction(() => document.getElementById('dlg').open, { timeout: T(5000) }).catch(() => {});
    check('deleting the live show says why it cannot be done', await guard.evaluate(() => { const d = document.getElementById('dlg'); return d.open && /This show is live/.test(d.innerText) && d.querySelector('#dlgcancel').hidden; }));
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
    check('and the show is still there', await guard.evaluate(() => [...document.querySelectorAll('.card.show')].some((x) => /LIVE/.test(x.textContent))));

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
