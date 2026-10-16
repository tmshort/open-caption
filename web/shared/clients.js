// The audience count in the header of Setup and the operator page: its label, and the list of connected devices it opens.
// (Shared by both pages so they say the same thing.)
(function () {
  'use strict';
  var __ = function (s, v) { return window.OCI18n ? OCI18n.t(s, v) : s; };

  // "👥 5 (en 3 · es 2)": a language nobody is using is not listed, and with nobody in any language there is just the number.
  function label(p) {
    var txt = '👥 ' + p.audience, langs = p.languages || [], by = p.audienceByLang || {};
    if (langs.length > 1) {
      var parts = langs.filter(function (l) { return by[l.code] > 0; }).map(function (l) { return l.code + ' ' + by[l.code]; });
      if (parts.length) txt += ' (' + parts.join(' · ') + ')';
    }
    return txt;
  }

  // What kind of device: translated words for the kind, brand names as they are.
  function device(c) {
    var kind = { phone: __('Phone'), tablet: __('Tablet'), computer: __('Computer'), projector: __('Projector screen') }[c.kind] || __('Unknown device');
    var more = [c.os, c.browser].filter(Boolean).join(', ');
    return more ? kind + ' (' + more + ')' : kind;
  }

  function cell(tag, text) { var e = document.createElement(tag); e.textContent = text; return e; }

  // The table for the dialog: device, address, language.
  function table(rows) {
    var box = document.createElement('div');
    if (!rows.length) { box.appendChild(cell('p', __('Nobody is connected right now.'))); return box; }
    var t = document.createElement('table'); t.className = 'cheat clients';
    var head = document.createElement('tr');
    [__('Device'), __('IP address'), __('Language')].forEach(function (h) { head.appendChild(cell('th', h)); });
    t.appendChild(head);
    rows.forEach(function (c) {
      var tr = document.createElement('tr');
      tr.appendChild(cell('td', device(c))); tr.appendChild(cell('td', c.ip)); tr.appendChild(cell('td', c.langName || c.lang));
      t.appendChild(tr);
    });
    box.appendChild(t);
    return box;
  }

  window.OCClients = { label: label, table: table, title: function (n) { return __('{n} device(s) connected. Click to see which.', { n: n }); } };
})();
