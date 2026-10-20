// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Per-character colors for speaker names. Shared by the audience page and the Setup preview
// so both show the same colors. Colors come from a color-blind-friendly palette and are
// lightened or darkened as needed so every name stays readable on the chosen background.
(function (root) {
  'use strict';
  var PALETTE = ['#E69F00', '#56B4E9', '#009E73', '#F0E442', '#D55E00', '#CC79A7', '#0072B2', '#B3DE69', '#FB8072', '#BC80BD'];

  function rgb(hex) { return [1, 3, 5].map(function (i) { return parseInt(hex.slice(i, i + 2), 16); }); }
  function hex(c) { return '#' + c.map(function (v) { return ('0' + Math.max(0, Math.min(255, Math.round(v))).toString(16)).slice(-2); }).join(''); }
  function lum(c) {
    var l = c.map(function (v) { v /= 255; return v <= 0.03928 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4); });
    return 0.2126 * l[0] + 0.7152 * l[1] + 0.0722 * l[2];
  }
  function contrast(a, b) { var x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); }
  function mix(a, b, t) { return a.map(function (v, i) { return v + (b[i] - v) * t; }); }

  // color returns a readable color for the character at `index` of the show's speaker list.
  function color(index, bg) {
    var back = rgb(bg || '#000000');
    var c = rgb(PALETTE[index % PALETTE.length]);
    var round = Math.floor(index / PALETTE.length); // more characters than colors: vary the shade
    var toward = lum(back) < 0.4 ? [255, 255, 255] : [0, 0, 0];
    if (round) c = mix(c, toward, Math.min(0.5, 0.25 * round));
    for (var t = 0; t <= 1 && contrast(c, back) < 4.5; t += 0.05) c = mix(c, toward, 0.05);
    return hex(c);
  }

  function hash(s) { var h = 0; for (var i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) | 0; return Math.abs(h); }

  // forName: the same name always gets the same color, in the same order for every device
  // (the server sends the speakers in order of first appearance in the show file).
  function forName(name, speakers, bg) {
    var i = (speakers || []).indexOf(name);
    return color(i >= 0 ? i : hash(name), bg);
  }

  root.OCSpeakers = { color: color, forName: forName, contrast: function (a, b) { return contrast(rgb(a), rgb(b)); } };
})(window);
