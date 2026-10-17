(() => {
  const root = document.querySelector('[data-path]');
  if (!root) return;
  const key = 'sn-learning-v1:' + root.dataset.path;
  const boxes = [...root.querySelectorAll('[data-step]')];
  const label = root.querySelector('[data-progress]');
  const bar = root.querySelector('progress');
  const arabic = document.documentElement.lang === 'ar';
  let saved = [];
  let storageAvailable = true;
  try { const data = JSON.parse(localStorage.getItem(key) || '[]'); if (Array.isArray(data)) saved = data; } catch { storageAvailable = false; }
  boxes.forEach(box => { box.checked = saved.includes(box.dataset.step); });
  function update(persist) {
    const complete = boxes.filter(box => box.checked).map(box => box.dataset.step);
    bar.value = complete.length;
    label.textContent = `${complete.length} / ${boxes.length} ${arabic ? 'مكتمل' : 'complete'}`;
    if (persist && storageAvailable) {
      try { localStorage.setItem(key, JSON.stringify(complete)); } catch { storageAvailable = false; }
    }
    if (!storageAvailable) root.querySelector('.progress-note').textContent = arabic ? 'التخزين غير متاح. سيُفقد التقدّم عند إغلاق الصفحة.' : 'Browser storage is unavailable. Progress will not be saved after you leave this page.';
  }
  boxes.forEach(box => box.addEventListener('change', () => update(true)));
  update(false);
})();
