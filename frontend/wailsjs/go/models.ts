export namespace main {
	
	export class RutaProcesosHoja {
	    id: number;
	    nombre: string;
	    fecha_inicio: string;
	    fecha_fin: string;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosHoja(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nombre = source["nombre"];
	        this.fecha_inicio = source["fecha_inicio"];
	        this.fecha_fin = source["fecha_fin"];
	    }
	}
	export class RutaProcesosProceso {
	    id: number;
	    modulo: string;
	    descripcion: string;
	    db_id: number;
	    activo: boolean;
	    solped: string;
	    estatus_detalle: string;
	    receptor: string;
	    timeline: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosProceso(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.modulo = source["modulo"];
	        this.descripcion = source["descripcion"];
	        this.db_id = source["db_id"];
	        this.activo = source["activo"];
	        this.solped = source["solped"];
	        this.estatus_detalle = source["estatus_detalle"];
	        this.receptor = source["receptor"];
	        this.timeline = source["timeline"];
	    }
	}
	export class RutaProcesosLegend {
	    id: number;
	    status_name: string;
	    hex_color: string;
	    orden: number;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosLegend(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status_name = source["status_name"];
	        this.hex_color = source["hex_color"];
	        this.orden = source["orden"];
	    }
	}
	export class RutaProcesosGanttData {
	    legend: RutaProcesosLegend[];
	    columns: any[];
	    processes: RutaProcesosProceso[];
	    hojas: RutaProcesosHoja[];
	    current_hoja?: RutaProcesosHoja;
	    offset_weeks: number;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosGanttData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.legend = this.convertValues(source["legend"], RutaProcesosLegend);
	        this.columns = source["columns"];
	        this.processes = this.convertValues(source["processes"], RutaProcesosProceso);
	        this.hojas = this.convertValues(source["hojas"], RutaProcesosHoja);
	        this.current_hoja = this.convertValues(source["current_hoja"], RutaProcesosHoja);
	        this.offset_weeks = source["offset_weeks"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

