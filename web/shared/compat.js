// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Compatibility for older browsers, loaded first on every page. Written in plain old JavaScript on purpose:
// it must run (and explain itself) in a browser that cannot run the rest of the page.
//
//  1. Gate: if the browser lacks something the page cannot work without, say so in plain words instead of
//     showing a blank screen. The audience page needs about Safari 12 / iOS 12 / Chrome 66 or newer; Setup,
//     the operator page and Help need about Safari 14.1 / iOS 14.5 / Chrome 87 / Firefox 66 or newer.
//  2. Fill-ins for features that are missing in browsers that are otherwise fine: Element.replaceChildren
//     (before Safari 14) and <dialog> (before Safari 15.4 / Firefox 98).
(function () {
  'use strict';
  var modern = document.documentElement.getAttribute('data-floor') === 'modern';

  function supported() {
    // ES2018 (async functions, object spread): Symbol.asyncIterator arrived with it. (Not tested by compiling code:
    // the pages' Content-Security-Policy forbids that.)
    if (!(window.Symbol && Symbol.asyncIterator)) return false;
    if (!window.fetch || !window.Promise || !window.EventSource || !window.URLSearchParams) return false;
    if (!(window.CSS && CSS.supports && CSS.supports('color', 'var(--x)'))) return false;
    // Setup and the operator page lay themselves out with flex gaps and inset, which arrived together.
    if (modern && !CSS.supports('inset', '0')) return false;
    return true;
  }

  if (!supported()) {
    window.OC_UNSUPPORTED = true;
    document.addEventListener('DOMContentLoaded', function () {
      var box = document.createElement('div');
      box.setAttribute('role', 'alert');
      box.style.cssText = 'font:18px/1.5 sans-serif;max-width:34em;margin:2em auto;padding:1em 1.2em;border:3px solid #c00;border-radius:8px;background:#fff;color:#000';
      var h = document.createElement('b');
      h.textContent = 'This browser is too old for this page.';
      var p = document.createElement('p');
      p.textContent = modern
        ? 'Please use a current version of Safari, Chrome, Edge or Firefox (Safari 14.1 or iOS 14.5 and newer), or another computer.'
        : 'Please update your phone or tablet, or open this address in a current browser such as Safari, Chrome, Edge or Firefox.';
      box.appendChild(h); box.appendChild(p);
      document.body.insertBefore(box, document.body.firstChild);
    });
    return;
  }

  // ---- Element.replaceChildren (Safari < 14, Chrome < 86, Firefox < 78) ----
  if (!Element.prototype.replaceChildren) {
    Element.prototype.replaceChildren = function () {
      while (this.firstChild) this.removeChild(this.firstChild);
      for (var i = 0; i < arguments.length; i++) {
        var n = arguments[i];
        this.appendChild(typeof n === 'string' ? document.createTextNode(n) : n);
      }
    };
  }

  // ---- <dialog> (Safari < 15.4, Chrome < 37, Firefox < 98) ----
  // Gives each <dialog> showModal() and close(), a 'close' event (and onclose), returnValue, Escape to close,
  // Tab kept inside, and <form method="dialog"> buttons that close it with the button's value.
  if (!(window.HTMLDialogElement && HTMLDialogElement.prototype.showModal)) {
    document.addEventListener('DOMContentLoaded', function () {
      var dialogs = document.getElementsByTagName('dialog');
      for (var i = 0; i < dialogs.length; i++) install(dialogs[i]);
    });
  }

  function install(d) {
    var before = null, last = null;
    d.className += ' oc-fallback';
    d.returnValue = '';
    d.showModal = function () {
      if (d.hasAttribute('open')) return;
      before = document.activeElement;
      d.setAttribute('open', '');
      document.addEventListener('keydown', onKey, true);
      var f = d.querySelector('[autofocus], textarea, input, select, button');
      if (f) f.focus();
    };
    d.close = function (value) {
      if (!d.hasAttribute('open')) return;
      if (value !== undefined) d.returnValue = String(value);
      d.removeAttribute('open');
      document.removeEventListener('keydown', onKey, true);
      if (before && before.focus) before.focus();
      var ev = document.createEvent('Event');
      ev.initEvent('close', false, false);
      d.dispatchEvent(ev); // addEventListener('close', …)
      if (typeof d.onclose === 'function') d.onclose(ev); // d.onclose = … is not wired up on an unknown element
    };
    function onKey(e) {
      if (e.key === 'Escape' || e.key === 'Esc') { e.preventDefault(); e.stopPropagation(); d.close(); return; }
      if (e.key !== 'Tab') return;
      var items = d.querySelectorAll('a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled])');
      var vis = [];
      for (var k = 0; k < items.length; k++) if (items[k].offsetParent !== null) vis.push(items[k]);
      if (!vis.length) return;
      var first = vis[0], end = vis[vis.length - 1];
      if (e.shiftKey && (document.activeElement === first || !d.contains(document.activeElement))) { e.preventDefault(); end.focus(); }
      else if (!e.shiftKey && (document.activeElement === end || !d.contains(document.activeElement))) { e.preventDefault(); first.focus(); }
    }
    d.addEventListener('click', function (e) { var b = e.target; while (b && b !== d && b.tagName !== 'BUTTON') b = b.parentNode; last = b && b !== d ? b : null; }, true);
    var forms = d.getElementsByTagName('form');
    for (var j = 0; j < forms.length; j++) {
      if ((forms[j].getAttribute('method') || '').toLowerCase() !== 'dialog') continue;
      forms[j].addEventListener('submit', function (e) {
        e.preventDefault();
        var v = last && last.value ? last.value : '';
        last = null;
        d.close(v);
      });
    }
  }
})();
