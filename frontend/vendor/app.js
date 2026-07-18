const $ = id => document.getElementById(id);

// Claves de localStorage (migrado de schema-config.js)
const STORAGE_KEYS = { FRECUENTES: 'sidebarFrecuentes', RECIENTES: 'baseaccess_recientes' };

// --- Toast ---
function esc(v) {
    if (v == null) return '';
    return String(v).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#39;');
}

function toast(msg, tipo = 'info') {
    const el = document.createElement('div');
    el.className = 'toast ' + tipo;
    el.textContent = msg;
    $('toast-container').appendChild(el);
    requestAnimationFrame(() => el.classList.add('show'));
    setTimeout(() => { el.classList.remove('show'); setTimeout(() => el.remove(), 300); }, 3000);
}

// Sobrescribir alert globalmente para redirigir a toast
window.alert = function(msg) {
    toast(msg, 'error');
};

// --- Abrir BD (único punto que necesita Wails binding) ---
async function abrirBaseDatos() {
    let path;
    try {
        if (window.go && window.go.main && window.go.main.App) {
            path = await window.go.main.App.AbrirDialogoBD();
        } else if (window.runtime) {
            path = await window.runtime.OpenFileDialog({
                Title: 'Seleccionar base de datos',
                Filters: [{ DisplayName: 'SQLite DB', Pattern: '*.db;*.sqlite' }]
            });
        }
        if (path) {
            await fetch('/api/abrir-bd', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: 'path=' + encodeURIComponent(path)
            });
            location.reload();
        }
    } catch(e) {
        $('dbfile').click();
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const dbfile = $('dbfile');
    if (dbfile) {
        dbfile.addEventListener('change', async function(e) {
            const f = e.target.files[0];
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
                alert('Error al abrir BD: ' + err);
            }
        });
    }
});

// --- Drag & Drop ---
document.addEventListener('dragover', e => { e.preventDefault(); $('body').classList.add('drag-over'); });
document.addEventListener('dragleave', e => { if (!e.relatedTarget) $('body').classList.remove('drag-over'); });
document.addEventListener('drop', async e => {
    e.preventDefault(); $('body').classList.remove('drag-over');
    const f = e.dataTransfer?.files?.[0]; if (!f) return;
    const path = f.path || f.name;
    try {
        await fetch('/api/abrir-bd', { method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: 'path=' + encodeURIComponent(path) });
        location.reload();
    } catch(err) { alert('Error: ' + err); }
});

// --- Desplegable fila ---
function toggleDesplegable(id) {
    const sub = $('subfila-' + id);
    const btn = $('btn-' + id);
    if (!sub || !btn) return;
    sub.classList.toggle('hidden');
    btn.innerHTML = sub.classList.contains('hidden') ? '<i class="fas fa-plus"></i>' : '<i class="fas fa-minus"></i>';
}

// --- Stack de modales (cierre al clickear afuera con jerarquia) ---
const MODAL_STACK = [];
function pushModal(id) {
    const el = $(id);
    if (!el || !el.classList.contains('hidden')) return;
    el.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    MODAL_STACK.push(id);
}

document.addEventListener('click', function(e) {
    if (MODAL_STACK.length === 0) return;
    const modal = e.target.closest('.modal');
    if (!modal || e.target !== modal) return;
    if (MODAL_STACK[MODAL_STACK.length - 1] === modal.id) {
        cerrarModal(modal.id);
    }
});
function cerrarModal(id) {
    const el = $(id);
    if (!el) return;
    el.classList.add('hidden');
    const idx = MODAL_STACK.lastIndexOf(id);
    if (idx !== -1) MODAL_STACK.splice(idx, 1);
    if (MODAL_STACK.length === 0) document.body.style.overflow = '';
}

