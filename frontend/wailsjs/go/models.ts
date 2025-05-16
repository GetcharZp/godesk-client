export namespace godesk {
	
	export class UserLoginRequest {
	    username?: string;
	    password?: string;
	
	    static createFrom(source: any = {}) {
	        return new UserLoginRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class UserRegisterRequest {
	    username?: string;
	    password?: string;
	
	    static createFrom(source: any = {}) {
	        return new UserRegisterRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}

}

