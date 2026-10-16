// Setup page: a six-step wizard (Show, Show file, Images, Look, Test, Go Live).
(() => {
  'use strict';
  if (window.OC_UNSUPPORTED) return; // compat.js has already shown the "too old" message
  const $ = (id) => document.getElementById(id);
  const A = { shows: [], activeId: '', cur: null, step: 'show', text: '', saved: '', analysis: null, mainAnalysis: null, lang: '', loadedLang: null, langs: [], net: null, live: false, audience: 0, rehearsing: false };
  const STEPS = [['show', 'Shows'], ['showfile', 'Show file'], ['images', 'Images'], ['look', 'Look'], ['test', 'Test'], ['golive', 'Go Live']];
  const ICON = { caption: '💬', clear: '🧹', blank: '⬛', image: '🖼', fullscreen: '⛶', loop: '🔁', section: '📑', note: '📝' };

  // ---- helpers ----
  function h(tag, attrs, ...kids) {
    const e = document.createElement(tag);
    if (tag === 'button' && !(attrs && attrs.type)) e.setAttribute('type', 'button'); // never submit the dialog's form by accident
    for (const [k, v] of Object.entries(attrs || {})) {
      if (v === false || v == null) continue;
      if (k === 'class') e.className = v;
      else if (k.startsWith('on')) e.addEventListener(k.slice(2), v);
      else if (k === 'text') e.textContent = v;
      else e.setAttribute(k, v === true ? '' : v);
    }
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  }
  async function api(method, path, body, raw) {
    const opt = { method, headers: {} };
    if (raw) opt.body = raw; else if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
    const r = await fetch(path, opt);
    let data = {};
    try { data = await r.json(); } catch (e) { /* empty */ }
    if (r.status === 401) { showLogin(); throw new Error('login'); }
    return { ok: r.ok, status: r.status, data };
  }
  let toastTimer;
  function toast(msg, ms) { const t = $('toast'); t.textContent = __(msg); t.classList.add('show'); clearTimeout(toastTimer); toastTimer = setTimeout(() => t.classList.remove('show'), ms || 3000); }
  function fail(r, fallback) { toast((r.data && r.data.error) || fallback || __('Something went wrong')); }

  // A reusable dialog. body: node; resolves true on OK, false on cancel.
  function ask(title, body, okLabel, opts) {
    return new Promise((resolve) => {
      const d = $('dlg');
      d.classList.toggle('wide', !!(opts && opts.wide));
      $('dlgtitle').textContent = title;
      $('dlgbody').replaceChildren(body);
      $('dlgok').textContent = okLabel || __('OK');
      $('dlgok').classList.toggle('danger', !!(opts && opts.danger));
      $('dlgcancel').hidden = !!(opts && opts.noCancel);
      $('dlgcancel').onclick = () => d.close('cancel');
      const extra = $('dlgextra');
      extra.hidden = !(opts && opts.extra);
      if (opts && opts.extra) { extra.textContent = opts.extra.label; extra.onclick = () => { d.close('extra'); if (opts.extra.run) opts.extra.run(); }; }
      d.returnValue = '';
      d.onclose = () => resolve(d.returnValue === 'ok');
      d.showModal();
    });
  }
  const confirmBox = (title, text, ok, danger) => ask(title, h('p', {}, text), ok, { danger });
  // Line-by-line difference between two lists of lines: [{t: 'same' | 'add' | 'del', s}] (null if they are too big to compare).
  function lineDiff(a, b) {
    let p = 0; while (p < a.length && p < b.length && a[p] === b[p]) p++;
    let q = 0; while (q < a.length - p && q < b.length - p && a[a.length - 1 - q] === b[b.length - 1 - q]) q++;
    const A = a.slice(p, a.length - q), B = b.slice(p, b.length - q), N = A.length, M = B.length;
    if ((N + 1) * (M + 1) > 6e6) return null;
    const w = M + 1, dp = new Uint32Array((N + 1) * w);
    for (let i = N - 1; i >= 0; i--) for (let j = M - 1; j >= 0; j--) dp[i * w + j] = A[i] === B[j] ? dp[(i + 1) * w + j + 1] + 1 : Math.max(dp[(i + 1) * w + j], dp[i * w + j + 1]);
    const out = a.slice(0, p).map((s) => ({ t: 'same', s }));
    let i = 0, j = 0;
    while (i < N && j < M) {
      if (A[i] === B[j]) { out.push({ t: 'same', s: A[i] }); i++; j++; }
      else if (dp[(i + 1) * w + j] >= dp[i * w + j + 1]) out.push({ t: 'del', s: A[i++] });
      else out.push({ t: 'add', s: B[j++] });
    }
    while (i < N) out.push({ t: 'del', s: A[i++] });
    while (j < M) out.push({ t: 'add', s: B[j++] });
    return out.concat(a.slice(a.length - q).map((s) => ({ t: 'same', s })));
  }
  const fmtSize = (n) => (n >= 1 << 20 ? (n / (1 << 20)).toFixed(1) + ' MB' : Math.max(1, Math.round(n / 1024)) + ' KB');
  const curShow = () => A.shows.find((s) => s.id === A.cur && !s.archived) || null;
  const unarchived = () => A.shows.filter((s) => !s.archived);

  // ---- help drawer (topic follows the current step) ----
  const HELP_TOPIC = { show: 'admin-getting-started', showfile: 'admin-writing-a-show-file', images: 'admin-images', look: 'admin-look-and-feel', test: 'admin-getting-started', golive: 'admin-network' };
  function openHelp(topic) {
    $('help').hidden = false;
    $('help').querySelector('iframe').src = '/help?role=admin&embed=1&ui=' + OCI18n.lang + '&topic=' + (topic || HELP_TOPIC[A.step] || 'admin-getting-started');
  }
  function closeHelp() { $('help').hidden = true; $('help').querySelector('iframe').src = 'about:blank'; }
  $('helpbtn').addEventListener('click', () => ($('help').hidden ? openHelp() : closeHelp()));
  $('helpclose').addEventListener('click', closeHelp);
  window.addEventListener('message', (e) => { if (e.origin === location.origin && e.data && e.data.type === 'help-close') closeHelp(); });
  document.addEventListener('keydown', (e) => { if (e.key === 'F1') { e.preventDefault(); openHelp(); } });

  // ---- settings: passwords and where things are ----
  async function openSettings() {
    const r = await api('GET', '/api/admin/settings'); if (!r.ok) return fail(r);
    const s = r.data;
    const adminPw = h('input', { type: 'password', autocomplete: 'new-password', 'aria-label': __('New admin password'), minlength: '6' });
    const opPw = h('input', { type: 'password', autocomplete: 'new-password', 'aria-label': __('New operator password'), minlength: '6' });
    const shown = h('div', { class: 'hint', role: 'status' });
    const reveal = h('button', { type: 'button', class: 'quiet' }, __('Show the current operator password'));
    reveal.addEventListener('click', async () => { // the same button shows it and hides it again
      if (shown.textContent) { shown.textContent = ''; reveal.textContent = __('Show the current operator password'); return; }
      const x = await api('GET', '/api/admin/settings/operator-password');
      shown.textContent = x.ok ? __('Operator password: {password}', { password: x.data.password }) : __('Could not read it');
      if (x.ok) reveal.textContent = __('Hide the operator password');
    });
    const changePw = h('button', { type: 'button', onclick: async () => {
      if (!adminPw.value && !opPw.value) { adminPw.focus(); toast(__('Enter a new password in at least one box.')); return; }
      const res = await api('PUT', '/api/admin/settings/passwords', { admin: adminPw.value, operator: opPw.value });
      if (!res.ok) return fail(res);
      adminPw.value = opPw.value = ''; shown.textContent = ''; reveal.textContent = __('Show the current operator password');
      toast(__('Passwords changed'));
    } }, __('Change passwords'));
    const row = (k, v) => h('tr', {}, h('td', {}, __(k)), h('td', {}, v ? __(v) : '—'));
    const m = s.mdns || {};
    const mdnsOn = h('input', { type: 'checkbox', 'aria-label': __('Announce this computer by name') }); mdnsOn.checked = !!m.enabled;
    const mdnsName = h('input', { type: 'text', class: 'netname', maxlength: '63', size: '16', 'aria-label': __('Name'), autocapitalize: 'none', spellcheck: 'false' }); mdnsName.value = m.requested || 'open-caption';
    const mdnsMsg = h('div', { class: 'hint', role: 'status' });
    const showMdns = (x) => {
      mdnsMsg.className = x.error ? 'check warn' : x.renamed ? 'check warn' : 'hint';
      mdnsMsg.textContent = !x.enabled ? __('Not announced.') : x.error ? __('Could not announce this name: {error}', { error: x.error })
        : x.renamed ? __('Another computer on this network already uses {name}.local, so this one is http://{host}/', { name: x.requested, host: x.host })
        : x.active ? __('Devices on this network can open http://{host}/', { host: x.host }) : __('Starting…');
    };
    showMdns(m);
    async function applyMdns() {
      mdnsMsg.className = 'hint'; mdnsMsg.textContent = __('Checking that nobody else uses the name…');
      const x = await api('PUT', '/api/admin/settings/mdns', { enabled: mdnsOn.checked, name: mdnsName.value });
      if (!x.ok) { mdnsMsg.className = 'check warn'; mdnsMsg.textContent = __((x.data && x.data.error) || 'Could not change the name'); return; }
      showMdns(x.data.mdns); if (x.data.mdns.active) mdnsName.value = x.data.mdns.requested;
    }
    const logSelect = h('select', { 'aria-label': __('Log detail'), onchange: async (e) => {
      const x = await api('PUT', '/api/admin/settings/log-level', { level: e.target.value });
      if (!x.ok) { fail(x); e.target.value = s.logLevel; return; }
      s.logLevel = x.data.level; toast(x.data.note || __('Log detail: {level}', { level: x.data.level }), x.data.note ? 9000 : undefined);
    } }, ...[['error', 'Errors only'], ['warn', 'Problems (warnings and errors)'], ['info', 'Normal (events, sign-ins)'], ['debug', 'Debug (every request)']].map(([v, t]) => h('option', { value: v }, __(t))));
    logSelect.value = s.logLevel;
    const sec = (title, ...kids) => h('section', { class: 'setsec' }, h('h3', {}, __(title)), ...kids);
    const body = h('div', {},
      sec('About this copy',
        h('table', { class: 'cheat' },
          row('Program', 'open-caption ' + s.version), row('Settings file', s.configPath), row('Shows and data', s.dataDir),
          row('Setup page can be opened from', s.adminRestricted ? s.adminAllow.join(', ') : 'any computer on the network'),
          row('Operator page can be opened from', s.operatorRestricted ? s.operatorAllow.join(', ') : 'any computer on the network'),
          row('Web port (HTTP)', (s.httpAddr || '').replace(/^.*:/, '')), row('Web port (HTTPS)', (s.httpsAddr || '').replace(/^.*:/, ''))),
        h('p', { class: 'hint' }, __('These are fixed in the settings file. To change them, edit that file and start the program again.')),
        !s.adminRestricted && !s.httpsAddr ? h('p', { class: 'check warn' }, __('Anyone on this WiFi can try to sign in to Setup, and passwords travel unencrypted over plain http. To limit it, set access: admin_allow in the settings file (for example [localhost]); see the usage guide.')) : null),
      sec('Network name',
        h('p', { class: 'hint' }, __('Devices can open this computer by name instead of a number. It needs no router settings, but some guest WiFi networks block it, so keep the numeric address and QR code as the main way in.')),
        h('div', { class: 'row' }, mdnsOn, __(' Announce as '), mdnsName, __('.local '), h('button', { type: 'button', onclick: applyMdns }, __('Apply'))), mdnsMsg),
      sec('Log detail',
        logSelect,
        s.logLevelPinned ? h('p', { class: 'check warn' }, __('This copy was started with --log-level {level}, which overrides the settings file every time it starts. A change made here applies now but lasts only until the program restarts.', { level: s.logLevelPinned })) : null,
        h('p', { class: 'hint' }, s.canChangePasswords ? __('How much the program records (in the support bundle and in the log file, if one is kept). Info records sign-ins and events; Debug adds every request, for tracking down a problem. Applies at once and is kept in the settings file.') : __('How much the program records (in the support bundle and in the log file, if one is kept). Info records sign-ins and events; Debug adds every request, for tracking down a problem. Applies at once.'))),
      sec('Change passwords',
        s.canChangePasswords
          ? h('div', {}, h('p', { class: 'hint' }, __('Leave a box empty to keep that password. At least 6 characters. People signed in with a changed password are signed out; you stay signed in.')),
              h('label', { class: 'field' }, h('span', {}, __('New admin password')), adminPw),
              h('label', { class: 'field' }, h('span', {}, __('New operator password')), opPw),
              h('div', { class: 'row' }, changePw, reveal), shown)
          : h('p', { class: 'hint' }, __('This copy was started without a settings file, so passwords cannot be changed here.'))));
    // Each section above has its own button (or applies as soon as you choose); this one only closes the window.
    await ask(__('Settings'), body, __('Close'), { noCancel: true, wide: true });
  }
  // The program's recent log (the last 1000 lines it keeps in memory), so a problem can be looked at without
  // opening the log file or the support bundle.
  async function showLog() {
    const pre = h('pre', { class: 'logview', tabindex: '0', 'aria-label': __('Recent log') });
    const level = h('select', { 'aria-label': __('Show') }, ...[['debug', 'Everything'], ['info', 'Normal and up'], ['warn', 'Problems only (warnings and errors)'], ['error', 'Errors only']].map(([v, t]) => h('option', { value: v }, __(t))));
    level.value = 'info';
    const auto = h('input', { type: 'checkbox', checked: '' });
    const count = h('span', { class: 'hint' });
    let lastLines = [];
    async function load() {
      const r = await api('GET', '/api/admin/log?lines=400&level=' + level.value);
      if (!r.ok) { pre.textContent = __((r.data && r.data.error) || 'Could not read the log.'); return; }
      lastLines = r.data.lines;
      const atEnd = pre.scrollHeight - pre.scrollTop - pre.clientHeight < 40;
      pre.textContent = lastLines.length ? lastLines.join('\n') : __('Nothing at this level yet.');
      count.textContent = __('{n} line(s); the program logs at "{level}" and above', { n: lastLines.length, level: r.data.level });
      if (atEnd) pre.scrollTop = pre.scrollHeight;
    }
    level.addEventListener('change', load);
    const copy = h('button', { onclick: async () => { try { await navigator.clipboard.writeText(lastLines.join('\n')); toast(__('Copied')); } catch (e) { toast(__('Select the text and copy it by hand')); } } }, __('Copy'));
    const body = h('div', {}, h('div', { class: 'row' }, h('label', { class: 'row' }, __('Show: '), level), h('label', { class: 'row' }, auto, __(' Update every few seconds')), h('button', { onclick: load }, __('Refresh')), copy),
      pre, count, h('p', { class: 'hint' }, __('Passwords are never written to the log. Setup → Go Live has a support bundle with this log and more, for sending to someone who is helping.')));
    await load(); pre.scrollTop = pre.scrollHeight;
    const timer = setInterval(() => { if (auto.checked) load(); }, 3000);
    await ask(__('Recent log'), body, __('Close'), { noCancel: true, wide: true });
    clearInterval(timer);
  }
  $('settingsbtn').addEventListener('click', openSettings);

  // ---- login ----
  function showLogin() { $('app').hidden = true; $('login').hidden = false; $('pw').focus(); }
  $('login').addEventListener('submit', async (e) => {
    e.preventDefault();
    const r = await fetch('/api/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ role: 'admin', password: $('pw').value }) });
    if (r.ok) { $('loginerr').textContent = ''; $('pw').value = ''; boot(); return; }
    $('loginerr').textContent = r.status === 429 ? __('Too many attempts. Wait a minute and try again.') : __('That password was not accepted.');
  });
  $('logout').addEventListener('click', async () => { await fetch('/api/logout', { method: 'POST' }); showLogin(); });

  // ---- boot ----
  async function boot() {
    const ping = await fetch('/api/admin/ping');
    if (!ping.ok) { showLogin(); return; }
    $('login').hidden = true; $('app').hidden = false;
    await loadShows();
    try { A.cur = localStorage.getItem('oc.cur'); } catch (e) { /* storage unavailable */ }
    if (!curShow()) A.cur = A.activeId && curShowById(A.activeId) ? A.activeId : (unarchived()[0] ? unarchived()[0].id : null);
    A.step = A.cur ? (location.hash.slice(1) || 'show') : 'show';
    if (!STEPS.some((s) => s[0] === A.step)) A.step = 'show';
    openOpStream();
    render();
  }
  const curShowById = (id) => A.shows.find((s) => s.id === id);
  async function loadShows() {
    const r = await api('GET', '/api/admin/shows');
    if (r.ok) { A.shows = r.data.shows; A.activeId = r.data.activeId; }
  }
  function setCur(id) { A.cur = id; A.rehearsing = false; A.text = A.saved = ''; A.analysis = A.mainAnalysis = null; A.lang = ''; A.loadedLang = null; A.langs = []; try { localStorage.setItem('oc.cur', id || ''); } catch (e) { /* ignore */ } }

  // live status via the operator stream (admins may use it)
  function openOpStream() {
    const es = new EventSource('/api/operator/stream');
    es.addEventListener('opstate', (e) => {
      const p = JSON.parse(e.data);
      const wasLive = A.live;
      A.live = p.op.live; A.audience = p.audience; A.activeId = p.show.id;
      $('viewers').textContent = '👥 ' + p.audience;
      $('live').hidden = !p.op.live;
      if (A.step === 'golive') { const v = document.getElementById('audcount'); if (v) v.textContent = String(p.audience); }
      if (wasLive !== A.live && (A.step === 'show' || A.step === 'golive')) loadShows().then(() => A.step === 'show' && render());
    });
    es.onerror = () => { es.close(); setTimeout(openOpStream, 2000); };
  }

  // ---- shell ----
  function go(step) { A.step = step; location.hash = step; render(); }
  window.addEventListener('hashchange', () => { const s = location.hash.slice(1); if (s !== A.step && STEPS.some((x) => x[0] === s)) { A.step = s; render(); } });

  // "Look" has no wrong answer, so it is ticked once the person has saved it for this show.
  const lookSaved = () => { const s = curShow(); return !!(s && s.lookSaved); };
  const markLookSaved = () => { const s = curShow(); if (s) s.lookSaved = true; refreshNav(); };

  function done(step) {
    switch (step) {
      case 'look': return lookSaved();
      case 'test': return A.rehearsing;
      case 'golive': return !!A.cur && A.activeId === A.cur;
      case 'show': return !!A.cur;
      case 'showfile': return !!(A.mainAnalysis && A.mainAnalysis.cues.length && !A.mainAnalysis.diagnostics.some((d) => d.severity === 'error'));
      case 'images': return !!(A.mainAnalysis && !(A.mainAnalysis.missing || []).length && (A.mainAnalysis.assets || []).length);
      default: return false;
    }
  }

  function refreshNav() {
    const nav = $('steps'); nav.replaceChildren();
    STEPS.forEach(([id, rawLabel], i) => { const label = __(rawLabel);
      nav.append(h('button', { class: (A.step === id ? 'on ' : '') + (done(id) ? 'done' : ''), onclick: () => go(id), disabled: id !== 'show' && !A.cur, 'aria-current': A.step === id ? 'step' : false },
        h('span', { class: 'n' }, h('span', {}, String(i + 1))), label));
    });
  }

  // Ticks for Show file and Images depend on the show file analysis, so load it in the background when not yet known.
  async function warmAnalysis() {
    const id = A.cur;
    if (!id || A.mainAnalysis) return;
    const r = await api('GET', '/api/admin/shows/' + id + '/showfile');
    if (r.ok && A.cur === id && !A.mainAnalysis) { A.mainAnalysis = r.data.analysis; refreshNav(); }
  }

  async function render() {
    const s = curShow();
    $('working').textContent = s ? s.name : __('No show selected');
    $('working').hidden = false;
    refreshNav();
    warmAnalysis();
    const panel = $('panel'); panel.replaceChildren();
    document.title = __('Setup') + ' — ' + __((STEPS.find((x) => x[0] === A.step) || ['', ''])[1]);
    try { await ({ show: stepShow, showfile: stepShowFile, images: stepImages, look: stepLook, test: stepTest, golive: stepGoLive })[A.step](panel); } catch (e) { if (e.message !== 'login') console.error(e); }
  }

  // ---- step 1: show ----
  async function stepShow(p) {
    await loadShows();
    p.append(h('h2', {}, __('Choose a show')));
    const current = unarchived(), archived = A.shows.filter((s) => s.archived);
    if (!A.shows.length) {
      p.append(h('div', { class: 'card empty' },
        h('p', { class: 'big' }, __('Welcome! Start by creating a show, or look around with the sample.')),
        h('div', { class: 'row' }, ...newShowForm(), h('button', { onclick: makeSample }, __('Create a sample show')))));
    } else {
      if (!current.length) p.append(h('div', { class: 'card empty' }, h('p', { class: 'big' }, __('All your shows are archived. Create a new show below, or restore one from the archive.'))));
      current.forEach((s) => p.append(h('div', { class: 'card show' },
        h('div', { class: 'name' }, s.name, ' ', s.active ? h('span', { class: 'badge active' }, __('ACTIVE')) : null, ' ', s.live ? h('span', { class: 'badge live' }, __('LIVE')) : null,
          h('div', { class: 'hint' }, __('Updated {when}', { when: new Date(s.updated).toLocaleString(OCI18n.lang) }))),
        h('button', { class: A.cur === s.id ? '' : 'primary', onclick: () => { setCur(s.id); go('showfile'); } }, A.cur === s.id ? __('Continue with this show') : __('Work on this show')),
        h('button', { class: 'quiet', onclick: () => renameShow(s) }, __('Rename…')),
        h('button', { class: 'quiet', onclick: () => duplicateShow(s) }, __('Duplicate')),
        h('a', { class: 'quiet', href: '/api/admin/shows/' + s.id + '/export', download: s.id + '.zip' }, __('Export')),
        h('button', { class: 'quiet', disabled: s.active, title: s.active ? __('The show on air cannot be archived. Activate another show first.') : __('Put this show away; you can restore it later'), onclick: () => archiveShow(s, true) }, __('Archive')),
        h('button', { class: 'danger quiet', onclick: () => deleteShow(s) }, __('Delete…')))));
      p.append(h('h3', {}, __('Add another show')), h('div', { class: 'card row' }, ...newShowForm(), h('button', { onclick: makeSample }, __('Sample show')),
        h('label', { class: 'row' }, __('Import a show (a .zip, or a plain .md / .txt show file):'), h('input', { type: 'file', accept: '.zip,.md,.txt,text/plain,text/markdown', 'aria-label': __('Import a show (a .zip, or a plain .md / .txt show file):'), onchange: importShow }))));
    }
    if (archived.length) {
      p.append(h('details', { class: 'archive', open: A.showArchive ? '' : null, ontoggle: (e) => { A.showArchive = e.target.open; } },
        h('summary', {}, __('Archived shows ({n})', { n: archived.length })),
        h('p', { class: 'hint' }, __('Archived shows are kept in full but stay out of the way and cannot go on air. Restore one to use it again, or delete it for good.')),
        ...archived.map((s) => h('div', { class: 'card show' },
          h('div', { class: 'name' }, s.name, h('div', { class: 'hint' }, __('Updated {when}', { when: new Date(s.updated).toLocaleString(OCI18n.lang) }))),
          h('button', { class: 'primary', onclick: () => archiveShow(s, false) }, __('Restore')),
          h('a', { class: 'quiet', href: '/api/admin/shows/' + s.id + '/export', download: s.id + '.zip' }, __('Export')),
          h('button', { class: 'danger quiet', onclick: () => deleteShow(s) }, __('Delete…'))))));
    }
  }
  function newShowForm() {
    const inp = h('input', { type: 'text', placeholder: __('Name of the show'), maxlength: '100', 'aria-label': __('Show name') });
    const create = async () => {
      if (!inp.value.trim()) { inp.focus(); return; }
      const r = await api('POST', '/api/admin/shows', { name: inp.value });
      if (!r.ok) return fail(r);
      await loadShows(); setCur(r.data.id); go('showfile');
    };
    inp.addEventListener('keydown', (e) => { if (e.key === 'Enter') create(); });
    return [inp, h('button', { class: 'primary', onclick: create }, __('Create show'))]; // two controls for the caller's row, so they line up with its other buttons
  }
  async function makeSample() {
    const r = await api('POST', '/api/admin/shows', { sample: true });
    if (!r.ok) return fail(r);
    await loadShows(); setCur(r.data.id); toast(__('Sample show created')); go('showfile');
  }
  async function importShow(e) {
    const f = e.target.files[0]; if (!f) return;
    const fd = new FormData(); fd.append('file', f);
    const r = await api('POST', '/api/admin/shows/import', undefined, fd);
    if (!r.ok) return fail(r, __('That file could not be imported'));
    await loadShows(); setCur(r.data.show.id);
    toast(r.data.skippedFiles ? __('Imported ({n} file(s) skipped)', { n: r.data.skippedFiles }) : __('Imported')); go('showfile');
  }
  async function renameShow(s) {
    const inp = h('input', { type: 'text', maxlength: '100', 'aria-label': __('Show name') });
    inp.value = s.name;
    if (!await ask(__('Rename the show'), h('label', { class: 'field' }, h('span', {}, __('New name')), inp), __('Rename'))) return;
    const r = await api('PUT', '/api/admin/shows/' + s.id, { name: inp.value });
    if (!r.ok) return fail(r);
    toast(__('Renamed')); render();
  }
  async function duplicateShow(s) {
    const inp = h('input', { type: 'text', maxlength: '100', 'aria-label': __('Name of the copy') });
    inp.value = __('{name} (copy)', { name: s.name });
    if (!await ask(__('Duplicate the show'), h('div', {}, h('p', {}, __('Makes a copy of the show file, images and look. Nothing else is shared.')), h('label', { class: 'field' }, h('span', {}, __('Name of the copy')), inp)), __('Duplicate'))) return;
    const r = await api('POST', '/api/admin/shows/' + s.id + '/duplicate', { name: inp.value });
    if (!r.ok) return fail(r);
    await loadShows(); setCur(r.data.id); toast(__('Copy created')); render();
  }
  async function archiveShow(s, on) {
    const r = await api('POST', '/api/admin/shows/' + s.id + (on ? '/archive' : '/unarchive'));
    if (!r.ok) return fail(r);
    if (on && A.cur === s.id) setCur(null);
    await loadShows();
    if (!A.cur && unarchived()[0]) setCur(unarchived()[0].id);
    if (on) A.showArchive = true;
    toast(on ? __('“{name}” archived', { name: s.name }) : __('“{name}” restored', { name: s.name }));
    render();
  }
  async function deleteShow(s) {
    const ok = await confirmBox(__('Delete “{name}”?', { name: s.name }), s.archived ? __('This archived show, its show file and its images will be permanently deleted. Export it first if you may want it back.') : __('Its show file and images will be permanently deleted. Export it first if you may want it back. (Archive keeps a show without deleting it.)'), __('Delete'), true);
    if (!ok) return;
    const r = await api('DELETE', '/api/admin/shows/' + s.id);
    if (!r.ok) return fail(r);
    if (A.cur === s.id) setCur(null);
    await loadShows(); if (!A.cur && unarchived()[0]) setCur(unarchived()[0].id);
    render();
  }

  // ---- step 2: show file ----
  const SNIPPETS = {
    clear: '[clear]\n', blank: '[blank]\n', fullscreen: '[fullscreen: INTERMISSION]\n', section: '[section: Act 1, Scene 1]\n',
    note: '[note: reminder for the operator]\n', comment: '<!-- a comment nobody sees -->\n',
    loop: '[loop: 10s]\n[fullscreen: Welcome!]\nPlease silence your phones\n[/loop]\n',
    settings: '[settings]\ntitle: The name of my show\nfont: system\ntext size: 100%\n[/settings]\n',
    together: '[together]\nANNA: ~ first voice ~\nKRIS: ~ second voice sung at the same time ~\n[/together]\n',
    song: '[song: Title of the song]\nFirst line of the lyrics\nSecond line of the lyrics\n[/song]\n',
  };
  async function stepShowFile(p) {
    const show = curShow(); if (!show) return go('show');
    // Languages: the tabs above the editor. A.lang is '' for the main language.
    const lr = await api('GET', '/api/admin/shows/' + show.id + '/languages');
    if (!lr.ok) return fail(lr);
    A.langs = lr.data.languages; const known = lr.data.known;
    const main = A.langs.find((l) => l.primary) || { code: 'en', name: 'English' };
    if (A.lang && !A.langs.some((l) => l.code === A.lang)) A.lang = '';
    const cur = A.langs.find((l) => l.code === (A.lang || main.code)) || main;
    const q = A.lang ? '?lang=' + encodeURIComponent(A.lang) : '';
    const dirty = A.text !== A.saved;
    if (!A.analysis || A.loadedLang !== A.lang || !dirty) { // keep unsaved edits when coming back to this step
      const r = await api('GET', '/api/admin/shows/' + show.id + '/showfile' + q);
      if (!r.ok) return fail(r);
      A.text = A.saved = r.data.text; A.analysis = r.data.analysis; A.version = r.data.version; A.loadedLang = A.lang;
      if (!A.lang) A.mainAnalysis = A.analysis;
    }
    const ta = h('textarea', { spellcheck: 'false', 'aria-label': __('Show file'), wrap: 'off', placeholder: __('Paste your show file here, or drop a .txt / .md file onto this box.\n\nEach line becomes one caption.') });
    ta.value = A.text;
    const gutter = h('pre', { class: 'gutter', 'aria-hidden': 'true' });
    const status = h('span', { class: 'hint' });
    const diagList = h('ul', { class: 'diags', tabindex: '0', 'aria-label': __('Problems and notes') }), cueList = h('ul', { class: 'cues', tabindex: '0', 'aria-label': __('What the operator will step through') }), diagTitle = h('h3', {}, __('Problems and notes'));
    const editor = h('div', { class: 'editor' }, gutter, ta);
    const saveBtn = h('button', { class: 'primary', onclick: save }, __('Save'));

    function lines() { const n = ta.value.split('\n').length; gutter.textContent = Array.from({ length: n }, (_, i) => i + 1).join('\n'); }
    function setStatus() {
      const dirty = ta.value !== A.saved;
      status.textContent = dirty ? __('Unsaved changes') : __('All changes saved');
      status.className = dirty ? 'status-warn' : 'status-ok';
      saveBtn.disabled = !dirty;
    }
    function showAnalysis() {
      const a = A.analysis;
      const shown = a.diagnostics.filter((d) => d.line > 0);
      const summary = a.diagnostics.find((d) => d.line === 0);
      diagTitle.textContent = shown.length ? __('Problems and notes ({n})', { n: shown.length }) : __('No problems found ✓');
      diagList.replaceChildren(...shown.map((d) => h('li', { class: d.severity, onclick: () => jump(d.line) }, h('span', { class: 'ln' }, __('Line {n}', { n: d.line })), __(d.message))),
        ...(summary ? [h('li', { class: 'info' }, __(summary.message))] : []));
      cueList.replaceChildren(...a.cues.map((c) => h('li', { class: c.kind === 'section' || c.kind === 'note' ? 'op' : '', title: __('Line {n}', { n: c.line }) },
        h('span', { class: 'k' }, c.sung ? '🎵' : ICON[c.kind] || ''), c.kind === 'caption' ? (c.parts && c.parts.length ? '⏸ ' + c.parts.map((v) => (v.speaker ? v.speaker + ': ' : '') + v.text).join('  ║  ') : (c.speaker ? c.speaker + ': ' : '') + c.text.replace(/\n/g, ' ⏎ ')) : (c.label || c.text || c.kind.toUpperCase()))));
    }
    function jump(line) {
      const rows = ta.value.split('\n'); let pos = 0;
      for (let i = 0; i < line - 1 && i < rows.length; i++) pos += rows[i].length + 1;
      ta.focus(); ta.setSelectionRange(pos, pos + (rows[line - 1] || '').length);
      ta.scrollTop = Math.max(0, (line - 4) * parseFloat(getComputedStyle(ta).lineHeight));
      gutter.scrollTop = ta.scrollTop;
    }
    let timer;
    function changed() {
      A.text = ta.value; lines(); setStatus();
      clearTimeout(timer);
      timer = setTimeout(async () => {
        const r = await api('POST', '/api/admin/shows/' + show.id + '/showfile/validate' + q, { text: ta.value });
        if (r.ok && A.step === 'showfile') { A.analysis = r.data.analysis; showAnalysis(); }
      }, 400);
    }
    async function save(confirmName) {
      const body = { text: ta.value, baseVersion: A.version }; if (typeof confirmName === 'string') body.confirm = confirmName;
      const r = await api('PUT', '/api/admin/shows/' + show.id + '/showfile' + q, body);
      if (r.status === 409 && r.data.needConfirm) {
        const inp = h('input', { type: 'text', 'aria-label': __('Show name'), placeholder: r.data.showName });
        const ok = await ask(__('This show is live'), h('div', {}, h('p', {}, __('Changing the show file now affects what the operator sees. The current position is kept by line number. To continue, type the show’s name:')), h('p', {}, h('b', {}, r.data.showName)), inp), __('Save change'));
        if (ok) return save(inp.value);
        return;
      }
      if (r.status === 409 && r.data.conflict) {
        // Someone else saved this show file after it was opened here. Never overwrite silently.
        const theirs = r.data;
        const choice = await ask(__('This show file was changed by someone else'),
          h('div', {}, h('p', {}, __('Another person (or another window) saved this show file after you opened it. If you save yours, their changes are lost.')),
            h('p', { class: 'hint' }, __('“Use their version” replaces what is in the editor with theirs. Copy anything you want to keep first.'))),
          __('Save mine anyway'), { danger: true, extra: { label: __('Use their version'), run: () => {
            ta.value = theirs.text; A.version = theirs.version; A.saved = theirs.text; changed(); toast(__('Loaded their version'));
          } } });
        if (choice) { A.version = theirs.version; return save(confirmName); }
        return;
      }
      if (!r.ok) return fail(r);
      A.saved = ta.value; A.version = r.data.version; A.analysis = r.data.analysis; if (!A.lang) A.mainAnalysis = A.analysis; setStatus(); showAnalysis(); refreshNav();
      if (r.data.skippedLanguages && r.data.skippedLanguages.length) toast(__('Saved. These languages are not shown to the audience until their structure matches: {list}', { list: r.data.skippedLanguages.join(', ') }));
      else
      toast(r.data.countChanged ? __('Saved. The number of lines changed, so check the operator’s position.') : __('Saved'));
      refreshLangs();
    }
    // Every save keeps the version it replaced (the last 50). Choosing one only fills the editor: nothing is
    // changed until Save is pressed, so it goes through the same live-show and conflict checks.
    async function showHistory() {
      const base = '/api/admin/shows/' + show.id + '/showfile/history';
      const r = await api('GET', base + q);
      if (!r.ok) return fail(r);
      let versions = r.data.versions;
      const list = h('ul', { class: 'versions' }), diffBox = h('div', { class: 'diffbox' });
      const when = (v) => new Date(v.time).toLocaleString(OCI18n.lang);
      const text = async (v) => { const t = await api('GET', base + '/' + v.id + q); if (!t.ok) { fail(t); return null; } return t.data.text; };
      const load = async (v) => {
        const t = await text(v); if (t === null) return;
        $('dlg').close('cancel');
        ta.value = t; changed(); toast(__('Loaded the version from {when}. Press Save to keep it.', { when: when(v) }));
      };
      // What loading the version would change in the editor, line by line.
      const compare = async (v) => {
        const t = await text(v); if (t === null) return;
        const d = lineDiff(ta.value.split('\n'), t.split('\n'));
        if (!d) { diffBox.replaceChildren(h('p', { class: 'hint' }, __('The two versions are too different to compare line by line.'))); return; }
        const added = d.filter((x) => x.t === 'add').length, removed = d.filter((x) => x.t === 'del').length;
        if (!added && !removed) { diffBox.replaceChildren(h('p', { class: 'hint' }, __('Identical to what is in the editor now.'))); return; }
        const rows = [];
        for (let i = 0; i < d.length;) {
          if (d[i].t !== 'same') { rows.push(h('div', { class: 'd-' + d[i].t }, (d[i].t === 'add' ? '+ ' : '− ') + d[i].s)); i++; continue; }
          let j = i; while (j < d.length && d[j].t === 'same') j++;
          const run = d.slice(i, j), first = i === 0, last = j === d.length;
          const keepHead = first ? 0 : 2, keepTail = last ? 0 : 2;
          if (run.length > keepHead + keepTail + 2) {
            run.slice(0, keepHead).forEach((x) => rows.push(h('div', { class: 'd-same' }, '  ' + x.s)));
            rows.push(h('div', { class: 'd-skip' }, __('… {n} unchanged line(s) …', { n: run.length - keepHead - keepTail })));
            run.slice(run.length - keepTail).forEach((x) => rows.push(h('div', { class: 'd-same' }, '  ' + x.s)));
          } else run.forEach((x) => rows.push(h('div', { class: 'd-same' }, '  ' + x.s)));
          i = j;
        }
        diffBox.replaceChildren(h('p', { class: 'hint', role: 'status' }, __('Loading this version would add {added} line(s) and remove {removed} line(s).', { added, removed })), h('pre', { class: 'diff', tabindex: '0', 'aria-label': __('Compare') }, ...rows));
      };
      const paint = () => {
        list.replaceChildren(...versions.map((v) => {
          const name = h('input', { type: 'text', maxlength: '80', placeholder: __('Name this version…'), 'aria-label': __('Name this version…'), class: 'vname' });
          name.value = v.label || '';
          name.addEventListener('change', async () => {
            const x = await api('PUT', base + '/' + v.id + '/label' + q, { label: name.value });
            if (!x.ok) { fail(x); name.value = v.label || ''; return; }
            v.label = name.value.trim(); toast(__('Saved'));
          });
          return h('li', {}, h('button', { onclick: () => load(v) }, __('Load')), ' ', h('button', { onclick: () => compare(v) }, __('Compare')), ' ',
            h('span', {}, when(v) + ' · ' + __('{n} lines', { n: v.lines })), ' ', name);
        }));
      };
      paint();
      const keepName = h('input', { type: 'text', maxlength: '80', placeholder: __('Keep the saved version as…'), 'aria-label': __('Keep the saved version as…'), class: 'vname' });
      const keep = async () => {
        if (!keepName.value.trim()) { keepName.focus(); return; }
        const x = await api('POST', base + q, { label: keepName.value });
        if (!x.ok) return fail(x);
        versions = [x.data].concat(versions); keepName.value = ''; paint(); toast(__('Saved'));
      };
      keepName.addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); keep(); } });
      const body = h('div', {},
        h('p', { class: 'hint' }, __('Each time you save, the version it replaced is kept (the last 50). Choosing one puts it in the editor; nothing changes until you press Save.')),
        h('p', { class: 'hint' }, __('Named versions are never forgotten. Use Keep to name what is saved now (for example “Opening night”), or name any version below. Compare shows what loading it would change.')),
        h('div', { class: 'row' }, keepName, h('button', { onclick: keep }, __('Keep'))),
        versions.length ? list : h('p', {}, __('There are no earlier versions yet. After the next save, the version it replaces will be listed here.')), diffBox);
      list.hidden = false;
      await ask(A.lang ? __('Earlier versions ({name})', { name: cur.name }) : __('Earlier versions'), body, __('Close'), { noCancel: true, wide: true });
    }
    async function refreshLangs() {
      const x = await api('GET', '/api/admin/shows/' + show.id + '/languages');
      if (x.ok) { A.langs = x.data.languages; renderTabs(); }
    }
    function insert(text) {
      const s = ta.selectionStart, e = ta.selectionEnd, atLineStart = s === 0 || ta.value[s - 1] === '\n';
      const t = (atLineStart ? '' : '\n') + text;
      ta.setRangeText(t, s, e, 'end'); ta.focus(); changed();
    }
    function insertInline(text) { ta.setRangeText(text, ta.selectionStart, ta.selectionEnd, 'end'); ta.focus(); changed(); }
    function loadText(text) { ta.value = text; changed(); }
    // Read a file the preparer made elsewhere: UTF-8 (with or without a mark), else Windows-1252, which is
    // what Notepad's old "ANSI" saves.
    async function readFile(f) {
      if (!f) return;
      const buf = await f.arrayBuffer(); let text;
      try { text = new TextDecoder('utf-8', { fatal: true }).decode(buf); } catch (e) { text = new TextDecoder('windows-1252').decode(buf); }
      loadText(text.replace(/^\uFEFF/, ''));
      toast(__('Loaded {name}. Check the problems list, then press Save.', { name: f.name }), 7000);
    }

    ta.addEventListener('input', changed);
    ta.addEventListener('scroll', () => { gutter.scrollTop = ta.scrollTop; });
    ta.addEventListener('keydown', (e) => { if ((e.metaKey || e.ctrlKey) && e.key === 's') { e.preventDefault(); if (!saveBtn.disabled) save(); } });
    editor.addEventListener('dragover', (e) => { e.preventDefault(); editor.classList.add('drop'); });
    editor.addEventListener('dragleave', () => editor.classList.remove('drop'));
    editor.addEventListener('drop', (e) => { e.preventDefault(); editor.classList.remove('drop'); readFile(e.dataTransfer.files[0]); });

    const tools = h('div', { class: 'tools' },
      ...[['Clear', 'clear'], ['Blank', 'blank'], ['Fullscreen text', 'fullscreen'], ['Section', 'section'], ['Operator note', 'note'], ['Comment', 'comment'], ['Loop', 'loop']].map(([l, k]) => h('button', { onclick: () => insert(SNIPPETS[k]) }, __(l))),
      h('button', { onclick: pickImage }, __('Image…')),
      h('button', { onclick: insertSettings, title: __('Title, colors, font, size and other options for the whole show') }, __('Settings')),
      h('span', { class: 'tools-sep', 'aria-hidden': 'true' }, '|'),
      h('button', { onclick: () => insert(SNIPPETS.song), title: __('A song: every line becomes a sung lyric') }, __('🎵 Song')),
      h('button', { onclick: () => insert(SNIPPETS.together), title: __('Two or more singers at the same time') }, __('Together')),
      h('button', { onclick: () => insertInline('♪ '), title: __('Insert a music note'), 'aria-label': __('Insert a music note') }, '♪'),
      h('button', { onclick: () => insertInline('♫ '), title: __('Insert beamed music notes'), 'aria-label': __('Insert beamed music notes') }, '♫'));
    // The [settings] block goes at the top (after a [language: …] line), and only once.
    function insertSettings() {
      const at = ta.value.search(/^\s*\[settings\]/im);
      if (at >= 0) { jump(ta.value.slice(0, at).split('\n').length + (ta.value[at] === '\n' ? 1 : 0)); toast(__('This show file already has a settings block')); return; }
      const lang = ta.value.match(/^\s*\[language[^\]]*\]\s*$/m);
      const pos = lang && lang.index < 400 ? lang.index + lang[0].length + 1 : 0;
      const block = A.lang ? '[settings]\nidle message: Captions will begin shortly\n[/settings]\n\n' : SNIPPETS.settings + '\n';
      ta.setRangeText(block, pos, pos, 'end'); ta.focus(); changed();
    }
    async function pickImage() {
      const names = (A.analysis.assets || []).map((a) => a.name);
      if (!names.length) { toast(__('Upload images in the Images step first')); return; }
      const sel = h('select', {}, ...names.map((n) => h('option', { value: n }, n)));
      if (await ask(__('Insert an image'), h('label', { class: 'field' }, h('span', {}, __('Which image?')), sel), __('Insert'))) insert('[image: ' + sel.value + ']\n');
    }
    const fileInput = h('input', { type: 'file', accept: '.txt,.md,text/plain,text/markdown', 'aria-label': __('Open a file:'), onchange: (e) => readFile(e.target.files[0]) });

    // --- language tabs and translation helpers ---
    const tabList = h('div', { class: 'tablist', role: 'tablist', 'aria-label': __('Languages') });
    const tabs = h('div', { class: 'langtabs' }, tabList);
    async function switchTo(code) {
      if (ta.value !== A.saved && !await confirmBox(__('Discard your unsaved changes?'), __('The show file in the editor has changes that are not saved.'), __('Discard changes'), true)) return;
      A.lang = code === main.code ? '' : code; A.analysis = null; A.text = A.saved = ''; render();
    }
    function renderTabs() {
      tabList.replaceChildren(...A.langs.map((l) => h('button', { role: 'tab', class: (l.code === cur.code ? 'on ' : '') + (l.available ? '' : 'bad'), 'aria-selected': String(l.code === cur.code), onclick: () => switchTo(l.code),
        title: l.available ? '' : __('Not shown to the audience until its structure matches the main show file') },
        l.name, l.primary ? h('small', {}, __(' main')) : null, l.available ? null : ' ⚠')));
      if (!tabs.querySelector('.addlang')) tabs.append(h('button', { class: 'quiet addlang', onclick: addLanguage }, __('＋ Add a language')));
    }
    async function addLanguage() {
      const have = new Set(A.langs.map((l) => l.code));
      const sel = h('select', {}, ...known.filter((k) => !have.has(k.code)).map((k) => h('option', { value: k.code }, k.name + ' (' + k.code + ')')), h('option', { value: '' }, __('Another language…')));
      const code = h('input', { type: 'text', maxlength: '12', placeholder: __('language code, e.g. cy'), 'aria-label': __('Language code') });
      const name = h('input', { type: 'text', maxlength: '40', placeholder: __('name shown to the audience, e.g. Cymraeg'), 'aria-label': __('Language name') });
      const other = h('div', {}, h('label', { class: 'field' }, h('span', {}, __('Code')), code), h('label', { class: 'field' }, h('span', {}, __('Name the audience sees')), name)); other.hidden = true;
      sel.addEventListener('change', () => { other.hidden = sel.value !== ''; });
      const copy = h('input', { type: 'checkbox', checked: true, id: 'copymain' });
      const body = h('div', {}, h('p', {}, __('Each language is its own show file with exactly the same lines, commands and blank screens in the same order. Only the words change, so one press of Space shows the same moment in every language.')),
        h('label', { class: 'field' }, h('span', {}, __('Language')), sel), other,
        h('label', { class: 'tick' }, copy, h('span', {}, __('Start from a copy of the {name} show file, so I only have to change the words (recommended)', { name: main.name }))));
      if (!await ask(__('Add a language'), body, __('Add language'))) return;
      const c = (sel.value || code.value).trim().toLowerCase();
      const r = await api('POST', '/api/admin/shows/' + show.id + '/languages', { code: c, name: sel.value ? '' : name.value, copyFromMain: copy.checked });
      if (!r.ok) return fail(r);
      A.lang = r.data.code; A.analysis = null; A.text = A.saved = ''; toast(__('{name} added', { name: r.data.name })); render();
    }
    async function removeLanguage() {
      if (!await confirmBox(__('Remove {name}?', { name: cur.name }), __('The {name} show file will be deleted. The other languages are not affected.', { name: cur.name }), __('Remove'), true)) return;
      const r = await api('DELETE', '/api/admin/shows/' + show.id + '/languages/' + encodeURIComponent(cur.code));
      if (!r.ok) return fail(r);
      A.lang = ''; A.analysis = null; A.text = A.saved = ''; toast(__('{name} removed', { name: cur.name })); render();
    }
    renderTabs();
    let translating = null;
    if (A.lang) {
      const mainText = h('pre', { class: 'maintext' }, __('Loading…'));
      api('GET', '/api/admin/shows/' + show.id + '/showfile').then((m) => { mainText.textContent = m.ok ? m.data.text : ''; });
      translating = h('div', { class: 'card' },
        h('p', {}, h('b', {}, __('Translating into {name}.', { name: cur.name }) + ' '), __('Keep every line, command and blank screen in the same order as the {name} show file; change only the words. A problem is shown below if the shapes stop matching.', { name: main.name })),
        h('div', { class: 'row' }, h('details', {}, h('summary', {}, __('Show the {name} show file for reference', { name: main.name })), mainText), h('span', { class: 'grow' }), h('button', { class: 'danger quiet', onclick: removeLanguage }, __('Remove {name}…', { name: cur.name }))));
    }
    p.append(
      h('h2', {}, A.lang ? __('Show file — {name}', { name: cur.name }) : __('Show file')),
      tabs, ...(translating ? [translating] : []), // (append() would print a null as the word "null")
      h('p', { class: 'hint' }, __('This file holds everything about the show except its images: its title and look in a [settings] block, and the lines themselves. Each line is one caption; one press of the spacebar sends one line. You can paste, type, or drop a .txt or .md file. Click a problem to jump to its line.')),
      h('div', { class: 'row' }, h('label', { class: 'row' }, __('Open a file:'), fileInput), h('a', { class: 'btn', href: '/api/admin/shows/' + show.id + '/showfile/download' + q, download: show.id + (A.lang ? '.' + A.lang : '') + '.md', title: __('Edit it in any text editor, then bring it back with Open a file') }, __('Download the show file')), h('button', { onclick: () => openHelp('admin-writing-a-show-file') }, __('Show file help')), h('button', { onclick: showHistory, title: __('Each save keeps the version it replaced') }, __('Earlier versions…')), h('span', { class: 'grow' }), status, saveBtn),
      tools,
      h('div', { class: 'editor-grid' }, editor, h('div', {}, h('div', { class: 'card' }, diagTitle, diagList), h('div', { class: 'card' }, h('h3', {}, __('What the operator will step through')), cueList))),
      cheatSheet());
    lines(); setStatus(); showAnalysis();
    window.onbeforeunload = () => (ta.value !== A.saved ? __('You have unsaved changes.') : undefined);
  }

  function cheatSheet() {
    const rows = [
      ['A normal line', 'One caption. Long lines wrap on the screen.'], ['NAME: text', 'Speaker label (NAME in capitals)'], ['**bold**   *italic*', 'Bold and italic'],
      ['(thunder rumbles)', 'Shown as a sound or stage direction'], ['line one \\', 'A backslash at the end joins the next line into the same caption'], ['[[ … ]]', 'Several lines in one caption'],
      ['<!-- comment -->', 'A comment: never shown to anyone'], ['[note: text]', 'A reminder shown to the operator only'], ['[section: Act 2]', 'A bookmark for the operator'],
      ['[clear]  [blank]', 'Clear the captions / blank the screen'], ['[image: file.png | caption]', 'Fullscreen image (also  ![caption](file.png) )'], ['[fullscreen: text]', 'Fullscreen text'],
      ['~ lyric ~', 'A sung line, shown between ♪ notes'], ['[song: Title] … [/song]', 'A whole song: every line inside is a sung lyric'], ['[together] … [/together]', 'Voices shown at the same time (a duet or counterpoint), one press of Space. Also  A: ~ one ~ || B: ~ two ~  on a single line'], ['♪  ♫', 'Music notes can be typed or pasted anywhere; the ♪ and ♫ buttons insert them'],
      ['[settings] … [/settings]', 'The show’s title and look (font, colors, size, speaker names) and other options, as lines like  font: dyslexic. Usually at the top. The Look step writes it for you. A translation may only set  idle message'], ['[loop: 10s] … [/loop]', 'Repeating items for pre-show or intermission. Add  | 15s  to an item to change its time'], ['# …', 'Reserved for headings; the line is ignored'],
    ];
    return h('details', { class: 'card' }, h('summary', {}, __('Show file cheat sheet')), h('table', { class: 'cheat' }, ...rows.map(([a, b]) => h('tr', {}, h('td', {}, __(a)), h('td', {}, __(b))))));
  }

  // ---- step 3: images ----
  async function stepImages(p) {
    const show = curShow(); if (!show) return go('show');
    // Every language's show file can use images, so look at all of them together.
    const lr = await api('GET', '/api/admin/shows/' + show.id + '/languages');
    if (!lr.ok) return fail(lr);
    const langs = lr.data.languages;
    const rs = await Promise.all(langs.map((l) => api('GET', '/api/admin/shows/' + show.id + '/showfile' + (l.primary ? '' : '?lang=' + encodeURIComponent(l.code)))));
    if (rs.some((x) => !x.ok)) return fail(rs.find((x) => !x.ok));
    A.mainAnalysis = rs[0].data.analysis;
    const used = {}, miss = {};
    rs.forEach((x, i) => {
      const tag = langs.length > 1 ? ' (' + langs[i].code + ')' : '';
      const an = x.data.analysis;
      Object.keys(an.used).forEach((f) => { (used[f] = used[f] || []).push(...an.used[f].map((n) => n + tag)); });
      an.missing.forEach((m) => { (miss[m.name] = miss[m.name] || []).push(...m.lines.map((n) => n + tag)); });
    });
    const a = { assets: rs[0].data.analysis.assets, used, missing: Object.keys(miss).sort().map((name) => ({ name, lines: miss[name] })) };
    p.append(h('h2', {}, __('Images')), h('p', { class: 'hint' }, __('PNG, JPG, GIF or WebP, up to 10 MB each. Images are sent to every device as they are, so keep them under about 1 MB (about 1920 pixels wide is plenty).')));

    async function upload(files, renameTo) {
      const fd = new FormData();
      [...files].forEach((f) => fd.append('files', f, renameTo || f.name));
      const res = await api('POST', '/api/admin/shows/' + show.id + '/assets', undefined, fd);
      if (!res.ok) return fail(res, __('Upload failed'));
      (res.data.failed || []).forEach((f) => toast(f.name + ': ' + __(f.error)));
      if ((res.data.saved || []).length) toast(__('Uploaded {n} image(s)', { n: res.data.saved.length }));
      render();
    }
    if (a.missing.length) {
      p.append(h('div', { class: 'card missing' }, h('h3', {}, __('⚠ Images the show file uses but you have not uploaded')),
        ...a.missing.map((m) => h('div', { class: 'row' }, h('b', {}, m.name), h('span', { class: 'hint' }, __('used on line {lines}', { lines: m.lines.join(', ') })),
          h('label', { class: 'row' }, h('input', { type: 'file', accept: 'image/png,image/jpeg,image/gif,image/webp', 'aria-label': __('Images'), onchange: (e) => e.target.files[0] && upload([e.target.files[0]], m.name) }))))));
    }
    const dz = h('div', { class: 'dropzone' }, h('p', {}, __('Drag images here, or ')), h('input', { type: 'file', multiple: true, accept: 'image/png,image/jpeg,image/gif,image/webp', 'aria-label': __('Images'), onchange: (e) => upload(e.target.files) }));
    dz.addEventListener('dragover', (e) => { e.preventDefault(); dz.classList.add('drop'); });
    dz.addEventListener('dragleave', () => dz.classList.remove('drop'));
    dz.addEventListener('drop', (e) => { e.preventDefault(); dz.classList.remove('drop'); upload(e.dataTransfer.files); });
    p.append(dz);
    if (!a.assets.length) { p.append(h('p', { class: 'hint' }, __('No images yet. Shows work fine without them.'))); return; }
    p.append(h('div', { class: 'grid' }, ...a.assets.map((as) => {
      const lines = a.used[as.name];
      return h('div', { class: 'thumb' },
        h('img', { src: '/api/admin/shows/' + show.id + '/assets/' + encodeURIComponent(as.name), alt: as.name, loading: 'lazy' }),
        h('div', { class: 'meta' }, h('b', {}, as.name), fmtSize(as.size), as.large ? h('div', { class: 'status-warn' }, __('⚠ large: may load slowly on many devices')) : null,
          h('div', { class: lines ? '' : 'hint' }, lines ? __('Used on line {lines}', { lines: lines.join(', ') }) : __('Not used')),
          h('button', { class: 'danger quiet', onclick: async () => {
            if (!await confirmBox(__('Delete “{name}”?', { name: as.name }), lines ? __('The show file uses this image on line {lines}.', { lines: lines.join(', ') }) : __('It is not used in the show file.'), __('Delete'), true)) return;
            const d = await api('DELETE', '/api/admin/shows/' + show.id + '/assets/' + encodeURIComponent(as.name)); if (!d.ok) return fail(d); render();
          } }, __('Delete'))));
    })));
  }

  // ---- step 4: look ----
  const PRESETS = [['White on black (default)', '#000000', '#ffffff', '#ffffff'], ['Black on white', '#ffffff', '#000000', '#000000'], ['Yellow on black', '#000000', '#ffee00', '#ffee00'], ['Dark blue', '#001a33', '#ffffff', '#ffd166']];
  const FONTS = { hyperlegible: '"Atkinson Hyperlegible", system-ui, sans-serif', dyslexic: '"OpenDyslexic", system-ui, sans-serif', system: 'system-ui, sans-serif', serif: 'Georgia, "Times New Roman", serif', mono: 'ui-monospace, Menlo, Consolas, monospace' };
  function lum(hex) { const c = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255).map((v) => (v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4))); return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]; }
  function contrast(a, b) { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); }
  async function stepLook(p) {
    const show = curShow(); if (!show) return go('show');
    const r = await api('GET', '/api/admin/shows/' + show.id); if (!r.ok) return fail(r);
    const t = Object.assign({}, r.data.theme);
    const ctl = {};
    const mk = (name, label, input) => { ctl[name] = input; return h('label', { class: 'field' }, h('span', {}, __(label)), input); };
    mk('bg', 'Background', h('input', { type: 'color' })); mk('fg', 'Text', h('input', { type: 'color' })); mk('accent', 'Speaker name color', h('input', { type: 'color' }));
    mk('font', 'Font', h('select', {}, h('option', { value: 'system' }, __('Clean (system)')), h('option', { value: 'serif' }, __('Serif')), h('option', { value: 'mono' }, __('Monospace')), h('option', { value: 'dyslexic' }, __('Dyslexia-friendly (OpenDyslexic)')), h('option', { value: 'hyperlegible' }, __('High legibility (Atkinson Hyperlegible)'))));
    mk('scalePct', 'Text size', h('input', { type: 'range', min: '60', max: '200', step: '10' }));
    mk('speakers', 'Speaker names', h('select', {}, h('option', { value: 'label' }, __('Show above the line (in the speaker name color)')), h('option', { value: 'color' }, __('Show above the line, a different color for each character')), h('option', { value: 'hidden' }, __('Hide'))));
    mk('idleMessage', 'Message before the show starts', h('input', { type: 'text', maxlength: '200' }));
    mk('loopSeconds', 'Default time for each item in a loop (seconds)', h('input', { type: 'number', min: '1', max: '3600', step: 'any' }));
    mk('scrollbackLines', 'Captions a viewer can scroll back through', h('input', { type: 'number', min: '20', max: '2000', step: '10' }));
    // Other languages get their own pre-show message.
    const lgs = await api('GET', '/api/admin/shows/' + show.id + '/languages');
    const idleOthers = {};
    const idleLabels = (lgs.ok ? lgs.data.languages.filter((l) => !l.primary) : []).map((l) => {
      const inp = h('input', { type: 'text', maxlength: '200', dir: 'auto', 'aria-label': __('Message before the show starts ({name})', { name: l.name }) });
      inp.value = (t.idleMessages || {})[l.code] || '';
      idleOthers[l.code] = inp;
      return h('label', { class: 'field' }, h('span', {}, __('Message before the show starts ({name})', { name: l.name })), inp);
    });
    const ratio = h('div', { class: 'hint' });
    const sizeLabel = h('span', { class: 'hint' });
    const phone = h('div', { class: 'preview phone', role: 'img', 'aria-label': __('Phone preview') }), proj = h('div', { class: 'preview proj', role: 'img', 'aria-label': __('Projector preview') });
    const sample = () => [h('p', { class: 'dim2' }, __('The train was late, Margaret.')), h('p', { class: 'dim1' }, h('span', { class: 'spk' }, __('Margaret')), __('The train is never late.')), h('p', {}, h('span', { class: 'spk' }, __('Edwin')), __('Is that tea for '), h('b', {}, __('me?')))];

    function read() {
      for (const k of ['bg', 'fg', 'accent', 'font', 'speakers', 'idleMessage']) t[k] = ctl[k].value;
      t.scalePct = Number(ctl.scalePct.value);
      t.loopSeconds = Number(ctl.loopSeconds.value);
      t.scrollbackLines = Number(ctl.scrollbackLines.value);
      t.idleMessages = {};
      Object.keys(idleOthers).forEach((c) => { if (idleOthers[c].value.trim()) t.idleMessages[c] = idleOthers[c].value.trim(); });
    }
    function write() { for (const k of ['bg', 'fg', 'accent', 'font', 'speakers', 'idleMessage', 'scalePct', 'loopSeconds', 'scrollbackLines']) ctl[k].value = t[k]; }
    function paint() {
      const c = contrast(t.bg, t.fg);
      ratio.textContent = __('Contrast {ratio} : 1', { ratio: c.toFixed(1) }) + ' — ' + __(c >= 7 ? 'excellent' : c >= 4.5 ? 'acceptable, but higher is easier to read in a dark room' : 'too low to read comfortably');
      ratio.className = c >= 7 ? 'status-ok' : c >= 4.5 ? 'status-warn' : 'status-bad';
      sizeLabel.textContent = t.scalePct + '%';
      for (const f of [phone, proj]) {
        f.style.setProperty('--pbg', t.bg); f.style.setProperty('--pfg', t.fg); f.style.setProperty('--paccent', t.accent);
        f.style.setProperty('--pfont', FONTS[t.font] || FONTS.system); f.style.setProperty('--pscale', String(t.scalePct / 100));
        f.replaceChildren(...sample());
        if (t.speakers === 'hidden') f.querySelectorAll('.spk').forEach((e) => e.remove());
        if (t.speakers === 'color' && window.OCSpeakers) f.querySelectorAll('.spk').forEach((e) => { e.style.color = OCSpeakers.forName(e.textContent, ['Margaret', 'Edwin'], t.bg); });
      }
    }
    write(); paint();
    Object.values(ctl).forEach((i) => i.addEventListener('input', () => { read(); paint(); }));
    const presets = h('div', { class: 'presets' }, ...PRESETS.map(([n, bg, fg, ac]) => h('button', { onclick: () => { Object.assign(t, { bg, fg, accent: ac }); write(); paint(); } }, __(n))));
    p.append(h('h2', {}, __('Look')), h('p', { class: 'hint' }, __('Large, high-contrast text is the default and is best for accessibility. Saving writes these choices into the [settings] block of the show file (which you can also edit by hand in the Show file step), and applies them to connected devices straight away.')),
      h('div', { class: 'look-grid' },
        h('div', { class: 'card' }, presets, ...Object.keys(ctl).map((k) => ctl[k].closest('label')), ...idleLabels, sizeLabel, ratio,
          h('div', { class: 'row end' }, h('button', { class: 'primary', onclick: async () => {
            if (A.text !== A.saved && A.loadedLang === '') { toast(__('You have unsaved changes in the Show file step. Save them first: saving the look rewrites the settings in the saved show file.')); return; }
            read(); const s = await api('PUT', '/api/admin/shows/' + show.id, { theme: t }); if (!s.ok) return fail(s); A.analysis = null; A.text = A.saved = ''; markLookSaved(); toast(__('Look saved in the show file'));
          } }, __('Save look')))),
        h('div', { class: 'previews' }, h('div', {}, h('div', { class: 'hint' }, __('Phone')), phone), h('div', {}, h('div', { class: 'hint' }, __('Projector')), proj))));
  }

  // ---- step 5: test (rehearsal) ----
  async function stepTest(p) {
    const show = curShow(); if (!show) return go('show');
    p.append(h('h2', {}, __('Test with a rehearsal')),
      h('p', { class: 'hint' }, __('A rehearsal runs this show privately: you press the keys as the operator would and see the audience screen beside it. Nothing here reaches real audience devices, and it does not affect a show that is live.')));
    const area = h('div', {});
    p.append(h('div', { class: 'row' }, h('button', { class: 'primary', onclick: start }, A.rehearsing ? __('Restart rehearsal') : __('Start rehearsal'))), area);
    async function start() {
      const r = await api('POST', '/api/admin/rehearsal/start', { showId: show.id }); if (!r.ok) return fail(r);
      A.rehearsing = true; refreshNav();
      area.replaceChildren(h('div', { class: 'reh' },
        h('iframe', { src: '/operator?rehearsal=1', title: __('Operator (rehearsal)') }),
        h('div', { class: 'phoneframe' }, h('div', { class: 'hint' }, __('What the audience sees')), h('iframe', { src: '/?rehearsal=1', title: __('Audience (rehearsal)') }))),
        h('p', { class: 'hint' }, __('Click inside the operator panel, then use Space to send lines. Shift+Q stops a loop.')));
    }
    if (A.rehearsing) start();
  }

  // ---- step 6: go live ----
  async function stepGoLive(p) {
    const show = curShow(); if (!show) return go('show');
    const [net, sc, lg] = await Promise.all([api('GET', '/api/admin/network'), api('GET', '/api/admin/shows/' + show.id + '/showfile'), api('GET', '/api/admin/shows/' + show.id + '/languages')]);
    if (!net.ok || !sc.ok || !lg.ok) return fail(!net.ok ? net : !sc.ok ? sc : lg);
    const langs = lg.data.languages;
    A.analysis = sc.data.analysis; A.net = net.data;
    const a = A.analysis, errs = a.diagnostics.filter((d) => d.severity === 'error').length, warns = a.diagnostics.filter((d) => d.severity === 'warning').length;
    const checks = h('div', { class: 'card' }, h('h3', {}, __('Ready?')),
      h('div', { class: 'check' + (a.cues.length ? '' : ' bad') }, a.cues.length ? __('{n} lines in the show file', { n: a.cues.length }) : __('The show file is empty')),
      h('div', { class: 'check' + (errs ? ' bad' : warns ? ' warn' : '') }, errs ? __('{n} show file error(s) to fix', { n: errs }) : warns ? __('{n} show file warning(s) to review', { n: warns }) : __('No show file problems')),
      h('div', { class: 'check' + (a.missing.length ? ' warn' : '') }, a.missing.length ? __('Missing images: {list}', { list: a.missing.map((m) => m.name).join(', ') }) : __('All images are uploaded')),
      langs.length > 1 ? h('div', { class: 'check' }, __('Languages: {list}', { list: langs.filter((l) => l.available).map((l) => l.name).join(', ') })) : null,
      ...langs.filter((l) => !l.available).map((l) => h('div', { class: 'check bad' }, __('{name} will NOT be shown to the audience until it matches the main show file: {why}', { name: l.name, why: __((l.structure[0] || {}).message || '') }))));
    const urls = net.data.urls.filter((u) => u.public);
    let chosen = (urls.find((u) => u.kind === 'http' && !u.name) || urls.find((u) => u.kind === 'http') || urls[0] || net.data.urls[0] || {}).url; // the numeric address is the safest to print
    const big = h('div', { class: 'urlbig' }), qrBox = h('div', { class: 'qr' });
    const paint = () => { big.textContent = chosen || __('No address found'); qrBox.replaceChildren(chosen ? h('img', { src: '/api/admin/network/qr.png?size=512&url=' + encodeURIComponent(chosen), alt: __('QR code for {url}', { url: chosen }) }) : ''); };
    paint();
    const list = h('div', { class: 'urls' }, ...net.data.urls.map((u) => h('div', { class: 'row' }, h('label', { class: 'row' }, h('input', { type: 'radio', name: 'u', checked: u.url === chosen, onchange: () => { chosen = u.url; paint(); } }), h('a', { href: u.url, target: '_blank', rel: 'noopener', title: __('Opens in a new window') }, u.url), u.public ? '' : __(' (this computer only)'), u.name ? __(' (a name: easy to type, but some WiFi networks and phones cannot use it)') : ''))));
    const isActive = A.activeId === show.id;
    p.append(h('h2', {}, __('Go live')), checks,
      h('div', { class: 'card' }, h('h3', {}, __('Tell the audience this address')), big,
        h('div', { class: 'row' }, h('button', { onclick: async () => { try { await navigator.clipboard.writeText(chosen); toast(__('Copied')); } catch (e) { toast(__('Select the address and copy it')); } } }, __('Copy address')),
          h('a', { href: '/admin?print=qr&url=' + encodeURIComponent(chosen || ''), target: '_blank', rel: 'noopener' }, __('Print a QR sign for the lobby'))),
        h('div', { class: 'row' }, qrBox, list),
        ...(net.data.warnings || []).map((w) => h('div', { class: 'check warn' }, w)),
        h('p', { class: 'hint' }, __('Audience devices must be on the same WiFi as this computer. Devices connected right now: '), h('b', { id: 'audcount' }, String(A.audience)))),
      h('div', { class: 'card row' },
        h('button', { class: 'primary big', onclick: () => startShow(show), disabled: !a.cues.length }, isActive ? __('Restart the show and open the operator page') : __('Start show')),
        isActive ? h('a', { href: '/operator' }, __('Open the operator page without restarting')) : null,
        h('span', { class: 'hint' }, isActive ? __('This show is active.') : __('Starting makes this the show audiences see.'))),
      h('p', { class: 'hint' }, __('Something not working? '), h('a', { href: '#', onclick: (e) => { e.preventDefault(); openHelp('admin-troubleshooting'); } }, __('Troubleshooting help')), ' · ', h('a', { href: '/api/admin/support-bundle', download: 'support-bundle.zip' }, __('Download a support bundle')), ' · ', h('a', { href: '#', onclick: (e) => { e.preventDefault(); showLog(); } }, __('Recent log…'))));
  }
  async function startShow(show) {
    let r = await api('POST', '/api/admin/shows/' + show.id + '/activate', {});
    if (r.status === 409 && r.data.live) {
      if (!await confirmBox(__('A show is live'), __('Starting this show restarts it from the beginning and clears all audience screens. Continue?'), __('Restart the show'), true)) return;
      r = await api('POST', '/api/admin/shows/' + show.id + '/activate', { force: true });
    }
    if (!r.ok) return fail(r);
    A.activeId = show.id; refreshNav();
    location.href = '/operator';
  }

  // ---- printable QR sign ----
  function printSign(url) {
    document.body.className = 'print-qr';
    $('login').hidden = true; $('app').hidden = false;
    $('steps').hidden = true; document.querySelector('header').hidden = true;
    $('panel').replaceChildren(h('div', { class: 'sign' }, h('h1', {}, __('Open captions: scan to read along')),
      h('img', { src: '/api/admin/network/qr.png?size=1024&url=' + encodeURIComponent(url), alt: __('QR code') }), h('div', { class: 'u' }, url),
      h('p', {}, __('Join the venue’s guest WiFi first.')), h('button', { class: 'noprint primary', onclick: () => window.print() }, __('Print'))));
  }

  // The page's language: the strings in the page are translated once now; the menus switch language by reloading.
  OCI18n.apply();
  const askLeave = () => (A.text !== A.saved ? window.confirm(__('You have unsaved changes in the Show file step. Switch language and lose them?')) : true);
  $('settingsbtn').parentNode.insertBefore(OCI18n.picker(askLeave), $('settingsbtn'));
  $('login').insertBefore(OCI18n.picker(), $('login').firstChild);

  const q = new URLSearchParams(location.search);
  if (q.get('print') === 'qr' && q.get('url')) fetch('/api/admin/ping').then((r) => (r.ok ? printSign(q.get('url')) : showLogin()));
  else boot();
})();
