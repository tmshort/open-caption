// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Audience page: renders the live state pushed by the server. No interaction is needed.
(() => {
  'use strict';
  if (window.OC_UNSUPPORTED) return; // compat.js has already shown the "too old" message
  const $ = (id) => document.getElementById(id);
  const params = new URLSearchParams(location.search);
  const projector = params.get('display') === 'projector';
  if (projector) document.body.classList.add('projector');

  const state = { mode: 'idle', scrollback: [], scrollbackMax: 200, fullscreen: null, theme: {}, speakers: [], lang: 'en', languages: [] };

  // The page's own labels, in the languages we have translations for (others show English).
  const I18N = {
    en: { keepOn: "Screen stays on", keepOff: "Screen may sleep", reconnecting: 'Reconnecting…', tip: 'Tip: turn off Auto-Lock in your device settings so the screen stays on during the show.', textSize: 'Text size and colors', yourDisplay: 'Your display', smaller: 'Smaller text', bigger: 'Larger text', darkOnLight: 'Dark on light', lightOnDark: 'Light on dark', fontAsChosen: 'Font: as chosen', fontDyslexic: 'Dyslexia-friendly font', fontHyper: 'High-legibility font', language: 'Language', backToLive: 'Back to live', standby: 'Captions will begin shortly' },
    es: { keepOn: "La pantalla sigue encendida", keepOff: "La pantalla puede apagarse", reconnecting: 'Reconectando…', tip: 'Consejo: desactiva el bloqueo automático para que la pantalla siga encendida durante la función.', textSize: 'Tamaño del texto y colores', yourDisplay: 'Tu pantalla', smaller: 'Texto más pequeño', bigger: 'Texto más grande', darkOnLight: 'Oscuro sobre claro', lightOnDark: 'Claro sobre oscuro', fontAsChosen: 'Fuente: la del espectáculo', fontDyslexic: 'Fuente para dislexia', fontHyper: 'Fuente de alta legibilidad', language: 'Idioma', backToLive: 'Volver al directo', standby: 'Los subtítulos comenzarán en breve' },
    fr: { keepOn: "L'écran reste allumé", keepOff: "L'écran peut se mettre en veille", reconnecting: 'Reconnexion…', tip: "Astuce : désactivez le verrouillage automatique pour que l'écran reste allumé pendant le spectacle.", textSize: 'Taille du texte et couleurs', yourDisplay: 'Votre affichage', smaller: 'Texte plus petit', bigger: 'Texte plus grand', darkOnLight: 'Sombre sur clair', lightOnDark: 'Clair sur sombre', fontAsChosen: 'Police : celle du spectacle', fontDyslexic: 'Police pour la dyslexie', fontHyper: 'Police très lisible', language: 'Langue', backToLive: 'Retour au direct', standby: 'Les sous-titres vont bientôt commencer' },
    de: { keepOn: "Bildschirm bleibt an", keepOff: "Bildschirm darf sich ausschalten", reconnecting: 'Verbindung wird wiederhergestellt …', tip: 'Tipp: Schalten Sie die automatische Sperre aus, damit der Bildschirm während der Vorstellung an bleibt.', textSize: 'Textgröße und Farben', yourDisplay: 'Ihre Anzeige', smaller: 'Kleinerer Text', bigger: 'Größerer Text', darkOnLight: 'Dunkel auf hell', lightOnDark: 'Hell auf dunkel', fontAsChosen: 'Schrift: wie vorgegeben', fontDyslexic: 'Schrift für Legasthenie', fontHyper: 'Gut lesbare Schrift', language: 'Sprache', backToLive: 'Zurück zum Live-Text', standby: 'Die Untertitel beginnen gleich' },
    it: { keepOn: "Lo schermo resta acceso", keepOff: "Lo schermo può spegnersi", reconnecting: 'Riconnessione…', tip: 'Suggerimento: disattiva il blocco automatico per tenere lo schermo acceso durante lo spettacolo.', textSize: 'Dimensione del testo e colori', yourDisplay: 'Il tuo schermo', smaller: 'Testo più piccolo', bigger: 'Testo più grande', darkOnLight: 'Scuro su chiaro', lightOnDark: 'Chiaro su scuro', fontAsChosen: 'Carattere: quello dello spettacolo', fontDyslexic: 'Carattere per la dislessia', fontHyper: 'Carattere ad alta leggibilità', language: 'Lingua', backToLive: 'Torna al live', standby: 'I sottotitoli inizieranno a breve' },
    pt: { keepOn: "A tela fica acesa", keepOff: "A tela pode apagar", reconnecting: 'Reconectando…', tip: 'Dica: desative o bloqueio automático para a tela continuar acesa durante o espetáculo.', textSize: 'Tamanho do texto e cores', yourDisplay: 'Sua tela', smaller: 'Texto menor', bigger: 'Texto maior', darkOnLight: 'Escuro sobre claro', lightOnDark: 'Claro sobre escuro', fontAsChosen: 'Fonte: a do espetáculo', fontDyslexic: 'Fonte para dislexia', fontHyper: 'Fonte de alta legibilidade', language: 'Idioma', backToLive: 'Voltar ao vivo', standby: 'As legendas começarão em breve' },
  };
  const tr = (k) => (I18N[(state.lang || 'en').split('-')[0]] || I18N.en)[k] || I18N.en[k];
  function applyI18n() {
    document.documentElement.lang = state.lang || 'en';
    $('offline').textContent = tr('reconnecting');
    $('tip').textContent = tr('tip');
    $('gear').setAttribute('aria-label', tr('textSize'));
    $('ctl').setAttribute('aria-label', tr('yourDisplay'));
    $('smaller').setAttribute('aria-label', tr('smaller'));
    $('bigger').setAttribute('aria-label', tr('bigger'));
    const fs = $('fontsel');
    fs.setAttribute('aria-label', tr('fontAsChosen').split(':')[0]);
    fs.options[0].text = tr('fontAsChosen'); fs.options[1].text = tr('fontDyslexic'); fs.options[2].text = tr('fontHyper');
    $('langbtn').setAttribute('aria-label', tr('language'));
    paintLive();
    $('langmenu').setAttribute('aria-label', tr('language'));
    paintLocal();
  }

  // The language menu: choosing a language reloads the page at /?lang=xx, so that address can also be shared.
  function buildLangMenu() {
    const langs = state.languages || [];
    $('langbtn').hidden = langs.length < 2;
    const menu = $('langmenu');
    menu.replaceChildren(...langs.map((l) => {
      const b = document.createElement('button');
      b.type = 'button'; b.textContent = l.name; b.lang = l.code; b.dir = 'auto';
      b.setAttribute('role', 'menuitem');
      if (l.code === state.lang) b.setAttribute('aria-current', 'true');
      b.addEventListener('click', () => { const u = new URL(location.href); u.searchParams.set('lang', l.code); location.href = u.toString(); });
      return b;
    }));
    if (langs.length < 2) menu.hidden = true;
  }
  $('langbtn').addEventListener('click', () => { const open = $('langmenu').hidden; $('langmenu').hidden = !open; $('langbtn').setAttribute('aria-expanded', String(open)); });
  const preloaded = new Set();
  let pinned = true; // following the live captions (false while the reader has scrolled up)
  let unread = 0;    // lines that arrived while the reader was scrolled up
  let es = null, lastEvent = 0, backoff = 500, offlineTimer = null, retryTimer = null;

  const FONTS = { hyperlegible: '"Atkinson Hyperlegible", system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif', dyslexic: '"OpenDyslexic", system-ui, -apple-system, "Segoe UI", Roboto, Arial, sans-serif', system: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif', serif: 'Georgia, "Times New Roman", serif', mono: 'ui-monospace, Menlo, Consolas, monospace' };

  // ---- this device's own preferences (optional; never required) ----
  const store = {
    get(k) { try { return localStorage.getItem(k); } catch (e) { return null; } },
    set(k, v) { try { localStorage.setItem(k, v); } catch (e) { /* storage may be unavailable */ } },
  };
  let localStep = parseInt(store.get('oc.step'), 10) || 0; // -3..+6 steps of 15%
  let lightScheme = store.get('oc.light') === '1';
  let keepOn = store.get('oc.awake') !== '0'; // keep the screen on (on by default); a device can turn it off
  // This device's own choice of font ('' = the one the show chose).
  let localFont = ['dyslexic', 'hyperlegible'].includes(store.get('oc.font')) ? store.get('oc.font') : '';
  function paintLocal() {
    const t = state.theme || {}, r = document.documentElement.style;
    const urlPct = parseInt(params.get('size'), 10);
    const pct = (urlPct > 0 ? urlPct : (t.scalePct || 100)) * (1 + 0.15 * localStep);
    r.setProperty('--scale', String(Math.max(0.4, pct / 100) * (projector ? 1.5 : 1)));
    const fontKey = localFont || t.font;
    r.setProperty('--font', FONTS[fontKey] || FONTS.system);
    document.body.classList.toggle('dys', fontKey === 'dyslexic');
    $('fontsel').value = localFont;
    if (lightScheme) { r.setProperty('--bg', '#ffffff'); r.setProperty('--fg', '#000000'); r.setProperty('--accent', '#000000'); }
    else {
      r.setProperty('--bg', t.bg || '#000000'); r.setProperty('--fg', t.fg || '#ffffff');
      r.setProperty('--accent', t.accent || t.fg || '#ffffff');
    }
    $('scheme').setAttribute('aria-pressed', String(lightScheme));
    $('scheme').textContent = lightScheme ? tr('lightOnDark') : tr('darkOnLight');
    $('awake').setAttribute('aria-pressed', String(keepOn));
    $('awake').textContent = (keepOn ? '☀ ' : '☾ ') + tr(keepOn ? 'keepOn' : 'keepOff');
  }
  $('gear').addEventListener('click', () => { const open = $('ctl').hidden; $('ctl').hidden = !open; $('gear').setAttribute('aria-expanded', String(open)); });
  $('bigger').addEventListener('click', () => { localStep = Math.min(6, localStep + 1); store.set('oc.step', String(localStep)); paintLocal(); });
  $('smaller').addEventListener('click', () => { localStep = Math.max(-3, localStep - 1); store.set('oc.step', String(localStep)); paintLocal(); });
  $('fontsel').addEventListener('change', () => { localFont = $('fontsel').value; store.set('oc.font', localFont); paintLocal(); });
  $('awake').addEventListener('click', () => {
    keepOn = !keepOn; store.set('oc.awake', keepOn ? '1' : '0'); paintLocal();
    if (keepOn) keepAwake();
    else { if (lock) { lock.release().catch(() => {}); lock = null; } if (video) video.pause(); setAwake('off'); }
  });
  $('scheme').addEventListener('click', () => { lightScheme = !lightScheme; store.set('oc.light', lightScheme ? '1' : '0'); paintLocal(); });

  // ---- rendering ----
  function applyTheme(t) {
    state.theme = t || {};
    paintLocal();
    $('idle').textContent = t.idleMessage || '';
  }

  // With the "different color for each character" style, a speaker's name (and the bar beside
  // a simultaneous voice) takes that character's color.
  function tint(el, name) {
    if (state.theme.speakers !== 'color' || !window.OCSpeakers || lightScheme) return;
    el.style.color = OCSpeakers.forName(name, state.speakers, state.theme.bg || '#000000');
  }

  function voiceEl(v) {
    const d = document.createElement('div');
    d.dir = 'auto'; // right-to-left languages line up on the right; the speaker label has its own direction
    d.className = 'part' + (v.sung ? ' sung' : '') + (v.stage ? ' stage' : '');
    if (v.speaker && state.theme.speakers !== 'hidden') {
      const s = document.createElement('span');
      s.className = 'spk';
      s.dir = 'auto';
      s.textContent = v.speaker;
      tint(s, v.speaker);
      if (s.style.color) d.style.borderColor = s.style.color;
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
    p.dir = 'auto';
    p.className = 'cap' + (c.stage ? ' stage' : '') + (c.sung ? ' sung' : '');
    p.dataset.id = c.id;
    if (c.speaker && state.theme.speakers !== 'hidden') {
      const s = document.createElement('span');
      s.className = 'spk';
      s.dir = 'auto';
      s.textContent = c.speaker;
      tint(s, c.speaker);
      p.appendChild(s);
    }
    const body = document.createElement('span');
    body.innerHTML = c.html; // generated and escaped by the server
    p.appendChild(body);
    return p;
  }

  // ---- re-reading: scroll up freely; new lines wait behind a button ----
  const nearBottom = (box) => box.scrollHeight - box.scrollTop - box.clientHeight < 48;
  function paintLive() {
    const b = $('live');
    b.hidden = pinned;
    b.textContent = '↓ ' + tr('backToLive') + (unread ? ' (' + unread + ')' : '');
    document.body.classList.toggle('reading', !pinned);
  }
  // New lines glide into place (about a third of a second) instead of jumping. The target is re-read on every
  // frame so a line that arrives mid-glide extends it, and the scroll handler ignores the glide (it would
  // otherwise think the reader had scrolled up). Touching or wheeling the page stops it at once.
  let glide = 0;
  const reduceMotion = () => !!(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches);
  function stopGlide() { if (glide) { cancelAnimationFrame(glide); glide = 0; } }
  function toBottom(smooth) {
    const box = $('captions');
    stopGlide();
    const from = box.scrollTop;
    if (!smooth || reduceMotion() || box.scrollHeight - box.clientHeight - from < 2) { box.scrollTop = box.scrollHeight; return; }
    const t0 = performance.now(), ms = 320;
    const step = (now) => {
      const p = Math.min(1, (now - t0) / ms), e = 1 - Math.pow(1 - p, 3); // ease out
      const end = box.scrollHeight - box.clientHeight;
      box.scrollTop = from + (end - from) * e;
      if (p < 1) glide = requestAnimationFrame(step); else { glide = 0; box.scrollTop = box.scrollHeight; }
    };
    glide = requestAnimationFrame(step);
  }
  ['wheel', 'touchstart', 'keydown'].forEach((ev) => $('captions').addEventListener(ev, stopGlide, { passive: true }));
  // A new line comes up from below the bottom edge, and everything above it moves up together, as if the whole
  // column were scrolling. With a screenful of text the glide above does that. When the lines do not yet fill the
  // screen they sit at the bottom and a new line would simply appear, with the older ones jumping up by its
  // height; so put every line back where it was, with the new one just below the edge (the CSS clips the
  // column there), and release them all (FLIP: measure, move back, release).
  function slideIn(d, fresh) {
    const box = $('captions');
    if (box.scrollHeight - box.clientHeight > 1) return; // a full screen scrolls instead (toBottom)
    const kids = Array.prototype.slice.call(box.children);
    const edge = box.getBoundingClientRect().bottom - parseFloat(getComputedStyle(box).paddingBottom);
    const start = (k) => (k === fresh ? Math.max(d, edge - k.getBoundingClientRect().top) : d); // the new line starts below the edge
    const from = kids.map(start);
    if (!from.some((v) => v > 1)) return;
    box.classList.add('sliding'); // a moved-down line must not make the column scrollable for a moment
    kids.forEach((k, i) => { k.style.transition = 'none'; k.style.transform = 'translateY(' + from[i] + 'px)'; });
    void box.offsetHeight; // commit the starting positions
    kids.forEach((k) => { k.style.transition = 'transform .32s cubic-bezier(.215,.61,.355,1)'; k.style.transform = ''; });
    setTimeout(() => { kids.forEach((k) => { k.style.transition = ''; }); box.classList.remove('sliding'); }, 400);
  }
  function goLive() { pinned = true; unread = 0; toBottom(true); paintLive(); }
  $('live').addEventListener('click', goLive);
  $('captions').addEventListener('scroll', () => {
    if (projector || glide) return;
    const wasPinned = pinned;
    pinned = nearBottom($('captions'));
    if (pinned) unread = 0;
    if (pinned !== wasPinned || !pinned) paintLive();
  }, { passive: true });

  function fade() {
    // Older lines dim with age (newest = full brightness).
    const els = $('captions').children;
    for (let i = 0; i < els.length; i++) {
      const age = els.length - 1 - i;
      // Older lines fade, but never below a level that is still readable.
      els[i].style.opacity = age === 0 ? '1' : String(Math.max(0.55, 1 - 0.12 * age));
    }
  }

  // ---- countdown ("Back in 15 min") ----
  // The server sends the moment the countdown ends and, with the snapshot, its own clock; a phone whose clock
  // is minutes out still shows the right time left. One timer ticks while a countdown is on screen.
  let clockSkew = 0, timerTick = 0; // server time minus this device's time, in ms
  // Whole minutes, rounded up: nobody needs the seconds, and a figure that changes once a minute is calm to read.
  // The unit comes from the browser in the device's language ("15 min", "1 h 5 min"); an old browser gets "15 min".
  function unit(n, u) {
    try { return new Intl.NumberFormat(state.lang || document.documentElement.lang || 'en', { style: 'unit', unit: u, unitDisplay: 'short' }).format(n); }
    catch (e) { return n + (u === 'hour' ? ' h' : ' min'); }
  }
  function fmtLeft(ms) {
    const total = Math.max(0, Math.ceil(ms / 60000)), h = Math.floor(total / 60), m = total % 60;
    return h ? unit(h, 'hour') + (m ? ' ' + unit(m, 'minute') : '') : unit(total, 'minute');
  }
  function paintTimers() {
    const els = document.querySelectorAll('#fs [data-timer]');
    const it = state.fullscreen;
    if (!els.length || !it || !it.deadline) { clearInterval(timerTick); timerTick = 0; return; }
    const text = fmtLeft(it.deadline - (Date.now() + clockSkew));
    for (let i = 0; i < els.length; i++) if (els[i].textContent !== text) els[i].textContent = text;
  }
  function startTimers() {
    clearInterval(timerTick); timerTick = 0;
    if (!document.querySelector('#fs [data-timer]')) return;
    paintTimers();
    timerTick = setInterval(paintTimers, 250);
  }

  function renderFullscreen() {
    const fs = $('fs'), it = state.fullscreen;
    fs.replaceChildren();
    fs.classList.remove('bleed');
    document.body.classList.toggle('has-fs', !!it);
    if (!it) { fs.hidden = true; return; }
    fs.hidden = false;
    if (it.kind === 'text') {
      const d = document.createElement('div');
      d.dir = 'auto';
      d.innerHTML = it.html;
      fs.appendChild(d);
      startTimers();
    } else if (it.kind === 'image') {
      fs.classList.toggle('bleed', !it.caption); // a picture with no words fills the whole screen
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

  // The venue's standby screen: its logo and message, or the built-in words in the viewer's language.
  function renderStandby() {
    const s = state.standby || {}, img = $('standbylogo');
    const src = s.logo || '';
    if (img.getAttribute('src') !== src) { if (src) img.src = src; else img.removeAttribute('src'); }
    img.hidden = !src; img.className = s.default ? 'default' : '';
    $('standbymsg').textContent = s.message || tr('standby');
  }

  function renderAll() {
    document.body.classList.remove('mode-idle', 'mode-captions', 'mode-blank', 'mode-standby');
    document.body.classList.add('mode-' + state.mode);
    const box = $('captions');
    renderStandby();
    const top = box.scrollTop;
    box.replaceChildren(...state.scrollback.map(captionEl));
    fade();
    renderFullscreen();
    if (pinned) toBottom(); else box.scrollTop = top; // a snapshot must not move a reader who is re-reading
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
      if (d.ts) clockSkew = d.ts - Date.now();
      Object.assign(state, { mode: d.mode, scrollback: d.scrollback || [], scrollbackMax: d.scrollbackMax || 200, fullscreen: d.fullscreen, speakers: d.speakers || [], lang: d.lang || 'en', languages: d.languages || [], standby: d.standby || null });
      applyTheme(d.theme || {});
      applyI18n();
      buildLangMenu();
      renderAll();
      preload(d.preload);
    },
    caption(d) {
      state.mode = 'captions';
      state.fullscreen = null;
      state.scrollback.push(d.item);
      if (state.scrollback.length > state.scrollbackMax) state.scrollback.splice(0, state.scrollback.length - state.scrollbackMax);
      document.body.classList.remove('mode-idle', 'mode-blank', 'mode-standby');
      document.body.classList.add('mode-captions');
      const box = $('captions');
      const prev = box.lastElementChild, wasAt = prev ? prev.getBoundingClientRect().top : 0;
      box.appendChild(captionEl(d.item));
      while (box.children.length > state.scrollbackMax) {
        const before = box.scrollHeight;
        box.removeChild(box.firstChild);
        if (!pinned) box.scrollTop -= before - box.scrollHeight; // keep the reader's place when old lines are trimmed
      }
      fade();
      renderFullscreen();
      const moved = prev && prev.parentNode === box && pinned && !reduceMotion() ? wasAt - prev.getBoundingClientRect().top : 0;
      if (pinned) { toBottom(true); if (!reduceMotion()) slideIn(moved, box.lastElementChild); } else { unread++; paintLive(); }
    },
    clear() {
      state.mode = 'captions';
      state.fullscreen = null;
      state.scrollback = [];
      pinned = true; unread = 0; // an empty screen has nothing to scroll back to
      renderAll();
      paintLive();
    },
    blank(d) {
      if (d.on) { state.mode = 'blank'; state.fullscreen = null; } else { state.mode = 'captions'; }
      renderAll();
    },
    fullscreen(d) {
      state.mode = state.mode === 'idle' ? 'captions' : state.mode;
      state.fullscreen = d.item;
      renderFullscreen();
      document.body.classList.remove('mode-idle', 'mode-standby');
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
    const path = params.has('rehearsal') ? '/api/admin/rehearsal/events' : '/events';
    es = new EventSource(path + q + (projector ? (q ? '&' : '?') + 'display=projector' : ''));
    lastEvent = Date.now();
    es.onopen = () => { backoff = 500; setOnline(true); };
    es.onerror = () => {
      setOnline(false);
      // Some browsers give up on the stream; reconnect ourselves, with jitter.
      if (es) es.close();
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

  // Do not hold a connection once the page is left (a page kept in the back/forward cache would keep its stream open).
  window.addEventListener('pagehide', () => { clearTimeout(retryTimer); if (es) { es.close(); es = null; } });
  window.addEventListener('pageshow', (e) => { if (e.persisted) connect(); });

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
  // Where the Screen Wake Lock API is missing (every browser on plain http:// except localhost, and older
  // iPhones) a tiny silent, invisible video looping on the page keeps the screen on instead. Browsers
  // only allow it to start without a tap in some cases, so the first touch starts it if needed.
  let video = null, videoWaiting = false;
  function setAwake(how) { document.documentElement.dataset.awake = how; }
  function startVideo() {
    if (!video) {
      video = document.createElement('video');
      video.muted = true; video.defaultMuted = true; video.loop = true; video.playsInline = true;
      video.setAttribute('muted', ''); video.setAttribute('playsinline', ''); video.setAttribute('aria-hidden', 'true'); video.tabIndex = -1;
      video.setAttribute('title', 'Keeps the screen on');
      video.style.cssText = 'position:fixed;right:0;bottom:0;width:2px;height:2px;opacity:0.01;pointer-events:none;border:0';
      for (const [file, type] of [['nosleep.mp4', 'video/mp4'], ['nosleep.webm', 'video/webm']]) {
        if (video.canPlayType(type)) { const src = document.createElement('source'); src.src = '/static/audience/' + file; src.type = type; video.append(src); }
      }
      // Some browsers drop a very short looping clip; start it again if that happens.
      video.addEventListener('ended', () => { video.currentTime = 0; video.play().catch(() => {}); });
      video.addEventListener('playing', () => setAwake('video'));
      document.body.append(video);
    }
    const p = video.play();
    if (p && p.then) p.then(() => setAwake('video'), (e) => {
      // Stopped on purpose (the screen wake lock took over and paused the video): nothing to do.
      if (e && e.name === 'AbortError') return;
      // Not allowed yet: try again at the first tap, and show the Auto-Lock tip meanwhile.
      setAwake('none'); tip();
      if (videoWaiting) return;
      videoWaiting = true;
      const again = () => { videoWaiting = false; ['touchend', 'click', 'keydown'].forEach((ev) => document.removeEventListener(ev, again, true)); video.play().catch(() => {}); };
      ['touchend', 'click', 'keydown'].forEach((ev) => document.addEventListener(ev, again, true));
    });
  }
  async function keepAwake() {
    if (!keepOn) { setAwake('off'); return; }
    if ('wakeLock' in navigator) {
      try {
        if (document.visibilityState === 'visible' && !lock) {
          lock = await navigator.wakeLock.request('screen');
          lock.addEventListener('release', () => { lock = null; });
          setAwake('wakelock');
          if (video) video.pause();
        }
        return;
      } catch (e) { /* refused (for example low battery): use the video instead */ }
    }
    if (video && !video.paused) return;
    startVideo();
  }

  applyTheme({});
  connect();
  keepAwake();
})();
