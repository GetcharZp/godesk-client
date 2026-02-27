import {request} from "../utils/request.js";
import {DeviceInfo, GetDeviceList, AddDevice, EditDevice, DeleteDevice} from "../../wailsjs/go/main/App.js";

// 设备信息
export async function getDeviceInfo() {
    return await request(DeviceInfo())
}

// 获取设备列表
export async function getDeviceList() {
    return await request(GetDeviceList())
}

// 添加设备
export async function addDevice(req) {
    return await request(AddDevice(req))
}

// 编辑设备
export async function editDevice(req) {
    return await request(EditDevice(req))
}

// 删除设备
export async function deleteDevice(req) {
    return await request(DeleteDevice(req))
}
