// Operator page: follows the script, sends NEXT/BACK/STOP, shows what the audience sees.
(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  // In a rehearsal (started from Setup) the same page drives an isolated, private show.
  const REHEARSAL = new URLSearchParams(location.search).has('rehearsal');
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

  let toastTimer;
  function toast(msg) {
    const t = $('toast');
    t.textContent = msg;
    t.classList.add('show');
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => t.classList.remove('show'), 2500);
  }

  // ---- boot / login ----
  async function boot() {
    const r = await api('GET', API + '/state');
    if (r.status === 401 || r.status === 403) { showLogin(); return; }
    if (r.status === 404) { document.body.textContent = 'No rehearsal is running. Start one from the Test step in Setup.'; return; }
    if (REHEARSAL) document.title = 'Rehearsal — Operator';
    $('login').hidden = true;
    $('app').hidden = false;
    S.me = r.data.me;
    applyState(r.data.cues, r.data.payload);
    await claim(false);
    openStream();
  }
  function showLogin() { $('app').hidden = true; $('login').hidden = false; $('pw').focus(); }
  $('login').addEventListener('submit', async (e) => {
    e.preventDefault();
    const r = await api('POST', '/api/login', { role: 'operator', password: $('pw').value });
    if (r.ok) { $('loginerr').textContent = ''; $('pw').value = ''; boot(); return; }
    $('loginerr').textContent = r.status === 429 ? 'Too many attempts. Wait a minute.' : 'That password was not accepted.';
  });
  $('logout').addEventListener('click', async () => { if (S.es) S.es.close(); await api('POST', '/api/logout'); showLogin(); });

  async function claim(takeover) {
    const r = await api('POST', API + '/claim', { takeover });
    if (r.data.payload) setPayload(r.data.payload);
    return r.ok;
  }
  $('takeover').addEventListener('click', async () => { if (await claim(true)) toast('You are now in control'); });

  // ---- live updates ----
  function openStream() {
    if (S.es) S.es.close();
    const es = new EventSource(API + '/stream');
    S.es = es;
    S.lastEvent = Date.now();
    es.addEventListener('opstate', (e) => { S.lastEvent = Date.now(); $('dot').classList.add('on'); setPayload(JSON.parse(e.data)); });
    es.addEventListener('ping', () => { S.lastEvent = Date.now(); });
    es.onerror = () => { $('dot').classList.remove('on'); es.close(); setTimeout(openStream, 1000 + Math.random() * 500); };
  }
  setInterval(() => { if (S.es && Date.now() - S.lastEvent > 40000) openStream(); }, 5000);

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
      if (r.ok) { applyState(r.data.cues, r.data.payload); return; }
    }
    renderStatus();
    renderMarkers(true);
  }

  function applyState(cues, payload) {
    S.cues = cues;
    S.payload = payload;
    S.cuesVer = payload.cuesVer;
    S.sel = null;
    buildList();
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
          body.append(chip('⏸ TOGETHER'));
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
      case 'clear': body.append(chip(ICON.clear + ' CLEAR SCREEN')); break;
      case 'blank': body.append(chip(ICON.blank + ' BLANK SCREEN')); break;
      case 'image': body.append(chip(ICON.image + ' IMAGE: ' + c.label)); break; // name only, never the picture (REQ-031)
      case 'fullscreen': body.append(chip(ICON.fullscreen + ' FULLSCREEN'), document.createTextNode(c.label)); break;
      case 'loop': {
        const secs = c.period ? c.period + ' s' : 'default time';
        body.append(chip(ICON.loop + ' LOOP — ' + (c.items || []).length + ' items, ' + secs));
        const ul = document.createElement('ul'); ul.className = 'items';
        (c.items || []).forEach((it) => {
          const li = document.createElement('li');
          li.textContent = (it.kind === 'image' ? ICON.image + ' ' : it.kind === 'blank' ? ICON.blank + ' ' : '') + it.label + (it.seconds ? ' (' + it.seconds + ' s)' : '');
          ul.appendChild(li);
        });
        body.appendChild(ul);
        break;
      }
      case 'section': body.textContent = ICON.section + ' ' + c.text; break;
      case 'note': body.textContent = ICON.note + ' NOTE: ' + c.text; break;
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
      li.classList.toggle('sent', i < o.cursor && i !== now);
      li.classList.toggle('now', i === now);
      li.classList.toggle('next', i === o.nextIndex);
      li.classList.toggle('sel', i === S.sel);
    });
    if (scroll) scrollTo(o.nextIndex >= 0 ? o.nextIndex : now);
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
    $('show').textContent = (REHEARSAL ? 'REHEARSAL — ' : '') + (p.show.name || p.show.id || 'No show loaded');
    const live = $('live'); live.textContent = o.live ? '● LIVE' : 'NOT LIVE'; live.classList.toggle('on', o.live);
    $('viewers').textContent = '👥 ' + p.audience;
    const mine = isController();
    document.body.classList.toggle('viewonly', !mine);
    $('ctrl').textContent = mine ? 'You are in control' : (p.controller ? 'Another operator is in control' : 'Nobody is in control');
    $('takeover').hidden = mine;
    $('takeover').textContent = p.controller ? 'Take over' : 'Take control';
    $('stop').hidden = !o.loop;
    $('blank').firstChild.textContent = o.mode === 'blank' ? 'Unblank' : 'Blank';
    const b = $('banner');
    if (o.loop) { const c = S.cues[o.loop.cue]; b.hidden = false; b.className = 'loop'; b.textContent = '🔁 LOOP RUNNING — item ' + (o.loop.item + 1) + ' of ' + (c ? (c.items || []).length : '?') + '. NEXT ends it; STOP blanks the screen.'; }
    else if (o.mode === 'blank') { b.hidden = false; b.className = ''; b.textContent = '⬛ Audience screen is blank.'; }
    else if (o.nextIndex < 0 && o.total > 0 && o.cursor >= o.total) { b.hidden = false; b.className = ''; b.textContent = 'End of script.'; }
    else b.hidden = true;
  }

  // ---- commands ----
  async function send(name, extra) {
    if (!isController()) { toast('Another operator is in control — use Take over'); return; }
    if (S.inflight) return;
    S.inflight = true;
    try {
      const r = await api('POST', API + '/' + name, Object.assign({ rev: op().rev }, extra));
      if (r.data.payload) setPayload(r.data.payload);
      if (r.status === 409) toast(r.data.error || 'Out of date — refreshed');
      else if (r.data.notice) toast(r.data.notice);
      else if (!r.ok) toast(r.data.error || 'Something went wrong');
      if (r.ok && name !== 'jump') clearSel();
    } catch (e) { toast('Connection problem — try again'); }
    finally { S.inflight = false; }
  }

  const next = () => (S.sel !== null ? confirmJump() : send('next'));
  $('next').addEventListener('click', () => send('next'));
  $('back').addEventListener('click', () => send('back'));
  $('stop').addEventListener('click', () => send('stop'));
  $('blank').addEventListener('click', () => send('blank', { on: op().mode !== 'blank' }));
  $('clear').addEventListener('click', () => send('clear'));

  // jump: select a row, then confirm
  function select(i) {
    S.sel = i;
    const c = S.cues[i];
    $('jumpbar').hidden = false;
    $('jumptext').textContent = 'Jump so that line ' + (c.line || i + 1) + ' is next: ' + (c.text || c.label || c.kind).slice(0, 80);
    renderMarkers(false);
  }
  function clearSel() { S.sel = null; $('jumpbar').hidden = true; renderMarkers(false); }
  function confirmJump() { const i = S.sel; if (i === null) return; send('jump', { index: i }).then(clearSel); }
  $('list').addEventListener('click', (e) => {
    const li = e.target.closest('.cue'); if (!li) return;
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
  $('reset').addEventListener('click', () => $('resetdlg').showModal());
  $('resetdlg').addEventListener('close', () => { if ($('resetdlg').returnValue === 'go') send('reset'); });

  // ---- help ----
  // While help is open every show-control key is ignored, so reading help can never send a line.
  let helpOpen = false;
  function openHelp(topic) {
    helpOpen = true;
    document.body.classList.add('helpopen');
    $('help').hidden = false;
    $('help').querySelector('iframe').src = '/help?role=operator&embed=1&topic=' + (topic || 'operator-running-the-show');
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
    const typing = /^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement.tagName);
    if (typing) return;
    let handled = true;
    const now = Date.now();
    const quick = now - lastKey < 120; // debounce double-fire from bouncing clickers
    switch (e.key) {
      case ' ': case 'ArrowRight': case 'PageDown': case 'Enter': if (!quick) next(); break;
      case 'Backspace': case 'ArrowLeft': case 'PageUp': if (!quick) send('back'); break;
      case 'b': case 'B': if (!e.shiftKey && !quick) send('blank', { on: op().mode !== 'blank' }); break;
      case 'c': case 'C': if (!e.shiftKey && !quick) send('clear'); break;
      case 'Q': if (!quick) send('stop'); break;
      case '/': $('search').focus(); break;
      case 'Home': $('resetdlg').showModal(); break;
      case 'Escape': clearSel(); break;
      case '?': case 'F1': openHelp(); break;
      default: handled = false;
    }
    if (handled) { e.preventDefault(); if (document.activeElement && document.activeElement.blur && document.activeElement.tagName === 'BUTTON') document.activeElement.blur(); lastKey = now; }
  });
  // A focused button would also "click" on Space release; we already handled the key.
  document.addEventListener('keyup', (e) => { if (e.key === ' ' && !/^(INPUT|TEXTAREA)$/.test(document.activeElement.tagName)) e.preventDefault(); });

  boot();
})();
