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
      el.className = 'ui-toast ' + tipo;
      const icons = { info: 'fa-info-circle', success: 'fa-check-circle', error: 'fa-exclamation-circle', warning: 'fa-exclamation-triangle' };
      el.innerHTML = '<i class="fas ' + (icons[tipo] || 'fa-info-circle') + '"></i> ' + msg;
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
  Alpine.store('fijados', {
    lista: [],

    init() {
      const stored = localStorage.getItem('sidebarFrecuentes');
      try { this.lista = JSON.parse(stored || '[]'); } catch (e) { this.lista = []; }
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
  });
  Alpine.data('fijados', () => ({
    get lista() { return Alpine.store('fijados').lista; },

    toggle(id, solped, modulo) {
      Alpine.store('fijados').toggle(id, solped, modulo);
    },

    estaFijado(id, modulo) {
      return Alpine.store('fijados').estaFijado(id, modulo);
    },

    eliminar(idx) {
      Alpine.store('fijados').eliminar(idx);
    }
  }));

  // --- BD Recientes ---
  Alpine.data('bdRecientes', () => ({
    lista: [],

    init() {
      this.cargar();
      if (window.PAGE_DATA && window.PAGE_DATA.hasDB && window.PAGE_DATA.dbPath) {
        const path = window.PAGE_DATA.dbPath;
        const nombre = path.split('/').pop().split('\\').pop() || path;
        const idx = this.lista.findIndex(r => r.path === path);
        if (idx >= 0) this.lista.splice(idx, 1);
        this.lista.unshift({ nombre, path, timestamp: Date.now() });
        if (this.lista.length > 5) this.lista = this.lista.slice(0, 5);
        this.guardar();
      } else if (this.lista.length === 1 && !(window.PAGE_DATA && window.PAGE_DATA.hasDB)) {
        this.abrir(this.lista[0].path);
      }
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
    fijados: [],

    onInput(idx, event) {
      const input = event.target;
      let raw = input.value;

      raw = raw.replace(/[^0-9.,-]/g, '');
      const lastComma = raw.lastIndexOf(',');
      const lastDot = raw.lastIndexOf('.');
      let intRaw, decRaw, sep;
      if (lastComma > lastDot) {
        sep = ',';
        intRaw = raw.slice(0, lastComma);
        decRaw = raw.slice(lastComma + 1);
      } else if (lastDot > lastComma) {
        sep = '.';
        intRaw = raw.slice(0, lastDot);
        decRaw = raw.slice(lastDot + 1);
      } else {
        intRaw = raw;
        decRaw = '';
        sep = '';
      }
      intRaw = intRaw.replace(/[.,]/g, '');
      let isNeg = intRaw.startsWith('-') ? '-' : '';
      intRaw = intRaw.replace(/-/g, '').replace(/\D/g, '');
      decRaw = decRaw.replace(/\D/g, '').slice(0, 2);

      raw = sep ? isNeg + intRaw + sep + decRaw : isNeg + intRaw;
      input.value = raw;
      this.filas[idx].valor = raw;
      this.calcular();
    },

    onBlur(idx, event) {
      const raw = event.target.value;
      const clean = raw.replace(/\./g, '').replace(',', '.');
      const num = parseFloat(clean);
      if (!isNaN(num) && isFinite(num)) {
        const formatted = num.toLocaleString('es-ES', { minimumFractionDigits: 0, maximumFractionDigits: 2 });
        event.target.value = formatted;
        this.filas[idx].valor = formatted;
        this.calcular();
      }
    },

    calcular() {
      this.resultado = this.filas.reduce((sum, f) => {
        const clean = String(f.valor || '').replace(/\./g, '').replace(',', '.');
        const v = parseFloat(clean) || 0;
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

    fijarResultado() {
      if (!this.resultado) return;
      this.fijados.push(this.resultado);
    },

    quitarFijado(idx) {
      this.fijados.splice(idx, 1);
    },

    get totalFijados() {
      return this.fijados.reduce((s, v) => s + v, 0);
    },

    formatearNum(v) {
      return v.toLocaleString('es-ES', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      });
    },

    formatearResultado() {
      return this.formatearNum(this.resultado);
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

    init() {
      this.$watch('modulo', () => this.cargarColumnas());
      this._prevOpen = false;
      this.$watch('$store.modals.stack', () => {
        const open = Alpine.store('modals').tiene('export-modal');
        if (open && !this._prevOpen) this.cargarColumnas();
        this._prevOpen = open;
      });
    },

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
          const cb = document.createElement('input');
          cb.type = 'checkbox';
          cb.className = 'exp-col';
          cb.checked = c.seleccionada;
          cb.addEventListener('change', () => { c.seleccionada = cb.checked; });
          lbl.appendChild(cb);
          lbl.appendChild(document.createTextNode(' ' + c.nombre.replace(/_/g, ' ')));
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
                const label = document.createElement('label');
                label.className = 'label';
                label.textContent = catInfo.label;
                div.appendChild(label);
                const select = document.createElement('select');
                select.className = 'input exp-filter-input';
                select.name = col;
                const optAll = document.createElement('option');
                optAll.value = '';
                optAll.textContent = 'Todos';
                select.appendChild(optAll);
                items.forEach(i => {
                  const opt = document.createElement('option');
                  opt.value = String(i.id);
                  if (i.id_gerencia) opt.dataset.idGerencia = String(i.id_gerencia);
                  opt.textContent = i.nombre;
                  select.appendChild(opt);
                });
                select.addEventListener('change', () => { this.filtros[col] = select.value; });
                div.appendChild(select);
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

    validarFechas() {
      if (this.fechaDesde && this.fechaHasta && this.fechaDesde > this.fechaHasta) {
        Alpine.store('toast').mostrar('La fecha desde no puede ser mayor que la fecha hasta', 'error');
        this.fechaHasta = '';
      }
    },

    async exportar() {
      if (this.fechaDesde && this.fechaHasta && this.fechaDesde > this.fechaHasta) {
        Alpine.store('toast').mostrar('La fecha desde no puede ser mayor que la fecha hasta', 'error');
        this.cargando = false;
        return;
      }
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
  Alpine.data('formularioModulo', (modulo, registroInicial) => ({
    modulo: modulo || 'expedientes',
    registro: registroInicial || {},
    autoObs: '',
    autoHistorial: '',
    lastSource: {},
    _obsReady: false,
    _pendingObs: {},
    ordenExcel: false,
    _fieldStructure: null,
    history: [],
    historyIndex: -1,
    _historyTimer: null,

    init() {
      if (!this.registro.id_estatus) this.registro.id_estatus = '1';
      const obs = String(this.registro.observaciones || '');
      const sepIdx = obs.indexOf('\n---\n');
      if (sepIdx >= 0) {
        this.autoHistorial = obs.slice(0, sepIdx);
        this.registro.observaciones = obs.slice(sepIdx + 5);
      }
      this._trackInicial = {
        id_documento: this.registro.id_documento,
        fecha_recibido: this.registro.fecha_recibido,
        fecha_devuelto: this.registro.fecha_devuelto,
        id_estatus: this.registro.id_estatus
      };
      this._pushSnapshot();
      this._obsReady = true;
      this.$watch('registro', () => {
        clearTimeout(this._historyTimer);
        this._historyTimer = setTimeout(() => this._pushSnapshot(), 500);
      }, { deep: true });
      this.$watch('registro.tipo_cambio', () => this._syncAll());
      this.$watch('registro.id_documento', (v, old) => this._obsCambio('Documento', old, v));
      this.$watch('registro.fecha_recibido', (v, old) => this._obsCambio('Fecha Recibido', old, v));
      this.$watch('registro.fecha_devuelto', (v, old) => this._obsCambio('Fecha Devuelto', old, v));
      this.$watch('registro.id_estatus', (v, old) => this._obsCambio('Estatus', old, v));
    },

    get puedeDeshacer() { return this.historyIndex > 0; },
    get puedeRehacer() { return this.historyIndex < this.history.length - 1; },

    deshacer() {
      if (!this.puedeDeshacer) return;
      this.historyIndex--;
      Object.assign(this.registro, JSON.parse(JSON.stringify(this.history[this.historyIndex])));
    },

    rehacer() {
      if (!this.puedeRehacer) return;
      this.historyIndex++;
      Object.assign(this.registro, JSON.parse(JSON.stringify(this.history[this.historyIndex])));
    },

    _pushSnapshot() {
      this.history = this.history.slice(0, this.historyIndex + 1);
      this.history.push(JSON.parse(JSON.stringify(this.registro)));
      if (this.history.length > 50) this.history.shift();
      this.historyIndex = this.history.length - 1;
    },

    _obsLineaCompleta() {
      const doc = this._obsLabel('Documento', String(this.registro.id_documento || ''));
      const est = this._obsLabel('Estatus', String(this.registro.id_estatus || ''));
      const dev = this.registro.fecha_devuelto;
      const rec = this.registro.fecha_recibido;
      const parts = [];
      if (dev) parts.push(`Fecha Devuelto: ${dev}`);
      else if (rec) parts.push(`Fecha Recibido: ${rec}`);
      if (doc) parts.push(`Documento: ${doc}`);
      if (est) parts.push(`Estatus: ${est}`);
      return parts.length ? parts.join(', ') : '';
    },

    toggleOrden() {
      this.ordenExcel = !this.ordenExcel;
      this.$nextTick(() => this._reordenar());
    },

    _reordenar() {
      console.log('[OrdenExcel] this.$el=', this.$el ? this.$el.tagName + '.' + this.$el.className : 'null');

      const container = document.getElementById('excel-order-container');
      if (!container) {
        console.warn('[OrdenExcel] #excel-order-container no encontrado en documento');
        return;
      }
      console.log('[OrdenExcel] container encontrado, display=' + container.style.display);

      const allFields = Array.from(document.querySelectorAll('[data-orden-excel]'));
      console.log('[OrdenExcel] estado=' + this.ordenExcel + ', campos=' + allFields.length);
      if (allFields.length === 0) {
        console.warn('[OrdenExcel] No hay campos con data-orden-excel');
        return;
      }

      if (this.ordenExcel) {
        if (!this._fieldStructure) {
          this._fieldStructure = [];
          document.querySelectorAll('[data-orden-excel]').forEach(el => {
            this._fieldStructure.push({
              el: el,
              parent: el.parentNode
            });
          });
          console.log('[OrdenExcel] estructura guardada, ' + this._fieldStructure.length + ' campos');
        }

        document.querySelectorAll('fieldset').forEach(fs => fs.style.display = 'none');

        const fields = Array.from(document.querySelectorAll('[data-orden-excel]'));
        fields.sort((a, b) => parseInt(a.dataset.ordenExcel) - parseInt(b.dataset.ordenExcel));
        console.log('[OrdenExcel] ordenando ' + fields.length + ' campos por data-orden-excel');

        container.innerHTML = '';
        fields.forEach(f => container.appendChild(f));
        container.style.display = '';
        console.log('[OrdenExcel] campos movidos al contenedor plano');
      } else {
        container.style.display = 'none';
        if (this._fieldStructure) {
          this._fieldStructure.forEach(item => {
            item.parent.appendChild(item.el);
          });
          console.log('[OrdenExcel] campos restaurados a fieldsets originales');
        }
        document.querySelectorAll('fieldset').forEach(fs => fs.style.display = '');
      }
    },

    _obsCambio(label, oldVal, newVal) {
      if (!this._obsReady) return;
      const old = String(oldVal || '').trim();
      const nw = String(newVal || '').trim();
      if (old === nw) return;
      this._pendingObs[label] = true;
      const linea = this._obsLineaCompleta();
      if (!linea) return;
      this.autoObs = linea;
    },

    _obsLabel(label, val) {
      if (!val) return '(vacío)';
      const cats = window.PAGE_DATA?.catalogs || {};
      if (label === 'Documento') {
        const item = (cats.documento || []).find(i => i.id == val);
        if (item) return item.nombre;
      }
      if (label === 'Estatus') {
        const item = (cats.estatus_detalle || []).find(i => i.id == val);
        if (item) return item.nombre;
      }
      return val;
    },

    prepararObservaciones() {
      const manual = String(this.registro.observaciones || '').trim();
      const auto = this.autoObs;
      const combinado = auto ? (manual ? auto + '\n---\n' + manual : auto) : manual;
      this.registro.observaciones = combinado;
      const el = document.getElementById('f-observaciones');
      if (el) el.value = combinado;
      this._pendingObs = {};
      this.autoHistorial = this.autoObs;
      this.autoObs = '';
    },

    _syncAll() {
      const tc = this._parseValue(this.registro.tipo_cambio);
      if (!tc || !isFinite(tc)) return;
      this._syncPair('presupuesto_base_bs', 'presupuesto_base_usd', tc);
      this._syncPair('monto_adjudicado_bs', 'monto_adjudicado_usd', tc);
    },

    _syncPair(bsKey, usdKey, tc) {
      const bs = this._parseValue(this.registro[bsKey]);
      const usd = this._parseValue(this.registro[usdKey]);
      const source = this.lastSource[bsKey + '_' + usdKey];
      if (source === bsKey && bs) {
        this._setVal(usdKey, bs / tc);
      } else if (source === usdKey && usd) {
        this._setVal(bsKey, usd * tc);
      } else if (bs && !usd) {
        this._setVal(usdKey, bs / tc);
      } else if (usd && !bs) {
        this._setVal(bsKey, usd * tc);
      }
    },

    onMontoInput(event, origen) {
      const input = event.target;
      let raw = input.value;
      raw = raw.replace(/[^0-9.,-]/g, '');
      const lastComma = raw.lastIndexOf(',');
      const lastDot = raw.lastIndexOf('.');
      let sep, intRaw, decRaw;
      if (lastComma > lastDot) { sep = ','; intRaw = raw.slice(0, lastComma); decRaw = raw.slice(lastComma + 1); }
      else if (lastDot > lastComma) { sep = '.'; intRaw = raw.slice(0, lastDot); decRaw = raw.slice(lastDot + 1); }
      else { sep = ''; intRaw = raw; decRaw = ''; }
      intRaw = intRaw.replace(/[.,-]/g, '').replace(/\D/g, '');
      decRaw = decRaw.replace(/\D/g, '').slice(0, 2);
      raw = sep ? intRaw + sep + decRaw : intRaw;
      input.value = raw;
      input.dataset.raw = raw;

      const map = {
        'bs_presup': ['presupuesto_base_bs', 'presupuesto_base_usd'],
        'usd_presup': ['presupuesto_base_usd', 'presupuesto_base_bs'],
        'bs_adj': ['monto_adjudicado_bs', 'monto_adjudicado_usd'],
        'usd_adj': ['monto_adjudicado_usd', 'monto_adjudicado_bs']
      };
      const [origenKey, destinoKey] = map[origen] || [];
      if (!origenKey) return;
      this.lastSource[origenKey + '_' + destinoKey] = origenKey;
      this.registro[origenKey] = raw;
      this._conv(origenKey, destinoKey);
    },

    onMontoBlur(event, key) {
      const raw = event.target.value;
      const num = this._parseValue(raw);
      if (!num) return;
      const formatted = num.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
      event.target.value = formatted;
      event.target.dataset.raw = raw;
      this.registro[key] = raw;
    },

    onMontoFocus(event) {
      if (event.target.dataset.raw !== undefined) {
        event.target.value = event.target.dataset.raw;
      }
    },

    _conv(origenKey, destinoKey) {
      const tc = this._parseValue(this.registro.tipo_cambio);
      if (!tc || !isFinite(tc)) return;
      const val = this._parseValue(this.registro[origenKey]);
      if (!val) return;
      const isUsdOrigen = origenKey.includes('_usd');
      const res = isUsdOrigen ? val * tc : val / tc;
      this._setVal(destinoKey, res);
    },

    _setVal(key, val) {
      if (!isFinite(val)) return;
      const str = val.toFixed(2);
      this.registro[key] = str;
      this.$nextTick(() => {
        const el = document.getElementById('f-' + key);
        if (el) {
          el.dataset.raw = str;
          const num = parseFloat(str);
          if (num) el.value = num.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
        }
      });
    },

    validarFechas() {
      const pares = [
        ['fecha_recibido', 'fecha_devuelto', 'Recibido', 'Devuelto'],
        ['fecha_desde', 'fecha_hasta', 'Desde', 'Hasta'],
        ['fecha_inicio', 'fecha_final', 'Inicio', 'Final'],
        ['periodo_valuacion_desde', 'periodo_valuacion_hasta', 'Período Desde', 'Período Hasta']
      ];
      pares.forEach(([desde, hasta, labelDesde, labelHasta]) => {
        const vDesde = this.registro[desde];
        const vHasta = this.registro[hasta];
        if (vDesde && vHasta && vDesde > vHasta) {
          Alpine.store('toast').mostrar(`La fecha ${labelDesde} no puede ser mayor que la fecha ${labelHasta}`, 'error');
          this.registro[hasta] = '';
          const el = document.getElementById('f-' + hasta);
          if (el) el.value = '';
        }
      });
    },

    validarAntesGuardar() {
      let ok = true;
      const pares = [
        ['fecha_recibido', 'fecha_devuelto', 'Recibido', 'Devuelto'],
        ['fecha_desde', 'fecha_hasta', 'Desde', 'Hasta'],
        ['fecha_inicio', 'fecha_final', 'Inicio', 'Final'],
        ['periodo_valuacion_desde', 'periodo_valuacion_hasta', 'Período Desde', 'Período Hasta']
      ];
      pares.forEach(([desde, hasta, labelDesde, labelHasta]) => {
        const vDesde = this.registro[desde];
        const vHasta = this.registro[hasta];
        if (vDesde && vHasta && vDesde > vHasta) {
          Alpine.store('toast').mostrar(`La fecha ${labelDesde} no puede ser mayor que la fecha ${labelHasta}`, 'error');
          ok = false;
        }
      });
      return ok;
    },

    appendDias() {
      const v = String(this.registro.tiempo_ejecucion || '').trim();
      if (v && !v.toUpperCase().endsWith('DÍAS') && !v.toUpperCase().endsWith('DIAS')) {
        const nuevo = v + ' DÍAS';
        this.registro.tiempo_ejecucion = nuevo;
        const el = document.getElementById('f-tiempo_ejecucion');
        if (el) el.value = nuevo;
      }
    },

    spinFrente(delta) {
      const actual = parseInt(this.registro.cantidad_frentes || '0', 10);
      const nuevo = Math.max(0, actual + delta);
      this.registro.cantidad_frentes = String(nuevo);
    },

    _parseValue(v) {
      if (!v) return 0;
      v = String(v).trim();
      const hasComma = v.includes(',');
      const hasDot = v.includes('.');
      if (hasComma && hasDot) {
        const lastComma = v.lastIndexOf(',');
        const lastDot = v.lastIndexOf('.');
        if (lastComma > lastDot) {
          v = v.replace(/\./g, '').replace(',', '.');
        } else {
          v = v.replace(/,/g, '').replace(/\.(?=.*\.)/, '');
        }
      } else if (hasComma && !hasDot) {
        v = v.replace(',', '.');
      } else if (!hasComma && hasDot) {
        if ((v.match(/\./g) || []).length > 1) {
          const lastDot = v.lastIndexOf('.');
          v = v.slice(0, lastDot).replace(/\./g, '') + v.slice(lastDot);
        }
      }
      v = v.replace(/[^0-9.]/g, '');
      const n = parseFloat(v);
      return isNaN(n) ? 0 : n;
    }
  }));
});