// --- Formulario ---
function mostrarFormulario(id, modulo) {
    modulo = modulo || (window.PAGE_DATA && window.PAGE_DATA.modulos && $('active-module-val'))
        ? $('active-module-val').value
        : 'expedientes';
    const nombreModulo = (window.PAGE_DATA && window.PAGE_DATA.modulos && window.PAGE_DATA.modulos[modulo])
        ? window.PAGE_DATA.modulos[modulo].Nombre
        : 'Registro';
    $('form-titulo').textContent = id ? 'Editar ' + nombreModulo + ' #' + id : 'Nuevo ' + nombreModulo;
    pushModal('form-modal');
    setTimeout(cargarSuperintendencias, 50);
}

function cerrarFormulario() { cerrarModal('form-modal'); }

// --- Superintendencias (cachea todas las opciones para filtrar siempre desde el set completo) ---
var _superCache = null;
function cargarSuperintendencias() {
    const gerId = $('f-id_gerencia')?.value;
    const sel = $('f-id_superintendencia');
    if (!sel) return;
    var curVal = sel.value;
    if (_superCache === null) {
        _superCache = [];
        sel.querySelectorAll('option').forEach(function(o) {
            if (!o.value) return;
            _superCache.push({ v: o.value, t: o.textContent, g: o.getAttribute('data-id-gerencia') || '' });
        });
    }
    var frag = document.createDocumentFragment();
    var empty = document.createElement('option'); empty.value = '';
    frag.appendChild(empty);
    var found = false;
    _superCache.forEach(function(o) {
        if (!gerId || o.g === gerId) {
            var el = document.createElement('option');
            el.value = o.v; el.textContent = o.t; el.setAttribute('data-id-gerencia', o.g);
            if (o.v === curVal) found = true;
            frag.appendChild(el);
        }
    });
    sel.replaceChildren(frag);
    if (found) sel.value = curVal;
}
document.body.addEventListener('htmx:afterSettle', function() {
    _superCache = null;
    inicializarCamposNumericos();
    if ($('form-modal') && !$('form-modal').classList.contains('hidden')) {
        cargarSuperintendencias();
    }
});

function cerrarHistorial() { cerrarModal('historial-modal'); }
function cerrarRuta() { cerrarModal('modal-ruta'); }
function cerrarPendientes() { cerrarModal('modal-pendientes'); }

// --- BD Recientes ---
function registrarReciente(nombre, path) {
    if (!nombre || !path) return;
    let recientes = [];
    try {
        recientes = JSON.parse(localStorage.getItem(STORAGE_KEYS.RECIENTES) || '[]');
    } catch(e) { recientes = []; }
    const idx = recientes.findIndex(r => r.path === path);
    if (idx !== -1) recientes.splice(idx, 1);
    recientes.unshift({ nombre, path, timestamp: Date.now() });
    if (recientes.length > 5) recientes = recientes.slice(0, 5);
    localStorage.setItem(STORAGE_KEYS.RECIENTES, JSON.stringify(recientes));
}

function eliminarReciente(path) {
    let recientes = [];
    try { recientes = JSON.parse(localStorage.getItem(STORAGE_KEYS.RECIENTES) || '[]'); } catch(e) {}
    recientes = recientes.filter(r => r.path !== path);
    localStorage.setItem(STORAGE_KEYS.RECIENTES, JSON.stringify(recientes));
    abrirRecientes();
}

function eliminarRecienteIndex(index) {
    let recientes = [];
    try { recientes = JSON.parse(localStorage.getItem(STORAGE_KEYS.RECIENTES) || '[]'); } catch(e) {}
    if (index >= 0 && index < recientes.length) {
        recientes.splice(index, 1);
    }
    localStorage.setItem(STORAGE_KEYS.RECIENTES, JSON.stringify(recientes));
    abrirRecientes();
}

async function abrirBaseDatosReciente(path) {
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
            alert("No se pudo reabrir la base de datos: " + err);
            eliminarReciente(path);
        }
    } catch (err) {
        alert("Error al abrir base de datos: " + err);
        eliminarReciente(path);
    }
    cerrarRecientes();
}

