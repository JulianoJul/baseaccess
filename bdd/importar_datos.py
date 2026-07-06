import sqlite3
import re
import os

DB = os.path.join(os.path.dirname(__file__), 'expedientes.db')
SCHEMA = os.path.join(os.path.dirname(__file__), 'Tablas8.sql')
DATA = os.path.join(os.path.dirname(__file__), 'datos_excel.sql')

COLUMNS = [
    'solped', 'id_gerencia', 'id_superintendencia', 'id_emisor', 'id_documento',
    'fecha_presupuesto_base', 'presupuesto_base_usd', 'tipo_cambio', 'presupuesto_base_bs',
    'id_plan', 'descripcion_proceso', 'id_modalidad', 'id_art', 'id_tipo_contrato',
    'nro_acta_apertura', 'cantidad_frentes', 'nro_resolucion_jd', 'id_estatus',
    'fecha_recibido', 'fecha_devuelto', 'id_receptor', 'nro_proceso', 'id_resultado',
    'nro_contrato_sicac', 'nro_contrato_sap', 'id_empresa', 'tiempo_ejecucion',
    'monto_adjudicado_bs', 'monto_adjudicado_usd', 'fecha_firma_contrato', 'observaciones_generales'
]

OBS_COL = COLUMNS.index('observaciones_generales')

def parse_row_values(line):
    line = line.strip()
    if line.endswith(';'):
        line = line[:-1]
    if line.endswith(','):
        line = line[:-1]
    if line.startswith('(') and line.endswith(')'):
        line = line[1:-1]
    else:
        return None

    vals = []
    current = ''
    in_quote = False
    depth = 0
    for ch in line:
        if ch == "'" and not in_quote:
            in_quote = True
            current += ch
        elif ch == "'" and in_quote:
            in_quote = False
            current += ch
        elif ch == '(' and not in_quote:
            depth += 1
            current += ch
        elif ch == ')' and not in_quote:
            depth -= 1
            current += ch
        elif ch == ',' and depth == 0 and not in_quote:
            vals.append(current.strip())
            current = ''
        else:
            current += ch
    if current.strip():
        vals.append(current.strip())
    return vals

def parse_value(v):
    if v.upper() == 'NULL':
        return None
    if v.startswith("'") and v.endswith("'"):
        return v[1:-1]
    if v == '':
        return None
    try:
        if '.' in v:
            return float(v)
        return int(v)
    except ValueError:
        return v

def generar_auto_linea(cur, vals):
    id_estatus = vals[COLUMNS.index('id_estatus')]
    id_documento = vals[COLUMNS.index('id_documento')]
    fecha_recibido = vals[COLUMNS.index('fecha_recibido')]
    fecha_devuelto = vals[COLUMNS.index('fecha_devuelto')]

    estatus_nombre = cur.execute(
        "SELECT nombre FROM cat_estatus_detalle WHERE id = ?", (id_estatus,)
    ).fetchone()
    estatus = estatus_nombre[0] if estatus_nombre else ''

    doc_nombre = cur.execute(
        "SELECT nombre FROM cat_documento WHERE id = ?", (id_documento,)
    ).fetchone()
    doc = doc_nombre[0] if doc_nombre else ''

    partes = [estatus, doc]
    if fecha_recibido:
        partes.append(f'*Fecha recibido* {fecha_recibido}')
    if fecha_devuelto:
        partes.append(f'*Fecha devuelto* {fecha_devuelto}')
    return ' - '.join(partes)

