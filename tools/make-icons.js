// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Regenerates web/shared/favicon.ico and web/shared/apple-touch-icon.png from web/shared/favicon.svg.
//   cd test/e2e && npm install && cd ../.. && CHROME=/path/to/chrome node tools/make-icons.js
'use strict';
const puppeteer = require('../test/e2e/node_modules/puppeteer-core');
const fs = require('fs'), path = require('path');
const CHROME = process.env.CHROME || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const shared = path.join(__dirname, '..', 'web', 'shared');
const svg = fs.readFileSync(path.join(shared, 'favicon.svg'), 'utf8');

async function render(page, size, background) {
  await page.setViewport({ width: size, height: size });
  await page.setContent(`<body style="margin:0;background:${background || 'transparent'}"><div style="width:${size}px;height:${size}px">${svg.replace('<svg ', `<svg width="${size}" height="${size}" `)}</div>`);
  return page.screenshot({ omitBackground: !background, clip: { x: 0, y: 0, width: size, height: size } });
}

// An .ico file is a small table in front of ordinary PNG images.
function ico(images) {
  const head = Buffer.alloc(6 + 16 * images.length);
  head.writeUInt16LE(1, 2); head.writeUInt16LE(images.length, 4);
  let offset = head.length;
  images.forEach(({ size, png }, i) => {
    const e = 6 + 16 * i;
    head[e] = size % 256; head[e + 1] = size % 256; head.writeUInt16LE(1, e + 4); head.writeUInt16LE(32, e + 6);
    head.writeUInt32LE(png.length, e + 8); head.writeUInt32LE(offset, e + 12);
    offset += png.length;
  });
  return Buffer.concat([head, ...images.map((i) => i.png)]);
}

(async () => {
  const b = await puppeteer.launch({ executablePath: CHROME, headless: 'new', args: ['--no-sandbox'] });
  const p = await b.newPage();
  const images = [];
  for (const size of [16, 32, 48]) images.push({ size, png: Buffer.from(await render(p, size)) });
  fs.writeFileSync(path.join(shared, 'favicon.ico'), ico(images));
  fs.writeFileSync(path.join(shared, 'apple-touch-icon.png'), Buffer.from(await render(p, 180, '#ffffff')));
  await b.close();
  console.log('wrote favicon.ico and apple-touch-icon.png');
})();