function abrirRecientes() {
    const modal = $('modal-recientes');
    const cont = $('recientes-contenido');
    if (!modal || !cont) return;
    pushModal('modal-recientes');
    let recientes = [];
    try { recientes = JSON.parse(localStorage.getItem(STORAGE_KEYS.RECIENTES) || '[]'); } catch(e) {}
    if (recientes.length === 0) {
        cont.innerHTML = '<p class="text-gray-500 text-center py-8 italic">Sin bases de datos recientes.</p>';
    } else {
        cont.innerHTML = '<div class="flex flex-col gap-2">' + recientes.map((r, i) => `
            <div class="reciente-item flex items-center gap-2 px-4 py-3 cursor-pointer rounded-lg bg-gray-900/60 border border-gray-700/60 hover:border-teal-500/60 transition-all"
                 data-df-path="${r.path}" onclick="abrirBaseDatosReciente(this.dataset.dfPath)">
                <div class="min-w-0 flex-1">
                    <div class="text-sm text-gray-200 truncate font-semibold">${esc(r.nombre)}</div>
                    <div class="text-xs text-gray-500 truncate">${esc(r.path)}</div>
                </div>
                <button onclick="event.stopPropagation(); eliminarRecienteIndex(${i})" class="text-xs text-gray-400 hover:text-red-400 shrink-0 p-1" title="Quitar de recientes"><i class="fas fa-times"></i></button>
            </div>
        `).join('') + '</div>';
    }
}

function cerrarRecientes() { cerrarModal('modal-recientes'); }

// --- Acceso Rápido / Pinned items ---
function toggleFrecuente(id, solped, modulo) {
    modulo = modulo || (window.PAGE_DATA ? window.PAGE_DATA.ActiveModule : 'expedientes') || 'expedientes';
    let list = [];
    try {
        list = JSON.parse(localStorage.getItem(STORAGE_KEYS.FRECUENTES) || '[]');
    } catch(e) { list = []; }
    const idx = list.findIndex(i => Number(i.id) === Number(id) && i.modulo === modulo);
    if (idx !== -1) {
        list.splice(idx, 1);
    } else {
        list.push({ id: Number(id), solped: solped || '', modulo: modulo });
    }
    localStorage.setItem(STORAGE_KEYS.FRECUENTES, JSON.stringify(list));
    
    // Si el modal de fijados está abierto, refrescarlo
    if (!$('modal-frecuentes').classList.contains('hidden')) {
        abrirFrecuentes();
    }
    
    // Actualizar color del botón en la tabla principal
    const btn = $(`pin-btn-${id}`);
    if (btn) {
        if (idx === -1) {
            btn.className = "btn-icon text-emerald-500 hover:text-emerald-400";
        } else {
            btn.className = "btn-icon text-blue-500 hover:text-blue-400";
        }
    }
}

