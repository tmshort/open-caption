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

  // ---- rendering ----
  function applyTheme(t) {
    state.theme = t || {};
    const r = document.documentElement.style;
    if (t.bg) r.setProperty('--bg', t.bg);
    if (t.fg) r.setProperty('--fg', t.fg);
    if (t.accent) r.setProperty('--accent', t.accent);
    const local = parseInt(params.get('size'), 10);
    const pct = local > 0 ? local : (t.scalePct || 100);
    r.setProperty('--scale', String((pct / 100) * (projector ? 1.5 : 1)));
    $('idle').textContent = t.idleMessage || '';
  }

  function captionEl(c) {
    const p = document.createElement('p');
    p.className = 'cap' + (c.stage ? ' stage' : '');
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
      els[i].style.opacity = age === 0 ? '1' : String(Math.max(0.3, 1 - 0.17 * age));
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
    es = new EventSource('/events' + q);
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
  async function keepAwake() {
    try {
      if ('wakeLock' in navigator && document.visibilityState === 'visible' && !lock) {
        lock = await navigator.wakeLock.request('screen');
        lock.addEventListener('release', () => { lock = null; });
      }
    } catch (e) { /* not available (e.g. insecure context); the page still works */ }
  }

  applyTheme({});
  connect();
  keepAwake();
})();
