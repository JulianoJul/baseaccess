export namespace main {
	
	export class RutaProcesosCronogramaEntry {
	    id: number;
	    id_junta_proceso: number;
	    fecha: string;
	    id_leyenda: number;
	    nota?: string;
	    status_name?: string;
	    hex_color?: string;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosCronogramaEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.id_junta_proceso = source["id_junta_proceso"];
	        this.fecha = source["fecha"];
	        this.id_leyenda = source["id_leyenda"];
	        this.nota = source["nota"];
	        this.status_name = source["status_name"];
	        this.hex_color = source["hex_color"];
	    }
	}
	export class RutaProcesosJuntaLeyenda {
	    id: number;
	    id_junta: number;
	    id_leyenda: number;
	    orden: number;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosJuntaLeyenda(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.id_junta = source["id_junta"];
	        this.id_leyenda = source["id_leyenda"];
	        this.orden = source["orden"];
	    }
	}
	export class RutaProcesosLegend {
	    id: number;
	    nombre: string;
	    color: string;
	    ambito: string;
	    id_hoja?: number;
	    bloqueado: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosLegend(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nombre = source["nombre"];
	        this.color = source["color"];
	        this.ambito = source["ambito"];
	        this.id_hoja = source["id_hoja"];
	        this.bloqueado = source["bloqueado"];
	    }
	}
	export class RutaProcesosJuntaProceso {
	    id: number;
	    id_junta: number;
	    numero: number;
	    proceso: string;
	    timeline: Record<string, Array<RutaProcesosCronogramaEntry>>;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosJuntaProceso(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.id_junta = source["id_junta"];
	        this.numero = source["numero"];
	        this.proceso = source["proceso"];
	        this.timeline = this.convertValues(source["timeline"], Array<RutaProcesosCronogramaEntry>, true);
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
	export class RutaProcesosJuntaSemana {
	    id: number;
	    id_junta: number;
	    numero: number;
	    fecha_inicio: string;
	    fecha_fin: string;
	    dias: string[];
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosJuntaSemana(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.id_junta = source["id_junta"];
	        this.numero = source["numero"];
	        this.fecha_inicio = source["fecha_inicio"];
	        this.fecha_fin = source["fecha_fin"];
	        this.dias = source["dias"];
	    }
	}
	export class RutaProcesosJunta {
	    id: number;
	    id_hoja: number;
	    numero: number;
	    consecutiva: number;
	    fecha: string;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosJunta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.id_hoja = source["id_hoja"];
	        this.numero = source["numero"];
	        this.consecutiva = source["consecutiva"];
	        this.fecha = source["fecha"];
	    }
	}
	export class RutaProcesosHoja {
	    id: number;
	    nombre: string;
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosHoja(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nombre = source["nombre"];
	    }
	}
	export class RutaProcesosGanttData {
	    hojas: RutaProcesosHoja[];
	    current_hoja?: RutaProcesosHoja;
	    juntas: RutaProcesosJunta[];
	    current_junta?: RutaProcesosJunta;
	    semanas: RutaProcesosJuntaSemana[];
	    procesos: RutaProcesosJuntaProceso[];
	    legend: RutaProcesosLegend[];
	    junta_legend: RutaProcesosJuntaLeyenda[];
	
	    static createFrom(source: any = {}) {
	        return new RutaProcesosGanttData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hojas = this.convertValues(source["hojas"], RutaProcesosHoja);
	        this.current_hoja = this.convertValues(source["current_hoja"], RutaProcesosHoja);
	        this.juntas = this.convertValues(source["juntas"], RutaProcesosJunta);
	        this.current_junta = this.convertValues(source["current_junta"], RutaProcesosJunta);
	        this.semanas = this.convertValues(source["semanas"], RutaProcesosJuntaSemana);
	        this.procesos = this.convertValues(source["procesos"], RutaProcesosJuntaProceso);
	        this.legend = this.convertValues(source["legend"], RutaProcesosLegend);
	        this.junta_legend = this.convertValues(source["junta_legend"], RutaProcesosJuntaLeyenda);
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

