--- plan_migracion_alpine.md (原始)


+++ plan_migracion_alpine.md (修改后)
# Plan de Migración a Alpine.js

**Objetivo:** Reemplazar JS vainilla con Alpine.js para estado UI local, manteniendo HTMX para comunicación servidor. **Cero gluecode** — solo atributos HTML y definiciones `Alpine.data()`.

**Principios:**
- Alpine maneja estado UI local (modales, localStorage, toggles)
- HTMX maneja toda comunicación servidor (fetch, POST, GET)
- JS residual solo para lógica que Alpine no puede expresar (ej: Gantt canvas, dialogs nativos)
- No modificar Go backend

---

## Fase 1: Modales (Stack y Backdrop)

### Funciones a migrar
- `pushModal(id)` (app.js:102)
- `cerrarModal(id)` (app.js:118)
- `cerrarSiOverlay` (click en backdrop)
- `mostrarFormulario`, `cerrarFormulario`, `cerrarHistorial`, `cerrarRuta`, `cerrarPendientes`, `cerrarRecientes`, `cerrarFrecuentes`, `cerrarSumas`, `cerrarModalExportar`

### Patrón Alpine

**HTML actual (index.html):**
```html
<div id="form-modal" class="modal hidden fixed inset-0 ...">
    <button onclick="cerrarFormulario()">Cerrar</button>
</div>
```

**HTML migrado:**
```html
<div x-data="{ modals: [] }"
     x-on:keydown.escape.window="modals.length && modals.pop()"
     :class="modals.length ? 'overflow-hidden' : ''">

    <!-- Modal individual -->
    <div id="form-modal"
         class="modal fixed inset-0 ..."
         :class="modals.includes('form-modal') ? '' : 'hidden'"
         @click.self="modals = modals.filter(m => m !== 'form-modal')">
        <button @click="modals = modals.filter(m => m !== 'form-modal')">Cerrar</button>
    </div>

    <!-- Otros modales con mismo patrón -->
</div>
```

**Para abrir modal desde HTMX:**
```html
<button hx-get="/api/cargar-expediente"
        hx-target="#form-expediente"
        @htmx:after-request="$dispatch('abrir-modal', 'form-modal')"
        class="btn btn-primary">
    Nuevo Registro
</button>

<script>
document.addEventListener('abrir-modal', e => {
    // Acceder al scope Alpine más cercano
    document.querySelector('[x-data]').__x.$data.modals.push(e.detail);
});
</script>
```

**Alternativa más simple (sin JS):**
```html
<!-- Botón que abre modal directamente con Alpine -->
<button @click="$dispatch('abrir-modal', 'form-modal')"
        hx-get="/api/cargar-expediente"
        hx-target="#form-expediente"
        class="btn btn-primary">
    Nuevo Registro
</button>

<!-- Escuchar evento HTMX y abrir modal -->
<div x-data="{ modalAbierto: false }"
     @abrir-modal.window="modalAbierto = true; $nextTick(() => $el.querySelector('.modal').classList.remove('hidden'))"
     @cerrar-modal.window="modalAbierto = false">
```

**Mejor enfoque - Alpine.data reusable:**
```js
// En frontend/vendor/alpine-app.js (nuevo archivo mínimo)
document.addEventListener('alpine:init', () => {
    Alpine.data('modalStack', () => ({
        stack: [],
        abrir(id) { this.stack.push(id); },
        cerrar(id) { this.stack = this.stack.filter(m => m !== id); },
        tiene(id) { return this.stack.includes(id); }
    }));

    Alpine.data('modalItem', (id) => ({
        init() {
            this.$watch('$el.classList.contains("hidden")', value => {
                if (!value && !modalStack().stack.includes(id)) {
                    modalStack().stack.push(id);
                }
            });
        }
    }));
});
```

### Estado reemplazado
- Elimina `MODAL_STACK` (array global)
- Elimina `document.body.style.overflow` manipulación manual
- Elimina listener global de click en modales

### JS residual
**Ninguno** para modales básicos.

Para integración con HTMX, se necesita un pequeño puente:
```js
// frontend/vendor/alpine-htmx-bridge.js (opcional, ~20 líneas)
document.addEventListener('htmx:afterRequest', function(evt) {
    const abrirId = evt.detail.target.getAttribute('hx-on::after-request');
    if (abrirId && abrirId.includes('pushModal')) {
        const match = abrirId.match(/pushModal\('([^']+)'\)/);
        if (match && window.Alpine) {
            window.Alpine.store('modals').abrir(match[1]);
        }
    }
});
```