function abrirFrecuentes() {
    const modal = $('modal-frecuentes');
    const cont = $('frecuentes-contenido');
    if (!modal || !cont) return;
    pushModal('modal-frecuentes');
    
    let list = [];
    try {
        list = JSON.parse(localStorage.getItem(STORAGE_KEYS.FRECUENTES) || '[]');
    } catch(e) { list = []; }
    
    if (list.length === 0) {
        cont.innerHTML = `
            <div class="text-center py-12 text-gray-500 space-y-2">
                <i class="fas fa-thumbtack text-gray-600 text-2xl rotate-45 animate-pulse"></i>
                <p class="text-sm italic">Sin expedientes fijados.</p>
                <p class="text-xs text-gray-600">Haz clic en el pin (<i class="fas fa-thumbtack text-blue-500"></i>) de un expediente en la tabla para anclarlo aquí.</p>
            </div>
        `;
        return;
    }
    cont.innerHTML = '<div class="grid grid-cols-1 md:grid-cols-2 gap-3">' + list.map(item => `
        <div data-df-id="${item.id}" data-df-modulo="${item.modulo || 'expedientes'}" class="flex items-center justify-between bg-gray-900/60 hover:bg-gray-900 border border-gray-700/60 hover:border-teal-500/30 rounded-xl p-4 cursor-pointer transition-all duration-200"
             onclick="event.stopPropagation(); cerrarFrecuentes(); hxGetFormulario(this.dataset.dfId, this.dataset.dfModulo)">
            <div class="min-w-0 flex-1">
                <div class="text-[10px] text-gray-500 uppercase font-bold tracking-wider">SOLPED</div>
                <div class="text-sm text-gray-200 font-bold truncate">${esc(item.solped) || 'SIN SOLPED'}</div>
            </div>
            <button data-df-id="${item.id}" data-df-solped="${item.solped || ''}" onclick="event.stopPropagation(); var d=this.dataset; toggleFrecuente(d.dfId, d.dfSolped)" class="text-gray-500 hover:text-red-400 p-2 rounded-lg hover:bg-gray-800 transition-all duration-200" title="Desanclar">
                <i class="fas fa-times text-xs"></i>
            </button>
        </div>
    `).join('') + '</div>';
}

function cerrarFrecuentes() { cerrarModal('modal-frecuentes'); }

function inicializarPines() {
    let list = [];
    try {
        list = JSON.parse(localStorage.getItem(STORAGE_KEYS.FRECUENTES) || '[]');
    } catch(e) { list = []; }
    
    // Restablecer todos los botones de pin visibles a azul
    document.querySelectorAll('[id^="pin-btn-"]').forEach(btn => {
        btn.className = "btn-icon text-blue-500 hover:text-blue-400";
    });
    
    // Colorear de verde los que estén guardados
    list.forEach(item => {
        const btn = $(`pin-btn-${item.id}`);
        if (btn) {
            btn.className = "btn-icon text-emerald-500 hover:text-emerald-400";
        }
    });
}

// --- Paginación del lado del cliente ---
let currentPage = 1;
const pageSize = 10;

function irPagina(pagina) {
    currentPage = pagina;
    aplicarPaginacionDOM();
}

function aplicarPaginacionDOM() {
    // Obtener todas las filas principales (omitir las subfilas de detalles)
    const filas = Array.from(document.querySelectorAll('#tabla-cuerpo > tr')).filter(tr => !tr.id.startsWith('subfila-'));
    const totalItems = filas.length;
    const totalPages = Math.ceil(totalItems / pageSize);

    if (currentPage > totalPages) currentPage = Math.max(1, totalPages);

    const startIdx = (currentPage - 1) * pageSize;
    const endIdx = currentPage * pageSize;

    // Ocultar o mostrar las filas principales
    filas.forEach((fila, idx) => {
        const clickAttr = fila.getAttribute('onclick') || '';
        const match = clickAttr.match(/\d+/);
        const idExp = match ? match[0] : null;
        const subfila = idExp ? $('subfila-' + idExp) : null;
        
        if (idx >= startIdx && idx < endIdx) {
            fila.style.display = '';
        } else {
            fila.style.display = 'none';
            if (subfila) subfila.classList.add('hidden'); // Forzar a ocultar la subfila si la principal está oculta
        }
    });

    renderPaginacionControles(totalPages);
}

