import {request} from "../utils/request.js";
import {GetUserInfo, UserLogin, UserLogout, UserRegister} from "../../wailsjs/go/main/App.js";

// 获取用户信息
export async function getUserInfo() {
    return await request(GetUserInfo())
}

// 用户登录
export async function userLogin(req) {
    return await request(UserLogin(req))
}

// 退出登录
export async function userLogout() {
    return await request(UserLogout())
}

// 用户注册
export async function userRegister(req) {
    return await request(UserRegister(req))
}