### Archivos a modificar
- `templates/index.html` — agregar `x-data` en body o contenedor principal
- `templates/index.html` — modificar cada modal para usar `x-show` o `:class`
- `frontend/vendor/alpine.min.js` — copiar Alpine.js (ya existe o crear)
- Opcional: `frontend/vendor/alpine-htmx-bridge.js` — puente mínimo

---

## Fase 2: Pines / Fijados (Acceso Rápido)

### Funciones a migrar
- `toggleFrecuente(id, solped, modulo)` (app.js:263)
- `abrirFrecuentes()` (app.js:293)
- `cerrarFrecuentes()` (app.js:328)
- `inicializarPines()` (app.js:330)

### Patrón Alpine

**Alpine.data definition:**
```js
Alpine.data('fijados', () => ({
    lista: [],

    init() {
        this.cargar();
        this.$watch('lista', () => this.guardar(), { deep: true });
    },

    cargar() {
        try {
            this.lista = JSON.parse(localStorage.getItem('sidebarFrecuentes') || '[]');
        } catch(e) { this.lista = []; }
    },

    guardar() {
        localStorage.setItem('sidebarFrecuentes', JSON.stringify(this.lista));
    },

    toggle(id, solped, modulo = 'expedientes') {
        const idx = this.lista.findIndex(i => i.id === id && i.modulo === modulo);
        if (idx !== -1) {
            this.lista.splice(idx, 1);
        } else {
            this.lista.push({ id, solped: solped || '', modulo });
        }
    },

    estaFijado(id, modulo = 'expedientes') {
        return this.lista.some(i => i.id === id && i.modulo === modulo);
    },

    eliminar(index) {
        this.lista.splice(index, 1);
    }
}));
```

**HTML para botón pin en tabla:**
```html
<!-- En tabla_expedientes.html unificado -->
<button @click="$dispatch('toggle-fijado', { id: {{rowGetStr $exp "id_expediente"}}, solped: '{{rowGetStr $exp "solped"}}' })"
        class="btn-icon"
        :class="$refs.fijados?.estaFijado({{rowGetStr $exp "id_expediente"}}) ? 'text-emerald-500' : 'text-blue-500'"
        title="Anclar">
    <i class="fas fa-thumbtack rotate-45 text-xs"></i>
</button>
```

**Modal de fijados:**
```html
<div id="modal-frecuentes"
     x-show="modales.includes('modal-frecuentes')"
     @abrir-frecuentes.window="$store.modales.abrir('modal-frecuentes')">

    <div x-data="fijados" class="modal-body">
        <template x-if="lista.length === 0">
            <p class="text-gray-500 text-center">Sin expedientes fijados.</p>
        </template>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3" x-show="lista.length > 0">
            <template x-for="(item, idx) in lista" :key="item.id">
                <div @click="$dispatch('cerrar-frecuentes'); $dispatch('abrir-formulario', item)"
                     class="flex items-center justify-between bg-gray-900/60 ...">
                    <div>
                        <div class="text-[10px] text-gray-500 uppercase">SOLPED</div>
                        <div class="text-sm text-gray-200 font-bold" x-text="item.solped || 'SIN SOLPED'"></div>
                    </div>
                    <button @click.stop="eliminar(idx)"
                            class="text-gray-500 hover:text-red-400">
                        <i class="fas fa-times text-xs"></i>
                    </button>
                </div>
            </template>
        </div>
    </div>
</div>
```

**Botón para abrir modal:**
```html
<button @click="$dispatch('abrir-frecuentes')"
        class="btn btn-primary"
        :disabled="!hasDB">
    <i class="fas fa-thumbtack rotate-45 mr-1"></i> Fijados
</button>
```

### Estado reemplazado
- Elimina `localStorage.sidebarFrecuentes` acceso manual
- Elimina `inicializarPines()` que recorre DOM
- El color del pin se calcula reactivamente con `:class`

### JS residual
**Ninguno.** Todo el estado y lógica cabe en `Alpine.data('fijados')`.

### Archivos a modificar
- `templates/index.html` — agregar Alpine.data en script inicial
- `templates/tabla_expedientes.html` (y otras) — modificar botón pin
- `templates/index.html` — modificar modal de fijados

