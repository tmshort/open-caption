// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// HTTPS in a real browser: starts the built program with HTTPS on (once with a certificate file that names a host, once
// with its own self-signed certificate) and checks what Setup's Share step and Settings > Certificate show, that the QR
// code and the lobby sign follow the chosen address, and that the audience page works over https. The main browser test
// (run.js) runs over plain http and cannot see any of this. Needs the openssl command (skipped without it).
//
//   make build && cd test/e2e && npm install && node https.js
'use strict';
const puppeteer = require('puppeteer-core');
const { spawn, spawnSync } = require('child_process');
const fs = require('fs'), os = require('os'), path = require('path'), net = require('net');

const CHROME = process.env.CHROME || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const BIN = process.env.OC_BIN || path.join(__dirname, '..', '..', 'bin', 'open-caption' + (process.platform === 'win32' ? '.exe' : ''));
const PW = 'https-test-pw';
const SLOW = Number(process.env.OC_E2E_SLOW) || (process.env.CI ? 2 : 1);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms * SLOW));
const free = () => new Promise((res) => { const s = net.createServer(); s.listen(0, '127.0.0.1', () => { const p = s.address().port; s.close(() => res(p)); }); });
const NAME = 'captions.example.test';

let failures = 0, passes = 0;
function check(name, ok, detail) {
  if (ok) { passes++; console.log('  ok   ' + name); }
  else { failures++; console.log('  FAIL ' + name + (detail !== undefined ? '  -> ' + JSON.stringify(detail) : '')); }
}

const made = [];
// Start the program with HTTPS on; certFiles = [cert, key] or null for the program's own self-signed certificate.
async function start(certFiles) {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'oc-https-'));
  const http = await free(), https = await free();
  let yaml = `admin_password: ${PW}\noperator_password: ${PW}2\nmdns:\n  enabled: false\naccess:\n  local_auto_login: false\n` +
    `http:\n  addr: "127.0.0.1:${http}"\nhttps:\n  enabled: true\n  addr: "127.0.0.1:${https}"\n`;
  if (certFiles) yaml += `  cert_file: "${certFiles[0].replace(/\\/g, '/')}"\n  key_file: "${certFiles[1].replace(/\\/g, '/')}"\n`;
  fs.writeFileSync(path.join(dir, 'config.yaml'), yaml);
  const proc = spawn(BIN, ['--data', dir, '--no-browser'], { stdio: 'ignore' });
  const rec = { dir, proc, http, https };
  made.push(rec);
  for (let i = 0; i < 80; i++) { try { if ((await fetch(`http://127.0.0.1:${http}/healthz`)).ok) return rec; } catch (e) { /* not up yet */ } await sleep(100); }
  throw new Error('the program did not start');
}

async function cleanup(browser) {
  try { await browser.close(); } catch (e) { /* gone */ }
  for (const m of made) { try { m.proc.kill('SIGTERM'); } catch (e) { /* gone */ } }
  await sleep(500);
  for (const m of made) { try { fs.rmSync(m.dir, { recursive: true, force: true }); } catch (e) { /* still held on Windows */ } }
}

// A certificate for NAME and *.example.test (the wildcard must be left out of the offered addresses).
function makeCert() {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'oc-cert-'));
  const cert = path.join(dir, 'cert.pem'), key = path.join(dir, 'key.pem');
  const r = spawnSync('openssl', ['req', '-x509', '-newkey', 'rsa:2048', '-nodes', '-keyout', key, '-out', cert, '-days', '2', '-subj', '/CN=' + NAME,
    '-addext', `subjectAltName=DNS:${NAME},DNS:*.example.test`], { encoding: 'utf8' });
  if (r.status !== 0) return null;
  return [cert, key, dir];
}

