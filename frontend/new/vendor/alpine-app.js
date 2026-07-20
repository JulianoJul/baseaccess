document.addEventListener('alpine:init', () => {
  // --- App Shell (drag & drop, estado global UI) ---
  Alpine.data('appShell', () => ({
    dragOver: false,

    async onDrop(e) {
      this.dragOver = false;
      const f = e.dataTransfer?.files?.[0];
      if (!f) return;
      const path = f.path || f.name;
      try {
        await fetch('/api/abrir-bd', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: 'path=' + encodeURIComponent(path)
        });
        location.reload();
      } catch(err) {
        Alpine.store('toast').error('Error: ' + err);
      }
    }
  }));

  // --- Global Stores ---
  Alpine.store('modals', {
    stack: [],

    get abierto() { return this.stack.length > 0; },

    abrir(id) {
      if (this.stack.includes(id)) return;
      this.stack.push(id);
      document.body.style.overflow = 'hidden';
    },

    cerrar(id) {
      const idx = this.stack.lastIndexOf(id);
      if (idx >= 0) this.stack.splice(idx, 1);
      if (this.stack.length === 0) document.body.style.overflow = '';
    },

    toggle(id) {
      this.tiene(id) ? this.cerrar(id) : this.abrir(id);
    },

    tiene(id) {
      return this.stack.includes(id);
    },

    cerrarClickFuera(e, id) {
      if (e.target === e.currentTarget) {
        this.cerrar(id);
      }
    }
  });

  Alpine.store('toast', {
    mostrar(msg, tipo = 'info') {
      const el = document.createElement('div');
      el.className = 'toast ' + tipo;
      el.textContent = msg;
      document.getElementById('toast-container')?.appendChild(el);
      requestAnimationFrame(() => el.classList.add('show'));
      setTimeout(() => {
        el.classList.remove('show');
        setTimeout(() => el.remove(), 300);
      }, 3000);
    },

    error(msg) { this.mostrar(msg, 'error'); },
    success(msg) { this.mostrar(msg, 'success'); },
    info(msg) { this.mostrar(msg, 'info'); }
  });

  // --- Fijados (Pines / Acceso Rápido) ---
  Alpine.data('fijados', () => ({
    lista: [],

    init() {
      this.cargar();
    },

    cargar() {
      try {
        this.lista = JSON.parse(localStorage.getItem('sidebarFrecuentes') || '[]');
      } catch (e) { this.lista = []; }
    },

    guardar() {
      localStorage.setItem('sidebarFrecuentes', JSON.stringify(this.lista));
    },

    toggle(id, solped, modulo) {
      modulo = modulo || 'expedientes';
      const idx = this.lista.findIndex(i => Number(i.id) === Number(id) && i.modulo === modulo);
      if (idx >= 0) {
        this.lista.splice(idx, 1);
      } else {
        this.lista.push({ id: Number(id), solped: solped || '', modulo });
      }
      this.guardar();
    },

    estaFijado(id, modulo) {
      modulo = modulo || 'expedientes';
      return this.lista.some(i => Number(i.id) === Number(id) && i.modulo === modulo);
    },

    eliminar(idx) {
      this.lista.splice(idx, 1);
      this.guardar();
    }
  }));

  // --- BD Recientes ---
  Alpine.data('bdRecientes', () => ({
    lista: [],

    init() {
      this.cargar();
    },

    cargar() {
      try {
        this.lista = JSON.parse(localStorage.getItem('baseaccess_recientes') || '[]');
      } catch (e) { this.lista = []; }
    },

    guardar() {
      localStorage.setItem('baseaccess_recientes', JSON.stringify(this.lista));
    },

    registrar(nombre, path) {
      if (!nombre || !path) return;
      const idx = this.lista.findIndex(r => r.path === path);
      if (idx >= 0) this.lista.splice(idx, 1);
      this.lista.unshift({ nombre, path, timestamp: Date.now() });
      if (this.lista.length > 5) this.lista = this.lista.slice(0, 5);
      this.guardar();
    },

    async abrir(path) {
      if (!path) return;
      try {
        const res = await fetch('/api/abrir-bd', {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: 'path=' + encodeURIComponent(path)
        });
        if (res.ok) {
          location.reload();
        } else {
          const err = await res.text();
          Alpine.store('toast').mostrar('Error: ' + err, 'error');
          this.eliminarPorPath(path);
        }
      } catch (err) {
        Alpine.store('toast').mostrar('Error de conexión: ' + err, 'error');
        this.eliminarPorPath(path);
      }
    },

    eliminarPorIndex(idx) {
      this.lista.splice(idx, 1);
      this.guardar();
    },

    eliminarPorPath(path) {
      this.lista = this.lista.filter(r => r.path !== path);
      this.guardar();
    }
  }));

  // --- Sumas (Calculadora) ---
  Alpine.data('calculadoraSumas', () => ({
    filas: [{ valor: '' }],
    resultado: 0,

    calcular() {
      this.resultado = this.filas.reduce((sum, f) => {
        const v = parseFloat(String(f.valor || '').replace(',', '.').replace(/[^0-9.]/g, '')) || 0;
        return sum + v;
      }, 0);
    },

    añadirFila() {
      this.filas.push({ valor: '' });
      this.$nextTick(() => {
        const inputs = this.$el.querySelectorAll('input[type="text"]');
        if (inputs.length > 0) inputs[inputs.length - 1].focus();
      });
    },

    quitarFila(idx) {
      if (this.filas.length > 1) {
        this.filas.splice(idx, 1);
        this.calcular();
      }
    },

    limpiar() {
      this.filas = [{ valor: '' }];
      this.resultado = 0;
    },

    formatearResultado() {
      return this.resultado.toLocaleString('es-ES', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      });
    }
  }));

  // --- Exportar Excel ---
  Alpine.data('exportarExcel', () => ({
    modulo: 'expedientes',
    columnas: [],
    filtros: {},
    fechaDesde: '',
    fechaHasta: '',
    cargando: false,

    async cargarColumnas() {
      this.cargando = true;
      const contCols = this.$el.querySelector('#exp-columnas');
      const contFiltros = this.$el.querySelector('#exp-filtros-dinamicos');
      if (!contCols) { this.cargando = false; return; }
      contCols.innerHTML = '<p class="text-gray-500 text-xs col-span-full">Cargando...</p>';
      if (contFiltros) contFiltros.innerHTML = '';

      try {
        const res = await fetch('/api/columnas-modulo?modulo=' + encodeURIComponent(this.modulo));
        if (!res.ok) { contCols.innerHTML = '<p class="text-red-400 text-xs col-span-full">Error al cargar</p>'; return; }
        const data = await res.json();

        this.columnas = (data.view_cols || []).map(c => ({ nombre: c, seleccionada: true }));

        contCols.innerHTML = '';
        this.columnas.forEach(c => {
          const lbl = document.createElement('label');
          lbl.className = 'exp-checkbox-label flex items-center gap-2 text-sm cursor-pointer';
          lbl.innerHTML = `<input type="checkbox" x-model="columnas[${this.columnas.indexOf(c)}].seleccionada" class="exp-col"> ${c.nombre.replace(/_/g, ' ')}`;
          contCols.appendChild(lbl);
        });

        if (contFiltros) {
          contFiltros.innerHTML = '';
          const catalogs = window.PAGE_DATA?.catalogs || {};
          const catalogFilters = window.PAGE_DATA?.CatalogFilters || {};
          (data.table_cols || []).forEach(col => {
            const catInfo = catalogFilters[col];
            if (catInfo && catalogs[catInfo.key]) {
              const items = catalogs[catInfo.key];
              if (items.length > 0) {
                const div = document.createElement('div');
                div.className = 'flex flex-col';
                div.innerHTML = `<label class="label">${catInfo.label}</label>
                  <select x-model="filtros.${col}" name="${col}" class="input exp-filter-input">
                    <option value="">Todos</option>
                    ${items.map(i => `<option value="${i.id}"${i.id_gerencia ? ' data-id-gerencia="' + i.id_gerencia + '"' : ''}>${i.nombre}</option>`).join('')}
                  </select>`;
                contFiltros.appendChild(div);
              }
            }
          });
        }
      } catch(e) {
        contCols.innerHTML = '<p class="text-red-400 text-xs col-span-full">Error de conexión</p>';
      } finally {
        this.cargando = false;
      }
    },

    async exportar() {
      this.cargando = true;
      const spinner = document.getElementById('spinner-overlay');
      if (spinner) { document.getElementById('spinner-text').textContent = 'Generando Excel...'; spinner.classList.remove('hidden'); }

      try {
        const cols = this.columnas.filter(c => c.seleccionada).map(c => c.nombre);
        const params = new URLSearchParams();
        params.set('modulo', this.modulo);
        if (this.fechaDesde) params.set('fecha_desde', this.fechaDesde);
        if (this.fechaHasta) params.set('fecha_hasta', this.fechaHasta);
        if (cols.length) params.set('columnas', cols.join(','));
        Object.entries(this.filtros).forEach(([k, v]) => { if (v) params.set(k, v); });

        const res = await fetch('/api/exportar-excel?' + params.toString());
        if (!res.ok) {
          const text = await res.text();
          Alpine.store('toast').mostrar(text || 'Error al exportar', 'error');
          return;
        }

        const disposition = res.headers.get('Content-Disposition');
        let filename = 'reporte.xlsx';
        if (disposition && disposition.includes('attachment')) {
          const m = disposition.match(/filename=(.+)/);
          if (m) filename = m[1].replace(/['"]/g, '');
        }

        const blob = await res.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url; a.download = filename;
        document.body.appendChild(a); a.click(); a.remove();
        window.URL.revokeObjectURL(url);
        Alpine.store('toast').mostrar('Excel descargado con éxito', 'success');
      } catch(e) {
        Alpine.store('toast').mostrar('Error de conexión al exportar', 'error');
      } finally {
        if (spinner) spinner.classList.add('hidden');
        this.cargando = false;
      }
    }
  }));

  // --- Filtro Superintendencias ---
  Alpine.data('filtroSuperintendencias', (supOpts, gerInicial) => ({
    superintendencias: supOpts || [],
    gerenciaSeleccionada: gerInicial || '',
    superintendenciaSeleccionada: '',

    init() {
      this.$watch('gerenciaSeleccionada', () => {
        this.superintendenciaSeleccionada = '';
      });
    },

    get superintendenciasFiltradas() {
      if (!this.gerenciaSeleccionada) return this.superintendencias;
      return this.superintendencias.filter(s =>
        String(s.id_gerencia) === String(this.gerenciaSeleccionada)
      );
    }
  }));

  // --- Formulario Módulo (conversión USD/Bs) ---
  Alpine.data('formularioModulo', (modulo, registroInicial, tipoCambioInicial) => ({
    modulo: modulo || 'expedientes',
    registro: registroInicial || {},
    tipoCambio: tipoCambioInicial || 0,
    convLock: false,

    init() {
      if (this.registro && this.registro.tipo_cambio) {
        this.tipoCambio = this._parseValue(String(this.registro.tipo_cambio));
      }
    },

    convertirMoneda(origen) {
      if (this.convLock || !this.tipoCambio) return;
      this.convLock = true;
      try {
        const getRaw = (key) => this._parseValue(String(this.registro[key] || ''));
        const setVal = (key, val) => { this.registro[key] = val ? val.toFixed(2) : ''; };

        if (origen === 'bs_presup') {
          const bs = getRaw('presupuesto_base_bs');
          if (bs) setVal('presupuesto_base_usd', bs / this.tipoCambio);
        } else if (origen === 'usd_adj') {
          const usd = getRaw('monto_adjudicado_usd');
          if (usd) setVal('monto_adjudicado_bs', usd * this.tipoCambio);
        } else if (origen === 'bs_adj') {
          const bs = getRaw('monto_adjudicado_bs');
          if (bs) setVal('monto_adjudicado_usd', bs / this.tipoCambio);
        } else {
          const usd = getRaw('presupuesto_base_usd');
          if (usd) setVal('presupuesto_base_bs', usd * this.tipoCambio);
        }
      } finally {
        this.convLock = false;
      }
    },

    _parseValue(v) {
      if (!v) return 0;
      v = String(v).replace(/,/g, '.').replace(/[^0-9.]/g, '');
      const n = parseFloat(v);
      return isNaN(n) ? 0 : n;
    }
  }));
});
