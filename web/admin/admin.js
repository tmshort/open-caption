(async () => {
  const $ = (id) => document.getElementById(id);
  async function me() {
    const r = await fetch('/api/admin/ping');
    return r.ok;
  }
  function show(ok) { $('login').hidden = ok; $('app').hidden = !ok; if (!ok) $('pw').focus(); }
  show(await me());
  $('login').addEventListener('submit', async (e) => {
    e.preventDefault();
    const r = await fetch('/api/login', { method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ role: 'admin', password: $('pw').value }) });
    if (r.ok) { $('err').textContent = ''; show(true); return; }
    $('err').textContent = r.status === 429 ? 'Too many attempts. Wait a minute and try again.' : 'That password was not accepted.';
  });
  $('logout').addEventListener('click', async () => { await fetch('/api/logout', { method: 'POST' }); show(false); });
})();
