// Translation of Setup, the operator page and Help into the six languages the audience page has
// (English, Spanish, French, German, Italian, Portuguese). English text is the key: a translation is looked
// up by the English sentence, and a sentence with no translation simply shows in English.
//
//   OCI18n.t('Save')                            -> 'Guardar'
//   OCI18n.t('Updated {when}', { when: '...' }) -> 'Actualizado ...'
//
// A key containing {name} placeholders also matches text produced by the server ("Image "x" not found"):
// the dictionary entry 'Image {0} not found.' translates it, with the matched parts filled back in.
// Dictionaries live in /static/shared/i18n/<code>.js and are loaded before the page's own script.
(function () {
  'use strict';
  var LANGS = [['en', 'English'], ['es', 'Español'], ['fr', 'Français'], ['de', 'Deutsch'], ['it', 'Italiano'], ['pt', 'Português']];
  var KEY = 'oc.ui.lang';
  var dicts = {};
  var compiled = {}; // code -> [{re, names, text, specificity}]

  function supported(c) { for (var i = 0; i < LANGS.length; i++) if (LANGS[i][0] === c) return true; return false; }
  function store(op, v) { try { return op === 'get' ? localStorage.getItem(KEY) : localStorage.setItem(KEY, v); } catch (e) { return null; } }
  function pick() {
    var q = /[?&]ui=([a-z-]+)/.exec(location.search);
    if (q && supported(q[1].split('-')[0])) return q[1].split('-')[0];
    var s = store('get');
    if (s && supported(s)) return s;
    var prefs = navigator.languages && navigator.languages.length ? navigator.languages : [navigator.language || 'en'];
    for (var i = 0; i < prefs.length; i++) { var c = String(prefs[i]).toLowerCase().split('-')[0]; if (supported(c)) return c; }
    return 'en';
  }
  var lang = pick();

  function esc(s) { return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'); }
  function compile(code) {
    var d = dicts[code] || {}, out = [];
    Object.keys(d).forEach(function (k) {
      if (!/\{\w+\}/.test(k)) return;
      var names = [], re = '^' + k.split(/(\{\w+\})/).map(function (part) {
        var m = /^\{(\w+)\}$/.exec(part);
        if (m) { names.push(m[1]); return '([\\s\\S]+?)'; }
        return esc(part);
      }).join('') + '$';
      out.push({ re: new RegExp(re), names: names, text: d[k], literal: k.replace(/\{\w+\}/g, '').length });
    });
    out.sort(function (a, b) { return b.literal - a.literal; }); // the most specific template first
    compiled[code] = out;
    return out;
  }
  function fill(text, vars) {
    if (!vars) return text;
    return text.replace(/\{(\w+)\}/g, function (m, n) { return Object.prototype.hasOwnProperty.call(vars, n) ? String(vars[n]) : m; });
  }
  function lookup(s) {
    var d = dicts[lang];
    if (!d) return null;
    if (Object.prototype.hasOwnProperty.call(d, s)) return d[s];
    return null;
  }
  // translate one piece of text; null if there is no translation for it
  function one(s) {
    var hit = lookup(s);
    if (hit !== null) return hit;
    var ts = compiled[lang] || compile(lang);
    for (var i = 0; i < ts.length; i++) {
      var m = ts[i].re.exec(s);
      if (m) {
        var v = {};
        for (var j = 0; j < ts[i].names.length; j++) v[ts[i].names[j]] = piece(m[j + 1]); // pieces may be messages too
        return fill(ts[i].text, v);
      }
    }
    return null;
  }
  // a captured piece: translate it if it is a known message, keeping the spaces around it
  function piece(p) {
    var core = p.trim();
    if (!core) return p;
    var out = one(core);
    return out === null ? p : p.slice(0, p.indexOf(core)) + out + p.slice(p.indexOf(core) + core.length);
  }
  // t translates s. vars fill {name} placeholders (after translation, so a translator can reorder them).
  function t(s, vars) {
    if (typeof s !== 'string' || s === '' || lang === 'en') return fill(s, vars);
    var out = one(s);
    if (out !== null) return fill(out, vars);
    if (s !== s.trim()) return fill(piece(s), vars); // text with spaces at either end (parts of a longer line)
    // "invalid name: give the show a name": a message built from parts joined by ": "
    if (s.indexOf(': ') > 0) {
      var parts = s.split(': '), changed = false;
      var tr = parts.map(function (x) { var y = one(x); if (y !== null) changed = true; return y === null ? x : y; });
      if (changed) return fill(tr.join(': '), vars);
    }
    return fill(s, vars);
  }

  var ATTRS = ['placeholder', 'title', 'aria-label', 'alt'];
  // apply translates the text and the placeholder/title/aria-label/alt attributes already in the page.
  function apply(root) {
    if (lang === 'en') return;
    root = root || document.body;
    // Elements whose text has inline markup are translated as a whole: the key is the inner HTML.
    var wholes = root.querySelectorAll ? root.querySelectorAll('[data-t]') : [];
    for (var k = 0; k < wholes.length; k++) {
      var html = wholes[k].innerHTML.trim(), th = t(html);
      if (th !== html) wholes[k].innerHTML = th;
      wholes[k].removeAttribute('data-t'); wholes[k].setAttribute('data-t-done', '');
    }
    var w = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, null, false), n, nodes = [];
    while ((n = w.nextNode())) {
      var p = n.parentNode && n.parentNode.nodeName;
      if (p === 'SCRIPT' || p === 'STYLE' || p === 'TEXTAREA' || p === 'PRE' || p === 'CODE') continue;
      if (n.parentNode.closest && n.parentNode.closest('[data-t-done]')) continue;
      nodes.push(n);
    }
    nodes.forEach(function (node) {
      var raw = node.nodeValue, trimmed = raw.trim();
      if (!trimmed) return;
      var out = t(trimmed);
      if (out !== trimmed) node.nodeValue = raw.replace(trimmed, out);
    });
    var els = root.querySelectorAll ? root.querySelectorAll('[placeholder],[title],[aria-label],[alt]') : [];
    for (var i = 0; i < els.length; i++) {
      for (var a = 0; a < ATTRS.length; a++) {
        var v = els[i].getAttribute(ATTRS[a]);
        if (v) { var tv = t(v); if (tv !== v) els[i].setAttribute(ATTRS[a], tv); }
      }
    }
    if (root === document.body && document.title) document.title = t(document.title);
  }

  // A small language menu (a <select>), shown in the page header.
  function picker(onChange) {
    var sel = document.createElement('select');
    sel.className = 'uilang';
    sel.setAttribute('aria-label', 'Language of this page');
    LANGS.forEach(function (l) { var o = document.createElement('option'); o.value = l[0]; o.textContent = l[1]; sel.appendChild(o); });
    sel.value = lang;
    sel.addEventListener('change', function () {
      if (onChange && onChange(sel.value) === false) { sel.value = lang; return; }
      store('set', sel.value);
      // a ?ui= in the address would win over the choice just made: take it out. Going to an address that only differs by its #hash
      // (or not at all) is not a navigation, so when nothing was taken out the page is reloaded on purpose.
      var search = location.search.replace(/([?&])ui=[^&]*&?/, '$1').replace(/[?&]$/, '');
      if (search !== location.search) location.replace(location.pathname + search + location.hash);
      else location.reload();
    });
    return sel;
  }

  window.OCI18n = {
    lang: lang, langs: LANGS, t: t, apply: apply, picker: picker, supported: supported,
    add: function (code, dict) { dicts[code] = Object.assign(dicts[code] || {}, dict); delete compiled[code]; },
    keys: function (code) { return Object.keys(dicts[code] || {}); },
  };
  window.__ = t;

  // Load the dictionary now, before the page's own script (which comes next) runs.
  if (lang !== 'en') document.write('<script src="/static/shared/i18n/' + lang + '.js"><\/script>');
  document.documentElement.setAttribute('lang', lang);
})();
