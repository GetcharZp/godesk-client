import {request} from "../utils/request.js";
import {DeviceInfo} from "../../wailsjs/go/main/App.js";

// 设备信息
export async function getDeviceInfo() {
    return await request(DeviceInfo())
}