async function main() {
  if (!fs.existsSync(BIN)) throw new Error('build the program first: make build (looked for ' + BIN + ')');
  const cert = makeCert();
  if (!cert) { console.log('openssl is not available: the HTTPS browser test is skipped'); return 0; }
  const browser = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox', '--ignore-certificate-errors'] });
  const errs = [];
  const open = async (width = 1300) => {
    const p = await browser.newPage(); p.setDefaultTimeout(8000 * SLOW);
    await p.setViewport({ width, height: 950 });
    p.on('pageerror', (e) => errs.push(e.message));
    p.on('dialog', (d) => d.accept());
    return p;
  };
  const click = (p, label, sel = 'button') => p.evaluate((l, s) => { const e = [...document.querySelectorAll(s)].find((x) => x.textContent.includes(l)); if (!e) throw new Error('no ' + s + ' with text ' + l); e.click(); }, label, sel);
  const toShare = async (base) => {
    const p = await open();
    await p.goto(base + '/admin'); await p.waitForSelector('#pw'); await p.type('#pw', PW); await p.keyboard.press('Enter'); await p.waitForSelector('#steps button');
    await p.evaluate(async () => { await fetch('/api/admin/shows', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ sample: true }) }); });
    await click(p, 'Show', '#steps button'); await sleep(700);
    await click(p, 'Work on this show'); await sleep(900);
    await click(p, 'Share', '#steps button'); await p.waitForSelector('.urls'); await sleep(600);
    return p;
  };
  const rows = (p) => p.evaluate(() => [...document.querySelectorAll('.urls .row label a')].map((a) => a.getAttribute('href')));
  const certTab = async (p) => {
    await p.evaluate(() => document.getElementById('settingsbtn').click()); await p.waitForSelector('#settab-cert', { visible: true });
    await p.evaluate(() => document.getElementById('settab-cert').click()); await sleep(500);
    const text = await p.evaluate(() => document.getElementById('setpanel-cert').innerText);
    await p.keyboard.press('Escape'); await sleep(200);
    return text;
  };

  // ---- a certificate file that names a host
  console.log('\nHTTPS with a certificate file');
  const a = await start(cert);
  const A = `https://localhost:${a.https}`;
  let p = await toShare(A);
  const names = await rows(p);
  check('the Share step is reached over https', (await p.evaluate(() => location.protocol)) === 'https:');
  check('the addresses are grouped as Secure (https) and Plain (http)', await p.evaluate(() => [...document.querySelectorAll('.urls h4')].map((h) => h.textContent).join('|') === 'Secure (https)|Plain (http)'));
  check('the certificate name is the first address, without the wildcard', names[0] === `https://${NAME}:${a.https}/` && !names.some((u) => u.includes('*')), names);
  check('https://localhost is never offered as an address', !names.some((u) => /localhost/.test(u)), names);
  check('the big address and the checked choice are the certificate name', await p.evaluate((want) => document.querySelector('.urlbig').textContent === want && document.querySelector('.urls input[type=radio]:checked').closest('.row').querySelector('a').getAttribute('href') === want, `https://${NAME}:${a.https}/`));
  await sleep(500);
  check('the QR code for the certificate name is drawn', await p.evaluate(() => { const i = document.querySelector('.qr img'); return !!i && i.complete && i.naturalWidth > 0; }));
  check('no self-signed warning is shown for a certificate file', await p.evaluate(() => !/self-signed/i.test(document.getElementById('panel').innerText)));
  const printHref = () => p.evaluate(() => new URL(document.getElementById('printsign').href).searchParams.get('url'));
  check('the lobby sign is for the certificate name', (await printHref()) === `https://${NAME}:${a.https}/`);
  await p.evaluate(() => { const r = [...document.querySelectorAll('.urls input[type=radio]')].find((x) => /^http:\/\//.test(x.closest('.row').querySelector('a').getAttribute('href'))); r.click(); });
  const plain = await p.evaluate(() => document.querySelector('.urlbig').textContent);
  check('choosing a plain http address changes the big address and the lobby sign', /^http:\/\/127\.|^http:\/\/\d/.test(plain) && (await printHref()) === plain, plain);
  await sleep(500);
  check('and the QR code for it is drawn too', await p.evaluate(() => { const i = document.querySelector('.qr img'); return !!i && i.complete && i.naturalWidth > 0; }));
  const tabText = await certTab(p);
  check('Settings > Certificate says the certificate files are used', /certificate files named in the settings file/.test(tabText), tabText);
  // The session cookie is Secure over https (and a reload keeps the browser signed in); over plain http it is not Secure
  // (Chrome would still keep a Secure cookie on 127.0.0.1, but a browser on the venue's real address would drop it, and nobody
  // could stay signed in).
  const sess = (await p.cookies()).filter((c) => c.httpOnly);
  check('over https the session cookie is Secure, HttpOnly and SameSite=Strict', sess.length === 1 && sess[0].secure && sess[0].sameSite === 'Strict', sess);
  await p.reload(); await p.waitForSelector('#steps button');
  check('and a reload over https keeps the browser signed in', await p.evaluate(() => !document.getElementById('login') || document.getElementById('login').hidden));
  await p.close();
  const hp = await open();
  await hp.goto(`http://127.0.0.1:${a.http}/admin`); await hp.waitForSelector('#pw'); await hp.type('#pw', PW); await hp.keyboard.press('Enter'); await hp.waitForSelector('#steps button');
  const plainSess = (await hp.cookies()).filter((c) => c.httpOnly);
  check('over plain http the session cookie is not Secure', plainSess.length === 1 && !plainSess[0].secure && plainSess[0].sameSite === 'Strict', plainSess);
  await hp.reload(); await hp.waitForSelector('#steps button');
  check('and a reload over plain http keeps the browser signed in', await hp.evaluate(() => !document.getElementById('login') || document.getElementById('login').hidden));
  await hp.close();

  // ---- the audience page over https
  const aud = await open(420);
  await aud.goto(`${A}/`); await sleep(1200);
  check('the audience page works over https (no script errors, not offline)', errs.length === 0 && (await aud.evaluate(() => /^mode-/.test([...document.body.classList].find((c) => /^mode-/.test(c)) || '') && document.getElementById('offline').hidden !== false)), errs);
  await aud.close();

  // ---- the program's own self-signed certificate
  console.log('\nHTTPS with a self-signed certificate');
  const b = await start(null);
  p = await toShare(`https://localhost:${b.https}`);
  const own = await rows(p);
  check('with no certificate name, the first choice is the numeric plain address', await p.evaluate(() => /^http:\/\/\d+\.\d+\.\d+\.\d+/.test(document.querySelector('.urlbig').textContent)), own);
  check('both groups are still listed', await p.evaluate(() => document.querySelectorAll('.urls h4').length === 2));
  check('the self-signed warning is shown', await p.evaluate(() => /self-signed/i.test(document.getElementById('panel').innerText)));
  const own2 = await certTab(p);
  check('Settings > Certificate says the program made the certificate itself', /made itself/.test(own2), own2);
  await p.close();

  await cleanup(browser);
  try { fs.rmSync(cert[2], { recursive: true, force: true }); } catch (e) { /* ignore */ }
  console.log(`\n${passes} passed, ${failures} failed`);
  return failures ? 1 : 0;
}

main().then((code) => process.exit(code), (e) => { console.error(e); for (const m of made) { try { m.proc.kill('SIGKILL'); } catch (x) { /* gone */ } } process.exit(2); });
