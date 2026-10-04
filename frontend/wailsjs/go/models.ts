export namespace app {
	
	export class DroppedPath {
	    sourceFolderPath: string;
	
	    static createFrom(source: any = {}) {
	        return new DroppedPath(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceFolderPath = source["sourceFolderPath"];
	    }
	}
	export class OrganizationState {
	    hasUndo: boolean;
	    sourceFolderPath: string;
	    destinationFolderPath: string;
	
	    static createFrom(source: any = {}) {
	        return new OrganizationState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUndo = source["hasUndo"];
	        this.sourceFolderPath = source["sourceFolderPath"];
	        this.destinationFolderPath = source["destinationFolderPath"];
	    }
	}

}

export namespace criteria {
	
	export class RawOptions {
	    byDuration: boolean;
	    byPages: boolean;
	    byResolution: boolean;
	    byDate: boolean;
	    bySize: boolean;
	    byExtension?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RawOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.byDuration = source["byDuration"];
	        this.byPages = source["byPages"];
	        this.byResolution = source["byResolution"];
	        this.byDate = source["byDate"];
	        this.bySize = source["bySize"];
	        this.byExtension = source["byExtension"];
	    }
	}

}

export namespace organizer {
	
	export class Request {
	    sourceFolderPath: string;
	    destinationFolderPath: string;
	    organizationOptions: criteria.RawOptions;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceFolderPath = source["sourceFolderPath"];
	        this.destinationFolderPath = source["destinationFolderPath"];
	        this.organizationOptions = this.convertValues(source["organizationOptions"], criteria.RawOptions);
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
	export class Result {
	    sourceFolderPath: string;
	    destinationFolderPath: string;
	    processedFiles: number;
	    movedFiles: number;
	    failedFiles: number;
	    unchangedFiles: number;
	    ignoredWithoutExtension: number;
	    ignoredFolders: number;
	    canUndo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceFolderPath = source["sourceFolderPath"];
	        this.destinationFolderPath = source["destinationFolderPath"];
	        this.processedFiles = source["processedFiles"];
	        this.movedFiles = source["movedFiles"];
	        this.failedFiles = source["failedFiles"];
	        this.unchangedFiles = source["unchangedFiles"];
	        this.ignoredWithoutExtension = source["ignoredWithoutExtension"];
	        this.ignoredFolders = source["ignoredFolders"];
	        this.canUndo = source["canUndo"];
	    }
	}

}

export namespace undo {
	
	export class Result {
	    restoredFiles: number;
	    renamedOnRestore: number;
	    skippedMissing: number;
	    failedFiles: number;
	    canUndo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.restoredFiles = source["restoredFiles"];
	        this.renamedOnRestore = source["renamedOnRestore"];
	        this.skippedMissing = source["skippedMissing"];
	        this.failedFiles = source["failedFiles"];
	        this.canUndo = source["canUndo"];
	    }
	}

}

