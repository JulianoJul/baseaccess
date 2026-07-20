(function() {
  if (!window.Alpine) return;

  document.addEventListener('htmx:afterSettle', function(evt) {
    Alpine.initTree(evt.detail.target);
  });

  document.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target.id === 'vista-tabla') {
      if (window.Alpine) {
        const store = Alpine.store('fijados');
        if (store) {
          store.init();
          const ids = store.lista.map(i => i.id);
          document.querySelectorAll('[id^="pin-btn-"]').forEach(btn => {
            const match = btn.id.match(/pin-btn-(\d+)/);
            btn.className = match && ids.includes(Number(match[1]))
              ? 'btn-icon text-emerald-500 hover:text-emerald-400'
              : 'btn-icon text-blue-500 hover:text-blue-400';
          });
        }
      }
    }
  });
})();