---

## Fase 3: BD Recientes

### Funciones a migrar
- `registrarReciente(nombre, path)` (app.js:184)
- `abrirRecientes()` (app.js:237)
- `cerrarRecientes()` (app.js:260)
- `eliminarReciente(path)` (app.js:197)
- `eliminarRecienteIndex(index)` (app.js:205)
- `abrirBaseDatosReciente(path)` (app.js:215)

### Patrón Alpine

**Alpine.data definition:**
```js
Alpine.data('bdRecientes', () => ({
    lista: [],

    init() {
        this.cargar();
        this.$watch('lista', () => this.guardar(), { deep: true });
    },

    cargar() {
        try {
            this.lista = JSON.parse(localStorage.getItem('baseaccess_recientes') || '[]');
        } catch(e) { this.lista = []; }
    },

    guardar() {
        localStorage.setItem('baseaccess_recientes', JSON.stringify(this.lista));
    },

    registrar(nombre, path) {
        const idx = this.lista.findIndex(r => r.path === path);
        if (idx !== -1) this.lista.splice(idx, 1);
        this.lista.unshift({ nombre, path, timestamp: Date.now() });
        if (this.lista.length > 5) this.lista = this.lista.slice(0, 5);
    },

    eliminarPorPath(path) {
        this.lista = this.lista.filter(r => r.path !== path);
    },

    eliminarPorIndex(index) {
        this.lista.splice(index, 1);
    },

    abrir(path) {
        fetch('/api/abrir-bd', {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            body: 'path=' + encodeURIComponent(path)
        }).then(res => {
            if (res.ok) location.reload();
            else {
                res.text().then(err => {
                    alert('Error: ' + err);
                    this.eliminarPorPath(path);
                });
            }
        }).catch(err => {
            alert('Error: ' + err);
            this.eliminarPorPath(path);
        });
    }
}));
```

**Registro automático al cargar página:**
```html
<script>
document.addEventListener('DOMContentLoaded', () => {
    if (window.PAGE_DATA?.hasDB && window.PAGE_DATA?.dbPath) {
        const recientesEl = document.querySelector('[x-data="bdRecientes"]');
        if (recientesEl && recientesEl.__x) {
            recientesEl.__x.data.registrar(
                window.PAGE_DATA.dbPath.split('/').pop(),
                window.PAGE_DATA.dbPath
            );
        }
    }
});
</script>
```

**Modal de recientes:**
```html
<div id="modal-recientes" x-show="modales.includes('modal-recientes')">
    <div x-data="bdRecientes" class="modal-body">
        <template x-if="lista.length === 0">
            <p class="text-gray-500 text-center py-8 italic">Sin bases de datos recientes.</p>
        </template>

        <div class="flex flex-col gap-2" x-show="lista.length > 0">
            <template x-for="(r, i) in lista" :key="r.path">
                <div @click="abrir(r.path)"
                     class="reciente-item flex items-center gap-2 px-4 py-3 cursor-pointer ...">
                    <div class="min-w-0 flex-1">
                        <div class="text-sm text-gray-200 truncate font-semibold" x-text="r.nombre"></div>
                        <div class="text-xs text-gray-500 truncate" x-text="r.path"></div>
                    </div>
                    <button @click.stop="eliminarPorIndex(i)"
                            class="text-xs text-gray-400 hover:text-red-400">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </template>
        </div>
    </div>
</div>
```

### Estado reemplazado
- Elimina `localStorage.baseaccess_recientes` acceso manual
- Elimina funciones de renderizado HTML manual

### JS residual
**Mínimo:** El registro automático al cargar página requiere un pequeño script porque `PAGE_DATA` viene de Go. Alternativamente, se puede hacer desde el backend enviando un evento HTMX.

### Archivos a modificar
- `templates/index.html` — agregar Alpine.data y modificar modal
- `templates/index.html` — script de registro automático

---

## Fase 4: Sumas (Calculadora)

### Funciones a migrar
- `abrirSumas()` (app.js:639)
- `cerrarSumas()` (app.js:642)
- `anyadirFilaSuma()` (app.js:646)
- `calcularSumas()` (app.js:650)
- `limpiarSumas()` (app.js:655)

### Patrón Alpine

