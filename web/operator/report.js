// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// The stand-alone audience report page (REQ-070): the toolbar, the sign-in and loading; the report itself is drawn by
// web/shared/report.js, which Setup's Report step uses too.
(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const h = (tag, attrs, ...kids) => {
    const e = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs || {})) { if (v == null || v === false) continue; if (k === 'class') e.className = v; else e.setAttribute(k, v === true ? '' : v); }
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  };
  let timer = null, current = 'current';

  async function api(path) {
    let r;
    try { r = await fetch(path, { headers: {} }); } catch (e) { return { ok: false, status: 0, data: {} }; }
    let data = {}; try { data = await r.json(); } catch (e) { /* not JSON (CSV, or an error page) */ }
    return { ok: r.ok, status: r.status, data };
  }

  async function loadList() {
    const r = await api('/api/operator/reports');
    const sel = $('which');
    const items = r.ok ? r.data.reports : [];
    sel.replaceChildren(...items.map((x) => h('option', { value: x.id }, OCReport.menuLabel(x))));
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
    $('none').hidden = true; $('report').hidden = false;
    $('report').replaceChildren(OCReport.build(r.data));
    $('csv').dataset.url = '/api/operator/report.csv' + (current === 'current' ? '' : '?id=' + encodeURIComponent(current));
    document.title = __('Audience report') + ' — ' + (r.data.report.show || r.data.report.showId);
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
  // CSV is a file: a link made on the spot and clicked, so the server names it and the page stays where it is
  $('csv').addEventListener('click', () => { const a = document.createElement('a'); a.href = $('csv').dataset.url; a.download = ''; document.body.append(a); a.click(); a.remove(); });
  $('back').addEventListener('click', () => { location.href = $('back').dataset.url; });

  async function start() { await loadList(); await load(); }
  start();
  if (window.OCI18n) OCI18n.apply(document.body);
})();
