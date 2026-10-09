export namespace app {
	
	export class Notification {
	    id: number;
	    kind: string;
	    code: string;
	    action?: string;
	    path?: string;
	    organize?: organizer.Result;
	    undo?: undo.Result;
	    // Go type: time
	    at: any;
	
	    static createFrom(source: any = {}) {
	        return new Notification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.code = source["code"];
	        this.action = source["action"];
	        this.path = source["path"];
	        this.organize = this.convertValues(source["organize"], organizer.Result);
	        this.undo = this.convertValues(source["undo"], undo.Result);
	        this.at = this.convertValues(source["at"], null);
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
	export class PreviewState {
	    status: string;
	    totalFiles: number;
	    folders: organizer.FolderCount[];
	    otherFiles: number;
	
	    static createFrom(source: any = {}) {
	        return new PreviewState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.totalFiles = source["totalFiles"];
	        this.folders = this.convertValues(source["folders"], organizer.FolderCount);
	        this.otherFiles = source["otherFiles"];
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
	export class ViewState {
	    version: number;
	    sourceFolderPath: string;
	    destinationFolderPath: string;
	    hasUndo: boolean;
	    busy: string;
	    unread: boolean;
	    preview: PreviewState;
	    settings: settings.View;
	    notifications: Notification[];
	
	    static createFrom(source: any = {}) {
	        return new ViewState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.sourceFolderPath = source["sourceFolderPath"];
	        this.destinationFolderPath = source["destinationFolderPath"];
	        this.hasUndo = source["hasUndo"];
	        this.busy = source["busy"];
	        this.unread = source["unread"];
	        this.preview = this.convertValues(source["preview"], PreviewState);
	        this.settings = this.convertValues(source["settings"], settings.View);
	        this.notifications = this.convertValues(source["notifications"], Notification);
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

export namespace organizer {
	
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

export namespace settings {
	
	export class Criterion {
	    key: string;
	    enabled: boolean;
	    locked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Criterion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.enabled = source["enabled"];
	        this.locked = source["locked"];
	    }
	}
	export class View {
	    language: string;
	    theme: string;
	    criteria: Criterion[];
	
	    static createFrom(source: any = {}) {
	        return new View(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.criteria = this.convertValues(source["criteria"], Criterion);
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

