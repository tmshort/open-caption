// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// The audience report (REQ-070), drawn for both the stand-alone page (/operator/report) and Setup's Report step:
// OCReport.build(data) returns the whole report as one element: what the numbers say, a graph of the devices over time
// with the loops marked, the loops, the languages and a row for each device. The caller fetches the data and adds
// the toolbar. Page text goes through __() like the rest of Setup and the operator page.
(() => {
  'use strict';
  const lang = (window.OCI18n && OCI18n.lang) || 'en';
  const h = (tag, attrs, ...kids) => {
    const e = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs || {})) { if (v == null || v === false) continue; if (k === 'class') e.className = v; else e.setAttribute(k, v === true ? '' : v); }
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  };
  const svgNS = 'http://www.w3.org/2000/svg';
  const sv = (tag, attrs, ...kids) => {
    const e = document.createElementNS(svgNS, tag);
    for (const [k, v] of Object.entries(attrs || {})) if (v != null && v !== false) e.setAttribute(k, String(v));
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  };
  const when = (t) => (t && !String(t).startsWith('0001') ? new Date(t).toLocaleString(lang) : '');
  const clock = (t) => (t && !String(t).startsWith('0001') ? new Date(t).toLocaleTimeString(lang) : '');
  const hhmm = (ms) => new Date(ms).toLocaleTimeString(lang, { hour: '2-digit', minute: '2-digit' });
  const span = (s) => { s = Math.round(s || 0); const p = (n) => (n < 10 ? '0' : '') + n; return Math.floor(s / 3600) + ':' + p(Math.floor(s / 60) % 60) + ':' + p(s % 60); };
  const kindName = (k) => k === 'phone' ? __('Phone') : k === 'tablet' ? __('Tablet') : __('Computer');
  // what a loop is called: the name its show file gives it, else a guess from where it stands among the lines
  const loopName = (k, name) => name ? name : k === 'preshow' ? __('Pre-show') : k === 'intermission' ? __('Intermission') : k === 'postshow' ? __('Post-show') : __('Loop');
  // a "nice" step (1, 2, 5, 10, 20, 50 ...) so that about n ticks cover the range
  const niceStep = (range, n) => { const raw = Math.max(range / n, 1e-9), mag = Math.pow(10, Math.floor(Math.log(raw) / Math.LN10)), f = raw / mag; return (f <= 1 ? 1 : f <= 2 ? 2 : f <= 5 ? 5 : 10) * mag; };
  const TIME_STEPS = [30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 14400].map((x) => x * 1000);

  // the graph: devices connected over time, with the stretches during which a loop was running
  function chart(r) {
    const W = 900, H = 300, L = 58, R = 18, T = 30, B = 40, pw = W - L - R, ph = H - T - B;
    const t0 = new Date(r.started).getTime(), end = r.running ? Date.now() : new Date(r.ended).getTime();
    const pts = (r.series || []).map((p) => ({ t: new Date(p.t).getTime(), n: p.n }));
    const t1 = Math.max(end, pts.length ? pts[pts.length - 1].t : 0, t0 + 60000);
    const peak = Math.max(1, r.totals.peak || 0, ...pts.map((p) => p.n));
    const ystep = Math.max(1, Math.ceil(niceStep(peak, 5))), ymax = Math.ceil(peak / ystep) * ystep; // whole devices only: there is no half a device
    const x = (t) => L + (pw * (t - t0)) / (t1 - t0), y = (n) => T + ph - (ph * n) / ymax;
    const label = __('Devices connected over time, from {start} to {end}: up to {peak} at once.', { start: when(t0), end: when(t1), peak: r.totals.peak });
    const kids = [sv('title', {}, __('Devices connected over time')), sv('desc', {}, label),
      sv('defs', {}, sv('pattern', { id: 'ochatch', width: 7, height: 7, patternUnits: 'userSpaceOnUse', patternTransform: 'rotate(45)' }, sv('rect', { width: 3, height: 7, class: 'hatchbar' })))];
    for (const lp of r.loops || []) { // the loops first, behind everything else
      const a = x(new Date(lp.start).getTime()), b = Math.max(a + 2, x(new Date(lp.end).getTime()));
      kids.push(sv('rect', { class: 'band', x: a, y: T, width: b - a, height: ph }, sv('title', {}, loopName(lp.kind, lp.name) + ': ' + clock(lp.start) + ' – ' + clock(lp.end))));
      kids.push(sv('rect', { x: a, y: T, width: b - a, height: ph, fill: 'url(#ochatch)' }));
      if (b - a >= 60) kids.push(sv('text', { x: (a + b) / 2, y: T - 8, 'text-anchor': 'middle' }, loopName(lp.kind, lp.name)));
    }
    for (let v = 0; v <= ymax; v += ystep) {
      kids.push(sv('line', { class: 'grid', x1: L, x2: W - R, y1: y(v), y2: y(v) }), sv('text', { class: 'tick', x: L - 6, y: y(v) + 4, 'text-anchor': 'end' }, String(v)));
    }
    const tstep = TIME_STEPS.find((s) => (t1 - t0) / s <= 8) || TIME_STEPS[TIME_STEPS.length - 1];
    // the start and the end are always on the axis, labelled; the round times between them are left out where they would crowd them
    for (let tk = Math.ceil(t0 / tstep) * tstep; tk <= t1; tk += tstep) {
      if (x(tk) - x(t0) < 72 || x(t1) - x(tk) < 72) continue;
      kids.push(sv('line', { class: 'axis', x1: x(tk), x2: x(tk), y1: T + ph, y2: T + ph + 5 }), sv('text', { class: 'tick', x: x(tk), y: T + ph + 20, 'text-anchor': 'middle' }, hhmm(tk)));
    }
    for (const [t, name, anchor] of [[t0, __('Start'), 'start'], [t1, r.running ? __('Now') : __('End'), 'end']]) { // the text sits inside the plot, so it is never cut off
      kids.push(sv('line', { class: 'axis', x1: x(t), x2: x(t), y1: T + ph, y2: T + ph + 5 }),
        sv('text', { class: 'edge', x: x(t), y: T + ph + 20, 'text-anchor': anchor }, hhmm(t)), sv('text', { class: 'tick', x: x(t), y: T + ph + 34, 'text-anchor': anchor }, name));
    }
    kids.push(sv('line', { class: 'axis', x1: L, x2: W - R, y1: T + ph, y2: T + ph }), sv('line', { class: 'axis', x1: L, x2: L, y1: T, y2: T + ph }));
    if (pts.length) { // a step line (the count stays until it changes), and the area under it
      let d = 'M' + x(pts[0].t) + ',' + y(pts[0].n);
      for (let i = 1; i < pts.length; i++) d += 'H' + x(pts[i].t) + 'V' + y(pts[i].n);
      d += 'H' + x(t1);
      kids.push(sv('path', { class: 'area', d: d + 'V' + y(0) + 'H' + x(pts[0].t) + 'Z' }), sv('path', { class: 'line', d }));
    }
    if (r.totals.peak > 0 && r.totals.peakAt) {
      kids.push(sv('circle', { class: 'peak', cx: x(new Date(r.totals.peakAt).getTime()), cy: y(r.totals.peak), r: 5 }, sv('title', {}, __('Most at once') + ': ' + r.totals.peak + ' ' + __('at {time}', { time: clock(r.totals.peakAt) }))));
    }
    kids.push(sv('text', { class: 'tick', x: 14, y: T + ph / 2, 'text-anchor': 'middle', transform: 'rotate(-90 14 ' + (T + ph / 2) + ')' }, __('Devices')));
    return sv('svg', { id: 'chart', class: 'chart', viewBox: '0 0 ' + W + ' ' + H, role: 'img', focusable: 'false', 'aria-label': label }, kids);
  }

  const table = (id, head, rows, cls) => h('table', { id, class: cls }, h('thead', {}, h('tr', {}, head.map((t) => h('th', { scope: 'col' }, t)))), h('tbody', {}, rows));

  function build(d) {
    const r = d.report, t = r.totals;
    const card = (name, value, note) => h('div', { class: 'card' }, h('dt', {}, name), h('dd', {}, String(value), note ? h('small', {}, note) : null));
    const names = d.languageNames || {}, langs = Object.keys(r.languages || {}).sort(), loops = r.loops || [];
    return h('div', { class: 'oc-report' },
      h('h2', { id: 'show' }, r.show || r.showId),
      h('p', { id: 'period', role: 'status', class: r.running ? 'running' : '' }, r.running
        ? __('The show is on air now: these are the numbers so far (from {start}).', { start: when(r.started) })
        : __('Final report: from {start} to {end}.', { start: when(r.started), end: when(r.ended) })),
      h('dl', { id: 'totals', class: 'totals' },
        card(__('Devices'), t.devices, __('different devices that used the service')),
        card(__('Phones'), t.phones), card(__('Tablets'), t.tablets), card(__('Computers'), t.computers), card(__('Projector screens'), t.projectors),
        card(__('Most at once'), t.peak, t.peak ? __('at {time}', { time: clock(t.peakAt) }) : ''),
        card(r.running ? __('Connected now') : __('Still connected at the end'), t.now),
        card(__('Connections'), t.connections, __('{n} of them reconnections', { n: t.reconnects })),
        card(__('Average time connected'), span(t.avgConnected)), card(__('Longest stay'), span(t.longest))),
      h('p', { class: 'hint' }, __('Devices are told apart by their address and browser type, so this is an estimate: phones that share both count as one, and a phone whose address changes counts twice. A device counts once however often it reconnects and however short its stay.')),
      h('h3', {}, __('Devices over time')),
      h('div', { class: 'scroll', tabindex: '0', role: 'region', 'aria-label': __('Devices over time') }, chart(r)),
      h('p', { id: 'legend', class: 'legend' }, h('span', {}, h('i', { class: 'swatch devices' }), __('Devices connected')), loops.length ? h('span', {}, h('i', { class: 'swatch looprun' }), __('A loop was running')) : null),
      loops.length ? h('h3', {}, __('Loops')) : null,
      loops.length ? table('loops', [__('Loop'), __('From'), __('To'), __('Duration')], loops.map((lp) => h('tr', {}, h('td', {}, loopName(lp.kind, lp.name)), h('td', {}, clock(lp.start)), h('td', {}, clock(lp.end)), h('td', { class: 'num' }, span((new Date(lp.end) - new Date(lp.start)) / 1000)))), 'loops') : null,
      h('h3', {}, __('Languages')),
      table('langs', [__('Language'), __('Devices')], langs.length ? langs.map((c) => h('tr', {}, h('td', {}, (names[c] || c) + ' (' + c + ')'), h('td', { class: 'num' }, String(r.languages[c])))) : [h('tr', {}, h('td', { colspan: '2' }, '—'))], 'langs'),
      h('h3', {}, __('Devices')),
      h('div', { class: 'scroll', tabindex: '0', role: 'region', 'aria-label': __('Devices') },
        table('devices', [__('Device'), __('Address'), __('Type'), __('Languages'), __('First seen'), __('Last seen'), __('Connections'), __('Time connected'), __('Longest stay'), __('At the end')],
          r.devices.length ? r.devices.map((x) => h('tr', {},
            h('td', {}, x.id), h('td', {}, x.address), h('td', {}, x.projector ? __('Projector screen') : kindName(x.kind)), h('td', {}, (x.languages || []).join(', ')),
            h('td', {}, when(x.first)), h('td', {}, when(x.last)), h('td', { class: 'num' }, String(x.connections)), h('td', { class: 'num' }, span(x.connected)), h('td', { class: 'num' }, span(x.longest)),
            h('td', {}, x.now ? __('yes') : __('no')))) : [h('tr', {}, h('td', { colspan: '10' }, '—'))], 'devices')));
  }

  // the text of a report in the menu of reports: the show and when it started
  const menuLabel = (x) => (x.running ? __('Now') + ' — ' : '') + (x.show || '') + ' — ' + when(x.started);

  window.OCReport = { build, menuLabel, when };
})();
