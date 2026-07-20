document.addEventListener('alpine:init', () => {
  Alpine.directive('currency', (el, { modifiers }, { evaluateLater, effect }) => {
    const locale = modifiers[0] || 'es-ES';

    const onInput = () => {
      let v = el.value.replace(/,/g, '.').replace(/[^0-9.]/g, '');
      const lastDot = v.lastIndexOf('.');
      if (lastDot >= 0) {
        v = v.substring(0, lastDot).replace(/\./g, '') + v.substring(lastDot);
      }
      const parts = v.split('.');
      if (parts.length === 2 && parts[1].length > 2) {
        v = parts[0] + '.' + parts[1].substring(0, 2);
      }
      el.dataset.raw = v;
      el.value = v;
    };

    const onBlur = () => {
      const raw = el.dataset.raw;
      if (raw && !isNaN(parseFloat(raw))) {
        const n = parseFloat(raw);
        el.value = n.toLocaleString(locale, {
          minimumFractionDigits: 2,
          maximumFractionDigits: 2
        });
      }
    };

    const onFocus = () => {
      if (el.dataset.raw !== undefined) {
        el.value = el.dataset.raw;
      }
    };

    el.addEventListener('input', onInput);
    el.addEventListener('blur', onBlur);
    el.addEventListener('focus', onFocus);

    if (el.value) onBlur();

    return () => {
      el.removeEventListener('input', onInput);
      el.removeEventListener('blur', onBlur);
      el.removeEventListener('focus', onFocus);
    };
  });
});
