export namespace define {
	
	export class SysConfig {
	    username: string;
	    token: string;
	    uuid: string;
	    password: string;
	    service_address: string;
	
	    static createFrom(source: any = {}) {
	        return new SysConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.token = source["token"];
	        this.uuid = source["uuid"];
	        this.password = source["password"];
	        this.service_address = source["service_address"];
	    }
	}

}

export namespace godesk {
	
	export class AddDeviceRequest {
	    code?: number;
	    password?: string;
	    remark?: string;
	
	    static createFrom(source: any = {}) {
	        return new AddDeviceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.password = source["password"];
	        this.remark = source["remark"];
	    }
	}
	export class DeleteDeviceRequest {
	    uuid?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteDeviceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uuid = source["uuid"];
	    }
	}
	export class EditDeviceRequest {
	    uuid?: string;
	    code?: number;
	    password?: string;
	    remark?: string;
	
	    static createFrom(source: any = {}) {
	        return new EditDeviceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uuid = source["uuid"];
	        this.code = source["code"];
	        this.password = source["password"];
	        this.remark = source["remark"];
	    }
	}
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