**Alpine.data definition:**
```js
Alpine.data('calculadoraSumas', () => ({
    filas: [{ valor: '' }],
    resultado: 0,

    calcular() {
        this.resultado = this.filas.reduce((sum, f) => {
            const val = parseFloat(f.valor.replace(',', '.')) || 0;
            return sum + val;
        }, 0);
    },

    añadirFila() {
        this.filas.push({ valor: '' });
        this.$nextTick(() => {
            this.$el.querySelector('input:last-of-type').focus();
        });
    },

    quitarFila(index) {
        if (this.filas.length > 1) {
            this.filas.splice(index, 1);
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
```

**HTML del modal:**
```html
<div id="sumas-modal" x-show="modales.includes('sumas-modal')">
    <div x-data="calculadoraSumas"
         @abrir-sumas.window="$store.modales.abrir('sumas-modal')"
         class="modal-body space-y-6">

        <div class="space-y-3">
            <template x-for="(fila, idx) in filas" :key="idx">
                <div class="suma-fila flex items-center gap-3 w-full">
                    <input type="text"
                           inputmode="decimal"
                           placeholder="0,00"
                           class="input flex-1"
                           x-model="fila.valor"
                           @input="calcular"
                           @keydown.enter.prevent="añadirFila()">
                    <button @click="quitarFila(idx)"
                            class="w-8 h-8 flex items-center justify-center text-red-400 hover:text-red-300">
                        <i class="fas fa-times text-sm"></i>
                    </button>
                </div>
            </template>
        </div>

        <button @click="añadirFila()" class="btn btn-secondary">
            <i class="fas fa-plus mr-1"></i> Añadir número
        </button>

        <div class="border-t border-gray-700 pt-4">
            <div class="flex justify-between items-center text-lg font-bold">
                <span class="text-gray-300">Resultado:</span>
                <span class="text-teal-400" x-text="formatearResultado()"></span>
            </div>
        </div>

        <div class="sticky bottom-0 bg-gray-800 ... flex justify-end gap-3">
            <button @click="limpiar()" class="btn btn-secondary">
                <i class="fas fa-eraser mr-1"></i> Limpiar todo
            </button>
            <button @click="$store.modales.cerrar('sumas-modal')" class="btn btn-primary">
                Cerrar
            </button>
        </div>
    </div>
</div>
```

### Estado reemplazado
- Elimina `_initNumInput`, `_fmtNum`, `_parseValue` para este caso específico
- Elimina manipulación manual del DOM para añadir/quitar filas

### JS residual
**Ninguno.** La calculadora es pura reactividad Alpine.

### Archivos a modificar
- `templates/index.html` — modificar modal de sumas

---

## Fase 5: Paginación Cliente (EVALUAR)

### Funciones a migrar
- `irPagina(pagina)` (app.js:353)
- `aplicarPaginacionDOM()` (app.js:358)
- `renderPaginacionControles(totalPages)` (app.js:387)

### Recomendación: **NO MIGRAR A ALPINE**

**Razón:** La paginación cliente actual opera sobre filas ya renderizadas por HTMX/Go. Migrar a Alpine requeriría:
1. Que Go envíe TODAS las filas en JSON
2. Alpine renderice la tabla completa
3. Alpine filtre/muestre según página

Esto contradice el principio de HTMX (renderizado servidor). Mejor opción:

**Opción A (Recomendada):** Mantener paginación servidor con HTMX
- Go ya soporta paginación (`currentPage`, `totalPages` en PAGE_DATA)
- HTMX puede hacer fetch de páginas específicas
- Eliminar paginación cliente completamente

**Opción B:** Si se requiere paginación cliente por performance, migrar pero con advertencia de complejidad.

### Si se decide migrar (Opción B):

**Alpine.data definition:**
```js
Alpine.data('tablaPaginada', (filasIniciales) => ({
    filas: filasIniciales,
    paginaActual: 1,
    pageSize: 10,

    get totalPages() {
        return Math.ceil(this.filas.length / this.pageSize);
    },

    get filasPagina() {
        const start = (this.paginaActual - 1) * this.pageSize;
        return this.filas.slice(start, start + this.pageSize);
    },

    irPagina(p) {
        if (p >= 1 && p <= this.totalPages) {
            this.paginaActual = p;
        }
    },

    get rangosPagina() {
        // Lógica compleja para mostrar ... entre páginas
        // Similar a renderPaginacionControles
    }
}));
```

### Estado reemplazado
- Elimina `currentPage` global
- Elimina `CONFIG.pageSize`
- Elimina manipulación directa de `style.display` en filas

