(function() {
  if (!window.Alpine) return;

  document.addEventListener('htmx:afterSettle', function(evt) {
    Alpine.initTree(evt.detail.target);
  });

  document.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target.id === 'vista-tabla') {
      const fijadosEl = document.querySelector('[x-data="fijados"]');
      if (fijadosEl && fijadosEl.__x) {
        const pinBtns = document.querySelectorAll('[id^="pin-btn-"]');
        pinBtns.forEach(btn => {
          btn.className = 'btn-icon text-blue-500 hover:text-blue-400';
        });
        fijadosEl.__x.$data.lista.forEach(item => {
          const btn = document.getElementById('pin-btn-' + item.id);
          if (btn) btn.className = 'btn-icon text-emerald-500 hover:text-emerald-400';
        });
      }

      if (window.PAGE_DATA && window.PAGE_DATA.hasDB) {
        const fijados = document.querySelector('[x-data="fijados"]');
        if (fijados && fijados.__x) {
          const ids = fijados.__x.$data.lista.map(i => i.id);
          const btnEliminar = document.querySelector('[hx-post*="eliminar"]');
          if (!btnEliminar) {
            document.querySelectorAll('[id^="pin-btn-"]').forEach(btn => {
              const match = btn.id.match(/pin-btn-(\d+)/);
              if (match && ids.includes(Number(match[1]))) {
                btn.className = 'btn-icon text-emerald-500 hover:text-emerald-400';
              }
            });
          }
        }
      }
    }
  });
})();
