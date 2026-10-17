const editor = document.querySelector('[data-content-editor]');
if (editor) {
  const language = editor.querySelector('[data-language]');
  const kind = editor.querySelector('[data-kind]');
  const templates = {
    section: editor.querySelector('.content-sections fieldset').cloneNode(true),
    step: editor.querySelector('.content-steps fieldset').cloneNode(true)
  };
  function refresh() {
    editor.querySelectorAll('[data-lang]').forEach(block => block.hidden = language.value !== 'both' && block.dataset.lang !== language.value);
    ['en', 'ar'].forEach(lang => editor.querySelector('[name="title_' + lang + '"]').required = language.value === 'both' || language.value === lang);
    editor.querySelector('[data-path-fields]').hidden = !kind || kind.value !== 'path';
  }
  language.addEventListener('change', refresh);
  kind?.addEventListener('change', refresh);
  editor.addEventListener('click', event => {
    const add = event.target.closest('[data-add]');
    const remove = event.target.closest('[data-remove]');
    if (add) {
      const name = add.dataset.add;
      const field = templates[name].cloneNode(true);
      field.querySelectorAll('input,textarea,select').forEach(input => input.value = '');
      editor.querySelector(name === 'step' ? '.content-steps' : '.content-sections').append(field);
      refresh();
      field.querySelector('div:not([hidden]) input')?.focus();
    }
    if (remove) remove.closest('fieldset').remove();
  });
  refresh();
}
const search = document.querySelector('[data-content-search]');
if (search) {
  const update = () => {
    let count = 0;
    document.querySelectorAll('[data-content-row]').forEach(row => {
      row.hidden = !row.textContent.toLowerCase().includes(search.value.toLowerCase());
      if (!row.hidden) count++;
    });
    document.querySelector('[data-empty-list]').hidden = count > 0;
  };
  search.addEventListener('input', update);
  update();
}
