export namespace main {
	
	export class RutaProcesosProceso {
	    id: number;
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
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosLegend(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status_name = source["status_name"];
	        this.hex_color = source["hex_color"];
	    }
	}
	export class RutaProcesosGanttData {
	    legend: RutaProcesosLegend[];
	    columns: any[];
	    processes: RutaProcesosProceso[];
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosGanttData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.legend = this.convertValues(source["legend"], RutaProcesosLegend);
	        this.columns = source["columns"];
	        this.processes = this.convertValues(source["processes"], RutaProcesosProceso);
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

