export namespace config {
	
	export class Config {
	    token: string;
	    owner: string;
	    repo: string;
	    branch: string;
	    dir: string;
	    autoCreate: boolean;
	    intervalSec: number;
	    retries: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.token = source["token"];
	        this.owner = source["owner"];
	        this.repo = source["repo"];
	        this.branch = source["branch"];
	        this.dir = source["dir"];
	        this.autoCreate = source["autoCreate"];
	        this.intervalSec = source["intervalSec"];
	        this.retries = source["retries"];
	    }
	}

}

export namespace main {
	
	export class QueueItem {
	    title: string;
	    repoPath: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.repoPath = source["repoPath"];
	    }
	}

}