### JS residual
**Posiblemente ninguno** si se migra completo, pero la complejidad del render de paginación puede justificar mantenerlo como está.

### Archivos a modificar
- Depende de decisión (evaluar en sprint planning)

---

## Fase 6: Exportar Excel (UI Solamente)

### Funciones a migrar
- `abrirModalExportar()` (app.js:489)
- `cerrarModalExportar()` (app.js:492)
- `cargarColumnasExportar()` (app.js:501)
- `filtrarSuperintendenciasExportar()` (app.js:567)
- `toggleTodasColumnas(sel)` (app.js:593)
- `ejecutarExportar()` (app.js:597)

### Patrón Alpine

**Alpine.data definition:**
```js
Alpine.data('exportarExcel', () => ({
    modulo: 'expedientes',
    columnas: [],
    filtros: {},
    fechaDesde: '',
    fechaHasta: '',
    cargando: false,

    async cargarColumnas() {
        this.cargando = true;
        try {
            const res = await fetch('/api/columnas-modulo?modulo=' + this.modulo);
            const data = await res.json();
            this.columnas = data.view_cols.map(c => ({ nombre: c, seleccionada: true }));
            this.filtros = {};
            // Cargar filtros dinámicos si hay catalogs
        } finally {
            this.cargando = false;
        }
    },

    toggleTodas(seleccionadas) {
        this.columnas.forEach(c => c.seleccionada = seleccionadas);
    },

    async exportar() {
        this.cargando = true;
        const cols = this.columnas.filter(c => c.seleccionada).map(c => c.nombre);
        const params = new URLSearchParams({
            modulo: this.modulo,
            ...(this.fechaDesde && { fecha_desde: this.fechaDesde }),
            ...(this.fechaHasta && { fecha_hasta: this.fechaHasta }),
            ...(cols.length && { columnas: cols.join(',') }),
            ...this.filtros
        });

        try {
            const res = await fetch('/api/exportar-excel?' + params);
            if (!res.ok) throw new Error(await res.text());

            const blob = await res.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = res.headers.get('Content-Disposition')?.match(/filename=(.+)/)?.[1] || 'reporte.xlsx';
            a.click();
            window.URL.revokeObjectURL(url);
        } finally {
            this.cargando = false;
        }
    }
}));
```

### Estado reemplazado
- Elimina `_expSuperCache`
- Elimina renderizado manual de checkboxes

### JS residual
**Fetch y descarga de blob** permanecen como JS vainilla dentro de Alpine.data porque son operaciones asíncronas que Alpine no abstrae.

### Archivos a modificar
- `templates/index.html` — modificar modal de exportar

---

## Fase 7: Campos Numéricos / Conversión USD/Bs

### Funciones a migrar
- `inicializarCamposNumericos()` (app.js:716)
- `convertirMoneda(origen)` (app.js:728)
- `_initNumInput`, `_fmtNum`, `_parseValue`, `_rawNum`

### Patrón Alpine (Directiva Custom)

**Crear directiva personalizada:**
```js
// frontend/vendor/alpine-directives.js
Alpine.directive('currency', (el, { modifiers }, { evaluateLater, effect }) => {
    const locale = modifiers[0] || 'es-ES';

    // Input handler
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

    // Blur handler (format)
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

    // Focus handler (show raw)
    const onFocus = () => {
        if (el.dataset.raw !== undefined) {
            el.value = el.dataset.raw;
        }
    };

    el.addEventListener('input', onInput);
    el.addEventListener('blur', onBlur);
    el.addEventListener('focus', onFocus);

    // Initial format
    if (el.value) onBlur();

    // Cleanup
    return () => {
        el.removeEventListener('input', onInput);
        el.removeEventListener('blur', onBlur);
        el.removeEventListener('focus', onFocus);
    };
});
```

**Uso en HTML:**
```html
<input type="text"
       x-currency
       x-model="presupuestoUsd"
       inputmode="decimal"
       class="input">
```

