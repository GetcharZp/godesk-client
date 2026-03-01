import {request} from "../utils/request.js";
import {GetSysConfig, Reconnect, SaveSysConfig, GetConnectionStatus} from "../../wailsjs/go/main/App.js";

// 获取系统配置
export const getSysConfig = () => {
    return GetSysConfig()
}

// 保存系统配置
export const saveSysConfig = (config) => {
    return SaveSysConfig(config)
}

// 重新连接服务
export const reconnect = () => {
    return Reconnect()
}

// 获取连接状态
export const getConnectionStatus = () => {
    return GetConnectionStatus()
}