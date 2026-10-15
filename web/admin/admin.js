// Setup page: a six-step wizard (Show, Script, Images, Look, Test, Go Live).
(() => {
  'use strict';
  const $ = (id) => document.getElementById(id);
  const A = { shows: [], activeId: '', cur: null, step: 'show', text: '', saved: '', analysis: null, net: null, live: false, audience: 0, rehearsing: false };
  const STEPS = [['show', 'Show'], ['script', 'Script'], ['images', 'Images'], ['look', 'Look'], ['test', 'Test'], ['golive', 'Go Live']];
  const ICON = { caption: '💬', clear: '🧹', blank: '⬛', image: '🖼', fullscreen: '⛶', loop: '🔁', section: '📑', note: '📝' };

  // ---- helpers ----
  function h(tag, attrs, ...kids) {
    const e = document.createElement(tag);
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
  function toast(msg) { const t = $('toast'); t.textContent = msg; t.classList.add('show'); clearTimeout(toastTimer); toastTimer = setTimeout(() => t.classList.remove('show'), 3000); }
  function fail(r, fallback) { toast((r.data && r.data.error) || fallback || 'Something went wrong'); }

  // A reusable dialog. body: node; resolves true on OK, false on cancel.
  function ask(title, body, okLabel, opts) {
    return new Promise((resolve) => {
      const d = $('dlg');
      $('dlgtitle').textContent = title;
      $('dlgbody').replaceChildren(body);
      $('dlgok').textContent = okLabel || 'OK';
      $('dlgok').classList.toggle('danger', !!(opts && opts.danger));
      $('dlgcancel').hidden = !!(opts && opts.noCancel);
      $('dlgcancel').onclick = () => d.close('cancel');
      d.returnValue = '';
      d.onclose = () => resolve(d.returnValue === 'ok');
      d.showModal();
    });
  }
  const confirmBox = (title, text, ok, danger) => ask(title, h('p', {}, text), ok, { danger });
  const fmtSize = (n) => (n >= 1 << 20 ? (n / (1 << 20)).toFixed(1) + ' MB' : Math.max(1, Math.round(n / 1024)) + ' KB');
  const curShow = () => A.shows.find((s) => s.id === A.cur) || null;

  // ---- help drawer (topic follows the current step) ----
  const HELP_TOPIC = { show: 'admin-getting-started', script: 'admin-writing-a-script', images: 'admin-images', look: 'admin-look-and-feel', test: 'admin-getting-started', golive: 'admin-network' };
  function openHelp(topic) {
    $('help').hidden = false;
    $('help').querySelector('iframe').src = '/help?role=admin&embed=1&topic=' + (topic || HELP_TOPIC[A.step] || 'admin-getting-started');
  }
  function closeHelp() { $('help').hidden = true; $('help').querySelector('iframe').src = 'about:blank'; }
  $('helpbtn').addEventListener('click', () => ($('help').hidden ? openHelp() : closeHelp()));
  $('helpclose').addEventListener('click', closeHelp);
  window.addEventListener('message', (e) => { if (e.origin === location.origin && e.data && e.data.type === 'help-close') closeHelp(); });
  document.addEventListener('keydown', (e) => { if (e.key === 'F1') { e.preventDefault(); openHelp(); } });

  // ---- login ----
  function showLogin() { $('app').hidden = true; $('login').hidden = false; $('pw').focus(); }
  $('login').addEventListener('submit', async (e) => {
    e.preventDefault();
    const r = await fetch('/api/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ role: 'admin', password: $('pw').value }) });
    if (r.ok) { $('loginerr').textContent = ''; $('pw').value = ''; boot(); return; }
    $('loginerr').textContent = r.status === 429 ? 'Too many attempts. Wait a minute and try again.' : 'That password was not accepted.';
  });
  $('logout').addEventListener('click', async () => { await fetch('/api/logout', { method: 'POST' }); showLogin(); });

  // ---- boot ----
  async function boot() {
    const ping = await fetch('/api/admin/ping');
    if (!ping.ok) { showLogin(); return; }
    $('login').hidden = true; $('app').hidden = false;
    await loadShows();
    try { A.cur = localStorage.getItem('oc.cur'); } catch (e) { /* storage unavailable */ }
    if (!curShow()) A.cur = A.activeId && curShowById(A.activeId) ? A.activeId : (A.shows[0] ? A.shows[0].id : null);
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
  function setCur(id) { A.cur = id; A.rehearsing = false; A.text = A.saved = ''; A.analysis = null; try { localStorage.setItem('oc.cur', id || ''); } catch (e) { /* ignore */ } }

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
  const lookKey = () => 'oc.look.' + A.cur;
  const lookSaved = () => { try { return localStorage.getItem(lookKey()) === '1'; } catch (e) { return false; } };
  const markLookSaved = () => { try { localStorage.setItem(lookKey(), '1'); } catch (e) { /* ignore */ } refreshNav(); };

  function done(step) {
    switch (step) {
      case 'look': return lookSaved();
      case 'test': return A.rehearsing;
      case 'golive': return !!A.cur && A.activeId === A.cur;
      case 'show': return !!A.cur;
      case 'script': return !!(A.analysis && A.analysis.cues.length && !A.analysis.diagnostics.some((d) => d.severity === 'error'));
      case 'images': return !!(A.analysis && !(A.analysis.missing || []).length && (A.analysis.assets || []).length);
      default: return false;
    }
  }

  function refreshNav() {
    const nav = $('steps'); nav.replaceChildren();
    STEPS.forEach(([id, label], i) => {
      nav.append(h('button', { class: (A.step === id ? 'on ' : '') + (done(id) ? 'done' : ''), onclick: () => go(id), disabled: id !== 'show' && !A.cur, 'aria-current': A.step === id ? 'step' : false },
        h('span', { class: 'n' }, h('span', {}, String(i + 1))), label));
    });
  }

  // Ticks for Script and Images depend on the script analysis, so load it in the background when not yet known.
  async function warmAnalysis() {
    const id = A.cur;
    if (!id || A.analysis) return;
    const r = await api('GET', '/api/admin/shows/' + id + '/script');
    if (r.ok && A.cur === id && !A.analysis) { A.analysis = r.data.analysis; refreshNav(); }
  }

  async function render() {
    const s = curShow();
    $('working').textContent = s ? s.name : 'No show selected';
    $('working').hidden = false;
    refreshNav();
    warmAnalysis();
    const panel = $('panel'); panel.replaceChildren();
    document.title = 'Setup — ' + (STEPS.find((x) => x[0] === A.step) || ['', ''])[1];
    try { await ({ show: stepShow, script: stepScript, images: stepImages, look: stepLook, test: stepTest, golive: stepGoLive })[A.step](panel); } catch (e) { if (e.message !== 'login') console.error(e); }
  }

  // ---- step 1: show ----
  async function stepShow(p) {
    await loadShows();
    p.append(h('h2', {}, 'Choose a show'));
    if (!A.shows.length) {
      p.append(h('div', { class: 'card empty' },
        h('p', { class: 'big' }, 'Welcome! Start by creating a show, or look around with the sample.'),
        h('div', { class: 'row' }, newShowForm(), h('button', { onclick: makeSample }, 'Create a sample show'))));
    } else {
      A.shows.forEach((s) => p.append(h('div', { class: 'card show' },
        h('div', { class: 'name' }, s.name, ' ', s.active ? h('span', { class: 'badge active' }, 'ACTIVE') : null, ' ', s.live ? h('span', { class: 'badge live' }, 'LIVE') : null,
          h('div', { class: 'hint' }, 'Updated ' + new Date(s.updated).toLocaleString())),
        h('button', { class: A.cur === s.id ? '' : 'primary', onclick: () => { setCur(s.id); go('script'); } }, A.cur === s.id ? 'Continue with this show' : 'Work on this show'),
        h('a', { href: '/api/admin/shows/' + s.id + '/export', download: s.id + '.zip' }, 'Export'),
        h('button', { class: 'danger quiet', onclick: () => deleteShow(s) }, 'Delete…'))));
      p.append(h('h3', {}, 'Add another show'), h('div', { class: 'card row' }, newShowForm(), h('button', { onclick: makeSample }, 'Sample show'),
        h('label', { class: 'row' }, 'Import a show (.zip):', h('input', { type: 'file', accept: '.zip', onchange: importShow }))));
    }
  }
  function newShowForm() {
    const inp = h('input', { type: 'text', placeholder: 'Name of the show', maxlength: '100', 'aria-label': 'Show name' });
    const create = async () => {
      if (!inp.value.trim()) { inp.focus(); return; }
      const r = await api('POST', '/api/admin/shows', { name: inp.value });
      if (!r.ok) return fail(r);
      await loadShows(); setCur(r.data.id); go('script');
    };
    inp.addEventListener('keydown', (e) => { if (e.key === 'Enter') create(); });
    return h('span', { class: 'row' }, inp, h('button', { class: 'primary', onclick: create }, 'Create show'));
  }
  async function makeSample() {
    const r = await api('POST', '/api/admin/shows', { sample: true });
    if (!r.ok) return fail(r);
    await loadShows(); setCur(r.data.id); toast('Sample show created'); go('script');
  }
  async function importShow(e) {
    const f = e.target.files[0]; if (!f) return;
    const fd = new FormData(); fd.append('file', f);
    const r = await api('POST', '/api/admin/shows/import', undefined, fd);
    if (!r.ok) return fail(r, 'That file could not be imported');
    await loadShows(); setCur(r.data.show.id);
    toast(r.data.skippedFiles ? 'Imported (' + r.data.skippedFiles + ' file(s) skipped)' : 'Imported'); go('script');
  }
  async function deleteShow(s) {
    const ok = await confirmBox('Delete “' + s.name + '”?', 'Its script and images will be permanently deleted. Export it first if you may want it back.', 'Delete', true);
    if (!ok) return;
    const r = await api('DELETE', '/api/admin/shows/' + s.id);
    if (!r.ok) return fail(r);
    if (A.cur === s.id) setCur(null);
    await loadShows(); if (!A.cur && A.shows[0]) setCur(A.shows[0].id);
    render();
  }

  // ---- step 2: script ----
  const SNIPPETS = {
    clear: '[clear]\n', blank: '[blank]\n', fullscreen: '[fullscreen: INTERMISSION]\n', section: '[section: Act 1, Scene 1]\n',
    note: '[note: reminder for the operator]\n', comment: '<!-- a comment nobody sees -->\n',
    loop: '[loop: 10s]\n[fullscreen: Welcome!]\nPlease silence your phones\n[/loop]\n',
    together: '[together]\nANNA: ~ first voice ~\nKRIS: ~ second voice sung at the same time ~\n[/together]\n',
    song: '[song: Title of the song]\nFirst line of the lyrics\nSecond line of the lyrics\n[/song]\n',
  };
  async function stepScript(p) {
    const show = curShow(); if (!show) return go('show');
    if (!A.analysis || A.text === '') {
      const r = await api('GET', '/api/admin/shows/' + show.id + '/script');
      if (!r.ok) return fail(r);
      A.text = A.saved = r.data.text; A.analysis = r.data.analysis;
    }
    const ta = h('textarea', { spellcheck: 'false', 'aria-label': 'Script', wrap: 'off', placeholder: 'Paste your script here, or drop a .txt / .md file onto this box.\n\nEach line becomes one caption.' });
    ta.value = A.text;
    const gutter = h('pre', { class: 'gutter', 'aria-hidden': 'true' });
    const status = h('span', { class: 'hint' });
    const diagList = h('ul', { class: 'diags' }), cueList = h('ul', { class: 'cues' }), diagTitle = h('h3', {}, 'Problems and notes');
    const editor = h('div', { class: 'editor' }, gutter, ta);
    const saveBtn = h('button', { class: 'primary', onclick: save }, 'Save');

    function lines() { const n = ta.value.split('\n').length; gutter.textContent = Array.from({ length: n }, (_, i) => i + 1).join('\n'); }
    function setStatus() {
      const dirty = ta.value !== A.saved;
      status.textContent = dirty ? 'Unsaved changes' : 'All changes saved';
      status.className = dirty ? 'status-warn' : 'status-ok';
      saveBtn.disabled = !dirty;
    }
    function showAnalysis() {
      const a = A.analysis;
      const shown = a.diagnostics.filter((d) => d.line > 0);
      const summary = a.diagnostics.find((d) => d.line === 0);
      diagTitle.textContent = shown.length ? 'Problems and notes (' + shown.length + ')' : 'No problems found ✓';
      diagList.replaceChildren(...shown.map((d) => h('li', { class: d.severity, onclick: () => jump(d.line) }, h('span', { class: 'ln' }, 'Line ' + d.line), d.message)),
        ...(summary ? [h('li', { class: 'info' }, summary.message)] : []));
      cueList.replaceChildren(...a.cues.map((c) => h('li', { class: c.kind === 'section' || c.kind === 'note' ? 'op' : '', title: 'Line ' + c.line },
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
        const r = await api('POST', '/api/admin/shows/' + show.id + '/script/validate', { text: ta.value });
        if (r.ok && A.step === 'script') { A.analysis = r.data.analysis; showAnalysis(); }
      }, 400);
    }
    async function save(confirmName) {
      const body = { text: ta.value }; if (typeof confirmName === 'string') body.confirm = confirmName;
      const r = await api('PUT', '/api/admin/shows/' + show.id + '/script', body);
      if (r.status === 409 && r.data.needConfirm) {
        const inp = h('input', { type: 'text', 'aria-label': 'Show name', placeholder: r.data.showName });
        const ok = await ask('This show is live', h('div', {}, h('p', {}, 'Changing the script now affects what the operator sees. The current position is kept by line number. To continue, type the show’s name:'), h('p', {}, h('b', {}, r.data.showName)), inp), 'Save change');
        if (ok) return save(inp.value);
        return;
      }
      if (!r.ok) return fail(r);
      A.saved = ta.value; A.analysis = r.data.analysis; setStatus(); showAnalysis();
      toast(r.data.countChanged ? 'Saved. The number of lines changed, so check the operator’s position.' : 'Saved');
    }
    function insert(text) {
      const s = ta.selectionStart, e = ta.selectionEnd, atLineStart = s === 0 || ta.value[s - 1] === '\n';
      const t = (atLineStart ? '' : '\n') + text;
      ta.setRangeText(t, s, e, 'end'); ta.focus(); changed();
    }
    function insertInline(text) { ta.setRangeText(text, ta.selectionStart, ta.selectionEnd, 'end'); ta.focus(); changed(); }
    function loadText(text) { ta.value = text; changed(); }
    async function readFile(f) { if (f) loadText(await f.text()); }

    ta.addEventListener('input', changed);
    ta.addEventListener('scroll', () => { gutter.scrollTop = ta.scrollTop; });
    ta.addEventListener('keydown', (e) => { if ((e.metaKey || e.ctrlKey) && e.key === 's') { e.preventDefault(); if (!saveBtn.disabled) save(); } });
    editor.addEventListener('dragover', (e) => { e.preventDefault(); editor.classList.add('drop'); });
    editor.addEventListener('dragleave', () => editor.classList.remove('drop'));
    editor.addEventListener('drop', (e) => { e.preventDefault(); editor.classList.remove('drop'); readFile(e.dataTransfer.files[0]); });

    const tools = h('div', { class: 'tools' },
      ...[['Clear', 'clear'], ['Blank', 'blank'], ['Fullscreen text', 'fullscreen'], ['Section', 'section'], ['Operator note', 'note'], ['Comment', 'comment'], ['Loop', 'loop']].map(([l, k]) => h('button', { onclick: () => insert(SNIPPETS[k]) }, l)),
      h('button', { onclick: pickImage }, 'Image…'),
      h('span', { class: 'tools-sep', 'aria-hidden': 'true' }, '|'),
      h('button', { onclick: () => insert(SNIPPETS.song), title: 'A song: every line becomes a sung lyric' }, '🎵 Song'),
      h('button', { onclick: () => insert(SNIPPETS.together), title: 'Two or more singers at the same time' }, 'Together'),
      h('button', { onclick: () => insertInline('♪ '), title: 'Insert a music note', 'aria-label': 'Insert a music note' }, '♪'),
      h('button', { onclick: () => insertInline('♫ '), title: 'Insert beamed music notes', 'aria-label': 'Insert beamed music notes' }, '♫'));
    async function pickImage() {
      const names = (A.analysis.assets || []).map((a) => a.name);
      if (!names.length) { toast('Upload images in the Images step first'); return; }
      const sel = h('select', {}, ...names.map((n) => h('option', { value: n }, n)));
      if (await ask('Insert an image', h('label', { class: 'field' }, h('span', {}, 'Which image?'), sel), 'Insert')) insert('[image: ' + sel.value + ']\n');
    }
    const fileInput = h('input', { type: 'file', accept: '.txt,.md,text/plain,text/markdown', onchange: (e) => readFile(e.target.files[0]) });
    p.append(
      h('h2', {}, 'Script'),
      h('p', { class: 'hint' }, 'Each line is one caption; one press of the spacebar sends one line. You can paste, type, or drop a .txt or .md file. Click a problem to jump to its line.'),
      h('div', { class: 'row' }, h('label', { class: 'row' }, 'Open a file:', fileInput), h('button', { onclick: () => openHelp('admin-writing-a-script') }, 'Script help'), h('span', { class: 'grow' }), status, saveBtn),
      tools,
      h('div', { class: 'editor-grid' }, editor, h('div', {}, h('div', { class: 'card' }, diagTitle, diagList), h('div', { class: 'card' }, h('h3', {}, 'What the operator will step through'), cueList))),
      cheatSheet());
    lines(); setStatus(); showAnalysis();
    window.onbeforeunload = () => (ta.value !== A.saved ? 'You have unsaved changes.' : undefined);
  }

  function cheatSheet() {
    const rows = [
      ['A normal line', 'One caption. Long lines wrap on the screen.'], ['NAME: text', 'Speaker label (NAME in capitals)'], ['**bold**   *italic*', 'Bold and italic'],
      ['(thunder rumbles)', 'Shown as a sound or stage direction'], ['line one \\', 'A backslash at the end joins the next line into the same caption'], ['[[ … ]]', 'Several lines in one caption'],
      ['<!-- comment -->', 'A comment: never shown to anyone'], ['[note: text]', 'A reminder shown to the operator only'], ['[section: Act 2]', 'A bookmark for the operator'],
      ['[clear]  [blank]', 'Clear the captions / blank the screen'], ['[image: file.png | caption]', 'Fullscreen image (also  ![caption](file.png) )'], ['[fullscreen: text]', 'Fullscreen text'],
      ['~ lyric ~', 'A sung line, shown between ♪ notes'], ['[song: Title] … [/song]', 'A whole song: every line inside is a sung lyric'], ['[together] … [/together]', 'Voices shown at the same time (a duet or counterpoint), one press of Space. Also  A: ~ one ~ || B: ~ two ~  on a single line'], ['♪  ♫', 'Music notes can be typed or pasted anywhere; the ♪ and ♫ buttons insert them'],
      ['[loop: 10s] … [/loop]', 'Repeating items for pre-show or intermission. Add  | 15s  to an item to change its time'], ['# …', 'Reserved for headings; the line is ignored'],
    ];
    return h('details', { class: 'card' }, h('summary', {}, 'Script cheat sheet'), h('table', { class: 'cheat' }, ...rows.map(([a, b]) => h('tr', {}, h('td', {}, a), h('td', {}, b)))));
  }

  // ---- step 3: images ----
  async function stepImages(p) {
    const show = curShow(); if (!show) return go('show');
    const r = await api('GET', '/api/admin/shows/' + show.id + '/script');
    if (!r.ok) return fail(r);
    A.analysis = r.data.analysis; A.text = A.saved = r.data.text;
    const a = A.analysis;
    p.append(h('h2', {}, 'Images'), h('p', { class: 'hint' }, 'PNG, JPG, GIF or WebP, up to 10 MB each. Images are sent to every device as they are, so keep them under about 1 MB (about 1920 pixels wide is plenty).'));

    async function upload(files, renameTo) {
      const fd = new FormData();
      [...files].forEach((f) => fd.append('files', f, renameTo || f.name));
      const res = await api('POST', '/api/admin/shows/' + show.id + '/assets', undefined, fd);
      if (!res.ok) return fail(res, 'Upload failed');
      (res.data.failed || []).forEach((f) => toast(f.name + ': ' + f.error));
      if ((res.data.saved || []).length) toast('Uploaded ' + res.data.saved.length + ' image(s)');
      render();
    }
    if (a.missing.length) {
      p.append(h('div', { class: 'card missing' }, h('h3', {}, '⚠ Images the script uses but you have not uploaded'),
        ...a.missing.map((m) => h('div', { class: 'row' }, h('b', {}, m.name), h('span', { class: 'hint' }, 'used on line ' + m.lines.join(', ')),
          h('label', { class: 'row' }, h('input', { type: 'file', accept: 'image/png,image/jpeg,image/gif,image/webp', onchange: (e) => e.target.files[0] && upload([e.target.files[0]], m.name) }))))));
    }
    const dz = h('div', { class: 'dropzone' }, h('p', {}, 'Drag images here, or '), h('input', { type: 'file', multiple: true, accept: 'image/png,image/jpeg,image/gif,image/webp', onchange: (e) => upload(e.target.files) }));
    dz.addEventListener('dragover', (e) => { e.preventDefault(); dz.classList.add('drop'); });
    dz.addEventListener('dragleave', () => dz.classList.remove('drop'));
    dz.addEventListener('drop', (e) => { e.preventDefault(); dz.classList.remove('drop'); upload(e.dataTransfer.files); });
    p.append(dz);
    if (!a.assets.length) { p.append(h('p', { class: 'hint' }, 'No images yet. Shows work fine without them.')); return; }
    p.append(h('div', { class: 'grid' }, ...a.assets.map((as) => {
      const lines = a.used[as.name];
      return h('div', { class: 'thumb' },
        h('img', { src: '/api/admin/shows/' + show.id + '/assets/' + encodeURIComponent(as.name), alt: as.name, loading: 'lazy' }),
        h('div', { class: 'meta' }, h('b', {}, as.name), fmtSize(as.size), as.large ? h('div', { class: 'status-warn' }, '⚠ large: may load slowly on many devices') : null,
          h('div', { class: lines ? '' : 'hint' }, lines ? 'Used on line ' + lines.join(', ') : 'Not used'),
          h('button', { class: 'danger quiet', onclick: async () => {
            if (!await confirmBox('Delete “' + as.name + '”?', lines ? 'The script uses this image on line ' + lines.join(', ') + '.' : 'It is not used in the script.', 'Delete', true)) return;
            const d = await api('DELETE', '/api/admin/shows/' + show.id + '/assets/' + encodeURIComponent(as.name)); if (!d.ok) return fail(d); render();
          } }, 'Delete')));
    })));
  }

  // ---- step 4: look ----
  const PRESETS = [['White on black (default)', '#000000', '#ffffff', '#ffffff'], ['Black on white', '#ffffff', '#000000', '#000000'], ['Yellow on black', '#000000', '#ffee00', '#ffee00'], ['Dark blue', '#001a33', '#ffffff', '#ffd166']];
  const FONTS = { system: 'system-ui, sans-serif', serif: 'Georgia, "Times New Roman", serif', mono: 'ui-monospace, Menlo, Consolas, monospace' };
  function lum(hex) { const c = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255).map((v) => (v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4))); return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]; }
  function contrast(a, b) { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); }
  async function stepLook(p) {
    const show = curShow(); if (!show) return go('show');
    const r = await api('GET', '/api/admin/shows/' + show.id); if (!r.ok) return fail(r);
    const t = Object.assign({}, r.data.theme);
    const ctl = {};
    const mk = (name, label, input) => { ctl[name] = input; return h('label', { class: 'field' }, h('span', {}, label), input); };
    mk('bg', 'Background', h('input', { type: 'color' })); mk('fg', 'Text', h('input', { type: 'color' })); mk('accent', 'Speaker names', h('input', { type: 'color' }));
    mk('font', 'Font', h('select', {}, h('option', { value: 'system' }, 'Clean (system)'), h('option', { value: 'serif' }, 'Serif'), h('option', { value: 'mono' }, 'Monospace')));
    mk('scalePct', 'Text size', h('input', { type: 'range', min: '60', max: '200', step: '10' }));
    mk('speakers', 'Speaker names', h('select', {}, h('option', { value: 'label' }, 'Show above the line'), h('option', { value: 'hidden' }, 'Hide'), h('option', { value: 'color' }, 'Show in the accent color')));
    mk('idleMessage', 'Message before the show starts', h('input', { type: 'text', maxlength: '200' }));
    const ratio = h('div', { class: 'hint' });
    const sizeLabel = h('span', { class: 'hint' });
    const phone = h('div', { class: 'preview phone', role: 'img', 'aria-label': 'Phone preview' }), proj = h('div', { class: 'preview proj', role: 'img', 'aria-label': 'Projector preview' });
    const sample = () => [h('p', { class: 'dim2' }, 'The train was late, Margaret.'), h('p', { class: 'dim1' }, h('span', { class: 'spk' }, 'Margaret'), 'The train is never late.'), h('p', {}, h('span', { class: 'spk' }, 'Edwin'), 'Is that tea for ', h('b', {}, 'me?'))];

    function read() { for (const k of ['bg', 'fg', 'accent', 'font', 'speakers', 'idleMessage']) t[k] = ctl[k].value; t.scalePct = Number(ctl.scalePct.value); }
    function write() { for (const k of ['bg', 'fg', 'accent', 'font', 'speakers', 'idleMessage', 'scalePct']) ctl[k].value = t[k]; }
    function paint() {
      const c = contrast(t.bg, t.fg);
      ratio.textContent = 'Contrast ' + c.toFixed(1) + ' : 1 — ' + (c >= 7 ? 'excellent' : c >= 4.5 ? 'acceptable, but higher is easier to read in a dark room' : 'too low to read comfortably');
      ratio.className = c >= 7 ? 'status-ok' : c >= 4.5 ? 'status-warn' : 'status-bad';
      sizeLabel.textContent = t.scalePct + '%';
      for (const f of [phone, proj]) {
        f.style.setProperty('--pbg', t.bg); f.style.setProperty('--pfg', t.fg); f.style.setProperty('--paccent', t.speakers === 'color' ? t.accent : t.fg);
        f.style.setProperty('--pfont', FONTS[t.font] || FONTS.system); f.style.setProperty('--pscale', String(t.scalePct / 100));
        f.replaceChildren(...sample());
        if (t.speakers === 'hidden') f.querySelectorAll('.spk').forEach((e) => e.remove());
      }
    }
    write(); paint();
    Object.values(ctl).forEach((i) => i.addEventListener('input', () => { read(); paint(); }));
    const presets = h('div', { class: 'presets' }, ...PRESETS.map(([n, bg, fg, ac]) => h('button', { onclick: () => { Object.assign(t, { bg, fg, accent: ac }); write(); paint(); } }, n)));
    p.append(h('h2', {}, 'Look'), h('p', { class: 'hint' }, 'Large, high-contrast text is the default and is best for accessibility. Changes apply to connected devices as soon as you save.'),
      h('div', { class: 'look-grid' },
        h('div', { class: 'card' }, presets, ...Object.keys(ctl).map((k) => ctl[k].closest('label')), sizeLabel, ratio,
          h('div', { class: 'row end' }, h('button', { class: 'primary', onclick: async () => { read(); const s = await api('PUT', '/api/admin/shows/' + show.id, { theme: t }); if (!s.ok) return fail(s); markLookSaved(); toast('Look saved'); } }, 'Save look'))),
        h('div', { class: 'previews' }, h('div', {}, h('div', { class: 'hint' }, 'Phone'), phone), h('div', {}, h('div', { class: 'hint' }, 'Projector'), proj))));
  }

  // ---- step 5: test (rehearsal) ----
  async function stepTest(p) {
    const show = curShow(); if (!show) return go('show');
    p.append(h('h2', {}, 'Test with a rehearsal'),
      h('p', { class: 'hint' }, 'A rehearsal runs this show privately: you press the keys as the operator would and see the audience screen beside it. Nothing here reaches real audience devices, and it does not affect a show that is live.'));
    const area = h('div', {});
    p.append(h('div', { class: 'row' }, h('button', { class: 'primary', onclick: start }, A.rehearsing ? 'Restart rehearsal' : 'Start rehearsal')), area);
    async function start() {
      const r = await api('POST', '/api/admin/rehearsal/start', { showId: show.id }); if (!r.ok) return fail(r);
      A.rehearsing = true; refreshNav();
      area.replaceChildren(h('div', { class: 'reh' },
        h('iframe', { src: '/operator?rehearsal=1', title: 'Operator (rehearsal)' }),
        h('div', { class: 'phoneframe' }, h('div', { class: 'hint' }, 'What the audience sees'), h('iframe', { src: '/?rehearsal=1', title: 'Audience (rehearsal)' }))),
        h('p', { class: 'hint' }, 'Click inside the operator panel, then use Space to send lines. Shift+Q stops a loop.'));
    }
    if (A.rehearsing) start();
  }

  // ---- step 6: go live ----
  async function stepGoLive(p) {
    const show = curShow(); if (!show) return go('show');
    const [net, sc] = await Promise.all([api('GET', '/api/admin/network'), api('GET', '/api/admin/shows/' + show.id + '/script')]);
    if (!net.ok || !sc.ok) return fail(net.ok ? sc : net);
    A.analysis = sc.data.analysis; A.net = net.data;
    const a = A.analysis, errs = a.diagnostics.filter((d) => d.severity === 'error').length, warns = a.diagnostics.filter((d) => d.severity === 'warning').length;
    const checks = h('div', { class: 'card' }, h('h3', {}, 'Ready?'),
      h('div', { class: 'check' + (a.cues.length ? '' : ' bad') }, a.cues.length ? a.cues.length + ' lines in the script' : 'The script is empty'),
      h('div', { class: 'check' + (errs ? ' bad' : warns ? ' warn' : '') }, errs ? errs + ' script error(s) to fix' : warns ? warns + ' script warning(s) to review' : 'No script problems'),
      h('div', { class: 'check' + (a.missing.length ? ' warn' : '') }, a.missing.length ? 'Missing images: ' + a.missing.map((m) => m.name).join(', ') : 'All images are uploaded'));
    const urls = net.data.urls.filter((u) => u.public);
    let chosen = (urls.find((u) => u.kind === 'http') || urls[0] || net.data.urls[0] || {}).url;
    const big = h('div', { class: 'urlbig' }), qrBox = h('div', { class: 'qr' });
    const paint = () => { big.textContent = chosen || 'No address found'; qrBox.replaceChildren(chosen ? h('img', { src: '/api/admin/network/qr.png?size=512&url=' + encodeURIComponent(chosen), alt: 'QR code for ' + chosen }) : ''); };
    paint();
    const list = h('div', { class: 'urls' }, ...net.data.urls.map((u) => h('div', { class: 'row' }, h('label', { class: 'row' }, h('input', { type: 'radio', name: 'u', checked: u.url === chosen, onchange: () => { chosen = u.url; paint(); } }), u.url, u.public ? '' : ' (this computer only)'))));
    const isActive = A.activeId === show.id;
    p.append(h('h2', {}, 'Go live'), checks,
      h('div', { class: 'card' }, h('h3', {}, 'Tell the audience this address'), big,
        h('div', { class: 'row' }, h('button', { onclick: async () => { try { await navigator.clipboard.writeText(chosen); toast('Copied'); } catch (e) { toast('Select the address and copy it'); } } }, 'Copy address'),
          h('a', { href: '/admin?print=qr&url=' + encodeURIComponent(chosen || ''), target: '_blank', rel: 'noopener' }, 'Print a QR sign for the lobby')),
        h('div', { class: 'row' }, qrBox, list),
        ...(net.data.warnings || []).map((w) => h('div', { class: 'check warn' }, w)),
        h('p', { class: 'hint' }, 'Audience devices must be on the same WiFi as this computer. Devices connected right now: ', h('b', { id: 'audcount' }, String(A.audience)))),
      h('div', { class: 'card row' },
        h('button', { class: 'primary big', onclick: () => startShow(show), disabled: !a.cues.length }, isActive ? 'Restart the show and open the operator page' : 'Start show'),
        isActive ? h('a', { href: '/operator' }, 'Open the operator page without restarting') : null,
        h('span', { class: 'hint' }, isActive ? 'This show is active.' : 'Starting makes this the show audiences see.')),
      h('p', { class: 'hint' }, 'Something not working? ', h('a', { href: '#', onclick: (e) => { e.preventDefault(); openHelp('admin-troubleshooting'); } }, 'Troubleshooting help'), ' · ', h('a', { href: '/api/admin/support-bundle', download: 'support-bundle.zip' }, 'Download a support bundle')));
  }
  async function startShow(show) {
    let r = await api('POST', '/api/admin/shows/' + show.id + '/activate', {});
    if (r.status === 409 && r.data.live) {
      if (!await confirmBox('A show is live', 'Starting this show restarts it from the beginning and clears all audience screens. Continue?', 'Restart the show', true)) return;
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
    $('panel').replaceChildren(h('div', { class: 'sign' }, h('h1', {}, 'Open captions: scan to read along'),
      h('img', { src: '/api/admin/network/qr.png?size=1024&url=' + encodeURIComponent(url), alt: 'QR code' }), h('div', { class: 'u' }, url),
      h('p', {}, 'Join the venue’s guest WiFi first.'), h('button', { class: 'noprint primary', onclick: () => window.print() }, 'Print')));
  }

  const q = new URLSearchParams(location.search);
  if (q.get('print') === 'qr' && q.get('url')) fetch('/api/admin/ping').then((r) => (r.ok ? printSign(q.get('url')) : showLogin()));
  else boot();
})();