**Para conversión USD/Bs:**
```js
Alpine.data('formularioExpediente', (registroInicial, tipoCambioInicial) => ({
    registro: registroInicial,
    tipoCambio: tipoCambioInicial,
    convLock: false,

    convertirMoneda(origen) {
        if (this.convLock || !this.tipoCambio) return;
        this.convLock = true;

        try {
            if (origen === 'bs_presup') {
                const bs = parseFloat(this.registro.presupuesto_base_bs?.replace(/,/g, '.')) || 0;
                if (bs) {
                    this.registro.presupuesto_base_usd = (bs / this.tipoCambio).toFixed(2);
                }
            } else if (origen === 'usd_adj') {
                const usd = parseFloat(this.registro.monto_adjudicado_usd?.replace(/,/g, '.')) || 0;
                if (usd) {
                    this.registro.monto_adjudicado_bs = (usd * this.tipoCambio).toFixed(2);
                }
            }
            // ... otros casos
        } finally {
            this.convLock = false;
        }
    }
}));
```

### Estado reemplazado
- Elimina `_superCache` (se maneja en cada formulario)
- Elimina listeners manuales en cada input

### JS residual
**La directiva custom `x-currency`** es JS necesario porque Alpine no tiene formato numérico nativo. Pero es código genérico reusable, no gluecode específico.

### Archivos a modificar
- `frontend/vendor/alpine-directives.js` — crear directiva
- `templates/form_*.html` — aplicar directiva y x-data

---

## Fase 8: Superintendencias (Filtro Dependiente)

### Funciones a migrar
- `cargarSuperintendencias()` (app.js:144)

### Patrón Alpine

**Alpine.data definition:**
```js
Alpine.data('filtroSuperintendencias', (superintendenciasTodos, gerenciaInicial) => ({
    superintendencias: superintendenciasTodos,
    gerenciaSeleccionada: gerenciaInicial,
    superintendenciaSeleccionada: '',

    get superintendenciasFiltradas() {
        if (!this.gerenciaSeleccionada) {
            return this.superintendencias;
        }
        return this.superintendencias.filter(s => s.id_gerencia === this.gerenciaSeleccionada);
    },

    onChangeGerencia() {
        this.superintendenciaSeleccionada = '';
    }
}));
```

**HTML:**
```html
<div x-data="filtroSuperintendencias(@json($superintendencias), @json($registro?.id_gerencia))">
    <select x-model="gerenciaSeleccionada"
            @change="onChangeGerencia()"
            name="id_gerencia"
            class="input">
        <option value="">Seleccione...</option>
        <template x-for="ger in gerencias" :key="ger.id">
            <option :value="ger.id" x-text="ger.nombre"></option>
        </template>
    </select>

    <select x-model="superintendenciaSeleccionada"
            name="id_superintendencia"
            class="input">
        <option value="">Seleccione...</option>
        <template x-for="sup in superintendenciasFiltradas" :key="sup.id">
            <option :value="sup.id" x-text="sup.nombre"></option>
        </template>
    </select>
</div>
```

### Estado reemplazado
- Elimina `_superCache` global
- Elimina manipulación manual de options del select

### JS residual
**Ninguno.** El filtro dependiente es el caso de uso perfecto para Alpine.

### Archivos a modificar
- `templates/components.html` — modificar template `form_gerencia_superintendencia`

---

## Fase 9: Gantt / Ruta Procesos (NO MIGRAR)

### Funciones a migrar
- **NINGUNA** — mantener IIFE actual

### Justificación

El Gantt tiene:
1. **Estado complejo** (processes, timeline, legend, columns)
2. **Renderizado Canvas/SVG** (tabla dinámica con celdas interactivas)
3. **Lógica de negocio embebida** (agregar proceso, toggle, eliminar, cronograma)
4. **IIFE ya encapsulada** — no contamina global scope

**Recomendación:** Mantener la IIFE actual en `ruta_procesos.html`. Alpine no aporta valor significativo aquí y la migración sería costosa.

### JS residual
**Todo el código del Gantt** (~300 líneas en ruta_procesos.html). Es aceptable porque:
- Ya está bien encapsulado
- No interfiere con el resto de la app
- La complejidad justifica JS imperativo

---

## Resumen de Priorización

| Fase | ROI | Complejidad | JS Residual | Prioridad |
|------|-----|-------------|-------------|-----------|
| 1. Modales | Alto | Baja | Mínimo (puente HTMX) | **1** |
| 2. Pines/Fijados | Alto | Media | Ninguno | **2** |
| 3. BD Recientes | Alto | Media | Mínimo (registro auto) | **3** |
| 4. Sumas | Medio | Baja | Ninguno | **4** |
| 5. Paginación | Bajo | Alta | Posiblemente alto | **Evaluar** |
| 6. Exportar | Medio | Media | Fetch/download | **5** |
| 7. Campos Numéricos | Medio | Media | Directiva custom | **6** |
| 8. Superintendencias | Medio | Baja | Ninguno | **7** |
| 9. Gantt | N/A | N/A | Todo | **No migrar** |