function renderPaginacionControles(totalPages) {
    const cont = $('paginacion');
    if (!cont) return;

    if (totalPages <= 1) {
        cont.innerHTML = '';
        return;
    }

    const buildBtn = (label, onclick, disabled, active) =>
        `<button onclick="${onclick}" class="btn btn-sm ${active ? 'bg-teal-600 text-white' : 'bg-gray-700 text-gray-300 hover:bg-gray-600'}" ${disabled ? 'disabled' : ''}>${label}</button>`;

    let html = '<div class="flex items-center justify-center gap-1.5 flex-wrap">';

    html += buildBtn('<i class="fas fa-angle-double-left text-[10px]"></i>', 'irPagina(1)', currentPage === 1, false);
    html += buildBtn('<i class="fas fa-chevron-left text-[10px]"></i>', `irPagina(${currentPage - 1})`, currentPage === 1, false);

    const maxVisible = 7;
    let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
    let endPage = Math.min(totalPages, startPage + maxVisible - 1);
    if (endPage - startPage + 1 < maxVisible) {
        startPage = Math.max(1, endPage - maxVisible + 1);
    }

    if (startPage > 1) {
        html += buildBtn('1', 'irPagina(1)', false, false);
        if (startPage > 2) html += '<span class="text-gray-500 px-1 text-xs">...</span>';
    }

    for (let i = startPage; i <= endPage; i++) {
        html += buildBtn(String(i), `irPagina(${i})`, false, i === currentPage);
    }

    if (endPage < totalPages) {
        if (endPage < totalPages - 1) html += '<span class="text-gray-500 px-1 text-xs">...</span>';
        html += buildBtn(String(totalPages), `irPagina(${totalPages})`, false, false);
    }

    html += buildBtn('<i class="fas fa-chevron-right text-[10px]"></i>', `irPagina(${currentPage + 1})`, currentPage === totalPages, false);
    html += buildBtn('<i class="fas fa-angle-double-right text-[10px]"></i>', `irPagina(${totalPages})`, currentPage === totalPages, false);

    html += '</div>';
    cont.innerHTML = html;
}

// Registrar evento htmx:afterSwap para colorear pines y paginar al recargar la tabla
document.addEventListener('DOMContentLoaded', () => {
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (evt.detail.target.id === 'tabla-cuerpo' || evt.detail.target.id === 'vista-tabla') {
            currentPage = 1;
            aplicarPaginacionDOM();
            inicializarPines();
        }
    });

    // Auto-registro en carga de página y refresco de Acceso Rápido y Paginación
    if (window.PAGE_DATA && window.PAGE_DATA.hasDB && window.PAGE_DATA.dbPath) {
        const path = window.PAGE_DATA.dbPath;
        const nombre = path.split('/').pop().split('\\').pop() || path;
        registrarReciente(nombre, path);
    }
    currentPage = 1;
    aplicarPaginacionDOM();
    if (window.PAGE_DATA && window.PAGE_DATA.hasDB) inicializarPines();

    // Auto-abrir si no hay BD y solo hay una reciente
    if (!window.PAGE_DATA || !window.PAGE_DATA.hasDB) {
        let recientes = [];
        try { recientes = JSON.parse(localStorage.getItem(STORAGE_KEYS.RECIENTES) || '[]'); } catch(e) {}
        if (recientes.length === 1) {
            abrirBaseDatosReciente(recientes[0].path);
        }
    }
});

function toggleSortDir() {
    const valEl = $('sort-dir-val');
    const dir = valEl.value === 'ASC' ? 'DESC' : 'ASC';
    valEl.value = dir;
    const icon = $('sort-dir').querySelector('i');
    if (dir === 'ASC') {
        icon.className = 'fas fa-arrow-up-short-wide';
        $('sort-dir').title = 'Cambiar dirección (Ascendente)';
    } else {
        icon.className = 'fas fa-arrow-down-wide-short';
        $('sort-dir').title = 'Cambiar dirección (Descendente)';
    }
    htmx.trigger('#sort-order', 'change');
}

function hxGetFormulario(id, modulo) {
    modulo = modulo || (window.PAGE_DATA ? window.PAGE_DATA.ActiveModule : 'expedientes') || 'expedientes';
    htmx.ajax('GET', '/api/cargar-expediente?modulo=' + encodeURIComponent(modulo) + '&id=' + id, {target: '#form-expediente'}).then(() => {
        mostrarFormulario(id, modulo);
    });
}

