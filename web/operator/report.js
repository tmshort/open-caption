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
