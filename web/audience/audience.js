// Audience page: renders the live state pushed by the server. No interaction is needed.
(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const params = new URLSearchParams(location.search);
  const projector = params.get('display') === 'projector';
  if (projector) document.body.classList.add('projector');

  const state = { mode: 'idle', scrollback: [], fullscreen: null, theme: {} };
  const preloaded = new Set();
  let es = null, lastEvent = 0, backoff = 500, offlineTimer = null, retryTimer = null;

  const FONTS = { system: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif', serif: 'Georgia, "Times New Roman", serif', mono: 'ui-monospace, Menlo, Consolas, monospace' };

  // ---- this device's own preferences (optional; never required) ----
  const store = {
    get(k) { try { return localStorage.getItem(k); } catch (e) { return null; } },
    set(k, v) { try { localStorage.setItem(k, v); } catch (e) { /* storage may be unavailable */ } },
  };
  let localStep = parseInt(store.get('oc.step'), 10) || 0; // -3..+6 steps of 15%
  let lightScheme = store.get('oc.light') === '1';
  function paintLocal() {
    const t = state.theme || {}, r = document.documentElement.style;
    const urlPct = parseInt(params.get('size'), 10);
    const pct = (urlPct > 0 ? urlPct : (t.scalePct || 100)) * (1 + 0.15 * localStep);
    r.setProperty('--scale', String(Math.max(0.4, pct / 100) * (projector ? 1.5 : 1)));
    if (lightScheme) { r.setProperty('--bg', '#ffffff'); r.setProperty('--fg', '#000000'); r.setProperty('--accent', '#000000'); }
    else {
      r.setProperty('--bg', t.bg || '#000000'); r.setProperty('--fg', t.fg || '#ffffff');
      r.setProperty('--accent', t.speakers === 'color' ? (t.accent || t.fg || '#ffffff') : (t.fg || '#ffffff'));
    }
    $('scheme').setAttribute('aria-pressed', String(lightScheme));
    $('scheme').textContent = lightScheme ? 'Light on dark' : 'Dark on light';
  }
  $('gear').addEventListener('click', () => { const open = $('ctl').hidden; $('ctl').hidden = !open; $('gear').setAttribute('aria-expanded', String(open)); });
  $('bigger').addEventListener('click', () => { localStep = Math.min(6, localStep + 1); store.set('oc.step', String(localStep)); paintLocal(); });
  $('smaller').addEventListener('click', () => { localStep = Math.max(-3, localStep - 1); store.set('oc.step', String(localStep)); paintLocal(); });
  $('scheme').addEventListener('click', () => { lightScheme = !lightScheme; store.set('oc.light', lightScheme ? '1' : '0'); paintLocal(); });

  // ---- rendering ----
  function applyTheme(t) {
    state.theme = t || {};
    const r = document.documentElement.style;
    r.setProperty('--font', FONTS[t.font] || FONTS.system);
    paintLocal();
    $('idle').textContent = t.idleMessage || '';
  }

  function voiceEl(v) {
    const d = document.createElement('div');
    d.className = 'part' + (v.sung ? ' sung' : '') + (v.stage ? ' stage' : '');
    if (v.speaker && state.theme.speakers !== 'hidden') {
      const s = document.createElement('span');
      s.className = 'spk';
      s.textContent = v.speaker;
      d.appendChild(s);
    }
    const body = document.createElement('span');
    body.innerHTML = v.html; // generated and escaped by the server
    d.appendChild(body);
    return d;
  }

  function captionEl(c) {
    if (c.parts && c.parts.length) { // several voices at the same time
      const g = document.createElement('div');
      g.className = 'cap group';
      g.dataset.id = c.id;
      g.setAttribute('role', 'group');
      c.parts.forEach((v) => g.appendChild(voiceEl(v)));
      return g;
    }
    const p = document.createElement('p');
    p.className = 'cap' + (c.stage ? ' stage' : '') + (c.sung ? ' sung' : '');
    p.dataset.id = c.id;
    if (c.speaker && state.theme.speakers !== 'hidden') {
      const s = document.createElement('span');
      s.className = 'spk';
      s.textContent = c.speaker;
      p.appendChild(s);
    }
    const body = document.createElement('span');
    body.innerHTML = c.html; // generated and escaped by the server
    p.appendChild(body);
    return p;
  }

  function fade() {
    // Older lines dim with age (newest = full brightness).
    const els = $('captions').children;
    for (let i = 0; i < els.length; i++) {
      const age = els.length - 1 - i;
      // Older lines fade, but never below a level that is still readable.
      els[i].style.opacity = age === 0 ? '1' : String(Math.max(0.55, 1 - 0.12 * age));
    }
  }

  function renderFullscreen() {
    const fs = $('fs'), it = state.fullscreen;
    fs.replaceChildren();
    document.body.classList.toggle('has-fs', !!it);
    if (!it) { fs.hidden = true; return; }
    fs.hidden = false;
    if (it.kind === 'text') {
      const d = document.createElement('div');
      d.innerHTML = it.html;
      fs.appendChild(d);
    } else if (it.kind === 'image') {
      const fig = document.createElement('figure');
      const img = document.createElement('img');
      img.alt = it.caption || '';
      img.src = it.src;
      fig.appendChild(img);
      if (it.caption) {
        const cap = document.createElement('figcaption');
        cap.textContent = it.caption;
        fig.appendChild(cap);
      }
      fs.appendChild(fig);
    } // kind "blank": empty overlay
  }

  function renderAll() {
    document.body.classList.remove('mode-idle', 'mode-captions', 'mode-blank');
    document.body.classList.add('mode-' + state.mode);
    const box = $('captions');
    box.replaceChildren(...state.scrollback.map(captionEl));
    fade();
    renderFullscreen();
  }

  function preload(urls) {
    (urls || []).forEach((u) => {
      if (preloaded.has(u)) return;
      preloaded.add(u);
      const im = new Image();
      im.decoding = 'async';
      if ('fetchPriority' in im) im.fetchPriority = 'low';
      im.src = u;
    });
  }

  // ---- event handling ----
  const on = {
    snapshot(d) {
      Object.assign(state, { mode: d.mode, scrollback: d.scrollback || [], fullscreen: d.fullscreen });
      applyTheme(d.theme || {});
      renderAll();
      preload(d.preload);
    },
    caption(d) {
      state.mode = 'captions';
      state.fullscreen = null;
      state.scrollback.push(d.item);
      document.body.classList.remove('mode-idle', 'mode-blank');
      document.body.classList.add('mode-captions');
      const box = $('captions');
      box.appendChild(captionEl(d.item));
      while (box.children.length > 200) box.removeChild(box.firstChild);
      fade();
      renderFullscreen();
    },
    clear() {
      state.mode = 'captions';
      state.fullscreen = null;
      state.scrollback = [];
      renderAll();
    },
    blank(d) {
      if (d.on) { state.mode = 'blank'; state.fullscreen = null; } else { state.mode = 'captions'; }
      renderAll();
    },
    fullscreen(d) {
      state.mode = state.mode === 'idle' ? 'captions' : state.mode;
      state.fullscreen = d.item;
      renderFullscreen();
      document.body.classList.remove('mode-idle');
      document.body.classList.add('mode-' + state.mode);
    },
    preload(d) { preload(d.urls); },
    reload() { location.reload(); },
    ping() {},
  };

  // ---- connection ----
  function setOnline(ok) {
    $('dot').classList.toggle('off', !ok);
    clearTimeout(offlineTimer);
    if (ok) { $('offline').hidden = true; return; }
    offlineTimer = setTimeout(() => { $('offline').hidden = false; }, 5000);
  }

  function connect() {
    clearTimeout(retryTimer);
    if (es) es.close();
    const q = params.get('lang') ? '?lang=' + encodeURIComponent(params.get('lang')) : '';
    // A rehearsal (started from Setup) streams from a private, authenticated endpoint.
    es = new EventSource(params.has('rehearsal') ? '/api/admin/rehearsal/events' : '/events' + q);
    lastEvent = Date.now();
    es.onopen = () => { backoff = 500; setOnline(true); };
    es.onerror = () => {
      setOnline(false);
      // Some browsers give up on the stream; reconnect ourselves, with jitter.
      es.close();
      retryTimer = setTimeout(connect, backoff + Math.random() * 300);
      backoff = Math.min(backoff * 2, 5000);
    };
    Object.keys(on).forEach((type) => {
      es.addEventListener(type, (e) => {
        lastEvent = Date.now();
        setOnline(true);
        try { on[type](JSON.parse(e.data)); } catch (err) { console.error(type, err); }
      });
    });
  }

  // Watchdog: the server pings every 15 s. Silence means a dead connection (sleeping device, NAT timeout).
  setInterval(() => { if (Date.now() - lastEvent > 40000) connect(); }, 5000);
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState !== 'visible') return;
    if (Date.now() - lastEvent > 20000) connect();
    keepAwake();
  });

  // Keep the screen on where the browser allows it.
  let lock = null;
  function tip() {
    if (projector || store.get('oc.tip') === '1') return;
    store.set('oc.tip', '1');
    $('tip').hidden = false;
    setTimeout(() => { $('tip').hidden = true; }, 12000);
  }
  async function keepAwake() {
    if (!('wakeLock' in navigator)) { tip(); return; }
    try {
      if ('wakeLock' in navigator && document.visibilityState === 'visible' && !lock) {
        lock = await navigator.wakeLock.request('screen');
        lock.addEventListener('release', () => { lock = null; });
      }
    } catch (e) { tip(); }
  }

  applyTheme({});
  connect();
  keepAwake();
})();