// --- Exportar Excel ---
function abrirModalExportar() {
    pushModal('export-modal');
    cargarColumnasExportar();
}
function cerrarModalExportar() {
    cerrarModal('export-modal');
}

async function cargarColumnasExportar() {
    const modulo = $('exp-modulo').value;
    const contCols = $('exp-columnas');
    const contFiltros = $('exp-filtros-dinamicos');
    contCols.innerHTML = '<p class="text-gray-500 text-xs col-span-full">Cargando...</p>';
    contFiltros.innerHTML = '';
    
    const res = await fetch('/api/columnas-modulo?modulo=' + encodeURIComponent(modulo));
    if (!res.ok) { 
        contCols.innerHTML = '<p class="text-red-400 text-xs col-span-full">Error al cargar</p>'; 
        return; 
    }
    const data = await res.json();
    const viewCols = data.view_cols;
    const tableCols = data.table_cols;
    
    // Render columns checkboxes
    contCols.innerHTML = '';
    viewCols.forEach(c => {
        const lbl = document.createElement('label');
        lbl.className = 'exp-checkbox-label';
        lbl.innerHTML = '<input type="checkbox" value="' + c + '" class="exp-col"> ' + c.replace(/_/g, ' ');
        contCols.appendChild(lbl);
    });
    
    // Render dynamic filters
    const catalogKeys = (window.PAGE_DATA && window.PAGE_DATA.CatalogFilters) || {};
    
    tableCols.forEach(col => {
        const catInfo = catalogKeys[col];
        if (catInfo && window.PAGE_DATA && window.PAGE_DATA.catalogs) {
            const items = window.PAGE_DATA.catalogs[catInfo.key] || [];
            if (items.length > 0) {
                const div = document.createElement('div');
                div.className = 'flex flex-col';
                
                let optionsHtml = '<option value="">Todos</option>';
                items.forEach(item => {
                    let dataGerAttr = '';
                    if (col === 'id_superintendencia' && item.id_gerencia) {
                        dataGerAttr = ` data-id-gerencia="${item.id_gerencia}"`;
                    }
                    optionsHtml += `<option value="${item.id}"${dataGerAttr}>${item.nombre}</option>`;
                });
                
                let onchangeHtml = '';
                if (col === 'id_gerencia') {
                    onchangeHtml = ' onchange="filtrarSuperintendenciasExportar()"';
                }
                
                div.innerHTML = `
                    <label class="label">${catInfo.label}</label>
                    <select id="exp-filt-${col}" name="${col}" class="input exp-filter-input"${onchangeHtml}>
                        ${optionsHtml}
                    </select>
                `;
                contFiltros.appendChild(div);
            }
        }
    });
    _expSuperCache = null;
    filtrarSuperintendenciasExportar();
}

var _expSuperCache = null;
function filtrarSuperintendenciasExportar() {
    const gerId = $('exp-filt-id_gerencia')?.value;
    const sel = $('exp-filt-id_superintendencia');
    if (!sel) return;
    var curVal = sel.value;
    if (_expSuperCache === null) {
        _expSuperCache = [];
        sel.querySelectorAll('option').forEach(function(o) {
            if (!o.value) return;
            _expSuperCache.push({ v: o.value, t: o.textContent, g: o.getAttribute('data-id-gerencia') || '' });
        });
    }
    var frag = document.createDocumentFragment();
    var empty = document.createElement('option'); empty.value = '';
    frag.appendChild(empty);
    var found = false;
    _expSuperCache.forEach(function(o) {
        if (!gerId || o.g === gerId) {
            var el = document.createElement('option');
            el.value = o.v; el.textContent = o.t; el.setAttribute('data-id-gerencia', o.g);
            if (o.v === curVal) found = true;
            frag.appendChild(el);
        }
    });
    sel.replaceChildren(frag);
    if (found) sel.value = curVal;
}

