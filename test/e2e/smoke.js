// Browser smoke test: drives Setup (rehearsal, go live), Operator and Audience in real Chrome.
// See README.md in this directory.
const puppeteer = require('puppeteer-core');
const BASE = process.env.OC_BASE || 'http://localhost:18095';
const CHROME = process.env.CHROME || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const ADMIN_PW = process.env.OC_ADMIN_PW || 'adminpw';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
(async () => {
  const dir = process.argv[2] || require('os').tmpdir();
  const b = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  const p = await b.newPage(); await p.setViewport({ width: 1280, height: 1000 });
  const errs = [];
  const watch = (pg, tag) => { pg.on('pageerror', (e) => errs.push(tag + ' pageerror: ' + e.message)); pg.on('console', (m) => { if (m.type() === 'error' && !/401|404/.test(m.text())) errs.push(tag + ' console: ' + m.text()); }); };
  watch(p, 'admin');
  await p.goto(BASE + '/admin'); await p.waitForSelector('#pw');
  await p.type('#pw', ADMIN_PW); await p.keyboard.press('Enter');
  await p.waitForSelector('#steps button', { timeout: 5000 }); await sleep(500);
  const click = (label) => p.evaluate((l) => [...document.querySelectorAll('#steps button')].find((x) => x.textContent.includes(l)).click(), label);
  // --- rehearsal
  await click('Test'); await sleep(400);
  await p.evaluate(() => [...document.querySelectorAll('button')].find((x) => x.textContent.includes('Start rehearsal')).click());
  await p.waitForSelector('.reh iframe', { timeout: 5000 }); await sleep(1500);
  const frames = p.frames();
  const opf = frames.find((f) => f.url().includes('/operator?rehearsal=1'));
  const audf = frames.find((f) => f.url().includes('/?rehearsal=1'));
  await opf.click('#list'); // focus
  for (let i = 0; i < 4; i++) { await p.keyboard.press('Space'); await sleep(250); }
  await sleep(500);
  await p.screenshot({ path: dir + '/10-rehearsal.png' });
  console.log('rehearsal audience text:', JSON.stringify(await audf.evaluate(() => document.getElementById('captions').innerText)));
  // --- make sure the live audience is untouched
  const live = await b.newPage(); watch(live, 'live'); await live.setViewport({ width: 390, height: 780, deviceScaleFactor: 2 });
  await live.goto(BASE + '/'); await sleep(800);
  console.log('live audience (should be idle):', JSON.stringify(await live.evaluate(() => document.body.className + ' | ' + document.getElementById('idle').innerText)));
  // --- go live
  await click('Go Live'); await sleep(800);
  await p.evaluate(() => [...document.querySelectorAll('button')].find((x) => x.textContent.includes('Start show')).click());
  await p.waitForFunction(() => location.pathname === '/operator', { timeout: 5000 }); await sleep(1200);
  await p.screenshot({ path: dir + '/11-operator-start.png' });
  console.log('live audience after start:', JSON.stringify(await live.evaluate(() => document.body.className + ' | fs=' + (document.getElementById('fs').hidden ? 'none' : document.getElementById('fs').innerText.slice(0, 40)))));
  await live.screenshot({ path: dir + '/12-audience-preshow.png' });
  await p.keyboard.press('Space'); await sleep(400); // ends loop, sends first caption (after clear)
  for (let i = 0; i < 6; i++) { await p.keyboard.press('Space'); await sleep(300); }
  await p.screenshot({ path: dir + '/13-operator-mid.png' });
  await live.screenshot({ path: dir + '/14-audience-captions.png' });
  console.log('live audience captions:', JSON.stringify(await live.evaluate(() => document.getElementById('captions').innerText)));
  console.log('ERRORS:', JSON.stringify(errs));
  await b.close();
})().catch((e) => { console.error('FAIL', e); process.exit(1); });
