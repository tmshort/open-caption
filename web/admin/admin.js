// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Setup page: a six-step wizard (Show, Show file, Images, Look, Test, Share), then a Report of how the show went.
(() => {
  'use strict';
  if (window.OC_UNSUPPORTED) return; // compat.js has already shown the "too old" message
  const $ = (id) => document.getElementById(id);
  const A = { shows: [], activeId: '', cur: null, step: 'show', text: '', saved: '', analysis: null, mainAnalysis: null, lang: '', loadedLang: null, langs: [], net: null, live: false, audience: 0, rehearsing: false };
  const STEPS = [['show', 'Shows'], ['showfile', 'Show file'], ['images', 'Images'], ['look', 'Look'], ['test', 'Test'], ['golive', 'Share'], ['report', 'Report']];
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
      else if (k === 'ref') v(e); // (a page's CSP forbids style attributes, so styles are set through the DOM)
      else e.setAttribute(k, v === true ? '' : v);
    }
    for (const kid of kids.flat()) if (kid != null && kid !== false) e.append(kid.nodeType ? kid : document.createTextNode(kid));
    return e;
  }
  async function api(method, path, body, raw) {
    const opt = { method, headers: {} };
    if (raw) opt.body = raw; else if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
    let r;
    try { r = await fetch(path, opt); } catch (e) {
      // The program is not answering (restarting, or the WiFi dropped): say so rather than leave a blank panel.
      toast(__('Could not reach the program. Check the connection, then try again.'), 6000);
      throw new Error('network', { cause: e });
    }
    let data = {};
    try { data = await r.json(); } catch (e) { /* empty */ }
    if (r.status === 401) { showLogin(); throw new Error('login'); }
    return { ok: r.ok, status: r.status, data };
  }
  // Fetches a file and hands it to the browser to save. Unlike a plain download link, a failure is explained here
  // (the program says why) instead of leaving the browser's own "download failed".
  async function saveDownload(path, filename) {
    let r;
    try { r = await fetch(path); } catch (e) { toast(__('Could not reach the program. Check the connection, then try again.'), 6000); return; }
    if (r.status === 401) { showLogin(); return; }
    if (!r.ok) {
      let data = {};
      try { data = await r.json(); } catch (e) { /* no reason given */ }
      fail({ ok: false, status: r.status, data });
      return;
    }
    const url = URL.createObjectURL(await r.blob());
    const a = h('a', { href: url, download: filename });
    document.body.appendChild(a);
    a.click();
    a.remove();
    setTimeout(() => URL.revokeObjectURL(url), 10000);
  }
  // What to tell the person when the program gave no reason of its own: say what the status means and what to do.
  function whyFailed(r) {
    if (r.data && r.data.error) return r.data.error;
    const s = r.status;
    if (s === 403) return 'That is not allowed from this device or with this password.';
    if (s === 413) return 'That file is too large.';
    if (s === 429) return 'Too many tries. Wait a minute, then try again.';
    if (s >= 500) return 'The program had a problem. Try again; if it keeps happening, restart the program.';
    return 'That did not work. Check the connection and try again.';
  }
  let toastTimer;
  function toast(msg, ms) { const t = $('toast'); t.textContent = __(msg); t.classList.add('show'); clearTimeout(toastTimer); toastTimer = setTimeout(() => t.classList.remove('show'), ms || 3000); }
  function fail(r, fallback) { toast(__((r.data && r.data.error) || fallback || whyFailed(r))); }

  // A reusable dialog. body: node; resolves true on OK, false on cancel.
  function ask(title, body, okLabel, opts) {
    return new Promise((resolve) => {
      const d = $('dlg');
      d.classList.toggle('wide', !!(opts && opts.wide));
      d.classList.toggle('tabbed', !!(opts && opts.tabbed));
      if (opts && opts.kind) d.dataset.kind = opts.kind; else delete d.dataset.kind;
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
  const HELP_TOPIC = { show: 'admin-getting-started', showfile: 'admin-writing-a-show-file', images: 'admin-images', look: 'admin-look-and-feel', test: 'admin-getting-started', golive: 'admin-network', report: 'operator-running-the-show' };
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
  async function openSettings(openTab) {
    const r = await api('GET', '/api/admin/settings'); if (!r.ok) return fail(r);
    const s = r.data;
    let netChanged = false;
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
    // The standby screen: the venue's own logo and message, shown while no show is on air.
    const sb = s.standby || {};
    const sbMsg = h('input', { type: 'text', maxlength: '200', size: '40', 'aria-label': __('Standby message'), placeholder: __('Captions will begin shortly') }); sbMsg.value = sb.message || '';
    const sbImg = h('img', { class: 'sblogo', alt: __('Current standby logo') });
    const sbNone = h('span', { class: 'hint' }, __('Showing the built-in Open Caption logo until you choose your own.'));
    const sbRemove = h('button', { type: 'button', onclick: async () => { const x = await api('DELETE', '/api/admin/settings/standby-logo'); if (!x.ok) return fail(x); paintSb(x.data.standby); toast(__('Logo removed')); } }, __('Remove logo'));
    function paintSb(v) { sbImg.hidden = !v.logo; if (v.logo) sbImg.src = v.logo; sbNone.hidden = !v.default; sbRemove.disabled = !!v.default; }
    paintSb(sb);
    // The browser's own "Choose File / No file chosen" control is hidden; a button like the others opens the chooser.
    const [sbChoose, sbFile] = fileButton(__('Choose a logo…'), { accept: IMAGE_TYPES, label: __('Choose a logo'), title: __('Choose a PNG, JPEG, GIF or WebP picture from this computer, up to 2 MB') }, async (files) => {
      const fd = new FormData(); fd.append('file', files[0]);
      const x = await api('POST', '/api/admin/settings/standby-logo', undefined, fd);
      if (!x.ok) return fail(x); paintSb(x.data.standby); toast(__('Logo saved'));
    });
    const sbSave = h('button', { type: 'button', onclick: async () => { const x = await api('PUT', '/api/admin/settings/standby', { message: sbMsg.value }); if (!x.ok) return fail(x); toast(__('Message saved')); } }, __('Save message'));
    const logSelect = h('select', { 'aria-label': __('Log detail'), onchange: async (e) => {
      const x = await api('PUT', '/api/admin/settings/log-level', { level: e.target.value });
      if (!x.ok) { fail(x); e.target.value = s.logLevel; return; }
      s.logLevel = x.data.level; toast(x.data.note || __('Log detail: {level}', { level: x.data.level }), x.data.note ? 9000 : undefined);
    } }, ...[['error', 'Errors only'], ['warn', 'Problems (warnings and errors)'], ['info', 'Normal (events, sign-ins)'], ['debug', 'Debug (every request)']].map(([v, t]) => h('option', { value: v }, __(t))));
    logSelect.value = s.logLevel;
    const sec = (title, ...kids) => h('section', { class: 'setsec' }, h('h3', {}, __(title)), ...kids);
    // Which of this computer's network interfaces are used (network.* in config.yaml): a setting of this computer, not of a
    // show, so it lives here and not on the Share step. Each tick is saved as soon as it is clicked.
    const net0 = s.network || {};
    const saved = { restrictListening: !!net0.restrictListening };
    let ifaces = s.interfaces || [];
    const ifList = h('div', { class: 'iflist' });
    const netMsg = h('div', { class: 'check warn', role: 'alert', hidden: '' });
    const kindName = (k) => k === 'wifi' ? __('Wi-Fi') : k === 'ethernet' ? __('Ethernet') : '';
    const stateText = (i) => !i.selected ? __('not used') : i.used ? __('used') : __('not used: same network as {name} ({ip})', { name: i.instead, ip: i.insteadIp });
    const paintIfaces = () => ifList.replaceChildren(...(ifaces.length ? ifaces.map((i) => {
      const box = h('input', { type: 'checkbox', 'aria-label': __('Use {name}', { name: i.name }), disabled: i.current, onchange: () => putNet(box.checked ? ifaces.filter((x) => x.selected || x === i).map((x) => x.name) : ifaces.filter((x) => x.selected && x !== i).map((x) => x.name), saved.restrictListening) });
      box.checked = i.selected;
      return h('div', { class: 'ifrow' + (i.selected && i.used ? '' : ' off') },
        h('label', { class: 'row' }, box, h('b', {}, i.name), kindName(i.kind) ? ' (' + kindName(i.kind) + ')' : ''),
        h('div', { class: 'hint ifinfo' }, i.addresses.join(', '), ' — ', stateText(i), i.current ? '. ' + __('you are using Setup through this interface') : ''));
    }) : [h('p', { class: 'hint' }, __('No network address was found.'))]));
    // Saves the choice and shows what the program does with it; returns whether it worked.
    async function putNet(names, restrict) {
      netMsg.hidden = true;
      const x = await api('PUT', '/api/admin/settings/network', { interfaces: names, restrictListening: restrict });
      if (!x.ok) { netMsg.textContent = __((x.data && x.data.error) || 'Could not change the network settings'); netMsg.hidden = false; paintIfaces(); return false; }
      saved.restrictListening = restrict; ifaces = x.data.interfaces || []; paintIfaces(); netChanged = true; toast(__('Network settings saved'));
      return true;
    }
    paintIfaces();
    const restrict = h('input', { type: 'checkbox', onchange: () => putNet(ifaces.filter((x) => x.selected).map((x) => x.name), restrict.checked).then((ok) => { if (!ok) restrict.checked = saved.restrictListening; }) });
    restrict.checked = saved.restrictListening;
    const networkSec = sec('Network interfaces',
      h('div', { class: 'addrbox' },
        h('p', { class: 'hint' }, __('The audience is given the addresses of the interfaces that are ticked: in the address list, the QR code, the HTTPS certificate and the .local name. Tick only the networks the audience uses. When two ticked interfaces are on the same network (a cable and Wi-Fi), only one address is offered: the wired one, else the main connection. Untick the other to choose.')),
        ifList,
        h('label', { class: 'row addrrestrict' }, restrict, __('Also refuse connections to interfaces that are not ticked')),
        h('p', { class: 'hint' }, __('Connections to this computer itself are always accepted. If a phone then cannot reach the program, clear this box.')),
        s.networkPinned ? h('p', { class: 'hint' }, __('This copy was started with --network-interfaces, which overrides these ticks each time it starts.')) : null,
        netMsg));
    // The name this computer is announced by on the local network (<name>.local, multicast DNS): also a network setting.
    const m = s.mdns || {};
    const nameIn = h('input', { type: 'text', class: 'netname', maxlength: '63', size: '16', 'aria-label': __('Name'), autocapitalize: 'none', spellcheck: 'false' }); nameIn.value = m.requested || 'open-caption';
    const useName = h('input', { type: 'checkbox', 'aria-label': __('Announce this computer by name'), onchange: () => putName(useName.checked) });
    useName.checked = !!m.enabled;
    const nameInfo = h('div', { class: 'namestatus' });
    const paintName = (st) => {
      const kids = [];
      if (st.active) kids.push(h('p', { class: 'hint' }, __('Devices on this network can open http://{host}/', { host: st.host })));
      if (st.error) kids.push(h('div', { class: 'check warn', role: 'alert' }, __('Could not announce this name: {error}', { error: st.error })));
      else if (st.renamed) kids.push(h('div', { class: 'check warn', role: 'alert' }, __('Another computer on this network already uses {name}.local, so this one is http://{host}/', { name: st.requested, host: st.host })));
      nameInfo.replaceChildren(...kids);
    };
    paintName(m);
    async function putName(enabled) { // the checkbox and the Apply button both save what is shown
      const x = await api('PUT', '/api/admin/settings/mdns', { enabled, name: nameIn.value });
      if (!x.ok) { nameInfo.replaceChildren(h('div', { class: 'check warn', role: 'alert' }, __((x.data && x.data.error) || 'Could not change the name'))); useName.checked = !!m.enabled; return; }
      Object.assign(m, x.data.mdns); useName.checked = !!m.enabled; nameIn.value = m.requested || nameIn.value; paintName(m); netChanged = true; toast(__('Network settings saved'));
    }
    const nameSec = sec('Name on the network',
      h('p', { class: 'hint' }, __('Devices can open this computer by name instead of a number. It needs no router settings, but some guest WiFi networks block it, so keep the numeric address and QR code as the main way in.')),
      h('label', { class: 'row' }, useName, __('Use a name')),
      h('div', { class: 'row' }, 'http://', nameIn, __('.local '), h('button', { type: 'button', onclick: () => putName(useName.checked) }, __('Apply'))),
      nameInfo);
    // The HTTPS certificate: where it comes from and, for an automatic one, its state (docs/10). Filled in once the program answers.
    const certBox = h('div', { class: 'certbox' });
    const paintCert = (c) => {
      const when = (t) => (t && !String(t).startsWith('0001') ? new Date(t).toLocaleString(OCI18n.lang) : '');
      const states = { waiting: 'Waiting for the first certificate', ok: 'Valid', renewing: 'Getting a certificate…', failed: 'The last attempt failed' };
      if (c.source === 'none') return certBox.replaceChildren(h('p', { class: 'hint' }, __('HTTPS is off, so devices use plain http. Turn it on in the settings file (https: enabled: yes).')));
      if (c.source === 'self-signed') return certBox.replaceChildren(h('p', { class: 'hint' }, __('HTTPS uses a certificate this program made itself, so browsers show a warning until it is accepted. For a trusted one, set up automatic certificates (docs/https-certificate-guide.md).')));
      if (c.source === 'file') return certBox.replaceChildren(h('p', { class: 'hint' }, __('HTTPS uses the certificate files named in the settings file. When they are replaced (for example after a renewal) the new ones are used within about 30 seconds, with no restart.')));
      const st = c.status || {};
      const renew = h('button', { type: 'button', disabled: st.state === 'renewing', onclick: async () => {
        const x = await api('POST', '/api/admin/settings/certificate/renew'); if (!x.ok) return fail(x);
        toast(__('Getting a certificate; this can take a few minutes'), 6000); poll(100);
      } }, __('Renew now'));
      certBox.replaceChildren(
        h('p', { class: 'hint' }, __('Automatic certificate from {directory}, using the DNS provider {provider}.', { directory: c.directory, provider: c.provider })),
        h('table', { class: 'cheat' },
          row('Name', st.domain), row('State', states[st.state] || ''), row('Issued by', st.issuer),
          row('Valid until', st.notAfter && !String(st.notAfter).startsWith('0001') ? when(st.notAfter) + ' (' + __('{n} days left', { n: st.daysLeft }) + ')' : ''),
          row('Last attempt', when(st.lastAttempt)), st.lastError ? row('Last problem', st.lastError) : null),
        h('div', { class: 'row' }, renew));
    };
    const loadCert = async () => { const x = await api('GET', '/api/admin/settings/certificate'); if (x.ok) { paintCert(x.data); return x.data; } };
    const poll = (tries) => { if (tries <= 0) return; setTimeout(async () => { if (!certBox.isConnected) return; const c = await loadCert(); if (c && c.status && c.status.state === 'renewing' || tries === 100) poll(tries - 1); }, 3000); };
    loadCert();
    // Five tabs, so the window is short and the controls people come for are one click away; the title, the tabs and Close stay in
    // view while a tab scrolls. Every tab stays in the page (only hidden), so something typed but not yet saved survives a switch.
    const tabDefs = [
      ['aud', 'Audience screen', [
        sec('Standby screen',
          h('p', { class: 'hint' }, __('What the audience sees while no show is on air: before the operator presses START, after Go to standby, and when no show is loaded. It belongs to the venue, not to a show.')),
          h('div', { class: 'row' }, sbImg, sbNone), h('div', { class: 'row' }, sbChoose, sbFile, sbRemove, h('span', { class: 'hint' }, __('PNG, JPEG, GIF or WebP, up to 2 MB'))),
          h('label', { class: 'field' }, h('span', {}, __('Message (optional; empty uses “Captions will begin shortly” in each viewer’s language)')), sbMsg), h('div', { class: 'row' }, sbSave)),
        ]],
      ['net', 'Network', [networkSec, nameSec]],
      ['sec', 'Security', [
          !s.adminRestricted && !s.httpsAddr ? h('p', { class: 'check warn' }, __('Anyone on this WiFi can try to sign in to Setup, and passwords travel unencrypted over plain http. To limit it, set access: admin_allow in the settings file (for example [localhost]); see the usage guide.')) : null,
        sec('Change passwords',
          s.canChangePasswords
            ? h('div', {}, h('p', { class: 'hint' }, __('Leave a box empty to keep that password. At least 6 characters. People signed in with a changed password are signed out; you stay signed in.')),
                h('label', { class: 'field' }, h('span', {}, __('New admin password')), adminPw),
                h('label', { class: 'field' }, h('span', {}, __('New operator password')), opPw),
                h('div', { class: 'row' }, changePw, reveal), shown)
            : h('p', { class: 'hint' }, __('This copy was started without a settings file, so passwords cannot be changed here.'))),
      ]],
      ['cert', 'Certificate', [sec('Certificate', certBox)]],
      ['abt', 'About', [
        sec('About this copy',
          h('table', { class: 'cheat' },
            row('Program', 'open-caption ' + s.version), row('Settings file', s.configPath), row('Shows and data', s.dataDir),
            row('Setup page can be opened from', s.adminRestricted ? s.adminAllow.join(', ') : 'any computer on the network'),
            row('Operator page can be opened from', s.operatorRestricted ? s.operatorAllow.join(', ') : 'any computer on the network'),
            row('Web port (HTTP)', (s.httpAddr || '').replace(/^.*:/, '')), row('Web port (HTTPS)', (s.httpsAddr || '').replace(/^.*:/, ''))),
          h('p', { class: 'hint' }, __('These are fixed in the settings file. To change them, edit that file and start the program again.'))),
        sec('Log detail',
          logSelect,
          s.logLevelPinned ? h('p', { class: 'check warn' }, __('This copy was started with --log-level {level}, which overrides the settings file every time it starts. A change made here applies now but lasts only until the program restarts.', { level: s.logLevelPinned })) : null,
          h('p', { class: 'hint' }, s.canChangePasswords ? __('How much the program records (in the support bundle and in the log file, if one is kept). Info records sign-ins and events; Debug adds every request, for tracking down a problem. Applies at once and is kept in the settings file.') : __('How much the program records (in the support bundle and in the log file, if one is kept). Info records sign-ins and events; Debug adds every request, for tracking down a problem. Applies at once.'))),
      ]]];
    const tabs = h('div', { class: 'tabs', role: 'tablist', 'aria-label': __('Settings sections') });
    const panels = tabDefs.map(([id, label, kids]) => h('div', { class: 'tabpanel inactive', role: 'tabpanel', id: 'setpanel-' + id, 'aria-labelledby': 'settab-' + id }, ...kids));
    const tabBtns = tabDefs.map(([id, label], i) => h('button', { type: 'button', role: 'tab', id: 'settab-' + id, 'aria-controls': 'setpanel-' + id, onclick: () => pick(i, false),
      onkeydown: (e) => { const n = tabDefs.length, k = e.key; const to = k === 'ArrowRight' ? (i + 1) % n : k === 'ArrowLeft' ? (i + n - 1) % n : k === 'Home' ? 0 : k === 'End' ? n - 1 : -1; if (to >= 0) { e.preventDefault(); pick(to, true); } } }, __(label)));
    function pick(i, focus) {
      tabBtns.forEach((b, k) => { b.setAttribute('aria-selected', k === i ? 'true' : 'false'); b.tabIndex = k === i ? 0 : -1; panels[k].classList.toggle('inactive', k !== i); });
      if (focus) tabBtns[i].focus();
      $('dlgbody').scrollTop = 0;
      try { localStorage.setItem('oc.settings.tab', tabDefs[i][0]); } catch (e) { /* storage unavailable */ }
    }
    tabs.append(...tabBtns);
    // All the panels share one grid cell (the unselected ones are only invisible), so the window is as tall as the tallest tab and
    // never changes size when you switch: the About tab fits without scrolling wherever the screen allows.
    const body = h('div', { class: 'settingsbody' }, tabs, h('div', { class: 'tabpanels' }, ...panels));
    let first = 0;
    try { const last = localStorage.getItem('oc.settings.tab'); const k = tabDefs.findIndex((t) => t[0] === last); if (k >= 0) first = k; } catch (e) { /* storage unavailable */ }
    if (openTab) { const k = tabDefs.findIndex((t) => t[0] === openTab); if (k >= 0) first = k; }
    pick(first, false);
    // Each section above has its own button (or applies as soon as you choose); this one only closes the window.
    await ask(__('Settings'), body, __('Close'), { noCancel: true, wide: true, tabbed: true });
    if (netChanged && A.step === 'golive') render(); // the Share step lists the addresses that are offered
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
      pre, count, h('p', { class: 'hint' }, __('Passwords are never written to the log. Setup → Share has a support bundle with this log and more, for sending to someone who is helping.')));
    await load(); pre.scrollTop = pre.scrollHeight;
    const timer = setInterval(() => { if (auto.checked) load(); }, 3000);
    await ask(__('Recent log'), body, __('Close'), { noCancel: true, wide: true });
    clearInterval(timer);
  }
  $('settingsbtn').addEventListener('click', () => openSettings());

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
    try { if ((await ping.json()).auto) $('logout').hidden = true; } catch (e) { /* no sign-out to hide */ } // opened on this computer: no password, so nothing to sign out of
    $('login').hidden = true; $('app').hidden = false;
    await loadShows();
    try { A.cur = localStorage.getItem('oc.cur'); } catch (e) { /* storage unavailable */ }
    if (!curShow()) A.cur = A.activeId && curShowById(A.activeId) ? A.activeId : (unarchived()[0] ? unarchived()[0].id : null);
    A.step = A.cur || location.hash.slice(1) === 'report' ? (location.hash.slice(1) || 'show') : 'show';
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
    if (A.parked) return; // leaving the page: never hold a connection (see the pagehide handler below)
    if (A.es) A.es.close();
    const es = new EventSource('/api/operator/stream');
    A.es = es;
    es.addEventListener('opstate', (e) => {
      $('dot').classList.add('on'); clearTimeout(A.lostTimer); A.lostTimer = 0; $('connbar').hidden = true; // the server is answering
      const p = JSON.parse(e.data);
      const wasLive = A.live, wasStart = A.atStart;
      A.atStart = p.op.atStart; A.live = p.op.live; A.audience = p.audience; A.activeId = p.show.id;
      $('viewers').textContent = OCClients.label(p); $('viewers').title = OCClients.title(p.audience);
      $('live').hidden = !p.op.live; $('live').textContent = '● ' + __('ON AIR');
      if (A.step === 'golive') { const v = document.getElementById('audcount'); if (v) v.textContent = String(p.audience); }
      if ((wasLive !== A.live || wasStart !== A.atStart) && (A.step === 'show' || A.step === 'golive')) loadShows().then(() => (A.step === 'show' || A.step === 'golive') && render());
    });
    es.onerror = () => { $('dot').classList.remove('on'); es.close(); if (!A.lostTimer) A.lostTimer = setTimeout(() => { if (!A.parked) $('connbar').hidden = false; }, 2000); if (A.es === es) setTimeout(openOpStream, 2000); };
  }

  // A browser keeps only six connections to one address and every open page holds a stream; a page left through a link
  // can stay alive in the back/forward cache with its stream open, so going back and forth between Setup and the
  // operator page used up all six and the next page hung. Give the stream up as soon as the page is left.
  window.addEventListener('pagehide', () => { A.parked = true; if (A.es) { A.es.close(); A.es = null; } $('dot').classList.remove('on'); });
  window.addEventListener('pageshow', (e) => { if (e.persisted) { A.parked = false; openOpStream(); loadShows().then(() => render()); } });

  // The audience count opens the list of connected devices (device, address, language), kept up to date while it is open.
  $('viewers').addEventListener('click', async () => {
    const box = h('div', {});
    const fill = async () => { const r = await api('GET', '/api/operator/clients'); if (r.ok) box.replaceChildren(OCClients.table(r.data.clients)); };
    await fill();
    const timer = setInterval(fill, 3000);
    await ask(__('Connected devices'), box, __('Close'), { noCancel: true, kind: 'clients' });
    clearInterval(timer);
  });

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
      nav.append(h('button', { class: (A.step === id ? 'on ' : '') + (done(id) ? 'done' : ''), onclick: () => go(id), disabled: id !== 'show' && id !== 'report' && !A.cur, // the report of the last show does not depend on the show being worked on
        'aria-current': A.step === id ? 'step' : false },
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
    clearTimeout(A.reportTimer); // the Report step refreshes itself while the show is on the air; any other step does not
    const s = curShow();
    $('working').textContent = s ? s.name : __('No show selected');
    $('working').hidden = false;
    refreshNav();
    warmAnalysis();
    const panel = $('panel'); panel.replaceChildren();
    panel.classList.toggle('wide', A.step === 'showfile'); // the editor uses the whole width of the window
    document.title = __('Setup') + ' — ' + __((STEPS.find((x) => x[0] === A.step) || ['', ''])[1]);
    nextStep();
    try { await ({ show: stepShow, showfile: stepShowFile, images: stepImages, look: stepLook, test: stepTest, golive: stepGoLive, report: stepReport })[A.step](panel); } catch (e) { if (e.message !== 'login' && e.message !== 'network') console.error(e); }
  }
  // A button whose request failed for lack of a connection has already told the person (see api()).
  window.addEventListener('unhandledrejection', (ev) => { if (ev.reason && ev.reason.message === 'network') ev.preventDefault(); });

  // Under every step but the last: a button for the next one, so a newcomer is always shown the way on.
  function nextStep() {
    const box = $('nextstep'), i = STEPS.findIndex((x) => x[0] === A.step), nx = STEPS[i + 1];
    box.replaceChildren();
    box.hidden = !nx || !A.cur;
    if (box.hidden) return;
    box.append(h('button', { class: done(A.step) ? 'primary' : '', onclick: () => { go(nx[0]); window.scrollTo(0, 0); } }, __('Next: {step} →', { step: __(nx[1]) })));
  }

  // ---- step 1: show ----
  async function stepShow(p) {
    await loadShows();
    p.append(h('h2', {}, A.shows.length ? __('Choose a show') : __('Add a show')));
    const current = unarchived(), archived = A.shows.filter((s) => s.archived);
    if (!A.shows.length) {
      p.append(h('div', { class: 'card empty' },
        h('p', { class: 'big' }, __('Welcome! Start by creating a show, importing one, or look around with the sample.')),
        h('div', { class: 'row addrow' }, ...newShowForm(), h('button', { onclick: makeSample }, __('Create a sample show')), ...importButton())));
    } else {
      if (!current.length) p.append(h('div', { class: 'card empty' }, h('p', { class: 'big' }, __('All your shows are archived. Create a new show below, or restore one from the archive.'))));
      current.forEach((s) => p.append(h('div', { class: 'card show' },
        h('div', { class: 'name' }, s.name, ' ', s.active ? h('span', { class: 'badge active' }, __('ACTIVE')) : null, ' ', s.live ? h('span', { class: 'badge live' }, __('ON AIR')) : null,
          h('div', { class: 'hint' }, __('Updated {when}', { when: new Date(s.updated).toLocaleString(OCI18n.lang) }))),
        h('button', { class: A.cur === s.id ? '' : 'primary', onclick: () => { setCur(s.id); go('showfile'); } }, A.cur === s.id ? __('Continue with this show') : __('Work on this show')),
        moreMenu(s.name, [
          { label: __('Rename…'), run: () => renameShow(s) },
          { label: __('Duplicate…'), run: () => duplicateShow(s) },
          { label: __('Export'), run: () => saveDownload('/api/admin/shows/' + s.id + '/export', s.id + '.zip') },
          { label: __('Archive'), run: () => archiveShow(s, true), why: s.live ? __('The show on air cannot be archived. Activate another show first.') : '', disabled: s.live, hint: __('Put this show away; you can restore it later') },
          { sep: true },
          { label: __('Delete…'), run: () => deleteShow(s), danger: true, why: s.live ? __('This show is on air. Restart it from the Run page, or activate another show first.') : '' }])))); 
      p.append(h('h3', {}, __('Add another show')), h('div', { class: 'card row addrow' }, ...newShowForm(), h('button', { onclick: makeSample }, __('Create a sample show')), ...importButton()));
    }
    if (archived.length) {
      p.append(h('details', { class: 'archive', open: A.showArchive ? '' : null, ontoggle: (e) => { A.showArchive = e.target.open; } },
        h('summary', {}, __('Archived shows ({n})', { n: archived.length })),
        h('p', { class: 'hint' }, __('Archived shows are kept in full but stay out of the way and cannot go on air. Restore one to use it again, or delete it for good.')),
        ...archived.map((s) => h('div', { class: 'card show' },
          h('div', { class: 'name' }, s.name, h('div', { class: 'hint' }, __('Updated {when}', { when: new Date(s.updated).toLocaleString(OCI18n.lang) }))),
          h('button', { class: 'primary', onclick: () => archiveShow(s, false) }, __('Restore')),
          moreMenu(s.name, [
            { label: __('Export'), run: () => saveDownload('/api/admin/shows/' + s.id + '/export', s.id + '.zip') },
            { sep: true },
            { label: __('Delete…'), run: () => deleteShow(s), danger: true, why: s.live ? __('This show is on air. Restart it from the Run page, or activate another show first.') : '' }])))));
    }
  }
  // A "More ▾" button with a menu of the less common actions of one thing (a show). items: { label, run | href+download, danger,
  // disabled, why (shown under a disabled item), hint } or { sep: true }. Keyboard: Enter/Space/Down opens, arrows move, Home/End,
  // Esc closes and returns to the button, Tab leaves. A disabled item stays in the menu (so it can be found) and says why.
  let openMenu = null;
  function closeMenu() { if (openMenu) { openMenu.menu.hidden = true; openMenu.btn.setAttribute('aria-expanded', 'false'); openMenu = null; } }
  document.addEventListener('click', (e) => { if (openMenu && !openMenu.box.contains(e.target)) closeMenu(); });
  function moreMenu(forName, items) {
    const box = h('div', { class: 'moremenu' });
    const btn = h('button', { type: 'button', class: 'more', 'aria-haspopup': 'menu', 'aria-expanded': 'false', 'aria-label': __('More actions for {name}', { name: forName }) }, __('More'), ' ▾');
    const menu = h('div', { class: 'menu', role: 'menu', hidden: '' });
    const live = [];
    items.forEach((it) => {
      if (it.sep) { menu.append(h('div', { class: 'sep', role: 'separator' })); return; }
      const props = { role: 'menuitem', tabindex: '-1', class: 'item' + (it.danger ? ' danger' : '') };
      if (it.href) { props.href = it.href; props.download = it.download; }
      const el = h(it.href ? 'a' : 'button', props, h('span', {}, it.label));
      if (it.disabled) el.setAttribute('aria-disabled', 'true');
      if (it.why) { const id = 'mw' + Math.random().toString(36).slice(2, 8); el.append(h('small', { id }, it.why)); el.setAttribute('aria-describedby', id); }
      else if (it.hint) el.title = it.hint;
      el.addEventListener('click', (e) => {
        if (it.disabled) { e.preventDefault(); return; }
        closeMenu(); btn.focus();
        if (it.run) { e.preventDefault(); it.run(); }
      });
      menu.append(el); live.push(el);
    });
    const show = (focusIdx) => { closeMenu(); menu.hidden = false; btn.setAttribute('aria-expanded', 'true'); openMenu = { box, menu, btn }; if (live[focusIdx]) live[focusIdx].focus(); };
    btn.addEventListener('click', () => { if (openMenu && openMenu.menu === menu) { closeMenu(); } else show(0); });
    btn.addEventListener('keydown', (e) => { if (e.key === 'ArrowDown') { e.preventDefault(); show(0); } else if (e.key === 'ArrowUp') { e.preventDefault(); show(live.length - 1); } });
    menu.addEventListener('keydown', (e) => {
      const i = live.indexOf(document.activeElement);
      if (e.key === 'ArrowDown') { e.preventDefault(); live[(i + 1) % live.length].focus(); }
      else if (e.key === 'ArrowUp') { e.preventDefault(); live[(i + live.length - 1) % live.length].focus(); }
      else if (e.key === 'Home') { e.preventDefault(); live[0].focus(); }
      else if (e.key === 'End') { e.preventDefault(); live[live.length - 1].focus(); }
      else if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); closeMenu(); btn.focus(); }
      else if (e.key === 'Tab') closeMenu();
    });
    box.append(btn, menu);
    return box;
  }
  // Every file chooser in Setup is a button like any other: it opens the browser's chooser through a hidden
  // input, and hands the chosen files to onFiles. Returns [button, input] to put in a row.
  const IMAGE_TYPES = 'image/png,image/jpeg,image/gif,image/webp';
  function fileButton(text, opts, onFiles) {
    const inp = h('input', { type: 'file', accept: opts.accept, multiple: opts.multiple ? true : null, 'aria-label': opts.label || text, hidden: '', tabindex: '-1',
      onchange: (e) => { const files = [...e.target.files]; e.target.value = ''; if (files.length) onFiles(files); } });
    return [h('button', { title: opts.title, onclick: () => inp.click() }, text), inp];
  }
  function importButton() {
    return fileButton(__('Import a show…'), { accept: '.zip,.md,.txt,text/plain,text/markdown', title: __('Choose an exported show (.zip) or a plain .md / .txt show file from this computer') }, (files) => importShow(files[0]));
  }
  function newShowForm() {
    const inp = h('input', { type: 'text', class: 'newname', placeholder: __('Name of the show'), maxlength: '100', 'aria-label': __('Show name') });
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
  async function importShow(f) {
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
  // A refusal is shown in a window that has to be dismissed: a toast is easy to miss, and "nothing happened" looks like a bug.
  function refuseDelete(why) { return ask(__("Can't delete this show"), h('p', {}, why), __('OK'), { noCancel: true }); }
  async function deleteShow(s) {
    if (s.live) return refuseDelete(__('This show is on air. Restart it from the Run page, or activate another show first.'));
    const ok = await confirmBox(__('Delete “{name}”?', { name: s.name }), s.archived ? __('This archived show, its show file and its images will be permanently deleted. Export it first if you may want it back.') : __('Its show file and images will be permanently deleted. Export it first if you may want it back. (Archive keeps a show without deleting it.)'), __('Delete'), true);
    if (!ok) return;
    const r = await api('DELETE', '/api/admin/shows/' + s.id);
    if (!r.ok) return refuseDelete(__(whyFailed(r)));
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
    const saveBtn = h('button', { class: 'primary', title: __('Save (Ctrl+S, or Cmd+S on a Mac)'), onclick: save }, __('Save'));

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
        const ok = await ask(__('This show is on air'), h('div', {}, h('p', {}, __('Changing the show file now affects what the operator sees. The current position is kept by line number. To continue, type the show’s name:')), h('p', {}, h('b', {}, r.data.showName)), inp), __('Save change'));
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
    const [uploadBtn, fileInput] = fileButton(__('Upload a show file…'), { accept: '.txt,.md,text/plain,text/markdown', title: __('Choose a .txt or .md file from this computer to put in the editor (nothing is saved until you press Save)') }, (files) => readFile(files[0]));

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
      h('div', { class: 'row filebar' }, uploadBtn, fileInput, h('a', { class: 'btn', href: '/api/admin/shows/' + show.id + '/showfile/download' + q, download: show.id + (A.lang ? '.' + A.lang : '') + '.md', title: __('Edit it in any text editor, then bring it back with Upload a show file') }, __('Download the show file')), h('button', { onclick: showHistory, title: __('Each save keeps the version it replaced') }, __('Earlier versions…')), h('span', { class: 'grow' }), h('span', { class: 'savebar' }, status, saveBtn)),
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
          ...fileButton(__('Upload this image…'), { accept: IMAGE_TYPES, title: __('Choose the picture for {name} from this computer', { name: m.name }), label: __('Upload {name}', { name: m.name }) }, (files) => upload([files[0]], m.name))))));
    }
    const dz = h('div', { class: 'dropzone' }, h('p', {}, __('Drag images here, or ')), ...fileButton(__('Upload images…'), { accept: IMAGE_TYPES, multiple: true, title: __('Choose one or more pictures from this computer') }, (files) => upload(files)));
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
    mk('idleMessage', 'Pre-show message', h('input', { type: 'text', maxlength: '200' }));
    mk('loopSeconds', 'Default time for each item in a loop (seconds)', h('input', { type: 'number', min: '1', max: '3600', step: 'any' }));
    mk('scrollbackLines', 'Captions a viewer can scroll back through', h('input', { type: 'number', min: '20', max: '2000', step: '10' }));
    // Other languages get their own pre-show message.
    const lgs = await api('GET', '/api/admin/shows/' + show.id + '/languages');
    const idleOthers = {};
    const idleLabels = (lgs.ok ? lgs.data.languages.filter((l) => !l.primary) : []).map((l) => {
      const inp = h('input', { type: 'text', maxlength: '200', dir: 'auto', 'aria-label': __('Pre-show message ({name})', { name: l.name }) });
      inp.value = (t.idleMessages || {})[l.code] || '';
      idleOthers[l.code] = inp;
      return h('label', { class: 'field' }, h('span', {}, __('Pre-show message ({name})', { name: l.name })), inp);
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
    // Two columns of controls: the colors with their contrast just under them, then the text options, then the
    // pre-show message in every language together (full width), then the loop and scrollback numbers.
    const L = (k) => ctl[k].closest('label'), wide = (e) => { e.classList.add('wide'); return e; };
    L('scalePct').append(sizeLabel);
    const fieldOrder = [L('bg'), L('fg'), wide(h('div', { class: 'contrastrow' }, ratio)), L('accent'), L('font'), wide(L('scalePct')), wide(L('speakers')), wide(L('idleMessage')), ...idleLabels.map(wide), L('loopSeconds'), L('scrollbackLines')];
    write(); paint();
    Object.values(ctl).forEach((i) => i.addEventListener('input', () => { read(); paint(); }));
    const presets = h('div', { class: 'presets' }, ...PRESETS.map(([n, bg, fg, ac]) => h('button', { class: 'swatch', ref: (b) => { b.style.background = bg; b.style.color = fg; b.style.borderColor = ac; }, title: __('Use these colors'), onclick: () => { Object.assign(t, { bg, fg, accent: ac }); write(); paint(); } }, __(n))));
    p.append(h('h2', {}, __('Look')), h('p', { class: 'hint' }, __('Large, high-contrast text is the default and is best for accessibility. Saving writes these choices into the [settings] block of the show file (which you can also edit by hand in the Show file step), and applies them to connected devices straight away.')),
      h('div', { class: 'look-grid' },
        h('div', { class: 'card' }, presets, h('div', { class: 'fields' }, ...fieldOrder),
          h('div', { class: 'row end savelook' }, h('button', { class: 'primary', onclick: async () => {
            if (A.text !== A.saved && A.loadedLang === '') { toast(__('You have unsaved changes in the Show file step. Save them first: saving the look rewrites the settings in the saved show file.')); return; }
            read(); const s = await api('PUT', '/api/admin/shows/' + show.id, { theme: t }); if (!s.ok) return fail(s); A.analysis = null; A.text = A.saved = ''; markLookSaved(); toast(__('Look saved in the show file'));
          } }, __('Save look')))),
        h('div', { class: 'previews' }, h('div', {}, h('div', { class: 'hint' }, __('Phone')), phone), h('div', {}, h('div', { class: 'hint' }, __('Projector')), proj))));
  }

  // ---- step 5: test (rehearsal) ----
  async function stepTest(p) {
    const show = curShow(); if (!show) return go('show');
    p.append(h('h2', {}, __('Test with a rehearsal')),
      h('p', { class: 'hint' }, __('A rehearsal runs this show privately: you press the keys as the operator would and see the audience screen beside it. Nothing here reaches real audience devices, and it does not affect a show that is on air.')));
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

  // ---- step 6: run (share the address, open the operator page) ----
  async function stepGoLive(p) {
    const show = curShow(); if (!show) return go('show');
    const [net, sc, lg] = await Promise.all([api('GET', '/api/admin/network'), api('GET', '/api/admin/shows/' + show.id + '/showfile'), api('GET', '/api/admin/shows/' + show.id + '/languages'), loadShows()]);
    if (!net.ok || !sc.ok || !lg.ok) return fail(!net.ok ? net : !sc.ok ? sc : lg);
    const langs = lg.data.languages;
    A.analysis = sc.data.analysis; A.net = net.data;
    const a = A.analysis, errs = a.diagnostics.filter((d) => d.severity === 'error').length, warns = a.diagnostics.filter((d) => d.severity === 'warning').length;
    const isActive = A.activeId === show.id;
    const checklist = h('div', { class: 'readylist' }, h('h3', {}, __('Ready?')),
      h('div', { class: 'check' + (a.cues.length ? '' : ' bad') }, a.cues.length ? __('{n} lines in the show file', { n: a.cues.length }) : __('The show file is empty')),
      h('div', { class: 'check' + (errs ? ' bad' : warns ? ' warn' : '') }, errs ? __('{n} show file error(s) to fix', { n: errs }) : warns ? __('{n} show file warning(s) to review', { n: warns }) : __('No show file problems')),
      h('div', { class: 'check' + (a.missing.length ? ' warn' : '') }, a.missing.length ? __('Missing images: {list}', { list: a.missing.map((m) => m.name).join(', ') }) : __('All images are uploaded')),
      langs.length > 1 ? h('div', { class: 'check' }, __('Languages: {list}', { list: langs.filter((l) => l.available).map((l) => l.name).join(', ') })) : null,
      ...langs.filter((l) => !l.available).map((l) => h('div', { class: 'check bad' }, __('{name} will NOT be shown to the audience until it matches the main show file: {why}', { name: l.name, why: __((l.structure[0] || {}).message || '') }))));
    // The Start show button sits beside the checklist, where it is seen without scrolling.
    // Opening the operator page loads the show in standby: the audience keeps seeing the venue's standby screen until
    // the operator presses START there. So there is nothing here that puts a show in front of the audience.
    const onAir = !!((A.shows || []).find((x) => x.id === show.id) || {}).live;
    const startBox = h('div', { class: 'startbox' },
      isActive ? h('a', { class: 'btn primary big', href: '/operator' }, __('Open the Run page'))
        : h('button', { class: 'primary big', onclick: () => loadShow(show), disabled: !a.cues.length }, __('Open the Run page')),
      h('span', { class: 'hint' }, isActive ? (onAir ? __('This show is on air now.') : __('This show is loaded and waiting in standby.'))
        : __('Opening it loads it in standby: the audience keeps seeing the standby screen until the operator presses START.')),
      isActive ? h('button', { class: 'quiet', disabled: !!((A.shows || []).find((x) => x.id === show.id) || {}).fresh, title: ((A.shows || []).find((x) => x.id === show.id) || {}).fresh ? __('The show is at the start and in standby, so there is nothing to restart') : null, onclick: () => restartShow(show) }, __('Restart show…')) : null);
    const checks = h('div', { class: 'card ready' }, checklist, startBox);
    const urls = net.data.urls.filter((u) => u.public);
    let chosen = (urls.find((u) => u.cert) || urls.find((u) => u.kind === 'http' && !u.name) || urls.find((u) => u.kind === 'http') || urls[0] || net.data.urls[0] || {}).url; // the numeric address is the safest to print
    const big = h('div', { class: 'urlbig' }), qrBox = h('div', { class: 'qr' });
    const paint = () => { big.textContent = chosen || __('No address found'); qrBox.replaceChildren(chosen ? h('img', { src: '/api/admin/network/qr.png?size=512&url=' + encodeURIComponent(chosen), alt: __('QR code for {url}', { url: chosen }) }) : ''); };
    paint();
    // Every address is one row: [radio] link, and what belongs to that address. The name (open-caption.local) is a row while it is being
    // announced; whether it is, and what it is called, is set in Settings > Network.
    const m = net.data.mdns || {};
    const httpUrl = net.data.urls.find((u) => u.kind === 'http' && !u.name);
    const radio = (u) => h('input', { type: 'radio', name: 'u', checked: u.url === chosen, onchange: () => { chosen = u.url; paint(); } });
    const link = (url) => h('a', { href: url, target: '_blank', rel: 'noopener', title: __('Opens in a new window') }, url);
    const nameU = net.data.urls.find((u) => u.name);
    const nameRows = () => { // the name is set in Settings > Network; here it is only listed when it is being announced
      const out = [];
      if (nameU) out.push(h('div', { class: 'row' }, h('label', { class: 'row' }, radio(nameU), link(nameU.url))));
      if (nameU && !m.error && !m.renamed) out.push(h('div', { class: 'hint namehint' }, __('If a phone cannot open this name (some Android phones and guest WiFi networks cannot), give it the numeric address or the QR code instead.')));
      if (m.error) out.push(h('div', { class: 'check warn' }, __('Could not announce this name: {error}', { error: m.error })));
      else if (m.renamed) out.push(h('div', { class: 'check warn' }, __('Another computer on this network already uses {name}.local, so this one is http://{host}/', { name: m.requested, host: m.host })));
      return out;
    };
    // The projector link: the audience page for a shared screen (big text, no buttons, no pointer), opened on this computer,
    // which is the one the projector is plugged into, so it uses this computer's own address.
    // always the numeric address (never the .local name), so it can be copied, and chosen like the others
    const projBase = (net.data.urls.find((u) => u.kind === 'http' && !u.name && u.public) || httpUrl || net.data.urls[0] || {}).url;
    const projU = projBase ? { url: projBase + (projBase.includes('?') ? '&' : '?') + 'display=projector' } : null;
    const projRow = projU ? h('div', { class: 'row projrow' }, h('label', { class: 'row' }, radio(projU), link(projU.url)),
      h('span', { class: 'hint' }, __('(for a projector or shared screen; open it on that computer)'))) : null;
    const where = (u) => (u.adapter === 'wifi' ? ' (' + __('Wi-Fi') + ')' : u.adapter === 'ethernet' ? ' (' + __('Ethernet') + ')' : '');
    const addrRow = (u) => h('div', { class: 'row' }, h('label', { class: 'row' }, radio(u), link(u.url), where(u)));
    const numeric = (kind) => net.data.urls.filter((u) => !u.name && !u.cert && u.public && u.kind === kind).map(addrRow);
    const certRows = net.data.urls.filter((u) => u.cert).map(addrRow);
    const hasHttps = net.data.urls.some((u) => u.kind === 'https');
    const group = (title, ...kids) => h('div', { class: 'urlgroup' }, h('h4', { class: 'urlhead' }, title), ...kids);
    // localhost is of no use to the audience: one quiet line instead of rows
    const here = net.data.urls.filter((u) => !u.public && !u.name);
    const hereLine = here.length ? h('div', { class: 'hint localonly' }, __('This computer only:'), ' ', ...here.flatMap((u, i) => (i ? [' · ', link(u.url)] : [link(u.url)]))) : null;
    // The audience report of the last run (how many devices used the service), when there is one.
    const reportLink = h('p', { class: 'hint', hidden: '' }, h('a', { href: '#report', onclick: (e) => { e.preventDefault(); go('report'); window.scrollTo(0, 0); } }, __('Audience report')), ' ', __('(how many devices used the audience page during the last run of the show)'));
    api('GET', '/api/operator/report').then((x) => { if (x.ok) reportLink.hidden = false; }).catch(() => {});
    const plain = [...nameRows(), ...numeric('http')]; // the .local name first, then the numeric addresses
    const list = h('div', { class: 'urls' },
      ...(hasHttps ? [group(__('Secure (https)'), ...certRows, ...(certRows.length ? [h('div', { class: 'hint namehint' }, __('The name in the HTTPS certificate. It works on devices that can find this computer by that name.'))] : []), ...numeric('https')),
        group(__('Plain (http)'), ...plain)] : plain),
      projRow);
    p.append(h('h2', {}, __('Share')), checks,
      h('div', { class: 'card' }, h('h3', {}, __('Tell the audience this address')), big,
        h('div', { class: 'row' }, h('button', { onclick: async () => { try { await navigator.clipboard.writeText(chosen); toast(__('Copied')); } catch (e) { toast(__('Select the address and copy it')); } } }, __('Copy address')),
          h('a', { class: 'btn', href: '/admin?print=qr&url=' + encodeURIComponent(chosen || ''), target: '_blank', rel: 'noopener' }, __('Print a QR sign for the lobby'))),
        h('div', { class: 'row qrrow' }, qrBox, list), hereLine,
        ...(net.data.warnings || []).map((w) => h('div', { class: 'check warn' }, w)),
        h('p', { class: 'hint' }, __('Audience devices must be on the same WiFi as this computer. Devices connected right now: '), h('b', { id: 'audcount' }, String(A.audience))),
        reportLink,
        h('p', { class: 'hint' }, __('Two networks, or the wrong address? '), h('a', { href: '#', onclick: (e) => { e.preventDefault(); openSettings('net'); } }, __('Choose which network interfaces are used')))),
      h('p', { class: 'hint' }, __('Something not working? '), h('a', { href: '#', onclick: (e) => { e.preventDefault(); openHelp('admin-troubleshooting'); } }, __('Troubleshooting help')), ' · ', h('a', { href: '/api/admin/support-bundle', download: 'support-bundle.zip' }, __('Download a support bundle')), ' · ', h('a', { href: '#', onclick: (e) => { e.preventDefault(); showLog(); } }, __('Recent log…'))));
  }
  // ---- step 7: report (REQ-070) ----
  // How many devices used the audience page while the show was on the air: the same report as the stand-alone page, drawn by
  // web/shared/report.js. It refreshes itself every few seconds while the show is on the air.
  async function stepReport(p) {
    let current = A.reportId || 'current';
    const sel = h('select', { 'aria-label': __('Report') });
    const csv = h('a', { class: 'btn', href: '/api/operator/report.csv', download: '' }, __('Download CSV'));
    const del = h('button', { class: 'danger', onclick: async () => {
      const r0 = (sel.selectedOptions[0] || {}).textContent || '';
      if (!await confirmBox(__('Delete this report?'), __('“{name}” will be permanently deleted. Download the CSV first if you may want it back.', { name: r0 }), __('Delete'), true)) return;
      const r = await api('DELETE', '/api/admin/reports?id=' + encodeURIComponent(current));
      if (!r.ok) return fail(r);
      current = A.reportId = 'current'; toast(__('The report was deleted')); await loadList(); load();
    } }, __('Delete…'));
    const none = h('p', { hidden: '' }, __('There is no audience report yet. One is made each time the show is put on air.'));
    const body = h('div', {});
    async function loadList() {
      const r = await api('GET', '/api/operator/reports');
      const items = r.ok ? r.data.reports : [];
      sel.replaceChildren(...items.map((x) => h('option', { value: x.id }, OCReport.menuLabel(x))));
      if (!items.length) sel.replaceChildren(h('option', { value: 'current' }, '—'));
      else if (!items.some((x) => x.id === current)) current = items[0].id;
      sel.value = current;
      del.hidden = !items.length || current === 'current';
    }
    async function load() {
      clearTimeout(A.reportTimer);
      const r = await api('GET', '/api/operator/report' + (current === 'current' ? '' : '?id=' + encodeURIComponent(current)));
      if (A.step !== 'report') return;
      del.hidden = current === 'current';
      if (r.status === 404) { body.replaceChildren(); none.hidden = false; del.hidden = true; return; }
      if (!r.ok) return;
      none.hidden = true;
      body.replaceChildren(OCReport.build(r.data));
      csv.href = '/api/operator/report.csv' + (current === 'current' ? '' : '?id=' + encodeURIComponent(current));
      if (r.data.report.running) A.reportTimer = setTimeout(() => { if (A.step === 'report') load(); }, 5000);
    }
    sel.addEventListener('change', () => { current = A.reportId = sel.value; del.hidden = current === 'current'; load(); });
    const print = () => { document.body.classList.add('printreport'); window.print(); document.body.classList.remove('printreport'); }; // only the report is printed, not Setup around it
    p.append(h('h2', {}, __('Audience report')),
      h('div', { class: 'rtools noprint' }, h('label', {}, __('Report') + ' ', sel), h('button', { onclick: async () => { await loadList(); load(); } }, __('Refresh')), h('button', { onclick: print }, __('Print')), csv, del,
        h('a', { class: 'btn', href: '/operator/report', target: '_blank', rel: 'noopener' }, __('Open the report page'))),
      none, body);
    await loadList(); await load();
  }

  // Load a show into standby and go to the operator page. Another show that is on air is only replaced after asking.
  async function loadShow(show) {
    let r = await api('POST', '/api/admin/shows/' + show.id + '/activate', {});
    if (r.status === 409 && r.data.live) {
      if (!await confirmBox(__('Replace the show on air?'), __('Another show is on air right now. Loading this one takes it off the air and replaces it. Continue?'), __('Load this show'), true)) return;
      r = await api('POST', '/api/admin/shows/' + show.id + '/activate', { force: true });
    }
    if (!r.ok) return fail(r);
    A.activeId = show.id; refreshNav();
    location.href = '/operator';
  }
  // Take the loaded show back to its beginning, in standby.
  async function restartShow(show) {
    if (((A.shows || []).find((x) => x.id === show.id) || {}).live && !await confirmBox(__('Restart the show?'), __('The show goes back to the beginning and the audience sees the standby screen.'), __('Restart'), true)) return;
    const r = await api('POST', '/api/admin/shows/' + show.id + '/activate', { force: true });
    if (!r.ok) return fail(r);
    toast(__('The show is back in standby')); render();
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
