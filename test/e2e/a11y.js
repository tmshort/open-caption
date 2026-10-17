// Accessibility audit: runs axe-core (WCAG 2.1 A/AA and best practices) against every page and Setup step, in dark and light
// colour schemes and at a phone width. Usage: node a11y.js   (exit 1 if any violation)
'use strict';
const puppeteer = require('puppeteer-core');
const axeSource = require('fs').readFileSync(require.resolve('axe-core/axe.min.js'), 'utf8');
const { spawn } = require('child_process');
const fs = require('fs'), os = require('os'), path = require('path'), net = require('net');
const CHROME = process.env.CHROME || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const BIN = process.env.OC_BIN || path.join(__dirname, '..', '..', 'bin', 'open-caption' + (process.platform === 'win32' ? '.exe' : ''));
const PW = 'a11y-test-pw';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const free = () => new Promise((res) => { const s = net.createServer(); s.listen(0, '127.0.0.1', () => { const p = s.address().port; s.close(() => res(p)); }); });
let bad = 0;
async function audit(page, label) {
  await page.evaluate(axeSource);
  const r = await page.evaluate(() => axe.run(document, { runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa', 'best-practice'] } }));
  if (!r.violations.length) { console.log('  ok   ' + label); return; }
  for (const v of r.violations) {
    bad++;
    console.log(`  FAIL ${label}: [${v.impact}] ${v.id} - ${v.help}`);
    for (const n of v.nodes.slice(0, 3)) console.log('         ' + n.target.join(' ') + '  ' + (n.failureSummary || '').split('\n').slice(0, 2).join(' | ').slice(0, 220));
  }
}
(async () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'oc-a11y-'));
  fs.writeFileSync(path.join(dir, 'config.yaml'), `admin_password: ${PW}\noperator_password: ${PW}2\nmdns:\n  enabled: false\naccess:\n  local_auto_login: false\n`);
  const port = await free(), B = 'http://localhost:' + port;
  const server = spawn(BIN, ['--data', dir, '--http-addr', ':' + port, '--no-browser'], { stdio: 'ignore' });
  for (let i = 0; i < 50; i++) { try { if ((await fetch(B + '/healthz')).ok) break; } catch (e) { /* wait */ } await sleep(100); }
  const b = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  try {
    for (const scheme of ['dark', 'light']) {
      for (const [w, h, tag] of [[1200, 900, 'desktop'], [390, 800, 'phone']]) {
        const p = await b.newPage(); await p.setViewport({ width: w, height: h });
        await p.emulateMediaFeatures([{ name: 'prefers-color-scheme', value: scheme }]);
        const t = `${scheme}/${tag}`;
        // audience: idle, then with captions
        await p.goto(B + '/'); await sleep(600); await audit(p, t + ' audience (waiting)');
        await p.goto(B + '/admin'); await p.waitForSelector('#pw'); await audit(p, t + ' setup sign-in');
        await p.type('#pw', PW); await p.keyboard.press('Enter'); await p.waitForSelector('#steps button');
        if (scheme === 'dark' && tag === 'desktop') {
          await p.waitForFunction(() => [...document.querySelectorAll('button')].some((x) => x.textContent.includes('sample')), { timeout: 20000 });
          await p.evaluate(() => [...document.querySelectorAll('button')].find((x) => x.textContent.includes('sample')).click()); await sleep(1500);
        }
        const steps = await p.$$eval('#steps button', (bs) => bs.length);
        for (let i = 0; i < steps; i++) {
          await p.evaluate((k) => document.querySelectorAll('#steps button')[k].click(), i); await sleep(900);
          const name = await p.evaluate((k) => document.querySelectorAll('#steps button')[k].innerText.replace(/\s+/g, ' ').trim(), i);
          await audit(p, `${t} setup: ${name}`);
        }
        await p.evaluate(() => document.getElementById('settingsbtn').click()); await sleep(800);
        for (const tab of ['aud', 'sec', 'abt']) { await p.evaluate((i) => document.getElementById('settab-' + i).click(), tab); await sleep(300); await audit(p, t + ' setup: settings dialog, ' + tab + ' tab'); }
        await p.evaluate(() => document.getElementById('dlgok').click()); await sleep(300);
        await p.goto(B + '/operator'); await p.waitForSelector('#message', { visible: true }).catch(() => {});
        await sleep(800); await audit(p, t + ' operator');
        await p.goto(B + '/help?role=admin'); await sleep(1000); await audit(p, t + ' help');
        // audience with captions on screen
        await p.goto(B + '/'); await sleep(800); await audit(p, t + ' audience (captions)');
        await p.close();
      }
    }
    // Reflow (WCAG 1.4.10): at 320 px wide nothing needs sideways scrolling; and keyboard focus is always visible (2.4.7).
    const pages = [['/', 'audience'], ['/admin', 'setup sign-in'], ['/operator', 'operator sign-in'], ['/help?role=admin', 'help']];
    for (const [url, name] of pages) {
      const p = await b.newPage(); await p.setViewport({ width: 320, height: 640 });
      await p.goto(B + url); await sleep(700);
      const over = await p.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
      if (over > 1) { bad++; const who = await p.evaluate(() => [...document.querySelectorAll('body *')].filter((e) => { if (getComputedStyle(e).display === 'none' || e.getBoundingClientRect().right <= innerWidth + 1) return false; for (let a = e.parentElement; a && a !== document.body; a = a.parentElement) { const o = getComputedStyle(a).overflowX; if (o === 'auto' || o === 'scroll' || o === 'hidden') return false; } return true; }).map((e) => e.tagName + '#' + e.id + '.' + e.className).slice(0, 5)); console.log(`  FAIL 320px ${name}: the page is ${over}px wider than the screen: ${who.join(' ')}`); } else console.log(`  ok   320px ${name} reflows`);
      await p.close();
    }
    {
      const p = await b.newPage(); await p.setViewport({ width: 1100, height: 800 });
      await p.goto(B + '/admin'); await p.waitForSelector('#pw'); await p.type('#pw', PW); await p.keyboard.press('Enter'); await p.waitForSelector('#steps button');
      const seen = new Set(); let hidden = 0;
      for (let i = 0; i < 14; i++) {
        await p.keyboard.press('Tab');
        const r = await p.evaluate(() => { const e = document.activeElement; if (!e || e === document.body) return null; const cs = getComputedStyle(e); return { tag: e.tagName + (e.id ? '#' + e.id : ''), ring: cs.outlineStyle !== 'none' && parseFloat(cs.outlineWidth) >= 2 || cs.boxShadow !== 'none' }; });
        if (r) { seen.add(r.tag); if (!r.ring) { hidden++; console.log('  FAIL keyboard focus has no visible indicator on ' + r.tag); } }
      }
      if (hidden) bad += hidden; else console.log(`  ok   keyboard: Tab reaches ${seen.size} controls in turn and each shows a focus ring`);
      if (seen.size < 6) { bad++; console.log('  FAIL keyboard: Tab reached only ' + seen.size + ' controls'); }
      await p.close();
    }
  } finally { await b.close(); server.kill(); fs.rmSync(dir, { recursive: true, force: true }); }
  console.log(bad ? `\n${bad} accessibility violation(s)` : '\nno accessibility violations found');
  process.exit(bad ? 1 : 0);
})();
