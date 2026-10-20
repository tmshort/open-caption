// The audience report (REQ-070): how many devices used the audience page during a run of the show, however long each stayed.
(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const h = (tag, attrs, ...kids) => {
    const e = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs || {})) { if (v == null || v === false) continue; if (k === 'class') e.className = v; else e.setAttribute(k, v === true ? '' : v); }
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  };
  const lang = (window.OCI18n && OCI18n.lang) || 'en';
  const when = (t) => (t && !String(t).startsWith('0001') ? new Date(t).toLocaleString(lang) : '');
  const clock = (t) => (t && !String(t).startsWith('0001') ? new Date(t).toLocaleTimeString(lang) : '');
  const span = (s) => { s = Math.round(s || 0); const p = (n) => (n < 10 ? '0' : '') + n; return Math.floor(s / 3600) + ':' + p(Math.floor(s / 60) % 60) + ':' + p(s % 60); };
  const kindName = (k) => k === 'phone' ? __('Phone') : k === 'tablet' ? __('Tablet') : __('Computer');
  let timer = null, current = 'current';

  async function api(path) {
    let r;
    try { r = await fetch(path, { headers: {} }); } catch (e) { return { ok: false, status: 0, data: {} }; }
    let data = {}; try { data = await r.json(); } catch (e) { /* not JSON (CSV, or an error page) */ }
    return { ok: r.ok, status: r.status, data };
  }

  // ---- the graph: devices connected over time, with the stretches during which a loop was running ----
  const svgNS = 'http://www.w3.org/2000/svg';
  const sv = (tag, attrs, ...kids) => {
    const e = document.createElementNS(svgNS, tag);
    for (const [k, v] of Object.entries(attrs || {})) if (v != null && v !== false) e.setAttribute(k, String(v));
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  };
  const loopName = (k, name) => name ? name : k === 'preshow' ? __('Pre-show') : k === 'intermission' ? __('Intermission') : k === 'postshow' ? __('Post-show') : __('Loop');
  const hhmm = (ms) => new Date(ms).toLocaleTimeString(lang, { hour: '2-digit', minute: '2-digit' });
  // a "nice" step (1, 2, 5, 10, 20, 50 ...) so that about n ticks cover the range
  const niceStep = (range, n) => { const raw = Math.max(range / n, 1e-9), mag = Math.pow(10, Math.floor(Math.log(raw) / Math.LN10)), f = raw / mag; return (f <= 1 ? 1 : f <= 2 ? 2 : f <= 5 ? 5 : 10) * mag; };
  const TIME_STEPS = [30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 14400].map((x) => x * 1000);

  function drawChart(r) {
    const svg = $('chart'), W = 900, H = 300, L = 58, R = 18, T = 30, B = 40, pw = W - L - R, ph = H - T - B;
    const t0 = new Date(r.started).getTime(), end = r.running ? Date.now() : new Date(r.ended).getTime();
    const pts = (r.series || []).map((p) => ({ t: new Date(p.t).getTime(), n: p.n }));
    const t1 = Math.max(end, pts.length ? pts[pts.length - 1].t : 0, t0 + 60000);
    const peak = Math.max(1, r.totals.peak || 0, ...pts.map((p) => p.n));
    const ystep = niceStep(peak, 5), ymax = Math.ceil(peak / ystep) * ystep;
    const x = (t) => L + (pw * (t - t0)) / (t1 - t0), y = (n) => T + ph - (ph * n) / ymax;
    const kids = [sv('title', {}, __('Devices connected over time')),
      sv('desc', {}, __('Devices connected over time, from {start} to {end}: up to {peak} at once.', { start: when(t0), end: when(t1), peak: r.totals.peak })),
      sv('defs', {}, sv('pattern', { id: 'hatch', width: 7, height: 7, patternUnits: 'userSpaceOnUse', patternTransform: 'rotate(45)' }, sv('rect', { width: 3, height: 7, class: 'hatchbar' })))];
    // the loops first, behind everything else
    for (const lp of r.loops || []) {
      const a = x(new Date(lp.start).getTime()), b = Math.max(a + 2, x(new Date(lp.end).getTime()));
      kids.push(sv('rect', { class: 'band', x: a, y: T, width: b - a, height: ph }, sv('title', {}, loopName(lp.kind, lp.name) + ': ' + clock(lp.start) + ' – ' + clock(lp.end))));
      kids.push(sv('rect', { x: a, y: T, width: b - a, height: ph, fill: 'url(#hatch)' }));
      if (b - a >= 60) kids.push(sv('text', { x: (a + b) / 2, y: T - 8, 'text-anchor': 'middle' }, loopName(lp.kind, lp.name)));
    }
    for (let v = 0; v <= ymax; v += ystep) {
      kids.push(sv('line', { class: 'grid', x1: L, x2: W - R, y1: y(v), y2: y(v) }), sv('text', { class: 'tick', x: L - 6, y: y(v) + 4, 'text-anchor': 'end' }, String(v)));
    }
    const tstep = TIME_STEPS.find((s) => (t1 - t0) / s <= 8) || TIME_STEPS[TIME_STEPS.length - 1];
    for (let tk = Math.ceil(t0 / tstep) * tstep; tk <= t1; tk += tstep) {
      kids.push(sv('line', { class: 'axis', x1: x(tk), x2: x(tk), y1: T + ph, y2: T + ph + 5 }), sv('text', { class: 'tick', x: x(tk), y: T + ph + 20, 'text-anchor': 'middle' }, hhmm(tk)));
    }
    kids.push(sv('line', { class: 'axis', x1: L, x2: W - R, y1: T + ph, y2: T + ph }), sv('line', { class: 'axis', x1: L, x2: L, y1: T, y2: T + ph }));
    // devices connected: a step line (the count stays until it changes), and the area under it
    if (pts.length) {
      let d = 'M' + x(pts[0].t) + ',' + y(pts[0].n);
      for (let i = 1; i < pts.length; i++) d += 'H' + x(pts[i].t) + 'V' + y(pts[i].n);
      d += 'H' + x(t1);
      kids.push(sv('path', { class: 'area', d: d + 'V' + y(0) + 'H' + x(pts[0].t) + 'Z' }), sv('path', { class: 'line', d }));
    }
    if (r.totals.peak > 0 && r.totals.peakAt) {
      const px = x(new Date(r.totals.peakAt).getTime());
      kids.push(sv('circle', { class: 'peak', cx: px, cy: y(r.totals.peak), r: 5 }, sv('title', {}, __('Most at once') + ': ' + r.totals.peak + ' ' + __('at {time}', { time: clock(r.totals.peakAt) }))));
    }
    kids.push(sv('text', { class: 'tick', x: 14, y: T + ph / 2, 'text-anchor': 'middle', transform: 'rotate(-90 14 ' + (T + ph / 2) + ')' }, __('Devices')));
    svg.setAttribute('viewBox', '0 0 ' + W + ' ' + H);
    svg.setAttribute('aria-label', __('Devices connected over time, from {start} to {end}: up to {peak} at once.', { start: when(t0), end: when(t1), peak: r.totals.peak }));
    svg.replaceChildren(...kids);
    $('legend').replaceChildren(h('span', {}, h('i', { class: 'swatch devices' }), __('Devices connected')), (r.loops || []).length ? h('span', {}, h('i', { class: 'swatch looprun' }), __('A loop was running')) : null);
    const lt = $('loops').querySelector('tbody');
    $('loopsh').hidden = $('loops').hidden = !(r.loops || []).length;
    lt.replaceChildren(...(r.loops || []).map((lp) => h('tr', {}, h('td', {}, loopName(lp.kind, lp.name)), h('td', {}, clock(lp.start)), h('td', {}, clock(lp.end)), h('td', { class: 'num' }, span((new Date(lp.end) - new Date(lp.start)) / 1000)))));
  }

  function paint(d) {
    const r = d.report, t = r.totals;
    $('none').hidden = true; $('report').hidden = false;
    $('show').textContent = r.show || r.showId;
    const p = $('period'); p.className = r.running ? 'running' : '';
    p.textContent = r.running
      ? __('The show is on air now: these are the numbers so far (from {start}).', { start: when(r.started) })
      : __('Final report: from {start} to {end}.', { start: when(r.started), end: when(r.ended) });
    const card = (name, value, note) => h('div', { class: 'card' }, h('dt', {}, name), h('dd', {}, String(value), note ? h('small', {}, note) : null));
    $('totals').replaceChildren(
      card(__('Devices'), t.devices, __('different devices that used the service')),
      card(__('Phones'), t.phones), card(__('Tablets'), t.tablets), card(__('Computers'), t.computers), card(__('Projector screens'), t.projectors),
      card(__('Most at once'), t.peak, t.peak ? __('at {time}', { time: clock(t.peakAt) }) : ''),
      card(r.running ? __('Connected now') : __('Still connected at the end'), t.now),
      card(__('Connections'), t.connections, __('{n} of them reconnections', { n: t.reconnects })),
      card(__('Average time connected'), span(t.avgConnected)), card(__('Longest stay'), span(t.longest)));
    $('caveat').textContent = __('Devices are told apart by their address and browser type, so this is an estimate: phones that share both count as one, and a phone whose address changes counts twice. A device counts once however often it reconnects and however short its stay.');
    drawChart(r);
    const names = d.languageNames || {};
    const langs = Object.keys(r.languages || {}).sort();
    $('langs').querySelector('tbody').replaceChildren(...(langs.length ? langs.map((c) => h('tr', {}, h('td', {}, (names[c] || c) + ' (' + c + ')'), h('td', { class: 'num' }, String(r.languages[c])))) : [h('tr', {}, h('td', { colspan: '2' }, '—'))]));
    $('devices').querySelector('tbody').replaceChildren(...(r.devices.length ? r.devices.map((x) => h('tr', {},
      h('td', {}, x.id), h('td', {}, x.address), h('td', {}, x.projector ? __('Projector screen') : kindName(x.kind)), h('td', {}, (x.languages || []).join(', ')),
      h('td', {}, when(x.first)), h('td', {}, when(x.last)), h('td', { class: 'num' }, String(x.connections)), h('td', { class: 'num' }, span(x.connected)), h('td', { class: 'num' }, span(x.longest)),
      h('td', {}, x.now ? __('yes') : __('no')))) : [h('tr', {}, h('td', { colspan: '10' }, '—'))]));
    $('csv').href = '/api/operator/report.csv' + (current === 'current' ? '' : '?id=' + encodeURIComponent(current));
    document.title = __('Audience report') + ' — ' + (r.show || r.showId);
  }

  async function loadList() {
    const r = await api('/api/operator/reports');
    const sel = $('which');
    const items = r.ok ? r.data.reports : [];
    sel.replaceChildren(...items.map((x) => h('option', { value: x.id }, (x.running ? __('Now') + ' — ' : '') + (x.show || '') + ' — ' + when(x.started))));
    if (!items.length) sel.replaceChildren(h('option', { value: 'current' }, '—'));
    else if (!items.some((x) => x.id === current)) current = items[0].id;
    sel.value = current;
  }

  async function load() {
    const r = await api('/api/operator/report' + (current === 'current' ? '' : '?id=' + encodeURIComponent(current)));
    if (r.status === 401 || r.status === 403) { $('app').hidden = true; $('login').hidden = false; $('pw').focus(); return; }
    $('login').hidden = true; $('app').hidden = false;
    clearTimeout(timer);
    if (r.status === 404) { $('report').hidden = true; $('none').hidden = false; return; }
    if (!r.ok) { timer = setTimeout(load, 3000); return; }
    paint(r.data);
    if (r.data.report.running) timer = setTimeout(load, 5000); // the numbers move while the show is on
  }

  $('login').addEventListener('submit', async (e) => {
    e.preventDefault();
    let r;
    try { r = await fetch('/api/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ role: 'operator', password: $('pw').value }) }); } catch (x) { return; }
    if (r.ok) { $('loginerr').textContent = ''; $('pw').value = ''; start(); return; }
    $('loginerr').textContent = r.status === 429 ? __('Too many attempts. Wait a minute.') : __('That password was not accepted.');
  });
  $('which').addEventListener('change', (e) => { current = e.target.value; load(); });
  $('refresh').addEventListener('click', async () => { await loadList(); load(); });
  $('print').addEventListener('click', () => window.print());

  async function start() { await loadList(); await load(); }
  start();
  if (window.OCI18n) OCI18n.apply(document.body);
})();