---

## Estructura de Archivos Final Propuesta

```
frontend/
├── vendor/
│   ├── alpine.min.js              # Alpine.js core (copiar)
│   ├── alpine-app.js              # Alpine.data definitions
│   ├── alpine-directives.js       # Directivas custom (x-currency)
│   └── alpine-htmx-bridge.js      # Puente opcional HTMX ↔ Alpine
templates/
├── index.html                     # Con x-data en body
├── components.html                # Templates con Alpine
├── form.html                      # Template unificado (ver sección siguiente)
├── tabla.html                     # Template unificado (ver sección siguiente)
└── ruta_procesos.html             # Sin cambios (IIFE propia)
```

---

## Unificación de Templates (Bonus)

Además de migrar a Alpine, se pueden reducir los 18 templates (9 form + 9 tabla) a 2:

### form.html unificado

```html
{{define "form_unificado"}}
<div class="modal-body space-y-6 overflow-y-auto max-h-[75vh]"
     x-data="formularioModulo(@json .ModuloKey, @json .Registro)">

    <input type="hidden" name="{{.IDColumna}}" :value="registro.{{.IDColumna}}">

    {{range $campo := .Campos}}
        {{if eq $campo.Tipo "select_catalog"}}
            {{template "form_select_alpine" dict "Campo" $campo "Catalogo" (index $.Catalogos $campo.CatalogName)}}
        {{else if eq $campo.Tipo "number"}}
            {{template "form_input_number_alpine" dict "Campo" $campo}}
        {{else if eq $campo.Tipo "textarea"}}
            {{template "form_textarea_alpine" dict "Campo" $campo}}
        {{else}}
            {{template "form_input_text_alpine" dict "Campo" $campo}}
        {{end}}
    {{end}}

    {{template "form_buttons_unificadas" dict "Modulo" .ModuloKey "IDColumna" .IDColumna}}
</div>
{{end}}
```

### tabla.html unificada

```html
{{define "tabla_unificada"}}
<div class="bg-gray-800 rounded-xl ...">
    <input type="hidden" id="active-module-val" value="{{.ModuloKey}}">

    <table class="w-full text-left">
        <thead>
            <tr>
                <th>Ver</th>
                <th>Acción</th>
                {{range $col := .ColsMostrar}}
                    <th>{{$col.Label}}</th>
                {{end}}
            </tr>
        </thead>
        <tbody id="tabla-cuerpo">
            {{range $exp := .Filas}}
                <tr @click="toggleDesplegable({{rowGetStr $exp $.IDColumna}})">
                    <td><button><i class="fas fa-plus"></i></button></td>
                    <td>
                        <button hx-get="/api/cargar-expediente?modulo={{$.ModuloKey}}&id={{rowGetStr $exp $.IDColumna}}"
                                @htmx:after-request="$dispatch('abrir-formulario')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button @click="$dispatch('toggle-fijado', {id: {{rowGetStr $exp $.IDColumna}}})">
                            <i class="fas fa-thumbtack"></i>
                        </button>
                    </td>
                    {{range $col := .ColsMostrar}}
                        <td>{{rowGetStr $exp $col.Campo}}</td>
                    {{end}}
                </tr>
                <tr :id="'subfila-' + {{rowGetStr $exp $.IDColumna}}" class="hidden">
                    <td colspan="{{len .ColsMostrar}}">
                        <!-- Detalles expandibles -->
                    </td>
                </tr>
            {{end}}
        </tbody>
    </table>
</div>
{{end}}
```

**Nota:** La unificación completa requiere cambios en Go (handler.go) para pasar `Campos`, `ColsMostrar` dinámicamente. Esto está fuera del scope de migración Alpine pero es el siguiente paso lógico.

---

## Siguientes Pasos

1. **Copiar Alpine.js** a `frontend/vendor/alpine.min.js`
2. **Implementar Fase 1 (Modales)** — mayor ROI, menor riesgo
3. **Validar** que HTMX sigue funcionando correctamente
4. **Iterar** por fases restantes según prioridad
5. **Documentar** patrones exitosos en `docs/funciones.md`