export namespace licensing {
	
	export class LicenseStatus {
	    is_licensed: boolean;
	    organization: string;
	    tier: string;
	    expires_at: string;
	    features: string[];
	
	    static createFrom(source: any = {}) {
	        return new LicenseStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_licensed = source["is_licensed"];
	        this.organization = source["organization"];
	        this.tier = source["tier"];
	        this.expires_at = source["expires_at"];
	        this.features = source["features"];
	    }
	}

}

export namespace validator {
	
	export class Issue {
	    resource: string;
	    path: string;
	    rule_id: string;
	    severity: string;
	    message: string;
	    diagnostic: string;
	    suggestion: string;
	    can_auto_fix: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Issue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.resource = source["resource"];
	        this.path = source["path"];
	        this.rule_id = source["rule_id"];
	        this.severity = source["severity"];
	        this.message = source["message"];
	        this.diagnostic = source["diagnostic"];
	        this.suggestion = source["suggestion"];
	        this.can_auto_fix = source["can_auto_fix"];
	    }
	}
	export class ValidationReport {
	    total_resources: number;
	    passed: boolean;
	    error_count: number;
	    warning_count: number;
	    issues: Issue[];
	    bundle_type?: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidationReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_resources = source["total_resources"];
	        this.passed = source["passed"];
	        this.error_count = source["error_count"];
	        this.warning_count = source["warning_count"];
	        this.issues = this.convertValues(source["issues"], Issue);
	        this.bundle_type = source["bundle_type"];
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