function toggleTodasColumnas(sel) {
    document.querySelectorAll('.exp-col').forEach(cb => cb.checked = sel);
}
async function ejecutarExportar() {
    const modulo = $('exp-modulo').value;
    const fd = $('exp-fecha-desde').value;
    const fh = $('exp-fecha-hasta').value;
    const cols = Array.from(document.querySelectorAll('.exp-col:checked')).map(cb => cb.value);
    const params = new URLSearchParams();
    params.set('modulo', modulo);
    if (fd) params.set('fecha_desde', fd);
    if (fh) params.set('fecha_hasta', fh);
    if (cols.length) params.set('columnas', cols.join(','));
    
    document.querySelectorAll('.exp-filter-input').forEach(sel => {
        if (sel.value) {
            params.set(sel.name, sel.value);
        }
    });
    
    const url = '/api/exportar-excel?' + params.toString();
    
    // Show spinner
    const spinner = $('spinner-overlay');
    if (spinner) {
        $('spinner-text').textContent = 'Generando Excel...';
        spinner.classList.remove('hidden');
    }
    
    try {
        const res = await fetch(url);
        if (!res.ok) {
            const text = await res.text();
            toast(text || 'Error al exportar los datos', 'error');
            return;
        }
        
        // Get filename from header
        const disposition = res.headers.get('Content-Disposition');
        let filename = 'reporte.xlsx';
        if (disposition && disposition.indexOf('attachment') !== -1) {
            const filenameRegex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
            const matches = filenameRegex.exec(disposition);
            if (matches != null && matches[1]) { 
                filename = matches[1].replace(/['"]/g, '');
            }
        }
        
        const blob = await res.blob();
        const downloadUrl = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = downloadUrl;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        a.remove();
        window.URL.revokeObjectURL(downloadUrl);
        toast('Excel descargado con éxito', 'success');
        cerrarModalExportar();
    } catch (err) {
        toast('Error de conexión al exportar', 'error');
    } finally {
        if (spinner) {
            spinner.classList.add('hidden');
        }
    }
}

// --- Sumas (calculadora con 2 decimales) ---
function abrirSumas() {
    pushModal('sumas-modal');
    if (document.querySelectorAll('.suma-fila').length === 0) anyadirFilaSuma();
    inicializarCamposNumericos();
    calcularSumas();
}
function cerrarSumas() { cerrarModal('sumas-modal'); }
function anyadirFilaSuma() {
    const div = document.createElement('div');
    div.className = 'suma-fila flex items-center gap-3 w-full pr-1';
    div.innerHTML = '<input type="text" inputmode="decimal" placeholder="0,00" class="input flex-1 suma-input" oninput="calcularSumas()"><button onclick="this.parentElement.remove(); calcularSumas();" class="w-8 h-8 shrink-0 flex items-center justify-center text-red-400 hover:text-red-300 hover:bg-gray-700/50 rounded-lg transition-colors" title="Quitar"><i class="fas fa-times text-sm"></i></button>';
    $('sumas-filas').appendChild(div);
    _initNumInput(div.querySelector('input'));
    div.querySelector('input').focus();
}
function calcularSumas() {
    let total = 0;
    document.querySelectorAll('.suma-input').forEach(inp => {
        total += _parseValue(inp);
    });
    $('sumas-resultado').textContent = total.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
function limpiarSumas() {
    $('sumas-filas').innerHTML = '';
    anyadirFilaSuma();
    $('sumas-resultado').textContent = '0.00';
}

// --- Helpers JS ---
function formatNum(v) {
    if (!v && v !== 0) return '';
    const s = String(v).replace(/\./g, '').replace(',', '.');
    const n = parseFloat(s); if (isNaN(n)) return String(v);
    return n.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

// --- Campos numéricos: formato, separador miles, max 2 decimales ---
function _rawNum(input) {
    if (input.dataset.raw !== undefined) return input.dataset.raw;
    var v = input.value.replace(/,/g, '.').replace(/[^0-9.]/g, '');
    if (v.endsWith('.')) v = v.slice(0, -1);
    var lastDot = v.lastIndexOf('.');
    if (lastDot >= 0) v = v.substring(0, lastDot).replace(/\./g, '') + v.substring(lastDot);
    return v;
}
function _parseValue(input) {
    var v = (input.value || '').replace(/,/g, '.').replace(/[^0-9.]/g, '');
    if (v.endsWith('.')) v = v.slice(0, -1);
    var lastDot = v.lastIndexOf('.');
    if (lastDot >= 0) v = v.substring(0, lastDot).replace(/\./g, '') + v.substring(lastDot);
    var n = parseFloat(v);
    return isNaN(n) ? 0 : n;
}
function _fmtNum(input) {
    var raw = _rawNum(input);
    if (!raw || isNaN(parseFloat(raw))) return;
    var n = parseFloat(raw);
    input.dataset.raw = n.toFixed(2);
    input.value = n.toLocaleString('es-ES', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
function _initNumInput(input) {
    if (input.dataset.numInited) return;
    input.dataset.numInited = '1';
    input.addEventListener('input', function() {
        var v = this.value.replace(/,/g, '.').replace(/[^0-9.]/g, '');
        var lastDot = v.lastIndexOf('.');
        if (lastDot >= 0) v = v.substring(0, lastDot).replace(/\./g, '') + v.substring(lastDot);
        var parts = v.split('.');
        if (parts.length === 2 && parts[1].length > 2) v = parts[0] + '.' + parts[1].substring(0, 2);
        this.dataset.raw = v;
        this.value = v;
    });
    input.addEventListener('focus', function() {
        if (this.dataset.raw !== undefined) {
            this.value = this.dataset.raw;
        }
    });
    input.addEventListener('blur', function() { _fmtNum(this); });
    if (input.value) _fmtNum(input);
}
function inicializarCamposNumericos() {
    document.querySelectorAll('input[inputmode="decimal"], input.num-field').forEach(_initNumInput);
    document.querySelectorAll('.suma-input').forEach(_initNumInput);
}

// --- Conversión USD ↔ Bs ---
var _convLock = false;
function convertirMoneda(origen) {
    if (_convLock) return;
    var tcEl = document.getElementById('f-tipo_cambio');
    var tc = tcEl ? (parseFloat(tcEl.dataset.raw) || 0) : 0;
    if (!tc) return;
    _convLock = true;
    try {
        function getRaw(id) {
            var el = document.getElementById(id);
            if (!el) return 0;
            return _parseValue(el);
        }
        function setVal(id, val) {
            var el = document.getElementById(id);
            if (!el) return;
            el.dataset.raw = val.toFixed(2);
            el.dispatchEvent(new Event('blur', {bubbles: true}));
        }
        // Presupuesto: solo se convierte cuando se editan campos de presupuesto
        if (origen === 'bs_presup') {
            var bs = getRaw('f-presupuesto_base_bs');
            if (bs) setVal('f-presupuesto_base_usd', bs / tc);
        } else if (origen === undefined || origen === 'usd_presup' || origen === '') {
            var usd = getRaw('f-presupuesto_base_usd');
            if (usd) setVal('f-presupuesto_base_bs', usd * tc);
        }
        // Monto Adjudicado: solo se convierte cuando se editan campos de adjudicación
        if (origen === 'bs_adj') {
            var adjBs = getRaw('f-monto_adjudicado_bs');
            if (adjBs) setVal('f-monto_adjudicado_usd', adjBs / tc);
        } else if (origen === 'usd_adj' || (origen === undefined && getRaw('f-monto_adjudicado_usd') && !getRaw('f-monto_adjudicado_bs'))) {
            var adjUsd = getRaw('f-monto_adjudicado_usd');
            if (adjUsd) setVal('f-monto_adjudicado_bs', adjUsd * tc);
        }
    } finally {
        _convLock = false;
    }
}
