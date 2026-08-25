export namespace configdir {
	
	export class FilePreview {
	    path: string;
	    content: string;
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FilePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.content = source["content"];
	        this.truncated = source["truncated"];
	    }
	}
	export class Node {
	    name: string;
	    path: string;
	    isDir: boolean;
	    children?: Node[];
	
	    static createFrom(source: any = {}) {
	        return new Node(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.children = this.convertValues(source["children"], Node);
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

export namespace contentgen {
	
	export class Article {
	    name: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new Article(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.body = source["body"];
	    }
	}
	export class Draft {
	    title: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new Draft(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.body = source["body"];
	    }
	}
	export class Library {
	    titleTemplate: string;
	    bodyTemplates: string[];
	    keywords: string[];
	    vars: string[][];
	    images: string[];
	    articles: Article[];
	
	    static createFrom(source: any = {}) {
	        return new Library(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.titleTemplate = source["titleTemplate"];
	        this.bodyTemplates = source["bodyTemplates"];
	        this.keywords = source["keywords"];
	        this.vars = source["vars"];
	        this.images = source["images"];
	        this.articles = this.convertValues(source["articles"], Article);
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
	export class Options {
	    count: number;
	    keywordOrder: string;
	    keywordTransform: string;
	    shuffleParagraphs: boolean;
	    dedupeLines: boolean;
	    chineseOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.keywordOrder = source["keywordOrder"];
	        this.keywordTransform = source["keywordTransform"];
	        this.shuffleParagraphs = source["shuffleParagraphs"];
	        this.dedupeLines = source["dedupeLines"];
	        this.chineseOnly = source["chineseOnly"];
	    }
	}

}

