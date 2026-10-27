// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Operator page: follows the show file, sends NEXT/BACK/STOP, shows what the audience sees.
(() => {
  'use strict';
  if (window.OC_UNSUPPORTED) return; // compat.js has already shown the "too old" message
  const $ = (id) => document.getElementById(id);
  // In a rehearsal (started from Setup) the same page drives an isolated, private show.
  const REHEARSAL = new URLSearchParams(location.search).has('rehearsal');
  const FOLLOW = /^\/follow\/?$/.test(location.pathname); // the page for a second device that only watches: no controls, and its own sign-in
  const API = REHEARSAL ? '/api/admin/rehearsal' : '/api/operator';
  const S = { cues: [], payload: null, me: '', cuesVer: -1, sel: null, inflight: false, es: null, lastEvent: 0, rows: [] };

  const ICON = { clear: '🧹', blank: '⬛', image: '🖼', fullscreen: '⛶', loop: '🔁', section: '📑', note: '📝' };
  const isOpOnly = (k) => k === 'section' || k === 'note';
  const op = () => (S.payload && S.payload.op) || { cursor: 0, nextIndex: -1, rev: 0, total: 0, mode: 'idle' };
  const isController = () => S.payload && S.payload.controller === S.me;

  // ---- api ----
  async function api(method, path, body) {
    const r = await fetch(path, { method, headers: { 'Content-Type': 'application/json' }, body: body ? JSON.stringify(body) : undefined });
    let data = {};
    try { data = await r.json(); } catch (e) { /* empty body */ }
    return { ok: r.ok, status: r.status, data };
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
  function toast(msg, ms) {
    const t = $('toast');
    t.textContent = __(msg);
    t.classList.add('show');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => t.classList.remove('show'), ms || 2500);
  }

  // ---- boot / login ----
  async function boot() {
    let r;
    try { r = await api('GET', API + '/state'); } catch (e) { r = { ok: false, status: 0, data: {} }; }
    if (r.status === 401 || r.status === 403) { showLogin(); return; }
    if (r.status === 404) { document.body.textContent = __('No rehearsal is running. Start one from the Test step in Setup.'); return; }
    if (!r.ok) { // not answering (restarting?) or a fault: say so and keep trying rather than fail on missing data
      $('connbar').hidden = false;
      setTimeout(boot, 2000);
      return;
    }
    if (REHEARSAL) document.title = __('Rehearsal') + ' — ' + __('Operator');
    $('login').hidden = true;
    $('app').hidden = false;
    S.me = r.data.me;
    if (r.data.role === 'viewer' && !FOLLOW) { location.replace('/follow'); return; } // a viewer password opens only the page that watches
    S.viewer = FOLLOW || r.data.role === 'viewer'; // this page only watches (nothing can be sent or taken over)
    document.body.classList.toggle('follower', S.viewer);
    $('runlink').hidden = !FOLLOW || (r.data.role !== 'admin' && r.data.role !== 'operator'); // an operator or admin may go from watching to running
    $('setuplink').hidden = r.data.role !== 'admin' || REHEARSAL; // signed in with the admin password: a way back to Setup
    $('logout').hidden = !!r.data.auto; // on the computer that runs the program there is no password to sign out of
    applyState(r.data.cues, r.data.payload);
    if (!S.viewer) await claim(false);
    openStream();
  }
  function showLogin() {
    $('app').hidden = true; $('login').hidden = false;
    document.querySelectorAll('.for-run').forEach((e) => { e.hidden = FOLLOW; });
    document.querySelectorAll('.for-follow').forEach((e) => { e.hidden = !FOLLOW; });
    $('pw').focus();
  }
  $('login').addEventListener('submit', async (e) => {
    e.preventDefault();
    const r = await api('POST', '/api/login', { role: FOLLOW ? 'viewer' : 'operator', password: $('pw').value });
    if (r.ok) { $('loginerr').textContent = ''; $('pw').value = ''; boot(); return; }
    $('loginerr').textContent = r.status === 429 ? __('Too many attempts. Wait a minute.') : __('That password was not accepted.');
  });
  $('logout').addEventListener('click', async () => { if (S.es) S.es.close(); await api('POST', '/api/logout'); showLogin(); });

  async function claim(takeover) {
    const r = await api('POST', API + '/claim', { takeover });
    if (r.data.payload) setPayload(r.data.payload);
    return r.ok;
  }
  $('takeover').addEventListener('click', async () => { if (await claim(true)) toast(__('You are now in control')); });

  // ---- live updates ----
  function openStream() {
    if (S.parked) return; // the page is on its way out (or in the back/forward cache): it must not hold a connection
    if (S.es) S.es.close();
    const es = new EventSource(API + '/stream' + (S.viewer ? '?follow=1' : '')); // (a page that only watches says so, to be listed with the devices)
    S.es = es;
    S.lastEvent = Date.now();
    const up = () => { S.lastEvent = Date.now(); clearTimeout(S.lostTimer); S.lostTimer = 0; $('connbar').hidden = true; };
    es.addEventListener('opstate', (e) => { up(); setPayload(JSON.parse(e.data)); });
    es.addEventListener('ping', up);
    // (a failed retry must not push the notice back, so the timer is only started once)
    es.onerror = () => { es.close(); if (!S.lostTimer) S.lostTimer = setTimeout(() => { if (!S.parked) $('connbar').hidden = false; }, 2000); setTimeout(openStream, 1000 + Math.random() * 500); };
  }
  setInterval(() => { if (S.es && Date.now() - S.lastEvent > 40000) openStream(); }, 5000);
  // A browser keeps only six connections to one address, and every open page holds a live stream. A page that has been
  // navigated away from can stay alive in the back/forward cache with its stream open, so moving back and forth between
  // Setup and this page used up all six and made the next page hang. Give the stream up as soon as the page is left.
  window.addEventListener('pagehide', () => { S.parked = true; if (S.es) { S.es.close(); S.es = null; } });
  window.addEventListener('pageshow', (e) => { if (e.persisted) { S.parked = false; boot(); } }); // restored from the cache: start afresh

  async function setPayload(p) {
    const needCues = S.cuesVer !== p.cuesVer;
    // After a server restart nobody holds control; take it back if this page had it.
    if (S.hadControl && !p.controller && !S.reclaiming) {
      S.reclaiming = true;
      claim(false).finally(() => { S.reclaiming = false; });
    }
    S.hadControl = p.controller === S.me;
    S.payload = p;
    if (needCues) {
      const r = await api('GET', API + '/state');
      if (r.ok) {
        // A newer update may have arrived while the show was being fetched; never put an older one back over it.
        const newer = S.payload !== p;
        if (newer && S.payload.cuesVer !== r.data.payload.cuesVer) return; // that update fetches the show itself
        applyState(r.data.cues, newer ? S.payload : r.data.payload);
        return;
      }
    }
    renderStatus();
    renderMarkers(true);
  }

  // The loop's countdown, shown to the operator too ("⏱ 13 min left"), so they know when to start the next line.
  function paintBannerTimer() {
    const t = $('bannertimer');
    if (!t || !S.timerEnds) return;
    const total = Math.max(0, Math.ceil((S.timerEnds - Date.now()) / 60000)), h = Math.floor(total / 60), m = total % 60;
    const unit = (n, u) => { try { return new Intl.NumberFormat(OCI18n.lang, { style: 'unit', unit: u, unitDisplay: 'short' }).format(n); } catch (e) { return n + (u === 'hour' ? ' h' : ' min'); } };
    t.textContent = '⏱ ' + __('{time} left on the timer', { time: h ? unit(h, 'hour') + (m ? ' ' + unit(m, 'minute') : '') : unit(total, 'minute') });
  }
  setInterval(paintBannerTimer, 500);

  // Sections double as a mini-map: pick one to jump to it (with the usual confirmation).
  function buildSections() {
    const sel = $('sections');
    const secs = S.cues.filter((c) => c.kind === 'section');
    sel.replaceChildren(new Option(__('Go to section…'), ''), ...secs.map((c) => new Option(c.text, String(c.index))));
    sel.hidden = secs.length === 0;
  }
  $('sections').addEventListener('change', () => {
    const v = $('sections').value;
    $('sections').value = '';
    $('sections').blur(); // give the keys (Enter to jump, Esc to cancel) back to the page
    if (v === '') return;
    const i = Number(v);
    select(i); scrollTo(i);
  });

  // "Last sent 12 s ago": a quick check that the show and this page are still in step.
  const span = (s) => (s < 60 ? s + ' s' : Math.floor(s / 60) + ' min ' + (s % 60) + ' s');
  function renderLastSent() {
    const el = $('lastsent'), o = S.payload && S.payload.op, at = o && o.lastAt;
    if (!at) { el.textContent = ''; el.classList.remove('due'); return; }
    const s = Math.max(0, Math.round((Date.now() - at) / 1000));
    let text = __('Last sent {when} ago', { when: span(s) }), due = false;
    // A [pace] in the show file says how long the line on screen usually stays up: count down to when the next press is due.
    const c = S.cues[currentIndex()];
    if (c && c.hint && !o.standby) {
      const left = Math.round(c.hint - (Date.now() - (o.sentAt || at)) / 1000); // from when that line was sent, not from the last action of any kind
      due = left <= 0;
      text += ' · ' + (due ? __('Next is due') : __('Next is due in {when}', { when: span(left) }));
    }
    el.textContent = text;
    el.classList.toggle('due', due);
    el.classList.toggle('stale', s >= 120);
  }
  setInterval(renderLastSent, 1000);

  function applyState(cues, payload) {
    S.cues = cues;
    S.payload = payload;
    S.cuesVer = payload.cuesVer;
    S.sel = null;
    buildList();
    buildSections();
    renderStatus();
    renderMarkers(true);
  }

  // ---- rendering ----
  function buildList() {
    const ol = $('list');
    ol.replaceChildren();
    S.rows = S.cues.map((c) => {
      const li = document.createElement('li');
      li.className = 'cue k-' + c.kind + (c.stage ? ' stage' : '') + (c.sung ? ' sung' : '');
      li.dataset.i = c.index;
      const gut = document.createElement('span'); gut.className = 'gut';
      const body = document.createElement('div'); body.className = 'body';
      fillBody(body, c);
      if (c.hint) body.append(chip('⏱ ' + __('about {when}', { when: span(Math.round(c.hint)) }))); // from [pace] in the show file
      const ln = document.createElement('span'); ln.className = 'ln'; ln.textContent = c.line ? String(c.line) : '';
      li.append(gut, body, ln);
      ol.appendChild(li);
      return li;
    });
  }

  function chip(text) { const s = document.createElement('span'); s.className = 'chip'; s.textContent = text; return s; }

  function fillBody(body, c) {
    switch (c.kind) {
      case 'caption': {
        if (c.parts && c.parts.length) { // several voices at once
          body.append(chip('⏸ ' + __('TOGETHER')));
          c.parts.forEach((v) => {
            const d = document.createElement('div'); d.className = 'part' + (v.sung ? ' sung' : '');
            if (v.speaker) { const s = document.createElement('span'); s.className = 'spk'; s.textContent = v.speaker; d.appendChild(s); }
            const t = document.createElement('span'); t.className = 'txt'; t.textContent = v.text; d.appendChild(t);
            body.appendChild(d);
          });
          break;
        }
        if (c.speaker) { const s = document.createElement('span'); s.className = 'spk'; s.textContent = c.speaker; body.appendChild(s); }
        const t = document.createElement('span'); t.className = 'txt'; t.textContent = c.text; t.style.whiteSpace = 'pre-line'; t.dir = 'auto'; body.appendChild(t);
        break;
      }
      case 'clear': body.append(chip(ICON.clear + ' ' + __('CLEAR SCREEN'))); break;
      case 'blank': body.append(chip(ICON.blank + ' ' + __('BLANK SCREEN'))); break;
      case 'image': body.append(chip(ICON.image + ' ' + __('IMAGE: {name}', { name: c.label }))); break; // name only, never the picture (REQ-031)
      case 'fullscreen': body.append(chip(ICON.fullscreen + ' ' + __('FULLSCREEN')), document.createTextNode(c.label)); break;
      case 'loop': {
        const secs = c.period ? c.period + ' s' : __('default time');
        body.append(chip(ICON.loop + ' ' + __('LOOP — {n} items, {time}', { n: (c.items || []).length, time: secs })));
        const ul = document.createElement('ul'); ul.className = 'items';
        (c.items || []).forEach((it) => {
          const li = document.createElement('li');
          const mark = document.createElement('span'); mark.className = 'pl'; mark.setAttribute('aria-hidden', 'true'); li.appendChild(mark); // ▶ while this item is on the audience screens
          li.appendChild(document.createTextNode((it.kind === 'image' ? ICON.image + ' ' : it.kind === 'blank' ? ICON.blank + ' ' : '') + it.label + (it.seconds ? ' (' + it.seconds + ' s)' : '')));
          ul.appendChild(li);
        });
        body.appendChild(ul);
        break;
      }
      case 'section': body.textContent = ICON.section + ' ' + c.text; break;
      case 'note': body.textContent = ICON.note + ' ' + __('NOTE: {text}', { text: c.text }); break;
      default: body.textContent = c.text || c.kind;
    }
  }

  function currentIndex() {
    const o = op();
    for (let i = o.cursor - 1; i >= 0; i--) if (S.cues[i] && !isOpOnly(S.cues[i].kind)) return i;
    return -1;
  }

  function renderMarkers(scroll) {
    const o = op(), now = currentIndex();
    S.rows.forEach((li, i) => {
      li.classList.toggle('sent', i < o.cursor && i !== now && i !== o.nextIndex);
      li.classList.toggle('now', i === now && !o.standby); // nothing is on the audience screens while in standby
      li.classList.toggle('next', i === o.nextIndex);
      li.classList.toggle('sel', i === S.sel);
      // a line passed over with Skip: struck through, and says so in words
      const skipped = !!o.skipped && o.skipped.indexOf(i) >= 0, tag = li.querySelector('.skiptag');
      li.classList.toggle('skipped', skipped);
      if (skipped && !tag) { const t = document.createElement('span'); t.className = 'chip skiptag'; t.textContent = __('SKIPPED'); li.querySelector('.body').prepend(t); }
      else if (!skipped && tag) tag.remove();
      // a running loop: a small play triangle beside the item the audience is looking at now
      const shown = o.loop && o.loop.cue === i ? o.loop.item : -1;
      li.querySelectorAll('.items li').forEach((it, k) => { it.classList.toggle('playing', k === shown); if (k === shown) it.setAttribute('aria-current', 'true'); else it.removeAttribute('aria-current'); });
    });
    if (scroll) scrollTo(o.nextIndex >= 0 ? o.nextIndex : now);
    // "Line 34 of 210": the audience-visible lines up to the one on screen (notes and section headings are not counted)
    let total = 0, at = 0;
    S.cues.forEach((c, i) => { if (!isOpOnly(c.kind)) { total++; if (i <= now) at++; } });
    $('progress').textContent = total ? (o.standby ? __('{total} items', { total }) : __('Item {n} of {total}', { n: at, total })) : '';
  }

  function scrollTo(i) {
    const li = S.rows[i];
    if (!li) return;
    const list = $('list');
    list.scrollTo({ top: li.offsetTop - list.offsetTop - list.clientHeight / 3, behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' });
  }

  function renderStatus() {
    const p = S.payload; if (!p) return;
    const o = p.op;
    $('show').textContent = (REHEARSAL ? __('REHEARSAL') + ' — ' : '') + (p.show.name || p.show.id || __('No show loaded'));
    if (!REHEARSAL) document.title = (S.viewer ? __('Follow') : __('Run')) + ' — ' + (p.show.name || p.show.id || __('No show loaded')) + ' (' + (o.live ? __('ON AIR') : __('STANDBY')) + ')';
    const live = $('live'); live.textContent = o.live ? '● ' + __('ON AIR') : __('STANDBY'); live.classList.toggle('on', o.live); live.classList.toggle('standby', !!o.standby);
    $('viewers').textContent = OCClients.label(p); // languages nobody is using are left out
    $('viewers').title = OCClients.title(p.audience);
    const mine = isController();
    document.body.classList.toggle('viewonly', !mine);
    $('ctrl').textContent = S.viewer ? __('You are following this show (view only)') : mine ? __('You are in control') : (p.controller ? __('Another operator is in control') : __('Nobody is in control'));
    $('takeover').hidden = mine || S.viewer; $('statusbar').classList.toggle('other', !mine && !S.viewer);
    $('takeover').textContent = p.controller ? __('Take over') : __('Take control');
    $('stop').disabled = !o.loop; // always in place, so the buttons never move while the show runs
    // Standby: the show is not on air. NEXT becomes START (same place, same key), the rest wait.
    document.body.classList.toggle('standby', !!o.standby);
    $('offair').disabled = !!o.standby;
    $('reset').disabled = !!o.atStart; // just started or just restarted: nothing to restart
    ['back', 'skip', 'blank', 'clear', 'message'].forEach((id) => { $(id).disabled = !!o.standby; }); // (a disabled control is exempt from the contrast rule, and says what it is)
    $('next').firstChild.textContent = o.standby ? __('START ▶') : __('NEXT ▶');
    $('next').title = o.standby ? __('Put the show on air (Space)') : __('Next (Space)');
    renderLastSent();
    $('blank').firstChild.textContent = o.mode === 'blank' ? __('Unblank') : __('Blank');
    const b = $('banner');
    S.timerEnds = o.loop && o.loop.deadline ? o.loop.deadline : 0;
    if (o.standby) { b.hidden = false; b.className = 'standby'; b.textContent = '⏸ ' + __('STANDBY — the audience sees the standby screen. Press Space (START) to put the show on air; it carries on from where it is.'); }
    else if (o.announcement) { b.hidden = false; b.className = 'announce'; b.textContent = '📢 ' + __('ANNOUNCEMENT ON SCREEN: “{text}” — Back removes it; the next line replaces it.', { text: o.announcement.length > 100 ? o.announcement.slice(0, 100) + '…' : o.announcement }); }
    else if (o.loop) {
      const c = S.cues[o.loop.cue]; b.hidden = false; b.className = 'loop';
      b.textContent = '🔁 ' + __('LOOP RUNNING — item {n} of {total}. NEXT ends it; STOP blanks the screen.', { n: o.loop.item + 1, total: c ? (c.items || []).length : '?' });
      if (S.timerEnds) { const t = document.createElement('b'); t.id = 'bannertimer'; b.appendChild(document.createTextNode(' ')); b.appendChild(t); paintBannerTimer(); }
    }
    else if (o.mode === 'blank') { b.hidden = false; b.className = ''; b.textContent = '⬛ ' + __('Audience screen is blank. Space (NEXT) sends the next line, B shows the earlier lines again, Back undoes.'); }
    else if (o.nextIndex < 0 && o.total > 0 && o.cursor >= o.total) { b.hidden = false; b.className = ''; b.textContent = __('End of the show.') + ' '; const rl = document.createElement('a'); rl.href = '/operator/report'; rl.target = '_blank'; rl.rel = 'noopener'; rl.className = 'reportlink'; rl.textContent = __('Audience report'); if (!REHEARSAL && !S.viewer) b.appendChild(rl); }
    else b.hidden = true;
  }

  // ---- commands ----
  async function send(name, extra) {
    if (S.viewer) { toast(__('This screen is view only')); return; }
    if (!isController()) { toast(__('Another operator is in control — use Take over'), 6000); return; }
    if (S.inflight) return;
    S.inflight = true;
    const passed = name === 'skip' ? S.cues[op().nextIndex] : null; // the line a Skip passes over
    try {
      const r = await api('POST', API + '/' + name, Object.assign({ rev: op().rev }, extra));
      if (r.data.payload) setPayload(r.data.payload);
      if (r.status === 409) toast(__(r.data.error || 'Out of date — refreshed'));
      else if (r.data.notice) toast(__(r.data.notice));
      else if (!r.ok) toast(__(whyFailed(r)), 6000); // a problem stays long enough to read
      else if (passed) { const t = passed.text || passed.label || ''; toast(t ? __('Skipped “{text}”', { text: t.length > 60 ? t.slice(0, 60) + '…' : t }) : __('Skipped'), 4000); }
      if (r.ok && name !== 'jump') clearSel();
    } catch (e) { toast(__('Connection problem — try again'), 6000); }
    finally { S.inflight = false; }
  }

  const next = () => (op().standby ? send('onair') : S.sel !== null ? confirmJump() : send('next'));
  $('next').addEventListener('click', () => send(op().standby ? 'onair' : 'next'));
  $('back').addEventListener('click', () => send('back'));
  $('stop').addEventListener('click', () => send('stop'));
  $('blank').addEventListener('click', () => send('blank', { on: op().mode !== 'blank' }));
  $('clear').addEventListener('click', () => send('clear'));
  $('skip').addEventListener('click', () => send('skip'));

  // jump: select a row, then confirm
  function select(i) {
    S.sel = i;
    const c = S.cues[i];
    $('jumpbar').hidden = false;
    $('jumptext').textContent = __('Jump so that line {n} is next: {text}', { n: c.line || i + 1, text: (c.text || c.label || c.kind).slice(0, 80) });
    renderMarkers(false);
  }
  function clearSel() { S.sel = null; $('jumpbar').hidden = true; renderMarkers(false); }
  function confirmJump() { const i = S.sel; if (i === null) return; send('jump', { index: i }).then(clearSel); }
  $('list').addEventListener('click', (e) => {
    const li = e.target.closest('.cue'); if (!li || S.viewer) return;
    const i = Number(li.dataset.i);
    if (S.sel === i) confirmJump(); else select(i);
  });
  $('jumpgo').addEventListener('click', confirmJump);
  $('jumpcancel').addEventListener('click', clearSel);

  // search
  $('search').addEventListener('input', () => {
    const q = $('search').value.trim().toLowerCase();
    let first = -1;
    S.rows.forEach((li, i) => {
      const c = S.cues[i];
      const hit = q && ((c.text || '') + ' ' + (c.label || '') + ' ' + (c.speaker || '')).toLowerCase().includes(q);
      li.classList.toggle('match', !!hit);
      if (hit && first < 0) first = i;
    });
    if (first >= 0) { select(first); scrollTo(first); } else if (!q) clearSel();
  });
  $('search').addEventListener('keydown', (e) => { if (e.key === 'Enter') { e.preventDefault(); $('search').blur(); } if (e.key === 'Escape') { $('search').value = ''; $('search').dispatchEvent(new Event('input')); $('search').blur(); } });

  // dialogs
  $('message').addEventListener('click', () => { $('msgtext').value = ''; $('msgdlg').showModal(); $('msgtext').focus(); });
  $('msgdlg').addEventListener('close', () => { if ($('msgdlg').returnValue === 'send' && $('msgtext').value.trim()) send('message', { text: $('msgtext').value.trim() }); });
  $('offair').addEventListener('click', () => $('offdlg').showModal());
  $('offdlg').addEventListener('close', () => { if ($('offdlg').returnValue === 'go') send('offair'); });
  $('reset').addEventListener('click', () => $('resetdlg').showModal());
  // The audience count opens the list of connected devices (device, address, language), kept up to date while it is open.
  $('viewers').addEventListener('click', async () => {
    if (S.viewer) return; // the list of devices is for the operator
    let timer = 0;
    const fill = async () => { const r = await api('GET', API + '/clients'); if (r.ok) $('clientsbody').replaceChildren(OCClients.table(r.data.clients)); };
    await fill();
    timer = setInterval(fill, 3000);
    $('clientsdlg').addEventListener('close', () => clearInterval(timer), { once: true });
    $('clientsdlg').showModal();
  });
  $('resetdlg').addEventListener('close', () => { if ($('resetdlg').returnValue === 'go') send('reset'); });

  // ---- help ----
  // While help is open every show-control key is ignored, so reading help can never send a line.
  let helpOpen = false;
  function openHelp(topic) {
    helpOpen = true;
    document.body.classList.add('helpopen');
    $('help').hidden = false;
    $('help').querySelector('iframe').src = '/help?role=operator&embed=1&ui=' + OCI18n.lang + '&topic=' + (topic || 'operator-running-the-show');
    $('help').querySelector('iframe').focus();
  }
  function closeHelp() { helpOpen = false; document.body.classList.remove('helpopen'); $('help').hidden = true; $('help').querySelector('iframe').src = 'about:blank'; document.body.focus(); }
  $('helpbtn').addEventListener('click', () => (helpOpen ? closeHelp() : openHelp()));
  $('helpclose').addEventListener('click', closeHelp);
  window.addEventListener('message', (e) => { if (e.origin === location.origin && e.data && e.data.type === 'help-close') closeHelp(); });

  // ---- keyboard ----
  // Keys are ignored while typing or while a dialog is open, so they can never advance the show by accident.
  let lastKey = 0;
  document.addEventListener('keydown', (e) => {
    if (e.repeat || e.ctrlKey || e.metaKey || e.altKey) return;
    if (helpOpen) { if (e.key === 'Escape' || e.key === 'F1' || e.key === '?') { e.preventDefault(); closeHelp(); } return; }
    if (document.querySelector('dialog[open]')) return;
    if (S.viewer && e.key !== '?' && e.key !== 'F1') return; // a viewer's keys do nothing
    const typing = /^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement.tagName);
    if (typing) return;
    let handled = true;
    const now = Date.now();
    const quick = now - lastKey < 120; // debounce double-fire from bouncing clickers
    // Not on air: only START (Space and the other advance keys), search, Restart and Help do anything.
    if (op().standby && ['Backspace', 'ArrowLeft', 'PageUp', 'b', 'B', 'c', 'C', 's', 'S', 'Q'].indexOf(e.key) >= 0) return;
    switch (e.key) {
      case ' ': case 'ArrowRight': case 'PageDown': case 'Enter': if (!quick) next(); break;
      case 'Backspace': case 'ArrowLeft': case 'PageUp': if (!quick) send('back'); break;
      case 'b': case 'B': if (!e.shiftKey && !quick) send('blank', { on: op().mode !== 'blank' }); break;
      case 'c': case 'C': if (!e.shiftKey && !quick) send('clear'); break;
      case 's': case 'S': if (!e.shiftKey && !quick) send('skip'); break;
      case 'Q': if (!quick) send('stop'); break;
      case '/': $('search').focus(); break;
      case 'Home': if (!op().atStart) $('resetdlg').showModal(); break;
      case 'Escape': clearSel(); break;
      case '?': case 'F1': openHelp(); break;
      default: handled = false;
    }
    if (handled) { e.preventDefault(); if (document.activeElement && document.activeElement.blur && document.activeElement.tagName === 'BUTTON') document.activeElement.blur(); lastKey = now; }
  });
  // A focused button would also "click" on Space release; we already handled the key.
  document.addEventListener('keyup', (e) => { if (e.key === ' ' && !/^(INPUT|TEXTAREA)$/.test(document.activeElement.tagName)) e.preventDefault(); });

  OCI18n.apply();
  $('offair').parentNode.insertBefore(OCI18n.picker(), $('offair')); // after the audience count, before the actions
  $('login').insertBefore(OCI18n.picker(), $('login').firstChild);

  boot();
})();
