// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// A guided tour: a dimmed page with a ring round one control and a pop-up beside it with an arrow, Back / Next / Skip.
// It has no words of its own; admin.js passes the steps and the labels (so they are translated with the rest of Setup).
//
//   OCTour.start({ steps: [{ target: '#x' | () => element | null, title, text, before }], labels: { next, back, skip, done, step }, onEnd(finished) })
//
// A step whose target is missing or hidden is shown in the middle of the page without a ring. Keys: Right arrow / Enter =
// next, Left arrow = back, Esc = skip. The pop-up is a dialog that holds the keyboard focus until the tour ends.
(function () {
  'use strict';
  let live = null;

  function el(tag, cls, text) { const e = document.createElement(tag); if (cls) e.className = cls; if (text) e.textContent = text; return e; }

  function start(opts) {
    if (live) live.end(false);
    const steps = opts.steps, L = opts.labels;
    const before = document.activeElement;
    const ring = el('div', 'tour-ring'), pop = el('div', 'tour-pop'), arrow = el('div', 'tour-arrow');
    const shade = el('div', 'tour-shade');
    const title = el('h2', 'tour-title'), text = el('p', 'tour-text'), count = el('span', 'tour-count');
    const skip = el('button', 'tour-skip', L.skip), back = el('button', 'tour-back', L.back), next = el('button', 'primary tour-next');
    title.id = 'tourtitle'; text.id = 'tourtext';
    pop.setAttribute('role', 'dialog'); pop.setAttribute('aria-modal', 'true'); pop.setAttribute('aria-labelledby', 'tourtitle'); pop.setAttribute('aria-describedby', 'tourtext');
    [skip, back, next].forEach((b) => { b.type = 'button'; });
    const foot = el('div', 'tour-foot'); foot.append(count, skip, back, next);
    pop.append(arrow, title, text, foot);
    document.body.append(shade, ring, pop);
    let i = 0, target = null;

    const find = (s) => { try { const t = typeof s.target === 'function' ? s.target() : s.target ? document.querySelector(s.target) : null; return t && !t.closest('[hidden]') && t.getClientRects().length ? t : null; } catch (e) { return null; } };

    function place() {
      const vw = document.documentElement.clientWidth, vh = document.documentElement.clientHeight;
      const pw = Math.min(pop.offsetWidth, vw - 16), ph = pop.offsetHeight;
      if (!target) { // no control to point at: the middle of the page
        ring.hidden = true; arrow.hidden = true; shade.className = 'tour-shade bare';
        pop.style.left = Math.max(8, (vw - pw) / 2) + 'px'; pop.style.top = Math.max(8, (vh - ph) / 2) + 'px';
        return;
      }
      const r = target.getBoundingClientRect(), pad = 4;
      ring.hidden = false; arrow.hidden = false; shade.className = 'tour-shade';
      ring.style.left = r.left - pad + 'px'; ring.style.top = r.top - pad + 'px'; ring.style.width = r.width + 2 * pad + 'px'; ring.style.height = r.height + 2 * pad + 'px';
      const gap = 14, below = r.bottom + pad + gap + ph <= vh - 8, above = r.top - pad - gap - ph >= 8;
      const under = below || !above; // below when it fits, else above, else below and let it overlap
      const top = under ? Math.min(r.bottom + pad + gap, Math.max(8, vh - ph - 8)) : r.top - pad - gap - ph;
      const left = Math.min(Math.max(8, r.left + r.width / 2 - pw / 2), vw - pw - 8);
      pop.style.left = left + 'px'; pop.style.top = Math.max(8, top) + 'px';
      const ax = Math.min(Math.max(18, r.left + r.width / 2 - left), pw - 18);
      arrow.style.left = ax - 7 + 'px';
      arrow.className = 'tour-arrow ' + (under ? 'up' : 'down');
    }

    async function show() {
      const s = steps[i];
      if (s.before) await s.before();
      target = find(s);
      if (target && target.scrollIntoView) target.scrollIntoView({ block: 'nearest', inline: 'nearest' });
      title.textContent = s.title; text.textContent = s.text;
      count.textContent = L.step.replace('{n}', String(i + 1)).replace('{total}', String(steps.length));
      back.hidden = i === 0;
      next.textContent = i === steps.length - 1 ? L.done : L.next;
      skip.hidden = i === steps.length - 1;
      place();
      next.focus();
    }

    function end(finished) {
      document.removeEventListener('keydown', onKey, true); window.removeEventListener('resize', place); window.removeEventListener('scroll', place, true);
      shade.remove(); ring.remove(); pop.remove(); live = null;
      if (before && before.focus && document.contains(before)) before.focus();
      if (opts.onEnd) opts.onEnd(finished);
    }
    const go = (d) => { if (i + d >= steps.length) { end(true); return; } if (i + d < 0) return; i += d; show(); };
    function onKey(e) {
      if (e.key === 'Escape') { e.preventDefault(); e.stopPropagation(); end(false); }
      else if (e.key === 'ArrowRight') { e.preventDefault(); go(1); }
      else if (e.key === 'ArrowLeft') { e.preventDefault(); go(-1); }
      else if (e.key === 'Tab') { // the keyboard stays in the pop-up
        const bs = [skip, back, next].filter((b) => !b.hidden), k = bs.indexOf(document.activeElement);
        e.preventDefault(); bs[(k + (e.shiftKey ? bs.length - 1 : 1)) % bs.length].focus();
      }
    }
    next.addEventListener('click', () => go(1)); back.addEventListener('click', () => go(-1)); skip.addEventListener('click', () => end(false));
    document.addEventListener('keydown', onKey, true); window.addEventListener('resize', place); window.addEventListener('scroll', place, true);
    live = { end };
    show();
  }

  window.OCTour = { start, running: () => !!live };
})();
