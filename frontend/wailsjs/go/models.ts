export namespace account {
	
	export class Account {
	    ck: string;
	    ua: string;
	    ip: string;
	    status: string;
	    success: number;
	    fail: number;
	    total: number;
	    bad: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ck = source["ck"];
	        this.ua = source["ua"];
	        this.ip = source["ip"];
	        this.status = source["status"];
	        this.success = source["success"];
	        this.fail = source["fail"];
	        this.total = source["total"];
	        this.bad = source["bad"];
	    }
	}

}

export namespace config {
	
	export class Config {
	    threads: number;
	    intervalSec: number;
	    perAccountCount: number;
	    failSwitchCount: number;
	    cycleRounds: number;
	    roundIntervalSec: number;
	    keywordSlots: number;
	    createRepo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.threads = source["threads"];
	        this.intervalSec = source["intervalSec"];
	        this.perAccountCount = source["perAccountCount"];
	        this.failSwitchCount = source["failSwitchCount"];
	        this.cycleRounds = source["cycleRounds"];
	        this.roundIntervalSec = source["roundIntervalSec"];
	        this.keywordSlots = source["keywordSlots"];
	        this.createRepo = source["createRepo"];
	    }
	}

}

export namespace main {
	
	export class PublishResult {
	    time: string;
	    ck: string;
	    title: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new PublishResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.ck = source["ck"];
	        this.title = source["title"];
	        this.value = source["value"];
	    }
	}
	export class QueueItem {
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	    }
	}

}