def main():
    if os.path.exists(DB):
        os.remove(DB)

    con = sqlite3.connect(DB)
    con.execute("PRAGMA foreign_keys = OFF")
    cur = con.cursor()

    with open(SCHEMA) as f:
        cur.executescript(f.read())

    with open(DATA) as f:
        data_sql = f.read()

    catalog_sections = data_sql.split('INSERT INTO expedientes')[0]
    cur.executescript(catalog_sections)

    rows_part = 'INSERT INTO expedientes' + data_sql.split('INSERT INTO expedientes')[1]
    lines = rows_part.split('\n')

    all_rows = []
    for line in lines:
        raw = parse_row_values(line)
        if raw is not None and len(raw) == len(COLUMNS):
            row = [parse_value(v) for v in raw]
            all_rows.append(row)

    insert_cols = ', '.join(COLUMNS)
    placeholders = ', '.join(['?' for _ in COLUMNS])
    update_set = ', '.join([f"{c} = ?" for c in COLUMNS])

    seen_empty = False
    insert_count = 0
    update_count = 0
    solped_to_id = {}

    for vals in all_rows:
        solped = vals[0]
        auto_linea = generar_auto_linea(cur, vals)

        if solped and solped.strip():
            if solped in solped_to_id:
                existing_id = solped_to_id[solped]
                obs_actual = cur.execute(
                    "SELECT observaciones_generales FROM expedientes WHERE id_expediente = ?",
                    (existing_id,)
                ).fetchone()[0]
                vals[OBS_COL] = (obs_actual + '\n' + auto_linea) if obs_actual else auto_linea
                vals.append(existing_id)
                try:
                    cur.execute(f"UPDATE expedientes SET {update_set} WHERE id_expediente = ?", vals)
                except Exception as e:
                    print(f"UPDATE err solped={solped}: {e}")
                    print(f"  vals={vals}")
                    raise
                update_count += 1
            else:
                vals[OBS_COL] = auto_linea
                try:
                    cur.execute(f"INSERT INTO expedientes ({insert_cols}) VALUES ({placeholders})", vals)
                except Exception as e:
                    print(f"INSERT err solped={solped}: {e}")
                    print(f"  vals={vals}")
                    raise
                new_id = cur.lastrowid
                solped_to_id[solped] = new_id
                insert_count += 1
        else:
            if not seen_empty:
                vals[OBS_COL] = auto_linea
                try:
                    cur.execute(f"INSERT INTO expedientes ({insert_cols}) VALUES ({placeholders})", vals)
                except Exception as e:
                    print(f"INSERT err solped=empty: {e}")
                    print(f"  vals={vals}")
                    raise
                seen_empty = True
                insert_count += 1
            else:
                cur.execute("SELECT id_expediente FROM expedientes WHERE (solped IS NULL OR solped = '') LIMIT 1")
                existing = cur.fetchone()
                if existing:
                    existing_id = existing[0]
                    obs_actual = cur.execute(
                        "SELECT observaciones_generales FROM expedientes WHERE id_expediente = ?",
                        (existing_id,)
                    ).fetchone()[0]
                    vals[OBS_COL] = (obs_actual + '\n' + auto_linea) if obs_actual else auto_linea
                    vals.append(existing_id)
                    try:
                        cur.execute(f"UPDATE expedientes SET {update_set} WHERE id_expediente = ?", vals)
                    except Exception as e:
                        print(f"UPDATE err solped=empty: {e}")
                        print(f"  vals={vals}")
                        raise
                    update_count += 1

    con.commit()

    hist_count = cur.execute("SELECT COUNT(*) FROM historial_movimientos").fetchone()[0]
    exp_count = cur.execute("SELECT COUNT(*) FROM expedientes").fetchone()[0]

    print(f"Expedientes: {exp_count} (insert={insert_count}, update={update_count})")
    print(f"Historial movimientos: {hist_count}")

    dups = cur.execute("SELECT solped, COUNT(*) FROM expedientes GROUP BY solped HAVING COUNT(*) > 1").fetchall()
    if dups:
        print(f"\nATENCIÓN: {len(dups)} solpeds con duplicados:")
        for s, c in dups:
            print(f"  {s}: {c} veces")

    print("\n--- Muestra de observaciones ---")
    for row in cur.execute("SELECT id_expediente, solped, observaciones_generales FROM expedientes WHERE observaciones_generales IS NOT NULL LIMIT 3"):
        print(f"  id={row[0]} solped={row[1]}: {row[2][:80]}...")
    for row in cur.execute("SELECT id_expediente, solped, observaciones_generales FROM historial_movimientos WHERE observaciones_generales IS NOT NULL AND observaciones_generales != '' LIMIT 3"):
        print(f"  HIST id={row[0]} solped={row[1]}: {row[2][:80]}...")

    con.close()

if __name__ == '__main__':
    main()
