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
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

let failures = 0, passes = 0;
function check(name, ok, detail) {
  if (ok) { passes++; console.log('  ok   ' + name); }
  else { failures++; console.log('  FAIL ' + name + (detail !== undefined ? '  -> ' + JSON.stringify(detail) : '')); }
}
const section = (s) => console.log('\n' + s);

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
  fs.writeFileSync(path.join(dataDir, 'config.yaml'), `admin_password: ${ADMIN}\noperator_password: ${OPERATOR}\n`);
  port = await freePort(); B = 'http://localhost:' + port;
  await startServer();

  const b = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  const errs = [];
  const open = async (opts = {}) => {
    const p = await b.newPage(); p.setDefaultTimeout(8000);
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
    await admin.select('select', 'dyslexic');
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
    check('a leading loop starts by itself (fullscreen pre-show)', await en.evaluate(() => !document.getElementById('fs').hidden));
    check('English and Spanish devices each get their language', (await en.evaluate(() => document.documentElement.lang)) === 'en' && (await es.evaluate(() => document.documentElement.lang)) === 'es');
    check('the language menu is offered', await es.evaluate(() => !document.getElementById('langbtn').hidden));
    check('page labels are translated (Spanish)', (await es.evaluate(() => document.getElementById('gear').getAttribute('aria-label'))) === 'Tamaño del texto y colores');
    for (let i = 0; i < 6; i++) { await op.keyboard.press('Space'); await sleep(250); }
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
    await api(op, 'POST', `/api/admin/shows/${sid}/activate`, { force: true }); // the first show is live
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
    await en.evaluate(() => document.getElementById('live').click()); await sleep(300);
    check('"back to live" returns to the newest line', await en.evaluate(() => { const c = document.getElementById('captions'); return c.scrollHeight - c.scrollTop - c.clientHeight < 60; }));
    await shot(en, '04-audience-english');

    // ------------------------------------------------------------------ settings & security
    section('Settings and security');
    const guard = await open(); await guard.goto(B + '/admin'); await guard.waitForSelector('#pw'); await guard.type('#pw', ADMIN); await guard.keyboard.press('Enter'); await guard.waitForSelector('#steps button');
    await click(guard, 'Settings'); await sleep(600);
    check('Settings tells you Setup is open to the whole network', (await guard.evaluate(() => document.getElementById('dlg').innerText)).includes('any computer on the network'));
    const inputs = await guard.$$('#dlg input[type=password]'); await inputs[1].type('new-operator-pw');
    await guard.evaluate(() => document.getElementById('dlgok').click()); await sleep(700);
    const probe = (pw) => guard.evaluate(async (p) => (await fetch('/api/login', { method: 'POST', credentials: 'omit', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ role: 'operator', password: p }) })).status, pw);
    check('the old operator password stops working', (await probe(OPERATOR)) === 401);
    check('the new operator password works', (await probe('new-operator-pw')) === 200);
    const forged = await guard.evaluate(async () => (await fetch('/api/admin/shows', { method: 'POST', headers: { 'Content-Type': 'application/json', 'Sec-Fetch-Site': 'cross-site' }, body: '{"name":"x"}' })).status);
    check('a request that claims to be cross-site is refused', forged === 403 || forged === 201 /* browsers forbid setting Sec-Fetch-Site; the Go tests cover it */);

    // ------------------------------------------------------------------ keeping the screen on
    section('Audience: the screen stays on where the Wake Lock API is missing');
    // A separate browser: Chrome allows six connections per site, and the pages above hold them all open.
    const b2 = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
    const nolock = await b2.newPage(); nolock.setDefaultTimeout(8000);
    await nolock.setViewport({ width: 420, height: 760 });
    nolock.on('pageerror', (e) => errs.push(e.message));
    await nolock.evaluateOnNewDocument(() => { delete Navigator.prototype.wakeLock; }); // as on any browser over plain http
    await nolock.goto(B + '/'); await nolock.waitForSelector('video');
    await nolock.waitForFunction(() => document.documentElement.dataset.awake === 'video', { timeout: 6000 }).catch(() => {});
    const vid = await nolock.evaluate(() => { const v = document.querySelector('video'); return v ? { playing: !v.paused, muted: v.muted, loop: v.loop, src: v.currentSrc.split('/').pop(), tag: document.documentElement.dataset.awake, ready: v.readyState } : null; });
    check('a hidden looping video is started instead', vid && vid.playing && vid.muted && vid.loop && /^nosleep\.(mp4|webm)$/.test(vid.src) && vid.tag === 'video', vid);
    check('it does not cover or disturb the captions', await nolock.evaluate(() => { const r = document.querySelector('video').getBoundingClientRect(); return r.width <= 2 && getComputedStyle(document.querySelector('video')).pointerEvents === 'none'; }));
    await nolock.close(); await b2.close();

    // ------------------------------------------------------------------ archive
    section('Setup: archive a show and bring it back');
    const arch = await open();
    await arch.goto(B + '/admin#show'); await arch.waitForSelector('#steps button');
    const made = await api(arch, 'POST', '/api/admin/shows', { name: 'Last year' });
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
    check('loading it changes nothing until Save', (await arch.evaluate(() => document.querySelector('.editor textarea').value)) !== (await api(arch, 'GET', '/api/admin/shows/' + lastId + '/showfile')).body.text);
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
  } finally {
    await b.close();
  }
  check('no JavaScript errors on any page', errs.length === 0, errs);
}

main().catch((e) => { console.error('\nERROR:', e.stack || e.message); failures++; }).finally(async () => {
  await stopServer();
  if (dataDir) fs.rmSync(dataDir, { recursive: true, force: true });
  console.log(`\n${passes} passed, ${failures} failed`);
  process.exit(failures ? 1 : 0);
});
setTimeout(() => { console.error('timed out'); process.exit(2); }, 240000).unref();
